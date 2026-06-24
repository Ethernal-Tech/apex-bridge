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
	updateMinBridgingAmountTokenIDFlag           = "token-id"
	updateMinBridgingAmountMinBridgingAmountFlag = "min-bridging-amount"

	updateMinBridgingAmountTokenIDFlagDesc           = "token ID"
	updateMinBridgingAmountMinBridgingAmountFlagDesc = "minimal token amount to bridge (lamports)"
)

type updateMinBridgingAmountParams struct {
	altCommonParams

	tokenID           uint16
	minBridgingAmount uint64
}

func (p *updateMinBridgingAmountParams) setFlags(cmd *cobra.Command) {
	p.setCommonFlags(cmd.Flags())

	cmd.Flags().Uint16Var(
		&p.tokenID,
		updateMinBridgingAmountTokenIDFlag,
		0,
		updateMinBridgingAmountTokenIDFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.minBridgingAmount,
		updateMinBridgingAmountMinBridgingAmountFlag,
		0,
		updateMinBridgingAmountMinBridgingAmountFlagDesc,
	)
}

func (p *updateMinBridgingAmountParams) validateFlags() error {
	if err := p.validateCommonFlags(); err != nil {
		return err
	}

	if p.tokenID == 0 {
		return fmt.Errorf("token ID must be greater than 0: --%s", updateMinBridgingAmountTokenIDFlag)
	}

	if p.minBridgingAmount == 0 {
		return fmt.Errorf("min bridging amount must be greater than 0: --%s", updateMinBridgingAmountMinBridgingAmountFlag)
	}

	return nil
}

func (p *updateMinBridgingAmountParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()

	provider, err := solanawallet.NewProvider(p.rpcURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}

	recentBlockhash, err := provider.GetLatestBlockhash(ctx)
	if err != nil {
		return nil, fmt.Errorf("get latest blockhash: %w", err)
	}

	txSender := solsendtx.NewTxSender(provider, nil)

	txDto := solsendtx.UpdateMinBridgingAmountDto{
		ProgramID:         p.programPublicKey,
		AuthorityAddr:     p.adminPrivateKey.PublicKey().String(),
		TokenID:           p.tokenID,
		MinBridgingAmount: p.minBridgingAmount,
	}

	tx, err := txSender.CreateTx(
		ctx,
		p.adminPrivateKey.PublicKey(),
		solsendtx.InstructionTypeUpdateMinBridgingAmount,
		recentBlockhash,
		txDto,
	)
	if err != nil {
		return nil, fmt.Errorf("create update min bridging amount tx: %w", err)
	}

	_, _ = outputter.Write([]byte("Submitting update-min-bridging-amount transaction..."))
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
		return nil, fmt.Errorf("send update min bridging amount tx: %w", err)
	}

	return &deployProgramResult{
		Output: "update min bridging amount finalized for token id " +
			strconv.Itoa(int(p.tokenID)) +
			" with amount " + strconv.FormatUint(p.minBridgingAmount, 10),
		TxSignature: sig.String(),
	}, nil
}
