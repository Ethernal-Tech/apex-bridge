package processor

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	"github.com/hashicorp/go-hclog"

	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

type CardanoTxsReceiverImpl struct {
	appConfig                   *cCore.AppConfig
	db                          core.CardanoTxsProcessorDB
	txProcessors                *txProcessorsCollection
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater
	logger                      hclog.Logger
}

var _ core.CardanoTxsReceiver = (*CardanoTxsReceiverImpl)(nil)

func NewCardanoTxsReceiverImpl(
	appConfig *cCore.AppConfig,
	db core.CardanoTxsProcessorDB,
	txProcessors *txProcessorsCollection,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
	logger hclog.Logger,
) *CardanoTxsReceiverImpl {
	return &CardanoTxsReceiverImpl{
		appConfig:                   appConfig,
		db:                          db,
		txProcessors:                txProcessors,
		bridgingRequestStateUpdater: bridgingRequestStateUpdater,
		logger:                      logger,
	}
}

func (r *CardanoTxsReceiverImpl) NewUnprocessedTxs(originChainID string, txs []*indexer.Tx) error {
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

		r.logger.Info("Checking if tx is relevant", "chain", originChainID, "tx", tx)

		txProcessor, err := r.txProcessors.getSuccess(cardanoTx, r.appConfig)
		if err != nil {
			r.logger.Error("Failed to get tx processor for new tx", "tx", tx, "err", err)

			onIrrelevantTx(cardanoTx)

			continue
		}

		txProcessorType := txProcessor.GetType()
		if txProcessorType == common.BridgingTxTypeBatchExecution ||
			txProcessorType == common.TxTypeHotWalletFund {
			cardanoTx.Priority = 0
		}

		relevantTxs = append(relevantTxs, cardanoTx)

		if txProcessorType == common.BridgingTxTypeBridgingRequest {
			bridgingRequests = append(
				bridgingRequests,
				&common.NewBridgingRequestStateModel{
					SourceTxHash: common.Hash(tx.Hash),
				},
			)
		}
	}

	if len(bridgingRequests) > 0 {
		r.logger.Debug("Adding multiple new bridging request states to db",
			"chainID", originChainID, "states", bridgingRequests)

		err := r.bridgingRequestStateUpdater.NewMultiple(originChainID, bridgingRequests)
		if err != nil {
			r.logger.Error("error while adding new bridging request states", "err", err)
		}
	}

	// we should update db only if there are some changes needed
	if len(processedTxs)+len(relevantTxs) > 0 {
		r.logger.Debug("Adding txs to db", "processed", processedTxs, "unprocessed", relevantTxs)

		if err := r.db.AddTxs(processedTxs, relevantTxs); err != nil {
			r.logger.Error("Failed to add processed and unprocessed txs", "err", err)

			return err
		}
	}

	if invalidTxsCounter > 0 {
		telemetry.UpdateOracleClaimsInvalidMetaDataCounter(originChainID, invalidTxsCounter) // update telemetry
	}

	return nil
}
