package telemetry

import (
	"fmt"

	"github.com/hashicorp/go-metrics"
)

const (
	oracleMetricsPrefix    = "oracle"
	batcherMetricsPrefix   = "batcher"
	indexersMetricsPrefix  = "indexers"
	hotWalletMetricsPrefix = "hotwallet"
)

func UpdateOracleTxsReceivedCounter(chain string, cnt int) {
	metrics.IncrCounter([]string{oracleMetricsPrefix, "txs_received_counter", chain}, float32(cnt))
}

func UpdateOracleClaimsSubmitCounter(cnt int) {
	metrics.IncrCounter([]string{oracleMetricsPrefix, "claims_submit_counter"}, float32(cnt))
}

func UpdateOracleClaimsInvalidCounter(chain string, cnt int) {
	metrics.IncrCounter([]string{oracleMetricsPrefix, "claims_invalid_counter", chain}, float32(cnt))
}

func UpdateOracleClaimsInvalidMetaDataCounter(chain string, cnt int) {
	metrics.IncrCounter([]string{oracleMetricsPrefix, "claims_invalid_metadata_counter", chain}, float32(cnt))
}

func UpdateBatcherBatchSubmitSucceeded(chain string, id uint64) {
	metrics.SetGauge([]string{batcherMetricsPrefix, "batch_submit_succeeded", chain}, float32(id))
}

func UpdateBatcherBatchSubmitFailed(chain string, id uint64) {
	metrics.SetGauge([]string{batcherMetricsPrefix, "batch_submit_failed", chain}, float32(id))
}

func UpdateIndexersBlockCounter(chain string, cnt int) {
	metrics.IncrCounter([]string{indexersMetricsPrefix, "block_counter", chain}, float32(cnt))
}

func UpdateHotWalletState(chain string, typeWallet string, val uint64) {
	stateLow := fmt.Sprintf("%s_%s_low", hotWalletMetricsPrefix, typeWallet)
	stateHigh := fmt.Sprintf("%s_%s_high", hotWalletMetricsPrefix, typeWallet)

	metrics.SetGauge([]string{batcherMetricsPrefix, stateHigh, chain}, float32(val>>32))
	metrics.SetGauge([]string{batcherMetricsPrefix, stateLow, chain}, float32(uint32(val))) //nolint:gosec
}
