package eth

import (
	"encoding/hex"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/bn256"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	cardanoWallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/Ethernal-Tech/ethgo"
	"github.com/hashicorp/go-hclog"
)

var (
	BN256Domain, _ = common.Keccak256([]byte("DOMAIN_APEX_BRIDGE_EVM"))
)

func GetBatcherEVMPrivateKey(secretsManager secrets.SecretsManager, chain string) (*bn256.PrivateKey, error) {
	keyName := fmt.Sprintf("%s%s_batcher_evm_key", secrets.OtherKeyLocalPrefix, chain)

	pkBytes, err := secretsManager.GetSecret(keyName)
	if err != nil {
		return nil, err
	}

	return bn256.UnmarshalPrivateKey(pkBytes)
}

func CreateAndSaveBatcherEVMPrivateKey(
	secretsManager secrets.SecretsManager, chain string, forceRegenerate bool,
) (*bn256.PrivateKey, error) {
	keyName := fmt.Sprintf("%s%s_batcher_evm_key", secrets.OtherKeyLocalPrefix, chain)

	if secretsManager.HasSecret(keyName) {
		if !forceRegenerate {
			return GetBatcherEVMPrivateKey(secretsManager, chain)
		}

		if err := secretsManager.RemoveSecret(keyName); err != nil {
			return nil, err
		}
	}

	privateKey, err := bn256.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}

	bytes, err := privateKey.Marshal()
	if err != nil {
		return nil, err
	}

	return privateKey, secretsManager.SetSecret(keyName, bytes)
}

func GetRelayerEVMPrivateKey(secretsManager secrets.SecretsManager, chain string) (*ethtxhelper.EthTxWallet, error) {
	keyName := fmt.Sprintf("%s%s_relayer_evm_key", secrets.OtherKeyLocalPrefix, chain)

	pkBytes, err := secretsManager.GetSecret(keyName)
	if err != nil {
		return nil, err
	}

	return ethtxhelper.NewEthTxWallet(string(pkBytes))
}

func CreateAndSaveRelayerEVMPrivateKey(
	secretsManager secrets.SecretsManager, chain string, forceRegenerate bool,
) (*ethtxhelper.EthTxWallet, error) {
	keyName := fmt.Sprintf("%s%s_relayer_evm_key", secrets.OtherKeyLocalPrefix, chain)

	if secretsManager.HasSecret(keyName) {
		if !forceRegenerate {
			return GetRelayerEVMPrivateKey(secretsManager, chain)
		}

		if err := secretsManager.RemoveSecret(keyName); err != nil {
			return nil, err
		}
	}

	ethWallet, err := ethtxhelper.GenerateNewEthTxWallet()
	if err != nil {
		return nil, err
	}

	return ethWallet, ethWallet.Save(secretsManager, keyName)
}

func GetEventSignatures(events []string) ([]ethgo.Hash, error) {
	abi, err := contractbinding.GatewayMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	hashes := make([]ethgo.Hash, len(events))
	for i, ev := range events {
		hashes[i] = ethgo.Hash(abi.Events[ev].ID)
	}

	return hashes, nil
}

func GetNexusEventSignatures() ([]ethgo.Hash, error) {
	return GetEventSignatures([]string{"Deposit", "Withdraw"})
}

func GetPolicyScripts(
	validatorsData []ValidatorChainData, logger hclog.Logger,
) (*cardanoWallet.PolicyScript, *cardanoWallet.PolicyScript, error) {
	multisigKeyHashes := make([]string, len(validatorsData))
	multisigFeeKeyHashes := make([]string, len(validatorsData))

	for i, x := range validatorsData {
		keyHash, err := cardanoWallet.GetKeyHash(x.Key[0].Bytes())
		if err != nil {
			return nil, nil, err
		}

		keyHashFee, err := cardanoWallet.GetKeyHash(x.Key[1].Bytes())
		if err != nil {
			return nil, nil, err
		}

		multisigKeyHashes[i] = keyHash
		multisigFeeKeyHashes[i] = keyHashFee
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
	multisigPolicyScript := cardanoWallet.NewPolicyScript(multisigKeyHashes, atLeastSignersCount)
	multisigFeePolicyScript := cardanoWallet.NewPolicyScript(multisigFeeKeyHashes, atLeastSignersCount)

	return multisigPolicyScript, multisigFeePolicyScript, nil
}

func GetMultisigAddresses(
	cardanoCliBinary string, networkMagic uint,
	multisigPolicyScript, multisigFeePolicyScript *cardanoWallet.PolicyScript,
) (string, string, error) {
	cliUtils := cardanoWallet.NewCliUtils(cardanoCliBinary)

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
