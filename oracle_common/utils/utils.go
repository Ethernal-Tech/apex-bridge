package utils

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oChain "github.com/Ethernal-Tech/apex-bridge/oracle_common/chain"
	"github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	solcore "github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
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

func (ccr *ChainConfigResult) ProcessReceiver(
	destCfg *ChainConfigResult,
	receiver *solcore.BridgingRequestSolMetadataTransaction,
	totalTokensAmount *core.TotalTokensAmount,
	cardanoChainInfos map[string]*oChain.CardanoChainInfo,
) (*core.BridgingRequestReceiver, error) {
	switch {
	case destCfg.Cardano != nil:
		return processReceiverCardano(ccr, destCfg.Cardano, receiver, totalTokensAmount, cardanoChainInfos)
	case destCfg.Eth != nil:
		return processReceiverEth(ccr, destCfg.Eth, receiver, totalTokensAmount)
	case destCfg.Solana != nil:
		// return processReceiverSolana(srcCfg, destCfg.Solana, receiver, totalTokensAmount)
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown destination chain config")
	}
}

func processReceiverCardano(
	srcCfg *ChainConfigResult,
	cardanoDestCfg *core.CardanoChainConfig,
	receiver *solcore.BridgingRequestSolMetadataTransaction,
	totalTokensAmount *core.TotalTokensAmount,
	cardanoChainInfos map[string]*oChain.CardanoChainInfo,
) (*core.BridgingRequestReceiver, error) {
	srcDestinationChains, srcChainID, srcTokens, alwaysTrackCurrencyAndWrappedCurrency, srcCurrencyID, err :=
		getSourceChainCommonFields(srcCfg)
	if err != nil {
		return nil, err
	}

	destCurrencyID, err := cardanoDestCfg.GetCurrencyID()
	if err != nil {
		return nil, fmt.Errorf("failed to get currency ID for destination chain %s: %w", cardanoDestCfg.ChainID, err)
	}

	tokenPair, err := GetTokenPair(
		srcDestinationChains,
		srcChainID,
		cardanoDestCfg.ChainID,
		receiver.TokenID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get token pair for source chain %s to destination chain %s: %w",
			srcChainID, cardanoDestCfg.ChainID, err,
		)
	}

	var (
		amount        *big.Int
		amountWrapped = big.NewInt(0)
	)

	if tokenPair.DestinationTokenID == destCurrencyID {
		amount = receiver.Amount

		if cardanoDestCfg.AlwaysTrackCurrencyAndWrappedCurrency || tokenPair.TrackDestinationToken {
			totalTokensAmount.TrackDestTokenAmount(
				receiver.Amount,
				big.NewInt(0),
			)
		}
	} else {
		nativeTokensSum := map[uint16]*big.Int{
			tokenPair.DestinationTokenID: common.WeiToDfm(receiver.Amount),
		}

		dstMinUtxo, err := calculateMinUtxo(cardanoDestCfg, receiver.Address, nativeTokensSum, cardanoChainInfos)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate destination minUtxo for chainID: %s. err: %w",
				cardanoDestCfg.ChainID, err)
		}

		amount = common.DfmToWei(new(big.Int).SetUint64(dstMinUtxo))
		totalTokensAmount.TrackDestTokenAmount(amount, big.NewInt(0))

		amountWrapped = receiver.Amount

		// wrapped token on destination
		if (cardanoDestCfg.AlwaysTrackCurrencyAndWrappedCurrency || tokenPair.TrackDestinationToken) &&
			cardanoDestCfg.Tokens[tokenPair.DestinationTokenID].IsWrappedCurrency {
			totalTokensAmount.TrackDestTokenAmount(
				big.NewInt(0),
				receiver.Amount,
			)
		}
	}

	if alwaysTrackCurrencyAndWrappedCurrency || tokenPair.TrackSourceToken {
		totalTokensAmount.TrackSourceTokenAmount(
			tokenPair.SourceTokenID,
			srcCurrencyID,
			receiver.Amount,
			srcTokens,
		)
	}

	return &core.BridgingRequestReceiver{
		DestinationAddress: receiver.Address,
		Amount:             amount,
		AmountWrapped:      amountWrapped,
		TokenId:            tokenPair.DestinationTokenID,
	}, nil
}

func getSourceChainCommonFields(
	srcCfg *ChainConfigResult,
) (
	srcDestinationChains map[string]common.TokenPairs,
	srcChainID string,
	srcTokens map[uint16]common.Token,
	alwaysTrackCurrencyAndWrappedCurrency bool,
	srcCurrencyID uint16,
	err error,
) {
	switch {
	case srcCfg.Cardano != nil:
		srcDestinationChains = srcCfg.Cardano.DestinationChains
		srcChainID = srcCfg.Cardano.ChainID
		srcTokens = srcCfg.Cardano.Tokens
		alwaysTrackCurrencyAndWrappedCurrency = srcCfg.Cardano.AlwaysTrackCurrencyAndWrappedCurrency

		srcCurrencyID, err = srcCfg.Cardano.GetCurrencyID()
		if err != nil {
			return nil, "", nil, false, 0,
				fmt.Errorf("failed to get currency ID for source chain %s: %w", srcCfg.Cardano.ChainID, err)
		}
	case srcCfg.Eth != nil:
		srcDestinationChains = srcCfg.Eth.DestinationChains
		srcChainID = srcCfg.Eth.ChainID
		srcTokens = srcCfg.Eth.Tokens
		alwaysTrackCurrencyAndWrappedCurrency = srcCfg.Eth.AlwaysTrackCurrencyAndWrappedCurrency

		srcCurrencyID, err = srcCfg.Eth.GetCurrencyID()
		if err != nil {
			return nil, "", nil, false, 0,
				fmt.Errorf("failed to get currency ID for source chain %s: %w", srcCfg.Eth.ChainID, err)
		}
	case srcCfg.Solana != nil:
		srcDestinationChains = srcCfg.Solana.DestinationChains
		srcChainID = srcCfg.Solana.ChainID
		srcTokens = srcCfg.Solana.Tokens
		// alwaysTrackCurrencyAndWrappedCurrency = srcCfg.Solana.AlwaysTrackCurrencyAndWrappedCurrency

		srcCurrencyID, err = srcCfg.Solana.GetCurrencyID()
		if err != nil {
			return nil, "", nil, false, 0,
				fmt.Errorf("failed to get currency ID for source chain %s: %w", srcCfg.Solana.ChainID, err)
		}
	default:
		return nil, "", nil, false, 0, fmt.Errorf("unknown source chain config")
	}

	return srcDestinationChains, srcChainID, srcTokens, alwaysTrackCurrencyAndWrappedCurrency, srcCurrencyID, nil
}

func processReceiverEth(
	srcCfg *ChainConfigResult,
	ethDestConfig *core.EthChainConfig,
	receiver *solcore.BridgingRequestSolMetadataTransaction,
	totalTokensAmount *core.TotalTokensAmount,
) (*core.BridgingRequestReceiver, error) {
	srcDestinationChains, srcChainID, srcTokens, alwaysTrackCurrencyAndWrappedCurrency, srcCurrencyID, err :=
		getSourceChainCommonFields(srcCfg)
	if err != nil {
		return nil, err
	}

	tokenPair, err := GetTokenPair(
		srcDestinationChains, srcChainID,
		ethDestConfig.ChainID,
		receiver.TokenID)
	if err != nil {
		return nil, err
	}

	amount := big.NewInt(0)
	amountWrapped := big.NewInt(0)

	destCurrencyID, err := ethDestConfig.GetCurrencyID()
	if err != nil {
		return nil, fmt.Errorf("failed to get currency ID for destination chain %s: %w", ethDestConfig.ChainID, err)
	}

	// currency on destination
	if tokenPair.DestinationTokenID == destCurrencyID {
		amount = receiver.Amount

		if ethDestConfig.AlwaysTrackCurrencyAndWrappedCurrency || tokenPair.TrackDestinationToken {
			totalTokensAmount.TrackDestTokenAmount(
				receiver.Amount, big.NewInt(0),
			)
		}
	} else {
		amountWrapped = receiver.Amount

		// wrapped token on destination
		if (ethDestConfig.AlwaysTrackCurrencyAndWrappedCurrency || tokenPair.TrackDestinationToken) &&
			ethDestConfig.Tokens[tokenPair.DestinationTokenID].IsWrappedCurrency {
			totalTokensAmount.TrackDestTokenAmount(
				big.NewInt(0), receiver.Amount,
			)
		}
	}

	if alwaysTrackCurrencyAndWrappedCurrency || tokenPair.TrackSourceToken {
		totalTokensAmount.TrackSourceTokenAmount(
			tokenPair.SourceTokenID, srcCurrencyID, receiver.Amount, srcTokens,
		)
	}

	return &core.BridgingRequestReceiver{
		DestinationAddress: receiver.Address,
		Amount:             amount,
		AmountWrapped:      amountWrapped,
		TokenId:            tokenPair.DestinationTokenID,
	}, nil
}

func calculateMinUtxo(
	config *core.CardanoChainConfig,
	receiverAddr string,
	nativeTokensSum map[uint16]*big.Int,
	cardanoChainInfos map[string]*oChain.CardanoChainInfo,
) (uint64, error) {
	builder, err := cardanowallet.NewTxBuilder(cardanowallet.ResolveCardanoCliBinary(config.NetworkID))
	if err != nil {
		return 0, err
	}

	defer builder.Dispose()

	chainInfo, exists := cardanoChainInfos[config.ChainID]
	if !exists {
		return 0, fmt.Errorf("chain info not found for chainID: %s", config.ChainID)
	}

	builder.SetProtocolParameters(chainInfo.ProtocolParams)

	tokenAmounts := make([]cardanowallet.TokenAmount, 0, len(nativeTokensSum))

	for tokenID, tokenAmount := range nativeTokensSum {
		tokenName := config.Tokens[tokenID].ChainSpecific

		nativeToken, err := cardanowallet.NewTokenWithFullNameTry(tokenName)
		if err != nil {
			return 0, err
		}

		tokenAmounts = append(tokenAmounts, cardanowallet.NewTokenAmount(nativeToken, tokenAmount.Uint64()))
	}

	potentialTokenCost, err := cardanowallet.GetMinUtxoForSumMap(
		builder,
		receiverAddr,
		cardanowallet.GetTokensSumMap(tokenAmounts...),
		nil,
	)
	if err != nil {
		return 0, err
	}

	return max(config.UtxoMinAmount, potentialTokenCost), nil
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
		return nil, fmt.Errorf("destination chain not registered: %s", destChainID)
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
