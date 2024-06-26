package chain

import (
	"context"
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

	if err := initUtxos(indexerDB, config.InitialUtxos); err != nil {
		return nil, err
	}

	/*
		if err := oracleDB.ClearUnprocessedTxs(config.ChainID); err != nil {
			return nil, err
		}

		if err := oracleDB.ClearExpectedTxs(config.ChainID); err != nil {
			return nil, err
		}
	*/

	confirmedBlockHandler := func(block *indexer.CardanoBlock, blockTxs []*indexer.Tx) error {
		logger.Info("Confirmed Block Handler invoked",
			"block", block.Hash, "slot", block.Slot, "block txs", len(blockTxs))

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

		err = indexerDB.MarkConfirmedTxsProcessed(txs)
		if err != nil {
			return err
		}

		telemetry.UpdateOracleTxsReceivedCounter(config.ChainID, len(txs))

		return nil
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
	bp, err := co.indexerDB.GetLatestBlockPoint()
	if err == nil && bp != nil {
		co.logger.Debug("Started...", "hash", bp.BlockHash, "slot", bp.BlockSlot)
	}

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
	networkAddress := strings.TrimPrefix(
		strings.TrimPrefix(config.NetworkAddress, "http://"),
		"https://")

	addressesOfInterest := append([]string{
		config.BridgingAddresses.BridgingAddress,
		config.BridgingAddresses.FeeAddress,
	}, config.OtherAddressesOfInterest...)

	indexerConfig := &indexer.BlockIndexerConfig{
		StartingBlockPoint: &indexer.BlockPoint{
			BlockSlot:   config.StartSlot,
			BlockHash:   indexer.NewHashFromHexString(config.StartBlockHash),
			BlockNumber: config.StartBlockNumber,
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
	if len(utxos) == 0 {
		return nil
	}

	return db.OpenTx().DeleteAllTxOutputsPhysically().AddTxOutputs(utxos).Execute()
}
