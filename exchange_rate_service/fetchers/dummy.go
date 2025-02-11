package fetchers

import (
	"github.com/Ethernal-Tech/apex-bridge/exchange_rate_service/core"
	"github.com/Ethernal-Tech/apex-bridge/exchange_rate_service/model"
)

type DummyFetcher struct {
	config *core.ExchangeRateServiceConfig
}

var _ core.ExchangeRateFetcher = (*DummyFetcher)(nil)

func (d *DummyFetcher) FetchRate(parms model.FetchRateParams) (float64, error) {
	return 0, nil
}
