package clideployevm

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethcontracts "github.com/Ethernal-Tech/apex-bridge/eth/contracts"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
)

const (
	apexBridgeSmartContracts = "apex-bridge-smartcontracts"

	nodeFlag              = "url"
	contractFlag          = "contract"
	repositoryURLFlag     = "repo"
	minFeeAmountFlag      = "min-fee"
	minBridgingAmountFlag = "min-bridging-amount"

	nodeFlagDessc             = "node url"
	contractFlagDesc          = "contractName:proxyAddr[:updateFunctionName] contract name is solidity file name, proxyAddr is address or proxy contract" //nolint:lll
	repositoryURLFlagDesc     = "smart contracts github repository url"
	minFeeAmountFlagDesc      = "minimal fee amount"
	minBridgingAmountFlagDesc = "minimal amount to bridge"
)

type upgradeEVMParams struct {
	nodeURL       string
	privateKey    string
	dir           string
	repositoryURL string
	clone         bool
	branchName    string
	dynamicTx     bool
	contracts     []string

	minFeeString            string
	minBridgingAmountString string

	gatewayInitParams *gatewayInitParams
	ethTxHelper       ethtxhelper.IEthTxHelper
}

func (ip *upgradeEVMParams) validateFlags() error {
	if !common.IsValidHTTPURL(ip.nodeURL) {
		return fmt.Errorf("invalid --%s flag", evmNodeURLFlag)
	}

	if ip.privateKey == "" {
		return fmt.Errorf("not specified --%s flag", evmPrivateKeyFlag)
	}

	if len(ip.contracts) == 0 {
		return fmt.Errorf("not specified --%s flag", contractFlag)
	}

	if ip.clone && !common.IsValidHTTPURL(ip.repositoryURL) {
		return fmt.Errorf("invalid --%s flag", repositoryURLFlag)
	}

	gatewayInitParams, err := validateAndSetGatewayParams(
		ip.minFeeString, ip.minBridgingAmountString,
	)
	if err != nil {
		return err
	}

	ip.gatewayInitParams = gatewayInitParams

	ethTxHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithNodeURL(ip.nodeURL))
	if err != nil {
		return fmt.Errorf("failed to connect to the bridge node: %w", err)
	}

	ip.ethTxHelper = ethTxHelper

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
		"audit/APEX-472",
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

	cmd.Flags().StringVar(
		&ip.minFeeString,
		minFeeAmountFlag,
		"",
		minFeeAmountFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.minBridgingAmountString,
		minBridgingAmountFlag,
		"",
		minBridgingAmountFlagDesc,
	)
}

func (ip *upgradeEVMParams) Execute(
	outputter common.OutputFormatter,
) (common.ICommandResult, error) {
	dir := filepath.Clean(ip.dir)
	ctx := context.Background()
	contracts := make([]string, len(ip.contracts))
	proxyAddrs := make([]ethcommon.Address, len(ip.contracts))
	updateFuncs := make([]string, len(ip.contracts))

	for i, x := range ip.contracts {
		ss := strings.Split(x, ":")
		if len(ss) != 2 && len(ss) != 3 {
			return nil, fmt.Errorf("invalid --%s number %d", contractFlag, i)
		}

		if ss[0] == "" {
			return nil, fmt.Errorf("invalid contract name for --%s number %d", contractFlag, i)
		}

		if !common.IsValidAddress(common.ChainIDStrNexus, ss[1]) {
			return nil, fmt.Errorf("invalid address for --%s number %d", contractFlag, i)
		}

		if len(ss) > 2 {
			updateFuncs[i] = ss[2] // empty function names will be skipped
		}

		contracts[i] = ss[0]
		proxyAddrs[i] = common.HexToAddress(ss[1])
	}

	if ip.clone {
		_, _ = outputter.Write([]byte("Cloning and building the smart contracts repository has started..."))
		outputter.WriteOutput()

		lastSlashIndex := strings.LastIndex(strings.TrimSuffix(ip.repositoryURL, "/"), "/")
		if lastSlashIndex == -1 {
			return nil, fmt.Errorf("invalid --%s", repositoryURLFlag)
		}

		repositoryName := ip.repositoryURL[lastSlashIndex+1:]

		newDir, err := ethcontracts.CloneAndBuildContracts(
			dir, ip.repositoryURL, repositoryName, evmRepositoryArtifactsDir, ip.branchName)
		if err != nil {
			return nil, err
		}

		dir = newDir
	}

	artifacts, err := ethcontracts.LoadArtifacts(
		dir, contracts...)
	if err != nil {
		return nil, err
	}

	wallet, err := ethtxhelper.NewEthTxWallet(ip.privateKey)
	if err != nil {
		return nil, err
	}

	txHelper, err := ethtxhelper.NewEThTxHelper(
		ethtxhelper.WithNodeURL(ip.nodeURL),
		ethtxhelper.WithDynamicTx(ip.dynamicTx),
		ethtxhelper.WithDefaultGasLimit(defaultGasLimit),
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
			initializationData, err = artifacts[contractName].Abi.Pack(fn)
			if err != nil {
				return nil, fmt.Errorf("upgrade %s has been failed: %w", contractName, err)
			}
		}

		tx, deployTxInfo, err := ethContractUtils.Upgrade(
			ctx, artifacts[contractName], proxyAddrs[i], initializationData)
		if err != nil {
			return nil, fmt.Errorf("upgrade %s has been failed: %w", contractName, err)
		}

		_, _ = outputter.Write([]byte(fmt.Sprintf("%s upgrade has been sent", contractName)))
		outputter.WriteOutput()

		txHashes[i*2] = tx.Hash().String()
		txHashes[i*2+1] = deployTxInfo.Hash
		contractInfos[i] = contractInfo{
			Name: contractName,
			Addr: deployTxInfo.Address,
		}

		if contractName == Gateway {
			transaction, err := infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (*types.Transaction, error) {
				contract, err := contractbinding.NewGateway(
					(contractInfos[i].Addr), ip.ethTxHelper.GetClient())
				if err != nil {
					return nil, fmt.Errorf("failed to connect to gateway smart contract: %w", err)
				}

				return ip.ethTxHelper.SendTx(
					ctx,
					wallet,
					bind.TransactOpts{
						GasLimit: defaultGasLimit,
					},
					func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
						return contract.SetMinAmounts(
							txOpts,
							ip.gatewayInitParams.minFeeAmount,
							ip.gatewayInitParams.minBridgingAmount,
						)
					},
				)
			})
			if err != nil {
				return nil, err
			}

			_, _ = outputter.Write([]byte(fmt.Sprintf("transaction has been submitted: %s", transaction.Hash())))
			outputter.WriteOutput()
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
