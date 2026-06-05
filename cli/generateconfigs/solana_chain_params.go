package cligenerateconfigs

import (
	"encoding/json"
	"fmt"
	"math/big"
	"path/filepath"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	rCore "github.com/Ethernal-Tech/apex-bridge/relayer/core"
	solanatx "github.com/Ethernal-Tech/apex-bridge/solana"
	vcCore "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	wallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/spf13/cobra"
)

const (
	solanaChainNodeURLFlag        = "sol-node-url"
	solanaChainTrackedProgramFlag = "sol-tracked-program"
	solanaBlockFetchDelayFlag     = "sol-block-fetch-delay"
	solanaMinFeeForBridgingFlag   = "sol-min-fee-for-bridging"
	solanaMinOperationFeeFlag     = "sol-min-operation-fee"
	solanaTTLNumberIncFlag        = "sol-ttl-number-inc"
	solanaConfirmationTimeoutFlag = "sol-confirmation-timeout"
	solanaTrackerStartBlockFlag   = "sol-tracker-start-block"

	solanaChainNodeURLFlagDesc        = "solana chain node URL"
	solanaChainTrackedProgramFlagDesc = "(mandatory) solana program address to track"
	solanaBlockFetchDelayFlagDesc     = "delay in milliseconds between block fetches for solana chain"
	solanaMinFeeForBridgingFlagDesc   = "minimal bridging fee for solana chain"
	solanaMinOperationFeeFlagDesc     = "minimal operation fee for solana chain"
	solanaTTLNumberIncFlagDesc        = "TTL increment for solana chain"
	solanaConfirmationTimeoutFlagDesc = "confirmation timeout for solana chain txs in milliseconds"
	solanaTrackerStartBlockFlagDesc   = "block to start solana chain tracker from in a form of slot:blockNumber (default 0)"

	defaultSolanaRetryIntervalMiliseconds   = 400 * time.Millisecond
	defaultSolanaBlockFetchDelay            = uint64(250)
	defaultSolanaMinFeeForBridging          = uint64(1_000_010)
	defaultSolanaMinOperationFee            = uint64(0)
	defaultSolanaMinColCoinsAllowedToBridge = uint64(1)
	defaultSolanaSlotBuffSize               = uint8(20)
	defaultSolanaEventBuffSize              = uint8(100)
	defaultSolanaErrorBuffSize              = uint8(10)
	defaultSolanaTTLSlotNumberInc           = uint64(0)
	defaultSolanaConfirmationTimeout        = int64(60000) // 1 minute

	altPublicKeyFlag     = "alt-public-key"
	altPublicKeyFlagDesc = "alt public key for solana chain"

	defaultSlotRoundingThresholdSolana = 10
	defaultNoBatchPeriodPercentSolana  = 0
)

type solanaChainGenerateConfigsParams struct {
	chainIDString string

	solanaChainNodeURL      string
	solanaTrackedProgram    string
	solanaBlockFetchDelay   uint64
	solanaMinFeeForBridging uint64
	solanaMinOperationFee   uint64
	solanaSlotBuffSize      uint8
	solanaEventBuffSize     uint8
	solanaErrorBuffSize     uint8
	solanaTrackerStartBlock string

	solanaTrackerStartSlot     uint64
	solanaTrackerStartBlockNum uint64

	emptyBlocksThreshold      uint
	solanaTTLNumberInc        uint64
	solanaConfirmationTimeout int64

	outputDir                         string
	outputValidatorComponentsFileName string
	outputRelayerFileName             string

	dbsPath string

	relayerDataDir    string
	relayerConfigPath string
	treasuryAddress   string
	altPublicKey      string

	slotRoundingThreshold uint64
}

func (p *solanaChainGenerateConfigsParams) validateFlags() error {
	if p.chainIDString == "" {
		return fmt.Errorf("missing %s", chainIDStringFlag)
	}

	if !common.IsValidHTTPURL(p.solanaChainNodeURL) {
		return fmt.Errorf("invalid %s: %s", solanaChainNodeURLFlag, p.solanaChainNodeURL)
	}

	if p.solanaTrackedProgram == "" {
		return fmt.Errorf("missing %s", solanaChainTrackedProgramFlag)
	}

	if _, err := wallet.PublicKeyFromAddress(p.treasuryAddress); err != nil {
		return fmt.Errorf("invalid %s: %s", treasuryAddressFlag, p.treasuryAddress)
	}

	var startSlot, startBlock uint64
	if p.solanaTrackerStartBlock != "" {
		_, err := fmt.Sscanf(p.solanaTrackerStartBlock, "%d:%d", &startSlot, &startBlock)
		if err != nil {
			return fmt.Errorf("invalid format for %s, expected slot:blockNumber: %w", solanaTrackerStartBlockFlag, err)
		}
	}

	p.solanaTrackerStartSlot = startSlot
	p.solanaTrackerStartBlockNum = startBlock

	return nil
}

func (p *solanaChainGenerateConfigsParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&p.chainIDString,
		chainIDStringFlag,
		"",
		chainIDStringFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.solanaChainNodeURL,
		solanaChainNodeURLFlag,
		"",
		solanaChainNodeURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.solanaTrackedProgram,
		solanaChainTrackedProgramFlag,
		"",
		solanaChainTrackedProgramFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.solanaBlockFetchDelay,
		solanaBlockFetchDelayFlag,
		defaultSolanaBlockFetchDelay,
		solanaBlockFetchDelayFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.solanaMinFeeForBridging,
		solanaMinFeeForBridgingFlag,
		defaultSolanaMinFeeForBridging,
		solanaMinFeeForBridgingFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.solanaMinOperationFee,
		solanaMinOperationFeeFlag,
		defaultSolanaMinOperationFee,
		solanaMinOperationFeeFlagDesc,
	)

	cmd.Flags().UintVar(
		&p.emptyBlocksThreshold,
		emptyBlocksThresholdFlag,
		defaultEmptyBlocksThreshold,
		emptyBlocksThresholdFlagDesc,
	)

	cmd.Flags().Uint64Var(
		&p.solanaTTLNumberInc,
		solanaTTLNumberIncFlag,
		defaultSolanaTTLSlotNumberInc,
		solanaTTLNumberIncFlagDesc,
	)

	cmd.Flags().Int64Var(
		&p.solanaConfirmationTimeout,
		solanaConfirmationTimeoutFlag,
		defaultSolanaConfirmationTimeout,
		solanaConfirmationTimeoutFlagDesc,
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

	cmd.Flags().Uint64Var(
		&p.slotRoundingThreshold,
		slotRoundingThresholdFlag,
		defaultSlotRoundingThresholdSolana,
		slotRoundingThresholdFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.solanaTrackerStartBlock,
		solanaTrackerStartBlockFlag,
		"",
		solanaTrackerStartBlockFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.outputRelayerFileName,
		outputRelayerFileNameFlag,
		defaultOutputRelayerFileName,
		outputRelayerFileNameFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.dbsPath,
		dbsPathFlag,
		defaultDBsPath,
		dbsPathFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.treasuryAddress,
		treasuryAddressFlag,
		"",
		treasuryAddressFlagDesc,
	)

	cmd.Flags().StringVar(
		&p.altPublicKey,
		altPublicKeyFlag,
		"",
		altPublicKeyFlagDesc,
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

	cmd.MarkFlagsMutuallyExclusive(relayerDataDirFlag, relayerConfigPathFlag)
}

func (p *solanaChainGenerateConfigsParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	outputDirPath := filepath.Clean(p.outputDir)
	if err := common.CreateDirectoryIfNotExists(outputDirPath, 0770); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	vcConfigPath := filepath.Join(outputDirPath, p.outputValidatorComponentsFileName)

	vcConfig, err := common.LoadJSON[vcCore.AppConfig](vcConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load validator components config json: %w", err)
	}

	if vcConfig.SolanaChains == nil {
		vcConfig.SolanaChains = make(map[string]*oCore.SolanaChainConfig)
	}

	vcConfig.SolanaChains[p.chainIDString] = &oCore.SolanaChainConfig{
		SolanaChainConfig: solanatx.SolanaChainConfig{
			ConfirmationTimeout:   time.Duration(p.solanaConfirmationTimeout),
			TTLNumberInc:          p.solanaTTLNumberInc,
			MinFeeForBridging:     new(big.Int).SetUint64(p.solanaMinFeeForBridging),
			TxProviderEndpoint:    p.solanaChainNodeURL,
			SlotRoundingThreshold: p.slotRoundingThreshold,
			NoBatchPeriodPercent:  defaultNoBatchPeriodPercentSolana,
			ALTPublicKey:          p.altPublicKey,
			ProgramID:             p.solanaTrackedProgram,
		},
		TrackerStartSlot:           p.solanaTrackerStartSlot,
		TrackerStartBlockNumber:    p.solanaTrackerStartBlockNum,
		TrackedProgram:             p.solanaTrackedProgram,
		BlockFetchDelayMiliseconds: time.Duration(p.solanaBlockFetchDelay), //nolint:gosec
		RetryTimeoutMiliseconds:    defaultSolanaRetryIntervalMiliseconds,
		RestartTrackerPullCheck:    time.Second * 150,
		FeeAddrBridgingAmount:      p.solanaMinFeeForBridging,
		MinColCoinsAllowedToBridge: defaultSolanaMinColCoinsAllowedToBridge,
		MinOperationFee:            p.solanaMinOperationFee,
		TreasuryAddress:            p.treasuryAddress,
	}

	if vcConfig.Bridge.SubmitConfig.EmptyBlocksThreshold == nil {
		vcConfig.Bridge.SubmitConfig.EmptyBlocksThreshold = make(map[string]uint)
	}

	vcConfig.Bridge.SubmitConfig.EmptyBlocksThreshold[p.chainIDString] = p.emptyBlocksThreshold

	if err := common.SaveJSON(vcConfigPath, vcConfig, true); err != nil {
		return nil, fmt.Errorf("failed to update validator components config json: %w", err)
	}

	rConfigPath := filepath.Join(outputDirPath, p.outputRelayerFileName)

	rConfig, err := common.LoadJSON[rCore.RelayerManagerConfiguration](rConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load relayer config json: %w", err)
	}

	chainSpecificJSONRaw, err := json.Marshal(vcConfig.SolanaChains[p.chainIDString].SolanaChainConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chain specific config to json: %w", err)
	}

	if rConfig.Chains == nil {
		rConfig.Chains = make(map[string]rCore.ChainConfig)
	}

	rConfig.Chains[p.chainIDString] = rCore.ChainConfig{
		ChainID:           p.chainIDString,
		ChainType:         common.ChainTypeSolanaStr,
		DbsPath:           filepath.Join(p.dbsPath, "relayer"),
		ChainSpecific:     chainSpecificJSONRaw,
		RelayerDataDir:    cleanPath(p.relayerDataDir),
		RelayerConfigPath: cleanPath(p.relayerConfigPath),
	}

	if err := common.SaveJSON(rConfigPath, rConfig, true); err != nil {
		return nil, fmt.Errorf("failed to update relayer config json: %w", err)
	}

	return &CmdResult{
		validatorComponentsConfigPath: vcConfigPath,
		relayerConfigPath:             rConfigPath,
	}, nil
}
