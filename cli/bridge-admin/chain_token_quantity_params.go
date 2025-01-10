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
	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
)

const (
	chainTokenQuantityAmountFlagDesc = "amount to add or subtract"
)

type getChainTokenQuantityParams struct {
	bridgeNodeURL string
	chainIDs      []string
}

// ValidateFlags implements common.CliCommandValidator.
func (g *getChainTokenQuantityParams) ValidateFlags() error {
	if !common.IsValidHTTPURL(g.bridgeNodeURL) {
		return fmt.Errorf("invalid --%s flag", bridgeNodeURLFlag)
	}

	if len(g.chainIDs) == 0 {
		return fmt.Errorf("--%s flag not specified", chainIDFlag)
	}

	return nil
}

// Execute implements common.CliCommandExecutor.
func (g *getChainTokenQuantityParams) Execute(_ common.OutputFormatter) (common.ICommandResult, error) {
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

	results := make([]chainTokenQuantity, len(g.chainIDs))

	for i, chainID := range g.chainIDs {
		amount, err := contract.GetChainTokenQuantity(&bind.CallOpts{}, common.ToNumChainID(chainID))
		if err != nil {
			return nil, err
		}

		results[i] = chainTokenQuantity{
			chainID: chainID,
			amount:  amount,
		}
	}

	return &chainTokenQuantityResult{
		results: results,
	}, nil
}

func (g *getChainTokenQuantityParams) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&g.bridgeNodeURL,
		bridgeNodeURLFlag,
		"",
		bridgeNodeURLFlagDesc,
	)
	cmd.Flags().StringSliceVar(
		&g.chainIDs,
		chainIDFlag,
		nil,
		chainIDFlagDesc,
	)
}

type updateChainTokenQuantityParams struct {
	bridgeNodeURL string
	chainID       string
	amountStr     string
	privateKeyRaw string
}

// ValidateFlags implements common.CliCommandValidator.
func (g *updateChainTokenQuantityParams) ValidateFlags() error {
	if !common.IsValidHTTPURL(g.bridgeNodeURL) {
		return fmt.Errorf("invalid --%s flag", bridgeNodeURLFlag)
	}

	if g.chainID == "" {
		return fmt.Errorf("--%s flag not specified", chainIDFlag)
	}

	g.amountStr = strings.TrimSpace(g.amountStr)

	amount, ok := new(big.Int).SetString(g.amountStr, 0)
	if !ok || amount.BitLen() == 0 {
		return fmt.Errorf("--%s flag must be greater or lower than zero", amountFlag)
	}

	if g.privateKeyRaw == "" {
		return fmt.Errorf("flag --%s not specified", privateKeyFlag)
	}

	return nil
}

// Execute implements common.CliCommandExecutor.
func (g *updateChainTokenQuantityParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()
	chainIDInt := common.ToNumChainID(g.chainID)
	amount, _ := new(big.Int).SetString(g.amountStr, 0)
	increment := amount.Sign() > 0
	amount = amount.Abs(amount)

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

	tx, err := infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (*types.Transaction, error) {
		estimatedGas, _, err := txHelper.EstimateGas(
			ctx, wallet.GetAddress(), apexBridgeAdminScAddress, nil, gasLimitMultiplier, abi,
			"updateChainTokenQuantity", chainIDInt, increment, amount)
		if err != nil {
			return nil, err
		}

		return txHelper.SendTx(
			ctx, wallet, bind.TransactOpts{}, func(opts *bind.TransactOpts) (*types.Transaction, error) {
				opts.GasLimit = estimatedGas

				return contract.UpdateChainTokenQuantity(opts, chainIDInt, increment, amount)
			})
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

	return &successResult{}, err
}

func (g *updateChainTokenQuantityParams) RegisterFlags(cmd *cobra.Command) {
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
		chainTokenQuantityAmountFlagDesc,
	)
}

var (
	_ common.CliCommandExecutor = (*getChainTokenQuantityParams)(nil)
	_ common.CliCommandExecutor = (*updateChainTokenQuantityParams)(nil)
)
