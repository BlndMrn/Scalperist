package rest

type OHLC2 struct {
	Symbol   string  `json:"symbol"`
	Interval string  `json:"interval"`
	OpenTime int64   `json:"open_time"`
	Open     float64 `json:"open,string"`
	High     float64 `json:"high,string"`
	Low      float64 `json:"low,string"`
	Close    float64 `json:"close,string"`
	Volume   float64 `json:"volume,string"`
	Turnover float64 `json:"turnover,string"`
}

type GetKlineResult2 struct {
	RetCode int     `json:"ret_code"`
	RetMsg  string  `json:"ret_msg"`
	ExtCode string  `json:"ext_code"`
	ExtInfo string  `json:"ext_info"`
	Result  []OHLC2 `json:"result"`
	TimeNow string  `json:"time_now"`
}

type OHLC2Futures struct {
	Symbol   string  `json:"symbol"`
	Period   string  `json:"period"`
	Start    int64   `json:"start_at"`
	Volume   float64 `json:"volume"`
	Open     float64 `json:"open"`
	High     float64 `json:"high"`
	Low      float64 `json:"low"`
	Close    float64 `json:"close"`
	Interval string  `json:"interval"`
	OpenTime int64   `json:"open_time"`
	Turnover float64 `json:"turnover"`
}

type GetKlineResult2Futures struct {
	RetCode int            `json:"ret_code"`
	RetMsg  string         `json:"ret_msg"`
	ExtCode string         `json:"ext_code"`
	ExtInfo string         `json:"ext_info"`
	Result  []OHLC2Futures `json:"result"`
	TimeNow string         `json:"time_now"`
}
