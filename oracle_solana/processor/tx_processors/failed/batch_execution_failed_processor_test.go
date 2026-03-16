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
		require.NoError(t, err)
		require.Equal(t, 0, claims.Count())
	})

	t.Run("ValidateAndAddClaim with claims and config", func(t *testing.T) {
		claims := &oCore.BridgeClaims{}
		appConfig := &oCore.AppConfig{
			ChainIDConverter: common.NewTestChainIDConverter(),
		}

		err := proc.ValidateAndAddClaim(claims, &core.BridgeExpectedSolanaTx{
			ChainID: common.ChainIDStrSolana,
		}, appConfig)
		require.NoError(t, err)
		require.Equal(t, 0, claims.Count())
	})
}
