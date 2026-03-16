package successtxprocessors

import (
	"fmt"
	"math/big"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	oChain "github.com/Ethernal-Tech/apex-bridge/oracle_common/chain"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	oUtils "github.com/Ethernal-Tech/apex-bridge/oracle_common/utils"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	goEthCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
)

var _ core.SolanaTxSuccessProcessor = (*BridgingRequestedProcessorImpl)(nil)

type BridgingRequestedProcessorImpl struct {
	logger hclog.Logger

	cardanoChainInfos map[string]*oChain.CardanoChainInfo
}

type receiverValidationCtxSolanaSrc struct {
	oCore.ReceiverValidationContext
	solanaSrcConfig *oCore.SolanaChainConfig
	metadata        *core.BridgingRequestSolMetadata
	feeSum          uint64
}

func NewSolanaBridgingRequestedProcessor(
	logger hclog.Logger,
	cardanoChainInfos map[string]*oChain.CardanoChainInfo,
) *BridgingRequestedProcessorImpl {
	return &BridgingRequestedProcessorImpl{
		logger:            logger.Named("solana_bridging_requested_processor"),
		cardanoChainInfos: cardanoChainInfos,
	}
}

func (*BridgingRequestedProcessorImpl) GetType() common.BridgingTxType {
	return common.BridgingTxTypeBridgingRequest
}

func (*BridgingRequestedProcessorImpl) PreValidate(tx *core.SolanaTx, appConfig *oCore.AppConfig) error {
	return nil
}

func (p *BridgingRequestedProcessorImpl) ValidateAndAddClaim(
	claims *oCore.BridgeClaims, tx *core.SolanaTx, appConfig *oCore.AppConfig,
) error {
	metadata, err := core.UnmarshalSolMetadata[core.BridgingRequestSolMetadata](tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal sol metadata: %w", err)
	}

	if metadata.BridgingTxType != p.GetType() {
		// wTODO: Refund
		return fmt.Errorf("irrelevant tx. Tx type: %s", metadata.BridgingTxType)
	}

	p.logger.Debug("Validating relevant tx", "txHash", tx.TxSignature, "metadata", metadata)

	err = p.validate(tx, metadata, appConfig)
	if err == nil {
		return p.addBridgingRequestClaim(claims, tx, metadata, appConfig)
	} else {
		// wTODO: Refund
		return err
	}
}

func (p *BridgingRequestedProcessorImpl) validate(
	tx *core.SolanaTx, metadata *core.BridgingRequestSolMetadata, appConfig *oCore.AppConfig,
) error {
	originChainConfig := oUtils.GetChainConfigResult(appConfig, tx.OriginChainID)
	if originChainConfig.IsNone() {
		return fmt.Errorf("origin chain not registered: %v", tx.OriginChainID)
	}

	destinationChainConfig := oUtils.GetChainConfigResult(appConfig, metadata.DestinationChainID)
	if destinationChainConfig.IsNone() {
		return fmt.Errorf("destination chain not registered: %v", metadata.DestinationChainID)
	}

	destChainInfo, err := oUtils.GetDestChainInfoResult(metadata.DestinationChainID, destinationChainConfig, appConfig)
	if err != nil {
		return err
	}

	if err := p.validateOperationAndReceiverLimits(metadata, originChainConfig, appConfig); err != nil {
		return err
	}

	srcCurrencyID, err := originChainConfig.Solana.GetCurrencyID()
	if err != nil {
		return err
	}

	receiverCtx := &receiverValidationCtxSolanaSrc{
		solanaSrcConfig: originChainConfig.Solana,
		metadata:        metadata,
		ReceiverValidationContext: oCore.ReceiverValidationContext{
			CardanoDestConfig: destinationChainConfig.Cardano,
			EthDestConfig:     destinationChainConfig.Eth,
			DestFeeAddress:    destChainInfo.FeeAddress,
			BridgingSettings:  &appConfig.BridgingSettings,
			MinColCoinsAllowedToBridge: oUtils.MaxBigInt(
				common.LamportToWei(new(big.Int).SetUint64(originChainConfig.Solana.MinColCoinsAllowedToBridge)),
				destChainInfo.MinColCoinsAllowedToBridge),
			AmountsSums:    make(map[uint16]*big.Int),
			CurrencySrcID:  srcCurrencyID,
			CurrencyDestID: destChainInfo.CurrencyTokenID,
		},
	}

	for _, receiver := range metadata.Transactions {
		if err := p.validateReceiver(&receiver, receiverCtx); err != nil {
			return err
		}
	}

	if err := p.validateTokenAmounts(tx.Value, receiverCtx); err != nil {
		return err
	}

	return nil
}

func (p *BridgingRequestedProcessorImpl) validateReceiver(
	receiver *core.BridgingRequestSolMetadataTransaction,
	ctx *receiverValidationCtxSolanaSrc,
) error {
	tokenPair, err := oUtils.GetTokenPair(
		ctx.solanaSrcConfig.DestinationChains,
		ctx.solanaSrcConfig.ChainID,
		ctx.metadata.DestinationChainID,
		receiver.TokenID,
	)
	if err != nil {
		return fmt.Errorf("invalid receiver. metadata: %v, receiver: %v, err: %w", ctx.metadata, receiver, err)
	}

	if ctx.CardanoDestConfig != nil {
		return p.validateReceiverCardano(ctx, receiver, tokenPair)
	}

	return p.validateReceiverEth(ctx, receiver, tokenPair)
}

func (p *BridgingRequestedProcessorImpl) validateReceiverCardano(
	ctx *receiverValidationCtxSolanaSrc,
	receiver *core.BridgingRequestSolMetadataTransaction,
	tokenPair *common.TokenPair,
) error {
	if !cardanotx.IsValidOutputAddress(receiver.Address, ctx.CardanoDestConfig.NetworkID) {
		return fmt.Errorf(
			"found an invalid receiver addr in metadata. metadata: %v, receiver: %v", ctx.metadata, receiver)
	}

	if tokenPair.DestinationTokenID == ctx.CurrencyDestID {
		utxoMinWeiDest := common.DfmToWei(new(big.Int).SetUint64(ctx.CardanoDestConfig.UtxoMinAmount))
		if receiver.Amount.Cmp(utxoMinWeiDest) < 0 {
			return fmt.Errorf("found an utxo value below minimum value in metadata receivers: %v", ctx.metadata)
		}
	} else if receiver.Amount.Cmp(ctx.MinColCoinsAllowedToBridge) < 0 {
		return fmt.Errorf("token amount below minimum allowed in metadata receivers: %v", ctx.metadata)
	}

	if tokensSum, ok := ctx.AmountsSums[tokenPair.SourceTokenID]; ok {
		tokensSum.Add(tokensSum, receiver.Amount)
	} else {
		ctx.AmountsSums[tokenPair.SourceTokenID] = new(big.Int).Set(receiver.Amount)
	}

	return nil
}

func (p *BridgingRequestedProcessorImpl) validateReceiverEth(
	ctx *receiverValidationCtxSolanaSrc,
	receiver *core.BridgingRequestSolMetadataTransaction,
	tokenPair *common.TokenPair,
) error {
	if !goEthCommon.IsHexAddress(receiver.Address) {
		return fmt.Errorf(
			"found an invalid eth receiver addr in metadata. metadata: %v, receiver: %v", ctx.metadata, receiver)
	}

	if tokensSum, ok := ctx.AmountsSums[tokenPair.SourceTokenID]; ok {
		tokensSum.Add(tokensSum, receiver.Amount)
	} else {
		ctx.AmountsSums[tokenPair.SourceTokenID] = new(big.Int).Set(receiver.Amount)
	}

	if receiver.Amount.Cmp(ctx.MinColCoinsAllowedToBridge) < 0 {
		return fmt.Errorf("token amount below minimum allowed in metadata receivers: %v", ctx.metadata)
	}

	return nil
}

func (p *BridgingRequestedProcessorImpl) validateTokenAmounts(
	txValue *big.Int, receiverCtx *receiverValidationCtxSolanaSrc,
) error {
	nativeCurrencySum, ok := receiverCtx.AmountsSums[receiverCtx.CurrencySrcID]
	if !ok {
		nativeCurrencySum = new(big.Int).SetInt64(0)
	}

	delete(receiverCtx.AmountsSums, receiverCtx.CurrencySrcID)

	maxCurrAmt := receiverCtx.BridgingSettings.MaxAmountAllowedToBridge
	if maxCurrAmt != nil && maxCurrAmt.Sign() > 0 && nativeCurrencySum.Cmp(maxCurrAmt) == 1 {
		return fmt.Errorf("sum of receiver amounts: %v greater than maximum allowed: %v",
			nativeCurrencySum, maxCurrAmt)
	}

	metadata := receiverCtx.metadata

	if metadata.BridgingFee < receiverCtx.solanaSrcConfig.MinFeeForBridging {
		return fmt.Errorf("bridging fee in metadata is less than minimum: %v", metadata)
	}

	return nil
}

func (p *BridgingRequestedProcessorImpl) validateOperationAndReceiverLimits(
	metadata *core.BridgingRequestSolMetadata,
	originChainConfig oUtils.ChainConfigResult,
	appConfig *oCore.AppConfig,
) error {
	if metadata.OperationFee < originChainConfig.Solana.MinOperationFee {
		return fmt.Errorf("operation fee in metadata is less than minimum: %v", metadata)
	}

	if len(metadata.Transactions) > appConfig.BridgingSettings.MaxReceiversPerBridgingRequest {
		return fmt.Errorf("number of receivers in metadata greater than maximum allowed - no: %v, max: %v, metadata: %v",
			len(metadata.Transactions), appConfig.BridgingSettings.MaxReceiversPerBridgingRequest, metadata)
	}

	return nil
}

func (p *BridgingRequestedProcessorImpl) addBridgingRequestClaim(
	claims *oCore.BridgeClaims,
	tx *core.SolanaTx,
	metadata *core.BridgingRequestSolMetadata,
	appConfig *oCore.AppConfig,
) error {
	chainIDConverter := appConfig.ChainIDConverter

	originChainConfig := oUtils.GetChainConfigResult(appConfig, tx.OriginChainID)
	if originChainConfig.IsNone() {
		return fmt.Errorf("origin chain not registered: %v", tx.OriginChainID)
	}

	destinationChainConfig := oUtils.GetChainConfigResult(appConfig, metadata.DestinationChainID)
	if destinationChainConfig.IsNone() {
		return fmt.Errorf("destination chain not registered: %v", metadata.DestinationChainID)
	}

	receivers := make([]oCore.BridgingRequestReceiver, 0, len(metadata.Transactions))
	totalTokensAmount := oCore.NewTotalTokensAmount()

	for _, receiver := range metadata.Transactions {
		brReceiver, err := originChainConfig.ProcessReceiver(
			&destinationChainConfig,
			&receiver,
			totalTokensAmount,
			p.cardanoChainInfos,
		)
		if err != nil {
			return fmt.Errorf(
				"failed to process receiver (chain %s, receiver address: %v): %w",
				metadata.DestinationChainID,
				receiver.Address,
				err,
			)
		}

		receivers = append(receivers, *brReceiver)
	}

	claim := oCore.BridgingRequestClaim{
		ObservedTransactionHash:         tx.TxSignature[:],
		SourceChainId:                   chainIDConverter.ToChainIDNum(tx.OriginChainID),
		DestinationChainId:              chainIDConverter.ToChainIDNum(metadata.DestinationChainID),
		Receivers:                       receivers,
		NativeCurrencyAmountSource:      totalTokensAmount.TotalAmountCurrencySrc,
		NativeCurrencyAmountDestination: totalTokensAmount.TotalAmountCurrencyDst,
		WrappedTokenAmountSource:        totalTokensAmount.TotalAmountWrappedSrc,
		WrappedTokenAmountDestination:   totalTokensAmount.TotalAmountWrappedDst,
		RetryCounter:                    big.NewInt(int64(tx.BatchTryCount)),
	}

	claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, claim)

	p.logger.Info("Added BridgingRequestClaim",
		"txHash", tx.TxSignature, "metadata", metadata, "claim", oCore.BridgingRequestClaimString(claim, chainIDConverter))

	return nil
}
