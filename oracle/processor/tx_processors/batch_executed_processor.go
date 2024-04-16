package tx_processors

import (
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/utils"
)

var _ core.CardanoTxProcessor = (*BatchExecutedProcessorImpl)(nil)

type BatchExecutedProcessorImpl struct {
}

func NewBatchExecutedProcessor() *BatchExecutedProcessorImpl {
	return &BatchExecutedProcessorImpl{}
}

func (*BatchExecutedProcessorImpl) GetType() core.TxProcessorType {
	return core.TxProcessorTypeBatchExecuted
}

func (*BatchExecutedProcessorImpl) IsTxRelevant(tx *core.CardanoTx) (bool, error) {
	metadata, err := common.UnmarshalMetadata[common.BaseMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)

	if err == nil && metadata != nil {
		return metadata.BridgingTxType == common.BridgingTxTypeBatchExecution, err
	}

	return false, err
}

func (p *BatchExecutedProcessorImpl) ValidateAndAddClaim(claims *core.BridgeClaims, tx *core.CardanoTx, appConfig *core.AppConfig) error {
	relevant, err := p.IsTxRelevant(tx)
	if err != nil {
		return err
	}

	if !relevant {
		return fmt.Errorf("ValidateAndAddClaim called for irrelevant tx: %v", tx)
	}

	metadata, err := common.UnmarshalMetadata[common.BatchExecutedMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: tx: %v,\n err: %v", tx, err)
	}

	if err := p.validate(tx, metadata, appConfig); err != nil {
		return fmt.Errorf("validation failed for tx: %v, err: %v", tx, err)
	}

	p.addBatchExecutedClaim(claims, tx, metadata)

	return nil
}

func (*BatchExecutedProcessorImpl) addBatchExecutedClaim(claims *core.BridgeClaims, tx *core.CardanoTx, metadata *common.BatchExecutedMetadata) {
	var utxos []core.UTXO
	for _, utxo := range tx.Outputs {
		utxos = append(utxos, core.UTXO{
			TxHash:  tx.Hash,
			TxIndex: new(big.Int).SetUint64(uint64(tx.Indx)),
			Amount:  new(big.Int).SetUint64(utxo.Amount),
		})
	}

	claim := core.BatchExecutedClaim{
		ObservedTransactionHash: tx.Hash,
		ChainID:                 tx.OriginChainId,
		BatchNonceID:            new(big.Int).SetUint64(metadata.BatchNonceId),
		OutputUTXOs: core.UTXOs{
			MultisigOwnedUTXOs: utxos,
		},
	}

	claims.BatchExecutedClaims = append(claims.BatchExecutedClaims, claim)
}

func (*BatchExecutedProcessorImpl) validate(tx *core.CardanoTx, metadata *common.BatchExecutedMetadata, appConfig *core.AppConfig) error {
	// after BridgingTxType and inputs are validated, no further validation needed
	return utils.ValidateTxInputs(tx, appConfig)
}
