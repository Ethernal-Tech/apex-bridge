package fetchers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/exchange_rate_service/core"
	"github.com/Ethernal-Tech/apex-bridge/exchange_rate_service/model"
)

type KrakenFetcher struct{}

var _ core.ExchangeRateFetcher = (*KrakenFetcher)(nil)

func (k *KrakenFetcher) FetchRate(ctx context.Context, params model.FetchRateParams) (float64, error) {
	pair := params.Currency + params.Base
	url := fmt.Sprintf("https://api.kraken.com/0/public/Ticker?pair=%s", pair)

	krakenResponse, err := common.HTTPGet[*model.KrakenResponse](ctx, url)
	if err != nil {
		return 0, err
	}

	res, ok := krakenResponse.Result[pair]
	if !ok {
		return 0, fmt.Errorf("no price found in Kraken response: %w", err)
	}

	price, err := strconv.ParseFloat(res.C[0], 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert price from Kraken to float: %w", err)
	}

	return price, nil
}
