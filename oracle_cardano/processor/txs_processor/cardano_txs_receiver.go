package processor

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_common/utils"
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
		bridgingRequests []*common.NewBridgingRequestStateModel
		relevantTxs      = make([]*core.CardanoTx, 0)
		processedTxs     []*core.ProcessedCardanoTx
	)

	for _, tx := range txs {
		cardanoTx := &core.CardanoTx{
			OriginChainID: originChainID,
			Tx:            *tx,
		}

		r.logger.Info("Checking if tx is relevant", "chain", originChainID, "tx", tx)

		txProcessor, err := r.txProcessors.getSuccess(cardanoTx, r.appConfig)
		if err != nil {
			r.logger.Error("Failed to get tx processor for new tx", "tx", tx, "err", err)

			processedTxs = append(processedTxs, cardanoTx.ToProcessedCardanoTx(false))

			continue
		}

		txProcessorType := txProcessor.GetType()
		cardanoTx.Priority = utils.GetTxPriority(txProcessorType)

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

		updateTelemetry(originChainID, processedTxs, relevantTxs)
	}

	return nil
}

func updateTelemetry(originChainID string, processedTxs []*core.ProcessedCardanoTx, relevantTxs []*core.CardanoTx) {
	telemetry.UpdateOracleTxsReceivedCounter(originChainID, len(processedTxs)+len(relevantTxs))

	invalidCnt := 0

	for _, x := range processedTxs {
		if x.IsInvalid {
			invalidCnt++
		}
	}

	if invalidCnt > 0 {
		telemetry.UpdateOracleClaimsInvalidMetaDataCounter(originChainID, invalidCnt)
	}
}
