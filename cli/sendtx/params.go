package clisendtx

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
)

const (
	privateKeyFlag      = "key"
	stakePrivateKeyFlag = "stake-key"
	ogmiosURLSrcFlag    = "ogmios-src"
	receiverFlag        = "receiver"
	networkIDSrcFlag    = "network-id-src"
	testnetMagicFlag    = "testnet-src"
	srcChainIDFlag      = "chain-src"
	dstChainIDFlag      = "chain-dst"
	multisigAddrSrcFlag = "addr-multisig-src"
	feeAmountFlag       = "fee"
	ogmiosURLDstFlag    = "ogmios-dst"
	txTypeFlag          = "tx-type"
	gatewayAddressFlag  = "gateway-addr"
	nexusURLFlag        = "nexus-url"
	currencyTokenIDFlag = "currency-token-id"
	chainIDsConfigFlag  = "chain-ids-config"

	privateKeyFlagDesc      = "wallet payment signing key"
	stakePrivateKeyFlagDesc = "wallet stake signing key"
	ogmiosURLSrcFlagDesc    = "source chain ogmios url"
	receiverFlagDesc        = "receiver addr:amount"
	testnetMagicFlagDesc    = "source testnet magic number. leave 0 for mainnet"
	networkIDSrcFlagDesc    = "source network id"
	srcChainIDFlagDesc      = "source chain ID (prime, vector, etc)"
	dstChainIDFlagDesc      = "destination chain ID (prime, vector, etc)"
	multisigAddrSrcFlagDesc = "source multisig address"
	feeAmountFlagDesc       = "amount for multisig fee addr"
	ogmiosURLDstFlagDesc    = "destination chain ogmios url"
	txTypeFlagDesc          = "type of transaction (evm, default: cardano)"
	gatewayAddressFlagDesc  = "address of gateway contract"
	nexusURLFlagDesc        = "nexus chain URL"
	currencyTokenIDFlagDesc = "currency token ID on evm chain"
	chainIDsConfigFlagDesc  = "path to the chain IDs config file"

	ttlSlotNumberInc = 500

	gasLimitMultiplier = 1.6
	potentialFee       = 500_000

	waitForAmountRetryCount = 240 // 240 * 10 = 40 min
	waitForAmountWaitTime   = time.Second * 10
)

var minNexusBridgingFee = new(big.Int).SetUint64(1000010000000000000)

const nexusCurrencyTokenID = 1

type receiverAmount struct {
	ReceiverAddr string
	Amount       *big.Int
}

type sendTxParams struct {
	txType string // cardano or evm

	// common
	privateKeyRaw      string
	stakePrivateKeyRaw string
	receivers          []string
	chainIDSrc         string
	chainIDDst         string
	feeString          string
	chainIDsConfig     string

	// apex
	ogmiosURLSrc    string
	networkIDSrc    uint
	testnetMagicSrc uint
	multisigAddrSrc string
	ogmiosURLDst    string

	// nexus
	gatewayAddress  string
	nexusURL        string
	currencyTokenID uint16

	feeAmount        *big.Int
	receiversParsed  []*receiverAmount
	wallet           *cardanowallet.Wallet
	chainIDConverter *common.ChainIDConverter
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

	if ip.chainIDsConfig == "" {
		return fmt.Errorf("--%s flag not specified", chainIDsConfigFlag)
	}

	if _, err := os.Stat(ip.chainIDsConfig); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config file does not exist: %s", ip.chainIDsConfig)
		}

		return fmt.Errorf("failed to check config file: %s. err: %w", ip.chainIDsConfig, err)
	}

	chainIDsConfig, err := common.LoadConfig[common.ChainIDsConfig](ip.chainIDsConfig, "")
	if err != nil {
		return fmt.Errorf("failed to load chain IDs config: %w", err)
	}

	ip.chainIDConverter = chainIDsConfig.ToChainIDConverter()

	if !ip.chainIDConverter.IsExistingChainID(ip.chainIDSrc) {
		return fmt.Errorf("--%s flag not specified", srcChainIDFlag)
	}

	if !ip.chainIDConverter.IsExistingChainID(ip.chainIDDst) {
		return fmt.Errorf("--%s flag not specified", dstChainIDFlag)
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
		if ip.feeAmount.Uint64() < common.MinFeeForBridgingDefault {
			return fmt.Errorf("--%s invalid amount: %d", feeAmountFlag, ip.feeAmount)
		}

		bytes, err := cardanotx.GetCardanoPrivateKeyBytes(ip.privateKeyRaw)
		if err != nil {
			return fmt.Errorf("invalid --%s value %s", privateKeyFlag, ip.privateKeyRaw)
		}

		var stakeBytes []byte
		if len(ip.stakePrivateKeyRaw) > 0 {
			stakeBytes, err = cardanotx.GetCardanoPrivateKeyBytes(ip.stakePrivateKeyRaw)
			if err != nil {
				return fmt.Errorf("invalid --%s value %s", stakePrivateKeyFlag, ip.stakePrivateKeyRaw)
			}
		}

		ip.wallet = cardanowallet.NewWallet(bytes, stakeBytes)

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

		if !common.IsValidAddress(ip.chainIDDst, vals[0], ip.chainIDConverter) {
			return fmt.Errorf("--%s number %d has invalid address: %s", receiverFlag, i, x)
		}

		if ip.chainIDDst != common.ChainIDStrNexus &&
			amount.Cmp(new(big.Int).SetUint64(common.MinUtxoAmountDefault)) < 0 {
			return fmt.Errorf("--%s number %d has insufficient amount: %s", receiverFlag, i, x)
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

	cmd.Flags().StringVar(
		&ip.stakePrivateKeyRaw,
		stakePrivateKeyFlag,
		"",
		stakePrivateKeyFlagDesc,
	)

	cmd.Flags().StringArrayVar(
		&ip.receivers,
		receiverFlag,
		nil,
		receiverFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.chainIDSrc,
		srcChainIDFlag,
		"",
		srcChainIDFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.chainIDDst,
		dstChainIDFlag,
		"",
		dstChainIDFlagDesc,
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

	cmd.Flags().Uint16Var(
		&ip.currencyTokenID,
		currencyTokenIDFlag,
		nexusCurrencyTokenID,
		currencyTokenIDFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.chainIDsConfig,
		chainIDsConfigFlag,
		"",
		chainIDsConfigFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(gatewayAddressFlag, testnetMagicFlag)
	cmd.MarkFlagsMutuallyExclusive(gatewayAddressFlag, networkIDSrcFlag)
	cmd.MarkFlagsMutuallyExclusive(gatewayAddressFlag, ogmiosURLSrcFlag)
}

func (ip *sendTxParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()

	switch ip.txType {
	case common.ChainTypeEVMStr:
		return ip.executeEvm(ctx, outputter)
	case common.ChainTypeCardanoStr, "":
		return ip.executeCardano(ctx, outputter)
	default:
		return nil, fmt.Errorf("txType not supported")
	}
}

func (ip *sendTxParams) executeCardano(ctx context.Context, outputter common.OutputFormatter) (
	common.ICommandResult, error,
) {
	receivers := toCardanoMetadata(ip.receiversParsed)
	networkID := cardanowallet.CardanoNetworkType(ip.networkIDSrc)
	txSender := sendtx.NewTxSender(
		map[string]sendtx.ChainConfig{
			ip.chainIDSrc: {
				CardanoCliBinary:         cardanowallet.ResolveCardanoCliBinary(networkID),
				TxProvider:               cardanowallet.NewTxProviderOgmios(ip.ogmiosURLSrc),
				TestNetMagic:             ip.testnetMagicSrc,
				TTLSlotNumberInc:         ttlSlotNumberInc,
				DefaultMinFeeForBridging: common.MinFeeForBridgingDefault,
				MinFeeForBridgingTokens:  common.MinFeeForBridgingDefault,
				MinUtxoValue:             common.MinUtxoAmountDefault,
				PotentialFee:             potentialFee,
				Tokens: map[uint16]sendtx.ApexToken{
					0: {FullName: cardanowallet.AdaTokenName},
				},
			},
			ip.chainIDDst: {
				TxProvider:               cardanowallet.NewTxProviderOgmios(ip.ogmiosURLDst),
				DefaultMinFeeForBridging: common.MinFeeForBridgingDefault,
				MinFeeForBridgingTokens:  common.MinFeeForBridgingDefault,
				PotentialFee:             potentialFee,
			},
		},
		sendtx.WithMinAmountToBridge(common.MinUtxoAmountDefault),
	)

	senderAddr, err := cardanotx.GetAddress(networkID, ip.wallet)
	if err != nil {
		return nil, err
	}

	txInfo, _, err := txSender.CreateBridgingTx(
		ctx,
		sendtx.BridgingTxDto{
			SrcChainID:      ip.chainIDSrc,
			DstChainID:      ip.chainIDDst,
			SenderAddr:      senderAddr.String(),
			Receivers:       receivers,
			BridgingAddress: ip.multisigAddrSrc,
			BridgingFee:     ip.feeAmount.Uint64(),
			OperationFee:    0,
		},
	)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Submiting bridging transaction..."))
	outputter.WriteOutput()

	err = txSender.SubmitTx(ctx, ip.chainIDSrc, txInfo.TxRaw, ip.wallet)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte(fmt.Sprintf("transaction has been submitted: %s", txInfo.TxHash)))
	outputter.WriteOutput()

	if ip.ogmiosURLDst != "" {
		err = waitForTxOnCardano(
			ctx, cardanowallet.NewTxProviderOgmios(ip.ogmiosURLDst),
			ip.receiversParsed)
		if err != nil {
			return nil, err
		}
	} else if ip.nexusURL != "" {
		txHelper, err := getTxHelper(ip.nexusURL)
		if err != nil {
			return nil, err
		}

		receivers := make([]*receiverAmount, len(ip.receiversParsed))
		for i, rec := range ip.receiversParsed {
			receivers[i] = &receiverAmount{
				ReceiverAddr: rec.ReceiverAddr,
				Amount:       common.DfmToWei(rec.Amount),
			}
		}

		err = waitForTxOnEvm(ctx, txHelper.GetClient(), receivers)
		if err != nil {
			return nil, err
		}
	}

	_, _ = outputter.Write([]byte("Transaction has been bridged"))
	outputter.WriteOutput()

	return CmdResult{
		SenderAddr: senderAddr.String(),
		ChainID:    ip.chainIDDst,
		Receipts:   ip.receiversParsed,
		TxHash:     txInfo.TxHash,
	}, nil
}

func (ip *sendTxParams) executeEvm(ctx context.Context, outputter common.OutputFormatter) (
	common.ICommandResult, error,
) {
	contractAddress := common.HexToAddress(ip.gatewayAddress)
	chainID := ip.chainIDConverter.ToNumChainID(ip.chainIDDst)
	receivers, totalAmount := toGatewayStruct(ip.receiversParsed, ip.currencyTokenID)
	totalAmount.Add(totalAmount, ip.feeAmount)

	minOperationFee := common.DfmToWei(new(big.Int).SetUint64(common.MinOperationFeeDefault))

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
		ctx, wallet.GetAddress(), contractAddress, totalAmount, gasLimitMultiplier,
		abi, "withdraw", chainID, receivers, ip.feeAmount, minOperationFee)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Submiting bridging transaction..."))
	outputter.WriteOutput()

	tx, err := txHelper.SendTx(ctx, wallet,
		bind.TransactOpts{
			GasLimit: estimatedGas,
			Value:    totalAmount,
		},
		func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
			return contract.Withdraw(
				txOpts, chainID, receivers, ip.feeAmount, minOperationFee,
			)
		})
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte(fmt.Sprintf("transaction has been submitted: %s", tx.Hash())))
	outputter.WriteOutput()

	receipt, err := txHelper.WaitForReceipt(context.Background(), tx.Hash().String())
	if err != nil {
		return nil, err
	} else if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, errors.New("transaction receipt status is unsuccessful")
	}

	if ip.ogmiosURLDst != "" {
		receivers := make([]*receiverAmount, len(ip.receiversParsed))
		for i, rec := range ip.receiversParsed {
			receivers[i] = &receiverAmount{
				ReceiverAddr: rec.ReceiverAddr,
				Amount:       common.WeiToDfm(rec.Amount),
			}
		}

		err = waitForTxOnCardano(
			ctx, cardanowallet.NewTxProviderOgmios(ip.ogmiosURLDst),
			receivers)
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

func waitForTxOnEvm(
	ctx context.Context, client *ethclient.Client, receivers []*receiverAmount) error {
	return waitForTx(ctx, receivers, func(ctx context.Context, addr string) (*big.Int, error) {
		return client.BalanceAt(ctx, common.HexToAddress(addr), nil)
	})
}

func waitForTxOnCardano(
	ctx context.Context, txUtxoRetriever cardanowallet.IUTxORetriever, receivers []*receiverAmount) error {
	return waitForTx(ctx, receivers, func(ctx context.Context, addr string) (*big.Int, error) {
		utxos, err := txUtxoRetriever.GetUtxos(ctx, addr)
		if err != nil {
			return nil, err
		}

		sum := cardanowallet.GetUtxosSum(utxos)

		return new(big.Int).SetUint64(sum[cardanowallet.AdaTokenName]), nil
	})
}

func waitForTx(ctx context.Context, receivers []*receiverAmount,
	getBalanceFn func(ctx context.Context, addr string) (*big.Int, error),
) error {
	errs := make([]error, len(receivers))
	wg := sync.WaitGroup{}

	for i, x := range receivers {
		wg.Add(1)

		go func(idx int, recv *receiverAmount) {
			defer wg.Done()

			_, errs[idx] = common.WaitForAmount(ctx, recv.Amount, func(ctx context.Context) (*big.Int, error) {
				return getBalanceFn(ctx, recv.ReceiverAddr)
			}, infracommon.WithRetryCount(waitForAmountRetryCount),
				infracommon.WithRetryWaitTime(waitForAmountWaitTime))
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

func toCardanoMetadata(receivers []*receiverAmount) []sendtx.BridgingTxReceiver {
	metadataReceivers := make([]sendtx.BridgingTxReceiver, len(receivers))
	for idx, rec := range receivers {
		metadataReceivers[idx] = sendtx.BridgingTxReceiver{
			Addr:    rec.ReceiverAddr,
			Amount:  rec.Amount.Uint64(),
			TokenID: 0,
		}
	}

	return metadataReceivers
}

func toGatewayStruct(receivers []*receiverAmount, currencyTokenID uint16) (
	[]contractbinding.IGatewayStructsReceiverWithdraw, *big.Int,
) {
	total := big.NewInt(0)

	gatewayOutputs := make([]contractbinding.IGatewayStructsReceiverWithdraw, len(receivers))
	for idx, rec := range receivers {
		gatewayOutputs[idx] = contractbinding.IGatewayStructsReceiverWithdraw{
			Receiver: rec.ReceiverAddr,
			Amount:   rec.Amount,
			TokenId:  currencyTokenID,
		}

		total.Add(total, rec.Amount)
	}

	return gatewayOutputs, total
}
