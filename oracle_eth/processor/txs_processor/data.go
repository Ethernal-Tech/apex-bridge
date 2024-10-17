package processor

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
)

type perTickState struct {
	allInvalidRelevantExpired []*core.BridgeExpectedEthTx
	allProcessedExpected      []*core.BridgeExpectedEthTx
	allProcessedInvalid       []*core.EthTx
	allProcessedValid         []*core.EthTx
	allUnprocessed            []*core.EthTx

	expectedTxsMap map[string]*core.BridgeExpectedEthTx
	unprocessedTxs []*core.EthTx
	blockInfo      *core.BridgeClaimsBlockInfo
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

func (pc *txProcessorsCollection) getSuccess(metadataJSON []byte) (
	core.EthTxSuccessProcessor, error,
) {
	metadata, err := core.UnmarshalEthMetadata[core.BaseEthMetadata](
		metadataJSON)
	if err != nil {
		return nil, err
	}

	txProcessor, relevant := pc.successTxProcessors[string(metadata.BridgingTxType)]
	if !relevant {
		return nil, fmt.Errorf("irrelevant tx. Tx type: %s", metadata.BridgingTxType)
	}

	return txProcessor, nil
}

func (pc *txProcessorsCollection) getFailed(metadataJSON []byte) (
	core.EthTxFailedProcessor, error,
) {
	metadata, err := core.UnmarshalEthMetadata[core.BaseEthMetadata](
		metadataJSON)
	if err != nil {
		return nil, err
	}

	txProcessor, relevant := pc.failedTxProcessors[string(metadata.BridgingTxType)]
	if !relevant {
		return nil, fmt.Errorf("irrelevant tx. Tx type: %s", metadata.BridgingTxType)
	}

	return txProcessor, nil
}
