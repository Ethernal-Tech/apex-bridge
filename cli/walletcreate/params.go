package cliwalletcreate

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	solanatx "github.com/Ethernal-Tech/apex-bridge/solana"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/spf13/cobra"
)

const (
	validatorDataDirFlag = "validator-data-dir"
	validatorConfigFlag  = "validator-config"
	chainIDFlag          = "chain"
	walletTypeFlag       = "type"
	networkIDFlag        = "network-id"
	forceRegenerateFlag  = "force"
	showPrivateKeyFlag   = "show-pk"

	validatorDataDirFlagDesc = "(mandatory validator-config not specified) path to bridge chain data directory when using local secrets manager" //nolint:lll
	validatorConfigFlagDesc  = "(mandatory validator-data not specified) path to bridge chain secrets manager config file"                       //nolint:lll
	chainIDFlagDesc          = "chain ID (prime, vector, etc)"
	walletTypeFlagDesc       = "wallet type (cardano, evm, solana_admin, new_solana_admin, solana_program, etc.)"
	networkIDFlagDesc        = "network id"
	forceRegenerateFlagDesc  = "force regenerating keys even if they exist in specified directory"
	showPrivateKeyFlagDesc   = "show private key in output"

	keyFileFlag = "key-file"

	importKeyFlagDesc = "existing solana private key to store instead of generating a new one (base58, base64 or keypair JSON); solana types only. Prefer --key-file: a key on the command line ends up in shell history" //nolint:lll
	keyFileFlagDesc   = "path to an existing solana keypair file (solana-keygen JSON) to store instead of generating a new one; solana types only"                                                                        //nolint:lll

	solanaAdminKeyType    = "solana_admin"
	newSolanaAdminKeyType = "new_solana_admin"
	solanaProgramKeyType  = "solana_program"
	relayerSolanaKeyType  = "relayer-solana"
	batcherSolanaKeyType  = "batcher-solana"
	solanaKeyTypesForHelp = relayerSolanaKeyType + ", " + batcherSolanaKeyType + ", " +
		solanaAdminKeyType + ", " + newSolanaAdminKeyType + ", " + solanaProgramKeyType
)

type walletCreateParams struct {
	validatorDataDir string
	validatorConfig  string
	chainID          string
	walletType       string
	networkID        uint32
	forceRegenerate  bool
	showPrivateKey   bool
	key              string
	keyFile          string

	importKey solana.PrivateKey
}

func (ip *walletCreateParams) validateFlags() error {
	if ip.validatorDataDir == "" && ip.validatorConfig == "" {
		return fmt.Errorf("specify at least one of: %s, %s", validatorDataDirFlag, validatorConfigFlag)
	}

	if ip.chainID == "" {
		return fmt.Errorf("--%s flag not specified", chainIDFlag)
	}

	return ip.resolveImportKey()
}

// resolveImportKey parses --key or --key-file, when one is set, into importKey.
func (ip *walletCreateParams) resolveImportKey() error {
	ip.importKey = nil

	if ip.key == "" && ip.keyFile == "" {
		return nil
	}

	if ip.key != "" && ip.keyFile != "" {
		return fmt.Errorf("specify only one of --%s or --%s", keyFlag, keyFileFlag)
	}

	if !isSolanaKeyType(ip.walletType) {
		return fmt.Errorf("--%s and --%s are only supported for --%s %s",
			keyFlag, keyFileFlag, walletTypeFlag, solanaKeyTypesForHelp)
	}

	if ip.key != "" {
		privateKey, err := solanatx.PrivateKeyFromWalletString(ip.key)
		if err != nil {
			return fmt.Errorf("invalid --%s: %w", keyFlag, err)
		}

		ip.importKey = privateKey

		return nil
	}

	content, err := os.ReadFile(filepath.Clean(ip.keyFile))
	if err != nil {
		return fmt.Errorf("--%s: %w", keyFileFlag, err)
	}

	privateKey, err := solanatx.PrivateKeyFromWalletString(string(content))
	if err != nil {
		return fmt.Errorf("--%s is not a valid solana keypair file: %w", keyFileFlag, err)
	}

	ip.importKey = privateKey

	return nil
}

func isSolanaKeyType(walletType string) bool {
	switch strings.ToLower(walletType) {
	case relayerSolanaKeyType, batcherSolanaKeyType, solanaAdminKeyType, newSolanaAdminKeyType, solanaProgramKeyType:
		return true
	default:
		return false
	}
}

func (ip *walletCreateParams) setFlags(cmd *cobra.Command) {
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
	cmd.Flags().Uint32Var(
		&ip.networkID,
		networkIDFlag,
		uint32(cardanowallet.TestNetNetwork),
		networkIDFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.key,
		keyFlag,
		"",
		importKeyFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.keyFile,
		keyFileFlag,
		"",
		keyFileFlagDesc,
	)

	cmd.MarkFlagsMutuallyExclusive(validatorDataDirFlag, validatorConfigFlag)
	cmd.MarkFlagsMutuallyExclusive(keyFlag, keyFileFlag)
}

func (ip *walletCreateParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
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

	case relayerSolanaKeyType, batcherSolanaKeyType, solanaAdminKeyType, newSolanaAdminKeyType, solanaProgramKeyType:
		return ip.executeSolana(secretsManager, walletType)

	case "relayer-cardano":
		wallet, err := cardanotx.GenerateWallet(secretsManager, ip.chainID, false, ip.forceRegenerate)
		if err != nil {
			return nil, err
		}

		addr, err := cardanowallet.NewBaseAddress(cardanowallet.CardanoNetworkType(ip.networkID), //nolint:gosec
			wallet.MultiSig.VerificationKey, wallet.MultiSig.StakeVerificationKey)
		if err != nil {
			return nil, err
		}

		return &cardanoRelayerCmdResult{
			ChainID:        ip.chainID,
			Wallet:         wallet.MultiSig,
			Address:        addr.String(),
			showPrivateKey: ip.showPrivateKey,
		}, nil

	default:
		isStake := walletType == "stake"

		wallet, err := cardanotx.GenerateWallet(secretsManager, ip.chainID, isStake, ip.forceRegenerate)
		if err != nil {
			return nil, err
		}

		var stakeKeyHash, stakeKeyHashFee string

		keyHash, err := cardanowallet.GetKeyHash(wallet.MultiSig.VerificationKey)
		if err != nil {
			return nil, err
		}

		keyHashFee, err := cardanowallet.GetKeyHash(wallet.Fee.VerificationKey)
		if err != nil {
			return nil, err
		}

		if isStake {
			stakeKeyHash, err = cardanowallet.GetKeyHash(wallet.MultiSig.StakeVerificationKey)
			if err != nil {
				return nil, err
			}

			stakeKeyHashFee, err = cardanowallet.GetKeyHash(wallet.Fee.StakeVerificationKey)
			if err != nil {
				return nil, err
			}
		}

		return &cardanoCmdResult{
			Multisig:        wallet.MultiSig,
			Fee:             wallet.Fee,
			KeyHash:         keyHash,
			KeyHashFee:      keyHashFee,
			StakeKeyHash:    stakeKeyHash,
			StakeKeyHashFee: stakeKeyHashFee,
			showPrivateKey:  ip.showPrivateKey,
			ChainID:         ip.chainID,
		}, nil
	}
}

// executeSolana stores the key given with --key/--key-file, or generates one when neither is set.
func (ip *walletCreateParams) executeSolana(
	secretsManager secrets.SecretsManager, walletType string,
) (common.ICommandResult, error) {
	var (
		keyName  string
		resultID string
		generate func() (*solana.PrivateKey, error)
	)

	switch walletType {
	case relayerSolanaKeyType:
		keyName, resultID = solanatx.GetKeyNameForRelayerSolana(ip.chainID), ip.chainID
		generate = func() (*solana.PrivateKey, error) {
			return solanatx.GenerateAndStoreRelayerSolanaPrivateKey(secretsManager, ip.chainID, ip.forceRegenerate)
		}
	case batcherSolanaKeyType:
		keyName, resultID = solanatx.GetKeyNameForBatcherSolana(ip.chainID), ip.chainID
		generate = func() (*solana.PrivateKey, error) {
			return solanatx.GenerateAndStoreBatcherSolanaPrivateKey(secretsManager, ip.chainID, ip.forceRegenerate)
		}
	case solanaAdminKeyType:
		keyName, resultID = solanatx.GetKeyNameForSolanaAdmin(), solanaAdminKeyType
		generate = func() (*solana.PrivateKey, error) {
			return solanatx.GenerateAndStoreSolanaAdminPrivateKey(secretsManager, ip.forceRegenerate)
		}
	case newSolanaAdminKeyType:
		keyName, resultID = solanatx.GetKeyNameForNewSolanaAdmin(), newSolanaAdminKeyType
		generate = func() (*solana.PrivateKey, error) {
			return solanatx.GenerateAndStoreNewSolanaAdminPrivateKey(secretsManager, ip.forceRegenerate)
		}
	case solanaProgramKeyType:
		keyName, resultID = solanatx.GetKeyNameForSolanaProgram(), solanaProgramKeyType
		generate = func() (*solana.PrivateKey, error) {
			return solanatx.GenerateAndStoreSolanaProgramPrivateKey(secretsManager, ip.forceRegenerate)
		}
	default:
		return nil, fmt.Errorf("unsupported solana wallet type: %s", walletType)
	}

	var privateKey solana.PrivateKey

	if ip.importKey != nil {
		if err := solanatx.ImportSolanaPrivateKey(
			secretsManager, keyName, ip.importKey, ip.forceRegenerate); err != nil {
			return nil, err
		}

		privateKey = ip.importKey
	} else {
		generated, err := generate()
		if err != nil {
			return nil, err
		}

		privateKey = *generated
	}

	return &evmCmdResult{
		ChainID:        resultID,
		PrivateKey:     privateKey.String(),
		PublicKey:      privateKey.PublicKey().String(),
		Address:        privateKey.PublicKey().String(),
		showPrivateKey: ip.showPrivateKey,
	}, nil
}
