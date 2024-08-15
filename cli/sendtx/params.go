package clisendtx

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
)

const (
	privateKeyFlag      = "key"
	ogmiosURLSrcFlag    = "ogmios-src"
	receiverFlag        = "receiver"
	networkIDSrcFlag    = "network-id-src"
	testnetMagicFlag    = "testnet-src"
	chainIDFlag         = "chain-dst"
	multisigAddrSrcFlag = "addr-multisig-src"
	feeAmountFlag       = "fee"
	ogmiosURLDstFlag    = "ogmios-dst"
	txTypeFlag          = "tx-type"
	gatewayAddressFlag  = "gateway-addr"
	nexusURLFlag        = "nexus-url"

	privateKeyFlagDesc      = "wallet private signing key"
	ogmiosURLSrcFlagDesc    = "source chain ogmios url"
	receiverFlagDesc        = "receiver addr:amount"
	testnetMagicFlagDesc    = "source testnet magic number. leave 0 for mainnet"
	networkIDSrcFlagDesc    = "source network id"
	chainIDFlagDesc         = "destination chain ID (prime, vector, etc)"
	multisigAddrSrcFlagDesc = "source multisig address"
	feeAmountFlagDesc       = "amount for multisig fee addr"
	ogmiosURLDstFlagDesc    = "destination chain ogmios url"
	txTypeFlagDesc          = "type of transaction (evm, default: cardano)"
	gatewayAddressFlagDesc  = "address of gateway contract"
	nexusURLFlagDesc        = "nexus chain URL"

	defaultFeeAmount = 1_100_000
	ttlSlotNumberInc = 500
)

var minNexusBridgingFee = new(big.Int).SetUint64(1000010000000000000)

type receiverAmount struct {
	ReceiverAddr string
	Amount       *big.Int
}

func ToTxOutput(receivers []receiverAmount) []cardanowallet.TxOutput {
	txOutputs := make([]cardanowallet.TxOutput, len(receivers))
	for idx, rec := range receivers {
		txOutputs[idx] = cardanowallet.TxOutput{
			Addr:   rec.ReceiverAddr,
			Amount: rec.Amount.Uint64(),
		}
	}

	return txOutputs
}

func ToGatewayStruct(receivers []receiverAmount) []contractbinding.IGatewayStructsReceiverWithdraw {
	gatewayOutputs := make([]contractbinding.IGatewayStructsReceiverWithdraw, len(receivers))
	for idx, rec := range receivers {
		gatewayOutputs[idx] = contractbinding.IGatewayStructsReceiverWithdraw{
			Receiver: rec.ReceiverAddr,
			Amount:   rec.Amount,
		}
	}

	return gatewayOutputs
}

type sendTxParams struct {
	txType string // cardano or evm

	// common
	privateKeyRaw string
	receivers     []string
	chainIDDst    string
	feeString     string

	// apex
	ogmiosURLSrc    string
	networkIDSrc    uint
	testnetMagicSrc uint
	multisigAddrSrc string
	ogmiosURLDst    string

	// nexus
	gatewayAddress string
	nexusURL       string

	feeAmount       *big.Int
	receiversParsed []receiverAmount
	wallet          cardanowallet.IWallet
}

func (ip *sendTxParams) validateFlags() error {
	if ip.txType != "" && ip.txType != common.ChainTypeEVMStr && ip.txType != common.ChainTypeCardanoStr {
		return fmt.Errorf("invalid --%s type not supported", txTypeFlag)
	}

	if ip.privateKeyRaw == "" {
		return fmt.Errorf("flag --%s not specified", privateKeyFlag)
	}

	if len(ip.receivers) == 0 {
		return fmt.Errorf("--%s not specified", receiverFlag)
	}

	if ip.chainIDDst == "" {
		return fmt.Errorf("--%s flag not specified", chainIDFlag)
	}

	feeAmount, ok := new(big.Int).SetString(ip.feeString, 0)
	if !ok {
		return fmt.Errorf("--%s invalid amount: %s", feeAmountFlag, ip.feeString)
	}

	ip.feeAmount = feeAmount

	if ip.txType == common.ChainTypeEVMStr {
		if ip.feeAmount.Cmp(minNexusBridgingFee) < 0 {
			return fmt.Errorf("--%s invalid amount: %d", feeAmountFlag, ip.feeAmount)
		}

		if ip.gatewayAddress == "" {
			return fmt.Errorf("--%s not specified", gatewayAddressFlag)
		}

		if ip.nexusURL == "" {
			return fmt.Errorf("--%s not specified", nexusURLFlag)
		}
	} else {
		if ip.feeAmount.Uint64() < cardanowallet.MinUTxODefaultValue {
			return fmt.Errorf("--%s invalid amount: %d", feeAmountFlag, ip.feeAmount)
		}

		bytes, err := cardanowallet.GetKeyBytes(ip.privateKeyRaw)
		if err != nil || len(bytes) != 32 {
			return fmt.Errorf("invalid --%s value %s", privateKeyFlag, ip.privateKeyRaw)
		}

		ip.wallet = cardanowallet.NewWallet(cardanowallet.GetVerificationKeyFromSigningKey(bytes), bytes)

		if ip.ogmiosURLSrc == "" || !common.IsValidURL(ip.ogmiosURLSrc) {
			return fmt.Errorf("invalid --%s: %s", ogmiosURLSrcFlag, ip.ogmiosURLSrc)
		}

		if ip.multisigAddrSrc == "" {
			return fmt.Errorf("--%s not specified", multisigAddrSrcFlag)
		}

		if ip.ogmiosURLDst != "" && !common.IsValidURL(ip.ogmiosURLDst) {
			return fmt.Errorf("invalid --%s: %s", ogmiosURLDstFlag, ip.ogmiosURLDst)
		}
	}

	receivers := make([]receiverAmount, 0, len(ip.receivers))

	for i, x := range ip.receivers {
		vals := strings.Split(x, ":")
		if len(vals) != 2 {
			return fmt.Errorf("--%s number %d is invalid: %s", receiverFlag, i, x)
		}

		amount, ok := new(big.Int).SetString(vals[1], 0)
		if !ok {
			return fmt.Errorf("--%s number %d has invalid amount: %s", receiverFlag, i, x)
		}

		if ip.txType != common.ChainTypeEVMStr {
			if amount.Uint64() < cardanowallet.MinUTxODefaultValue {
				return fmt.Errorf("--%s number %d has insufficient amount: %s", receiverFlag, i, x)
			}
		}

		_, err := cardanowallet.NewAddress(vals[0])
		if err != nil {
			return fmt.Errorf("--%s number %d has invalid address: %s", receiverFlag, i, x)
		}

		receivers = append(receivers, receiverAmount{
			ReceiverAddr: vals[0],
			Amount:       amount,
		})
	}

	ip.receiversParsed = receivers

	return nil
}

func (ip *sendTxParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&ip.txType,
		txTypeFlag,
		"",
		txTypeFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.privateKeyRaw,
		privateKeyFlag,
		"",
		privateKeyFlagDesc,
	)

	cmd.Flags().StringArrayVar(
		&ip.receivers,
		receiverFlag,
		nil,
		receiverFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.chainIDDst,
		chainIDFlag,
		"",
		chainIDFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.feeString,
		feeAmountFlag,
		"0",
		feeAmountFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.ogmiosURLSrc,
		ogmiosURLSrcFlag,
		"",
		ogmiosURLSrcFlagDesc,
	)

	cmd.Flags().UintVar(
		&ip.networkIDSrc,
		networkIDSrcFlag,
		0,
		networkIDSrcFlagDesc,
	)

	cmd.Flags().UintVar(
		&ip.testnetMagicSrc,
		testnetMagicFlag,
		0,
		testnetMagicFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.multisigAddrSrc,
		multisigAddrSrcFlag,
		"",
		multisigAddrSrcFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.ogmiosURLDst,
		ogmiosURLDstFlag,
		"",
		ogmiosURLDstFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.gatewayAddress,
		gatewayAddressFlag,
		"",
		gatewayAddressFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.nexusURL,
		nexusURLFlag,
		"",
		nexusURLFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(gatewayAddressFlag, testnetMagicFlag)
	cmd.MarkFlagsMutuallyExclusive(gatewayAddressFlag, networkIDSrcFlag)
	cmd.MarkFlagsMutuallyExclusive(gatewayAddressFlag, ogmiosURLSrcFlag)
	cmd.MarkFlagsMutuallyExclusive(gatewayAddressFlag, ogmiosURLDstFlag)
	cmd.MarkFlagsMutuallyExclusive(nexusURLFlag, testnetMagicFlag)
	cmd.MarkFlagsMutuallyExclusive(nexusURLFlag, networkIDSrcFlag)
	cmd.MarkFlagsMutuallyExclusive(nexusURLFlag, ogmiosURLSrcFlag)
	cmd.MarkFlagsMutuallyExclusive(nexusURLFlag, ogmiosURLDstFlag)
}

func (ip *sendTxParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	switch ip.txType {
	case common.ChainTypeEVMStr:
		return ip.executeEvm(outputter)
	case common.ChainTypeCardanoStr, "":
		return ip.executeCardano(outputter)
	default:
		return nil, fmt.Errorf("txType not supported")
	}
}

func (ip *sendTxParams) executeCardano(outputter common.OutputFormatter) (common.ICommandResult, error) {
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

	receivers := ToTxOutput(ip.receiversParsed)

	txRaw, txHash, err := txSender.CreateTx(
		context.Background(), ip.chainIDDst, senderAddr.String(), receivers, ip.feeAmount.Uint64())
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
		err = txSender.WaitForTx(context.Background(), receivers)
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

func (ip *sendTxParams) executeEvm(outputter common.OutputFormatter) (common.ICommandResult, error) {
	wallet, err := ethtxhelper.NewEthTxWalletFromPk(ip.privateKeyRaw)
	if err != nil {
		return nil, err
	}

	txHelper, err := ethtxhelper.NewEThTxHelper(
		//nolint:godox
		// TODO: Estimate gas manually until https://github.com/ethereum/go-ethereum/issues/29798 is implemented
		ethtxhelper.WithNodeURL(ip.nexusURL), ethtxhelper.WithGasFeeMultiplier(150),
		ethtxhelper.WithZeroGasPrice(false), ethtxhelper.WithDefaultGasLimit(0))
	if err != nil {
		return nil, err
	}

	contract, err := contractbinding.NewGateway(common.HexToAddress(ip.gatewayAddress), txHelper.GetClient())
	if err != nil {
		return nil, err
	}

	tx, err := txHelper.SendTx(context.Background(), wallet, bind.TransactOpts{
		From: wallet.GetAddress(),
	},
		func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
			return contract.Withdraw(
				txOpts,
				common.ToNumChainID(ip.chainIDDst),
				ToGatewayStruct(ip.receiversParsed),
				ip.feeAmount,
			)
		})
	if err != nil {
		return nil, err
	}

	receipt, err := txHelper.WaitForReceipt(context.Background(), tx.Hash().String(), true)
	if types.ReceiptStatusSuccessful != receipt.Status {
		return nil, err
	}

	_, _ = outputter.Write([]byte(fmt.Sprintf("transaction has been submitted: %s", receipt.TxHash.String())))
	outputter.WriteOutput()

	return CmdResult{
		SenderAddr: wallet.GetAddress().String(),
		ChainID:    ip.chainIDDst,
		Receipts:   ip.receiversParsed,
		TxHash:     receipt.TxHash.String(),
	}, nil
}
