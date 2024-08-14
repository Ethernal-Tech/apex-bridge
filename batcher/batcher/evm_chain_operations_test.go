package batcher

import (
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"testing"

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

	chainSpecificJSONRaw, err := (cardanotx.BatcherEVMChainConfig{
		TTLBlockNumberInc:      ttlBlockNumberInc,
		BlockRoundingThreshold: 6,
		NoBatchPeriodPercent:   0.1,
	}).Serialize()
	require.NoError(t, err)

	dbMock := eventTrackerStore.NewTestTrackerStore(t)

	require.NoError(t, dbMock.InsertLastProcessedBlock(uint64(4)))

	t.Run("pass", func(t *testing.T) {
		confirmedTxs := []eth.ConfirmedTransaction{
			{
				SourceChainId: 2,
				Receivers: []eth.BridgeReceiver{
					{
						Amount:             new(big.Int).SetUint64(100),
						DestinationAddress: "0xff",
					},
				},
			},
		}
		ops, err := NewEVMChainOperations(
			chainSpecificJSONRaw, secretsMngr, dbMock, chainID, hclog.NewNullLogger())
		require.NoError(t, err)

		dt, err := ops.GenerateBatchTransaction(ctx, nil, chainID, confirmedTxs, batchNonceID)
		require.NoError(t, err)

		txs := newEVMSmartContractTransaction(batchNonceID, uint64(6)+ttlBlockNumberInc, confirmedTxs)

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

		bytes, _, err := ops.SignBatchTransaction(hash)
		require.NoError(t, err)

		require.Equal(t, expected, hex.EncodeToString(bytes))
	})
}

func TestEthChain_newEVMSmartContractTransaction(t *testing.T) {
	batchNonceID := uint64(213)
	ttl := uint64(39203902)
	feeAmount := new(big.Int).SetUint64(11)

	confirmedTxs := []eth.ConfirmedTransaction{
		{
			Receivers: []eth.BridgeReceiver{
				{
					Amount:             new(big.Int).SetUint64(100),
					DestinationAddress: "0xff",
				},
				{
					Amount:             new(big.Int).SetUint64(200),
					DestinationAddress: "0xfa",
				},
			},
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					Amount:             new(big.Int).SetUint64(10),
					DestinationAddress: "0xff",
				},
			},
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					Amount:             new(big.Int).SetUint64(15),
					DestinationAddress: "0xf0",
				},
				{
					Amount:             new(big.Int).SetUint64(11),
					DestinationAddress: "0xff",
				},
			},
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					Amount:             new(big.Int).SetUint64(15),
					DestinationAddress: "0xf0",
				},
				{
					Amount:             feeAmount,
					DestinationAddress: common.ETH_ZERO_ADDR,
				},
			},
		},
	}

	result := newEVMSmartContractTransaction(batchNonceID, ttl, confirmedTxs)
	require.Equal(t, eth.EVMSmartContractTransaction{
		BatchNonceID: batchNonceID,
		TTL:          ttl,
		FeeAmount:    common.DfmToWei(feeAmount),
		Receivers: []eth.EVMSmartContractTransactionReceiver{
			{
				Address: common.HexToAddress("0xf0"),
				Amount:  common.DfmToWei(new(big.Int).SetUint64(30)),
			},
			{
				Address: common.HexToAddress("0xfa"),
				Amount:  common.DfmToWei(new(big.Int).SetUint64(200)),
			},
			{
				Address: common.HexToAddress("0xff"),
				Amount:  common.DfmToWei(new(big.Int).SetUint64(121)),
			},
		},
	}, result)
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
