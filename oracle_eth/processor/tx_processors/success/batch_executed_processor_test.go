package successtxprocessors

import (
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	"github.com/Ethernal-Tech/ethgo"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestBatchExecutedProcessor(t *testing.T) {
	proc := NewEthBatchExecutedProcessor(hclog.NewNullLogger())

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &oCore.BridgeClaims{}

		err := proc.ValidateAndAddClaim(claims, &core.EthTx{}, nil)
		require.Error(t, err)
	})

	t.Run("ValidateAndAddClaim irrelevant metadata", func(t *testing.T) {
		irrelevantMetadata, err := core.MarshalEthMetadata(core.BaseEthMetadata{
			BridgingTxType: common.BridgingTxTypeBridgingRequest,
		})
		require.NoError(t, err)
		require.NotNil(t, irrelevantMetadata)

		claims := &oCore.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, &core.EthTx{
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
		err = proc.ValidateAndAddClaim(claims, &core.EthTx{
			OriginChainID: common.ChainIDStrNexus,
			Metadata:      relevantButNotFullMetadata,
		}, &oCore.AppConfig{
			ChainIDConverter: common.NewChainIDConverterForTest(),
		})
		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		require.Len(t, claims.BatchExecutedClaims, 1)
		require.Equal(t, [32]byte{}, claims.BatchExecutedClaims[0].ObservedTransactionHash)
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
		txHash := ethgo.Hash{1, 20}

		err = proc.ValidateAndAddClaim(claims, &core.EthTx{
			Hash:          txHash,
			OriginChainID: common.ChainIDStrNexus,
			Metadata:      relevantFullMetadata,
		}, &oCore.AppConfig{
			ChainIDConverter: common.NewChainIDConverterForTest(),
		})
		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		require.Len(t, claims.BatchExecutedClaims, 1)
		require.Equal(t, txHash[:], claims.BatchExecutedClaims[0].ObservedTransactionHash[:])
		require.Equal(t, batchNonceID, claims.BatchExecutedClaims[0].BatchNonceId)
	})
}
