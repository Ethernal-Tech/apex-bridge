package failedtxprocessors

import (
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestBatchExecutionFailedProcessor(t *testing.T) {
	proc := NewEthBatchExecutionFailedProcessor(hclog.NewNullLogger())
	appConfig := &oCore.AppConfig{
		ChainIDConverter: common.NewChainIDConverterForTest(),
	}

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &oCore.BridgeClaims{}

		err := proc.ValidateAndAddClaim(claims, &core.BridgeExpectedEthTx{}, nil)
		require.Error(t, err)
	})

	t.Run("ValidateAndAddClaim irrelevant metadata", func(t *testing.T) {
		irrelevantMetadata, err := core.MarshalEthMetadata(core.BaseEthMetadata{
			BridgingTxType: common.BridgingTxTypeBridgingRequest,
		})
		require.NoError(t, err)
		require.NotNil(t, irrelevantMetadata)

		claims := &oCore.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, &core.BridgeExpectedEthTx{
			Metadata: irrelevantMetadata,
		}, nil)
		require.Error(t, err)
	})

	t.Run("ValidateAndAddClaim valid but metadata not full", func(t *testing.T) {
		relevantButNotFullMetadata, err := core.MarshalEthMetadata(core.BaseEthMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
		})
		require.NoError(t, err)
		require.NotNil(t, relevantButNotFullMetadata)

		claims := &oCore.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, &core.BridgeExpectedEthTx{
			Metadata: relevantButNotFullMetadata,
		}, appConfig)
		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		require.Len(t, claims.BatchExecutionFailedClaims, 1)
		require.Equal(t, [32]byte{}, claims.BatchExecutionFailedClaims[0].ObservedTransactionHash)
	})

	t.Run("ValidateAndAddClaim valid full metadata", func(t *testing.T) {
		batchNonceID := uint64(1)
		relevantFullMetadata, err := core.MarshalEthMetadata(core.BatchExecutedEthMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
			BatchNonceID:   batchNonceID,
		})
		require.NoError(t, err)
		require.NotNil(t, relevantFullMetadata)

		claims := &oCore.BridgeClaims{}
		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))

		err = proc.ValidateAndAddClaim(claims, &core.BridgeExpectedEthTx{
			Metadata: relevantFullMetadata,
			Hash:     txHash,
		}, appConfig)
		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		require.Len(t, claims.BatchExecutionFailedClaims, 1)
		require.Equal(t, txHash, claims.BatchExecutionFailedClaims[0].ObservedTransactionHash)
		require.Equal(t, batchNonceID, claims.BatchExecutionFailedClaims[0].BatchNonceId)
	})
}
