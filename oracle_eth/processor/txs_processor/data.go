package processor

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
)

type perTickState struct {
	updateData *core.EthUpdateTxsData

	// duplicated data, used for easier marking of invalid state for bridging request history
	allProcessedInvalid           []*core.EthTx
	innerActionHashToActualTxHash map[string]common.Hash

	expectedTxsMap map[string]*core.BridgeExpectedEthTx
	unprocessedTxs []*core.EthTx
	blockInfo      *core.BridgeClaimsBlockInfo

	lastObservedPerChain map[string]uint64
}

type txProcessorsCollection struct {
	successTxProcessors map[string]core.EthTxSuccessProcessor
	failedTxProcessors  map[string]core.EthTxFailedProcessor
}

func NewTxProcessorsCollection(
	successTxProcessors []core.EthTxSuccessProcessor,
	failedTxProcessors []core.EthTxFailedProcessor,
) *txProcessorsCollection {
	successTxProcessorsMap := make(map[string]core.EthTxSuccessProcessor, len(successTxProcessors))
	for _, txProcessor := range successTxProcessors {
		successTxProcessorsMap[string(txProcessor.GetType())] = txProcessor
	}

	failedTxProcessorsMap := make(map[string]core.EthTxFailedProcessor, len(failedTxProcessors))
	for _, txProcessor := range failedTxProcessors {
		failedTxProcessorsMap[string(txProcessor.GetType())] = txProcessor
	}

	return &txProcessorsCollection{
		successTxProcessors: successTxProcessorsMap,
		failedTxProcessors:  failedTxProcessorsMap,
	}
}

func (pc *txProcessorsCollection) getSuccess(tx *core.EthTx, appConfig *cCore.AppConfig) (
	core.EthTxSuccessProcessor, error,
) {
	var (
		txProcessor core.EthTxSuccessProcessor
		relevant    bool
	)

	if len(tx.Metadata) != 0 {
		metadata, err := core.UnmarshalEthMetadata[core.BaseEthMetadata](tx.Metadata)
		if err != nil {
			return nil, err
		}

		txProcessor, relevant = pc.successTxProcessors[string(metadata.BridgingTxType)]
		if !relevant {
			txProcessor, relevant = pc.successTxProcessors[string(common.TxTypeRefundRequest)]
			if !relevant {
				return nil, fmt.Errorf("irrelevant tx. Tx type: %s", metadata.BridgingTxType)
			}
		}
	} else {
		txProcessor = pc.successTxProcessors[string(common.TxTypeHotWalletFund)]
	}

	if err := txProcessor.PreValidate(tx, appConfig); err != nil {
		return nil, err
	}

	return txProcessor, nil
}

func (pc *txProcessorsCollection) getFailed(tx *core.BridgeExpectedEthTx, appConfig *cCore.AppConfig) (
	core.EthTxFailedProcessor, error,
) {
	metadata, err := core.UnmarshalEthMetadata[core.BaseEthMetadata](tx.Metadata)
	if err != nil {
		return nil, err
	}

	txProcessor, relevant := pc.failedTxProcessors[string(metadata.BridgingTxType)]
	if !relevant {
		return nil, fmt.Errorf("irrelevant tx. Tx type: %s", metadata.BridgingTxType)
	}

	if err = txProcessor.PreValidate(tx, appConfig); err != nil {
		return nil, err
	}

	return txProcessor, nil
}
