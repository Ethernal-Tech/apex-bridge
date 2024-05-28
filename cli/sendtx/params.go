package clisendtx

import (
	"context"
	"fmt"
	"path"
	"strconv"
	"strings"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/spf13/cobra"
)

const (
	privateKeyDirFlag      = "key-dir"
	privateKeyFlag         = "key"
	ogmiosURLSrcFlag       = "ogmios-src"
	receiverFlag           = "receiver"
	testnetMagicFlag       = "testnet-src"
	chainIDFlag            = "chain-dst"
	multisigAddrSrcFlag    = "addr-multisig-src"
	multisigFeeAddrDstFlag = "addr-fee-dst"
	feeAmountFlag          = "fee"
	ogmiosURLDstFlag       = "ogmios-dst"

	privateKeyDirFlagDesc      = "wallet directory"
	privateKeyFlagDesc         = "wallet private signing key"
	ogmiosURLSrcFlagDesc       = "source chain ogmios url"
	receiverFlagDesc           = "receiver addr:amount"
	testnetMagicFlagDesc       = "source testnet magic number. leave 0 for mainnet"
	chainIDFlagDesc            = "destination chain ID (prime, vector, etc)"
	multisigAddrSrcFlagDesc    = "source multisig address"
	multisigFeeAddrDstFlagDesc = "destination fee payer address"
	feeAmountFlagDesc          = "amount for multisig fee addr"
	ogmiosURLDstFlagDesc       = "destination chain ogmios url"

	defaultFeeAmount = 1_100_000
	ttlSlotNumberInc = 500
)

type sendTxParams struct {
	privateKeyDirectory string
	privateKeyRaw       string
	ogmiosURLSrc        string
	receivers           []string
	testnetMagicSrc     uint
	chainIDDst          string
	multisigAddrSrc     string
	multisigFeeAddrDst  string
	feeAmount           uint64
	ogmiosURLDst        string

	receiversParsed []cardanowallet.TxOutput
	wallet          cardanowallet.IWallet
}

func (ip *sendTxParams) validateFlags() error {
	if ip.privateKeyDirectory == "" && ip.privateKeyRaw == "" {
		return fmt.Errorf("--%s or --%s must be specified", privateKeyDirFlag, privateKeyFlag)
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

	if ip.privateKeyRaw != "" {
		bytes, err := (cardanowallet.Key{
			Hex: ip.privateKeyRaw,
		}).GetKeyBytes()
		if err != nil || len(bytes) != 32 {
			return fmt.Errorf("invalid --%s value %s", privateKeyFlag, ip.privateKeyRaw)
		}

		ip.wallet = cardanowallet.NewWallet(cardanowallet.GetVerificationKeyFromSigningKey(bytes),
			bytes)
	} else {
		// first try to load with stake wallet manager, then with wallet manager
		wallet, err := cardanowallet.NewStakeWalletManager().Load(path.Clean(ip.privateKeyDirectory))
		if err != nil {
			wallet, err = cardanowallet.NewWalletManager().Load(path.Clean(ip.privateKeyDirectory))
			if err != nil {
				return fmt.Errorf("invalid --%s value %s, err: %w", privateKeyDirFlag, ip.privateKeyDirectory, err)
			}
		}

		ip.wallet = wallet
	}

	receivers := make([]cardanowallet.TxOutput, len(ip.receivers))

	for i, x := range ip.receivers {
		vals := strings.Split(x, ":")
		if len(vals) != 2 {
			return fmt.Errorf("--%s number %d is invalid: %s", receiverFlag, i, x)
		}

		amount, err := strconv.ParseUint(vals[1], 0, 64)
		if err != nil {
			return fmt.Errorf("--%s number %d has invalid amount: %s", receiverFlag, i, x)
		}

		_, err = cardanowallet.NewAddress(vals[0])
		if err != nil {
			return fmt.Errorf("--%s number %d has invalid address: %s", receiverFlag, i, x)
		}

		receivers[i] = cardanowallet.TxOutput{
			Addr:   vals[0],
			Amount: amount,
		}
	}

	ip.receiversParsed = receivers

	return nil
}

func (ip *sendTxParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&ip.privateKeyDirectory,
		privateKeyDirFlag,
		"",
		privateKeyDirFlagDesc,
	)

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

	cmd.MarkFlagsMutuallyExclusive(privateKeyDirFlag, privateKeyFlag)
}

func (ip *sendTxParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	txSender := cardanotx.NewBridgingTxSender(
		cardanowallet.NewTxProviderOgmios(ip.ogmiosURLSrc),
		cardanowallet.NewTxProviderOgmios(ip.ogmiosURLDst),
		ip.testnetMagicSrc, ip.multisigAddrSrc, ip.multisigFeeAddrDst, ip.feeAmount, ttlSlotNumberInc)

	senderAddr, _, err := cardanowallet.GetWalletAddressCli(ip.wallet, ip.testnetMagicSrc)
	if err != nil {
		return nil, err
	}

	txRaw, txHash, err := txSender.CreateTx(context.Background(), ip.chainIDDst, senderAddr, ip.receiversParsed)
	if err != nil {
		return nil, err
	}

	err = txSender.SendTx(context.Background(), ip.wallet, txRaw, txHash)
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
		SenderAddr: senderAddr,
		ChainID:    ip.chainIDDst,
		Receipts:   ip.receiversParsed,
		TxHash:     txHash,
	}, nil
}
