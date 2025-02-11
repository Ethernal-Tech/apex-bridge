package ratefetcher

import (
	"context"
	"fmt"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/exchange_rate_service/core"
	"github.com/Ethernal-Tech/apex-bridge/exchange_rate_service/fetchers"
	"github.com/Ethernal-Tech/apex-bridge/exchange_rate_service/model"
)

const (
	ADA  = "ADA"
	USD  = "USD"
	USDT = "USDT"
)

type RateFetcher struct {
	fetchers map[core.ExchangeProvider]core.ExchangeRateFetcher
}

func NewRateFetcher(config *core.ExchangeRateServiceConfig) *RateFetcher {
	return &RateFetcher{
		fetchers: map[core.ExchangeProvider]core.ExchangeRateFetcher{
			core.Binance: &fetchers.BinanceFetcher{},
			core.Kraken:  &fetchers.KrakenFetcher{},
			core.KuCoin:  &fetchers.KuCoinFetcher{},
			core.Dummy:   &fetchers.DummyFetcher{},
		},
	}
}

func (r *RateFetcher) FetchRateByExchange(ctx context.Context, exchange core.ExchangeProvider) error {
	fetcher, exists := r.fetchers[exchange]
	if !exists {
		return fmt.Errorf("unsupported exchange: %d", exchange)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(5 * time.Second):
			rate, err := fetcher.FetchRate(ctx, model.FetchRateParams{Base: USDT, Currency: ADA})
			if err != nil {
				return fmt.Errorf("error fetching rate from %d: %w", exchange, err)
			}

			fmt.Printf("Fetched rate from %s: %f\n", exchange.String(), rate)
		}
	}
}
