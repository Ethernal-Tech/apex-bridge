package clideploysolana

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	solsendtx "github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	solanawallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/spf13/cobra"
)

const (
	programVersionRPCURLFlag    = "url"
	programVersionProgramIDFlag = "program-id"

	programVersionRPCURLFlagDesc    = "Solana RPC URL"
	programVersionProgramIDFlagDesc = "skyline program ID"
)

type programVersionParams struct {
	rpcURL    string
	programID string

	programPublicKey solana.PublicKey
}

type programVersionResult struct {
	ProgramID     string
	VersionString string
	DeployedAt    int64
	Authority     string
}

func (r programVersionResult) GetOutput() string {
	var buffer bytes.Buffer

	buffer.WriteString("[PROGRAM VERSION]\n")
	buffer.WriteString(common.FormatKV([]string{
		fmt.Sprintf("Program ID|%s", r.ProgramID),
		fmt.Sprintf("Version|%s", r.VersionString),
		fmt.Sprintf("Deployed at|%s", time.Unix(r.DeployedAt, 0).UTC().Format(time.RFC3339)),
		fmt.Sprintf("Authority|%s", r.Authority),
	}))

	return buffer.String()
}

func (p *programVersionParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&p.rpcURL,
		programVersionRPCURLFlag,
		"",
		programVersionRPCURLFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.programID,
		programVersionProgramIDFlag,
		"",
		programVersionProgramIDFlagDesc,
	)
}

func (p *programVersionParams) validateFlags() error {
	if !common.IsValidHTTPURL(p.rpcURL) {
		return fmt.Errorf(
			"invalid --%s flag (must be a valid http or https URL)",
			programVersionRPCURLFlag,
		)
	}

	if p.programID == "" {
		return fmt.Errorf("program ID not specified: --%s", programVersionProgramIDFlag)
	}

	programPublicKey, err := solanawallet.PublicKeyFromAddress(p.programID)
	if err != nil {
		return fmt.Errorf("invalid program id: %w", err)
	}

	p.programPublicKey = programPublicKey

	return nil
}

func (p *programVersionParams) Execute(_ common.OutputFormatter) (common.ICommandResult, error) {
	ctx := context.Background()

	provider, err := solanawallet.NewProvider(p.rpcURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}

	txSender := solsendtx.NewTxSender(provider, nil)

	programConfig, err := txSender.GetProgramConfig(ctx, p.programPublicKey)
	if err != nil {
		return nil, fmt.Errorf("get program config: %w", err)
	}

	return &programVersionResult{
		ProgramID:     p.programPublicKey.String(),
		VersionString: programConfig.VersionString,
		DeployedAt:    programConfig.DeployedAt,
		Authority:     programConfig.Authority.String(),
	}, nil
}
