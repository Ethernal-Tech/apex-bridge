package successtxprocessors

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
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
		feeAddrBridgingAmount = uint64(1000001)
		primeBridgingAddr     = "addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj660"
		primeBridgingFeeAddr  = "addr_test1vqqj5apwf5npsmudw0ranypkj9jw98t25wk4h83jy5mwypswekttt"
		vectorBridgingAddr    = "addr_test1w2h482rf4gf44ek0rekamxksulazkr64yf2fhmm7f5gxjpsdm4zsg"
		vectorBridgingFeeAddr = "addr_test1vzv206r2s6c5y3rr9eexxnlppz8lm048empp8zvtwjkn9cqleec9x"
		validTestAddress      = "addr_test1vz68kkm248u5yze6cphql743lv3y34z65njw3x4j8vfcqwg0shpwd"
	)

	maxAmountAllowedToBridge := new(big.Int).SetUint64(100000000)

	getAppConfig := func(refundEnabled bool) *cCore.AppConfig {
		appConfig := &cCore.AppConfig{
			CardanoChains: map[string]*cCore.CardanoChainConfig{
				common.ChainIDStrPrime: {
					BridgingAddresses: cCore.BridgingAddresses{
						BridgingAddress: primeBridgingAddr,
						FeeAddress:      primeBridgingFeeAddr,
					},
					CardanoChainConfig: cardanotx.CardanoChainConfig{
						NetworkID:     wallet.TestNetNetwork,
						UtxoMinAmount: utxoMinValue,
					},
					MinFeeForBridging:     minFeeForBridging,
					FeeAddrBridgingAmount: feeAddrBridgingAmount,
				},
				common.ChainIDStrVector: {
					BridgingAddresses: cCore.BridgingAddresses{
						BridgingAddress: vectorBridgingAddr,
						FeeAddress:      vectorBridgingFeeAddr,
					},
					CardanoChainConfig: cardanotx.CardanoChainConfig{
						NetworkID:     wallet.TestNetNetwork,
						UtxoMinAmount: utxoMinValue,
					},
					MinFeeForBridging:     minFeeForBridging,
					FeeAddrBridgingAmount: feeAddrBridgingAmount,
				},
			},
			BridgingSettings: cCore.BridgingSettings{
				MaxReceiversPerBridgingRequest: 3,
				MaxAmountAllowedToBridge:       maxAmountAllowedToBridge,
			},
			RefundEnabled: refundEnabled,
		}
		appConfig.FillOut()

		return appConfig
	}

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &cCore.BridgeClaims{}

		appConfig := getAppConfig(false)

		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, &core.CardanoTx{}, appConfig).Return(nil)

		proc := NewBridgingRequestedProcessor(
			refundRequestProcessorMock, hclog.NewNullLogger())

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

		proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

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

		proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

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

		proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

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

		proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		refundRequestProcessorMock.On("ValidateAndAddClaim", claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: irrelevantMetadata,
			},
		}, appConfig).Return(nil)

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
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

		proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "unsupported chain id found in tx")
	})

	t.Run("ValidateAndAddClaim insufficient metadata with refund", func(t *testing.T) {
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

		appConfig := getAppConfig(true)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim destination chain not registered", func(t *testing.T) {
		destinationChainNonRegisteredMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: "invalid",
			SenderAddr:         cardanotx.AddrToMetaDataAddr("addr1"),
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

		proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "destination chain not registered")
	})

	t.Run("ValidateAndAddClaim origin chain not registered", func(t *testing.T) {
		destinationChainNonRegisteredMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         cardanotx.AddrToMetaDataAddr("addr1"),
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
		cardanoTx := &core.CardanoTx{
			Tx:            tx,
			OriginChainID: "invalid",
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "unsupported chain id found in tx")
	})

	t.Run("ValidateAndAddClaim forbidden transaction direction", func(t *testing.T) {
		destinationChainNonRegisteredMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrNexus,
			SenderAddr:         cardanotx.AddrToMetaDataAddr("addr1"),
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
		cardanoTx := &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrVector,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.CardanoTxSuccessRefundProcessorMock{
			SuccessProc: &core.CardanoTxSuccessProcessorMock{},
		}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, cardanoTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", cardanoTx, appConfig).Return(nil)

		proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "transaction direction not allowed")
	})

	t.Run("ValidateAndAddClaim bridging addr not in utxos", func(t *testing.T) {
		bridgingAddrNotFoundInUtxosMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         cardanotx.AddrToMetaDataAddr("addr1"),
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

		proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "bridging address")
		require.ErrorContains(t, err, "not found in tx outputs")
	})

	t.Run("ValidateAndAddClaim multiple utxos to bridging addr", func(t *testing.T) {
		multipleUtxosToBridgingAddrMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         cardanotx.AddrToMetaDataAddr("addr1"),
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

		proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found multiple tx outputs to the bridging address")
	})

	t.Run("ValidateAndAddClaim number of receivers greater than maximum allowed", func(t *testing.T) {
		feeAddrNotInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         cardanotx.AddrToMetaDataAddr("addr1"),
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: cardanotx.AddrToMetaDataAddr(vectorBridgingFeeAddr), Amount: 2},
				{Address: cardanotx.AddrToMetaDataAddr(vectorBridgingFeeAddr), Amount: 2},
				{Address: cardanotx.AddrToMetaDataAddr(vectorBridgingFeeAddr), Amount: 2},
				{Address: cardanotx.AddrToMetaDataAddr(vectorBridgingFeeAddr), Amount: 2},
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

		proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "number of receivers in metadata greater than maximum allowed")
	})

	t.Run("ValidateAndAddClaim fee amount is too low", func(t *testing.T) {
		feeAddrNotInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         cardanotx.AddrToMetaDataAddr("addr1"),
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: cardanotx.AddrToMetaDataAddr(validTestAddress), Amount: utxoMinValue},
			},
			BridgingFee: minFeeForBridging - 1,
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

		proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "bridging fee in metadata receivers is less than minimum")
	})

	t.Run("ValidateAndAddClaim fee amount is specified in receivers", func(t *testing.T) {
		metadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         cardanotx.AddrToMetaDataAddr("addr1"),
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: cardanotx.AddrToMetaDataAddr(validTestAddress), Amount: utxoMinValue},
				{Address: cardanotx.AddrToMetaDataAddr(vectorBridgingFeeAddr), Amount: minFeeForBridging},
			},
			BridgingFee: 100,
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: utxoMinValue + minFeeForBridging + 100},
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

		proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim utxo value below minimum in receivers in metadata", func(t *testing.T) {
		utxoValueBelowMinInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         cardanotx.AddrToMetaDataAddr("addr1"),
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: cardanotx.AddrToMetaDataAddr(validTestAddress), Amount: utxoMinValue},
				{Address: cardanotx.AddrToMetaDataAddr(vectorBridgingFeeAddr), Amount: 2},
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

		proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found a utxo value below minimum value in metadata receivers")
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim invalid receiver addr in metadata 1", func(t *testing.T) {
		invalidAddrInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         cardanotx.AddrToMetaDataAddr("addr1"),
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: cardanotx.AddrToMetaDataAddr(vectorBridgingFeeAddr), Amount: utxoMinValue},
				{Address: cardanotx.AddrToMetaDataAddr(
					"addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj661"), Amount: utxoMinValue},
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

		proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found an invalid receiver addr in metadata")
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim invalid receiver addr in metadata 2", func(t *testing.T) {
		invalidAddrInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         cardanotx.AddrToMetaDataAddr("addr1"),
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: cardanotx.AddrToMetaDataAddr(vectorBridgingFeeAddr), Amount: utxoMinValue},
				{Address: cardanotx.AddrToMetaDataAddr(
					"stake_test1urrzuuwrq6lfq82y9u642qzcwvkljshn0743hs0rpd5wz8s2pe23d"), Amount: utxoMinValue},
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

		proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found an invalid receiver addr in metadata")
	})

	t.Run("ValidateAndAddClaim receivers amounts and multisig amount missmatch less", func(t *testing.T) {
		invalidAddrInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         cardanotx.AddrToMetaDataAddr("addr1"),
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: cardanotx.AddrToMetaDataAddr(vectorBridgingFeeAddr), Amount: minFeeForBridging},
				{Address: cardanotx.AddrToMetaDataAddr(validTestAddress), Amount: utxoMinValue},
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

		proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "multisig amount is not equal to sum of receiver amounts + fee")
	})

	t.Run("ValidateAndAddClaim receivers amounts and multisig amount missmatch more", func(t *testing.T) {
		invalidAddrInReceiversMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         cardanotx.AddrToMetaDataAddr("addr1"),
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: cardanotx.AddrToMetaDataAddr(vectorBridgingFeeAddr), Amount: minFeeForBridging},
				{Address: cardanotx.AddrToMetaDataAddr(validTestAddress), Amount: utxoMinValue},
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

		proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "multisig amount is not equal to sum of receiver amounts + fee")
	})

	t.Run("ValidateAndAddClaim fee in receivers less than minimum", func(t *testing.T) {
		feeInReceiversLessThanMinMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         cardanotx.AddrToMetaDataAddr("addr1"),
			Transactions: []common.BridgingRequestMetadataTransaction{
				{Address: cardanotx.AddrToMetaDataAddr(vectorBridgingFeeAddr), Amount: minFeeForBridging - 1},
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

		proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "bridging fee in metadata receivers is less than minimum")
	})

	t.Run("ValidateAndAddClaim more than allowed", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrVector

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []common.BridgingRequestMetadataTransaction{
			{Address: cardanotx.AddrToMetaDataAddr(vectorBridgingFeeAddr), Amount: minFeeForBridging},
			{Address: cardanotx.AddrToMetaDataAddr(validTestAddress), Amount: maxAmountAllowedToBridge.Uint64() + 1},
		}

		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: destinationChainID,
			SenderAddr:         cardanotx.AddrToMetaDataAddr("addr1"),
			Transactions:       receivers,
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: minFeeForBridging + maxAmountAllowedToBridge.Uint64() + 1},
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

		proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "sum of receiver amounts + fee")
		require.ErrorContains(t, err, "greater than maximum allowed")
	})

	t.Run("ValidateAndAddClaim valid", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrVector

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []common.BridgingRequestMetadataTransaction{
			{Address: cardanotx.AddrToMetaDataAddr(vectorBridgingFeeAddr), Amount: minFeeForBridging},
			{Address: cardanotx.AddrToMetaDataAddr(validTestAddress), Amount: utxoMinValue},
		}

		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: destinationChainID,
			SenderAddr:         cardanotx.AddrToMetaDataAddr("addr1"),
			Transactions:       receivers,
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: minFeeForBridging + utxoMinValue},
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

		proc := NewBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, cardanoTx, appConfig)
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
		require.Equal(t, feeAddrBridgingAmount, claims.BridgingRequestClaims[0].Receivers[1].Amount.Uint64())
	})
}
