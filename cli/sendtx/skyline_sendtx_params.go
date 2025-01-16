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
	adaNativeCurrencyDenom  = "ada"
	apexNativeCurrencyDenom = "apex"
	adaNativeTokenDenom     = "wada"
	apexNativeTokenDenom    = "wapex"
)

const (
	tokenDenomFlag     = "token-denom"
	tokenDenomFlagDesc = "denom of the token to transfer" //nolint:gosec
)

func ToCardanoMetadataForSkyline(receivers []*receiverAmount, tokenDenom string) []sendtx.BridgingTxReceiver {
	metadataReceivers := make([]sendtx.BridgingTxReceiver, len(receivers))
	for idx, rec := range receivers {
		metadataReceivers[idx] = sendtx.BridgingTxReceiver{
			Addr:   rec.ReceiverAddr,
			Amount: rec.Amount.Uint64(),
		}
		if tokenDenom == adaNativeCurrencyDenom || tokenDenom == apexNativeCurrencyDenom {
			metadataReceivers[idx].BridgingType = sendtx.BridgingTypeCurrencyOnSource
		} else {
			metadataReceivers[idx].BridgingType = sendtx.BridgingTypeNativeTokenOnSource
		}
	}

	return metadataReceivers
}

type sendSkylineTxParams struct {
	privateKeyRaw string
	receivers     []string
	chainIDSrc    string
	chainIDDst    string
	feeString     string
	tokenDenom    string

	ogmiosURLSrc    string
	networkIDSrc    uint
	testnetMagicSrc uint
	multisigAddrSrc string
	ogmiosURLDst    string

	feeAmount       *big.Int
	receiversParsed []*receiverAmount
	wallet          *cardanowallet.Wallet
}

func (p *sendSkylineTxParams) validateFlags() error {
	if p.privateKeyRaw == "" {
		return fmt.Errorf("flag --%s not specified", privateKeyFlag)
	}

	if len(p.receivers) == 0 {
		return fmt.Errorf("--%s not specified", receiverFlag)
	}

	if !common.IsExistingChainID(p.chainIDSrc) {
		return fmt.Errorf("--%s flag not specified", srcChainIDFlag)
	}

	if !common.IsExistingChainID(p.chainIDDst) {
		return fmt.Errorf("--%s flag not specified", dstChainIDFlag)
	}

	if p.chainIDSrc == common.ChainIDStrCardano {
		if p.tokenDenom != adaNativeCurrencyDenom && p.tokenDenom != apexNativeTokenDenom {
			return fmt.Errorf("--%s invalid denom for chain: %s", tokenDenomFlag, p.chainIDSrc)
		}
	} else {
		if p.tokenDenom != apexNativeCurrencyDenom && p.tokenDenom != adaNativeTokenDenom {
			return fmt.Errorf("--%s invalid denom for chain: %s", tokenDenomFlag, p.chainIDSrc)
		}
	}

	feeAmount, ok := new(big.Int).SetString(p.feeString, 0)
	if !ok {
		return fmt.Errorf("--%s invalid amount: %s", feeAmountFlag, p.feeString)
	}

	p.feeAmount = feeAmount

	if p.feeAmount.Uint64() < common.MinFeeForBridgingDefault {
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

		if p.chainIDDst != common.ChainIDStrNexus &&
			amount.Cmp(new(big.Int).SetUint64(common.MinUtxoAmountDefault)) < 0 {
			return fmt.Errorf("--%s number %d has insufficient amount: %s", receiverFlag, i, x)
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
		&p.tokenDenom,
		tokenDenomFlag,
		"",
		feeAmountFlagDesc,
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
	receivers := ToCardanoMetadataForSkyline(p.receiversParsed, p.tokenDenom)
	networkID := cardanowallet.CardanoNetworkType(p.networkIDSrc)
	txSender := sendtx.NewTxSender(
		p.feeAmount.Uint64(),
		common.MinUtxoAmountDefault,
		common.PotentialFeeDefault,
		common.MaxInputsPerBridgingTxDefault,
		map[string]sendtx.ChainConfig{
			p.chainIDSrc: {
				CardanoCliBinary: cardanowallet.ResolveCardanoCliBinary(networkID),
				TxProvider:       cardanowallet.NewTxProviderOgmios(p.ogmiosURLSrc),
				MultiSigAddr:     p.multisigAddrSrc,
				TestNetMagic:     p.testnetMagicSrc,
				TTLSlotNumberInc: ttlSlotNumberInc,
				MinUtxoValue:     common.MinUtxoAmountDefault,
				ExchangeRate:     make(map[string]float64),
			},
			p.chainIDDst: {
				TxProvider: cardanowallet.NewTxProviderOgmios(p.ogmiosURLDst),
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
		senderAddr.String(), receivers, sendtx.NewExchangeRate())
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

	err = waitForTxOnCardano(
		ctx, cardanowallet.NewTxProviderOgmios(p.ogmiosURLDst),
		p.receiversParsed)
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
