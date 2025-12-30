package clicreateaddress

import (
	"context"
	"fmt"
	"os"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	vcCore "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
)

const (
	networkIDFlag        = "network-id"
	testnetMagicFlag     = "testnet-magic"
	chainIDFlag          = "chain"
	bridgeNodeURLFlag    = "bridge-url"
	bridgeSCAddrFlag     = "bridge-addr"
	bridgePrivateKeyFlag = "bridge-key"
	privateKeyConfigFlag = "key-config"
	showPolicyScrFlag    = "show-policy-script"
	custodialAddrFlag    = "generate-custodial-address"
	configFlag           = "config"

	networkIDFlagDesc        = "network ID"
	testnetMagicFlagDesc     = "testnet magic number. leave 0 for mainnet"
	bridgeNodeURLFlagDesc    = "bridge node url"
	bridgeSCAddrFlagDesc     = "bridge smart contract address"
	chainIDFlagDesc          = "cardano chain ID (prime, vector, etc)"
	bridgePrivateKeyFlagDesc = "private key for bridge admin"
	privateKeyConfigFlagDesc = "path to secrets manager config file"
	showPolicyScrFlagDesc    = "show policy script"
	custodialAddrFlagDesc    = "generate custodial address"
	configFlagDesc           = "path to config json file"
)

type createAddressParams struct {
	networkID    uint
	testnetMagic uint

	bridgeNodeURL    string
	bridgeSCAddr     string
	chainID          string
	bridgePrivateKey string
	privateKeyConfig string
	showPolicyScript bool
	custodialAddress bool
	config           string

	chainIDConverter *common.ChainIDConverter
}

func (ip *createAddressParams) validateFlags() error {
	if !common.IsValidHTTPURL(ip.bridgeNodeURL) {
		return fmt.Errorf("invalid --%s flag", bridgeNodeURLFlag)
	}

	if !ethcommon.IsHexAddress(ip.bridgeSCAddr) {
		return fmt.Errorf("invalid --%s flag", bridgeSCAddrFlag)
	}

	if params.config == "" {
		return fmt.Errorf("--%s flag not specified", configFlag)
	}

	if _, err := os.Stat(params.config); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config file does not exist: %s", params.config)
		}

		return fmt.Errorf("failed to check config file: %s. err: %w", params.config, err)
	}

	config, err := common.LoadConfig[vcCore.AppConfig](ip.config, "")
	if err != nil {
		return fmt.Errorf("failed to load config file: %w", err)
	}

	if err := config.SetupChainIDs(); err != nil {
		return fmt.Errorf("failed to setup chain ids: %w", err)
	}

	ip.chainIDConverter = config.ChainIDConverter

	if !ip.chainIDConverter.IsExistingChainID(ip.chainID) {
		return fmt.Errorf("unexisting chain: %s", ip.chainID)
	}

	return nil
}

func (ip *createAddressParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().UintVar(
		&ip.networkID,
		networkIDFlag,
		0,
		networkIDFlagDesc,
	)

	cmd.Flags().UintVar(
		&ip.testnetMagic,
		testnetMagicFlag,
		0,
		testnetMagicFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.bridgeNodeURL,
		bridgeNodeURLFlag,
		"",
		bridgeNodeURLFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.bridgeSCAddr,
		bridgeSCAddrFlag,
		"",
		bridgeSCAddrFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.chainID,
		chainIDFlag,
		common.ChainIDStrPrime,
		chainIDFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.bridgePrivateKey,
		bridgePrivateKeyFlag,
		"",
		bridgePrivateKeyFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.privateKeyConfig,
		privateKeyConfigFlag,
		"",
		privateKeyConfigFlagDesc,
	)

	cmd.Flags().BoolVar(
		&ip.showPolicyScript,
		showPolicyScrFlag,
		false,
		showPolicyScrFlagDesc,
	)

	cmd.Flags().BoolVar(
		&ip.custodialAddress,
		custodialAddrFlag,
		false,
		custodialAddrFlagDesc,
	)

	cmd.Flags().StringVar(
		&params.config,
		configFlag,
		"",
		configFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(privateKeyConfigFlag, bridgePrivateKeyFlag)
}

func (ip *createAddressParams) Execute(
	ctx context.Context, outputter common.OutputFormatter,
) (common.ICommandResult, error) {
	txHelperBridge, err := ip.getTxHelperBridge()
	if err != nil {
		return nil, err
	}

	bridgeContract := eth.NewBridgeSmartContract(ip.bridgeSCAddr, txHelperBridge, ip.chainIDConverter)
	cliBinary := wallet.ResolveCardanoCliBinary(wallet.CardanoNetworkType(ip.networkID))

	validatorsData, err := bridgeContract.GetValidatorsChainData(ctx, ip.chainID)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Validators chain data retrieved:\n"))
	_, _ = outputter.Write([]byte(eth.GetChainValidatorsDataInfoString(ip.chainID, validatorsData, ip.chainIDConverter)))
	_, _ = outputter.Write([]byte("\n"))
	outputter.WriteOutput()

	bridgingAddressesCnt, err := bridgeContract.GetBridgingAddressesCount(ctx, ip.chainID)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte(fmt.Sprintf("Bridging addresses count retreived %d:\n", bridgingAddressesCnt)))
	_, _ = outputter.Write([]byte("\n"))
	outputter.WriteOutput()

	cmdResult := CmdResult{
		AddressAndPolicyScripts: make([]AddressAndPolicyScripts, bridgingAddressesCnt),
		ShowPolicyScripts:       ip.showPolicyScript,
	}

	keyHashes, err := cardanotx.NewApexKeyHashes(validatorsData)
	if err != nil {
		return nil, err
	}

	for i := range bridgingAddressesCnt {
		policyScripts := cardanotx.NewApexPolicyScripts(keyHashes, uint64(i))

		addrs, err := cardanotx.NewApexAddresses(cliBinary, ip.testnetMagic, policyScripts)
		if err != nil {
			return nil, err
		}

		if ip.bridgePrivateKey != "" && i == 0 {
			_, _ = outputter.Write(fmt.Appendf(nil, "Configuring bridge smart contract at %s...", ip.bridgeSCAddr))
			outputter.WriteOutput()

			err := bridgeContract.SetChainAdditionalData(ctx, ip.chainID, addrs.Multisig.Payment, addrs.Fee.Payment)
			if err != nil {
				return nil, err
			}
		}

		cmdResult.AddressAndPolicyScripts[i] = AddressAndPolicyScripts{
			ApexAddresses: addrs,
			PolicyScripts: policyScripts}
	}

	if ip.custodialAddress {
		policyScript := cardanotx.NewCustodialPolicyScriptContainer(keyHashes.Multisig, 0)

		addr, err := cardanotx.NewAddressContainer(
			cliBinary, ip.testnetMagic, policyScript)
		if err != nil {
			return nil, err
		}

		cmdResult.Custodial = &Custodial{
			Address:      addr.Payment,
			PolicyScript: policyScript.Payment,
		}
	}

	return &cmdResult, nil
}

func (ip *createAddressParams) getTxHelperBridge() (*eth.EthHelperWrapper, error) {
	ctx := context.Background()

	if ip.bridgePrivateKey == "" && ip.privateKeyConfig == "" {
		return eth.NewEthHelperWrapper(
			hclog.NewNullLogger(),
			ethtxhelper.WithNodeURL(ip.bridgeNodeURL),
			ethtxhelper.WithInitClientAndChainIDFn(ctx),
			ethtxhelper.WithDynamicTx(false)), nil
	}

	wallet, err := eth.GetEthWalletForBladeAdmin(false, ip.bridgePrivateKey, ip.privateKeyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create bridge admin wallet: %w", err)
	}

	return eth.NewEthHelperWrapperWithWallet(
		wallet, hclog.NewNullLogger(),
		ethtxhelper.WithNodeURL(ip.bridgeNodeURL),
		ethtxhelper.WithTxPoolCheck(true),
		ethtxhelper.WithInitClientAndChainIDFn(ctx),
		ethtxhelper.WithDynamicTx(false)), nil
}
