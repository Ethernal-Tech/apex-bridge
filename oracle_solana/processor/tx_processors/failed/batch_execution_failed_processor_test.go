package failedtxprocessors

import (
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestBatchExecutionFailedProcessor(t *testing.T) {
	proc := NewSolanaBatchExecutionFailedProcessor(hclog.NewNullLogger())

	t.Run("GetType", func(t *testing.T) {
		require.Equal(t, common.BridgingTxTypeBatchExecution, proc.GetType())
	})

	t.Run("PreValidate empty tx", func(t *testing.T) {
		err := proc.PreValidate(&core.BridgeExpectedSolanaTx{}, nil)
		require.NoError(t, err)
	})

	t.Run("PreValidate with app config", func(t *testing.T) {
		appConfig := &oCore.AppConfig{}
		err := proc.PreValidate(&core.BridgeExpectedSolanaTx{
			ChainID: common.ChainIDStrSolana,
		}, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &oCore.BridgeClaims{}
		err := proc.ValidateAndAddClaim(claims, &core.BridgeExpectedSolanaTx{}, nil)
		require.Error(t, err)
		require.Equal(t, 0, claims.Count())
	})

	t.Run("ValidateAndAddClaim with claims and config", func(t *testing.T) {
		claims := &oCore.BridgeClaims{}
		appConfig := &oCore.AppConfig{
			ChainIDConverter: common.NewTestChainIDConverter(),
		}

		metadata, err := core.MarshalSolMetadata(core.BatchExecutedSolMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
			BatchNonceID:   7,
		})
		require.NoError(t, err)

		err = proc.ValidateAndAddClaim(claims, &core.BridgeExpectedSolanaTx{
			ChainID:  common.ChainIDStrSolana,
			Metadata: metadata,
		}, appConfig)
		require.NoError(t, err)
		require.Len(t, claims.BatchExecutionFailedClaims, 1)
		require.Equal(t, uint64(7), claims.BatchExecutionFailedClaims[0].BatchNonceId)
	})
}
