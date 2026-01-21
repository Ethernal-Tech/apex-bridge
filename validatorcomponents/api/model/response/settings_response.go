package response

import (
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
)

type SettingsResponse struct {
	// For each chain, the minimum fee required to cover the submission of the currency transaction
	// on the destination chain
	MinChainFeeForBridging map[string]*big.Int `json:"minChainFeeForBridging"`
	// For each chain, the minimum fee required to cover the submission of the native token transaction
	// on the destination chain
	MinChainFeeForBridgingTokens map[string]uint64 `json:"minChainFeeForBridgingTokens"`
	// For each chain, the minimum fee required to cover operational costs
	MinOperationFee map[string]*big.Int `json:"minOperationFee"`
	// For each chain, the minimum allowed UTXO value
	MinUtxoChainValue map[string]uint64 `json:"minUtxoChainValue"`
	// For each chain, the direction config
	DirectionConfig map[string]common.DirectionConfig `json:"directionConfig"`
	// All defined tokens across the whole ecosystem
	EcosystemTokens []common.EcosystemToken `json:"ecosystemTokens"`
	// Minimum value allowed to be bridged
	MinValueToBridge uint64 `json:"minValueToBridge"`
	// Maximum amount of currency allowed to be bridged
	MaxAmountAllowedToBridge string `json:"maxAmountAllowedToBridge"`
	// Maximum amount of native tokens allowed to be bridged
	MaxTokenAmountAllowedToBridge string `json:"maxTokenAmountAllowedToBridge"`
	// Minimum amount of colored tokens allowed to be bridged for each chain
	MinColCoinsAllowedToBridge map[string]*big.Int `json:"minColCoinsAllowedToBridge"`
	// Maximum number of receivers allowed in a bridging request
	MaxReceiversPerBridgingRequest int `json:"maxReceiversPerBridgingRequest"`
} // @name SettingsResponse

func NewSettingsResponse(
	appConfig *core.AppConfig,
) *SettingsResponse {
	minUtxoMap := make(map[string]uint64)
	minFeeForBridgingMap := make(map[string]*big.Int)
	minFeeForBridgingTokensMap := make(map[string]uint64)
	minOperationFeeMap := make(map[string]*big.Int)
	minColCoinsAllowedToBridgeMap := make(map[string]*big.Int)

	var maxUtxoValue uint64 = 0

	for chainID, chainConfig := range appConfig.CardanoChains {
		minUtxoMap[chainID] = chainConfig.UtxoMinAmount
		minFeeForBridgingMap[chainID] = new(big.Int).SetUint64(chainConfig.DefaultMinFeeForBridging)
		minFeeForBridgingTokensMap[chainID] = chainConfig.MinFeeForBridgingTokens
		minOperationFeeMap[chainID] = new(big.Int).SetUint64(chainConfig.MinOperationFee)
		minColCoinsAllowedToBridgeMap[chainID] = new(big.Int).SetUint64(chainConfig.MinColCoinsAllowedToBridge)

		if chainConfig.UtxoMinAmount > maxUtxoValue {
			maxUtxoValue = chainConfig.UtxoMinAmount
		}
	}

	for chainID, ethConfig := range appConfig.EthChains {
		minFeeForBridgingMap[chainID] = ethConfig.MinFeeForBridging
		minOperationFeeMap[chainID] = ethConfig.MinOperationFee
		minColCoinsAllowedToBridgeMap[chainID] = ethConfig.MinColCoinsAllowedToBridge
	}

	return &SettingsResponse{
		MinChainFeeForBridging:         minFeeForBridgingMap,
		MinChainFeeForBridgingTokens:   minFeeForBridgingTokensMap,
		MinOperationFee:                minOperationFeeMap,
		MinUtxoChainValue:              minUtxoMap,
		DirectionConfig:                appConfig.DirectionConfig,
		EcosystemTokens:                appConfig.EcosystemTokens,
		MinValueToBridge:               maxUtxoValue,
		MaxAmountAllowedToBridge:       appConfig.BridgingSettings.MaxAmountAllowedToBridge.String(),
		MaxTokenAmountAllowedToBridge:  appConfig.BridgingSettings.MaxTokenAmountAllowedToBridge.String(),
		MinColCoinsAllowedToBridge:     minColCoinsAllowedToBridgeMap,
		MaxReceiversPerBridgingRequest: appConfig.BridgingSettings.MaxReceiversPerBridgingRequest,
	}
}
