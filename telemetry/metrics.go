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
	relayerMetricsPrefix   = "relayer"
)

func UpdateOracleTxsReceivedCounter(chain string, cnt int) {
	metrics.IncrCounter([]string{oracleMetricsPrefix, "txs_received_counter", chain}, float32(cnt))
}

func UpdateOracleClaimsSubmitCounter(cnt int) {
	metrics.IncrCounter([]string{oracleMetricsPrefix, "claims_submit_counter"}, float32(cnt))
}

func UpdateOracleClaimsSubmitFailedCounter(chain string, cnt int) {
	metrics.IncrCounter([]string{oracleMetricsPrefix, "claims_submit_failed_counter", chain}, float32(cnt))
}

func UpdateOracleClaimsInvalidCounter(chain string, cnt int) {
	metrics.IncrCounter([]string{oracleMetricsPrefix, "claims_invalid_counter", chain}, float32(cnt))
}

func UpdateOracleRefundRequestCounter(chain string, cnt int) {
	if cnt <= 0 {
		return
	}

	metrics.IncrCounter([]string{oracleMetricsPrefix, "refund_request_counter", chain}, float32(cnt))
}

func UpdateOracleRefundRetryCounter(chain string, cnt int) {
	if cnt <= 0 {
		return
	}

	metrics.IncrCounter([]string{oracleMetricsPrefix, "refund_retry_counter", chain}, float32(cnt))
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

func UpdateIndexersFinalizedBlockCounter(chain string, cnt int) {
	metrics.IncrCounter([]string{indexersMetricsPrefix, "finalized_block_counter", chain}, float32(cnt))
}

func UpdateHotWalletState(chain string, typeWallet string, val uint64) {
	stateLow := fmt.Sprintf("%s_%s_low", hotWalletMetricsPrefix, typeWallet)
	stateHigh := fmt.Sprintf("%s_%s_high", hotWalletMetricsPrefix, typeWallet)

	metrics.SetGauge([]string{batcherMetricsPrefix, stateHigh, chain}, float32(val>>32))
	metrics.SetGauge([]string{batcherMetricsPrefix, stateLow, chain}, float32(uint32(val))) //nolint:gosec
}

func UpdateRelayerBalance(chain string, val uint64) {
	metrics.SetGauge([]string{relayerMetricsPrefix, "balance_high", chain}, float32(val>>32))
	metrics.SetGauge([]string{relayerMetricsPrefix, "balance_low", chain}, float32(uint32(val))) //nolint:gosec
}

func UpdateRelayerSendTxFailed(chain string) {
	metrics.IncrCounter([]string{relayerMetricsPrefix, "send_tx_failed_counter", chain}, 1)
}
