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

	minUtxoAmount := new(big.Int).SetUint64(minUtxoAmount)

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
			"testnetMagic": 42
			}`))
	dbMock := &indexer.DatabaseMock{}
	txProviderMock := &cardano.TxProviderTestMock{
		ReturnDefaultParameters: true,
	}

	cco, err := NewCardanoChainOperations(configRaw, dbMock, secretsMngr, "prime", hclog.NewNullLogger())
	require.NoError(t, err)

	cco.txProvider = txProviderMock
	cco.config.SlotRoundingThreshold = 100

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
		witnessMultiSig, witnessMultiSigFee, err := cco.SignBatchTransaction(
			"26a9d1a894c7e3719a79342d0fc788989e5d55f076581327c54bcc0c7693905a")
		require.NoError(t, err)
		require.NotNil(t, witnessMultiSig)
		require.NotNil(t, witnessMultiSigFee)
	})
}

func Test_getSlotNumberWithRoundingThreshold(t *testing.T) {
	_, err := getSlotNumberWithRoundingThreshold(66, 60, 0.125)
	assert.ErrorIs(t, err, errNonActiveBatchPeriod)

	_, err = getSlotNumberWithRoundingThreshold(12, 60, 0.2)
	assert.ErrorIs(t, err, errNonActiveBatchPeriod)

	_, err = getSlotNumberWithRoundingThreshold(115, 60, 0.125)
	assert.ErrorIs(t, err, errNonActiveBatchPeriod)

	_, err = getSlotNumberWithRoundingThreshold(224, 80, 0.2)
	assert.ErrorIs(t, err, errNonActiveBatchPeriod)

	_, err = getSlotNumberWithRoundingThreshold(336, 80, 0.2)
	assert.ErrorIs(t, err, errNonActiveBatchPeriod)

	_, err = getSlotNumberWithRoundingThreshold(0, 60, 0.125)
	assert.ErrorContains(t, err, "slot number is zero")

	val, err := getSlotNumberWithRoundingThreshold(75, 60, 0.125)
	assert.NoError(t, err)
	assert.Equal(t, uint64(120), val)

	val, err = getSlotNumberWithRoundingThreshold(105, 60, 0.125)
	assert.NoError(t, err)
	assert.Equal(t, uint64(120), val)

	val, err = getSlotNumberWithRoundingThreshold(40, 60, 0.125)
	assert.NoError(t, err)
	assert.Equal(t, uint64(60), val)

	val, err = getSlotNumberWithRoundingThreshold(270, 80, 0.125)
	assert.NoError(t, err)
	assert.Equal(t, uint64(320), val)

	val, err = getSlotNumberWithRoundingThreshold(223, 80, 0.2)
	assert.NoError(t, err)
	assert.Equal(t, uint64(240), val)

	val, err = getSlotNumberWithRoundingThreshold(337, 80, 0.2)
	assert.NoError(t, err)
	assert.Equal(t, uint64(400), val)
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
		require.ErrorContains(t, err, "couldn't select UTXOs for sum")
	})
}

func Test_getOutputs(t *testing.T) {
	txs := []eth.ConfirmedTransaction{
		{
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "0x1",
					Amount:             big.NewInt(100),
				},
				{
					DestinationAddress: "0x2",
					Amount:             big.NewInt(200),
				},
				{
					DestinationAddress: "0x3",
					Amount:             big.NewInt(400),
				},
			},
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "0x4",
					Amount:             big.NewInt(50),
				},
				{
					DestinationAddress: "0x3",
					Amount:             big.NewInt(900),
				},
				{
					DestinationAddress: "0x11",
					Amount:             big.NewInt(0),
				},
			},
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "0x5",
					Amount:             big.NewInt(3000),
				},
			},
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "0x1",
					Amount:             big.NewInt(2000),
				},
				{
					DestinationAddress: "0x4",
					Amount:             big.NewInt(170),
				},
				{
					DestinationAddress: "0x3",
					Amount:             big.NewInt(10),
				},
			},
		},
	}

	res := getOutputs(txs)

	assert.Equal(t, uint64(6830), res.Sum)
	assert.Equal(t, []cardanowallet.TxOutput{
		{
			Addr:   "0x1",
			Amount: 2100,
		},
		{
			Addr:   "0x2",
			Amount: 200,
		},
		{
			Addr:   "0x3",
			Amount: 1310,
		},
		{
			Addr:   "0x4",
			Amount: 220,
		},
		{
			Addr:   "0x5",
			Amount: 3000,
		},
	}, res.Outputs)
}

func Test_getUTXOs(t *testing.T) {
	dbMock := &indexer.DatabaseMock{}
	multisigAddr := "0x001"
	feeAddr := "0x002"
	testErr := errors.New("test err")
	ops := &CardanoChainOperations{
		db:     dbMock,
		logger: hclog.NewNullLogger(),
	}
	txOutputs := cardano.TxOutputs{
		Outputs: []cardanowallet.TxOutput{
			{}, {}, {},
		},
		Sum: 2_000_000,
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
