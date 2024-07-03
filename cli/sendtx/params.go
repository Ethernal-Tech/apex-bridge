package clisendtx

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/spf13/cobra"
)

const (
	privateKeyFlag         = "key"
	ogmiosURLSrcFlag       = "ogmios-src"
	receiverFlag           = "receiver"
	networkIDSrcFlag       = "network-id-src"
	testnetMagicFlag       = "testnet-src"
	chainIDFlag            = "chain-dst"
	multisigAddrSrcFlag    = "addr-multisig-src"
	multisigFeeAddrDstFlag = "addr-fee-dst"
	feeAmountFlag          = "fee"
	ogmiosURLDstFlag       = "ogmios-dst"

	privateKeyFlagDesc         = "wallet private signing key"
	ogmiosURLSrcFlagDesc       = "source chain ogmios url"
	receiverFlagDesc           = "receiver addr:amount"
	testnetMagicFlagDesc       = "source testnet magic number. leave 0 for mainnet"
	networkIDSrcFlagDesc       = "source network id"
	chainIDFlagDesc            = "destination chain ID (prime, vector, etc)"
	multisigAddrSrcFlagDesc    = "source multisig address"
	multisigFeeAddrDstFlagDesc = "destination fee payer address"
	feeAmountFlagDesc          = "amount for multisig fee addr"
	ogmiosURLDstFlagDesc       = "destination chain ogmios url"

	defaultFeeAmount = 1_100_000
	ttlSlotNumberInc = 500
)

type sendTxParams struct {
	privateKeyRaw      string
	ogmiosURLSrc       string
	receivers          []string
	networkIDSrc       uint
	testnetMagicSrc    uint
	chainIDDst         string
	multisigAddrSrc    string
	multisigFeeAddrDst string
	feeAmount          uint64
	ogmiosURLDst       string

	receiversParsed []cardanowallet.TxOutput
	wallet          cardanowallet.IWallet
}

func (ip *sendTxParams) validateFlags() error {
	if ip.privateKeyRaw == "" {
		return fmt.Errorf("flag --%s not specified", privateKeyFlag)
	}

	if ip.ogmiosURLSrc == "" || !common.IsValidURL(ip.ogmiosURLSrc) {
		return fmt.Errorf("invalid --%s: %s", ogmiosURLSrcFlag, ip.ogmiosURLSrc)
	}

	if ip.ogmiosURLDst != "" && !common.IsValidURL(ip.ogmiosURLDst) {
		return fmt.Errorf("invalid --%s: %s", ogmiosURLDstFlag, ip.ogmiosURLDst)
	}

	if ip.chainIDDst == "" {
		return fmt.Errorf("--%s flag not specified", chainIDFlag)
	}

	if len(ip.receivers) == 0 {
		return fmt.Errorf("--%s not specified", receiverFlag)
	}

	if ip.multisigFeeAddrDst == "" {
		return fmt.Errorf("--%s not specified", multisigFeeAddrDstFlag)
	}

	if ip.multisigAddrSrc == "" {
		return fmt.Errorf("--%s not specified", multisigAddrSrcFlag)
	}

	if ip.feeAmount < cardanowallet.MinUTxODefaultValue {
		return fmt.Errorf("--%s invalid amount: %d", feeAmountFlag, ip.feeAmount)
	}

	bytes, err := cardanowallet.GetKeyBytes(ip.privateKeyRaw)
	if err != nil || len(bytes) != 32 {
		return fmt.Errorf("invalid --%s value %s", privateKeyFlag, ip.privateKeyRaw)
	}

	ip.wallet = cardanowallet.NewWallet(cardanowallet.GetVerificationKeyFromSigningKey(bytes), bytes)

	receivers := make([]cardanowallet.TxOutput, 0, len(ip.receivers))

	for i, x := range ip.receivers {
		vals := strings.Split(x, ":")
		if len(vals) != 2 {
			return fmt.Errorf("--%s number %d is invalid: %s", receiverFlag, i, x)
		}

		amount, err := strconv.ParseUint(vals[1], 0, 64)
		if err != nil {
			return fmt.Errorf("--%s number %d has invalid amount: %s", receiverFlag, i, x)
		}

		if amount < cardanowallet.MinUTxODefaultValue {
			return fmt.Errorf("--%s number %d has insufficient amount: %s", receiverFlag, i, x)
		}

		_, err = cardanowallet.NewAddress(vals[0])
		if err != nil {
			return fmt.Errorf("--%s number %d has invalid address: %s", receiverFlag, i, x)
		}

		receivers = append(receivers, cardanowallet.TxOutput{
			Addr:   vals[0],
			Amount: amount,
		})
	}

	if cardanotx.IsAddressInOutputs(receivers, ip.multisigFeeAddrDst) {
		return errors.New("fee address can not be in receivers list")
	}

	receivers = append(receivers, cardanowallet.TxOutput{
		Addr:   ip.multisigFeeAddrDst,
		Amount: ip.feeAmount,
	})

	ip.receiversParsed = receivers

	return nil
}

func (ip *sendTxParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&ip.privateKeyRaw,
		privateKeyFlag,
		"",
		privateKeyFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.ogmiosURLSrc,
		ogmiosURLSrcFlag,
		"",
		ogmiosURLSrcFlagDesc,
	)

	cmd.Flags().StringArrayVar(
		&ip.receivers,
		receiverFlag,
		nil,
		receiverFlagDesc,
	)

	cmd.Flags().UintVar(
		&ip.testnetMagicSrc,
		testnetMagicFlag,
		0,
		testnetMagicFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.chainIDDst,
		chainIDFlag,
		"",
		chainIDFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.multisigAddrSrc,
		multisigAddrSrcFlag,
		"",
		multisigAddrSrcFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.multisigFeeAddrDst,
		multisigFeeAddrDstFlag,
		"",
		multisigFeeAddrDstFlagDesc,
	)

	cmd.Flags().Uint64Var(
		&ip.feeAmount,
		feeAmountFlag,
		defaultFeeAmount,
		feeAmountFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.ogmiosURLDst,
		ogmiosURLDstFlag,
		"",
		ogmiosURLDstFlagDesc,
	)

	cmd.Flags().UintVar(
		&ip.networkIDSrc,
		networkIDSrcFlag,
		0,
		networkIDSrcFlagDesc,
	)
}

func (ip *sendTxParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	networkID := cardanowallet.CardanoNetworkType(ip.networkIDSrc)
	cardanoCliBinary := cardanowallet.ResolveCardanoCliBinary(networkID)
	txSender := cardanotx.NewBridgingTxSender(
		cardanoCliBinary,
		cardanowallet.NewTxProviderOgmios(ip.ogmiosURLSrc),
		cardanowallet.NewTxProviderOgmios(ip.ogmiosURLDst),
		ip.testnetMagicSrc, ip.multisigAddrSrc, ttlSlotNumberInc)

	senderAddr, err := cardanotx.GetAddress(networkID, ip.wallet)
	if err != nil {
		return nil, err
	}

	txRaw, txHash, err := txSender.CreateTx(
		context.Background(), ip.chainIDDst, senderAddr.String(), ip.receiversParsed)
	if err != nil {
		return nil, err
	}

	err = txSender.SendTx(context.Background(), txRaw, txHash, ip.wallet)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte(fmt.Sprintf("transaction has been submitted: %s", txHash)))
	outputter.WriteOutput()

	if ip.ogmiosURLDst != "" {
		err = txSender.WaitForTx(context.Background(), ip.receiversParsed)
		if err != nil {
			return nil, err
		}
	}

	return CmdResult{
		SenderAddr: senderAddr.String(),
		ChainID:    ip.chainIDDst,
		Receipts:   ip.receiversParsed,
		TxHash:     txHash,
	}, nil
}
