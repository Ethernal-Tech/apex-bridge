package core

type ExchangeProvider int

const (
	Binance ExchangeProvider = iota
	Kraken
	Coinbase
	DummyExchange
)
