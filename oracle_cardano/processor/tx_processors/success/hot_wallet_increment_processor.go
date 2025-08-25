package successtxprocessors

import (
	"fmt"
	"math/big"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
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
		return fmt.Errorf("validation failed for tx: %s, err: %w", tx.Hash, err)
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

	var (
		totalAmount        = big.NewInt(0)
		totalAmountWrapped = big.NewInt(0)
	)

	for _, output := range tx.Outputs {
		if utils.IsBridgingAddrForChain(appConfig, tx.OriginChainID, output.Address) {
			totalAmount.Add(totalAmount, new(big.Int).SetUint64(output.Amount))

			if len(chainConfig.NativeTokens) > 0 {
				wrappedToken, err := cardanotx.GetNativeTokenFromConfig(chainConfig.NativeTokens[0])
				if err != nil {
					return err
				}

				totalAmountWrapped.Add(
					totalAmountWrapped, new(big.Int).SetUint64(cardanotx.GetTokenAmount(output, wrappedToken.String())))
			}
		}
	}

	claims.HotWalletIncrementClaims = append(claims.HotWalletIncrementClaims, cCore.HotWalletIncrementClaim{
		ChainId:       common.ToNumChainID(tx.OriginChainID),
		Amount:        totalAmount,
		AmountWrapped: totalAmountWrapped,
	})

	p.logger.Info("Added HotWalletIncrementClaim",
		"chain", tx.OriginChainID, "Amount", totalAmount, "Increment", true)

	return nil
}

func (p *HotWalletIncrementProcessor) validate(
	tx *core.CardanoTx, appConfig *cCore.AppConfig,
) error {
	chainConfig := appConfig.CardanoChains[tx.OriginChainID]
	if chainConfig == nil {
		return fmt.Errorf("unsupported chain id found in tx. chain id: %v", tx.OriginChainID)
	}

	if err := utils.ValidateOutputsHaveUnknownTokens(tx, appConfig); err != nil {
		return err
	}

	if _, err := utils.ValidateTxOutputs(tx, appConfig, true); err != nil {
		return err
	}

	if len(tx.Metadata) != 0 {
		return fmt.Errorf("metadata should be empty")
	}

	return nil
}
