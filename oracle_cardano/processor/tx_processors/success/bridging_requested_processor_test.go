package successtxprocessors

import (
	"math/big"
	"strings"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestBridgingRequestedProcessor(t *testing.T) {
	const (
		utxoMinValue          = 1000000
		minFeeForBridging     = 1000010
		primeBridgingAddr     = "addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj660"
		primeBridgingFeeAddr  = "addr_test1vqqj5apwf5npsmudw0ranypkj9jw98t25wk4h83jy5mwypswekttt"
		vectorBridgingAddr    = "vector_test1w2h482rf4gf44ek0rekamxksulazkr64yf2fhmm7f5gxjpsdm4zsg"
		vectorBridgingFeeAddr = "vector_test1wtyslvqxffyppmzhs7ecwunsnpq6g2p6kf9r4aa8ntfzc4qj925fr"
		validTestAddress      = "vector_test1vgrgxh4s35a5pdv0dc4zgq33crn34emnk2e7vnensf4tezq3tkm9m"
	)

	maxAmountAllowedToBridge := new(big.Int).SetUint64(100000000)

	refundRequestProcessorMock := &core.CardanoTxSuccessProcessorMock{}
	proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

	appConfig := &cCore.AppConfig{
		CardanoChains: map[string]*cCore.CardanoChainConfig{
			common.ChainIDStrPrime: {
				NetworkID: wallet.TestNetNetwork,
				BridgingAddresses: cCore.BridgingAddresses{
					BridgingAddress: primeBridgingAddr,
					FeeAddress:      primeBridgingFeeAddr,
				},
				UtxoMinAmount:     utxoMinValue,
				MinFeeForBridging: minFeeForBridging,
			},
			common.ChainIDStrVector: {
				NetworkID: wallet.VectorTestNetNetwork,
				BridgingAddresses: cCore.BridgingAddresses{
					BridgingAddress: vectorBridgingAddr,
					FeeAddress:      vectorBridgingFeeAddr,
				},
				UtxoMinAmount:     utxoMinValue,
				MinFeeForBridging: minFeeForBridging,
			},
		},
		BridgingSettings: cCore.BridgingSettings{
			MaxReceiversPerBridgingRequest: 3,
			MaxAmountAllowedToBridge:       maxAmountAllowedToBridge,
		},
	}
	appConfig.FillOut()

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &cCore.BridgeClaims{}

		refundRequestProcessorMock.On("ValidateAndAddClaim", claims, &core.CardanoTx{}, appConfig).Return(nil)

		err := proc.ValidateAndAddClaim(claims, &core.CardanoTx{}, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim irrelevant metadata", func(t *testing.T) {
		irrelevantMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BaseMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
		})
		require.NoError(t, err)
		require.NotNil(t, irrelevantMetadata)

		claims := &cCore.BridgeClaims{}

		refundRequestProcessorMock.On("ValidateAndAddClaim", claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: irrelevantMetadata,
			},
		}, appConfig).Return(nil)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: irrelevantMetadata,
			},
		}, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim insufficient metadata", func(t *testing.T) {
		relevantButNotFullMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BaseMetadata{
			BridgingTxType: common.BridgingTxTypeBridgingRequest,
		})
		require.NoError(t, err)
		require.NotNil(t, relevantButNotFullMetadata)

		claims := &cCore.BridgeClaims{}

		refundRequestProcessorMock.On("ValidateAndAddClaim", claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: relevantButNotFullMetadata,
			},
		}, appConfig).Return(nil)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: relevantButNotFullMetadata,
			},
		}, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim destination chain not registered", func(t *testing.T) {
		destinationChainNonRegisteredMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: "invalid",
			SenderAddr:         []string{"addr1"},
			Transactions:       []common.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, destinationChainNonRegisteredMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: "addr1", Amount: 1},
			{Address: "addr2", Amount: 2},
			{Address: primeBridgingAddr, Amount: 3},
			{Address: primeBridgingFeeAddr, Amount: 4},
		}

		tx := indexer.Tx{
			Metadata: destinationChainNonRegisteredMetadata,
			Outputs:  txOutputs,
		}

		refundRequestProcessorMock.On("ValidateAndAddClaim", claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig).Return(nil)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim origin chain not registered", func(t *testing.T) {
		destinationChainNonRegisteredMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"addr1"},
			Transactions:       []common.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, destinationChainNonRegisteredMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: "addr1", Amount: 1},
			{Address: "addr2", Amount: 2},
			{Address: primeBridgingAddr, Amount: 3},
			{Address: primeBridgingFeeAddr, Amount: 4},
		}

		tx := indexer.Tx{
			Metadata: destinationChainNonRegisteredMetadata,
			Outputs:  txOutputs,
		}

		refundRequestProcessorMock.On("ValidateAndAddClaim", claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: "invalid",
		}, appConfig).Return(nil)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: "invalid",
		}, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim forbidden transaction direction", func(t *testing.T) {
		destinationChainNonRegisteredMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrNexus,
			SenderAddr:         []string{"addr1"},
			Transactions:       []common.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, destinationChainNonRegisteredMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: "addr1", Amount: 1},
			{Address: "addr2", Amount: 2},
			{Address: vectorBridgingAddr, Amount: 3},
			{Address: vectorBridgingAddr, Amount: 4},
		}

		tx := indexer.Tx{
			Metadata: destinationChainNonRegisteredMetadata,
			Outputs:  txOutputs,
		}

		refundRequestProcessorMock.On("ValidateAndAddClaim", claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrVector,
		}, appConfig).Return(nil)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrVector,
		}, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim bridging addr not in utxos", func(t *testing.T) {
		bridgingAddrNotFoundInUtxosMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"addr1"},
			Transactions:       []common.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, bridgingAddrNotFoundInUtxosMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: "addr1", Amount: 1},
			{Address: "addr2", Amount: 2},
		}

		tx := indexer.Tx{
			Metadata: bridgingAddrNotFoundInUtxosMetadata,
			Outputs:  txOutputs,
		}

		refundRequestProcessorMock.On("ValidateAndAddClaim", claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig).Return(nil)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim multiple utxos to bridging addr", func(t *testing.T) {
		multipleUtxosToBridgingAddrMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"addr1"},
			Transactions:       []common.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, multipleUtxosToBridgingAddrMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: 1},
			{Address: primeBridgingAddr, Amount: 2},
		}

		tx := indexer.Tx{
			Metadata: multipleUtxosToBridgingAddrMetadata,
			Outputs:  txOutputs,
		}

		refundRequestProcessorMock.On("ValidateAndAddClaim", claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig).Return(nil)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim number of receivers greater than maximum allowed", func(t *testing.T) {
		feeAddrNotInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
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

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: 1},
		}

		tx := indexer.Tx{
			Metadata: feeAddrNotInReceiversMetadata,
			Outputs:  txOutputs,
		}

		refundRequestProcessorMock.On("ValidateAndAddClaim", claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig).Return(nil)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim fee amount is too low", func(t *testing.T) {
		feeAddrNotInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"addr1"},
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: []string{validTestAddress}, Amount: utxoMinValue},
			},
			FeeAmount: minFeeForBridging - 1,
		})
		require.NoError(t, err)
		require.NotNil(t, feeAddrNotInReceiversMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: utxoMinValue},
		}

		tx := indexer.Tx{
			Metadata: feeAddrNotInReceiversMetadata,
			Outputs:  txOutputs,
		}

		refundRequestProcessorMock.On("ValidateAndAddClaim", claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig).Return(nil)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim fee amount is specified in receivers", func(t *testing.T) {
		metadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"addr1"},
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: []string{validTestAddress}, Amount: utxoMinValue},
				{Address: common.SplitString(vectorBridgingFeeAddr, 40), Amount: minFeeForBridging},
			},
			FeeAmount: 100,
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: utxoMinValue + minFeeForBridging + 100},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: metadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim utxo value below minimum in receivers in metadata", func(t *testing.T) {
		utxoValueBelowMinInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"addr1"},
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: []string{validTestAddress}, Amount: utxoMinValue},
				{Address: []string{vectorBridgingFeeAddr}, Amount: 2},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, utxoValueBelowMinInReceiversMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: utxoMinValue},
		}

		tx := indexer.Tx{
			Metadata: utxoValueBelowMinInReceiversMetadata,
			Outputs:  txOutputs,
		}

		refundRequestProcessorMock.On("ValidateAndAddClaim", claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig).Return(nil)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.NoError(t, err)
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim invalid receiver addr in metadata 1", func(t *testing.T) {
		invalidAddrInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
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

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: utxoMinValue},
		}

		tx := indexer.Tx{
			Metadata: invalidAddrInReceiversMetadata,
			Outputs:  txOutputs,
		}

		refundRequestProcessorMock.On("ValidateAndAddClaim", claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig).Return(nil)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.NoError(t, err)
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim invalid receiver addr in metadata 2", func(t *testing.T) {
		invalidAddrInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"addr1"},
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: []string{vectorBridgingFeeAddr}, Amount: utxoMinValue},
				{Address: []string{
					"stake_test1urrzuuwrq6lfq82y9u642qzcwvkljshn0743hs0rpd5wz8s2pe23d",
				}, Amount: utxoMinValue},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, invalidAddrInReceiversMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: utxoMinValue},
		}

		tx := indexer.Tx{
			Metadata: invalidAddrInReceiversMetadata,
			Outputs:  txOutputs,
		}

		refundRequestProcessorMock.On("ValidateAndAddClaim", claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig).Return(nil)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim receivers amounts and multisig amount missmatch less", func(t *testing.T) {
		invalidAddrInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"addr1"},
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: []string{vectorBridgingFeeAddr}, Amount: minFeeForBridging},
				{Address: []string{validTestAddress}, Amount: utxoMinValue},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, invalidAddrInReceiversMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: utxoMinValue + 1},
		}

		tx := indexer.Tx{
			Metadata: invalidAddrInReceiversMetadata,
			Outputs:  txOutputs,
		}

		refundRequestProcessorMock.On("ValidateAndAddClaim", claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig).Return(nil)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim receivers amounts and multisig amount missmatch more", func(t *testing.T) {
		invalidAddrInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"addr1"},
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: []string{vectorBridgingFeeAddr}, Amount: minFeeForBridging},
				{Address: []string{validTestAddress}, Amount: utxoMinValue},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, invalidAddrInReceiversMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: utxoMinValue*2 + 1},
		}

		tx := indexer.Tx{
			Metadata: invalidAddrInReceiversMetadata,
			Outputs:  txOutputs,
		}

		refundRequestProcessorMock.On("ValidateAndAddClaim", claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig).Return(nil)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim fee in receivers less than minimum", func(t *testing.T) {
		feeInReceiversLessThanMinMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"addr1"},
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: []string{vectorBridgingFeeAddr}, Amount: minFeeForBridging - 1},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, feeInReceiversLessThanMinMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: minFeeForBridging - 1},
		}

		tx := indexer.Tx{
			Metadata: feeInReceiversLessThanMinMetadata,
			Outputs:  txOutputs,
		}

		refundRequestProcessorMock.On("ValidateAndAddClaim", claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig).Return(nil)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim more than allowed", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrVector

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []common.BridgingRequestMetadataTransaction{
			{Address: common.SplitString(vectorBridgingFeeAddr, 40), Amount: minFeeForBridging},
			{Address: []string{validTestAddress}, Amount: maxAmountAllowedToBridge.Uint64()},
		}

		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: destinationChainID,
			SenderAddr:         []string{"addr1"},
			Transactions:       receivers,
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: minFeeForBridging + maxAmountAllowedToBridge.Uint64()},
		}

		tx := indexer.Tx{
			Hash:     txHash,
			Metadata: validMetadata,
			Outputs:  txOutputs,
		}

		refundRequestProcessorMock.On("ValidateAndAddClaim", claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig).Return(nil)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim valid", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrVector

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []common.BridgingRequestMetadataTransaction{
			{Address: common.SplitString(vectorBridgingFeeAddr, 40), Amount: minFeeForBridging},
			{Address: []string{validTestAddress}, Amount: utxoMinValue},
		}

		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: destinationChainID,
			SenderAddr:         []string{"addr1"},
			Transactions:       receivers,
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: minFeeForBridging + utxoMinValue},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Hash:     txHash,
				Metadata: validMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		require.Len(t, claims.BridgingRequestClaims, 1)
		require.Equal(t, txHash, claims.BridgingRequestClaims[0].ObservedTransactionHash)
		require.Equal(t, destinationChainID, common.ToStrChainID(claims.BridgingRequestClaims[0].DestinationChainId))
		require.Len(t, claims.BridgingRequestClaims[0].Receivers, len(receivers))
		require.Equal(t, strings.Join(receivers[1].Address, ""),
			claims.BridgingRequestClaims[0].Receivers[0].DestinationAddress)
		require.Equal(t, receivers[1].Amount, claims.BridgingRequestClaims[0].Receivers[0].Amount.Uint64())
		require.Equal(t, strings.Join(receivers[0].Address, ""),
			claims.BridgingRequestClaims[0].Receivers[1].DestinationAddress)
		require.Equal(t, receivers[0].Amount, claims.BridgingRequestClaims[0].Receivers[1].Amount.Uint64())
	})
}
