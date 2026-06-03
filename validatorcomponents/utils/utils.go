package utils

import (
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
)

func GetChainConfig(
	appConfig *core.AppConfig, chainID string,
) (*oCore.CardanoChainConfig, *oCore.EthChainConfig, *oCore.SolanaChainConfig) {
	if cardanoChainConfig, exists := appConfig.CardanoChains[chainID]; exists {
		return cardanoChainConfig, nil, nil
	}

	if ethChainConfig, exists := appConfig.EthChains[chainID]; exists {
		return nil, ethChainConfig, nil
	}

	if solanaChainConfig, exists := appConfig.SolanaChains[chainID]; exists {
		return nil, nil, solanaChainConfig
	}

	return nil, nil, nil
}
