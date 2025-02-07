package fetchers

import "github.com/Ethernal-Tech/apex-bridge/exchange_rate_service/core"

type Kraken struct{}

var _ core.ExchangeRateFetcher = (*Kraken)(nil)

func (k *Kraken) FetchRate(pair string) (float64, error) {
	return 0, nil
}
