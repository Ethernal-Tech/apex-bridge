package txprocessors

import (
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestFundProcessor(t *testing.T) {
	proc := NewFundProcessor(hclog.NewNullLogger())
	appConfig := core.AppConfig{
		CardanoChains: map[string]*core.CardanoChainConfig{"prime": {
			BridgingAddresses: core.BridgingAddresses{
				BridgingAddress: "addr_bridging",
				FeeAddress:      "addr_fee",
			},
		}},
		BridgingSettings: core.BridgingSettings{
			UtxoMinValue: 1000000,
		},
	}

	appConfig.FillOut()

	t.Run("ValidateAndAddClaim irrelevant metadata", func(t *testing.T) {
		irrelevantMetadata := []byte{1}
		require.NotNil(t, irrelevantMetadata)

		claims := &core.BridgeClaims{}
		err := proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: irrelevantMetadata,
			},
		}, &appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "ValidateAndAddClaim called for irrelevant tx")
	})

	t.Run("ValidateAndAddClaim fail on validate", func(t *testing.T) {
		relevantFullMetadata := []byte{}

		claims := &core.BridgeClaims{}

		const txHash = "test_hash"

		txOutputs := []*indexer.TxOutput{
			{Address: "addr_bridging", Amount: 1},
			{Address: "addr2", Amount: 2},
		}

		err := proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			OriginChainID: "prime",
			Tx: indexer.Tx{
				Hash:     txHash,
				Metadata: relevantFullMetadata,
				Outputs:  txOutputs,
			},
		}, &appConfig)

		require.Error(t, err)
		require.ErrorContains(t, err, "not enough to fund the bridging address")
	})

	t.Run("ValidateAndAddClaim valid full metadata", func(t *testing.T) {
		relevantFullMetadata := []byte{}

		claims := &core.BridgeClaims{}

		const txHash = "test_hash"

		txOutputs := []*indexer.TxOutput{
			{Address: "addr_bridging", Amount: 1000000},
			{Address: "addr2", Amount: 2},
		}

		err := proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			OriginChainID: "prime",
			Tx: indexer.Tx{
				Hash:     txHash,
				Metadata: relevantFullMetadata,
				Outputs:  txOutputs,
			},
		}, &appConfig)

		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		// require.Len(t, claims.BatchExecutedClaims, 1)
		// require.Equal(t, txHash, claims.BatchExecutedClaims[0].ObservedTransactionHash)
		// require.Equal(t, new(big.Int).SetUint64(batchNonceID), claims.BatchExecutedClaims[0].BatchNonceID)
		// require.NotNil(t, claims.BatchExecutedClaims[0].OutputUTXOs.MultisigOwnedUTXOs)
		// require.Len(t, claims.BatchExecutedClaims[0].OutputUTXOs.MultisigOwnedUTXOs, 1)
		// require.Equal(t, claims.BatchExecutedClaims[0].OutputUTXOs.MultisigOwnedUTXOs[0].Amount, new(big.Int).SetUint64(txOutputs[0].Amount))
		// require.NotNil(t, claims.BatchExecutedClaims[0].OutputUTXOs.FeePayerOwnedUTXOs)
		// require.Len(t, claims.BatchExecutedClaims[0].OutputUTXOs.FeePayerOwnedUTXOs, 1)
		// require.Equal(t, claims.BatchExecutedClaims[0].OutputUTXOs.FeePayerOwnedUTXOs[0].Amount, new(big.Int).SetUint64(txOutputs[1].Amount))
	})
}
