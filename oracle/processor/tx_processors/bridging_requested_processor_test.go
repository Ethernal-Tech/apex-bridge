package tx_processors

import (
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/require"
)

func TestBridgingRequestedProcessor(t *testing.T) {

	const utxoMinValue = 1000000
	const minFeeForBridging = 10000010
	const primeBridgingAddr = "addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj660"
	const primeBridgingFeeAddr = "addr_test1vqqj5apwf5npsmudw0ranypkj9jw98t25wk4h83jy5mwypswekttt"
	const vectorBridgingAddr = "addr_test1vr076kzqu8ejq22y4e3j0rpck54nlvryd8sjkewjxzsrjgq2lszpw"
	const vectorBridgingFeeAddr = "addr_test1vpg5t5gv784rmlze9ye0r9nud706d2v5v94d5h7kpvllamgq6yfx4"

	proc := NewBridgingRequestedProcessor()
	appConfig := &core.AppConfig{
		CardanoChains: map[string]core.CardanoChainConfig{
			"prime": {
				ChainId: "prime",
				BridgingAddresses: core.BridgingAddresses{
					BridgingAddress: primeBridgingAddr,
					FeeAddress:      primeBridgingFeeAddr,
				},
			},
			"vector": {
				ChainId: "vector",
				BridgingAddresses: core.BridgingAddresses{
					BridgingAddress: vectorBridgingAddr,
					FeeAddress:      vectorBridgingFeeAddr,
				},
			},
		},
		BridgingSettings: core.BridgingSettings{
			MinFeeForBridging:              minFeeForBridging,
			UtxoMinValue:                   utxoMinValue,
			MaxReceiversPerBridgingRequest: 3,
		},
	}

	t.Run("IsTxRelevant", func(t *testing.T) {
		relevant, err := proc.IsTxRelevant(&core.CardanoTx{}, appConfig)
		require.Error(t, err)
		require.False(t, relevant)

		irrelevantMetadata, err := cbor.Marshal(core.BaseMetadataMap{
			Value: core.BaseMetadata{
				BridgingTxType: core.BridgingTxTypeBatchExecution,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, irrelevantMetadata)

		relevant, err = proc.IsTxRelevant(&core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: irrelevantMetadata,
			},
		}, appConfig)
		require.NoError(t, err)
		require.False(t, relevant)

		relevantMetadata, err := cbor.Marshal(core.BaseMetadataMap{
			Value: core.BaseMetadata{
				BridgingTxType: core.BridgingTxTypeBridgingRequest,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, relevantMetadata)

		relevant, err = proc.IsTxRelevant(&core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: relevantMetadata,
			},
		}, appConfig)
		require.NoError(t, err)
		require.True(t, relevant)
	})

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &core.BridgeClaims{}

		err := proc.ValidateAndAddClaim(claims, &core.CardanoTx{}, appConfig)
		require.Error(t, err)
	})

	t.Run("ValidateAndAddClaim irrelevant metadata", func(t *testing.T) {
		irrelevantMetadata, err := cbor.Marshal(core.BaseMetadataMap{
			Value: core.BaseMetadata{
				BridgingTxType: core.BridgingTxTypeBatchExecution,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, irrelevantMetadata)

		claims := &core.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: irrelevantMetadata,
			},
		}, appConfig)
		require.Error(t, err)
	})

	t.Run("ValidateAndAddClaim insufficient metadata", func(t *testing.T) {
		relevantButNotFullMetadata, err := cbor.Marshal(core.BaseMetadataMap{
			Value: core.BaseMetadata{
				BridgingTxType: core.BridgingTxTypeBridgingRequest,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, relevantButNotFullMetadata)

		claims := &core.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: relevantButNotFullMetadata,
			},
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "validation failed for tx")
	})

	t.Run("ValidateAndAddClaim destination chain not registered", func(t *testing.T) {
		destinationChainNonRegisteredMetadata, err := cbor.Marshal(core.BridgingRequestMetadataMap{
			Value: core.BridgingRequestMetadata{
				BridgingTxType:     core.BridgingTxTypeBridgingRequest,
				DestinationChainId: "invalid",
				SenderAddr:         "addr1",
				Transactions:       []core.BridgingRequestMetadataTransaction{},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, destinationChainNonRegisteredMetadata)

		claims := &core.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: "addr1", Amount: 1},
			{Address: "addr2", Amount: 2},
			{Address: primeBridgingAddr, Amount: 3},
			{Address: primeBridgingFeeAddr, Amount: 4},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: destinationChainNonRegisteredMetadata,
				Outputs:  txOutputs,
			},
			OriginChainId: "prime",
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "destination chain not registered")
	})

	t.Run("ValidateAndAddClaim bridging addr not in utxos", func(t *testing.T) {
		bridgingAddrNotFoundInUtxosMetadata, err := cbor.Marshal(core.BridgingRequestMetadataMap{
			Value: core.BridgingRequestMetadata{
				BridgingTxType:     core.BridgingTxTypeBridgingRequest,
				DestinationChainId: "vector",
				SenderAddr:         "addr1",
				Transactions:       []core.BridgingRequestMetadataTransaction{},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, bridgingAddrNotFoundInUtxosMetadata)

		claims := &core.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: "addr1", Amount: 1},
			{Address: "addr2", Amount: 2},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: bridgingAddrNotFoundInUtxosMetadata,
				Outputs:  txOutputs,
			},
			OriginChainId: "prime",
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "bridging address on origin not found in utxos")
	})

	t.Run("ValidateAndAddClaim multiple utxos to bridging addr", func(t *testing.T) {
		multipleUtxosToBridgingAddrMetadata, err := cbor.Marshal(core.BridgingRequestMetadataMap{
			Value: core.BridgingRequestMetadata{
				BridgingTxType:     core.BridgingTxTypeBridgingRequest,
				DestinationChainId: "vector",
				SenderAddr:         "addr1",
				Transactions:       []core.BridgingRequestMetadataTransaction{},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, multipleUtxosToBridgingAddrMetadata)

		claims := &core.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: 1},
			{Address: primeBridgingAddr, Amount: 2},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: multipleUtxosToBridgingAddrMetadata,
				Outputs:  txOutputs,
			},
			OriginChainId: "prime",
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found multiple utxos to the bridging address on origin")
	})

	t.Run("ValidateAndAddClaim 6", func(t *testing.T) {
		feeAddrNotInReceiversMetadata, err := cbor.Marshal(core.BridgingRequestMetadataMap{
			Value: core.BridgingRequestMetadata{
				BridgingTxType:     core.BridgingTxTypeBridgingRequest,
				DestinationChainId: "vector",
				SenderAddr:         "addr1",
				Transactions: []core.BridgingRequestMetadataTransaction{
					{Address: vectorBridgingFeeAddr, Amount: 2},
					{Address: vectorBridgingFeeAddr, Amount: 2},
					{Address: vectorBridgingFeeAddr, Amount: 2},
					{Address: vectorBridgingFeeAddr, Amount: 2},
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, feeAddrNotInReceiversMetadata)

		claims := &core.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: 1},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: feeAddrNotInReceiversMetadata,
				Outputs:  txOutputs,
			},
			OriginChainId: "prime",
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "number of receivers in metadata greater than maximum allowed")
	})

	t.Run("ValidateAndAddClaim fee addr not in receivers in metadata", func(t *testing.T) {
		feeAddrNotInReceiversMetadata, err := cbor.Marshal(core.BridgingRequestMetadataMap{
			Value: core.BridgingRequestMetadata{
				BridgingTxType:     core.BridgingTxTypeBridgingRequest,
				DestinationChainId: "vector",
				SenderAddr:         "addr1",
				Transactions:       []core.BridgingRequestMetadataTransaction{},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, feeAddrNotInReceiversMetadata)

		claims := &core.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: 1},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: feeAddrNotInReceiversMetadata,
				Outputs:  txOutputs,
			},
			OriginChainId: "prime",
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "destination chain fee address not found in receiver addrs in metadata")
	})

	t.Run("ValidateAndAddClaim utxo value below minimum in receivers in metadata", func(t *testing.T) {
		utxoValueBelowMinInReceiversMetadata, err := cbor.Marshal(core.BridgingRequestMetadataMap{
			Value: core.BridgingRequestMetadata{
				BridgingTxType:     core.BridgingTxTypeBridgingRequest,
				DestinationChainId: "vector",
				SenderAddr:         "addr1",
				Transactions: []core.BridgingRequestMetadataTransaction{
					{Address: vectorBridgingFeeAddr, Amount: 2},
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, utxoValueBelowMinInReceiversMetadata)

		claims := &core.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: 1},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: utxoValueBelowMinInReceiversMetadata,
				Outputs:  txOutputs,
			},
			OriginChainId: "prime",
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found a utxo value below minimum value in metadata receivers")
	})

	t.Run("ValidateAndAddClaim invalid receiver addr in metadata", func(t *testing.T) {
		invalidAddrInReceiversMetadata, err := cbor.Marshal(core.BridgingRequestMetadataMap{
			Value: core.BridgingRequestMetadata{
				BridgingTxType:     core.BridgingTxTypeBridgingRequest,
				DestinationChainId: "vector",
				SenderAddr:         "addr1",
				Transactions: []core.BridgingRequestMetadataTransaction{
					{Address: vectorBridgingFeeAddr, Amount: utxoMinValue},
					{Address: "addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj661", Amount: utxoMinValue},
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, invalidAddrInReceiversMetadata)

		claims := &core.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: 1},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: invalidAddrInReceiversMetadata,
				Outputs:  txOutputs,
			},
			OriginChainId: "prime",
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found an invalid receiver addr in metadata")
	})

	t.Run("ValidateAndAddClaim fee in receivers less than minimum", func(t *testing.T) {
		feeInReceiversLessThanMinMetadata, err := cbor.Marshal(core.BridgingRequestMetadataMap{
			Value: core.BridgingRequestMetadata{
				BridgingTxType:     core.BridgingTxTypeBridgingRequest,
				DestinationChainId: "vector",
				SenderAddr:         "addr1",
				Transactions: []core.BridgingRequestMetadataTransaction{
					{Address: vectorBridgingFeeAddr, Amount: minFeeForBridging - 1},
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, feeInReceiversLessThanMinMetadata)

		claims := &core.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: 1},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: feeInReceiversLessThanMinMetadata,
				Outputs:  txOutputs,
			},
			OriginChainId: "prime",
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "bridging fee in metadata receivers is less than minimum")
	})

	t.Run("ValidateAndAddClaim valid", func(t *testing.T) {
		const destinationChainId = "vector"
		const txHash = "test_hash"
		receivers := []core.BridgingRequestMetadataTransaction{
			{Address: vectorBridgingFeeAddr, Amount: minFeeForBridging},
		}

		validMetadata, err := cbor.Marshal(core.BridgingRequestMetadataMap{
			Value: core.BridgingRequestMetadata{
				BridgingTxType:     core.BridgingTxTypeBridgingRequest,
				DestinationChainId: destinationChainId,
				SenderAddr:         "addr1",
				Transactions:       receivers,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &core.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: 1},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Hash:     txHash,
				Metadata: validMetadata,
				Outputs:  txOutputs,
			},
			OriginChainId: "prime",
		}, appConfig)
		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		require.Len(t, claims.BridgingRequest, 1)
		require.Equal(t, txHash, claims.BridgingRequest[0].TxHash)
		require.Equal(t, destinationChainId, claims.BridgingRequest[0].DestinationChainId)
		require.Len(t, claims.BridgingRequest[0].Receivers, len(receivers))
		require.Equal(t, receivers[0].Address, claims.BridgingRequest[0].Receivers[0].Address)
		require.Equal(t, receivers[0].Amount, claims.BridgingRequest[0].Receivers[0].Amount)

		require.Len(t, claims.BridgingRequest[0].OutputUtxos, len(txOutputs))
		require.Equal(t, txOutputs[0].Address, claims.BridgingRequest[0].OutputUtxos[0].Address)
		require.Equal(t, txOutputs[0].Amount, claims.BridgingRequest[0].OutputUtxos[0].Amount)
	})
}
