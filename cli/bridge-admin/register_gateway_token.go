package clibridgeadmin

import (
	"context"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	vcCore "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
)

const (
	nodeURLFlag        = "node-url"
	gatewayAddressFlag = "gateway-addr"
	gasLimitFlag       = "gas-limit"
	tokenSCAddressFlag = "token-sc-addr" //nolint:gosec
	tokIDFlag          = "token-id"
	tokNameFlag        = "token-name"
	tokSymbolFlag      = "token-symbol"

	nodeURLFlagDesc        = "evm node url"
	gatewayAddressFlagDesc = "address of the gateway smart contract"
	gasLimitFlagDesc       = "gas limit for transaction"
	tokenSCAddressFlagDesc = "address of the token smart contract. passed only if the bridge does not own the contract" //nolint:lll
	tokIDFlagDesc          = "id of the token"
	tokNameFlagDesc        = "name of the token"
	tokSymbolFlagDesc      = "symbol of the token"

	defaultGasFeeMultiplier = 200 // 200%
	defaultGasLimitValue    = uint64(8_242_880)
)

type registerGatewayTokenParams struct {
	nodeURL           string
	privateKey        string
	privateKeyConfig  string
	gatewayAddress    string
	gasLimit          uint64
	tokenSCAddressStr string
	tokenID           uint16
	tokenName         string
	tokenSymbol       string

	tokenSCAddress ethcommon.Address
	config         string
}

// ValidateFlags implements common.CliCommandValidator.
func (g *registerGatewayTokenParams) ValidateFlags() error {
	if !common.IsValidHTTPURL(g.nodeURL) {
		return fmt.Errorf("invalid --%s flag", nodeURLFlag)
	}

	if g.privateKey == "" && g.privateKeyConfig == "" {
		return fmt.Errorf("specify at least one: --%s or --%s", privateKeyFlag, privateKeyConfigFlag)
	}

	if err := validateConfigFilePath(g.config); err != nil {
		return err
	}

	config, err := common.LoadConfig[vcCore.AppConfig](g.config, "")
	if err != nil {
		return fmt.Errorf("failed to load config file: %w", err)
	}

	config.SetupChainIDs()

	if !common.IsValidAddress(common.ChainIDStrNexus, g.gatewayAddress, config.ChainIDConverter) {
		return fmt.Errorf("invalid address: --%s", gatewayAddressFlag)
	}

	if !common.IsValidAddress(common.ChainIDStrNexus, g.tokenSCAddressStr, config.ChainIDConverter) {
		return fmt.Errorf("invalid address: --%s", tokenSCAddressFlag)
	}

	g.tokenSCAddress = ethcommon.HexToAddress(g.tokenSCAddressStr)

	if g.tokenID == 0 {
		return fmt.Errorf("invalid tokenID: --%s", tokIDFlag)
	}

	if g.tokenName == "" {
		return fmt.Errorf("invalid tokenName: --%s", tokNameFlag)
	}

	if g.tokenSymbol == "" {
		return fmt.Errorf("invalid tokenSymbol: --%s", tokSymbolFlag)
	}

	return nil
}

// Execute implements common.CliCommandExecutor.
func (g *registerGatewayTokenParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	var (
		ctx    = context.Background()
		logger = hclog.NewNullLogger()
	)

	wallet, err := eth.GetEthWalletForBladeAdmin(true, g.privateKey, g.privateKeyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create smart contracts admin wallet: %w", err)
	}

	txHelper := eth.NewEthHelperWrapperWithWallet(wallet, logger,
		ethtxhelper.WithNodeURL(g.nodeURL),
		ethtxhelper.WithInitClientAndChainIDFn(ctx),
		ethtxhelper.WithDefaultGasLimit(g.gasLimit),
		ethtxhelper.WithZeroGasPrice(false),
		ethtxhelper.WithGasFeeMultiplier(defaultGasFeeMultiplier),
	)

	evmSmartContract, err := eth.NewSimpleEVMGatewaySmartContract(
		g.gatewayAddress, txHelper, logger)
	if err != nil {
		return nil, err
	}

	tokenRegisteredEvent, err := evmSmartContract.RegisterToken(
		ctx, g.tokenSCAddress, g.tokenID, g.tokenName, g.tokenSymbol)
	if err != nil {
		return nil, err
	}

	return &registerGatewayTokenResult{tokenRegisteredEvent}, nil
}

func (g *registerGatewayTokenParams) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&g.nodeURL,
		nodeURLFlag,
		"",
		nodeURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&g.privateKey,
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
		&g.gatewayAddress,
		gatewayAddressFlag,
		"",
		gatewayAddressFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&g.gasLimit,
		gasLimitFlag,
		defaultGasLimitValue,
		gasLimitFlagDesc,
	)
	cmd.Flags().StringVar(
		&g.tokenSCAddressStr,
		tokenSCAddressFlag,
		common.EthZeroAddr,
		tokenSCAddressFlagDesc,
	)
	cmd.Flags().Uint16Var(
		&g.tokenID,
		tokIDFlag,
		0,
		tokIDFlagDesc,
	)
	cmd.Flags().StringVar(
		&g.tokenName,
		tokNameFlag,
		"",
		tokNameFlagDesc,
	)
	cmd.Flags().StringVar(
		&g.tokenSymbol,
		tokSymbolFlag,
		"",
		tokSymbolFlagDesc,
	)
	cmd.Flags().StringVar(
		&g.config,
		configFlag,
		"",
		configFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(privateKeyConfigFlag, privateKeyFlag)
}

var (
	_ common.CliCommandExecutor = (*registerGatewayTokenParams)(nil)
)
