package batcher

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	batcherCore "github.com/Ethernal-Tech/apex-bridge/batcher/core"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	solanatx "github.com/Ethernal-Tech/apex-bridge/solana"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	secretsHelper "github.com/Ethernal-Tech/cardano-infrastructure/secrets/helper"
	"github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	solanaTrackerStore "github.com/Ethernal-Tech/solana-infrastructure/tracker/store"
	"github.com/gagliardetto/solana-go"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestSolanaChainOperations(t *testing.T) (string, *SolanaChainOperations, *solanaTrackerStore.MockStorageHandler, func()) {
	t.Helper()

	destChainID := common.ChainIDStrSolana

	testDir, err := os.MkdirTemp("", "bat-chain-ops-tx")
	require.NoError(t, err)

	cleanup := func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}

	secretsMngr, err := secretsHelper.CreateSecretsManager(&secrets.SecretsManagerConfig{
		Path: filepath.Join(testDir, "stp"),
		Type: secrets.Local,
	})
	require.NoError(t, err)

	_, err = solanatx.GenerateAndStoreBatcherSolanaPrivateKey(secretsMngr, destChainID, true)
	require.NoError(t, err)

	const (
		currencyID        uint16 = 1
		wrappedCurrencyID        = 2
	)

	wrappedTokenMint, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)

	batcherConfig := solanatx.SolanaChainConfig{
		TTLNumberInc:          5,
		SlotRoundingThreshold: 6,
		NoBatchPeriodPercent:  0.1,
		Tokens: map[uint16]common.Token{
			currencyID: {
				ChainSpecific: common.WSOLTokenMint,
			},
			wrappedCurrencyID: {
				ChainSpecific:     wrappedTokenMint.PublicKey().String(),
				IsWrappedCurrency: true,
			},
		},
		MinFeeForBridging: big.NewInt(1_000_000),
	}

	chainSpecificJSONRaw, err := batcherConfig.Serialize()
	require.NoError(t, err)

	chainIDConverter := common.NewTestChainIDConverter()
	dbMock := &solanaTrackerStore.MockStorageHandler{}

	sco, err := NewSolanaChainOperations(chainSpecificJSONRaw, chainIDConverter, dbMock, secretsMngr, destChainID, hclog.NewNullLogger())
	require.NoError(t, err)

	return destChainID, sco, dbMock, cleanup
}

func TestSolanaChain_GenerateBatchTransaction(t *testing.T) {
	ctx := context.Background()
	batchNonceID := uint64(7834)
	minAmount := new(big.Int).SetInt64(2_000_000)

	destChainID, sco, dbMock, cleanup := newTestSolanaChainOperations(t)
	defer cleanup()

	receiverWallet, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		solanaHash, err := solana.NewRandomPrivateKey()
		require.NoError(t, err)

		slotNumber := uint64(10)
		roundedSlot, err := getNumberWithRoundingThresholdRoundDown(
			slotNumber, sco.config.SlotRoundingThreshold, sco.config.NoBatchPeriodPercent,
		)
		require.NoError(t, err)

		latestBlockPoint := &solanaTrackerStore.BlockPoint{
			BlockSlot: slotNumber,
		}

		dbMock.On("GetLatestBlockPoint").Return(latestBlockPoint, nil).Once()
		dbMock.On("GetBlockhashBySlot", roundedSlot).Return(solana.Hash(solanaHash.PublicKey()), nil).Once()

		confirmedTransactions := make([]eth.ConfirmedTransaction, 1)
		confirmedTransactions[0] = eth.ConfirmedTransaction{
			Nonce:       1,
			BlockHeight: big.NewInt(1),
			Receivers: []eth.BridgeReceiver{{
				DestinationAddress: receiverWallet.PublicKey().String(),
				Amount:             minAmount,
				AmountWrapped:      big.NewInt(0),
				TokenId:            1,
			}},
			TransactionType: uint8(common.BridgingConfirmedTxType),
		}

		batchTxData, err := sco.GenerateBatchTransaction(ctx, destChainID, confirmedTransactions, batchNonceID)
		require.NoError(t, err)
		require.NotNil(t, batchTxData)
	})

	t.Run("payload matches manual build", func(t *testing.T) {
		dbMock.ExpectedCalls = nil

		solanaHash, err := solana.NewRandomPrivateKey()
		require.NoError(t, err)

		slotNumber := uint64(4)
		roundedSlot, err := getNumberWithRoundingThresholdRoundDown(
			slotNumber, sco.config.SlotRoundingThreshold, sco.config.NoBatchPeriodPercent,
		)
		require.NoError(t, err)

		expectedBlockhash := solana.Hash(solanaHash.PublicKey())

		latestBlockPoint := &solanaTrackerStore.BlockPoint{
			BlockSlot: slotNumber,
		}

		dbMock.On("GetLatestBlockPoint").Return(latestBlockPoint, nil).Once()
		dbMock.On("GetBlockhashBySlot", roundedSlot).Return(expectedBlockhash, nil).Once()

		confirmedTransactions := []eth.ConfirmedTransaction{
			{
				Nonce:       1,
				BlockHeight: big.NewInt(1),
				Receivers: []eth.BridgeReceiver{
					{
						DestinationAddress: receiverWallet.PublicKey().String(),
						Amount:             minAmount,
						AmountWrapped:      big.NewInt(0),
						TokenId:            1,
					},
				},
				TransactionType: uint8(common.BridgingConfirmedTxType),
			},
		}

		gbt, err := sco.GenerateBatchTransaction(ctx, destChainID, confirmedTransactions, batchNonceID)
		require.NoError(t, err)
		require.NotNil(t, gbt)

		receivers, feeAmount, err := sco.newSolanaReceivers(
			sco.config,
			confirmedTransactions,
			sco.config.MinFeeForBridging,
		)
		require.NoError(t, err)

		payload := sendtx.SolanaPayload{
			Blockhash: [32]byte(expectedBlockhash),
			Receivers: receivers,
			BatchID:   batchNonceID,
			FeeAmount: feeAmount.Uint64(),
		}
		expectedRaw, err := payload.Marshal()
		require.NoError(t, err)

		expectedHash := sha256.Sum256(expectedRaw)
		expectedPayloadHash := hex.EncodeToString(expectedHash[:])

		require.Equal(t, expectedPayloadHash, gbt.TxHash)
		require.Equal(t, expectedRaw, gbt.TxRaw)
	})

	t.Run("error when GetLatestBlockPoint fails", func(t *testing.T) {
		dbMock.ExpectedCalls = nil
		dbMock.On("GetLatestBlockPoint").
			Return(&solanaTrackerStore.BlockPoint{BlockSlot: 0}, errors.New("db error")).
			Once()

		amountAboveMinFee := common.LamportToWei(big.NewInt(2))

		confirmedTransactions := []eth.ConfirmedTransaction{
			{
				Receivers: []eth.BridgeReceiver{
					{
						DestinationAddress: receiverWallet.PublicKey().String(),
						Amount:             amountAboveMinFee,
						AmountWrapped:      big.NewInt(0),
						TokenId:            1,
					},
				},
			},
		}

		batchTxData, err := sco.GenerateBatchTransaction(ctx, destChainID, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "db error")
		require.Nil(t, batchTxData)
	})

	t.Run("error when slot is in non-active batch period (rounding threshold)", func(t *testing.T) {
		dbMock.ExpectedCalls = nil

		latestBlockPoint := &solanaTrackerStore.BlockPoint{
			BlockSlot: 6,
		}

		dbMock.On("GetLatestBlockPoint").Return(latestBlockPoint, nil).Once()

		confirmedTransactions := []eth.ConfirmedTransaction{
			{
				Receivers: []eth.BridgeReceiver{
					{
						DestinationAddress: receiverWallet.PublicKey().String(),
						Amount:             minAmount,
						AmountWrapped:      big.NewInt(0),
						TokenId:            1,
					},
				},
			},
		}

		batchTxData, err := sco.GenerateBatchTransaction(ctx, destChainID, confirmedTransactions, batchNonceID)
		require.ErrorIs(t, err, errNonActiveBatchPeriod)
		require.Nil(t, batchTxData)
	})

	t.Run("error when reading slot", func(t *testing.T) {
		dbMock.ExpectedCalls = nil

		confirmedTransactions := []eth.ConfirmedTransaction{
			{
				Receivers: []eth.BridgeReceiver{
					{
						DestinationAddress: receiverWallet.PublicKey().String(),
						Amount:             minAmount,
						AmountWrapped:      big.NewInt(0),
						TokenId:            1,
					},
				},
			},
		}

		expectedErr := errors.New("read slot error")
		latestBlockPoint := &solanaTrackerStore.BlockPoint{
			BlockSlot: 0,
		}
		dbMock.On("GetLatestBlockPoint").Return(latestBlockPoint, expectedErr).Once()

		batchTxData, err := sco.GenerateBatchTransaction(ctx, destChainID, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.Nil(t, batchTxData)
	})

	t.Run("error when getting blockhash", func(t *testing.T) {
		dbMock.ExpectedCalls = nil
		expectedErr := errors.New("blockhash error")

		slotNumber := uint64(10)
		roundedSlot, err := getNumberWithRoundingThresholdRoundDown(
			slotNumber, sco.config.SlotRoundingThreshold, sco.config.NoBatchPeriodPercent,
		)
		require.NoError(t, err)

		latestBlockPoint := &solanaTrackerStore.BlockPoint{
			BlockSlot: slotNumber,
		}
		dbMock.On("GetLatestBlockPoint").Return(latestBlockPoint, nil).Once()
		dbMock.On("GetBlockhashBySlot", roundedSlot).Return(solana.Hash{}, expectedErr).Once()

		confirmedTransactions := []eth.ConfirmedTransaction{
			{
				Receivers: []eth.BridgeReceiver{
					{
						DestinationAddress: receiverWallet.PublicKey().String(),
						Amount:             minAmount,
						AmountWrapped:      big.NewInt(0),
						TokenId:            1,
					},
				},
			},
		}

		batchTxData, err := sco.GenerateBatchTransaction(ctx, destChainID, confirmedTransactions, batchNonceID)
		require.ErrorIs(t, err, expectedErr)
		require.Nil(t, batchTxData)
	})
}

func TestSolanaChain_SignBatchTransaction(t *testing.T) {
	ctx := context.Background()
	batchNonceID := uint64(77)

	destChainID, sco, dbMock, cleanup := newTestSolanaChainOperations(t)
	defer cleanup()

	buildGeneratedBatchData := func(t *testing.T) *batcherCore.GeneratedBatchTxData {
		t.Helper()

		receiverWallet, err := solana.NewRandomPrivateKey()
		require.NoError(t, err)

		slotNumber := uint64(10)
		roundedSlot, err := getNumberWithRoundingThresholdRoundDown(
			slotNumber, sco.config.SlotRoundingThreshold, sco.config.NoBatchPeriodPercent,
		)
		require.NoError(t, err)

		blockHashKey, err := solana.NewRandomPrivateKey()
		require.NoError(t, err)

		dbMock.ExpectedCalls = nil
		latestBlockPoint := &solanaTrackerStore.BlockPoint{
			BlockSlot: slotNumber,
		}
		dbMock.On("GetLatestBlockPoint").Return(latestBlockPoint, nil).Once()
		dbMock.On("GetBlockhashBySlot", roundedSlot).Return(solana.Hash(blockHashKey.PublicKey()), nil).Once()

		confirmed := []eth.ConfirmedTransaction{
			{
				Receivers: []eth.BridgeReceiver{
					{
						DestinationAddress: receiverWallet.PublicKey().String(),
						Amount:             big.NewInt(2_000_000),
						AmountWrapped:      big.NewInt(0),
						TokenId:            1,
					},
				},
			},
		}

		data, err := sco.GenerateBatchTransaction(ctx, destChainID, confirmed, batchNonceID)
		require.NoError(t, err)
		require.NotNil(t, data)

		return data
	}

	t.Run("signs generated batch payload and verifies with public key", func(t *testing.T) {
		data := buildGeneratedBatchData(t)

		signatures, err := sco.SignBatchTransaction(data)
		require.NoError(t, err)
		require.NotNil(t, signatures)
		require.Len(t, signatures.Multisig, solana.SignatureLength)

		signature := solana.SignatureFromBytes(signatures.Multisig)
		require.True(t, signature.Verify(sco.privateKey.PublicKey(), data.TxRaw))
	})

	t.Run("returns error with invalid private key", func(t *testing.T) {
		data := buildGeneratedBatchData(t)

		opsWithInvalidKey := *sco
		opsWithInvalidKey.privateKey = solana.PrivateKey{}

		signatures, err := opsWithInvalidKey.SignBatchTransaction(data)
		require.Error(t, err)
		require.Nil(t, signatures)
	})
}

func TestSolanaChain_IsSynchronized(t *testing.T) {
	t.Run("synchronized when oracle ahead or equal", func(t *testing.T) {
		destChainID, sco, dbMock, cleanup := newTestSolanaChainOperations(t)
		defer cleanup()

		ctx := context.Background()

		bridgeMock := &eth.BridgeSmartContractMock{}
		lastObserved := eth.CardanoBlock{
			BlockSlot: big.NewInt(10),
		}

		bridgeMock.On("GetLastObservedBlock", ctx, destChainID).Return(lastObserved, nil).Once()
		dbMock.On("GetLatestBlockPoint").Return(&solanaTrackerStore.BlockPoint{BlockSlot: uint64(11)}, nil).Once()

		synced, err := sco.IsSynchronized(ctx, bridgeMock, destChainID)
		require.NoError(t, err)
		require.True(t, synced)

		bridgeMock.AssertExpectations(t)
	})

	t.Run("not synchronized when oracle behind", func(t *testing.T) {
		destChainID, sco, dbMock, cleanup := newTestSolanaChainOperations(t)
		defer cleanup()

		ctx := context.Background()

		bridgeMock := &eth.BridgeSmartContractMock{}
		lastObserved := eth.CardanoBlock{
			BlockSlot: big.NewInt(10),
		}

		bridgeMock.On("GetLastObservedBlock", ctx, destChainID).Return(lastObserved, nil).Once()
		dbMock.On("GetLatestBlockPoint").Return(&solanaTrackerStore.BlockPoint{BlockSlot: uint64(9)}, nil).Once()

		synced, err := sco.IsSynchronized(ctx, bridgeMock, destChainID)
		require.NoError(t, err)
		require.False(t, synced)

		bridgeMock.AssertExpectations(t)
	})

	t.Run("error from smart contract is returned", func(t *testing.T) {
		destChainID, sco, _, cleanup := newTestSolanaChainOperations(t)
		defer cleanup()

		ctx := context.Background()

		bridgeMock := &eth.BridgeSmartContractMock{}
		expectedErr := errors.New("get last observed block error")

		bridgeMock.On("GetLastObservedBlock", ctx, destChainID).Return(eth.CardanoBlock{}, expectedErr).Once()

		synced, err := sco.IsSynchronized(ctx, bridgeMock, destChainID)
		require.Error(t, err, expectedErr)
		require.False(t, synced)

		bridgeMock.AssertExpectations(t)
	})

	t.Run("error from db is returned", func(t *testing.T) {
		destChainID, sco, dbMock, cleanup := newTestSolanaChainOperations(t)
		defer cleanup()

		ctx := context.Background()

		bridgeMock := &eth.BridgeSmartContractMock{}
		bridgeMock.On("GetLastObservedBlock", ctx, destChainID).Return(
			eth.CardanoBlock{BlockSlot: big.NewInt(10)}, nil,
		).Once()

		expectedErr := errors.New("db error")
		dbMock.On("GetLatestBlockPoint").Return((*solanaTrackerStore.BlockPoint)(nil), expectedErr).Once()

		synced, err := sco.IsSynchronized(ctx, bridgeMock, destChainID)
		require.ErrorIs(t, err, expectedErr)
		require.False(t, synced)
	})

	t.Run("not synchronized when db returns nil block point", func(t *testing.T) {
		destChainID, sco, dbMock, cleanup := newTestSolanaChainOperations(t)
		defer cleanup()

		ctx := context.Background()

		bridgeMock := &eth.BridgeSmartContractMock{}
		bridgeMock.On("GetLastObservedBlock", ctx, destChainID).Return(
			eth.CardanoBlock{BlockSlot: big.NewInt(10)}, nil,
		).Once()

		dbMock.On("GetLatestBlockPoint").Return((*solanaTrackerStore.BlockPoint)(nil), error(nil)).Once()

		synced, err := sco.IsSynchronized(ctx, bridgeMock, destChainID)
		require.NoError(t, err)
		require.False(t, synced)
	})
}

func TestSolanaChain_Submit(t *testing.T) {
	destChainID, sco, _, cleanup := newTestSolanaChainOperations(t)
	defer cleanup()

	ctx := context.Background()
	_ = destChainID

	bridgeMock := &eth.BridgeSmartContractMock{}
	signedBatch := eth.SignedBatch{}

	t.Run("submit success", func(t *testing.T) {
		bridgeMock.ExpectedCalls = nil
		bridgeMock.On("SubmitSignedBatchSolana", ctx, signedBatch, mock.AnythingOfType("uint64")).Return(nil).Once()

		err := sco.Submit(ctx, bridgeMock, signedBatch)
		require.NoError(t, err)

		bridgeMock.AssertExpectations(t)
	})

	t.Run("submit error is returned", func(t *testing.T) {
		bridgeMock.ExpectedCalls = nil
		expectedErr := errors.New("submit failed")
		bridgeMock.On("SubmitSignedBatchSolana", ctx, signedBatch, mock.AnythingOfType("uint64")).Return(expectedErr).Once()

		err := sco.Submit(ctx, bridgeMock, signedBatch)
		require.ErrorIs(t, err, expectedErr)

		bridgeMock.AssertExpectations(t)
	})
}

func TestSolanaChain_newSolanaReceivers_AggregationAndSorting(t *testing.T) {
	_, sco, _, cleanup := newTestSolanaChainOperations(t)
	defer cleanup()

	receiver1Key, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)
	receiver2Key, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)

	receiver1 := receiver1Key.PublicKey().String()
	receiver2 := receiver2Key.PublicKey().String()

	amount1 := common.LamportToWei(big.NewInt(1_000_005))
	amount2 := common.LamportToWei(big.NewInt(1_000_002))
	wrappedAmount := common.LamportToWei(big.NewInt(3))

	confirmed := []eth.ConfirmedTransaction{
		{
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: receiver2,
					Amount:             amount1,
					AmountWrapped:      big.NewInt(0),
					TokenId:            1,
				},
				{
					DestinationAddress: receiver1,
					Amount:             amount2,
					AmountWrapped:      big.NewInt(0),
					TokenId:            1,
				},
				{
					DestinationAddress: receiver1,
					Amount:             big.NewInt(0),
					AmountWrapped:      wrappedAmount,
					TokenId:            0,
				},
			},
		},
	}

	receivers, feeAmount, err := sco.newSolanaReceivers(
		sco.config,
		confirmed,
		sco.config.MinFeeForBridging,
	)
	require.NoError(t, err)
	require.NotNil(t, receivers)
	require.NotNil(t, feeAmount)
	require.Len(t, receivers, 3)

	for i := 0; i < len(receivers)-1; i++ {
		r1 := receivers[i]
		r2 := receivers[i+1]

		if r1.Address == r2.Address {
			require.LessOrEqual(t, r1.TokenAmount.TokenID, r2.TokenAmount.TokenID)
		} else {
			require.LessOrEqual(t, bytes.Compare(r1.Address[:], r2.Address[:]), 0)
		}
	}

	for _, r := range receivers {
		require.NotZero(t, r.TokenAmount.Amount)
	}
}

func TestSolanaChain_newSolanaReceivers_ErrorPaths(t *testing.T) {
	_, sco, _, cleanup := newTestSolanaChainOperations(t)
	defer cleanup()

	minFee := sco.config.MinFeeForBridging

	t.Run("GetTokenMint error for non-existing token", func(t *testing.T) {
		_, err := sco.config.GetTokenMint(99)
		require.Error(t, err)
		require.ErrorContains(t, err, "token not found in chain config")
	})

	t.Run("GetWrappedTokenID error when no wrapped token configured", func(t *testing.T) {
		configCopy := *sco.config
		for k, v := range configCopy.Tokens {
			v.IsWrappedCurrency = false
			configCopy.Tokens[k] = v
		}

		receiverWallet, err := solana.NewRandomPrivateKey()
		require.NoError(t, err)

		confirmed := []eth.ConfirmedTransaction{
			{
				Receivers: []eth.BridgeReceiver{
					{
						DestinationAddress: receiverWallet.PublicKey().String(),
						Amount:             big.NewInt(0),
						AmountWrapped:      common.LamportToWei(big.NewInt(2)),
						TokenId:            0,
					},
				},
			},
		}

		receivers, feeAmount, err := sco.newSolanaReceivers(&configCopy, confirmed, minFee)
		require.Error(t, err)
		require.Nil(t, receivers)
		require.Nil(t, feeAmount)
	})
}
