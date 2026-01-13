package batcher

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/validatorobserver"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer/gouroboros"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	secretsHelper "github.com/Ethernal-Tech/cardano-infrastructure/secrets/helper"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type IndexerUpdaterMock struct {
	mock.Mock
}

func (*IndexerUpdaterMock) AddNewAddressesOfInterest(address ...string) {}

func TestCardanoChainOperations_IsSynchronized(t *testing.T) {
	chainID := "prime"
	dbMock := &indexer.DatabaseMock{}
	bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
	ctx := context.Background()
	scBlock1 := eth.CardanoBlock{
		BlockSlot: big.NewInt(15),
	}
	scBlock2 := eth.CardanoBlock{
		BlockSlot: big.NewInt(20),
	}
	oracleBlock1 := &indexer.BlockPoint{
		BlockSlot: uint64(10),
	}
	oracleBlock2 := &indexer.BlockPoint{
		BlockSlot: uint64(20),
	}
	testErr1 := errors.New("test error 1")
	testErr2 := errors.New("test error 2")

	bridgeSmartContractMock.On("GetLastObservedBlock", ctx, chainID).Return(eth.CardanoBlock{}, testErr1).Once()
	bridgeSmartContractMock.On("GetLastObservedBlock", ctx, chainID).Return(eth.CardanoBlock{}, nil).Once()
	bridgeSmartContractMock.On("GetLastObservedBlock", ctx, chainID).Return(scBlock1, nil).Times(3)
	bridgeSmartContractMock.On("GetLastObservedBlock", ctx, chainID).Return(scBlock2, nil).Once()

	dbMock.On("GetLatestBlockPoint").Return((*indexer.BlockPoint)(nil), testErr2).Once()
	dbMock.On("GetLatestBlockPoint").Return(oracleBlock1, nil).Once()
	dbMock.On("GetLatestBlockPoint").Return(oracleBlock2, nil).Twice()

	cco := &CardanoChainOperations{
		db:     dbMock,
		logger: hclog.NewNullLogger(),
	}

	// sc error
	_, err := cco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.ErrorIs(t, err, testErr1)

	// database error
	_, err = cco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.ErrorIs(t, err, testErr2)

	// not in sync
	val, err := cco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.NoError(t, err)
	require.False(t, val)

	// in sync
	val, err = cco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.NoError(t, err)
	require.True(t, val)

	// in sync again
	val, err = cco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.NoError(t, err)
	require.True(t, val)
}

func TestGenerateBatchTransaction(t *testing.T) {
	testDir, err := os.MkdirTemp("", "bat-chain-ops-tx")
	require.NoError(t, err)

	minUtxoAmount := new(big.Int).SetUint64(1_000)

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	secretsMngr, err := secretsHelper.CreateSecretsManager(&secrets.SecretsManagerConfig{
		Path: filepath.Join(testDir, "stp"),
		Type: secrets.Local,
	})
	require.NoError(t, err)

	wallet, err := cardano.GenerateWallet(secretsMngr, "prime", true, false)
	require.NoError(t, err)

	configRaw := json.RawMessage([]byte(`{
			"socketPath": "./socket",
			"testnetMagic": 42,
			"minUtxoAmount": 1000
			}`))
	dbMock := &indexer.DatabaseMock{}
	txProviderMock := &cardano.TxProviderTestMock{
		ReturnDefaultParameters: true,
	}

	cco, err := NewCardanoChainOperations(configRaw, dbMock, nil, secretsMngr, "prime", 1*time.Millisecond, hclog.NewNullLogger())
	require.NoError(t, err)

	cco.txProvider = txProviderMock
	cco.config.SlotRoundingThreshold = 100
	cco.config.MaxFeeUtxoCount = 4
	cco.config.MaxUtxoCount = 50

	testError := errors.New("test err")
	confirmedTransactions := make([]eth.ConfirmedTransaction, 1)
	confirmedTransactions[0] = eth.ConfirmedTransaction{
		Nonce:       1,
		BlockHeight: big.NewInt(1),
		Receivers: []eth.BridgeReceiver{{
			DestinationAddress: "addr_test1vqeux7xwusdju9dvsj8h7mca9aup2k439kfmwy773xxc2hcu7zy99",
			Amount:             minUtxoAmount,
		}},
	}
	batchNonceID := uint64(1)
	destinationChain := common.ChainIDStrVector
	validValidatorsChainData := []eth.ValidatorChainData{
		{
			Key: [4]*big.Int{
				new(big.Int).SetBytes(wallet.MultiSig.VerificationKey),
				new(big.Int).SetBytes(wallet.Fee.VerificationKey),
				new(big.Int).SetBytes(wallet.MultiSig.StakeVerificationKey),
				new(big.Int).SetBytes(wallet.Fee.StakeVerificationKey),
			},
		},
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	t.Run("GetValidatorsChainData returns error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(nil, testError).Once()

		_, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("no vkey for multisig address error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		getValidatorsCardanoDataRet := []eth.ValidatorChainData{
			{
				Key: [4]*big.Int{
					new(big.Int), new(big.Int),
				},
			},
		}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil).Once()

		_, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "verifying keys of current batcher wasn't found in validators data queried from smart contract")
	})

	t.Run("no vkey for multisig fee address error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		getValidatorsCardanoDataRet := []eth.ValidatorChainData{
			{
				Key: [4]*big.Int{
					new(big.Int).SetBytes(wallet.MultiSig.VerificationKey), new(big.Int),
					new(big.Int).SetBytes(wallet.MultiSig.VerificationKey), new(big.Int).SetBytes(wallet.MultiSig.VerificationKey),
				},
			},
			{
				Key: [4]*big.Int{
					new(big.Int).SetBytes(wallet.MultiSig.StakeSigningKey), new(big.Int).SetBytes(wallet.MultiSig.StakeSigningKey),
					new(big.Int).SetBytes(wallet.MultiSig.VerificationKey), new(big.Int).SetBytes(wallet.MultiSig.VerificationKey),
				},
			},
		}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil).Once()

		_, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "verifying keys of current batcher wasn't found in validators data queried from smart contract")
	})

	t.Run("GetLatestBlockPoint return error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(validValidatorsChainData, nil).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 2000000,
					},
				},
			}, error(nil)).Twice()
		dbMock.On("GetLatestBlockPoint").Return((*indexer.BlockPoint)(nil), testError).Once()

		_, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GetAllTxOutputs return error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(validValidatorsChainData, nil).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput(nil), testError).Once()

		_, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GenerateBatchTransaction fee multisig does not have any utxo", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(validValidatorsChainData, nil).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 2000000,
					},
				},
			}, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{}, error(nil)).Once()

		_, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.ErrorContains(t, err, "fee multisig does not have any utxo")
	})

	t.Run("GenerateBatchTransaction should pass", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(validValidatorsChainData, nil).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 2_000_000,
					},
				},
			}, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0xFF"),
					},
					Output: indexer.TxOutput{
						Amount: 300_000,
					},
				},
			}, error(nil)).Once()

		result, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.NoError(t, err)
		require.NotNil(t, result.TxRaw)
		require.NotEqual(t, "", result.TxHash)
	})

	t.Run("Test SignBatchTransaction", func(t *testing.T) {
		txRaw, err := hex.DecodeString("84a5008282582000000000000000000000000000000000000000000000000000000000000000120082582000000000000000000000000000000000000000000000000000000000000000ff00018282581d6033c378cee41b2e15ac848f7f6f1d2f78155ab12d93b713de898d855f1903e882581d702b5398fcb481e94163a6b5cca889c54bcd9d340fb71c5eaa9f2c8d441a001e8098021a0002e76d031864075820c5e403ad2ee72ff4eb1ab7e988c1e1b4cb34df699cb9112d6bded8e8f3195f34a10182830301818200581ce67d6de92a4abb3712e887fe2cf0f07693028fad13a3e510dbe73394830301818200581c31a31e2f2cd4e1d66fc25f400aa02ab0fe6ca5a3d735c2974e842a89f5d90103a100a101a2616e016174656261746368")
		require.NoError(t, err)

		witnessMultiSig, witnessMultiSigFee, err := cco.SignBatchTransaction(
			&core.GeneratedBatchTxData{TxRaw: txRaw})
		require.NoError(t, err)
		require.NotNil(t, witnessMultiSig)
		require.NotNil(t, witnessMultiSigFee)
	})
}

func Test_createBatchInitialData(t *testing.T) {
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

	wallet, err := cardano.GenerateWallet(secretsMngr, "prime", true, false)
	require.NoError(t, err)

	configRaw := json.RawMessage([]byte(`{
			"socketPath": "./socket",
			"testnetMagic": 42,
			"minUtxoAmount": 1000
			}`))
	dbMock := &indexer.DatabaseMock{}
	txProviderMock := &cardano.TxProviderTestMock{
		ReturnDefaultParameters: true,
	}

	cco, err := NewCardanoChainOperations(configRaw, dbMock, nil, secretsMngr, "prime", 1*time.Millisecond, hclog.NewNullLogger())
	require.NoError(t, err)

	cco.txProvider = txProviderMock

	testError := errors.New("test err")
	batchID := uint64(1)
	destinationChain := common.ChainIDStrVector
	validValidatorsChainData := []eth.ValidatorChainData{
		{
			Key: [4]*big.Int{
				new(big.Int).SetBytes(wallet.MultiSig.VerificationKey),
				new(big.Int).SetBytes(wallet.Fee.VerificationKey),
				new(big.Int).SetBytes(wallet.MultiSig.StakeVerificationKey),
				new(big.Int).SetBytes(wallet.Fee.StakeVerificationKey),
			},
		},
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	t.Run("GetValidatorsChainData returns error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(nil, testError).Once()

		_, err := cco.createBatchInitialData(ctx, bridgeSmartContractMock, destinationChain, batchID)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("no vkey for multisig address error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		getValidatorsCardanoDataRet := []eth.ValidatorChainData{
			{
				Key: [4]*big.Int{
					new(big.Int), new(big.Int).SetBytes(wallet.Fee.VerificationKey),
				},
			},
		}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil).Once()

		_, err := cco.createBatchInitialData(ctx, bridgeSmartContractMock, destinationChain, batchID)
		require.ErrorContains(t, err, "verifying keys of current batcher wasn't found in validators data queried from smart contract")
	})

	t.Run("no vkey for multisig fee address error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		getValidatorsCardanoDataRet := []eth.ValidatorChainData{
			{
				Key: [4]*big.Int{
					new(big.Int).SetBytes(wallet.MultiSig.VerificationKey), new(big.Int),
				},
			},
		}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil).Once()

		_, err := cco.createBatchInitialData(ctx, bridgeSmartContractMock, destinationChain, batchID)
		require.ErrorContains(t, err, "verifying keys of current batcher wasn't found in validators data queried from smart contract")
	})

	t.Run("GetProtocolParameters error", func(t *testing.T) {
		desiredErr := errors.New("hello")

		txProviderMock.On("GetProtocolParameters", ctx).Return(nil, desiredErr)

		txProviderMock.ReturnDefaultParameters = false
		defer func() {
			txProviderMock.ReturnDefaultParameters = true
		}()

		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).
			Return(validValidatorsChainData, nil).Once()

		_, err := cco.createBatchInitialData(ctx, bridgeSmartContractMock, destinationChain, batchID)
		require.ErrorIs(t, err, desiredErr)
	})

	t.Run("good", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).
			Return(validValidatorsChainData, nil).Once()

		data, err := cco.createBatchInitialData(ctx, bridgeSmartContractMock, destinationChain, batchID)
		require.NoError(t, err)

		assert.Contains(t, data.MultisigAddr, "addr_test1")
		assert.Contains(t, data.FeeAddr, "addr_test1")
		assert.NotEqual(t, data.MultisigAddr, data.FeeAddr)
		assert.Equal(t, batchID, data.BatchNonceID)
		assert.Greater(t, len(data.Metadata), 5)
	})
}

func TestGenerateConsolidationTransaction(t *testing.T) {
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

	wallet, err := cardano.GenerateWallet(secretsMngr, "prime", false, false)
	require.NoError(t, err)

	configRaw := json.RawMessage([]byte(`{
			"socketPath": "./socket",
			"testnetMagic": 42,
			"minUtxoAmount": 1000
			}`))
	dbMock := &indexer.DatabaseMock{}
	txProviderMock := &cardano.TxProviderTestMock{
		ReturnDefaultParameters: true,
	}

	cco, err := NewCardanoChainOperations(configRaw, dbMock, nil, secretsMngr, "prime", 1*time.Millisecond, hclog.NewNullLogger())
	require.NoError(t, err)

	cco.txProvider = txProviderMock
	cco.config.SlotRoundingThreshold = 100
	cco.config.MaxFeeUtxoCount = 4
	cco.config.MaxUtxoCount = 50

	testError := errors.New("test err")
	batchID := uint64(1)
	destinationChain := common.ChainIDStrVector
	validValidatorsChainData := []eth.ValidatorChainData{
		{
			Key: [4]*big.Int{
				new(big.Int).SetBytes(wallet.MultiSig.VerificationKey),
				new(big.Int).SetBytes(wallet.Fee.VerificationKey),
			},
		},
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	t.Run("GetLatestBlockPoint return error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(validValidatorsChainData, nil).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 2000000,
					},
				},
			}, error(nil)).Twice()
		dbMock.On("GetLatestBlockPoint").Return((*indexer.BlockPoint)(nil), testError).Once()

		data, err := cco.createBatchInitialData(ctx, bridgeSmartContractMock, destinationChain, batchID)
		require.NoError(t, err)

		_, err = cco.generateConsolidationTransaction(data)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GetAllTxOutputs return error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(validValidatorsChainData, nil).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput(nil), testError).Once()

		data, err := cco.createBatchInitialData(ctx, bridgeSmartContractMock, destinationChain, batchID)
		require.NoError(t, err)

		_, err = cco.generateConsolidationTransaction(data)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GenerateConsolidationTransaction fee multisig does not have any utxo", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(validValidatorsChainData, nil).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 2000000,
					},
				},
			}, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{}, error(nil)).Once()

		data, err := cco.createBatchInitialData(ctx, bridgeSmartContractMock, destinationChain, batchID)
		require.NoError(t, err)

		_, err = cco.generateConsolidationTransaction(data)
		require.ErrorContains(t, err, "fee multisig does not have any utxo")
	})

	t.Run("GenerateConsolidationTransaction should pass", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(validValidatorsChainData, nil).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 2_000_000,
					},
				},
			}, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0xFF"),
					},
					Output: indexer.TxOutput{
						Amount: 300_000,
					},
				},
			}, error(nil)).Once()

		data, err := cco.createBatchInitialData(ctx, bridgeSmartContractMock, destinationChain, batchID)
		require.NoError(t, err)

		result, err := cco.generateConsolidationTransaction(data)

		require.NoError(t, err)
		require.NotNil(t, result.TxRaw)
		require.NotEqual(t, "", result.TxHash)
	})

	t.Run("getUTXOsForConsolidation should pass when there is more utxos than maxUtxo", func(t *testing.T) {
		multisigUtxoOutputs, _ := generateSmallUtxoOutputs(10, 50)
		feePayerUtxoOutputs, _ := generateSmallUtxoOutputs(10, 10)

		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return(multisigUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return(feePayerUtxoOutputs, error(nil)).Once()

		multisigUtxos, feeUtxos, err := cco.getUTXOsForConsolidation("aaa", "bbb")
		require.NoError(t, err)
		require.Equal(t, 46, len(multisigUtxos))
		require.Equal(t, 4, len(feeUtxos))
	})

	t.Run("getUTXOsForConsolidation should pass when there is les utxos than maxUtxo", func(t *testing.T) {
		multisigUtxoOutputs, _ := generateSmallUtxoOutputs(10, 30)
		feePayerUtxoOutputs, _ := generateSmallUtxoOutputs(10, 3)

		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return(multisigUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return(feePayerUtxoOutputs, error(nil)).Once()

		multisigUtxos, feeUtxos, err := cco.getUTXOsForConsolidation("aaa", "bbb")
		require.NoError(t, err)
		require.Equal(t, 30, len(multisigUtxos))
		require.Equal(t, 3, len(feeUtxos))
	})

	t.Run("GenerateBatchTransaction execute consolidation and pass", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		batchNonceID := uint64(1)

		multisigUtxoOutputs, multisigUtxoOutputsSum := generateSmallUtxoOutputs(1000, 100)
		feePayerUtxoOutputs, _ := generateSmallUtxoOutputs(1_000_000, 10)

		confirmedTransactions := make([]eth.ConfirmedTransaction, 1)
		confirmedTransactions[0] = eth.ConfirmedTransaction{
			Nonce:       1,
			BlockHeight: big.NewInt(1),
			Receivers: []eth.BridgeReceiver{{
				DestinationAddress: "addr_test1vqeux7xwusdju9dvsj8h7mca9aup2k439kfmwy773xxc2hcu7zy99",
				Amount:             new(big.Int).SetUint64(multisigUtxoOutputsSum - 10000),
			}},
		}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(validValidatorsChainData, nil).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return(multisigUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return(feePayerUtxoOutputs, error(nil)).Once()

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(validValidatorsChainData, nil).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 55}, nil).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return(multisigUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return(feePayerUtxoOutputs, error(nil)).Once()

		result, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.NoError(t, err)
		require.Equal(t, uint8(Consolidation), result.BatchType)
		require.NotNil(t, result.TxRaw)
		require.NotEqual(t, "", result.TxHash)
	})
}

func Test_getNeededUtxos(t *testing.T) {
	inputs := []*indexer.TxInputOutput{
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0x1"), Index: 100},
			Output: indexer.TxOutput{Amount: 10},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0x1"), Index: 0},
			Output: indexer.TxOutput{Amount: 20},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0x2"), Index: 7},
			Output: indexer.TxOutput{Amount: 5},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0x4"), Index: 5},
			Output: indexer.TxOutput{Amount: 30},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0x4"), Index: 6},
			Output: indexer.TxOutput{Amount: 15},
		},
	}

	t.Run("pass", func(t *testing.T) {
		result, err := getNeededUtxos(inputs, 65, 5, 25, 1)

		require.NoError(t, err)
		require.Equal(t, inputs[:len(inputs)-1], result)

		result, err = getNeededUtxos(inputs, 50, 6, 2, 1)

		require.NoError(t, err)
		require.Equal(t, []*indexer.TxInputOutput{inputs[3], inputs[1]}, result)
	})

	t.Run("pass with change", func(t *testing.T) {
		result, err := getNeededUtxos(inputs, 67, 4, 25, 1)

		require.NoError(t, err)
		require.Equal(t, inputs, result)
	})

	t.Run("pass with at least", func(t *testing.T) {
		result, err := getNeededUtxos(inputs, 10, 4, 25, 3)

		require.NoError(t, err)
		require.Equal(t, inputs[:3], result)
	})

	t.Run("not enough sum", func(t *testing.T) {
		_, err := getNeededUtxos(inputs, 160, 5, 25, 1)
		require.ErrorIs(t, err, errUTXOsCouldNotSelect)
	})

	t.Run("errUTXOsLimitReached case change", func(t *testing.T) {
		_, err := getNeededUtxos(inputs, 66, 14, 3, 1)
		require.ErrorIs(t, err, errUTXOsLimitReached)
	})

	t.Run("errUTXOsLimitReached case exactly", func(t *testing.T) {
		_, err := getNeededUtxos(inputs, 80, 20, 3, 1)
		require.ErrorIs(t, err, errUTXOsLimitReached)
	})
}

func Test_getUtxosFromRefundTransactions(t *testing.T) {
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

	_, err = cardano.GenerateWallet(secretsMngr, "prime", true, false)
	require.NoError(t, err)

	configRaw := json.RawMessage([]byte(`{
			"socketPath": "./socket",
			"testnetMagic": 42,
			"minUtxoAmount": 1000
			}`))
	dbMock := &indexer.DatabaseMock{}
	txProviderMock := &cardano.TxProviderTestMock{
		ReturnDefaultParameters: true,
	}

	cco, err := NewCardanoChainOperations(configRaw, dbMock, nil, secretsMngr, "prime", 1*time.Millisecond, hclog.NewNullLogger())
	require.NoError(t, err)

	cco.txProvider = txProviderMock
	cco.config.SlotRoundingThreshold = 100
	cco.config.MaxFeeUtxoCount = 4
	cco.config.MaxUtxoCount = 50
	txs := []eth.ConfirmedTransaction{
		{
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
					Amount:             big.NewInt(100),
				},
				{
					DestinationAddress: "addr128phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtupnz75xxcrtw79hu",
					Amount:             big.NewInt(200),
				},
				{
					DestinationAddress: "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8",
					Amount:             big.NewInt(400),
				},
			},
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "addr1w8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcyjy7wx",
					Amount:             big.NewInt(50),
				},
				{
					DestinationAddress: "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8",
					Amount:             big.NewInt(900),
				},
				{
					DestinationAddress: "addr1z8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gten0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgs9yc0hh",
					Amount:             big.NewInt(0),
				},
			},
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "addr1qx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer3n0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgse35a3x",
					Amount:             big.NewInt(3000),
				},
				{
					// this one will be skipped
					DestinationAddress: "stake178phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcccycj5",
					Amount:             big.NewInt(3000),
				},
			},
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
					Amount:             big.NewInt(2000),
				},
				{
					DestinationAddress: "addr1w8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcyjy7wx",
					Amount:             big.NewInt(170),
				},
				{
					DestinationAddress: "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8",
					Amount:             big.NewInt(10),
				},
			},
		},
	}

	t.Run("getUtxosFromRefundTransactions no refund pass", func(t *testing.T) {
		refundUtxosPerConfirmedTx, err := cco.getUtxosFromRefundTransactions(txs)
		require.NoError(t, err)

		for _, refundUtxo := range refundUtxosPerConfirmedTx {
			require.Empty(t, refundUtxo)
		}
	})

	t.Run("getUtxosFromRefundTransactions with 1 output index pass", func(t *testing.T) {
		refundTokenAmount := uint64(100)
		dbMock.On("GetTxOutput", mock.Anything).Return(indexer.TxOutput{Amount: refundTokenAmount}, nil).Once()

		txs := append(slices.Clone(txs), eth.ConfirmedTransaction{
			TransactionType: uint8(common.RefundConfirmedTxType),
			OutputIndexes:   common.PackNumbersToBytes([]common.TxOutputIndex{2}),
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
				},
			},
		})

		refundUtxosPerConfirmedTx, err := cco.getUtxosFromRefundTransactions(txs)
		require.NoError(t, err)

		for i, refundUtxo := range refundUtxosPerConfirmedTx {
			// if it is the last transaction in the collection (the one assigned as a refund), it should contain outputs, while others should be empty.
			if i == len(txs)-1 {
				require.Equal(t, refundTokenAmount, refundUtxosPerConfirmedTx[i][0].Output.Amount)
			} else {
				require.Empty(t, refundUtxo)
			}
		}

		require.Equal(t, refundTokenAmount, refundUtxosPerConfirmedTx[len(txs)-1][0].Output.Amount)
	})

	t.Run("getUtxosFromRefundTransactions with more output indexes pass", func(t *testing.T) {
		refundTokenAmount := uint64(100)
		dbMock.On("GetTxOutput", mock.Anything).Return(indexer.TxOutput{Amount: refundTokenAmount}, nil).Once()
		dbMock.On("GetTxOutput", mock.Anything).Return(indexer.TxOutput{Amount: 2 * refundTokenAmount}, nil).Once()
		dbMock.On("GetTxOutput", mock.Anything).Return(indexer.TxOutput{Amount: 3 * refundTokenAmount}, nil).Once()

		txs := append(slices.Clone(txs), eth.ConfirmedTransaction{
			TransactionType: uint8(common.RefundConfirmedTxType),
			OutputIndexes:   common.PackNumbersToBytes([]common.TxOutputIndex{2, 3, 5}),
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
				},
			},
		})

		refundUtxosPerConfirmedTx, err := cco.getUtxosFromRefundTransactions(txs)
		require.NoError(t, err)

		for i, refundUtxo := range refundUtxosPerConfirmedTx {
			// if it is the last transaction in the collection (the one assigned as a refund), it should contain outputs, while others should be empty.
			if i == len(txs)-1 {
				for j, txInputOutput := range refundUtxosPerConfirmedTx[i] {
					require.Equal(t, (uint64(j)+1)*refundTokenAmount, txInputOutput.Output.Amount)
				}
			} else {
				require.Empty(t, refundUtxo)
			}
		}
	})
}

func Test_getOutputs(t *testing.T) {
	feeAddr := "0x002"
	minFeeForBridging := uint64(100)
	//nolint:dupl
	txs := []eth.ConfirmedTransaction{
		{
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
					Amount:             big.NewInt(100),
				},
				{
					DestinationAddress: "addr128phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtupnz75xxcrtw79hu",
					Amount:             big.NewInt(200),
				},
				{
					DestinationAddress: "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8",
					Amount:             big.NewInt(400),
				},
			},
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "addr1w8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcyjy7wx",
					Amount:             big.NewInt(50),
				},
				{
					DestinationAddress: "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8",
					Amount:             big.NewInt(900),
				},
				{
					DestinationAddress: "addr1z8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gten0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgs9yc0hh",
					Amount:             big.NewInt(0),
				},
			},
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "addr1qx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer3n0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgse35a3x",
					Amount:             big.NewInt(3000),
				},
				{
					// this one will be skipped
					DestinationAddress: "stake178phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcccycj5",
					Amount:             big.NewInt(3000),
				},
			},
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
					Amount:             big.NewInt(2000),
				},
				{
					DestinationAddress: "addr1w8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcyjy7wx",
					Amount:             big.NewInt(170),
				},
				{
					DestinationAddress: "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8",
					Amount:             big.NewInt(10),
				},
			},
		},
	}

	t.Run("getOutputs pass", func(t *testing.T) {
		res := getOutputs(txs, cardanowallet.MainNetNetwork,
			[][]*indexer.TxInputOutput{}, "", 100, hclog.NewNullLogger())

		assert.Equal(t, uint64(6830), res.Sum[cardanowallet.AdaTokenName])
		assert.Equal(t, []cardanowallet.TxOutput{
			{
				Addr:   "addr128phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtupnz75xxcrtw79hu",
				Amount: 200,
			},
			{
				Addr:   "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
				Amount: 2100,
			},
			{
				Addr:   "addr1qx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer3n0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgse35a3x",
				Amount: 3000,
			},
			{
				Addr:   "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8",
				Amount: 1310,
			},
			{
				Addr:   "addr1w8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcyjy7wx",
				Amount: 220,
			},
		}, res.Outputs)
	})

	t.Run("getOutputs with refund pass", func(t *testing.T) {
		refundTxAmount := uint64(300)
		txs := append(slices.Clone(txs), eth.ConfirmedTransaction{
			TransactionType: uint8(common.RefundConfirmedTxType),
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
					Amount:             new(big.Int).SetUint64(refundTxAmount),
				},
			},
		})

		refundUtxos := make([][]*indexer.TxInputOutput, len(txs))
		refundUtxos[len(refundUtxos)-1] = []*indexer.TxInputOutput{
			{
				Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0x1"), Index: 3},
				Output: indexer.TxOutput{Amount: 250},
			},
			{
				Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0x2"), Index: 2},
				Output: indexer.TxOutput{Amount: 50},
			},
		}

		res := getOutputs(txs, cardanowallet.MainNetNetwork,
			refundUtxos, feeAddr, minFeeForBridging, hclog.NewNullLogger())

		assert.Equal(t, uint64(7030), res.Sum[cardanowallet.AdaTokenName])
		assert.Equal(t, []cardanowallet.TxOutput{
			{
				Addr:   "addr128phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtupnz75xxcrtw79hu",
				Amount: 200,
			},
			{
				Addr:   "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
				Amount: 2100 + refundTxAmount - minFeeForBridging,
			},
			{
				Addr:   "addr1qx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer3n0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgse35a3x",
				Amount: 3000,
			},
			{
				Addr:   "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8",
				Amount: 1310,
			},
			{
				Addr:   "addr1w8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcyjy7wx",
				Amount: 220,
			},
		}, res.Outputs)
	})

	t.Run("getOutputs with refund pass with tokens", func(t *testing.T) {
		refundTxAmount := uint64(300)
		txs = append(txs, eth.ConfirmedTransaction{
			TransactionType: uint8(common.RefundConfirmedTxType),
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
					Amount:             new(big.Int).SetUint64(refundTxAmount),
				},
			},
		})

		refundUtxos := make([][]*indexer.TxInputOutput, len(txs))
		refundUtxos[len(refundUtxos)-1] = []*indexer.TxInputOutput{
			{
				Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("0x1"), Index: 3},
				Output: indexer.TxOutput{
					Amount: 200,
					Tokens: []indexer.TokenAmount{
						{
							PolicyID: "1",
							Name:     "1",
							Amount:   15,
						},
					},
				},
			},
			{
				Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("0x21"), Index: 2},
				Output: indexer.TxOutput{
					Amount: 100,
					Tokens: []indexer.TokenAmount{
						{
							PolicyID: "1",
							Name:     "3",
							Amount:   15,
						},
					},
				},
			},
		}

		res := getOutputs(txs, cardanowallet.MainNetNetwork,
			refundUtxos, feeAddr, minFeeForBridging, hclog.NewNullLogger())

		assert.Equal(t, uint64(7030), res.Sum[cardanowallet.AdaTokenName])
		assert.Equal(t, []cardanowallet.TxOutput{
			{
				Addr:   "addr128phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtupnz75xxcrtw79hu",
				Amount: 200,
			},
			{
				Addr:   "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
				Amount: 2100 + refundTxAmount - minFeeForBridging,
				Tokens: []cardanowallet.TokenAmount{
					cardanowallet.NewTokenAmount(cardanowallet.NewToken("1", "1"), 15),
					cardanowallet.NewTokenAmount(cardanowallet.NewToken("1", "3"), 15),
				},
			},
			{
				Addr:   "addr1qx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer3n0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgse35a3x",
				Amount: 3000,
			},
			{
				Addr:   "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8",
				Amount: 1310,
			},
			{
				Addr:   "addr1w8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcyjy7wx",
				Amount: 220,
			},
		}, res.Outputs)
	})
}

func Test_getUTXOs(t *testing.T) {
	dbMock := &indexer.DatabaseMock{}
	multisigAddr := "0x001"
	feeAddr := "0x002"
	testErr := errors.New("test err")
	ops := &CardanoChainOperations{
		db: dbMock,
		config: &cardano.CardanoChainConfig{
			NoBatchPeriodPercent: 0.1,
			MaxFeeUtxoCount:      4,
			MaxUtxoCount:         50,
		},
		logger: hclog.NewNullLogger(),
	}

	t.Run("GetAllTxOutputs multisig error", func(t *testing.T) {
		dbMock.On("GetAllTxOutputs", multisigAddr, true).Return(([]*indexer.TxInputOutput)(nil), testErr).Once()

		_, _, err := ops.getUTXOs(multisigAddr, feeAddr, nil, 2_000_000)
		require.Error(t, err)
	})

	t.Run("GetAllTxOutputs fee error", func(t *testing.T) {
		dbMock.On("GetAllTxOutputs", multisigAddr, true).Return([]*indexer.TxInputOutput{}, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).Return(([]*indexer.TxInputOutput)(nil), testErr).Once()

		_, _, err := ops.getUTXOs(multisigAddr, feeAddr, nil, 2_000_000)
		require.Error(t, err)
	})

	t.Run("pass", func(t *testing.T) {
		expectedUtxos := []*indexer.TxInputOutput{
			{
				Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0x1"), Index: 2},
				Output: indexer.TxOutput{Amount: 1_000_000, Slot: 80},
			},
			{
				Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0x1"), Index: 3},
				Output: indexer.TxOutput{Amount: 1_000_000, Slot: 1900},
			},
			{
				Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0xAA"), Index: 100},
				Output: indexer.TxOutput{Amount: 10},
			},
		}
		allMultisigUtxos := expectedUtxos[0:2]
		allFeeUtxos := expectedUtxos[2:]

		dbMock.On("GetAllTxOutputs", multisigAddr, true).Return(allMultisigUtxos, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).Return(allFeeUtxos, error(nil)).Once()

		multisigUtxos, feeUtxos, err := ops.getUTXOs(multisigAddr, feeAddr, nil, 2_000_000)

		require.NoError(t, err)
		require.Equal(t, expectedUtxos[0:2], multisigUtxos)
		require.Equal(t, expectedUtxos[2:], feeUtxos)
	})

	t.Run("pass with refund", func(t *testing.T) {
		expectedUtxos := []*indexer.TxInputOutput{
			{
				Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0x1"), Index: 2},
				Output: indexer.TxOutput{Amount: 1_000_000, Slot: 80},
			},
			{
				Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0x1"), Index: 3},
				Output: indexer.TxOutput{Amount: 1_000_000, Slot: 1900},
			},
			{
				Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0xAA"), Index: 100},
				Output: indexer.TxOutput{Amount: 10},
			},
		}
		refundUtxos := []*indexer.TxInputOutput{
			{
				Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0xAAEB"), Index: 121},
				Output: indexer.TxOutput{Amount: 30},
			},
		}
		allMultisigUtxos := expectedUtxos[0:2]
		allFeeUtxos := expectedUtxos[2:]

		dbMock.On("GetAllTxOutputs", multisigAddr, true).Return(allMultisigUtxos, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).Return(allFeeUtxos, error(nil)).Once()

		multisigUtxos, feeUtxos, err := ops.getUTXOs(multisigAddr, feeAddr, refundUtxos, 2_000_000)

		require.NoError(t, err)
		require.Equal(t, expectedUtxos[2:], feeUtxos)
		require.Equal(t, append(expectedUtxos[0:2], refundUtxos...), multisigUtxos)
	})
}

func generateSmallUtxoOutputs(value uint64, n uint64) ([]*indexer.TxInputOutput, uint64) {
	utxoOutput := make([]*indexer.TxInputOutput, 0, n)
	returnSum := uint64(0)

	for i := uint64(0); i < n; i++ {
		utxoOutput = append(utxoOutput,
			&indexer.TxInputOutput{
				Input: indexer.TxInput{
					Hash: indexer.NewHashFromHexString(fmt.Sprintf("0x00%d", i)),
				},
				Output: indexer.TxOutput{
					Amount: value,
				},
			},
		)
		returnSum += value
	}

	return utxoOutput, returnSum
}

func Test_CreateValidatorSetChangeTx(t *testing.T) {
	testDir, err := os.MkdirTemp("", "bat-chain-ops-tx")
	require.NoError(t, err)

	bridgeSmartContractMock := &eth.BridgeSmartContractMock{}

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	secretsMngr, err := secretsHelper.CreateSecretsManager(&secrets.SecretsManagerConfig{
		Path: filepath.Join(testDir, "stp"),
		Type: secrets.Local,
	})
	require.NoError(t, err)

	wallet, err := cardano.GenerateWallet(secretsMngr, "prime", true, false)
	require.NoError(t, err)

	wallet2, err := cardano.GenerateWallet(secretsMngr, "prime1", true, false)
	require.NoError(t, err)

	txProviderMock := &cardano.TxProviderTestMock{
		ReturnDefaultParameters: true,
	}

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

	validatorPerChain := validatorobserver.ValidatorsPerChain{
		"prime": validatorobserver.ValidatorsChainData{
			Keys: validatorsChainData,
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

	newValidatorPerChain := validatorobserver.ValidatorsPerChain{
		"prime": validatorobserver.ValidatorsChainData{
			Keys: newValidatorChainData,
		},
	}

	indxUpdaterMock := &IndexerUpdaterMock{}
	indxUpdaterMock.On("AddNewAddressesOfInterest", mock.Anything, mock.Anything).Return()

	cco, err := NewCardanoChainOperations(configRaw, nil, indxUpdaterMock, secretsMngr, "prime", 1*time.Millisecond, hclog.NewNullLogger())
	require.NoError(t, err)

	cco.txProvider = txProviderMock

	_, activeAddresses, err := generatePolicyAndMultisig(&validatorPerChain, "prime", cco.cardanoCliBinary, cco.config.NetworkMagic)
	require.NoError(t, err)

	_, newAddresses, err := generatePolicyAndMultisig(&newValidatorPerChain, "prime", cco.cardanoCliBinary, cco.config.NetworkMagic)
	require.NoError(t, err)

	bridgeSmartContractMock.On("GetValidatorsChainData", mock.Anything, mock.Anything).Return(validatorsChainData, nil)
	bridgeSmartContractMock.On("SubmitSignedBatch", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	const nextBatchID = 1

	runTest := func(multisigNum, feeNum uint) *indexer.TxInfo {
		dbMock := &indexer.DatabaseMock{}
		cco.db = dbMock

		multisig := generateUTXO(multisigNum)
		fee := generateUTXO(feeNum)

		dbMock.On("GetAllTxOutputs", activeAddresses.Multisig.Payment, mock.Anything).Return(multisig, nil)
		dbMock.On("GetAllTxOutputs", activeAddresses.Fee.Payment, mock.Anything).Return(fee, nil)

		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{
			BlockSlot: 4,
			BlockHash: indexer.Hash{},
		}, nil)

		generatedData, err := cco.CreateValidatorSetChangeTx(context.TODO(), common.ChainIDStrPrime,
			nextBatchID, bridgeSmartContractMock, validatorobserver.ValidatorsPerChain{
				common.ChainIDStrPrime: {
					Keys:       newValidatorChainData,
					SlotNumber: 0,
				},
			})
		require.NoError(t, err)

		if generatedData.BatchType == uint8(ValidatorSetFinal) {
			return nil
		}

		info, err := gouroboros.ParseTxInfo(generatedData.TxRaw, true)
		require.NoError(t, err)

		return &info
	}

	t.Run("Test 10 multisig, 2 fee UTXOs", func(t *testing.T) {
		info := runTest(10, 2)

		require.Equal(t, len(info.Outputs), 2)
		require.Equal(t, info.Outputs[0].Address, newAddresses.Multisig.Payment)
		require.Equal(t, info.Outputs[1].Address, activeAddresses.Fee.Payment)

		// max utxos = 5 => output = 3 multisig utxos & 2 fee utxos to old fee address
		require.Equal(t, info.Outputs[0].Amount, uint64(3*1000000))
		require.True(t, info.Outputs[1].Amount > 0)
	})

	t.Run("Test 0 multisig, 10 fee UTXOs", func(t *testing.T) {
		info := runTest(0, 10)

		// max utxos = 5 => output = 5 utxos to new fee address reduced for fee amount
		require.Equal(t, len(info.Outputs), 1)
		require.Equal(t, info.Outputs[0].Address, newAddresses.Fee.Payment)
		require.True(t, info.Outputs[0].Amount < 5*1000000)
	})

	t.Run("Test 10 multisig, 10 fee UTXOs", func(t *testing.T) {
		info := runTest(10, 10)

		require.Equal(t, len(info.Outputs), 2)
		require.Equal(t, info.Outputs[0].Address, newAddresses.Multisig.Payment)
		require.Equal(t, info.Outputs[1].Address, activeAddresses.Fee.Payment)

		// max utxos = 5 => output = 3 multisig utxos & 2 fee utxos to old fee address
		require.Equal(t, info.Outputs[0].Amount, uint64(3*1000000))
		require.True(t, info.Outputs[1].Amount > 0)
	})

	t.Run("Test 2 multisig, 2 fee UTXOs", func(t *testing.T) {
		info := runTest(2, 2)

		require.Equal(t, len(info.Outputs), 2)
		require.Equal(t, info.Outputs[0].Address, newAddresses.Multisig.Payment)
		require.Equal(t, info.Outputs[1].Address, activeAddresses.Fee.Payment)

		// max utxos = 5 => output = 2 multisig utxos & 2 fee utxos to old fee address
		require.Equal(t, info.Outputs[0].Amount, uint64(2*1000000))
		require.True(t, info.Outputs[1].Amount > 0)
	})

	t.Run("Test 10 multisig, 0 fee UTXOs", func(t *testing.T) {
		info := runTest(10, 0)

		// no fee utxos => finalize batch
		require.Nil(t, info)
	})

	t.Run("Test 10 multisig, 1 fee UTXOs", func(t *testing.T) {
		info := runTest(10, 1)

		// 1 fee utxo < MinUtxoAmountDefault*2 => finalize batch
		require.Nil(t, info)
	})
}

func TestGetUTXOsForValidatorChange(t *testing.T) {
	db := indexer.DatabaseMock{}
	db.On("GetAllTxOutputs", "fee", mock.Anything).Return([]*indexer.TxInputOutput{
		{},
		{},
	}, nil)
	db.On("GetAllTxOutputs", "multisig", mock.Anything).Return([]*indexer.TxInputOutput{
		{
			Output: indexer.TxOutput{Address: "1", Slot: 56}, // should be omitted from the returned UTXOs
		},
		{
			Output: indexer.TxOutput{Address: "2", Slot: 20},
		},
		{
			Output: indexer.TxOutput{Address: "3", Slot: 83}, // should be omitted from the returned UTXOs
		},
		{
			Output: indexer.TxOutput{Address: "4", Slot: 44},
		},
		{
			Output: indexer.TxOutput{Address: "5", Slot: 12},
		},
	}, nil)

	cco := CardanoChainOperations{
		logger: hclog.NewNullLogger(),
		db:     &db,
		config: &cardano.CardanoChainConfig{
			MaxUtxoCount:    10,
			MaxFeeUtxoCount: 3,
		},
	}

	ms, fee, onlyFee, err := cco.getUTXOsForValidatorChange("multisig", "fee", 50)

	require.NoError(t, err)
	require.Len(t, ms, 3)

	correctAddr := []string{"2", "4", "5"}

	for _, o := range ms {
		require.Contains(t, correctAddr, o.Output.Address)
	}

	require.Len(t, fee, 2)
	require.False(t, onlyFee)
}
