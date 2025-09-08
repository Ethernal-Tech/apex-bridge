package batcher

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/validatorobserver"
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
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "0x333",
					Amount:             big.NewInt(10),
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
		operationsMock.On("ShouldConsolidate", testError).Return(false)

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
		batchData := &core.GeneratedBatchTxData{
			TxRaw:  []byte{0},
			TxHash: "txHash",
		}

		bridgeSmartContractMock.On("GetNextBatchID", ctx, common.ChainIDStrPrime).Return(batchNonceID, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, common.ChainIDStrPrime).Return(getConfirmedTransactionsRet, nil)
		operationsMock.On("GenerateBatchTransaction", ctx, bridgeSmartContractMock, common.ChainIDStrPrime, getConfirmedTransactionsRet, batchNonceID).
			Return(batchData, nil)
		operationsMock.On("SignBatchTransaction", batchData).Return(nil, nil, testError)

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
		batchData := &core.GeneratedBatchTxData{
			TxRaw:  []byte{0},
			TxHash: "txHash",
		}

		bridgeSmartContractMock.On("GetNextBatchID", ctx, common.ChainIDStrPrime).Return(batchNonceID, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, common.ChainIDStrPrime).Return(getConfirmedTransactionsRet, nil)
		operationsMock.On("GenerateBatchTransaction", ctx, bridgeSmartContractMock, common.ChainIDStrPrime, getConfirmedTransactionsRet, batchNonceID).
			Return(batchData, nil)
		operationsMock.On("SignBatchTransaction", batchData).Return([]byte{}, []byte{}, nil)
		operationsMock.On("Submit", ctx, bridgeSmartContractMock, mock.Anything).Return(testError)

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
		batchData := &core.GeneratedBatchTxData{
			TxRaw:  []byte{0},
			TxHash: "txHash",
		}

		bridgeSmartContractMock.On("GetNextBatchID", ctx, common.ChainIDStrPrime).Return(batchNonceID, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, common.ChainIDStrPrime).Return(getConfirmedTransactionsRet, nil)
		operationsMock.On("GenerateBatchTransaction", ctx, bridgeSmartContractMock, common.ChainIDStrPrime, getConfirmedTransactionsRet, batchNonceID).
			Return(batchData, nil)
		operationsMock.On("SignBatchTransaction", batchData).Return([]byte{}, []byte{}, nil)
		operationsMock.On("Submit", ctx, bridgeSmartContractMock, mock.Anything).Return(error(nil))

		b := NewBatcher(config, operationsMock,
			bridgeSmartContractMock, &common.BridgingRequestStateUpdaterMock{ReturnNil: true}, hclog.NewNullLogger())
		batchID, err := b.execute(ctx)

		require.NoError(t, err)
		require.Equal(t, batchNonceID, batchID)
	})

	t.Run("Test Validator set change - five multisig one fee", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}

		testDir, err := os.MkdirTemp("", "bat-chain-ops-tx")
		require.NoError(t, err)

		defer func() {
			os.RemoveAll(testDir)
			os.Remove(testDir)
		}()

		secretsMngr, err := secretsHelper.CreateSecretsManager(&secrets.SecretsManagerConfig{
			Path: filepath.Join(testDir, "stp"),
			Type: secrets.Local,
		})
		require.NoError(t, err)

		wallet, err := cardanotx.GenerateWallet(secretsMngr, "prime", true, false)
		require.NoError(t, err)

		wallet2, err := cardanotx.GenerateWallet(secretsMngr, "prime2", true, false)
		require.NoError(t, err)

		dbMock := &indexer.DatabaseMock{}

		configRaw := json.RawMessage([]byte(`{
			"socketPath": "./socket",
			"testnetMagic": 42,
			"minUtxoAmount": 1000,
			"maxFeeUtxoCount": 2,
			"maxUtxoCount": 5,
			"slotRoundingThreshold": 2
			}`))

		validatorsChainData := []eth.ValidatorChainData{
			{
				Key: [4]*big.Int{
					new(big.Int).SetBytes(wallet.MultiSig.VerificationKey),
					new(big.Int).SetBytes(wallet.Fee.VerificationKey),
					new(big.Int).SetBytes(wallet.MultiSig.StakeVerificationKey),
					new(big.Int).SetBytes(wallet.Fee.StakeVerificationKey),
				},
			},
		}

		newValidatorChainData := append([]eth.ValidatorChainData{}, validatorsChainData[0], eth.ValidatorChainData{
			Key: [4]*big.Int{
				new(big.Int).SetBytes(wallet2.MultiSig.VerificationKey),
				new(big.Int).SetBytes(wallet2.Fee.VerificationKey),
				new(big.Int).SetBytes(wallet2.MultiSig.StakeVerificationKey),
				new(big.Int).SetBytes(wallet2.Fee.StakeVerificationKey),
			},
		})

		txProviderMock := &cardanotx.TxProviderTestMock{
			ReturnDefaultParameters: true,
		}

		indxUpdaterMock := &IndexerUpdaterMock{}
		indxUpdaterMock.On("AddNewAddressesOfInterest", mock.Anything, mock.Anything).Return()

		operations, err := NewCardanoChainOperations(configRaw, dbMock, nil, secretsMngr, common.ChainIDStrPrime, hclog.NewNullLogger())
		require.NoError(t, err)

		operations.txProvider = txProviderMock
		operations.indxUpdater = indxUpdaterMock

		multisig := generateUTXO(5)
		fee := generateUTXO(1)

		dbMock.On("GetAllTxOutputs", mock.Anything, mock.Anything).Return(multisig, nil)
		dbMock.On("GetAllTxOutputs", mock.Anything, mock.Anything).Return(fee, nil)

		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{
			BlockSlot: 4,
			BlockHash: indexer.Hash{},
		}, nil)

		bridgeSmartContractMock.On("GetNextBatchID", ctx, common.ChainIDStrPrime).Return(batchNonceID, nil)
		bridgeSmartContractMock.On("GetConfirmedTransactions", ctx, common.ChainIDStrPrime).Return(getConfirmedTransactionsRet, nil)
		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, common.ChainIDStrPrime).Return(validatorsChainData, nil)
		bridgeSmartContractMock.On("SubmitSignedBatch", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		b := NewBatcher(config, operations, bridgeSmartContractMock, &common.BridgingRequestStateUpdaterMock{ReturnNil: true}, hclog.NewNullLogger())

		b.UpdateValidatorSet(&validatorobserver.ValidatorsPerChain{
			common.ChainIDStrPrime: {
				Keys:       newValidatorChainData,
				SlotNumber: 0,
			},
		},
		)

		_, err = b.execute(ctx)
		require.NoError(t, err)
	})
}

func generateUTXO(size uint) []*indexer.TxInputOutput {
	txs := make([]*indexer.TxInputOutput, size)

	for i := range txs {
		txs[i] = &indexer.TxInputOutput{
			Input: indexer.TxInput{},
			Output: indexer.TxOutput{
				Amount: 1000000,
			},
		}
	}

	return txs
}

func TestBatcherGetChainSpecificOperations(t *testing.T) {
	validPath, err := os.MkdirTemp("", "cardano-prime")
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(validPath)
		os.Remove(validPath)
	}()

	secretsMngr, err := secretsHelper.CreateSecretsManager(&secrets.SecretsManagerConfig{
		Path: filepath.Join(validPath, "stp"),
		Type: secrets.Local,
	})
	require.NoError(t, err)

	_, err = cardanotx.GenerateWallet(secretsMngr, "prime", false, true)
	require.NoError(t, err)

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
			common.NewBridgingRequestStateKey(common.ChainIDStrPrime, included[0]),
			common.NewBridgingRequestStateKey(common.ChainIDStrVector, included[1]),
			common.NewBridgingRequestStateKey(common.ChainIDStrPrime, included[2]),
		}, res)
	})
}

type cardanoChainOperationsMock struct {
	mock.Mock
}

// CreateValidatorSetChangeTx implements core.ChainOperations.
func (c *cardanoChainOperationsMock) CreateValidatorSetChangeTx(ctx context.Context, chainID string,
	nextBatchID uint64, bridgeSmartContract eth.IBridgeSmartContract,
	validatorsKeys validatorobserver.ValidatorsPerChain, lastBatchID uint64, lastBatchType uint8,
) (bool, *core.GeneratedBatchTxData, error) {
	args := c.Called(ctx, chainID, nextBatchID, bridgeSmartContract, validatorsKeys, lastBatchID, lastBatchType)

	return args.Get(0).(bool), args.Get(1).(*core.GeneratedBatchTxData), args.Error(2)
}

var _ core.ChainOperations = (*cardanoChainOperationsMock)(nil)

func (c *cardanoChainOperationsMock) GenerateMultisigAddress(
	validators *validatorobserver.ValidatorsPerChain, chainID string) error {
	return nil
}

// GenerateBatchTransaction implements core.ChainOperations.
func (c *cardanoChainOperationsMock) GenerateBatchTransaction(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract,
	destinationChain string, confirmedTransactions []eth.ConfirmedTransaction, batchNonceID uint64,
) (*core.GeneratedBatchTxData, error) {
	args := c.Called(ctx, bridgeSmartContract, destinationChain, confirmedTransactions, batchNonceID)

	return args.Get(0).(*core.GeneratedBatchTxData), args.Error(1)
}

// SignBatchTransaction implements core.ChainOperations.
func (c *cardanoChainOperationsMock) SignBatchTransaction(generatedBatchData *core.GeneratedBatchTxData) ([]byte, []byte, error) {
	args := c.Called(generatedBatchData)

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

func (c *cardanoChainOperationsMock) Submit(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, batch eth.SignedBatch,
) error {
	return c.Called(ctx, bridgeSmartContract, batch).Error(0)
}
