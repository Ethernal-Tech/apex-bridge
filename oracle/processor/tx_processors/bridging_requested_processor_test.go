package txprocessors

import (
	"math/big"
	"strings"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/stretchr/testify/require"
)

func TestBridgingRequestedProcessor(t *testing.T) {
	const (
		utxoMinValue          = 1000000
		minFeeForBridging     = 10000010
		primeBridgingAddr     = "addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj660"
		primeBridgingFeeAddr  = "addr_test1vqqj5apwf5npsmudw0ranypkj9jw98t25wk4h83jy5mwypswekttt"
		vectorBridgingAddr    = "addr_test1vr076kzqu8ejq22y4e3j0rpck54nlvryd8sjkewjxzsrjgq2lszpw"
		vectorBridgingFeeAddr = "addr_test1vpg5t5gv784rmlze9ye0r9nud706d2v5v94d5h7kpvllamgq6yfx4"
		validTestAddress      = "addr_test1vq6zkfat4rlmj2nd2sylpjjg5qhcg9mk92wykaw4m2dp2rqneafvl"
	)

	proc := NewBridgingRequestedProcessor()
	appConfig := &core.AppConfig{
		CardanoChains: map[string]*core.CardanoChainConfig{
			"prime": {
				BridgingAddresses: core.BridgingAddresses{
					BridgingAddress: primeBridgingAddr,
					FeeAddress:      primeBridgingFeeAddr,
				},
			},
			"vector": {
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
	appConfig.FillOut()

	t.Run("bridging requested processor - IsTxRelevant", func(t *testing.T) {
		relevant, err := proc.IsTxRelevant(&core.CardanoTx{})
		require.Error(t, err)
		require.False(t, relevant)

		irrelevantMetadata, err := common.MarshalMetadata(common.MetadataEncodingTypeCbor, common.BaseMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
		})
		require.NoError(t, err)
		require.NotNil(t, irrelevantMetadata)

		relevant, err = proc.IsTxRelevant(&core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: irrelevantMetadata,
			},
		})
		require.NoError(t, err)
		require.False(t, relevant)

		relevantMetadata, err := common.MarshalMetadata(common.MetadataEncodingTypeCbor, common.BaseMetadata{
			BridgingTxType: common.BridgingTxTypeBridgingRequest,
		})
		require.NoError(t, err)
		require.NotNil(t, relevantMetadata)

		relevant, err = proc.IsTxRelevant(&core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: relevantMetadata,
			},
		})
		require.NoError(t, err)
		require.True(t, relevant)
	})

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &core.BridgeClaims{}

		err := proc.ValidateAndAddClaim(claims, &core.CardanoTx{}, appConfig)
		require.Error(t, err)
	})

	t.Run("ValidateAndAddClaim irrelevant metadata", func(t *testing.T) {
		irrelevantMetadata, err := common.MarshalMetadata(common.MetadataEncodingTypeCbor, common.BaseMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
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
		relevantButNotFullMetadata, err := common.MarshalMetadata(common.MetadataEncodingTypeCbor, common.BaseMetadata{
			BridgingTxType: common.BridgingTxTypeBridgingRequest,
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

	//nolint:dupl
	t.Run("ValidateAndAddClaim origin chain not registered", func(t *testing.T) {
		destinationChainNonRegisteredMetadata, err := common.MarshalMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: "invalid",
			SenderAddr:         []string{"addr1"},
			Transactions:       []common.BridgingRequestMetadataTransaction{},
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
			OriginChainID: "prime",
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "destination chain not registered")
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim destination chain not registered", func(t *testing.T) {
		destinationChainNonRegisteredMetadata, err := common.MarshalMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: "vector",
			SenderAddr:         []string{"addr1"},
			Transactions:       []common.BridgingRequestMetadataTransaction{},
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
			OriginChainID: "invalid",
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "unsupported chain id found in tx")
	})

	t.Run("ValidateAndAddClaim bridging addr not in utxos", func(t *testing.T) {
		bridgingAddrNotFoundInUtxosMetadata, err := common.MarshalMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: "vector",
			SenderAddr:         []string{"addr1"},
			Transactions:       []common.BridgingRequestMetadataTransaction{},
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
			OriginChainID: "prime",
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "bridging address on origin not found in utxos")
	})

	t.Run("ValidateAndAddClaim multiple utxos to bridging addr", func(t *testing.T) {
		multipleUtxosToBridgingAddrMetadata, err := common.MarshalMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: "vector",
			SenderAddr:         []string{"addr1"},
			Transactions:       []common.BridgingRequestMetadataTransaction{},
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
			OriginChainID: "prime",
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found multiple utxos to the bridging address on origin")
	})

	t.Run("ValidateAndAddClaim 6", func(t *testing.T) {
		feeAddrNotInReceiversMetadata, err := common.MarshalMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: "vector",
			SenderAddr:         []string{"addr1"},
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: []string{vectorBridgingFeeAddr}, Amount: 2},
				{Address: []string{vectorBridgingFeeAddr}, Amount: 2},
				{Address: []string{vectorBridgingFeeAddr}, Amount: 2},
				{Address: []string{vectorBridgingFeeAddr}, Amount: 2},
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
			OriginChainID: "prime",
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "number of receivers in metadata greater than maximum allowed")
	})

	t.Run("ValidateAndAddClaim fee addr not in receivers in metadata", func(t *testing.T) {
		feeAddrNotInReceiversMetadata, err := common.MarshalMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: "vector",
			SenderAddr:         []string{"addr1"},
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: []string{validTestAddress}, Amount: utxoMinValue},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, feeAddrNotInReceiversMetadata)

		claims := &core.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: utxoMinValue},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: feeAddrNotInReceiversMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: "prime",
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "destination chain fee address not found in receiver addrs in metadata")
	})

	t.Run("ValidateAndAddClaim utxo value below minimum in receivers in metadata", func(t *testing.T) {
		utxoValueBelowMinInReceiversMetadata, err := common.MarshalMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: "vector",
			SenderAddr:         []string{"addr1"},
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: []string{validTestAddress}, Amount: utxoMinValue},
				{Address: []string{vectorBridgingFeeAddr}, Amount: 2},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, utxoValueBelowMinInReceiversMetadata)

		claims := &core.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: utxoMinValue},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: utxoValueBelowMinInReceiversMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: "prime",
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found a utxo value below minimum value in metadata receivers")
	})

	t.Run("ValidateAndAddClaim invalid receiver addr in metadata", func(t *testing.T) {
		invalidAddrInReceiversMetadata, err := common.MarshalMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: "vector",
			SenderAddr:         []string{"addr1"},
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: []string{vectorBridgingFeeAddr}, Amount: utxoMinValue},
				{Address: []string{
					"addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj661",
				}, Amount: utxoMinValue},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, invalidAddrInReceiversMetadata)

		claims := &core.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: utxoMinValue},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: invalidAddrInReceiversMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: "prime",
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found an invalid receiver addr in metadata")
	})

	t.Run("ValidateAndAddClaim receivers amounts and multisig amount missmatch less", func(t *testing.T) {
		invalidAddrInReceiversMetadata, err := common.MarshalMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: "vector",
			SenderAddr:         []string{"addr1"},
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: []string{vectorBridgingFeeAddr}, Amount: utxoMinValue},
				{Address: []string{validTestAddress}, Amount: utxoMinValue},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, invalidAddrInReceiversMetadata)

		claims := &core.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: utxoMinValue + 1},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: invalidAddrInReceiversMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: "prime",
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "receivers amounts and multisig amount missmatch: expected 2000000 but got 1000001")
	})

	t.Run("ValidateAndAddClaim receivers amounts and multisig amount missmatch more", func(t *testing.T) {
		invalidAddrInReceiversMetadata, err := common.MarshalMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: "vector",
			SenderAddr:         []string{"addr1"},
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: []string{vectorBridgingFeeAddr}, Amount: utxoMinValue},
				{Address: []string{validTestAddress}, Amount: utxoMinValue},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, invalidAddrInReceiversMetadata)

		claims := &core.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: utxoMinValue*2 + 1},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: invalidAddrInReceiversMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: "prime",
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "receivers amounts and multisig amount missmatch: expected 2000000 but got 2000001")
	})

	t.Run("ValidateAndAddClaim fee in receivers less than minimum", func(t *testing.T) {
		feeInReceiversLessThanMinMetadata, err := common.MarshalMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: "vector",
			SenderAddr:         []string{"addr1"},
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: []string{vectorBridgingFeeAddr}, Amount: minFeeForBridging - 1},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, feeInReceiversLessThanMinMetadata)

		claims := &core.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: minFeeForBridging - 1},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: feeInReceiversLessThanMinMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: "prime",
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "bridging fee in metadata receivers is less than minimum")
	})

	t.Run("ValidateAndAddClaim valid", func(t *testing.T) {
		const (
			destinationChainID = "vector"
			txHash             = "test_hash"
		)

		receivers := []common.BridgingRequestMetadataTransaction{
			{Address: []string{
				vectorBridgingFeeAddr[:5],
				vectorBridgingFeeAddr[5:],
			}, Amount: minFeeForBridging},
			{Address: []string{validTestAddress}, Amount: utxoMinValue},
		}

		validMetadata, err := common.MarshalMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: destinationChainID,
			SenderAddr:         []string{"addr1"},
			Transactions:       receivers,
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &core.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: minFeeForBridging + utxoMinValue},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Hash:     txHash,
				Metadata: validMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: "prime",
		}, appConfig)
		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		require.Len(t, claims.BridgingRequestClaims, 1)
		require.Equal(t, txHash, claims.BridgingRequestClaims[0].ObservedTransactionHash)
		require.Equal(t, destinationChainID, claims.BridgingRequestClaims[0].DestinationChainID)
		require.Len(t, claims.BridgingRequestClaims[0].Receivers, len(receivers))
		require.Equal(t, strings.Join(receivers[0].Address, ""),
			claims.BridgingRequestClaims[0].Receivers[0].DestinationAddress)
		require.Equal(t, new(big.Int).SetUint64(receivers[0].Amount), claims.BridgingRequestClaims[0].Receivers[0].Amount)

		require.NotNil(t, claims.BridgingRequestClaims[0].OutputUTXO)
		require.Equal(t, new(big.Int).SetUint64(txOutputs[0].Amount), claims.BridgingRequestClaims[0].OutputUTXO.Amount)
	})
}
