package batcher_manager

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/batcher/batcher"
	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	infraCommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
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

	jsonData := []byte(fmt.Sprintf(`{
		"socketPath": "./socket",
		"testnetMagic": 2,
		"atLeastValidators": 0.6666666666666666,
		"potentialFee": 300000,
		"keysDirPath": "%s"
		}`, testDir))

	config := &core.BatcherManagerConfiguration{
		Chains: []core.ChainConfig{
			{
				ChainId:       "prime",
				ChainType:     "Cardano",
				ChainSpecific: json.RawMessage(jsonData),
			},
		},
		Bridge:        core.BridgeConfig{},
		PullTimeMilis: 2500,
	}

	for _, chainConfig := range config.Chains {
		wallet, err := cardanotx.GenerateWallet(testDir, false, true)
		require.NoError(t, err)

		chainOp, err := batcher.GetChainSpecificOperations(chainConfig)
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

	_, err = cardanotx.GenerateWallet(testDir, false, true)
	require.NoError(t, err)

	ecdsaValidatoSecretDirPath := path.Join(testDir, secrets.ConsensusFolderLocal)
	require.NoError(t, infraCommon.CreateDirSafe(ecdsaValidatoSecretDirPath, 0770))

	ecdsaValidatoSecretFilePath := path.Join(ecdsaValidatoSecretDirPath, secrets.ValidatorKeyLocal)
	require.NoError(t, os.WriteFile(ecdsaValidatoSecretFilePath, []byte(
		"6a9d5cf2d80878afcd6c268fc4972f23eab59ac258435d8c9ac5790b5e15da6d",
	), 0770))

	t.Run("creation fails - invalid operations", func(t *testing.T) {
		invalidConfig := &core.BatcherManagerConfiguration{
			Chains: []core.ChainConfig{
				{
					ChainId:       "prime",
					ChainType:     "Cardano",
					ChainSpecific: json.RawMessage(""),
				},
			},
		}

		_, err := NewBatcherManager(context.Background(), invalidConfig, &common.BridgingRequestStateUpdaterMock{ReturnNil: true}, hclog.NewNullLogger())
		require.Error(t, err)
	})
}
