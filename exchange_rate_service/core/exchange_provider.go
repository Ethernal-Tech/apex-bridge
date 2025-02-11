package core

type ExchangeProvider int

const (
	Binance ExchangeProvider = iota
	Kraken
	KuCoin
	DummyExchange
)

func (e ExchangeProvider) String() string {
	switch e {
	case Binance:
		return "Binance"
	case Kraken:
		return "Kraken"
	case KuCoin:
		return "KuCoin"
	case DummyExchange:
		return "DummyExchange"
	default:
		return "Unknown"
	}
}
