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
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
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

	cco, err := NewCardanoChainOperations(configRaw, dbMock, secretsMngr, "prime", &CardanoChainOperationReactorStrategy{}, hclog.NewNullLogger())
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
						Amount: 2_000_000,
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

	cco, err := NewCardanoChainOperations(configRaw, dbMock, secretsMngr, "prime", &CardanoChainOperationReactorStrategy{}, hclog.NewNullLogger())
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

	t.Run("GetValidatorsChainData returns error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetValidatorsChainData", ctx, destinationChain).Return(nil, testError).Once()

		_, err := cco.generateConsolidationTransaction(ctx, bridgeSmartContractMock, destinationChain, batchID)
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

		_, err := cco.generateConsolidationTransaction(ctx, bridgeSmartContractMock, destinationChain, batchID)
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

		_, err := cco.generateConsolidationTransaction(ctx, bridgeSmartContractMock, destinationChain, batchID)
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

		_, err := cco.generateConsolidationTransaction(ctx, bridgeSmartContractMock, destinationChain, batchID)
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

		_, err := cco.generateConsolidationTransaction(ctx, bridgeSmartContractMock, destinationChain, batchID)
		require.Error(t, err)
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

		_, err := cco.generateConsolidationTransaction(ctx, bridgeSmartContractMock, destinationChain, batchID)
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
						Amount: 200000,
					},
				},
			}, error(nil)).Once()

		result, err := cco.generateConsolidationTransaction(ctx, bridgeSmartContractMock, destinationChain, batchID)
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

		multisigUtxoOutputs, _ := generateSmallUtxoOutputs(10000, 100)
		feePayerUtxoOutputs, _ := generateSmallUtxoOutputs(200000, 10)

		confirmedTransactions := make([]eth.ConfirmedTransaction, 1)
		confirmedTransactions[0] = eth.ConfirmedTransaction{
			Nonce:       1,
			BlockHeight: big.NewInt(1),
			Receivers: []eth.BridgeReceiver{{
				DestinationAddress: "addr_test1vqeux7xwusdju9dvsj8h7mca9aup2k439kfmwy773xxc2hcu7zy99",
				Amount:             new(big.Int).SetUint64(10000),
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

func TestSkylineConsolidation(t *testing.T) {
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

	_ = wallet

	token1, _ := cardanowallet.NewTokenWithFullName("29f8873beb52e126f207a2dfd50f7cff556806b5b4cba9834a7b26a8.4b6173685f546f6b656e", true)
	token2, _ := cardanowallet.NewTokenWithFullName("29f8873beb52e126f207a2dfd50f7cff556806b5b4cba9834a7b26a8.526f75746533", true)

	configRaw := json.RawMessage([]byte(`{
			"socketPath": "./socket",
			"testnetMagic": 42,
			"minUtxoAmount": 1000
			}`))
	dbMock := &indexer.DatabaseMock{}
	txProviderMock := &cardano.TxProviderTestMock{
		ReturnDefaultParameters: true,
	}

	cco, err := NewCardanoChainOperations(
		configRaw, dbMock, secretsMngr, common.ChainIDStrPrime, &CardanoChainOperationSkylineStrategy{}, hclog.NewNullLogger())
	require.NoError(t, err)

	destinationChain := common.ChainIDStrCardano
	cco.txProvider = txProviderMock
	cco.config.SlotRoundingThreshold = 100
	cco.config.MaxFeeUtxoCount = 4
	cco.config.MaxUtxoCount = 40
	cco.config.NativeTokens = []sendtx.TokenExchangeConfig{
		{
			DstChainID: destinationChain,
			TokenName:  token1.String(),
		},
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	t.Run("getUTXOsForConsolidation should pass when there is more utxos than maxUtxo", func(t *testing.T) {
		multisigUtxoOutputs, _ := generateSmallUtxoOutputs(10, 100, token1, token2)
		feePayerUtxoOutputs, _ := generateSmallUtxoOutputs(10, 10)

		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return(multisigUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return(feePayerUtxoOutputs, error(nil)).Once()

		multisigUtxos, feeUtxos, err := cco.getUTXOsForConsolidation("aaa", "bbb")
		require.NoError(t, err)
		require.Equal(t, cco.config.MaxUtxoCount-len(feeUtxos), len(multisigUtxos))
		require.Equal(t, 4, len(feeUtxos))
	})

	t.Run("getUTXOsForConsolidation should pass when there is les utxos than maxUtxo", func(t *testing.T) {
		multisigUtxoOutputs, _ := generateSmallUtxoOutputs(10, 30, token1, token2)
		feePayerUtxoOutputs, _ := generateSmallUtxoOutputs(10, 3)

		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return(multisigUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return(feePayerUtxoOutputs, error(nil)).Once()

		multisigUtxos, feeUtxos, err := cco.getUTXOsForConsolidation("aaa", "bbb")
		require.NoError(t, err)
		require.Equal(t, 15, len(multisigUtxos))
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

		multisigUtxoOutputs, _ := generateSmallUtxoOutputs(60_000, 100, token1, token2)
		feePayerUtxoOutputs, _ := generateSmallUtxoOutputs(200000, 10)

		confirmedTransactions := make([]eth.ConfirmedTransaction, 1)
		confirmedTransactions[0] = eth.ConfirmedTransaction{
			Nonce:         1,
			BlockHeight:   big.NewInt(1),
			SourceChainId: common.ChainIDIntCardano,
			Receivers: []eth.BridgeReceiver{{
				DestinationAddress: "addr_test1vqeux7xwusdju9dvsj8h7mca9aup2k439kfmwy773xxc2hcu7zy99",
				Amount:             big.NewInt(1_250_000),
				AmountWrapped:      big.NewInt(2_500_000),
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
	const minUtxoAmount = 5

	desiredAmounts := map[string]uint64{
		cardanowallet.AdaTokenName: 0,
	}
	inputs := []*indexer.TxInputOutput{
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("01"), Index: 100},
			Output: indexer.TxOutput{Amount: 100},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("02"), Index: 0},
			Output: indexer.TxOutput{Amount: 50},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("03"), Index: 7},
			Output: indexer.TxOutput{Amount: 150},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("04"), Index: 5},
			Output: indexer.TxOutput{Amount: 200},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("05"), Index: 6},
			Output: indexer.TxOutput{Amount: 160},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("06"), Index: 8},
			Output: indexer.TxOutput{Amount: 400},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("07"), Index: 10},
			Output: indexer.TxOutput{Amount: 200},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("08"), Index: 9},
			Output: indexer.TxOutput{Amount: 50},
		},
	}

	t.Run("exact amount", func(t *testing.T) {
		desiredAmounts[cardanowallet.AdaTokenName] = 605
		result, err := getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 4, 1)

		require.NoError(t, err)
		require.Equal(t, []*indexer.TxInputOutput{inputs[0], inputs[2], inputs[3], inputs[4]}, result)

		desiredAmounts[cardanowallet.AdaTokenName] = 245
		result, err = getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 2, 1)

		require.NoError(t, err)
		require.Equal(t, []*indexer.TxInputOutput{inputs[0], inputs[2]}, result)
	})

	t.Run("pass with change", func(t *testing.T) {
		desiredAmounts[cardanowallet.AdaTokenName] = 706
		result, err := getNeededUtxos(inputs, desiredAmounts, 4, 3, 1)

		require.NoError(t, err)
		require.Equal(t, inputs[3:6], result)
	})

	t.Run("pass with at least", func(t *testing.T) {
		desiredAmounts[cardanowallet.AdaTokenName] = 5
		result, err := getNeededUtxos(inputs, desiredAmounts, 4, 30, 3)

		require.NoError(t, err)
		require.Equal(t, inputs[:3], result)
	})

	t.Run("not enough sum", func(t *testing.T) {
		desiredAmounts[cardanowallet.AdaTokenName] = 1550
		_, err := getNeededUtxos(inputs, desiredAmounts, 5, 30, 1)
		require.ErrorIs(t, err, cardanowallet.ErrUTXOsCouldNotSelect)
	})
}

func Test_getNeededSkylineUtxos(t *testing.T) {
	inputs := []*indexer.TxInputOutput{
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("01"), Index: 100},
			Output: indexer.TxOutput{Amount: 100},
		},
		{
			Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("02"), Index: 0},
			Output: indexer.TxOutput{
				Amount: 50,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "1",
						Name:     "1",
						Amount:   100,
					},
				},
			},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("03"), Index: 7},
			Output: indexer.TxOutput{Amount: 150},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("04"), Index: 5},
			Output: indexer.TxOutput{Amount: 200},
		},
		{
			Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("05"), Index: 6},
			Output: indexer.TxOutput{
				Amount: 160,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "1",
						Name:     "1",
						Amount:   50,
					},
				},
			},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("06"), Index: 8},
			Output: indexer.TxOutput{Amount: 400},
		},
		{
			Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("07"), Index: 10},
			Output: indexer.TxOutput{
				Amount: 200,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "1",
						Name:     "1",
						Amount:   400,
					},
				},
			},
		},
		{
			Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("08"), Index: 9},
			Output: indexer.TxOutput{
				Amount: 50,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "1",
						Name:     "1",
						Amount:   200,
					},
				},
			},
		},
	}

	t.Run("pass", func(t *testing.T) {
		const minUtxoAmount = 5

		desiredAmounts := map[string]uint64{
			cardanowallet.AdaTokenName: 590,
		}

		result, err := getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 4, 1)

		require.NoError(t, err)
		require.Equal(t, []*indexer.TxInputOutput{inputs[0], inputs[2], inputs[3], inputs[4]}, result)

		desiredAmounts = map[string]uint64{
			cardanowallet.AdaTokenName: minUtxoAmount,
			"1.31":                     100,
		}

		result, err = getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 1, 1)

		require.NoError(t, err)
		require.Equal(t, []*indexer.TxInputOutput{inputs[1]}, result)
	})

	t.Run("pass with change", func(t *testing.T) {
		const minUtxoAmount = 4

		desiredAmounts := map[string]uint64{
			cardanowallet.AdaTokenName: minUtxoAmount,
			"1.31":                     350,
		}
		result, err := getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 2, 1)

		require.NoError(t, err)
		require.Equal(t, []*indexer.TxInputOutput{inputs[1], inputs[6]}, result)
	})

	t.Run("pass with at least", func(t *testing.T) {
		const minUtxoAmount = 4

		desiredAmounts := map[string]uint64{
			cardanowallet.AdaTokenName: minUtxoAmount,
			"1.31":                     20,
		}
		result, err := getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 30, 3)

		require.NoError(t, err)
		require.Equal(t, inputs[:3], result)

		desiredAmounts = map[string]uint64{
			cardanowallet.AdaTokenName: 12,
		}
		result, err = getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 30, 3)

		require.NoError(t, err)
		require.Equal(t, inputs[:3], result)
	})

	t.Run("not enough sum", func(t *testing.T) {
		const minUtxoAmount = 5

		desiredAmounts := map[string]uint64{
			cardanowallet.AdaTokenName: 1600,
		}
		_, err := getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 30, 1)
		require.ErrorIs(t, err, cardanowallet.ErrUTXOsCouldNotSelect)

		desiredAmounts = map[string]uint64{
			cardanowallet.AdaTokenName: minUtxoAmount,
			"1.31":                     2500,
		}
		_, err = getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 30, 1)
		require.ErrorIs(t, err, cardanowallet.ErrUTXOsCouldNotSelect)
	})
}

func Test_reactorGetOutputs(t *testing.T) {
	configRaw := json.RawMessage([]byte(`{
			"socketPath": "./socket",
			"testnetMagic": 42,
			"minUtxoAmount": 1000
			}`))

	cardanoConfig, err := cardano.NewCardanoChainConfig(configRaw)
	require.NoError(t, err)

	cco := &CardanoChainOperations{
		strategy: &CardanoChainOperationReactorStrategy{},
		config:   cardanoConfig,
	}
	cco.config.NetworkID = cardanowallet.MainNetNetwork

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

	res, err := cco.strategy.GetOutputs(txs, cco.config, hclog.NewNullLogger())
	require.NoError(t, err)

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

func Test_skylineGetOutputs(t *testing.T) {
	// vector -> prime
	const (
		addr1 = "vector_test1vgxk3ha6hmftgjzrjlrxrndmqrg43y862pu909r87q8kpas0c0mzc"
		addr2 = "vector_test1v25acu09yv4z2jc026ss5hhgfu5nunfp9z7gkamae43t6fc8gx3pf"
		addr3 = "vector_test1w2h482rf4gf44ek0rekamxksulazkr64yf2fhmm7f5gxjpsdm4zsg"
	)

	policyID := "584ffccecba8a7c6a18037152119907b6b5c2ed063798ee68b012c41"
	tokenName, _ := hex.DecodeString("526f75746533")
	token := cardanowallet.NewToken(policyID, string(tokenName))
	bactherStrategyPrime := &CardanoChainOperationSkylineStrategy{}
	config := &cardano.CardanoChainConfig{
		NetworkID: cardanowallet.VectorTestNetNetwork,
		NativeTokens: []sendtx.TokenExchangeConfig{
			{
				DstChainID: common.ChainIDStrVector,
				TokenName:  token.String(),
			},
		},
	}

	txs := []eth.ConfirmedTransaction{
		{
			SourceChainId: common.ChainIDIntVector,
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: addr1,
					Amount:             big.NewInt(100),
					AmountWrapped:      big.NewInt(200),
				},
				{
					DestinationAddress: addr2,
					Amount:             big.NewInt(51),
					AmountWrapped:      big.NewInt(102),
				},
			},
		},
		{
			SourceChainId: common.ChainIDIntVector,
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: addr3,
					Amount:             big.NewInt(8),
				},
				{
					DestinationAddress: addr1,
					Amount:             big.NewInt(2),
					AmountWrapped:      big.NewInt(5),
				},
			},
		},
	}

	outputs, err := bactherStrategyPrime.GetOutputs(txs, config, hclog.NewNullLogger())
	require.NoError(t, err)

	require.Equal(t, []cardanowallet.TxOutput{
		{
			Addr:   addr2,
			Amount: 51,
			Tokens: []cardanowallet.TokenAmount{
				cardanowallet.NewTokenAmount(token, 102),
			},
		},
		{
			Addr:   addr1,
			Amount: 102,
			Tokens: []cardanowallet.TokenAmount{
				cardanowallet.NewTokenAmount(token, 205),
			},
		},
		{
			Addr:   addr3,
			Amount: 8,
		},
	}, outputs.Outputs)
	require.Len(t, outputs.Sum, 2)
	require.Equal(t, uint64(307), outputs.Sum[token.String()])
	require.Equal(t, uint64(161), outputs.Sum[cardanowallet.AdaTokenName])
}

func Test_getUTXOs(t *testing.T) {
	feeAddr := "0x002"
	config := &cardano.CardanoChainConfig{
		NoBatchPeriodPercent: 0.1,
		MaxFeeUtxoCount:      4,
		MaxUtxoCount:         50,
	}
	txOutputs := cardano.TxOutputs{
		Outputs: []cardanowallet.TxOutput{
			{}, {}, {},
		},
		Sum: map[string]uint64{
			cardanowallet.AdaTokenName: 2_000_000,
		},
	}

	t.Run("empty fee", func(t *testing.T) {
		allMultisigUtxos := []*indexer.TxInputOutput{
			{
				Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0x1"), Index: 2},
				Output: indexer.TxOutput{Amount: 1_000_000, Slot: 80},
			},
		}

		_, _, err := getUTXOsForAmounts(
			config, feeAddr, allMultisigUtxos, nil, txOutputs.Sum, 0)
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

		multisigUtxos, feeUtxos, err := getUTXOsForAmounts(
			config, feeAddr, allMultisigUtxos, allFeeUtxos, txOutputs.Sum, 0)

		require.NoError(t, err)
		require.Equal(t, expectedUtxos[0:2], multisigUtxos)
		require.Equal(t, expectedUtxos[2:], feeUtxos)
	})
}

func Test_getSkylineUTXOs(t *testing.T) {
	feeAddr := "0x002"
	config := &cardano.CardanoChainConfig{
		NoBatchPeriodPercent: 0.1,
		MaxFeeUtxoCount:      4,
		MaxUtxoCount:         50,
		NativeTokens: []sendtx.TokenExchangeConfig{
			{
				DstChainID: "testChainID",
				TokenName:  "1.31",
			},
		},
	}
	txOutputs := cardano.TxOutputs{
		Outputs: []cardanowallet.TxOutput{
			{
				Amount: 5,
				Tokens: []cardanowallet.TokenAmount{
					{
						Token:  cardanowallet.NewToken("1", "1"),
						Amount: 20,
					},
				},
			},
			{
				Amount: 15,
			},
			{
				Amount: 30,
				Tokens: []cardanowallet.TokenAmount{
					{
						Token:  cardanowallet.NewToken("1", "1"),
						Amount: 40,
					},
				},
			},
			{
				Amount: 10,
			},
		},
		Sum: map[string]uint64{
			cardanowallet.AdaTokenName: 60,
			"1.31":                     60,
		},
	}

	t.Run("pass", func(t *testing.T) {
		expectedUtxos := []*indexer.TxInputOutput{
			{
				Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("01"), Index: 2},
				Output: indexer.TxOutput{
					Amount: 30,
					Slot:   80,
					Tokens: []indexer.TokenAmount{
						{
							PolicyID: "1",
							Name:     "1",
							Amount:   40,
						},
					},
				},
			},

			{
				Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("01"), Index: 3},
				Output: indexer.TxOutput{
					Amount: 40,
					Slot:   1900,
					Tokens: []indexer.TokenAmount{
						{
							PolicyID: "1",
							Name:     "1",
							Amount:   30,
						},
					},
				},
			},
			{
				Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("AA"), Index: 100},
				Output: indexer.TxOutput{
					Amount: 10,
				},
			},
		}
		allMultisigUtxos := expectedUtxos[0:2]
		allFeeUtxos := expectedUtxos[2:]

		multisigUtxos, feeUtxos, err := getUTXOsForAmounts(
			config, feeAddr, allMultisigUtxos, allFeeUtxos, txOutputs.Sum, 0)

		require.NoError(t, err)
		require.Equal(t, expectedUtxos[0:2], multisigUtxos)
		require.Equal(t, expectedUtxos[2:], feeUtxos)
	})
}

func Test_filterOutTokenUtxos(t *testing.T) {
	multisigUtxos := []*indexer.TxInputOutput{
		{
			Input: indexer.TxInput{Index: 0},
			Output: indexer.TxOutput{
				Amount: 30,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "1",
						Name:     "1",
						Amount:   40,
					},
				},
			},
		},
		{
			Input: indexer.TxInput{Index: 1},
			Output: indexer.TxOutput{
				Amount: 40,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "3",
						Name:     "1",
						Amount:   30,
					},
					{
						PolicyID: "1",
						Name:     "2",
						Amount:   30,
					},
				},
			},
		},
		{
			Input: indexer.TxInput{Index: 2},
			Output: indexer.TxOutput{
				Amount: 50,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "1",
						Name:     "1",
						Amount:   51,
					},
					{
						PolicyID: "3",
						Name:     "1",
						Amount:   21,
					},
				},
			},
		},
		{
			Input: indexer.TxInput{Index: 3},
			Output: indexer.TxOutput{
				Amount: 2,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "3",
						Name:     "1",
						Amount:   7,
					},
				},
			},
		},
	}

	t.Run("filter out all the tokens", func(t *testing.T) {
		resTxInputOutput := filterOutUtxosWithUnknownTokens(multisigUtxos)
		require.Equal(t, 0, len(resTxInputOutput))
	})

	t.Run("filter out all the tokens except the one with specified token name", func(t *testing.T) {
		tok, err := cardano.GetNativeTokenFromName("1.31")
		require.NoError(t, err)

		resTxInputOutput := filterOutUtxosWithUnknownTokens(multisigUtxos, tok)
		require.Equal(t, 1, len(resTxInputOutput))
		require.Equal(
			t,
			indexer.TxInput{Index: 0},
			resTxInputOutput[0].Input,
		)
	})

	t.Run("filter out InputOutput with invalid token even if it contains valid token as well", func(t *testing.T) {
		tok, err := cardano.GetNativeTokenFromName("3.31")
		require.NoError(t, err)

		resTxInputOutput := filterOutUtxosWithUnknownTokens(multisigUtxos, tok)
		require.Equal(t, 1, len(resTxInputOutput))
		require.Equal(
			t,
			indexer.TxInput{Index: 3},
			resTxInputOutput[0].Input,
		)
	})

	t.Run("filter out all the tokens except those with specified token names", func(t *testing.T) {
		tok1, err := cardano.GetNativeTokenFromName("3.31")
		require.NoError(t, err)

		tok2, err := cardano.GetNativeTokenFromName("1.31")
		require.NoError(t, err)

		resTxInputOutput := filterOutUtxosWithUnknownTokens(multisigUtxos, tok1, tok2)
		require.Equal(t, 3, len(resTxInputOutput))
		require.Equal(
			t,
			indexer.TxInput{Index: 0},
			resTxInputOutput[0].Input,
		)
		require.Equal(
			t,
			2,
			len(resTxInputOutput[1].Output.Tokens),
		)
	})
}

// if tokens are passed as parameters, two of them are required
func generateSmallUtxoOutputs(value, n uint64, tokens ...cardanowallet.Token) ([]*indexer.TxInputOutput, uint64) {
	utxoOutput := make([]*indexer.TxInputOutput, 0, n)
	returnCurrencySum := uint64(0)

	for i := uint64(0); i < n; i++ {
		tx := &indexer.TxInputOutput{
			Input: indexer.TxInput{
				Hash: indexer.NewHashFromHexString(fmt.Sprintf("0x00%d", i)),
			},
			Output: indexer.TxOutput{
				Amount: value,
			},
		}

		if len(tokens) > 0 {
			token := tokens[int(i)%len(tokens)]

			tx.Output.Tokens = append(tx.Output.Tokens, indexer.TokenAmount{
				PolicyID: token.PolicyID,
				Name:     token.Name,
				Amount:   value,
			})
		}

		utxoOutput = append(utxoOutput, tx)
		returnCurrencySum += value
	}

	return utxoOutput, returnCurrencySum
}
