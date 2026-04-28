package processor

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	solana "github.com/gagliardetto/solana-go"
)

type perTickState struct {
	updateData *core.SolanaUpdateTxsData

	// duplicated data, used for easier marking of invalid state for bridging request history
	allProcessedInvalid           []*core.SolanaTx
	innerActionHashToActualTxHash map[string]solana.Signature

	expectedTxsMap map[string]*core.BridgeExpectedSolanaTx
	unprocessedTxs []*core.SolanaTx
	blockInfo      *core.BridgeClaimsSlotInfo
}

type txProcessorsCollection struct {
	successTxProcessors map[string]core.SolanaTxSuccessProcessor
	failedTxProcessors  map[string]core.SolanaTxFailedProcessor
}

func NewTxProcessorsCollection(
	successTxProcessors []core.SolanaTxSuccessProcessor,
	failedTxProcessors []core.SolanaTxFailedProcessor,
) *txProcessorsCollection {
	successTxProcessorsMap := make(map[string]core.SolanaTxSuccessProcessor, len(successTxProcessors))
	for _, txProcessor := range successTxProcessors {
		successTxProcessorsMap[string(txProcessor.GetType())] = txProcessor
	}

	failedTxProcessorsMap := make(map[string]core.SolanaTxFailedProcessor, len(failedTxProcessors))
	for _, txProcessor := range failedTxProcessors {
		failedTxProcessorsMap[string(txProcessor.GetType())] = txProcessor
	}

	return &txProcessorsCollection{
		successTxProcessors: successTxProcessorsMap,
		failedTxProcessors:  failedTxProcessorsMap,
	}
}

func (c *txProcessorsCollection) getFailed(tx *core.BridgeExpectedSolanaTx, appConfig *oCore.AppConfig) (
	core.SolanaTxFailedProcessor, error,
) {
	metadata, err := core.UnmarshalSolMetadata[core.BaseSolMetadata](tx.Metadata)
	if err != nil {
		return nil, err
	}

	txProcessor, relevant := c.failedTxProcessors[string(metadata.BridgingTxType)]
	if !relevant {
		return nil, fmt.Errorf("irrelevant tx. Tx type: %s", metadata.BridgingTxType)
	}

	if err = txProcessor.PreValidate(tx, appConfig); err != nil {
		return nil, err
	}

	return txProcessor, nil
}

func (c *txProcessorsCollection) getSuccess(tx *core.SolanaTx, appConfig *oCore.AppConfig) (
	core.SolanaTxSuccessProcessor, error,
) {
	var (
		txProcessor core.SolanaTxSuccessProcessor
		relevant    bool
	)

	if len(tx.Metadata) != 0 {
		metadata, err := core.UnmarshalSolMetadata[core.BaseSolMetadata](tx.Metadata)
		if err != nil {
			return nil, err
		}

		txProcessor, relevant = c.successTxProcessors[string(metadata.BridgingTxType)]
		if !relevant {
			txProcessor, relevant = c.successTxProcessors[string(common.TxTypeRefundRequest)]
			if !relevant {
				return nil, fmt.Errorf("irrelevant tx. Tx type: %s", metadata.BridgingTxType)
			}
		}
	} else {
		txProcessor = c.successTxProcessors[string(common.TxTypeHotWalletFund)]
	}

	if err := txProcessor.PreValidate(tx, appConfig); err != nil {
		return nil, err
	}

	return txProcessor, nil
}
