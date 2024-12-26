package clibridgeadmin

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
)

const (
	nativeTokenAmountFlag = "native-token-amount"

	defundAddressFlagDesc     = "address where defund amount goes"
	defundAmountFlagDesc      = "amount to withdraw from the hot wallet in DFM (or in native tokens if the --native-token-amount flag is specified)" //nolint:lll
	defundTokenAmountFlagDesc = "amount to withdraw native tokens from the hot wallet in DFM"
	nativeTokenAmountFlagDesc = "use at your own risk (see the --amount flag)" //nolint:gosec
)

type defundParams struct {
	bridgeNodeURL        string
	chainID              string
	currencyAmountStr    string
	nativeTokenAmountStr string
	privateKeyRaw        string
	address              string
	nativeTokenAmount    bool
}

// ValidateFlags implements common.CliCommandValidator.
func (g *defundParams) ValidateFlags() error {
	if !common.IsValidHTTPURL(g.bridgeNodeURL) {
		return fmt.Errorf("invalid --%s flag", bridgeNodeURLFlag)
	}

	if g.chainID == "" {
		return fmt.Errorf("--%s flag not specified", chainIDFlag)
	}

	g.currencyAmountStr = strings.TrimSpace(g.currencyAmountStr)

	currencyAmount, ok := new(big.Int).SetString(g.currencyAmountStr, 0)
	if !ok || currencyAmount.Sign() <= 0 {
		return fmt.Errorf(" --%s flag must specify a value greater than %d in dfm",
			currencyAmountFlag, common.MinUtxoAmountDefault)
	}

	g.nativeTokenAmountStr = strings.TrimSpace(g.nativeTokenAmountStr)

	nativeTokenAmount, ok := new(big.Int).SetString(g.nativeTokenAmountStr, 0)
	if !ok || nativeTokenAmount.Sign() <= 0 {
		return fmt.Errorf(" --%s flag must specify a value greater than %d in dfm",
			nativeTokenAmountFlag, 0) // should this be 0 by default?
	}

	if currencyAmount.Cmp(new(big.Int).SetUint64(common.MinUtxoAmountDefault)) < 0 {
		return fmt.Errorf(" --%s flag must specify a value greater than %d in dfm",
			currencyAmountFlag, common.MinUtxoAmountDefault)
	}

	if g.privateKeyRaw == "" {
		return fmt.Errorf("flag --%s not specified", privateKeyFlag)
	}

	if !common.IsValidAddress(g.chainID, g.address) {
		return fmt.Errorf("invalid address: --%s", addressFlag)
	}

	return nil
}

// Execute implements common.CliCommandExecutor.
func (g *defundParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()
	chainIDInt := common.ToNumChainID(g.chainID)
	amount, _ := new(big.Int).SetString(g.currencyAmountStr, 0)
	tokenAmount, _ := new(big.Int).SetString(g.nativeTokenAmountStr, 0)

	_, _ = outputter.Write([]byte("creating and sending transaction..."))
	outputter.WriteOutput()

	wallet, err := ethtxhelper.NewEthTxWallet(g.privateKeyRaw)
	if err != nil {
		return nil, err
	}

	txHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithNodeURL(g.bridgeNodeURL))
	if err != nil {
		return nil, err
	}

	contract, err := contractbinding.NewAdminContract(
		apexBridgeAdminScAddress,
		txHelper.GetClient())
	if err != nil {
		return nil, err
	}

	abi, err := contractbinding.AdminContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	// should wrapped amount affect gas estimation?
	estimatedGas, _, err := txHelper.EstimateGas(
		ctx, wallet.GetAddress(), apexBridgeAdminScAddress, nil, gasLimitMultiplier, abi,
		"defund", chainIDInt, g.address, amount)
	if err != nil {
		return nil, err
	}

	tx, err := txHelper.SendTx(
		ctx, wallet, bind.TransactOpts{}, func(opts *bind.TransactOpts) (*types.Transaction, error) {
			opts.GasLimit = estimatedGas

			return contract.Defund(opts, chainIDInt, amount, tokenAmount, g.address)
		})
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte(fmt.Sprintf("transaction has been submitted: %s", tx.Hash())))
	outputter.WriteOutput()

	receipt, err := txHelper.WaitForReceipt(ctx, tx.Hash().String(), true)
	if err != nil {
		return nil, err
	} else if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, errors.New("transaction receipt status is unsuccessful")
	}

	return &chainTokenQuantityResult{}, err
}

func (g *defundParams) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&g.bridgeNodeURL,
		bridgeNodeURLFlag,
		"",
		bridgeNodeURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&g.chainID,
		chainIDFlag,
		"",
		chainIDFlagDesc,
	)
	cmd.Flags().StringVar(
		&g.privateKeyRaw,
		privateKeyFlag,
		"",
		privateKeyFlagDesc,
	)
	cmd.Flags().StringVar(
		&g.currencyAmountStr,
		currencyAmountFlag,
		"0",
		defundAmountFlagDesc,
	)
	cmd.Flags().StringVar(
		&g.nativeTokenAmountStr,
		nativeTokenAmountFlag,
		"0",
		defundTokenAmountFlagDesc,
	)
	cmd.Flags().StringVar(
		&g.address,
		addressFlag,
		"0",
		defundAddressFlagDesc,
	)
	cmd.Flags().BoolVar(
		&g.nativeTokenAmount,
		nativeTokenAmountFlag,
		false,
		nativeTokenAmountFlagDesc,
	)
}

var (
	_ common.CliCommandExecutor = (*defundParams)(nil)
)
