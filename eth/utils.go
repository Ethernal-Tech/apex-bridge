package eth

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/bn256"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
)

var (
	BN256Domain, _ = common.Keccak256([]byte("DOMAIN_APEX_BRIDGE_EVM"))
)

func GetValidatorPrivateKey(secretsManager secrets.SecretsManager, _ string) (*bn256.PrivateKey, error) {
	pkBytes, err := secretsManager.GetSecret(secrets.ValidatorBLSKey)
	if err != nil {
		return nil, err
	}

	return bn256.UnmarshalPrivateKey(pkBytes)
}
