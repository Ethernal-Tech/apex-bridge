package cligenerateconfigs

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	vcCore "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/spf13/cobra"
)

const (
	solanaChainNodeURLFlag        = "sol-node-url"
	solanaChainTrackedProgramFlag = "sol-tracked-program"
	solanaBlockFetchDelayFlag     = "sol-block-fetch-delay"
	solanaMinFeeForBridgingFlag   = "sol-min-fee-for-bridging"
	solanaMinOperationFeeFlag     = "sol-min-operation-fee"
	solanaFeeAddrBridgingFlag     = "sol-fee-addr-bridging"
	solanaSlotBuffSizeFlag        = "sol-slot-buff-size"
	solanaEventBuffSizeFlag       = "sol-event-buff-size"
	solanaErrorBuffSizeFlag       = "sol-error-buff-size"

	solanaChainNodeURLFlagDesc        = "solana chain node URL"
	solanaChainTrackedProgramFlagDesc = "(mandatory) solana program address to track"
	solanaBlockFetchDelayFlagDesc     = "delay in milliseconds between block fetches for solana chain"
	solanaMinFeeForBridgingFlagDesc   = "minimal bridging fee for solana chain"
	solanaMinOperationFeeFlagDesc     = "minimal operation fee for solana chain"
	solanaFeeAddrBridgingFlagDesc     = "minimal addr fee bridging for solana chain"
	solanaSlotBuffSizeFlagDesc        = "slot buffer size for solana chain tracker"
	solanaEventBuffSizeFlagDesc       = "event buffer size for solana chain tracker"
	solanaErrorBuffSizeFlagDesc       = "error buffer size for solana chain tracker"

	defaultSolanaPoolIntervalMiliseconds    = time.Duration(1500)
	defaultSolanaBlockFetchDelay            = uint64(500)
	defaultSolanaMinFeeForBridging          = uint64(1_000_010)
	defaultSolanaMinOperationFee            = uint64(0)
	defaultSolanaFeeAddrBridging            = uint64(1_000_000)
	defaultSolanaMinColCoinsAllowedToBridge = uint64(1)
	defaultSolanaSlotBuffSize               = uint8(10)
	defaultSolanaEventBuffSize              = uint8(10)
	defaultSolanaErrorBuffSize              = uint8(10)
)

type solanaChainGenerateConfigsParams struct {
	chainIDString string

	solanaChainNodeURL      string
	solanaTrackedProgram    string
	solanaBlockFetchDelay   uint64
	solanaMinFeeForBridging uint64
	solanaMinOperationFee   uint64
	solanaFeeAddrBridging   uint64
	solanaSlotBuffSize      uint8
	solanaEventBuffSize     uint8
	solanaErrorBuffSize     uint8

	emptyBlocksThreshold uint

	outputDir                         string
	outputValidatorComponentsFileName string

	dbsPath string
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
	cmd.Flags().Uint64Var(
		&p.solanaFeeAddrBridging,
		solanaFeeAddrBridgingFlag,
		defaultSolanaFeeAddrBridging,
		solanaFeeAddrBridgingFlagDesc,
	)
	cmd.Flags().Uint8Var(
		&p.solanaSlotBuffSize,
		solanaSlotBuffSizeFlag,
		defaultSolanaSlotBuffSize,
		solanaSlotBuffSizeFlagDesc,
	)
	cmd.Flags().Uint8Var(
		&p.solanaEventBuffSize,
		solanaEventBuffSizeFlag,
		defaultSolanaEventBuffSize,
		solanaEventBuffSizeFlagDesc,
	)
	cmd.Flags().Uint8Var(
		&p.solanaErrorBuffSize,
		solanaErrorBuffSizeFlag,
		defaultSolanaErrorBuffSize,
		solanaErrorBuffSizeFlagDesc,
	)

	cmd.Flags().UintVar(
		&p.emptyBlocksThreshold,
		emptyBlocksThresholdFlag,
		defaultEmptyBlocksThreshold,
		emptyBlocksThresholdFlagDesc,
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
		&p.dbsPath,
		dbsPathFlag,
		defaultDBsPath,
		dbsPathFlagDesc,
	)
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
		NodeURL:                    p.solanaChainNodeURL,
		TrackedProgram:             p.solanaTrackedProgram,
		BlockFetchDelayMiliseconds: time.Duration(p.solanaBlockFetchDelay), //nolint:gosec
		PoolIntervalMiliseconds:    defaultSolanaPoolIntervalMiliseconds,
		RestartTrackerPullCheck:    time.Second * 150,
		FeeAddrBridgingAmount:      p.solanaFeeAddrBridging,
		MinColCoinsAllowedToBridge: defaultSolanaMinColCoinsAllowedToBridge,
		MinOperationFee:            p.solanaMinOperationFee,
		MinFeeForBridging:          p.solanaMinFeeForBridging,
		SlotBuffSize:               p.solanaSlotBuffSize,
		EventBuffSize:              p.solanaEventBuffSize,
		ErrorBuffSize:              p.solanaErrorBuffSize,
	}

	if vcConfig.Bridge.SubmitConfig.EmptyBlocksThreshold == nil {
		vcConfig.Bridge.SubmitConfig.EmptyBlocksThreshold = make(map[string]uint)
	}

	vcConfig.Bridge.SubmitConfig.EmptyBlocksThreshold[p.chainIDString] = p.emptyBlocksThreshold

	if err := common.SaveJSON(vcConfigPath, vcConfig, true); err != nil {
		return nil, fmt.Errorf("failed to update validator components config json: %w", err)
	}

	return &CmdResult{
		validatorComponentsConfigPath: vcConfigPath,
	}, nil
}
