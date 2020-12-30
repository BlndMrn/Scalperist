package wsphemex

func (b *ByBitWS) processKLine(symbol string, data KLine) {
	b.Emit(WSKLine, symbol, data)
}
