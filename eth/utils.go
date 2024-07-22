package eth

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/bn256"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
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
