package batcher

import (
	"context"
	"encoding/hex"
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
		witnessMultiSig, witnessMultiSigFee, err := cco.SignBatchTransaction(
			"26a9d1a894c7e3719a79342d0fc788989e5d55f076581327c54bcc0c7693905a")
		require.NoError(t, err)
		require.NotNil(t, witnessMultiSig)
		require.NotNil(t, witnessMultiSigFee)
	})
}

func Test_getNeededUtxos(t *testing.T) {
	const minUtxoAmount = 5

	reactorStrategy := &CardanoChainOperationReactorStrategy{}
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
		result, err := reactorStrategy.getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 4, 0, 1)

		require.NoError(t, err)
		require.Equal(t, []*indexer.TxInputOutput{inputs[0], inputs[2], inputs[3], inputs[4]}, result)

		desiredAmounts[cardanowallet.AdaTokenName] = 245
		result, err = reactorStrategy.getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 2, 0, 1)

		require.NoError(t, err)
		require.Equal(t, []*indexer.TxInputOutput{inputs[0], inputs[2]}, result)
	})

	t.Run("pass with change", func(t *testing.T) {
		desiredAmounts[cardanowallet.AdaTokenName] = 706
		result, err := reactorStrategy.getNeededUtxos(inputs, desiredAmounts, 4, 3, 0, 1)

		require.NoError(t, err)
		require.Equal(t, inputs[3:6], result)
	})

	t.Run("pass with at least", func(t *testing.T) {
		desiredAmounts[cardanowallet.AdaTokenName] = 5
		result, err := reactorStrategy.getNeededUtxos(inputs, desiredAmounts, 4, 30, 0, 3)

		require.NoError(t, err)
		require.Equal(t, inputs[:3], result)
	})

	t.Run("not enough sum", func(t *testing.T) {
		desiredAmounts[cardanowallet.AdaTokenName] = 1550
		_, err := reactorStrategy.getNeededUtxos(inputs, desiredAmounts, 5, 30, 0, 1)
		require.ErrorContains(t, err, "not enough funds for the transaction")
	})
}

func Test_getNeededSkylineUtxos(t *testing.T) {
	strategy := &CardanoChainOperationSkylineStrategy{}
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

	var outputsWithTokens uint64 = 4

	t.Run("pass", func(t *testing.T) {
		const minUtxoAmount = 5

		desiredAmounts := map[string]uint64{
			cardanowallet.AdaTokenName: 590,
		}

		result, err := strategy.getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 4, outputsWithTokens, 1)

		require.NoError(t, err)
		require.Equal(t, []*indexer.TxInputOutput{inputs[0], inputs[2], inputs[3], inputs[4]}, result)

		desiredAmounts = map[string]uint64{
			cardanowallet.AdaTokenName: outputsWithTokens * minUtxoAmount,
			"1.31":                     100,
		}

		result, err = strategy.getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 1, outputsWithTokens, 1)

		require.NoError(t, err)
		require.Equal(t, []*indexer.TxInputOutput{inputs[1]}, result)
	})

	t.Run("pass with change", func(t *testing.T) {
		const minUtxoAmount = 4

		desiredAmounts := map[string]uint64{
			cardanowallet.AdaTokenName: outputsWithTokens * minUtxoAmount,
			"1.31":                     350,
		}
		result, err := strategy.getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 2, outputsWithTokens, 1)

		require.NoError(t, err)
		require.Equal(t, []*indexer.TxInputOutput{inputs[1], inputs[6]}, result)
	})

	t.Run("pass with at least", func(t *testing.T) {
		const minUtxoAmount = 4

		desiredAmounts := map[string]uint64{
			cardanowallet.AdaTokenName: outputsWithTokens * minUtxoAmount,
			"1.31":                     20,
		}
		result, err := strategy.getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 30, outputsWithTokens, 3)

		require.NoError(t, err)
		require.Equal(t, inputs[:3], result)

		desiredAmounts = map[string]uint64{
			cardanowallet.AdaTokenName: 12,
		}
		result, err = strategy.getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 30, outputsWithTokens, 3)

		require.NoError(t, err)
		require.Equal(t, inputs[:3], result)
	})

	t.Run("not enough sum", func(t *testing.T) {
		const minUtxoAmount = 5

		desiredAmounts := map[string]uint64{
			cardanowallet.AdaTokenName: 1600,
		}
		_, err := strategy.getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 30, outputsWithTokens, 1)
		require.ErrorContains(t, err, "not enough funds for the transaction")

		desiredAmounts = map[string]uint64{
			cardanowallet.AdaTokenName: outputsWithTokens * minUtxoAmount,
			"1.31":                     2500,
		}
		_, err = strategy.getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 30, outputsWithTokens, 1)
		require.ErrorContains(t, err, "not enough funds for the transaction")
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

	res, _, _ := cco.strategy.GetOutputs(txs, cco.config, hclog.NewNullLogger())

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

	outputs, tokenOutputs, err := bactherStrategyPrime.GetOutputs(txs, config, hclog.NewNullLogger())
	require.NoError(t, err)

	require.Equal(t, uint64(2), tokenOutputs)
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
	dbMock := &indexer.DatabaseMock{}
	multisigAddr := "0x001"
	feeAddr := "0x002"
	testErr := errors.New("test err")
	ops := &CardanoChainOperations{
		db: dbMock,
		config: &cardano.CardanoChainConfig{
			NoBatchPeriodPercent: 0.1,
		},
		strategy: &CardanoChainOperationReactorStrategy{},
		logger:   hclog.NewNullLogger(),
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

		_, _, err := ops.strategy.GetUTXOs(multisigAddr, feeAddr, txOutputs, 0, ops.config, ops.db, ops.logger)
		require.Error(t, err)
	})

	t.Run("GetAllTxOutputs fee error", func(t *testing.T) {
		dbMock.On("GetAllTxOutputs", multisigAddr, true).Return([]*indexer.TxInputOutput{}, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).Return(([]*indexer.TxInputOutput)(nil), testErr).Once()

		_, _, err := ops.strategy.GetUTXOs(multisigAddr, feeAddr, txOutputs, 0, ops.config, ops.db, ops.logger)
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

		multisigUtxos, feeUtxos, err := ops.strategy.GetUTXOs(multisigAddr, feeAddr, txOutputs, 0, ops.config, ops.db, ops.logger)

		require.NoError(t, err)
		require.Equal(t, expectedUtxos[0:2], multisigUtxos)
		require.Equal(t, expectedUtxos[2:], feeUtxos)
	})
}

func Test_getSkylineUTXOs(t *testing.T) {
	dbMock := &indexer.DatabaseMock{}
	multisigAddr := "0x001"
	feeAddr := "0x002"
	testErr := errors.New("test err")

	ops := &CardanoChainOperations{
		db: dbMock,
		config: &cardano.CardanoChainConfig{
			NoBatchPeriodPercent: 0.1,
		},
		strategy: &CardanoChainOperationSkylineStrategy{},
		logger:   hclog.NewNullLogger(),
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

	t.Run("GetAllTxOutputs multisig error", func(t *testing.T) {
		dbMock.On("GetAllTxOutputs", multisigAddr, true).Return(([]*indexer.TxInputOutput)(nil), testErr).Once()

		_, _, err := ops.strategy.GetUTXOs(multisigAddr, feeAddr, txOutputs, 2, ops.config, ops.db, ops.logger)
		require.Error(t, err)
	})

	t.Run("GetAllTxOutputs fee error", func(t *testing.T) {
		dbMock.On("GetAllTxOutputs", multisigAddr, true).Return([]*indexer.TxInputOutput{}, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).Return(([]*indexer.TxInputOutput)(nil), testErr).Once()

		_, _, err := ops.strategy.GetUTXOs(multisigAddr, feeAddr, txOutputs, 2, ops.config, ops.db, ops.logger)
		require.Error(t, err)
	})

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

		dbMock.On("GetAllTxOutputs", multisigAddr, true).Return(allMultisigUtxos, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).Return(allFeeUtxos, error(nil)).Once()

		multisigUtxos, feeUtxos, err := ops.strategy.GetUTXOs(multisigAddr, feeAddr, txOutputs, 2, ops.config, ops.db, ops.logger)

		require.NoError(t, err)
		require.Equal(t, expectedUtxos[0:2], multisigUtxos)
		require.Equal(t, expectedUtxos[2:], feeUtxos)
	})
}
