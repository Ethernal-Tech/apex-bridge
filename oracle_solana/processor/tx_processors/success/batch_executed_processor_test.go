package successtxprocessors

import (
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	"github.com/gagliardetto/solana-go"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestBatchExecutedProcessor(t *testing.T) {
	proc := NewSolanaBatchExecutedProcessor(hclog.NewNullLogger())

	t.Run("GetType", func(t *testing.T) {
		require.Equal(t, common.BridgingTxTypeBatchExecution, proc.GetType())
	})

	t.Run("PreValidate empty tx", func(t *testing.T) {
		err := proc.PreValidate(&core.SolanaTx{}, nil)
		require.NoError(t, err)
	})

	t.Run("PreValidate with app config", func(t *testing.T) {
		err := proc.PreValidate(&core.SolanaTx{
			OriginChainID: common.ChainIDStrSolana,
		}, &oCore.AppConfig{})
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &oCore.BridgeClaims{}

		err := proc.ValidateAndAddClaim(claims, &core.SolanaTx{}, nil)
		require.Error(t, err)
		require.Equal(t, 0, claims.Count())
	})

	t.Run("ValidateAndAddClaim irrelevant metadata", func(t *testing.T) {
		irrelevantMetadata, err := core.MarshalSolMetadata(core.BaseSolMetadata{
			BridgingTxType: common.BridgingTxTypeBridgingRequest,
		})
		require.NoError(t, err)
		require.NotNil(t, irrelevantMetadata)

		claims := &oCore.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, &core.SolanaTx{
			Metadata: irrelevantMetadata,
		}, &oCore.AppConfig{
			ChainIDConverter: common.NewTestChainIDConverter(),
		})
		require.Error(t, err)
		require.Equal(t, 0, claims.Count())
	})

	t.Run("ValidateAndAddClaim valid but metadata not full", func(t *testing.T) {
		relevantButNotFullMetadata, err := core.MarshalSolMetadata(core.BaseSolMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
		})
		require.NoError(t, err)
		require.NotNil(t, relevantButNotFullMetadata)

		claims := &oCore.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, &core.SolanaTx{
			OriginChainID: common.ChainIDStrSolana,
			Metadata:      relevantButNotFullMetadata,
		}, &oCore.AppConfig{
			ChainIDConverter: common.NewTestChainIDConverter(),
		})
		require.NoError(t, err)
		require.Equal(t, 1, claims.Count())
		require.Len(t, claims.BatchExecutedClaims, 1)
		require.Equal(t, [32]byte{}, [32]byte(claims.BatchExecutedClaims[0].ObservedTransactionHash))
		require.Equal(t, uint64(0), claims.BatchExecutedClaims[0].BatchNonceId)
	})

	t.Run("ValidateAndAddClaim valid full metadata", func(t *testing.T) {
		batchNonceID := uint64(7)
		relevantFullMetadata, err := core.MarshalSolMetadata(core.BatchExecutedSolMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
			BatchNonceID:   batchNonceID,
		})
		require.NoError(t, err)
		require.NotNil(t, relevantFullMetadata)

		claims := &oCore.BridgeClaims{}
		txSig := solana.Signature{1, 20}

		err = proc.ValidateAndAddClaim(claims, &core.SolanaTx{
			TxSignature:   txSig,
			OriginChainID: common.ChainIDStrSolana,
			Metadata:      relevantFullMetadata,
		}, &oCore.AppConfig{
			ChainIDConverter: common.NewTestChainIDConverter(),
		})
		require.NoError(t, err)
		require.Equal(t, 1, claims.Count())
		require.Len(t, claims.BatchExecutedClaims, 1)
		require.Equal(t, txSig[:], claims.BatchExecutedClaims[0].ObservedTransactionHash[:])
		require.Equal(t, batchNonceID, claims.BatchExecutedClaims[0].BatchNonceId)
		require.Equal(
			t,
			common.NewTestChainIDConverter().ToChainIDNum(common.ChainIDStrSolana),
			claims.BatchExecutedClaims[0].ChainId,
		)
	})

	t.Run("validate helper passes", func(t *testing.T) {
		err := proc.validate(&core.SolanaTx{}, &core.BatchExecutedSolMetadata{}, &oCore.AppConfig{})
		require.NoError(t, err)
	})
}
