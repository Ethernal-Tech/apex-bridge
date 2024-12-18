package utils

import (
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
