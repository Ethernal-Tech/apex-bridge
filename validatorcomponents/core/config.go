package core

import (
	"encoding/json"

	batcherCore "github.com/Ethernal-Tech/apex-bridge/batcher/core"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/hashicorp/go-hclog"
)

type CardanoChainConfig struct {
	NetworkAddress           string                       `json:"networkAddress"`
	NetworkMagic             uint32                       `json:"networkMagic"`
	StartBlockHash           string                       `json:"startBlockHash"`
	StartSlot                uint64                       `json:"startSlot"`
	StartBlockNumber         uint64                       `json:"startBlockNumber"`
	ConfirmationBlockCount   uint                         `json:"confirmationBlockCount"`
	BridgingAddresses        oracleCore.BridgingAddresses `json:"bridgingAddresses"`
	OtherAddressesOfInterest []string                     `json:"otherAddressesOfInterest"`
	KeysDirPath              string                       `json:"keysDirPath"`
	BlockfrostUrl            string                       `json:"blockfrostUrl"`
	BlockfrostAPIKey         string                       `json:"blockfrostApiKey"`
	AtLeastValidators        float64                      `json:"atLeastValidators"`
	PotentialFee             uint64                       `json:"potentialFee"`
}

type AppConfig struct {
	CardanoChains        map[string]*CardanoChainConfig `json:"cardanoChains"`
	Bridge               oracleCore.BridgeConfig        `json:"bridge"`
	BridgingSettings     oracleCore.BridgingSettings    `json:"bridgingSettings"`
	Settings             oracleCore.AppSettings         `json:"appSettings"`
	BatcherPullTimeMilis uint64                         `json:"batcherPullTime"`
	InitialUtxos         oracleCore.InitialUtxos        `json:"initialUtxos"`
}

func (appConfig *AppConfig) SeparateConfigs() (*oracleCore.AppConfig, *batcherCore.BatcherManagerConfiguration) {
	oracleCardanoChains := make(map[string]*oracleCore.CardanoChainConfig)
	batcherChains := make(map[string]batcherCore.ChainConfig)

	for chainId, ccConfig := range appConfig.CardanoChains {
		oracleCardanoChains[chainId] = &oracleCore.CardanoChainConfig{
			ChainId:                  chainId,
			NetworkAddress:           ccConfig.NetworkAddress,
			NetworkMagic:             ccConfig.NetworkMagic,
			StartBlockHash:           ccConfig.StartBlockHash,
			StartSlot:                ccConfig.StartSlot,
			StartBlockNumber:         ccConfig.StartBlockNumber,
			ConfirmationBlockCount:   ccConfig.ConfirmationBlockCount,
			BridgingAddresses:        ccConfig.BridgingAddresses,
			OtherAddressesOfInterest: ccConfig.OtherAddressesOfInterest,
		}

		chainSpecificJsonRaw, _ := json.Marshal(batcherCore.CardanoChainConfig{
			TestNetMagic:      uint(ccConfig.NetworkMagic),
			BlockfrostUrl:     ccConfig.BlockfrostUrl,
			BlockfrostAPIKey:  ccConfig.BlockfrostAPIKey,
			AtLeastValidators: ccConfig.AtLeastValidators,
			PotentialFee:      ccConfig.PotentialFee,
		})

		batcherChains[chainId] = batcherCore.ChainConfig{
			Base: batcherCore.BaseConfig{
				ChainId:     chainId,
				KeysDirPath: ccConfig.KeysDirPath,
			},
			ChainSpecific: batcherCore.ChainSpecific{
				ChainType: "Cardano",
				Config:    chainSpecificJsonRaw,
			},
		}
	}

	oracleConfig := &oracleCore.AppConfig{
		Bridge:           appConfig.Bridge,
		Settings:         appConfig.Settings,
		BridgingSettings: appConfig.BridgingSettings,
		InitialUtxos:     appConfig.InitialUtxos,
		CardanoChains:    oracleCardanoChains,
	}

	batcherConfig := &batcherCore.BatcherManagerConfiguration{
		Bridge: batcherCore.BridgeConfig{
			NodeUrl:              appConfig.Bridge.NodeUrl,
			SmartContractAddress: appConfig.Bridge.SmartContractAddress,
			SecretsManager:       appConfig.Bridge.SecretsManager,
		},
		PullTimeMilis: appConfig.BatcherPullTimeMilis,
		Logger: logger.LoggerConfig{
			LogFilePath:   appConfig.Settings.LogsPath + "batcher",
			LogLevel:      hclog.Level(appConfig.Settings.LogLevel),
			JSONLogFormat: false,
			AppendFile:    true,
		},
		Chains: batcherChains,
	}

	return oracleConfig, batcherConfig
}
