package relayer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/batcher/batcher"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/bn256"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	secretsHelper "github.com/Ethernal-Tech/cardano-infrastructure/secrets/helper"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEVMChainOperations(t *testing.T) {
	const chainID = "nexus"

	testDir, err := os.MkdirTemp("", "relayer-www")
	require.NoError(t, err)

	secretsDir := filepath.Join(testDir, "stp")

	defer os.RemoveAll(testDir)

	secretsMngr, err := secretsHelper.CreateSecretsManager(&secrets.SecretsManagerConfig{
		Path: secretsDir,
		Type: secrets.Local,
	})
	require.NoError(t, err)

	t.Run("NewEVMChainOperations", func(t *testing.T) {
		const (
			nodeURL = "localhost:5000"
		)

		_, err = eth.CreateAndSaveRelayerEVMPrivateKey(secretsMngr, chainID, true)
		require.NoError(t, err)

		configRaw := json.RawMessage([]byte(fmt.Sprintf(`{
			"dataDir": "%s",
			"nodeUrl": "%s"
		}`, secretsDir, nodeURL)))

		ops, err := NewEVMChainOperations(configRaw, chainID, "0x0ff", hclog.NewNullLogger())
		require.NoError(t, err)

		require.Equal(t, chainID, ops.chainID)
		require.Equal(t, nodeURL, ops.config.NodeURL)
	})

	t.Run("SendTx", func(t *testing.T) {
		ctx := context.Background()
		scMock := &eth.EVMGatewaySmartContractMock{}
		batch := &eth.ConfirmedBatch{
			RawTransaction: []byte{1, 2, 3},
			Bitmap:         new(big.Int).SetBytes([]byte{1, 7, 4}),
		}
		domain := []byte("domain")
		message := [32]byte{1, 2, 89, 100, 245, 78, 3, 0, 8}

		pk1, err := eth.CreateAndSaveBatcherEVMPrivateKey(secretsMngr, chainID, true)
		require.NoError(t, err)

		pk2, err := eth.CreateAndSaveBatcherEVMPrivateKey(secretsMngr, chainID, true)
		require.NoError(t, err)

		signature1, err := pk1.Sign(message[:], domain)
		require.NoError(t, err)

		signature2, err := pk2.Sign(message[:], domain)
		require.NoError(t, err)

		sigBytes1, err := signature1.Marshal()
		require.NoError(t, err)

		sigBytes2, err := signature2.Marshal()
		require.NoError(t, err)

		batch.Signatures = [][]byte{
			sigBytes1, sigBytes2,
		}

		finalSigBytes, err := bn256.Signatures{signature1, signature2}.Aggregate().Marshal()
		require.NoError(t, err)

		scMock.On("Deposit", ctx, finalSigBytes, batch.Bitmap, batch.RawTransaction).Return(errors.New("hello")).Once()
		scMock.On("Deposit", ctx, finalSigBytes, batch.Bitmap, batch.RawTransaction).Return(error(nil)).Once()

		ops := &EVMChainOperations{
			evmSmartContract: scMock,
			logger:           hclog.NewNullLogger(),
		}

		require.Error(t, ops.SendTx(ctx, nil, batch))
		require.NoError(t, ops.SendTx(ctx, nil, batch))

		scMock.AssertExpectations(t)
	})

	t.Run("SendTx - behavior for each batch type", func(t *testing.T) {
		createFn := func() (*eth.EVMGatewaySmartContractMock, *EVMChainOperations) {
			gateway := &eth.EVMGatewaySmartContractMock{}
			gateway.On("Deposit",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything).Return(nil)
			gateway.On("UpdateValidatorsChainData",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything).Return(nil)

			op := &EVMChainOperations{
				evmSmartContract: gateway,
				logger:           hclog.NewNullLogger(),
			}

			return gateway, op
		}

		t.Run("Normal batch", func(t *testing.T) {
			gateway, op := createFn()

			_ = op.SendTx(nil, nil, &eth.ConfirmedBatch{
				BatchType: uint8(batcher.Normal),
			})

			gateway.AssertCalled(t, "Deposit",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything)
		})

		t.Run("Validator set batch", func(t *testing.T) {
			gateway, op := createFn()

			_ = op.SendTx(nil, nil, &eth.ConfirmedBatch{
				BatchType: uint8(batcher.ValidatorSet),
			})

			gateway.AssertCalled(t, "UpdateValidatorsChainData",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything)
		})

		t.Run("Validator set batch", func(t *testing.T) {
			gateway, op := createFn()

			err := op.SendTx(nil, nil, &eth.ConfirmedBatch{
				BatchType: uint8(batcher.ValidatorSetFinal),
			})

			require.NoError(t, err)

			gateway.AssertNotCalled(t, "Deposit",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything)
			gateway.AssertNotCalled(t, "UpdateValidatorsChainData",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything)
		})
	})
}
