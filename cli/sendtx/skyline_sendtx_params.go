package clisendtx

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/spf13/cobra"
)

const (
	operationFeeFlag     = "operation-fee"
	fullSrcTokenNameFlag = "src-token-name" //nolint:gosec
	fullDstTokenNameFlag = "dst-token-name" //nolint:gosec

	operationFeeFlagDesc     = "operation fee"
	fullSrcTokenNameFlagDesc = "denom of the token to transfer from source chain"    //nolint:gosec
	fullDstTokenNameFlagDesc = "denom of the token to transfer to destination chain" //nolint:gosec
)

type sendSkylineTxParams struct {
	privateKeyRaw      string
	stakePrivateKeyRaw string
	receivers          []string
	chainIDSrc         string
	chainIDDst         string
	feeString          string
	operationFeeString string
	tokenFullNameSrc   string
	tokenFullNameDst   string

	ogmiosURLSrc    string
	networkIDSrc    uint
	testnetMagicSrc uint
	multisigAddrSrc string
	ogmiosURLDst    string

	feeAmount          *big.Int
	operationFeeAmount *big.Int
	receiversParsed    []*receiverAmount
	wallet             *cardanowallet.Wallet
}

func (p *sendSkylineTxParams) validateFlags() error {
	if p.privateKeyRaw == "" {
		return fmt.Errorf("flag --%s not specified", privateKeyFlag)
	}

	if len(p.receivers) == 0 {
		return fmt.Errorf("--%s not specified", receiverFlag)
	}

	if !common.IsExistingSkylineChainID(p.chainIDSrc) {
		return fmt.Errorf("--%s flag not specified", srcChainIDFlag)
	}

	if !common.IsExistingSkylineChainID(p.chainIDDst) {
		return fmt.Errorf("--%s flag not specified", dstChainIDFlag)
	}

	switch {
	case p.tokenFullNameSrc != "":
		if p.tokenFullNameDst != "" {
			return fmt.Errorf("only one flag between --%s and --%s should be specified",
				fullSrcTokenNameFlag, fullDstTokenNameFlag)
		}

		token, err := getToken(p.tokenFullNameSrc)
		if err != nil {
			return fmt.Errorf("--%s invalid token name: %s", fullSrcTokenNameFlag, p.tokenFullNameSrc)
		}

		p.tokenFullNameSrc = token.String()
		p.tokenFullNameDst = cardanowallet.AdaTokenName

	case p.tokenFullNameDst != "":
		token, err := getToken(p.tokenFullNameDst)
		if err != nil {
			return fmt.Errorf("--%s invalid token name: %s", fullDstTokenNameFlag, p.tokenFullNameDst)
		}

		p.tokenFullNameSrc = cardanowallet.AdaTokenName
		p.tokenFullNameDst = token.String()

	default:
		return fmt.Errorf("specify at least one of: --%s, --%s", fullSrcTokenNameFlag, fullDstTokenNameFlag)
	}

	feeAmount, ok := new(big.Int).SetString(p.feeString, 0)
	if !ok {
		return fmt.Errorf("--%s invalid amount: %s", feeAmountFlag, p.feeString)
	}

	p.feeAmount = feeAmount

	minFeeForBridging := common.MinFeeForBridgingOnPrime

	if p.chainIDSrc == common.ChainIDStrCardano {
		minFeeForBridging = common.MinFeeForBridgingOnCardano
	}

	if p.feeAmount.Uint64() < minFeeForBridging {
		return fmt.Errorf("--%s invalid amount: %d", feeAmountFlag, p.feeAmount)
	}

	operationFeeAmount, ok := new(big.Int).SetString(p.operationFeeString, 0)
	if !ok {
		return fmt.Errorf("--%s invalid amount: %s", operationFeeFlag, p.operationFeeString)
	}

	p.operationFeeAmount = operationFeeAmount

	minOperationFee := common.MinOperationFeeOnPrime

	if p.chainIDSrc == common.ChainIDStrCardano {
		minOperationFee = common.MinOperationFeeOnCardano
	}

	if p.operationFeeAmount.Uint64() < minOperationFee {
		return fmt.Errorf("--%s invalid amount: %d", operationFeeFlag, p.operationFeeAmount)
	}

	bytes, err := getCardanoPrivateKeyBytes(p.privateKeyRaw)
	if err != nil {
		return fmt.Errorf("invalid --%s value %s", privateKeyFlag, p.privateKeyRaw)
	}

	var stakeBytes []byte
	if len(p.stakePrivateKeyRaw) > 0 {
		stakeBytes, err = getCardanoPrivateKeyBytes(p.stakePrivateKeyRaw)
		if err != nil {
			return fmt.Errorf("invalid --%s value %s", stakePrivateKeyFlag, p.stakePrivateKeyRaw)
		}
	}

	p.wallet = cardanowallet.NewWallet(bytes, stakeBytes)

	if !common.IsValidHTTPURL(p.ogmiosURLSrc) {
		return fmt.Errorf("invalid --%s: %s", ogmiosURLSrcFlag, p.ogmiosURLSrc)
	}

	if p.multisigAddrSrc == "" {
		return fmt.Errorf("--%s not specified", multisigAddrSrcFlag)
	}

	if p.ogmiosURLDst == "" {
		return fmt.Errorf("--%s and --%s not specified", ogmiosURLDstFlag, nexusURLFlag)
	}

	if p.ogmiosURLDst != "" && !common.IsValidHTTPURL(p.ogmiosURLDst) {
		return fmt.Errorf("invalid --%s: %s", ogmiosURLDstFlag, p.ogmiosURLDst)
	}

	receivers := make([]*receiverAmount, 0, len(p.receivers))

	for i, x := range p.receivers {
		vals := strings.Split(x, ":")
		if len(vals) != 2 {
			return fmt.Errorf("--%s number %d is invalid: %s", receiverFlag, i, x)
		}

		amount, ok := new(big.Int).SetString(vals[1], 0)
		if !ok {
			return fmt.Errorf("--%s number %d has invalid amount: %s", receiverFlag, i, x)
		}

		if !common.IsValidAddress(p.chainIDDst, vals[0]) {
			return fmt.Errorf("--%s number %d has invalid address: %s", receiverFlag, i, x)
		}

		receivers = append(receivers, &receiverAmount{
			ReceiverAddr: vals[0],
			Amount:       amount,
		})
	}

	p.receiversParsed = receivers

	return nil
}

func (p *sendSkylineTxParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&p.privateKeyRaw,
		privateKeyFlag,
		"",
		privateKeyFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.stakePrivateKeyRaw,
		stakePrivateKeyFlag,
		"",
		stakePrivateKeyFlagDesc,
	)

	cmd.Flags().StringArrayVar(
		&p.receivers,
		receiverFlag,
		nil,
		receiverFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.chainIDSrc,
		srcChainIDFlag,
		"",
		srcChainIDFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.chainIDDst,
		dstChainIDFlag,
		"",
		dstChainIDFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.feeString,
		feeAmountFlag,
		"0",
		feeAmountFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.operationFeeString,
		operationFeeFlag,
		"0",
		operationFeeFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.tokenFullNameSrc,
		fullSrcTokenNameFlag,
		"",
		fullSrcTokenNameFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.tokenFullNameDst,
		fullDstTokenNameFlag,
		"",
		fullDstTokenNameFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.ogmiosURLSrc,
		ogmiosURLSrcFlag,
		"",
		ogmiosURLSrcFlagDesc,
	)

	cmd.Flags().UintVar(
		&p.networkIDSrc,
		networkIDSrcFlag,
		0,
		networkIDSrcFlagDesc,
	)

	cmd.Flags().UintVar(
		&p.testnetMagicSrc,
		testnetMagicFlag,
		0,
		testnetMagicFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.multisigAddrSrc,
		multisigAddrSrcFlag,
		"",
		multisigAddrSrcFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.ogmiosURLDst,
		ogmiosURLDstFlag,
		"",
		ogmiosURLDstFlagDesc,
	)
}

func (p *sendSkylineTxParams) Execute(
	outputter common.OutputFormatter,
) (common.ICommandResult, error) {
	ctx := context.Background()
	receivers := toCardanoMetadataForSkyline(p.receiversParsed, p.tokenFullNameSrc)
	networkID := cardanowallet.CardanoNetworkType(p.networkIDSrc)

	utxoAmountSrc := common.MinUtxoAmountDefaultPrime
	utxoAmountDest := common.MinUtxoAmountDefaultCardano

	if p.chainIDSrc == common.ChainIDStrCardano {
		utxoAmountSrc = common.MinUtxoAmountDefaultCardano
		utxoAmountDest = common.MinUtxoAmountDefaultPrime
	}

	minFeeForBridgingSrc := common.MinFeeForBridgingOnPrime
	minFeeForBridgingDest := common.MinFeeForBridgingOnCardano

	if p.chainIDSrc == common.ChainIDStrCardano {
		minFeeForBridgingSrc = common.MinFeeForBridgingOnCardano
		minFeeForBridgingDest = common.MinFeeForBridgingOnPrime
	}

	minOperationFeeSrc := common.MinOperationFeeOnPrime
	minOperationFeeDest := common.MinOperationFeeOnCardano

	if p.chainIDSrc == common.ChainIDStrCardano {
		minOperationFeeSrc = common.MinOperationFeeOnCardano
		minOperationFeeDest = common.MinOperationFeeOnPrime
	}

	var srcNativeTokens []sendtx.TokenExchangeConfig
	if p.tokenFullNameSrc != cardanowallet.AdaTokenName {
		srcNativeTokens = append(srcNativeTokens, sendtx.TokenExchangeConfig{
			DstChainID: p.chainIDDst,
			TokenName:  p.tokenFullNameSrc,
		})
	}

	txSender := sendtx.NewTxSender(
		map[string]sendtx.ChainConfig{
			p.chainIDSrc: {
				CardanoCliBinary:      cardanowallet.ResolveCardanoCliBinary(networkID),
				TxProvider:            cardanowallet.NewTxProviderOgmios(p.ogmiosURLSrc),
				MultiSigAddr:          p.multisigAddrSrc,
				TestNetMagic:          p.testnetMagicSrc,
				TTLSlotNumberInc:      ttlSlotNumberInc,
				MinBridgingFeeAmount:  minFeeForBridgingSrc,
				MinOperationFeeAmount: minOperationFeeSrc,
				MinUtxoValue:          utxoAmountSrc,
				NativeTokens:          srcNativeTokens,
			},
			p.chainIDDst: {
				MinUtxoValue:          utxoAmountDest,
				MinBridgingFeeAmount:  minFeeForBridgingDest,
				MinOperationFeeAmount: minOperationFeeDest,
			},
		},
	)

	senderAddr, err := cardanotx.GetAddress(networkID, p.wallet)
	if err != nil {
		return nil, err
	}

	txInfo, _, err := txSender.CreateBridgingTx(
		ctx,
		p.chainIDSrc, p.chainIDDst,
		senderAddr.String(), receivers,
		p.feeAmount.Uint64(), p.operationFeeAmount.Uint64())
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Submiting bridging transaction..."))
	outputter.WriteOutput()

	err = txSender.SubmitTx(ctx, p.chainIDSrc, txInfo.TxRaw, p.wallet)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte(fmt.Sprintf("transaction has been submitted: %s", txInfo.TxHash)))
	outputter.WriteOutput()

	err = waitForSkylineTx(
		ctx, cardanowallet.NewTxProviderOgmios(p.ogmiosURLDst), p.tokenFullNameDst, p.receiversParsed)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Transaction has been bridged"))
	outputter.WriteOutput()

	return CmdResult{
		SenderAddr: senderAddr.String(),
		ChainID:    p.chainIDDst,
		Receipts:   p.receiversParsed,
		TxHash:     txInfo.TxHash,
	}, nil
}

func waitForSkylineTx(
	ctx context.Context, txUtxoRetriever cardanowallet.IUTxORetriever,
	tokenName string, receivers []*receiverAmount) error {
	return waitForTx(ctx, receivers, func(ctx context.Context, addr string) (*big.Int, error) {
		utxos, err := txUtxoRetriever.GetUtxos(ctx, addr)
		if err != nil {
			return nil, err
		}

		return new(big.Int).SetUint64(cardanowallet.GetUtxosSum(utxos)[tokenName]), nil
	})
}

func toCardanoMetadataForSkyline(receivers []*receiverAmount, sourceTokenName string) []sendtx.BridgingTxReceiver {
	metadataReceivers := make([]sendtx.BridgingTxReceiver, len(receivers))
	for idx, rec := range receivers {
		metadataReceivers[idx] = sendtx.BridgingTxReceiver{
			Addr:   rec.ReceiverAddr,
			Amount: rec.Amount.Uint64(),
		}
		if sourceTokenName == cardanowallet.AdaTokenName {
			metadataReceivers[idx].BridgingType = sendtx.BridgingTypeCurrencyOnSource
		} else {
			metadataReceivers[idx].BridgingType = sendtx.BridgingTypeNativeTokenOnSource
		}
	}

	return metadataReceivers
}

func getToken(fullName string) (token cardanowallet.Token, err error) {
	token, err = cardanowallet.NewTokenWithFullName(fullName, false)
	if err == nil {
		return token, nil
	}

	token, err = cardanowallet.NewTokenWithFullName(fullName, true)
	if err == nil {
		return token, nil
	}

	return token, err
}
