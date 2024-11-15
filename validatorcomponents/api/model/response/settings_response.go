package response

import (
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
)

type SettingsResponse struct {
	MinFeeForBridging              uint64 `json:"minFeeForBridging"`
	MinUtxoValue                   uint64 `json:"minUtxoValue"`
	MaxAmountAllowedToBridge       string `json:"maxAmountAllowedToBridge"`
	MaxReceiversPerBridgingRequest int    `json:"maxReceiversPerBridgingRequest"`
}

func NewSettingsResponse(
	appConfig *core.AppConfig,
) *SettingsResponse {
	return &SettingsResponse{
		MinFeeForBridging:              appConfig.BridgingSettings.MinFeeForBridging,
		MinUtxoValue:                   appConfig.BridgingSettings.UtxoMinValue,
		MaxAmountAllowedToBridge:       appConfig.BridgingSettings.MaxAmountAllowedToBridge.String(),
		MaxReceiversPerBridgingRequest: appConfig.BridgingSettings.MaxReceiversPerBridgingRequest,
	}
}
