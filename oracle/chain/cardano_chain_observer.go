package chain

import (
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"

	"github.com/hashicorp/go-hclog"
)

type CardanoChainObserverImpl struct {
	ctx       context.Context
	indexerDB indexer.Database
	syncer    indexer.BlockSyncer
	logger    hclog.Logger
	config    *core.CardanoChainConfig
}

var _ core.CardanoChainObserver = (*CardanoChainObserverImpl)(nil)

func NewCardanoChainObserver(
	ctx context.Context,
	config *core.CardanoChainConfig,
	txsProcessor core.CardanoTxsProcessor, oracleDB core.CardanoTxsProcessorDB,
	indexerDB indexer.Database, bridgeDataFetcher core.BridgeDataFetcher,
	logger hclog.Logger,
) (*CardanoChainObserverImpl, error) {
	indexerConfig, syncerConfig := loadSyncerConfigs(config)

	if len(config.InitialUtxos) > 0 {
		logger.Debug("trying to insert utxos", "utxos", config.InitialUtxos)

		err := initUtxos(indexerDB, config.InitialUtxos)
		if err != nil {
			return nil, err
		}

		logger.Info("inserted utxos", "utxos", config.InitialUtxos)
	}

	err := updateLastConfirmedBlockFromSc(ctx, indexerDB, oracleDB, bridgeDataFetcher, config.ChainID, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to updateLastConfirmedBlockFromSc. err: %w", err)
	}

	confirmedBlockHandler := func(cb *indexer.CardanoBlock, txs []*indexer.Tx) error {
		logger.Info("Confirmed Txs", "txs", len(txs))

		telemetry.UpdateOracleTxsReceivedCounter(config.ChainID, len(txs))

		txs, err := indexerDB.GetUnprocessedConfirmedTxs(0)
		if err != nil {
			return err
		}

		// Process confirmed Txs
		err = txsProcessor.NewUnprocessedTxs(config.ChainID, txs)
		if err != nil {
			return err
		}

		logger.Info("Txs have been processed", "txs", txs)

		return indexerDB.MarkConfirmedTxsProcessed(txs)
	}

	blockIndexer := indexer.NewBlockIndexer(indexerConfig, confirmedBlockHandler, indexerDB, logger.Named("block_indexer"))

	syncer := indexer.NewBlockSyncer(syncerConfig, blockIndexer, logger.Named("block_syncer"))

	return &CardanoChainObserverImpl{
		ctx:       ctx,
		indexerDB: indexerDB,
		syncer:    syncer,
		logger:    logger,
		config:    config,
	}, nil
}

func (co *CardanoChainObserverImpl) Start() error {
	go func() {
		_ = common.RetryForever(co.ctx, 5*time.Second, func(context.Context) (err error) {
			err = co.syncer.Sync()
			if err != nil {
				co.logger.Error(
					"Failed to Start syncer while starting CardanoChainObserver. Retrying...",
					"chainId", co.config.ChainID, "err", err)
			}

			return err
		})
	}()

	return nil
}

func (co *CardanoChainObserverImpl) Dispose() error {
	err := co.syncer.Close()
	if err != nil {
		co.logger.Error("Syncer close failed", "err", err)

		return fmt.Errorf("syncer close failed. err: %w", err)
	}

	return nil
}

func (co *CardanoChainObserverImpl) GetConfig() *core.CardanoChainConfig {
	return co.config
}

func (co *CardanoChainObserverImpl) ErrorCh() <-chan error {
	return co.syncer.ErrorCh()
}

func loadSyncerConfigs(config *core.CardanoChainConfig) (*indexer.BlockIndexerConfig, *indexer.BlockSyncerConfig) {
	var startBlockHash []byte
	if config.StartBlockHash == "" {
		startBlockHash = []byte(nil)
	} else {
		startBlockHash, _ = hex.DecodeString(config.StartBlockHash)
	}

	networkAddress := config.NetworkAddress
	networkAddress = strings.TrimPrefix(networkAddress, "http://")
	networkAddress = strings.TrimPrefix(networkAddress, "https://")

	addressesOfInterest := []string{
		config.BridgingAddresses.BridgingAddress,
		config.BridgingAddresses.FeeAddress,
	}

	addressesOfInterest = append(addressesOfInterest, config.OtherAddressesOfInterest...)

	indexerConfig := &indexer.BlockIndexerConfig{
		StartingBlockPoint: &indexer.BlockPoint{
			BlockSlot:   config.StartSlot,
			BlockHash:   indexer.NewHashFromBytes(startBlockHash),
			BlockNumber: config.StartBlockNumber - 1,
		},
		AddressCheck:           indexer.AddressCheckAll,
		ConfirmationBlockCount: config.ConfirmationBlockCount,
		AddressesOfInterest:    addressesOfInterest,
		SoftDeleteUtxo:         false,
	}
	syncerConfig := &indexer.BlockSyncerConfig{
		NetworkMagic:   config.NetworkMagic,
		NodeAddress:    networkAddress,
		RestartOnError: true,
		RestartDelay:   time.Second * 5,
		KeepAlive:      true,
		SyncStartTries: math.MaxInt,
	}

	return indexerConfig, syncerConfig
}

func initUtxos(db indexer.Database, utxos []*indexer.TxInputOutput) error {
	var nonExistingUtxos []*indexer.TxInputOutput

	for _, x := range utxos {
		r, err := db.GetTxOutput(x.Input)
		if err != nil {
			return err
		} else if r.Address == "" {
			nonExistingUtxos = append(nonExistingUtxos, x)
		}
	}

	return db.OpenTx().AddTxOutputs(nonExistingUtxos).Execute()
}

func updateLastConfirmedBlockFromSc(
	ctx context.Context,
	indexerDB indexer.Database,
	oracleDB core.CardanoTxsProcessorDB,
	bridgeDataFetcher core.BridgeDataFetcher,
	chainID string,
	logger hclog.Logger,
) error {
	var blockPointSc *indexer.BlockPoint

	l := logger.Named("updateLastConfirmedBlockFromSc")

	l.Debug("trying to update last confirmed block")

	err := common.RetryForever(ctx, 2*time.Second, func(context.Context) (err error) {
		blockPointSc, err = bridgeDataFetcher.FetchLatestBlockPoint(chainID)
		if err != nil {
			l.Error(
				"Failed to FetchLatestBlockPoint while creating CardanoChainObserver. Retrying...",
				"chainId", chainID, "err", err)
		}

		return err
	})

	if err != nil {
		return fmt.Errorf("error while RetryForever of FetchLatestBlockPoint. err: %w", err)
	}

	l.Debug("done FetchLatestBlockPoint", "blockPointSc", blockPointSc)

	if blockPointSc == nil {
		return nil
	}

	blockPointDB, err := indexerDB.GetLatestBlockPoint()
	if err != nil {
		return err
	}

	l.Debug("done GetLatestBlockPoint", "blockPointDB", blockPointDB)

	if blockPointDB != nil {
		if blockPointDB.BlockSlot > blockPointSc.BlockSlot {
			l.Debug("slot from db higher than from sc",
				"blockPointDB.BlockSlot", blockPointDB.BlockSlot, "blockPointSc.BlockSlot", blockPointSc.BlockSlot)

			return nil
		}

		if blockPointDB.BlockHash == blockPointSc.BlockHash &&
			blockPointDB.BlockSlot == blockPointSc.BlockSlot {
			l.Debug("slot from db same as from sc",
				"blockPointDB.BlockSlot", blockPointDB.BlockSlot, "blockPointSc.BlockSlot", blockPointSc.BlockSlot)

			return nil
		}
	}

	if err := oracleDB.ClearUnprocessedTxs(chainID); err != nil {
		return err
	}

	if err := oracleDB.ClearExpectedTxs(chainID); err != nil {
		return err
	}

	l.Info("updating last confirmed block", "to blockPoint", blockPointSc)

	return indexerDB.OpenTx().SetLatestBlockPoint(blockPointSc).Execute()
}
