package clideploysolana

import (
	"context"
	"fmt"
	"time"

	solanawallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// signUpgradeAuthorityChange builds the loader-v3 SetAuthorityChecked instruction.
// The program-data PDA is writable and both authorities must sign.
func signUpgradeAuthorityChange(
	program solana.PublicKey, blockhash solana.Hash, current, next, payer solana.PrivateKey,
) (*solana.Transaction, error) {
	loader := solana.BPFLoaderUpgradeableProgramID

	programData, _, err := solana.FindProgramAddress([][]byte{program.Bytes()}, loader)
	if err != nil {
		return nil, fmt.Errorf("derive program-data address: %w", err)
	}

	const setAuthorityChecked = 7 // bincode u32 enum discriminant in solana-loader-v3-interface

	ix := solana.NewInstruction(loader, solana.AccountMetaSlice{
		solana.Meta(programData).WRITE(),
		solana.Meta(current.PublicKey()).SIGNER(),
		solana.Meta(next.PublicKey()).SIGNER(),
	}, []byte{setAuthorityChecked, 0, 0, 0})

	tx, err := solana.NewTransaction([]solana.Instruction{ix}, blockhash, solana.TransactionPayer(payer.PublicKey()))
	if err != nil {
		return nil, err
	}

	keys := map[solana.PublicKey]*solana.PrivateKey{
		current.PublicKey(): &current,
		next.PublicKey():    &next,
		payer.PublicKey():   &payer,
	}

	if _, err := tx.Sign(func(key solana.PublicKey) *solana.PrivateKey { return keys[key] }); err != nil {
		return nil, err
	}

	return tx, nil
}

func (p *setUpgradeAuthorityParams) submitUpgradeAuthorityChange() (string, error) {
	provider, err := solanawallet.NewProvider(p.rpcURL, nil)
	if err != nil {
		return "", err
	}

	program, err := solana.PublicKeyFromBase58(p.programID)
	if err != nil {
		return "", err
	}

	payer := p.currentPrivateKey
	if p.feePayerKeyPath != "" {
		payer, err = solana.PrivateKeyFromSolanaKeygenFile(p.feePayerKeyPath)
		if err != nil {
			return "", fmt.Errorf("read fee payer keypair: %w", err)
		}
	}

	const timeout = 2 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	blockhash, err := provider.GetLatestBlockhash(ctx)
	if err != nil {
		return "", fmt.Errorf("get latest blockhash: %w", err)
	}

	tx, err := signUpgradeAuthorityChange(program, blockhash, p.currentPrivateKey, p.newPrivateKey, payer)
	if err != nil {
		return "", fmt.Errorf("sign upgrade authority change: %w", err)
	}

	sig, err := provider.SendTransaction(ctx, tx)
	if err != nil {
		return "", fmt.Errorf("submit upgrade authority change %s: %w", tx.Signatures[0], err)
	}

	if err := provider.WaitForSignature(ctx, sig, rpc.CommitmentType(p.commitment), timeout); err != nil {
		return "", fmt.Errorf("confirm upgrade authority change %s: %w", sig, err)
	}

	return sig.String(), nil
}
