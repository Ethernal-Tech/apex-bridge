package batchermanager

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestBatcherManagerCreation(t *testing.T) {
	testDir, err := os.MkdirTemp("", "cardano-prime")
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	secretsPath := filepath.Join(testDir, "stp")

	secretsMngr, err := common.GetSecretsManager(secretsPath, "", true)
	require.NoError(t, err)

	err = secretsMngr.SetSecret(secrets.ValidatorKey, []byte("6a9d5cf2d80878afcd6c268fc4972f23eab59ac258435d8c9ac5790b5e15da6d"))
	require.NoError(t, err)

	_, err = cardanotx.GenerateWallet(secretsMngr, "prime", true, true)
	require.NoError(t, err)

	t.Run("creation fails - secrets manager", func(t *testing.T) {
		invalidConfig := &core.BatcherManagerConfiguration{
			Chains: []core.ChainConfig{
				{
					ChainID:       common.ChainIDStrPrime,
					ChainType:     "Cardano",
					ChainSpecific: json.RawMessage(""),
				},
			},
		}

		_, err := NewBatcherManager(context.Background(),
			invalidConfig, nil, &common.BridgingRequestStateUpdaterMock{ReturnNil: true}, hclog.NewNullLogger())
		require.ErrorContains(t, err, "failed to create secrets manager")
	})

	t.Run("creation fails - invalid operations", func(t *testing.T) {
		invalidConfig := &core.BatcherManagerConfiguration{
			ValidatorDataDir: secretsPath,
			Chains: []core.ChainConfig{
				{
					ChainID:       common.ChainIDStrPrime,
					ChainType:     "Cardano",
					ChainSpecific: json.RawMessage(""),
				},
			},
		}

		_, err := NewBatcherManager(context.Background(),
			invalidConfig, map[string]indexer.Database{
				common.ChainIDStrPrime: &indexer.DatabaseMock{},
			}, &common.BridgingRequestStateUpdaterMock{ReturnNil: true}, hclog.NewNullLogger())
		require.ErrorContains(t, err, "failed to unmarshal Cardano configuration")
	})

	t.Run("creation fails - database for chain not exists", func(t *testing.T) {
		invalidConfig := &core.BatcherManagerConfiguration{
			ValidatorDataDir: secretsPath,
			Chains: []core.ChainConfig{
				{
					ChainID:       common.ChainIDStrPrime,
					ChainType:     "Cardano",
					ChainSpecific: json.RawMessage(""),
				},
			},
		}

		_, err := NewBatcherManager(context.Background(),
			invalidConfig, map[string]indexer.Database{}, &common.BridgingRequestStateUpdaterMock{ReturnNil: true}, hclog.NewNullLogger())
		require.ErrorContains(t, err, "database not exists")
	})

	t.Run("pass", func(t *testing.T) {
		invalidConfig := &core.BatcherManagerConfiguration{
			ValidatorDataDir: secretsPath,
			Chains: []core.ChainConfig{
				{
					ChainID:       common.ChainIDStrPrime,
					ChainType:     "Cardano",
					ChainSpecific: json.RawMessage([]byte(`{ "testnetMagic": 2, "socketPath": "./" }`)),
				},
			},
		}

		_, err := NewBatcherManager(context.Background(),
			invalidConfig, map[string]indexer.Database{
				common.ChainIDStrPrime: &indexer.DatabaseMock{},
			}, &common.BridgingRequestStateUpdaterMock{ReturnNil: true}, hclog.NewNullLogger())
		require.NoError(t, err)
	})
}
