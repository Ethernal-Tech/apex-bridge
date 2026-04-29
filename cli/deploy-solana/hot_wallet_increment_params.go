package clideploysolana

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	solsendtx "github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	solanawallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/spf13/cobra"
)

const (
	hotWalletIncrementRPCURLFlag                     = "url"
	hotWalletIncrementKeyPathFlag                    = "key"
	hotWalletIncrementMintFlag                       = "mint"
	hotWalletIncrementAmountFlag                     = "amount"
	hotWalletIncrementProgramFlag                    = "program"
	hotWalletIncrementConfirmationTimeoutSecondsFlag = "confirmation-timeout-seconds"

	hotWalletIncrementRPCURLFlagDesc                     = "Solana RPC URL"
	hotWalletIncrementKeyPathFlagDesc                    = "path to Solana signer keypair file"
	hotWalletIncrementMintFlagDesc                       = "token mint public key"
	hotWalletIncrementAmountFlagDesc                     = "hot wallet increment amount in token base units (lamports for SOL)" //nolint:lll
	hotWalletIncrementProgramFlagDesc                    = "skyline program public key"
	hotWalletIncrementConfirmationTimeoutSecondsFlagDesc = "max wait time in seconds for tx finalization"

	defaultHotWalletIncrementConfirmationTimeoutSeconds = uint64(120)
)

type hotWalletIncrementParams struct {
	rpcURL                     string
	keyPath                    string
	mintAddress                string
	amount                     uint64
	programAddress             string
	confirmationTimeoutSeconds uint64

	senderPrivateKey    solana.PrivateKey
	confirmationTimeout time.Duration
	programPublicKey    solana.PublicKey
}

func (p *hotWalletIncrementParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&p.rpcURL,
		hotWalletIncrementRPCURLFlag,
		"",
		hotWalletIncrementRPCURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.keyPath,
		hotWalletIncrementKeyPathFlag,
		"",
		hotWalletIncrementKeyPathFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.mintAddress,
		hotWalletIncrementMintFlag,
		"",
		hotWalletIncrementMintFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.amount,
		hotWalletIncrementAmountFlag,
		0,
		hotWalletIncrementAmountFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.programAddress,
		hotWalletIncrementProgramFlag,
		"",
		hotWalletIncrementProgramFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.confirmationTimeoutSeconds,
		hotWalletIncrementConfirmationTimeoutSecondsFlag,
		defaultHotWalletIncrementConfirmationTimeoutSeconds,
		hotWalletIncrementConfirmationTimeoutSecondsFlagDesc,
	)
}

func (p *hotWalletIncrementParams) validateFlags() error {
	if !common.IsValidHTTPURL(p.rpcURL) {
		return fmt.Errorf(
			"invalid --%s flag (must be a valid http or https URL)",
			hotWalletIncrementRPCURLFlag,
		)
	}

	if p.keyPath == "" {
		return fmt.Errorf("key path not specified: --%s", hotWalletIncrementKeyPathFlag)
	}

	p.keyPath = filepath.Clean(p.keyPath)
	if _, err := os.Stat(p.keyPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("key file does not exist: %s", p.keyPath)
		}

		return fmt.Errorf("failed to check key file: %w", err)
	}

	senderPrivateKey, err := solana.PrivateKeyFromSolanaKeygenFile(p.keyPath)
	if err != nil {
		return fmt.Errorf("failed to load signer keypair file: %w", err)
	}

	p.senderPrivateKey = senderPrivateKey

	if p.mintAddress == "" {
		return fmt.Errorf("mint not specified: --%s", hotWalletIncrementMintFlag)
	}

	if _, err := solanawallet.PublicKeyFromAddress(p.mintAddress); err != nil {
		return fmt.Errorf("invalid mint address: %w", err)
	}

	if p.amount == 0 {
		return fmt.Errorf("amount must be greater than 0: --%s", hotWalletIncrementAmountFlag)
	}

	if p.programAddress == "" {
		return fmt.Errorf("program not specified: --%s", hotWalletIncrementProgramFlag)
	}

	programPublicKey, err := solanawallet.PublicKeyFromAddress(p.programAddress)
	if err != nil {
		return fmt.Errorf("invalid program address: %w", err)
	}

	p.programPublicKey = programPublicKey

	if p.confirmationTimeoutSeconds == 0 {
		return fmt.Errorf(
			"confirmation timeout must be greater than 0: --%s",
			hotWalletIncrementConfirmationTimeoutSecondsFlag,
		)
	}

	p.confirmationTimeout = time.Duration(p.confirmationTimeoutSeconds) * time.Second //nolint:gosec

	return nil
}

func (p *hotWalletIncrementParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()

	provider, err := solanawallet.NewProvider(p.rpcURL)
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}

	recentBlockhash, err := provider.GetLatestBlockhash(ctx)
	if err != nil {
		return nil, fmt.Errorf("get latest blockhash: %w", err)
	}

	txSender := solsendtx.NewTxSender(provider, nil)

	txDto := solsendtx.HotWalletIncrementDto{
		SenderAddr: p.senderPrivateKey.PublicKey().String(),
		TokenMint:  p.mintAddress,
		Amount:     p.amount,
	}

	tx, err := txSender.CreateTx(
		ctx,
		p.senderPrivateKey.PublicKey(),
		solsendtx.InstructionTypeHotWalletIncrement,
		recentBlockhash,
		txDto,
	)
	if err != nil {
		return nil, fmt.Errorf("create hot wallet increment tx: %w", err)
	}

	_, _ = outputter.Write([]byte("Submitting hot-wallet-increment transaction..."))
	outputter.WriteOutput()

	sig, err := sendSignedTransaction(
		ctx,
		provider,
		txSender,
		tx,
		p.confirmationTimeout,
		p.senderPrivateKey,
	)
	if err != nil {
		return nil, fmt.Errorf("send hot wallet increment tx: %w", err)
	}

	return &deployProgramResult{
		Output: "hot wallet increment finalized for mint " + p.mintAddress +
			" with amount " + strconv.FormatUint(p.amount, 10),
		TxSignature: sig.String(),
		MintAddress: p.mintAddress,
	}, nil
}
