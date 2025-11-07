package clibridgeadmin

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
)

const (
	validatorChangeFlag = "validator-change"
	validatorChangeDesc = "notify the web application that a validator set change is in progress"
)

type setValidatorChangeParams struct {
	nodeURL          string
	privateKey       string
	privateKeyConfig string
	validatorChange  *bool
	contractAddress  string
}

func (v *setValidatorChangeParams) ValidateFlags() error {
	if !common.IsValidHTTPURL(v.nodeURL) {
		return fmt.Errorf("invalid --%s flag", bridgeNodeURLFlag)
	}

	if v.privateKey == "" && v.privateKeyConfig == "" {
		return fmt.Errorf("specify at least one: --%s or --%s", evmPrivateKeyFlag, privateKeyConfigFlag)
	}

	if !ethcommon.IsHexAddress(v.contractAddress) {
		return fmt.Errorf("invalid --%s flag", bridgeSCAddrFlag)
	}

	return nil
}

func (v *setValidatorChangeParams) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&v.nodeURL,
		nodeFlag,
		"",
		nodeFlagDesc,
	)

	cmd.Flags().StringVar(
		&v.privateKey,
		privateKeyFlag,
		"",
		bridgePrivateKeyFlagDesc,
	)

	cmd.Flags().StringVar(
		&v.privateKeyConfig,
		privateKeyConfigFlag,
		"",
		privateKeyConfigFlagDesc,
	)

	cmd.Flags().BoolVar(
		v.validatorChange,
		validatorChangeFlag,
		false,
		validatorChangeDesc,
	)

	cmd.Flags().StringVar(
		&v.contractAddress,
		contractAddressFlag,
		"",
		contractAddressFlagDesc,
	)
}

func (v *setValidatorChangeParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()

	_, _ = outputter.Write([]byte("preparing transaction to return values..."))
	outputter.WriteOutput()

	wallet, err := eth.GetEthWalletForBladeAdmin(false, v.privateKey, v.privateKeyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create smart contracts admin wallet: %w", err)
	}

	txHelper, err := ethtxhelper.NewEThTxHelper(
		ethtxhelper.WithNodeURL(v.nodeURL), ethtxhelper.WithGasFeeMultiplier(150),
		ethtxhelper.WithZeroGasPrice(false), ethtxhelper.WithDefaultGasLimit(0),
	)
	if err != nil {
		return nil, err
	}

	contractAddress := common.HexToAddress(v.contractAddress)

	contract, err := contractbinding.NewAdminContract(
		contractAddress,
		txHelper.GetClient())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to admin smart contract: %w", err)
	}

	parsedABI, err := contractbinding.AdminContractMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to parse admin smart contract abi: %w", err)
	}

	transaction, err := infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (*types.Transaction, error) {
		estimatedGas, _, err := txHelper.EstimateGas(
			ctx, wallet.GetAddress(),
			contractAddress, nil, 1.2,
			parsedABI, "setValidatorSetChange", v.validatorChange)
		if err != nil {
			return nil, fmt.Errorf("failed to estimate gas for admin smart contract: %w", err)
		}

		return txHelper.SendTx(
			ctx,
			wallet,
			bind.TransactOpts{},
			func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
				txOpts.GasLimit = estimatedGas

				return contract.SetValidatorChange(
					txOpts,
					*v.validatorChange,
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
