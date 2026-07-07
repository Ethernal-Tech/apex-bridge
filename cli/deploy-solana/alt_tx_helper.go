package clideploysolana

import (
	"context"
	"fmt"
	"time"

	solsendtx "github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	solanawallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func sendALTInstructions(
	ctx context.Context,
	provider *solanawallet.Provider,
	txSender *solsendtx.TxSender,
	adminPrivateKey solana.PrivateKey,
	ixs []solana.Instruction,
	confirmationTimeout time.Duration,
) (*solana.Signature, error) {
	recentBlockhash, err := provider.GetLatestBlockhash(ctx)
	if err != nil {
		return nil, fmt.Errorf("get latest blockhash: %w", err)
	}

	builder := solana.NewTransactionBuilder().SetRecentBlockHash(recentBlockhash).SetFeePayer(adminPrivateKey.PublicKey())

	for _, ix := range ixs {
		builder = builder.AddInstruction(ix)
	}

	tx, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("build transaction: %w", err)
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		return &adminPrivateKey
	})
	if err != nil {
		return nil, fmt.Errorf("sign transaction: %w", err)
	}

	sig, err := txSender.SendTx(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("send transaction: %w", err)
	}

	if err := provider.WaitForSignature(ctx, *sig, rpc.CommitmentFinalized, confirmationTimeout); err != nil {
		return nil, fmt.Errorf("wait for confirmation: %w", err)
	}

	return sig, nil
}
