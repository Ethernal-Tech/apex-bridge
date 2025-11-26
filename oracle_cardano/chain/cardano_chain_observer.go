package chain

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer/gouroboros"

	"github.com/hashicorp/go-hclog"
)

const (
	indexerRestartDelay   = time.Second * 5
	indexerKeepAlive      = true
	indexerSyncStartTries = math.MaxInt
)

type CardanoChainObserverImpl struct {
	ctx       context.Context
	indexerDB indexer.Database
	syncer    indexer.BlockSyncer
	config    *cCore.CardanoChainConfig
	isClosed  uint32
	logger    hclog.Logger
}

var _ core.CardanoChainObserver = (*CardanoChainObserverImpl)(nil)

func NewCardanoChainObserver(
	ctx context.Context,
	config *cCore.CardanoChainConfig,
	txsReceiver core.CardanoTxsReceiver, oracleDB core.CardanoTxsProcessorDB,
	indexerDB indexer.Database,
	bridgingAddressesManager common.BridgingAddressesManager,
	logger hclog.Logger,
) (*CardanoChainObserverImpl, error) {
	indexerConfig, syncerConfig := loadSyncerConfigs(config, bridgingAddressesManager)

	err := initOracleState(indexerDB,
		oracleDB, config.StartBlockHash, config.StartSlot, config.InitialUtxos, config.ChainID, logger)
	if err != nil {
		return nil, err
	}

	confirmedBlockHandler := func(block *indexer.CardanoBlock, blockTxs []*indexer.Tx) error {
		logger.Info("Confirmed Block Handler invoked",
			"block", block.Hash, "slot", block.Slot, "block txs", len(blockTxs))

		// do not rely only on blockTx, instead retrieve all unprocessed transactions from the database
		// to account for any previous errors
		txs, err := indexerDB.GetUnprocessedConfirmedTxs(0)
		if err != nil {
			return err
		}

		// Process confirmed Txs
		err = txsReceiver.NewUnprocessedTxs(config.ChainID, txs)
		if err != nil {
			return err
		}

		logger.Info("Txs have been processed", "txs", txs)

		err = indexerDB.MarkConfirmedTxsProcessed(txs)
		if err != nil {
			return err
		}

		return nil
	}

	blockIndexer := indexer.NewBlockIndexer(indexerConfig, confirmedBlockHandler, indexerDB, logger.Named("block_indexer"))
	syncer := gouroboros.NewBlockSyncer(syncerConfig, blockIndexer, logger.Named("block_syncer"))

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

		for {
			select {
			case <-co.ctx.Done():
				return
			case err, ok := <-co.syncer.ErrorCh():
				if !ok {
					return
				}

				co.logger.Error("Syncer fatal error", "chainID", co.config.ChainID, "err", err)

				if err := co.Dispose(); err != nil {
					co.logger.Error("cardano chain observer dispose failed", "err", err)
				}
			}
		}
	}()

	return nil
}

func (co *CardanoChainObserverImpl) Dispose() error {
	if atomic.CompareAndSwapUint32(&co.isClosed, 0, 1) {
		if err := co.syncer.Close(); err != nil {
			co.logger.Error("Failed to close syncer", "err", err)
		}

		if err := co.indexerDB.Close(); err != nil {
			co.logger.Error("Failed to close indexerDB", "err", err)
		}
	}

	return nil
}

func (co *CardanoChainObserverImpl) GetConfig() *cCore.CardanoChainConfig {
	return co.config
}

func (co *CardanoChainObserverImpl) ErrorCh() <-chan error {
	return co.syncer.ErrorCh()
}

func loadSyncerConfigs(
	config *cCore.CardanoChainConfig,
	bridgingAddressesManager common.BridgingAddressesManager,
) (*indexer.BlockIndexerConfig, *gouroboros.BlockSyncerConfig) {
	networkAddress := strings.TrimPrefix(
		strings.TrimPrefix(config.NetworkAddress, "http://"),
		"https://")

	chainID := common.ToNumChainID(config.ChainID)

	addressesOfInterest := append(
		append(bridgingAddressesManager.GetAllPaymentAddresses(chainID),
			bridgingAddressesManager.GetFeeMultisigAddress(chainID),
		),
		config.OtherAddressesOfInterest...,
	)

	indexerConfig := &indexer.BlockIndexerConfig{
		StartingBlockPoint: &indexer.BlockPoint{
			BlockSlot: config.StartSlot,
			BlockHash: indexer.NewHashFromHexString(config.StartBlockHash),
		},
		AddressCheck:           indexer.AddressCheckAll,
		ConfirmationBlockCount: config.ConfirmationBlockCount,
		AddressesOfInterest:    addressesOfInterest,
	}
	syncerConfig := &gouroboros.BlockSyncerConfig{
		NetworkMagic:   config.NetworkMagic,
		NodeAddress:    networkAddress,
		RestartOnError: true, // always try to restart on non-fatal errors
		RestartDelay:   indexerRestartDelay,
		KeepAlive:      indexerKeepAlive,
		SyncStartTries: indexerSyncStartTries,
	}

	return indexerConfig, syncerConfig
}

func initOracleState(
	db indexer.Database, oracleDB core.CardanoTxsProcessorDB,
	blockHashStr string, blockSlot uint64, utxos []cCore.CardanoChainConfigUtxo,
	chainID string, logger hclog.Logger,
) error {
	blockHash := indexer.NewHashFromHexString(blockHashStr)
	if blockHash == (indexer.Hash{}) {
		logger.Info("Configuration block hash is zero hash", "slot", blockSlot)

		return nil
	}

	latestBlockPoint, err := db.GetLatestBlockPoint()
	if err != nil {
		return fmt.Errorf("could not retrieve latest block point while initializing utxos: %w", err)
	}

	currentBlockSlot := uint64(0)
	if latestBlockPoint != nil {
		currentBlockSlot = latestBlockPoint.BlockSlot
	}

	// in oracle we already have more recent block
	if currentBlockSlot >= blockSlot {
		logger.Info("Oracle database contains more recent block",
			"hash", currentBlockSlot, "slot", currentBlockSlot)

		return nil
	}

	return db.OpenTx().SetLatestBlockPoint(&indexer.BlockPoint{
		BlockSlot: blockSlot,
		BlockHash: blockHash,
	}).Execute()

	// if err := oracleDB.ClearAllTxs(chainID); err != nil {
	// 	return err
	// }

	// return db.OpenTx().DeleteAllTxOutputsPhysically().SetLatestBlockPoint(&indexer.BlockPoint{
	// 	BlockSlot: blockSlot,
	// 	BlockHash: blockHash,
	// }).AddTxOutputs(convertUtxos(utxos)).Execute()
}

func convertUtxos(input []cCore.CardanoChainConfigUtxo) (output []*indexer.TxInputOutput) {
	output = make([]*indexer.TxInputOutput, len(input))
	for i, inp := range input {
		utxo := &indexer.TxInputOutput{
			Input: indexer.TxInput{
				Hash:  inp.Hash,
				Index: inp.Index,
			},
			Output: indexer.TxOutput{
				Address: inp.Address,
				Amount:  inp.Amount,
				Slot:    inp.Slot,
				Tokens:  make([]indexer.TokenAmount, 0, len(inp.Tokens)),
			},
		}
		for _, t := range inp.Tokens {
			utxo.Output.Tokens = append(utxo.Output.Tokens, indexer.TokenAmount{
				PolicyID: t.Token.PolicyID,
				Name:     t.Token.Name,
				Amount:   t.Amount,
			})
		}

		output[i] = utxo
	}

	return output
}
