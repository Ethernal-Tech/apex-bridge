package clibridgeadmin

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
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
	nodeURL         string
	privateKey      string
	contractAddress string

	minFeeString            string
	minBridgingAmountString string

	minFeeAmount      *big.Int
	minBridgingAmount *big.Int
}

func (ip *setMinAmountsParams) ValidateFlags() error {
	if !common.IsValidHTTPURL(ip.nodeURL) {
		return fmt.Errorf("invalid --%s flag", nodeFlag)
	}

	if ip.privateKey == "" {
		return fmt.Errorf("not specified --%s flag", evmPrivateKeyFlag)
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

	ip.minFeeAmount = feeAmount

	bridgingAmount, ok := new(big.Int).SetString(ip.minBridgingAmountString, 0)
	if !ok {
		return fmt.Errorf("--%s invalid amount", minBridgingAmountFlag)
	}

	if bridgingAmount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("--%s invalid amount: %d", minBridgingAmountFlag, bridgingAmount)
	}

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
}

func (ip *setMinAmountsParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()

	_, _ = outputter.Write([]byte("preparing transaction to update minimum values..."))
	outputter.WriteOutput()

	wallet, err := ethtxhelper.NewEthTxWallet(ip.privateKey)
	if err != nil {
		return nil, err
	}

	txHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithNodeURL(ip.nodeURL))
	if err != nil {
		return nil, err
	}

	contractAddress := common.HexToAddress(ip.contractAddress)

	transaction, err := infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (*types.Transaction, error) {
		contract, err := contractbinding.NewGateway(
			contractAddress,
			txHelper.GetClient())
		if err != nil {
			return nil, fmt.Errorf("failed to connect to gateway smart contract: %w", err)
		}

		return txHelper.SendTx(
			ctx,
			wallet,
			bind.TransactOpts{},
			func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
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

	receipt, err := txHelper.WaitForReceipt(ctx, transaction.Hash().String(), true)
	if err != nil {
		return nil, err
	} else if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, errors.New("transaction receipt status is unsuccessful")
	}

	return &successResult{}, err
}
