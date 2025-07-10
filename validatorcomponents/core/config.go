package core

import (
	apiCore "github.com/Ethernal-Tech/apex-bridge/api/core"
	batcherCore "github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
)

type AppConfig struct {
	RunMode                      common.VCRunMode                          `json:"runMode"`
	RefundEnabled                bool                                      `json:"refundEnabled"`
	ValidatorDataDir             string                                    `json:"validatorDataDir"`
	ValidatorConfigPath          string                                    `json:"validatorConfigPath"`
	CardanoChains                map[string]*oracleCore.CardanoChainConfig `json:"cardanoChains"`
	EthChains                    map[string]*oracleCore.EthChainConfig     `json:"ethChains"`
	Bridge                       oracleCore.BridgeConfig                   `json:"bridge"`
	BridgingSettings             oracleCore.BridgingSettings               `json:"bridgingSettings"`
	Settings                     oracleCore.AppSettings                    `json:"appSettings"`
	RelayerImitatorPullTimeMilis uint64                                    `json:"relayerImitatorPullTime"`
	BatcherPullTimeMilis         uint64                                    `json:"batcherPullTime"`
	APIConfig                    apiCore.APIConfig                         `json:"api"`
	Telemetry                    telemetry.TelemetryConfig                 `json:"telemetry"`
	RetryUnprocessedSettings     oracleCore.RetryUnprocessedSettings       `json:"retryUnprocessedSettings"`
	TryCountLimits               oracleCore.TryCountLimits                 `json:"tryCountLimits"`
}

func (appConfig *AppConfig) SeparateConfigs() (
	*oracleCore.AppConfig, *batcherCore.BatcherManagerConfiguration,
) {
	oracleCardanoChains := make(map[string]*oracleCore.CardanoChainConfig, len(appConfig.CardanoChains))
	batcherChains := make([]batcherCore.ChainConfig, 0, len(appConfig.CardanoChains)+len(appConfig.EthChains))
	oracleEthChains := make(map[string]*oracleCore.EthChainConfig, len(appConfig.EthChains))

	for _, ccConfig := range appConfig.CardanoChains {
		oracleCardanoChains[ccConfig.ChainID] = ccConfig

		chainSpecificJSONRaw, _ := ccConfig.CardanoChainConfig.Serialize()

		batcherChains = append(batcherChains, batcherCore.ChainConfig{
			ChainID:       ccConfig.ChainID,
			ChainType:     common.ChainTypeCardanoStr,
			ChainSpecific: chainSpecificJSONRaw,
		})
	}

	for _, ecConfig := range appConfig.EthChains {
		oracleEthChains[ecConfig.ChainID] = ecConfig

		chainSpecificJSONRaw, _ := (cardanotx.BatcherEVMChainConfig{
			TTLBlockNumberInc:      ecConfig.TTLBlockNumberInc,
			BlockRoundingThreshold: ecConfig.BlockRoundingThreshold,
			NoBatchPeriodPercent:   ecConfig.NoBatchPeriodPercent,
			MinFeeForBridging:      ecConfig.MinFeeForBridging,
			TestMode:               ecConfig.TestMode,
		}).Serialize()

		batcherChains = append(batcherChains, batcherCore.ChainConfig{
			ChainID:       ecConfig.ChainID,
			ChainType:     common.ChainTypeEVMStr,
			ChainSpecific: chainSpecificJSONRaw,
		})
	}

	oracleConfig := &oracleCore.AppConfig{
		RunMode:                  appConfig.RunMode,
		RefundEnabled:            appConfig.RefundEnabled,
		ValidatorDataDir:         appConfig.ValidatorDataDir,
		ValidatorConfigPath:      appConfig.ValidatorConfigPath,
		Bridge:                   appConfig.Bridge,
		Settings:                 appConfig.Settings,
		BridgingSettings:         appConfig.BridgingSettings,
		RetryUnprocessedSettings: appConfig.RetryUnprocessedSettings,
		TryCountLimits:           appConfig.TryCountLimits,
		CardanoChains:            oracleCardanoChains,
		EthChains:                oracleEthChains,
	}

	batcherConfig := &batcherCore.BatcherManagerConfiguration{
		PullTimeMilis: appConfig.BatcherPullTimeMilis,
		Chains:        batcherChains,
	}

	return oracleConfig, batcherConfig
}
