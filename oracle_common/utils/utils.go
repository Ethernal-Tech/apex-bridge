package utils

import "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"

func GetChainConfig(appConfig *core.AppConfig, chainID string) (*core.CardanoChainConfig, *core.EthChainConfig) {
	if cardanoChainConfig, exists := appConfig.CardanoChains[chainID]; exists {
		return cardanoChainConfig, nil
	}

	if ethChainConfig, exists := appConfig.EthChains[chainID]; exists {
		return nil, ethChainConfig
	}

	return nil, nil
}
