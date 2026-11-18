package clideploysolana

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	solanatx "github.com/Ethernal-Tech/apex-bridge/solana"
	solanawallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/spf13/cobra"
)

const (
	newUpgradeAuthorityKeyFlag = "new-upgrade-authority-key"

	setUpgradeAuthorityKeyConfigFlagDesc = "secrets config for other_solana_admin (current) and other_new_solana_admin (incoming)" //nolint:lll
	newUpgradeAuthorityKeyFlagDesc       = "incoming keypair file (overrides other_new_solana_admin)"
	setUpgradeAuthorityFeePayerFlagDesc  = "path to fee payer keypair file (defaults to the current upgrade authority key)"
)

type setUpgradeAuthorityParams struct {
	rpcURL                     string
	programID                  string
	upgradeAuthorityKeyPath    string
	privateKeyConfig           string
	newUpgradeAuthorityKeyPath string
	feePayerKeyPath            string
	commitment                 string

	currentPrivateKey solana.PrivateKey
	newPrivateKey     solana.PrivateKey
}

func (p *setUpgradeAuthorityParams) validateFlags() error {
	if !common.IsValidHTTPURL(p.rpcURL) {
		return fmt.Errorf("invalid --%s flag (must be a valid http or https URL)", rpcURLFlag)
	}

	if p.programID == "" {
		return fmt.Errorf("program id not specified: --%s", programIDFlag)
	}

	if _, err := solanawallet.PublicKeyFromAddress(p.programID); err != nil {
		return fmt.Errorf("invalid --%s: %w", programIDFlag, err)
	}

	if err := validateCommitment(p.commitment); err != nil {
		return err
	}

	if (p.upgradeAuthorityKeyPath == "") == (p.privateKeyConfig == "") {
		return fmt.Errorf("specify exactly one of --%s or --%s", upgradeAuthorityKeyFlag, privateKeyConfigFlag)
	}

	var err error

	if p.upgradeAuthorityKeyPath != "" {
		p.upgradeAuthorityKeyPath, err = resolveExistingFile(p.upgradeAuthorityKeyPath, "upgrade authority key")
		if err != nil {
			return err
		}
	}

	p.currentPrivateKey, err = solanatx.GetSolanaPrivateKey(
		p.upgradeAuthorityKeyPath, p.privateKeyConfig, solanatx.GetKeyNameForSolanaAdmin())
	if err != nil {
		return fmt.Errorf("failed to load current upgrade authority key: %w", err)
	}

	if p.newUpgradeAuthorityKeyPath == "" && p.privateKeyConfig == "" {
		return fmt.Errorf("new upgrade authority key not specified: --%s", newUpgradeAuthorityKeyFlag)
	}

	newKeySource := solanatx.GetKeyNameForNewSolanaAdmin()
	if p.newUpgradeAuthorityKeyPath != "" {
		newKeySource = "--" + newUpgradeAuthorityKeyFlag
		p.newUpgradeAuthorityKeyPath, err = resolveExistingFile(p.newUpgradeAuthorityKeyPath, "new upgrade authority key")
		if err != nil {
			return fmt.Errorf("%s: %w", newKeySource, err)
		}
	}

	p.newPrivateKey, err = solanatx.GetSolanaPrivateKey(
		p.newUpgradeAuthorityKeyPath, p.privateKeyConfig, solanatx.GetKeyNameForNewSolanaAdmin())
	if err != nil {
		return fmt.Errorf("load incoming upgrade authority from %s: %w", newKeySource, err)
	}

	if p.newPrivateKey.PublicKey() == p.currentPrivateKey.PublicKey() {
		return fmt.Errorf("%s is the same as the current upgrade authority %s", newKeySource, p.currentPrivateKey.PublicKey())
	}

	if p.feePayerKeyPath != "" {
		p.feePayerKeyPath, err = resolveExistingFile(p.feePayerKeyPath, "fee payer key")
		if err != nil {
			return fmt.Errorf("--%s: %w", feePayerKeyFlag, err)
		}
	}

	return nil
}

func (p *setUpgradeAuthorityParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.rpcURL, rpcURLFlag, "", rpcURLFlagDesc)
	cmd.Flags().StringVar(&p.programID, programIDFlag, "", "public key (address) of the program")
	cmd.Flags().StringVar(
		&p.upgradeAuthorityKeyPath, upgradeAuthorityKeyFlag, "", "path to the current upgrade authority keypair file")
	cmd.Flags().StringVar(
		&p.privateKeyConfig, privateKeyConfigFlag, "", setUpgradeAuthorityKeyConfigFlagDesc)
	cmd.Flags().StringVar(
		&p.newUpgradeAuthorityKeyPath, newUpgradeAuthorityKeyFlag, "", newUpgradeAuthorityKeyFlagDesc)
	cmd.Flags().StringVar(&p.feePayerKeyPath, feePayerKeyFlag, "", setUpgradeAuthorityFeePayerFlagDesc)
	cmd.Flags().StringVar(&p.commitment, commitmentFlag, defaultCommitment, commitmentFlagDesc)

	cmd.MarkFlagsMutuallyExclusive(upgradeAuthorityKeyFlag, privateKeyConfigFlag)
}

// buildArgs builds the `solana program set-upgrade-authority` arguments. All paths must already be
// absolute. The fee payer (-k) defaults to the current authority; --final and the signer-check
// skip flag are never passed.
func (p *setUpgradeAuthorityParams) buildArgs(currentAuthorityPath string) []string {
	feePayer := p.feePayerKeyPath
	if feePayer == "" {
		feePayer = currentAuthorityPath
	}

	return []string{
		solanaCLIProgramArg, "set-upgrade-authority", p.programID,
		solanaCLIURLArg, p.rpcURL,
		"-k", feePayer,
		solanaCLIUpgradeAuthorityArg, currentAuthorityPath,
		"--new-upgrade-authority", p.newUpgradeAuthorityKeyPath,
		solanaCLICommitmentArg, p.commitment,
	}
}

func (p *setUpgradeAuthorityParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	_, _ = outputter.Write([]byte("Changing Solana program upgrade authority..."))
	outputter.WriteOutput()

	oldKey := p.currentPrivateKey.PublicKey()
	newKey := p.newPrivateKey.PublicKey()

	var output string
	var err error
	if p.privateKeyConfig != "" {
		output, err = p.submitUpgradeAuthorityChange()
	} else {
		output, err = common.ExecuteCLICommand("solana", p.buildArgs(p.upgradeAuthorityKeyPath), ".")
	}
	if err != nil {
		return nil, fmt.Errorf("solana program set-upgrade-authority failed: %w", err)
	}

	return &setUpgradeAuthorityResult{
		Output:       strings.TrimSpace(output),
		ProgramID:    p.programID,
		OldAuthority: oldKey.String(),
		NewAuthority: newKey.String(),
	}, nil
}

type setUpgradeAuthorityResult struct {
	Output       string
	ProgramID    string
	OldAuthority string
	NewAuthority string
}

func (r setUpgradeAuthorityResult) GetOutput() string {
	var buffer bytes.Buffer

	if r.Output != "" {
		buffer.WriteString(r.Output + "\n")
	}

	buffer.WriteString("\n[UPGRADE AUTHORITY CHANGED]\n")
	buffer.WriteString(common.FormatKV([]string{
		fmt.Sprintf("Program ID|%s", r.ProgramID),
		fmt.Sprintf("Old Authority|%s", r.OldAuthority),
		fmt.Sprintf("New Authority|%s", r.NewAuthority),
	}))
	buffer.WriteString("\n")

	return buffer.String()
}
