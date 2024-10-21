package cligenerateconfigs

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	rCore "github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	vcCore "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
)

const (
	primeNetworkAddressFlag        = "prime-network-address"
	primeNetworkMagicFlag          = "prime-network-magic"
	primeNetworkIDFlag             = "prime-network-id"
	primeOgmiosURLFlag             = "prime-ogmios-url"
	primeBlockfrostURLFlag         = "prime-blockfrost-url"
	primeBlockfrostAPIKeyFlag      = "prime-blockfrost-api-key"
	primeSocketPathFlag            = "prime-socket-path"
	primeTTLSlotIncFlag            = "prime-ttl-slot-inc"
	primeSlotRoundingThresholdFlag = "prime-slot-rounding-threshold"
	primeStartingBlockFlag         = "prime-starting-block"

	vectorNetworkAddressFlag        = "vector-network-address"
	vectorNetworkMagicFlag          = "vector-network-magic"
	vectorNetworkIDFlag             = "vector-network-id"
	vectorOgmiosURLFlag             = "vector-ogmios-url"
	vectorBlockfrostURLFlag         = "vector-blockfrost-url"
	vectorBlockfrostAPIKeyFlag      = "vector-blockfrost-api-key"
	vectorSocketPathFlag            = "vector-socket-path"
	vectorTTLSlotIncFlag            = "vector-ttl-slot-inc"
	vectorSlotRoundingThresholdFlag = "vector-slot-rounding-threshold"
	vectorStartingBlockFlag         = "vector-starting-block"

	bridgeNodeURLFlag   = "bridge-node-url"
	bridgeSCAddressFlag = "bridge-sc-address"

	validatorDataDirFlag = "validator-data-dir"
	validatorConfigFlag  = "validator-config"

	logsPathFlag = "logs-path"
	dbsPathFlag  = "dbs-path"

	apiPortFlag = "api-port"
	apiKeysFlag = "api-keys"

	outputDirFlag                         = "output-dir"
	outputValidatorComponentsFileNameFlag = "output-validator-components-file-name"
	outputRelayerFileNameFlag             = "output-relayer-file-name"

	telemetryFlag = "telemetry"

	nexusNodeURLFlag                = "nexus-node-url"
	nexusTTLBlockNumberIncFlag      = "nexus-ttl-block-inc"
	nexusBlockRoundingThresholdFlag = "nexus-block-rounding-threshold"
	nexusStartingBlockFlag          = "nexus-starting-block"
	relayerDataDirFlag              = "relayer-data-dir"
	relayerConfigPathFlag           = "relayer-config"

	primeNetworkAddressFlagDesc        = "(mandatory) address of prime network"
	primeNetworkMagicFlagDesc          = "prime network magic (default 0)"
	primeNetworkIDFlagDesc             = "prime network id"
	primeOgmiosURLFlagDesc             = "ogmios URL for prime network"
	primeBlockfrostURLFlagDesc         = "blockfrost URL for prime network"
	primeBlockfrostAPIKeyFlagDesc      = "blockfrost API key for prime network" //nolint:gosec
	primeSocketPathFlagDesc            = "socket path for prime network"
	primeTTLSlotIncFlagDesc            = "TTL slot increment for prime"
	primeSlotRoundingThresholdFlagDesc = "defines the upper limit used for rounding slot values for prime. Any slot value between 0 and `slotRoundingThreshold` will be rounded to `slotRoundingThreshold` etc" //nolint:lll
	primeStartingBlockFlagDesc         = "slot: hash of the block from where to start prime oracle"

	vectorNetworkAddressFlagDesc        = "(mandatory) address of vector network"
	vectorNetworkMagicFlagDesc          = "vector network magic (default 0)"
	vectorNetworkIDFlagDesc             = "vector network id"
	vectorOgmiosURLFlagDesc             = "ogmios URL for vector network"
	vectorBlockfrostURLFlagDesc         = "blockfrost URL for vector network"
	vectorBlockfrostAPIKeyFlagDesc      = "blockfrost API key for vector network" //nolint:gosec
	vectorSocketPathFlagDesc            = "socket path for vector network"
	vectorTTLSlotIncFlagDesc            = "TTL slot increment for vector"
	vectorSlotRoundingThresholdFlagDesc = "defines the upper limit used for rounding slot values for vector. Any slot value between 0 and `slotRoundingThreshold` will be rounded to `slotRoundingThreshold` etc" //nolint:lll
	vectorStartingBlockFlagDesc         = "slot: hash of the block from where to start vector oracle"

	bridgeNodeURLFlagDesc   = "(mandatory) node URL of bridge chain"
	bridgeSCAddressFlagDesc = "(mandatory) bridging smart contract address on bridge chain"

	validatorDataDirFlagDesc = "path to bridge chain data directory when using local secrets manager"
	validatorConfigFlagDesc  = "path to to bridge chain secrets manager config file"

	logsPathFlagDesc = "path to where logs will be stored"
	dbsPathFlagDesc  = "path to where databases will be stored"

	apiPortFlagDesc = "port at which API should run"
	apiKeysFlagDesc = "(mandatory) list of keys for API access"

	outputDirFlagDesc                         = "path to config jsons output directory"
	outputValidatorComponentsFileNameFlagDesc = "validator components config json output file name"
	outputRelayerFileNameFlagDesc             = "relayer config json output file name"

	telemetryFlagDesc = "prometheus_ip:port,datadog_ip:port"

	nexusNodeURLFlagDesc                = "nexus node URL"
	nexusTTLBlockNumberIncFlagDesc      = "TTL block increment for nexus"
	nexusBlockRoundingThresholdFlagDesc = "defines the upper limit used for rounding block values for nexus. Any block value between 0 and `blockRoundingThreshold` will be rounded to `blockRoundingThreshold` etc" //nolint:lll
	relayerDataDirFlagDesc              = "path to relayer secret directory when using local secrets manager"
	relayerConfigPathFlagDesc           = "path to relayer secrets manager config file"
	nexusStartingBlockFlagDesc          = "block from where to start nexus oracle"

	defaultPrimeBlockConfirmationCount       = 10
	defaultVectorBlockConfirmationCount      = 10
	defaultNetworkMagic                      = 0
	defaultLogsPath                          = "./logs"
	defaultDBsPath                           = "./db"
	defaultAPIPort                           = 10000
	defaultOutputDir                         = "./"
	defaultOutputValidatorComponentsFileName = "config.json"
	defaultOutputRelayerFileName             = "relayer_config.json"
	defaultPrimeTTLSlotNumberInc             = 1800 + defaultPrimeBlockConfirmationCount*10 // BlockTimeSeconds
	defaultPrimeSlotRoundingThreshold        = 60
	defaultVectorTTLSlotNumberInc            = 1800 + defaultVectorBlockConfirmationCount*10 // BlockTimeSeconds
	defaultVectorSlotRoundingThreshold       = 60
	defaultNexusBlockConfirmationCount       = 1 // try zero also because nexus is instant finality chain
	defaultNexusSyncBatchSize                = 20
	defaultNexusPoolIntervalMiliseconds      = 1500
	defaultNexusNoBatchPeriodPercent         = 0.2
	defaultNoBatchPeriodPercent              = 0.0625
	defaultTakeAtLeastUtxoCount              = 6
	defaultNexusTTLBlockRoundingThreshold    = 10
	defaultNexusTTLBlockNumberInc            = 20
)

type generateConfigsParams struct {
	primeNetworkAddress        string
	primeNetworkMagic          uint32
	primeNetworkID             uint32
	primeOgmiosURL             string
	primeBlockfrostURL         string
	primeBlockfrostAPIKey      string
	primeSocketPath            string
	primeTTLSlotInc            uint64
	primeSlotRoundingThreshold uint64
	primeStartingBlock         string

	vectorNetworkAddress        string
	vectorNetworkMagic          uint32
	vectorNetworkID             uint32
	vectorOgmiosURL             string
	vectorBlockfrostURL         string
	vectorBlockfrostAPIKey      string
	vectorSocketPath            string
	vectorTTLSlotInc            uint64
	vectorSlotRoundingThreshold uint64
	vectorStartingBlock         string

	bridgeNodeURL   string
	bridgeSCAddress string

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

	nexusNodeURL                string
	nexusTTLBlockNumberInc      uint64
	nexusBlockRoundingThreshold uint64
	nexusStartingBlock          uint64

	relayerDataDir    string
	relayerConfigPath string
}

func (p *generateConfigsParams) validateFlags() error {
	if !common.IsValidNetworkAddress(p.primeNetworkAddress) {
		return fmt.Errorf("invalid %s: %s", primeNetworkAddressFlag, p.primeNetworkAddress)
	}

	if p.primeBlockfrostURL == "" && p.primeSocketPath == "" && p.primeOgmiosURL == "" {
		return fmt.Errorf("specify at least one of: %s, %s, %s",
			primeBlockfrostURLFlag, primeSocketPathFlag, primeOgmiosURLFlag)
	}

	if p.primeBlockfrostURL != "" && !common.IsValidHTTPURL(p.primeBlockfrostURL) {
		return fmt.Errorf("invalid prime blockfrost url: %s", p.primeBlockfrostURL)
	}

	if p.primeOgmiosURL != "" && !common.IsValidHTTPURL(p.primeOgmiosURL) {
		return fmt.Errorf("invalid prime ogmios url: %s", p.primeOgmiosURL)
	}

	if !common.IsValidNetworkAddress(p.vectorNetworkAddress) {
		return fmt.Errorf("invalid %s: %s", vectorNetworkAddressFlag, p.vectorNetworkAddress)
	}

	if p.vectorBlockfrostURL == "" && p.vectorSocketPath == "" && p.vectorOgmiosURL == "" {
		return fmt.Errorf("specify at least one of: %s, %s, %s",
			vectorBlockfrostURLFlag, vectorSocketPathFlag, vectorOgmiosURLFlag)
	}

	if p.vectorBlockfrostURL != "" && !common.IsValidHTTPURL(p.vectorBlockfrostURL) {
		return fmt.Errorf("invalid vector blockfrost url: %s", p.vectorBlockfrostURL)
	}

	if p.vectorOgmiosURL != "" && !common.IsValidHTTPURL(p.vectorOgmiosURL) {
		return fmt.Errorf("invalid vector ogmios url: %s", p.vectorOgmiosURL)
	}

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

	if p.telemetry != "" && !common.IsValidNetworkAddress(p.telemetry) {
		return fmt.Errorf("invalid telemetry: %s", p.telemetry)
	}

	if p.primeStartingBlock != "" {
		parts := strings.Split(p.primeStartingBlock, ":")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("invalid prime starting block: %s", p.primeStartingBlock)
		}
	}

	if p.vectorStartingBlock != "" {
		parts := strings.Split(p.vectorStartingBlock, ":")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("invalid vector starting block: %s", p.vectorStartingBlock)
		}
	}

	if !common.IsValidHTTPURL(p.nexusNodeURL) {
		return fmt.Errorf("invalid %s: %s", nexusNodeURLFlag, p.nexusNodeURL)
	}

	if p.relayerDataDir == "" && p.relayerConfigPath == "" {
		return fmt.Errorf("specify at least one of: %s, %s", relayerDataDirFlag, relayerConfigPathFlag)
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
	cmd.Flags().Uint32Var(
		&p.primeNetworkID,
		primeNetworkIDFlag,
		uint32(wallet.MainNetNetwork),
		primeNetworkIDFlagDesc,
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
		defaultPrimeTTLSlotNumberInc,
		primeTTLSlotIncFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.primeSlotRoundingThreshold,
		primeSlotRoundingThresholdFlag,
		defaultPrimeSlotRoundingThreshold,
		primeSlotRoundingThresholdFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.primeStartingBlock,
		primeStartingBlockFlag,
		"",
		primeStartingBlockFlagDesc,
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
	cmd.Flags().Uint32Var(
		&p.vectorNetworkID,
		vectorNetworkIDFlag,
		uint32(wallet.VectorMainNetNetwork),
		vectorNetworkIDFlagDesc,
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
		defaultVectorTTLSlotNumberInc,
		vectorTTLSlotIncFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.vectorSlotRoundingThreshold,
		vectorSlotRoundingThresholdFlag,
		defaultVectorSlotRoundingThreshold,
		vectorSlotRoundingThresholdFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.vectorStartingBlock,
		vectorStartingBlockFlag,
		"",
		vectorStartingBlockFlagDesc,
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

	cmd.Flags().StringVar(
		&p.nexusNodeURL,
		nexusNodeURLFlag,
		"",
		nexusNodeURLFlagDesc,
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
	cmd.Flags().Uint64Var(
		&p.nexusTTLBlockNumberInc,
		nexusTTLBlockNumberIncFlag,
		defaultNexusTTLBlockNumberInc,
		nexusTTLBlockNumberIncFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.nexusBlockRoundingThreshold,
		nexusBlockRoundingThresholdFlag,
		defaultNexusTTLBlockRoundingThreshold,
		nexusBlockRoundingThresholdFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.nexusStartingBlock,
		nexusStartingBlockFlag,
		0,
		nexusStartingBlockFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(validatorDataDirFlag, validatorConfigFlag)
	cmd.MarkFlagsMutuallyExclusive(relayerDataDirFlag, relayerConfigPathFlag)
	cmd.MarkFlagsMutuallyExclusive(primeBlockfrostAPIKeyFlag, primeSocketPathFlag, primeOgmiosURLFlag)
	cmd.MarkFlagsMutuallyExclusive(vectorBlockfrostURLFlag, vectorSocketPathFlag, vectorOgmiosURLFlag)
}

func (p *generateConfigsParams) Execute() (common.ICommandResult, error) {
	telemetryConfig := telemetry.TelemetryConfig{}

	if p.telemetry != "" {
		parts := strings.Split(p.telemetry, ",")

		telemetryConfig.PrometheusAddr = parts[0]
		telemetryConfig.DataDogAddr = parts[1]
	}

	primeStartingSlot, primeStartingHash, err := parseStartingBlock(p.primeStartingBlock)
	if err != nil {
		return nil, err
	}

	vectorStartingSlot, vectorStartingHash, err := parseStartingBlock(p.vectorStartingBlock)
	if err != nil {
		return nil, err
	}

	vcConfig := &vcCore.AppConfig{
		ValidatorDataDir:    cleanPath(p.validatorDataDir),
		ValidatorConfigPath: cleanPath(p.validatorConfig),
		CardanoChains: map[string]*oCore.CardanoChainConfig{
			common.ChainIDStrPrime: {
				NetworkAddress:           p.primeNetworkAddress,
				NetworkMagic:             p.primeNetworkMagic,
				NetworkID:                wallet.CardanoNetworkType(p.primeNetworkID),
				StartBlockHash:           primeStartingHash,
				StartSlot:                primeStartingSlot,
				ConfirmationBlockCount:   defaultPrimeBlockConfirmationCount,
				TTLSlotNumberInc:         p.primeTTLSlotInc,
				OtherAddressesOfInterest: []string{},
				OgmiosURL:                p.primeOgmiosURL,
				BlockfrostURL:            p.primeBlockfrostURL,
				BlockfrostAPIKey:         p.primeBlockfrostAPIKey,
				SocketPath:               p.primeSocketPath,
				PotentialFee:             300000,
				SlotRoundingThreshold:    p.primeSlotRoundingThreshold,
				NoBatchPeriodPercent:     defaultNoBatchPeriodPercent,
				TakeAtLeastUtxoCount:     defaultTakeAtLeastUtxoCount,
			},
			common.ChainIDStrVector: {
				NetworkAddress:           p.vectorNetworkAddress,
				NetworkMagic:             p.vectorNetworkMagic,
				NetworkID:                wallet.CardanoNetworkType(p.vectorNetworkID),
				StartBlockHash:           vectorStartingHash,
				StartSlot:                vectorStartingSlot,
				ConfirmationBlockCount:   defaultVectorBlockConfirmationCount,
				TTLSlotNumberInc:         p.vectorTTLSlotInc,
				OtherAddressesOfInterest: []string{},
				OgmiosURL:                p.vectorOgmiosURL,
				BlockfrostURL:            p.vectorBlockfrostURL,
				BlockfrostAPIKey:         p.vectorBlockfrostAPIKey,
				SocketPath:               p.vectorSocketPath,
				PotentialFee:             300000,
				SlotRoundingThreshold:    p.vectorSlotRoundingThreshold,
				NoBatchPeriodPercent:     defaultNoBatchPeriodPercent,
				TakeAtLeastUtxoCount:     defaultTakeAtLeastUtxoCount,
			},
		},
		EthChains: map[string]*oCore.EthChainConfig{
			common.ChainIDStrNexus: {
				NodeURL:                 p.nexusNodeURL,
				SyncBatchSize:           defaultNexusSyncBatchSize,
				NumBlockConfirmations:   defaultNexusBlockConfirmationCount,
				StartBlockNumber:        p.nexusStartingBlock,
				PoolIntervalMiliseconds: defaultNexusPoolIntervalMiliseconds,
				TTLBlockNumberInc:       p.nexusTTLBlockNumberInc,
				BlockRoundingThreshold:  p.nexusBlockRoundingThreshold,
				NoBatchPeriodPercent:    defaultNexusNoBatchPeriodPercent,
				DynamicTx:               true,
			},
		},
		Bridge: oCore.BridgeConfig{
			NodeURL:              p.bridgeNodeURL,
			DynamicTx:            false,
			SmartContractAddress: p.bridgeSCAddress,
			SubmitConfig: oCore.SubmitConfig{
				ConfirmedBlocksThreshold:  20,
				ConfirmedBlocksSubmitTime: 3000,
			},
		},
		BridgingSettings: oCore.BridgingSettings{
			MinFeeForBridging:              1000010,
			UtxoMinValue:                   1000000,
			MaxReceiversPerBridgingRequest: 4, // 4 + 1 for fee
			MaxBridgingClaimsToGroup:       5,
		},
		Settings: oCore.AppSettings{
			Logger: logger.LoggerConfig{
				LogFilePath:   filepath.Join(p.logsPath, "validator-components.log"),
				LogLevel:      hclog.Debug,
				JSONLogFormat: false,
				AppendFile:    true,
			},
			DbsPath: filepath.Join(p.dbsPath, "validatorcomponents"),
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
		NetworkID:        wallet.CardanoNetworkType(p.primeNetworkID),
		TestNetMagic:     p.primeNetworkMagic,
		OgmiosURL:        p.primeOgmiosURL,
		BlockfrostURL:    p.primeBlockfrostURL,
		BlockfrostAPIKey: p.primeBlockfrostAPIKey,
		SocketPath:       p.primeSocketPath,
		PotentialFee:     300000,
	})

	vectorChainSpecificJSONRaw, _ := json.Marshal(cardanotx.CardanoChainConfig{
		NetworkID:        wallet.CardanoNetworkType(p.vectorNetworkID),
		TestNetMagic:     p.vectorNetworkMagic,
		OgmiosURL:        p.vectorOgmiosURL,
		BlockfrostURL:    p.vectorBlockfrostURL,
		BlockfrostAPIKey: p.vectorBlockfrostAPIKey,
		SocketPath:       p.vectorSocketPath,
		PotentialFee:     300000,
	})

	nexusChainSpecificJSONRaw, _ := json.Marshal(cardanotx.RelayerEVMChainConfig{
		NodeURL:    p.nexusNodeURL,
		DataDir:    cleanPath(p.relayerDataDir),
		ConfigPath: cleanPath(p.relayerConfigPath),
		DynamicTx:  true,
	})

	rConfig := &rCore.RelayerManagerConfiguration{
		Bridge: rCore.BridgeConfig{
			NodeURL:              p.bridgeNodeURL,
			DynamicTx:            false,
			SmartContractAddress: p.bridgeSCAddress,
		},
		Chains: map[string]rCore.ChainConfig{
			common.ChainIDStrPrime: {
				ChainType:     common.ChainTypeCardanoStr,
				DbsPath:       filepath.Join(p.dbsPath, "relayer"),
				ChainSpecific: primeChainSpecificJSONRaw,
			},
			common.ChainIDStrVector: {
				ChainType:     common.ChainTypeCardanoStr,
				DbsPath:       filepath.Join(p.dbsPath, "relayer"),
				ChainSpecific: vectorChainSpecificJSONRaw,
			},
			common.ChainIDStrNexus: {
				ChainType:     common.ChainTypeEVMStr,
				DbsPath:       filepath.Join(p.dbsPath, "relayer"),
				ChainSpecific: nexusChainSpecificJSONRaw,
			},
		},
		PullTimeMilis: 1000,
		Logger: logger.LoggerConfig{
			LogFilePath:   filepath.Join(p.logsPath, "relayer.log"),
			LogLevel:      hclog.Debug,
			JSONLogFormat: false,
			AppendFile:    true,
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

func cleanPath(path string) string {
	if path != "" {
		return filepath.Clean(path)
	}

	return ""
}

func parseStartingBlock(s string) (uint64, string, error) {
	if s == "" {
		return 0, "", nil
	}

	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, "", errors.New("invalid starting block")
	}

	val, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, "", err
	}

	return val, parts[1], nil
}
