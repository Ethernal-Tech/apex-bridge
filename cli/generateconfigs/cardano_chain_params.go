package cligenerateconfigs

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	rCore "github.com/Ethernal-Tech/apex-bridge/relayer/core"
	vcCore "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/spf13/cobra"
)

const (
	chainIDStringFlag          = "chain-id"
	networkAddressFlag         = "network-address"
	networkMagicFlag           = "network-magic"
	networkIDFlag              = "network-id"
	ogmiosURLFlag              = "ogmios-url"
	blockfrostURLFlag          = "blockfrost-url"
	blockfrostAPIKeyFlag       = "blockfrost-api-key"
	socketPathFlag             = "socket-path"
	ttlSlotIncFlag             = "ttl-slot-inc"
	slotRoundingThresholdFlag  = "slot-rounding-threshold"
	startingBlockFlag          = "starting-block"
	utxoMinAmountFlag          = "utxo-min-amount"
	minFeeForBridgingFlag      = "min-fee-for-bridging"
	blockConfirmationCountFlag = "block-confirmation-count"
	allowedDirectionsFlag      = "allowed-directions"

	chainIDStringFlagDesc          = "(mandatory) chain id string for the chain config"
	networkAddressFlagDesc         = "(mandatory) address of network"
	networkMagicFlagDesc           = "network magic (default 0)"
	networkIDFlagDesc              = "network id"
	ogmiosURLFlagDesc              = "ogmios URL chain network"
	blockfrostURLFlagDesc          = "blockfrost URL for chain network"
	blockfrostAPIKeyFlagDesc       = "blockfrost API key for chain network" //nolint:gosec
	socketPathFlagDesc             = "socket path for chain network"
	ttlSlotIncFlagDesc             = "TTL slot increment"
	slotRoundingThresholdFlagDesc  = "defines the upper limit used for rounding slot values for the chain. Any slot value between 0 and `slotRoundingThreshold` will be rounded to `slotRoundingThreshold` etc" //nolint:lll
	startingBlockFlagDesc          = "slot: hash of the block from where to start oracle / block submitter for the chain"                                                                                       //nolint:lll
	utxoMinAmountFlagDesc          = "minimal UTXO value for the chain"
	minFeeForBridgingFlagDesc      = "minimal bridging fee for the chain"
	blockConfirmationCountFlagDesc = "block confirmation count for the chain"
	allowedDirectionsFlagDesc      = "allowed bridging directions for the chain"

	defaultBlockConfirmationCount = 10
	defaultTTLSlotNumberInc       = 1800 + defaultBlockConfirmationCount*10 // BlockTimeSeconds
	defaultSlotRoundingThreshold  = 60

	defaultNoBatchPeriodPercent = 0.0625
)

type cardanoChainGenerateConfigsParams struct {
	chainIDString string

	networkAddress         string
	networkMagic           uint32
	networkID              uint32
	ogmiosURL              string
	blockfrostURL          string
	blockfrostAPIKey       string
	socketPath             string
	ttlSlotInc             uint64
	slotRoundingThreshold  uint64
	startingBlock          string
	utxoMinAmount          uint64
	minFeeForBridging      uint64
	blockConfirmationCount uint
	allowedDirections      []string

	dbsPath string

	outputDir                         string
	outputValidatorComponentsFileName string
	outputRelayerFileName             string

	emptyBlocksThreshold uint
}

func (p *cardanoChainGenerateConfigsParams) validateFlags() error {
	if p.chainIDString == "" {
		return fmt.Errorf("missing %s", chainIDStringFlag)
	}

	if !common.IsValidNetworkAddress(p.networkAddress) {
		return fmt.Errorf("invalid %s: %s", networkAddressFlag, p.networkAddress)
	}

	if p.blockfrostURL == "" && p.socketPath == "" && p.ogmiosURL == "" {
		return fmt.Errorf("specify at least one of: %s, %s, %s",
			blockfrostURLFlag, socketPathFlag, ogmiosURLFlag)
	}

	if p.blockfrostURL != "" && !common.IsValidHTTPURL(p.blockfrostURL) {
		return fmt.Errorf("invalid %s blockfrost url: %s", p.chainIDString, p.blockfrostURL)
	}

	if p.ogmiosURL != "" && !common.IsValidHTTPURL(p.ogmiosURL) {
		return fmt.Errorf("invalid %s ogmios url: %s", p.chainIDString, p.ogmiosURL)
	}

	if p.startingBlock != "" {
		parts := strings.Split(p.startingBlock, ":")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("invalid %s starting block: %s", p.chainIDString, p.startingBlock)
		}
	}

	if p.minFeeForBridging < p.utxoMinAmount {
		return fmt.Errorf("%s minimal fee for bridging: %d should't be less than minimal UTXO amount: %d",
			p.chainIDString, p.minFeeForBridging, p.utxoMinAmount)
	}

	return nil
}

func (p *cardanoChainGenerateConfigsParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&p.chainIDString,
		chainIDStringFlag,
		"",
		chainIDStringFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.networkAddress,
		networkAddressFlag,
		"",
		networkAddressFlagDesc,
	)
	cmd.Flags().Uint32Var(
		&p.networkMagic,
		networkMagicFlag,
		defaultNetworkMagic,
		networkMagicFlagDesc,
	)
	cmd.Flags().Uint32Var(
		&p.networkID,
		networkIDFlag,
		uint32(wallet.MainNetNetwork),
		networkIDFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.ogmiosURL,
		ogmiosURLFlag,
		"",
		ogmiosURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.blockfrostURL,
		blockfrostURLFlag,
		"",
		blockfrostURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.blockfrostAPIKey,
		blockfrostAPIKeyFlag,
		"",
		blockfrostAPIKeyFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.socketPath,
		socketPathFlag,
		"",
		socketPathFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.ttlSlotInc,
		ttlSlotIncFlag,
		defaultTTLSlotNumberInc,
		ttlSlotIncFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.slotRoundingThreshold,
		slotRoundingThresholdFlag,
		defaultSlotRoundingThreshold,
		slotRoundingThresholdFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.startingBlock,
		startingBlockFlag,
		"",
		startingBlockFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.utxoMinAmount,
		utxoMinAmountFlag,
		common.MinUtxoAmountDefault,
		utxoMinAmountFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.minFeeForBridging,
		minFeeForBridgingFlag,
		common.MinFeeForBridgingDefault,
		minFeeForBridgingFlagDesc,
	)
	cmd.Flags().UintVar(
		&p.blockConfirmationCount,
		blockConfirmationCountFlag,
		defaultBlockConfirmationCount,
		blockConfirmationCountFlagDesc,
	)
	cmd.Flags().UintVar(
		&p.emptyBlocksThreshold,
		emptyBlocksThresholdFlag,
		defaultEmptyBlocksThreshold,
		emptyBlocksThresholdFlagDesc,
	)
	cmd.Flags().StringSliceVar(
		&p.allowedDirections,
		allowedDirectionsFlag,
		nil,
		allowedDirectionsFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.dbsPath,
		dbsPathFlag,
		defaultDBsPath,
		dbsPathFlagDesc,
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

	cmd.MarkFlagsMutuallyExclusive(blockfrostURLFlag, socketPathFlag, ogmiosURLFlag)
}

func (p *cardanoChainGenerateConfigsParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	outputDirPath := filepath.Clean(p.outputDir)
	if err := common.CreateDirectoryIfNotExists(outputDirPath, 0770); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	vcConfigPath := filepath.Join(outputDirPath, p.outputValidatorComponentsFileName)

	vcConfig, err := common.LoadJSON[vcCore.AppConfig](vcConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load validator components config json: %w", err)
	}

	startingSlot, startingHash, err := parseStartingBlock(p.startingBlock)
	if err != nil {
		return nil, err
	}

	if vcConfig.CardanoChains == nil {
		vcConfig.CardanoChains = make(map[string]*oCore.CardanoChainConfig)
	}

	vcConfig.CardanoChains[p.chainIDString] = &oCore.CardanoChainConfig{
		CardanoChainConfig: cardanotx.CardanoChainConfig{
			NetworkMagic:          p.networkMagic,
			NetworkID:             wallet.CardanoNetworkType(p.networkID),
			TTLSlotNumberInc:      p.ttlSlotInc,
			OgmiosURL:             p.ogmiosURL,
			BlockfrostURL:         p.blockfrostURL,
			BlockfrostAPIKey:      p.blockfrostAPIKey,
			SocketPath:            p.socketPath,
			PotentialFee:          300000,
			SlotRoundingThreshold: p.slotRoundingThreshold,
			NoBatchPeriodPercent:  defaultNoBatchPeriodPercent,
			UtxoMinAmount:         p.utxoMinAmount,
			MaxFeeUtxoCount:       defaultMaxFeeUtxoCount,
			MaxUtxoCount:          defaultMaxUtxoCount,
			TakeAtLeastUtxoCount:  defaultTakeAtLeastUtxoCount,
		},
		NetworkAddress:           p.networkAddress,
		StartBlockHash:           startingHash,
		StartSlot:                startingSlot,
		ConfirmationBlockCount:   p.blockConfirmationCount,
		OtherAddressesOfInterest: []string{},
		FeeAddrBridgingAmount:    p.utxoMinAmount,
		MinFeeForBridging:        p.minFeeForBridging,
	}

	if vcConfig.Bridge.SubmitConfig.EmptyBlocksThreshold == nil {
		vcConfig.Bridge.SubmitConfig.EmptyBlocksThreshold = make(map[string]uint)
	}

	vcConfig.Bridge.SubmitConfig.EmptyBlocksThreshold[p.chainIDString] = p.emptyBlocksThreshold

	if vcConfig.BridgingSettings.AllowedDirections == nil {
		vcConfig.BridgingSettings.AllowedDirections = make(map[string][]string)
	}

	vcConfig.BridgingSettings.AllowedDirections[p.chainIDString] = p.allowedDirections

	if err := common.SaveJSON(vcConfigPath, vcConfig, true); err != nil {
		return nil, fmt.Errorf("failed to create validator components config json: %w", err)
	}

	rConfigPath := filepath.Join(outputDirPath, p.outputRelayerFileName)

	rConfig, err := common.LoadJSON[rCore.RelayerManagerConfiguration](rConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load relayer components config json: %w", err)
	}

	chainSpecificJSONRaw, err := json.Marshal(vcConfig.CardanoChains[p.chainIDString].CardanoChainConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chain specific config to json: %w", err)
	}

	if rConfig.Chains == nil {
		rConfig.Chains = make(map[string]rCore.ChainConfig)
	}

	rConfig.Chains[p.chainIDString] = rCore.ChainConfig{
		ChainType:     common.ChainTypeCardanoStr,
		DbsPath:       filepath.Join(p.dbsPath, "relayer"),
		ChainSpecific: chainSpecificJSONRaw,
	}

	if err := common.SaveJSON(rConfigPath, rConfig, true); err != nil {
		return nil, fmt.Errorf("failed to create relayer config json: %w", err)
	}

	return &CmdResult{
		validatorComponentsConfigPath: vcConfigPath,
		relayerConfigPath:             rConfigPath,
	}, nil
}
