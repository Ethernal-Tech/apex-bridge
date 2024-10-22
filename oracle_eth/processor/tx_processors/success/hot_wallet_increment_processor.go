package successtxprocessors

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	oUtils "github.com/Ethernal-Tech/apex-bridge/oracle_common/utils"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	"github.com/hashicorp/go-hclog"
)

var _ core.EthTxSuccessProcessor = (*HotWalletIncrementProcessor)(nil)

type HotWalletIncrementProcessor struct {
	logger hclog.Logger
}

func NewHotWalletIncrementProcessor(logger hclog.Logger) *HotWalletIncrementProcessor {
	return &HotWalletIncrementProcessor{
		logger: logger.Named("hot_wallet_increment_processor"),
	}
}

func (*HotWalletIncrementProcessor) GetType() common.BridgingTxType {
	return common.BridgingTxTypeHotWalletFund
}

func (p *HotWalletIncrementProcessor) ValidateAndAddClaim(
	claims *oCore.BridgeClaims, tx *core.EthTx, appConfig *oCore.AppConfig,
) error {
	if err := p.validate(tx, appConfig); err != nil {
		return fmt.Errorf("validation failed for tx: %v, err: %w", tx, err)
	}

	claims.HotWalletIncrementClaims = append(claims.HotWalletIncrementClaims, oCore.HotWalletIncrementClaim{
		ChainId:     common.ToNumChainID(tx.OriginChainID),
		Amount:      tx.Value,
		IsIncrement: true,
	})

	p.logger.Info("Added HotWalletIncrementClaim",
		"chain", tx.OriginChainID, "Amount", tx.Value, "Increment", true)

	return nil
}

func (p *HotWalletIncrementProcessor) validate(
	tx *core.EthTx, appConfig *oCore.AppConfig,
) error {
	chainConfig := appConfig.EthChains[tx.OriginChainID]

	_, ethSrcConfig := oUtils.GetChainConfig(appConfig, tx.OriginChainID)
	if ethSrcConfig == nil {
		return fmt.Errorf("origin chain not registered: %v", tx.OriginChainID)
	}

	if tx.Value == nil {
		return fmt.Errorf("tx value is nil")
	}

	if tx.Address.String() != chainConfig.BridgingAddresses.BridgingAddress {
		return fmt.Errorf("wrong address")
	}

	return nil
}
