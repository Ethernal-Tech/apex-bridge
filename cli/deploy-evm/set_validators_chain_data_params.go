package clideployevm

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/Ethernal-Tech/apex-bridge/common"
	ethcontracts "github.com/Ethernal-Tech/apex-bridge/eth/contracts"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

const (
	validatorsProxyAddrFlag     = "validators-proxy-addr"
	validatorsProxyAddrFlagDesc = "Validators SC proxy addr"
)

type setValidatorsChainDataEVMParams struct {
	deployEVMParams

	validatorsProxyAddr string
}

func (ip *setValidatorsChainDataEVMParams) validateFlags() error {
	if !common.IsValidHTTPURL(ip.evmNodeURL) {
		return fmt.Errorf("invalid --%s flag", evmNodeURLFlag)
	}

	if ip.evmPrivateKey == "" {
		return fmt.Errorf("invalid --%s flag", evmPrivateKeyFlag)
	}

	if ip.bridgeNodeURL != "" {
		if !common.IsValidHTTPURL(ip.bridgeNodeURL) {
			return fmt.Errorf("invalid --%s flag", bridgeNodeURLFlag)
		}

		if !ethcommon.IsHexAddress(ip.bridgeSCAddr) {
			return fmt.Errorf("invalid --%s flag", bridgeSCAddrFlag)
		}
	} else if len(ip.evmBlsKeys) == 0 {
		return fmt.Errorf("bls keys not specified: --%s", evmBlsKeyFlag)
	}

	if ip.validatorsProxyAddr == "" || !ethcommon.IsHexAddress(ip.validatorsProxyAddr) {
		return fmt.Errorf("invalid --%s flag", validatorsProxyAddrFlag)
	}

	return nil
}

func (ip *setValidatorsChainDataEVMParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(
		&ip.evmBlsKeys,
		evmBlsKeyFlag,
		nil,
		evmBlsKeyFlagDesc,
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
		evmSCDirFlag,
		"",
		evmSCDirFlagDesc,
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
		&ip.validatorsProxyAddr,
		validatorsProxyAddrFlag,
		"",
		validatorsProxyAddrFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(bridgeNodeURLFlag, evmBlsKeyFlag)
	cmd.MarkFlagsMutuallyExclusive(bridgeSCAddrFlag, evmBlsKeyFlag)
}

func (ip *setValidatorsChainDataEVMParams) Execute(
	outputter common.OutputFormatter,
) (common.ICommandResult, error) {
	dir := filepath.Clean(ip.evmDir)
	ctx := context.Background()

	contractNames := []string{
		Validators,
	}

	if ip.evmClone {
		_, _ = outputter.Write([]byte("Cloning and building the smart contracts repository has started..."))
		outputter.WriteOutput()

		newDir, err := ethcontracts.CloneAndBuildContracts(
			dir, evmGatewayRepositoryURL, evmGatewayRepositoryName, evmRepositoryArtifactsDir, ip.evmBranchName)
		if err != nil {
			return nil, err
		}

		dir = newDir
	}

	artifacts, err := ethcontracts.LoadArtifacts(
		dir, append([]string{ercProxyContractName}, contractNames...)...)
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
		ethtxhelper.WithNonceStrategyType(ethtxhelper.NonceNodePendingStrategy),
		ethtxhelper.WithZeroGasPrice(false),
		ethtxhelper.WithGasFeeMultiplier(defaultGasFeeMultiplier),
		ethtxhelper.WithInitClientAndChainIDFn(ctx),
	)
	if err != nil {
		return nil, err
	}

	txHelperBridge, err := ip.getTxHelperBridge()
	if err != nil {
		return nil, err
	}

	validatorsData, err := ip.getValidatorsChainData(ctx, txHelperBridge, outputter)
	if err != nil {
		return nil, err
	}

	ethContractUtils := ethcontracts.NewEthContractUtils(txHelper, wallet, defaultGasLimitMultiplier)

	setValidatorsChainDataTx, err := ethContractUtils.ExecuteMethod(
		ctx, artifacts[Validators], ethcommon.HexToAddress(ip.validatorsProxyAddr), "setValidatorsChainData", validatorsData)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("setValidatorsChainData transaction has been sent. Waiting for the receipts..."))
	outputter.WriteOutput()

	_, err = ethtxhelper.WaitForTransactions(ctx, txHelper, setValidatorsChainDataTx.Hash().String())
	if err != nil {
		return nil, err
	}

	return &cmdResult{}, nil
}
