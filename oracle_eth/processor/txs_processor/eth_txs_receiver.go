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

	if _, exists := r.appConfig.EthChains[originChainID]; !exists {
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
func (r *EthTxsReceiverImpl) getBridgingRequestStateDetails(
	originChainID string, txMetadata []byte,
) (string, *common.BridgingRequestStateDetails) {
	chainConfig := r.appConfig.EthChains[originChainID]
	if chainConfig == nil {
		r.logger.Warn("no chain config for bridging request state details", "chainID", originChainID)

		return "", nil
	}

	metadata, err := core.UnmarshalEthMetadata[core.BridgingRequestEthMetadata](txMetadata)
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
			Address: receiver.Address,
			Amount:  receiver.Amount,
			TokenID: receiver.TokenID,
		}
	}

	return metadata.DestinationChainID, common.NewBridgingRequestStateDetails(
		metadata.SenderAddr, receivers, metadata.BridgingFee, metadata.OperationFee,
		currencyID)
}

func (r *EthTxsReceiverImpl) logToTx(originChainID string, log *ethgo.Log) (*core.EthTx, error) {
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

	logEventType := log.Topics[0]

	metadata, innerActionTxHash, txValue, err := r.processLog(log, parsedLog, logEventType)
	if err != nil {
		return nil, fmt.Errorf("failed to process log. err: %w", err)
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

func (r *EthTxsReceiverImpl) processLog(log *ethgo.Log, parsedLog types.Log, logEventType ethgo.Hash) (
	[]byte, ethgo.Hash, *big.Int, error,
) {
	var (
		metadata          []byte
		innerActionTxHash ethgo.Hash
		txValue           *big.Int
	)

	events, err := eth.GetGatewayEventSignatures()
	if err != nil {
		return nil, ethgo.Hash{}, nil, fmt.Errorf("failed to get gateway event signatures. err: %w", err)
	}

	depositEventSig := events[0]
	withdrawEventSig := events[1]
	fundedEventSig := events[2]

	gatewayContract, err := contractbinding.NewGateway(ethereum_common.Address{}, nil)
	if err != nil {
		r.logger.Error("failed to get contractbinding gateway", "err", err)

		return nil, ethgo.Hash{}, nil, fmt.Errorf("failed to get contractbinding gateway. err: %w", err)
	}

	switch logEventType {
	case depositEventSig:
		deposit, err := gatewayContract.GatewayFilterer.ParseDeposit(parsedLog)
		if err != nil {
			r.logger.Error("failed to parse deposit event", "err", err)

			return nil, ethgo.Hash{}, nil, err
		}

		evmTx, err := eth.NewEVMSmartContractTransaction(deposit.Data)
		if err != nil {
			r.logger.Error("failed to create new evm smart contract tx", "err", err)

			return nil, ethgo.Hash{}, nil, err
		}

		batchExecutedMetadata := core.BatchExecutedEthMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
			BatchNonceID:   evmTx.BatchNonceID,
		}

		metadata, err = core.MarshalEthMetadata(batchExecutedMetadata)
		if err != nil {
			r.logger.Error("failed to marshal metadata", "err", err)

			return nil, ethgo.Hash{}, nil, err
		}

		evmTxHash, err := common.Keccak256(deposit.Data)
		if err != nil {

			return nil, ethgo.Hash{}, nil, fmt.Errorf("failed to create txHash. err: %w", err)
		}

		innerActionTxHash = ethgo.BytesToHash(evmTxHash)
	case withdrawEventSig:
		withdraw, err := gatewayContract.GatewayFilterer.ParseWithdraw(parsedLog)
		if err != nil {
			r.logger.Error("failed to parse withdraw event", "err", err)

			return nil, ethgo.Hash{}, nil, err
		}

		txs := make([]core.BridgingRequestEthMetadataTransaction, len(withdraw.Receivers))
		for idx, tx := range withdraw.Receivers {
			txs[idx] = core.BridgingRequestEthMetadataTransaction{
				Amount:  tx.Amount,
				Address: tx.Receiver,
				TokenID: tx.TokenId,
			}
		}

		bridgingRequestMetadata := core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: r.appConfig.ChainIDConverter.ToChainIDStr(withdraw.DestinationChainId),
			SenderAddr:         withdraw.Sender.String(),
			Transactions:       txs,
			BridgingFee:        withdraw.Fee,
			OperationFee:       withdraw.OperationFee,
		}

		metadata, err = core.MarshalEthMetadata(bridgingRequestMetadata)
		if err != nil {
			r.logger.Error("failed to marshal metadata", "err", err)

			return nil, ethgo.Hash{}, nil, err
		}

		txValue = withdraw.Value
	case fundedEventSig:
		funded, err := gatewayContract.GatewayFilterer.ParseFundsDeposited(parsedLog)
		if err != nil {
			r.logger.Error("failed to parse funds deposited event", "err", err)

			return nil, ethgo.Hash{}, nil, err
		}

		txValue = funded.Value

	default:
		r.logger.Error("unknown event type in log", "log", log)

		return nil, ethgo.Hash{}, nil, fmt.Errorf("unknown event type in unprocessed log")
	}

	return metadata, innerActionTxHash, txValue, nil
}
