package cliwalletcreate

import (
	"encoding/hex"
	"fmt"
	"strings"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/spf13/cobra"
)

const (
	validatorDataDirFlag = "validator-data-dir"
	validatorConfigFlag  = "validator-config"
	chainIDFlag          = "chain"
	walletTypeFlag       = "type"
	forceRegenerateFlag  = "force"
	showPrivateKeyFlag   = "show-pk"

	validatorDataDirFlagDesc = "(mandatory validator-config not specified) Path to bridge chain data directory when using local secrets manager" //nolint:lll
	validatorConfigFlagDesc  = "(mandatory validator-data not specified) Path to to bridge chain secrets manager config file"                    //nolint:lll
	chainIDFlagDesc          = "chain ID (prime, vector, etc)"
	walletTypeFlagDesc       = "type of wallet (cardano, evm, relayer evm, etc)"
	forceRegenerateFlagDesc  = "force regenerating keys even if they exist in specified directory"
	showPrivateKeyFlagDesc   = "show private key in output"
)

type initParams struct {
	validatorDataDir string
	validatorConfig  string
	chainID          string
	walletType       string
	forceRegenerate  bool
	showPrivateKey   bool
}

func (ip *initParams) validateFlags() error {
	if ip.validatorDataDir == "" && ip.validatorConfig == "" {
		return fmt.Errorf("specify at least one of: %s, %s", validatorDataDirFlag, validatorConfigFlag)
	}

	if ip.chainID == "" {
		return fmt.Errorf("--%s flag not specified", chainIDFlag)
	}

	return nil
}

func (ip *initParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&ip.validatorDataDir,
		validatorDataDirFlag,
		"",
		validatorDataDirFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.validatorConfig,
		validatorConfigFlag,
		"",
		validatorConfigFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.chainID,
		chainIDFlag,
		"",
		chainIDFlagDesc,
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

	cmd.Flags().StringVar(
		&ip.walletType,
		walletTypeFlag,
		"",
		walletTypeFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(validatorDataDirFlag, validatorConfigFlag)
}

func (ip *initParams) Execute() (common.ICommandResult, error) {
	secretsManager, err := common.GetSecretsManager(ip.validatorDataDir, ip.validatorConfig, true)
	if err != nil {
		return nil, err
	}

	switch walletType := strings.ToLower(ip.walletType); walletType {
	case "relayer-evm":
		evmWallet, err := eth.CreateAndSaveRelayerEVMPrivateKey(secretsManager, ip.chainID, ip.forceRegenerate)
		if err != nil {
			return nil, err
		}

		pk, pub, addr := evmWallet.GetHexData()

		return &evmCmdResult{
			ChainID:        ip.chainID,
			PrivateKey:     pk,
			PublicKey:      pub,
			Address:        addr,
			showPrivateKey: ip.showPrivateKey,
		}, nil

	case "evm", "batcher-evm":
		privateKey, err := eth.CreateAndSaveBatcherEVMPrivateKey(secretsManager, ip.chainID, ip.forceRegenerate)
		if err != nil {
			return nil, err
		}

		pkBytes, err := privateKey.Marshal()
		if err != nil {
			return nil, err
		}

		pubBytes := privateKey.PublicKey().Marshal()

		return &evmCmdResult{
			ChainID:        ip.chainID,
			PrivateKey:     string(pkBytes),
			PublicKey:      hex.EncodeToString(pubBytes),
			showPrivateKey: ip.showPrivateKey,
		}, nil

	default:
		wallet, err := cardanotx.GenerateWallet(secretsManager, ip.chainID, walletType == "stake", ip.forceRegenerate)
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

		return &cardanoCmdResult{
			SigningKey:      wallet.MultiSig.GetSigningKey(),
			VerifyingKey:    wallet.MultiSig.GetVerificationKey(),
			KeyHash:         keyHash,
			SigningKeyFee:   wallet.MultiSigFee.GetSigningKey(),
			VerifyingKeyFee: wallet.MultiSigFee.GetVerificationKey(),
			KeyHashFee:      keyHashFee,
			showPrivateKey:  ip.showPrivateKey,
			ChainID:         ip.chainID,
		}, nil
	}
}
