package common

import (
	"errors"
	"fmt"

	secretsInfra "github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	secretsInfraHelper "github.com/Ethernal-Tech/cardano-infrastructure/secrets/helper"
	secretsInfraLocal "github.com/Ethernal-Tech/cardano-infrastructure/secrets/local"
)

// GetSecretsManager function resolves secrets manager instance based on provided data or config paths.
// insecureLocalStore defines if utilization of local secrets manager is allowed.
func GetSecretsManager(
	dataPath, configPath string, insecureLocalStore bool,
) (secretsInfra.SecretsManager, error) {
	if configPath != "" {
		secretsConfig, err := secretsInfra.ReadConfig(configPath)
		if err != nil {
			return nil, fmt.Errorf("invalid secrets configuration: %w", err)
		}

		return secretsInfraHelper.CreateSecretsManager(secretsConfig)
	}

	// Storing secrets on a local file system should only be allowed with --insecure flag,
	// to raise awareness that it should be only used in development/testing environments.
	// Production setups should use one of the supported secrets managers
	if !insecureLocalStore {
		return nil, errors.New("insecure local storage not supported")
	}

	return secretsInfraLocal.SecretsManagerFactory(&secretsInfra.SecretsManagerConfig{
		Path: dataPath,
	})
}
