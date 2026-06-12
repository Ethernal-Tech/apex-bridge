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

	upgradeProgramAdminKeyPathFlag                   = "admin-key"                                                                                        //nolint:lll //nolint:gofmt
	upgradeProgramAdminKeyPathFlagDesc               = "path to bridge admin keypair (signs update_program_version; must match program config authority)" //nolint:lll
	upgradeProgramConfirmationTimeoutSecondsFlag     = "confirmation-timeout-seconds"
	upgradeProgramConfirmationTimeoutSecondsFlagDesc = "max wait time in seconds for update-program-version tx finalization" //nolint:lll

	defaultUpgradeProgramConfirmationTimeoutSeconds = uint64(120)
)

type upgradeProgramParams struct {
	rpcURL                     string
	feePayerKeyPath            string
	programKeyPath             string
	programID                  string
	buildPath                  string
	commitment                 string
	upgradeProgramVersion      string
	adminKeyPath               string
	confirmationTimeoutSeconds uint64

	programPublicKey    solana.PublicKey
	adminPrivateKey     solana.PrivateKey
	confirmationTimeout time.Duration
}

func (p *upgradeProgramParams) validateFlags() error {
	if !common.IsValidHTTPURL(p.rpcURL) {
		return fmt.Errorf("invalid --%s flag (must be a valid http or https URL)", rpcURLFlag)
	}

	if p.feePayerKeyPath == "" {
		return fmt.Errorf("fee payer key path not specified: --%s", feePayerKeyFlag)
	}

	if _, err := os.Stat(p.feePayerKeyPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("fee payer key file does not exist: %s", p.feePayerKeyPath)
		}

		return fmt.Errorf("failed to check fee payer key file: %w", err)
	}

	if p.programKeyPath == "" {
		return fmt.Errorf("program key path not specified: --%s", programKeyFlag)
	}

	if _, err := os.Stat(p.programKeyPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("program key file does not exist: %s", p.programKeyPath)
		}

		return fmt.Errorf("failed to check program key file: %w", err)
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

	if _, err := os.Stat(p.buildPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("build path does not exist: %s", p.buildPath)
		}

		return fmt.Errorf("failed to check build path: %w", err)
	}

	if p.commitment == "" {
		p.commitment = defaultCommitment
	}

	p.upgradeProgramVersion = strings.TrimSpace(p.upgradeProgramVersion)
	if p.upgradeProgramVersion == "" {
		return fmt.Errorf("upgrade program version not specified: --%s", upgradeProgramVersionFlag)
	}

	if _, err := parseSemanticVersionScore(p.upgradeProgramVersion); err != nil {
		return fmt.Errorf("invalid --%s: %w", upgradeProgramVersionFlag, err)
	}

	if p.adminKeyPath == "" {
		return fmt.Errorf("admin key path not specified: --%s", upgradeProgramAdminKeyPathFlag)
	}

	p.adminKeyPath = filepath.Clean(p.adminKeyPath)
	if _, err := os.Stat(p.adminKeyPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("admin key file does not exist: %s", p.adminKeyPath)
		}

		return fmt.Errorf("failed to check admin key file: %w", err)
	}

	adminPrivateKey, err := solana.PrivateKeyFromSolanaKeygenFile(p.adminKeyPath)
	if err != nil {
		return fmt.Errorf("failed to load admin keypair file: %w", err)
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
		&p.programKeyPath,
		programKeyFlag,
		"",
		programKeyFlagDesc,
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
	cmd.Flags().Uint64Var(
		&p.confirmationTimeoutSeconds,
		upgradeProgramConfirmationTimeoutSecondsFlag,
		defaultUpgradeProgramConfirmationTimeoutSeconds,
		upgradeProgramConfirmationTimeoutSecondsFlagDesc,
	)
}

func (p *upgradeProgramParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()
	buildPath := filepath.Clean(p.buildPath)

	_, _ = outputter.Write([]byte("Upgrading Solana program..."))
	outputter.WriteOutput()

	args := []string{
		"program", "deploy",
		"--url", p.rpcURL,
		"--fee-payer", p.feePayerKeyPath,
		"-k", p.programKeyPath,
		"--program-id", p.programID,
		"--commitment", p.commitment,
		buildPath,
	}

	output, err := common.ExecuteCLICommand("solana", args, ".")
	if err != nil {
		return nil, fmt.Errorf("solana program upgrade failed: %w", err)
	}

	_, _ = outputter.Write([]byte("Updating on-chain program version..."))
	outputter.WriteOutput()

	provider, err := solanawallet.NewProvider(p.rpcURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}

	txSender := solsendtx.NewTxSender(provider, nil)

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
