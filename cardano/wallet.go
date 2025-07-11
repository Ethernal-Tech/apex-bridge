package cardanotx

import (
	"encoding/json"
	"fmt"

	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type ApexCardanoWallet struct {
	MultiSig *cardanowallet.Wallet `json:"multisig"`
	Fee      *cardanowallet.Wallet `json:"fee"`
}

func GenerateWallet(
	mngr secrets.SecretsManager, chain string, isStake bool, forceRegenerate bool,
) (*ApexCardanoWallet, error) {
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

	cardanoWallet := &ApexCardanoWallet{
		MultiSig: multisigWallet,
		Fee:      feeWallet,
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

func LoadWallet(mngr secrets.SecretsManager, chain string) (*ApexCardanoWallet, error) {
	keyName := fmt.Sprintf("%s%s_key", secrets.CardanoKeyLocalPrefix, chain)

	bytes, err := mngr.GetSecret(keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to load wallet: %w", err)
	}

	var cardanoWallet *ApexCardanoWallet

	if err := json.Unmarshal(bytes, &cardanoWallet); err != nil {
		return nil, fmt.Errorf("failed to load wallet: %w", err)
	}

	return cardanoWallet, nil
}

func GetAddress(
	networkID cardanowallet.CardanoNetworkType, cardanoWallet *cardanowallet.Wallet,
) (*cardanowallet.CardanoAddress, error) {
	if len(cardanoWallet.StakeVerificationKey) > 0 {
		return cardanowallet.NewBaseAddress(networkID,
			cardanoWallet.VerificationKey, cardanoWallet.StakeVerificationKey)
	}

	return cardanowallet.NewEnterpriseAddress(networkID, cardanoWallet.VerificationKey)
}
