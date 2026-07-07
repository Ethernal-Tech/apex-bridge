package utils

import (
	"fmt"
	"math/big"
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

type ChainConfigResult struct {
	Cardano *core.CardanoChainConfig
	Eth     *core.EthChainConfig
	Solana  *core.SolanaChainConfig
}

func GetChainConfigResult(appConfig *core.AppConfig, chainID string) ChainConfigResult {
	if cfg, exists := appConfig.CardanoChains[chainID]; exists {
		return ChainConfigResult{Cardano: cfg}
	}

	if cfg, exists := appConfig.EthChains[chainID]; exists {
		return ChainConfigResult{Eth: cfg}
	}

	if cfg, exists := appConfig.SolanaChains[chainID]; exists {
		return ChainConfigResult{Solana: cfg}
	}

	return ChainConfigResult{}
}

func (ccr *ChainConfigResult) IsNone() bool {
	return ccr.Cardano == nil && ccr.Eth == nil && ccr.Solana == nil
}

func (ccr *ChainConfigResult) GetChainType() string {
	if ccr.Cardano != nil {
		return common.ChainTypeCardanoStr
	}

	if ccr.Eth != nil {
		return common.ChainTypeEVMStr
	}

	if ccr.Solana != nil {
		return common.ChainTypeSolanaStr
	}

	return ""
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
	FeeAddress                 string
	FeeAddrBridgingWei         *big.Int
	CurrencyTokenID            uint16
	MinColCoinsAllowedToBridge *big.Int
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
			FeeAddress:                 appConfig.GetFeeMultisigAddress(destChainID),
			FeeAddrBridgingWei:         common.DfmToWei(new(big.Int).SetUint64(cardanoDestConfig.FeeAddrBridgingAmount)),
			CurrencyTokenID:            currencyDestID,
			MinColCoinsAllowedToBridge: common.DfmToWei(new(big.Int).SetUint64(cardanoDestConfig.MinColCoinsAllowedToBridge)),
		}, nil
	case ethDestConfig != nil:
		currencyDestID, err := ethDestConfig.GetCurrencyID()
		if err != nil {
			return nil, fmt.Errorf("failed to get currency ID for destination chain %s: %w", destChainID, err)
		}

		return &DestChainInfo{
			FeeAddress:                 common.EthZeroAddr,
			FeeAddrBridgingWei:         ethDestConfig.FeeAddrBridgingAmount,
			CurrencyTokenID:            currencyDestID,
			MinColCoinsAllowedToBridge: ethDestConfig.MinColCoinsAllowedToBridge,
		}, nil
	default:
		return nil, fmt.Errorf("destination chain not registered ovde: %s", destChainID)
	}
}

func GetDestChainInfoResult(
	destChainID string,
	destChainConfig ChainConfigResult,
	appConfig *core.AppConfig,
) (*DestChainInfo, error) {
	if destChainConfig.IsNone() {
		return nil, fmt.Errorf("destination chain not registered")
	}

	switch {
	case destChainConfig.Cardano != nil:
		return GetDestChainInfo(destChainConfig.Cardano.ChainID, appConfig, destChainConfig.Cardano, nil)
	case destChainConfig.Eth != nil:
		return GetDestChainInfo(destChainConfig.Eth.ChainID, appConfig, nil, destChainConfig.Eth)
	case destChainConfig.Solana != nil:
		currencyDestID, err := destChainConfig.Solana.GetCurrencyID()
		if err != nil {
			return nil, fmt.Errorf("failed to get currency ID for destination chain %s: %w", destChainID, err)
		}

		return &DestChainInfo{
			FeeAddress: common.EthZeroAddr,
			FeeAddrBridgingWei: common.LamportToWei(
				new(big.Int).SetUint64(destChainConfig.Solana.FeeAddrBridgingAmount)),
			CurrencyTokenID: currencyDestID,
			MinColCoinsAllowedToBridge: common.LamportToWei(
				new(big.Int).SetUint64(destChainConfig.Solana.MinColCoinsAllowedToBridge)),
		}, nil
	default:
		return nil, fmt.Errorf("unknown destination chain config")
	}
}

func NormalizeAddr(addr string) string {
	addr = strings.ToLower(addr)

	return strings.TrimPrefix(addr, "0x")
}

func MaxBigInt(a, b *big.Int) *big.Int {
	if a.Cmp(b) >= 0 { // a >= b
		return a
	}

	return b
}
