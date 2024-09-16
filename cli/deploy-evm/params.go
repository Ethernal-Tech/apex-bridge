package clideployevm

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethcontracts "github.com/Ethernal-Tech/apex-bridge/eth/contracts"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
)

const (
	defaultGasFeeMultiplier   = 200 // 170%
	defaultGasLimit           = uint64(5_242_880)
	defaultGasLimitMultiplier = float64(1.7)

	bridgeNodeURLFlag    = "bridge-url"
	bridgeSCAddrFlag     = "bridge-addr"
	evmNodeURLFlag       = "url"
	evmPrivateKeyFlag    = "key"
	evmCompiledSCDirFlag = "dir"
	evmChainIDFlag       = "chain"
	evmDynamicTxFlag     = "dynamic-tx"
	evmCloneEvmRepoFlag  = "clone"
	evmBranchNameFlag    = "branch"

	bridgeNodeURLFlagDesc    = "bridge node url"
	bridgeSCAddrFlagDesc     = "bridge smart contract address"
	evmNodeURLFlagDesc       = "evm node url"
	evmCompiledSCDirFlagDesc = "compiled evm smart contracts directory (json files)"
	evmPrivateKeyFlagDesc    = "private key for evm chain"
	evmChainIDFlagDesc       = "evm chain ID (prime, vector, etc)"
	evmDynamicTxFlagDesc     = "dynamic tx"
	evmCloneEvmRepoFlagDesc  = "clone evm gateway repository and build smart contracts"
	evmBranchNameFlagDesc    = "branch to use if the evm gateway repository is cloned"

	defaultBridgeSCAddr = "0xABEF000000000000000000000000000000000005"
	defaultEVMChainID   = common.ChainIDStrNexus

	evmGatewayRepositoryName        = "apex-evm-gateway"
	evmGatewayRepositoryURL         = "https://github.com/Ethernal-Tech/" + evmGatewayRepositoryName
	evmGatewayRepositoryArtifactDir = "artifacts"
)

type deployEVMParams struct {
	bridgeNodeURL string
	bridgeSCAddr  string
	evmNodeURL    string
	evmPrivateKey string
	evmDir        string
	evmClone      bool
	evmChainID    string
	evmBranchName string
	evmDynamicTx  bool
}

func (ip *deployEVMParams) validateFlags() error {
	if !common.IsValidHTTPURL(ip.bridgeNodeURL) {
		return fmt.Errorf("invalid --%s flag", bridgeNodeURLFlag)
	}

	if !common.IsValidHTTPURL(ip.evmNodeURL) {
		return fmt.Errorf("invalid --%s flag", evmNodeURLFlag)
	}

	if !ethcommon.IsHexAddress(ip.bridgeSCAddr) {
		return fmt.Errorf("invalid --%s flag", bridgeSCAddrFlag)
	}

	if !common.IsExistingChainID(ip.evmChainID) {
		return fmt.Errorf("unexisting chain: %s", ip.evmChainID)
	}

	if ip.evmPrivateKey == "" {
		return fmt.Errorf("invalid --%s flag", evmChainIDFlag)
	}

	return nil
}

func (ip *deployEVMParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&ip.bridgeNodeURL,
		bridgeNodeURLFlag,
		"",
		bridgeNodeURLFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.bridgeSCAddr,
		bridgeSCAddrFlag,
		defaultBridgeSCAddr,
		bridgeSCAddrFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.evmNodeURL,
		evmNodeURLFlag,
		"",
		evmNodeURLFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.evmPrivateKey,
		evmPrivateKeyFlag,
		"",
		evmPrivateKeyFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.evmDir,
		evmCompiledSCDirFlag,
		"",
		evmCompiledSCDirFlagDesc,
	)

	cmd.Flags().BoolVar(
		&ip.evmClone,
		evmCloneEvmRepoFlag,
		false,
		evmCloneEvmRepoFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.evmChainID,
		evmChainIDFlag,
		defaultEVMChainID,
		evmChainIDFlagDesc,
	)

	cmd.Flags().BoolVar(
		&ip.evmDynamicTx,
		evmDynamicTxFlag,
		false,
		evmDynamicTxFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.evmBranchName,
		evmBranchNameFlag,
		"main",
		evmBranchNameFlagDesc,
	)
}

func (ip *deployEVMParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()
	dir := filepath.Clean(ip.evmDir)

	if ip.evmClone {
		newDir, err := cloneSmartContract(dir, ip.evmBranchName, outputter)
		if err != nil {
			return nil, err
		}

		dir = newDir
	}

	artifacts, err := ethcontracts.LoadArtifacts(
		dir, "ERC1967Proxy", "Gateway", "NativeTokenPredicate", "NativeTokenWallet", "Validators")
	if err != nil {
		return nil, err
	}

	validatorsData, err := ip.getValidatorsChainData(ctx)
	if err != nil {
		return nil, err
	}

	wallet, err := ethtxhelper.NewEthTxWallet(ip.evmPrivateKey)
	if err != nil {
		return nil, err
	}

	txHelper, err := ethtxhelper.NewEThTxHelper(
		ethtxhelper.WithNodeURL(ip.evmNodeURL),
		ethtxhelper.WithDynamicTx(ip.evmDynamicTx),
		ethtxhelper.WithDefaultGasLimit(defaultGasLimit),
		ethtxhelper.WithZeroGasPrice(false),
		ethtxhelper.WithGasFeeMultiplier(defaultGasFeeMultiplier),
	)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Deploying the smart contracts has started..."))
	outputter.WriteOutput()

	gatewayProxyAddr, gatewayAddr, err := ethcontracts.DeployContractWithProxy(
		ctx, txHelper, wallet, artifacts["Gateway"], artifacts["ERC1967Proxy"])
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Gateway has been deployed"))
	outputter.WriteOutput()

	nativeTokenPredicateProxyAddr, nativeTokenPredicateAddr, err := ethcontracts.DeployContractWithProxy(
		ctx, txHelper, wallet, artifacts["NativeTokenPredicate"], artifacts["ERC1967Proxy"])
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("NativeTokenPredicate has been deployed"))
	outputter.WriteOutput()

	nativeTokenWalletProxyAddr, nativeTokenWalletAddr, err := ethcontracts.DeployContractWithProxy(
		ctx, txHelper, wallet, artifacts["NativeTokenWallet"], artifacts["ERC1967Proxy"])
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("NativeTokenWallet has been deployed"))
	outputter.WriteOutput()

	validatorsProxyAddr, validatorsAddr, err := ethcontracts.DeployContractWithProxy(
		ctx, txHelper, wallet, artifacts["Validators"], artifacts["ERC1967Proxy"])
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Validators has been deployed"))
	outputter.WriteOutput()

	_, err = ethcontracts.ExecuteContractMethod(
		ctx, txHelper, wallet, artifacts["Gateway"], defaultGasLimitMultiplier, true,
		gatewayProxyAddr, "setDependencies", nativeTokenPredicateProxyAddr, validatorsProxyAddr)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Gateway has been initialized"))
	outputter.WriteOutput()

	_, err = ethcontracts.ExecuteContractMethod(
		ctx, txHelper, wallet, artifacts["NativeTokenPredicate"], defaultGasLimitMultiplier, true,
		nativeTokenPredicateProxyAddr, "setDependencies", gatewayProxyAddr, nativeTokenWalletProxyAddr)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("NativeTokenPredicate has been initialized"))
	outputter.WriteOutput()

	_, err = ethcontracts.ExecuteContractMethod(
		ctx, txHelper, wallet, artifacts["NativeTokenWallet"], defaultGasLimitMultiplier, true,
		nativeTokenWalletProxyAddr, "setDependencies", nativeTokenPredicateProxyAddr)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("NativeTokenWallet has been initialized"))
	outputter.WriteOutput()

	_, err = ethcontracts.ExecuteContractMethod(
		ctx, txHelper, wallet, artifacts["Validators"], defaultGasLimitMultiplier, true,
		validatorsProxyAddr, "setValidatorsChainData", validatorsData)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Validators has been initialized"))
	outputter.WriteOutput()

	return &CmdResult{
		gatewayProxyAddr:              gatewayProxyAddr.String(),
		gatewayAddr:                   gatewayAddr.String(),
		nativeTokenPredicateProxyAddr: nativeTokenPredicateProxyAddr.String(),
		nativeTokenPredicateAddr:      nativeTokenPredicateAddr.String(),
		nativeTokenWalletProxyAddr:    nativeTokenWalletProxyAddr.String(),
		nativeTokenWalletAddr:         nativeTokenWalletAddr.String(),
		validatorsProxyAddr:           validatorsProxyAddr.String(),
		validatorsAddr:                validatorsAddr.String(),
	}, nil
}

func (ip *deployEVMParams) getValidatorsChainData(ctx context.Context) ([]eth.ValidatorChainData, error) {
	bridgeSC := eth.NewBridgeSmartContract(ip.bridgeNodeURL, ip.bridgeSCAddr, false, hclog.NewNullLogger())

	return bridgeSC.GetValidatorsChainData(ctx, ip.evmChainID)
}

func executeCLICommand(binary string, args []string, workingDir string, envVariables ...string) (string, error) {
	var (
		stdErrBuffer bytes.Buffer
		stdOutBuffer bytes.Buffer
	)

	cmd := exec.Command(binary, args...)
	cmd.Stderr = &stdErrBuffer
	cmd.Stdout = &stdOutBuffer
	cmd.Dir = workingDir

	cmd.Env = append(os.Environ(), envVariables...)

	err := cmd.Run()

	if stdErrBuffer.Len() > 0 {
		return "", fmt.Errorf("error while executing command: %s", stdErrBuffer.String())
	} else if err != nil {
		return "", err
	}

	return stdOutBuffer.String(), nil
}

func cloneSmartContract(dir, evmBranchName string, outputter common.OutputFormatter) (string, error) {
	_, _ = outputter.Write([]byte("Cloning and building the smart contracts repository has started..."))
	outputter.WriteOutput()

	if _, err := executeCLICommand(
		"git", []string{"clone", "--progress", evmGatewayRepositoryURL}, dir); err != nil {
		// git clone writes to stderror, check if messages are ok...
		// or there is already
		str := strings.TrimSpace(err.Error())
		if !strings.Contains(str, "Cloning into") && !strings.HasSuffix(str, "done.") &&
			!strings.Contains(str, fmt.Sprintf("'%s' already exists", evmGatewayRepositoryName)) {
			return "", err
		}
	}

	dir = filepath.Join(dir, evmGatewayRepositoryName)

	// do not listen for errors on following commands
	_, _ = executeCLICommand("git", []string{"checkout", evmBranchName}, dir)
	_, _ = executeCLICommand("git", []string{"pull", "origin"}, dir)
	_, _ = executeCLICommand("npm", []string{"install"}, dir)

	if _, err := executeCLICommand("npx", []string{"hardhat", "compile"}, dir); err != nil {
		return "", err
	}

	return filepath.Join(dir, evmGatewayRepositoryArtifactDir), nil
}
