package processor

import (
	"math/big"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	successtxprocessors "github.com/Ethernal-Tech/apex-bridge/oracle_cardano/processor/tx_processors/success"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_common/utils"
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

		if txProcessorType == common.BridgingTxTypeBridgingRequest ||
			txProcessorType == common.TxTypeRefundRequest {
			isRefund := txProcessorType == common.TxTypeRefundRequest

			dstChainID, details := "", (*common.BridgingRequestStateDetails)(nil)
			if !isRefund {
				dstChainID, details = r.getBridgingRequestStateDetails(originChainID, tx.Metadata)
			}

			bridgingRequests = append(
				bridgingRequests,
				&common.NewBridgingRequestStateModel{
					SourceTxHash:       tx.Hash[:],
					IsRefund:           isRefund,
					DestinationChainID: dstChainID,
					Details:            details,
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

		utils.UpdateTxReceivedTelemetry(originChainID, processedTxs, len(relevantTxs))
	}

	return nil
}

// getBridgingRequestStateDetails reads what was requested to be bridged out of the tx metadata.
// It is reporting data only, so any failure yields empty details instead of an error.
func (r *CardanoTxsReceiverImpl) getBridgingRequestStateDetails(
	originChainID string, txMetadata []byte,
) (string, *common.BridgingRequestStateDetails) {
	chainConfig := r.appConfig.CardanoChains[originChainID]
	if chainConfig == nil {
		r.logger.Warn("no chain config for bridging request state details", "chainID", originChainID)

		return "", nil
	}

	metadata, err := successtxprocessors.UnmarshalBridgingRequestMetadata(chainConfig, txMetadata)
	if err != nil {
		r.logger.Warn("failed to unmarshal metadata for bridging request state details",
			"chainID", originChainID, "err", err)

		return "", nil
	}

	currencyID, err := chainConfig.GetCurrencyID()
	if err != nil {
		r.logger.Warn("failed to get currency id for bridging request state details",
			"chainID", originChainID, "err", err)

		return metadata.DestinationChainID, nil
	}

	receivers := make([]common.BridgingRequestStateReceiver, len(metadata.Transactions))
	for i, receiver := range metadata.Transactions {
		receivers[i] = common.BridgingRequestStateReceiver{
			Address: strings.Join(receiver.Address, ""),
			Amount:  common.DfmToWei(new(big.Int).SetUint64(receiver.Amount)),
			TokenID: receiver.TokenID,
		}
	}

	return metadata.DestinationChainID, common.NewBridgingRequestStateDetails(
		strings.Join(metadata.SenderAddr, ""), receivers,
		common.DfmToWei(new(big.Int).SetUint64(metadata.BridgingFee)),
		common.DfmToWei(new(big.Int).SetUint64(metadata.OperationFee)),
		currencyID)
}
