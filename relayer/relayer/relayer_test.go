package relayer

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	databaseaccess "github.com/Ethernal-Tech/apex-bridge/relayer/database_access"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	secretsHelper "github.com/Ethernal-Tech/cardano-infrastructure/secrets/helper"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRelayerExecute(t *testing.T) {
	relayerConfig := &core.RelayerConfiguration{
		Bridge: core.BridgeConfig{},
		Chain: core.ChainConfig{
			ChainID: common.ChainIDStrPrime,
		},
		PullTimeMilis: 1000,
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	confirmedBatchRet := &eth.ConfirmedBatch{
		ID:             0,
		RawTransaction: []byte{},
		Signatures:     [][]byte{},
		FeeSignatures:  [][]byte{},
	}

	testError := errors.New("test err")

	t.Run("execute test fail to retrieve", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &databaseaccess.CardanoChainOperationsMock{}

		bridgeSmartContractMock.On("GetConfirmedBatch", ctx, common.ChainIDStrPrime).Return(nil, testError)

		dbMock := &databaseaccess.DBMock{}

		r := NewRelayer(relayerConfig, bridgeSmartContractMock, operationsMock, dbMock, hclog.Default())
		err := r.execute(ctx)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to retrieve confirmed batch")
	})

	t.Run("execute test fail to get last submitted batch", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &databaseaccess.CardanoChainOperationsMock{}

		bridgeSmartContractMock.On("GetConfirmedBatch", ctx, common.ChainIDStrPrime).Return(confirmedBatchRet, nil)

		dbMock := &databaseaccess.DBMock{}
		dbMock.On("GetLastSubmittedBatchID", common.ChainIDStrPrime).Return(nil, testError)

		r := NewRelayer(relayerConfig, bridgeSmartContractMock, operationsMock, dbMock, hclog.Default())
		err := r.execute(ctx)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get last submitted batch id from db")
	})

	lastConfirmedBatchID := big.NewInt(1)

	confirmedBatchRet.ID = 0

	t.Run("execute test db returns nil", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &databaseaccess.CardanoChainOperationsMock{}

		bridgeSmartContractMock.On("GetConfirmedBatch", ctx, common.ChainIDStrPrime).Return(confirmedBatchRet, nil)

		dbMock := &databaseaccess.DBMock{}
		dbMock.On("GetLastSubmittedBatchID", common.ChainIDStrPrime).Return(nil, nil)

		r := NewRelayer(relayerConfig, bridgeSmartContractMock, operationsMock, dbMock, hclog.Default())
		err := r.execute(ctx)
		require.NoError(t, err)
	})

	t.Run("execute test last submitted id greater than received id", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &databaseaccess.CardanoChainOperationsMock{}

		bridgeSmartContractMock.On("GetConfirmedBatch", ctx, common.ChainIDStrPrime).Return(confirmedBatchRet, nil)

		dbMock := &databaseaccess.DBMock{}
		dbMock.On("GetLastSubmittedBatchID", common.ChainIDStrPrime).Return(lastConfirmedBatchID, nil)

		r := NewRelayer(relayerConfig, bridgeSmartContractMock, operationsMock, dbMock, hclog.Default())
		err := r.execute(ctx)
		require.Error(t, err)
		require.ErrorContains(t, err, "last submitted batch id greater than received: last submitted 1 > received 0")
	})

	confirmedBatchRet.ID = 1

	t.Run("execute test same ids", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &databaseaccess.CardanoChainOperationsMock{}

		bridgeSmartContractMock.On("GetConfirmedBatch", ctx, common.ChainIDStrPrime).Return(confirmedBatchRet, nil)

		dbMock := &databaseaccess.DBMock{}
		dbMock.On("GetLastSubmittedBatchID", common.ChainIDStrPrime).Return(lastConfirmedBatchID, nil)

		r := NewRelayer(relayerConfig, bridgeSmartContractMock, operationsMock, dbMock, hclog.Default())
		err := r.execute(ctx)
		require.NoError(t, err)
	})

	confirmedBatchRet.ID = 2

	t.Run("execute test fail to send", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &databaseaccess.CardanoChainOperationsMock{}

		bridgeSmartContractMock.On("GetConfirmedBatch", ctx, common.ChainIDStrPrime).Return(confirmedBatchRet, nil)

		dbMock := &databaseaccess.DBMock{}
		dbMock.On("GetLastSubmittedBatchID", common.ChainIDStrPrime).Return(lastConfirmedBatchID, nil)
		operationsMock.On("SendTx", ctx, bridgeSmartContractMock, confirmedBatchRet).Return(testError)

		r := NewRelayer(relayerConfig, bridgeSmartContractMock, operationsMock, dbMock, hclog.Default())
		err := r.execute(ctx)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to send confirmed batch")
	})

	t.Run("execute test test fail to add to db", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &databaseaccess.CardanoChainOperationsMock{}

		bridgeSmartContractMock.On("GetConfirmedBatch", ctx, common.ChainIDStrPrime).Return(confirmedBatchRet, nil)

		dbMock := &databaseaccess.DBMock{}
		dbMock.On("GetLastSubmittedBatchID", common.ChainIDStrPrime).Return(lastConfirmedBatchID, nil)
		operationsMock.On("SendTx", ctx, bridgeSmartContractMock, confirmedBatchRet).Return(nil)
		dbMock.On("AddLastSubmittedBatchID", common.ChainIDStrPrime, mock.Anything).Return(testError)

		r := NewRelayer(relayerConfig, bridgeSmartContractMock, operationsMock, dbMock, hclog.Default())
		err := r.execute(ctx)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to insert last submitted batch id into db")
	})

	t.Run("execute test valid", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &databaseaccess.CardanoChainOperationsMock{}

		bridgeSmartContractMock.On("GetConfirmedBatch", ctx, common.ChainIDStrPrime).Return(confirmedBatchRet, nil)

		dbMock := &databaseaccess.DBMock{}
		dbMock.On("GetLastSubmittedBatchID", common.ChainIDStrPrime).Return(lastConfirmedBatchID, nil)
		operationsMock.On("SendTx", ctx, bridgeSmartContractMock, confirmedBatchRet).Return(nil)
		dbMock.On("AddLastSubmittedBatchID", common.ChainIDStrPrime, mock.Anything).Return(nil)

		r := NewRelayer(relayerConfig, bridgeSmartContractMock, operationsMock, dbMock, hclog.Default())
		require.NoError(t, r.execute(ctx))
	})
}

func TestRelayerGetChainSpecificOperations(t *testing.T) {
	jsonData := []byte(`{
		"socketPath": "./socket",
		"testnetMagic": 2,
		"potentialFee": 300000
		}`)

	testDir, err := os.MkdirTemp("", "relayer-www")
	require.NoError(t, err)

	secretsDir := filepath.Join(testDir, "stp")

	defer os.RemoveAll(testDir)

	secretsMngr, err := secretsHelper.CreateSecretsManager(&secrets.SecretsManagerConfig{
		Path: secretsDir,
		Type: secrets.Local,
	})
	require.NoError(t, err)

	_, err = cardanotx.GenerateWallet(secretsMngr, common.ChainIDStrPrime, true, true)
	require.NoError(t, err)

	t.Run("invalid chain type", func(t *testing.T) {
		chainSpecificConfig := core.ChainConfig{
			ChainType:      "Invalid",
			ChainID:        common.ChainIDStrPrime,
			ChainSpecific:  json.RawMessage(""),
			RelayerDataDir: secretsDir,
		}

		chainOp, err := GetChainSpecificOperations(chainSpecificConfig, eth.Chain{}, hclog.NewNullLogger())
		require.Nil(t, chainOp)
		require.Error(t, err)
		require.ErrorContains(t, err, "unknown chain type")
	})

	t.Run("invalid cardano json config", func(t *testing.T) {
		chainSpecificConfig := core.ChainConfig{
			ChainType:      "Cardano",
			ChainID:        common.ChainIDStrPrime,
			ChainSpecific:  json.RawMessage(""),
			RelayerDataDir: secretsDir,
		}

		chainOp, err := GetChainSpecificOperations(chainSpecificConfig, eth.Chain{}, hclog.NewNullLogger())
		require.Nil(t, chainOp)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to unmarshal Cardano configuration")
	})

	t.Run("valid cardano config", func(t *testing.T) {
		chainSpecificConfig := core.ChainConfig{
			ChainType:      "Cardano",
			ChainID:        common.ChainIDStrPrime,
			ChainSpecific:  json.RawMessage(jsonData),
			RelayerDataDir: secretsDir,
		}

		chainOp, err := GetChainSpecificOperations(chainSpecificConfig, eth.Chain{}, hclog.NewNullLogger())
		require.NotNil(t, chainOp)
		require.NoError(t, err)
	})

	t.Run("valid cardano config check case sensitivity", func(t *testing.T) {
		chainSpecificConfig := core.ChainConfig{
			ChainType:      "CaRdAnO",
			ChainID:        common.ChainIDStrPrime,
			ChainSpecific:  json.RawMessage(jsonData),
			RelayerDataDir: secretsDir,
		}

		chainOp, err := GetChainSpecificOperations(chainSpecificConfig, eth.Chain{}, hclog.NewNullLogger())
		require.NotNil(t, chainOp)
		require.NoError(t, err)
	})
}
