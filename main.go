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
	//baseURL := "https://api.bybit.com/" // 主网络
	//b := rest.New(nil, baseURL, "9b6hMVhXjuOT4u94Sd", "0WGlSwICaAa0Uy2BZEh72AzvHogY7pL7bDXG", false) //scalperist@yandex.ru
	//key, skey := "9b6hMVhXjuOT4u94Sd", "0WGlSwICaAa0Uy2BZEh72AzvHogY7pL7bDXG"
	//baseURL := "https://api-testnet.bybit.com/" // 测试网络
	//b := rest.New(nil, baseURL, "EZZzge5Yi7devetcWC", "jqFcDNCTlAkPT6BypbVoAyqQMzsLeTaWbA1u", false)
	var account UserAccount
	//main ETH
	account.key, account.skey, account.exchange = "OkIoivdlojFI9hqSZ6", "edHKGzQ3dqbbtT28gIYJt3O05GlnFQ47ucoI", "Bybit"
	//account.symbols = "ETHUSDT"
	//main BTC
	//account.key, account.skey, account.exchange = "nCFx2rgcLEM7hZW446", "oZGzvxCneQWnFWwq5Rexn1mMyQYC4yhZnZ6Y", "Bybit"
	//account.symbols = "BTCUSD"

	//testnet
	//account.key, account.skey, account.exchange = "9b6hMVhXjuOT4u94Sd", "0WGlSwICaAa0Uy2BZEh72AzvHogY7pL7bDXG", "Bybit"
	//account.symbol = "BTCUSD"
	//data base logic
	fmt.Println("Symbol: ")
	fmt.Scan(&account.symbol)

	switch account.symbol {
	case "BTCUSDT":
		account.risk = 2
	case "ETHUSDT":
		account.risk = 2
		account.key, account.skey = "fjNsQPESdYpngeHlYH", "KKsNtFXsw03u0JpO4vqQ130PTdVDornFi1S9"
	case "BCHUSDT":
		account.risk = 1
		account.key, account.skey = "FUjXQHqo8dNvt6my9o", "xAjDIpPMCBHrGq9VKebS5fEen19ogOt0CnTY"
	case "LINKUSDT":
		account.risk = 1
		account.key, account.skey = "2GptAQmTBIKpuGUVDf", "DqtD1EoerbqmqLRbG26Ua7HnYx248muauo6F"
	case "LTCUSDT":
		account.risk = 1
		account.key, account.skey = "x0KCJTbFQirGN5tv4h", "xbGnm4jkr2UTpm6n9R8hZdBWXPhEW0Ixfm54"
	case "XTZUSDT":
		account.risk = 1
		account.key, account.skey = "x0KCJTbFQirGN5tv4h", "xbGnm4jkr2UTpm6n9R8hZdBWXPhEW0Ixfm54"
	}
	go usdt.MainLogic(account.symbol, account.risk, account.key, account.skey, account.exchange, "https://api.bybit.com/")
	//account.symbols = "BTCUSD"
	//mainLogic(account.symbols, account.key, account.skey, account.exchange) //when using go it makes duplicate orders need test. Also rate isnt enough
	for {

	}
}
