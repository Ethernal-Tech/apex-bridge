package chain

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"

	"github.com/hashicorp/go-hclog"
)

type CardanoChainObserverImpl struct {
	indexerDb indexer.Database
	syncer    indexer.BlockSyncer
	logger    hclog.Logger
	config    *core.CardanoChainConfig
}

var _ core.CardanoChainObserver = (*CardanoChainObserverImpl)(nil)

func NewCardanoChainObserver(
	settings core.AppSettings, config *core.CardanoChainConfig, initialUtxosForChain []*indexer.TxInputOutput,
	txsProcessor core.CardanoTxsProcessor, oracleDb core.CardanoTxsProcessorDb, indexerDb indexer.Database, bridgeDataFetcher core.BridgeDataFetcher,
) *CardanoChainObserverImpl {
	logger, err := logger.NewLogger(logger.LoggerConfig{
		LogLevel:      hclog.Level(settings.LogLevel),
		JSONLogFormat: false,
		AppendFile:    true,
		LogFilePath:   settings.LogsPath + config.ChainId,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return nil
	}

	indexerConfig, syncerConfig := loadSyncerConfigs(config)

	if len(initialUtxosForChain) > 0 {
		err := initUtxos(indexerDb, initialUtxosForChain)
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		logger.Error("Failed to insert initial UTXOs", "err", err)
	}

	err = updateLastConfirmedBlockFromSc(indexerDb, oracleDb, bridgeDataFetcher, config.ChainId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		logger.Error("Update latest confirmed block from Smart Contract failed", "err", err)
		return nil
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
		indexerDb: indexerDb,
		syncer:    syncer,
		logger:    logger,
		config:    config,
	}
}

func (co *CardanoChainObserverImpl) Start() error {
	err := co.syncer.Sync()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		co.logger.Error("Start syncing failed", "err", err)
	}

	return err
}

func (co *CardanoChainObserverImpl) Stop() error {

	err := co.syncer.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		co.logger.Error("Syncer close failed", "err", err)
	}

	return err
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
		RestartDelay:   time.Second * 2,
		KeepAlive:      true,
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

func updateLastConfirmedBlockFromSc(indexerDb indexer.Database, oracleDb core.CardanoTxsProcessorDb, bridgeDataFetcher core.BridgeDataFetcher, chainId string) error {
	blockPointSc, err := bridgeDataFetcher.FetchLatestBlockPoint(chainId)
	if blockPointSc == nil || err != nil {
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
