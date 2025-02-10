package fetchers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/exchange_rate_service/core"
	"github.com/Ethernal-Tech/apex-bridge/exchange_rate_service/model"
)

type BinanceFetcher struct{}

var _ core.ExchangeRateFetcher = (*BinanceFetcher)(nil)

func (b *BinanceFetcher) FetchRate(params model.FetchRateParams) (float64, error) {
	pair := params.Symbol + params.Currency
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s", pair)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)

	if err != nil {
		return 0, fmt.Errorf("failed to fetch price rate from Binance: %w", err)
	}

	defer resp.Body.Close()

	var binanceResponse model.BinanceResponse
	err = json.NewDecoder(resp.Body).Decode(&binanceResponse)

	if err != nil {
		return 0, fmt.Errorf("error decoding Binance response: %w", err)
	}

	price, err := strconv.ParseFloat(binanceResponse.Price, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert price from Binance to float: %w", err)
	}

	return price, nil
}
