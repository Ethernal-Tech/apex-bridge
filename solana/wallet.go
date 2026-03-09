package solanatx

import (
	"encoding/json"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	solanawallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
)

type ApexSolanaWallet struct {
	BridgeWallet *solanawallet.Wallet
	FeeWallet    *solanawallet.Wallet
}

func GenerateWallet(
	mngr secrets.SecretsManager, chain string, forceRegenerate bool,
) (*ApexSolanaWallet, error) {
	keyName := fmt.Sprintf("%s%s_key", secrets.OtherKeyLocalPrefix, chain)

	if mngr.HasSecret(keyName) {
		if !forceRegenerate {
			return LoadWallet(mngr, chain)
		}

		if err := mngr.RemoveSecret(keyName); err != nil {
			return nil, err
		}
	}

	bridgeWallet, err := solanawallet.NewWallet()
	if err != nil {
		return nil, fmt.Errorf("failed to generate bridging wallet: %w", err)
	}

	feeWallet, err := solanawallet.NewWallet()
	if err != nil {
		return nil, fmt.Errorf("failed to generate fee wallet: %w", err)
	}

	solanaWallet := &ApexSolanaWallet{
		BridgeWallet: bridgeWallet,
		FeeWallet:    feeWallet,
	}

	bytes, err := json.Marshal(solanaWallet)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal wallet: %w", err)
	}

	if err := mngr.SetSecret(keyName, bytes); err != nil {
		return nil, fmt.Errorf("failed to store wallet: %w", err)
	}

	return solanaWallet, err
}

func LoadWallet(mngr secrets.SecretsManager, chain string) (*ApexSolanaWallet, error) {
	keyName := fmt.Sprintf("%s%s_key", secrets.OtherKeyLocalPrefix, chain)

	bytes, err := mngr.GetSecret(keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to load wallet: %w", err)
	}

	var solanaWallet *ApexSolanaWallet

	if err := json.Unmarshal(bytes, &solanaWallet); err != nil {
		return nil, fmt.Errorf("failed to load wallet: %w", err)
	}

	return solanaWallet, nil
}

func StoreSolanaProgramKeyPair(
	mngr secrets.SecretsManager, solanaProgramKeyPair solanawallet.Wallet, forceRegenerate bool,
) (*solanawallet.Wallet, error) {
	keyName := fmt.Sprintf("%s%s_program_key", secrets.OtherKeyLocalPrefix, common.ChainIDStrSolana)

	if mngr.HasSecret(keyName) {
		if !forceRegenerate {
			return LoadSolanaProgramKeyPair(mngr)
		}

		if err := mngr.RemoveSecret(keyName); err != nil {
			return nil, err
		}
	}

	bytes, err := json.Marshal(solanaProgramKeyPair)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal solana program key pair: %w", err)
	}

	if err := mngr.SetSecret(keyName, bytes); err != nil {
		return nil, fmt.Errorf("failed to store solana program key pair: %w", err)
	}

	return &solanaProgramKeyPair, nil
}

func LoadSolanaProgramKeyPair(mngr secrets.SecretsManager) (*solanawallet.Wallet, error) {
	keyName := fmt.Sprintf("%s%s_program_key", secrets.OtherKeyLocalPrefix, common.ChainIDStrSolana)

	bytes, err := mngr.GetSecret(keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to load wallet: %w", err)
	}

	var solanaKeyPair *solanawallet.Wallet

	if err := json.Unmarshal(bytes, &solanaKeyPair); err != nil {
		return nil, fmt.Errorf("failed to load solana program key pair: %w", err)
	}

	return solanaKeyPair, nil
}
