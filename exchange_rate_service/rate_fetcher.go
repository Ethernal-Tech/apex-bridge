package ratefetcher

import (
	"context"
	"fmt"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/exchange_rate_service/core"
	"github.com/Ethernal-Tech/apex-bridge/exchange_rate_service/fetchers"
)

type RateFetcher struct {
	ctx      context.Context
	fetchers map[core.ExchangeProvider]core.ExchangeRateFetcher
}

func NewRateFetcher(ctx context.Context) *RateFetcher {
	return &RateFetcher{
		ctx: ctx,
		fetchers: map[core.ExchangeProvider]core.ExchangeRateFetcher{
			core.Binance: &fetchers.BinanceFetcher{},
			core.Kraken:  &fetchers.Kraken{},
		},
	}
}

func (r *RateFetcher) FetchRateByExchange(exchange core.ExchangeProvider) error {
	fetcher, exists := r.fetchers[exchange]
	if !exists {
		return fmt.Errorf("unsupported exchange: %d", exchange)
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return nil
		case <-ticker.C:
			rate, err := fetcher.FetchRate("ADAUSDT")
			if err != nil {
				return fmt.Errorf("error fetching rate from %d: %w", exchange, err)
			}

			fmt.Printf("Fetched rate from %d: %f\n", exchange, rate)
		}
	}
}
