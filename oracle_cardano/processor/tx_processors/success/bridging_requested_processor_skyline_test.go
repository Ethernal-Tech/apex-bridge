package successtxprocessors

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"testing"

	brAddrManager "github.com/Ethernal-Tech/apex-bridge/bridging_addresses_manager"
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
		utxoMinValue             = 1000000
		defaultMinFeeForBridging = 2000010
		minFeeForBridgingTokens  = 1000010
		minOperationFee          = 1000010
		primeBridgingAddr        = "addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj660"
		primeBridgingAddr2       = "addr_test1wrz24vv4tvfqsywkxn36rv5zagys2d7euafcgt50gmpgqpq4ju9uv"
		primeBridgingFeeAddr     = "addr_test1vqqj5apwf5npsmudw0ranypkj9jw98t25wk4h83jy5mwypswekttt"
		cardanoBridgingAddr      = "addr_test1wrz24vv4tvfqsywkxn36rv5zagys2d7euafcgt50gmpgqpq4ju9uv"
		cardanoBridgingFeeAddr   = "addr_test1wq5dw0g9mpmjy0xd6g58kncapdf6vgcka9el4llhzwy5vhqz80tcq"
		validTestAddress         = "addr_test1wrz24vv4tvfqsywkxn36rv5zagys2d7euafcgt50gmpgqpq4ju9uv"
		validPrimeTestAddress    = "addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj660"

		policyID    = "29f8873beb52e126f207a2dfd50f7cff556806b5b4cba9834a7b26a8"
		testChainID = "test"
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
	maxTokenAmountAllowedToBridge := new(big.Int).SetUint64(100000000)

	brAddrManagerMock := &brAddrManager.BridgingAddressesManagerMock{}
	brAddrManagerMock.On("GetAllPaymentAddresses", common.ChainIDIntPrime).Return([]string{primeBridgingAddr, primeBridgingAddr2}, nil)
	brAddrManagerMock.On("GetPaymentAddressFromIndex", common.ChainIDIntPrime, uint8(0)).Return(primeBridgingAddr, true)
	brAddrManagerMock.On("GetPaymentAddressFromIndex", common.ChainIDIntPrime, uint8(1)).Return(primeBridgingAddr2, true)
	brAddrManagerMock.On("GetPaymentAddressFromIndex", common.ChainIDIntCardano, uint8(0)).Return(cardanoBridgingAddr, true)
	brAddrManagerMock.On("GetFeeMultisigAddress", common.ChainIDIntPrime).Return(primeBridgingFeeAddr)
	brAddrManagerMock.On("GetAllPaymentAddresses", common.ChainIDIntCardano).Return([]string{cardanoBridgingAddr}, nil)
	brAddrManagerMock.On("GetFeeMultisigAddress", common.ChainIDIntCardano).Return(cardanoBridgingFeeAddr)

	getAppConfig := func(refundEnabled bool) *cCore.AppConfig {
		appConfig := &cCore.AppConfig{
			BridgingAddressesManager: brAddrManagerMock,
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
						DefaultMinFeeForBridging: defaultMinFeeForBridging,
						MinFeeForBridgingTokens:  minFeeForBridgingTokens,
					},
					MinOperationFee: minOperationFee,
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
						DefaultMinFeeForBridging: defaultMinFeeForBridging,
						MinFeeForBridgingTokens:  minFeeForBridgingTokens,
					},
					MinOperationFee: minOperationFee,
				},
			},
			BridgingSettings: cCore.BridgingSettings{
				MaxReceiversPerBridgingRequest: 3,
				MaxAmountAllowedToBridge:       maxAmountAllowedToBridge,
				MaxTokenAmountAllowedToBridge:  maxTokenAmountAllowedToBridge,
				AllowedDirections: map[string][]string{
					common.ChainIDStrPrime:   {common.ChainIDStrCardano, testChainID},
					common.ChainIDStrCardano: {common.ChainIDStrPrime},
					testChainID:              {common.ChainIDStrPrime},
				},
			},
			RefundEnabled: refundEnabled,
		}
		appConfig.FillOut()

		return appConfig
	}

	chainInfos := map[string]*chain.CardanoChainInfo{
		common.ChainIDStrPrime:   {ProtocolParams: protocolParameters},
		common.ChainIDStrCardano: {ProtocolParams: protocolParameters},
	}

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &cCore.BridgeClaims{}

		appConfig := getAppConfig(false)

		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"ValidateAndAddClaim", claims, &core.CardanoTx{}, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err := proc.ValidateAndAddClaim(claims, &core.CardanoTx{}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to unmarshal metadata, err: EOF")
	})

	t.Run("ValidateAndAddClaim empty tx with refund", func(t *testing.T) {
		claims := &cCore.BridgeClaims{}

		appConfig := getAppConfig(true)

		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, &core.CardanoTx{}, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err := proc.ValidateAndAddClaim(claims, &core.CardanoTx{}, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim empty tx with refund err", func(t *testing.T) {
		claims := &cCore.BridgeClaims{}

		appConfig := getAppConfig(true)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, &core.CardanoTx{}, appConfig).Return(
			fmt.Errorf("test err"))

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err := proc.ValidateAndAddClaim(claims, &core.CardanoTx{}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("ValidateAndAddClaim irrelevant metadata", func(t *testing.T) {
		irrelevantMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BaseMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
		})
		require.NoError(t, err)
		require.NotNil(t, irrelevantMetadata)

		claims := &cCore.BridgeClaims{}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, &core.CardanoTx{}, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: irrelevantMetadata,
			},
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "ValidateAndAddClaim called for irrelevant tx")
	})

	t.Run("ValidateAndAddClaim irrelevant metadata with refund", func(t *testing.T) {
		irrelevantMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BaseMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
		})
		require.NoError(t, err)
		require.NotNil(t, irrelevantMetadata)

		claims := &cCore.BridgeClaims{}
		cardanoTx := &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: irrelevantMetadata,
			},
		}

		appConfig := getAppConfig(true)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

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

		cardanoTx := &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: relevantButNotFullMetadata,
			},
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "validation failed for tx")
	})

	t.Run("ValidateAndAddClaim - transaction direction not allowed - invalid destination chain", func(t *testing.T) {
		transactionDirectionNotSupportedMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: "invalid",
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, transactionDirectionNotSupportedMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: "addr1", Amount: 1},
			{Address: "addr2", Amount: 2},
			{Address: primeBridgingAddr, Amount: 3},
			{Address: primeBridgingFeeAddr, Amount: 4},
		}

		cardanoTx := &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: transactionDirectionNotSupportedMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrCardano,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "transaction direction not allowed")
	})

	t.Run("ValidateAndAddClaim destination chain not registered", func(t *testing.T) {
		destinationChainNonRegisteredMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: testChainID,
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

		cardanoTx := &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: destinationChainNonRegisteredMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "destination chain not registered")
	})

	t.Run("ValidateAndAddClaim unsupported chain id found in tx", func(t *testing.T) {
		destinationChainNonRegisteredMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrPrime,
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

		cardanoTx := &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: destinationChainNonRegisteredMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: testChainID,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "unsupported chain id found in tx")
	})

	t.Run("ValidateAndAddClaim less than minOperationFee", func(t *testing.T) {
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
			{Address: primeBridgingAddr, Amount: 1},
		}

		cardanoTx := &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: destinationChainNonRegisteredMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: destinationChainNonRegisteredMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "operation fee in metadata receivers is less than minimum")
	})

	t.Run("ValidateAndAddClaim multiple bridging addresses", func(t *testing.T) {
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
			{Address: primeBridgingAddr, Amount: 1},
			{Address: primeBridgingAddr2, Amount: 2},
		}

		cardanoTx := &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: destinationChainNonRegisteredMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: destinationChainNonRegisteredMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found multiple tx outputs to the bridging addresses on prime")
	})

	t.Run("ValidateAndAddClaim no bridging addrs in outputs", func(t *testing.T) {
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
			{Address: "addr1", Amount: 1},
			{Address: "addr2", Amount: 2},
		}

		cardanoTx := &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: multipleUtxosToBridgingAddrMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, fmt.Sprintf("none of bridging addresses on %s", common.ChainIDStrPrime))
	})

	t.Run("ValidateAndAddClaim multiple utxos to different bridging addr", func(t *testing.T) {
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

		cardanoTx := &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: multipleUtxosToBridgingAddrMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found multiple tx outputs to the bridging address")
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

		cardanoTx := &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "unknown tokens")
	})

	t.Run("ValidateAndAddClaim unknown tokens 2", func(t *testing.T) {
		appConfig := getAppConfig(false)
		appConfig.CardanoChains[common.ChainIDStrPrime].NativeTokens = nil

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
			}},
		}
		proc := NewSkylineBridgingRequestedProcessor(
			&RefundDisabledProcessorImpl{},
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: metadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "with some unknown tokens")
	})

	t.Run("ValidateAndAddClaim number of receivers greater than maximum allowed", func(t *testing.T) {
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

		cardanoTx := &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "number of receivers in metadata greater than maximum allowed")
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim fee amount is too low - default", func(t *testing.T) {
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
			BridgingFee:  defaultMinFeeForBridging - 2,
			OperationFee: minOperationFee,
		})
		require.NoError(t, err)
		require.NotNil(t, feeAddrNotInReceiversMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{
				Address: primeBridgingAddr,
				Amount:  utxoMinValue + 2_000_000 + (defaultMinFeeForBridging-1)*2,
			},
		}

		tx := indexer.Tx{
			Metadata: feeAddrNotInReceiversMetadata,
			Outputs:  txOutputs,
		}

		cardanoTx := &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "bridging fee in metadata receivers is less than minimum")
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim fee amount is too low - tokens", func(t *testing.T) {
		feeAddrNotInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
				{
					Address:            sendtx.AddrToMetaDataAddr(validTestAddress),
					Amount:             utxoMinValue,
					IsNativeTokenOnSrc: 1,
				},
			},
			BridgingFee:  minFeeForBridgingTokens - 2,
			OperationFee: minOperationFee,
		})
		require.NoError(t, err)
		require.NotNil(t, feeAddrNotInReceiversMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{
				Address: primeBridgingAddr,
				Amount:  utxoMinValue + 2_000_000 + (minFeeForBridgingTokens-1)*2,
			},
		}

		tx := indexer.Tx{
			Metadata: feeAddrNotInReceiversMetadata,
			Outputs:  txOutputs,
		}

		cardanoTx := &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "bridging fee in metadata receivers is less than minimum")
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
					Amount:             defaultMinFeeForBridging * 2,
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
				Amount:  utxoMinValue + minOperationFee + defaultMinFeeForBridging*2 + 200,
			},
		}

		cardanoTx := &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: metadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
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

		cardanoTx := &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found a utxo value below minimum value in metadata receivers")
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

		cardanoTx := &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found an invalid receiver addr in metadata")
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

		cardanoTx := &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found an invalid receiver addr in metadata")
	})

	t.Run("ValidateAndAddClaim receivers amounts and multisig amount missmatch less", func(t *testing.T) {
		invalidAddrInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
				{
					Address:            sendtx.AddrToMetaDataAddr(cardanoBridgingFeeAddr),
					Amount:             defaultMinFeeForBridging * 2,
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

		cardanoTx := &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "multisig amount is not equal to sum of receiver amounts + fee")
	})

	t.Run("ValidateAndAddClaim receivers amounts and multisig amount missmatch more", func(t *testing.T) {
		invalidAddrInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
				{
					Address:            sendtx.AddrToMetaDataAddr(cardanoBridgingFeeAddr),
					Amount:             defaultMinFeeForBridging * 2,
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

		cardanoTx := &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "multisig amount is not equal to sum of receiver amounts + fee")
	})

	t.Run("ValidateAndAddClaim fee in receivers less than minimum", func(t *testing.T) {
		feeInReceiversLessThanMinMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrCardano,
			SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
				{
					Address:            sendtx.AddrToMetaDataAddr(cardanoBridgingFeeAddr),
					Amount:             defaultMinFeeForBridging - 1,
					IsNativeTokenOnSrc: 0,
				},
			},
			OperationFee: minOperationFee,
		})
		require.NoError(t, err)
		require.NotNil(t, feeInReceiversLessThanMinMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: defaultMinFeeForBridging - 1},
		}

		tx := indexer.Tx{
			Metadata: feeInReceiversLessThanMinMetadata,
			Outputs:  txOutputs,
		}

		cardanoTx := &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "bridging fee in metadata receivers is less than minimum")
	})

	t.Run("ValidateAndAddClaim direction not allowed currency+native", func(t *testing.T) {
		for _, isNativeTokenOnSource := range []byte{0, 1} {
			// deep copy (clone) with json marshalling
			var newAppConfig *cCore.AppConfig

			appConfig := getAppConfig(false)

			bytes, err := json.Marshal(appConfig)
			require.NoError(t, err)

			require.NoError(t, json.Unmarshal(bytes, &newAppConfig))
			// because of `json:"-"`
			newAppConfig.CardanoChains[common.ChainIDStrCardano].NativeTokens = nil
			newAppConfig.CardanoChains[common.ChainIDStrCardano].ChainID = common.ChainIDStrCardano
			newAppConfig.CardanoChains[common.ChainIDStrPrime].ChainID = common.ChainIDStrPrime
			newAppConfig.BridgingAddressesManager = appConfig.BridgingAddressesManager

			srcChainID, dstChainID := common.ChainIDStrPrime, common.ChainIDStrCardano
			txOutput := &indexer.TxOutput{
				Address: primeBridgingAddr,
				Amount:  1_000_000,
			}

			if isNativeTokenOnSource == 1 {
				srcChainID, dstChainID = dstChainID, srcChainID
				txOutput.Address = cardanoBridgingAddr
			}

			validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
				BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
				DestinationChainID: dstChainID,
				SenderAddr:         sendtx.AddrToMetaDataAddr("addr1"),
				Transactions: []sendtx.BridgingRequestMetadataTransaction{
					{
						Address:            common.SplitString(validTestAddress, 40),
						Amount:             1_000_000,
						IsNativeTokenOnSrc: isNativeTokenOnSource,
					},
				},
				OperationFee: minOperationFee,
				BridgingFee:  defaultMinFeeForBridging,
			})
			require.NoError(t, err)
			require.NotNil(t, validMetadata)

			proc := NewSkylineBridgingRequestedProcessor(
				&RefundDisabledProcessorImpl{},
				hclog.NewNullLogger(),
				chainInfos,
			)

			err = proc.ValidateAndAddClaim(&cCore.BridgeClaims{}, &core.CardanoTx{
				Tx: indexer.Tx{
					Hash:     [32]byte(common.NewHashFromHexString("0x2244FF")),
					Metadata: validMetadata,
					Outputs:  []*indexer.TxOutput{txOutput},
				},
				OriginChainID: srcChainID,
			}, newAppConfig)
			require.Error(t, err)
			require.ErrorContains(t, err, "no native token specified for destination: "+common.ChainIDStrPrime)
		}
	})

	t.Run("ValidateAndAddClaim more than allowed", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrCardano

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []sendtx.BridgingRequestMetadataTransaction{
			{
				Address:            common.SplitString(cardanoBridgingFeeAddr, 40),
				Amount:             defaultMinFeeForBridging * 2,
				IsNativeTokenOnSrc: 0,
			},
			{
				Address:            sendtx.AddrToMetaDataAddr(validTestAddress),
				IsNativeTokenOnSrc: 0,
				Amount:             maxAmountAllowedToBridge.Uint64() + 1,
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
				Amount:  minOperationFee + defaultMinFeeForBridging*2 + maxAmountAllowedToBridge.Uint64() + 1,
			},
		}

		tx := indexer.Tx{
			Hash:     txHash,
			Metadata: validMetadata,
			Outputs:  txOutputs,
		}

		cardanoTx := &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "sum of receiver amounts + fee")
		require.ErrorContains(t, err, "greater than maximum allowed")
	})

	t.Run("ValidateAndAddClaim more tokens than allowed", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrCardano

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []sendtx.BridgingRequestMetadataTransaction{
			{
				Address:            common.SplitString(cardanoBridgingFeeAddr, 40),
				Amount:             minFeeForBridgingTokens * 2,
				IsNativeTokenOnSrc: 0,
			},
			{
				Address:            sendtx.AddrToMetaDataAddr(validTestAddress),
				IsNativeTokenOnSrc: 1,
				Amount:             maxTokenAmountAllowedToBridge.Uint64() * 2,
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
				Amount:  minOperationFee + minFeeForBridgingTokens*2,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: policyID,
						Name:     wrappedTokenPrime.Name,
						Amount:   maxTokenAmountAllowedToBridge.Uint64() * 2,
					},
				},
			},
		}

		cardanoTx := &core.CardanoTx{
			Tx: indexer.Tx{
				Hash:     txHash,
				Metadata: validMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "greater than maximum allowed")
	})

	t.Run("ValidateAndAddClaim - native token - currency under min allowed", func(t *testing.T) {
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
			BridgingFee:        minFeeForBridgingTokens,
			OperationFee:       minOperationFee,
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{
				Address: primeBridgingAddr,
				Amount:  minOperationFee + minFeeForBridgingTokens,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: policyID,
						Name:     wrappedTokenPrime.Name,
						Amount:   utxoMinValue,
					},
				},
			},
		}

		cardanoTx := &core.CardanoTx{
			Tx: indexer.Tx{
				Hash:     txHash,
				Metadata: validMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "sum of receiver amounts + fee is under the minimum allowed")
	})

	t.Run("ValidateAndAddClaim valid", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrCardano

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []sendtx.BridgingRequestMetadataTransaction{
			{
				Address:            common.SplitString(cardanoBridgingFeeAddr, 40),
				Amount:             minFeeForBridgingTokens * 2,
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
				Amount:  minOperationFee + minFeeForBridgingTokens*2,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: policyID,
						Name:     wrappedTokenPrime.Name,
						Amount:   utxoMinValue,
					},
				},
			},
		}

		cardanoTx := &core.CardanoTx{
			Tx: indexer.Tx{
				Hash:     txHash,
				Metadata: validMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
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

	t.Run("ValidateAndAddClaim - native token - not first bridging address", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrCardano

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []sendtx.BridgingRequestMetadataTransaction{
			{
				Address: common.SplitString(cardanoBridgingFeeAddr, 40),
				Amount:  minFeeForBridgingTokens * 2,
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
				Address: primeBridgingAddr2,
				Amount:  minOperationFee + minFeeForBridgingTokens*2,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: policyID,
						Name:     wrappedTokenPrime.Name,
						Amount:   utxoMinValue,
					},
				},
			},
		}

		appConfig := getAppConfig(false)

		cardanoTx := &core.CardanoTx{
			Tx: indexer.Tx{
				Hash:     txHash,
				Metadata: validMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}

		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "with some unknown tokens")
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim - native token - valid", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrCardano

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []sendtx.BridgingRequestMetadataTransaction{
			{
				Address: common.SplitString(cardanoBridgingFeeAddr, 40),
				Amount:  minFeeForBridgingTokens * 2,
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
				Amount:  minOperationFee + minFeeForBridgingTokens*2,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: policyID,
						Name:     wrappedTokenPrime.Name,
						Amount:   utxoMinValue,
					},
				},
			},
		}

		cardanoTx := &core.CardanoTx{
			Tx: indexer.Tx{
				Hash:     txHash,
				Metadata: validMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
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
				Amount:  minFeeForBridgingTokens * 2,
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
				Amount:  minOperationFee + minFeeForBridgingTokens*2,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: policyID,
						Name:     wrappedTokenPrime.Name,
						Amount:   utxoMinValue,
					},
				},
			},
		}

		cardanoTx := &core.CardanoTx{
			Tx: indexer.Tx{
				Hash:     txHash,
				Metadata: validMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
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
			BridgingFee:        minFeeForBridgingTokens * 3,
			OperationFee:       minOperationFee,
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{
				Address: primeBridgingAddr,
				Amount:  minOperationFee + minFeeForBridgingTokens*3,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: policyID,
						Name:     wrappedTokenPrime.Name,
						Amount:   utxoMinValue,
					},
				},
			},
		}

		cardanoTx := &core.CardanoTx{
			Tx: indexer.Tx{
				Hash:     txHash,
				Metadata: validMetadata,
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
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
			{Address: primeBridgingAddr, Amount: minFeeForBridgingTokens + utxoMinValue},
		}

		tx := indexer.Tx{
			Hash:     txHash,
			Metadata: validMetadata,
			Outputs:  txOutputs,
		}

		cardanoTx := &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "bridging fee in metadata receivers is less than minimum")
	})

	t.Run("ValidateAndAddClaim - native token - invalid #2", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrCardano

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
			{Address: primeBridgingAddr, Amount: minFeeForBridgingTokens + utxoMinValue},
		}

		tx := indexer.Tx{
			Hash:     txHash,
			Metadata: validMetadata,
			Outputs:  txOutputs,
		}

		cardanoTx := &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "bridging fee in metadata receivers is less than minimum")
	})

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
			BridgingFee:        minFeeForBridgingTokens * 3,
			OperationFee:       minOperationFee,
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{
				Address: cardanoBridgingAddr,
				Amount:  minOperationFee + minFeeForBridgingTokens*3,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: policyID,
						Name:     wrappedTokenCardano.Name,
						Amount:   utxoMinValue,
					},
				},
			},
		}

		tx := indexer.Tx{
			Hash:     txHash,
			Metadata: validMetadata,
			Outputs:  txOutputs,
		}

		cardanoTx := &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrCardano,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewSkylineBridgingRequestedProcessor(
			refundRequestProcessorMock,
			hclog.NewNullLogger(),
			chainInfos,
		)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
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
