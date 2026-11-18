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

const (
	setAuthorityNewAuthorityFlag     = "new-authority"
	setAuthorityNewAuthorityFlagDesc = "expected base58 public key of the new bridge admin"
	setAuthorityNewAuthorityKeyFlag  = "new-authority-key"
)

// setAuthorityParams hands the bridge admin role (FeeConfig.authority and ProgramConfig.authority)
// over to a new key.
type setAuthorityParams struct {
	programAdminParams

	newAuthority        string
	newAuthorityKeyPath string
	newAuthorityKey     solana.PrivateKey
}

func (p *setAuthorityParams) setFlags(cmd *cobra.Command) {
	p.setCommonFlagsOnCommand(cmd)

	cmd.Flags().StringVar(
		&p.newAuthority, setAuthorityNewAuthorityFlag, "", setAuthorityNewAuthorityFlagDesc)
	cmd.Flags().StringVar(&p.newAuthorityKeyPath, setAuthorityNewAuthorityKeyFlag, "",
		"path to the new bridge admin keypair that co-signs the handover")
}

func (p *setAuthorityParams) validateFlags() error {
	if p.adminKeyPath == "" && p.privateKeyConfig == "" {
		return fmt.Errorf("specify one of --%s or --%s", programAdminAdminKeyPathFlag, privateKeyConfigFlag)
	}

	if err := p.validateCommonFlags(); err != nil {
		return err
	}

	if p.newAuthority == "" {
		return fmt.Errorf("new authority not specified: --%s", setAuthorityNewAuthorityFlag)
	}

	newKey, err := solanawallet.PublicKeyFromAddress(p.newAuthority)
	if err != nil {
		return fmt.Errorf("invalid --%s: %w", setAuthorityNewAuthorityFlag, err)
	}

	if newKey == (solana.PublicKey{}) {
		return fmt.Errorf("--%s must not be the default (all zero) public key", setAuthorityNewAuthorityFlag)
	}

	if !newKey.IsOnCurve() {
		return fmt.Errorf("--%s not on ed25519 curve: no private key can sign for it", setAuthorityNewAuthorityFlag)
	}

	if newKey == p.adminPrivateKey.PublicKey() {
		return fmt.Errorf("new authority equals current admin %s", newKey)
	}

	if p.newAuthorityKeyPath == "" {
		return fmt.Errorf("new authority keypair not specified: --%s", setAuthorityNewAuthorityKeyFlag)
	}

	signer, err := solana.PrivateKeyFromSolanaKeygenFile(p.newAuthorityKeyPath)
	if err != nil {
		return fmt.Errorf("invalid --%s: %w", setAuthorityNewAuthorityKeyFlag, err)
	}

	if signer.PublicKey() != newKey {
		return fmt.Errorf("--%s does not match --%s", setAuthorityNewAuthorityKeyFlag, setAuthorityNewAuthorityFlag)
	}

	p.newAuthorityKey = signer

	return nil
}

// checkCurrentAuthority refuses a handover that cannot succeed or would change nothing. Only
// ProgramConfig is read; a FeeConfig mismatch fails atomically on-chain at has_one.
func checkCurrentAuthority(cfgAuthority, admin, newAuthority solana.PublicKey) error {
	if cfgAuthority != admin {
		return fmt.Errorf("admin key %s does not match program config authority %s", admin, cfgAuthority)
	}

	if newAuthority == cfgAuthority {
		return fmt.Errorf("new authority %s is already the program config authority", newAuthority)
	}

	return nil
}

func buildSetAuthorityDto(programID, admin, newAuthority solana.PublicKey) solsendtx.SetAuthorityDto {
	return solsendtx.SetAuthorityDto{
		ProgramID:        programID,
		AuthorityAddr:    admin.String(),
		NewAuthorityAddr: newAuthority.String(),
	}
}

// checkAuthorityChanged verifies the read-back authority after the transaction finalized.
func checkAuthorityChanged(sig string, got, want solana.PublicKey) error {
	if got != want {
		return fmt.Errorf("tx %s finalized but program config authority is %s, expected %s", sig, got, want)
	}

	return nil
}

func (p *setAuthorityParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	provider, err := solanawallet.NewProvider(p.rpcURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}

	return p.rotateAuthority(context.Background(), provider, outputter)
}

func (p *setAuthorityParams) rotateAuthority(
	ctx context.Context, provider solanawallet.ITxProvider, outputter common.OutputFormatter,
) (common.ICommandResult, error) {
	admin := p.adminPrivateKey.PublicKey()
	newAuthority := p.newAuthorityKey.PublicKey()
	txSender := solsendtx.NewTxSender(provider, nil)

	programConfig, err := txSender.GetProgramConfig(ctx, p.programPublicKey)
	if err != nil {
		return nil, fmt.Errorf("get program config: %w", err)
	}

	if err := checkCurrentAuthority(programConfig.Authority, admin, newAuthority); err != nil {
		return nil, err
	}

	recentBlockhash, err := provider.GetLatestBlockhash(ctx)
	if err != nil {
		return nil, fmt.Errorf("get latest blockhash: %w", err)
	}

	tx, err := txSender.CreateTx(
		ctx,
		admin,
		solsendtx.InstructionTypeSetAuthority,
		recentBlockhash,
		buildSetAuthorityDto(p.programPublicKey, admin, newAuthority),
	)
	if err != nil {
		return nil, fmt.Errorf("create set authority tx: %w", err)
	}

	_, _ = outputter.Write([]byte("Submitting set-authority transaction..."))
	outputter.WriteOutput()

	sig, err := sendSignedTransaction(
		ctx, provider, txSender, tx, p.confirmationTimeout, p.adminPrivateKey, p.newAuthorityKey)
	if err != nil {
		return nil, fmt.Errorf("send set authority tx: %w", err)
	}

	updated, err := txSender.GetProgramConfig(ctx, p.programPublicKey)
	if err != nil {
		return nil, fmt.Errorf("tx %s finalized but reading back program config failed: %w", sig, err)
	}

	if err := checkAuthorityChanged(sig.String(), updated.Authority, newAuthority); err != nil {
		return nil, err
	}

	return &deployProgramResult{
		Output:      fmt.Sprintf("bridge admin changed from %s to %s", admin, newAuthority),
		TxSignature: sig.String(),
	}, nil
}
