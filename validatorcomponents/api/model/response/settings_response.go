package response

import (
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
)

type SettingsResponse struct {
	MinFeeForBridging              uint64            `json:"minFeeForBridging"`
	MinUtxoChainValue              map[string]uint64 `json:"minUtxoMap"`
	MinValueToBridge               uint64            `json:"minValueToBridge"`
	MaxAmountAllowedToBridge       string            `json:"maxAmountAllowedToBridge"`
	MaxReceiversPerBridgingRequest int               `json:"maxReceiversPerBridgingRequest"`
}

func NewSettingsResponse(
	appConfig *core.AppConfig,
) *SettingsResponse {
	minUtxoMap := make(map[string]uint64)
	var maxUtxoValue uint64 = 0
	for chainID, chainConfig := range appConfig.CardanoChains {
		minUtxoMap[chainID] = chainConfig.UtxoMinAmount
		if chainConfig.UtxoMinAmount > maxUtxoValue {
			maxUtxoValue = chainConfig.UtxoMinAmount
		}
	}
	return &SettingsResponse{
		MinFeeForBridging:              appConfig.BridgingSettings.MinFeeForBridging,
		MinUtxoChainValue:              minUtxoMap,
		MinValueToBridge:               maxUtxoValue,
		MaxAmountAllowedToBridge:       appConfig.BridgingSettings.MaxAmountAllowedToBridge.String(),
		MaxReceiversPerBridgingRequest: appConfig.BridgingSettings.MaxReceiversPerBridgingRequest,
	}
}
