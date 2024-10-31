package successtxprocessors

import (
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/utils"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/hashicorp/go-hclog"
)

var _ core.CardanoTxSuccessProcessor = (*HotWalletIncrementProcessor)(nil)

type HotWalletIncrementProcessor struct {
	logger hclog.Logger
}

func NewHotWalletIncrementProcessor(logger hclog.Logger) *HotWalletIncrementProcessor {
	return &HotWalletIncrementProcessor{
		logger: logger.Named("hot_wallet_increment_processor"),
	}
}

func (*HotWalletIncrementProcessor) GetType() common.BridgingTxType {
	return common.TxTypeHotWalletFund
}

func (p *HotWalletIncrementProcessor) PreValidate(tx *core.CardanoTx, appConfig *cCore.AppConfig) error {
	if err := p.validate(tx, appConfig); err != nil {
		return fmt.Errorf("validation failed for tx: %v, err: %w", tx, err)
	}

	return nil
}

func (p *HotWalletIncrementProcessor) ValidateAndAddClaim(
	claims *cCore.BridgeClaims, tx *core.CardanoTx, appConfig *cCore.AppConfig,
) error {
	chainConfig := appConfig.CardanoChains[tx.OriginChainID]
	if chainConfig == nil {
		return fmt.Errorf("origin chain not registered: %v", tx.OriginChainID)
	}

	p.logger.Debug("Validating relevant tx", "txHash", tx.Hash)

	if err := p.validate(tx, appConfig); err != nil {
		return fmt.Errorf("validation failed for tx: %v, err: %w", tx, err)
	}

	totalAmount := big.NewInt(0)

	for _, output := range tx.Outputs {
		if output.Address == chainConfig.BridgingAddresses.BridgingAddress {
			totalAmount.Add(totalAmount, new(big.Int).SetUint64(output.Amount))
		}
	}

	claims.HotWalletIncrementClaims = append(claims.HotWalletIncrementClaims, cCore.HotWalletIncrementClaim{
		ChainId:     common.ToNumChainID(tx.OriginChainID),
		Amount:      totalAmount,
		IsIncrement: true,
	})

	p.logger.Info("Added HotWalletIncrementClaim",
		"chain", tx.OriginChainID, "Amount", totalAmount, "Increment", true)

	return nil
}

func (p *HotWalletIncrementProcessor) validate(
	tx *core.CardanoTx, appConfig *cCore.AppConfig,
) error {
	if _, err := utils.ValidateTxOutputs(tx, appConfig, true); err != nil {
		return err
	}

	if len(tx.Metadata) != 0 {
		return fmt.Errorf("metadata should be empty")
	}

	return nil
}
