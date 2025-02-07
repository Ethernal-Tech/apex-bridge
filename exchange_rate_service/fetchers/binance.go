package fetchers

import "github.com/Ethernal-Tech/apex-bridge/exchange_rate_service/core"

type BinanceFetcher struct{}

var _ core.ExchangeRateFetcher = (*BinanceFetcher)(nil)

func (b *BinanceFetcher) FetchRate() (float64, error) {
	return 0, nil
}
