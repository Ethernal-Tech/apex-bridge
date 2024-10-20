package processor

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
)

type perTickState struct {
	allInvalidRelevantExpired []*core.BridgeExpectedCardanoTx
	allProcessedExpected      []*core.BridgeExpectedCardanoTx
	allProcessedInvalid       []*core.CardanoTx
	allProcessedValid         []*core.CardanoTx
	allUnprocessed            []*core.CardanoTx

	expectedTxsMap map[string]*core.BridgeExpectedCardanoTx
	unprocessedTxs []*core.CardanoTx
	blockInfo      *core.BridgeClaimsBlockInfo
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

func (pc *txProcessorsCollection) getSuccess(metadataBytes []byte) (
	core.CardanoTxSuccessProcessor, error,
) {
	// return HotWallet fund processor if no metadata is found, address validation will be done later
	if len(metadataBytes) == 0 {
		return pc.successTxProcessors[string(common.BridgingTxTypeHotWalletFund)], nil
	}

	metadata, err := common.UnmarshalMetadata[common.BaseMetadata](common.MetadataEncodingTypeCbor, metadataBytes)
	if err != nil {
		return nil, err
	}

	txProcessor, relevant := pc.successTxProcessors[string(metadata.BridgingTxType)]
	if !relevant {
		return nil, fmt.Errorf("irrelevant tx. Tx type: %s", metadata.BridgingTxType)
	}

	return txProcessor, nil
}

func (pc *txProcessorsCollection) getFailed(metadataBytes []byte) (
	core.CardanoTxFailedProcessor, error,
) {
	metadata, err := common.UnmarshalMetadata[common.BaseMetadata](common.MetadataEncodingTypeCbor, metadataBytes)
	if err != nil {
		return nil, err
	}

	txProcessor, relevant := pc.failedTxProcessors[string(metadata.BridgingTxType)]
	if !relevant {
		return nil, fmt.Errorf("irrelevant tx. Tx type: %s", metadata.BridgingTxType)
	}

	return txProcessor, nil
}
