package clideploysolana

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	solanatx "github.com/Ethernal-Tech/apex-bridge/solana"
	solsendtx "github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	solanawallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/spf13/cobra"
)

const (
	programIDFlag     = "program-id"
	programIDFlagDesc = "public key (address) of the program account to upgrade"

	upgradeProgramVersionFlag     = "upgrade-program-version"
	upgradeProgramVersionFlagDesc = "semver written to on-chain program config after deploy (e.g. 1.2.3); must not be smaller than the current version" //nolint:lll

	upgradeProgramAdminKeyPathFlag     = "admin-key"
	upgradeProgramAdminKeyPathFlagDesc = "path to bridge admin keypair (signs update_program_version; must match program config authority)" //nolint:lll

	upgradeAuthorityKeyFlag     = "upgrade-authority-key"
	upgradeAuthorityKeyFlagDesc = "path to the program upgrade authority keypair (defaults to the admin key)"

	upgradeProgramConfirmationTimeoutSecondsFlag     = "confirmation-timeout-seconds"
	upgradeProgramConfirmationTimeoutSecondsFlagDesc = "max wait time in seconds for update-program-version tx finalization" //nolint:lll

	defaultUpgradeProgramConfirmationTimeoutSeconds = uint64(120)

	// solana CLI arguments shared by the commands that shell out to it
	solanaCLIProgramArg          = "program"
	solanaCLIURLArg              = "--url"
	solanaCLIUpgradeAuthorityArg = "--upgrade-authority"
	solanaCLICommitmentArg       = "--commitment"
)

type upgradeProgramParams struct {
	rpcURL                     string
	feePayerKeyPath            string
	upgradeAuthorityKeyPath    string
	programID                  string
	buildPath                  string
	commitment                 string
	upgradeProgramVersion      string
	adminKeyPath               string
	privateKeyConfig           string
	confirmationTimeoutSeconds uint64

	programPublicKey    solana.PublicKey
	adminPrivateKey     solana.PrivateKey
	confirmationTimeout time.Duration
}

// validateCommitment accepts only the commitment levels the solana CLI and RPC understand.
func validateCommitment(commitment string) error {
	switch commitment {
	case "processed", "confirmed", "finalized":
		return nil
	default:
		return fmt.Errorf("invalid --%s %q (must be processed, confirmed or finalized)", commitmentFlag, commitment)
	}
}

// resolveExistingFile makes path clean and absolute (so a value starting with "-" can never be
// read as a flag by the solana CLI) and checks that the file exists.
func resolveExistingFile(path, description string) (string, error) {
	abs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("failed to resolve %s path %s: %w", description, path, err)
	}

	if _, err := os.Stat(abs); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%s file does not exist: %s", description, abs)
		}

		return "", fmt.Errorf("failed to check %s file %s: %w", description, abs, err)
	}

	return abs, nil
}

func (p *upgradeProgramParams) validateFlags() error {
	if !common.IsValidHTTPURL(p.rpcURL) {
		return fmt.Errorf("invalid --%s flag (must be a valid http or https URL)", rpcURLFlag)
	}

	var err error

	if p.feePayerKeyPath != "" {
		if p.feePayerKeyPath, err = resolveExistingFile(p.feePayerKeyPath, "fee payer key"); err != nil {
			return err
		}
	}

	if p.upgradeAuthorityKeyPath != "" {
		p.upgradeAuthorityKeyPath, err = resolveExistingFile(p.upgradeAuthorityKeyPath, "upgrade authority key")
		if err != nil {
			return err
		}
	}

	if p.programID == "" {
		return fmt.Errorf("program id not specified: --%s", programIDFlag)
	}

	programPublicKey, err := solanawallet.PublicKeyFromAddress(p.programID)
	if err != nil {
		return fmt.Errorf("invalid program id: %w", err)
	}

	p.programPublicKey = programPublicKey

	if p.buildPath == "" {
		return fmt.Errorf("build path not specified: --%s", buildPathFlag)
	}

	if p.buildPath, err = resolveExistingFile(p.buildPath, "build path"); err != nil {
		return err
	}

	if p.commitment == "" {
		p.commitment = defaultCommitment
	}

	if err := validateCommitment(p.commitment); err != nil {
		return err
	}

	p.upgradeProgramVersion = strings.TrimSpace(p.upgradeProgramVersion)
	if p.upgradeProgramVersion == "" {
		return fmt.Errorf("upgrade program version not specified: --%s", upgradeProgramVersionFlag)
	}

	if _, err := parseSemanticVersionScore(p.upgradeProgramVersion); err != nil {
		return fmt.Errorf("invalid --%s: %w", upgradeProgramVersionFlag, err)
	}

	if p.adminKeyPath == "" && p.privateKeyConfig == "" {
		return fmt.Errorf("specify at least one: --%s or --%s", upgradeProgramAdminKeyPathFlag, privateKeyConfigFlag)
	}

	if p.privateKeyConfig == "" {
		if p.adminKeyPath, err = resolveExistingFile(p.adminKeyPath, "admin key"); err != nil {
			return err
		}
	}

	adminPrivateKey, err := solanatx.GetSolanaPrivateKey(
		p.adminKeyPath, p.privateKeyConfig, solanatx.GetKeyNameForSolanaAdmin())
	if err != nil {
		return fmt.Errorf("failed to load admin key: %w", err)
	}

	p.adminPrivateKey = adminPrivateKey

	if p.confirmationTimeoutSeconds == 0 {
		return fmt.Errorf("confirmation timeout must be greater than 0: --%s",
			upgradeProgramConfirmationTimeoutSecondsFlag)
	}

	p.confirmationTimeout = time.Duration(p.confirmationTimeoutSeconds) * time.Second //nolint:gosec

	return nil
}

func (p *upgradeProgramParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&p.rpcURL,
		rpcURLFlag,
		"",
		rpcURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.feePayerKeyPath,
		feePayerKeyFlag,
		"",
		feePayerKeyFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.upgradeAuthorityKeyPath,
		upgradeAuthorityKeyFlag,
		"",
		upgradeAuthorityKeyFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.programID,
		programIDFlag,
		"",
		programIDFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.buildPath,
		buildPathFlag,
		"",
		buildPathFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.commitment,
		commitmentFlag,
		defaultCommitment,
		commitmentFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.upgradeProgramVersion,
		upgradeProgramVersionFlag,
		"",
		upgradeProgramVersionFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.adminKeyPath,
		upgradeProgramAdminKeyPathFlag,
		"",
		upgradeProgramAdminKeyPathFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.privateKeyConfig,
		privateKeyConfigFlag,
		"",
		privateKeyConfigFlagDesc,
	)
	cmd.Flags().Uint64Var(
		&p.confirmationTimeoutSeconds,
		upgradeProgramConfirmationTimeoutSecondsFlag,
		defaultUpgradeProgramConfirmationTimeoutSeconds,
		upgradeProgramConfirmationTimeoutSecondsFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(upgradeProgramAdminKeyPathFlag, privateKeyConfigFlag)
}

// buildUpgradeDeployArgs builds the `solana program deploy` arguments for an upgrade. All paths
// must already be absolute. -k is the fee payer: the solana CLI resolves the default signer even
// when --fee-payer and --upgrade-authority are given, and fails without ~/.config/solana/id.json.
func buildUpgradeDeployArgs(rpcURL, feePayer, programID, upgradeAuthority, commitment, buildPath string) []string {
	return []string{
		solanaCLIProgramArg, "deploy",
		solanaCLIURLArg, rpcURL,
		"-k", feePayer,
		"--fee-payer", feePayer,
		"--program-id", programID,
		solanaCLIUpgradeAuthorityArg, upgradeAuthority,
		solanaCLICommitmentArg, commitment,
		buildPath,
	}
}

// deployArgs picks the fee payer (default: admin key) and the upgrade authority (default: admin
// key) and builds the deploy arguments. adminKeyPath is the absolute path of the admin keypair.
func (p *upgradeProgramParams) deployArgs(adminKeyPath string) []string {
	feePayer := p.feePayerKeyPath
	if feePayer == "" {
		feePayer = adminKeyPath
	}

	upgradeAuthority := p.upgradeAuthorityKeyPath
	if upgradeAuthority == "" {
		upgradeAuthority = adminKeyPath
	}

	return buildUpgradeDeployArgs(
		p.rpcURL, feePayer, p.programID, upgradeAuthority, p.commitment, p.buildPath)
}

func (p *upgradeProgramParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()

	_, _ = outputter.Write([]byte("Checking on-chain program config..."))
	outputter.WriteOutput()

	provider, err := solanawallet.NewProvider(p.rpcURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}

	txSender := solsendtx.NewTxSender(provider, nil)

	// every on-chain check runs before the deploy, so a wrong admin key or version never leaves
	// a deployed binary with a stale on-chain version
	cfg, err := txSender.GetProgramConfig(ctx, p.programPublicKey)
	if err != nil {
		return nil, fmt.Errorf("get program config: %w", err)
	}

	adminPK := p.adminPrivateKey.PublicKey()
	if cfg.Authority != adminPK {
		return nil, fmt.Errorf(
			"admin key %s does not match program config authority %s",
			adminPK.String(),
			cfg.Authority.String(),
		)
	}

	currentScore, err := parseSemanticVersionScore(cfg.VersionString)
	if err != nil {
		return nil, fmt.Errorf("parse on-chain version %q: %w", cfg.VersionString, err)
	}

	newScore, err := parseSemanticVersionScore(p.upgradeProgramVersion)
	if err != nil {
		return nil, fmt.Errorf("parse new version: %w", err)
	}

	if newScore < currentScore {
		return nil, fmt.Errorf(
			"--%s %q (score %d) is smaller than on-chain version %q (score %d)",
			upgradeProgramVersionFlag,
			p.upgradeProgramVersion,
			newScore,
			cfg.VersionString,
			currentScore,
		)
	}

	_, _ = outputter.Write([]byte("Upgrading Solana program..."))
	outputter.WriteOutput()

	adminKeyPath, cleanupAdminKey, err := solanaKeypairPathForCLI(p.adminKeyPath, p.adminPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("prepare admin keypair: %w", err)
	}

	defer cleanupAdminKey()

	output, err := common.ExecuteCLICommand("solana", p.deployArgs(adminKeyPath), ".")
	if err != nil {
		return nil, fmt.Errorf("solana program upgrade failed: %w", err)
	}

	_, _ = outputter.Write([]byte("Updating on-chain program version..."))
	outputter.WriteOutput()

	recentBlockhash, err := provider.GetLatestBlockhash(ctx)
	if err != nil {
		return nil, fmt.Errorf("get latest blockhash: %w", err)
	}

	txDto := solsendtx.UpdateProgramVersionDto{
		ProgramID:     p.programPublicKey,
		AuthorityAddr: adminPK.String(),
		VersionString: p.upgradeProgramVersion,
	}

	tx, err := txSender.CreateTx(
		ctx,
		adminPK,
		solsendtx.InstructionTypeUpdateProgramVersion,
		recentBlockhash,
		txDto,
	)
	if err != nil {
		return nil, fmt.Errorf("create update program version tx: %w", err)
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		return &p.adminPrivateKey
	})
	if err != nil {
		return nil, fmt.Errorf("sign update program version tx: %w", err)
	}

	sig, err := txSender.SendTx(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("send update program version tx: %w", err)
	}

	if err := provider.WaitForSignature(ctx, *sig, rpc.CommitmentFinalized, p.confirmationTimeout); err != nil {
		return nil, fmt.Errorf("wait for update program version confirmation: %w", err)
	}

	combined := strings.TrimSpace(output) + "\nupdate program version tx finalized: " + sig.String()

	return &deployProgramResult{
		Output:      combined,
		TxSignature: sig.String(),
	}, nil
}

// parseSemanticVersionScore maps major.minor.patch to 1000*major + 100*minor + 10*patch for ordering.
// Pre-release (-) and build (+) suffixes are stripped before parsing. A leading "v" is allowed.
func parseSemanticVersionScore(v string) (uint64, error) {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(strings.ToLower(v), "v")

	if v == "" {
		return 0, fmt.Errorf("empty version")
	}

	if i := strings.IndexByte(v, '+'); i >= 0 {
		v = v[:i]
	}

	if i := strings.IndexByte(v, '-'); i >= 0 {
		v = v[:i]
	}

	parts := strings.Split(v, ".")
	if len(parts) == 0 {
		return 0, fmt.Errorf("invalid version %q", v)
	}

	parsePart := func(idx int) (uint64, error) {
		if idx >= len(parts) {
			return 0, nil
		}

		s := strings.TrimSpace(parts[idx])
		if s == "" {
			return 0, fmt.Errorf("empty version segment in %q", v)
		}

		for _, r := range s {
			if r < '0' || r > '9' {
				return 0, fmt.Errorf("non-numeric version segment %q in %q", s, v)
			}
		}

		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid version segment %q: %w", s, err)
		}

		return n, nil
	}

	major, err := parsePart(0)
	if err != nil {
		return 0, err
	}

	minor, err := parsePart(1)
	if err != nil {
		return 0, err
	}

	patch, err := parsePart(2)
	if err != nil {
		return 0, err
	}

	return major*1000 + minor*100 + patch*10, nil
}
