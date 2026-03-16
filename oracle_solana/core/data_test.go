package core

import (
	"testing"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

func TestSolanaTx(t *testing.T) {
	sig := solana.Signature{1, 2, 3, 4, 5}
	iaSig := solana.Signature{10, 20, 30}

	tx := &SolanaTx{
		OriginChainID:   "solana",
		Priority:        1,
		SubmitTryCount:  2,
		BatchTryCount:   3,
		RefundTryCount:  1,
		SlotNumber:      100,
		TxSignature:     sig,
		InnerActionHash: iaSig,
	}

	t.Run("GetChainID", func(t *testing.T) {
		require.Equal(t, "solana", tx.GetChainID())
	})

	t.Run("GetPriority", func(t *testing.T) {
		require.Equal(t, uint8(1), tx.GetPriority())
	})

	t.Run("GetSubmitTryCount", func(t *testing.T) {
		require.Equal(t, uint32(2), tx.GetSubmitTryCount())
	})

	t.Run("GetTxHash", func(t *testing.T) {
		require.Equal(t, sig[:], tx.GetTxHash())
	})

	t.Run("IncrementBatchTryCount", func(t *testing.T) {
		origCount := tx.BatchTryCount
		tx.IncrementBatchTryCount()
		require.Equal(t, origCount+1, tx.BatchTryCount)
	})

	t.Run("IncrementRefundTryCount", func(t *testing.T) {
		origCount := tx.RefundTryCount
		tx.IncrementRefundTryCount()
		require.Equal(t, origCount+1, tx.RefundTryCount)
	})

	t.Run("ResetSubmitTryCount", func(t *testing.T) {
		tx.SubmitTryCount = 5
		tx.ResetSubmitTryCount()
		require.Equal(t, uint32(0), tx.SubmitTryCount)
	})

	t.Run("SetLastTimeTried", func(t *testing.T) {
		now := time.Now().UTC()
		tx.SetLastTimeTried(now)
		require.Equal(t, now, tx.LastTimeTried)
	})

	t.Run("UnprocessedDBKey", func(t *testing.T) {
		key := tx.UnprocessedDBKey()
		require.NotEmpty(t, key)
		require.Equal(t, tx.Priority, key[0])
	})

	t.Run("ToSolanaTxKey", func(t *testing.T) {
		key := tx.ToSolanaTxKey()
		require.NotEmpty(t, key)
		require.True(t, len(key) > len(sig))
	})

	t.Run("ToExpectedSolanaTxKey", func(t *testing.T) {
		key := tx.ToExpectedSolanaTxKey()
		require.NotEmpty(t, key)
	})

	t.Run("ToProcessed valid", func(t *testing.T) {
		processed := tx.ToProcessed(false)
		require.NotNil(t, processed)
		require.False(t, processed.GetIsInvalid())
	})

	t.Run("ToProcessed invalid", func(t *testing.T) {
		processed := tx.ToProcessed(true)
		require.NotNil(t, processed)
		require.True(t, processed.GetIsInvalid())
	})

	t.Run("ToProcessedSolanaTx", func(t *testing.T) {
		processed := tx.ToProcessedSolanaTx(false)
		require.NotNil(t, processed)
		require.Equal(t, tx.OriginChainID, processed.OriginChainID)
		require.Equal(t, tx.SlotNumber, processed.SlotNumber)
		require.Equal(t, tx.TxSignature, processed.TxSignature)
		require.Equal(t, tx.Priority, processed.Priority)
		require.Equal(t, tx.InnerActionHash, processed.InnerActionHash)
		require.False(t, processed.IsInvalid)
	})
}

func TestProcessedSolanaTx(t *testing.T) {
	sig := solana.Signature{5, 6, 7}
	iaSig := solana.Signature{15, 16, 17}

	ptx := &ProcessedSolanaTx{
		SlotNumber:      200,
		OriginChainID:   "solana",
		TxSignature:     sig,
		Priority:        1,
		IsInvalid:       false,
		InnerActionHash: iaSig,
	}

	t.Run("GetChainID", func(t *testing.T) {
		require.Equal(t, "solana", ptx.GetChainID())
	})

	t.Run("GetTxHash", func(t *testing.T) {
		require.Equal(t, sig[:], ptx.GetTxHash())
	})

	t.Run("HasInnerActionTxHash true", func(t *testing.T) {
		require.True(t, ptx.HasInnerActionTxHash())
	})

	t.Run("HasInnerActionTxHash false", func(t *testing.T) {
		ptxNoIA := &ProcessedSolanaTx{}
		require.False(t, ptxNoIA.HasInnerActionTxHash())
	})

	t.Run("GetInnerActionTxHash", func(t *testing.T) {
		require.Equal(t, iaSig[:], ptx.GetInnerActionTxHash())
	})

	t.Run("UnprocessedDBKey", func(t *testing.T) {
		key := ptx.UnprocessedDBKey()
		require.NotEmpty(t, key)
		require.Equal(t, ptx.Priority, key[0])
	})

	t.Run("GetIsInvalid false", func(t *testing.T) {
		require.False(t, ptx.GetIsInvalid())
	})

	t.Run("GetIsInvalid true", func(t *testing.T) {
		invalidPtx := &ProcessedSolanaTx{IsInvalid: true}
		require.True(t, invalidPtx.GetIsInvalid())
	})

	t.Run("ToSolanaTxKey", func(t *testing.T) {
		key := ptx.ToSolanaTxKey()
		require.NotEmpty(t, key)
	})
}

func TestBridgeExpectedSolanaTx(t *testing.T) {
	sig := solana.Signature{8, 9, 10}

	etx := &BridgeExpectedSolanaTx{
		ChainID:     "solana",
		TxSignature: sig,
		Metadata:    []byte("test_metadata"),
		TTL:         500,
		Priority:    0,
		IsProcessed: false,
		IsInvalid:   false,
	}

	t.Run("DBKey", func(t *testing.T) {
		key := etx.DBKey()
		require.NotEmpty(t, key)
	})

	t.Run("GetChainID", func(t *testing.T) {
		require.Equal(t, "solana", etx.GetChainID())
	})

	t.Run("GetTxHash", func(t *testing.T) {
		require.Equal(t, sig[:], etx.GetTxHash())
	})

	t.Run("GetPriority", func(t *testing.T) {
		require.Equal(t, uint8(0), etx.GetPriority())
	})

	t.Run("GetIsInvalid", func(t *testing.T) {
		require.False(t, etx.GetIsInvalid())
	})

	t.Run("GetIsProcessed", func(t *testing.T) {
		require.False(t, etx.GetIsProcessed())
	})

	t.Run("SetProcessed", func(t *testing.T) {
		etx.SetProcessed()
		require.True(t, etx.GetIsProcessed())
	})

	t.Run("SetInvalid", func(t *testing.T) {
		etx.SetInvalid()
		require.True(t, etx.GetIsInvalid())
	})

	t.Run("ToSolanaTxKey", func(t *testing.T) {
		key := etx.ToSolanaTxKey()
		require.NotEmpty(t, key)
	})

	t.Run("ToExpectedTxKey", func(t *testing.T) {
		key := etx.ToExpectedTxKey()
		require.NotEmpty(t, key)
		require.Equal(t, etx.Priority, key[0])
	})
}

func TestBridgeClaimsSlotInfo(t *testing.T) {
	bi := &BridgeClaimsSlotInfo{
		ChainID: "solana",
		Number:  100,
	}

	t.Run("EqualWithUnprocessed matching", func(t *testing.T) {
		tx := &SolanaTx{OriginChainID: "solana", SlotNumber: 100}
		require.True(t, bi.EqualWithUnprocessed(tx))
	})

	t.Run("EqualWithUnprocessed different chain", func(t *testing.T) {
		tx := &SolanaTx{OriginChainID: "other", SlotNumber: 100}
		require.False(t, bi.EqualWithUnprocessed(tx))
	})

	t.Run("EqualWithUnprocessed different slot", func(t *testing.T) {
		tx := &SolanaTx{OriginChainID: "solana", SlotNumber: 200}
		require.False(t, bi.EqualWithUnprocessed(tx))
	})

	t.Run("EqualWithProcessed matching", func(t *testing.T) {
		ptx := &ProcessedSolanaTx{OriginChainID: "solana", SlotNumber: 100}
		require.True(t, bi.EqualWithProcessed(ptx))
	})

	t.Run("EqualWithProcessed different", func(t *testing.T) {
		ptx := &ProcessedSolanaTx{OriginChainID: "other", SlotNumber: 100}
		require.False(t, bi.EqualWithProcessed(ptx))
	})

	t.Run("EqualWithExpected matching", func(t *testing.T) {
		etx := &BridgeExpectedSolanaTx{ChainID: "solana"}
		require.True(t, bi.EqualWithExpected(etx, 100))
	})

	t.Run("EqualWithExpected different chain", func(t *testing.T) {
		etx := &BridgeExpectedSolanaTx{ChainID: "other"}
		require.False(t, bi.EqualWithExpected(etx, 100))
	})

	t.Run("EqualWithExpected different slot", func(t *testing.T) {
		etx := &BridgeExpectedSolanaTx{ChainID: "solana"}
		require.False(t, bi.EqualWithExpected(etx, 200))
	})
}

func TestToSolanaTxKey(t *testing.T) {
	sig := solana.Signature{1, 2, 3}
	key := ToSolanaTxKey("solana", sig)

	require.NotEmpty(t, key)
	require.Equal(t, append([]byte("solana"), sig[:]...), key)
}

func TestToUnprocessedSolanaTxKey(t *testing.T) {
	sig := solana.Signature{1, 2, 3}
	key := toUnprocessedSolanaTxKey(1, 100, sig)

	require.NotEmpty(t, key)
	require.Equal(t, uint8(1), key[0])
	require.Len(t, key, 9+len(sig))
}
