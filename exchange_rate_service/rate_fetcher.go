package ratefetcher

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/exchange_rate_service/core"
	"github.com/Ethernal-Tech/apex-bridge/exchange_rate_service/fetchers"
)

type RateFetcher struct {
	fetchers map[core.ExchangeProvider]core.ExchangeRateFetcher
}

func NewRateFetcher() *RateFetcher {
	return &RateFetcher{
		fetchers: map[core.ExchangeProvider]core.ExchangeRateFetcher{
			core.Binance: &fetchers.BinanceFetcher{},
			core.Kraken:  &fetchers.Kraken{},
		},
	}
}

func (r *RateFetcher) GetRateByExchange(exchange core.ExchangeProvider) (float64, error) {
	fetcher, exists := r.fetchers[exchange]
	if !exists {
		return 0, fmt.Errorf("unsupported exchange: %d", exchange)
	}

	rate, err := fetcher.FetchRate()
	if err != nil {
		return 0, fmt.Errorf("error fetching rate from %d: %w", exchange, err)
	}

	return rate, nil
}
