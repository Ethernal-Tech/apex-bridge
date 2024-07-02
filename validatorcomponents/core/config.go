package core

import (
	batcherCore "github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	ethOracleCore "github.com/Ethernal-Tech/apex-bridge/eth_oracle/core"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
)

type CardanoChainConfig struct {
	NetworkAddress           string   `json:"networkAddress"`
	NetworkID                uint32   `json:"networkID"`
	NetworkMagic             uint32   `json:"networkMagic"`
	StartBlockHash           string   `json:"startBlockHash"`
	StartSlot                uint64   `json:"startSlot"`
	StartBlockNumber         uint64   `json:"startBlockNumber"`
	TTLSlotNumberInc         uint64   `json:"ttlSlotNumberIncrement"`
	ConfirmationBlockCount   uint     `json:"confirmationBlockCount"`
	OtherAddressesOfInterest []string `json:"otherAddressesOfInterest"`
	OgmiosURL                string   `json:"ogmiosUrl"`
	BlockfrostURL            string   `json:"blockfrostUrl"`
	BlockfrostAPIKey         string   `json:"blockfrostApiKey"`
	SocketPath               string   `json:"socketPath"`
	PotentialFee             uint64   `json:"potentialFee"`
	SlotRoundingThreshold    uint64   `json:"slotRoundingThreshold"` // empty if we want to use value from sc
}

type APIConfig struct {
	Port           uint32   `json:"port"`
	PathPrefix     string   `json:"pathPrefix"`
	AllowedHeaders []string `json:"allowedHeaders"`
	AllowedOrigins []string `json:"allowedOrigins"`
	AllowedMethods []string `json:"allowedMethods"`
	APIKeyHeader   string   `json:"apiKeyHeader"`
	APIKeys        []string `json:"apiKeys"`
}

type AppConfig struct {
	ValidatorDataDir             string                                   `json:"validatorDataDir"`
	ValidatorConfigPath          string                                   `json:"validatorConfigPath"`
	CardanoChains                map[string]*CardanoChainConfig           `json:"cardanoChains"`
	EthChains                    map[string]*ethOracleCore.EthChainConfig `json:"ethChains"`
	Bridge                       oracleCore.BridgeConfig                  `json:"bridge"`
	BridgingSettings             oracleCore.BridgingSettings              `json:"bridgingSettings"`
	Settings                     oracleCore.AppSettings                   `json:"appSettings"`
	RelayerImitatorPullTimeMilis uint64                                   `json:"relayerImitatorPullTime"`
	BatcherPullTimeMilis         uint64                                   `json:"batcherPullTime"`
	APIConfig                    APIConfig                                `json:"api"`
	Telemetry                    telemetry.TelemetryConfig                `json:"telemetry"`
}

func (appConfig *AppConfig) SeparateConfigs() (
	*oracleCore.AppConfig, *ethOracleCore.AppConfig, *batcherCore.BatcherManagerConfiguration,
) {
	oracleCardanoChains := make(map[string]*oracleCore.CardanoChainConfig, len(appConfig.CardanoChains))
	batcherChains := make([]batcherCore.ChainConfig, 0, len(appConfig.CardanoChains))
	oracleEthChains := make(map[string]*ethOracleCore.EthChainConfig, len(appConfig.EthChains))

	for chainID, ccConfig := range appConfig.CardanoChains {
		oracleCardanoChains[chainID] = &oracleCore.CardanoChainConfig{
			ChainID:                  chainID,
			NetworkAddress:           ccConfig.NetworkAddress,
			NetworkMagic:             ccConfig.NetworkMagic,
			NetworkID:                ccConfig.NetworkID,
			StartBlockHash:           ccConfig.StartBlockHash,
			StartSlot:                ccConfig.StartSlot,
			StartBlockNumber:         ccConfig.StartBlockNumber,
			ConfirmationBlockCount:   ccConfig.ConfirmationBlockCount,
			OtherAddressesOfInterest: ccConfig.OtherAddressesOfInterest,
		}

		chainSpecificJSONRaw, _ := (cardanotx.CardanoChainConfig{
			TestNetMagic:          ccConfig.NetworkMagic,
			OgmiosURL:             ccConfig.OgmiosURL,
			BlockfrostURL:         ccConfig.BlockfrostURL,
			BlockfrostAPIKey:      ccConfig.BlockfrostAPIKey,
			SocketPath:            ccConfig.SocketPath,
			PotentialFee:          ccConfig.PotentialFee,
			TTLSlotNumberInc:      ccConfig.TTLSlotNumberInc,
			SlotRoundingThreshold: ccConfig.SlotRoundingThreshold,
		}).Serialize()

		batcherChains = append(batcherChains, batcherCore.ChainConfig{
			ChainID:       chainID,
			ChainType:     "Cardano",
			ChainSpecific: chainSpecificJSONRaw,
		})
	}

	for chainID := range appConfig.EthChains {
		oracleEthChains[chainID] = &ethOracleCore.EthChainConfig{
			ChainID: chainID,
		}
	}

	oracleConfig := &oracleCore.AppConfig{
		ValidatorDataDir:    appConfig.ValidatorDataDir,
		ValidatorConfigPath: appConfig.ValidatorConfigPath,
		Bridge:              appConfig.Bridge,
		Settings:            appConfig.Settings,
		BridgingSettings:    appConfig.BridgingSettings,
		CardanoChains:       oracleCardanoChains,
	}

	ethOracleConfig := &ethOracleCore.AppConfig{
		ValidatorDataDir:    appConfig.ValidatorDataDir,
		ValidatorConfigPath: appConfig.ValidatorConfigPath,
		EthChains:           oracleEthChains,
		Bridge:              appConfig.Bridge,
		Settings:            appConfig.Settings,
		BridgingSettings:    appConfig.BridgingSettings,
	}

	batcherConfig := &batcherCore.BatcherManagerConfiguration{
		ValidatorDataDir:    appConfig.ValidatorDataDir,
		ValidatorConfigPath: appConfig.ValidatorConfigPath,
		Bridge: batcherCore.BridgeConfig{
			NodeURL:              appConfig.Bridge.NodeURL,
			DynamicTx:            appConfig.Bridge.DynamicTx,
			SmartContractAddress: appConfig.Bridge.SmartContractAddress,
		},
		PullTimeMilis: appConfig.BatcherPullTimeMilis,
		Chains:        batcherChains,
	}

	return oracleConfig, ethOracleConfig, batcherConfig
}
