package batcher

import (
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	"github.com/Ethernal-Tech/bn256"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	secretsHelper "github.com/Ethernal-Tech/cardano-infrastructure/secrets/helper"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestEthChain_GenerateBatchTransaction(t *testing.T) {
	chainID := common.ChainIDStrNexus
	ctx := context.Background()
	batchNonceID := uint64(7834)
	ttlBlockNumberInc := uint64(5)

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

	_, err = eth.CreateAndSaveBatcherEVMPrivateKey(secretsMngr, chainID, true)
	require.NoError(t, err)

	currencyID := uint16(1)
	wrappedCurrencyID := uint16(2)
	tokenID := uint16(3)

	batcherConfig := cardanotx.BatcherEVMChainConfig{
		TTLBlockNumberInc:      ttlBlockNumberInc,
		BlockRoundingThreshold: 6,
		NoBatchPeriodPercent:   0.1,
		Tokens: map[uint16]common.Token{
			currencyID: {
				ChainSpecific: "lovelace",
			},
			wrappedCurrencyID: {
				ChainSpecific:     "cc",
				IsWrappedCurrency: true,
			},
			tokenID: {
				ChainSpecific: "dd",
			},
		},
	}

	chainSpecificJSONRaw, err := batcherConfig.Serialize()
	require.NoError(t, err)

	dbMock := eventTrackerStore.NewTestTrackerStore(t)

	require.NoError(t, dbMock.InsertLastProcessedBlock(uint64(4)))

	t.Run("pass", func(t *testing.T) {
		confirmedTxs := []eth.ConfirmedTransaction{
			{
				SourceChainId: 2,
				Receivers: []eth.BridgeReceiver{
					{
						Amount:             common.DfmToWei(big.NewInt(100)),
						AmountWrapped:      common.DfmToWei(big.NewInt(0)),
						DestinationAddress: "0xff",
						TokenId:            0,
					},
					{
						Amount:             common.DfmToWei(big.NewInt(1000)),
						AmountWrapped:      common.DfmToWei(big.NewInt(0)),
						DestinationAddress: "0xaa",
						TokenId:            0,
					},
					{
						Amount:             common.DfmToWei(big.NewInt(0)),
						AmountWrapped:      common.DfmToWei(big.NewInt(1000)),
						DestinationAddress: "0xaa",
						TokenId:            0,
					},
					{
						Amount:             common.DfmToWei(big.NewInt(0)),
						AmountWrapped:      common.DfmToWei(big.NewInt(1000)),
						DestinationAddress: "0xaa",
						TokenId:            wrappedCurrencyID,
					},
					{
						Amount:             common.DfmToWei(big.NewInt(1000)),
						AmountWrapped:      common.DfmToWei(big.NewInt(1000)),
						DestinationAddress: "0xaa",
						TokenId:            tokenID,
					},
					{
						Amount:             common.DfmToWei(big.NewInt(10)),
						AmountWrapped:      common.DfmToWei(big.NewInt(0)),
						DestinationAddress: "cc",
						TokenId:            0,
					},
				},
			},
		}
		ops, err := NewEVMChainOperations(
			chainSpecificJSONRaw, secretsMngr, dbMock, chainID, hclog.NewNullLogger())
		require.NoError(t, err)

		dt, err := ops.GenerateBatchTransaction(ctx, chainID, confirmedTxs, batchNonceID)
		require.NoError(t, err)

		txs, err := newEVMSmartContractTransaction(&batcherConfig, batchNonceID, uint64(6)+ttlBlockNumberInc, confirmedTxs, big.NewInt(0))
		require.NoError(t, err)

		require.Equal(t, []eth.EVMSmartContractTransactionReceiver{
			{
				Address: common.HexToAddress("0xaa"),
				Amount:  common.DfmToWei(big.NewInt(2000)),
				TokenID: currencyID,
			},
			{
				Address: common.HexToAddress("0xaa"),
				Amount:  common.DfmToWei(big.NewInt(2000)),
				TokenID: wrappedCurrencyID,
			},
			{
				Address: common.HexToAddress("0xaa"),
				Amount:  common.DfmToWei(big.NewInt(1000)),
				TokenID: tokenID,
			},
			{
				Address: common.HexToAddress("0xcc"),
				Amount:  common.DfmToWei(big.NewInt(10)),
				TokenID: currencyID,
			},
			{
				Address: common.HexToAddress("0xff"),
				Amount:  common.DfmToWei(big.NewInt(100)),
				TokenID: currencyID,
			},
		}, txs.Receivers)

		txsBytes, err := txs.Pack()
		require.NoError(t, err)

		hash, err := common.Keccak256(txsBytes)
		require.NoError(t, err)

		require.Equal(t, hex.EncodeToString(hash), dt.TxHash)
	})
}

func TestEthChain_SignBatchTransaction(t *testing.T) {
	hash := "7fc42dc2cecb683c88f5646d6afc6360e088ffebebb8232f2f59ccd30614b4b9"
	secret := new(big.Int).SetUint64(uint64(3824728647346735412))
	expected := "1291681c0d2c6f48e3fdef436a5638995ee90d4ac072279a1ea95519abb69cd10a97fb741dc9f3eeae3c7f68599f307b5eeb19071d4988e0a5f2cb5830ae7a26"

	t.Run("pass", func(t *testing.T) {
		ops := &EVMChainOperations{
			privateKey: bn256.NewPrivateKey(secret),
			logger:     hclog.NewNullLogger(),
		}

		signatures, err := ops.SignBatchTransaction(&core.GeneratedBatchTxData{TxHash: hash})
		require.NoError(t, err)

		require.Equal(t, expected, hex.EncodeToString(signatures.Multisig))
	})
}

func TestEthChain_newEVMSmartContractTransaction(t *testing.T) {
	batchNonceID := uint64(213)
	ttl := uint64(39203902)
	feeAmount := new(big.Int).SetUint64(11)

	currencyID := uint16(1)
	wrappedCurrencyID := uint16(2)
	tokenID := uint16(3)

	batcherConfig := cardanotx.BatcherEVMChainConfig{
		TTLBlockNumberInc:      5,
		BlockRoundingThreshold: 6,
		NoBatchPeriodPercent:   0.1,
		Tokens: map[uint16]common.Token{
			currencyID: {
				ChainSpecific: "lovelace",
			},
			wrappedCurrencyID: {
				ChainSpecific:     "cc",
				IsWrappedCurrency: true,
			},
			tokenID: {
				ChainSpecific: "dd",
			},
		},
	}

	confirmedTxs := []eth.ConfirmedTransaction{
		{
			Receivers: []eth.BridgeReceiver{
				{
					Amount:             common.DfmToWei(big.NewInt(100)),
					AmountWrapped:      common.DfmToWei(big.NewInt(100)),
					DestinationAddress: "0xff",
					TokenId:            wrappedCurrencyID,
				},
				{
					Amount:             common.DfmToWei(big.NewInt(200)),
					AmountWrapped:      common.DfmToWei(big.NewInt(200)),
					DestinationAddress: "0xfa",
					TokenId:            tokenID,
				},
			},
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					Amount:             common.DfmToWei(big.NewInt(10)),
					AmountWrapped:      common.DfmToWei(big.NewInt(0)),
					DestinationAddress: "0xff",
					TokenId:            0,
				},
			},
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					Amount:             common.DfmToWei(big.NewInt(15)),
					AmountWrapped:      common.DfmToWei(big.NewInt(0)),
					DestinationAddress: "0xf0",
					TokenId:            0,
				},
				{
					Amount:             common.DfmToWei(big.NewInt(11)),
					AmountWrapped:      common.DfmToWei(big.NewInt(0)),
					DestinationAddress: "0xff",
					TokenId:            0,
				},
			},
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					Amount:             common.DfmToWei(big.NewInt(15)),
					AmountWrapped:      common.DfmToWei(big.NewInt(0)),
					DestinationAddress: "0xf0",
					TokenId:            0,
				},
				{
					Amount:             common.DfmToWei(feeAmount),
					AmountWrapped:      common.DfmToWei(big.NewInt(0)),
					DestinationAddress: common.EthZeroAddr,
					TokenId:            0,
				},
			},
		},
	}

	result, err := newEVMSmartContractTransaction(&batcherConfig, batchNonceID, ttl, confirmedTxs, big.NewInt(0))
	require.NoError(t, err)

	require.Equal(t, eth.EVMSmartContractTransaction{
		BatchNonceID: batchNonceID,
		TTL:          ttl,
		FeeAmount:    common.DfmToWei(feeAmount),
		Receivers: []eth.EVMSmartContractTransactionReceiver{
			{
				Address: common.HexToAddress("0xf0"),
				Amount:  common.DfmToWei(big.NewInt(30)),
				TokenID: currencyID,
			},
			{
				Address: common.HexToAddress("0xfa"),
				Amount:  common.DfmToWei(big.NewInt(200)),
				TokenID: currencyID,
			},
			{
				Address: common.HexToAddress("0xfa"),
				Amount:  common.DfmToWei(big.NewInt(200)),
				TokenID: tokenID,
			},
			{
				Address: common.HexToAddress("0xff"),
				Amount:  common.DfmToWei(big.NewInt(121)),
				TokenID: currencyID,
			},
			{
				Address: common.HexToAddress("0xff"),
				Amount:  common.DfmToWei(big.NewInt(100)),
				TokenID: wrappedCurrencyID,
			},
		},
	}, *result)
}

func TestEthChain_newEVMSmartContractTransactionRefund(t *testing.T) {
	batchNonceID := uint64(213)
	ttl := uint64(39203902)
	feeAmount := new(big.Int).SetUint64(11)
	minFeeForBridging := new(big.Int).SetUint64(10)

	currencyID := uint16(1)
	wrappedCurrencyID := uint16(2)
	tokenID := uint16(3)

	batcherConfig := cardanotx.BatcherEVMChainConfig{
		TTLBlockNumberInc:      5,
		BlockRoundingThreshold: 6,
		NoBatchPeriodPercent:   0.1,
		Tokens: map[uint16]common.Token{
			currencyID: {
				ChainSpecific: "lovelace",
			},
			wrappedCurrencyID: {
				ChainSpecific:     "cc",
				IsWrappedCurrency: true,
			},
			tokenID: {
				ChainSpecific: "dd",
			},
		},
	}

	confirmedTxs := []eth.ConfirmedTransaction{
		{
			TransactionType: uint8(common.RefundConfirmedTxType),
			Receivers: []eth.BridgeReceiver{
				{
					Amount:             common.DfmToWei(big.NewInt(100)),
					AmountWrapped:      common.DfmToWei(big.NewInt(100)),
					DestinationAddress: "0xff",
					TokenId:            0,
				},
				{
					Amount:             common.DfmToWei(big.NewInt(200)),
					AmountWrapped:      common.DfmToWei(big.NewInt(2)),
					DestinationAddress: "0xfa",
					TokenId:            wrappedCurrencyID,
				},
			},
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					Amount:             common.DfmToWei(big.NewInt(10)),
					AmountWrapped:      common.DfmToWei(big.NewInt(50)),
					DestinationAddress: "0xff",
					TokenId:            tokenID,
				},
			},
		},
		{
			TransactionType: uint8(common.RefundConfirmedTxType),
			Receivers: []eth.BridgeReceiver{
				{
					Amount:             common.DfmToWei(big.NewInt(15)),
					AmountWrapped:      common.DfmToWei(big.NewInt(15)),
					DestinationAddress: "0xf0",
					TokenId:            0,
				},
				{
					Amount:             common.DfmToWei(big.NewInt(11)),
					AmountWrapped:      common.DfmToWei(big.NewInt(11)),
					DestinationAddress: "0xff",
					TokenId:            wrappedCurrencyID,
				},
			},
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					Amount:             common.DfmToWei(big.NewInt(15)),
					AmountWrapped:      common.DfmToWei(big.NewInt(2)),
					DestinationAddress: "0xf0",
					TokenId:            0,
				},
				{
					Amount:             common.DfmToWei(feeAmount),
					AmountWrapped:      common.DfmToWei(big.NewInt(0)),
					DestinationAddress: common.EthZeroAddr,
					TokenId:            1,
				},
			},
		},
	}

	result, err := newEVMSmartContractTransaction(&batcherConfig, batchNonceID, ttl, confirmedTxs, minFeeForBridging)
	require.NoError(t, err)

	require.Equal(t, eth.EVMSmartContractTransaction{
		BatchNonceID: batchNonceID,
		TTL:          ttl,
		FeeAmount:    big.NewInt(0).Add(common.DfmToWei(feeAmount), big.NewInt(40)),
		Receivers: []eth.EVMSmartContractTransactionReceiver{
			{
				Address: common.HexToAddress("0xf0"),
				// 30 - 2 * minFeeForBridging due to refund tx
				Amount: big.NewInt(0).Sub(
					common.DfmToWei(big.NewInt(30)), minFeeForBridging),
				TokenID: currencyID,
			},
			{
				Address: common.HexToAddress("0xf0"),
				Amount:  common.DfmToWei(big.NewInt(17)),
				TokenID: wrappedCurrencyID,
			},
			{
				Address: common.HexToAddress("0xfa"),
				// 200 - 1 * minFeeForBridging due to refund tx
				Amount:  big.NewInt(0).Sub(common.DfmToWei(big.NewInt(200)), minFeeForBridging),
				TokenID: currencyID,
			},
			{
				Address: common.HexToAddress("0xfa"),
				Amount:  common.DfmToWei(big.NewInt(2)),
				TokenID: wrappedCurrencyID,
			},
			{
				Address: common.HexToAddress("0xff"),
				// 121 - 2 * minFeeForBridging due to refund txs
				Amount:  big.NewInt(0).Sub(common.DfmToWei(big.NewInt(121)), big.NewInt(0).Mul(minFeeForBridging, big.NewInt(2))),
				TokenID: currencyID,
			},
			{
				Address: common.HexToAddress("0xff"),
				Amount:  common.DfmToWei(big.NewInt(111)),
				TokenID: wrappedCurrencyID,
			},
			{
				Address: common.HexToAddress("0xff"),
				Amount:  common.DfmToWei(big.NewInt(50)),
				TokenID: tokenID,
			},
		},
	}, *result)
}

func TestEthChain_IsSynchronized(t *testing.T) {
	chainID := "nexus"
	dbMock := eventTrackerStore.NewTestTrackerStore(t)
	bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
	ctx := context.Background()
	scBlock := eth.CardanoBlock{BlockSlot: big.NewInt(15)}
	testErr := errors.New("test error 1")

	bridgeSmartContractMock.On("GetLastObservedBlock", ctx, chainID).Return(eth.CardanoBlock{}, testErr).Once()
	bridgeSmartContractMock.On("GetLastObservedBlock", ctx, chainID).Return(scBlock, nil).Times(6)

	cco := &EVMChainOperations{
		db:     dbMock,
		logger: hclog.NewNullLogger(),
	}

	// sc error
	_, err := cco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.ErrorIs(t, err, testErr)

	for _, i := range []uint64{5, 10, 12, 15, 16, 18} {
		require.NoError(t, dbMock.InsertLastProcessedBlock(i))

		val, err := cco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
		require.NoError(t, err)
		require.Equal(t, i >= scBlock.BlockSlot.Uint64(), val)
	}
}
