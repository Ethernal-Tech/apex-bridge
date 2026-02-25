package clideploysolana

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

const (
	rpcURLFlag        = "url"
	feePayerKeyFlag   = "fee-payer"
	programKeyFlag    = "key"
	buildPathFlag     = "build-path"
	commitmentFlag    = "commitment"
	defaultCommitment = "finalized"

	rpcURLFlagDesc      = "Solana RPC URL"
	feePayerKeyFlagDesc = "path to fee payer keypair file"
	programKeyFlagDesc  = "path to program keypair file"
	buildPathFlagDesc   = "path to the compiled program (.so file)"
	commitmentFlagDesc  = "commitment level (processed, confirmed, finalized)"
)

type deployProgramParams struct {
	rpcURL          string
	feePayerKeyPath string
	programKeyPath  string
	buildPath       string
	commitment      string
}

func (p *deployProgramParams) validateFlags() error {
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

	return nil
}

func (p *deployProgramParams) setFlags(cmd *cobra.Command) {
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
}

func (p *deployProgramParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	buildPath := filepath.Clean(p.buildPath)

	_, _ = outputter.Write([]byte("Deploying Solana program..."))
	outputter.WriteOutput()

	args := []string{
		"program", "deploy",
		"--url", p.rpcURL,
		"--fee-payer", p.feePayerKeyPath,
		"-k", p.programKeyPath,
		"--commitment", p.commitment,
		buildPath,
	}

	output, err := common.ExecuteCLICommand("solana", args, ".")
	if err != nil {
		return nil, fmt.Errorf("solana program deploy failed: %w", err)
	}

	return &deployProgramResult{Output: output}, nil
}
