package chain

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"

	"github.com/hashicorp/go-hclog"
)

type CardanoChainObserverImpl struct {
	ctx       context.Context
	indexerDb indexer.Database
	syncer    indexer.BlockSyncer
	logger    hclog.Logger
	config    *core.CardanoChainConfig
}

var _ core.CardanoChainObserver = (*CardanoChainObserverImpl)(nil)

func NewCardanoChainObserver(
	ctx context.Context,
	config *core.CardanoChainConfig,
	txsProcessor core.CardanoTxsProcessor, oracleDb core.CardanoTxsProcessorDb,
	indexerDb indexer.Database, bridgeDataFetcher core.BridgeDataFetcher,
	logger hclog.Logger,
) (*CardanoChainObserverImpl, error) {
	indexerConfig, syncerConfig := loadSyncerConfigs(config)

	if len(config.InitialUtxos) > 0 {
		err := initUtxos(indexerDb, config.InitialUtxos)
		if err != nil {
			return nil, err
		}
	}

	err := updateLastConfirmedBlockFromSc(ctx, indexerDb, oracleDb, bridgeDataFetcher, config.ChainId, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to updateLastConfirmedBlockFromSc. err: %w", err)
	}

	confirmedBlockHandler := func(cb *indexer.CardanoBlock, txs []*indexer.Tx) error {
		logger.Info("Confirmed Txs", "txs", len(txs))

		txs, err := indexerDb.GetUnprocessedConfirmedTxs(0)
		if err != nil {
			return err
		}

		// Process confirmed Txs
		err = txsProcessor.NewUnprocessedTxs(config.ChainId, txs)
		if err != nil {
			return err
		}

		logger.Info("Txs have been processed", "txs", txs)

		return indexerDb.MarkConfirmedTxsProcessed(txs)
	}

	blockIndexer := indexer.NewBlockIndexer(indexerConfig, confirmedBlockHandler, indexerDb, logger.Named("block_indexer"))

	syncer := indexer.NewBlockSyncer(syncerConfig, blockIndexer, logger.Named("block_syncer"))

	return &CardanoChainObserverImpl{
		ctx:       ctx,
		indexerDb: indexerDb,
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
					"chainId", co.config.ChainId, "err", err)
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

	addressesOfInterest := []string{
		config.BridgingAddresses.BridgingAddress,
		config.BridgingAddresses.FeeAddress,
	}

	addressesOfInterest = append(addressesOfInterest, config.OtherAddressesOfInterest...)

	indexerConfig := &indexer.BlockIndexerConfig{
		StartingBlockPoint: &indexer.BlockPoint{
			BlockSlot:   config.StartSlot,
			BlockHash:   startBlockHash,
			BlockNumber: config.StartBlockNumber - 1,
		},
		AddressCheck:           indexer.AddressCheckOutputs,
		ConfirmationBlockCount: config.ConfirmationBlockCount,
		AddressesOfInterest:    addressesOfInterest,
		SoftDeleteUtxo:         true,
	}
	syncerConfig := &indexer.BlockSyncerConfig{
		NetworkMagic:   config.NetworkMagic,
		NodeAddress:    config.NetworkAddress,
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
	indexerDb indexer.Database,
	oracleDb core.CardanoTxsProcessorDb,
	bridgeDataFetcher core.BridgeDataFetcher,
	chainId string,
	logger hclog.Logger,
) error {
	var blockPointSc *indexer.BlockPoint
	err := common.RetryForever(ctx, 2*time.Second, func(context.Context) (err error) {
		blockPointSc, err = bridgeDataFetcher.FetchLatestBlockPoint(chainId)
		if err != nil {
			logger.Error(
				"Failed to FetchLatestBlockPoint while creating CardanoChainObserver. Retrying...",
				"chainId", chainId, "err", err)
		}

		return err
	})

	if err != nil {
		return fmt.Errorf("error while RetryForever of FetchLatestBlockPoint. err: %w", err)
	}

	if blockPointSc == nil {
		return nil
	}

	blockPointDb, err := indexerDb.GetLatestBlockPoint()
	if err != nil {
		return err
	}

	if blockPointDb != nil {
		if blockPointDb.BlockSlot > blockPointSc.BlockSlot {
			return nil
		}

		if bytes.Equal(blockPointDb.BlockHash, blockPointSc.BlockHash) &&
			blockPointDb.BlockSlot == blockPointSc.BlockSlot {
			return nil
		}
	}

	if err := oracleDb.ClearUnprocessedTxs(chainId); err != nil {
		return err
	}

	if err := oracleDb.ClearExpectedTxs(chainId); err != nil {
		return err
	}

	return indexerDb.OpenTx().SetLatestBlockPoint(blockPointSc).Execute()
}
