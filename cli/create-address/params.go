package clicreateaddress

import (
	"context"
	"fmt"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
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
	addressIndexFlag     = "addr-index"

	networkIDFlagDesc        = "network ID"
	testnetMagicFlagDesc     = "testnet magic number. leave 0 for mainnet"
	bridgeNodeURLFlagDesc    = "bridge node url"
	bridgeSCAddrFlagDesc     = "bridge smart contract address"
	chainIDFlagDesc          = "cardano chain ID (prime, vector, etc)"
	bridgePrivateKeyFlagDesc = "private key for bridge admin"
	privateKeyConfigFlagDesc = "path to secrets manager config file"
	showPolicyScrFlagDesc    = "show policy script"
	addrIndexFlagDesc        = "address index"
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
	addrIndex        uint
}

func (ip *createAddressParams) validateFlags() error {
	if !common.IsValidHTTPURL(ip.bridgeNodeURL) {
		return fmt.Errorf("invalid --%s flag", bridgeNodeURLFlag)
	}

	if !ethcommon.IsHexAddress(ip.bridgeSCAddr) {
		return fmt.Errorf("invalid --%s flag", bridgeSCAddrFlag)
	}

	if !common.IsExistingChainID(ip.chainID) {
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

	cmd.Flags().UintVar(
		&ip.addrIndex,
		addressIndexFlag,
		0,
		addrIndexFlagDesc,
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

	cmd.MarkFlagsMutuallyExclusive(privateKeyConfigFlag, bridgePrivateKeyFlag)
}

func (ip *createAddressParams) Execute(
	ctx context.Context, outputter common.OutputFormatter,
) (common.ICommandResult, error) {
	txHelperBridge, err := ip.getTxHelperBridge()
	if err != nil {
		return nil, err
	}

	bridgeContract := eth.NewBridgeSmartContract(ip.bridgeSCAddr, txHelperBridge)
	cliBinary := wallet.ResolveCardanoCliBinary(wallet.CardanoNetworkType(ip.networkID))

	validatorsData, err := bridgeContract.GetValidatorsChainData(ctx, ip.chainID)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Validators chain data retrieved:\n"))
	_, _ = outputter.Write([]byte(eth.GetChainValidatorsDataInfoString(ip.chainID, validatorsData)))
	_, _ = outputter.Write([]byte("\n"))
	outputter.WriteOutput()

	keyHashes, err := cardanotx.NewApexKeyHashes(validatorsData)
	if err != nil {
		return nil, err
	}

	policyScripts := cardanotx.NewApexPolicyScripts(keyHashes, uint64(ip.addrIndex))

	addrs, err := cardanotx.NewApexAddresses(cliBinary, ip.testnetMagic, policyScripts)
	if err != nil {
		return nil, err
	}

	if ip.bridgePrivateKey != "" {
		_, _ = outputter.Write(fmt.Appendf(nil, "Configuring bridge smart contract at %s...", ip.bridgeSCAddr))
		outputter.WriteOutput()

		err := bridgeContract.SetChainAdditionalData(ctx, ip.chainID, addrs.Multisig.Payment, addrs.Fee.Payment)
		if err != nil {
			return nil, err
		}
	}

	return &CmdResult{
		ApexAddresses:     addrs,
		PolicyScripts:     policyScripts,
		ShowPolicyScripts: ip.showPolicyScript,
	}, nil
}

func (ip *createAddressParams) getTxHelperBridge() (*eth.EthHelperWrapper, error) {
	if ip.bridgePrivateKey == "" && ip.privateKeyConfig == "" {
		return eth.NewEthHelperWrapper(
			hclog.NewNullLogger(),
			ethtxhelper.WithNodeURL(ip.bridgeNodeURL),
			ethtxhelper.WithInitClientAndChainIDFn(context.Background()),
			ethtxhelper.WithDynamicTx(false)), nil
	}

	wallet, err := eth.GetEthWalletForBladeAdmin(false, ip.bridgePrivateKey, ip.privateKeyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create bridge admin wallet: %w", err)
	}

	return eth.NewEthHelperWrapperWithWallet(
		wallet, hclog.NewNullLogger(),
		ethtxhelper.WithNodeURL(ip.bridgeNodeURL),
		ethtxhelper.WithInitClientAndChainIDFn(context.Background()),
		ethtxhelper.WithDynamicTx(false)), nil
}
