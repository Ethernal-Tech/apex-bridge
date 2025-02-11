package fetchers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/exchange_rate_service/core"
	"github.com/Ethernal-Tech/apex-bridge/exchange_rate_service/model"
)

type KuCoinFetcher struct{}

var _ core.ExchangeRateFetcher = (*KuCoinFetcher)(nil)

func (k *KuCoinFetcher) FetchRate(ctx context.Context, params model.FetchRateParams) (float64, error) {
	url := fmt.Sprintf("https://api.kucoin.com/api/v1/prices?base=%s&currencies=%s", params.Base, params.Currency)

	kuCoinResponse, err := common.HTTPGet[*model.KuCoinResponse](ctx, url)
	if err != nil {
		return 0, err
	}

	res, ok := kuCoinResponse.Data[params.Currency]
	if !ok {
		return 0, fmt.Errorf("no price found in KuCoin response: %w", err)
	}

	price, err := strconv.ParseFloat(res, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert price from KuCoin to float: %w", err)
	}

	return price, nil
}
