package clideploysolana

import (
	"fmt"
	"path/filepath"

	"github.com/Ethernal-Tech/apex-bridge/common"
	solanatx "github.com/Ethernal-Tech/apex-bridge/solana"
	"github.com/spf13/cobra"
)

const (
	exportKeypairPrivateKeyFlag = "key"
	exportKeypairOutputFlag     = "output"

	exportKeypairPrivateKeyFlagDesc = "base58 solana private key string from wallet"
	exportKeypairOutputFlagDesc     = "path to write the solana keypair json file (default: key.json)"

	defaultExportKeypairOutputPath = "key.json"
)

type exportKeypairParams struct {
	privateKey string
	outputPath string
}

func (p *exportKeypairParams) validateFlags() error {
	if p.privateKey == "" {
		return fmt.Errorf("--%s not specified", exportKeypairPrivateKeyFlag)
	}

	if p.outputPath == "" {
		p.outputPath = defaultExportKeypairOutputPath
	}

	return nil
}

func (p *exportKeypairParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&p.privateKey,
		exportKeypairPrivateKeyFlag,
		"",
		exportKeypairPrivateKeyFlagDesc,
	)
	cmd.Flags().StringVar(
		&p.outputPath,
		exportKeypairOutputFlag,
		"",
		exportKeypairOutputFlagDesc,
	)
}

func (p *exportKeypairParams) Execute(_ common.OutputFormatter) (common.ICommandResult, error) {
	privateKey, err := solanatx.PrivateKeyFromWalletString(p.privateKey)
	if err != nil {
		return nil, err
	}

	outputPath := filepath.Clean(p.outputPath)
	if err := solanatx.WriteSolanaKeypairFile(outputPath, privateKey); err != nil {
		return nil, fmt.Errorf("write solana keypair file: %w", err)
	}

	return &exportKeypairResult{
		OutputPath: outputPath,
		PublicKey:  privateKey.PublicKey().String(),
	}, nil
}
