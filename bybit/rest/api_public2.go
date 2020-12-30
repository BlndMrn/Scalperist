package rest

import (
	"net/http"
)

// GetKLine2 (USDT永续)
// https://bybit-exchange.github.io/docs/zh-cn/linear/#t-querykline
// interval: 1 3 5 15 30 60 120 240 360 720 "D" "M" "W" "Y"
// from: From timestamp in seconds
// limit: Limit for data size per page, max size is 200. Default as showing 200 pieces of data per page
func (b *ByBit) GetKLine2(symbol string, interval string, from int64, limit int) (result []OHLC2, err error) {
	var ret GetKlineResult2
	params := map[string]interface{}{}
	params["symbol"] = symbol
	params["interval"] = interval
	params["from"] = from
	if limit > 0 {
		params["limit"] = limit
	}
	_, err = b.PublicRequest(http.MethodGet, "v2/public/kline/list", params, &ret)
	if err != nil {
		return
	}
	result = ret.Result
	return
}

//GetKLineFutures GetKLineFutures
func (b *ByBit) GetKLineFutures(symbol string, interval string, from int64, limit int) (result []OHLC2Futures, err error) {
	var ret GetKlineResult2Futures
	params := map[string]interface{}{}
	params["symbol"] = symbol
	params["interval"] = interval
	params["from"] = from
	if limit > 0 {
		params["limit"] = limit
	}
	_, err = b.PublicRequest(http.MethodGet, "public/linear/kline", params, &ret)
	if err != nil {
		return
	}
	result = ret.Result
	return
}
