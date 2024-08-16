package cardanotx

import (
	"encoding/hex"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

func GetPolicyScripts(
	validatorsData []eth.ValidatorChainData, logger hclog.Logger,
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

	if logger != nil {
		pubKeys := make([]string, len(validatorsData))
		feePubKeys := make([]string, len(validatorsData))

		for i, x := range validatorsData {
			pubKeys[i] = hex.EncodeToString(x.Key[0].Bytes())
			feePubKeys[i] = hex.EncodeToString(x.Key[1].Bytes())
		}

		logger.Debug("Validator public keys/hashes multisig", "pubs", pubKeys, "hashes", multisigKeyHashes)
		logger.Debug("Validator public keys/hashes fee", "pubs", feePubKeys, "hashes", multisigFeeKeyHashes)
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
