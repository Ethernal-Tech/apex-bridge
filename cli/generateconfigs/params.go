package cligenerateconfigs

import (
	"encoding/json"
	"fmt"
	"path"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
	rCore "github.com/Ethernal-Tech/apex-bridge/relayer/core"
	vcCore "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
)

const (
	primeNetworkAddressFlag   = "prime-network-address"
	primeNetworkMagicFlag     = "prime-network-magic"
	primeKeysDirFlag          = "prime-keys-dir"
	primeBlockfrostUrlFlag    = "prime-blockfrost-url"
	primeBlockfrostApiKeyFlag = "prime-blockfrost-api-key"
	primeSocketPathFlag       = "prime-socket-path"

	vectorNetworkAddressFlag   = "vector-network-address"
	vectorNetworkMagicFlag     = "vector-network-magic"
	vectorKeysDirFlag          = "vector-keys-dir"
	vectorBlockfrostUrlFlag    = "vector-blockfrost-url"
	vectorBlockfrostApiKeyFlag = "vector-blockfrost-api-key"
	vectorSocketPathFlag       = "vector-socket-path"

	bridgeNodeUrlFlag          = "bridge-node-url"
	bridgeSCAddressFlag        = "bridge-sc-address"
	bridgeValidatorDataDirFlag = "bridge-validator-data-dir"
	bridgeValidatorConfigFlag  = "bridge-validator-config-path"

	logsPathFlag = "logs-path"
	dbsPathFlag  = "dbs-path"

	apiPortFlag = "api-port"
	apiKeysFlag = "api-keys"

	outputDirFlag                         = "output-dir"
	outputValidatorComponentsFileNameFlag = "output-validator-components-file-name"
	outputRelayerFileNameFlag             = "output-relayer-file-name"

	primeNetworkAddressFlagDesc   = "(Mandatory) Address of prime network"
	primeNetworkMagicFlagDesc     = "Network magic of prime network (default 0)"
	primeKeysDirFlagDesc          = "Path to cardano keys directory for prime network"
	primeBlockfrostUrlFlagDesc    = "(Mandatory if prime-socket-path not specified) Blockfrost URL for prime network"
	primeBlockfrostApiKeyFlagDesc = "Blockfrost API key for prime network"
	primeSocketPathFlagDesc       = "(Mandatory if prime-blockfrost-url not specified) Socket path for prime network"

	vectorNetworkAddressFlagDesc   = "(Mandatory) Address of vector network"
	vectorNetworkMagicFlagDesc     = "Network magic of vector network (default 0)"
	vectorKeysDirFlagDesc          = "Path to cardano keys directory for vector network"
	vectorBlockfrostUrlFlagDesc    = "(Mandatory if vector-socket-path not specified) Blockfrost URL for vector network"
	vectorBlockfrostApiKeyFlagDesc = "Blockfrost API key for vector network"
	vectorSocketPathFlagDesc       = "(Mandatory if vector-blockfrost-url not specified) Socket path for vector network"

	bridgeNodeUrlFlagDesc          = "(Mandatory) Node URL of bridge chain"
	bridgeSCAddressFlagDesc        = "(Mandatory) Bridging smart contract address on bridge chain"
	bridgeValidatorDataDirFlagDesc = "(Mandatory if bridge-validator-config not specified) Path to bridge chain data directory when using local secrets manager"
	bridgeValidatorConfigFlagDesc  = "(Mandatory if bridge-validator-data not specified) Path to to bridge chain secrets manager config file"

	logsPathFlagDesc = "Path to where logs will be stored"
	dbsPathFlagDesc  = "Path to where databases will be stored"

	apiPortFlagDesc = "Port at which API should run"
	apiKeysFlagDesc = "(Mandatory) List of keys for API access"

	outputDirFlagDesc                         = "Path to config jsons output directory"
	outputValidatorComponentsFileNameFlagDesc = "Validator components config json output file name"
	outputRelayerFileNameFlagDesc             = "Relayer config json output file name"

	defaultNetworkMagic                      = 0
	defaultPrimeKeysDir                      = "./keys/prime"
	defaultVectorKeysDir                     = "./keys/vector"
	defaultLogsPath                          = "./logs"
	defaultDBsPath                           = "./db"
	defaultApiPort                           = 10000
	defaultOutputDir                         = "./"
	defaultOutputValidatorComponentsFileName = "config.json"
	defaultOutputRelayerFileName             = "relayer_config.json"
)

type generateConfigsParams struct {
	primeNetworkAddress   string
	primeNetworkMagic     uint32
	primeKeysDir          string
	primeBlockfrostUrl    string
	primeBlockfrostApiKey string
	primeSocketPath       string

	vectorNetworkAddress   string
	vectorNetworkMagic     uint32
	vectorKeysDir          string
	vectorBlockfrostUrl    string
	vectorBlockfrostApiKey string
	vectorSocketPath       string

	bridgeNodeUrl          string
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
}

func (p *generateConfigsParams) validateFlags() error {
	if p.primeNetworkAddress == "" || !common.IsValidURL(p.primeNetworkAddress) {
		return fmt.Errorf("invalid %s: %s", primeNetworkAddressFlag, p.primeNetworkAddress)
	}
	if p.primeBlockfrostUrl == "" && p.primeSocketPath == "" {
		return fmt.Errorf("specify at least one of: %s, %s", primeBlockfrostUrlFlag, primeSocketPathFlag)
	}

	if p.vectorNetworkAddress == "" || !common.IsValidURL(p.vectorNetworkAddress) {
		return fmt.Errorf("invalid %s: %s", vectorNetworkAddressFlag, p.vectorNetworkAddress)
	}
	if p.vectorBlockfrostUrl == "" && p.vectorSocketPath == "" {
		return fmt.Errorf("specify at least one of: %s, %s", vectorBlockfrostUrlFlag, vectorSocketPathFlag)
	}

	if p.bridgeNodeUrl == "" || !common.IsValidURL(p.bridgeNodeUrl) {
		return fmt.Errorf("invalid %s: %s", bridgeNodeUrlFlag, p.bridgeNodeUrl)
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
		&p.primeBlockfrostUrl,
		primeBlockfrostUrlFlag,
		"",
		primeBlockfrostUrlFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.primeBlockfrostApiKey,
		primeBlockfrostApiKeyFlag,
		"",
		primeBlockfrostApiKeyFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.primeSocketPath,
		primeSocketPathFlag,
		"",
		primeSocketPathFlagDesc,
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
		&p.vectorBlockfrostUrl,
		vectorBlockfrostUrlFlag,
		"",
		vectorBlockfrostUrlFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.vectorBlockfrostApiKey,
		vectorBlockfrostApiKeyFlag,
		"",
		vectorBlockfrostApiKeyFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.vectorSocketPath,
		vectorSocketPathFlag,
		"",
		vectorSocketPathFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.bridgeNodeUrl,
		bridgeNodeUrlFlag,
		"",
		bridgeNodeUrlFlagDesc,
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
		defaultApiPort,
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
	vcConfig := &vcCore.AppConfig{
		CardanoChains: map[string]*vcCore.CardanoChainConfig{
			"prime": {
				NetworkAddress:           p.primeNetworkAddress,
				NetworkMagic:             p.primeNetworkMagic,
				StartBlockHash:           "",
				StartSlot:                0,
				StartBlockNumber:         0,
				ConfirmationBlockCount:   10,
				OtherAddressesOfInterest: []string{},
				KeysDirPath:              path.Clean(p.primeKeysDir),
				BlockfrostUrl:            p.primeBlockfrostUrl,
				BlockfrostAPIKey:         p.primeBlockfrostApiKey,
				SocketPath:               p.primeSocketPath,
				PotentialFee:             300000,
			},
			"vector": {
				NetworkAddress:           p.vectorNetworkAddress,
				NetworkMagic:             p.vectorNetworkMagic,
				StartBlockHash:           "",
				StartSlot:                0,
				StartBlockNumber:         0,
				ConfirmationBlockCount:   10,
				OtherAddressesOfInterest: []string{},
				KeysDirPath:              path.Clean(p.vectorKeysDir),
				BlockfrostUrl:            p.vectorBlockfrostUrl,
				BlockfrostAPIKey:         p.vectorBlockfrostApiKey,
				SocketPath:               p.vectorSocketPath,
				PotentialFee:             300000,
			},
		},
		Bridge: oCore.BridgeConfig{
			NodeUrl:              p.bridgeNodeUrl,
			SmartContractAddress: p.bridgeSCAddress,
			ValidatorDataDir:     validatorDataDir,
			ValidatorConfigPath:  validatorConfig,
			SubmitConfig: oCore.SubmitConfig{
				ConfirmedBlocksThreshold:  10,
				ConfirmedBlocksSubmitTime: 5000,
			},
		},
		BridgingSettings: oCore.BridgingSettings{
			MinFeeForBridging:              1000010,
			UtxoMinValue:                   1000000,
			MaxReceiversPerBridgingRequest: 5,
			MaxBridgingClaimsToGroup:       10,
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
		ApiConfig: vcCore.ApiConfig{
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
			ApiKeyHeader: "x-api-key",
			ApiKeys:      p.apiKeys,
		},
	}

	primeChainSpecificJsonRaw, _ := json.Marshal(cardanotx.CardanoChainConfig{
		TestNetMagic:     p.primeNetworkMagic,
		BlockfrostUrl:    p.primeBlockfrostUrl,
		BlockfrostAPIKey: p.primeBlockfrostApiKey,
		SocketPath:       p.primeSocketPath,
		PotentialFee:     300000,
	})

	vectorChainSpecificJsonRaw, _ := json.Marshal(cardanotx.CardanoChainConfig{
		TestNetMagic:     p.vectorNetworkMagic,
		BlockfrostUrl:    p.vectorBlockfrostUrl,
		BlockfrostAPIKey: p.vectorBlockfrostApiKey,
		SocketPath:       p.vectorSocketPath,
		PotentialFee:     300000,
	})

	rConfig := &rCore.RelayerManagerConfiguration{
		Bridge: rCore.BridgeConfig{
			NodeUrl:              p.bridgeNodeUrl,
			SmartContractAddress: p.bridgeSCAddress,
		},
		Chains: map[string]rCore.ChainConfig{
			"prime": {
				ChainType:     "Cardano",
				DbsPath:       path.Join(p.dbsPath, "relayer"),
				ChainSpecific: primeChainSpecificJsonRaw,
			},
			"vector": {
				ChainType:     "Cardano",
				DbsPath:       path.Join(p.dbsPath, "relayer"),
				ChainSpecific: vectorChainSpecificJsonRaw,
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
	if err := common.SaveJson(vcConfigPath, vcConfig, true); err != nil {
		return nil, fmt.Errorf("failed to create validator components config json: %w", err)
	}

	rConfigPath := path.Join(outputDirPath, p.outputRelayerFileName)
	if err := common.SaveJson(rConfigPath, rConfig, true); err != nil {
		return nil, fmt.Errorf("failed to create relayer config json: %w", err)
	}

	return &CmdResult{
		validatorComponentsConfigPath: vcConfigPath,
		relayerConfigPath:             rConfigPath,
	}, nil
}
