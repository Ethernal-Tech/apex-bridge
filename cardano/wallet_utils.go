package cardanotx

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type KeyHashesContainer struct {
	Payment []string
	Stake   []string
}

func (k KeyHashesContainer) String() string {
	return fmt.Sprintf("(payment=%s, stake=%s)", strings.Join(k.Payment, ","), strings.Join(k.Stake, ","))
}

type PolicyScriptsContainer struct {
	Payment *wallet.PolicyScript
	Stake   *wallet.PolicyScript
}

type AddressContainer struct {
	Payment string
	Stake   string
}

type ApexKeyHashes struct {
	Multisig KeyHashesContainer
	Fee      KeyHashesContainer
}

func (k ApexKeyHashes) String() string {
	return fmt.Sprintf("multisig=%s, fee=%s", k.Multisig, k.Fee)
}

type ApexPolicyScripts struct {
	Multisig PolicyScriptsContainer
	Fee      PolicyScriptsContainer
}

type ApexAddresses struct {
	Multisig AddressContainer
	Fee      AddressContainer
}

func NewApexKeyHashes(
	validatorsData []eth.ValidatorChainData,
) (ApexKeyHashes, error) {
	multisig, err := getKeyHashes(validatorsData, false)
	if err != nil {
		return ApexKeyHashes{}, fmt.Errorf("failed to create key hashes for multisig: %w", err)
	}

	fee, err := getKeyHashes(validatorsData, true)
	if err != nil {
		return ApexKeyHashes{}, fmt.Errorf("failed to create key hashes for fee: %w", err)
	}

	return ApexKeyHashes{
		Multisig: multisig,
		Fee:      fee,
	}, nil
}

func NewPolicyScriptsContainer(keyHashes KeyHashesContainer) PolicyScriptsContainer {
	//nolint:gosec
	quorumCount := int(common.GetRequiredSignaturesForConsensus(uint64(len(keyHashes.Payment))))
	//  if needed create policy script for payment only
	if len(keyHashes.Stake) == 0 {
		return PolicyScriptsContainer{
			Payment: wallet.NewPolicyScript(keyHashes.Payment, quorumCount),
		}
	}

	return PolicyScriptsContainer{
		Payment: wallet.NewPolicyScript(keyHashes.Payment, quorumCount),
		Stake:   wallet.NewPolicyScript(keyHashes.Stake, quorumCount),
	}
}

func NewApexPolicyScripts(keyHashes ApexKeyHashes) ApexPolicyScripts {
	return ApexPolicyScripts{
		Multisig: NewPolicyScriptsContainer(keyHashes.Multisig),
		Fee:      NewPolicyScriptsContainer(keyHashes.Fee),
	}
}

func NewAddressContainer(
	cardanoCliBinary string, networkMagic uint, policyScripts PolicyScriptsContainer,
) (addr AddressContainer, err error) {
	cliUtils := wallet.NewCliUtils(cardanoCliBinary)

	if policyScripts.Stake == nil {
		addr.Payment, err = cliUtils.GetPolicyScriptEnterpriseAddress(networkMagic, policyScripts.Payment)
		if err != nil {
			return addr, fmt.Errorf("payment address: %w", err)
		}

		return addr, nil
	}

	addr.Stake, err = cliUtils.GetPolicyScriptRewardAddress(networkMagic, policyScripts.Stake)
	if err != nil {
		return addr, fmt.Errorf("stake address: %w", err)
	}

	addr.Payment, err = cliUtils.GetPolicyScriptBaseAddress(
		networkMagic, policyScripts.Payment, policyScripts.Stake)
	if err != nil {
		return addr, fmt.Errorf("payment address: %w", err)
	}

	return addr, nil
}

func NewApexAddresses(
	cardanoCliBinary string, networkMagic uint, policyScripts ApexPolicyScripts,
) (ApexAddresses, error) {
	multisig, err := NewAddressContainer(cardanoCliBinary, networkMagic, policyScripts.Multisig)
	if err != nil {
		return ApexAddresses{}, fmt.Errorf("failed to create address for multisig: %w", err)
	}

	fee, err := NewAddressContainer(cardanoCliBinary, networkMagic, policyScripts.Fee)
	if err != nil {
		return ApexAddresses{}, fmt.Errorf("failed to create address for fee: %w", err)
	}

	return ApexAddresses{
		Multisig: multisig,
		Fee:      fee,
	}, nil
}

func AreVerifyingKeysTheSame(w *ApexCardanoWallet, data eth.ValidatorChainData) bool {
	return bytes.Equal(w.MultiSig.VerificationKey, bigIntToKey(data.Key[0])) &&
		bytes.Equal(w.Fee.VerificationKey, bigIntToKey(data.Key[1])) &&
		bytes.Equal(w.MultiSig.StakeVerificationKey, bigIntToKey(data.Key[2])) &&
		bytes.Equal(w.Fee.StakeVerificationKey, bigIntToKey(data.Key[3]))
}

func getKeyHashes(validatorsData []eth.ValidatorChainData, isFee bool) (KeyHashesContainer, error) {
	paymentKeyHashes := make([]string, len(validatorsData))
	stakeKeyHashes := make([]string, len(validatorsData))
	quorumCount := int(common.GetRequiredSignaturesForConsensus(uint64(len(validatorsData)))) //nolint:gosec
	countWithStake := 0

	indx := 0
	if isFee {
		indx = 1
	}

	for i, x := range validatorsData {
		payment, stake, err := getKeyHashPair(x.Key[indx], x.Key[indx+2])
		if err != nil {
			return KeyHashesContainer{}, err
		}

		paymentKeyHashes[i] = payment
		stakeKeyHashes[i] = stake

		if stake != "" {
			countWithStake++
		}
	}
	// if less than quorum num of validators sent stake verifying key, do not create stake hashes
	if countWithStake < quorumCount {
		stakeKeyHashes = nil
	}

	return KeyHashesContainer{
		Payment: paymentKeyHashes,
		Stake:   stakeKeyHashes,
	}, nil
}

func getKeyHashPair(paymentVerificationKey, stakeVerificationKey *big.Int) (string, string, error) {
	paymentKeyBytes := bigIntToKey(paymentVerificationKey)
	stakeKeyBytes := bigIntToKey(stakeVerificationKey)

	keyHash, err := wallet.GetKeyHash(paymentKeyBytes)
	if err != nil {
		return "", "", err
	}
	// do no generate stake key hash if not needed
	if len(stakeKeyBytes) == 0 {
		return keyHash, "", nil
	}

	stakeKeyHash, err := wallet.GetKeyHash(stakeKeyBytes)
	if err != nil {
		return "", "", err
	}

	return keyHash, stakeKeyHash, nil
}

func bigIntToKey(a *big.Int) []byte {
	if a == nil || a.BitLen() == 0 {
		return nil
	}

	return wallet.PadKeyToSize(a.Bytes())
}
