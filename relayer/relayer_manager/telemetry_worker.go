package relayermanager

import (
	"context"
	"math/big"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	"github.com/hashicorp/go-hclog"
)

// TelemetryWorker periodically reports the balance of the accounts the relayer
// submits batches with. A chain is reported only if it is enabled in its configuration
type TelemetryWorker struct {
	reporters      map[string]core.BalanceReporter
	waitTime       time.Duration
	latestBalances map[string]*big.Int
	logger         hclog.Logger
}

func NewTelemetryWorker(
	operations map[string]core.ChainOperations,
	chains map[string]core.ChainConfig,
	waitTime time.Duration,
	logger hclog.Logger,
) *TelemetryWorker {
	reporters := make(map[string]core.BalanceReporter, len(operations))

	for chainID, ops := range operations {
		if !chains[chainID].Telemetry {
			continue
		}

		reporter, ok := ops.(core.BalanceReporter)
		if !ok {
			logger.Warn("Relayer balance telemetry enabled for a chain that can not report it", "chain", chainID)

			continue
		}

		reporters[chainID] = reporter

		logger.Info("Reporting relayer balance", "chain", chainID, "addr", reporter.GetRelayerAddress())
	}

	return &TelemetryWorker{
		reporters:      reporters,
		waitTime:       waitTime,
		latestBalances: map[string]*big.Int{},
		logger:         logger,
	}
}

func (tw *TelemetryWorker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(tw.waitTime):
			tw.execute(ctx)
		}
	}
}

func (tw *TelemetryWorker) execute(ctx context.Context) {
	for chainID, reporter := range tw.reporters {
		balance, err := reporter.GetRelayerBalance(ctx)
		if err != nil {
			tw.logger.Warn("failed to retrieve relayer balance", "chain", chainID, "err", err)

			continue
		}

		tw.logger.Debug("TELEMETRY: Relayer balance retrieved", "chain", chainID, "balance", balance)

		if cache := tw.latestBalances[chainID]; cache == nil || cache.Cmp(balance) != 0 {
			tw.latestBalances[chainID] = balance

			telemetry.UpdateRelayerBalance(chainID, common.WeiToDfm(balance).Uint64())
		}
	}
}
