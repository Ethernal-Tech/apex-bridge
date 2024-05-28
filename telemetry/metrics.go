package telemetry

import (
	"github.com/armon/go-metrics"
)

const (
	oracleMetricsPrefix  = "oracle"
	batcherMetricsPrefix = "batcher"
)

func UpdateOracleTxsReceivedCounter(chain string, cnt int) {
	metrics.IncrCounter([]string{oracleMetricsPrefix, "txs_received_counter", chain}, float32(cnt))
}

func UpdateOracleClaimsSubmitCounter(cnt int) {
	metrics.IncrCounter([]string{oracleMetricsPrefix, "claims_submit_counter"}, float32(cnt))
}

func UpdateOracleClaimsInvalidCounter(cnt int) {
	metrics.IncrCounter([]string{oracleMetricsPrefix, "claims_invalid_counter"}, float32(cnt))
}

func UpdateBatcherBatchSubmitSucceeded(chain string, id uint64) {
	metrics.SetGauge([]string{batcherMetricsPrefix, "batch_submit_succeeded", chain}, float32(id))
}

func UpdateBatcherBatchSubmitFailed(chain string, id uint64) {
	metrics.SetGauge([]string{batcherMetricsPrefix, "batch_submit_failed", chain}, float32(id))
}
