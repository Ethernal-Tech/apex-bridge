package solanatx

import (
	"encoding/json"
	"fmt"

	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	solanawallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	"github.com/gagliardetto/solana-go"
)

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
