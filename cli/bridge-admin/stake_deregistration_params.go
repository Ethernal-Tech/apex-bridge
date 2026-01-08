package clibridgeadmin

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
)

type stakeDeregParams struct {
	chainID          string
	bridgeAddrIdx    int8
	bridgeNodeURL    string
	bridgePrivateKey string
	privateKeyConfig string
	chainIDsConfig   string
}

// ValidateFlags implements common.CliCommandValidator.
func (params *stakeDeregParams) ValidateFlags() error {
	if !common.IsValidHTTPURL(params.bridgeNodeURL) {
		return fmt.Errorf("invalid --%s flag", bridgeNodeURLFlag)
	}

	if params.chainID == "" {
		return fmt.Errorf("--%s flag not specified", chainIDFlag)
	}

	if params.bridgeAddrIdx < 0 {
		return fmt.Errorf("--%s flag not specified or negative", bridgeAddrIdxFlag)
	}

	if params.bridgePrivateKey == "" && params.privateKeyConfig == "" {
		return fmt.Errorf("specify at least one: --%s or --%s", privateKeyFlag, privateKeyConfigFlag)
	}

	if err := validateConfigFilePath(params.chainIDsConfig); err != nil {
		return err
	}

	return nil
}

// Execute implements common.CliCommandExecutor.
func (params *stakeDeregParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()

	chainIDsConfig, err := common.LoadConfig[common.ChainIDsConfigFile](params.chainIDsConfig, "")
	if err != nil {
		return nil, fmt.Errorf("failed to load chain IDs config: %w", err)
	}

	chainIDInt := chainIDsConfig.ToChainIDConverter().ToChainIDNum(params.chainID)
	bridgeAddrIndex := uint8(params.bridgeAddrIdx) //nolint:gosec

	_, _ = outputter.Write([]byte("creating and sending transaction..."))
	outputter.WriteOutput()

	wallet, err := eth.GetEthWalletForBladeAdmin(false, params.bridgePrivateKey, params.privateKeyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create bridge admin wallet: %w", err)
	}

	txHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithNodeURL(params.bridgeNodeURL))
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
		"stakeAddressOperation", chainIDInt, bridgeAddrIndex, "", uint8(common.StakeDeregConfirmedTxSubType))
	if err != nil {
		return nil, err
	}

	tx, err := txHelper.SendTx(
		ctx, wallet, bind.TransactOpts{}, func(opts *bind.TransactOpts) (*types.Transaction, error) {
			opts.GasLimit = estimatedGas

			return contract.StakeAddressOperation(
				opts,
				chainIDInt,
				bridgeAddrIndex,
				"",
				uint8(common.StakeDeregConfirmedTxSubType),
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

	return &chainTokenQuantityResult{}, err
}

func (params *stakeDeregParams) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&params.chainID,
		chainIDFlag,
		"",
		chainIDFlagDesc,
	)

	cmd.Flags().Int8Var(
		&params.bridgeAddrIdx,
		bridgeAddrIdxFlag,
		-1,
		bridgeAddrIdxFlagDesc,
	)

	cmd.Flags().StringVar(
		&params.bridgeNodeURL,
		bridgeNodeURLFlag,
		"",
		bridgeNodeURLFlagDesc,
	)

	cmd.Flags().StringVar(
		&params.bridgePrivateKey,
		privateKeyFlag,
		"",
		bridgePrivateKeyFlagDesc,
	)

	cmd.Flags().StringVar(
		&params.privateKeyConfig,
		privateKeyConfigFlag,
		"",
		privateKeyConfigFlagDesc,
	)

	cmd.Flags().StringVar(
		&params.chainIDsConfig,
		chainIDsConfigFlag,
		"",
		chainIDsConfigFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(privateKeyConfigFlag, privateKeyFlag)
}

var (
	_ common.CliCommandExecutor = (*stakeDeregParams)(nil)
)
