package bridge

import (
	"context"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestBridgeSubmitter(t *testing.T) {
	t.Run("submit claims", func(t *testing.T) {
		bridgeSC := eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("SubmitClaims").Return(nil)

		bridgeSubmitter := NewBridgeSubmitter(context.Background(), &bridgeSC, hclog.NewNullLogger())
		require.NotNil(t, bridgeSubmitter)

		err := bridgeSubmitter.SubmitClaims(&core.BridgeClaims{
			ContractClaims: core.ContractClaims{
				BridgingRequestClaims: []core.BridgingRequestClaim{
					{
						ObservedTransactionHash: common.NewHashFromHexString("0x11"),
						SourceChainId:           common.ToNumChainID(common.ChainIDStrVector),
						DestinationChainId:      common.ToNumChainID(common.ChainIDStrPrime),
						Receivers:               []core.BridgingRequestReceiver{},
					},
				},
				BatchExecutedClaims: []core.BatchExecutedClaim{
					{
						ObservedTransactionHash: common.NewHashFromHexString("0x11"),
						BatchNonceId:            1,
					},
				},
			},
		}, nil)

		require.NoError(t, err)
	})

	t.Run("submit confirmed blocks", func(t *testing.T) {
		bridgeSC := eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("SubmitLastObservedBlocks").Return(nil)

		bridgeSubmitter := NewBridgeSubmitter(context.Background(), &bridgeSC, hclog.NewNullLogger())
		require.NotNil(t, bridgeSubmitter)

		err := bridgeSubmitter.SubmitConfirmedBlocks(common.ChainIDStrPrime, 0, 1)

		require.NoError(t, err)
	})
}
