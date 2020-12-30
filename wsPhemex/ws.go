package wsphemex

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/chuckpreslar/emission"
	"github.com/gorilla/websocket"
	"github.com/recws-org/recws"
	"github.com/tidwall/gjson"
)

const (
	MaxTryTimes = 10
)

// https://github.com/bybit-exchange/bybit-official-api-docs/blob/master/zh_cn/websocket.md

// 测试网地址
// wss://stream-testnet.bybit.com/realtime

// 主网地址
// wss://stream.bybit.com/realtime

const (
	HostReal = "wss://phemex.com/ws"
)

const (
	WSKLine = "kline.subscribe" // K线: kline.BTCUSD.1m

	WSDisconnected = "disconnected" // WS断开事件
)

type Configuration struct {
	Addr          string `json:"addr"`
	ApiKey        string `json:"api_key"`
	SecretKey     string `json:"secret_key"`
	AutoReconnect bool   `json:"auto_reconnect"`
	DebugMode     bool   `json:"debug_mode"`
}

type ByBitWS struct {
	cfg    *Configuration
	ctx    context.Context
	cancel context.CancelFunc
	conn   *recws.RecConn
	mu     sync.RWMutex

	subscribeCmds []Cmd

	emitter *emission.Emitter
}

func New(config *Configuration) *ByBitWS {
	b := &ByBitWS{
		cfg:     config,
		emitter: emission.NewEmitter(),
	}
	b.ctx, b.cancel = context.WithCancel(context.Background())
	b.conn = &recws.RecConn{
		KeepAliveTimeout: 60 * time.Second,
	}
	b.conn.SubscribeHandler = b.subscribeHandler
	return b
}

func (b *ByBitWS) subscribeHandler() error {
	log.Printf("subscribeHandler")

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cfg.ApiKey != "" && b.cfg.SecretKey != "" {
		err := b.Auth()
		if err != nil {
			log.Printf("auth error: %v", err)
		}
	}

	for _, cmd := range b.subscribeCmds {
		err := b.SendCmd(cmd)
		if err != nil {
			log.Printf("SendCmd return error: %v", err)
		}
	}

	return nil
}

func (b *ByBitWS) closeHandler(code int, text string) error {
	log.Printf("close handle executed code=%v text=%v",
		code, text)
	return nil
}

// IsConnected returns the WebSocket connection state
func (b *ByBitWS) IsConnected() bool {
	return b.conn.IsConnected()
}

func (b *ByBitWS) Subscribe(symbol string, interval int) {

	cmd := Cmd{
		Id:     135,
		Method: "kline.subscribe",
		Params: []interface{}{symbol, interval},
	}
	b.subscribeCmds = append(b.subscribeCmds, cmd)
	b.SendCmd(cmd)
}

func (b *ByBitWS) SendCmd(cmd Cmd) error {
	data, err := json.Marshal(cmd)
	if err != nil {
		log.Printf("cmd %#v", err)
		return err
	}
	return b.Send(string(data))
}

func (b *ByBitWS) Send(msg string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("send error: %v", r))
		}
	}()

	err = b.conn.WriteMessage(websocket.TextMessage, []byte(msg))
	return
}

func (b *ByBitWS) Start() error {
	b.connect()

	cancel := make(chan struct{})

	go func() {
		t := time.NewTicker(time.Second * 5)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				b.ping()
			case <-cancel:
				return
			}
		}
	}()

	go func() {
		defer close(cancel)

		for {
			messageType, data, err := b.conn.ReadMessage()
			if err != nil {
				log.Printf("Read error: %v", err)
				time.Sleep(100 * time.Millisecond)
				return
			}

			b.processMessage(messageType, data)
		}
	}()

	return nil
}

func (b *ByBitWS) connect() {
	b.conn.Dial(b.cfg.Addr, nil)
}

func (b *ByBitWS) ping() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ping error: %v", r)
		}
	}()

	if !b.IsConnected() {
		return
	}
	err := b.conn.WriteMessage(websocket.TextMessage, []byte(`{"id": 200, "method":"server.ping"}`))
	if err != nil {
		log.Printf("ping error: %v", err)
	}
}

func (b *ByBitWS) Auth() error {
	// 单位:毫秒
	expires := time.Now().Unix()*1000 + 10000
	req := fmt.Sprintf("GET/realtime%d", expires)
	sig := hmac.New(sha256.New, []byte(b.cfg.SecretKey))
	sig.Write([]byte(req))
	signature := hex.EncodeToString(sig.Sum(nil))

	cmd := Cmd{
		Method: "user.auth",
		Params: []interface{}{
			b.cfg.ApiKey,
			//fmt.Sprintf("%v", expires),
			expires,
			signature,
		},
		Id: 200,
	}
	err := b.SendCmd(cmd)
	return err
}

func (b *ByBitWS) processMessage(messageType int, data []byte) {
	ret := gjson.ParseBytes(data)

	if b.cfg.DebugMode {
		log.Printf("%v", string(data))
	}

	// 处理心跳包
	if ret.Get("ret_msg").String() == "pong" {
		b.handlePong()
	}

	var dataK KLineData
	err := json.Unmarshal([]byte(data), &dataK)
	if err != nil {
		log.Printf("Kline %v", err)
		return
	}

	if topicValue := ret.Get("type"); topicValue.Exists() {
		topic := topicValue.String()
		if strings.Contains(topic, "incremental") {
			var res KLine
			err := json.Unmarshal([]byte(data), &res)
			if err != nil {
				log.Printf("Kline %v", err)
				return
			}
			symbol := res.Symbol
			b.processKLine(symbol, res)
		}
		return
	}
}

func (b *ByBitWS) handlePong() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("handlePong error: %v", r))
		}
	}()
	pongHandler := b.conn.PongHandler()
	if pongHandler != nil {
		pongHandler("pong")
	}
	return nil
}

func (b *ByBitWS) Close() {
	b.conn.Close()
}
