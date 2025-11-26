package utils

import (
	"fmt"

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

	return nil, fmt.Errorf("no token pair found for source token ID %d", tokenID)
}
