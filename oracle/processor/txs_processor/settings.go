package txsprocessor

import "github.com/Ethernal-Tech/apex-bridge/oracle/core"

const (
	TickTimeMs                  = 2000
	MinBridgingClaimsToGroup    = 1
	GasLimitMultiplierDefault   = float32(1)
	GasLimitMultiplierIncrement = float32(0.5)
	GasLimitMultiplierMax       = float32(3)
)

type txsProcessorSettings struct {
	appConfig                *core.AppConfig
	maxBridgingClaimsToGroup map[string]int
	gasLimitMultiplier       map[string]float32
}

func NewTxsProcessorSettings(appConfig *core.AppConfig) *txsProcessorSettings {
	defaultMaxClaimsToGroup := appConfig.BridgingSettings.MaxBridgingClaimsToGroup
	defaultGasLimitMultiplier := float32(1)

	maxBridgingClaimsToGroup := make(map[string]int, len(appConfig.CardanoChains)+len(appConfig.EthChains))
	for _, chain := range appConfig.CardanoChains {
		maxBridgingClaimsToGroup[chain.ChainID] = defaultMaxClaimsToGroup
	}

	for _, chain := range appConfig.EthChains {
		maxBridgingClaimsToGroup[chain.ChainID] = defaultMaxClaimsToGroup
	}

	gasLimitMultiplier := make(map[string]float32, len(appConfig.CardanoChains)+len(appConfig.EthChains))
	for _, chain := range appConfig.CardanoChains {
		gasLimitMultiplier[chain.ChainID] = defaultGasLimitMultiplier
	}

	for _, chain := range appConfig.EthChains {
		gasLimitMultiplier[chain.ChainID] = defaultGasLimitMultiplier
	}

	return &txsProcessorSettings{
		appConfig:                appConfig,
		maxBridgingClaimsToGroup: maxBridgingClaimsToGroup,
		gasLimitMultiplier:       gasLimitMultiplier,
	}
}

func (s *txsProcessorSettings) OnSubmitClaimsFailed(chainID string, claimsCount int) {
	s.maxBridgingClaimsToGroup[chainID] = claimsCount - 1
	if s.maxBridgingClaimsToGroup[chainID] < MinBridgingClaimsToGroup {
		s.maxBridgingClaimsToGroup[chainID] = MinBridgingClaimsToGroup
	}

	if claimsCount <= MinBridgingClaimsToGroup &&
		s.gasLimitMultiplier[chainID]+GasLimitMultiplierIncrement <= GasLimitMultiplierMax {
		s.gasLimitMultiplier[chainID] += GasLimitMultiplierIncrement
	}
}

func (s *txsProcessorSettings) ResetSubmitClaimsSettings(chainID string) {
	s.maxBridgingClaimsToGroup[chainID] = s.appConfig.BridgingSettings.MaxBridgingClaimsToGroup
	s.gasLimitMultiplier[chainID] = GasLimitMultiplierDefault
}
