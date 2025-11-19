package cligenerateconfigs

import (
	"fmt"
	"math"
	"math/big"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	apiCore "github.com/Ethernal-Tech/apex-bridge/api/core"
	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	rCore "github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	vcCore "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
)

var defaultMaxTokenAmountAllowedToBridge = new(big.Int).SetUint64(1_000_000_000_000)

type skylineGenerateConfigsParams struct {
	bridgeNodeURL   string
	bridgeSCAddress string

	minColCoinsAmount uint64

	validatorDataDir string
	validatorConfig  string

	logsPath string
	dbsPath  string

	apiPort uint32
	apiKeys []string

	outputDir                         string
	outputValidatorComponentsFileName string
	outputRelayerFileName             string

	telemetry string

	emptyBlocksThreshold uint

	coloredCoins []string
}

func (p *skylineGenerateConfigsParams) validateFlags() error {
	if !common.IsValidHTTPURL(p.bridgeNodeURL) {
		return fmt.Errorf("invalid %s: %s", bridgeNodeURLFlag, p.bridgeNodeURL)
	}

	if p.bridgeSCAddress == "" {
		return fmt.Errorf("missing %s", bridgeSCAddressFlag)
	}

	if p.validatorDataDir == "" && p.validatorConfig == "" {
		return fmt.Errorf("specify at least one of: %s, %s", validatorDataDirFlag, validatorConfigFlag)
	}

	if len(p.apiKeys) == 0 {
		return fmt.Errorf("specify at least one %s", apiKeysFlag)
	}

	if p.telemetry != "" {
		parts := strings.Split(p.telemetry, ",")
		if len(parts) < 1 || len(parts) > 2 || !common.IsValidNetworkAddress(strings.TrimSpace(parts[0])) ||
			(len(parts) == 2 && !common.IsValidNetworkAddress(strings.TrimSpace(parts[1]))) {
			return fmt.Errorf("invalid telemetry: %s", p.telemetry)
		}
	}

	for _, coinStr := range p.coloredCoins {
		if err := validateColoredCoinConfigFormat(coinStr); err != nil {
			return fmt.Errorf("invalid %s format: %w", coloredCoinsFlag, err)
		}
	}

	return nil
}

func validateColoredCoinConfigFormat(coinStr string) error {
	parts := strings.Split(coinStr, ":")
	if len(parts) != 3 {
		return fmt.Errorf("invalid %s format: expected 'id:name:ecosystemOriginChainID', got: %s", coloredCoinsFlag, coinStr)
	}

	id, err := strconv.ParseUint(parts[0], 10, 8)
	if err != nil {
		return fmt.Errorf("invalid %s format: %w", coloredCoinsFlag, err)
	}

	if id > math.MaxUint8 {
		return fmt.Errorf("invalid %s format: id must be less than %d", coloredCoinsFlag, math.MaxUint8)
	}

	if id == 0 {
		return fmt.Errorf("invalid %s format: id must be greater than 0", coloredCoinsFlag)
	}

	name := strings.TrimSpace(parts[1])
	if name == "" {
		return fmt.Errorf("invalid %s format: %s", coloredCoinsFlag, coinStr)
	}

	ecosystemOriginChainID := strings.TrimSpace(parts[2])
	if ecosystemOriginChainID == "" {
		return fmt.Errorf("invalid %s format: %s", coloredCoinsFlag, coinStr)
	}

	return nil
}

func (p *skylineGenerateConfigsParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&p.bridgeNodeURL,
		bridgeNodeURLFlag,
		"",
		bridgeNodeURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.bridgeSCAddress,
		bridgeSCAddressFlag,
		"",
		bridgeSCAddressFlagDesc,
	)

	cmd.Flags().Uint64Var(
		&p.minColCoinsAmount,
		minColCoinsAmountFlag,
		0, //TODO: set default val
		minColCoinsAmountFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.validatorDataDir,
		validatorDataDirFlag,
		"",
		validatorDataDirFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.validatorConfig,
		validatorConfigFlag,
		"",
		validatorConfigFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.logsPath,
		logsPathFlag,
		defaultLogsPath,
		logsPathFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.dbsPath,
		dbsPathFlag,
		defaultDBsPath,
		dbsPathFlagDesc,
	)

	cmd.Flags().Uint32Var(
		&p.apiPort,
		apiPortFlag,
		defaultAPIPort,
		apiPortFlagDesc,
	)

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
	cmd.Flags().StringArrayVar(
		&p.apiKeys,
		apiKeysFlag,
		nil,
		apiKeysFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.telemetry,
		telemetryFlag,
		"",
		telemetryFlagDesc,
	)

	cmd.Flags().UintVar(
		&p.emptyBlocksThreshold,
		emptyBlocksThresholdFlag,
		defaultEmptyBlocksThreshold,
		emptyBlocksThresholdFlagDesc,
	)

	cmd.Flags().StringArrayVar(
		&p.coloredCoins,
		coloredCoinsFlag,
		nil,
		coloredCoinsConfigFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(validatorDataDirFlag, validatorConfigFlag)
}

func (p *skylineGenerateConfigsParams) Execute(
	outputter common.OutputFormatter,
) (common.ICommandResult, error) {
	telemetryConfig := telemetry.TelemetryConfig{
		PullTime: time.Second * 10,
	}

	if p.telemetry != "" {
		parts := strings.Split(p.telemetry, ",")

		telemetryConfig.PrometheusAddr = strings.TrimSpace(parts[0])
		if len(parts) == 2 {
			telemetryConfig.DataDogAddr = strings.TrimSpace(parts[1])
		}
	}

	coloredCoins := make([]vcCore.ColoredCoinSettings, 0)
	for _, coinStr := range p.coloredCoins {
		coin, err := parseColoredCoinConfig(coinStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse colored coin: %w", err)
		}

		coloredCoins = append(coloredCoins, coin)
	}

	vcConfig := &vcCore.AppConfig{
		RunMode:             common.SkylineMode,
		RefundEnabled:       true,
		ValidatorDataDir:    cleanPath(p.validatorDataDir),
		ValidatorConfigPath: cleanPath(p.validatorConfig),
		CardanoChains:       map[string]*oCore.CardanoChainConfig{},
		Bridge: oCore.BridgeConfig{
			NodeURL:              p.bridgeNodeURL,
			DynamicTx:            false,
			SmartContractAddress: p.bridgeSCAddress,
			SubmitConfig: oCore.SubmitConfig{
				ConfirmedBlocksThreshold:  20,
				ConfirmedBlocksSubmitTime: 3000,
				EmptyBlocksThreshold:      map[string]uint{},
			},
		},
		BridgingSettings: oCore.BridgingSettings{
			MaxAmountAllowedToBridge:       defaultMaxAmountAllowedToBridge,
			MaxTokenAmountAllowedToBridge:  defaultMaxTokenAmountAllowedToBridge,
			MaxReceiversPerBridgingRequest: 4, // 4 + 1 for fee
			MaxBridgingClaimsToGroup:       5,
			MinColCoinsAllowedToBridge:     p.minColCoinsAmount,
			AllowedDirections:              oCore.AllowedDirections{},
		},
		ColoredCoins: coloredCoins,
		RetryUnprocessedSettings: oCore.RetryUnprocessedSettings{
			BaseTimeout: time.Second * 60,
			MaxTimeout:  time.Second * 60 * 2048,
		},
		TryCountLimits: oCore.TryCountLimits{
			MaxBatchTryCount:  70,
			MaxSubmitTryCount: 50,
			MaxRefundTryCount: 50,
		},
		Settings: oCore.AppSettings{
			Logger: logger.LoggerConfig{
				LogFilePath:         filepath.Join(p.logsPath, "validator-components.log"),
				LogLevel:            hclog.Debug,
				JSONLogFormat:       false,
				AppendFile:          true,
				RotatingLogsEnabled: false,
				RotatingLogerConfig: logger.RotatingLoggerConfig{
					MaxSizeInMB:  100,
					MaxBackups:   30,
					MaxAgeInDays: 30,
					Compress:     false,
				},
			},
			DbsPath: filepath.Join(p.dbsPath, "validatorcomponents"),
		},
		RelayerImitatorPullTimeMilis: 1000,
		BatcherPullTimeMilis:         2500,
		APIConfig: apiCore.APIConfig{
			Port:       p.apiPort,
			PathPrefix: "api",
			AllowedHeaders: []string{
				"Content-Type",
			},
			AllowedOrigins: []string{
				"*",
			},
			AllowedMethods: []string{
				"GET",
				"HEAD",
				"POST",
				"PUT",
				"OPTIONS",
				"DELETE",
			},
			APIKeyHeader: "x-api-key",
			APIKeys:      p.apiKeys,
		},
		Telemetry: telemetryConfig,
	}

	rConfig := &rCore.RelayerManagerConfiguration{
		RunMode: common.SkylineMode,
		Bridge: rCore.BridgeConfig{
			NodeURL:              p.bridgeNodeURL,
			DynamicTx:            false,
			SmartContractAddress: p.bridgeSCAddress,
		},
		Chains:        map[string]rCore.ChainConfig{},
		PullTimeMilis: 1000,
		Logger: logger.LoggerConfig{
			LogFilePath:         filepath.Join(p.logsPath, "relayer.log"),
			LogLevel:            hclog.Debug,
			JSONLogFormat:       false,
			AppendFile:          true,
			RotatingLogsEnabled: false,
			RotatingLogerConfig: logger.RotatingLoggerConfig{
				MaxSizeInMB:  100,
				MaxBackups:   30,
				MaxAgeInDays: 30,
				Compress:     false,
			},
		},
	}

	outputDirPath := filepath.Clean(p.outputDir)
	if err := common.CreateDirectoryIfNotExists(outputDirPath, 0770); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	vcConfigPath := filepath.Join(outputDirPath, p.outputValidatorComponentsFileName)
	if err := common.SaveJSON(vcConfigPath, vcConfig, true); err != nil {
		return nil, fmt.Errorf("failed to create validator components config json: %w", err)
	}

	rConfigPath := filepath.Join(outputDirPath, p.outputRelayerFileName)
	if err := common.SaveJSON(rConfigPath, rConfig, true); err != nil {
		return nil, fmt.Errorf("failed to create relayer config json: %w", err)
	}

	return &CmdResult{
		validatorComponentsConfigPath: vcConfigPath,
		relayerConfigPath:             rConfigPath,
	}, nil
}

func parseColoredCoinConfig(coinStr string) (vcCore.ColoredCoinSettings, error) {
	parts := strings.Split(coinStr, ":")

	id, err := strconv.ParseUint(parts[0], 10, 8)
	if err != nil {
		return vcCore.ColoredCoinSettings{}, fmt.Errorf("invalid %s format: %w", coloredCoinsFlag, err)
	}

	return vcCore.ColoredCoinSettings{
		ID:                     uint8(id),
		Name:                   strings.TrimSpace(parts[1]),
		EcosystemOriginChainID: strings.TrimSpace(parts[2]),
	}, nil
}
