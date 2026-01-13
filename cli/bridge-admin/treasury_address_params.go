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
	treasuryAddressFlag = "treasury-addr"

	treasuryAddressFlagDesc = "evm treasury address"
)

type treasuryBaseParams struct {
	nodeURL          string
	privateKey       string
	privateKeyConfig string
	gatewayAddress   string
	chainIDsConfig   string
}

func (bp *treasuryBaseParams) ValidateBaseFlags() error {
	if !common.IsValidHTTPURL(bp.nodeURL) {
		return fmt.Errorf("invalid --%s flag", nodeFlag)
	}

	if bp.privateKey == "" && bp.privateKeyConfig == "" {
		return fmt.Errorf("specify at least one: --%s or --%s", evmPrivateKeyFlag, privateKeyConfigFlag)
	}

	chainIDsConfig, err := common.LoadConfig[common.ChainIDsConfigFile](bp.chainIDsConfig, "")
	if err != nil {
		return fmt.Errorf("failed to load chain IDs config: %w", err)
	}

	if !common.IsValidAddress(common.ChainIDStrNexus, bp.gatewayAddress, chainIDsConfig.ToChainIDConverter()) {
		return fmt.Errorf("invalid address: --%s", gatewayAddressFlag)
	}

	return nil
}

func (bp *treasuryBaseParams) RegisterBaseFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&bp.nodeURL,
		nodeFlag,
		"",
		nodeFlagDesc,
	)

	cmd.Flags().StringVar(
		&bp.privateKey,
		evmPrivateKeyFlag,
		"",
		evmPrivateKeyFlagDesc,
	)

	cmd.Flags().StringVar(
		&bp.privateKeyConfig,
		privateKeyConfigFlag,
		"",
		privateKeyConfigFlagDesc,
	)

	cmd.Flags().StringVar(
		&bp.gatewayAddress,
		gatewayAddressFlag,
		"",
		gatewayAddressFlagDesc,
	)

	cmd.Flags().StringVar(
		&bp.chainIDsConfig,
		chainIDsConfigFlag,
		"",
		chainIDsConfigFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(privateKeyConfigFlag, evmPrivateKeyFlag)
}

type setTreasuryAddressParams struct {
	treasuryBaseParams
	treasuryAddressStr string
	treasuryAddress    ethcommon.Address
	chainIDsConfig     string
}

func (sp *setTreasuryAddressParams) ValidateFlags() error {
	if err := sp.ValidateBaseFlags(); err != nil {
		return err
	}

	chainIDsConfig, err := common.LoadConfig[common.ChainIDsConfigFile](sp.chainIDsConfig, "")
	if err != nil {
		return fmt.Errorf("failed to load chain IDs config: %w", err)
	}

	if !common.IsValidAddress(common.ChainIDStrNexus, sp.treasuryAddressStr, chainIDsConfig.ToChainIDConverter()) {
		return fmt.Errorf("invalid address: --%s", treasuryAddressFlag)
	}

	sp.treasuryAddress = ethcommon.HexToAddress(sp.treasuryAddressStr)

	return nil
}

func (sp *setTreasuryAddressParams) RegisterFlags(cmd *cobra.Command) {
	sp.RegisterBaseFlags(cmd)

	cmd.Flags().StringVar(
		&sp.treasuryAddressStr,
		treasuryAddressFlag,
		"",
		treasuryAddressFlagDesc,
	)

	cmd.Flags().StringVar(
		&sp.chainIDsConfig,
		chainIDsConfigFlag,
		"",
		chainIDsConfigFlagDesc,
	)
}

func (sp *setTreasuryAddressParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()

	_, _ = outputter.Write([]byte("preparing transaction to update treasury address..."))
	outputter.WriteOutput()

	wallet, err := eth.GetEthWalletForBladeAdmin(true, sp.privateKey, sp.privateKeyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create smart contracts admin wallet: %w", err)
	}

	txHelper, err := ethtxhelper.NewEThTxHelper(
		ethtxhelper.WithNodeURL(sp.nodeURL), ethtxhelper.WithGasFeeMultiplier(150),
		ethtxhelper.WithZeroGasPrice(false), ethtxhelper.WithDefaultGasLimit(0))
	if err != nil {
		return nil, err
	}

	gatewayAddress := common.HexToAddress(sp.gatewayAddress)

	transaction, err := infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (*types.Transaction, error) {
		contract, err := contractbinding.NewGateway(
			gatewayAddress,
			txHelper.GetClient())
		if err != nil {
			return nil, fmt.Errorf("failed to connect to gateway smart contract: %w", err)
		}

		currentTreasuryAddress, err := contract.TreasuryAddress(&bind.CallOpts{})
		if err != nil {
			return nil, err
		}

		if currentTreasuryAddress == sp.treasuryAddress {
			return nil, fmt.Errorf("treasury address is already set to %s", sp.treasuryAddress)
		}

		parsedABI, err := contractbinding.GatewayMetaData.GetAbi()
		if err != nil {
			return nil, fmt.Errorf("failed to parse gateway smart contract abi: %w", err)
		}

		estimatedGas, _, err := txHelper.EstimateGas(
			ctx, wallet.GetAddress(),
			gatewayAddress, nil, 1.2,
			parsedABI, "setTreasuryAddress",
			sp.treasuryAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to estimate gas for gateway smart contract: %w", err)
		}

		return txHelper.SendTx(
			ctx,
			wallet,
			bind.TransactOpts{},
			func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
				txOpts.GasLimit = estimatedGas

				return contract.SetTreasuryAddress(
					txOpts,
					sp.treasuryAddress,
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

type getTreasuryAddressParams struct {
	treasuryBaseParams
}

func (gp *getTreasuryAddressParams) ValidateFlags() error {
	return gp.ValidateBaseFlags()
}

func (gp *getTreasuryAddressParams) RegisterFlags(cmd *cobra.Command) {
	gp.RegisterBaseFlags(cmd)
}

func (gp *getTreasuryAddressParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	txHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithNodeURL(gp.nodeURL))
	if err != nil {
		return nil, err
	}

	_, err = eth.GetEthWalletForBladeAdmin(true, gp.privateKey, gp.privateKeyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create smart contracts admin wallet: %w", err)
	}

	gatewayAddress := common.HexToAddress(gp.gatewayAddress)

	contract, err := contractbinding.NewGateway(
		gatewayAddress,
		txHelper.GetClient())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gateway smart contract: %w", err)
	}

	treasuryAddress, err := contract.TreasuryAddress(&bind.CallOpts{})
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte(fmt.Sprintf("Current treasury address: %s", treasuryAddress.Hex())))
	outputter.WriteOutput()

	return &successResult{}, nil
}

var (
	_ common.CliCommandExecutor = (*setTreasuryAddressParams)(nil)
	_ common.CliCommandExecutor = (*getTreasuryAddressParams)(nil)
)

func NewTreasuryAddressCommand() *cobra.Command {
	getTreasuryAddressCmd := &cobra.Command{
		Use:   "get",
		Short: "get the current treasury address",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return getTreasuryAddressParamsData.ValidateFlags()
		},
		Run: common.GetCliRunCommand(getTreasuryAddressParamsData),
	}
	setTreasuryAddressCmd := &cobra.Command{
		Use:   "set",
		Short: "set the treasury address",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return setTreasuryAddressParamsData.ValidateFlags()
		},
		Run: common.GetCliRunCommand(setTreasuryAddressParamsData),
	}

	getTreasuryAddressParamsData.RegisterFlags(getTreasuryAddressCmd)
	setTreasuryAddressParamsData.RegisterFlags(setTreasuryAddressCmd)

	cmd := &cobra.Command{
		Use:   "treasury-addr",
		Short: "treasury address functions",
	}

	cmd.AddCommand(
		getTreasuryAddressCmd,
		setTreasuryAddressCmd,
	)

	return cmd
}
