package clisendtx

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
)

const (
	operationFeeFlag                  = "operation-fee"
	fullSrcTokenNameFlag              = "src-token-name"          //nolint:gosec
	fullDstTokenNameFlag              = "dst-token-name"          //nolint:gosec
	tokenIDSrcFlag                    = "src-token-id"            //nolint:gosec
	tokenContractAddrSrcFlag          = "src-token-contract-addr" //nolint:gosec
	tokenContractAddrDstFlag          = "dst-token-contract-addr" //nolint:gosec
	nativeTokenWalletContractAddrFlag = "native-token-wallet-contract-addr"

	operationFeeFlagDesc                  = "operation fee"
	fullSrcTokenNameFlagDesc              = "denom of the token to transfer from source chain"    //nolint:gosec
	fullDstTokenNameFlagDesc              = "denom of the token to transfer to destination chain" //nolint:gosec
	tokenIDSrcFlagDesc                    = "token id from source chain"
	tokenContractAddrSrcFlagDesc          = "contract address of the token on src"
	tokenContractAddrDstFlagDesc          = "contract address of the token on destination"
	nativeTokenWalletContractAddrFlagDesc = "address of native token wallet contract"

	apexTokenID = uint16(1)
	adaTokenID  = uint16(2)
)

type sendSkylineTxParams struct {
	txType string // cardano or evm

	privateKeyRaw      string
	stakePrivateKeyRaw string
	receivers          []string
	chainIDSrc         string
	chainIDDst         string
	feeString          string
	operationFeeString string
	tokenIDSrc         uint16
	tokenFullNameSrc   string
	tokenFullNameDst   string

	ogmiosURLSrc    string
	networkIDSrc    uint
	testnetMagicSrc uint
	multisigAddrSrc string
	treasuryAddrSrc string
	ogmiosURLDst    string

	// nexus
	gatewayAddress                   string
	nativeTokenWalletContractAddress string
	nexusURL                         string
	tokenContractAddrSrc             string
	tokenContractAddrDst             string

	feeAmount          *big.Int
	operationFeeAmount *big.Int
	receiversParsed    []*receiverAmount
	wallet             *cardanowallet.Wallet
}

func (p *sendSkylineTxParams) validateFlags() error {
	if p.txType != "" && p.txType != common.ChainTypeEVMStr && p.txType != common.ChainTypeCardanoStr {
		return fmt.Errorf("invalid --%s type not supported", txTypeFlag)
	}

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

	if p.tokenFullNameSrc == "" {
		return fmt.Errorf("--%s flag not specified", fullSrcTokenNameFlag)
	}

	if p.tokenFullNameSrc != cardanowallet.AdaTokenName {
		token, err := cardanowallet.NewTokenWithFullNameTry(p.tokenFullNameSrc)
		if err != nil {
			return fmt.Errorf("--%s invalid token name: %s", fullSrcTokenNameFlag, p.tokenFullNameSrc)
		}

		p.tokenFullNameSrc = token.String()
	}

	if p.tokenFullNameDst == "" {
		return fmt.Errorf("--%s flag not specified", fullDstTokenNameFlag)
	}

	if p.tokenFullNameDst != cardanowallet.AdaTokenName {
		token, err := cardanowallet.NewTokenWithFullNameTry(p.tokenFullNameDst)
		if err != nil {
			return fmt.Errorf("--%s invalid token name: %s", fullDstTokenNameFlag, p.tokenFullNameDst)
		}

		p.tokenFullNameDst = token.String()
	}

	if p.gatewayAddress != "" &&
		!common.IsValidAddress(common.ChainIDStrNexus, p.gatewayAddress) {
		return fmt.Errorf("invalid address for flag --%s", gatewayAddressFlag)
	}

	if p.tokenContractAddrSrc != "" && !common.IsValidAddress(common.ChainIDStrNexus, p.tokenContractAddrSrc) {
		return fmt.Errorf("invalid address for flag --%s", tokenContractAddrDstFlag)
	}

	if p.tokenContractAddrDst != "" && !common.IsValidAddress(common.ChainIDStrNexus, p.tokenContractAddrDst) {
		return fmt.Errorf("invalid address for flag --%s", tokenContractAddrDstFlag)
	}

	if p.nativeTokenWalletContractAddress != "" &&
		!common.IsValidAddress(common.ChainIDStrNexus, p.nativeTokenWalletContractAddress) {
		return fmt.Errorf("invalid address for flag --%s", nativeTokenWalletContractAddrFlag)
	}

	feeAmount, ok := new(big.Int).SetString(p.feeString, 0)
	if !ok {
		return fmt.Errorf("--%s invalid amount: %s", feeAmountFlag, p.feeString)
	}

	p.feeAmount = feeAmount

	operationFeeAmount, ok := new(big.Int).SetString(p.operationFeeString, 0)
	if !ok {
		return fmt.Errorf("--%s invalid amount: %s", operationFeeFlag, p.operationFeeString)
	}

	p.operationFeeAmount = operationFeeAmount

	if p.txType == common.ChainTypeEVMStr {
		srcChainConfig := common.GetChainConfig(p.chainIDSrc)
		minFeeForBridging, minOperationFee :=
			common.DfmToWei(new(big.Int).SetUint64(srcChainConfig.MinFeeForBridging)),
			common.DfmToWei(new(big.Int).SetUint64(srcChainConfig.MinOperationFee))

		if p.feeAmount.Cmp(minFeeForBridging) == -1 {
			return fmt.Errorf("--%s invalid amount: %d", feeAmountFlag, p.feeAmount)
		}

		if p.operationFeeAmount.Cmp(minOperationFee) == -1 {
			return fmt.Errorf("--%s invalid amount: %d", operationFeeFlag, p.operationFeeAmount)
		}

		if p.gatewayAddress == "" {
			return fmt.Errorf("--%s not specified", gatewayAddressFlag)
		}

		if !common.IsValidHTTPURL(p.nexusURL) {
			return fmt.Errorf("invalid --%s flag", nexusURLFlag)
		}
	} else {
		srcChainConfig := common.GetChainConfig(p.chainIDSrc)
		minFeeForBridging, minOperationFee := srcChainConfig.MinFeeForBridging, srcChainConfig.MinOperationFee

		if p.feeAmount.Uint64() < minFeeForBridging {
			return fmt.Errorf("--%s invalid amount: %d", feeAmountFlag, p.feeAmount)
		}

		if p.operationFeeAmount.Uint64() < minOperationFee {
			return fmt.Errorf("--%s invalid amount: %d", operationFeeFlag, p.operationFeeAmount)
		}

		bytes, err := cardanotx.GetCardanoPrivateKeyBytes(p.privateKeyRaw)
		if err != nil {
			return fmt.Errorf("invalid --%s value %s", privateKeyFlag, p.privateKeyRaw)
		}

		var stakeBytes []byte
		if len(p.stakePrivateKeyRaw) > 0 {
			stakeBytes, err = cardanotx.GetCardanoPrivateKeyBytes(p.stakePrivateKeyRaw)
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

		if p.treasuryAddrSrc == "" {
			return fmt.Errorf("--%s not specified", treasuryAddrSrcFlag)
		}

		if p.nexusURL == "" && p.ogmiosURLDst == "" {
			return fmt.Errorf("--%s and --%s not specified", ogmiosURLDstFlag, nexusURLFlag)
		}

		if p.ogmiosURLDst != "" && !common.IsValidHTTPURL(p.ogmiosURLDst) {
			return fmt.Errorf("invalid --%s: %s", ogmiosURLDstFlag, p.ogmiosURLDst)
		}

		if p.nexusURL != "" && !common.IsValidHTTPURL(p.nexusURL) {
			return fmt.Errorf("invalid --%s: %s", nexusURLFlag, p.nexusURL)
		}
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
		&p.txType,
		txTypeFlag,
		"",
		txTypeFlagDesc,
	)

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

	cmd.Flags().Uint16Var(
		&p.tokenIDSrc,
		tokenIDSrcFlag,
		0,
		tokenIDSrcFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.tokenFullNameSrc,
		fullSrcTokenNameFlag,
		cardanowallet.AdaTokenName,
		fullSrcTokenNameFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.tokenFullNameDst,
		fullDstTokenNameFlag,
		cardanowallet.AdaTokenName,
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
		&p.treasuryAddrSrc,
		treasuryAddrSrcFlag,
		"",
		treasuryAddrSrcFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.ogmiosURLDst,
		ogmiosURLDstFlag,
		"",
		ogmiosURLDstFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.gatewayAddress,
		gatewayAddressFlag,
		"",
		gatewayAddressFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.nativeTokenWalletContractAddress,
		nativeTokenWalletContractAddrFlag,
		"",
		nativeTokenWalletContractAddrFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.nexusURL,
		nexusURLFlag,
		"",
		nexusURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.tokenContractAddrDst,
		tokenContractAddrDstFlag,
		"",
		tokenContractAddrDstFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.tokenContractAddrSrc,
		tokenContractAddrSrcFlag,
		"",
		tokenContractAddrSrcFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(gatewayAddressFlag, testnetMagicFlag)
	cmd.MarkFlagsMutuallyExclusive(gatewayAddressFlag, networkIDSrcFlag)
	cmd.MarkFlagsMutuallyExclusive(gatewayAddressFlag, ogmiosURLSrcFlag)
}

func (p *sendSkylineTxParams) Execute(
	outputter common.OutputFormatter,
) (common.ICommandResult, error) {
	ctx := context.Background()

	switch p.txType {
	case common.ChainTypeEVMStr:
		return p.executeEvm(ctx, outputter)
	case common.ChainTypeCardanoStr, "":
		return p.executeCardano(ctx, outputter)
	default:
		return nil, fmt.Errorf("txType not supported")
	}
}

func (p *sendSkylineTxParams) executeCardano(ctx context.Context, outputter common.OutputFormatter) (
	common.ICommandResult, error,
) {
	receivers := toSkylineCardanoMetadata(p.receiversParsed, p.tokenIDSrc)
	networkID := cardanowallet.CardanoNetworkType(p.networkIDSrc)

	srcConfig := common.GetChainConfig(p.chainIDSrc)
	dstConfig := common.GetChainConfig(p.chainIDDst)

	srcTokens := map[uint16]sendtx.ApexToken{
		p.tokenIDSrc: {
			FullName:          p.tokenFullNameSrc,
			IsWrappedCurrency: p.tokenFullNameDst == cardanowallet.AdaTokenName,
		},
	}

	currencyTokenID := apexTokenID
	if p.chainIDSrc == common.ChainIDStrCardano {
		currencyTokenID = adaTokenID
	}

	if p.tokenIDSrc != currencyTokenID {
		srcTokens[currencyTokenID] = sendtx.ApexToken{
			FullName:          cardanowallet.AdaTokenName,
			IsWrappedCurrency: false,
		}
	}

	txSender := sendtx.NewTxSender(
		map[string]sendtx.ChainConfig{
			p.chainIDSrc: {
				CardanoCliBinary:           cardanowallet.ResolveCardanoCliBinary(networkID),
				TxProvider:                 cardanowallet.NewTxProviderOgmios(p.ogmiosURLSrc),
				TestNetMagic:               p.testnetMagicSrc,
				TTLSlotNumberInc:           ttlSlotNumberInc,
				DefaultMinFeeForBridging:   srcConfig.MinFeeForBridging,
				MinFeeForBridgingTokens:    srcConfig.MinFeeForBridging,
				MinOperationFeeAmount:      srcConfig.MinOperationFee,
				MinUtxoValue:               srcConfig.MinUtxoAmount,
				MinColCoinsAllowedToBridge: srcConfig.MinColCoinsAllowedToBridge,
				Tokens:                     srcTokens,
				TreasuryAddress:            p.treasuryAddrSrc,
			},
			p.chainIDDst: {
				MinUtxoValue:             dstConfig.MinUtxoAmount,
				DefaultMinFeeForBridging: dstConfig.MinFeeForBridging,
				MinFeeForBridgingTokens:  dstConfig.MinFeeForBridging,
				MinOperationFeeAmount:    dstConfig.MinOperationFee,
			},
		},
		sendtx.WithMinAmountToBridge(srcConfig.MinUtxoAmount),
	)

	senderAddr, err := cardanotx.GetAddress(networkID, p.wallet)
	if err != nil {
		return nil, err
	}

	txInfo, _, err := txSender.CreateBridgingTx(
		ctx,
		sendtx.BridgingTxDto{
			SrcChainID:      p.chainIDSrc,
			DstChainID:      p.chainIDDst,
			SenderAddr:      senderAddr.String(),
			Receivers:       receivers,
			BridgingAddress: p.multisigAddrSrc,
			BridgingFee:     p.feeAmount.Uint64(),
			OperationFee:    p.operationFeeAmount.Uint64(),
		},
	)
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

	if p.ogmiosURLDst != "" {
		err = waitForCardanoSkylineTx(
			ctx, cardanowallet.NewTxProviderOgmios(p.ogmiosURLDst), p.tokenFullNameDst, p.receiversParsed)
		if err != nil {
			return nil, err
		}
	} else if p.nexusURL != "" {
		txHelper, err := getTxHelper(p.nexusURL)
		if err != nil {
			return nil, err
		}

		receivers := make([]*receiverAmount, len(p.receiversParsed))
		for i, rec := range p.receiversParsed {
			receivers[i] = &receiverAmount{
				ReceiverAddr: rec.ReceiverAddr,
				Amount:       common.DfmToWei(rec.Amount),
			}
		}

		err = waitForEvmSkylineTx(ctx, txHelper.GetClient(), p.tokenContractAddrDst, receivers)
		if err != nil {
			return nil, err
		}
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

func (p *sendSkylineTxParams) executeEvm(ctx context.Context, outputter common.OutputFormatter) (
	common.ICommandResult, error,
) {
	contractAddress := common.HexToAddress(p.gatewayAddress)
	chainID := common.ToNumChainID(p.chainIDDst)
	receivers, totalTokenAmount := toSkylineGatewayStruct(p.receiversParsed, p.tokenIDSrc)

	totalAmount := big.NewInt(0)
	totalAmount.Add(totalAmount, p.feeAmount)
	totalAmount.Add(totalAmount, p.operationFeeAmount)

	wallet, err := ethtxhelper.NewEthTxWallet(p.privateKeyRaw)
	if err != nil {
		return nil, err
	}

	txHelper, err := getTxHelper(p.nexusURL)
	if err != nil {
		return nil, err
	}

	if p.tokenContractAddrSrc != "" {
		_, _ = outputter.Write([]byte("submitting approve tx..."))
		outputter.WriteOutput()

		parsed, err := abi.JSON(strings.NewReader(approveERC20ABI))
		if err != nil {
			return nil, err
		}

		client := txHelper.GetClient()

		erc20Contract := bind.NewBoundContract(
			common.HexToAddress(p.tokenContractAddrSrc),
			parsed,
			client,
			client,
			client,
		)

		tx, err := infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (*types.Transaction, error) {
			return txHelper.SendTx(ctx, wallet, bind.TransactOpts{},
				func(opts *bind.TransactOpts) (*types.Transaction, error) {
					return erc20Contract.Transact(
						opts, "approve", ethcommon.HexToAddress(p.nativeTokenWalletContractAddress), totalTokenAmount)
				})
		})
		if err != nil {
			return nil, err
		}

		_, _ = outputter.Write([]byte(fmt.Sprintf("approve transaction has been submitted: %s", tx.Hash())))
		outputter.WriteOutput()

		receipt, err := txHelper.WaitForReceipt(ctx, tx.Hash().String())
		if err != nil {
			return nil, err
		} else if receipt.Status != types.ReceiptStatusSuccessful {
			return nil, errors.New("approve transaction receipt status is unsuccessful")
		}
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
		abi, "withdraw", chainID, receivers, p.feeAmount, p.operationFeeAmount)
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
				txOpts, chainID, receivers, p.feeAmount, p.operationFeeAmount,
			)
		})
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte(fmt.Sprintf("transaction has been submitted: %s", tx.Hash())))
	outputter.WriteOutput()

	receipt, err := txHelper.WaitForReceipt(ctx, tx.Hash().String())
	if err != nil {
		return nil, err
	} else if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, errors.New("transaction receipt status is unsuccessful")
	}

	if p.ogmiosURLDst != "" {
		receivers := make([]*receiverAmount, len(p.receiversParsed))
		for i, rec := range p.receiversParsed {
			receivers[i] = &receiverAmount{
				ReceiverAddr: rec.ReceiverAddr,
				Amount:       common.WeiToDfm(rec.Amount),
			}
		}

		err = waitForCardanoSkylineTx(
			ctx, cardanowallet.NewTxProviderOgmios(p.ogmiosURLDst),
			p.tokenFullNameDst, receivers)
		if err != nil {
			return nil, err
		}

		_, _ = outputter.Write([]byte("Transaction has been bridged"))
		outputter.WriteOutput()
	}

	return CmdResult{
		SenderAddr: wallet.GetAddress().String(),
		ChainID:    p.chainIDDst,
		Receipts:   p.receiversParsed,
		TxHash:     receipt.TxHash.String(),
	}, nil
}

func waitForEvmSkylineTx(
	ctx context.Context, client *ethclient.Client, tokenContractAddr string, receivers []*receiverAmount) error {
	if tokenContractAddr == "" {
		return waitForTx(ctx, receivers, func(ctx context.Context, addr string) (*big.Int, error) {
			return client.BalanceAt(ctx, common.HexToAddress(addr), nil)
		})
	}

	return waitForTx(ctx, receivers, func(ctx context.Context, addr string) (*big.Int, error) {
		return getERC20Balance(client, common.HexToAddress(tokenContractAddr), common.HexToAddress(addr))
	})
}

func waitForCardanoSkylineTx(
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

const approveERC20ABI = `
[
    {
      "inputs": [
        {
          "internalType": "address",
          "name": "spender",
          "type": "address"
        },
        {
          "internalType": "uint256",
          "name": "value",
          "type": "uint256"
        }
      ],
      "name": "approve",
      "outputs": [
        {
          "internalType": "bool",
          "name": "",
          "type": "bool"
        }
      ],
      "stateMutability": "nonpayable",
      "type": "function"
    }
]`

const balanceOfERC20ABI = `
[
    {
      "inputs": [
        {
          "internalType": "address",
          "name": "account",
          "type": "address"
        }
      ],
      "name": "balanceOf",
      "outputs": [
        {
          "internalType": "uint256",
          "name": "",
          "type": "uint256"
        }
      ],
      "stateMutability": "view",
      "type": "function"
    }
]`

func getERC20Balance(
	client *ethclient.Client, tokenContractAddr ethcommon.Address, addr ethcommon.Address,
) (*big.Int, error) {
	parsedABI, err := abi.JSON(strings.NewReader(balanceOfERC20ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse abi. err: %w", err)
	}

	contract := bind.NewBoundContract(tokenContractAddr, parsedABI, client, client, client)

	var out []interface{}

	err = contract.Call(&bind.CallOpts{}, &out, "balanceOf", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract. err: %w", err)
	}

	balance, ok := out[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("failed to convert erc20 balanceOf result to big.Int")
	}

	return balance, nil
}

func toSkylineCardanoMetadata(receivers []*receiverAmount, tokenID uint16) []sendtx.BridgingTxReceiver {
	metadataReceivers := make([]sendtx.BridgingTxReceiver, len(receivers))
	for idx, rec := range receivers {
		metadataReceivers[idx] = sendtx.BridgingTxReceiver{
			Addr:    rec.ReceiverAddr,
			Amount:  rec.Amount.Uint64(),
			TokenID: tokenID,
		}
	}

	return metadataReceivers
}

func toSkylineGatewayStruct(receivers []*receiverAmount, tokenID uint16) (
	[]contractbinding.IGatewayStructsReceiverWithdraw, *big.Int,
) {
	total := big.NewInt(0)

	gatewayOutputs := make([]contractbinding.IGatewayStructsReceiverWithdraw, len(receivers))
	for idx, rec := range receivers {
		gatewayOutputs[idx] = contractbinding.IGatewayStructsReceiverWithdraw{
			Receiver: rec.ReceiverAddr,
			Amount:   rec.Amount,
			TokenId:  tokenID,
		}

		total.Add(total, rec.Amount)
	}

	return gatewayOutputs, total
}
