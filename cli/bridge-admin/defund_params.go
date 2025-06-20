package clibridgeadmin

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
)

const (
	nativeTokenAmountFlag = "native-token-amount"

	defundAddressFlagDesc     = "address where defund amount goes"
	defundAmountFlagDesc      = "amount to withdraw from the hot wallet in DFM"
	defundTokenAmountFlagDesc = "amount to withdraw from the hot wallet in native tokens"
)

type defundParams struct {
	chainID              string
	currencyAmountStr    string
	nativeTokenAmountStr string
	bridgeNodeURL        string
	bridgePrivateKey     string
	privateKeyConfig     string
	address              string
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
			amountFlag, common.MinUtxoAmountDefault)
	}

	g.nativeTokenAmountStr = strings.TrimSpace(g.nativeTokenAmountStr)

	nativeTokenAmount, ok := new(big.Int).SetString(g.nativeTokenAmountStr, 0)
	if !ok || nativeTokenAmount.Sign() < 0 {
		return fmt.Errorf(" --%s flag must specify a value greater or equal than %d in dfm",
			nativeTokenAmountFlag, 0)
	}

	if currencyAmount.Cmp(new(big.Int).SetUint64(common.MinUtxoAmountDefault)) < 0 {
		return fmt.Errorf(" --%s flag must specify a value greater than %d in dfm",
			amountFlag, common.MinUtxoAmountDefault)
	}

	if !common.IsValidAddress(g.chainID, g.address) {
		return fmt.Errorf("invalid address: --%s", addressFlag)
	}

	if g.bridgePrivateKey == "" || g.privateKeyConfig == "" {
		return fmt.Errorf("specify at least one: --%s or --%s", privateKeyFlag, privateKeyConfigFlag)
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

	wallet, err := eth.GetEthWalletForBladeAdmin(false, g.bridgePrivateKey, g.privateKeyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create bridge admin wallet: %w", err)
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
		"defund", chainIDInt, amount, tokenAmount, g.address)
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

	receipt, err := txHelper.WaitForReceipt(ctx, tx.Hash().String())
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
		&g.bridgePrivateKey,
		privateKeyFlag,
		"",
		bridgePrivateKeyFlagDesc,
	)
	cmd.Flags().StringVar(
		&g.privateKeyConfig,
		privateKeyConfigFlag,
		"",
		privateKeyConfigFlagDesc,
	)
	cmd.Flags().StringVar(
		&g.currencyAmountStr,
		amountFlag,
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

	cmd.MarkFlagsMutuallyExclusive(privateKeyConfigFlag, privateKeyFlag)
}

var (
	_ common.CliCommandExecutor = (*defundParams)(nil)
)
