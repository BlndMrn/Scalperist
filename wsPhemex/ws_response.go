package wsphemex

type KLine struct {
	Data     [][]int64 `json:"kline"` // 563
	Sequence int64     `json:"sequence"`
	Symbol   string    `json:"symbol"` // 0.0013844
	Type     string    `json:"type"`   // 1m
}

type KLineData struct {
	Timestamp string
	Interval  string
	LastClose string
	OpenEp    string
	HighEp    float64
	LowEp     float64
	CloseEp   float64
	Volume    float64
	Turnover  float64
}
