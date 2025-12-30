package clideployevm

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"

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
	defaultGasLimitValue      = uint64(5_242_880)
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
	gasLimitFlag        = "gas-limit"

	bridgeNodeURLFlag    = "bridge-url"
	bridgeSCAddrFlag     = "bridge-addr"
	bridgePrivateKeyFlag = "bridge-key"

	minFeeAmountFlag           = "min-fee"
	minBridgingAmountFlag      = "min-bridging-amount"
	minTokenBridgingAmountFlag = "min-token-bridging-amount" //nolint:gosec
	minOperationFeeFlag        = "min-operation-fee"
	currencyTokIDFlag          = "currency-token-id"
	chainIDsConfigFlag         = "chain-ids-config"

	evmNodeURLFlagDesc      = "evm node url"
	evmSCDirFlagDesc        = "the directory where the repository will be cloned, or the directory where the compiled evm smart contracts (JSON files) are located." //nolint:lll
	evmPrivateKeyFlagDesc   = "private key for smart contract admin"
	evmBlsKeyFlagDesc       = "bls key of the bridge validator. it can be used multiple times, but the order must be the same as on the bridge" //nolint:lll
	evmChainIDFlagDesc      = "evm chain ID (prime, vector, etc)"
	evmDynamicTxFlagDesc    = "dynamic tx"
	evmCloneEvmRepoFlagDesc = "clone evm gateway repository and build smart contracts"
	evmBranchNameFlagDesc   = "branch to use if the evm gateway repository is cloned"
	gasLimitFlagDesc        = "gas limit for transaction"

	bridgeNodeURLFlagDesc    = "bridge node url"
	bridgeSCAddrFlagDesc     = "bridge smart contract address"
	bridgePrivateKeyFlagDesc = "bridge admin private key"

	privateKeyConfigFlag     = "key-config"
	privateKeyConfigFlagDesc = "path to secrets manager config file"

	minFeeAmountFlagDesc           = "minimal fee amount"
	minBridgingAmountFlagDesc      = "minimal amount to bridge"
	minTokenBridgingAmountFlagDesc = "minimal amount to bridge tokens"
	minOperationFeeFlagDesc        = "minimal operation fee"
	currencyTokIDFlagDesc          = "token ID of the currency of the chain"
	chainIDsConfigFlagDesc         = "path to the chain IDs config file"

	defaultEVMChainID = common.ChainIDStrNexus

	evmGatewayRepositoryName  = "apex-evm-gateway"
	evmGatewayRepositoryURL   = "https://github.com/Ethernal-Tech/" + evmGatewayRepositoryName
	evmRepositoryArtifactsDir = "artifacts"
)

const (
	Gateway              = "Gateway"
	NativeTokenPredicate = "NativeTokenPredicate"
	NativeTokenWallet    = "NativeTokenWallet"
	Validators           = "Validators"
	MyToken              = "MyToken"
	TokenFactory         = "TokenFactory"

	MyTokenTestName   = "Test Token"
	MyTokenTestSymbol = "TTK"

	defaultEvmBranch       = "feat/skyline"
	defaultCurrencyTokenID = 1
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

	privateKeyConfig string

	minFeeString                 string
	minBridgingAmountString      string
	minTokenBridgingAmountString string
	minOperationFeeString        string

	minFeeAmount           *big.Int
	minBridgingAmount      *big.Int
	minTokenBridgingAmount *big.Int
	minOperationFee        *big.Int

	currencyTokenID uint16

	gasLimit       uint64
	chainIDsConfig string

	chainIDConverter *common.ChainIDConverter
}

func (ip *deployEVMParams) validateFlags() error {
	if !common.IsValidHTTPURL(ip.evmNodeURL) {
		return fmt.Errorf("invalid --%s flag", evmNodeURLFlag)
	}

	if ip.chainIDsConfig == "" {
		return fmt.Errorf("--%s flag not specified", chainIDsConfigFlag)
	}

	if _, err := os.Stat(ip.chainIDsConfig); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config file does not exist: %s", ip.chainIDsConfig)
		}

		return fmt.Errorf("failed to check config file: %s. err: %w", ip.chainIDsConfig, err)
	}

	chainIDsConfig, err := common.LoadConfig[common.ChainIDsConfig](ip.chainIDsConfig, "")
	if err != nil {
		return fmt.Errorf("failed to load chain IDs config: %w", err)
	}

	ip.chainIDConverter = chainIDsConfig.ToChainIDConverter()

	if !ip.chainIDConverter.IsExistingChainID(ip.evmChainID) {
		return fmt.Errorf("unexisting chain: %s", ip.evmChainID)
	}

	if ip.evmPrivateKey == "" && ip.privateKeyConfig == "" {
		return fmt.Errorf("specify at least one: --%s or --%s", evmPrivateKeyFlag, privateKeyConfigFlag)
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

	feeAmount, ok := new(big.Int).SetString(ip.minFeeString, 0)
	if !ok {
		return fmt.Errorf("--%s invalid amount", minFeeAmountFlag)
	}

	if feeAmount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("--%s invalid amount: %d", minFeeAmountFlag, feeAmount)
	}

	bridgingAmount, ok := new(big.Int).SetString(ip.minBridgingAmountString, 0)
	if !ok {
		return fmt.Errorf("--%s invalid amount", minBridgingAmountFlag)
	}

	if bridgingAmount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("--%s invalid amount: %d", minBridgingAmountFlag, bridgingAmount)
	}

	tokenBridgingAmount, ok := new(big.Int).SetString(ip.minTokenBridgingAmountString, 0)
	if !ok {
		return fmt.Errorf("--%s invalid amount", minTokenBridgingAmountFlag)
	}

	if tokenBridgingAmount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("--%s invalid amount: %d", minBridgingAmountFlag, tokenBridgingAmount)
	}

	operationFeeAmount, ok := new(big.Int).SetString(ip.minOperationFeeString, 0)
	if !ok {
		return fmt.Errorf("--%s invalid amount", minOperationFeeFlag)
	}

	if operationFeeAmount.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("--%s invalid amount: %d", minOperationFeeFlag, operationFeeAmount)
	}

	if ip.currencyTokenID == 0 {
		return fmt.Errorf("--%s invalid value: %d", currencyTokIDFlag, ip.currencyTokenID)
	}

	ip.minFeeAmount = feeAmount
	ip.minBridgingAmount = bridgingAmount
	ip.minTokenBridgingAmount = tokenBridgingAmount
	ip.minOperationFee = operationFeeAmount

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
		defaultEvmBranch,
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
		&ip.minFeeString,
		minFeeAmountFlag,
		common.DfmToWei(new(big.Int).SetUint64(common.MinFeeForBridgingDefault)).String(),
		minFeeAmountFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.minBridgingAmountString,
		minBridgingAmountFlag,
		common.DfmToWei(new(big.Int).SetUint64(common.MinUtxoAmountDefault)).String(),
		minBridgingAmountFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.minTokenBridgingAmountString,
		minTokenBridgingAmountFlag,
		common.DfmToWei(new(big.Int).SetUint64(common.MinColCoinsAllowedToBridgeDefault)).String(),
		minTokenBridgingAmountFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.minOperationFeeString,
		minOperationFeeFlag,
		common.DfmToWei(new(big.Int).SetUint64(common.MinOperationFeeDefault)).String(),
		minOperationFeeFlagDesc,
	)

	cmd.Flags().Uint16Var(
		&ip.currencyTokenID,
		currencyTokIDFlag,
		defaultCurrencyTokenID,
		currencyTokIDFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.evmPrivateKey,
		evmPrivateKeyFlag,
		"",
		evmPrivateKeyFlagDesc,
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

	cmd.Flags().Uint64Var(
		&ip.gasLimit,
		gasLimitFlag,
		defaultGasLimitValue,
		gasLimitFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.chainIDsConfig,
		chainIDsConfigFlag,
		"",
		chainIDsConfigFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(bridgePrivateKeyFlag, privateKeyConfigFlag)
	cmd.MarkFlagsMutuallyExclusive(evmPrivateKeyFlag, privateKeyConfigFlag)

	cmd.MarkFlagsMutuallyExclusive(bridgeNodeURLFlag, evmBlsKeyFlag)
	cmd.MarkFlagsMutuallyExclusive(bridgeSCAddrFlag, evmBlsKeyFlag)
	cmd.MarkFlagsMutuallyExclusive(bridgePrivateKeyFlag, evmBlsKeyFlag)
}

func (ip *deployEVMParams) Execute(
	outputter common.OutputFormatter,
) (common.ICommandResult, error) {
	dir := filepath.Clean(ip.evmDir)
	ctx := context.Background()

	contractNames := []string{
		Gateway, NativeTokenPredicate, NativeTokenWallet, Validators, MyToken, TokenFactory,
	}
	setDependenciesData := map[string][]string{
		Gateway:              {NativeTokenPredicate, TokenFactory, Validators},
		NativeTokenPredicate: {Gateway, NativeTokenWallet},
		NativeTokenWallet:    {NativeTokenPredicate},
		Validators:           {Gateway},
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

	wallet, err := eth.GetEthWalletForBladeAdmin(true, ip.evmPrivateKey, ip.privateKeyConfig)
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

	txHelperBridge, err := ip.getTxHelperBridge()
	if err != nil {
		return nil, err
	}

	validatorsData, err := ip.getValidatorsChainData(ctx, txHelperBridge, ip.chainIDConverter, outputter)
	if err != nil {
		return nil, err
	}

	_, _ = outputter.Write([]byte("Deploying the smart contracts has started..."))
	outputter.WriteOutput()

	ethContractUtils := ethcontracts.NewEthContractUtils(txHelper, wallet, defaultGasLimitMultiplier)
	contracts := make([]contractInfo, len(contractNames)*2)
	txHashes := make([]string, len(contractNames)*2)
	addresses := make(map[string]ethcommon.Address, len(contractNames))
	implAddresses := make(map[string]ethcommon.Address, len(contractNames))

	for i, contractName := range contractNames {
		initParams, err := ip.getInitParams(contractName, addresses, implAddresses)
		if err != nil {
			return nil, fmt.Errorf("failed to get init parameters for contract %s: %w", contractName, err)
		}

		proxyTx, tx, err := ethContractUtils.DeployWithProxy(
			ctx, artifacts[contractName], artifacts[ercProxyContractName], initParams...)
		if err != nil {
			return nil, fmt.Errorf("deploy %s has been failed: %w", contractName, err)
		}

		_, _ = outputter.Write(fmt.Appendf(nil, "%s has been sent", contractName))
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
		implAddresses[contractName] = tx.Address
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
		dependencies := make([]any, len(dependencyNames))
		for i, x := range dependencyNames {
			dependencies[i] = addresses[x]
		}

		txInfo, err := ethContractUtils.ExecuteMethod(
			ctx, artifacts[contractName], addresses[contractName], "setDependencies", dependencies...)
		if err != nil {
			return nil, fmt.Errorf("setDependecies for %s has been failed: %w", contractName, err)
		}

		_, _ = outputter.Write(fmt.Appendf(nil, "%s initialization transaction has been sent", contractName))
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

	if err := ip.setChainAdditionalData(
		ctx, addresses[Gateway], txHelperBridge,
		ip.chainIDConverter, outputter,
	); err != nil {
		return nil, err
	}

	return &cmdResult{
		Contracts: contracts,
	}, nil
}

func (ip *deployEVMParams) setChainAdditionalData(
	ctx context.Context, gatewayProxyAddr ethcommon.Address,
	txHelper *eth.EthHelperWrapper, chainIDConverter *common.ChainIDConverter, outputter common.OutputFormatter,
) error {
	if ip.bridgePrivateKey == "" && ip.privateKeyConfig == "" {
		return nil
	}

	sc := eth.NewBridgeSmartContract(ip.bridgeSCAddr, txHelper, chainIDConverter)

	_, _ = outputter.Write(fmt.Appendf(nil, "Configuring bridge smart contract at %s...", ip.bridgeSCAddr))
	outputter.WriteOutput()

	_, err := infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (bool, error) {
		return true, sc.SetChainAdditionalData(ctx, ip.evmChainID, gatewayProxyAddr.String(), "")
	})

	return err
}

func (ip *deployEVMParams) getValidatorsChainData(
	ctx context.Context, txHelper *eth.EthHelperWrapper,
	chainIDConverter *common.ChainIDConverter, outputter common.OutputFormatter,
) ([]eth.ValidatorChainData, error) {
	if ip.bridgeNodeURL != "" {
		_, _ = outputter.Write(fmt.Appendf(nil, "Get data from bridge smart contract at %s...", ip.bridgeSCAddr))
		outputter.WriteOutput()

		bridgeSC := eth.NewBridgeSmartContract(
			ip.bridgeSCAddr,
			txHelper,
			chainIDConverter,
		)

		return bridgeSC.GetValidatorsChainData(ctx, ip.evmChainID)
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

	if ip.bridgePrivateKey == "" && ip.privateKeyConfig == "" {
		return eth.NewEthHelperWrapper(
			hclog.NewNullLogger(),
			ethtxhelper.WithNodeURL(ip.bridgeNodeURL),
			ethtxhelper.WithDynamicTx(false)), nil
	}

	wallet, err := eth.GetEthWalletForBladeAdmin(false, ip.bridgePrivateKey, ip.privateKeyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create smart contracts admin wallet: %w", err)
	}

	return eth.NewEthHelperWrapperWithWallet(
		wallet, hclog.NewNullLogger(),
		ethtxhelper.WithNodeURL(ip.bridgeNodeURL),
		ethtxhelper.WithInitClientAndChainIDFn(context.Background()),
		ethtxhelper.WithNonceStrategyType(ethtxhelper.NonceInMemoryStrategy),
		ethtxhelper.WithDynamicTx(false)), nil
}

func (ip *deployEVMParams) getInitParams(
	contractName string,
	addresses map[string]ethcommon.Address,
	implAddresses map[string]ethcommon.Address,
) ([]any, error) {
	switch strings.ToLower(contractName) {
	case strings.ToLower(Gateway):
		return []any{
			ip.minFeeAmount,
			ip.minBridgingAmount,
			ip.minTokenBridgingAmount,
			ip.minOperationFee,
			ip.currencyTokenID,
		}, nil
	case strings.ToLower(TokenFactory):
		gatewayProxy, ok := addresses[Gateway]
		if !ok {
			return nil, fmt.Errorf("missing Gateway address for TokenFactory")
		}

		myTokenImpl, ok := implAddresses[MyToken]
		if !ok {
			return nil, fmt.Errorf("missing MyToken implementation address for TokenFactory")
		}

		nativeWallet, ok := addresses[NativeTokenWallet]
		if !ok {
			return nil, fmt.Errorf("missing NativeTokenWallet address for TokenFactory")
		}

		return []any{
			gatewayProxy,
			myTokenImpl,
			nativeWallet,
		}, nil
	case strings.ToLower(MyToken):
		nativeTokenWallet, ok := addresses[NativeTokenWallet]
		if !ok {
			return nil, fmt.Errorf("missing NativeTokenWallet address for MyToken")
		}

		return []any{
			MyTokenTestName,
			MyTokenTestSymbol,
			nativeTokenWallet,
		}, nil
	case
		strings.ToLower(NativeTokenPredicate),
		strings.ToLower(NativeTokenWallet),
		strings.ToLower(Validators):
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown contract: %s", contractName)
	}
}
