package batcher

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBatcherExecute(t *testing.T) {
	config := &core.BatcherConfiguration{
		Base: core.BaseConfig{
			ChainId:     "prime",
			KeysDirPath: "../keys/prime",
		},
		Bridge: core.BridgeConfig{
			NodeUrl:              "https://polygon-mumbai-pokt.nodies.app", // will be our node,
			SmartContractAddress: "0x4c7aBbe2c5A44d758b70BE5C3c07eEB573304Db4",
			SigningKey:           "3761f6deeb2e0b2aa8b843e804d880afa6e5fecf1631f411e267641a72d0ca20",
		},
		PullTimeMilis: 2500,
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	testError := errors.New("test err")

	t.Run("should create batch returns err", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}
		bridgeSmartContractMock.On("ShouldCreateBatch", ctx, "prime").Return(false, testError)

		b := NewBatcher(config, hclog.Default(), operationsMock, bridgeSmartContractMock)
		err := b.execute(ctx)

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to query bridge.ShouldCreateBatch")
	})

	t.Run("should create batch returns false", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}
		bridgeSmartContractMock.On("ShouldCreateBatch", ctx, "prime").Return(false, nil)

		b := NewBatcher(config, hclog.Default(), operationsMock, bridgeSmartContractMock)
		err := b.execute(ctx)

		require.NoError(t, err)
	})

	t.Run("GetConfirmedTransactions returns error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}
		bridgeSmartContractMock.On("ShouldCreateBatch", ctx, "prime").Return(true, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, "prime").Return(nil, testError)

		b := NewBatcher(config, hclog.Default(), operationsMock, bridgeSmartContractMock)
		err := b.execute(ctx)

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to query bridge.GetConfirmedTransactions")
	})

	getConfirmedTransactionsRet := []eth.ConfirmedTransaction{}

	t.Run("GenerateBatchTransaction returns error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}
		bridgeSmartContractMock.On("ShouldCreateBatch", ctx, "prime").Return(true, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, "prime").Return(getConfirmedTransactionsRet, nil)
		operationsMock.On("GenerateBatchTransaction", ctx, bridgeSmartContractMock, "prime", getConfirmedTransactionsRet).Return(nil, "", nil, testError)

		b := NewBatcher(config, hclog.Default(), operationsMock, bridgeSmartContractMock)
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
		bridgeSmartContractMock.On("ShouldCreateBatch", ctx, "prime").Return(true, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, "prime").Return(getConfirmedTransactionsRet, nil)
		operationsMock.On("GenerateBatchTransaction", ctx, bridgeSmartContractMock, "prime", getConfirmedTransactionsRet).Return([]byte{0}, "txHash", utxos, nil)
		operationsMock.On("SignBatchTransaction", "txHash").Return(nil, nil, testError)

		b := NewBatcher(config, hclog.Default(), operationsMock, bridgeSmartContractMock)
		err := b.execute(ctx)

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to sign batch transaction")
	})

	t.Run("SubmitSignedBatch returns error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}
		bridgeSmartContractMock.On("ShouldCreateBatch", ctx, "prime").Return(true, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, "prime").Return(getConfirmedTransactionsRet, nil)
		operationsMock.On("GenerateBatchTransaction", ctx, bridgeSmartContractMock, "prime", getConfirmedTransactionsRet).Return([]byte{0}, "txHash", utxos, nil)
		operationsMock.On("SignBatchTransaction", "txHash").Return([]byte{}, []byte{}, nil)
		bridgeSmartContractMock.On("SubmitSignedBatch", ctx, mock.Anything).Return(testError)

		b := NewBatcher(config, hclog.Default(), operationsMock, bridgeSmartContractMock)
		err := b.execute(ctx)

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to submit signed batch")
	})

	t.Run("execute pass", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}
		bridgeSmartContractMock.On("ShouldCreateBatch", ctx, "prime").Return(true, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, "prime").Return(getConfirmedTransactionsRet, nil)
		operationsMock.On("GenerateBatchTransaction", ctx, bridgeSmartContractMock, "prime", getConfirmedTransactionsRet).Return([]byte{0}, "txHash", utxos, nil)
		operationsMock.On("SignBatchTransaction", "txHash").Return([]byte{}, []byte{}, nil)
		bridgeSmartContractMock.On("SubmitSignedBatch", ctx, mock.Anything).Return(nil)

		b := NewBatcher(config, hclog.Default(), operationsMock, bridgeSmartContractMock)
		err := b.execute(ctx)

		require.NoError(t, err)
	})
}

func TestBatcherGetChainSpecificOperations(t *testing.T) {
	jsonData := []byte(`{
		"testnetMagic": 2,
		"blockfrostUrl": "https://cardano-preview.blockfrost.io/api/v0",
		"blockfrostApiKey": "preview7mGSjpyEKb24OxQ4cCxomxZ5axMs5PvE",
		"atLeastValidators": 0.6666666666666666,
		"potentialFee": 300000
		}`)

	validPath := "../keys/prime"
	invalidPath := "invalidPath"

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
		require.NotNil(t, chainOp)
		require.NoError(t, err)
	})

	t.Run("valid cardano config check case sensitivity", func(t *testing.T) {
		chainSpecificConfig := core.ChainSpecific{
			ChainType: "CaRdAnO",
			Config:    json.RawMessage(jsonData),
		}

		chainOp, err := GetChainSpecificOperations(chainSpecificConfig, validPath)
		require.NotNil(t, chainOp)
		require.NoError(t, err)
	})
}

type cardanoChainOperationsMock struct {
	mock.Mock
}

var _ core.ChainOperations = (*cardanoChainOperationsMock)(nil)

// GenerateBatchTransaction implements core.ChainOperations.
func (c *cardanoChainOperationsMock) GenerateBatchTransaction(ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, destinationChain string, confirmedTransactions []eth.ConfirmedTransaction) ([]byte, string, *eth.UTXOs, error) {
	args := c.Called(ctx, bridgeSmartContract, destinationChain, confirmedTransactions)

	if args.Get(0) == nil {
		return []byte{}, "", nil, args.Error(3)
	}

	return args.Get(0).([]byte), args.String(1), args.Get(2).(*eth.UTXOs), args.Error(3)
}

// SignBatchTransaction implements core.ChainOperations.
func (c *cardanoChainOperationsMock) SignBatchTransaction(txHash string) ([]byte, []byte, error) {
	args := c.Called(txHash)

	if args.Get(0) == nil {
		return []byte{}, []byte{}, args.Error(2)
	}

	return args.Get(0).([]byte), args.Get(1).([]byte), args.Error(2)
}
