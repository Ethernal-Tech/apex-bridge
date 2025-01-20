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
	fullSrcTokenNameFlag      = "src-token-name"                                      //nolint:gosec
	fullDestTokenNameFlag     = "dest-token-name"                                     //nolint:gosec
	fullSrcTokenNameFlagDesc  = "denom of the token to transfer from source chain"    //nolint:gosec
	fullDestTokenNameFlagDesc = "denom of the token to transfer to destination chain" //nolint:gosec
)

func ToCardanoMetadataForSkyline(receivers []*receiverAmount, sourceTokenName string) []sendtx.BridgingTxReceiver {
	metadataReceivers := make([]sendtx.BridgingTxReceiver, len(receivers))
	for idx, rec := range receivers {
		metadataReceivers[idx] = sendtx.BridgingTxReceiver{
			Addr:   rec.ReceiverAddr,
			Amount: rec.Amount.Uint64(),
		}
		if sourceTokenName == "" {
			metadataReceivers[idx].BridgingType = sendtx.BridgingTypeCurrencyOnSource
		} else {
			metadataReceivers[idx].BridgingType = sendtx.BridgingTypeNativeTokenOnSource
		}
	}

	return metadataReceivers
}

type sendSkylineTxParams struct {
	privateKeyRaw           string
	receivers               []string
	chainIDSrc              string
	chainIDDst              string
	feeString               string
	fullSrcTokenNameString  string
	fullDestTokenNameString string

	ogmiosURLSrc    string
	networkIDSrc    uint
	testnetMagicSrc uint
	multisigAddrSrc string
	ogmiosURLDst    string

	feeAmount         *big.Int
	fullSrcTokenName  cardanowallet.Token
	fullDestTokenName cardanowallet.Token
	receiversParsed   []*receiverAmount
	wallet            *cardanowallet.Wallet
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

	if (p.fullSrcTokenNameString != "" && p.fullDestTokenNameString != "") ||
		(p.fullSrcTokenNameString == "" && p.fullDestTokenNameString == "") {
		return fmt.Errorf("only one flag between %s and %s should be specified",
			p.fullSrcTokenNameString, p.fullDestTokenNameString)
	}

	if p.fullSrcTokenNameString != "" {
		tokenName, err := cardanowallet.NewTokenWithFullName(p.fullSrcTokenNameString, false)
		if err != nil {
			tokenName, err = cardanowallet.NewTokenWithFullName(p.fullSrcTokenNameString, true)
			if err != nil {
				return fmt.Errorf("--%s invalid token name: %s", fullSrcTokenNameFlag, p.fullSrcTokenNameString)
			}
		}

		p.fullSrcTokenName = tokenName
		p.fullDestTokenName.Name = cardanowallet.AdaTokenName
	}

	if p.fullDestTokenNameString != "" {
		tokenName, err := cardanowallet.NewTokenWithFullName(p.fullDestTokenNameString, false)
		if err != nil {
			tokenName, err = cardanowallet.NewTokenWithFullName(p.fullDestTokenNameString, true)
			if err != nil {
				return fmt.Errorf("--%s invalid token name: %s", fullDestTokenNameFlag, p.fullDestTokenNameString)
			}
		}

		p.fullDestTokenName = tokenName
		p.fullSrcTokenName.Name = cardanowallet.AdaTokenName
	}

	feeAmount, ok := new(big.Int).SetString(p.feeString, 0)
	if !ok {
		return fmt.Errorf("--%s invalid amount: %s", feeAmountFlag, p.feeString)
	}

	p.feeAmount = feeAmount

	minFeeForBridging := common.MinFeeForBridgingToPrime

	if p.chainIDDst == common.ChainIDStrCardano {
		minFeeForBridging = common.MinFeeForBridgingToCardano
	}

	if p.feeAmount.Uint64() < minFeeForBridging {
		return fmt.Errorf("--%s invalid amount: %d", feeAmountFlag, p.feeAmount)
	}

	bytes, err := cardanowallet.GetKeyBytes(p.privateKeyRaw)
	if err != nil || len(bytes) != 32 {
		return fmt.Errorf("invalid --%s value %s", privateKeyFlag, p.privateKeyRaw)
	}

	p.wallet = cardanowallet.NewWallet(cardanowallet.GetVerificationKeyFromSigningKey(bytes), bytes)

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
		&p.fullSrcTokenNameString,
		fullSrcTokenNameFlag,
		"",
		fullSrcTokenNameFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.fullDestTokenNameString,
		fullDestTokenNameFlag,
		"",
		fullDestTokenNameFlagDesc,
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
	receivers := ToCardanoMetadataForSkyline(p.receiversParsed, p.fullDestTokenName.Name)
	networkID := cardanowallet.CardanoNetworkType(p.networkIDSrc)

	utxoAmountSrc := common.MinUtxoAmountDefaultPrime
	utxoAmountDest := common.MinUtxoAmountDefaultCardano

	if p.chainIDSrc == common.ChainIDStrCardano {
		utxoAmountSrc = common.MinUtxoAmountDefaultCardano
		utxoAmountDest = common.MinUtxoAmountDefaultPrime
	}

	minFeeForBridgingSrc := common.MinFeeForBridgingToPrime
	minFeeForBridgingDest := common.MinFeeForBridgingToCardano

	if p.chainIDSrc == common.ChainIDStrCardano {
		minFeeForBridgingSrc = common.MinFeeForBridgingToCardano
		minFeeForBridgingDest = common.MinFeeForBridgingToPrime
	}

	txSender := sendtx.NewTxSender(
		map[string]sendtx.ChainConfig{
			p.chainIDSrc: {
				CardanoCliBinary:     cardanowallet.ResolveCardanoCliBinary(networkID),
				TxProvider:           cardanowallet.NewTxProviderOgmios(p.ogmiosURLSrc),
				MultiSigAddr:         p.multisigAddrSrc,
				TestNetMagic:         p.testnetMagicSrc,
				TTLSlotNumberInc:     ttlSlotNumberInc,
				MinBridgingFeeAmount: minFeeForBridgingSrc,
				MinUtxoValue:         utxoAmountSrc,
				NativeTokens: []sendtx.TokenExchangeConfig{
					{
						DstChainID: p.chainIDDst,
						TokenName:  p.fullSrcTokenName.Name,
					},
				},
			},
			p.chainIDDst: {
				TxProvider:           cardanowallet.NewTxProviderOgmios(p.ogmiosURLDst),
				MinUtxoValue:         utxoAmountDest,
				MinBridgingFeeAmount: minFeeForBridgingDest,
			},
		},
	)

	senderAddr, err := cardanotx.GetAddress(networkID, p.wallet)
	if err != nil {
		return nil, err
	}

	txRaw, txHash, _, err := txSender.CreateBridgingTx(
		ctx,
		p.chainIDSrc, p.chainIDDst,
		senderAddr.String(), receivers,
		p.feeAmount.Uint64(), sendtx.NewExchangeRate())
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Submiting bridging transaction..."))
	outputter.WriteOutput()

	err = txSender.SubmitTx(ctx, p.chainIDSrc, txRaw, p.wallet)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte(fmt.Sprintf("transaction has been submitted: %s", txHash)))
	outputter.WriteOutput()

	err = waitForSkylineTx(
		ctx, cardanowallet.NewTxProviderOgmios(p.ogmiosURLDst),
		p.fullSrcTokenName.Name, p.fullDestTokenName.Name, p.receiversParsed)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Transaction has been bridged"))
	outputter.WriteOutput()

	return CmdResult{
		SenderAddr: senderAddr.String(),
		ChainID:    p.chainIDDst,
		Receipts:   p.receiversParsed,
		TxHash:     txHash,
	}, nil
}

func waitForSkylineTx(
	ctx context.Context, txUtxoRetriever cardanowallet.IUTxORetriever,
	sourceTokenName string, destinationTokenName string, receivers []*receiverAmount) error {
	return waitForTx(ctx, receivers, func(ctx context.Context, addr string) (*big.Int, error) {
		utxos, err := txUtxoRetriever.GetUtxos(ctx, addr)
		if err != nil {
			return nil, err
		}

		sum := cardanowallet.GetUtxosSum(utxos)

		var receivingTokenName string

		if sourceTokenName == cardanowallet.AdaTokenName {
			receivingTokenName = destinationTokenName
		} else {
			receivingTokenName = cardanowallet.AdaTokenName
		}

		return new(big.Int).SetUint64(sum[receivingTokenName]), nil
	})
}
