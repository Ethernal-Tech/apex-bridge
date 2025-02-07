package core

type ExchangeRateFetcher interface {
	FetchRate(symbol string) (float64, error)
}
