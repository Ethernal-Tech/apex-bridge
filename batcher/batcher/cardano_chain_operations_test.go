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
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	secretsHelper "github.com/Ethernal-Tech/cardano-infrastructure/secrets/helper"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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

	cco, err := NewCardanoChainOperations(configRaw, dbMock, secretsMngr, "prime", hclog.NewNullLogger())
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
		require.ErrorContains(t, err, "verifying key of current batcher wasn't found in validators data queried from smart contract")
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

		_, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "verifying fee key of current batcher wasn't found in validators data queried from smart contract")
	})

	t.Run("GetLatestBlockPoint return error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		getValidatorsCardanoDataRet := []eth.ValidatorChainData{
			{
				Key: [4]*big.Int{
					new(big.Int).SetBytes(wallet.MultiSig.VerificationKey),
					new(big.Int).SetBytes(wallet.MultiSigFee.VerificationKey),
				},
			},
		}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil).Once()
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
		getValidatorsCardanoDataRet := []eth.ValidatorChainData{
			{
				Key: [4]*big.Int{
					new(big.Int).SetBytes(wallet.MultiSig.VerificationKey),
					new(big.Int).SetBytes(wallet.MultiSigFee.VerificationKey),
				},
			},
		}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput(nil), testError).Once()

		_, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GenerateBatchTransaction fee multisig does not have any utxo", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		getValidatorsCardanoDataRet := []eth.ValidatorChainData{
			{
				Key: [4]*big.Int{
					new(big.Int).SetBytes(wallet.MultiSig.VerificationKey),
					new(big.Int).SetBytes(wallet.MultiSigFee.VerificationKey),
				},
			},
		}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil).Once()
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
		getValidatorsCardanoDataRet := []eth.ValidatorChainData{
			{
				Key: [4]*big.Int{
					new(big.Int).SetBytes(wallet.MultiSig.VerificationKey),
					new(big.Int).SetBytes(wallet.MultiSigFee.VerificationKey),
				},
			},
		}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil).Once()
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
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0xFF"),
					},
					Output: indexer.TxOutput{
						Amount: 20000,
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

	cco, err := NewCardanoChainOperations(configRaw, dbMock, secretsMngr, "prime", hclog.NewNullLogger())
	require.NoError(t, err)

	cco.txProvider = txProviderMock

	getValidatorsCardanoDataRetGood := []eth.ValidatorChainData{
		{
			Key: [4]*big.Int{
				new(big.Int).SetBytes(wallet.MultiSig.VerificationKey),
				new(big.Int).SetBytes(wallet.MultiSigFee.VerificationKey),
			},
		},
	}

	testError := errors.New("test err")
	batchID := uint64(1)
	destinationChain := common.ChainIDStrVector

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
					new(big.Int), new(big.Int),
				},
			},
		}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil).Once()

		_, err := cco.createBatchInitialData(ctx, bridgeSmartContractMock, destinationChain, batchID)
		require.ErrorContains(t, err, "verifying key of current batcher wasn't found in validators data queried from smart contract")
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
		require.ErrorContains(t, err, "verifying fee key of current batcher wasn't found in validators data queried from smart contract")
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
			Return(getValidatorsCardanoDataRetGood, nil).Once()

		_, err := cco.createBatchInitialData(ctx, bridgeSmartContractMock, destinationChain, batchID)
		require.ErrorIs(t, err, desiredErr)
	})

	t.Run("good", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).
			Return(getValidatorsCardanoDataRetGood, nil).Once()

		data, err := cco.createBatchInitialData(ctx, bridgeSmartContractMock, destinationChain, batchID)
		require.NoError(t, err)

		assert.Contains(t, data.MultisigAddress, "addr_test1")
		assert.Contains(t, data.MultisigFeeAddress, "addr_test1")
		assert.NotEqual(t, data.MultisigAddress, data.MultisigFeeAddress)
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

	cco, err := NewCardanoChainOperations(configRaw, dbMock, secretsMngr, "prime", hclog.NewNullLogger())
	require.NoError(t, err)

	cco.txProvider = txProviderMock
	cco.config.SlotRoundingThreshold = 100
	cco.config.MaxFeeUtxoCount = 4
	cco.config.MaxUtxoCount = 50

	testError := errors.New("test err")
	batchID := uint64(1)
	destinationChain := common.ChainIDStrVector

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	t.Run("GetLatestBlockPoint return error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		getValidatorsCardanoDataRet := []eth.ValidatorChainData{
			{
				Key: [4]*big.Int{
					new(big.Int).SetBytes(wallet.MultiSig.VerificationKey),
					new(big.Int).SetBytes(wallet.MultiSigFee.VerificationKey),
				},
			},
		}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil).Once()
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
		getValidatorsCardanoDataRet := []eth.ValidatorChainData{
			{
				Key: [4]*big.Int{
					new(big.Int).SetBytes(wallet.MultiSig.VerificationKey),
					new(big.Int).SetBytes(wallet.MultiSigFee.VerificationKey),
				},
			},
		}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil).Once()
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
		getValidatorsCardanoDataRet := []eth.ValidatorChainData{
			{
				Key: [4]*big.Int{
					new(big.Int).SetBytes(wallet.MultiSig.VerificationKey),
					new(big.Int).SetBytes(wallet.MultiSigFee.VerificationKey),
				},
			},
		}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil).Once()
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
		getValidatorsCardanoDataRet := []eth.ValidatorChainData{
			{
				Key: [4]*big.Int{
					new(big.Int).SetBytes(wallet.MultiSig.VerificationKey),
					new(big.Int).SetBytes(wallet.MultiSigFee.VerificationKey),
				},
			},
		}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil).Once()
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
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0xFF"),
					},
					Output: indexer.TxOutput{
						Amount: 20000,
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
		getValidatorsCardanoDataRet := []eth.ValidatorChainData{
			{
				Key: [4]*big.Int{
					new(big.Int).SetBytes(wallet.MultiSig.VerificationKey),
					new(big.Int).SetBytes(wallet.MultiSigFee.VerificationKey),
				},
			},
		}

		batchNonceID := uint64(1)

		multisigUtxoOutputs, multisigUtxoOutputsSum := generateSmallUtxoOutputs(1000, 100)
		feePayerUtxoOutputs, _ := generateSmallUtxoOutputs(2000, 10)

		confirmedTransactions := make([]eth.ConfirmedTransaction, 1)
		confirmedTransactions[0] = eth.ConfirmedTransaction{
			Nonce:       1,
			BlockHeight: big.NewInt(1),
			Receivers: []eth.BridgeReceiver{{
				DestinationAddress: "addr_test1vqeux7xwusdju9dvsj8h7mca9aup2k439kfmwy773xxc2hcu7zy99",
				Amount:             new(big.Int).SetUint64(multisigUtxoOutputsSum - 10000),
			}},
		}

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return(multisigUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return(feePayerUtxoOutputs, error(nil)).Once()

		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 55}, nil).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return(multisigUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return(feePayerUtxoOutputs, error(nil)).Once()

		result, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.NoError(t, err)
		require.Equal(t, true, result.IsConsolidation)
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
		result, err := getNeededUtxos(inputs, 65, 5, 5, 30, 1)

		require.NoError(t, err)
		require.Equal(t, inputs[:len(inputs)-1], result)

		result, err = getNeededUtxos(inputs, 50, 6, 0, 2, 1)

		require.NoError(t, err)
		require.Equal(t, []*indexer.TxInputOutput{inputs[3], inputs[1]}, result)
	})

	t.Run("pass with change", func(t *testing.T) {
		result, err := getNeededUtxos(inputs, 67, 4, 5, 30, 1)

		require.NoError(t, err)
		require.Equal(t, inputs, result)
	})

	t.Run("pass with at least", func(t *testing.T) {
		result, err := getNeededUtxos(inputs, 10, 4, 5, 30, 3)

		require.NoError(t, err)
		require.Equal(t, inputs[:3], result)
	})

	t.Run("not enough sum", func(t *testing.T) {
		_, err := getNeededUtxos(inputs, 160, 5, 5, 30, 1)
		require.ErrorIs(t, err, errUTXOsCouldNotSelect)
	})
}

func Test_getOutputs(t *testing.T) {
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

	res := getOutputs(txs, cardanowallet.MainNetNetwork, hclog.NewNullLogger())

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
	txOutputs := cardano.TxOutputs{
		Outputs: []cardanowallet.TxOutput{
			{}, {}, {},
		},
		Sum: map[string]uint64{
			cardanowallet.AdaTokenName: 2_000_000,
		},
	}

	t.Run("GetAllTxOutputs multisig error", func(t *testing.T) {
		dbMock.On("GetAllTxOutputs", multisigAddr, true).Return(([]*indexer.TxInputOutput)(nil), testErr).Once()

		_, _, err := ops.getUTXOs(multisigAddr, feeAddr, txOutputs)
		require.Error(t, err)
	})

	t.Run("GetAllTxOutputs fee error", func(t *testing.T) {
		dbMock.On("GetAllTxOutputs", multisigAddr, true).Return([]*indexer.TxInputOutput{}, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).Return(([]*indexer.TxInputOutput)(nil), testErr).Once()

		_, _, err := ops.getUTXOs(multisigAddr, feeAddr, txOutputs)
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

		multisigUtxos, feeUtxos, err := ops.getUTXOs(multisigAddr, feeAddr, txOutputs)

		require.NoError(t, err)
		require.Equal(t, expectedUtxos[0:2], multisigUtxos)
		require.Equal(t, expectedUtxos[2:], feeUtxos)
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
