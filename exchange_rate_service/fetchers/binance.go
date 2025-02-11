package fetchers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/exchange_rate_service/core"
	"github.com/Ethernal-Tech/apex-bridge/exchange_rate_service/model"
)

type BinanceFetcher struct{}

var _ core.ExchangeRateFetcher = (*BinanceFetcher)(nil)

func (b *BinanceFetcher) FetchRate(ctx context.Context, params model.FetchRateParams) (float64, error) {
	pair := params.Currency + params.Base
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s", pair)

	binanceResponse, err := common.HTTPGet[*model.BinanceResponse](ctx, url)
	if err != nil {
		return 0, err
	}

	price, err := strconv.ParseFloat(binanceResponse.Price, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert price from Binance to float: %w", err)
	}

	return price, nil
}
