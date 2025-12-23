package clideployevm

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethcontracts "github.com/Ethernal-Tech/apex-bridge/eth/contracts"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

const (
	contractProxyAddrFlag     = "proxy-addr"
	contractProxyAddrFlagDesc = "proxy address of the deployed contract to configure"
)

type setDependenciesParams struct {
	contractName             string
	contractDir              string
	dependenciesAddressesStr string
	dependenciesAddresses    []ethcommon.Address
	contractProxyAddr        string

	evmPrivateKey    string
	privateKeyConfig string

	repositoryURL string
	clone         bool
	branchName    string

	evmNodeURL   string
	evmDynamicTx bool
	gasLimit     uint64
}

func (ip *setDependenciesParams) validateFlags() error {
	if !common.IsValidHTTPURL(ip.evmNodeURL) {
		return fmt.Errorf("invalid --%s flag", evmNodeURLFlag)
	}

	if ip.contractName == "" {
		return fmt.Errorf("contract name not specified: --%s", contractNameFlag)
	}

	if ip.contractDir == "" {
		return fmt.Errorf("contract directory not specified: --%s", contractDirFlag)
	}

	if ip.dependenciesAddressesStr != "" {
		dependenciesAddrs, err := parseAddresses(ip.dependenciesAddressesStr)
		if err != nil {
			return fmt.Errorf("dependencies addresses are invalid: err: %w --%s", err, dependenciesAddressesFlag)
		}

		ip.dependenciesAddresses = make([]ethcommon.Address, 0, len(dependenciesAddrs))
		for _, addr := range dependenciesAddrs {
			ip.dependenciesAddresses = append(ip.dependenciesAddresses, ethcommon.HexToAddress(addr))
		}
	}

	if ip.contractProxyAddr == "" || !ethcommon.IsHexAddress(ip.contractProxyAddr) {
		return fmt.Errorf("invalid --%s flag", contractProxyAddrFlag)
	}

	if ip.evmPrivateKey == "" && ip.privateKeyConfig == "" {
		return fmt.Errorf("specify at least one: --%s or --%s", evmPrivateKeyFlag, privateKeyConfigFlag)
	}

	if ip.clone && !common.IsValidHTTPURL(ip.repositoryURL) {
		return fmt.Errorf("invalid --%s flag", repositoryURLFlag)
	}

	return nil
}

func (ip *setDependenciesParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&ip.contractName,
		contractNameFlag,
		"",
		contractNameFlagDesc,
	)
	cmd.Flags().StringVar(
		&ip.contractDir,
		contractDirFlag,
		"",
		contractDirFlagDesc,
	)
	cmd.Flags().StringVar(
		&ip.dependenciesAddressesStr,
		dependenciesAddressesFlag,
		"",
		dependenciesAddressesFlagDesc,
	)
	cmd.Flags().StringVar(
		&ip.contractProxyAddr,
		contractProxyAddrFlag,
		"",
		contractProxyAddrFlagDesc,
	)
	cmd.Flags().StringVar(
		&ip.evmPrivateKey,
		evmPrivateKeyFlag,
		"",
		evmPrivateKeyFlagDesc,
	)
	cmd.Flags().StringVar(
		&ip.privateKeyConfig,
		privateKeyConfigFlag,
		"",
		privateKeyConfigFlagDesc,
	)
	cmd.Flags().StringVar(
		&ip.evmNodeURL,
		evmNodeURLFlag,
		"",
		evmNodeURLFlagDesc,
	)
	cmd.Flags().BoolVar(
		&ip.evmDynamicTx,
		evmDynamicTxFlag,
		false,
		evmDynamicTxFlagDesc,
	)

	cmd.Flags().BoolVar(
		&ip.clone,
		evmCloneEvmRepoFlag,
		false,
		evmCloneEvmRepoFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.branchName,
		evmBranchNameFlag,
		"main",
		evmBranchNameFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.repositoryURL,
		repositoryURLFlag,
		"",
		repositoryURLFlagDesc,
	)

	cmd.Flags().Uint64Var(
		&ip.gasLimit,
		gasLimitFlag,
		defaultGasLimitValue,
		gasLimitFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(evmPrivateKeyFlag, privateKeyConfigFlag)
}

func (ip *setDependenciesParams) Execute(
	outputter common.OutputFormatter,
) (common.ICommandResult, error) {
	ctx := context.Background()
	contractDir := filepath.Clean(ip.contractDir)

	_, _ = outputter.Write(fmt.Appendf(nil, "Building the smart contracts from %s repository has started...", ip.contractDir)) //nolint:lll
	outputter.WriteOutput()

	if ip.clone {
		newContractDir, err := cloneAndBuildRepo(contractDir, ip.repositoryURL, ip.branchName, outputter)
		if err != nil {
			return nil, err
		}

		contractDir = newContractDir
	} else {
		if err := buildRepo(contractDir, outputter); err != nil {
			return nil, err
		}
	}

	artifacts, err := ethcontracts.LoadArtifacts(contractDir, []string{ip.contractName}...)
	if err != nil {
		return nil, err
	}

	wallet, err := eth.GetEthWalletForBladeAdmin(false, ip.evmPrivateKey, ip.privateKeyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create smart contracts admin wallet: %w", err)
	}

	txHelper, err := ethtxhelper.NewEThTxHelper(
		ethtxhelper.WithNodeURL(ip.evmNodeURL),
		ethtxhelper.WithDynamicTx(ip.evmDynamicTx),
		ethtxhelper.WithDefaultGasLimit(ip.gasLimit),
		ethtxhelper.WithNonceStrategyType(ethtxhelper.NonceInMemoryStrategy),
		ethtxhelper.WithZeroGasPrice(false),
		ethtxhelper.WithGasFeeMultiplier(defaultGasFeeMultiplier),
		ethtxhelper.WithInitClientAndChainIDFn(ctx),
	)
	if err != nil {
		return nil, err
	}

	ethContractUtils := ethcontracts.NewEthContractUtils(txHelper, wallet, defaultGasLimitMultiplier)

	if len(ip.dependenciesAddresses) == 0 {
		_, _ = outputter.Write([]byte("No dependencies provided; nothing to set."))
		outputter.WriteOutput()

		return &cmdResult{}, nil
	}

	dependencies := make([]any, len(ip.dependenciesAddresses))
	for i, addr := range ip.dependenciesAddresses {
		dependencies[i] = addr
	}

	txInfo, err := ethContractUtils.ExecuteMethod(
		ctx, artifacts[ip.contractName], ethcommon.HexToAddress(ip.contractProxyAddr), "setDependencies", dependencies...)
	if err != nil {
		return nil, fmt.Errorf("setDependecies for %s has been failed: %w", ip.contractName, err)
	}

	_, _ = outputter.Write(fmt.Appendf(
		nil, "%s setDependencies transaction has been sent. Waiting for the receipt...", ip.contractName))

	outputter.WriteOutput()

	receipts, err := ethtxhelper.WaitForTransactions(ctx, txHelper, txInfo.Hash().String())
	if err != nil {
		return nil, err
	}

	for _, receipt := range receipts {
		_, _ = outputter.Write(fmt.Appendf(
			nil, "%s setDependencies transaction has been sent. txHash: %s", ip.contractName, receipt.TxHash.String()))

		outputter.WriteOutput()
	}

	return &cmdResult{}, nil
}
