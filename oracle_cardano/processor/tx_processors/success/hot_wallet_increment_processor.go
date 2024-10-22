package successtxprocessors

import (
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	cUtils "github.com/Ethernal-Tech/apex-bridge/oracle_common/utils"
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
	totalAmount := big.NewInt(0)

	for _, output := range tx.Outputs {
		totalAmount.Add(totalAmount, new(big.Int).SetUint64(output.Amount))
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
	chainConfig := appConfig.CardanoChains[tx.OriginChainID]
	if chainConfig == nil {
		return fmt.Errorf("unsupported chain id found in tx. chain id: %v", tx.OriginChainID)
	}

	if len(tx.Tx.Outputs) == 0 {
		return fmt.Errorf("no outputs found in tx")
	}

	for _, utxo := range tx.Tx.Outputs {
		if utxo.Address != chainConfig.BridgingAddresses.BridgingAddress {
			return fmt.Errorf("bridging address on origin not found in utxos")
		}
	}

	if len(tx.Metadata) != 0 {
		return fmt.Errorf("metadata should be empty")
	}

	cardanoSrcConfig, _ := cUtils.GetChainConfig(appConfig, tx.OriginChainID)
	if cardanoSrcConfig == nil {
		return fmt.Errorf("origin chain not registered: %v", tx.OriginChainID)
	}

	return nil
}
