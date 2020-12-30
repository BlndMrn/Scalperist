package usdt

import (
	"ScalperistBybit/bybit/rest"
	"log"
	"math"
	"strconv"
	"time"
)

const (
	differ = 0.1 // volatiliy difference
	step   = 3   //step between qty of next order
)

//MainLogic MainLogic
func MainLogic(symbol string, risk float64, key string, skey string, exchange string, baseURL string) {
	//check for position entering
	var positionEntered bool
	//CancelAllOrders(symbol, key, skey)
	switch exchange {
	case "Bybit":
		for {
			position := GetPositions(symbol, key, skey, baseURL) //check active positions
			if position == true {
				log.Printf("Active position detected for %#v", symbol)
				positionEntered = true
				CreateMoreOrders(symbol, risk, key, skey, baseURL)
			} else {
				orders, _ := CheckOrders(symbol, key, skey, baseURL) //check open orders
				if orders == true {
					log.Printf("Waiting for fill for %#v", symbol)
					if positionEntered {
						CancelAllOrder(symbol, key, skey, baseURL) //delete orders without position after position closing
					}
				} else { //if no open positions or orders detected start orders creation
					positionEntered = CreateOrder(symbol, risk, key, skey, baseURL)
				}
			}
		}
	}
}

//CreateOrder open short/long positions
func CreateOrder(symbol string, risk float64, key string, skey string, baseURL string) (positionEntered bool) {
	var price, percent float64
	open, close := GetCurrentMarketPrice(symbol, key, skey, baseURL)
	log.Printf("Current price %#v %#v", symbol, close)
	if ((open-close)/open)*100 > 0.12 { //if open less than close on 0.10% start Long trade
		log.Printf("Starting long trade for %#v", symbol)
		qty := CalculateOrderQty(symbol, close, key, skey, risk, baseURL)
		percent = close * 0.1 / 100 //percent where to place new order
		price = CalculateOrderPrice(symbol, "Long", close, percent)
		for i := 0; i < 6; i++ {
			if i == 0 {
				log.Printf("Placed buy odrers from: %v for %#v", price, symbol)
			}
			OpenPosition("Long", price, Round(qty, 3), symbol, key, skey, baseURL)
			price = CalculateOrderPrice(symbol, "Long", price, percent)
			qty = qty + (qty * step / 100)
			positionEntered = false
		}
		go PositionCheck(symbol, key, skey, baseURL)
		for {
			orders, _ := CheckOrders(symbol, key, skey, baseURL)
			position := GetPositions(symbol, key, skey, baseURL)
			if orders || position {
				break
			}
			time.Sleep(time.Second * 2)
		}
	}
	if ((open-close)/open)*100 < -0.12 { // if open more than close on 0.10% start Short trade
		log.Printf("Starting short trade for %#v", symbol)
		qty := CalculateOrderQty(symbol, close, key, skey, risk, baseURL)
		percent = close * 0.1 / 100                                  //percent where to place new order
		price = CalculateOrderPrice(symbol, "Short", close, percent) // first order
		for i := 0; i < 6; i++ {
			if i == 0 {
				log.Printf("Placed sell odrers from: %v for %#v", price, symbol)
			}
			OpenPosition("Short", price, Round(qty, 3), symbol, key, skey, baseURL)
			price = CalculateOrderPrice(symbol, "Short", price, percent)
			qty = qty + (qty * step / 100)
			positionEntered = false
		}
		go PositionCheck(symbol, key, skey, baseURL)
		for {
			orders, _ := CheckOrders(symbol, key, skey, baseURL)
			position := GetPositions(symbol, key, skey, baseURL)
			if orders || position {
				break
			}
			time.Sleep(time.Second * 2)
		}
	}

	return positionEntered
}

//CreateMoreOrders limit to reduce order and more open position orders
func CreateMoreOrders(symbol string, risk float64, key string, skey string, baseURL string) {
	var tpOrder bool
	b := rest.New(nil, baseURL, key, skey, false)
	pos, err := b.GetPositionFutures(symbol)
	if err != nil {
		log.Printf("%v", err)
		return
	}
	//stoploss
	/*
		if pos.WalletBalance*0.1 <= math.Abs(pos.UnrealisedPnl) {
			b.CancelAllOrder(symbol)
			switch pos.Side {
			case "Buy":
				b.CreateOrderV2("Sell", "Market", pos.EntryPrice, int(pos.Size), "", 0, 0, true, false, "", symbol)
			case "Sell":
				b.CreateOrderV2("Buy", "Market", pos.EntryPrice, int(pos.Size), "", 0, 0, true, false, "", symbol)
			}
		}*/
	if pos[0].Size > 0 { //Long
		ordersNew, err := b.GetOrdersFutures("", "", 1, 20, "New", symbol)
		if err != nil {
			log.Printf("%v", err)
			return
		}
		for i := 0; i < len(ordersNew); i++ {
			if ordersNew[i].ReduceOnly == true {
				if pos[0].Size != ordersNew[i].Qty {
					b.CancelOrderFutures(ordersNew[i].OrderID, "", symbol)
					_, rate, _ := b.CreateOrderFutures("Sell", "Limit", CalculateOrderPrice(symbol, "Short", pos[0].EntryPrice, pos[0].EntryPrice*0.12/100), pos[0].Size, "PostOnly", 0, 0, true, "", "", false, "", symbol)
					log.Printf("Rate for: %#v", rate)
					tpOrder = true
					break
				} else if pos[0].Size == ordersNew[i].Qty {
					tpOrder = true
					break
				}
			}
		}
		if tpOrder == false {
			_, rate, _ := b.CreateOrderFutures("Sell", "Limit", CalculateOrderPrice(symbol, "Short", pos[0].EntryPrice, pos[0].EntryPrice*0.12/100), pos[0].Size, "PostOnly", 0, 0, true, "", "", false, "", symbol)
			log.Printf("Rate single: %#v", rate)
		}

		if len(ordersNew) < 5 && len(ordersNew) > 1 {
			qty := ordersNew[len(ordersNew)-1].Qty
			price := ordersNew[len(ordersNew)-1].Price
			for i := 0; i < len(ordersNew); i++ {
				if ordersNew[i].Price < price {
					price = ordersNew[i].Price
					qty = ordersNew[i].Qty
				}
			}
			for i := 0; i < 7; i++ {
				price = CalculateOrderPrice(symbol, "Long", price, price*0.1/100)
				qty = qty + (qty * step / 100)
				OpenPosition("Long", price, Round(qty, 4), symbol, key, skey, baseURL)
			}
			for {
				_, count := CheckOrders(symbol, key, skey, baseURL)
				log.Printf("orders count: %#v", count)
				if count > 3 {
					break
				}
				time.Sleep(time.Second * 2)
			}
		} else {
			if len(ordersNew) <= 1 {
				_, price := GetCurrentMarketPrice(symbol, key, skey, baseURL)
				qty := CalculateOrderQty(symbol, price, key, skey, risk, baseURL)
				for i := 0; i < 7; i++ {
					price = CalculateOrderPrice(symbol, "Long", price, price*0.1/100)
					qty = qty + (qty * step / 100)
					OpenPosition("Long", price, Round(qty, 4), symbol, key, skey, baseURL)
				}
				for {
					_, count := CheckOrders(symbol, key, skey, baseURL)
					log.Printf("orders count: %#v", count)
					if count > 3 {
						break
					}
					time.Sleep(time.Second * 2)
				}
			}
		}
	}
	if pos[1].Size > 0 { //Short
		ordersNew, err := b.GetOrdersFutures("", "", 1, 20, "New", symbol)
		if err != nil {
			log.Printf("%v", err)
			return
		}

		for i := 0; i < len(ordersNew); i++ {
			if ordersNew[i].ReduceOnly == true {
				if pos[1].Size != ordersNew[i].Qty {
					b.CancelOrderFutures(ordersNew[i].OrderID, "", symbol)
					_, rate, _ := b.CreateOrderFutures("Buy", "Limit", CalculateOrderPrice(symbol, "Long", pos[1].EntryPrice, pos[1].EntryPrice*0.12/100), pos[1].Size, "PostOnly", 0, 0, true, "", "", false, "", symbol)
					log.Printf("Rate for: %#v", rate)
					tpOrder = true
					break
				} else if pos[1].Size == ordersNew[i].Qty {
					tpOrder = true
					break
				}
			}
		}
		if tpOrder == false {
			_, rate, _ := b.CreateOrderFutures("Buy", "Limit", CalculateOrderPrice(symbol, "Long", pos[1].EntryPrice, pos[1].EntryPrice*0.12/100), pos[1].Size, "PostOnly", 0, 0, true, "", "", false, "", symbol)
			log.Printf("Rate single: %#v", rate)
		}
		if len(ordersNew) < 5 && len(ordersNew) > 1 {
			qty := ordersNew[len(ordersNew)-1].Qty
			price := ordersNew[len(ordersNew)-1].Price
			for i := 0; i < len(ordersNew); i++ {
				if ordersNew[i].Price > price {
					price = ordersNew[i].Price
					qty = ordersNew[i].Qty
				}
			}
			for i := 0; i < 7; i++ {
				price = CalculateOrderPrice(symbol, "Short", price, price*0.1/100)
				qty = qty + (qty * step / 100)
				OpenPosition("Short", price, Round(qty, 4), symbol, key, skey, baseURL)

			}
			for {
				_, count := CheckOrders(symbol, key, skey, baseURL)
				log.Printf("orders count: %#v", count)
				if count > 3 {
					break
				}
				time.Sleep(time.Second * 2)
			}
		} else {
			if len(ordersNew) <= 1 {
				_, price := GetCurrentMarketPrice(symbol, key, skey, baseURL)
				qty := CalculateOrderQty(symbol, price, key, skey, risk, baseURL)
				for i := 0; i < 7; i++ {
					price = CalculateOrderPrice(symbol, "Short", price, price*0.1/100)
					qty = qty + (qty * step / 100)
					OpenPosition("Short", price, Round(qty, 4), symbol, key, skey, baseURL)
				}
				for {
					_, count := CheckOrders(symbol, key, skey, baseURL)
					log.Printf("orders count: %#v", count)
					if count > 3 {
						break
					}
					time.Sleep(time.Second * 2)
				}
			}
		}
	}
}

//PositionCheck check position
func PositionCheck(symbol string, key string, skey string, baseURL string) {
	time.Sleep(time.Minute)
	position := GetPositions(symbol, key, skey, baseURL)
	if position == false {
		CancelAllOrder(symbol, key, skey, baseURL)
	}
}

//CalculateOrderQty calculate position size
func CalculateOrderQty(symbol string, price float64, key string, skey string, percent float64, baseURL string) (qty float64) {
	b := rest.New(nil, baseURL, key, skey, false)
	balance, err := b.GetWalletBalance("USDT")
	if err != nil {
		log.Printf("%v", err)
		return
	}
	qty = balance.AvailableBalance / price
	qty = Round(qty*percent/100, 4) * 10
	log.Printf("balance %#v, qty %#v", balance.AvailableBalance, qty)
	return qty
}

//GetCurrentMarketPrice get market data
func GetCurrentMarketPrice(symbol string, key string, skey string, baseURL string) (open float64, close float64) {
	b := rest.New(nil, baseURL, key, skey, false)
	//GetMarketData Api endpoint
	time := b.GetServerTime() //get server time
	timeStr := strconv.FormatInt(time, 10)
	if len(timeStr) > 10 {
		time, _ = strconv.ParseInt(timeStr[0:10], 10, 64)          //Delete more than 10 symbols
		market, err := b.GetKLineFutures(symbol, "1", time-120, 0) //Get data per 2 min
		if err != nil {
			log.Printf("%v", err)
			return
		}
		return market[len(market)-1].Open, market[len(market)-1].Close
	}
	return
}

//OpenPosition OpenPosition
func OpenPosition(side string, price float64, qty float64, symbol string, key string, skey string, baseURL string) (rate int) {
	b := rest.New(nil, baseURL, key, skey, false)

	switch side {
	case "Long":
		//open position long
		_, rate, err := b.CreateOrderFutures("Buy", "Limit", price, qty, "PostOnly", 0, 0, false, "", "", false, "", symbol)
		if err != nil {
			log.Printf("%v", err)
			return 0
		}
		log.Printf("Rate: %#v", rate)
	case "Short":
		//open position short
		_, rate, err := b.CreateOrderFutures("Sell", "Limit", price, qty, "PostOnly", 0, 0, false, "", "", false, "", symbol)
		if err != nil {
			log.Printf("%v", err)
			return 0
		}
		log.Printf("Rate: %#v", rate)
	}
	return rate
}

//GetPositions get current positions
func GetPositions(symbol string, key string, skey string, baseURL string) (res bool) {
	b := rest.New(nil, baseURL, key, skey, false)
	positions, err := b.GetPositionFutures(symbol)
	if err != nil {
		log.Printf("%v", err)
		return
	}
	if positions[0].EntryPrice != 0 || positions[1].EntryPrice != 0 {
		return true
	}
	return false
}

//CheckOrders get active orders
func CheckOrders(symbol string, key string, skey string, baseURL string) (res bool, count int) {
	b := rest.New(nil, baseURL, key, skey, false)
	orders, err := b.GetOrdersFutures("", "", 1, 20, "New", symbol)
	if err != nil {
		log.Printf("%v", err)
		return
	}
	for i := 0; i < len(orders); i++ {
		if orders[i].OrderStatus == "New" {
			count++
		}
	}
	if count > 0 {
		return true, count
	}
	return false, count
}

//CalculateOrderPrice 123
func CalculateOrderPrice(symbol string, side string, x float64, y float64) (price float64) {
	if side == "Long" {
		switch symbol {
		case "LINKUSDT":
			price = Round(x-y, 3) // first order
		case "XTZUSDT":
			price = Round(x-y, 3) // first order
		default:
			price = Round(x-y, 2) // first order BTCUSDT ETHUSDT BCHUSDT LTCUSDT
		}
	}
	if side == "Short" {
		switch symbol {
		case "LINKUSDT":
			price = Round(x+y, 3) // first order for LINKUSDT
		case "XTZUSDT":
			price = Round(x+y, 3) // first order XTZUSDT
		default:
			price = Round(x+y, 2) // first order BTCUSDT ETHUSDT BCHUSDT LTCUSDT
		}
	}
	return price
}

//Round with prec precision
func Round(x float64, prec int) float64 {
	var rounder float64
	pow := math.Pow(10, float64(prec))
	intermed := x * pow
	_, frac := math.Modf(intermed)
	if frac >= 0.5 {
		rounder = math.Ceil(intermed)
	} else {
		rounder = math.Floor(intermed)
	}

	return rounder / pow
}

//CancelAllOrder CancelAllOrderFutures
func CancelAllOrder(symbol string, key string, skey string, baseURL string) {
	b := rest.New(nil, baseURL, key, skey, false)
	b.CancelAllOrderFutures(symbol)
}
