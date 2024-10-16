package cardanotxsprocessor

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	"github.com/hashicorp/go-hclog"

	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

type CardanoTxsReceiverImpl struct {
	db                          core.CardanoTxsProcessorDB
	txProcessors                *txProcessorsCollection
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater
	logger                      hclog.Logger
}

var _ core.CardanoTxsReceiver = (*CardanoTxsReceiverImpl)(nil)

func NewCardanoTxsReceiverImpl(
	db core.CardanoTxsProcessorDB,
	txProcessors *txProcessorsCollection,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
	logger hclog.Logger,
) *CardanoTxsReceiverImpl {
	return &CardanoTxsReceiverImpl{
		db:                          db,
		txProcessors:                txProcessors,
		bridgingRequestStateUpdater: bridgingRequestStateUpdater,
		logger:                      logger,
	}
}

func (r *CardanoTxsReceiverImpl) NewUnprocessedTxs(originChainID string, txs []*indexer.Tx) error {
	r.logger.Info("NewUnprocessedTxs", "txs", txs)

	var (
		bridgingRequests  []*common.NewBridgingRequestStateModel
		relevantTxs       = make([]*core.CardanoTx, 0)
		processedTxs      []*core.ProcessedCardanoTx
		invalidTxsCounter int
	)

	onIrrelevantTx := func(cardanoTx *core.CardanoTx) {
		processedTxs = append(processedTxs, cardanoTx.ToProcessedCardanoTx(false))
		invalidTxsCounter++
	}

	for _, tx := range txs {
		cardanoTx := &core.CardanoTx{
			OriginChainID: originChainID,
			Tx:            *tx,
			Priority:      1,
		}

		r.logger.Debug("Checking if tx is relevant", "tx", tx)

		txProcessor, err := r.txProcessors.getSuccess(tx.Metadata)
		if err != nil {
			r.logger.Error("Failed to get tx processor for new tx", "tx", tx, "err", err)

			onIrrelevantTx(cardanoTx)

			continue
		}

		if txProcessor.GetType() == common.BridgingTxTypeBatchExecution {
			cardanoTx.Priority = 0
		}

		relevantTxs = append(relevantTxs, cardanoTx)

		if txProcessor.GetType() == common.BridgingTxTypeBridgingRequest {
			bridgingRequests = append(
				bridgingRequests,
				&common.NewBridgingRequestStateModel{
					SourceTxHash: common.Hash(tx.Hash),
				},
			)
		}
	}

	if len(processedTxs) > 0 {
		r.logger.Debug("Adding already processed txs to db", "txs", processedTxs)

		err := r.db.AddProcessedTxs(processedTxs)
		if err != nil {
			r.logger.Error("Failed to add already processed txs", "err", err)

			return err
		}
	}

	if len(relevantTxs) > 0 {
		r.logger.Debug("Adding relevant txs to db", "txs", relevantTxs)

		err := r.db.AddUnprocessedTxs(relevantTxs)
		if err != nil {
			r.logger.Error("Failed to add unprocessed txs", "err", err)

			return err
		}
	}

	if invalidTxsCounter > 0 {
		telemetry.UpdateOracleClaimsInvalidMetaDataCounter(originChainID, invalidTxsCounter) // update telemetry
	}

	if len(bridgingRequests) > 0 {
		r.logger.Debug("Adding multiple new bridging request states to db",
			"chainID", originChainID, "states", bridgingRequests)

		err := r.bridgingRequestStateUpdater.NewMultiple(originChainID, bridgingRequests)
		if err != nil {
			r.logger.Error("error while adding new bridging request states", "err", err)
		}
	}

	return nil
}
