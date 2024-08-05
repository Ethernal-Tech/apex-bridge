package clisendtx

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
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
	nexusUrlFlag        = "nexus-url"

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
	nexusUrlFlagDesc        = "nexus chain URL"

	defaultFeeAmount = 1_100_000
	ttlSlotNumberInc = 500
)

type receiverAmount struct {
	ReceiverAddr string
	Amount       uint64
}

func ToTxOutput(receivers []receiverAmount) []cardanowallet.TxOutput {
	txOutputs := make([]cardanowallet.TxOutput, len(receivers))
	for idx, rec := range receivers {
		txOutputs[idx] = cardanowallet.TxOutput{
			Addr:   rec.ReceiverAddr,
			Amount: rec.Amount,
		}
	}
	return txOutputs
}

func ToGatewayStruct(receivers []receiverAmount) []contractbinding.IGatewayStructsReceiverWithdraw {
	gatewayOutputs := make([]contractbinding.IGatewayStructsReceiverWithdraw, len(receivers))
	for idx, rec := range receivers {
		gatewayOutputs[idx] = contractbinding.IGatewayStructsReceiverWithdraw{
			Receiver: rec.ReceiverAddr,
			Amount:   big.NewInt(int64(rec.Amount)),
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
	feeAmount     uint64

	// apex
	ogmiosURLSrc    string
	networkIDSrc    uint
	testnetMagicSrc uint
	multisigAddrSrc string
	ogmiosURLDst    string

	// nexus
	gatewayAddress string
	nexusUrl       string

	receiversParsed []receiverAmount
	wallet          cardanowallet.IWallet
}

func (ip *sendTxParams) validateFlags() error {
	if ip.txType != "" && ip.txType != "evm" && ip.txType != "cardano" {
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

	if ip.feeAmount < cardanowallet.MinUTxODefaultValue {
		return fmt.Errorf("--%s invalid amount: %d", feeAmountFlag, ip.feeAmount)
	}

	if ip.txType == "evm" {
		bytes, err := cardanowallet.GetKeyBytes(ip.privateKeyRaw)
		if err != nil || len(bytes) != 20 {
			return fmt.Errorf("invalid --%s value %s", privateKeyFlag, ip.privateKeyRaw)
		}

		if ip.gatewayAddress == "" {
			return fmt.Errorf("--%s not specified", gatewayAddressFlag)
		}

		if ip.nexusUrl == "" {
			return fmt.Errorf("--%s not specified", nexusUrlFlag)
		}
	} else {
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

		amount, err := strconv.ParseUint(vals[1], 0, 64)
		if err != nil {
			return fmt.Errorf("--%s number %d has invalid amount: %s", receiverFlag, i, x)
		}

		// TODO:nexus check cardano/evm address
		if amount < cardanowallet.MinUTxODefaultValue {
			return fmt.Errorf("--%s number %d has insufficient amount: %s", receiverFlag, i, x)
		}

		_, err = cardanowallet.NewAddress(vals[0])
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

	cmd.Flags().StringVar(
		&ip.txType,
		txTypeFlag,
		"",
		txTypeFlagDesc,
	)
}

func (ip *sendTxParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	if ip.txType == "evm" {
		return ip.executeEvm(outputter)
	} else if ip.txType == "cardano" || ip.txType == "" {
		return ip.executeCardano(outputter)
	}

	return nil, fmt.Errorf("")
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
		context.Background(), ip.chainIDDst, senderAddr.String(), receivers, ip.feeAmount)
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
	wallet, err := ethtxhelper.NewEthTxWallet(ip.privateKeyRaw)
	if err != nil {
		return nil, err
	}

	txHelper, err := ethtxhelper.NewEThTxHelper(
		ethtxhelper.WithNodeURL(ip.nexusUrl), ethtxhelper.WithGasFeeMultiplier(150), // TODO:nexus check fee params
		ethtxhelper.WithZeroGasPrice(false), ethtxhelper.WithDefaultGasLimit(0))
	if err != nil {
		return nil, err
	}

	contract, err := contractbinding.NewGateway(common.HexToAddress(ip.gatewayAddress), txHelper.GetClient())
	if err != nil {
		return nil, err
	}

	tx, err := txHelper.SendTx(context.Background(), wallet, bind.TransactOpts{},
		func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
			return contract.Withdraw(txOpts, 1, ToGatewayStruct(ip.receiversParsed), big.NewInt(150))
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
