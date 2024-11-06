package bridge

import (
	"context"
	"fmt"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/hashicorp/go-hclog"
)

const (
	TickTimeMs = 5000
)

type ExpectedTxsFetcherImpl struct {
	ctx               context.Context
	bridgeDataFetcher core.CardanoBridgeDataFetcher
	appConfig         *cCore.AppConfig
	db                core.BridgeExpectedCardanoTxsDB
	logger            hclog.Logger
}

var _ cCore.ExpectedTxsFetcher = (*ExpectedTxsFetcherImpl)(nil)

func NewExpectedTxsFetcher(
	ctx context.Context,
	bridgeDataFetcher core.CardanoBridgeDataFetcher,
	appConfig *cCore.AppConfig,
	db core.BridgeExpectedCardanoTxsDB,
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

	for {
		select {
		case <-f.ctx.Done():
			return
		case <-time.After(TickTimeMs * time.Millisecond):
			err := f.fetchData()
			if err != nil {
				f.logger.Error("error while fetching data", "err", err)
			}
		}
	}
}

func (f *ExpectedTxsFetcherImpl) fetchData() error {
	var expectedTxs []*core.BridgeExpectedCardanoTx

	for chainID := range f.appConfig.CardanoChains {
		existingExpectedTxs, err := f.db.GetAllExpectedTxs(chainID, 0)
		if err != nil {
			f.logger.Error("Failed to GetExpectedTxs from db", "chainId", chainID, "err", err)

			continue
		}

		if len(existingExpectedTxs) > 0 {
			// no new batch can be executed until a claim is produced for the previous batch
			continue
		}

		f.logger.Debug("Fetching expected txs", "chainID", chainID)

		expectedTx, err := f.bridgeDataFetcher.FetchExpectedTx(chainID)
		if err != nil {
			f.logger.Error("Failed to fetch expected tx from bridge", "chainId", chainID, "err", err)

			continue
		}

		f.logger.Debug("Got expected tx", "chainID", chainID, "expectedTx", expectedTx)

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

	f.logger.Debug("Added expected txs to db", "expectedTxs", expectedTxs)

	return nil
}
