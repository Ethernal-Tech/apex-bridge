package cligenerateconfigs

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	rCore "github.com/Ethernal-Tech/apex-bridge/relayer/core"
	vcCore "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/spf13/cobra"
)

const (
	evmChainNodeURLFlag                = "evm-node-url"
	evmChainTTLBlockNumberIncFlag      = "evm-ttl-block-inc"
	evmChainBlockRoundingThresholdFlag = "evm-block-rounding-threshold"
	evmChainStartingBlockFlag          = "evm-starting-block"
	evmChainMinFeeForBridgingFlag      = "evm-min-fee-for-bridging"
	evmRelayerGasFeeMultiplierFlag     = "evm-relayer-gas-fee-multiplier"

	evmChainNodeURLFlagDesc                = "evm chain node URL"
	evmChainTTLBlockNumberIncFlagDesc      = "TTL block increment for evm chain"
	evmChainBlockRoundingThresholdFlagDesc = "defines the upper limit used for rounding block values for evm chain. Any block value between 0 and `blockRoundingThreshold` will be rounded to `blockRoundingThreshold` etc" //nolint:lll
	evmChainStartingBlockFlagDesc          = "block from where to start evm chain oracle / evm chain block submitter"
	evmChainMinFeeForBridgingFlagDesc      = "minimal bridging fee for evm chain"
	evmRelayerGasFeeMultiplierFlagDesc     = "gas fee multiplier for evm relayer"

	defaultEvmBlockConfirmationCount    = 1
	defaultEvmSyncBatchSize             = 20
	defaultEvmPoolIntervalMiliseconds   = 1500
	defaultEvmNoBatchPeriodPercent      = 0.2
	defaultEvmTTLBlockRoundingThreshold = 10
	defaultEvmTTLBlockNumberInc         = 20
	defaultEvmRelayerGasFeeMultiplier   = 140

	defaultEvmFeeAddrBridgingAmount = 1_000_000
)

type evmChainGenerateConfigsParams struct {
	chainIDString string

	evmChainNodeURL                string
	evmChainTTLBlockNumberInc      uint64
	evmChainBlockRoundingThreshold uint64
	evmChainStartingBlock          uint64
	evmChainMinFeeForBridging      uint64

	evmRelayerGasFeeMultiplier uint64
	emptyBlocksThreshold       uint

	allowedDirections []string
	coloredCoins      []string

	outputDir                         string
	outputValidatorComponentsFileName string
	outputRelayerFileName             string

	dbsPath           string
	relayerDataDir    string
	relayerConfigPath string
}

func (p *evmChainGenerateConfigsParams) validateFlags() error {
	if p.chainIDString == "" {
		return fmt.Errorf("missing %s", chainIDStringFlag)
	}

	if !common.IsValidHTTPURL(p.evmChainNodeURL) {
		return fmt.Errorf("invalid %s: %s", evmChainNodeURLFlag, p.evmChainNodeURL)
	}

	if p.relayerDataDir == "" && p.relayerConfigPath == "" {
		return fmt.Errorf("specify at least one of: %s, %s", relayerDataDirFlag, relayerConfigPathFlag)
	}

	// Validate allowed directions format
	for _, dirStr := range p.allowedDirections {
		if err := validateAllowedDirectionFormat(dirStr); err != nil {
			return fmt.Errorf("invalid %s format: %w", allowedDirectionsFlag, err)
		}
	}

	// Validate colored coins format
	for _, coinStr := range p.coloredCoins {
		if err := validateEthColoredCoinFormat(coinStr, p.chainIDString); err != nil {
			return fmt.Errorf("invalid %s format: %w", coloredCoinsFlag, err)
		}
	}

	return nil
}

func validateEthColoredCoinFormat(coinStr string, chainID string) error {
	parts := strings.Split(coinStr, ":")
	if len(parts) != 3 {
		return fmt.Errorf("invalid %s format: %s", coloredCoinsFlag, coinStr)
	}

	contractAddress := strings.TrimSpace(parts[0])
	if contractAddress == "" {
		return fmt.Errorf("invalid %s format: %s", coloredCoinsFlag, coinStr)
	}

	if !common.IsValidAddress(chainID, contractAddress) {
		return fmt.Errorf("invalid %s format: %s", coloredCoinsFlag, coinStr)
	}

	tokenName := strings.TrimSpace(parts[1])
	if tokenName == "" {
		return fmt.Errorf("invalid %s format: %s", coloredCoinsFlag, coinStr)
	}

	coloredCoinID, err := strconv.ParseUint(parts[2], 10, 8)
	if err != nil {
		return fmt.Errorf("invalid %s format: %s", coloredCoinsFlag, coinStr)
	}

	if coloredCoinID == 0 {
		return fmt.Errorf("invalid %s format: %s", coloredCoinsFlag, coinStr)
	}

	return nil
}

func (p *evmChainGenerateConfigsParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&p.chainIDString,
		chainIDStringFlag,
		"",
		chainIDStringFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.evmChainNodeURL,
		evmChainNodeURLFlag,
		"",
		evmChainNodeURLFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.evmChainTTLBlockNumberInc,
		evmChainTTLBlockNumberIncFlag,
		defaultEvmTTLBlockNumberInc,
		evmChainTTLBlockNumberIncFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.evmChainBlockRoundingThreshold,
		evmChainBlockRoundingThresholdFlag,
		defaultEvmTTLBlockRoundingThreshold,
		evmChainBlockRoundingThresholdFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.evmChainStartingBlock,
		evmChainStartingBlockFlag,
		0,
		evmChainStartingBlockFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.evmChainMinFeeForBridging,
		evmChainMinFeeForBridgingFlag,
		common.MinFeeForBridgingDefault,
		evmChainMinFeeForBridgingFlagDesc,
	)

	cmd.Flags().Uint64Var(
		&p.evmRelayerGasFeeMultiplier,
		evmRelayerGasFeeMultiplierFlag,
		defaultEvmRelayerGasFeeMultiplier,
		evmRelayerGasFeeMultiplierFlagDesc,
	)

	cmd.Flags().UintVar(
		&p.emptyBlocksThreshold,
		emptyBlocksThresholdFlag,
		defaultEmptyBlocksThreshold,
		emptyBlocksThresholdFlagDesc,
	)

	cmd.Flags().StringArrayVar(
		&p.allowedDirections,
		allowedDirectionsFlag,
		nil,
		allowedDirectionsFlagDesc,
	)

	cmd.Flags().StringArrayVar(
		&p.coloredCoins,
		coloredCoinsFlag,
		nil,
		coloredCoinsFlagDesc,
	)

	// Output params
	cmd.Flags().StringVar(
		&p.outputDir,
		outputDirFlag,
		defaultOutputDir,
		outputDirFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.outputValidatorComponentsFileName,
		outputValidatorComponentsFileNameFlag,
		defaultOutputValidatorComponentsFileName,
		outputValidatorComponentsFileNameFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.outputRelayerFileName,
		outputRelayerFileNameFlag,
		defaultOutputRelayerFileName,
		outputRelayerFileNameFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.relayerDataDir,
		relayerDataDirFlag,
		"",
		relayerDataDirFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.relayerConfigPath,
		relayerConfigPathFlag,
		"",
		relayerConfigPathFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.dbsPath,
		dbsPathFlag,
		defaultDBsPath,
		dbsPathFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(relayerDataDirFlag, relayerConfigPathFlag)
}

func (p *evmChainGenerateConfigsParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	outputDirPath := filepath.Clean(p.outputDir)
	if err := common.CreateDirectoryIfNotExists(outputDirPath, 0770); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	vcConfigPath := filepath.Join(outputDirPath, p.outputValidatorComponentsFileName)

	vcConfig, err := common.LoadJSON[vcCore.AppConfig](vcConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load validator components config json: %w", err)
	}

	if vcConfig.EthChains == nil {
		vcConfig.EthChains = make(map[string]*oCore.EthChainConfig)
	}

	vcConfig.EthChains[p.chainIDString] = &oCore.EthChainConfig{
		NodeURL:                 p.evmChainNodeURL,
		SyncBatchSize:           defaultEvmSyncBatchSize,
		NumBlockConfirmations:   defaultEvmBlockConfirmationCount,
		StartBlockNumber:        p.evmChainStartingBlock,
		PoolIntervalMiliseconds: defaultEvmPoolIntervalMiliseconds,
		TTLBlockNumberInc:       p.evmChainTTLBlockNumberInc,
		BlockRoundingThreshold:  p.evmChainBlockRoundingThreshold,
		NoBatchPeriodPercent:    defaultEvmNoBatchPeriodPercent,
		DynamicTx:               true,
		MinFeeForBridging:       p.evmChainMinFeeForBridging,
		RestartTrackerPullCheck: time.Second * 150,
		FeeAddrBridgingAmount:   defaultEvmFeeAddrBridgingAmount,
	}

	if vcConfig.Bridge.SubmitConfig.EmptyBlocksThreshold == nil {
		vcConfig.Bridge.SubmitConfig.EmptyBlocksThreshold = make(map[string]uint)
	}

	vcConfig.Bridge.SubmitConfig.EmptyBlocksThreshold[p.chainIDString] = p.emptyBlocksThreshold

	if vcConfig.BridgingSettings.AllowedDirections == nil {
		vcConfig.BridgingSettings.AllowedDirections = make(oCore.AllowedDirections)
	}

	// Parse allowed directions
	allowedDirs, err := parseAllowedDirections(p.allowedDirections)
	if err != nil {
		return nil, fmt.Errorf("failed to parse allowed directions: %w", err)
	}

	if vcConfig.BridgingSettings.AllowedDirections[p.chainIDString] == nil {
		vcConfig.BridgingSettings.AllowedDirections[p.chainIDString] = make(map[string]oCore.AllowedDirection)
	}

	for destChainID, direction := range allowedDirs {
		vcConfig.BridgingSettings.AllowedDirections[p.chainIDString][destChainID] = direction
	}

	// Parse colored coins
	coloredCoins, err := parseEthColoredCoins(p.coloredCoins)
	if err != nil {
		return nil, fmt.Errorf("failed to parse colored coins: %w", err)
	}

	if vcConfig.EthChains[p.chainIDString].ColoredCoins == nil {
		vcConfig.EthChains[p.chainIDString].ColoredCoins = make([]oCore.ColoredCoinEvm, 0)
	}

	vcConfig.EthChains[p.chainIDString].ColoredCoins = append(vcConfig.EthChains[p.chainIDString].ColoredCoins, coloredCoins...)

	if err := common.SaveJSON(vcConfigPath, vcConfig, true); err != nil {
		return nil, fmt.Errorf("failed to update validator components config json: %w", err)
	}

	rConfigPath := filepath.Join(outputDirPath, p.outputRelayerFileName)

	rConfig, err := common.LoadJSON[rCore.RelayerManagerConfiguration](rConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load relayer config json: %w", err)
	}

	chainSpecificJSONRaw, err := json.Marshal(cardanotx.RelayerEVMChainConfig{
		NodeURL:          p.evmChainNodeURL,
		DataDir:          cleanPath(p.relayerDataDir),
		ConfigPath:       cleanPath(p.relayerConfigPath),
		DynamicTx:        true,
		GasFeeMultiplier: p.evmRelayerGasFeeMultiplier,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chain specific config to json: %w", err)
	}

	if rConfig.Chains == nil {
		rConfig.Chains = make(map[string]rCore.ChainConfig)
	}

	rConfig.Chains[p.chainIDString] = rCore.ChainConfig{
		ChainType:     common.ChainTypeEVMStr,
		DbsPath:       filepath.Join(p.dbsPath, "relayer"),
		ChainSpecific: chainSpecificJSONRaw,
	}

	if err := common.SaveJSON(rConfigPath, rConfig, true); err != nil {
		return nil, fmt.Errorf("failed to update relayer config json: %w", err)
	}

	return &CmdResult{
		validatorComponentsConfigPath: vcConfigPath,
		relayerConfigPath:             rConfigPath,
	}, nil
}

func parseEthColoredCoins(s []string) ([]oCore.ColoredCoinEvm, error) {
	result := make([]oCore.ColoredCoinEvm, 0)
	for _, coinStr := range s {
		coin, err := parseEthColoredCoin(coinStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse colored coin: %w", err)
		}
		result = append(result, coin)
	}
	return result, nil
}

func parseEthColoredCoin(coinStr string) (oCore.ColoredCoinEvm, error) {
	parts := strings.Split(coinStr, ":")

	tokenName := strings.TrimSpace(parts[0])
	if tokenName == "" {
		return oCore.ColoredCoinEvm{}, fmt.Errorf("invalid %s format: %s", coloredCoinsFlag, coinStr)
	}

	coloredCoinID, err := strconv.ParseUint(parts[1], 10, 8)
	if err != nil {
		return oCore.ColoredCoinEvm{}, fmt.Errorf("invalid %s format: %s", coloredCoinsFlag, coinStr)
	}

	return oCore.ColoredCoinEvm{
		TokenName:       tokenName,
		ColoredCoinID:   uint8(coloredCoinID),
		ContractAddress: strings.TrimSpace(parts[2]),
	}, nil
}
