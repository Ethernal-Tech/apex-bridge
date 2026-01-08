package utils

import (
	"fmt"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
)

func GetChainConfig(appConfig *core.AppConfig, chainID string) (*core.CardanoChainConfig, *core.EthChainConfig) {
	if cardanoChainConfig, exists := appConfig.CardanoChains[chainID]; exists {
		return cardanoChainConfig, nil
	}

	if ethChainConfig, exists := appConfig.EthChains[chainID]; exists {
		return nil, ethChainConfig
	}

	return nil, nil
}

func GetTxPriority(txProcessorType common.BridgingTxType) uint8 {
	if txProcessorType == common.BridgingTxTypeBatchExecution || txProcessorType == common.TxTypeHotWalletFund {
		return 0
	}

	return 1
}

func UpdateTxReceivedTelemetry[T core.IIsInvalid](originChainID string, processedTxs []T, countRelevantTx int) {
	telemetry.UpdateOracleTxsReceivedCounter(originChainID, len(processedTxs)+countRelevantTx)

	invalidCnt := 0

	for _, x := range processedTxs {
		if x.GetIsInvalid() {
			invalidCnt++
		}
	}

	if invalidCnt > 0 {
		telemetry.UpdateOracleClaimsInvalidMetaDataCounter(originChainID, invalidCnt)
	}
}

func GetTokenPair(
	destinationChains map[string]common.TokenPairs,
	srcChainID, destChainID string,
	tokenID uint16,
) (*common.TokenPair, error) {
	tokenPairs, pathExists := destinationChains[destChainID]
	if !pathExists {
		return nil, fmt.Errorf("no bridging path from source chain %s to destination chain %s",
			srcChainID, destChainID)
	}

	for _, tokenPair := range tokenPairs {
		if tokenPair.SourceTokenID == tokenID {
			return &tokenPair, nil
		}
	}

	return nil, fmt.Errorf("no bridging path from source chain %s to destination chain %s with token ID %d",
		srcChainID, destChainID, tokenID)
}

type DestChainInfo struct {
	FeeAddress         string
	FeeAddrBridgingAmt uint64
	CurrencyTokenID    uint16
}

func GetDestChainInfo(
	destChainID string,
	appConfig *core.AppConfig,
	cardanoDestConfig *core.CardanoChainConfig,
	ethDestConfig *core.EthChainConfig,
) (*DestChainInfo, error) {
	switch {
	case cardanoDestConfig != nil:
		currencyDestID, err := cardanoDestConfig.GetCurrencyID()
		if err != nil {
			return nil, fmt.Errorf("failed to get currency ID for destination chain %s: %w", destChainID, err)
		}

		return &DestChainInfo{
			FeeAddress:         appConfig.GetFeeMultisigAddress(destChainID),
			FeeAddrBridgingAmt: cardanoDestConfig.FeeAddrBridgingAmount,
			CurrencyTokenID:    currencyDestID,
		}, nil
	case ethDestConfig != nil:
		currencyDestID, err := ethDestConfig.GetCurrencyID()
		if err != nil {
			return nil, fmt.Errorf("failed to get currency ID for destination chain %s: %w", destChainID, err)
		}

		return &DestChainInfo{
			FeeAddress:         common.EthZeroAddr,
			FeeAddrBridgingAmt: ethDestConfig.FeeAddrBridgingAmount,
			CurrencyTokenID:    currencyDestID,
		}, nil
	default:
		return nil, fmt.Errorf("destination chain not registered: %s", destChainID)
	}
}

func NormalizeAddr(addr string) string {
	addr = strings.ToLower(addr)

	return strings.TrimPrefix(addr, "0x")
}
