package clideploysolana

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	solsendtx "github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	solanawallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
)

const (
	rpcURLFlag        = "url"
	feePayerKeyFlag   = "fee-payer"
	programKeyFlag    = "key"
	buildPathFlag     = "build-path"
	commitmentFlag    = "commitment"
	defaultCommitment = "finalized"

	adminKeyPathFlag               = "admin-key"
	bridgeNodeURLFlag              = "bridge-url"
	bridgeSCAddrFlag               = "bridge-addr"
	chainIDsConfigFlag             = "chain-ids-config"
	lastIDFlag                     = "last-id"
	minOperationFeeAmountFlag      = "min-operation-fee"
	minFeeForBridgingFlag          = "min-fee-for-bridging"
	minAmountToBridgeFlag          = "min-amount-to-bridge"
	treasuryAddressFlag            = "treasury-address"
	confirmationTimeoutSecondsFlag = "confirmation-timeout-seconds"

	rpcURLFlagDesc                     = "Solana RPC URL"
	feePayerKeyFlagDesc                = "path to fee payer keypair file"
	programKeyFlagDesc                 = "path to program keypair file"
	buildPathFlagDesc                  = "path to the compiled program (.so file)"
	commitmentFlagDesc                 = "commitment level (processed, confirmed, finalized)"
	adminKeyPathFlagDesc               = "path to Solana admin keypair file"
	bridgeNodeURLFlagDesc              = "bridge node url"
	bridgeSCAddrFlagDesc               = "bridge smart contract address"
	chainIDsConfigFlagDesc             = "path to the chain IDs config file"
	lastIDFlagDesc                     = "initial batch last id"
	minOperationFeeAmountFlagDesc      = "minimal operation fee amount (lamports)"
	minFeeForBridgingFlagDesc          = "minimal fee for bridging (lamports)"
	minAmountToBridgeFlagDesc          = "minimal token amount to bridge (lamports)"
	treasuryAddressFlagDesc            = "treasury wallet address"
	confirmationTimeoutSecondsFlagDesc = "max wait time in seconds for initialize tx finalization"

	defaultConfirmationTimeoutSeconds = uint64(120)
)

type deployProgramParams struct {
	rpcURL                     string
	feePayerKeyPath            string
	programKeyPath             string
	programID                  string
	buildPath                  string
	commitment                 string
	adminKeyPath               string
	bridgeNodeURL              string
	bridgeSCAddr               string
	chainIDsConfig             string
	lastID                     uint64
	minOperationFeeAmount      uint64
	minFeeForBridging          uint64
	minAmountToBridge          uint64
	treasuryAddress            string
	confirmationTimeoutSeconds uint64

	adminPrivateKey     solana.PrivateKey
	treasuryPublicKey   solana.PublicKey
	confirmationTimeout time.Duration
	programPublicKey    solana.PublicKey
	chainIDConverter    *common.ChainIDConverter
}

func (p *deployProgramParams) validateFlags() error {
	if !common.IsValidHTTPURL(p.rpcURL) {
		return fmt.Errorf("invalid --%s flag (must be a valid http or https URL)", rpcURLFlag)
	}

	if p.feePayerKeyPath == "" {
		return fmt.Errorf("fee payer key path not specified: --%s", feePayerKeyFlag)
	}

	if _, err := os.Stat(p.feePayerKeyPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("fee payer key file does not exist: %s", p.feePayerKeyPath)
		}

		return fmt.Errorf("failed to check fee payer key file: %w", err)
	}

	if p.programKeyPath == "" {
		return fmt.Errorf("program key path not specified: --%s", programKeyFlag)
	}

	p.programKeyPath = filepath.Clean(p.programKeyPath)
	if _, err := os.Stat(p.programKeyPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("program key file does not exist: %s", p.programKeyPath)
		}

		return fmt.Errorf("failed to check program key file: %w", err)
	}

	if p.buildPath == "" {
		return fmt.Errorf("build path not specified: --%s", buildPathFlag)
	}

	if _, err := os.Stat(p.buildPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("build path does not exist: %s", p.buildPath)
		}

		return fmt.Errorf("failed to check build path: %w", err)
	}

	if p.commitment == "" {
		p.commitment = defaultCommitment
	}

	if p.adminKeyPath == "" {
		return fmt.Errorf("admin key path not specified: --%s", adminKeyPathFlag)
	}

	p.adminKeyPath = filepath.Clean(p.adminKeyPath)
	if _, err := os.Stat(p.adminKeyPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("admin key file does not exist: %s", p.adminKeyPath)
		}

		return fmt.Errorf("failed to check admin key file: %w", err)
	}

	adminPrivateKey, err := solana.PrivateKeyFromSolanaKeygenFile(p.adminKeyPath)
	if err != nil {
		return fmt.Errorf("failed to load admin keypair file: %w", err)
	}

	p.adminPrivateKey = adminPrivateKey

	if !common.IsValidHTTPURL(p.bridgeNodeURL) {
		return fmt.Errorf("invalid --%s flag", bridgeNodeURLFlag)
	}

	if !ethcommon.IsHexAddress(p.bridgeSCAddr) {
		return fmt.Errorf("invalid --%s flag", bridgeSCAddrFlag)
	}

	if p.chainIDsConfig == "" {
		return fmt.Errorf("--%s flag not specified", chainIDsConfigFlag)
	}

	if _, err := os.Stat(p.chainIDsConfig); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config file does not exist: %s", p.chainIDsConfig)
		}

		return fmt.Errorf("failed to check config file: %s. err: %w", p.chainIDsConfig, err)
	}

	chainIDsConfig, err := common.LoadConfig[common.ChainIDsConfigFile](p.chainIDsConfig, "")
	if err != nil {
		return fmt.Errorf("failed to load chain IDs config: %w", err)
	}

	p.chainIDConverter = chainIDsConfig.ToChainIDConverter()

	if p.programID == "" {
		programPrivateKey, err := solana.PrivateKeyFromSolanaKeygenFile(p.programKeyPath)
		if err != nil {
			return fmt.Errorf("failed to load program keypair file: %w", err)
		}

		p.programPublicKey = programPrivateKey.PublicKey()
	} else {
		programPublicKey, err := solanawallet.PublicKeyFromAddress(p.programID)
		if err != nil {
			return fmt.Errorf("invalid program id: %w", err)
		}

		p.programPublicKey = programPublicKey
	}

	if p.treasuryAddress == "" {
		return fmt.Errorf("treasury address not specified: --%s", treasuryAddressFlag)
	}

	treasuryPublicKey, err := solanawallet.PublicKeyFromAddress(p.treasuryAddress)
	if err != nil {
		return fmt.Errorf("invalid treasury address: %w", err)
	}

	p.treasuryPublicKey = treasuryPublicKey

	if p.confirmationTimeoutSeconds == 0 {
		return fmt.Errorf("confirmation timeout must be greater than 0: --%s", confirmationTimeoutSecondsFlag)
	}

	p.confirmationTimeout = time.Duration(p.confirmationTimeoutSeconds) * time.Second //nolint:gosec

	return nil
}

func (p *deployProgramParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&p.rpcURL,
		rpcURLFlag,
		"",
		rpcURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.feePayerKeyPath,
		feePayerKeyFlag,
		"",
		feePayerKeyFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.programKeyPath,
		programKeyFlag,
		"",
		programKeyFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.buildPath,
		buildPathFlag,
		"",
		buildPathFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.commitment,
		commitmentFlag,
		defaultCommitment,
		commitmentFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.programID,
		programIDFlag,
		"",
		programIDFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.adminKeyPath,
		adminKeyPathFlag,
		"",
		adminKeyPathFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.bridgeNodeURL,
		bridgeNodeURLFlag,
		"",
		bridgeNodeURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.bridgeSCAddr,
		bridgeSCAddrFlag,
		"",
		bridgeSCAddrFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.chainIDsConfig,
		chainIDsConfigFlag,
		"",
		chainIDsConfigFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.lastID,
		lastIDFlag,
		0,
		lastIDFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.minOperationFeeAmount,
		minOperationFeeAmountFlag,
		0,
		minOperationFeeAmountFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.minFeeForBridging,
		minFeeForBridgingFlag,
		0,
		minFeeForBridgingFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.minAmountToBridge,
		minAmountToBridgeFlag,
		0,
		minAmountToBridgeFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.treasuryAddress,
		treasuryAddressFlag,
		"",
		treasuryAddressFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.confirmationTimeoutSeconds,
		confirmationTimeoutSecondsFlag,
		defaultConfirmationTimeoutSeconds,
		confirmationTimeoutSecondsFlagDesc,
	)
}

func (p *deployProgramParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	deployOutput, err := p.deployProgram(outputter)
	if err != nil {
		return nil, err
	}

	initSig, err := p.initializeProgram(outputter)
	if err != nil {
		return nil, err
	}

	return &deployProgramResult{
		Output:      deployOutput + "\ninitialize tx finalized",
		TxSignature: initSig,
	}, nil
}

func (p *deployProgramParams) deployProgram(outputter common.OutputFormatter) (string, error) {
	buildPath := filepath.Clean(p.buildPath)

	_, _ = outputter.Write([]byte("Deploying Solana program..."))
	outputter.WriteOutput()

	args := []string{
		"program", "deploy",
		"--url", p.rpcURL,
		"--fee-payer", p.feePayerKeyPath,
		"-k", p.programKeyPath,
		"--commitment", p.commitment,
		buildPath,
	}

	if p.programID != "" {
		args = append(args, "--program-id", p.programID)
	}

	output, err := common.ExecuteCLICommand("solana", args, ".")
	if err != nil {
		return "", fmt.Errorf("solana program deploy failed: %w", err)
	}

	return output, nil
}

func (p *deployProgramParams) initializeProgram(outputter common.OutputFormatter) (string, error) {
	ctx := context.Background()

	txHelperBridge, err := p.getTxHelperBridge()
	if err != nil {
		return "", err
	}

	validators, err := p.getValidatorPubkeys(ctx, txHelperBridge, outputter)
	if err != nil {
		return "", err
	}

	provider, err := solanawallet.NewProvider(p.rpcURL, nil)
	if err != nil {
		return "", fmt.Errorf("create provider: %w", err)
	}

	recentBlockhash, err := provider.GetLatestBlockhash(ctx)
	if err != nil {
		return "", fmt.Errorf("get latest blockhash: %w", err)
	}

	txSender := solsendtx.NewTxSender(provider, &solsendtx.ChainConfig{
		MinOperationFeeAmount: p.minOperationFeeAmount,
		MinFeeForBridging:     p.minFeeForBridging,
		MinAmountToBridge:     p.minAmountToBridge,
		TreasuryAddress:       p.treasuryPublicKey,
	})

	txDto := solsendtx.InitializeDto{
		ProgramID:     p.programPublicKey,
		AuthorityAddr: p.adminPrivateKey.PublicKey().String(),
		Validators:    validators,
		LastID:        p.lastID,
	}

	_, _ = outputter.Write([]byte("Submitting Solana initialize-program transaction..."))
	outputter.WriteOutput()

	tx, err := txSender.CreateTx(
		ctx,
		p.adminPrivateKey.PublicKey(),
		solsendtx.InstructionTypeInitialize,
		recentBlockhash,
		txDto,
	)
	if err != nil {
		return "", fmt.Errorf("create initialize tx: %w", err)
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		return &p.adminPrivateKey
	})
	if err != nil {
		return "", fmt.Errorf("sign initialize tx: %w", err)
	}

	sig, err := txSender.SendTx(ctx, tx)
	if err != nil {
		return "", fmt.Errorf("send initialize tx: %w", err)
	}

	if err := provider.WaitForSignature(ctx, *sig, rpc.CommitmentFinalized, p.confirmationTimeout); err != nil {
		return "", fmt.Errorf("wait for initialize confirmation: %w", err)
	}

	return sig.String(), nil
}

func (p *deployProgramParams) getTxHelperBridge() (*eth.EthHelperWrapper, error) {
	return eth.NewEthHelperWrapper(
		hclog.NewNullLogger(),
		ethtxhelper.WithNodeURL(p.bridgeNodeURL),
		ethtxhelper.WithDynamicTx(false)), nil
}

func (p *deployProgramParams) getValidatorPubkeys(
	ctx context.Context,
	txHelper *eth.EthHelperWrapper,
	outputter common.OutputFormatter,
) ([]string, error) {
	_, _ = outputter.Write(
		fmt.Appendf(nil, "Get validators data from bridge smart contract at %s...", p.bridgeSCAddr))

	outputter.WriteOutput()

	bridgeSC := eth.NewBridgeSmartContract(
		p.bridgeSCAddr,
		txHelper,
		p.chainIDConverter,
	)

	validatorsData, err := bridgeSC.GetValidatorsChainData(ctx, common.ChainIDStrSolana)
	if err != nil {
		return nil, fmt.Errorf("get validators chain data: %w", err)
	}

	if len(validatorsData) == 0 {
		return nil, fmt.Errorf("no validators found for chain %s", common.ChainIDStrSolana)
	}

	return solanaValidatorPubkeysFromChainData(validatorsData)
}

func solanaValidatorPubkeysFromChainData(validatorsData []eth.ValidatorChainData) ([]string, error) {
	validators := make([]string, len(validatorsData))

	for i, validatorData := range validatorsData {
		rawPubKey := make([]byte, solana.PublicKeyLength)
		validatorData.Key[0].FillBytes(rawPubKey)

		pubKey, err := solanawallet.PublicKeyFromBytes(rawPubKey)
		if err != nil {
			return nil, fmt.Errorf("convert validator key %d to public key: %w", i, err)
		}

		validators[i] = pubKey.String()
	}

	return validators, nil
}
