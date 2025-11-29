package response

import (
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
)

type SettingsResponse struct {
	// For each chain, the minimum fee required to cover the submission of the currency transaction
	// on the destination chain
	MinChainFeeForBridging map[string]uint64 `json:"minChainFeeForBridging"`
	// For each chain, the minimum fee required to cover the submission of the native token transaction
	// on the destination chain
	MinChainFeeForBridgingTokens map[string]uint64 `json:"minChainFeeForBridgingTokens"`
	// For each chain, the minimum fee required to cover operational costs
	MinOperationFee map[string]uint64 `json:"minOperationFee"`
	// For each chain, the minimum allowed UTXO value
	MinUtxoChainValue map[string]uint64 `json:"minUtxoChainValue"`
	// For each chain, all allowed bridging directions
	AllowedDirections oracleCore.AllowedDirections `json:"allowedDirections"`
	// For each chain, all defined native tokens
	NativeTokens map[string][]sendtx.TokenExchangeConfig `json:"nativeTokens"`
	// Minimum value allowed to be bridged
	MinValueToBridge uint64 `json:"minValueToBridge"`
	// Maximum amount of currency allowed to be bridged
	MaxAmountAllowedToBridge string `json:"maxAmountAllowedToBridge"`
	// Maximum amount of native tokens allowed to be bridged
	MaxTokenAmountAllowedToBridge string `json:"maxTokenAmountAllowedToBridge"`
	// Maximum number of receivers allowed in a bridging request
	MaxReceiversPerBridgingRequest int `json:"maxReceiversPerBridgingRequest"`
	// List of colored coins allowed to be bridged
	ColoredCoins map[string]oracleCore.ColoredCoins `json:"coloredCoins"`
} // @name SettingsResponse

func NewSettingsResponse(
	appConfig *core.AppConfig,
) *SettingsResponse {
	minUtxoMap := make(map[string]uint64)
	minFeeForBridgingMap := make(map[string]uint64)
	minFeeForBridgingTokensMap := make(map[string]uint64)
	minOperationFeeMap := make(map[string]uint64)
	nativeTokensMap := make(map[string][]sendtx.TokenExchangeConfig)
	coloredCoins := make(map[string]oracleCore.ColoredCoins)

	var maxUtxoValue uint64 = 0

	for chainID, chainConfig := range appConfig.CardanoChains {
		minUtxoMap[chainID] = chainConfig.UtxoMinAmount
		minFeeForBridgingMap[chainID] = chainConfig.DefaultMinFeeForBridging
		minFeeForBridgingTokensMap[chainID] = chainConfig.MinFeeForBridgingTokens
		minOperationFeeMap[chainID] = chainConfig.MinOperationFee
		nativeTokensMap[chainID] = chainConfig.WrappedCurrencyTokens

		if chainConfig.UtxoMinAmount > maxUtxoValue {
			maxUtxoValue = chainConfig.UtxoMinAmount
		}

		for ccID, ccName := range chainConfig.ColoredCoins {
			if _, exists := coloredCoins[chainID]; !exists {
				coloredCoins[chainID] = make(oracleCore.ColoredCoins)
			}

			coloredCoins[chainID][ccID] = oracleCore.ColoredCoinEvm{
				TokenName: ccName,
			}
		}
	}

	for chainID, ethConfig := range appConfig.EthChains {
		minFeeForBridgingMap[chainID] = ethConfig.MinFeeForBridging
		coloredCoins[chainID] = ethConfig.ColoredCoins
	}

	return &SettingsResponse{
		MinChainFeeForBridging:         minFeeForBridgingMap,
		MinChainFeeForBridgingTokens:   minFeeForBridgingTokensMap,
		MinOperationFee:                minOperationFeeMap,
		MinUtxoChainValue:              minUtxoMap,
		AllowedDirections:              appConfig.BridgingSettings.AllowedDirections,
		NativeTokens:                   nativeTokensMap,
		MinValueToBridge:               maxUtxoValue,
		MaxAmountAllowedToBridge:       appConfig.BridgingSettings.MaxAmountAllowedToBridge.String(),
		MaxTokenAmountAllowedToBridge:  appConfig.BridgingSettings.MaxTokenAmountAllowedToBridge.String(),
		MaxReceiversPerBridgingRequest: appConfig.BridgingSettings.MaxReceiversPerBridgingRequest,
		ColoredCoins:                   coloredCoins,
	}
}
