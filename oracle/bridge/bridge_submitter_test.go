package bridge

import (
	"context"
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
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
						ObservedTransactionHash: "test",
						SourceChainID:           "vector",
						DestinationChainID:      "prime",
						OutputUTXO:              core.UTXO{},
						Receivers:               []core.BridgingRequestReceiver{},
					},
				},
				BatchExecutedClaims: []core.BatchExecutedClaim{
					{
						ObservedTransactionHash: "test",
						BatchNonceID:            big.NewInt(1),
						OutputUTXOs:             core.UTXOs{},
					},
				},
			},
		})

		require.NoError(t, err)
	})

	t.Run("submit confirmed blocks", func(t *testing.T) {
		bridgeSC := eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("SubmitLastObservedBlocks").Return(nil)

		bridgeSubmitter := NewBridgeSubmitter(context.Background(), &bridgeSC, hclog.NewNullLogger())
		require.NotNil(t, bridgeSubmitter)

		err := bridgeSubmitter.SubmitConfirmedBlocks("prime", []*indexer.CardanoBlock{})

		require.NoError(t, err)
	})
}
