package batchermanager

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/batcher/batcher"
	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	secretsHelper "github.com/Ethernal-Tech/cardano-infrastructure/secrets/helper"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatcherManagerOperations(t *testing.T) {
	testDir, err := os.MkdirTemp("", "cardano-prime")
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	jsonData := []byte(`{
		"socketPath": "./socket",
		"testnetMagic": 2,
		"potentialFee": 300000
		}`)

	config := &core.BatcherManagerConfiguration{
		Chains: []core.ChainConfig{
			{
				ChainID:       common.ChainIDStrPrime,
				ChainType:     "Cardano",
				ChainSpecific: json.RawMessage(jsonData),
			},
		},
		Bridge: core.BridgeConfig{},
	}

	secretsMngr, err := secretsHelper.CreateSecretsManager(&secrets.SecretsManagerConfig{
		Path: filepath.Join(testDir, "stp"),
		Type: secrets.Local,
	})
	require.NoError(t, err)

	for _, chainConfig := range config.Chains {
		wallet, err := cardanotx.GenerateWallet(secretsMngr, chainConfig.ChainID, true, false)
		require.NoError(t, err)

		chainOp, err := batcher.GetChainSpecificOperations(
			chainConfig, secretsMngr, hclog.NewNullLogger())
		assert.NoError(t, err)

		operationsType := reflect.TypeOf(chainOp)
		assert.NotNil(t, operationsType)

		// check keys
		concreteChainOp, ok := chainOp.(*batcher.CardanoChainOperations)
		if ok {
			// check config
			cardanoChainConfig, err := cardanotx.NewCardanoChainConfig(chainConfig.ChainSpecific)
			assert.NoError(t, err)
			assert.Equal(t, cardanoChainConfig, concreteChainOp.Config)

			// remove cbor prefix
			assert.Equal(t, wallet.MultiSig.GetSigningKey(), concreteChainOp.Wallet.MultiSig.GetSigningKey())
			assert.Equal(t, wallet.MultiSigFee.GetSigningKey(), concreteChainOp.Wallet.MultiSigFee.GetSigningKey())

			// test signatures
			sigWithString, err := cardanotx.CreateTxWitness("b335adf170a3df72dfba3864a1d09eb87d3848c98aac54d58bce1d544d1a63ea", wallet.MultiSig)
			assert.NoError(t, err)
			sigWithWallet, err := cardanotx.CreateTxWitness("b335adf170a3df72dfba3864a1d09eb87d3848c98aac54d58bce1d544d1a63ea", concreteChainOp.Wallet.MultiSig)
			assert.NoError(t, err)
			assert.Equal(t, sigWithString, sigWithWallet)
		}
	}
}

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
			invalidConfig, nil, &common.BridgingRequestStateUpdaterMock{ReturnNil: true}, hclog.NewNullLogger())
		require.ErrorContains(t, err, "failed to unmarshal Cardano configuration")
	})

	t.Run("creation fails - database for chain not exists", func(t *testing.T) {
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
