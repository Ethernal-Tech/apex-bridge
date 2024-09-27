package clideployevm

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethcontracts "github.com/Ethernal-Tech/apex-bridge/eth/contracts"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/bn256"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
)

const (
	defaultGasFeeMultiplier   = 200 // 170%
	defaultGasLimit           = uint64(5_242_880)
	defaultGasLimitMultiplier = float64(1.1)

	evmNodeURLFlag      = "url"
	evmSCDirFlag        = "dir"
	evmPrivateKeyFlag   = "key"
	evmBlsKeyFlag       = "bls-key"
	evmChainIDFlag      = "chain"
	evmDynamicTxFlag    = "dynamic-tx"
	evmCloneEvmRepoFlag = "clone"
	evmBranchNameFlag   = "branch"
	bridgeNodeURLFlag   = "bridge-url"
	bridgeSCAddrFlag    = "bridge-addr"

	evmNodeURLFlagDesc      = "evm node url"
	evmSCDirFlagDesc        = "the directory where the repository will be cloned, or the directory where the compiled evm smart contracts (JSON files) are located." //nolint:lll
	evmPrivateKeyFlagDesc   = "private key for evm chain"
	evmBlsKeyFlagDesc       = "bls key of the bridge validator. it can be used multiple times, but the order must be the same as on the bridge" //nolint:lll
	evmChainIDFlagDesc      = "evm chain ID (prime, vector, etc)"
	evmDynamicTxFlagDesc    = "dynamic tx"
	evmCloneEvmRepoFlagDesc = "clone evm gateway repository and build smart contracts"
	evmBranchNameFlagDesc   = "branch to use if the evm gateway repository is cloned"
	bridgeNodeURLFlagDesc   = "bridge node url"
	bridgeSCAddrFlagDesc    = "bridge smart contract address"

	defaultBridgeSCAddr = "0xABEF000000000000000000000000000000000005"
	defaultEVMChainID   = common.ChainIDStrNexus

	evmGatewayRepositoryName         = "apex-evm-gateway"
	evmGatewayRepositoryURL          = "https://github.com/Ethernal-Tech/" + evmGatewayRepositoryName
	evmGatewayRepositoryArtifactsDir = "artifacts"
)

type deployEVMParams struct {
	evmNodeURL    string
	evmPrivateKey string
	evmDir        string
	evmBlsKeys    []string
	evmClone      bool
	evmChainID    string
	evmBranchName string
	evmDynamicTx  bool

	bridgeNodeURL string
	bridgeSCAddr  string
}

func (ip *deployEVMParams) validateFlags() error {
	if !common.IsValidHTTPURL(ip.evmNodeURL) {
		return fmt.Errorf("invalid --%s flag", evmNodeURLFlag)
	}

	if !common.IsExistingChainID(ip.evmChainID) {
		return fmt.Errorf("unexisting chain: %s", ip.evmChainID)
	}

	if ip.evmPrivateKey == "" {
		return fmt.Errorf("invalid --%s flag", evmChainIDFlag)
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

	return nil
}

func (ip *deployEVMParams) setFlags(cmd *cobra.Command) {
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
		defaultBridgeSCAddr,
		bridgeSCAddrFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(bridgeNodeURLFlag, evmBlsKeyFlag)
	cmd.MarkFlagsMutuallyExclusive(bridgeSCAddrFlag, evmBlsKeyFlag)
}

func (ip *deployEVMParams) Execute(
	ctx context.Context, outputter common.OutputFormatter,
) (common.ICommandResult, error) {
	dir := filepath.Clean(ip.evmDir)

	if ip.evmClone {
		_, _ = outputter.Write([]byte("Cloning and building the smart contracts repository has started..."))
		outputter.WriteOutput()

		newDir, err := ethcontracts.CloneAndBuildContracts(
			dir, evmGatewayRepositoryURL, evmGatewayRepositoryName, evmGatewayRepositoryArtifactsDir, ip.evmBranchName)
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

	validatorsData, err := ip.getValidatorsChainData(ctx)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Deploying the smart contracts has started..."))
	outputter.WriteOutput()

	ethContractUtils := ethcontracts.NewEthContractUtils(txHelper, wallet, defaultGasLimitMultiplier)

	gatewayProxyAddr, gatewayProxyTxHash, gatewayAddr, gatewayTxHash, err := ethContractUtils.DeployWithProxy(
		ctx, artifacts["Gateway"], artifacts["ERC1967Proxy"])
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Gateway has been sent"))
	outputter.WriteOutput()

	predicateProxyAddr, predicateProxyTxHash, predicateAddr, predicateTxHash, err := ethContractUtils.DeployWithProxy(
		ctx, artifacts["NativeTokenPredicate"], artifacts["ERC1967Proxy"])
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("NativeTokenPredicate has been sent"))
	outputter.WriteOutput()

	walletProxyAddr, walletProxyTxHash, walletAddr, walletTxHash, err := ethContractUtils.DeployWithProxy(
		ctx, artifacts["NativeTokenWallet"], artifacts["ERC1967Proxy"])
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("NativeTokenWallet has been sent"))
	outputter.WriteOutput()

	validatorsProxyAddr, validsProxyTxHash, validatorsAddr, validsTxHash, err := ethContractUtils.DeployWithProxy(
		ctx, artifacts["Validators"], artifacts["ERC1967Proxy"])
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Validators has been sent. Waiting for the receipts..."))
	outputter.WriteOutput()

	_, err = ethtxhelper.WaitForTransactions(ctx, txHelper,
		gatewayProxyTxHash, gatewayTxHash, predicateProxyTxHash, predicateTxHash,
		walletProxyTxHash, walletTxHash, validsProxyTxHash, validsTxHash)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Transactions have been included in the blockchain. Initializing contracts..."))
	outputter.WriteOutput()

	txHash1, err := ethContractUtils.ExecuteMethod(
		ctx, artifacts["Gateway"], gatewayProxyAddr, "setDependencies",
		predicateProxyAddr, validatorsProxyAddr)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Gateway initialization transaction has been sent"))
	outputter.WriteOutput()

	txHash2, err := ethContractUtils.ExecuteMethod(
		ctx, artifacts["NativeTokenPredicate"], predicateProxyAddr, "setDependencies",
		gatewayProxyAddr, walletProxyAddr)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("NativeTokenPredicate initialization transaction has been sent"))
	outputter.WriteOutput()

	txHash3, err := ethContractUtils.ExecuteMethod(
		ctx, artifacts["NativeTokenWallet"], walletProxyAddr, "setDependencies", predicateProxyAddr)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("NativeTokenWallet initialization transaction has been sent"))
	outputter.WriteOutput()

	txHash4, err := ethContractUtils.ExecuteMethod(
		ctx, artifacts["Validators"], validatorsProxyAddr, "setValidatorsChainData", validatorsData)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Validators initialization transaction has been sent. Waiting for the receipts..."))
	outputter.WriteOutput()

	_, err = ethtxhelper.WaitForTransactions(ctx, txHelper, txHash1, txHash2, txHash3, txHash4)
	if err != nil {
		return nil, err
	}

	return &CmdResult{
		gatewayProxyAddr:              gatewayProxyAddr.String(),
		gatewayAddr:                   gatewayAddr.String(),
		nativeTokenPredicateProxyAddr: predicateProxyAddr.String(),
		nativeTokenPredicateAddr:      predicateAddr.String(),
		nativeTokenWalletProxyAddr:    walletProxyAddr.String(),
		nativeTokenWalletAddr:         walletAddr.String(),
		validatorsProxyAddr:           validatorsProxyAddr.String(),
		validatorsAddr:                validatorsAddr.String(),
	}, nil
}

func (ip *deployEVMParams) getValidatorsChainData(ctx context.Context) ([]eth.ValidatorChainData, error) {
	if ip.bridgeNodeURL != "" {
		bridgeSC := eth.NewBridgeSmartContract(ip.bridgeNodeURL, ip.bridgeSCAddr, false, hclog.NewNullLogger())

		return bridgeSC.GetValidatorsChainData(ctx, ip.evmChainID)
	}

	result := make([]eth.ValidatorChainData, len(ip.evmBlsKeys))

	for i, x := range ip.evmBlsKeys {
		blsRaw, err := common.DecodeHex(x)
		if err != nil {
			return nil, err
		}

		key, err := bn256.UnmarshalPublicKey(blsRaw)
		if err != nil {
			return nil, err
		}

		result[i] = eth.ValidatorChainData{
			Key: key.ToBigInt(),
		}
	}

	return result, nil
}
