package clideployevm

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethcontracts "github.com/Ethernal-Tech/apex-bridge/eth/contracts"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/bn256"
	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
)

const (
	defaultGasFeeMultiplier   = 200 // 170%
	defaultGasLimit           = uint64(5_242_880)
	defaultGasLimitMultiplier = float64(1.1)

	ercProxyContractName = "ERC1967Proxy"

	evmNodeURLFlag      = "url"
	evmSCDirFlag        = "dir"
	evmPrivateKeyFlag   = "key"
	evmBlsKeyFlag       = "bls-key"
	evmChainIDFlag      = "chain"
	evmDynamicTxFlag    = "dynamic-tx"
	evmCloneEvmRepoFlag = "clone"
	evmBranchNameFlag   = "branch"

	bridgeNodeURLFlag    = "bridge-url"
	bridgeSCAddrFlag     = "bridge-addr"
	bridgePrivateKeyFlag = "bridge-key"

	evmNodeURLFlagDesc      = "evm node url"
	evmSCDirFlagDesc        = "the directory where the repository will be cloned, or the directory where the compiled evm smart contracts (JSON files) are located." //nolint:lll
	evmPrivateKeyFlagDesc   = "private key for evm chain"
	evmBlsKeyFlagDesc       = "bls key of the bridge validator. it can be used multiple times, but the order must be the same as on the bridge" //nolint:lll
	evmChainIDFlagDesc      = "evm chain ID (prime, vector, etc)"
	evmDynamicTxFlagDesc    = "dynamic tx"
	evmCloneEvmRepoFlagDesc = "clone evm gateway repository and build smart contracts"
	evmBranchNameFlagDesc   = "branch to use if the evm gateway repository is cloned"

	bridgeNodeURLFlagDesc    = "bridge node url"
	bridgeSCAddrFlagDesc     = "bridge smart contract address"
	bridgePrivateKeyFlagDesc = "private key for bridge wallet (proxy admin)"

	defaultEVMChainID = common.ChainIDStrNexus

	evmGatewayRepositoryName  = "apex-evm-gateway"
	evmGatewayRepositoryURL   = "https://github.com/Ethernal-Tech/" + evmGatewayRepositoryName
	evmRepositoryArtifactsDir = "artifacts"
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

	bridgeNodeURL    string
	bridgeSCAddr     string
	bridgePrivateKey string
}

func (ip *deployEVMParams) validateFlags() error {
	if !common.IsValidHTTPURL(ip.evmNodeURL) {
		return fmt.Errorf("invalid --%s flag", evmNodeURLFlag)
	}

	if !common.IsExistingReactorChainID(ip.evmChainID) {
		return fmt.Errorf("unexisting chain: %s", ip.evmChainID)
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
		"",
		bridgeSCAddrFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.bridgePrivateKey,
		bridgePrivateKeyFlag,
		"",
		bridgePrivateKeyFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(bridgeNodeURLFlag, evmBlsKeyFlag)
	cmd.MarkFlagsMutuallyExclusive(bridgeSCAddrFlag, evmBlsKeyFlag)
	cmd.MarkFlagsMutuallyExclusive(bridgePrivateKeyFlag, evmBlsKeyFlag)
}

func (ip *deployEVMParams) Execute(
	outputter common.OutputFormatter,
) (common.ICommandResult, error) {
	dir := filepath.Clean(ip.evmDir)
	ctx := context.Background()

	const (
		Gateway              = "Gateway"
		NativeTokenPredicate = "NativeTokenPredicate"
		NativeTokenWallet    = "NativeTokenWallet"
		Validators           = "Validators"
	)

	contractNames := []string{
		Gateway, NativeTokenPredicate, NativeTokenWallet, Validators,
	}
	setDependenciesData := map[string][]string{
		Gateway:              {NativeTokenPredicate, Validators},
		NativeTokenPredicate: {Gateway, NativeTokenWallet},
		NativeTokenWallet:    {NativeTokenPredicate},
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
		ethtxhelper.WithNonceStrategyType(ethtxhelper.NonceInMemoryStrategy),
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

	_, _ = outputter.Write([]byte("Deploying the smart contracts has started..."))
	outputter.WriteOutput()

	ethContractUtils := ethcontracts.NewEthContractUtils(txHelper, wallet, defaultGasLimitMultiplier)
	contracts := make([]contractInfo, len(contractNames)*2)
	txHashes := make([]string, len(contractNames)*2)
	addresses := make(map[string]ethcommon.Address, len(contractNames))

	for i, contractName := range contractNames {
		proxyTx, tx, err := ethContractUtils.DeployWithProxy(
			ctx, artifacts[contractName], artifacts[ercProxyContractName])
		if err != nil {
			return nil, fmt.Errorf("deploy %s has been failed: %w", contractName, err)
		}

		_, _ = outputter.Write([]byte(fmt.Sprintf("%s has been sent", contractName)))
		outputter.WriteOutput()

		txHashes[i*2] = proxyTx.Hash
		txHashes[i*2+1] = tx.Hash
		contracts[i*2] = contractInfo{
			Name:    contractName,
			Addr:    proxyTx.Address,
			IsProxy: true,
		}
		contracts[i*2+1] = contractInfo{
			Name: contractName,
			Addr: tx.Address,
		}
		addresses[contractName] = proxyTx.Address
	}

	_, _ = outputter.Write([]byte("Waiting for receipts..."))
	outputter.WriteOutput()

	if _, err = ethtxhelper.WaitForTransactions(ctx, txHelper, txHashes...); err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Transactions have been included in the blockchain. Initializing contracts..."))
	outputter.WriteOutput()

	additionalTxHashes := make([]string, 0, len(setDependenciesData)+1) // + 1 for setValidatorsChainData

	for contractName, dependencyNames := range setDependenciesData {
		dependencies := make([]interface{}, len(dependencyNames))
		for i, x := range dependencyNames {
			dependencies[i] = addresses[x]
		}

		txInfo, err := ethContractUtils.ExecuteMethod(
			ctx, artifacts[contractName], addresses[contractName], "setDependencies", dependencies...)
		if err != nil {
			return nil, fmt.Errorf("setDependecies for %s has been failed: %w", contractName, err)
		}

		_, _ = outputter.Write([]byte(fmt.Sprintf("%s initialization transaction has been sent", contractName)))
		outputter.WriteOutput()

		additionalTxHashes = append(additionalTxHashes, txInfo.Hash().String())
	}

	setValidatorsChainDataTx, err := ethContractUtils.ExecuteMethod(
		ctx, artifacts[Validators], addresses[Validators], "setValidatorsChainData", validatorsData)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Validators initialization transaction has been sent. Waiting for the receipts..."))
	outputter.WriteOutput()

	_, err = ethtxhelper.WaitForTransactions(ctx, txHelper,
		append(additionalTxHashes, setValidatorsChainDataTx.Hash().String())...)
	if err != nil {
		return nil, err
	}

	if err := ip.setChainAdditionalData(ctx, addresses[Gateway], txHelperBridge, outputter); err != nil {
		return nil, err
	}

	return &cmdResult{
		Contracts: contracts,
	}, nil
}

func (ip *deployEVMParams) setChainAdditionalData(
	ctx context.Context, gatewayProxyAddr ethcommon.Address,
	txHelper *eth.EthHelperWrapper, outputter common.OutputFormatter,
) error {
	if ip.bridgePrivateKey == "" {
		return nil
	}

	sc := eth.NewBridgeSmartContract(ip.bridgeSCAddr, txHelper)

	_, _ = outputter.Write([]byte(fmt.Sprintf("Configuring bridge smart contract at %s...", ip.bridgeSCAddr)))
	outputter.WriteOutput()

	_, err := infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (bool, error) {
		return true, sc.SetChainAdditionalData(ctx, ip.evmChainID, gatewayProxyAddr.String(), "")
	})

	return err
}

func (ip *deployEVMParams) getValidatorsChainData(
	ctx context.Context, txHelper *eth.EthHelperWrapper, outputter common.OutputFormatter,
) ([]eth.ValidatorChainData, error) {
	if ip.bridgeNodeURL != "" {
		_, _ = outputter.Write([]byte(fmt.Sprintf("Get data from bridge smart contract at %s...", ip.bridgeSCAddr)))
		outputter.WriteOutput()

		return eth.NewBridgeSmartContract(ip.bridgeSCAddr, txHelper).GetValidatorsChainData(ctx, ip.evmChainID)
	}

	result := make([]eth.ValidatorChainData, len(ip.evmBlsKeys))
	existing := make(map[string]bool, len(ip.evmBlsKeys))

	for i, x := range ip.evmBlsKeys {
		if x == "" {
			return nil, errors.New("empty key")
		}

		blsRaw, err := common.DecodeHex(x)
		if err != nil {
			return nil, err
		}

		key, err := bn256.UnmarshalPublicKey(blsRaw)
		if err != nil {
			return nil, err
		}

		if existing[x] {
			return nil, fmt.Errorf("duplicate key: %s", x)
		}

		existing[x] = true
		result[i] = eth.ValidatorChainData{
			Key: key.ToBigInt(),
		}
	}

	return result, nil
}

func (ip *deployEVMParams) getTxHelperBridge() (*eth.EthHelperWrapper, error) {
	if ip.bridgeNodeURL == "" {
		return nil, nil
	}

	if ip.bridgePrivateKey == "" {
		return eth.NewEthHelperWrapper(
			hclog.NewNullLogger(),
			ethtxhelper.WithNodeURL(ip.bridgeNodeURL),
			ethtxhelper.WithDynamicTx(false)), nil
	}

	wallet, err := ethtxhelper.NewEthTxWallet(ip.bridgePrivateKey)
	if err != nil {
		return nil, err
	}

	return eth.NewEthHelperWrapperWithWallet(
		wallet, hclog.NewNullLogger(),
		ethtxhelper.WithNodeURL(ip.bridgeNodeURL),
		ethtxhelper.WithInitClientAndChainIDFn(context.Background()),
		ethtxhelper.WithNonceStrategyType(ethtxhelper.NonceInMemoryStrategy),
		ethtxhelper.WithDynamicTx(false)), nil
}
