package core

import (
	"context"

	"github.com/Ethernal-Tech/apex-bridge/exchange_rate_service/model"
)

type ExchangeRateFetcher interface {
	FetchRate(ctx context.Context, params model.FetchRateParams) (float64, error)
}
