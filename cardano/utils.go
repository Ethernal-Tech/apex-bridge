package cardanotx

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

func GetPolicyScripts(
	validatorsData []eth.ValidatorChainData,
) (multisigPolicyScript *wallet.PolicyScript, feePolicyScript *wallet.PolicyScript, err error) {
	multisigKeyHashes := make([]string, len(validatorsData))
	multisigFeeKeyHashes := make([]string, len(validatorsData))

	for i, x := range validatorsData {
		multisigKeyHashes[i], err = wallet.GetKeyHash(
			wallet.PadKeyToSize(x.Key[0].Bytes()))
		if err != nil {
			return nil, nil, err
		}

		multisigFeeKeyHashes[i], err = wallet.GetKeyHash(
			wallet.PadKeyToSize(x.Key[1].Bytes()))
		if err != nil {
			return nil, nil, err
		}
	}

	atLeastSignersCount := int(common.GetRequiredSignaturesForConsensus(uint64(len(validatorsData))))
	multisigPolicyScript = wallet.NewPolicyScript(multisigKeyHashes, atLeastSignersCount)
	feePolicyScript = wallet.NewPolicyScript(multisigFeeKeyHashes, atLeastSignersCount)

	return multisigPolicyScript, feePolicyScript, nil
}

func GetMultisigAddresses(
	cardanoCliBinary string, networkMagic uint,
	multisigPolicyScript, multisigFeePolicyScript *wallet.PolicyScript,
) (string, string, error) {
	cliUtils := wallet.NewCliUtils(cardanoCliBinary)

	multisigAddress, err := cliUtils.GetPolicyScriptAddress(networkMagic, multisigPolicyScript)
	if err != nil {
		return "", "", err
	}

	multisigFeeAddress, err := cliUtils.GetPolicyScriptAddress(networkMagic, multisigFeePolicyScript)
	if err != nil {
		return "", "", err
	}

	return multisigAddress, multisigFeeAddress, nil
}

func IsValidOutputAddress(addr string, networkID wallet.CardanoNetworkType) bool {
	cardAddr, err := wallet.NewAddress(addr)
	if err != nil {
		return false
	}

	return cardAddr.GetInfo().AddressType != wallet.RewardAddress && cardAddr.GetInfo().Network == networkID
}
