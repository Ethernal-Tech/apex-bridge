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
	"github.com/Ethernal-Tech/apex-bridge/eth"
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
		Chain: core.ChainConfig{
			ChainID:   "prime",
			ChainType: "Cardano",
			ChainSpecific: json.RawMessage([]byte(fmt.Sprintf(`{
				"socketPath": "./socket",
				"testnetMagic": 2,
				"potentialFee": 300000,
				"keysDirPath": "%s"
				}`, testDir))),
		},
		Bridge: core.BridgeConfig{
			ValidatorDataDir: path.Join(testDir, "some"),
		},
		PullTimeMilis: 2500,
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	testError := errors.New("test err")
	batchNonceID := big.NewInt(1)

	t.Run("GetNextBatchID returns err", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}

		bridgeSmartContractMock.On("GetNextBatchID", ctx, "prime").Return(big.NewInt(0), testError)

		b := NewBatcher(config, hclog.Default(), operationsMock, bridgeSmartContractMock, &common.BridgingRequestStateUpdaterMock{ReturnNil: true})
		err := b.execute(ctx)

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to query bridge.GetNextBatchID")
	})

	t.Run("GetNextBatchID returns 0", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}

		bridgeSmartContractMock.On("GetNextBatchID", ctx, "prime").Return(big.NewInt(0), nil)

		b := NewBatcher(config, hclog.Default(), operationsMock, bridgeSmartContractMock, &common.BridgingRequestStateUpdaterMock{ReturnNil: true})
		err := b.execute(ctx)

		require.NoError(t, err)
	})

	t.Run("GetConfirmedTransactions returns error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}

		bridgeSmartContractMock.On("GetNextBatchID", ctx, "prime").Return(batchNonceID, nil)
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

		bridgeSmartContractMock.On("GetNextBatchID", ctx, "prime").Return(batchNonceID, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, "prime").Return(getConfirmedTransactionsRet, nil)
		operationsMock.On("GenerateBatchTransaction", ctx, bridgeSmartContractMock, "prime", getConfirmedTransactionsRet, batchNonceID).Return(nil, "", nil, nil, testError)

		b := NewBatcher(config, hclog.Default(), operationsMock, bridgeSmartContractMock, &common.BridgingRequestStateUpdaterMock{ReturnNil: true})
		err := b.execute(ctx)

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to generate batch transaction")
	})

	utxos := &eth.UTXOs{
		MultisigOwnedUTXOs: []eth.UTXO{},
		FeePayerOwnedUTXOs: []eth.UTXO{},
	}

	includedTransactions := make(map[uint64]eth.ConfirmedTransaction)

	t.Run("SignBatchTransaction returns error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}

		bridgeSmartContractMock.On("GetNextBatchID", ctx, "prime").Return(batchNonceID, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, "prime").Return(getConfirmedTransactionsRet, nil)
		operationsMock.On("GenerateBatchTransaction", ctx, bridgeSmartContractMock, "prime", getConfirmedTransactionsRet, batchNonceID).Return([]byte{0}, "txHash", utxos, includedTransactions, nil)
		operationsMock.On("SignBatchTransaction", "txHash").Return(nil, nil, testError)

		b := NewBatcher(config, hclog.Default(), operationsMock, bridgeSmartContractMock, &common.BridgingRequestStateUpdaterMock{ReturnNil: true})
		err := b.execute(ctx)

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to sign batch transaction")
	})

	t.Run("SubmitSignedBatch returns error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		operationsMock := &cardanoChainOperationsMock{}

		bridgeSmartContractMock.On("GetNextBatchID", ctx, "prime").Return(batchNonceID, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, "prime").Return(getConfirmedTransactionsRet, nil)
		operationsMock.On("GenerateBatchTransaction", ctx, bridgeSmartContractMock, "prime", getConfirmedTransactionsRet, batchNonceID).Return([]byte{0}, "txHash", utxos, includedTransactions, nil)
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

		bridgeSmartContractMock.On("GetNextBatchID", ctx, "prime").Return(batchNonceID, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, "prime").Return(getConfirmedTransactionsRet, nil)
		operationsMock.On("GenerateBatchTransaction", ctx, bridgeSmartContractMock, "prime", getConfirmedTransactionsRet, batchNonceID).Return([]byte{0}, "txHash", utxos, includedTransactions, nil)
		operationsMock.On("SignBatchTransaction", "txHash").Return([]byte{}, []byte{}, nil)
		bridgeSmartContractMock.On("SubmitSignedBatch", ctx, mock.Anything).Return(nil)

		b := NewBatcher(config, hclog.Default(), operationsMock, bridgeSmartContractMock, &common.BridgingRequestStateUpdaterMock{ReturnNil: true})
		err := b.execute(ctx)

		require.NoError(t, err)
	})
}

func TestBatcherGetChainSpecificOperations(t *testing.T) {
	validPath, err := os.MkdirTemp("", "cardano-prime")
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(validPath)
		os.Remove(validPath)
	}()

	_, err = cardanotx.GenerateWallet(validPath, false, true)
	require.NoError(t, err)

	chainConfig := core.ChainConfig{
		ChainID:   "prime",
		ChainType: "Cardano",
		ChainSpecific: json.RawMessage([]byte(fmt.Sprintf(`{
			"socketPath": "./socket",
			"testnetMagic": 2,
			"potentialFee": 300000,
			"keysDirPath": "%s"
			}`, validPath))),
	}

	t.Run("invalid chain type", func(t *testing.T) {
		cfg := chainConfig
		cfg.ChainType = "invalid"

		chainOp, err := GetChainSpecificOperations(cfg)
		require.Nil(t, chainOp)
		require.Error(t, err)
		require.ErrorContains(t, err, "unknown chain type")
	})

	t.Run("invalid cardano json config", func(t *testing.T) {
		cfg := chainConfig
		cfg.ChainSpecific = json.RawMessage("")

		chainOp, err := GetChainSpecificOperations(cfg)
		require.Nil(t, chainOp)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to unmarshal Cardano configuration")
	})

	t.Run("invalid keys path", func(t *testing.T) {
		chainConfig := core.ChainConfig{
			ChainID:   "prime",
			ChainType: "Cardano",
			ChainSpecific: json.RawMessage([]byte(fmt.Sprintf(`{
				"testnetMagic": 2,
				"socketPath": "./socket",
				"potentialFee": 300000,
				"keysDirPath": "%s"
				}`, path.Join(validPath, "a1", "a2", "a3")))),
		}

		chainOp, err := GetChainSpecificOperations(chainConfig)
		require.Nil(t, chainOp)
		require.Error(t, err)
		require.ErrorContains(t, err, "error while loading wallet info")
	})

	t.Run("valid cardano config and keys path", func(t *testing.T) {
		chainOp, err := GetChainSpecificOperations(chainConfig)
		require.NoError(t, err)
		require.NotNil(t, chainOp)
	})

	t.Run("valid cardano config check case sensitivity", func(t *testing.T) {
		cfg := chainConfig
		cfg.ChainType = "CaRDaNo"

		chainOp, err := GetChainSpecificOperations(cfg)
		require.NoError(t, err)
		require.NotNil(t, chainOp)
	})
}

type cardanoChainOperationsMock struct {
	mock.Mock
}

var _ core.ChainOperations = (*cardanoChainOperationsMock)(nil)

// GenerateBatchTransaction implements core.ChainOperations.
func (c *cardanoChainOperationsMock) GenerateBatchTransaction(ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, destinationChain string, confirmedTransactions []eth.ConfirmedTransaction, batchNonceID *big.Int) ([]byte, string, *eth.UTXOs, map[uint64]eth.ConfirmedTransaction, error) {
	args := c.Called(ctx, bridgeSmartContract, destinationChain, confirmedTransactions, batchNonceID)

	if args.Get(0) == nil {
		return []byte{}, "", nil, nil, args.Error(4)
	}

	return args.Get(0).([]byte), args.String(1), args.Get(2).(*eth.UTXOs), args.Get(3).(map[uint64]eth.ConfirmedTransaction), args.Error(4)
}

// SignBatchTransaction implements core.ChainOperations.
func (c *cardanoChainOperationsMock) SignBatchTransaction(txHash string) ([]byte, []byte, error) {
	args := c.Called(txHash)

	if args.Get(0) == nil {
		return []byte{}, []byte{}, args.Error(2)
	}

	return args.Get(0).([]byte), args.Get(1).([]byte), args.Error(2)
}
