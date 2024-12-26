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
	"strings"
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

func splitTokenAmount(name string, isNameEncoded bool) (string, string, error) {
	parts := strings.Split(name, ".")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid full token name: %s", name)
	}

	if !isNameEncoded {
		name = parts[1]
	} else {
		decodedName, err := hex.DecodeString(parts[1])
		if err != nil {
			return "", "", fmt.Errorf("invalid full token name: %s", name)
		}

		name = string(decodedName)
	}

	return parts[0], name, nil
}

func TestSkylineCardanoChainOperations_IsSynchronized(t *testing.T) {
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

	scco := &SkylineCardanoChainOperations{
		db:     dbMock,
		logger: hclog.NewNullLogger(),
	}

	// sc error
	_, err := scco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.ErrorIs(t, err, testErr1)

	// database error
	_, err = scco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.ErrorIs(t, err, testErr2)

	// not in sync
	val, err := scco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.NoError(t, err)
	require.False(t, val)

	// in sync
	val, err = scco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.NoError(t, err)
	require.True(t, val)

	// in sync again
	val, err = scco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.NoError(t, err)
	require.True(t, val)
}

func TestGenerateSkylineBatchTransaction(t *testing.T) {
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

	strategy := &CardanoChainOperationSkylineStrategy{}

	scco, err := NewSkylineCardanoChainOperations(configRaw, dbMock, secretsMngr, "prime", strategy, hclog.NewNullLogger())
	require.NoError(t, err)

	scco.txProvider = txProviderMock
	scco.config.SlotRoundingThreshold = 100

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

		_, err := scco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
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

		_, err := scco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
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

		_, err := scco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
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

		_, err := scco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
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

		_, err := scco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
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

		_, err := scco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
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

		result, err := scco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.NoError(t, err)
		require.NotNil(t, result.TxRaw)
		require.NotEqual(t, "", result.TxHash)
	})

	t.Run("Test SignBatchTransaction", func(t *testing.T) {
		witnessMultiSig, witnessMultiSigFee, err := scco.SignBatchTransaction(
			"26a9d1a894c7e3719a79342d0fc788989e5d55f076581327c54bcc0c7693905a")
		require.NoError(t, err)
		require.NotNil(t, witnessMultiSig)
		require.NotNil(t, witnessMultiSigFee)
	})
}

func Test_getNeededSkylineUtxos(t *testing.T) {
	configRaw := json.RawMessage([]byte(`{
			"socketPath": "./socket",
			"testnetMagic": 42,
			"minUtxoAmount": 1000
			}`))
	dbMock := &indexer.DatabaseMock{}

	cardanoConfig, err := cardano.NewCardanoChainConfig(configRaw)
	require.NoError(t, err)

	strategy := &CardanoChainOperationSkylineStrategy{}

	scco := &CardanoChainOperations{
		db:       dbMock,
		logger:   hclog.NewNullLogger(),
		strategy: strategy,
		config:   cardanoConfig,
	}

	primeCardanoWrappedTokenName := "29f8873beb52e126f207a2dfd50f7cff556806b5b4cba9834a7b26a8.526f75746533"
	polID, tName, _ := splitTokenAmount(primeCardanoWrappedTokenName, true)

	var minUtxoAmount uint64 = 5

	inputs := []*indexer.TxInputOutput{
		{
			Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("0x1"), Index: 100},
			Output: indexer.TxOutput{
				Amount: 20,
			},
		},
		{
			Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("0x1"), Index: 0},
			Output: indexer.TxOutput{
				Amount: minUtxoAmount,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: polID,
						Name:     tName,
						Amount:   20,
					},
				},
			},
		},
		{
			Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("0x2"), Index: 7},
			Output: indexer.TxOutput{
				Amount: 15,
			},
		},
		{
			Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("0x4"), Index: 5},
			Output: indexer.TxOutput{
				Amount: minUtxoAmount,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: polID,
						Name:     tName,
						Amount:   40,
					},
				},
			},
		},
		{
			Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("0x4"), Index: 6},
			Output: indexer.TxOutput{
				Amount: 10,
			},
		},
	}

	t.Run("pass", func(t *testing.T) {
		desiredAmounts := map[string]uint64{
			fmt.Sprintf("%s", cardanowallet.AdaTokenName): 35,
		}

		result, err := scco.strategy.GetNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 0, 2, 1)

		require.NoError(t, err)
		require.Equal(t, []*indexer.TxInputOutput{inputs[0], inputs[2]}, result)

		desiredAmounts = map[string]uint64{
			fmt.Sprintf("%s", cardanowallet.AdaTokenName): minUtxoAmount,
			fmt.Sprintf("%s", tName):                      45,
		}

		result, err = scco.strategy.GetNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 5, 30, 1)

		require.NoError(t, err)
		require.Equal(t, inputs[:len(inputs)-1], result)
	})

	t.Run("pass with change", func(t *testing.T) {
		minUtxoAmount = 4
		desiredAmounts := map[string]uint64{
			fmt.Sprintf("%s", cardanowallet.AdaTokenName): minUtxoAmount,
			fmt.Sprintf("%s", tName):                      40,
		}
		result, err := scco.strategy.GetNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 5, 30, 1)

		require.NoError(t, err)
		require.Equal(t, inputs[:len(inputs)-1], result)
	})

	t.Run("pass with at least", func(t *testing.T) {
		minUtxoAmount = 4
		desiredAmounts := map[string]uint64{
			fmt.Sprintf("%s", cardanowallet.AdaTokenName): minUtxoAmount,
			fmt.Sprintf("%s", tName):                      20,
		}
		result, err := scco.strategy.GetNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 5, 30, 3)

		require.NoError(t, err)
		require.Equal(t, inputs[:3], result)

		desiredAmounts = map[string]uint64{
			fmt.Sprintf("%s", cardanowallet.AdaTokenName): 12,
		}
		result, err = scco.strategy.GetNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 5, 30, 3)

		require.NoError(t, err)
		require.Equal(t, inputs[:3], result)
	})

	t.Run("not enough sum", func(t *testing.T) {
		minUtxoAmount = 5
		desiredAmounts := map[string]uint64{
			fmt.Sprintf("%s", cardanowallet.AdaTokenName): 160,
		}
		_, err := scco.strategy.GetNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 5, 30, 1)
		require.ErrorContains(t, err, "couldn't select UTXOs for sum")

		desiredAmounts = map[string]uint64{
			fmt.Sprintf("%s", cardanowallet.AdaTokenName): minUtxoAmount,
			fmt.Sprintf("%s", tName):                      250,
		}
		_, err = scco.strategy.GetNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 5, 30, 1)
		require.ErrorContains(t, err, "couldn't select UTXOs for sum")
	})
}

func Test_getSkylineOutputs(t *testing.T) {
	configRaw := json.RawMessage([]byte(`{
			"socketPath": "./socket",
			"testnetMagic": 42,
			"minUtxoAmount": 1000
			}`))
	cardanoConfig, _ := cardano.NewCardanoChainConfig(configRaw)

	scco := &CardanoChainOperations{
		strategy: &CardanoChainOperationSkylineStrategy{},
		config:   cardanoConfig,
	}
	scco.config.NetworkID = cardanowallet.MainNetNetwork

	cardanoPrimeWrappedTokenName := "72f3d1e6c885e4d0bdcf5250513778dbaa851c0b4bfe3ed4e1bcceb0.4b6173685f546f6b656e"
	primeCardanoWrappedTokenName := "29f8873beb52e126f207a2dfd50f7cff556806b5b4cba9834a7b26a8.526f75746533"

	ccCardanoConfigExchange := []cardano.CardanoConfigTokenExchange{
		{
			Chain:        common.ChainIDStrPrime,
			SrcTokenName: cardanowallet.AdaTokenName,
			DstTokenName: primeCardanoWrappedTokenName,
		},
		{
			Chain:        common.ChainIDStrPrime,
			SrcTokenName: cardanoPrimeWrappedTokenName,
			DstTokenName: cardanowallet.AdaTokenName,
		},
	}

	ccPrimeTokenExchange := []cardano.CardanoConfigTokenExchange{
		{
			Chain:        common.ChainIDStrCardano,
			SrcTokenName: cardanowallet.AdaTokenName,
			DstTokenName: cardanoPrimeWrappedTokenName,
		},
		{
			Chain:        common.ChainIDStrCardano,
			SrcTokenName: primeCardanoWrappedTokenName,
			DstTokenName: cardanowallet.AdaTokenName,
		},
	}

	t.Run("from cardano to prime", func(t *testing.T) {
		scco.config.Destinations = ccCardanoConfigExchange
		txs := []eth.ConfirmedTransaction{
			{
				Receivers: []eth.BridgeReceiver{
					{
						DestinationAddress: "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
						Amount:             big.NewInt(100),
						AmountWrapped:      big.NewInt(10),
					},
					{
						DestinationAddress: "addr128phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtupnz75xxcrtw79hu",
						Amount:             big.NewInt(200),
						AmountWrapped:      big.NewInt(20),
					},
					{
						DestinationAddress: "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8",
						Amount:             big.NewInt(400),
						AmountWrapped:      big.NewInt(0),
					},
				},
				SourceChainId: 4,
			},
			{
				Receivers: []eth.BridgeReceiver{
					{
						DestinationAddress: "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8",
						Amount:             big.NewInt(900),
						AmountWrapped:      big.NewInt(80),
					},
				},
				SourceChainId: 4,
			},
		}

		polID, tName, _ := splitTokenAmount(primeCardanoWrappedTokenName, true)
		res, err := scco.strategy.GetOutputs(txs, scco.config, common.ChainIDStrPrime, hclog.NewNullLogger())
		assert.NoError(t, err)

		assert.Equal(t, map[string]uint64{
			fmt.Sprintf("%s", cardanowallet.AdaTokenName):   1600,
			fmt.Sprintf("%s", primeCardanoWrappedTokenName): 110,
		}, res.Sum)

		assert.Equal(t, []cardanowallet.TxOutput{
			{
				Addr:   "addr128phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtupnz75xxcrtw79hu",
				Amount: 200,
				Tokens: []cardanowallet.TokenAmount{
					{
						PolicyID: polID,
						Name:     tName,
						Amount:   20,
					},
				},
			},
			{
				Addr:   "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
				Amount: 100,
				Tokens: []cardanowallet.TokenAmount{
					{
						PolicyID: polID,
						Name:     tName,
						Amount:   10,
					},
				},
			},
			{
				Addr:   "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8",
				Amount: 1300,
				Tokens: []cardanowallet.TokenAmount{
					{
						PolicyID: polID,
						Name:     tName,
						Amount:   80,
					},
				},
			},
		}, res.Outputs)
	})

	t.Run("from prime to cardano", func(t *testing.T) {
		scco.config.Destinations = ccPrimeTokenExchange
		txs := []eth.ConfirmedTransaction{
			{
				Receivers: []eth.BridgeReceiver{
					{
						DestinationAddress: "addr1qx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer3n0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgse35a3x",
						Amount:             big.NewInt(3000),
						AmountWrapped:      big.NewInt(200),
					},
					{
						DestinationAddress: "addr1z8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gten0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgs9yc0hh",
						Amount:             big.NewInt(0),
						AmountWrapped:      big.NewInt(0),
					},
					{
						// this one will be skipped
						DestinationAddress: "stake178phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcccycj5",
						Amount:             big.NewInt(3000),
						AmountWrapped:      big.NewInt(1300),
					},
				},
				SourceChainId: 1,
			},
			{
				Receivers: []eth.BridgeReceiver{
					{
						DestinationAddress: "addr1w8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcyjy7wx",
						Amount:             big.NewInt(170),
						AmountWrapped:      big.NewInt(50),
					},
				},
				SourceChainId: 1,
			},
		}

		polID, tName, _ := splitTokenAmount(cardanoPrimeWrappedTokenName, true)
		res, err := scco.strategy.GetOutputs(txs, scco.config, common.ChainIDStrCardano, hclog.NewNullLogger())
		assert.NoError(t, err)
		assert.Equal(t, map[string]uint64{
			fmt.Sprintf("%s", cardanowallet.AdaTokenName):   3170,
			fmt.Sprintf("%s", cardanoPrimeWrappedTokenName): 250,
		}, res.Sum)

		assert.Equal(t, []cardanowallet.TxOutput{
			{
				Addr:   "addr1qx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer3n0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgse35a3x",
				Amount: 3000,
				Tokens: []cardanowallet.TokenAmount{
					{
						PolicyID: polID,
						Name:     tName,
						Amount:   200,
					},
				},
			},
			{
				Addr:   "addr1w8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcyjy7wx",
				Amount: 170,
				Tokens: []cardanowallet.TokenAmount{
					{
						PolicyID: polID,
						Name:     tName,
						Amount:   50,
					},
				},
			},
		}, res.Outputs)
	})
}
