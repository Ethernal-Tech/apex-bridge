package core

type ExchangeRateFetcher interface {
	FetchRate() (float64, error)
}
