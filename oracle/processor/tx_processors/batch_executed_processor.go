package txprocessors

import (
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/utils"
	"github.com/hashicorp/go-hclog"
)

var _ core.CardanoTxProcessor = (*BatchExecutedProcessorImpl)(nil)

type BatchExecutedProcessorImpl struct {
	logger hclog.Logger
}

func NewBatchExecutedProcessor(logger hclog.Logger) *BatchExecutedProcessorImpl {
	return &BatchExecutedProcessorImpl{
		logger: logger.Named("batch_executed_processor"),
	}
}

func (*BatchExecutedProcessorImpl) GetType() core.TxProcessorType {
	return core.TxProcessorTypeBatchExecuted
}

func (p *BatchExecutedProcessorImpl) IsTxRelevant(tx *core.CardanoTx) (bool, error) {
	p.logger.Debug("Checking if tx is relevant", "tx", tx)

	metadata, err := common.UnmarshalMetadata[common.BaseMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)

	p.logger.Debug("Unmarshaled metadata", "txHash", tx.Hash, "metadata", metadata, "err", err)

	if err == nil && metadata != nil {
		return metadata.BridgingTxType == common.BridgingTxTypeBatchExecution, err
	}

	return false, err
}

func (p *BatchExecutedProcessorImpl) ValidateAndAddClaim(
	claims *core.BridgeClaims, tx *core.CardanoTx, appConfig *core.AppConfig,
) error {
	relevant, err := p.IsTxRelevant(tx)
	if err != nil {
		return err
	}

	if !relevant {
		return fmt.Errorf("ValidateAndAddClaim called for irrelevant tx: %v", tx)
	}

	p.logger.Debug("tx is relevant", "txHash", tx.Hash)

	metadata, err := common.UnmarshalMetadata[common.BatchExecutedMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: tx: %v, err: %w", tx, err)
	}

	p.logger.Debug("Validating", "txHash", tx.Hash, "metadata", metadata)

	if err := p.validate(tx, metadata, appConfig); err != nil {
		return fmt.Errorf("validation failed for tx: %v, err: %w", tx, err)
	}

	p.addBatchExecutedClaim(appConfig, claims, tx, metadata)

	return nil
}

func (p *BatchExecutedProcessorImpl) addBatchExecutedClaim(
	appConfig *core.AppConfig, claims *core.BridgeClaims,
	tx *core.CardanoTx, metadata *common.BatchExecutedMetadata,
) {
	bridgingAddrUtxos := make([]core.UTXO, 0)
	feeAddrUtxos := make([]core.UTXO, 0)
	addrs := appConfig.CardanoChains[tx.OriginChainID].BridgingAddresses

	for _, utxo := range tx.Outputs {
		if utxo.Address == addrs.BridgingAddress {
			bridgingAddrUtxos = append(bridgingAddrUtxos, core.UTXO{
				TxHash:  tx.Hash,
				TxIndex: new(big.Int).SetUint64(uint64(tx.Indx)),
				Amount:  new(big.Int).SetUint64(utxo.Amount),
			})
		} else if utxo.Address == addrs.FeeAddress {
			feeAddrUtxos = append(feeAddrUtxos, core.UTXO{
				TxHash:  tx.Hash,
				TxIndex: new(big.Int).SetUint64(uint64(tx.Indx)),
				Amount:  new(big.Int).SetUint64(utxo.Amount),
			})
		}
	}

	claim := core.BatchExecutedClaim{
		ObservedTransactionHash: tx.Hash,
		ChainID:                 tx.OriginChainID,
		BatchNonceID:            new(big.Int).SetUint64(metadata.BatchNonceID),
		OutputUTXOs: core.UTXOs{
			MultisigOwnedUTXOs: bridgingAddrUtxos,
			FeePayerOwnedUTXOs: feeAddrUtxos,
		},
	}

	claims.BatchExecutedClaims = append(claims.BatchExecutedClaims, claim)

	p.logger.Info("Added BatchExecutedClaim",
		"txHash", tx.Hash, "metadata", metadata, "claim", core.BatchExecutedClaimString(claim))
}

func (*BatchExecutedProcessorImpl) validate(
	tx *core.CardanoTx, metadata *common.BatchExecutedMetadata, appConfig *core.AppConfig,
) error {
	// after BridgingTxType and inputs are validated, no further validation needed
	return utils.ValidateTxInputs(tx, appConfig)
}
