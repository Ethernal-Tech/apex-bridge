package clibridgeadmin

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet/bech32"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
)

const (
	stakePoolIDFlag   = "stake-pool"
	bridgeAddrIdxFlag = "bridge-address-index"

	stakePoolIDFlagDesc   = "identifier of the stake pool to delegate to"
	bridgeAddrIdxFlagDesc = "index of the bridging address to be delegated"
)

type stakeDelParams struct {
	chainID          string
	bridgeAddrIdx    int8
	stakePoolID      string
	bridgeNodeURL    string
	bridgePrivateKey string
	privateKeyConfig string
}

// ValidateFlags implements common.CliCommandValidator.
func (params *stakeDelParams) ValidateFlags() error {
	if !common.IsValidHTTPURL(params.bridgeNodeURL) {
		return fmt.Errorf("invalid --%s flag", bridgeNodeURLFlag)
	}

	if params.chainID == "" {
		return fmt.Errorf("--%s flag not specified", chainIDFlag)
	}

	if params.bridgeAddrIdx < 0 {
		return fmt.Errorf("--%s flag not specified or negative", bridgeAddrIdxFlag)
	}

	if params.stakePoolID == "" {
		return fmt.Errorf("--%s flag not specified", stakePoolIDFlag)
	} else {
		prefix, _, err := bech32.Decode(params.stakePoolID)
		if err != nil || prefix != "pool" {
			return fmt.Errorf("invalid --%s", stakePoolIDFlag)
		}
	}

	if params.bridgePrivateKey == "" && params.privateKeyConfig == "" {
		return fmt.Errorf("specify at least one: --%s or --%s", privateKeyFlag, privateKeyConfigFlag)
	}

	return nil
}

// Execute implements common.CliCommandExecutor.
func (params *stakeDelParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()
	chainIDInt := common.ToNumChainID(params.chainID)
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

	contract, err := contractbinding.NewBridgeContract(
		apexBridgeScAddress,
		txHelper.GetClient())
	if err != nil {
		return nil, err
	}

	abi, err := contractbinding.BridgeContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	estimatedGas, _, err := txHelper.EstimateGas(
		ctx, wallet.GetAddress(), apexBridgeScAddress, nil, gasLimitMultiplier, abi,
		"delegateAddrToStakePool", chainIDInt, bridgeAddrIndex, params.stakePoolID)
	if err != nil {
		return nil, err
	}

	tx, err := txHelper.SendTx(
		ctx, wallet, bind.TransactOpts{}, func(opts *bind.TransactOpts) (*types.Transaction, error) {
			opts.GasLimit = estimatedGas

			return contract.DelegateAddrToStakePool(opts, chainIDInt, bridgeAddrIndex, params.stakePoolID)
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

func (params *stakeDelParams) RegisterFlags(cmd *cobra.Command) {
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
		&params.stakePoolID,
		stakePoolIDFlag,
		"",
		stakePoolIDFlagDesc,
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
}

var (
	_ common.CliCommandExecutor = (*stakeDelParams)(nil)
)
