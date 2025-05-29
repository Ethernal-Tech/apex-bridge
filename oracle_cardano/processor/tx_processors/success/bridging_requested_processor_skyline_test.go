package successtxprocessors

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/chain"
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
		utxoMinValue           = 1000000
		minFeeForBridging      = 1000010
		minOperationFee        = 1000010
		primeBridgingAddr      = "addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj660"
		primeBridgingFeeAddr   = "addr_test1vqqj5apwf5npsmudw0ranypkj9jw98t25wk4h83jy5mwypswekttt"
		cardanoBridgingAddr    = "addr_test1wrz24vv4tvfqsywkxn36rv5zagys2d7euafcgt50gmpgqpq4ju9uv"
		cardanoBridgingFeeAddr = "addr_test1wq5dw0g9mpmjy0xd6g58kncapdf6vgcka9el4llhzwy5vhqz80tcq"
		validTestAddress       = "addr_test1wrz24vv4tvfqsywkxn36rv5zagys2d7euafcgt50gmpgqpq4ju9uv"
		validPrimeTestAddress  = "addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj660"

		policyID = "29f8873beb52e126f207a2dfd50f7cff556806b5b4cba9834a7b26a8"
	)

	wrappedTokenPrime, err := wallet.NewTokenWithFullName(
		fmt.Sprintf("%s.%s",
			policyID,
			hex.EncodeToString([]byte("wrappedApex"))), true,
	)
	require.NoError(t, err)

	wrappedTokenCardano, err := wallet.NewTokenWithFullName(
		fmt.Sprintf("%s.%s",
			policyID,
			hex.EncodeToString([]byte("wrappedCardano"))), true,
	)
	require.NoError(t, err)

	maxAmountAllowedToBridge := new(big.Int).SetUint64(100000000)

	refundRequestProcessorMock := &core.CardanoTxSuccessProcessorMock{}
	proc := NewSkylineBridgingRequestedProcessor(
		refundRequestProcessorMock,
		hclog.NewNullLogger(),
		map[string]*chain.CardanoChainInfo{
			common.ChainIDStrPrime:   {ProtocolParams: protocolParameters},
			common.ChainIDStrCardano: {ProtocolParams: protocolParameters},
		})

	//nolint:dupl
	appConfig := &cCore.AppConfig{
		CardanoChains: map[string]*cCore.CardanoChainConfig{
			common.ChainIDStrPrime: {
				CardanoChainConfig: cardanotx.CardanoChainConfig{
					NetworkID:     wallet.TestNetNetwork,
					UtxoMinAmount: utxoMinValue,
					NativeTokens: []sendtx.TokenExchangeConfig{
						{
							DstChainID: common.ChainIDStrVector,
							TokenName:  fmt.Sprintf("%s.%s", policyID, hex.EncodeToString([]byte("notimportant"))),
						},
						{
							DstChainID: common.ChainIDStrCardano,
							TokenName:  wrappedTokenPrime.String(),
						},
					},
				},
				BridgingAddresses: cCore.BridgingAddresses{
					BridgingAddress: primeBridgingAddr,
				},
				MinFeeForBridging: minFeeForBridging,
				MinOperationFee:   minOperationFee,
			},
			common.ChainIDStrCardano: {
				CardanoChainConfig: cardanotx.CardanoChainConfig{
					NetworkID:     wallet.TestNetNetwork,
					UtxoMinAmount: utxoMinValue,
					NativeTokens: []sendtx.TokenExchangeConfig{
						{
							DstChainID: common.ChainIDStrPrime,
							TokenName:  wrappedTokenCardano.String(),
						},
					},
				},
				BridgingAddresses: cCore.BridgingAddresses{
					BridgingAddress: cardanoBridgingAddr,
					FeeAddress:      cardanoBridgingFeeAddr,
				},
				MinFeeForBridging: minFeeForBridging,
				MinOperationFee:   minOperationFee,
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

	t.Run("ValidateAndAddClaim origin chain not registered", func(t *testing.T) {
		destinationChainNonRegisteredMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: "invalid",
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
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

		tx := indexer.Tx{
			Metadata: destinationChainNonRegisteredMetadata,
			Outputs:  txOutputs,
		}

		refundRequestProcessorMock.On("ValidateAndAddClaim", claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig).Return(nil)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: destinationChainNonRegisteredMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.NoError(t, err)
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim destination chain not registered", func(t *testing.T) {
		destinationChainNonRegisteredMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
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

	//nolint:dupl
	t.Run("ValidateAndAddClaim origin chain not registered", func(t *testing.T) {
		metadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: "addr1", Amount: 1},
			{Address: "addr2", Amount: 2},
			{Address: primeBridgingAddr, Amount: 3},
			{Address: primeBridgingFeeAddr, Amount: 4},
		}

		tx := indexer.Tx{
			Metadata: metadata,
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

	//nolint:dupl
	t.Run("ValidateAndAddClaim bridging addr not in utxos", func(t *testing.T) {
		bridgingAddrNotFoundInUtxosMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
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
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
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

	t.Run("ValidateAndAddClaim unknown tokens", func(t *testing.T) {
		metadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: 1, Tokens: []indexer.TokenAmount{
				{
					PolicyID: policyID,
					Name:     wrappedTokenPrime.Name,
					Amount:   utxoMinValue,
				},
				{
					PolicyID: "111",
					Name:     "222",
					Amount:   utxoMinValue,
				},
			}},
		}

		tx := indexer.Tx{
			Metadata: metadata,
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

	t.Run("ValidateAndAddClaim 6", func(t *testing.T) {
		feeAddrNotInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
				{Address: sendtx.AddrToMetaDataAddr(cardanoBridgingFeeAddr), Amount: 2},
				{Address: sendtx.AddrToMetaDataAddr(cardanoBridgingFeeAddr), Amount: 2},
				{Address: sendtx.AddrToMetaDataAddr(cardanoBridgingFeeAddr), Amount: 2},
				{Address: sendtx.AddrToMetaDataAddr(cardanoBridgingFeeAddr), Amount: 2},
			},
			OperationFee: minOperationFee,
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
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
				{
					Address:            sendtx.AddrToMetaDataAddr(validTestAddress),
					Amount:             utxoMinValue,
					IsNativeTokenOnSrc: 0,
				},
			},
			BridgingFee:  minFeeForBridging - 2,
			OperationFee: minOperationFee,
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
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
				{
					Address:            sendtx.AddrToMetaDataAddr(validTestAddress),
					Amount:             utxoMinValue,
					IsNativeTokenOnSrc: 0,
				},
				{
					Address:            common.SplitString(cardanoBridgingFeeAddr, 40),
					Amount:             minFeeForBridging * 2,
					IsNativeTokenOnSrc: 0,
				},
			},
			BridgingFee:  200,
			OperationFee: minOperationFee,
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{
				Address: primeBridgingAddr,
				Amount:  utxoMinValue + minOperationFee + minFeeForBridging*2 + 200,
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

	t.Run("ValidateAndAddClaim utxo value below minimum in receivers in metadata", func(t *testing.T) {
		utxoValueBelowMinInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
				{Address: sendtx.AddrToMetaDataAddr(validTestAddress), Amount: 2},
			},
			OperationFee: minOperationFee,
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
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
				{
					Address:            sendtx.AddrToMetaDataAddr(cardanoBridgingFeeAddr),
					Amount:             utxoMinValue,
					IsNativeTokenOnSrc: 0,
				},
				{
					Address:            sendtx.AddrToMetaDataAddr("addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj661"),
					Amount:             utxoMinValue,
					IsNativeTokenOnSrc: 0,
				},
			},
			OperationFee: minOperationFee,
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
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
				{
					Address:            sendtx.AddrToMetaDataAddr(cardanoBridgingFeeAddr),
					Amount:             utxoMinValue,
					IsNativeTokenOnSrc: 0,
				},
				{
					Address:            sendtx.AddrToMetaDataAddr("stake_test1urrzuuwrq6lfq82y9u642qzcwvkljshn0743hs0rpd5wz8s2pe23d"),
					Amount:             utxoMinValue,
					IsNativeTokenOnSrc: 0,
				},
			},
			OperationFee: minOperationFee,
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
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
				{
					Address:            sendtx.AddrToMetaDataAddr(cardanoBridgingFeeAddr),
					Amount:             minFeeForBridging * 2,
					IsNativeTokenOnSrc: 0,
				},
				{
					Address:            sendtx.AddrToMetaDataAddr(validTestAddress),
					Amount:             utxoMinValue,
					IsNativeTokenOnSrc: 0,
				},
			},
			OperationFee: minOperationFee,
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
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
				{
					Address:            sendtx.AddrToMetaDataAddr(cardanoBridgingFeeAddr),
					Amount:             minFeeForBridging * 2,
					IsNativeTokenOnSrc: 0,
				},
				{
					Address:            sendtx.AddrToMetaDataAddr(validTestAddress),
					Amount:             utxoMinValue,
					IsNativeTokenOnSrc: 0,
				},
			},
			OperationFee: minOperationFee,
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
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
				{
					Address:            sendtx.AddrToMetaDataAddr(cardanoBridgingFeeAddr),
					Amount:             minFeeForBridging - 1,
					IsNativeTokenOnSrc: 0,
				},
			},
			OperationFee: minOperationFee,
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
		const destinationChainID = common.ChainIDStrCardano

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []sendtx.BridgingRequestMetadataTransaction{
			{
				Address:            common.SplitString(cardanoBridgingFeeAddr, 40),
				Amount:             minFeeForBridging * 2,
				IsNativeTokenOnSrc: 0,
			},
			{
				Address:            sendtx.AddrToMetaDataAddr(validTestAddress),
				IsNativeTokenOnSrc: 0,
				Amount:             maxAmountAllowedToBridge.Uint64(),
			},
		}

		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: destinationChainID,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions:       receivers,
			OperationFee:       minOperationFee,
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{
				Address: primeBridgingAddr,
				Amount:  minOperationFee + minFeeForBridging*2 + maxAmountAllowedToBridge.Uint64(),
			},
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
		const destinationChainID = common.ChainIDStrCardano

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []sendtx.BridgingRequestMetadataTransaction{
			{
				Address:            common.SplitString(cardanoBridgingFeeAddr, 40),
				Amount:             minFeeForBridging * 2,
				IsNativeTokenOnSrc: 0,
			},
			{
				Address:            sendtx.AddrToMetaDataAddr(validTestAddress),
				IsNativeTokenOnSrc: 1,
				Amount:             utxoMinValue,
			},
		}

		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: destinationChainID,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions:       receivers,
			OperationFee:       minOperationFee,
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{
				Address: primeBridgingAddr,
				Amount:  minOperationFee + minFeeForBridging*2,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: policyID,
						Name:     wrappedTokenPrime.Name,
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
		const destinationChainID = common.ChainIDStrCardano

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []sendtx.BridgingRequestMetadataTransaction{
			{
				Address: common.SplitString(cardanoBridgingFeeAddr, 40),
				Amount:  minFeeForBridging * 2,
			},
			{
				Address:            sendtx.AddrToMetaDataAddr(validTestAddress),
				IsNativeTokenOnSrc: 1,
				Amount:             utxoMinValue,
			},
		}

		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: destinationChainID,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions:       receivers,
			OperationFee:       minOperationFee,
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{
				Address: primeBridgingAddr,
				Amount:  minOperationFee + minFeeForBridging*2,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: policyID,
						Name:     wrappedTokenPrime.Name,
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
		const destinationChainID = common.ChainIDStrCardano

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []sendtx.BridgingRequestMetadataTransaction{
			{
				Address: common.SplitString(cardanoBridgingFeeAddr, 40),
				Amount:  minFeeForBridging * 2,
			},
			{
				Address:            sendtx.AddrToMetaDataAddr(validTestAddress),
				IsNativeTokenOnSrc: 1,
				Amount:             utxoMinValue,
			},
		}

		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: destinationChainID,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions:       receivers,
			OperationFee:       minOperationFee,
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{
				Address: primeBridgingAddr,
				Amount:  minOperationFee + minFeeForBridging*2,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: policyID,
						Name:     wrappedTokenPrime.Name,
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
		const destinationChainID = common.ChainIDStrCardano

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []sendtx.BridgingRequestMetadataTransaction{
			{
				Address:            sendtx.AddrToMetaDataAddr(validTestAddress),
				IsNativeTokenOnSrc: 1,
				Amount:             utxoMinValue,
			},
		}

		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: destinationChainID,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions:       receivers,
			BridgingFee:        minFeeForBridging,
			OperationFee:       minOperationFee,
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{
				Address: primeBridgingAddr,
				Amount:  minOperationFee + minFeeForBridging,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: policyID,
						Name:     wrappedTokenPrime.Name,
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

	t.Run("ValidateAndAddClaim - native token - invalid", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrCardano

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []sendtx.BridgingRequestMetadataTransaction{
			{
				Address:            sendtx.AddrToMetaDataAddr(validTestAddress),
				IsNativeTokenOnSrc: 1,
				Amount:             utxoMinValue,
			},
		}

		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: destinationChainID,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions:       receivers,
			OperationFee:       minOperationFee,
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: minFeeForBridging + utxoMinValue},
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

	t.Run("ValidateAndAddClaim - native token - invalid #2", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrPrime

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []sendtx.BridgingRequestMetadataTransaction{
			{
				Address:            sendtx.AddrToMetaDataAddr(validPrimeTestAddress),
				IsNativeTokenOnSrc: 1,
				Amount:             utxoMinValue,
			},
		}

		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: destinationChainID,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions:       receivers,
			BridgingFee:        250_000,
			OperationFee:       minOperationFee,
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: minFeeForBridging + utxoMinValue},
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
			Tx: indexer.Tx{
				Hash:     txHash,
				Metadata: validMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.NoError(t, err)
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim - native token - valid #4", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrPrime

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []sendtx.BridgingRequestMetadataTransaction{
			{
				Address:            sendtx.AddrToMetaDataAddr(validPrimeTestAddress),
				IsNativeTokenOnSrc: 1,
				Amount:             utxoMinValue,
			},
		}

		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: destinationChainID,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions:       receivers,
			BridgingFee:        minFeeForBridging,
			OperationFee:       minOperationFee,
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{
				Address: cardanoBridgingAddr,
				Amount:  minOperationFee + minFeeForBridging,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: policyID,
						Name:     wrappedTokenCardano.Name,
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
			OriginChainID: common.ChainIDStrCardano,
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
}
