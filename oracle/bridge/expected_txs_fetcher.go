package bridge

import (
	"context"
	"fmt"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/hashicorp/go-hclog"
)

const (
	TickTimeMs = 5000
)

type ExpectedTxsFetcherImpl struct {
	ctx               context.Context
	bridgeDataFetcher core.BridgeDataFetcher
	appConfig         *core.AppConfig
	db                core.BridgeExpectedCardanoTxsDb
	logger            hclog.Logger
}

var _ core.ExpectedTxsFetcher = (*ExpectedTxsFetcherImpl)(nil)

func NewExpectedTxsFetcher(
	ctx context.Context,
	bridgeDataFetcher core.BridgeDataFetcher,
	appConfig *core.AppConfig,
	db core.BridgeExpectedCardanoTxsDb,
	logger hclog.Logger,
) *ExpectedTxsFetcherImpl {
	return &ExpectedTxsFetcherImpl{
		ctx:               ctx,
		bridgeDataFetcher: bridgeDataFetcher,
		appConfig:         appConfig,
		db:                db,
		logger:            logger,
	}
}

func (f *ExpectedTxsFetcherImpl) Start() {
	f.logger.Debug("Starting ExpectedTxsFetcher")

	ticker := time.NewTicker(TickTimeMs * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-f.ctx.Done():
			return
		case <-ticker.C:
			err := f.fetchData()
			if err != nil {
				f.logger.Error("error while fetching data", "err", err)
			}
		}
	}
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
			f.logger.Error("failed to add expected txs", "err", err)
			return fmt.Errorf("failed to add expected txs. err: %w", err)
		}
	}

	return nil
}
