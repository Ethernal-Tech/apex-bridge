package tx_processors

import (
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/require"
)

func TestBatchExecutedProcessor(t *testing.T) {

	proc := NewBatchExecutedProcessor()

	t.Run("IsTxRelevant", func(t *testing.T) {
		relevant, err := proc.IsTxRelevant(&core.CardanoTx{}, nil)
		require.Error(t, err)
		require.False(t, relevant)

		irrelevantMetadata, err := cbor.Marshal(core.BaseMetadataMap{
			Value: core.BaseMetadata{
				BridgingTxType: core.BridgingTxTypeBridgingRequest,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, irrelevantMetadata)

		relevant, err = proc.IsTxRelevant(&core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: irrelevantMetadata,
			},
		}, nil)
		require.NoError(t, err)
		require.False(t, relevant)

		relevantMetadata, err := cbor.Marshal(core.BaseMetadataMap{
			Value: core.BaseMetadata{
				BridgingTxType: core.BridgingTxTypeBatchExecution,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, relevantMetadata)

		relevant, err = proc.IsTxRelevant(&core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: relevantMetadata,
			},
		}, nil)
		require.NoError(t, err)
		require.True(t, relevant)
	})

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &core.BridgeClaims{}

		err := proc.ValidateAndAddClaim(claims, &core.CardanoTx{}, nil)
		require.Error(t, err)
	})

	t.Run("ValidateAndAddClaim irrelevant metadata", func(t *testing.T) {
		irrelevantMetadata, err := cbor.Marshal(core.BaseMetadataMap{
			Value: core.BaseMetadata{
				BridgingTxType: core.BridgingTxTypeBridgingRequest,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, irrelevantMetadata)

		claims := &core.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: irrelevantMetadata,
			},
		}, nil)
		require.Error(t, err)
	})

	t.Run("ValidateAndAddClaim valid but metadata not full", func(t *testing.T) {
		relevantButNotFullMetadata, err := cbor.Marshal(core.BaseMetadataMap{
			Value: core.BaseMetadata{
				BridgingTxType: core.BridgingTxTypeBatchExecution,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, relevantButNotFullMetadata)

		claims := &core.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: relevantButNotFullMetadata,
			},
		}, nil)
		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		require.Len(t, claims.BatchExecuted, 1)
		require.Equal(t, "", claims.BatchExecuted[0].TxHash)
		require.Equal(t, "", claims.BatchExecuted[0].BatchNonceId)
		require.Nil(t, claims.BatchExecuted[0].OutputUtxos)
	})

	t.Run("ValidateAndAddClaim valid full metadata", func(t *testing.T) {
		const batchNonceId = "1"
		relevantFullMetadata, err := cbor.Marshal(core.BatchExecutedMetadataMap{
			Value: core.BatchExecutedMetadata{
				BridgingTxType: core.BridgingTxTypeBatchExecution,
				BatchNonceId:   batchNonceId,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, relevantFullMetadata)

		claims := &core.BridgeClaims{}
		const txHash = "test_hash"
		txOutputs := []*indexer.TxOutput{
			{Address: "addr1", Amount: 1},
			{Address: "addr2", Amount: 2},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Hash:     txHash,
				Metadata: relevantFullMetadata,
				Outputs:  txOutputs,
			},
		}, nil)
		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		require.Len(t, claims.BatchExecuted, 1)
		require.Equal(t, txHash, claims.BatchExecuted[0].TxHash)
		require.Equal(t, batchNonceId, claims.BatchExecuted[0].BatchNonceId)
		require.NotNil(t, claims.BatchExecuted[0].OutputUtxos)
		require.Len(t, claims.BatchExecuted[0].OutputUtxos, len(txOutputs))
		require.Equal(t, claims.BatchExecuted[0].OutputUtxos[0].Address, txOutputs[0].Address)
		require.Equal(t, claims.BatchExecuted[0].OutputUtxos[0].Amount, txOutputs[0].Amount)
		require.Equal(t, claims.BatchExecuted[0].OutputUtxos[1].Address, txOutputs[1].Address)
		require.Equal(t, claims.BatchExecuted[0].OutputUtxos[1].Amount, txOutputs[1].Amount)
	})

	t.Run("validate method fail", func(t *testing.T) {
		var cardanoChains map[string]core.CardanoChainConfig = make(map[string]core.CardanoChainConfig)
		cardanoChains["prime"] = core.CardanoChainConfig{
			ChainId:    "prime",
			FeeAddress: "addr1",
			BridgingAddresses: map[string]core.BridgingAddress{"prime": core.BridgingAddress{
				ChainId:    "prime",
				Address:    "addr2",
				FeeAddress: "addr3",
			}},
		}

		config := core.AppConfig{
			CardanoChains:    cardanoChains,
			Bridge:           core.BridgeConfig{},
			Settings:         core.AppSettings{},
			BridgingSettings: core.BridgingSettings{},
		}
		tx := core.CardanoTx{
			OriginChainId: "prime",
			Tx: indexer.Tx{
				Inputs: append(make([]*indexer.TxInputOutput, 0), &indexer.TxInputOutput{
					Output: indexer.TxOutput{
						Address: "addr4",
						IsUsed:  true,
					},
				}),
			},
		}

		err := proc.validate(&tx, &core.BatchExecutedMetadata{}, &config)
		require.Error(t, err)
	})

	t.Run("validate method pass", func(t *testing.T) {
		var cardanoChains map[string]core.CardanoChainConfig = make(map[string]core.CardanoChainConfig)
		cardanoChains["prime"] = core.CardanoChainConfig{
			ChainId:    "prime",
			FeeAddress: "addr1",
			BridgingAddresses: map[string]core.BridgingAddress{"prime": core.BridgingAddress{
				ChainId:    "prime",
				Address:    "addr2",
				FeeAddress: "addr3",
			}},
		}

		config := core.AppConfig{
			CardanoChains:    cardanoChains,
			Bridge:           core.BridgeConfig{},
			Settings:         core.AppSettings{},
			BridgingSettings: core.BridgingSettings{},
		}
		tx := core.CardanoTx{
			OriginChainId: "prime",
			Tx: indexer.Tx{
				Inputs: append(make([]*indexer.TxInputOutput, 0), &indexer.TxInputOutput{
					Output: indexer.TxOutput{
						Address: "addr2",
						IsUsed:  true,
					},
				}),
			},
		}

		err := proc.validate(&tx, &core.BatchExecutedMetadata{}, &config)
		require.NoError(t, err)

		tx.Tx.Inputs[0].Output.Address = "addr3"
		err = proc.validate(&tx, &core.BatchExecutedMetadata{}, &config)
		require.NoError(t, err)
	})
}
