package batcher

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
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
	"github.com/mr-tron/base58"
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

	instructionConfig, err := sendtx.NewInstructionConfig()
	require.NoError(t, err)

	wrappedTokenMint, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)

	bridgingFeeWallet, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)

	batcherConfig := solanatx.SolanaChainConfig{
		TTLSlotNumberInc:      5,
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
		BridgingFeeAddress: bridgingFeeWallet.PublicKey().String(),
		MinFeeForBridging:  big.NewInt(1_000_000),
		InstructionConfig:  *instructionConfig,
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
	minAmount := new(big.Int).SetUint64(1_000)

	destChainID, sco, dbMock, cleanup := newTestSolanaChainOperations(t)
	defer cleanup()

	receiverWallet, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		solanaHash, err := solana.NewRandomPrivateKey()
		require.NoError(t, err)

		slotNumber := uint64(4)

		dbMock.On("ReadSlot").Return(slotNumber, nil).Once()
		dbMock.On("GetBlockhashBySlot", slotNumber+2).Return(solana.Hash(solanaHash.PublicKey()), nil).Once()

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
		fmt.Printf("\ntxHash: %s\nrawTx: %s\n", batchTxData.TxHash, hex.EncodeToString(batchTxData.TxRaw[:]))
		require.NoError(t, err)
		require.NotNil(t, batchTxData)
	})

	t.Run("error when building solana transaction (invalid token)", func(t *testing.T) {
		confirmedTransactions := []eth.ConfirmedTransaction{
			{
				Receivers: []eth.BridgeReceiver{
					{
						DestinationAddress: receiverWallet.PublicKey().String(),
						Amount:             minAmount,
						AmountWrapped:      big.NewInt(0),
						TokenId:            99,
					},
				},
			},
		}

		batchTxData, err := sco.GenerateBatchTransaction(ctx, destChainID, confirmedTransactions, batchNonceID)
		require.Error(t, err)
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
						TokenId:            0,
					},
				},
			},
		}

		expectedErr := errors.New("read slot error")
		dbMock.On("ReadSlot").Return(uint64(0), expectedErr).Once()

		batchTxData, err := sco.GenerateBatchTransaction(ctx, destChainID, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.Nil(t, batchTxData)
	})

	t.Run("error when getting blockhash", func(t *testing.T) {
		dbMock.ExpectedCalls = nil

		slotNumber := uint64(10)
		dbMock.On("ReadSlot").Return(slotNumber, nil).Once()
		dbMock.On("GetBlockhashBySlot", slotNumber+2).Return(solana.Hash{}, errors.New("blockhash error")).Once()

		confirmedTransactions := []eth.ConfirmedTransaction{
			{
				Receivers: []eth.BridgeReceiver{
					{
						DestinationAddress: receiverWallet.PublicKey().String(),
						Amount:             minAmount,
						AmountWrapped:      big.NewInt(0),
						TokenId:            0,
					},
				},
			},
		}

		batchTxData, err := sco.GenerateBatchTransaction(ctx, destChainID, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.Nil(t, batchTxData)
	})
}

func TestSolanaChain_SignBatchTransaction(t *testing.T) {
	destChainID, sco, _, cleanup := newTestSolanaChainOperations(t)
	defer cleanup()

	_ = destChainID

	t.Run("success", func(t *testing.T) {
		hashBytes := []byte("hash-to-sign")
		txHash := base58.Encode(hashBytes)

		data := &batcherCore.GeneratedBatchTxData{
			TxHash: txHash,
		}

		signatures, err := sco.SignBatchTransaction(data)
		require.NoError(t, err)
		require.NotNil(t, signatures)
		require.NotEmpty(t, signatures.Multisig)
	})

	t.Run("error on invalid hash", func(t *testing.T) {
		// invalid base58 string should cause Decode to fail before signing
		data := &batcherCore.GeneratedBatchTxData{
			TxHash: "###-invalid-base58-hash-###",
		}

		signatures, err := sco.SignBatchTransaction(data)
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
		require.Error(t, err)
		require.False(t, synced)

		bridgeMock.AssertExpectations(t)
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
		bridgeMock.On("SubmitSignedBatch", ctx, signedBatch, mock.AnythingOfType("uint64")).Return(nil).Once()

		err := sco.Submit(ctx, bridgeMock, signedBatch)
		require.NoError(t, err)

		bridgeMock.AssertExpectations(t)
	})

	t.Run("submit error is returned", func(t *testing.T) {
		bridgeMock.ExpectedCalls = nil
		expectedErr := errors.New("submit failed")
		bridgeMock.On("SubmitSignedBatch", ctx, signedBatch, mock.AnythingOfType("uint64")).Return(expectedErr).Once()

		err := sco.Submit(ctx, bridgeMock, signedBatch)
		require.ErrorIs(t, err, expectedErr)

		bridgeMock.AssertExpectations(t)
	})
}

func TestSolanaChain_newSolanaSmartContractTransaction_AggregationAndSorting(t *testing.T) {
	destChainID, sco, _, cleanup := newTestSolanaChainOperations(t)
	defer cleanup()

	receiver1 := "AddrA"
	receiver2 := "AddrB"

	amount1 := big.NewInt(5_000_000) // > minFeeForBridging
	amount2 := big.NewInt(2_000_000) // > minFeeForBridging
	wrappedAmount := big.NewInt(3_000_000)

	confirmed := []eth.ConfirmedTransaction{
		{
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: receiver2,
					Amount:             amount1,
					AmountWrapped:      big.NewInt(0),
					TokenId:            0,
				},
				{
					DestinationAddress: receiver1,
					Amount:             amount2,
					AmountWrapped:      big.NewInt(0),
					TokenId:            0,
				},
				{
					DestinationAddress: receiver1,
					Amount:             big.NewInt(0),
					AmountWrapped:      wrappedAmount,
					TokenId:            1,
				},
			},
		},
	}

	dto, err := sco.newSolanaSmartContractTransaction(
		sco.config,
		destChainID,
		42,
		confirmed,
		sco.config.MinFeeForBridging,
	)
	require.NoError(t, err)
	require.NotNil(t, dto)

	require.Equal(t, uint8(0), dto.SrcChainID)
	require.Equal(t, sco.chainIDConverter.StrToInt[destChainID], dto.DstChainID)
	require.Equal(t, sco.config.BridgingFeeAddress, dto.SenderAddr)
	require.Equal(t, uint64(42), dto.BatchID)

	require.Len(t, dto.Receivers, 3)

	for i := 0; i < len(dto.Receivers)-1; i++ {
		r1 := dto.Receivers[i]
		r2 := dto.Receivers[i+1]

		if r1.Address == r2.Address {
			require.LessOrEqual(t, r1.TokenAmount.TokenMint, r2.TokenAmount.TokenMint)
		} else {
			require.LessOrEqual(t, r1.Address, r2.Address)
		}
	}

	for _, r := range dto.Receivers {
		require.True(t, r.TokenAmount.Amount.Cmp(big.NewInt(0)) == 1)
	}
}

func TestSolanaChain_newSolanaSmartContractTransaction_ErrorPaths(t *testing.T) {
	destChainID, sco, _, cleanup := newTestSolanaChainOperations(t)
	defer cleanup()

	minFee := sco.config.MinFeeForBridging

	t.Run("GetTokenMint error for non-existing token", func(t *testing.T) {
		confirmed := []eth.ConfirmedTransaction{
			{
				Receivers: []eth.BridgeReceiver{
					{
						DestinationAddress: "SomeAddr",
						Amount:             new(big.Int).Set(minFee),
						AmountWrapped:      big.NewInt(0),
						TokenId:            99,
					},
				},
			},
		}

		dto, err := sco.newSolanaSmartContractTransaction(
			sco.config,
			destChainID,
			1,
			confirmed,
			minFee,
		)
		require.Error(t, err)
		require.Nil(t, dto)
	})

	t.Run("GetWrappedTokenID error when no wrapped token configured", func(t *testing.T) {
		configCopy := *sco.config
		for k, v := range configCopy.Tokens {
			v.IsWrappedCurrency = false
			configCopy.Tokens[k] = v
		}

		confirmed := []eth.ConfirmedTransaction{
			{
				Receivers: []eth.BridgeReceiver{
					{
						DestinationAddress: "SomeAddr",
						Amount:             big.NewInt(0),
						AmountWrapped:      new(big.Int).Set(minFee),
						TokenId:            0,
					},
				},
			},
		}

		dto, err := sco.newSolanaSmartContractTransaction(
			&configCopy,
			destChainID,
			1,
			confirmed,
			minFee,
		)
		require.Error(t, err)
		require.Nil(t, dto)
	})
}
