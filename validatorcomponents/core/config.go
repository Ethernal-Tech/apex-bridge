package core

import (
	"fmt"

	apiCore "github.com/Ethernal-Tech/apex-bridge/api/core"
	batcherCore "github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	goEthCommon "github.com/ethereum/go-ethereum/common"
)

type AppConfig struct {
	RunMode                      common.VCRunMode                          `json:"runMode"`
	RefundEnabled                bool                                      `json:"refundEnabled"`
	ValidatorDataDir             string                                    `json:"validatorDataDir"`
	ValidatorConfigPath          string                                    `json:"validatorConfigPath"`
	ChainIDConverter             *common.ChainIDConverter                  `json:"-"`
	CardanoChains                map[string]*oracleCore.CardanoChainConfig `json:"cardanoChains"`
	EthChains                    map[string]*oracleCore.EthChainConfig     `json:"ethChains"`
	DirectionConfig              map[string]common.DirectionConfig         `json:"directionConfig"`
	Bridge                       oracleCore.BridgeConfig                   `json:"bridge"`
	BridgingSettings             oracleCore.BridgingSettings               `json:"bridgingSettings"`
	Settings                     oracleCore.AppSettings                    `json:"appSettings"`
	RelayerImitatorPullTimeMilis uint64                                    `json:"relayerImitatorPullTime"`
	BatcherPullTimeMilis         uint64                                    `json:"batcherPullTime"`
	APIConfig                    apiCore.APIConfig                         `json:"api"`
	Telemetry                    telemetry.TelemetryConfig                 `json:"telemetry"`
	RetryUnprocessedSettings     oracleCore.RetryUnprocessedSettings       `json:"retryUnprocessedSettings"`
	TryCountLimits               oracleCore.TryCountLimits                 `json:"tryCountLimits"`
	EcosystemTokens              []common.EcosystemToken                   `json:"ecosystemTokens"`
}

func (appConfig *AppConfig) SetupChainIDs(chainIDsConfig *common.ChainIDsConfigFile) {
	appConfig.ChainIDConverter = chainIDsConfig.ToChainIDConverter()
}

func (appConfig *AppConfig) SetupDirectionConfig(directionConfig *common.DirectionConfigFile) error {
	appConfig.DirectionConfig = directionConfig.Directions
	appConfig.EcosystemTokens = directionConfig.EcosystemTokens

	for chainID, directionConfig := range directionConfig.Directions {
		if appConfig.ChainIDConverter.IsEVMChainID(chainID) {
			if _, ok := appConfig.EthChains[chainID]; !ok {
				return fmt.Errorf("invalid eth chain while setting up direction config. %s", chainID)
			}

			data := appConfig.EthChains[chainID]
			data.AlwaysTrackCurrencyAndWrappedCurrency = directionConfig.AlwaysTrackCurrencyAndWrappedCurrency
			data.DestinationChains = directionConfig.DestinationChains
			data.Tokens = directionConfig.Tokens
			appConfig.EthChains[chainID] = data
		} else {
			if _, ok := appConfig.CardanoChains[chainID]; !ok {
				return fmt.Errorf("invalid cardano chain while setting up direction config. %s", chainID)
			}

			data := appConfig.CardanoChains[chainID]
			data.CardanoChainConfig.DestinationChains = directionConfig.DestinationChains
			data.CardanoChainConfig.Tokens = directionConfig.Tokens
			data.CardanoChainConfig.AlwaysTrackCurrencyAndWrappedCurrency = directionConfig.AlwaysTrackCurrencyAndWrappedCurrency
			appConfig.CardanoChains[chainID] = data
		}
	}

	return nil
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
			DestinationChains:      ecConfig.DestinationChains,
			Tokens:                 ecConfig.Tokens,
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
		ChainIDConverter:         appConfig.ChainIDConverter,
	}

	batcherConfig := &batcherCore.BatcherManagerConfiguration{
		PullTimeMilis:    appConfig.BatcherPullTimeMilis,
		Chains:           batcherChains,
		ChainIDConverter: appConfig.ChainIDConverter,
	}

	return oracleConfig, batcherConfig
}

func (appConfig *AppConfig) ValidateDirectionConfig() error {
	if len(appConfig.EcosystemTokens) == 0 {
		return fmt.Errorf("no ecosystem tokens")
	}

	ecosystemTokensMap := make(map[uint16]string, len(appConfig.EcosystemTokens))

	for _, tok := range appConfig.EcosystemTokens {
		if tok.ID == 0 {
			return fmt.Errorf("found ecosystem token with id zero")
		}

		ecosystemTokensMap[tok.ID] = tok.Name
	}

	allChains := make([]string, 0, len(appConfig.CardanoChains)+len(appConfig.EthChains))
	for _, cc := range appConfig.CardanoChains {
		allChains = append(allChains, cc.ChainID)
	}

	for _, ec := range appConfig.EthChains {
		allChains = append(allChains, ec.ChainID)
	}

	for _, chainID := range allChains {
		dirConfig, ok := appConfig.DirectionConfig[chainID]
		if !ok {
			return fmt.Errorf("direction config not found for chain: %s", chainID)
		}

		if len(dirConfig.Tokens) == 0 {
			return fmt.Errorf("direction config for chain: %s, has no tokens defined", chainID)
		}

		var foundCurrency bool

		for tokID, tok := range dirConfig.Tokens {
			if tok.ChainSpecific == wallet.AdaTokenName {
				foundCurrency = true
			}

			if _, ok := ecosystemTokensMap[tokID]; !ok {
				return fmt.Errorf("tokenID: %v for chain %s not found in ecosystem tokens", tokID, chainID)
			}
		}

		if !foundCurrency {
			return fmt.Errorf("currency token not found in direction config for chain: %s", chainID)
		}
	}

	for _, cc := range appConfig.CardanoChains {
		dirConfig := appConfig.DirectionConfig[cc.ChainID]

		for _, tok := range dirConfig.Tokens {
			if tok.ChainSpecific != wallet.AdaTokenName {
				if _, err := wallet.NewTokenWithFullNameTry(tok.ChainSpecific); err != nil {
					return fmt.Errorf("invalid cardano token %s in direction config for chain: %s",
						tok.ChainSpecific, cc.ChainID)
				}
			}
		}
	}

	for _, ec := range appConfig.EthChains {
		dirConfig := appConfig.DirectionConfig[ec.ChainID]

		for _, tok := range dirConfig.Tokens {
			if tok.ChainSpecific != wallet.AdaTokenName {
				if len(tok.ChainSpecific) == 0 || !goEthCommon.IsHexAddress(tok.ChainSpecific) {
					return fmt.Errorf("invalid eth token contract addr %s in direction config for chain: %s",
						tok.ChainSpecific, ec.ChainID)
				}
			}
		}
	}

	for _, srcChainID := range allChains {
		srcDirConfig := appConfig.DirectionConfig[srcChainID]

		for dstChainID, tokenPairs := range srcDirConfig.DestinationChains {
			dstDirConfig, ok := appConfig.DirectionConfig[dstChainID]
			if !ok {
				return fmt.Errorf("direction config not found for chain: %s", dstChainID)
			}

			for _, tokenPair := range tokenPairs {
				if _, ok := ecosystemTokensMap[tokenPair.SourceTokenID]; !ok {
					return fmt.Errorf("tokenPair tokenID: %v not found in ecosystem tokens",
						tokenPair.SourceTokenID)
				}

				if _, ok := ecosystemTokensMap[tokenPair.DestinationTokenID]; !ok {
					return fmt.Errorf("tokenPair tokenID: %v not found in ecosystem tokens",
						tokenPair.DestinationTokenID)
				}

				if _, ok := srcDirConfig.Tokens[tokenPair.SourceTokenID]; !ok {
					return fmt.Errorf(
						"tokenPair tokenID: %v not found in direction config tokens for chain: %s",
						tokenPair.SourceTokenID, srcChainID)
				}

				if _, ok := dstDirConfig.Tokens[tokenPair.DestinationTokenID]; !ok {
					return fmt.Errorf(
						"tokenPair tokenID: %v not found in direction config tokens for chain: %s",
						tokenPair.DestinationTokenID, dstChainID)
				}
			}
		}
	}

	return nil
}
