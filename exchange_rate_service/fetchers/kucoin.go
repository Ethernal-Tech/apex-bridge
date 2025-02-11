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

type KuCoin struct{}

var _ core.ExchangeRateFetcher = (*KuCoin)(nil)

func (k *KuCoin) FetchRate(params model.FetchRateParams) (float64, error) {
	url := fmt.Sprintf("https://api.kucoin.com/api/v1/prices?base=%s&currencies=%s", params.Base, params.Currency)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)

	if err != nil {
		return 0, fmt.Errorf("failed to fetch price rate from KuCoin: %w", err)
	}

	defer resp.Body.Close()

	var kuCoin model.KuCoinResponse
	err = json.NewDecoder(resp.Body).Decode(&kuCoin)

	if err != nil {
		return 0, fmt.Errorf("error decoding KuCoin response: %w", err)
	}

	price, ok := kuCoin.Data[params.Currency]
	if !ok {
		return 0, fmt.Errorf("no price found in KuCoin response: %w", err)
	}

	priceF, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert price from Kraken to float: %w", err)
	}

	return priceF, nil
}
