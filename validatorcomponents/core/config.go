package core

import (
	batcherCore "github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
)

type CardanoChainConfig struct {
	NetworkAddress           string   `json:"networkAddress"`
	NetworkMagic             uint32   `json:"networkMagic"`
	StartBlockHash           string   `json:"startBlockHash"`
	StartSlot                uint64   `json:"startSlot"`
	StartBlockNumber         uint64   `json:"startBlockNumber"`
	ConfirmationBlockCount   uint     `json:"confirmationBlockCount"`
	OtherAddressesOfInterest []string `json:"otherAddressesOfInterest"`
	KeysDirPath              string   `json:"keysDirPath"`
	BlockfrostUrl            string   `json:"blockfrostUrl"`
	BlockfrostAPIKey         string   `json:"blockfrostApiKey"`
	SocketPath               string   `json:"socketPath"`
	PotentialFee             uint64   `json:"potentialFee"`
}

type ApiConfig struct {
	Port           uint32   `json:"port"`
	PathPrefix     string   `json:"pathPrefix"`
	AllowedHeaders []string `json:"allowedHeaders"`
	AllowedOrigins []string `json:"allowedOrigins"`
	AllowedMethods []string `json:"allowedMethods"`
	ApiKeyHeader   string   `json:"apiKeyHeader"`
	ApiKeys        []string `json:"apiKeys"`
}

type AppConfig struct {
	CardanoChains                map[string]*CardanoChainConfig `json:"cardanoChains"`
	Bridge                       oracleCore.BridgeConfig        `json:"bridge"`
	BridgingSettings             oracleCore.BridgingSettings    `json:"bridgingSettings"`
	Settings                     oracleCore.AppSettings         `json:"appSettings"`
	RelayerImitatorPullTimeMilis uint64                         `json:"relayerImitatorPullTime"`
	BatcherPullTimeMilis         uint64                         `json:"batcherPullTime"`
	ApiConfig                    ApiConfig                      `json:"api"`
}

func (appConfig *AppConfig) SeparateConfigs() (*oracleCore.AppConfig, *batcherCore.BatcherManagerConfiguration) {
	oracleCardanoChains := make(map[string]*oracleCore.CardanoChainConfig, len(appConfig.CardanoChains))
	batcherChains := make([]batcherCore.ChainConfig, 0, len(appConfig.CardanoChains))

	for chainId, ccConfig := range appConfig.CardanoChains {
		oracleCardanoChains[chainId] = &oracleCore.CardanoChainConfig{
			ChainId:                  chainId,
			NetworkAddress:           ccConfig.NetworkAddress,
			NetworkMagic:             ccConfig.NetworkMagic,
			StartBlockHash:           ccConfig.StartBlockHash,
			StartSlot:                ccConfig.StartSlot,
			StartBlockNumber:         ccConfig.StartBlockNumber,
			ConfirmationBlockCount:   ccConfig.ConfirmationBlockCount,
			OtherAddressesOfInterest: ccConfig.OtherAddressesOfInterest,
		}

		chainSpecificJsonRaw, _ := (cardanotx.CardanoChainConfig{
			TestNetMagic:     ccConfig.NetworkMagic,
			BlockfrostUrl:    ccConfig.BlockfrostUrl,
			BlockfrostAPIKey: ccConfig.BlockfrostAPIKey,
			SocketPath:       ccConfig.SocketPath,
			PotentialFee:     ccConfig.PotentialFee,
			KeysDirPath:      ccConfig.KeysDirPath,
		}).Serialize()

		batcherChains = append(batcherChains, batcherCore.ChainConfig{
			ChainId:       chainId,
			ChainType:     "Cardano",
			ChainSpecific: chainSpecificJsonRaw,
		})
	}

	oracleConfig := &oracleCore.AppConfig{
		Bridge:           appConfig.Bridge,
		Settings:         appConfig.Settings,
		BridgingSettings: appConfig.BridgingSettings,
		CardanoChains:    oracleCardanoChains,
	}

	batcherConfig := &batcherCore.BatcherManagerConfiguration{
		Bridge: batcherCore.BridgeConfig{
			NodeUrl:              appConfig.Bridge.NodeUrl,
			SmartContractAddress: appConfig.Bridge.SmartContractAddress,
			ValidatorDataDir:     appConfig.Bridge.ValidatorDataDir,
			ValidatorConfigPath:  appConfig.Bridge.ValidatorConfigPath,
		},
		PullTimeMilis: appConfig.BatcherPullTimeMilis,
		Chains:        batcherChains,
	}

	return oracleConfig, batcherConfig
}
