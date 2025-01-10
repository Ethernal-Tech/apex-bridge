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
	nativeTokenAmountFlagDesc = "use at your own risk (see the --amount flag)"                                                                       //nolint:gosec
)

type defundParams struct {
	bridgeNodeURL     string
	chainID           string
	amountStr         string
	privateKeyRaw     string
	address           string
	nativeTokenAmount bool
}

// ValidateFlags implements common.CliCommandValidator.
func (g *defundParams) ValidateFlags() error {
	if !common.IsValidHTTPURL(g.bridgeNodeURL) {
		return fmt.Errorf("invalid --%s flag", bridgeNodeURLFlag)
	}

	if g.chainID == "" {
		return fmt.Errorf("--%s flag not specified", chainIDFlag)
	}

	g.amountStr = strings.TrimSpace(g.amountStr)

	amount, ok := new(big.Int).SetString(g.amountStr, 0)
	if !ok || amount.Sign() <= 0 {
		return fmt.Errorf(" --%s flag must specify a value greater than %d in dfm",
			amountFlag, common.MinUtxoAmountDefault)
	}

	if g.nativeTokenAmount {
		amount = common.GetDfmAmount(g.chainID, amount)
	}

	if amount.Cmp(new(big.Int).SetUint64(common.MinUtxoAmountDefault)) < 0 {
		return fmt.Errorf(" --%s flag must specify a value greater than %d in dfm",
			amountFlag, common.MinUtxoAmountDefault)
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
	amount, _ := new(big.Int).SetString(g.amountStr, 0)

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

	estimatedGas, _, err := txHelper.EstimateGas(
		ctx, wallet.GetAddress(), apexBridgeAdminScAddress, nil, gasLimitMultiplier, abi,
		"defund", chainIDInt, amount, big.NewInt(0), g.address)
	if err != nil {
		return nil, err
	}

	tx, err := txHelper.SendTx(
		ctx, wallet, bind.TransactOpts{}, func(opts *bind.TransactOpts) (*types.Transaction, error) {
			opts.GasLimit = estimatedGas

			return contract.Defund(opts, chainIDInt, amount, big.NewInt(0), g.address) // a TODO - fix for skyline
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
		&g.amountStr,
		amountFlag,
		"0",
		defundAmountFlagDesc,
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
