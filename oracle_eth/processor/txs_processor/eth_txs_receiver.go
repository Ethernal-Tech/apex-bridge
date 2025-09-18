package processor

import (
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_common/utils"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	"github.com/Ethernal-Tech/ethgo"
	ethereum_common "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/hashicorp/go-hclog"
)

type EthTxsReceiverImpl struct {
	appConfig                   *oCore.AppConfig
	db                          core.EthTxsProcessorDB
	txProcessors                *txProcessorsCollection
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater
	logger                      hclog.Logger
}

var _ core.EthTxsReceiver = (*EthTxsReceiverImpl)(nil)

func NewEthTxsReceiverImpl(
	appConfig *oCore.AppConfig,
	db core.EthTxsProcessorDB,
	txProcessors *txProcessorsCollection,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
	logger hclog.Logger,
) *EthTxsReceiverImpl {
	return &EthTxsReceiverImpl{
		appConfig:                   appConfig,
		db:                          db,
		txProcessors:                txProcessors,
		bridgingRequestStateUpdater: bridgingRequestStateUpdater,
		logger:                      logger,
	}
}

func (r *EthTxsReceiverImpl) NewUnprocessedLog(originChainID string, log *ethgo.Log) error {
	r.logger.Info("NewUnprocessedLog", "log", log)

	_, exists := r.appConfig.EthChains[originChainID]
	if !exists {
		r.logger.Error("originChainID not registered", "originChainID", originChainID)

		return fmt.Errorf("originChainID not registered. originChainID: %s", originChainID)
	}

	var (
		bridgingRequests []*common.NewBridgingRequestStateModel
		relevantTxs      []*core.EthTx
		processedTxs     []*core.ProcessedEthTx
	)

	if log == nil || log.Data == nil || log.Topics == nil {
		r.logger.Error("empty log received")

		return nil
	}

	tx, err := r.logToTx(originChainID, log)
	if err != nil {
		r.logger.Error("failed to convert log into tx", "err", err)

		return err
	}

	r.logger.Debug("Checking if tx is relevant", "tx", tx)

	txProcessor, err := r.txProcessors.getSuccess(tx, r.appConfig)
	if err != nil {
		r.logger.Error("Failed to get tx processor for new tx", "tx", tx, "err", err)

		processedTxs = append(processedTxs, tx.ToProcessedEthTx(false))
	} else {
		txProcessorType := txProcessor.GetType()
		tx.Priority = utils.GetTxPriority(txProcessorType)

		relevantTxs = append(relevantTxs, tx)

		if txProcessorType == common.BridgingTxTypeBridgingRequest ||
			txProcessorType == common.TxTypeRefundRequest {
			bridgingRequests = append(
				bridgingRequests,
				&common.NewBridgingRequestStateModel{
					SourceTxHash: common.Hash(tx.Hash),
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

func (r *EthTxsReceiverImpl) logToTx(originChainID string, log *ethgo.Log) (*core.EthTx, error) {
	events, err := eth.GetNexusEventSignatures()
	if err != nil {
		r.logger.Error("failed to get nexus event signatures", "err", err)

		return nil, err
	}

	depositEventSig := events[0]
	withdrawEventSig := events[1]
	fundedEventSig := events[2]
	// validator set change occurred
	vscEventSig := events[3]

	contract, err := contractbinding.NewGateway(ethereum_common.Address{}, nil)
	if err != nil {
		r.logger.Error("failed to get contractbinding gateway", "err", err)

		return nil, err
	}

	topics := make([]ethereum_common.Hash, len(log.Topics))
	for idx, topic := range log.Topics {
		topics[idx] = ethereum_common.Hash(topic)
	}

	parsedLog := types.Log{
		Address:     ethereum_common.Address(log.Address),
		Data:        log.Data,
		BlockNumber: log.BlockNumber,
		TxHash:      ethereum_common.Hash(log.TransactionHash),
		TxIndex:     uint(log.TransactionIndex),
		BlockHash:   ethereum_common.Hash(log.BlockHash),
		Index:       uint(log.LogIndex),
		Removed:     log.Removed,
		Topics:      topics,
	}

	var (
		metadata          []byte
		innerActionTxHash ethgo.Hash
		txValue           *big.Int
	)

	logEventType := log.Topics[0]
	switch logEventType {
	case depositEventSig:
		deposit, err := contract.GatewayFilterer.ParseDeposit(parsedLog)
		if err != nil {
			r.logger.Error("failed to parse deposit event", "err", err)

			return nil, err
		}

		evmTx, err := eth.NewEVMSmartContractTransaction(deposit.Data)
		if err != nil {
			r.logger.Error("failed to create new evm smart contract tx", "err", err)

			return nil, err
		}

		batchExecutedMetadata := core.BatchExecutedEthMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
			BatchNonceID:   evmTx.BatchNonceID,
		}

		metadata, err = core.MarshalEthMetadata(batchExecutedMetadata)
		if err != nil {
			r.logger.Error("failed to marshal metadata", "err", err)

			return nil, err
		}

		evmTxHash, err := common.Keccak256(deposit.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to create txHash. err: %w", err)
		}

		innerActionTxHash = ethgo.BytesToHash(evmTxHash)
	case withdrawEventSig:
		withdraw, err := contract.GatewayFilterer.ParseWithdraw(parsedLog)
		if err != nil {
			r.logger.Error("failed to parse withdraw event", "err", err)

			return nil, err
		}

		txs := make([]core.BridgingRequestEthMetadataTransaction, len(withdraw.Receivers))
		for idx, tx := range withdraw.Receivers {
			txs[idx] = core.BridgingRequestEthMetadataTransaction{
				Amount:  tx.Amount,
				Address: tx.Receiver,
			}
		}

		bridgingRequestMetadata := core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ToStrChainID(withdraw.DestinationChainId),
			SenderAddr:         withdraw.Sender.String(),
			Transactions:       txs,
			FeeAmount:          withdraw.FeeAmount,
		}

		metadata, err = core.MarshalEthMetadata(bridgingRequestMetadata)
		if err != nil {
			r.logger.Error("failed to marshal metadata", "err", err)

			return nil, err
		}

		txValue = withdraw.Value
	case fundedEventSig:
		funded, err := contract.GatewayFilterer.ParseFundsDeposited(parsedLog)
		if err != nil {
			r.logger.Error("failed to parse funds deposited event", "err", err)

			return nil, err
		}

		txValue = funded.Value
	case vscEventSig:
		vsc, err := contract.GatewayFilterer.ParseValidatorSetUpdatedGW(parsedLog)
		if err != nil {
			r.logger.Error("failed to parse validator set updated gw event", "err", err)

			return nil, err
		}

		evmTx, err := eth.NewEVMValidatorSetChangeTransaction(vsc.Data)
		if err != nil {
			r.logger.Error("failed to create new evm smart contract tx", "err", err)

			return nil, err
		}

		batchExecutedMetadata := core.BatchExecutedEthMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
			BatchNonceID:   evmTx.BatchNonceID,
		}

		metadata, err = core.MarshalEthMetadata(batchExecutedMetadata)
		if err != nil {
			r.logger.Error("failed to marshal metadata", "err", err)

			return nil, err
		}

		evmTxHash, err := common.Keccak256(vsc.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to create txHash. err: %w", err)
		}

		innerActionTxHash = ethgo.BytesToHash(evmTxHash)
	default:
		r.logger.Error("unknown event type in log", "log", log)

		return nil, fmt.Errorf("unknown event type in unprocessed log")
	}

	return &core.EthTx{
		OriginChainID: originChainID,
		Priority:      1,

		BlockNumber:     log.BlockNumber,
		BlockHash:       log.BlockHash,
		Hash:            log.TransactionHash,
		TxIndex:         log.TransactionIndex,
		Removed:         log.Removed,
		LogIndex:        log.LogIndex,
		Address:         log.Address,
		Metadata:        metadata,
		Value:           txValue,
		InnerActionHash: innerActionTxHash,
	}, nil
}
