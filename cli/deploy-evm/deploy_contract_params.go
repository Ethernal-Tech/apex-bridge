package clideployevm

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethcontracts "github.com/Ethernal-Tech/apex-bridge/eth/contracts"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

const (
	contractNameFlag          = "contract-name"
	contractDirFlag           = "contract-dir"
	dependenciesAddressesFlag = "dependencies"
	contractOwnerFlag         = "owner"
	upgradeAdminFlag          = "upgrade-admin"

	contractNameFlagDesc          = "name of the smart contract to deploy"
	contractDirFlagDesc           = "the directory where the sc repository is cloned"
	dependenciesAddressesFlagDesc = "addresses of dependency contracts, separated by semicolons"
	contractOwnerFlagDesc         = "address of the contract's owner"
	upgradeAdminFlagDesc          = "address of the contract's upgrade admin"
)

type deployContractParams struct {
	contractName             string
	contractDir              string
	dependenciesAddressesStr string
	dependenciesAddresses    []ethcommon.Address

	contractOwner    string
	upgradeAdmin     string
	evmPrivateKey    string
	privateKeyConfig string

	repositoryURL string
	clone         bool
	branchName    string

	evmNodeURL   string
	evmDynamicTx bool
	gasLimit     uint64
}

func (ip *deployContractParams) validateFlags() error {
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

	if ip.contractOwner == "" {
		return fmt.Errorf("contract owner not specified: --%s", contractOwnerFlag)
	}

	if ip.upgradeAdmin == "" {
		return fmt.Errorf("upgrade admin not specified: --%s", upgradeAdminFlag)
	}

	if ip.evmPrivateKey == "" && ip.privateKeyConfig == "" {
		return fmt.Errorf("specify at least one: --%s or --%s", evmPrivateKeyFlag, privateKeyConfigFlag)
	}

	if ip.clone && !common.IsValidHTTPURL(ip.repositoryURL) {
		return fmt.Errorf("invalid --%s flag", repositoryURLFlag)
	}

	return nil
}

func (ip *deployContractParams) setFlags(cmd *cobra.Command) {
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
		&ip.contractOwner,
		contractOwnerFlag,
		"",
		contractOwnerFlagDesc,
	)
	cmd.Flags().StringVar(
		&ip.upgradeAdmin,
		upgradeAdminFlag,
		"",
		upgradeAdminFlagDesc,
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

func (ip *deployContractParams) Execute(
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

	artifacts, err := ethcontracts.LoadArtifacts(contractDir, []string{ercProxyContractName, ip.contractName}...)
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

	_, _ = outputter.Write([]byte("Deploying the smart contracts has started..."))
	outputter.WriteOutput()

	ethContractUtils := ethcontracts.NewEthContractUtils(txHelper, wallet, defaultGasLimitMultiplier)

	proxyTx, tx, err := ethContractUtils.DeployWithProxy(
		ctx, artifacts[ip.contractName], artifacts[ercProxyContractName], ip.getInitParams()...)
	if err != nil {
		return nil, fmt.Errorf("deploy %s has been failed: %w", ip.contractName, err)
	}

	_, _ = outputter.Write(fmt.Appendf(nil, "%s has been sent", ip.contractName))
	outputter.WriteOutput()

	contracts := []contractInfo{
		{Name: ip.contractName, Addr: proxyTx.Address, IsProxy: true},
		{Name: ip.contractName, Addr: tx.Address},
	}

	txHashes := []string{proxyTx.Hash, tx.Hash}

	_, _ = outputter.Write([]byte("Waiting for receipts..."))
	outputter.WriteOutput()

	if _, err = ethtxhelper.WaitForTransactions(ctx, txHelper, txHashes...); err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Transactions have been included in the blockchain. Initializing contracts..."))
	outputter.WriteOutput()

	dependencies := make([]any, len(ip.dependenciesAddresses))
	for i, addr := range ip.dependenciesAddresses {
		dependencies[i] = addr
	}

	txInfo, err := ethContractUtils.ExecuteMethod(
		ctx, artifacts[ip.contractName], proxyTx.Address, "setDependencies", dependencies...)
	if err != nil {
		return nil, fmt.Errorf("setDependecies for %s has been failed: %w", ip.contractName, err)
	}

	_, _ = outputter.Write(fmt.Appendf(nil, "%s initialization transaction has been sent", ip.contractName))
	outputter.WriteOutput()

	_, err = ethtxhelper.WaitForTransactions(ctx, txHelper, txInfo.Hash().String())
	if err != nil {
		return nil, err
	}

	return &cmdResult{
		Contracts: contracts,
	}, nil
}

func cloneAndBuildRepo(
	contractDir string, repositoryURL, branchName string, outputter common.OutputFormatter,
) (string, error) {
	lastSlashIndex := strings.LastIndex(strings.TrimSuffix(repositoryURL, "/"), "/")
	if lastSlashIndex == -1 {
		return "", fmt.Errorf("invalid repository url: %s", repositoryURL)
	}

	repositoryName := repositoryURL[lastSlashIndex+1:]

	_, _ = outputter.Write(fmt.Appendf(nil, "Cloning and building the smart contracts repository %s has been started...", repositoryURL)) //nolint:lll
	outputter.WriteOutput()

	newDir, err := ethcontracts.CloneAndBuildContracts(
		contractDir, repositoryURL, repositoryName, evmRepositoryArtifactsDir, branchName)
	if err != nil {
		return "", fmt.Errorf("failed to clone and build contracts: %w", err)
	}

	return newDir, nil
}

func buildRepo(contractDir string, outputter common.OutputFormatter) error {
	if _, err := common.ExecuteCLICommand("npm", []string{"install"}, contractDir); err != nil {
		_, _ = outputter.Write(fmt.Appendf(nil, "Failed to execute npm install: %s", err.Error()))
		outputter.WriteOutput()
	}

	if _, err := common.ExecuteCLICommand("npx", []string{"hardhat", "compile"}, contractDir); err != nil {
		return fmt.Errorf("failed to compile smart contracts: %w", err)
	}

	return nil
}

func parseAddresses(input string) ([]string, error) {
	addresses := strings.Split(input, ";")
	validated := make([]string, 0, len(addresses))

	for _, addr := range addresses {
		addr = strings.TrimSpace(addr)

		if !ethcommon.IsHexAddress(addr) {
			return nil, fmt.Errorf("invalid address: %s", addr)
		}

		validated = append(validated, addr)
	}

	return validated, nil
}

func (ip *deployContractParams) getInitParams() []any {
	return []any{
		ethcommon.HexToAddress(ip.contractOwner),
		ethcommon.HexToAddress(ip.upgradeAdmin),
	}
}
