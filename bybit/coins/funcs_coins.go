package coins

import (
	"ScalperistBybit/bybit/rest"
	"log"
	"math"
	"strconv"
	"time"
)

/*
GetCurrentMarketPrice
CalculateOrderQty
CalculateOrderPrice
OpenPosition
PositionCheck

GetKLine2
Rework for websocket

GetPosition
Rework for USDT futures +


*/

const differ = 0.1

func mainLogic(symbol string, key string, skey string, exchange string, baseURL string) {
	var positionEntered bool //check for position entering
	//CancelAllOrders(symbol, key, skey)

	switch exchange {
	case "Bybit":
		for {
			position := GetPositions(symbol, key, skey, baseURL) //check active positions
			if position == true {
				log.Printf("Active position detected for %#v", symbol)
				positionEntered = true
				CreateMoreOrders(symbol, key, skey, baseURL)
			} else {
				orders, _ := GetOrders(symbol, key, skey, baseURL) //check open orders
				if orders == true {
					log.Printf("Waiting for fill for %#v", symbol)
					if positionEntered {
						CreateMoreOrders(symbol, key, skey, baseURL) //delete orders without position after position closing
					}
				} else { //if no open positions or orders detected start orders creation
					positionEntered = CreateOrder(symbol, key, skey, baseURL)
				}
			}
			time.Sleep(time.Microsecond * 1500)
		}
	}
}

//CreateOrder open short/long positions
func CreateOrder(symbol string, key string, skey string, baseURL string) (positionEntered bool) {
	var price, percent float64
	open, close := GetCurrentMarketPrice(symbol, key, skey, baseURL)
	log.Print("Get market data for ", symbol)
	log.Printf("Current price %#v", close)
	qty := CalculateOrderQty(symbol, key, skey, 3, baseURL)
	if ((open-close)/open)*100 < -0.12 { //if open less than close on 0.10% start Long trade
		log.Printf("Volatility more than 1 for %#v", symbol)
		percent = close * 0.1 / 100 //percent where to place new order
		price = CalculateOrderPrice(symbol, "Long", close, percent)
		for i := 0; i < 6; i++ {
			if i == 0 {
				log.Printf("Placed buy odrer from: %v for %#v", price, symbol)
			}
			OpenPosition("Long", price, int(math.Round(qty)), symbol, key, skey, baseURL)
			price = CalculateOrderPrice(symbol, "Long", price, percent)
			qty = qty + (qty * 3 / 100)
			positionEntered = false
			go PositionCheck(symbol, key, skey, baseURL)
		}
	}
	if ((open-close)/open)*100 > 0.12 { // if open more than close on 0.10% start Short trade
		log.Printf("Volatility less than -1 for %#v", symbol)
		percent = close * 0.1 / 100                                  //percent where to place new order
		price = CalculateOrderPrice(symbol, "Short", close, percent) // first order
		for i := 0; i < 6; i++ {
			if i == 0 {
				log.Printf("Placed sell odrer from: %v for %#v", price, symbol)
			}
			OpenPosition("Short", price, int(math.Round(qty)), symbol, key, skey, baseURL)
			price = CalculateOrderPrice(symbol, "Short", price, percent)
			qty = qty + (qty * 3 / 100)
			positionEntered = false
			go PositionCheck(symbol, key, skey, baseURL)
		}
	}

	return positionEntered
}

//CreateMoreOrders limit to reduce order and more open position orders
func CreateMoreOrders(symbol string, key string, skey string, baseURL string) {
	b := rest.New(nil, baseURL, key, skey, false)
	pos, err := b.GetPosition(symbol)
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
	if pos.Size == 0 {
		b.CancelAllOrder(symbol)
	}
	if pos.Side == "Buy" {
		ordersNew, err := b.GetOrders("", "", 1, 20, "New", symbol)
		if err != nil {
			log.Printf("%v", err)
			return
		}
		needProfit := false
		for i := 0; i < len(ordersNew); i++ {
			if ordersNew[i].Side == "Sell" {
				if pos.Size != ordersNew[i].Qty {
					_, rate, _ := b.CreateOrderV2("Sell", "Limit", CalculateOrderPrice(symbol, "Short", pos.EntryPrice, pos.EntryPrice*0.12/100), int(pos.Size), "PostOnly", 0, 0, true, false, "", symbol)
					log.Printf("Rate: %#v", rate)
				}
				needProfit = false
				break
			} else {
				needProfit = true
			}
		}
		if needProfit {
			pos, rate, err := b.CreateOrderV2("Sell", "Limit", CalculateOrderPrice(symbol, "Short", pos.EntryPrice, pos.EntryPrice*0.12/100), int(pos.Size), "PostOnly", 0, 0, true, false, "", symbol)
			if err != nil {
				log.Printf("Rate: %#v", pos)
			}
			log.Printf("Rate: %#v", rate)
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
				qty = qty + (qty * 3 / 100)
				OpenPosition("Long", price, int(math.Round(qty)), symbol, key, skey, baseURL)
			}
		} else {
			if len(ordersNew) <= 1 {
				_, price := GetCurrentMarketPrice(symbol, key, skey, baseURL)
				qty := CalculateOrderQty(symbol, key, skey, 3, baseURL)
				for i := 0; i < 7; i++ {
					price = CalculateOrderPrice(symbol, "Long", price, price*0.1/100)
					qty = qty + (qty * 3 / 100)
					OpenPosition("Long", price, int(math.Round(qty)), symbol, key, skey, baseURL)
				}
			}
		}
	}
	if pos.Side == "Sell" {
		ordersNew, err := b.GetOrders("", "", 1, 20, "New", symbol)
		if err != nil {
			log.Printf("%v", err)
			return
		}
		needProfit := false
		for i := 0; i < len(ordersNew); i++ {
			if ordersNew[i].Side == "Buy" {
				if pos.Size != ordersNew[i].Qty {
					_, rate, _ := b.CreateOrderV2("Buy", "Limit", CalculateOrderPrice(symbol, "Long", pos.EntryPrice, pos.EntryPrice*0.12/100), int(pos.Size), "PostOnly", 0, 0, true, false, "", symbol)
					log.Printf("Rate: %#v", rate)
				}
				needProfit = false
				break
			} else {
				needProfit = true
			}
		}
		if needProfit {
			_, rate, _ := b.CreateOrderV2("Buy", "Limit", CalculateOrderPrice(symbol, "Long", pos.EntryPrice, pos.EntryPrice*0.12/100), int(pos.Size), "PostOnly", 0, 0, true, false, "", symbol)
			log.Printf("Rate: %#v", rate)
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
				qty = qty + (qty * 3 / 100)
				OpenPosition("Short", price, int(math.Round(qty)), symbol, key, skey, baseURL)

			}
		} else {
			if len(ordersNew) <= 1 {
				_, price := GetCurrentMarketPrice(symbol, key, skey, baseURL)
				qty := CalculateOrderQty(symbol, key, skey, 3, baseURL)
				for i := 0; i < 7; i++ {
					price = CalculateOrderPrice(symbol, "Short", price, price*0.1/100)
					qty = qty + (qty * 3 / 100)
					OpenPosition("Short", price, int(math.Round(qty)), symbol, key, skey, baseURL)
				}
			}
		}

	}

}

//PositionCheck check position
func PositionCheck(symbol string, key string, skey string, baseURL string) {
	time.Sleep(time.Minute)
	b := rest.New(nil, baseURL, key, skey, false)
	position := GetPositions(symbol, key, skey, baseURL)
	if position == false {
		b.CancelAllOrder(symbol)
	}
}

//CalculateOrderQty calculate position size
func CalculateOrderQty(symbol string, key string, skey string, percent float64, baseURL string) (qty float64) {
	b := rest.New(nil, baseURL, key, skey, false)
	position, err := b.GetPosition(symbol)
	if err != nil {
		log.Printf("%v", err)
		return
	}
	_, pirce := GetCurrentMarketPrice(symbol, key, skey, baseURL)
	usd := position.WalletBalance * pirce
	qty = math.Round(usd * percent / 100)
	return qty
}

//GetCurrentMarketPrice get market data
func GetCurrentMarketPrice(symbol string, key string, skey string, baseURL string) (open float64, close float64) {
	b := rest.New(nil, baseURL, key, skey, false)
	//GetMarketData Api endpoint
	time := b.GetServerTime() //get server time
	timeStr := strconv.FormatInt(time, 10)
	if len(timeStr) > 10 {
		time, _ = strconv.ParseInt(timeStr[0:10], 10, 64)    //Delete more than 10 symbols
		market, err := b.GetKLine2(symbol, "1", time-120, 0) //Get data per 2 min
		if err != nil {
			log.Printf("%v", err)
			return
		}
		return market[len(market)-1].Open, market[len(market)-1].Close
	}
	return
}

//OpenPosition OpenPosition
func OpenPosition(side string, price float64, qty int, symbol string, key string, skey string, baseURL string) (rate int) {
	b := rest.New(nil, baseURL, key, skey, false)

	switch side {
	case "Long":
		//open position long
		_, rate, err := b.CreateOrderV2("Buy", "Limit", price, qty, "PostOnly", 0, 0, false, false, "", symbol)
		if err != nil {
			log.Printf("%v", err)
			return 0
		}
		log.Printf("Rate: %#v", rate)
	case "Short":
		//open position short
		_, rate, err := b.CreateOrderV2("Sell", "Limit", price, qty, "PostOnly", 0, 0, false, false, "", symbol)
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
	positions, err := b.GetPosition(symbol)
	if err != nil {
		log.Printf("%v", err)
		return
	}
	if positions.EntryPrice != 0 {
		return true
	}
	return false
}

//GetOrders get active orders
func GetOrders(symbol string, key string, skey string, baseURL string) (res bool, count int) {
	b := rest.New(nil, baseURL, key, skey, false)
	orders, err := b.GetOrders("", "", 1, 20, "New", symbol)
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
		case "BTCUSD":
			price = math.Round(x - y) // first order
		case "ETHUSD":
			price = Round(x-y, 2) // first order
		}
	}
	if side == "Short" {
		switch symbol {
		case "BTCUSD":
			price = math.Round(x + y) // first order
		case "ETHUSD":
			price = Round(x+y, 2) // first order
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

//CancelAllOrders CancelAllOrders
func CancelAllOrders(symbol string, key string, skey string, baseURL string) {
	b := rest.New(nil, baseURL, key, skey, false)
	b.CancelAllOrder(symbol)
}
