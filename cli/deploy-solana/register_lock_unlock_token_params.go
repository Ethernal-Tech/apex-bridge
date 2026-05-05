package clideploysolana

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Ethernal-Tech/apex-bridge/common"
	solsendtx "github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	solanawallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/spf13/cobra"
)

const (
	registerLockUnlockTokenIDFlag                = "token-id"
	registerLockUnlockTokenMintFlag              = "token-mint"
	registerLockUnlockTokenMinBridgingAmountFlag = "min-bridging-amount" //nolint:gosec

	registerLockUnlockTokenIDFlagDesc                = "token ID"
	registerLockUnlockTokenMintFlagDesc              = "token mint public key" //nolint:gosec
	registerLockUnlockTokenMinBridgingAmountFlagDesc = "minimal token amount to bridge (lamports)"
)

type registerLockUnlockTokenParams struct {
	registerTokenCommonParams

	tokenID           uint16
	tokenMint         string
	minBridgingAmount uint64
}

func (p *registerLockUnlockTokenParams) setFlags(cmd *cobra.Command) {
	p.setRegisterTokenCommonFlags(cmd.Flags())

	cmd.Flags().Uint16Var(
		&p.tokenID,
		registerLockUnlockTokenIDFlag,
		0,
		registerLockUnlockTokenIDFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.tokenMint,
		registerLockUnlockTokenMintFlag,
		"",
		registerLockUnlockTokenMintFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.minBridgingAmount,
		registerLockUnlockTokenMinBridgingAmountFlag,
		0,
		registerLockUnlockTokenMinBridgingAmountFlagDesc,
	)
}

func (p *registerLockUnlockTokenParams) validateFlags() error {
	if err := p.validateRegisterTokenCommonFlags(); err != nil {
		return err
	}

	if p.tokenMint == "" {
		return fmt.Errorf("token mint not specified: --%s", registerLockUnlockTokenMintFlag)
	}

	if _, err := solanawallet.PublicKeyFromAddress(p.tokenMint); err != nil {
		return fmt.Errorf("invalid token mint address: %w", err)
	}

	return nil
}

func (p *registerLockUnlockTokenParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
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
		TreasuryAddress:    p.treasuryPublicKey,
		BridgingFeeAddress: p.relayerPublicKey,
	})

	txDto := solsendtx.RegisterTokenLockUnlockDto{
		ProgramID:         p.programPublicKey,
		AuthorityAddr:     p.adminPrivateKey.PublicKey().String(),
		TokenMint:         p.tokenMint,
		TokenID:           p.tokenID,
		MinBridgingAmount: p.minBridgingAmount,
	}

	tx, err := txSender.CreateTx(
		ctx,
		p.adminPrivateKey.PublicKey(),
		solsendtx.InstructionTypeRegisterTokensLockUnlock,
		recentBlockhash,
		txDto,
	)
	if err != nil {
		return nil, fmt.Errorf("create register token lock unlock tx: %w", err)
	}

	_, _ = outputter.Write([]byte("Submitting register-lock-unlock-token transaction..."))
	outputter.WriteOutput()

	sig, err := sendSignedTransaction(
		ctx,
		provider,
		txSender,
		tx,
		p.confirmationTimeout,
		p.adminPrivateKey,
	)
	if err != nil {
		return nil, fmt.Errorf("send register token lock unlock tx: %w", err)
	}

	return &deployProgramResult{
		Output:      "registered lock/unlock token id " + strconv.Itoa(int(p.tokenID)),
		MintAddress: p.tokenMint,
		TxSignature: sig.String(),
	}, nil
}
