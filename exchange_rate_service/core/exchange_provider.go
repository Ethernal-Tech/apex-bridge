package core

type ExchangeProvider int

const (
	Binance ExchangeProvider = iota
	Kraken
	KuCoin
	Dummy
)

func (e ExchangeProvider) String() string {
	switch e {
	case Binance:
		return "Binance"
	case Kraken:
		return "Kraken"
	case KuCoin:
		return "KuCoin"
	case Dummy:
		return "Dummy"
	default:
		return "Unknown"
	}
}
