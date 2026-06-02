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

const (
	registerTokenTreasuryAddressFlag = "treasury-address"

	registerTokenTreasuryAddressFlagDesc = "treasury wallet address"
)

type registerTokenCommonParams struct {
	altCommonParams

	treasuryAddress string

	treasuryPublicKey solana.PublicKey
}

func (p *registerTokenCommonParams) setRegisterTokenCommonFlags(cmd commonFlagSetter) {
	p.setCommonFlags(cmd)

	cmd.StringVar(
		&p.treasuryAddress,
		registerTokenTreasuryAddressFlag,
		"",
		registerTokenTreasuryAddressFlagDesc,
	)
}

func (p *registerTokenCommonParams) validateRegisterTokenCommonFlags() error {
	if err := p.validateCommonFlags(); err != nil {
		return err
	}

	if p.treasuryAddress == "" {
		return fmt.Errorf("treasury address not specified: --%s", registerTokenTreasuryAddressFlag)
	}

	treasuryPublicKey, err := solanawallet.PublicKeyFromAddress(p.treasuryAddress)
	if err != nil {
		return fmt.Errorf("invalid treasury address: %w", err)
	}

	p.treasuryPublicKey = treasuryPublicKey

	return nil
}

func sendSignedTransaction(
	ctx context.Context,
	provider *solanawallet.Provider,
	txSender *solsendtx.TxSender,
	tx *solana.Transaction,
	confirmationTimeout time.Duration,
	signers ...solana.PrivateKey,
) (*solana.Signature, error) {
	signerKeys := make(map[solana.PublicKey]*solana.PrivateKey, len(signers))

	for _, signer := range signers {
		signerCopy := signer
		signerKeys[signerCopy.PublicKey()] = &signerCopy
	}

	_, err := tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if pk, ok := signerKeys[key]; ok {
			return pk
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("sign instruction: %w", err)
	}

	sig, err := txSender.SendTx(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("send tx: %w", err)
	}

	if err := provider.WaitForSignature(ctx, *sig, rpc.CommitmentFinalized, confirmationTimeout); err != nil {
		return nil, fmt.Errorf("wait for confirmation: %w", err)
	}

	return sig, nil
}
