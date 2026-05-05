package clideploysolana

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	solanawallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
)

const (
	altRPCURLFlag                     = "url"
	altProgramIDFlag                  = "program-id"
	altAdminKeyPathFlag               = "admin-key"
	altConfirmationTimeoutSecondsFlag = "confirmation-timeout-seconds"

	altRPCURLFlagDesc                     = "Solana RPC URL"
	altProgramIDFlagDesc                  = "skyline program ID"
	altAdminKeyPathFlagDesc               = "path to Solana admin keypair file"
	altConfirmationTimeoutSecondsFlagDesc = "max wait time in seconds for tx finalization"

	defaultALTConfirmationTimeoutSeconds = uint64(120)
)

type altCommonParams struct {
	rpcURL                     string
	programID                  string
	adminKeyPath               string
	confirmationTimeoutSeconds uint64

	adminPrivateKey     solana.PrivateKey
	confirmationTimeout time.Duration
	programPublicKey    solana.PublicKey
}

func (p *altCommonParams) setCommonFlags(cmd commonFlagSetter) {
	cmd.StringVar(
		&p.rpcURL,
		altRPCURLFlag,
		"",
		altRPCURLFlagDesc,
	)
	cmd.StringVar(
		&p.programID,
		altProgramIDFlag,
		"",
		altProgramIDFlagDesc,
	)
	cmd.StringVar(
		&p.adminKeyPath,
		altAdminKeyPathFlag,
		"",
		altAdminKeyPathFlagDesc,
	)
	cmd.Uint64Var(
		&p.confirmationTimeoutSeconds,
		altConfirmationTimeoutSecondsFlag,
		defaultALTConfirmationTimeoutSeconds,
		altConfirmationTimeoutSecondsFlagDesc,
	)
}

func (p *altCommonParams) validateCommonFlags() error {
	if !common.IsValidHTTPURL(p.rpcURL) {
		return fmt.Errorf(
			"invalid --%s flag (must be a valid http or https URL)",
			altRPCURLFlag,
		)
	}

	if p.adminKeyPath == "" {
		return fmt.Errorf("admin key path not specified: --%s", altAdminKeyPathFlag)
	}

	if p.programID == "" {
		return fmt.Errorf("program ID not specified: --%s", altProgramIDFlag)
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
		return fmt.Errorf(
			"confirmation timeout must be greater than 0: --%s",
			altConfirmationTimeoutSecondsFlag,
		)
	}

	p.confirmationTimeout = time.Duration(p.confirmationTimeoutSeconds) * time.Second //nolint:gosec

	if p.programID == "" {
		return fmt.Errorf("program ID not specified: --%s", altProgramIDFlag)
	}

	programID, err := solanawallet.PublicKeyFromAddress(p.programID)
	if err != nil {
		return fmt.Errorf("invalid program ID: %w", err)
	}

	p.programPublicKey = programID

	return nil
}

type commonFlagSetter interface {
	StringVar(p *string, name string, value string, usage string)
	Uint64Var(p *uint64, name string, value uint64, usage string)
}
