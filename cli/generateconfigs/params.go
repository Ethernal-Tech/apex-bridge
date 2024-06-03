package cligenerateconfigs

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
	rCore "github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	vcCore "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
)

const (
	primeNetworkAddressFlag        = "prime-network-address"
	primeNetworkMagicFlag          = "prime-network-magic"
	primeKeysDirFlag               = "prime-keys-dir"
	primeOgmiosURLFlag             = "prime-ogmios-url"
	primeBlockfrostURLFlag         = "prime-blockfrost-url"
	primeBlockfrostAPIKeyFlag      = "prime-blockfrost-api-key"
	primeSocketPathFlag            = "prime-socket-path"
	primeTTLSlotIncFlag            = "prime-ttl-slot-inc"
	primeSlotRoundingThresholdFlag = "prime-slot-rounding-threshold"

	vectorNetworkAddressFlag        = "vector-network-address"
	vectorNetworkMagicFlag          = "vector-network-magic"
	vectorKeysDirFlag               = "vector-keys-dir"
	vectorOgmiosURLFlag             = "vector-ogmios-url"
	vectorBlockfrostURLFlag         = "vector-blockfrost-url"
	vectorBlockfrostAPIKeyFlag      = "vector-blockfrost-api-key"
	vectorSocketPathFlag            = "vector-socket-path"
	vectorTTLSlotIncFlag            = "vector-ttl-slot-inc"
	vectorSlotRoundingThresholdFlag = "vector-slot-rounding-threshold"

	bridgeNodeURLFlag          = "bridge-node-url"
	bridgeSCAddressFlag        = "bridge-sc-address"
	bridgeValidatorDataDirFlag = "bridge-validator-data-dir"
	bridgeValidatorConfigFlag  = "bridge-validator-config"

	logsPathFlag = "logs-path"
	dbsPathFlag  = "dbs-path"

	apiPortFlag = "api-port"
	apiKeysFlag = "api-keys"

	outputDirFlag                         = "output-dir"
	outputValidatorComponentsFileNameFlag = "output-validator-components-file-name"
	outputRelayerFileNameFlag             = "output-relayer-file-name"

	telemetryFlag = "telemetry"

	primeNetworkAddressFlagDesc        = "(mandatory) address of prime network"
	primeNetworkMagicFlagDesc          = "network magic of prime network (default 0)"
	primeKeysDirFlagDesc               = "path to cardano keys directory for prime network"
	primeOgmiosURLFlagDesc             = "ogmios URL for prime network"
	primeBlockfrostURLFlagDesc         = "blockfrost URL for prime network"
	primeBlockfrostAPIKeyFlagDesc      = "blockfrost API key for prime network" //nolint:gosec
	primeSocketPathFlagDesc            = "socket path for prime network"
	primeTTLSlotIncFlagDesc            = "TTL slot increment for prime"
	primeSlotRoundingThresholdFlagDesc = "defines the upper limit used for rounding slot values for prime. Any slot value between 0 and `slotRoundingThreshold` will be rounded to `slotRoundingThreshold` etc" //nolint:lll

	vectorNetworkAddressFlagDesc        = "(mandatory) address of vector network"
	vectorNetworkMagicFlagDesc          = "network magic of vector network (default 0)"
	vectorKeysDirFlagDesc               = "path to cardano keys directory for vector network"
	vectorOgmiosURLFlagDesc             = "ogmios URL for vector network"
	vectorBlockfrostURLFlagDesc         = "blockfrost URL for vector network"
	vectorBlockfrostAPIKeyFlagDesc      = "blockfrost API key for vector network" //nolint:gosec
	vectorSocketPathFlagDesc            = "socket path for vector network"
	vectorTTLSlotIncFlagDesc            = "TTL slot increment for vector"
	vectorSlotRoundingThresholdFlagDesc = "defines the upper limit used for rounding slot values for vector. Any slot value between 0 and `slotRoundingThreshold` will be rounded to `slotRoundingThreshold` etc" //nolint:lll

	bridgeNodeURLFlagDesc          = "(mandatory) node URL of bridge chain"
	bridgeSCAddressFlagDesc        = "(mandatory) bridging smart contract address on bridge chain"
	bridgeValidatorDataDirFlagDesc = "path to bridge chain data directory when using local secrets manager"
	bridgeValidatorConfigFlagDesc  = "path to to bridge chain secrets manager config file"

	logsPathFlagDesc = "path to where logs will be stored"
	dbsPathFlagDesc  = "path to where databases will be stored"

	apiPortFlagDesc = "port at which API should run"
	apiKeysFlagDesc = "(mandatory) list of keys for API access"

	outputDirFlagDesc                         = "path to config jsons output directory"
	outputValidatorComponentsFileNameFlagDesc = "validator components config json output file name"
	outputRelayerFileNameFlagDesc             = "relayer config json output file name"

	telemetryFlagDesc = "prometheus_ip:port,datadog_ip:port"

	defaultNetworkMagic                      = 0
	defaultPrimeKeysDir                      = "./keys/prime"
	defaultVectorKeysDir                     = "./keys/vector"
	defaultLogsPath                          = "./logs"
	defaultDBsPath                           = "./db"
	defaultAPIPort                           = 10000
	defaultOutputDir                         = "./"
	defaultOutputValidatorComponentsFileName = "config.json"
	defaultOutputRelayerFileName             = "relayer_config.json"
	defaultTTLSlotNumberInc                  = 1800 + 20*10 // ConfirmationBlockCount * BlockTimeSeconds
	defaultSlotRoundingThreshold             = 60
)

type generateConfigsParams struct {
	primeNetworkAddress        string
	primeNetworkMagic          uint32
	primeKeysDir               string
	primeOgmiosURL             string
	primeBlockfrostURL         string
	primeBlockfrostAPIKey      string
	primeSocketPath            string
	primeTTLSlotInc            uint64
	primeSlotRoundingThreshold uint64

	vectorNetworkAddress        string
	vectorNetworkMagic          uint32
	vectorKeysDir               string
	vectorOgmiosURL             string
	vectorBlockfrostURL         string
	vectorBlockfrostAPIKey      string
	vectorSocketPath            string
	vectorTTLSlotInc            uint64
	vectorSlotRoundingThreshold uint64

	bridgeNodeURL          string
	bridgeSCAddress        string
	bridgeValidatorDataDir string
	bridgeValidatorConfig  string

	logsPath string
	dbsPath  string

	apiPort uint32
	apiKeys []string

	outputDir                         string
	outputValidatorComponentsFileName string
	outputRelayerFileName             string

	telemetry string
}

func (p *generateConfigsParams) validateFlags() error {
	if p.primeNetworkAddress == "" || !common.IsValidURL(p.primeNetworkAddress) {
		return fmt.Errorf("invalid %s: %s", primeNetworkAddressFlag, p.primeNetworkAddress)
	}

	if p.primeBlockfrostURL == "" && p.primeSocketPath == "" && p.primeOgmiosURL == "" {
		return fmt.Errorf("specify at least one of: %s, %s, %s",
			primeBlockfrostURLFlag, primeSocketPathFlag, primeOgmiosURLFlag)
	}

	if p.primeBlockfrostURL != "" && !common.IsValidURL(p.primeBlockfrostURL) {
		return fmt.Errorf("invalid prime blockfrost url: %s", p.primeBlockfrostURL)
	}

	if p.primeOgmiosURL != "" && !common.IsValidURL(p.primeOgmiosURL) {
		return fmt.Errorf("invalid prime ogmios url: %s", p.primeOgmiosURL)
	}

	if p.vectorNetworkAddress == "" || !common.IsValidURL(p.vectorNetworkAddress) {
		return fmt.Errorf("invalid %s: %s", vectorNetworkAddressFlag, p.vectorNetworkAddress)
	}

	if p.vectorBlockfrostURL != "" && !common.IsValidURL(p.vectorBlockfrostURL) {
		return fmt.Errorf("invalid vector blockfrost url: %s", p.vectorBlockfrostURL)
	}

	if p.vectorOgmiosURL != "" && !common.IsValidURL(p.vectorOgmiosURL) {
		return fmt.Errorf("invalid vector ogmios url: %s", p.vectorOgmiosURL)
	}

	if p.vectorBlockfrostURL == "" && p.vectorSocketPath == "" && p.vectorOgmiosURL == "" {
		return fmt.Errorf("specify at least one of: %s, %s, %s",
			vectorBlockfrostURLFlag, vectorSocketPathFlag, vectorOgmiosURLFlag)
	}

	if p.bridgeNodeURL == "" || !common.IsValidURL(p.bridgeNodeURL) {
		return fmt.Errorf("invalid %s: %s", bridgeNodeURLFlag, p.bridgeNodeURL)
	}

	if p.bridgeSCAddress == "" {
		return fmt.Errorf("missing %s", bridgeSCAddressFlag)
	}

	if p.bridgeValidatorDataDir == "" && p.bridgeValidatorConfig == "" {
		return fmt.Errorf("specify at least one of: %s, %s", bridgeValidatorDataDirFlag, bridgeValidatorConfigFlag)
	}

	if len(p.apiKeys) == 0 {
		return fmt.Errorf("specify at least one %s", apiKeysFlag)
	}

	if p.telemetry != "" {
		parts := strings.Split(p.telemetry, ",")
		// common.IsValidURL(parts[0]) currently returns false for 0.0.0.0:5001
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" { //
			return fmt.Errorf("invalid telemetry: %s", p.telemetry)
		}
	}

	return nil
}

func (p *generateConfigsParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&p.primeNetworkAddress,
		primeNetworkAddressFlag,
		"",
		primeNetworkAddressFlagDesc,
	)
	cmd.Flags().Uint32Var(
		&p.primeNetworkMagic,
		primeNetworkMagicFlag,
		defaultNetworkMagic,
		primeNetworkMagicFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.primeKeysDir,
		primeKeysDirFlag,
		defaultPrimeKeysDir,
		primeKeysDirFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.primeOgmiosURL,
		primeOgmiosURLFlag,
		"",
		primeOgmiosURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.primeBlockfrostURL,
		primeBlockfrostURLFlag,
		"",
		primeBlockfrostURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.primeBlockfrostAPIKey,
		primeBlockfrostAPIKeyFlag,
		"",
		primeBlockfrostAPIKeyFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.primeSocketPath,
		primeSocketPathFlag,
		"",
		primeSocketPathFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.primeTTLSlotInc,
		primeTTLSlotIncFlag,
		defaultTTLSlotNumberInc,
		primeTTLSlotIncFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.primeSlotRoundingThreshold,
		primeSlotRoundingThresholdFlag,
		defaultSlotRoundingThreshold,
		primeSlotRoundingThresholdFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.vectorNetworkAddress,
		vectorNetworkAddressFlag,
		"",
		vectorNetworkAddressFlagDesc,
	)
	cmd.Flags().Uint32Var(
		&p.vectorNetworkMagic,
		vectorNetworkMagicFlag,
		defaultNetworkMagic,
		vectorNetworkMagicFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.vectorKeysDir,
		vectorKeysDirFlag,
		defaultVectorKeysDir,
		vectorKeysDirFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.vectorOgmiosURL,
		vectorOgmiosURLFlag,
		"",
		vectorOgmiosURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.vectorBlockfrostURL,
		vectorBlockfrostURLFlag,
		"",
		vectorBlockfrostURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.vectorBlockfrostAPIKey,
		vectorBlockfrostAPIKeyFlag,
		"",
		vectorBlockfrostAPIKeyFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.vectorSocketPath,
		vectorSocketPathFlag,
		"",
		vectorSocketPathFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.vectorTTLSlotInc,
		vectorTTLSlotIncFlag,
		defaultTTLSlotNumberInc,
		vectorTTLSlotIncFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.vectorSlotRoundingThreshold,
		vectorSlotRoundingThresholdFlag,
		defaultSlotRoundingThreshold,
		vectorSlotRoundingThresholdFlagDesc,
	)

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
	cmd.Flags().StringVar(
		&p.bridgeValidatorDataDir,
		bridgeValidatorDataDirFlag,
		"",
		bridgeValidatorDataDirFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.bridgeValidatorConfig,
		bridgeValidatorConfigFlag,
		"",
		bridgeValidatorConfigFlagDesc,
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

	cmd.MarkFlagsMutuallyExclusive(bridgeValidatorDataDirFlag, bridgeValidatorConfigFlag)
	cmd.MarkFlagsMutuallyExclusive(primeBlockfrostAPIKeyFlag, primeSocketPathFlag, primeOgmiosURLFlag)
	cmd.MarkFlagsMutuallyExclusive(vectorBlockfrostURLFlag, vectorSocketPathFlag, vectorOgmiosURLFlag)
}

func (p *generateConfigsParams) Execute() (common.ICommandResult, error) {
	validatorDataDir := p.bridgeValidatorDataDir
	if validatorDataDir != "" {
		validatorDataDir = path.Clean(validatorDataDir)
	}

	validatorConfig := p.bridgeValidatorConfig
	if validatorConfig != "" {
		validatorConfig = path.Clean(validatorConfig)
	}

	telemetryConfig := telemetry.TelemetryConfig{}

	if p.telemetry != "" {
		parts := strings.Split(p.telemetry, ",")

		telemetryConfig.PrometheusAddr = parts[0]
		telemetryConfig.DataDogAddr = parts[1]
	}

	vcConfig := &vcCore.AppConfig{
		CardanoChains: map[string]*vcCore.CardanoChainConfig{
			"prime": {
				NetworkAddress:           p.primeNetworkAddress,
				NetworkMagic:             p.primeNetworkMagic,
				StartBlockHash:           "",
				StartSlot:                0,
				StartBlockNumber:         0,
				ConfirmationBlockCount:   10,
				TTLSlotNumberInc:         p.primeTTLSlotInc,
				OtherAddressesOfInterest: []string{},
				KeysDirPath:              path.Clean(p.primeKeysDir),
				OgmiosURL:                p.primeOgmiosURL,
				BlockfrostURL:            p.primeBlockfrostURL,
				BlockfrostAPIKey:         p.primeBlockfrostAPIKey,
				SocketPath:               p.primeSocketPath,
				PotentialFee:             300000,
				SlotRoundingThreshold:    p.primeSlotRoundingThreshold,
			},
			"vector": {
				NetworkAddress:           p.vectorNetworkAddress,
				NetworkMagic:             p.vectorNetworkMagic,
				StartBlockHash:           "",
				StartSlot:                0,
				StartBlockNumber:         0,
				ConfirmationBlockCount:   10,
				TTLSlotNumberInc:         p.vectorTTLSlotInc,
				OtherAddressesOfInterest: []string{},
				KeysDirPath:              path.Clean(p.vectorKeysDir),
				OgmiosURL:                p.vectorOgmiosURL,
				BlockfrostURL:            p.vectorBlockfrostURL,
				BlockfrostAPIKey:         p.vectorBlockfrostAPIKey,
				SocketPath:               p.vectorSocketPath,
				PotentialFee:             300000,
				SlotRoundingThreshold:    p.vectorSlotRoundingThreshold,
			},
		},
		Bridge: oCore.BridgeConfig{
			NodeURL:              p.bridgeNodeURL,
			DynamicTx:            false,
			SmartContractAddress: p.bridgeSCAddress,
			ValidatorDataDir:     validatorDataDir,
			ValidatorConfigPath:  validatorConfig,
			SubmitConfig: oCore.SubmitConfig{
				ConfirmedBlocksThreshold:  20,
				ConfirmedBlocksSubmitTime: 3000,
			},
		},
		BridgingSettings: oCore.BridgingSettings{
			MinFeeForBridging:              1000010,
			UtxoMinValue:                   1000000,
			MaxReceiversPerBridgingRequest: 5,
			MaxBridgingClaimsToGroup:       5,
		},
		Settings: oCore.AppSettings{
			Logger: logger.LoggerConfig{
				LogFilePath:   path.Join(p.logsPath, "validator-components.log"),
				LogLevel:      hclog.Debug,
				JSONLogFormat: false,
				AppendFile:    true,
			},
			DbsPath: path.Join(p.dbsPath, "validatorcomponents"),
		},
		RelayerImitatorPullTimeMilis: 1000,
		BatcherPullTimeMilis:         2500,
		APIConfig: vcCore.APIConfig{
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

	primeChainSpecificJSONRaw, _ := json.Marshal(cardanotx.CardanoChainConfig{
		TestNetMagic:     p.primeNetworkMagic,
		OgmiosURL:        p.primeOgmiosURL,
		BlockfrostURL:    p.primeBlockfrostURL,
		BlockfrostAPIKey: p.primeBlockfrostAPIKey,
		SocketPath:       p.primeSocketPath,
		PotentialFee:     300000,
	})

	vectorChainSpecificJSONRaw, _ := json.Marshal(cardanotx.CardanoChainConfig{
		TestNetMagic:     p.vectorNetworkMagic,
		OgmiosURL:        p.vectorOgmiosURL,
		BlockfrostURL:    p.vectorBlockfrostURL,
		BlockfrostAPIKey: p.vectorBlockfrostAPIKey,
		SocketPath:       p.vectorSocketPath,
		PotentialFee:     300000,
	})

	rConfig := &rCore.RelayerManagerConfiguration{
		Bridge: rCore.BridgeConfig{
			NodeURL:              p.bridgeNodeURL,
			DynamicTx:            false,
			SmartContractAddress: p.bridgeSCAddress,
		},
		Chains: map[string]rCore.ChainConfig{
			"prime": {
				ChainType:     "Cardano",
				DbsPath:       path.Join(p.dbsPath, "relayer"),
				ChainSpecific: primeChainSpecificJSONRaw,
			},
			"vector": {
				ChainType:     "Cardano",
				DbsPath:       path.Join(p.dbsPath, "relayer"),
				ChainSpecific: vectorChainSpecificJSONRaw,
			},
		},
		PullTimeMilis: 1000,
		Logger: logger.LoggerConfig{
			LogFilePath:   path.Join(p.logsPath, "relayer.log"),
			LogLevel:      hclog.Debug,
			JSONLogFormat: false,
			AppendFile:    true,
		},
	}

	outputDirPath := path.Clean(p.outputDir)
	if err := common.CreateDirectoryIfNotExists(outputDirPath, 0770); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	vcConfigPath := path.Join(outputDirPath, p.outputValidatorComponentsFileName)
	if err := common.SaveJSON(vcConfigPath, vcConfig, true); err != nil {
		return nil, fmt.Errorf("failed to create validator components config json: %w", err)
	}

	rConfigPath := path.Join(outputDirPath, p.outputRelayerFileName)
	if err := common.SaveJSON(rConfigPath, rConfig, true); err != nil {
		return nil, fmt.Errorf("failed to create relayer config json: %w", err)
	}

	return &CmdResult{
		validatorComponentsConfigPath: vcConfigPath,
		relayerConfigPath:             rConfigPath,
	}, nil
}
