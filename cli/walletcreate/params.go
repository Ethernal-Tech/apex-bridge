package cliwalletcreate

import (
	"errors"
	"fmt"
	"path"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/spf13/cobra"
)

const (
	directoryFlag        = "dir"
	chainIDFlag          = "chain"
	generateStakeKeyFlag = "stake"
	forceRegenerateFlag  = "force"
	showPrivateKeyFlag   = "show-pk"

	directoryFlagDesc        = "wallet directory"
	chainIDFlagDesc          = "chain ID (prime, vector, etc)"
	generateStakeKeyFlagDesc = "stake wallet"
	forceRegenerateFlagDesc  = "force regenerating keys even if they exist in specified directory"
	showPrivateKeyFlagDesc   = "show private key in output"
)

type initParams struct {
	directory        string
	chainID          string
	generateStakeKey bool
	forceRegenerate  bool
	showPrivateKey   bool
}

func (ip *initParams) validateFlags() error {
	if ip.directory == "" {
		return fmt.Errorf("invalid directory: %s", ip.directory)
	}

	if ip.chainID == "" {
		return errors.New("--chain flag not specified")
	}

	return nil
}

func (ip *initParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&ip.directory,
		directoryFlag,
		"",
		directoryFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.chainID,
		chainIDFlag,
		"",
		chainIDFlagDesc,
	)

	cmd.Flags().BoolVar(
		&ip.generateStakeKey,
		generateStakeKeyFlag,
		false,
		generateStakeKeyFlagDesc,
	)

	cmd.Flags().BoolVar(
		&ip.forceRegenerate,
		forceRegenerateFlag,
		false,
		forceRegenerateFlagDesc,
	)

	cmd.Flags().BoolVar(
		&ip.showPrivateKey,
		showPrivateKeyFlag,
		false,
		showPrivateKeyFlagDesc,
	)
}

func (ip *initParams) Execute() (common.ICommandResult, error) {
	dir := path.Clean(ip.directory)

	wallet, err := cardanotx.GenerateWallet(path.Join(dir, ip.chainID), ip.generateStakeKey, ip.forceRegenerate)
	if err != nil {
		return nil, err
	}

	keyHash, err := cardanowallet.GetKeyHash(wallet.MultiSig.GetVerificationKey())
	if err != nil {
		return nil, err
	}

	keyHashFee, err := cardanowallet.GetKeyHash(wallet.MultiSigFee.GetVerificationKey())
	if err != nil {
		return nil, err
	}

	return &CmdResult{
		SigningKey:      wallet.MultiSig.GetSigningKey(),
		VerifyingKey:    wallet.MultiSig.GetVerificationKey(),
		KeyHash:         keyHash,
		SigningKeyFee:   wallet.MultiSigFee.GetSigningKey(),
		VerifyingKeyFee: wallet.MultiSigFee.GetVerificationKey(),
		KeyHashFee:      keyHashFee,
		showPrivateKey:  ip.showPrivateKey,
		chainID:         ip.chainID,
	}, nil
}
