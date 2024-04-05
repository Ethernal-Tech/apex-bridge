package failed_tx_processors

import (
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/stretchr/testify/require"
)

func TestBatchExecutionFailedProcessor(t *testing.T) {

	proc := NewBatchExecutionFailedProcessor()

	t.Run("IsTxRelevant", func(t *testing.T) {
		relevant, err := proc.IsTxRelevant(&core.BridgeExpectedCardanoTx{}, nil)
		require.Error(t, err)
		require.False(t, relevant)

		irrelevantMetadata, err := common.MarshalMetadata(common.MetadataEncodingTypeCbor, common.BaseMetadata{
			BridgingTxType: common.BridgingTxTypeBridgingRequest,
		})

		require.NoError(t, err)
		require.NotNil(t, irrelevantMetadata)

		relevant, err = proc.IsTxRelevant(&core.BridgeExpectedCardanoTx{
			Metadata: irrelevantMetadata,
		}, nil)
		require.NoError(t, err)
		require.False(t, relevant)

		relevantMetadata, err := common.MarshalMetadata(common.MetadataEncodingTypeCbor, common.BaseMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
		})
		require.NoError(t, err)
		require.NotNil(t, relevantMetadata)

		relevant, err = proc.IsTxRelevant(&core.BridgeExpectedCardanoTx{
			Metadata: relevantMetadata,
		}, nil)
		require.NoError(t, err)
		require.True(t, relevant)
	})

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &core.BridgeClaims{}

		err := proc.ValidateAndAddClaim(claims, &core.BridgeExpectedCardanoTx{}, nil)
		require.Error(t, err)
	})

	t.Run("ValidateAndAddClaim irrelevant metadata", func(t *testing.T) {
		irrelevantMetadata, err := common.MarshalMetadata(common.MetadataEncodingTypeCbor, common.BaseMetadata{
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
		relevantButNotFullMetadata, err := common.MarshalMetadata(common.MetadataEncodingTypeCbor, common.BaseMetadata{
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
		batchNonceId := uint64(1)
		relevantFullMetadata, err := common.MarshalMetadata(common.MetadataEncodingTypeCbor, common.BatchExecutedMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
			BatchNonceId:   batchNonceId,
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
		require.Equal(t, new(big.Int).SetUint64(batchNonceId), claims.BatchExecutionFailedClaims[0].BatchNonceID)
	})
}
