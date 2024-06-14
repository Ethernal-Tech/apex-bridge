package txprocessors

import (
	"encoding/hex"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestBatchExecutedProcessor(t *testing.T) {
	proc := NewBatchExecutedProcessor(hclog.NewNullLogger())

	appConfig := core.AppConfig{
		CardanoChains: map[string]*core.CardanoChainConfig{common.ChainIDStrPrime: {
			BridgingAddresses: core.BridgingAddresses{
				BridgingAddress: "addr_bridging",
				FeeAddress:      "addr_fee",
			},
		}},
	}
	appConfig.FillOut()

	txInputs := append(append(make([]*indexer.TxInputOutput, 0), &indexer.TxInputOutput{
		Input: indexer.TxInput{},
		Output: indexer.TxOutput{
			Address: "addr_bridging",
		},
	}), &indexer.TxInputOutput{
		Input: indexer.TxInput{},
		Output: indexer.TxOutput{
			Address: "addr_fee",
		},
	})

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &core.BridgeClaims{}

		err := proc.ValidateAndAddClaim(claims, &core.CardanoTx{}, &appConfig)
		require.Error(t, err)
	})

	t.Run("ValidateAndAddClaim irrelevant metadata", func(t *testing.T) {
		irrelevantMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BaseMetadata{
			BridgingTxType: common.BridgingTxTypeBridgingRequest,
		})
		require.NoError(t, err)
		require.NotNil(t, irrelevantMetadata)

		claims := &core.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: irrelevantMetadata,
			},
		}, &appConfig)
		require.Error(t, err)
	})

	t.Run("ValidateAndAddClaim valid but metadata not full", func(t *testing.T) {
		relevantButNotFullMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BaseMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
		})
		require.NoError(t, err)
		require.NotNil(t, relevantButNotFullMetadata)

		claims := &core.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			OriginChainID: common.ChainIDStrPrime,
			Tx: indexer.Tx{
				Metadata: relevantButNotFullMetadata,
				Inputs:   txInputs,
			},
		}, &appConfig)
		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		require.Len(t, claims.BatchExecutedClaims, 1)
		require.Equal(t, [32]byte{}, claims.BatchExecutedClaims[0].ObservedTransactionHash)
		require.Len(t, claims.BatchExecutedClaims[0].OutputUTXOs.MultisigOwnedUTXOs, 0)
	})

	t.Run("ValidateAndAddClaim fail on validate", func(t *testing.T) {
		const batchNonceID = uint64(1)
		relevantFullMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BatchExecutedMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
			BatchNonceID:   batchNonceID,
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
			OriginChainID: common.ChainIDStrPrime,
			Tx: indexer.Tx{
				Hash:     txHash,
				Metadata: relevantFullMetadata,
				Outputs:  txOutputs,
				Inputs: append(make([]*indexer.TxInputOutput, 0), &indexer.TxInputOutput{
					Input: indexer.TxInput{},
					Output: indexer.TxOutput{
						Address: "addr123",
					},
				}),
			},
		}, &appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "unexpected address found in tx input")
	})

	t.Run("ValidateAndAddClaim valid full metadata", func(t *testing.T) {
		batchNonceID := uint64(1)
		relevantFullMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BatchExecutedMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
			BatchNonceID:   batchNonceID,
		})
		require.NoError(t, err)
		require.NotNil(t, relevantFullMetadata)

		claims := &core.BridgeClaims{}
		txHashBytes := common.MustHashToBytes32("010203")
		txHash := hex.EncodeToString(txHashBytes[:])

		txOutputs := []*indexer.TxOutput{
			{Address: "addr_bridging", Amount: 1},
			{Address: "addr_fee", Amount: 2},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			OriginChainID: common.ChainIDStrPrime,
			Tx: indexer.Tx{
				Hash:     txHash,
				Metadata: relevantFullMetadata,
				Outputs:  txOutputs,
				Inputs:   txInputs,
			},
		}, &appConfig)
		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		require.Len(t, claims.BatchExecutedClaims, 1)
		require.Equal(t, txHash, hex.EncodeToString(claims.BatchExecutedClaims[0].ObservedTransactionHash[:]))
		require.Equal(t, batchNonceID, claims.BatchExecutedClaims[0].BatchNonceId)
		require.NotNil(t, claims.BatchExecutedClaims[0].OutputUTXOs.MultisigOwnedUTXOs)
		require.Len(t, claims.BatchExecutedClaims[0].OutputUTXOs.MultisigOwnedUTXOs, 1)
		require.Equal(t, claims.BatchExecutedClaims[0].OutputUTXOs.MultisigOwnedUTXOs[0].Amount, txOutputs[0].Amount)
		require.NotNil(t, claims.BatchExecutedClaims[0].OutputUTXOs.FeePayerOwnedUTXOs)
		require.Len(t, claims.BatchExecutedClaims[0].OutputUTXOs.FeePayerOwnedUTXOs, 1)
		require.Equal(t, claims.BatchExecutedClaims[0].OutputUTXOs.FeePayerOwnedUTXOs[0].Amount, txOutputs[1].Amount)
	})

	t.Run("validate method fail", func(t *testing.T) {
		cardanoChains := make(map[string]*core.CardanoChainConfig)
		cardanoChains[common.ChainIDStrPrime] = &core.CardanoChainConfig{
			BridgingAddresses: core.BridgingAddresses{
				BridgingAddress: "addr1",
				FeeAddress:      "addr2",
			},
		}

		config := &core.AppConfig{
			CardanoChains:    cardanoChains,
			Bridge:           core.BridgeConfig{},
			Settings:         core.AppSettings{},
			BridgingSettings: core.BridgingSettings{},
		}

		config.FillOut()

		inputs := []*indexer.TxInputOutput{
			{
				Output: indexer.TxOutput{
					Address: "addr3",
					IsUsed:  true,
				},
			},
		}
		tx := core.CardanoTx{
			OriginChainID: common.ChainIDStrPrime,
			Tx: indexer.Tx{
				Inputs: inputs,
			},
		}

		err := proc.validate(&tx, &common.BatchExecutedMetadata{}, config)
		require.Error(t, err)
		require.ErrorContains(t, err, "unexpected address found in tx input")

		tx.Inputs[0].Output.Address = "addr1"
		err = proc.validate(&tx, &common.BatchExecutedMetadata{}, config)
		require.Error(t, err)
		require.ErrorContains(t, err, "fee address not found in tx inputs")

		tx.Inputs[0].Output.Address = "addr2"
		err = proc.validate(&tx, &common.BatchExecutedMetadata{}, config)
		require.Error(t, err)
		require.ErrorContains(t, err, "bridging address not found in tx inputs")
	})

	t.Run("validate method origin chain not registered", func(t *testing.T) {
		cardanoChains := make(map[string]*core.CardanoChainConfig)

		config := &core.AppConfig{
			CardanoChains:    cardanoChains,
			Bridge:           core.BridgeConfig{},
			Settings:         core.AppSettings{},
			BridgingSettings: core.BridgingSettings{},
		}

		config.FillOut()

		inputs := []*indexer.TxInputOutput{
			{
				Output: indexer.TxOutput{
					Address: "addr3",
					IsUsed:  true,
				},
			},
		}
		tx := core.CardanoTx{
			OriginChainID: common.ChainIDStrPrime,
			Tx: indexer.Tx{
				Inputs: inputs,
			},
		}

		err := proc.validate(&tx, &common.BatchExecutedMetadata{}, config)
		require.Error(t, err)
		require.ErrorContains(t, err, "unsupported chain id found in tx")
	})

	t.Run("validate method pass", func(t *testing.T) {
		cardanoChains := make(map[string]*core.CardanoChainConfig)
		cardanoChains[common.ChainIDStrPrime] = &core.CardanoChainConfig{
			BridgingAddresses: core.BridgingAddresses{
				BridgingAddress: "addr1",
				FeeAddress:      "addr2",
			},
		}

		config := &core.AppConfig{
			CardanoChains:    cardanoChains,
			Bridge:           core.BridgeConfig{},
			Settings:         core.AppSettings{},
			BridgingSettings: core.BridgingSettings{},
		}

		config.FillOut()

		inputs := []*indexer.TxInputOutput{
			{
				Output: indexer.TxOutput{
					Address: "addr1",
					IsUsed:  true,
				},
			},
			{
				Output: indexer.TxOutput{
					Address: "addr2",
					IsUsed:  true,
				},
			},
		}
		tx := core.CardanoTx{
			OriginChainID: common.ChainIDStrPrime,
			Tx: indexer.Tx{
				Inputs: inputs,
			},
		}

		err := proc.validate(&tx, &common.BatchExecutedMetadata{}, config)
		require.NoError(t, err)

		tx.Tx.Inputs = append(tx.Tx.Inputs, &indexer.TxInputOutput{
			Output: indexer.TxOutput{
				Address: "addr1",
				IsUsed:  true,
			},
		})
		err = proc.validate(&tx, &common.BatchExecutedMetadata{}, config)
		require.NoError(t, err)

		tx.Tx.Inputs = append(tx.Tx.Inputs, &indexer.TxInputOutput{
			Output: indexer.TxOutput{
				Address: "addr2",
				IsUsed:  true,
			},
		})
		err = proc.validate(&tx, &common.BatchExecutedMetadata{}, config)
		require.NoError(t, err)
	})
}
