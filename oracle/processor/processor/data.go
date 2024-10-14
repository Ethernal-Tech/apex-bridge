package processor

import "github.com/Ethernal-Tech/apex-bridge/oracle/core"

const (
	TickTimeMs                  = 2000
	TTLInsuranceOffset          = 2
	MinBridgingClaimsToGroup    = 1
	GasLimitMultiplierDefault   = float32(1)
	GasLimitMultiplierIncrement = float32(0.5)
	GasLimitMultiplierMax       = float32(3)
)

type cardanoTxsState struct {
	invalidRelevantExpired []*core.BridgeExpectedCardanoTx
	processedExpected      []*core.BridgeExpectedCardanoTx
	processed              []*core.ProcessedCardanoTx
	unprocessed            []*core.CardanoTx
}

func (s *cardanoTxsState) addToInvalidRelevantExpired(newTxs []*core.BridgeExpectedCardanoTx) {
	s.invalidRelevantExpired = append(s.invalidRelevantExpired, newTxs...)
}

func (s *cardanoTxsState) addToProcessedExpected(newTxs []*core.BridgeExpectedCardanoTx) {
	s.processedExpected = append(s.processedExpected, newTxs...)
}

func (s *cardanoTxsState) addToProcessed(newTxs []*core.ProcessedCardanoTx) {
	s.processed = append(s.processed, newTxs...)
}

func (s *cardanoTxsState) addToUnprocessed(newTxs []*core.CardanoTx) {
	s.unprocessed = append(s.unprocessed, newTxs...)
}

type txsProcessorSettings struct {
	appConfig                *core.AppConfig
	maxBridgingClaimsToGroup map[string]int
	gasLimitMultiplier       map[string]float32
}

func NewTxsProcessorSettings(appConfig *core.AppConfig) *txsProcessorSettings {
	maxBridgingClaimsToGroup := make(map[string]int, len(appConfig.CardanoChains))
	for _, chain := range appConfig.CardanoChains {
		maxBridgingClaimsToGroup[chain.ChainID] = appConfig.BridgingSettings.MaxBridgingClaimsToGroup
	}

	gasLimitMultiplier := make(map[string]float32, len(appConfig.CardanoChains))
	for _, chain := range appConfig.CardanoChains {
		gasLimitMultiplier[chain.ChainID] = 1
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
