package clideployevm

import (
	"context"
	"fmt"
	"math/big"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethcontracts "github.com/Ethernal-Tech/apex-bridge/eth/contracts"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

const (
	apexBridgeSmartContracts = "apex-bridge-smartcontracts"

	nodeFlag          = "url"
	contractFlag      = "contract"
	repositoryURLFlag = "repo"

	nodeFlagDessc         = "node url"
	contractFlagDesc      = "contractName:proxyAddr[:updateFunctionName:args] contract name is solidity file name, proxyAddr is address or proxy contract" //nolint:lll
	repositoryURLFlagDesc = "smart contracts github repository url"
)

type upgradeEVMParams struct {
	nodeURL          string
	privateKey       string
	privateKeyConfig string

	dir           string
	repositoryURL string
	clone         bool
	branchName    string
	dynamicTx     bool
	contracts     []string
	gasLimit      uint64
}

func (ip *upgradeEVMParams) validateFlags() error {
	if !common.IsValidHTTPURL(ip.nodeURL) {
		return fmt.Errorf("invalid --%s flag", evmNodeURLFlag)
	}

	if ip.privateKey == "" && ip.privateKeyConfig == "" {
		return fmt.Errorf("specify at least one: --%s or --%s", evmPrivateKeyFlag, privateKeyConfigFlag)
	}

	if len(ip.contracts) == 0 {
		return fmt.Errorf("not specified --%s flag", contractFlag)
	}

	if ip.clone && !common.IsValidHTTPURL(ip.repositoryURL) {
		return fmt.Errorf("invalid --%s flag", repositoryURLFlag)
	}

	return nil
}

func (ip *upgradeEVMParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&ip.nodeURL,
		evmNodeURLFlag,
		"",
		evmNodeURLFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.privateKey,
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
		&ip.dir,
		evmSCDirFlag,
		"",
		evmSCDirFlagDesc,
	)

	cmd.Flags().BoolVar(
		&ip.clone,
		evmCloneEvmRepoFlag,
		false,
		evmCloneEvmRepoFlagDesc,
	)

	cmd.Flags().BoolVar(
		&ip.dynamicTx,
		evmDynamicTxFlag,
		false,
		evmDynamicTxFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.branchName,
		evmBranchNameFlag,
		"main",
		evmBranchNameFlagDesc,
	)

	cmd.Flags().StringSliceVar(
		&ip.contracts,
		contractFlag,
		nil,
		contractFlagDesc,
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

func (ip *upgradeEVMParams) Execute(
	outputter common.OutputFormatter,
) (common.ICommandResult, error) {
	dir := filepath.Clean(ip.dir)
	ctx := context.Background()
	contracts := make([]string, len(ip.contracts))
	proxyAddrs := make([]ethcommon.Address, len(ip.contracts))
	updateFuncs := make([]string, len(ip.contracts))
	updateFuncsArgs := make([]string, len(ip.contracts))

	for i, x := range ip.contracts {
		ss := strings.Split(x, ":")
		if n := len(ss); n < 2 || n > 4 {
			return nil, fmt.Errorf("invalid --%s number %d", contractFlag, i)
		}

		if ss[0] == "" {
			return nil, fmt.Errorf("invalid contract name for --%s number %d", contractFlag, i)
		}

		if !common.IsValidAddress(common.ChainIDStrNexus, ss[1]) {
			return nil, fmt.Errorf("invalid address for --%s number %d", contractFlag, i)
		}

		contracts[i] = ss[0]
		proxyAddrs[i] = common.HexToAddress(ss[1])

		if len(ss) > 2 {
			updateFuncs[i] = ss[2] // empty function names will be skipped
		}

		if len(ss) > 3 {
			updateFuncsArgs[i] = ss[3]
		}
	}

	if ip.clone {
		newContractDir, err := cloneAndBuildRepo(dir, ip.repositoryURL, ip.branchName, outputter)
		if err != nil {
			return nil, err
		}

		dir = newContractDir
	}

	artifacts, err := ethcontracts.LoadArtifacts(
		dir, contracts...)
	if err != nil {
		return nil, err
	}

	wallet, err := eth.GetEthWalletForBladeAdmin(true, ip.privateKey, ip.privateKeyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create smart contracts admin wallet: %w", err)
	}

	txHelper, err := ethtxhelper.NewEThTxHelper(
		ethtxhelper.WithNodeURL(ip.nodeURL),
		ethtxhelper.WithDynamicTx(ip.dynamicTx),
		ethtxhelper.WithDefaultGasLimit(ip.gasLimit),
		ethtxhelper.WithNonceStrategyType(ethtxhelper.NonceInMemoryStrategy),
		ethtxhelper.WithZeroGasPrice(strings.Contains(dir, apexBridgeSmartContracts)),
		ethtxhelper.WithGasFeeMultiplier(defaultGasFeeMultiplier),
		ethtxhelper.WithInitClientAndChainIDFn(ctx),
	)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Upgrading the smart contracts has started..."))
	outputter.WriteOutput()

	ethContractUtils := ethcontracts.NewEthContractUtils(txHelper, wallet, defaultGasLimitMultiplier)
	contractInfos := make([]contractInfo, len(contracts))
	txHashes := make([]string, len(contracts)*2)

	for i, contractName := range contracts {
		var initializationData []byte

		if fn := updateFuncs[i]; fn != "" {
			var args []any
			if argStr := updateFuncsArgs[i]; argStr != "" {
				args = parseFnArguments(argStr)
			}

			initializationData, err = artifacts[contractName].Abi.Pack(fn, args...)
			if err != nil {
				return nil, fmt.Errorf("upgrade %s has been failed: %w", contractName, err)
			}
		}

		tx, deployTxInfo, err := ethContractUtils.Upgrade(
			ctx, artifacts[contractName], proxyAddrs[i], initializationData)
		if err != nil {
			return nil, fmt.Errorf("upgrade %s has been failed: %w", contractName, err)
		}

		_, _ = outputter.Write(fmt.Appendf(nil, "%s upgrade has been sent", contractName))
		outputter.WriteOutput()

		txHashes[i*2] = tx.Hash().String()
		txHashes[i*2+1] = deployTxInfo.Hash
		contractInfos[i] = contractInfo{
			Name: contractName,
			Addr: deployTxInfo.Address,
		}
	}

	_, _ = outputter.Write([]byte("Waiting for receipts..."))
	outputter.WriteOutput()

	if _, err = ethtxhelper.WaitForTransactions(ctx, txHelper, txHashes...); err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Transactions have been included in the blockchain..."))
	outputter.WriteOutput()

	return &cmdResult{
		Contracts: contractInfos,
	}, nil
}

func parseFnArguments(input string) []any {
	inputArgs := strings.Split(input, ";")
	args := make([]any, len(inputArgs))

	for i, rawArg := range inputArgs {
		arg := strings.TrimSpace(rawArg)

		if ethcommon.IsHexAddress(arg) {
			args[i] = ethcommon.HexToAddress(arg)
		} else if isHash(arg) {
			args[i] = ethcommon.HexToHash(arg)
		} else if strings.EqualFold(arg, "true") {
			args[i] = true
		} else if strings.EqualFold(arg, "false") {
			args[i] = false
		} else if strings.HasSuffix(arg, "u") {
			numStr := strings.TrimSuffix(arg, "u")
			if n, err := strconv.ParseUint(numStr, 10, 8); err == nil {
				args[i] = uint8(n)
			} else {
				args[i] = arg // fallback to string if parsing fails
			}
		} else if n, err := strconv.ParseInt(arg, 10, 8); err == nil {
			args[i] = int8(n)
		} else if n, ok := new(big.Int).SetString(arg, 10); ok {
			args[i] = n
		} else {
			args[i] = arg
		}
	}

	return args
}

func isHash(s string) bool {
	has0xPrefix := strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X")

	return has0xPrefix && len(s[2:]) == 2*ethcommon.HashLength
}
