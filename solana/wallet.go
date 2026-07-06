package solanatx

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	solanawallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
)

func GetKeyNameForSolanaAdmin() string {
	return secrets.OtherKeyLocalPrefix + "solana_admin"
}

func GetKeyNameForSolanaProgram() string {
	return secrets.OtherKeyLocalPrefix + "solana_program_key"
}

func GetSolanaPrivateKey(keyPath, config, keyName string) (solana.PrivateKey, error) {
	if keyPath != "" {
		return solana.PrivateKeyFromSolanaKeygenFile(keyPath)
	}

	secretsManager, err := common.GetSecretsManager("", config, false)
	if err != nil {
		return nil, err
	}

	bytes, err := secretsManager.GetSecret(keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to load solana private key: %w", err)
	}

	var privateKey solana.PrivateKey

	if err := json.Unmarshal(bytes, &privateKey); err != nil {
		return nil, fmt.Errorf("failed to unmarshal solana private key: %w", err)
	}

	return privateKey, nil
}

func MarshalSolanaKeygenFileContent(privateKey solana.PrivateKey) ([]byte, error) {
	values := make([]int, len(privateKey))
	for i, b := range privateKey {
		values[i] = int(b)
	}

	content, err := json.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("marshal solana keypair: %w", err)
	}

	return content, nil
}

func WriteSolanaKeypairFile(path string, privateKey solana.PrivateKey) error {
	content, err := MarshalSolanaKeygenFileContent(privateKey)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, content, 0o600); err != nil {
		return fmt.Errorf("write solana keypair file: %w", err)
	}

	return nil
}

func WriteSolanaKeypairTempFile(privateKey solana.PrivateKey) (string, error) {
	file, err := os.CreateTemp("", "solana-keypair-*.json")
	if err != nil {
		return "", fmt.Errorf("create temp keypair file: %w", err)
	}

	path := file.Name()

	if err := file.Close(); err != nil {
		_ = os.Remove(path)

		return "", fmt.Errorf("close temp keypair file: %w", err)
	}

	if err := WriteSolanaKeypairFile(path, privateKey); err != nil {
		_ = os.Remove(path)

		return "", err
	}

	return path, nil
}

func PrivateKeyFromWalletString(privateKey string) (solana.PrivateKey, error) {
	privateKey = strings.TrimSpace(privateKey)
	if privateKey == "" {
		return nil, fmt.Errorf("empty solana private key")
	}

	if strings.HasPrefix(privateKey, "\"") {
		var encoded string

		if err := json.Unmarshal([]byte(privateKey), &encoded); err != nil {
			return nil, fmt.Errorf("invalid solana private key json string: %w", err)
		}

		return PrivateKeyFromWalletString(encoded)
	}

	if strings.HasPrefix(privateKey, "[") {
		var values []int

		if err := json.Unmarshal([]byte(privateKey), &values); err != nil {
			return nil, fmt.Errorf("invalid solana keypair json: %w", err)
		}

		return privateKeyFromIntSlice(values)
	}

	if key, err := solana.PrivateKeyFromBase58(privateKey); err == nil {
		return key, nil
	}

	raw, err := base64.StdEncoding.DecodeString(privateKey)
	if err == nil {
		if key, err := privateKeyFromBytes(raw); err == nil {
			return key, nil
		}
	}

	return nil, fmt.Errorf("invalid solana private key")
}

func privateKeyFromIntSlice(values []int) (solana.PrivateKey, error) {
	raw := make([]byte, len(values))

	for i, value := range values {
		if value < 0 || value > 255 {
			return nil, fmt.Errorf("invalid keypair byte at index %d: %d", i, value)
		}

		raw[i] = byte(value)
	}

	return privateKeyFromBytes(raw)
}

func privateKeyFromBytes(raw []byte) (solana.PrivateKey, error) {
	key := solana.PrivateKey(raw)

	if err := key.Validate(); err != nil {
		return nil, fmt.Errorf("invalid solana private key: %w", err)
	}

	return key, nil
}

func generateAndStoreSolanaPrivateKey(
	mngr secrets.SecretsManager, keyName string, forceRegenerate bool,
) (*solana.PrivateKey, error) {
	if mngr.HasSecret(keyName) {
		if !forceRegenerate {
			return loadSolanaPrivateKeyFromSecretsManager(mngr, keyName)
		}

		if err := mngr.RemoveSecret(keyName); err != nil {
			return nil, err
		}
	}

	privateKey, err := solana.NewRandomPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate wallet: %w", err)
	}

	bytes, err := json.Marshal(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal wallet: %w", err)
	}

	if err := mngr.SetSecret(keyName, bytes); err != nil {
		return nil, fmt.Errorf("failed to store wallet: %w", err)
	}

	return &privateKey, nil
}

func loadSolanaPrivateKeyFromSecretsManager(
	mngr secrets.SecretsManager, keyName string,
) (*solana.PrivateKey, error) {
	bytes, err := mngr.GetSecret(keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to load wallet: %w", err)
	}

	var privateKey solana.PrivateKey

	if err := json.Unmarshal(bytes, &privateKey); err != nil {
		return nil, fmt.Errorf("failed to load wallet: %w", err)
	}

	return &privateKey, nil
}

func GenerateAndStoreSolanaAdminPrivateKey(
	mngr secrets.SecretsManager, forceRegenerate bool,
) (*solana.PrivateKey, error) {
	return generateAndStoreSolanaPrivateKey(mngr, GetKeyNameForSolanaAdmin(), forceRegenerate)
}

func GenerateAndStoreSolanaProgramPrivateKey(
	mngr secrets.SecretsManager, forceRegenerate bool,
) (*solana.PrivateKey, error) {
	return generateAndStoreSolanaPrivateKey(mngr, GetKeyNameForSolanaProgram(), forceRegenerate)
}

type ApexSolanaWallet struct {
	BridgeWallet *solanawallet.Wallet
	FeeWallet    *solanawallet.Wallet
}

func GenerateAndStoreBatcherSolanaPrivateKey(
	mngr secrets.SecretsManager, chain string, forceRegenerate bool,
) (*solana.PrivateKey, error) {
	keyName := fmt.Sprintf("%s%s_key", secrets.OtherKeyLocalPrefix, chain)

	if mngr.HasSecret(keyName) {
		if !forceRegenerate {
			return LoadBatcherSolanaPrivateKey(mngr, chain)
		}

		if err := mngr.RemoveSecret(keyName); err != nil {
			return nil, err
		}
	}

	privateKey, err := solana.NewRandomPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate wallet: %w", err)
	}

	bytes, err := json.Marshal(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal wallet: %w", err)
	}

	if err := mngr.SetSecret(keyName, bytes); err != nil {
		return nil, fmt.Errorf("failed to store wallet: %w", err)
	}

	return &privateKey, err
}

func LoadBatcherSolanaPrivateKey(mngr secrets.SecretsManager, chain string) (*solana.PrivateKey, error) {
	keyName := fmt.Sprintf("%s%s_key", secrets.OtherKeyLocalPrefix, chain)

	bytes, err := mngr.GetSecret(keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to load wallet: %w", err)
	}

	var solanaPrivateKey *solana.PrivateKey

	if err := json.Unmarshal(bytes, &solanaPrivateKey); err != nil {
		return nil, fmt.Errorf("failed to load wallet: %w", err)
	}

	return solanaPrivateKey, nil
}

func GenerateAndStoreRelayerSolanaPrivateKey(
	mngr secrets.SecretsManager, chain string, forceRegenerate bool,
) (*solana.PrivateKey, error) {
	keyName := fmt.Sprintf("%s%s_relayer_solana_key", secrets.OtherKeyLocalPrefix, chain)

	if mngr.HasSecret(keyName) {
		if !forceRegenerate {
			return LoadRelayerSolanaPrivateKey(mngr, chain)
		}

		if err := mngr.RemoveSecret(keyName); err != nil {
			return nil, err
		}
	}

	privateKey, err := solana.NewRandomPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate wallet: %w", err)
	}

	bytes, err := json.Marshal(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal wallet: %w", err)
	}

	if err := mngr.SetSecret(keyName, bytes); err != nil {
		return nil, fmt.Errorf("failed to store wallet: %w", err)
	}

	return &privateKey, err
}

func LoadRelayerSolanaPrivateKey(mngr secrets.SecretsManager, chain string) (*solana.PrivateKey, error) {
	keyName := fmt.Sprintf("%s%s_relayer_solana_key", secrets.OtherKeyLocalPrefix, chain)

	bytes, err := mngr.GetSecret(keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to load wallet: %w", err)
	}

	var solanaPrivateKey *solana.PrivateKey

	if err := json.Unmarshal(bytes, &solanaPrivateKey); err != nil {
		return nil, fmt.Errorf("failed to load wallet: %w", err)
	}

	return solanaPrivateKey, nil
}
