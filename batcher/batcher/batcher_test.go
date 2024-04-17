package batcher

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"path"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBatcherExecute(t *testing.T) {
	testDir, err := os.MkdirTemp("", "cardano-prime")
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	config := &core.BatcherConfiguration{
		Base: core.BaseConfig{
			ChainId:     "prime",
			KeysDirPath: testDir,
		},
		Bridge: core.BridgeConfig{
			SecretsManager: &secrets.SecretsManagerConfig{
				Type: secrets.Local,
				Path: "dummy",
			},
		},
		PullTimeMilis: 2500,
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	testError := errors.New("test err")
	batchNonceId := big.NewInt(1)

	t.Run("GetNextBatchId returns err", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}
		bridgeSmartContractMock.On("GetNextBatchId", ctx, "prime").Return(big.NewInt(0), testError)

		b := NewBatcher(config, hclog.Default(), operationsMock, bridgeSmartContractMock, &common.BridgingRequestStateUpdaterMock{ReturnNil: true})
		err := b.execute(ctx)

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to query bridge.GetNextBatchId")
	})

	t.Run("GetNextBatchId returns 0", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}
		bridgeSmartContractMock.On("GetNextBatchId", ctx, "prime").Return(big.NewInt(0), nil)

		b := NewBatcher(config, hclog.Default(), operationsMock, bridgeSmartContractMock, &common.BridgingRequestStateUpdaterMock{ReturnNil: true})
		err := b.execute(ctx)

		require.NoError(t, err)
	})

	t.Run("GetConfirmedTransactions returns error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}
		bridgeSmartContractMock.On("GetNextBatchId", ctx, "prime").Return(batchNonceId, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, "prime").Return(nil, testError)

		b := NewBatcher(config, hclog.Default(), operationsMock, bridgeSmartContractMock, &common.BridgingRequestStateUpdaterMock{ReturnNil: true})
		err := b.execute(ctx)

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to query bridge.GetConfirmedTransactions")
	})

	getConfirmedTransactionsRet := []eth.ConfirmedTransaction{}

	t.Run("GenerateBatchTransaction returns error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}
		bridgeSmartContractMock.On("GetNextBatchId", ctx, "prime").Return(batchNonceId, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, "prime").Return(getConfirmedTransactionsRet, nil)
		operationsMock.On("GenerateBatchTransaction", ctx, bridgeSmartContractMock, "prime", getConfirmedTransactionsRet, batchNonceId).Return(nil, "", nil, nil, testError)

		b := NewBatcher(config, hclog.Default(), operationsMock, bridgeSmartContractMock, &common.BridgingRequestStateUpdaterMock{ReturnNil: true})
		err := b.execute(ctx)

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to generate batch transaction")
	})

	utxos := &eth.UTXOs{
		MultisigOwnedUTXOs: []eth.UTXO{},
		FeePayerOwnedUTXOs: []eth.UTXO{},
	}

	t.Run("SignBatchTransaction returns error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}
		bridgeSmartContractMock.On("GetNextBatchId", ctx, "prime").Return(batchNonceId, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, "prime").Return(getConfirmedTransactionsRet, nil)
		operationsMock.On("GenerateBatchTransaction", ctx, bridgeSmartContractMock, "prime", getConfirmedTransactionsRet, batchNonceId).Return([]byte{0}, "txHash", utxos, []*big.Int{big.NewInt(1)}, nil)
		operationsMock.On("SignBatchTransaction", "txHash").Return(nil, nil, testError)

		b := NewBatcher(config, hclog.Default(), operationsMock, bridgeSmartContractMock, &common.BridgingRequestStateUpdaterMock{ReturnNil: true})
		err := b.execute(ctx)

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to sign batch transaction")
	})

	t.Run("SubmitSignedBatch returns error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}
		bridgeSmartContractMock.On("GetNextBatchId", ctx, "prime").Return(batchNonceId, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, "prime").Return(getConfirmedTransactionsRet, nil)
		operationsMock.On("GenerateBatchTransaction", ctx, bridgeSmartContractMock, "prime", getConfirmedTransactionsRet, batchNonceId).Return([]byte{0}, "txHash", utxos, []*big.Int{big.NewInt(1)}, nil)
		operationsMock.On("SignBatchTransaction", "txHash").Return([]byte{}, []byte{}, nil)
		bridgeSmartContractMock.On("SubmitSignedBatch", ctx, mock.Anything).Return(testError)

		b := NewBatcher(config, hclog.Default(), operationsMock, bridgeSmartContractMock, &common.BridgingRequestStateUpdaterMock{ReturnNil: true})
		err := b.execute(ctx)

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to submit signed batch")
	})

	t.Run("execute pass", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}
		bridgeSmartContractMock.On("GetNextBatchId", ctx, "prime").Return(batchNonceId, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, "prime").Return(getConfirmedTransactionsRet, nil)
		operationsMock.On("GenerateBatchTransaction", ctx, bridgeSmartContractMock, "prime", getConfirmedTransactionsRet, batchNonceId).Return([]byte{0}, "txHash", utxos, []*big.Int{big.NewInt(1)}, nil)
		operationsMock.On("SignBatchTransaction", "txHash").Return([]byte{}, []byte{}, nil)
		bridgeSmartContractMock.On("SubmitSignedBatch", ctx, mock.Anything).Return(nil)

		b := NewBatcher(config, hclog.Default(), operationsMock, bridgeSmartContractMock, &common.BridgingRequestStateUpdaterMock{ReturnNil: true})
		err := b.execute(ctx)

		require.NoError(t, err)
	})
}

func TestBatcherGetChainSpecificOperations(t *testing.T) {
	jsonData := []byte(`{
		"testnetMagic": 2,
		"atLeastValidators": 3,
		"potentialFee": 300000
		}`)

	validPath, err := os.MkdirTemp("", "cardano-prime")
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(validPath)
		os.Remove(validPath)
	}()

	invalidPath := path.Join(validPath, "something_that_does_not_exist")

	_, err = cardanotx.GenerateWallet(validPath, false, true)
	require.NoError(t, err)

	t.Run("invalid chain type", func(t *testing.T) {
		chainSpecificConfig := core.ChainSpecific{
			ChainType: "Invalid",
			Config:    json.RawMessage(""),
		}

		chainOp, err := GetChainSpecificOperations(chainSpecificConfig, invalidPath)
		require.Nil(t, chainOp)
		require.Error(t, err)
		require.ErrorContains(t, err, "unknown chain type")
	})

	t.Run("invalid cardano json config", func(t *testing.T) {
		chainSpecificConfig := core.ChainSpecific{
			ChainType: "Cardano",
			Config:    json.RawMessage(""),
		}

		chainOp, err := GetChainSpecificOperations(chainSpecificConfig, invalidPath)
		require.Nil(t, chainOp)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to unmarshal Cardano configuration")
	})

	t.Run("invalid keys path", func(t *testing.T) {
		chainSpecificConfig := core.ChainSpecific{
			ChainType: "Cardano",
			Config:    json.RawMessage(jsonData),
		}

		chainOp, err := GetChainSpecificOperations(chainSpecificConfig, invalidPath)
		require.Nil(t, chainOp)
		require.Error(t, err)
		require.ErrorContains(t, err, "error while loading wallet info")
	})

	t.Run("valid cardano config and keys path", func(t *testing.T) {
		chainSpecificConfig := core.ChainSpecific{
			ChainType: "Cardano",
			Config:    json.RawMessage(jsonData),
		}

		chainOp, err := GetChainSpecificOperations(chainSpecificConfig, validPath)
		require.NoError(t, err)
		require.NotNil(t, chainOp)
	})

	t.Run("valid cardano config check case sensitivity", func(t *testing.T) {
		chainSpecificConfig := core.ChainSpecific{
			ChainType: "CaRdAnO",
			Config:    json.RawMessage(jsonData),
		}

		chainOp, err := GetChainSpecificOperations(chainSpecificConfig, validPath)
		require.NoError(t, err)
		require.NotNil(t, chainOp)
	})
}

type cardanoChainOperationsMock struct {
	mock.Mock
}

var _ core.ChainOperations = (*cardanoChainOperationsMock)(nil)

// GenerateBatchTransaction implements core.ChainOperations.
func (c *cardanoChainOperationsMock) GenerateBatchTransaction(ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, destinationChain string, confirmedTransactions []eth.ConfirmedTransaction, batchNonceId *big.Int) ([]byte, string, *eth.UTXOs, []*big.Int, error) {
	args := c.Called(ctx, bridgeSmartContract, destinationChain, confirmedTransactions, batchNonceId)

	if args.Get(0) == nil {
		return []byte{}, "", nil, nil, args.Error(4)
	}

	return args.Get(0).([]byte), args.String(1), args.Get(2).(*eth.UTXOs), args.Get(3).([]*big.Int), args.Error(4)
}

// SignBatchTransaction implements core.ChainOperations.
func (c *cardanoChainOperationsMock) SignBatchTransaction(txHash string) ([]byte, []byte, error) {
	args := c.Called(txHash)

	if args.Get(0) == nil {
		return []byte{}, []byte{}, args.Error(2)
	}

	return args.Get(0).([]byte), args.Get(1).([]byte), args.Error(2)
}
