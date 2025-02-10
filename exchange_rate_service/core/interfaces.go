package core

import "github.com/Ethernal-Tech/apex-bridge/exchange_rate_service/model"

type ExchangeRateFetcher interface {
	FetchRate(params model.FetchRateParams) (float64, error)
}
