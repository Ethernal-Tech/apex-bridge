package response

import (
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
)

type SettingsResponse struct {
	MinChainFeeForBridging         map[string]uint64   `json:"minChainFeeForBridging"`
	MinUtxoChainValue              map[string]uint64   `json:"minUtxoChainValue"`
	MinValueToBridge               uint64              `json:"minValueToBridge"`
	MaxAmountAllowedToBridge       string              `json:"maxAmountAllowedToBridge"`
	MaxReceiversPerBridgingRequest int                 `json:"maxReceiversPerBridgingRequest"`
	AllowedDirections              map[string][]string `json:"allowedDirections"`
}

func NewSettingsResponse(
	appConfig *core.AppConfig,
) *SettingsResponse {
	minUtxoMap := make(map[string]uint64)
	minFeeForBridgingMap := make(map[string]uint64)

	var maxUtxoValue uint64 = 0

	for chainID, chainConfig := range appConfig.CardanoChains {
		minUtxoMap[chainID] = chainConfig.UtxoMinAmount
		minFeeForBridgingMap[chainID] = chainConfig.MinFeeForBridging

		if chainConfig.UtxoMinAmount > maxUtxoValue {
			maxUtxoValue = chainConfig.UtxoMinAmount
		}
	}

	for chainID, ethConfig := range appConfig.EthChains {
		minFeeForBridgingMap[chainID] = ethConfig.MinFeeForBridging
	}

	return &SettingsResponse{
		MinChainFeeForBridging:         minFeeForBridgingMap,
		MinUtxoChainValue:              minUtxoMap,
		MinValueToBridge:               maxUtxoValue,
		MaxAmountAllowedToBridge:       appConfig.BridgingSettings.MaxAmountAllowedToBridge.String(),
		MaxReceiversPerBridgingRequest: appConfig.BridgingSettings.MaxReceiversPerBridgingRequest,
		AllowedDirections:              appConfig.BridgingSettings.AllowedDirections,
	}
}

type ValidatorChangeStatusReponse struct {
	InProgress bool `json:"inProgress"`
}

func NewValidatorChangeStatusResponse(
	inProgress bool,
) *ValidatorChangeStatusReponse {
	return &ValidatorChangeStatusReponse{
		InProgress: inProgress,
	}
}

type MultiSigAddressesResponse struct {
	CardanoChains map[string]*oracleCore.BridgingAddresses `json:"bridgingAddress"`
}

func NewMultiSigAddressesResponse(appConfig *core.AppConfig) *MultiSigAddressesResponse {
	cardanoChainsMap := make(map[string]*oracleCore.BridgingAddresses)

	for chainID, chainAppConfig := range appConfig.CardanoChains {
		cardanoChainsMap[chainID] = &chainAppConfig.BridgingAddresses
	}

	return &MultiSigAddressesResponse{
		CardanoChains: cardanoChainsMap,
	}
}
