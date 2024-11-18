package clisendtx

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
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

	gasLimitMultiplier       = 1.6
	amountCheckRetryWaitTime = time.Second * 5
	amountCheckRetryCount    = 144 // 12 minutes = 5 seconds * 144
)

var minNexusBridgingFee = new(big.Int).SetUint64(1000010000000000000)

type receiverAmount struct {
	ReceiverAddr string
	Amount       *big.Int
}

func ToTxOutput(receivers []*receiverAmount) []cardanowallet.TxOutput {
	txOutputs := make([]cardanowallet.TxOutput, len(receivers))
	for idx, rec := range receivers {
		txOutputs[idx] = cardanowallet.TxOutput{
			Addr:   rec.ReceiverAddr,
			Amount: rec.Amount.Uint64(),
		}
	}

	return txOutputs
}

func ToGatewayStruct(receivers []*receiverAmount) ([]contractbinding.IGatewayStructsReceiverWithdraw, *big.Int) {
	total := big.NewInt(0)

	gatewayOutputs := make([]contractbinding.IGatewayStructsReceiverWithdraw, len(receivers))
	for idx, rec := range receivers {
		gatewayOutputs[idx] = contractbinding.IGatewayStructsReceiverWithdraw{
			Receiver: rec.ReceiverAddr,
			Amount:   rec.Amount,
		}

		total.Add(total, rec.Amount)
	}

	return gatewayOutputs, total
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
	receiversParsed []*receiverAmount
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

	if !common.IsExistingChainID(ip.chainIDDst) {
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

		if !common.IsValidHTTPURL(ip.nexusURL) {
			return fmt.Errorf("invalid --%s flag", nexusURLFlag)
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

		if !common.IsValidHTTPURL(ip.ogmiosURLSrc) {
			return fmt.Errorf("invalid --%s: %s", ogmiosURLSrcFlag, ip.ogmiosURLSrc)
		}

		if ip.multisigAddrSrc == "" {
			return fmt.Errorf("--%s not specified", multisigAddrSrcFlag)
		}

		if ip.nexusURL == "" && ip.ogmiosURLDst == "" {
			return fmt.Errorf("--%s and --%s not specified", ogmiosURLDstFlag, nexusURLFlag)
		}

		if ip.ogmiosURLDst != "" && !common.IsValidHTTPURL(ip.ogmiosURLDst) {
			return fmt.Errorf("invalid --%s: %s", ogmiosURLDstFlag, ip.ogmiosURLDst)
		}

		if ip.nexusURL != "" && !common.IsValidHTTPURL(ip.nexusURL) {
			return fmt.Errorf("invalid --%s: %s", nexusURLFlag, ip.nexusURL)
		}
	}

	receivers := make([]*receiverAmount, 0, len(ip.receivers))

	for i, x := range ip.receivers {
		vals := strings.Split(x, ":")
		if len(vals) != 2 {
			return fmt.Errorf("--%s number %d is invalid: %s", receiverFlag, i, x)
		}

		amount, ok := new(big.Int).SetString(vals[1], 0)
		if !ok {
			return fmt.Errorf("--%s number %d has invalid amount: %s", receiverFlag, i, x)
		}

		switch ip.chainIDDst {
		case common.ChainIDStrNexus:
			if !ethcommon.IsHexAddress(vals[0]) {
				return fmt.Errorf("--%s number %d has invalid address: %s", receiverFlag, i, x)
			}
		default:
			if amount.Uint64() < cardanowallet.MinUTxODefaultValue {
				return fmt.Errorf("--%s number %d has insufficient amount: %s", receiverFlag, i, x)
			}

			_, err := cardanowallet.NewAddress(vals[0])
			if err != nil {
				return fmt.Errorf("--%s number %d has invalid address: %s", receiverFlag, i, x)
			}
		}

		receivers = append(receivers, &receiverAmount{
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
	receivers := ToTxOutput(ip.receiversParsed)
	networkID := cardanowallet.CardanoNetworkType(ip.networkIDSrc)
	cardanoCliBinary := cardanowallet.ResolveCardanoCliBinary(networkID)
	txSender := cardanotx.NewBridgingTxSender(
		cardanoCliBinary,
		cardanowallet.NewTxProviderOgmios(ip.ogmiosURLSrc),
		cardanowallet.NewTxProviderOgmios(ip.ogmiosURLDst),
		ip.testnetMagicSrc, ip.multisigAddrSrc, ttlSlotNumberInc, cardanotx.DefaultPotentialFee)

	senderAddr, err := cardanotx.GetAddress(networkID, ip.wallet)
	if err != nil {
		return nil, err
	}

	txRaw, txHash, err := txSender.CreateTx(
		context.Background(), ip.chainIDDst, senderAddr.String(), receivers, ip.feeAmount.Uint64())
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Submiting bridging transaction..."))
	outputter.WriteOutput()

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

		_, _ = outputter.Write([]byte("Transaction has been bridged"))
		outputter.WriteOutput()
	} else if ip.nexusURL != "" {
		txHelper, err := getTxHelper(ip.nexusURL)
		if err != nil {
			return nil, err
		}

		err = waitForAmounts(context.Background(), txHelper.GetClient(), ip.receiversParsed)
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
	contractAddress := common.HexToAddress(ip.gatewayAddress)
	chainID := common.ToNumChainID(ip.chainIDDst)
	receivers, totalAmount := ToGatewayStruct(ip.receiversParsed)
	totalAmount.Add(totalAmount, ip.feeAmount)

	wallet, err := ethtxhelper.NewEthTxWallet(ip.privateKeyRaw)
	if err != nil {
		return nil, err
	}

	txHelper, err := getTxHelper(ip.nexusURL)
	if err != nil {
		return nil, err
	}

	contract, err := contractbinding.NewGateway(contractAddress, txHelper.GetClient())
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Estimating gas..."))
	outputter.WriteOutput()

	abi, err := contractbinding.GatewayMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	estimatedGas, _, err := txHelper.EstimateGas(
		context.Background(), wallet.GetAddress(), contractAddress, totalAmount, gasLimitMultiplier,
		abi, "withdraw", chainID, receivers, ip.feeAmount)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Submiting bridging transaction..."))
	outputter.WriteOutput()

	tx, err := txHelper.SendTx(context.Background(), wallet,
		bind.TransactOpts{
			GasLimit: estimatedGas,
			Value:    totalAmount,
		},
		func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
			return contract.Withdraw(
				txOpts, chainID, receivers, ip.feeAmount,
			)
		})
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte(fmt.Sprintf("transaction has been submitted: %s", tx.Hash())))
	outputter.WriteOutput()

	receipt, err := txHelper.WaitForReceipt(context.Background(), tx.Hash().String(), true)
	if err != nil {
		return nil, err
	} else if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, errors.New("transaction receipt status is unsuccessful")
	}

	if ip.ogmiosURLDst != "" {
		cardanoReceivers := ToTxOutput(ip.receiversParsed)
		for i := range cardanoReceivers {
			cardanoReceivers[i].Amount = 1 // just need to see if there is some change
		}

		err = cardanotx.WaitForTx(
			context.Background(), cardanowallet.NewTxProviderOgmios(ip.ogmiosURLDst), cardanoReceivers)
		if err != nil {
			return nil, err
		}

		_, _ = outputter.Write([]byte("Transaction has been bridged"))
		outputter.WriteOutput()
	}

	return CmdResult{
		SenderAddr: wallet.GetAddress().String(),
		ChainID:    ip.chainIDDst,
		Receipts:   ip.receiversParsed,
		TxHash:     receipt.TxHash.String(),
	}, nil
}

func waitForAmounts(ctx context.Context, client *ethclient.Client, receivers []*receiverAmount) error {
	errs := make([]error, len(receivers))
	wg := sync.WaitGroup{}

	for i, x := range receivers {
		wg.Add(1)

		go func(idx int, recv *receiverAmount) {
			defer wg.Done()

			var (
				addr = common.HexToAddress(recv.ReceiverAddr)
			)

			oldBalance, err := infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (*big.Int, error) {
				return client.BalanceAt(ctx, addr, nil)
			}, infracommon.WithIsRetryableError(ethtxhelper.IsRetryableEthError))
			if err != nil {
				errs[idx] = err

				return
			}

			expectedBalance := oldBalance.Add(oldBalance, recv.Amount)

			_, errs[idx] = infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (bool, error) {
				balance, err := client.BalanceAt(ctx, addr, nil)
				if err != nil {
					return false, err
				}

				if balance.Cmp(expectedBalance) < 0 {
					return false, infracommon.ErrRetryTryAgain
				}

				return true, nil
			},
				infracommon.WithIsRetryableError(ethtxhelper.IsRetryableEthError),
				infracommon.WithRetryCount(amountCheckRetryCount),
				infracommon.WithRetryWaitTime(amountCheckRetryWaitTime))
		}(i, x)
	}

	wg.Wait()

	return errors.Join(errs...)
}

func getTxHelper(nexusURL string) (*ethtxhelper.EthTxHelperImpl, error) {
	return ethtxhelper.NewEThTxHelper(
		ethtxhelper.WithNodeURL(nexusURL), ethtxhelper.WithGasFeeMultiplier(150),
		ethtxhelper.WithZeroGasPrice(false), ethtxhelper.WithDefaultGasLimit(0))
}
