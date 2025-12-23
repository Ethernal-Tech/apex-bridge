package processor

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
)

type perTickState struct {
	updateData *core.CardanoUpdateTxsData

	// duplicated data, used for easier marking of invalid state for bridging request history
	allProcessedInvalid []*core.CardanoTx

	expectedTxsMap map[string]*core.BridgeExpectedCardanoTx
	unprocessedTxs []*core.CardanoTx
	blockInfo      *core.BridgeClaimsBlockInfo

	lastObservedPerChain map[string]uint64
}

type txProcessorsCollection struct {
	successTxProcessors map[string]core.CardanoTxSuccessProcessor
	failedTxProcessors  map[string]core.CardanoTxFailedProcessor
}

func NewTxProcessorsCollection(
	successTxProcessors []core.CardanoTxSuccessProcessor,
	failedTxProcessors []core.CardanoTxFailedProcessor,
) *txProcessorsCollection {
	successTxProcessorsMap := make(map[string]core.CardanoTxSuccessProcessor, len(successTxProcessors))
	for _, txProcessor := range successTxProcessors {
		successTxProcessorsMap[string(txProcessor.GetType())] = txProcessor
	}

	failedTxProcessorsMap := make(map[string]core.CardanoTxFailedProcessor, len(failedTxProcessors))
	for _, txProcessor := range failedTxProcessors {
		failedTxProcessorsMap[string(txProcessor.GetType())] = txProcessor
	}

	return &txProcessorsCollection{
		successTxProcessors: successTxProcessorsMap,
		failedTxProcessors:  failedTxProcessorsMap,
	}
}

func (pc *txProcessorsCollection) getSuccess(tx *core.CardanoTx, appConfig *cCore.AppConfig) (
	core.CardanoTxSuccessProcessor, error,
) {
	var (
		txProcessor core.CardanoTxSuccessProcessor
		relevant    bool
	)

	if len(tx.Metadata) != 0 {
		metadata, err := common.UnmarshalMetadata[common.BaseMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)
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

func (pc *txProcessorsCollection) getFailed(tx *core.BridgeExpectedCardanoTx, appConfig *cCore.AppConfig) (
	core.CardanoTxFailedProcessor, error,
) {
	metadata, err := common.UnmarshalMetadata[common.BaseMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)
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
