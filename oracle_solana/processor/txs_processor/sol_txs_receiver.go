package processor

import (
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_common/utils"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	skyline "github.com/Ethernal-Tech/solana-infrastructure/sendtx/skyline_program"
	"github.com/Ethernal-Tech/solana-infrastructure/tracker"
	"github.com/hashicorp/go-hclog"
)

type SolEventReceiverImpl struct {
	appConfig                   *oCore.AppConfig
	db                          core.SolanaTxsProcessorDB
	txProcessors                *txProcessorsCollection
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater
	logger                      hclog.Logger
}

var _ core.SolanaTxsReceiver = (*SolEventReceiverImpl)(nil)

func NewSolTxsReceiver(
	appConfig *oCore.AppConfig,
	db core.SolanaTxsProcessorDB,
	txProcessors *txProcessorsCollection,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
	logger hclog.Logger,
) *SolEventReceiverImpl {
	return &SolEventReceiverImpl{
		appConfig:                   appConfig,
		db:                          db,
		txProcessors:                txProcessors,
		bridgingRequestStateUpdater: bridgingRequestStateUpdater,
		logger:                      logger,
	}
}

func (r *SolEventReceiverImpl) NewUnprocessedEvent(originChainID string, event tracker.EventNotification) error {
	r.logger.Debug("NewUnprocessedEvent", "originChainID", originChainID, "event", event)

	if _, exists := r.appConfig.SolanaChains[originChainID]; !exists {
		r.logger.Error("originChainID not registered", "originChainID", originChainID)

		return fmt.Errorf("originChainID not registered. originChainID: %s", originChainID)
	}

	if event.EventName == "" || event.EventData == nil {
		r.logger.Error("empty event received")

		return nil
	}

	tx, err := r.parseEvent(originChainID, event)
	if err != nil {
		r.logger.Error("failed to parse event", "err", err)

		return err
	}

	r.logger.Debug("Parsed event", "tx", tx) // remove

	var (
		bridgingRequests []*common.NewBridgingRequestStateModel
		relevantTxs      = make([]*core.SolanaTx, 0)
		processedTxs     []*core.ProcessedSolanaTx
	)

	txProcessor, err := r.txProcessors.getSuccess(tx, r.appConfig)
	if err != nil {
		r.logger.Error("failed to get tx processor", "err", err)

		processedTxs = append(processedTxs, tx.ToProcessedSolanaTx(false))
	} else {
		txProcessorType := txProcessor.GetType()
		tx.Priority = utils.GetTxPriority(txProcessorType)

		relevantTxs = append(relevantTxs, tx)

		if txProcessorType == common.BridgingTxTypeBridgingRequest ||
			txProcessorType == common.TxTypeRefundRequest {
			bridgingRequests = append(
				bridgingRequests,
				&common.NewBridgingRequestStateModel{
					SourceTxHash: tx.TxSignature[:],
					IsRefund:     txProcessorType == common.TxTypeRefundRequest,
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

func (r *SolEventReceiverImpl) parseEvent(
	originChainID string, event tracker.EventNotification,
) (*core.SolanaTx, error) {
	switch event.EventName {
	case core.BridgeRequestEvent:
		return r.parseBridgeRequestEvent(originChainID, event)
	case core.TransactionExecutedEvent:
		return r.parseTransactionExecutedEvent(originChainID, event)
	case core.HotWalletIncrementEvent:
		return r.parseHotWalletIncrementEvent(originChainID, event)
	case core.ValidatorSetUpdatedEvent:
		panic("not implemented") //nolint:gocritic
	default:
		return nil, fmt.Errorf("unknown event name: %s", event.EventName)
	}
}

func (r *SolEventReceiverImpl) parseBridgeRequestEvent(
	originChainID string, event tracker.EventNotification,
) (*core.SolanaTx, error) {
	bridgeRequestEvent, ok := event.EventData.(*skyline.BridgeRequestEvent)
	if !ok {
		return nil, fmt.Errorf("failed to parse bridge request event")
	}

	tokenID, err := r.appConfig.SolanaChains[originChainID].GetTokenIDByName(bridgeRequestEvent.MintToken.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get token id by name: %w", err)
	}

	transactions := make([]core.BridgingRequestSolMetadataTransaction, 1)
	transactions[0] = core.BridgingRequestSolMetadataTransaction{
		Address: bridgeRequestEvent.Receiver,
		Amount:  common.LamportToWei(new(big.Int).SetUint64(bridgeRequestEvent.Amount)),
		TokenID: tokenID,
	}

	mtdata := core.BridgingRequestSolMetadata{
		BridgingTxType:     common.BridgingTxTypeBridgingRequest,
		DestinationChainID: bridgeRequestEvent.DestinationChain,
		SenderAddr:         bridgeRequestEvent.Sender.String(),
		Transactions:       transactions,
		BridgingFee:        common.LamportToWei(new(big.Int).SetUint64(bridgeRequestEvent.BridgeFee)),
		OperationFee:       common.LamportToWei(new(big.Int).SetUint64(bridgeRequestEvent.OperationalFee)),
	}

	metadata, err := core.MarshalSolMetadata(mtdata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return &core.SolanaTx{
		OriginChainID: originChainID,
		Priority:      1,

		SlotNumber:  event.SlotNumber,
		TxSignature: event.TxSignature,
		Value:       common.LamportToWei(new(big.Int).SetUint64(bridgeRequestEvent.Amount)),
		Metadata:    metadata,
	}, nil
}

func (r *SolEventReceiverImpl) parseHotWalletIncrementEvent(
	originChainID string, event tracker.EventNotification,
) (*core.SolanaTx, error) {
	hotWalletIncrementEvent, ok := event.EventData.(*skyline.HotWalletIncrementEvent)
	if !ok {
		return nil, fmt.Errorf("failed to parse hot wallet increment event")
	}

	if hotWalletIncrementEvent.Mint.String() != skyline.NATIVE_SOL_MINT.String() {
		if _, err := r.appConfig.SolanaChains[originChainID].GetTokenIDByName(
			hotWalletIncrementEvent.Mint.String()); err != nil {
			return nil, fmt.Errorf("failed to get token id by mint: %w", err)
		}
	}

	return &core.SolanaTx{
		OriginChainID: originChainID,
		Priority:      1,

		SlotNumber:  event.SlotNumber,
		TxSignature: event.TxSignature,
		Value:       common.LamportToWei(new(big.Int).SetUint64(hotWalletIncrementEvent.Amount)),
		Metadata:    []byte{},
	}, nil
}

func (r *SolEventReceiverImpl) parseTransactionExecutedEvent(
	originChainID string,
	event tracker.EventNotification,
) (*core.SolanaTx, error) {
	transactionExecutedEvent, ok := event.EventData.(*skyline.TransactionExecutedEvent)
	if !ok {
		return nil, fmt.Errorf("failed to parse transaction executed event")
	}

	metadata, err := core.MarshalSolMetadata(core.BatchExecutedSolMetadata{
		BridgingTxType: common.BridgingTxTypeBatchExecution,
		BatchNonceID:   transactionExecutedEvent.BatchId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return &core.SolanaTx{
		OriginChainID: originChainID,
		Priority:      1,

		SlotNumber:      event.SlotNumber,
		TxSignature:     event.TxSignature,
		InnerActionHash: event.InnerActionHash,
		Metadata:        metadata,
	}, nil
}
