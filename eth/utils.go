package eth

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/bn256"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/Ethernal-Tech/ethgo"
)

const (
	ForceBladeBlsKey = true
)

var (
	BN256Domain, _ = common.Keccak256([]byte("DOMAIN_APEX_BRIDGE_EVM"))
)

func GetBatcherEVMPrivateKey(secretsManager secrets.SecretsManager, chainID string) (*bn256.PrivateKey, error) {
	pkBytes, err := secretsManager.GetSecret(getBLSKeyName(chainID))
	if err != nil {
		return nil, err
	}

	return bn256.UnmarshalPrivateKey(pkBytes)
}

func CreateAndSaveBatcherEVMPrivateKey(
	secretsManager secrets.SecretsManager, chainID string, forceRegenerate bool,
) (*bn256.PrivateKey, error) {
	keyName := getBLSKeyName(chainID)

	if secretsManager.HasSecret(keyName) {
		if !forceRegenerate {
			return GetBatcherEVMPrivateKey(secretsManager, chainID)
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

func GetRelayerEVMPrivateKey(secretsManager secrets.SecretsManager, chainID string) (*ethtxhelper.EthTxWallet, error) {
	keyName := fmt.Sprintf("%s%s_relayer_evm_key", secrets.OtherKeyLocalPrefix, chainID)

	pkBytes, err := secretsManager.GetSecret(keyName)
	if err != nil {
		return nil, err
	}

	return ethtxhelper.NewEthTxWallet(string(pkBytes))
}

func CreateAndSaveRelayerEVMPrivateKey(
	secretsManager secrets.SecretsManager, chainID string, forceRegenerate bool,
) (*ethtxhelper.EthTxWallet, error) {
	keyName := fmt.Sprintf("%s%s_relayer_evm_key", secrets.OtherKeyLocalPrefix, chainID)

	if secretsManager.HasSecret(keyName) {
		if !forceRegenerate {
			return GetRelayerEVMPrivateKey(secretsManager, chainID)
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

func GetChainValidatorsDataInfoString(
	chainID string, data []ValidatorChainData,
) string {
	var sb strings.Builder

	for i, x := range data {
		if i > 0 {
			sb.WriteString(", ")
		}

		switch chainID {
		case common.ChainIDStrNexus:
			pub, err := bn256.UnmarshalPublicKeyFromBigInt(x.Key)
			if err != nil {
				return fmt.Sprintf("failed to unmarshal bls key for %s, error: %s", chainID, err)
			}

			sb.WriteString(hex.EncodeToString(pub.Marshal()))
		default:
			sb.WriteRune('(')
			sb.WriteString(hex.EncodeToString(wallet.PadKeyToSize(x.Key[0].Bytes())))
			sb.WriteRune(',')
			sb.WriteString(hex.EncodeToString(wallet.PadKeyToSize(x.Key[1].Bytes())))
			sb.WriteRune(')')
		}
	}

	return sb.String()
}

func getBLSKeyName(chainID string) string {
	if ForceBladeBlsKey {
		return secrets.ValidatorBLSKey
	}

	return fmt.Sprintf("%s%s_batcher_evm_key", secrets.OtherKeyLocalPrefix, chainID)
}
