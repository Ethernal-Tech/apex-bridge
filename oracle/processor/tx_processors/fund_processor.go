package txprocessors

import (
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/hashicorp/go-hclog"
)

var _ core.CardanoTxProcessor = (*FundProcessorImpl)(nil)

type FundProcessorImpl struct {
	logger hclog.Logger
}

func NewFundProcessor(logger hclog.Logger) *FundProcessorImpl {
	return &FundProcessorImpl{
		logger: logger.Named("fund_processor"),
	}
}

func (*FundProcessorImpl) GetType() common.BridgingTxType {
	return common.BridgingTxTypeFund
}

func (p *FundProcessorImpl) ValidateAndAddClaim(
	claims *core.BridgeClaims, tx *core.CardanoTx, appConfig *core.AppConfig,
) error {
	metadata, err := common.UnmarshalMetadata[common.FundMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: tx: %v, err: %w", tx, err)
	}

	if metadata.BridgingTxType != p.GetType() {
		return fmt.Errorf("ValidateAndAddClaim called for irrelevant tx: %v", tx)
	}

	p.logger.Debug("Validating relevant tx", "txHash", tx.Hash, "metadata", metadata)

	if err := p.validate(tx, appConfig); err != nil {
		return fmt.Errorf("validation failed for tx: %v, err: %w", tx, err)
	}

	p.addFundClaim(appConfig, claims, tx, metadata)

	return nil
}

func (p *FundProcessorImpl) addFundClaim(
	appConfig *core.AppConfig, claims *core.BridgeClaims,
	tx *core.CardanoTx, metadata *common.FundMetadata,
) {
	sum := big.NewInt(0)
	bridgingAddrUtxos := make([]core.UTXO, 0)
	addrs := appConfig.CardanoChains[tx.OriginChainID].BridgingAddresses

	for idx, utxo := range tx.Outputs {
		if utxo.Address == addrs.BridgingAddress {
			amount := new(big.Int).SetUint64(utxo.Amount)

			bridgingAddrUtxos = append(bridgingAddrUtxos, core.UTXO{
				TxHash:  tx.Hash,
				TxIndex: new(big.Int).SetUint64(uint64(idx)),
				Amount:  amount,
			})

			sum.Add(sum, amount)
		}
	}

	// claim := core.BatchExecutedClaim{
	// 	ObservedTransactionHash: tx.Hash,
	// 	ChainID:                 tx.OriginChainID,
	// 	BatchNonceID:            new(big.Int).SetUint64(metadata.BatchNonceID),
	// 	OutputUTXOs: core.UTXOs{
	// 		MultisigOwnedUTXOs: bridgingAddrUtxos,
	// 		FeePayerOwnedUTXOs: feeAddrUtxos,
	// 	},
	// }

	// claims.BatchExecutedClaims = append(claims.BatchExecutedClaims, claim)

	p.logger.Info("Added FundClaim",
		"txHash", tx.Hash, "metadata", metadata /*, "claim", core.BatchExecutedClaimString(claim)*/)
}

func (*FundProcessorImpl) validate(
	tx *core.CardanoTx, appConfig *core.AppConfig,
) error {
	sum := uint64(0)
	bridgingAddr := appConfig.CardanoChains[tx.OriginChainID].BridgingAddresses.BridgingAddress

	for _, utxo := range tx.Outputs {
		if utxo.Address == bridgingAddr {
			sum += utxo.Amount
		}
	}

	if sum < appConfig.BridgingSettings.UtxoMinValue {
		return fmt.Errorf("not enough to fund the bridging address. tried to fund with: %v", sum)
	}

	return nil
}
