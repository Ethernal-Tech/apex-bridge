package successtxprocessors

import (
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	"github.com/hashicorp/go-hclog"
)

var _ core.SolanaTxSuccessProcessor = (*HotWalletIncrementProcessor)(nil)

type HotWalletIncrementProcessor struct {
	logger hclog.Logger
}

func NewSolanaHotWalletIncrementProcessor(logger hclog.Logger) *HotWalletIncrementProcessor {
	return &HotWalletIncrementProcessor{
		logger: logger.Named("solana_hot_wallet_increment_processor"),
	}
}

func (*HotWalletIncrementProcessor) GetType() common.BridgingTxType {
	return common.TxTypeHotWalletFund
}

func (p *HotWalletIncrementProcessor) PreValidate(tx *core.SolanaTx, appConfig *oCore.AppConfig) error {
	if err := p.validate(tx, appConfig); err != nil {
		return fmt.Errorf("validation failed for tx: %v, err: %w", tx, err)
	}

	return nil
}

func (p *HotWalletIncrementProcessor) ValidateAndAddClaim(
	claims *oCore.BridgeClaims, tx *core.SolanaTx, appConfig *oCore.AppConfig,
) error {
	p.logger.Debug("Validating relevant tx", "txHash", tx.TxSignature, "metadata", tx.Metadata)

	if err := p.validate(tx, appConfig); err != nil {
		return fmt.Errorf("validation failed for tx: %v, err: %w", tx, err)
	}

	claims.HotWalletIncrementClaims = append(claims.HotWalletIncrementClaims, oCore.HotWalletIncrementClaim{
		ChainId:       appConfig.ChainIDConverter.ToChainIDNum(tx.OriginChainID),
		Amount:        tx.Value,
		AmountWrapped: big.NewInt(0),
		TxHash:        tx.TxSignature[:],
	})

	p.logger.Info("Added HotWalletIncrementClaim",
		"chain", tx.OriginChainID, "Amount", tx.Value, "Increment", true)

	return nil
}

func (*HotWalletIncrementProcessor) validate(tx *core.SolanaTx, appConfig *oCore.AppConfig) error {
	chainConfig := appConfig.SolanaChains[tx.OriginChainID]
	if chainConfig == nil {
		return fmt.Errorf("origin chain not registered: %v", tx.OriginChainID)
	}

	if len(tx.Metadata) != 0 {
		return fmt.Errorf("metadata should be empty")
	}

	if tx.Value == nil || tx.Value.BitLen() == 0 {
		return fmt.Errorf("tx value is zero or not set")
	}

	return nil
}
