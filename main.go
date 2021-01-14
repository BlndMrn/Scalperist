package main

import (
	"ScalperistBybit/bybit/usdt"
	"fmt"
)

//"ScalperisBybit/bybit"

//https://www.youtube.com/watch?v=ys-YOsoCF34

//Volatility 1 min scalp bot
//Bot name scalperist
//Изменение свечи от открытия до текущей цены более чем на 20 целых пунктов (17193.5 - 17173.5) влечет создание ордеров в противоположную сторону движения свечи. каждый следующий ордер больше предыдущего на 6%.
//(Можно наверно 7-8 %) шаг 5%
//Тейк на 0.13-0.15%.
//Ордера живы ровно 1 минуту после создания. Если позиция не создана, отмена ордеров

//UserAccount Users account from database
type UserAccount struct {
	key      string
	skey     string
	exchange string
	symbol   string
	risk     float64
}

func main() {
	var account UserAccount
	
	account.key, account.skey, account.exchange = "key", "skey", "Bybit"
	
	fmt.Println("Symbol: ")
	fmt.Scan(&account.symbol)

	switch account.symbol {
	case "BTCUSDT":
		account.risk = 2
	case "ETHUSDT":
		account.risk = 2
		account.key, account.skey = "", ""
	case "BCHUSDT":
		account.risk = 1
		account.key, account.skey = "", ""
	case "LINKUSDT":
		account.risk = 1
		account.key, account.skey = "", ""
	case "LTCUSDT":
		account.risk = 1
		account.key, account.skey = "", ""
	case "XTZUSDT":
		account.risk = 1
		account.key, account.skey = "", ""
	}
	go usdt.MainLogic(account.symbol, account.risk, account.key, account.skey, account.exchange, "https://api.bybit.com/")
	//account.symbols = "BTCUSD"
	//mainLogic(account.symbols, account.key, account.skey, account.exchange) //when using go it makes duplicate orders need test. Also rate isnt enough
	for {

	}
}
