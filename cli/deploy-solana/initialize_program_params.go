package clideploysolana

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	solsendtx "github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	solanawallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/spf13/cobra"
)

const (
	initializeProgramRPCURLFlag                     = "url"
	initializeProgramAdminKeyPathFlag               = "admin-key"
	initializeProgramValidatorFlag                  = "validator"
	initializeProgramLastIDFlag                     = "last-id"
	initializeProgramMinOperationFeeAmountFlag      = "min-operation-fee"
	initializeProgramMinFeeForBridgingFlag          = "min-fee-for-bridging"
	initializeProgramMinAmountToBridgeFlag          = "min-amount-to-bridge"
	initializeProgramTreasuryAddressFlag            = "treasury-address"
	initializeProgramRelayerAddressFlag             = "relayer-address"
	initializeProgramConfirmationTimeoutSecondsFlag = "confirmation-timeout-seconds"

	initializeProgramRPCURLFlagDesc                     = "Solana RPC URL"
	initializeProgramAdminKeyPathFlagDesc               = "path to Solana admin keypair file"
	initializeProgramValidatorFlagDesc                  = "validator public key (repeat flag for multiple validators)"
	initializeProgramLastIDFlagDesc                     = "initial batch last id"
	initializeProgramMinOperationFeeAmountFlagDesc      = "minimal operation fee amount (lamports)"
	initializeProgramMinFeeForBridgingFlagDesc          = "minimal fee for bridging (lamports)"
	initializeProgramMinAmountToBridgeFlagDesc          = "minimal token amount to bridge (lamports)"
	initializeProgramTreasuryAddressFlagDesc            = "treasury wallet address"
	initializeProgramRelayerAddressFlagDesc             = "relayer wallet address used as bridging fee receiver"
	initializeProgramConfirmationTimeoutSecondsFlagDesc = "max wait time in seconds for tx finalization"

	defaultInitializeProgramConfirmationTimeoutSeconds = uint64(120)
)

type initializeProgramParams struct {
	rpcURL                     string
	adminKeyPath               string
	validators                 []string
	lastID                     uint64
	minOperationFeeAmount      uint64
	minFeeForBridging          uint64
	minAmountToBridge          uint64
	treasuryAddress            string
	relayerAddress             string
	confirmationTimeoutSeconds uint64

	adminPrivateKey     solana.PrivateKey
	treasuryPublicKey   solana.PublicKey
	relayerPublicKey    solana.PublicKey
	confirmationTimeout time.Duration
}

func (p *initializeProgramParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&p.rpcURL,
		initializeProgramRPCURLFlag,
		"",
		initializeProgramRPCURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.adminKeyPath,
		initializeProgramAdminKeyPathFlag,
		"",
		initializeProgramAdminKeyPathFlagDesc,
	)
	cmd.Flags().StringArrayVar(
		&p.validators,
		initializeProgramValidatorFlag,
		nil,
		initializeProgramValidatorFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.lastID,
		initializeProgramLastIDFlag,
		0,
		initializeProgramLastIDFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.minOperationFeeAmount,
		initializeProgramMinOperationFeeAmountFlag,
		0,
		initializeProgramMinOperationFeeAmountFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.minFeeForBridging,
		initializeProgramMinFeeForBridgingFlag,
		0,
		initializeProgramMinFeeForBridgingFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.minAmountToBridge,
		initializeProgramMinAmountToBridgeFlag,
		0,
		initializeProgramMinAmountToBridgeFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.treasuryAddress,
		initializeProgramTreasuryAddressFlag,
		"",
		initializeProgramTreasuryAddressFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.relayerAddress,
		initializeProgramRelayerAddressFlag,
		"",
		initializeProgramRelayerAddressFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.confirmationTimeoutSeconds,
		initializeProgramConfirmationTimeoutSecondsFlag,
		defaultInitializeProgramConfirmationTimeoutSeconds,
		initializeProgramConfirmationTimeoutSecondsFlagDesc,
	)
}

func (p *initializeProgramParams) validateFlags() error {
	if !common.IsValidHTTPURL(p.rpcURL) {
		return fmt.Errorf(
			"invalid --%s flag (must be a valid http or https URL)",
			initializeProgramRPCURLFlag,
		)
	}

	if p.adminKeyPath == "" {
		return fmt.Errorf("admin key path not specified: --%s", initializeProgramAdminKeyPathFlag)
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

	if len(p.validators) == 0 {
		return fmt.Errorf("validators are not specified: --%s", initializeProgramValidatorFlag)
	}

	validators := make([]string, len(p.validators))
	for i, validatorAddr := range p.validators {
		validatorAddr = strings.TrimSpace(validatorAddr)
		if validatorAddr == "" {
			return fmt.Errorf("validator address %d is empty", i)
		}

		if _, err := solanawallet.PublicKeyFromAddress(validatorAddr); err != nil {
			return fmt.Errorf("invalid validator address %s: %w", validatorAddr, err)
		}

		validators[i] = validatorAddr
	}

	p.validators = validators

	if p.treasuryAddress == "" {
		return fmt.Errorf("treasury address not specified: --%s", initializeProgramTreasuryAddressFlag)
	}

	treasuryPublicKey, err := solanawallet.PublicKeyFromAddress(p.treasuryAddress)
	if err != nil {
		return fmt.Errorf("invalid treasury address: %w", err)
	}

	p.treasuryPublicKey = treasuryPublicKey

	if p.relayerAddress == "" {
		return fmt.Errorf("relayer address not specified: --%s", initializeProgramRelayerAddressFlag)
	}

	relayerPublicKey, err := solanawallet.PublicKeyFromAddress(p.relayerAddress)
	if err != nil {
		return fmt.Errorf("invalid relayer address: %w", err)
	}

	p.relayerPublicKey = relayerPublicKey

	if p.confirmationTimeoutSeconds == 0 {
		return fmt.Errorf("confirmation timeout must be greater than 0: --%s",
			initializeProgramConfirmationTimeoutSecondsFlag)
	}

	p.confirmationTimeout = time.Duration(p.confirmationTimeoutSeconds) * time.Second //nolint:gosec

	return nil
}

func (p *initializeProgramParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()

	provider, err := solanawallet.NewProvider(p.rpcURL)
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}

	recentBlockhash, err := provider.GetLatestBlockhash(ctx)
	if err != nil {
		return nil, fmt.Errorf("get latest blockhash: %w", err)
	}

	txSender := solsendtx.NewTxSender(provider, &solsendtx.ChainConfig{
		MinOperationFeeAmount: p.minOperationFeeAmount,
		MinFeeForBridging:     p.minFeeForBridging,
		MinAmountToBridge:     p.minAmountToBridge,
		TreasuryAddress:       p.treasuryPublicKey,
		BridgingFeeAddress:    p.relayerPublicKey,
	})

	txDto := solsendtx.InitializeDto{
		AuthorityAddr: p.adminPrivateKey.PublicKey().String(),
		Validators:    p.validators,
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
		return nil, fmt.Errorf("create initialize tx: %w", err)
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		return &p.adminPrivateKey
	})
	if err != nil {
		return nil, fmt.Errorf("sign initialize tx: %w", err)
	}

	sig, err := txSender.SendTx(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("send initialize tx: %w", err)
	}

	if err := provider.WaitForSignature(ctx, *sig, rpc.CommitmentFinalized, p.confirmationTimeout); err != nil {
		return nil, fmt.Errorf("wait for initialize confirmation: %w", err)
	}

	return &deployProgramResult{
		Output:      "initialize tx finalized",
		TxSignature: sig.String(),
	}, nil
}
