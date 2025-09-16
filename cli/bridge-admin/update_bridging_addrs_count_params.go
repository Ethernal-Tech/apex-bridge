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

const (
	bridgingAddrsCountFlag      = "bridging-addresses-count"
	stakeBridgingAddrsCountFlag = "stake-bridging-addresses-count"

	bridgingAddrsCountFlagDesc      = "count of bridging addresses"
	stakeBridgingAddrsCountFlagDesc = "count of stake bridging addresses"
)

type updateBridgingAddrsCountParams struct {
	chainID                 string
	bridgingAddrsCount      int8
	stakeBridgingAddrsCount int8
	bridgeNodeURL           string
	bridgePrivateKey        string
	privateKeyConfig        string
}

// ValidateFlags implements common.CliCommandValidator.
func (params *updateBridgingAddrsCountParams) ValidateFlags() error {
	if !common.IsValidHTTPURL(params.bridgeNodeURL) {
		return fmt.Errorf("invalid --%s flag", bridgeNodeURLFlag)
	}

	if params.chainID == "" {
		return fmt.Errorf("--%s flag not specified", chainIDFlag)
	}

	if params.bridgingAddrsCount == 0 {
		return fmt.Errorf("--%s flag cannot be zero", bridgeAddrIdxFlag)
	}

	if params.bridgePrivateKey == "" && params.privateKeyConfig == "" {
		return fmt.Errorf("specify at least one: --%s or --%s", privateKeyFlag, privateKeyConfigFlag)
	}

	return nil
}

// Execute implements common.CliCommandExecutor.
func (params *updateBridgingAddrsCountParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()
	chainIDInt := common.ToNumChainID(params.chainID)

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

	bridgeContract, err := contractbinding.NewBridgeContract(
		apexBridgeScAddress,
		txHelper.GetClient())
	if err != nil {
		return nil, err
	}

	bridgingAddrsCount, err := resolveAddrCountParam(ctx, func(opts *bind.CallOpts) (uint8, error) {
		return bridgeContract.GetBridgingAddressesCount(opts, chainIDInt)
	}, params.bridgingAddrsCount)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve bridging addresses count: %w", err)
	}

	stakeBridgingAddrsCount, err := resolveAddrCountParam(ctx, func(opts *bind.CallOpts) (uint8, error) {
		return bridgeContract.GetStakeBridgingAddressesCount(opts, chainIDInt)
	}, params.stakeBridgingAddrsCount)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve stake bridging addresses count: %w", err)
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
		"updateBridgingAddrsCount", chainIDInt, bridgingAddrsCount, stakeBridgingAddrsCount)
	if err != nil {
		return nil, err
	}

	tx, err := txHelper.SendTx(
		ctx, wallet, bind.TransactOpts{}, func(opts *bind.TransactOpts) (*types.Transaction, error) {
			opts.GasLimit = estimatedGas

			return contract.UpdateBridgingAddrsCount(opts, chainIDInt, bridgingAddrsCount, stakeBridgingAddrsCount)
		})
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write(fmt.Appendf(nil, "transaction has been submitted: %s", tx.Hash()))
	outputter.WriteOutput()

	receipt, err := txHelper.WaitForReceipt(ctx, tx.Hash().String())
	if err != nil {
		return nil, err
	} else if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, errors.New("transaction receipt status is unsuccessful")
	}

	return &successResult{}, err
}

func (params *updateBridgingAddrsCountParams) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&params.chainID,
		chainIDFlag,
		"",
		chainIDFlagDesc,
	)

	cmd.Flags().Int8Var(
		&params.bridgingAddrsCount,
		bridgingAddrsCountFlag,
		-1,
		bridgingAddrsCountFlagDesc,
	)

	cmd.Flags().Int8Var(
		&params.stakeBridgingAddrsCount,
		stakeBridgingAddrsCountFlag,
		-1,
		stakeBridgingAddrsCountFlagDesc,
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

	cmd.MarkFlagsMutuallyExclusive(privateKeyConfigFlag, privateKeyFlag)
	cmd.MarkFlagsOneRequired(bridgingAddrsCountFlag, stakeBridgingAddrsCountFlag)
}

func resolveAddrCountParam(
	ctx context.Context,
	contractFunc func(*bind.CallOpts) (uint8, error),
	paramValue int8,
) (uint8, error) {
	if paramValue < 0 {
		return contractFunc(&bind.CallOpts{Context: ctx})
	}

	return uint8(paramValue), nil
}

var (
	_ common.CliCommandExecutor = (*updateBridgingAddrsCountParams)(nil)
)
