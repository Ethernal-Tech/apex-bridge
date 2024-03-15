package bridge

import (
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestClaimsSubmitter(t *testing.T) {
	appConfig := &core.AppConfig{
		Bridge: core.BridgeConfig{
			NodeUrl:              "https://polygon-mumbai-pokt.nodies.app",
			SmartContractAddress: "0xb2B87f7e652Aa847F98Cc05e130d030b91c7B37d",
			SigningKey:           "93c91e490bfd3736d17d04f53a10093e9cf2435309f4be1f5751381c8e201d23",
		},
	}

	t.Run("submit claims", func(t *testing.T) {
		claimsSubmitter := NewClaimsSubmitter(appConfig, hclog.NewNullLogger())
		require.NotNil(t, claimsSubmitter)

		defer claimsSubmitter.Dispose()

		err := claimsSubmitter.SubmitClaims(&core.BridgeClaims{
			BridgingRequest: []core.BridgingRequestClaim{
				{
					TxHash:             "test",
					DestinationChainId: "prime",
					OutputUtxos:        []core.Utxo{},
					Receivers:          []core.BridgingRequestReceiver{},
				},
			},
			BatchExecuted: []core.BatchExecutedClaim{
				{
					TxHash:       "test",
					BatchNonceId: "test",
					OutputUtxos:  []core.Utxo{},
				},
			},
		})

		require.NoError(t, err)
	})

	t.Run("dispose", func(t *testing.T) {
		claimsSubmitter := NewClaimsSubmitter(appConfig, hclog.NewNullLogger())
		require.NotNil(t, claimsSubmitter)

		client, err := ethclient.Dial(appConfig.Bridge.NodeUrl)
		require.NoError(t, err)

		claimsSubmitter.ethClient = client

		err = claimsSubmitter.Dispose()
		require.NoError(t, err)
		require.Nil(t, claimsSubmitter.ethClient)
	})
}
