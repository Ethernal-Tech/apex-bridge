package clideploysolana

import (
	"context"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	solsendtx "github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	solanawallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/spf13/cobra"
)

type createALTParams struct {
	altCommonParams
}

func (p *createALTParams) setFlags(cmd *cobra.Command) {
	p.setCommonFlags(cmd.Flags())
}

func (p *createALTParams) validateFlags() error {
	return p.validateCommonFlags()
}

func (p *createALTParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()

	provider, err := solanawallet.NewProvider(p.rpcURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}

	txSender := solsendtx.NewTxSender(provider, nil)
	altAdmin := solanawallet.NewALTAdmin(provider)

	slot, err := provider.GetFinalizedSlot(ctx)
	if err != nil {
		return nil, fmt.Errorf("get finalized slot: %w", err)
	}

	createIx, altPubKey, err := altAdmin.NewCreateInstruction(
		p.adminPrivateKey.PublicKey(),
		p.adminPrivateKey.PublicKey(),
		slot,
	)
	if err != nil {
		return nil, fmt.Errorf("new create instruction: %w", err)
	}

	_, _ = outputter.Write([]byte("Submitting create-alt transaction..."))
	outputter.WriteOutput()

	sig, err := sendALTInstructions(
		ctx,
		provider,
		txSender,
		p.adminPrivateKey,
		[]solana.Instruction{createIx},
		p.confirmationTimeout,
	)
	if err != nil {
		return nil, fmt.Errorf("send create ALT tx: %w", err)
	}

	return &deployProgramResult{
		Output:      "ALT created: " + altPubKey.String(),
		TxSignature: sig.String(),
	}, nil
}
