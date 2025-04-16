package cardanotx

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type CardanoWallet struct {
	MultiSig    *cardanowallet.Wallet `json:"multisig"`
	MultiSigFee *cardanowallet.Wallet `json:"fee"`
}

func GenerateWallet(
	mngr secrets.SecretsManager, chain string, isStake bool, forceRegenerate bool,
) (*CardanoWallet, error) {
	keyName := fmt.Sprintf("%s%s_key", secrets.CardanoKeyLocalPrefix, chain)

	if mngr.HasSecret(keyName) {
		if !forceRegenerate {
			return LoadWallet(mngr, chain)
		}

		if err := mngr.RemoveSecret(keyName); err != nil {
			return nil, err
		}
	}

	multisigWallet, err := cardanowallet.GenerateWallet(isStake)
	if err != nil {
		return nil, fmt.Errorf("failed to generate multisig wallet: %w", err)
	}

	feeWallet, err := cardanowallet.GenerateWallet(isStake)
	if err != nil {
		return nil, fmt.Errorf("failed to generate fee wallet: %w", err)
	}

	cardanoWallet := &CardanoWallet{
		MultiSig:    multisigWallet,
		MultiSigFee: feeWallet,
	}

	bytes, err := json.Marshal(cardanoWallet)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal wallet: %w", err)
	}

	if err := mngr.SetSecret(keyName, bytes); err != nil {
		return nil, fmt.Errorf("failed to store wallet: %w", err)
	}

	return cardanoWallet, err
}

func LoadWallet(mngr secrets.SecretsManager, chain string) (*CardanoWallet, error) {
	keyName := fmt.Sprintf("%s%s_key", secrets.CardanoKeyLocalPrefix, chain)

	bytes, err := mngr.GetSecret(keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to load wallet: %w", err)
	}

	var cardanoWallet *CardanoWallet

	if err := json.Unmarshal(bytes, &cardanoWallet); err != nil {
		return nil, fmt.Errorf("failed to load wallet: %w", err)
	}

	return cardanoWallet, nil
}

func GetAddress(
	networkID cardanowallet.CardanoNetworkType, wallet *cardanowallet.Wallet,
) (*cardanowallet.CardanoAddress, error) {
	if len(wallet.StakeVerificationKey) > 0 {
		return cardanowallet.NewBaseAddress(networkID,
			wallet.VerificationKey, wallet.StakeVerificationKey)
	}

	return cardanowallet.NewEnterpriseAddress(networkID, wallet.VerificationKey)
}

func GetCardanoPrivateKeyBytes(str string) ([]byte, error) {
	bytes, err := cardanowallet.GetKeyBytes(str)
	if err != nil {
		bytes, err = hex.DecodeString(str)
		if err != nil {
			return nil, err
		}

		bytes = cardanowallet.PadKeyToSize(bytes)
	}

	return bytes, nil
}
