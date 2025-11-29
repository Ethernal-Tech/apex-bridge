package clibridgeadmin

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/Ethernal-Tech/apex-bridge/common"
	reactorgatewaycontractbinding "github.com/Ethernal-Tech/apex-bridge/contractbinding/gateway/reactor"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
)

const (
	nodeFlag              = "url"
	evmPrivateKeyFlag     = "key"
	contractAddressFlag   = "contract-addr"
	minFeeAmountFlag      = "min-fee"
	minBridgingAmountFlag = "min-bridging-amount"

	nodeFlagDesc              = "evm node url"
	evmPrivateKeyFlagDesc     = "private key for evm chain"
	contractAddressFlagDesc   = "address of the Gateway contract"
	minFeeAmountFlagDesc      = "minimal fee amount"
	minBridgingAmountFlagDesc = "minimal amount to bridge"
)

type setMinAmountsParams struct {
	nodeURL          string
	privateKey       string
	privateKeyConfig string
	contractAddress  string

	minFeeString            string
	minBridgingAmountString string

	minFeeAmount      *big.Int
	minBridgingAmount *big.Int
}

func (ip *setMinAmountsParams) ValidateFlags() error {
	if !common.IsValidHTTPURL(ip.nodeURL) {
		return fmt.Errorf("invalid --%s flag", nodeFlag)
	}

	if ip.privateKey == "" && ip.privateKeyConfig == "" {
		return fmt.Errorf("specify at least one: --%s or --%s", evmPrivateKeyFlag, privateKeyConfigFlag)
	}

	if !ethcommon.IsHexAddress(ip.contractAddress) {
		return fmt.Errorf("invalid --%s flag", bridgeSCAddrFlag)
	}

	feeAmount, ok := new(big.Int).SetString(ip.minFeeString, 0)
	if !ok {
		return fmt.Errorf("--%s invalid amount", minFeeAmountFlag)
	}

	if feeAmount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("--%s invalid amount: %d", minFeeAmountFlag, feeAmount)
	}

	bridgingAmount, ok := new(big.Int).SetString(ip.minBridgingAmountString, 0)
	if !ok {
		return fmt.Errorf("--%s invalid amount", minBridgingAmountFlag)
	}

	if bridgingAmount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("--%s invalid amount: %d", minBridgingAmountFlag, bridgingAmount)
	}

	ip.minFeeAmount = feeAmount
	ip.minBridgingAmount = bridgingAmount

	return nil
}

func (ip *setMinAmountsParams) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&ip.nodeURL,
		nodeFlag,
		"",
		nodeFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.privateKey,
		evmPrivateKeyFlag,
		"",
		evmPrivateKeyFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.privateKeyConfig,
		privateKeyConfigFlag,
		"",
		privateKeyConfigFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.contractAddress,
		contractAddressFlag,
		"",
		contractAddressFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.minFeeString,
		minFeeAmountFlag,
		strconv.FormatUint(common.MinFeeForBridgingDefault, 10),
		minFeeAmountFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.minBridgingAmountString,
		minBridgingAmountFlag,
		strconv.FormatUint(common.MinUtxoAmountDefault, 10),
		minBridgingAmountFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(privateKeyConfigFlag, evmPrivateKeyFlag)
}

func (ip *setMinAmountsParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()

	_, _ = outputter.Write([]byte("preparing transaction to update minimum values..."))
	outputter.WriteOutput()

	wallet, err := eth.GetEthWalletForBladeAdmin(true, ip.privateKey, ip.privateKeyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create smart contracts admin wallet: %w", err)
	}

	txHelper, err := ethtxhelper.NewEThTxHelper(
		ethtxhelper.WithNodeURL(ip.nodeURL), ethtxhelper.WithGasFeeMultiplier(150),
		ethtxhelper.WithZeroGasPrice(false), ethtxhelper.WithDefaultGasLimit(0))
	if err != nil {
		return nil, err
	}

	contractAddress := common.HexToAddress(ip.contractAddress)

	transaction, err := infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (*types.Transaction, error) {
		contract, err := reactorgatewaycontractbinding.NewGateway(
			contractAddress,
			txHelper.GetClient())
		if err != nil {
			return nil, fmt.Errorf("failed to connect to gateway smart contract: %w", err)
		}

		parsedABI, err := reactorgatewaycontractbinding.GatewayMetaData.GetAbi()
		if err != nil {
			return nil, fmt.Errorf("failed to parse gateway smart contract abi: %w", err)
		}

		estimatedGas, _, err := txHelper.EstimateGas(
			ctx, wallet.GetAddress(),
			contractAddress, nil, 1.2,
			parsedABI, "setMinAmounts", ip.minFeeAmount, ip.minBridgingAmount)
		if err != nil {
			return nil, fmt.Errorf("failed to estimate gas for gateway smart contract: %w", err)
		}

		return txHelper.SendTx(
			ctx,
			wallet,
			bind.TransactOpts{},
			func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
				txOpts.GasLimit = estimatedGas

				return contract.SetMinAmounts(
					txOpts,
					ip.minFeeAmount,
					ip.minBridgingAmount,
				)
			},
		)
	})
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte(fmt.Sprintf("transaction has been submitted: %s", transaction.Hash())))
	outputter.WriteOutput()

	receipt, err := txHelper.WaitForReceipt(ctx, transaction.Hash().String())
	if err != nil {
		return nil, err
	} else if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, errors.New("transaction receipt status is unsuccessful")
	}

	return &successResult{}, err
}
