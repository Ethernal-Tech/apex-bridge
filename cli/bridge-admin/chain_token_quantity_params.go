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
	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
)

const (
	isWrappedTokenFlag = "is-wrapped-token"

	isWrappedTokenFlagDesc           = "should refer to wrapped token"
	chainTokenQuantityAmountFlagDesc = "amount to add or subtract"
)

type getChainTokenQuantityParams struct {
	bridgeNodeURL  string
	chainIDs       []string
	chainIDsConfig string
}

// ValidateFlags implements common.CliCommandValidator.
func (g *getChainTokenQuantityParams) ValidateFlags() error {
	if !common.IsValidHTTPURL(g.bridgeNodeURL) {
		return fmt.Errorf("invalid --%s flag", bridgeNodeURLFlag)
	}

	if len(g.chainIDs) == 0 {
		return fmt.Errorf("--%s flag not specified", chainIDFlag)
	}

	if err := validateConfigFilePath(g.chainIDsConfig); err != nil {
		return err
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

	chainIDsConfig, err := common.LoadConfig[common.ChainIDsConfigFile](g.chainIDsConfig, "")
	if err != nil {
		return nil, err
	}

	chainIDConverter := chainIDsConfig.ToChainIDConverter()

	results := make([]chainTokenQuantity, len(g.chainIDs))

	for i, chainID := range g.chainIDs {
		amount, err := contract.GetChainTokenQuantity(&bind.CallOpts{}, chainIDConverter.ToChainIDNum(chainID))
		if err != nil {
			return nil, err
		}

		wrappedAmount, err := contract.GetChainWrappedTokenQuantity(&bind.CallOpts{}, chainIDConverter.ToChainIDNum(chainID))
		if err != nil {
			return nil, err
		}

		results[i] = chainTokenQuantity{
			chainID:       chainID,
			amount:        amount,
			wrappedAmount: wrappedAmount,
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
	cmd.Flags().StringVar(
		&g.chainIDsConfig,
		chainIDsConfigFlag,
		"",
		chainIDsConfigFlagDesc,
	)
}

type updateChainTokenQuantityParams struct {
	bridgeNodeURL    string
	chainID          string
	amountStr        string
	bridgePrivateKey string
	privateKeyConfig string
	isWrappedToken   bool
	chainIDsConfig   string
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

	if g.bridgePrivateKey == "" && g.privateKeyConfig == "" {
		return fmt.Errorf("specify at least one: --%s or --%s", privateKeyFlag, privateKeyConfigFlag)
	}

	if err := validateConfigFilePath(g.chainIDsConfig); err != nil {
		return err
	}

	return nil
}

// Execute implements common.CliCommandExecutor.
func (g *updateChainTokenQuantityParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()

	chainIDsConfig, err := common.LoadConfig[common.ChainIDsConfigFile](g.chainIDsConfig, "")
	if err != nil {
		return nil, err
	}

	chainIDConverter := chainIDsConfig.ToChainIDConverter()

	chainIDInt := chainIDConverter.ToChainIDNum(g.chainID)
	amount, _ := new(big.Int).SetString(g.amountStr, 0)
	increment := amount.Sign() > 0
	amount = amount.Abs(amount)

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

	methodName := "updateChainTokenQuantity"
	if g.isWrappedToken {
		methodName = "updateChainWrappedTokenQuantity"
	}

	tx, err := infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (*types.Transaction, error) {
		estimatedGas, _, err := txHelper.EstimateGas(
			ctx, wallet.GetAddress(), apexBridgeAdminScAddress, nil, gasLimitMultiplier, abi,
			methodName, chainIDInt, increment, amount)
		if err != nil {
			return nil, err
		}

		return txHelper.SendTx(
			ctx, wallet, bind.TransactOpts{}, func(opts *bind.TransactOpts) (*types.Transaction, error) {
				opts.GasLimit = estimatedGas

				if g.isWrappedToken {
					return contract.UpdateChainWrappedTokenQuantity(opts, chainIDInt, increment, amount)
				}

				return contract.UpdateChainTokenQuantity(opts, chainIDInt, increment, amount)
			})
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
		&g.amountStr,
		amountFlag,
		"0",
		chainTokenQuantityAmountFlagDesc,
	)
	cmd.Flags().BoolVar(
		&g.isWrappedToken,
		isWrappedTokenFlag,
		false,
		isWrappedTokenFlagDesc,
	)
	cmd.Flags().StringVar(
		&g.chainIDsConfig,
		chainIDsConfigFlag,
		"",
		chainIDsConfigFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(privateKeyConfigFlag, privateKeyFlag)
}

var (
	_ common.CliCommandExecutor = (*getChainTokenQuantityParams)(nil)
	_ common.CliCommandExecutor = (*updateChainTokenQuantityParams)(nil)
)
