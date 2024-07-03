package batcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	secretsHelper "github.com/Ethernal-Tech/cardano-infrastructure/secrets/helper"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBatcherExecute(t *testing.T) {
	config := &core.BatcherConfiguration{
		Chain: core.ChainConfig{
			ChainID:   common.ChainIDStrPrime,
			ChainType: "Cardano",
			ChainSpecific: json.RawMessage([]byte(`{
				"socketPath": "./socket",
				"testnetMagic": 2,
				"potentialFee": 300000,
				}`)),
		},
		PullTimeMilis: 2500,
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	testError := errors.New("test err")
	batchNonceID := uint64(1)

	t.Run("GetNextBatchID returns err", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}

		bridgeSmartContractMock.On("GetNextBatchID", ctx, common.ChainIDStrPrime).Return(uint64(0), testError)

		b := NewBatcher(config, operationsMock,
			bridgeSmartContractMock, &common.BridgingRequestStateUpdaterMock{ReturnNil: true},
			hclog.NewNullLogger())
		_, err := b.execute(ctx)

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to query bridge.GetNextBatchID")
	})

	t.Run("GetNextBatchID returns 0", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}

		bridgeSmartContractMock.On("GetNextBatchID", ctx, common.ChainIDStrPrime).Return(uint64(0), nil)

		b := NewBatcher(config, operationsMock,
			bridgeSmartContractMock, &common.BridgingRequestStateUpdaterMock{ReturnNil: true}, hclog.NewNullLogger())
		batchID, err := b.execute(ctx)

		require.NoError(t, err)
		require.Equal(t, uint64(0), batchID)
	})

	t.Run("GetConfirmedTransactions returns error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}

		bridgeSmartContractMock.On("GetNextBatchID", ctx, common.ChainIDStrPrime).Return(batchNonceID, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, common.ChainIDStrPrime).Return(nil, testError)

		b := NewBatcher(config, operationsMock,
			bridgeSmartContractMock, &common.BridgingRequestStateUpdaterMock{ReturnNil: true}, hclog.NewNullLogger())
		batchID, err := b.execute(ctx)

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to query bridge.GetConfirmedTransactions")
		require.Equal(t, batchNonceID, batchID)
	})

	getConfirmedTransactionsRet := []eth.ConfirmedTransaction{
		{
			Nonce:                   5,
			ObservedTransactionHash: common.NewHashFromHexString("0x6674"),
			BlockHeight:             big.NewInt(10),
			SourceChainId:           common.ToNumChainID(common.ChainIDStrPrime),
			Receivers: []contractbinding.IBridgeStructsReceiver{
				{
					DestinationAddress: "0x333",
					Amount:             10,
				},
			},
		},
	}

	t.Run("GenerateBatchTransaction returns error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}

		bridgeSmartContractMock.On("GetNextBatchID", ctx, common.ChainIDStrPrime).Return(batchNonceID, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, common.ChainIDStrPrime).Return(getConfirmedTransactionsRet, nil)
		operationsMock.On("GenerateBatchTransaction", ctx, bridgeSmartContractMock, common.ChainIDStrPrime, getConfirmedTransactionsRet, batchNonceID).
			Return((*core.GeneratedBatchTxData)(nil), testError)

		b := NewBatcher(config, operationsMock,
			bridgeSmartContractMock, &common.BridgingRequestStateUpdaterMock{ReturnNil: true}, hclog.NewNullLogger())
		batchID, err := b.execute(ctx)

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to generate batch transaction")
		require.Equal(t, batchNonceID, batchID)
	})

	t.Run("SignBatchTransaction returns error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}

		bridgeSmartContractMock.On("GetNextBatchID", ctx, common.ChainIDStrPrime).Return(batchNonceID, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, common.ChainIDStrPrime).Return(getConfirmedTransactionsRet, nil)
		operationsMock.On("GenerateBatchTransaction", ctx, bridgeSmartContractMock, common.ChainIDStrPrime, getConfirmedTransactionsRet, batchNonceID).
			Return(&core.GeneratedBatchTxData{
				TxRaw:  []byte{0},
				TxHash: "txHash",
			}, nil)
		operationsMock.On("SignBatchTransaction", "txHash").Return(nil, nil, testError)

		b := NewBatcher(config, operationsMock,
			bridgeSmartContractMock, &common.BridgingRequestStateUpdaterMock{ReturnNil: true}, hclog.NewNullLogger())
		batchID, err := b.execute(ctx)

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to sign batch transaction")
		require.Equal(t, batchNonceID, batchID)
	})

	t.Run("SubmitSignedBatch returns error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}

		bridgeSmartContractMock.On("GetNextBatchID", ctx, common.ChainIDStrPrime).Return(batchNonceID, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, common.ChainIDStrPrime).Return(getConfirmedTransactionsRet, nil)
		operationsMock.On("GenerateBatchTransaction", ctx, bridgeSmartContractMock, common.ChainIDStrPrime, getConfirmedTransactionsRet, batchNonceID).
			Return(&core.GeneratedBatchTxData{
				TxRaw:  []byte{0},
				TxHash: "txHash",
			}, nil)
		operationsMock.On("SignBatchTransaction", "txHash").Return([]byte{}, []byte{}, nil)
		bridgeSmartContractMock.On("SubmitSignedBatch", ctx, mock.Anything).Return(testError)

		b := NewBatcher(config, operationsMock,
			bridgeSmartContractMock, &common.BridgingRequestStateUpdaterMock{ReturnNil: true}, hclog.NewNullLogger())
		batchID, err := b.execute(ctx)

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to submit signed batch")
		require.Equal(t, batchNonceID, batchID)
	})

	t.Run("execute same tx hash", func(t *testing.T) {
		const txHash = "txHash"

		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}

		bridgeSmartContractMock.On("GetNextBatchID", ctx, common.ChainIDStrPrime).Return(batchNonceID, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, common.ChainIDStrPrime).Return(getConfirmedTransactionsRet, nil)
		operationsMock.On("GenerateBatchTransaction", ctx, bridgeSmartContractMock, common.ChainIDStrPrime, getConfirmedTransactionsRet, batchNonceID).
			Return(&core.GeneratedBatchTxData{
				TxRaw:  []byte{0},
				TxHash: txHash,
			}, nil)

		b := NewBatcher(config, operationsMock,
			bridgeSmartContractMock, &common.BridgingRequestStateUpdaterMock{ReturnNil: true}, hclog.NewNullLogger())
		b.lastBatch = lastBatchData{
			id:     1,
			txHash: txHash,
		}

		batchID, err := b.execute(ctx)

		require.NoError(t, err)
		require.Equal(t, batchNonceID, batchID)
	})

	t.Run("execute pass", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}

		bridgeSmartContractMock.On("GetNextBatchID", ctx, common.ChainIDStrPrime).Return(batchNonceID, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, common.ChainIDStrPrime).Return(getConfirmedTransactionsRet, nil)
		operationsMock.On("GenerateBatchTransaction", ctx, bridgeSmartContractMock, common.ChainIDStrPrime, getConfirmedTransactionsRet, batchNonceID).
			Return(&core.GeneratedBatchTxData{
				TxRaw:  []byte{0},
				TxHash: "txHash",
			}, nil)
		operationsMock.On("SignBatchTransaction", "txHash").Return([]byte{}, []byte{}, nil)
		bridgeSmartContractMock.On("SubmitSignedBatch", ctx, mock.Anything).Return(nil)

		b := NewBatcher(config, operationsMock,
			bridgeSmartContractMock, &common.BridgingRequestStateUpdaterMock{ReturnNil: true}, hclog.NewNullLogger())
		batchID, err := b.execute(ctx)

		require.NoError(t, err)
		require.Equal(t, batchNonceID, batchID)
	})
}

func TestBatcherGetChainSpecificOperations(t *testing.T) {
	validPath, err := os.MkdirTemp("", "cardano-prime")
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(validPath)
		os.Remove(validPath)
	}()

	secretsMngr, err := secretsHelper.CreateSecretsManager(&secrets.SecretsManagerConfig{
		Path: path.Join(validPath, "stp"),
		Type: secrets.Local,
	})
	require.NoError(t, err)

	_, err = cardanotx.GenerateWallet(secretsMngr, "prime", false, true)
	require.NoError(t, err)

	chainConfig := core.ChainConfig{
		ChainID:   common.ChainIDStrPrime,
		ChainType: "Cardano",
		ChainSpecific: json.RawMessage([]byte(`{
			"socketPath": "./socket",
			"testnetMagic": 2,
			"potentialFee": 300000
			}`)),
	}

	t.Run("invalid chain type", func(t *testing.T) {
		cfg := chainConfig
		cfg.ChainType = "invalid"

		chainOp, err := GetChainSpecificOperations(cfg, &indexer.DatabaseMock{}, secretsMngr, hclog.NewNullLogger())
		require.Nil(t, chainOp)
		require.Error(t, err)
		require.ErrorContains(t, err, "unknown chain type")
	})

	t.Run("invalid cardano json config", func(t *testing.T) {
		cfg := chainConfig
		cfg.ChainSpecific = json.RawMessage("")

		chainOp, err := GetChainSpecificOperations(cfg, &indexer.DatabaseMock{}, secretsMngr, hclog.NewNullLogger())
		require.Nil(t, chainOp)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to unmarshal Cardano configuration")
	})

	t.Run("non existing cardano wallet", func(t *testing.T) {
		_, err = cardanotx.GenerateWallet(secretsMngr, "prime", false, true)
		require.NoError(t, err)

		err = secretsMngr.RemoveSecret(fmt.Sprintf("%sprime_key", secrets.CardanoKeyLocalPrefix))
		require.NoError(t, err)

		chainOp, err := GetChainSpecificOperations(chainConfig, &indexer.DatabaseMock{}, secretsMngr, hclog.NewNullLogger())
		require.Nil(t, chainOp)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to load wallet")

		_, err = cardanotx.GenerateWallet(secretsMngr, "prime", false, true)
		require.NoError(t, err)
	})

	t.Run("valid cardano config and keys path", func(t *testing.T) {
		chainOp, err := GetChainSpecificOperations(chainConfig, &indexer.DatabaseMock{}, secretsMngr, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, chainOp)
	})

	t.Run("valid cardano config check case sensitivity", func(t *testing.T) {
		defer func() {
			chainConfig.ChainType = "Cardano"
		}()

		chainConfig.ChainType = "CaRDaNo"

		chainOp, err := GetChainSpecificOperations(chainConfig, &indexer.DatabaseMock{}, secretsMngr, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, chainOp)
	})

	t.Run("getFirstAndLastTxNonceID one item", func(t *testing.T) {
		f, l := getFirstAndLastTxNonceID([]eth.ConfirmedTransaction{
			{Nonce: 5},
		})

		assert.Equal(t, uint64(5), f)
		assert.Equal(t, uint64(5), l)
	})

	t.Run("getFirstAndLastTxNonceID multiple items", func(t *testing.T) {
		f, l := getFirstAndLastTxNonceID([]eth.ConfirmedTransaction{
			{Nonce: 5}, {Nonce: 7}, {Nonce: 2}, {Nonce: 9}, {Nonce: 5},
		})

		assert.Equal(t, uint64(2), f)
		assert.Equal(t, uint64(9), l)
	})

	t.Run("getBridgingRequestStateKeys", func(t *testing.T) {
		included := [][32]byte{
			{1},
			{4},
			{5},
		}
		res := getBridgingRequestStateKeys([]eth.ConfirmedTransaction{
			{
				ObservedTransactionHash: included[0],
				SourceChainId:           common.ToNumChainID(common.ChainIDStrPrime),
				Nonce:                   4,
			},
			{
				ObservedTransactionHash: [32]byte{2},
				SourceChainId:           common.ToNumChainID(common.ChainIDStrVector),
				Nonce:                   2,
			},
			{
				ObservedTransactionHash: [32]byte{3},
				SourceChainId:           common.ToNumChainID(common.ChainIDStrPrime),
				Nonce:                   6,
			},
			{
				ObservedTransactionHash: included[1],
				SourceChainId:           common.ToNumChainID(common.ChainIDStrVector),
				Nonce:                   3,
			},
			{
				ObservedTransactionHash: included[2],
				SourceChainId:           common.ToNumChainID(common.ChainIDStrPrime),
				Nonce:                   5,
			},
			{
				ObservedTransactionHash: [32]byte{6},
				SourceChainId:           common.ToNumChainID(common.ChainIDStrVector),
				Nonce:                   2,
			},
		}, 3, 5)

		require.Equal(t, []common.BridgingRequestStateKey{
			{
				SourceChainID: common.ChainIDStrPrime,
				SourceTxHash:  included[0],
			},
			{
				SourceChainID: common.ChainIDStrVector,
				SourceTxHash:  included[1],
			},
			{
				SourceChainID: common.ChainIDStrPrime,
				SourceTxHash:  included[2],
			},
		}, res)
	})
}

type cardanoChainOperationsMock struct {
	mock.Mock
}

var _ core.ChainOperations = (*cardanoChainOperationsMock)(nil)

// GenerateBatchTransaction implements core.ChainOperations.
func (c *cardanoChainOperationsMock) GenerateBatchTransaction(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract,
	destinationChain string, confirmedTransactions []eth.ConfirmedTransaction, batchNonceID uint64,
) (*core.GeneratedBatchTxData, error) {
	args := c.Called(ctx, bridgeSmartContract, destinationChain, confirmedTransactions, batchNonceID)

	return args.Get(0).(*core.GeneratedBatchTxData), args.Error(1)
}

// SignBatchTransaction implements core.ChainOperations.
func (c *cardanoChainOperationsMock) SignBatchTransaction(txHash string) ([]byte, []byte, error) {
	args := c.Called(txHash)

	if args.Get(0) == nil {
		return []byte{}, []byte{}, args.Error(2)
	}

	return args.Get(0).([]byte), args.Get(1).([]byte), args.Error(2)
}

// IsSynchronized implements core.ChainOperations.
func (c *cardanoChainOperationsMock) IsSynchronized(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, chainID string,
) (bool, error) {
	args := c.Called(ctx, bridgeSmartContract, chainID)

	return args.Get(0).(bool), args.Error(1)
}
