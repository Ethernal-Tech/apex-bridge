package failedtxprocessors

import (
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestBatchExecutionFailedProcessor(t *testing.T) {
	proc := NewBatchExecutionFailedProcessor(hclog.NewNullLogger())

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &core.BridgeClaims{}

		err := proc.ValidateAndAddClaim(claims, &core.BridgeExpectedCardanoTx{}, nil)
		require.Error(t, err)
	})

	t.Run("ValidateAndAddClaim irrelevant metadata", func(t *testing.T) {
		irrelevantMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BaseMetadata{
			BridgingTxType: common.BridgingTxTypeBridgingRequest,
		})
		require.NoError(t, err)
		require.NotNil(t, irrelevantMetadata)

		claims := &core.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, &core.BridgeExpectedCardanoTx{
			Metadata: irrelevantMetadata,
		}, nil)
		require.Error(t, err)
	})

	t.Run("ValidateAndAddClaim valid but metadata not full", func(t *testing.T) {
		relevantButNotFullMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BaseMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
		})
		require.NoError(t, err)
		require.NotNil(t, relevantButNotFullMetadata)

		claims := &core.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, &core.BridgeExpectedCardanoTx{
			Metadata: relevantButNotFullMetadata,
		}, nil)
		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		require.Len(t, claims.BatchExecutionFailedClaims, 1)
		require.Equal(t, "", claims.BatchExecutionFailedClaims[0].ObservedTransactionHash)
	})

	t.Run("ValidateAndAddClaim valid full metadata", func(t *testing.T) {
		batchNonceID := uint64(1)
		relevantFullMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BatchExecutedMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
			BatchNonceID:   batchNonceID,
		})
		require.NoError(t, err)
		require.NotNil(t, relevantFullMetadata)

		claims := &core.BridgeClaims{}

		const txHash = "test_hash"
		err = proc.ValidateAndAddClaim(claims, &core.BridgeExpectedCardanoTx{
			Metadata: relevantFullMetadata,
			Hash:     txHash,
		}, nil)
		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		require.Len(t, claims.BatchExecutionFailedClaims, 1)
		require.Equal(t, txHash, claims.BatchExecutionFailedClaims[0].ObservedTransactionHash)
		require.Equal(t, new(big.Int).SetUint64(batchNonceID), claims.BatchExecutionFailedClaims[0].BatchNonceID)
	})
}
