package bridge

import (
	"context"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestBridgeSubmitter(t *testing.T) {
	t.Run("submit claims", func(t *testing.T) {
		bridgeSC := eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("SubmitClaims").Return(nil, nil)

		bridgeSubmitter := NewBridgeSubmitter(context.Background(), &bridgeSC, hclog.NewNullLogger())
		require.NotNil(t, bridgeSubmitter)

		_, err := bridgeSubmitter.SubmitClaims(&cCore.BridgeClaims{
			ContractClaims: cCore.ContractClaims{
				BridgingRequestClaims: []cCore.BridgingRequestClaim{
					{
						ObservedTransactionHash: common.NewHashFromHexString("0x11"),
						SourceChainId:           common.ToNumChainID(common.ChainIDStrVector),
						DestinationChainId:      common.ToNumChainID(common.ChainIDStrPrime),
						Receivers:               []cCore.BridgingRequestReceiver{},
					},
				},
				BatchExecutedClaims: []cCore.BatchExecutedClaim{
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

		err := bridgeSubmitter.SubmitBlocks(common.ChainIDStrPrime, []eth.CardanoBlock{})

		require.NoError(t, err)
	})
}
