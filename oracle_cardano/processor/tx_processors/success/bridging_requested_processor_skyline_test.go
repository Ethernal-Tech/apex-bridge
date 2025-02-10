package successtxprocessors

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestBridgingRequestedProcessorSkyline(t *testing.T) {
	const (
		utxoMinValue          = 1000000
		minFeeForBridging     = 1000010
		primeBridgingAddr     = "addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj660"
		primeBridgingFeeAddr  = "addr_test1vqqj5apwf5npsmudw0ranypkj9jw98t25wk4h83jy5mwypswekttt"
		vectorBridgingAddr    = "vector_test1w2h482rf4gf44ek0rekamxksulazkr64yf2fhmm7f5gxjpsdm4zsg"
		vectorBridgingFeeAddr = "vector_test1wtyslvqxffyppmzhs7ecwunsnpq6g2p6kf9r4aa8ntfzc4qj925fr"
		validTestAddress      = "vector_test1vgrgxh4s35a5pdv0dc4zgq33crn34emnk2e7vnensf4tezq3tkm9m"
		validPrimeTestAddress = "addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj660"

		wrappedTokenPrime  = "wrappedApex"
		wrappedTokenVector = "wrappedVector"
	)

	maxAmountAllowedToBridge := new(big.Int).SetUint64(100000000)

	proc := NewSkylineBridgingRequestedProcessor(hclog.NewNullLogger())
	appConfig := &cCore.AppConfig{
		CardanoChains: map[string]*cCore.CardanoChainConfig{
			common.ChainIDStrPrime: {
				CardanoChainConfig: cardanotx.CardanoChainConfig{
					NetworkID:     wallet.TestNetNetwork,
					UtxoMinAmount: utxoMinValue,
					NativeTokens: []sendtx.TokenExchangeConfig{
						{
							DstChainID: common.ChainIDStrCardano,
							TokenName:  fmt.Sprintf("123.%s", hex.EncodeToString([]byte("notimportant"))),
						},
						{
							DstChainID: common.ChainIDStrVector,
							TokenName:  fmt.Sprintf("123.%s", hex.EncodeToString([]byte(wrappedTokenPrime))),
						},
					},
				},
				BridgingAddresses: cCore.BridgingAddresses{
					BridgingAddress: primeBridgingAddr,
				},
				MinFeeForBridging: minFeeForBridging,
			},
			common.ChainIDStrVector: {
				CardanoChainConfig: cardanotx.CardanoChainConfig{
					NetworkID:     wallet.VectorTestNetNetwork,
					UtxoMinAmount: utxoMinValue,
					NativeTokens: []sendtx.TokenExchangeConfig{
						{
							DstChainID: common.ChainIDStrPrime,
							TokenName:  fmt.Sprintf("%s.%s", "123", hex.EncodeToString([]byte(wrappedTokenVector))),
						},
					},
				},
				BridgingAddresses: cCore.BridgingAddresses{
					BridgingAddress: vectorBridgingAddr,
					FeeAddress:      vectorBridgingFeeAddr,
				},
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

		err := proc.ValidateAndAddClaim(claims, &core.CardanoTx{}, appConfig)
		require.Error(t, err)
	})

	t.Run("ValidateAndAddClaim irrelevant metadata", func(t *testing.T) {
		irrelevantMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BaseMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
		})
		require.NoError(t, err)
		require.NotNil(t, irrelevantMetadata)

		claims := &cCore.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: irrelevantMetadata,
			},
		}, appConfig)
		require.Error(t, err)
	})

	t.Run("ValidateAndAddClaim insufficient metadata", func(t *testing.T) {
		relevantButNotFullMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BaseMetadata{
			BridgingTxType: common.BridgingTxTypeBridgingRequest,
		})
		require.NoError(t, err)
		require.NotNil(t, relevantButNotFullMetadata)

		claims := &cCore.BridgeClaims{}
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
		destinationChainNonRegisteredMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: "invalid",
			SenderAddr:         []string{"addr1"},
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
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
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: destinationChainNonRegisteredMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "destination chain not registered")
	})

	t.Run("ValidateAndAddClaim destination chain not registered", func(t *testing.T) {
		destinationChainNonRegisteredMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"addr1"},
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
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
		bridgingAddrNotFoundInUtxosMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"addr1"},
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, bridgingAddrNotFoundInUtxosMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: "addr1", Amount: 1},
			{Address: "addr2", Amount: 2},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: bridgingAddrNotFoundInUtxosMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, fmt.Sprintf("bridging address %s on %s", primeBridgingAddr, common.ChainIDStrPrime))
	})

	t.Run("ValidateAndAddClaim multiple utxos to bridging addr", func(t *testing.T) {
		multipleUtxosToBridgingAddrMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"addr1"},
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, multipleUtxosToBridgingAddrMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: 1},
			{Address: primeBridgingAddr, Amount: 2},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: multipleUtxosToBridgingAddrMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found multiple tx outputs to the bridging address")
	})
	//nolint:dupl
	t.Run("ValidateAndAddClaim 6", func(t *testing.T) {
		feeAddrNotInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"addr1"},
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
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
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: feeAddrNotInReceiversMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "number of receivers in metadata greater than maximum allowed")
	})

	t.Run("ValidateAndAddClaim fee amount is too low", func(t *testing.T) {
		feeAddrNotInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"addr1"},
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
				{
					Address:            []string{validTestAddress},
					Amount:             utxoMinValue,
					IsNativeTokenOnSrc: 0,
					Additional: &sendtx.BridgingRequestMetadataCurrencyInfo{
						SrcAmount:  2_000_000,
						DestAmount: 1_000_000,
					}},
			},
			FeeAmount: sendtx.BridgingRequestMetadataCurrencyInfo{
				SrcAmount:  (minFeeForBridging - 2),
				DestAmount: (minFeeForBridging - 2) / 2},
		})
		require.NoError(t, err)
		require.NotNil(t, feeAddrNotInReceiversMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{
				Address: primeBridgingAddr,
				Amount:  utxoMinValue + 2_000_000 + (minFeeForBridging-1)*2,
			},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: feeAddrNotInReceiversMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "bridging fee in metadata receivers is less than minimum")
	})

	t.Run("ValidateAndAddClaim fee amount is specified in receivers", func(t *testing.T) {
		metadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"addr1"},
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
				{
					Address:            []string{validTestAddress},
					Amount:             utxoMinValue,
					IsNativeTokenOnSrc: 0,
					Additional: &sendtx.BridgingRequestMetadataCurrencyInfo{
						SrcAmount:  2_000_000,
						DestAmount: 1_000_000,
					},
				},
				{
					Address:            common.SplitString(vectorBridgingFeeAddr, 40),
					Amount:             minFeeForBridging * 2,
					IsNativeTokenOnSrc: 0,
				},
			},
			FeeAmount: sendtx.BridgingRequestMetadataCurrencyInfo{
				SrcAmount:  200,
				DestAmount: 100},
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{
				Address: primeBridgingAddr,
				Amount:  utxoMinValue + 2_000_000 + minFeeForBridging*2 + 200,
			},
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

	//nolint:dupl
	t.Run("ValidateAndAddClaim utxo value below minimum in receivers in metadata", func(t *testing.T) {
		utxoValueBelowMinInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"addr1"},
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
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
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: utxoValueBelowMinInReceiversMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found a utxo value below minimum value in metadata receivers")
	})

	t.Run("ValidateAndAddClaim invalid receiver addr in metadata 1", func(t *testing.T) {
		invalidAddrInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"addr1"},
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
				{
					Address:            []string{vectorBridgingFeeAddr},
					Amount:             utxoMinValue,
					IsNativeTokenOnSrc: 0,
				},
				{
					Address:            []string{"addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj661"},
					Amount:             utxoMinValue,
					IsNativeTokenOnSrc: 0,
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, invalidAddrInReceiversMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: utxoMinValue},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: invalidAddrInReceiversMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found an invalid receiver addr in metadata")
	})

	t.Run("ValidateAndAddClaim invalid receiver addr in metadata 2", func(t *testing.T) {
		invalidAddrInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"addr1"},
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
				{
					Address:            []string{vectorBridgingFeeAddr},
					Amount:             utxoMinValue,
					IsNativeTokenOnSrc: 0,
				},
				{
					Address:            []string{"stake_test1urrzuuwrq6lfq82y9u642qzcwvkljshn0743hs0rpd5wz8s2pe23d"},
					Amount:             utxoMinValue,
					IsNativeTokenOnSrc: 0,
					Additional: &sendtx.BridgingRequestMetadataCurrencyInfo{
						SrcAmount:  2_000_000,
						DestAmount: 1_000_000,
					},
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, invalidAddrInReceiversMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: utxoMinValue},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: invalidAddrInReceiversMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found an invalid receiver addr in metadata")
	})

	t.Run("ValidateAndAddClaim receivers amounts and multisig amount missmatch less", func(t *testing.T) {
		invalidAddrInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"addr1"},
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
				{
					Address:            []string{vectorBridgingFeeAddr},
					Amount:             minFeeForBridging * 2,
					IsNativeTokenOnSrc: 0,
				},
				{
					Address:            []string{validTestAddress},
					Amount:             utxoMinValue,
					IsNativeTokenOnSrc: 0,
					Additional: &sendtx.BridgingRequestMetadataCurrencyInfo{
						SrcAmount:  2_000_000,
						DestAmount: 1_000_000,
					},
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, invalidAddrInReceiversMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: utxoMinValue + 1},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: invalidAddrInReceiversMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "multisig amount is not equal to sum of receiver amounts + fee")
	})

	t.Run("ValidateAndAddClaim receivers amounts and multisig amount missmatch more", func(t *testing.T) {
		invalidAddrInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"addr1"},
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
				{
					Address:            []string{vectorBridgingFeeAddr},
					Amount:             minFeeForBridging * 2,
					IsNativeTokenOnSrc: 0,
				},
				{
					Address:            []string{validTestAddress},
					Amount:             utxoMinValue,
					IsNativeTokenOnSrc: 0,
					Additional: &sendtx.BridgingRequestMetadataCurrencyInfo{
						SrcAmount:  2_000_000,
						DestAmount: 1_000_000,
					},
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, invalidAddrInReceiversMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: utxoMinValue*2 + 1},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: invalidAddrInReceiversMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "multisig amount is not equal to sum of receiver amounts + fee")
	})

	t.Run("ValidateAndAddClaim fee in receivers less than minimum", func(t *testing.T) {
		feeInReceiversLessThanMinMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"addr1"},
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
				{
					Address:            []string{vectorBridgingFeeAddr},
					Amount:             minFeeForBridging - 1,
					IsNativeTokenOnSrc: 0,
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, feeInReceiversLessThanMinMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: minFeeForBridging - 1},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: feeInReceiversLessThanMinMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "bridging fee in metadata receivers is less than minimum")
	})

	t.Run("ValidateAndAddClaim more than allowed", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrVector

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []sendtx.BridgingRequestMetadataTransaction{
			{
				Address:            common.SplitString(vectorBridgingFeeAddr, 40),
				Amount:             minFeeForBridging * 2,
				IsNativeTokenOnSrc: 0,
			},
			{
				Address:            []string{validTestAddress},
				IsNativeTokenOnSrc: 0,
				Amount:             maxAmountAllowedToBridge.Uint64(),
				Additional: &sendtx.BridgingRequestMetadataCurrencyInfo{
					SrcAmount:  2_000_000,
					DestAmount: 1_000_000,
				},
			},
		}

		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: destinationChainID,
			SenderAddr:         []string{"addr1"},
			Transactions:       receivers,
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{
				Address: primeBridgingAddr,
				Amount:  minFeeForBridging*2 + maxAmountAllowedToBridge.Uint64() + 2_000_000,
			},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Hash:     txHash,
				Metadata: validMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "greater than maximum allowed")
	})

	t.Run("ValidateAndAddClaim valid", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrVector

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []sendtx.BridgingRequestMetadataTransaction{
			{
				Address:            common.SplitString(vectorBridgingFeeAddr, 40),
				Amount:             minFeeForBridging * 2,
				IsNativeTokenOnSrc: 0,
			},
			{
				Address:            []string{validTestAddress},
				IsNativeTokenOnSrc: 1,
				Amount:             utxoMinValue,
				Additional: &sendtx.BridgingRequestMetadataCurrencyInfo{
					SrcAmount:  1_000_000,
					DestAmount: 500_000,
				},
			},
		}

		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: destinationChainID,
			SenderAddr:         []string{"addr1"},
			Transactions:       receivers,
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{
				Address: primeBridgingAddr,
				Amount:  minFeeForBridging*2 + utxoMinValue,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "123",
						Name:     wrappedTokenPrime,
						Amount:   utxoMinValue,
					},
				},
			},
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
		// require.Equal(t, receivers[1].Amount, claims.BridgingRequestClaims[0].Receivers[0].Amount.Uint64())
		// require.Equal(t, receivers[0].Amount, claims.BridgingRequestClaims[0].Receivers[1].Amount.Uint64())
		require.Equal(t, strings.Join(receivers[0].Address, ""),
			claims.BridgingRequestClaims[0].Receivers[1].DestinationAddress)
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim - native token - valid", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrVector

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []sendtx.BridgingRequestMetadataTransaction{
			{
				Address: common.SplitString(vectorBridgingFeeAddr, 40),
				Amount:  minFeeForBridging * 2,
			},
			{
				Address:            []string{validTestAddress},
				IsNativeTokenOnSrc: 1,
				Amount:             utxoMinValue,
				Additional: &sendtx.BridgingRequestMetadataCurrencyInfo{
					SrcAmount:  1_000_000,
					DestAmount: 500_000,
				},
			},
		}

		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: destinationChainID,
			SenderAddr:         []string{"addr1"},
			Transactions:       receivers,
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{
				Address: primeBridgingAddr,
				Amount:  minFeeForBridging*2 + utxoMinValue,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "123",
						Name:     wrappedTokenPrime,
						Amount:   utxoMinValue,
					},
				},
			},
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
		// require.Equal(t, receivers[1].Amount, claims.BridgingRequestClaims[0].Receivers[0].Amount.Uint64())
		// require.Equal(t, receivers[0].Amount, claims.BridgingRequestClaims[0].Receivers[1].Amount.Uint64())
		require.Equal(t, strings.Join(receivers[0].Address, ""),
			claims.BridgingRequestClaims[0].Receivers[1].DestinationAddress)
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim - native token - valid #2", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrVector

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []sendtx.BridgingRequestMetadataTransaction{
			{
				Address: common.SplitString(vectorBridgingFeeAddr, 40),
				Amount:  minFeeForBridging * 2,
			},
			{
				Address:            []string{validTestAddress},
				IsNativeTokenOnSrc: 1,
				Amount:             utxoMinValue,
				Additional: &sendtx.BridgingRequestMetadataCurrencyInfo{
					SrcAmount:  1_000_000,
					DestAmount: 500_000,
				},
			},
		}

		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: destinationChainID,
			SenderAddr:         []string{"addr1"},
			Transactions:       receivers,
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{
				Address: primeBridgingAddr,
				Amount:  minFeeForBridging*2 + utxoMinValue,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "123",
						Name:     wrappedTokenPrime,
						Amount:   utxoMinValue,
					},
				},
			},
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
		// require.Equal(t, receivers[1].Amount, claims.BridgingRequestClaims[0].Receivers[0].Amount.Uint64())
		// require.Equal(t, receivers[0].Amount, claims.BridgingRequestClaims[0].Receivers[1].Amount.Uint64())
		require.Equal(t, strings.Join(receivers[0].Address, ""),
			claims.BridgingRequestClaims[0].Receivers[1].DestinationAddress)
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim - native token - valid #3", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrVector

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []sendtx.BridgingRequestMetadataTransaction{
			{
				Address:            []string{validTestAddress},
				IsNativeTokenOnSrc: 1,
				Amount:             utxoMinValue,
				Additional: &sendtx.BridgingRequestMetadataCurrencyInfo{
					SrcAmount:  2_000_000,
					DestAmount: 1_000_000,
				},
			},
		}

		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: destinationChainID,
			SenderAddr:         []string{"addr1"},
			Transactions:       receivers,
			FeeAmount: sendtx.BridgingRequestMetadataCurrencyInfo{
				SrcAmount:  2_000_020,
				DestAmount: 1_000_010,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{
				Address: primeBridgingAddr,
				Amount:  2_000_000 + 2_000_020,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "123",
						Name:     wrappedTokenPrime,
						Amount:   utxoMinValue,
					},
				},
			},
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
		require.Len(t, claims.BridgingRequestClaims[0].Receivers, len(receivers)+1)

		require.Equal(t, strings.Join(receivers[0].Address, ""),
			claims.BridgingRequestClaims[0].Receivers[0].DestinationAddress)
		// require.Equal(t, receivers[0].Amount, claims.BridgingRequestClaims[0].Receivers[0].Amount.Uint64())
		// require.Equal(t, receivers[0].Amount, claims.BridgingRequestClaims[0].Receivers[0].Amount.Uint64())
		require.Equal(t, strings.Join(receivers[0].Address, ""),
			claims.BridgingRequestClaims[0].Receivers[0].DestinationAddress)
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim - native token - invalid", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrVector

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []sendtx.BridgingRequestMetadataTransaction{
			{
				Address:            []string{validTestAddress},
				IsNativeTokenOnSrc: 1,
				Amount:             utxoMinValue,
				Additional: &sendtx.BridgingRequestMetadataCurrencyInfo{
					SrcAmount:  2_000_000,
					DestAmount: 1_000_000,
				},
			},
		}

		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: destinationChainID,
			SenderAddr:         []string{"addr1"},
			Transactions:       receivers,
			FeeAmount: sendtx.BridgingRequestMetadataCurrencyInfo{
				SrcAmount:  500_000,
				DestAmount: 250_000,
			},
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
		require.Error(t, err)
		require.ErrorContains(t, err, "bridging fee in metadata receivers is less than minimum")
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim - native token - invalid #2", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrPrime

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []sendtx.BridgingRequestMetadataTransaction{
			{
				Address:            []string{validPrimeTestAddress},
				IsNativeTokenOnSrc: 1,
				Amount:             utxoMinValue,
				Additional: &sendtx.BridgingRequestMetadataCurrencyInfo{
					SrcAmount:  500_000,
					DestAmount: 1_000_000,
				},
			},
		}

		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: destinationChainID,
			SenderAddr:         []string{"vector1"},
			Transactions:       receivers,
			FeeAmount: sendtx.BridgingRequestMetadataCurrencyInfo{
				SrcAmount:  250_000,
				DestAmount: 500_000,
			},
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
		require.Error(t, err)
		require.ErrorContains(t, err, "bridging fee in metadata receivers is less than minimum")
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim - native token - valid #4", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrPrime

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []sendtx.BridgingRequestMetadataTransaction{
			{
				Address:            []string{validPrimeTestAddress},
				IsNativeTokenOnSrc: 1,
				Amount:             utxoMinValue,
				Additional: &sendtx.BridgingRequestMetadataCurrencyInfo{
					SrcAmount:  500_000,
					DestAmount: 1_000_000,
				},
			},
		}

		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: destinationChainID,
			SenderAddr:         []string{"vector1"},
			Transactions:       receivers,
			FeeAmount: sendtx.BridgingRequestMetadataCurrencyInfo{
				SrcAmount:  1_000_010,
				DestAmount: 2_000_020,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{
				Address: vectorBridgingAddr,
				Amount:  1_000_010 + 500_000,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "123",
						Name:     wrappedTokenVector,
						Amount:   utxoMinValue,
					},
				},
			},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Hash:     txHash,
				Metadata: validMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrVector,
		}, appConfig)
		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		require.Len(t, claims.BridgingRequestClaims, 1)
		require.Equal(t, txHash, claims.BridgingRequestClaims[0].ObservedTransactionHash)
		require.Equal(t, destinationChainID, common.ToStrChainID(claims.BridgingRequestClaims[0].DestinationChainId))
		require.Len(t, claims.BridgingRequestClaims[0].Receivers, len(receivers)+1)

		require.Equal(t, strings.Join(receivers[0].Address, ""),
			claims.BridgingRequestClaims[0].Receivers[0].DestinationAddress)
		// require.Equal(t, receivers[0].Amount, claims.BridgingRequestClaims[0].Receivers[0].Amount.Uint64())
		// require.Equal(t, receivers[0].Amount, claims.BridgingRequestClaims[0].Receivers[0].Amount.Uint64())
		require.Equal(t, strings.Join(receivers[0].Address, ""),
			claims.BridgingRequestClaims[0].Receivers[0].DestinationAddress)
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim - native token - valid #5", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrPrime

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []sendtx.BridgingRequestMetadataTransaction{
			{
				Address:            []string{validPrimeTestAddress},
				IsNativeTokenOnSrc: 1,
				Amount:             utxoMinValue,
				Additional: &sendtx.BridgingRequestMetadataCurrencyInfo{
					SrcAmount:  1_000_000,
					DestAmount: 500_005,
				},
			},
		}

		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: destinationChainID,
			SenderAddr:         []string{"vector1"},
			Transactions:       receivers,
			FeeAmount: sendtx.BridgingRequestMetadataCurrencyInfo{
				SrcAmount:  500_000,
				DestAmount: 1_000_010,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: vectorBridgingAddr, Amount: minFeeForBridging + utxoMinValue},
		}
		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Hash:     txHash,
				Metadata: validMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrVector,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found an exchange rate error in metadata")
	})
}
