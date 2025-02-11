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

type KrakenFetcher struct{}

var _ core.ExchangeRateFetcher = (*KrakenFetcher)(nil)

func (k *KrakenFetcher) FetchRate(params model.FetchRateParams) (float64, error) {
	pair := params.Currency + params.Base
	url := fmt.Sprintf("https://api.kraken.com/0/public/Ticker?pair=%s", pair)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)

	if err != nil {
		return 0, fmt.Errorf("failed to fetch price rate from Kraken: %w", err)
	}

	defer resp.Body.Close()

	var kraken model.KrakenResponse
	err = json.NewDecoder(resp.Body).Decode(&kraken)

	if err != nil {
		return 0, fmt.Errorf("error decoding Kraken response: %w", err)
	}

	res := kraken.Result[pair]

	price, err := strconv.ParseFloat(res.C[0], 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert price from Kraken to float: %w", err)
	}

	return price, nil
}
