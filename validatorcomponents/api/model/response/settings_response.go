package response

import (
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
)

type SettingsResponse struct {
	// For each chain, the minimum fee required to cover the submission of the transaction on the destination chain
	MinChainFeeForBridging map[string]uint64 `json:"minChainFeeForBridging"`
	// For each chain, the minimum fee required to cover operational costs
	MinOperationFee map[string]uint64 `json:"minOperationFee"`
	// For each chain, the minimum allowed UTXO value
	MinUtxoChainValue map[string]uint64 `json:"minUtxoChainValue"`
	// Minimum value allowed to be bridged
	MinValueToBridge uint64 `json:"minValueToBridge"`
	// Maximum amount of currency allowed to be bridged
	MaxAmountAllowedToBridge string `json:"maxAmountAllowedToBridge"`
	// Maximum amount of native tokens allowed to be bridged
	MaxTokenAmountAllowedToBridge string `json:"maxTokenAmountAllowedToBridge"`
	// Maximum number of receivers allowed in a bridging request
	MaxReceiversPerBridgingRequest int `json:"maxReceiversPerBridgingRequest"`
} //@name SettingsResponse

func NewSettingsResponse(
	appConfig *core.AppConfig,
) *SettingsResponse {
	minUtxoMap := make(map[string]uint64)
	minFeeForBridgingMap := make(map[string]uint64)
	minOperationFeeMap := make(map[string]uint64)

	var maxUtxoValue uint64 = 0

	for chainID, chainConfig := range appConfig.CardanoChains {
		minUtxoMap[chainID] = chainConfig.UtxoMinAmount
		minFeeForBridgingMap[chainID] = chainConfig.MinFeeForBridging
		minOperationFeeMap[chainID] = chainConfig.MinOperationFee

		if chainConfig.UtxoMinAmount > maxUtxoValue {
			maxUtxoValue = chainConfig.UtxoMinAmount
		}
	}

	for chainID, ethConfig := range appConfig.EthChains {
		minFeeForBridgingMap[chainID] = ethConfig.MinFeeForBridging
	}

	return &SettingsResponse{
		MinChainFeeForBridging:         minFeeForBridgingMap,
		MinOperationFee:                minOperationFeeMap,
		MinUtxoChainValue:              minUtxoMap,
		MinValueToBridge:               maxUtxoValue,
		MaxAmountAllowedToBridge:       appConfig.BridgingSettings.MaxAmountAllowedToBridge.String(),
		MaxTokenAmountAllowedToBridge:  appConfig.BridgingSettings.MaxTokenAmountAllowedToBridge.String(),
		MaxReceiversPerBridgingRequest: appConfig.BridgingSettings.MaxReceiversPerBridgingRequest,
	}
}
