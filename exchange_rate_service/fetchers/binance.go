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

func (b *BinanceFetcher) FetchRate(symbol string) (float64, error) {
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s", symbol)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("error fetching Binance price: %v", err)
	}
	defer resp.Body.Close()
	var binanceResponse model.BinanceResponse
	err = json.NewDecoder(resp.Body).Decode(&binanceResponse)
	if err != nil {
		return 0, fmt.Errorf("error decoding Binance response: %v", err)
	}

	price, err := strconv.ParseFloat(binanceResponse.Price, 64)
	if err != nil {
		return 0, fmt.Errorf("error converting Binance price to float: %v", err)
	}

	return price, nil
}
