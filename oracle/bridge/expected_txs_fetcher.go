package bridge

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/hashicorp/go-hclog"
)

const (
	TickTimeMs = 5000
)

type ExpectedTxsFetcherImpl struct {
	bridgeDataFetcher core.BridgeDataFetcher
	appConfig         *core.AppConfig
	db                core.BridgeExpectedCardanoTxsDb
	ctx               context.Context
	cancelCtx         context.CancelFunc
	logger            hclog.Logger
}

var _ core.ExpectedTxsFetcher = (*ExpectedTxsFetcherImpl)(nil)

func NewExpectedTxsFetcher(
	bridgeDataFetcher core.BridgeDataFetcher,
	appConfig *core.AppConfig,
	db core.BridgeExpectedCardanoTxsDb,
	logger hclog.Logger,
) *ExpectedTxsFetcherImpl {
	ctx, cancelCtx := context.WithCancel(context.Background())
	return &ExpectedTxsFetcherImpl{
		bridgeDataFetcher: bridgeDataFetcher,
		appConfig:         appConfig,
		db:                db,
		ctx:               ctx,
		cancelCtx:         cancelCtx,
		logger:            logger,
	}
}

func (f *ExpectedTxsFetcherImpl) Start() error {
	f.logger.Debug("Starting ExpectedTxsFetcher")

	timerTime := TickTimeMs * time.Millisecond
	ticker := time.NewTicker(timerTime)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			f.fetchData()
		case <-f.ctx.Done():
			return nil
		}
	}
}

func (f *ExpectedTxsFetcherImpl) Stop() error {
	f.logger.Debug("Stopping ExpectedTxsFetcher")
	f.cancelCtx()
	return nil
}

func (f *ExpectedTxsFetcherImpl) fetchData() error {
	var expectedTxs []*core.BridgeExpectedCardanoTx

	for chainId := range f.appConfig.CardanoChains {
		existingExpectedTxs, err := f.db.GetExpectedTxs(chainId, 0)
		if err != nil {
			f.logger.Error("Failed to GetExpectedTxs from db", "chainId", chainId, "err", err)
			continue
		}

		if len(existingExpectedTxs) > 0 {
			// no new batch can be executed until a claim is produced for the previous batch
			continue
		}

		expectedTx, err := f.bridgeDataFetcher.FetchExpectedTx(chainId)
		if err != nil {
			f.logger.Error("Failed to fetch expected tx from bridge", "chainId", chainId, "err", err)
			continue
		}

		if expectedTx != nil {
			expectedTxs = append(expectedTxs, expectedTx)
		}
	}

	if len(expectedTxs) > 0 {
		err := f.db.AddExpectedTxs(expectedTxs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add expected txs. error: %v\n", err)
			f.logger.Error("Failed to add expected txs", "err", err)
		}

		return fmt.Errorf("failed to add expected txs. err: %v", err)
	}

	return nil
}
