package cligenerateconfigs

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	apiCore "github.com/Ethernal-Tech/apex-bridge/api/core"
	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	rCore "github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	vcCore "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
)

const (
	primeMinOperationFeeFlag     = "prime-min-operation-fee"
	primeMinOperationFeeFlagDesc = "minimal operation fee for prime"

	primeCardanoWrappedTokenNameFlag     = "prime-cardano-token-name"
	primeCardanoWrappedTokenNameFlagDesc = "wrapped token name for Cardano Ada"

	cardanoPrimeWrappedTokenNameFlag     = "cardano-prime-token-name"
	cardanoPrimeWrappedTokenNameFlagDesc = "wrapped token name for Prime Apex"

	cardanoNetworkAddressFlag         = "cardano-network-address"
	cardanoNetworkMagicFlag           = "cardano-network-magic"
	cardanoNetworkIDFlag              = "cardano-network-id"
	cardanoOgmiosURLFlag              = "cardano-ogmios-url"
	cardanoBlockfrostURLFlag          = "cardano-blockfrost-url"
	cardanoBlockfrostAPIKeyFlag       = "cardano-blockfrost-api-key" //nolint:gosec
	cardanoSocketPathFlag             = "cardano-socket-path"
	cardanoTTLSlotIncFlag             = "cardano-ttl-slot-inc"
	cardanoSlotRoundingThresholdFlag  = "cardano-slot-rounding-threshold"
	cardanoStartingBlockFlag          = "cardano-starting-block"
	cardanoUtxoMinAmountFlag          = "cardano-utxo-min-amount"
	cardanoMinFeeForBridgingFlag      = "cardano-min-fee-for-bridging"
	cardanoMinOperationFeeFlag        = "cardano-min-operation-fee"
	cardanoBlockConfirmationCountFlag = "cardano-block-confirmation-count"

	cardanoNetworkAddressFlagDesc         = "(mandatory) address of cardano network"
	cardanoNetworkMagicFlagDesc           = "cardano network magic (default 0)"
	cardanoNetworkIDFlagDesc              = "cardano network id"
	cardanoOgmiosURLFlagDesc              = "ogmios URL for cardano network"
	cardanoBlockfrostURLFlagDesc          = "blockfrost URL for cardano network"
	cardanoBlockfrostAPIKeyFlagDesc       = "blockfrost API key for cardano network" //nolint:gosec
	cardanoSocketPathFlagDesc             = "socket path for cardano network"
	cardanoTTLSlotIncFlagDesc             = "TTL slot increment for cardano"
	cardanoSlotRoundingThresholdFlagDesc  = "defines the upper limit used for rounding slot values for cardano. Any slot value between 0 and `slotRoundingThreshold` will be rounded to `slotRoundingThreshold` etc" //nolint:lll
	cardanoStartingBlockFlagDesc          = "slot: hash of the block from where to start cardano oracle / cardano block submitter"                                                                                   //nolint:lll
	cardanoUtxoMinAmountFlagDesc          = "minimal UTXO value for cardano"
	cardanoMinFeeForBridgingFlagDesc      = "minimal bridging fee for cardano"
	cardanoMinOperationFeeFlagDesc        = "minimal operation fee for cardano"
	cardanoBlockConfirmationCountFlagDesc = "block confirmation count for cardano"

	defaultCardanoBlockConfirmationCount = 10
	defaultCardanoTTLSlotNumberInc       = 1800 + defaultCardanoBlockConfirmationCount*10 // BlockTimeSeconds
	defaultCardanoSlotRoundingThreshold  = 60
)

type skylineGenerateConfigsParams struct {
	primeNetworkAddress         string
	primeNetworkMagic           uint32
	primeNetworkID              uint32
	primeOgmiosURL              string
	primeBlockfrostURL          string
	primeBlockfrostAPIKey       string
	primeSocketPath             string
	primeTTLSlotInc             uint64
	primeSlotRoundingThreshold  uint64
	primeStartingBlock          string
	primeUtxoMinAmount          uint64
	primeMinFeeForBridging      uint64
	primeMinOperationFee        uint64
	primeBlockConfirmationCount uint

	cardanoNetworkAddress         string
	cardanoNetworkMagic           uint32
	cardanoNetworkID              uint32
	cardanoOgmiosURL              string
	cardanoBlockfrostURL          string
	cardanoBlockfrostAPIKey       string
	cardanoSocketPath             string
	cardanoTTLSlotInc             uint64
	cardanoSlotRoundingThreshold  uint64
	cardanoStartingBlock          string
	cardanoUtxoMinAmount          uint64
	cardanoMinFeeForBridging      uint64
	cardanoMinOperationFee        uint64
	cardanoBlockConfirmationCount uint

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

	relayerDataDir    string
	relayerConfigPath string

	cardanoPrimeWrappedTokenName string
	primeCardanoWrappedTokenName string
}

func (p *skylineGenerateConfigsParams) validateFlags() error {
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

	if !common.IsValidNetworkAddress(p.cardanoNetworkAddress) {
		return fmt.Errorf("invalid %s: %s", cardanoNetworkAddressFlag, p.cardanoNetworkAddress)
	}

	if p.cardanoBlockfrostURL == "" && p.cardanoSocketPath == "" && p.cardanoOgmiosURL == "" {
		return fmt.Errorf("specify at least one of: %s, %s, %s",
			cardanoBlockfrostURLFlag, cardanoSocketPathFlag, cardanoOgmiosURLFlag)
	}

	if p.cardanoBlockfrostURL != "" && !common.IsValidHTTPURL(p.cardanoBlockfrostURL) {
		return fmt.Errorf("invalid cardano blockfrost url: %s", p.cardanoBlockfrostURL)
	}

	if p.cardanoOgmiosURL != "" && !common.IsValidHTTPURL(p.cardanoOgmiosURL) {
		return fmt.Errorf("invalid cardano ogmios url: %s", p.cardanoOgmiosURL)
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

	if p.telemetry != "" {
		parts := strings.Split(p.telemetry, ",")
		if len(parts) < 1 || len(parts) > 2 || !common.IsValidNetworkAddress(strings.TrimSpace(parts[0])) ||
			(len(parts) == 2 && !common.IsValidNetworkAddress(strings.TrimSpace(parts[1]))) {
			return fmt.Errorf("invalid telemetry: %s", p.telemetry)
		}
	}

	if p.primeStartingBlock != "" {
		parts := strings.Split(p.primeStartingBlock, ":")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("invalid prime starting block: %s", p.primeStartingBlock)
		}
	}

	if p.cardanoStartingBlock != "" {
		parts := strings.Split(p.cardanoStartingBlock, ":")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("invalid cardano starting block: %s", p.cardanoStartingBlock)
		}
	}

	if p.primeMinFeeForBridging < p.primeUtxoMinAmount {
		return fmt.Errorf("prime minimal fee for bridging: %d should't be less than minimal UTXO amount: %d",
			p.primeMinFeeForBridging, p.primeUtxoMinAmount)
	}

	if p.cardanoMinFeeForBridging < p.cardanoUtxoMinAmount {
		return fmt.Errorf("cardano minimal fee for bridging: %d should't be less than minimal UTXO amount: %d",
			p.cardanoMinFeeForBridging, p.cardanoUtxoMinAmount)
	}

	if p.relayerDataDir == "" && p.relayerConfigPath == "" {
		return fmt.Errorf("specify at least one of: %s, %s", relayerDataDirFlag, relayerConfigPathFlag)
	}

	if p.primeCardanoWrappedTokenName != "" {
		if _, err := wallet.NewTokenWithFullNameTry(p.primeCardanoWrappedTokenName); err != nil {
			return fmt.Errorf("invalid token name %s", primeCardanoWrappedTokenNameFlag)
		}
	}

	if p.cardanoPrimeWrappedTokenName != "" {
		if _, err := wallet.NewTokenWithFullNameTry(p.cardanoPrimeWrappedTokenName); err != nil {
			return fmt.Errorf("invalid token name %s", cardanoPrimeWrappedTokenNameFlag)
		}
	}

	return nil
}

func (p *skylineGenerateConfigsParams) setFlags(cmd *cobra.Command) {
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
	cmd.Flags().Uint64Var(
		&p.primeUtxoMinAmount,
		primeUtxoMinAmountFlag,
		common.MinUtxoAmountDefault,
		primeUtxoMinAmountFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.primeMinFeeForBridging,
		primeMinFeeForBridgingFlag,
		common.MinFeeForBridgingDefault,
		primeMinFeeForBridgingFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.primeMinOperationFee,
		primeMinOperationFeeFlag,
		common.MinOperationFeeOnPrime,
		primeMinOperationFeeFlagDesc,
	)
	cmd.Flags().UintVar(
		&p.primeBlockConfirmationCount,
		primeBlockConfirmationCountFlag,
		defaultPrimeBlockConfirmationCount,
		primeBlockConfirmationCountFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.cardanoNetworkAddress,
		cardanoNetworkAddressFlag,
		"",
		cardanoNetworkAddressFlagDesc,
	)
	cmd.Flags().Uint32Var(
		&p.cardanoNetworkMagic,
		cardanoNetworkMagicFlag,
		defaultNetworkMagic,
		cardanoNetworkMagicFlagDesc,
	)
	cmd.Flags().Uint32Var(
		&p.cardanoNetworkID,
		cardanoNetworkIDFlag,
		uint32(wallet.MainNetNetwork),
		cardanoNetworkIDFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.cardanoOgmiosURL,
		cardanoOgmiosURLFlag,
		"",
		cardanoOgmiosURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.cardanoBlockfrostURL,
		cardanoBlockfrostURLFlag,
		"",
		cardanoBlockfrostURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.cardanoBlockfrostAPIKey,
		cardanoBlockfrostAPIKeyFlag,
		"",
		cardanoBlockfrostAPIKeyFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.cardanoSocketPath,
		cardanoSocketPathFlag,
		"",
		cardanoSocketPathFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.cardanoTTLSlotInc,
		cardanoTTLSlotIncFlag,
		defaultCardanoTTLSlotNumberInc,
		cardanoTTLSlotIncFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.cardanoSlotRoundingThreshold,
		cardanoSlotRoundingThresholdFlag,
		defaultCardanoSlotRoundingThreshold,
		cardanoSlotRoundingThresholdFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.cardanoStartingBlock,
		cardanoStartingBlockFlag,
		"",
		cardanoStartingBlockFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.cardanoUtxoMinAmount,
		cardanoUtxoMinAmountFlag,
		common.MinUtxoAmountDefault,
		cardanoUtxoMinAmountFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.cardanoMinFeeForBridging,
		cardanoMinFeeForBridgingFlag,
		common.MinFeeForBridgingDefault,
		cardanoMinFeeForBridgingFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.cardanoMinOperationFee,
		cardanoMinOperationFeeFlag,
		common.MinOperationFeeOnCardano,
		cardanoMinOperationFeeFlagDesc,
	)
	cmd.Flags().UintVar(
		&p.cardanoBlockConfirmationCount,
		cardanoBlockConfirmationCountFlag,
		defaultCardanoBlockConfirmationCount,
		cardanoBlockConfirmationCountFlagDesc,
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
		&p.primeCardanoWrappedTokenName,
		primeCardanoWrappedTokenNameFlag,
		"",
		primeCardanoWrappedTokenNameFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.cardanoPrimeWrappedTokenName,
		cardanoPrimeWrappedTokenNameFlag,
		"",
		cardanoPrimeWrappedTokenNameFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(validatorDataDirFlag, validatorConfigFlag)
	cmd.MarkFlagsMutuallyExclusive(relayerDataDirFlag, relayerConfigPathFlag)
	cmd.MarkFlagsMutuallyExclusive(primeBlockfrostAPIKeyFlag, primeSocketPathFlag, primeOgmiosURLFlag)
	cmd.MarkFlagsMutuallyExclusive(cardanoBlockfrostURLFlag, cardanoSocketPathFlag, cardanoOgmiosURLFlag)
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

	primeStartingSlot, primeStartingHash, err := parseStartingBlock(p.primeStartingBlock)
	if err != nil {
		return nil, err
	}

	cardanoStartingSlot, cardanoStartingHash, err := parseStartingBlock(p.cardanoStartingBlock)
	if err != nil {
		return nil, err
	}

	var (
		nativeTokensPrime   []sendtx.TokenExchangeConfig
		nativeTokensCardano []sendtx.TokenExchangeConfig
	)

	if p.primeCardanoWrappedTokenName != "" {
		nativeTokensPrime = []sendtx.TokenExchangeConfig{
			{
				DstChainID: common.ChainIDStrCardano,
				TokenName:  p.primeCardanoWrappedTokenName,
			},
		}
	}

	if p.cardanoPrimeWrappedTokenName != "" {
		nativeTokensCardano = []sendtx.TokenExchangeConfig{
			{
				DstChainID: common.ChainIDStrPrime,
				TokenName:  p.cardanoPrimeWrappedTokenName,
			},
		}
	}

	vcConfig := &vcCore.AppConfig{
		RunMode:             common.SkylineMode,
		RefundEnabled:       true,
		ValidatorDataDir:    cleanPath(p.validatorDataDir),
		ValidatorConfigPath: cleanPath(p.validatorConfig),
		CardanoChains: map[string]*oCore.CardanoChainConfig{
			common.ChainIDStrPrime: {
				CardanoChainConfig: cardanotx.CardanoChainConfig{
					NetworkMagic:          p.primeNetworkMagic,
					NetworkID:             wallet.CardanoNetworkType(p.primeNetworkID),
					TTLSlotNumberInc:      p.primeTTLSlotInc,
					OgmiosURL:             p.primeOgmiosURL,
					BlockfrostURL:         p.primeBlockfrostURL,
					BlockfrostAPIKey:      p.primeBlockfrostAPIKey,
					SocketPath:            p.primeSocketPath,
					PotentialFee:          300000,
					SlotRoundingThreshold: p.primeSlotRoundingThreshold,
					NoBatchPeriodPercent:  defaultNoBatchPeriodPercent,
					UtxoMinAmount:         p.primeUtxoMinAmount,
					MaxFeeUtxoCount:       defaultMaxFeeUtxoCount,
					MaxUtxoCount:          defaultMaxUtxoCount,
					TakeAtLeastUtxoCount:  defaultTakeAtLeastUtxoCount,
					NativeTokens:          nativeTokensPrime,
				},
				NetworkAddress:           p.primeNetworkAddress,
				StartBlockHash:           primeStartingHash,
				StartSlot:                primeStartingSlot,
				ConfirmationBlockCount:   p.primeBlockConfirmationCount,
				OtherAddressesOfInterest: []string{},
				MinFeeForBridging:        p.primeMinFeeForBridging,
				MinOperationFee:          p.primeMinOperationFee,
			},
			common.ChainIDStrCardano: {
				CardanoChainConfig: cardanotx.CardanoChainConfig{
					NetworkMagic:          p.cardanoNetworkMagic,
					NetworkID:             wallet.CardanoNetworkType(p.cardanoNetworkID),
					TTLSlotNumberInc:      p.cardanoTTLSlotInc,
					OgmiosURL:             p.cardanoOgmiosURL,
					BlockfrostURL:         p.cardanoBlockfrostURL,
					BlockfrostAPIKey:      p.cardanoBlockfrostAPIKey,
					SocketPath:            p.cardanoSocketPath,
					PotentialFee:          300000,
					SlotRoundingThreshold: p.cardanoSlotRoundingThreshold,
					NoBatchPeriodPercent:  defaultNoBatchPeriodPercent,
					UtxoMinAmount:         p.cardanoUtxoMinAmount,
					MaxFeeUtxoCount:       defaultMaxFeeUtxoCount,
					MaxUtxoCount:          defaultMaxUtxoCount,
					TakeAtLeastUtxoCount:  defaultTakeAtLeastUtxoCount,
					NativeTokens:          nativeTokensCardano,
				},
				NetworkAddress:           p.cardanoNetworkAddress,
				StartBlockHash:           cardanoStartingHash,
				StartSlot:                cardanoStartingSlot,
				ConfirmationBlockCount:   p.cardanoBlockConfirmationCount,
				OtherAddressesOfInterest: []string{},
				MinFeeForBridging:        p.cardanoMinFeeForBridging,
				MinOperationFee:          p.cardanoMinOperationFee,
			},
		},
		Bridge: oCore.BridgeConfig{
			NodeURL:              p.bridgeNodeURL,
			DynamicTx:            false,
			SmartContractAddress: p.bridgeSCAddress,
			SubmitConfig: oCore.SubmitConfig{
				ConfirmedBlocksThreshold:  20,
				ConfirmedBlocksSubmitTime: 3000,
				EmptyBlocksThreshold:      200,
			},
		},
		BridgingSettings: oCore.BridgingSettings{
			MaxAmountAllowedToBridge:       defaultMaxAmountAllowedToBridge,
			MaxReceiversPerBridgingRequest: 4, // 4 + 1 for fee
			MaxBridgingClaimsToGroup:       5,
		},
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

	primeChainSpecificJSONRaw, _ := json.Marshal(vcConfig.CardanoChains[common.ChainIDStrPrime].CardanoChainConfig)
	cardanoChainSpecificJSONRaw, _ := json.Marshal(vcConfig.CardanoChains[common.ChainIDStrCardano].CardanoChainConfig)

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
			common.ChainIDStrCardano: {
				ChainType:     common.ChainTypeCardanoStr,
				DbsPath:       filepath.Join(p.dbsPath, "relayer"),
				ChainSpecific: cardanoChainSpecificJSONRaw,
			},
		},
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
