package successtxprocessors

import (
	"fmt"
	"math/big"
	"testing"

	brAddrManager "github.com/Ethernal-Tech/apex-bridge/bridging_addresses_manager"
	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestBridgingRequestedProcessor(t *testing.T) {
	const (
		utxoMinValue          = 1000000
		minFeeForBridging     = 1000010
		feeAddrBridgingAmount = uint64(1000005)
		primeBridgingAddr     = "addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj660"
		primeBridgingFeeAddr  = "addr_test1vqqj5apwf5npsmudw0ranypkj9jw98t25wk4h83jy5mwypswekttt"
		nexusBridgingAddr     = "0xA4d1233A67776575425Ab185f6a9251aa00fEA25"
		validTestAddress      = "addr_test1vq6zkfat4rlmj2nd2sylpjjg5qhcg9mk92wykaw4m2dp2rqneafvl"
	)

	maxAmountAllowedToBridge := new(big.Int).SetUint64(100000000)

	proc := NewEthBridgingRequestedProcessor(hclog.NewNullLogger())

	brAddrManagerMock := &brAddrManager.BridgingAddressesManagerMock{}
	brAddrManagerMock.On("GetAllPaymentAddresses", common.ChainIDIntPrime).Return([]string{primeBridgingAddr}, nil)
	brAddrManagerMock.On("GetFeeMultisigAddress", common.ChainIDIntPrime).Return(primeBridgingFeeAddr)

	appConfig := &oCore.AppConfig{
		BridgingAddressesManager: brAddrManagerMock,
		CardanoChains: map[string]*oCore.CardanoChainConfig{
			common.ChainIDStrPrime: {
				CardanoChainConfig: cardanotx.CardanoChainConfig{
					NetworkID:     wallet.TestNetNetwork,
					UtxoMinAmount: utxoMinValue,
				},
				MinFeeForBridging:     minFeeForBridging,
				FeeAddrBridgingAmount: feeAddrBridgingAmount,
			},
		},
		EthChains: map[string]*oCore.EthChainConfig{
			common.ChainIDStrNexus: {
				BridgingAddresses: oCore.EthBridgingAddresses{
					BridgingAddress: nexusBridgingAddr,
				},
				MinFeeForBridging: minFeeForBridging,
			},
		},
		BridgingSettings: oCore.BridgingSettings{
			MaxReceiversPerBridgingRequest: 3,
			MaxAmountAllowedToBridge:       maxAmountAllowedToBridge,
		},
	}

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &oCore.BridgeClaims{}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, &core.EthTx{}, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err := proc.ValidateAndAddClaim(claims, &core.EthTx{}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "unexpected end of JSON input")
	})

	t.Run("ValidateAndAddClaim empty tx with refund", func(t *testing.T) {
		claims := &oCore.BridgeClaims{}

		appConfig := getAppConfig(true)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, &core.EthTx{}, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err := proc.ValidateAndAddClaim(claims, &core.EthTx{}, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim empty tx with refund err", func(t *testing.T) {
		claims := &oCore.BridgeClaims{}

		appConfig := getAppConfig(true)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, &core.EthTx{}, appConfig).Return(fmt.Errorf("test err"))

		proc := NewEthBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err := proc.ValidateAndAddClaim(claims, &core.EthTx{}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("ValidateAndAddClaim irrelevant metadata", func(t *testing.T) {
		irrelevantMetadata, err := core.MarshalEthMetadata(core.BaseEthMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
		})
		require.NoError(t, err)
		require.NotNil(t, irrelevantMetadata)

		claims := &oCore.BridgeClaims{}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, &core.EthTx{}, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, &core.EthTx{
			Metadata: irrelevantMetadata,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "ValidateAndAddClaim called for irrelevant tx")
	})

	t.Run("ValidateAndAddClaim irrelevant metadata with refund", func(t *testing.T) {
		irrelevantMetadata, err := core.MarshalEthMetadata(core.BaseEthMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
		})
		require.NoError(t, err)
		require.NotNil(t, irrelevantMetadata)

		claims := &oCore.BridgeClaims{}
		ethTx := &core.EthTx{
			Metadata: irrelevantMetadata,
		}

		appConfig := getAppConfig(true)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim insufficient metadata", func(t *testing.T) {
		relevantButNotFullMetadata, err := core.MarshalEthMetadata(core.BaseEthMetadata{
			BridgingTxType: common.BridgingTxTypeBridgingRequest,
		})
		require.NoError(t, err)
		require.NotNil(t, relevantButNotFullMetadata)

		claims := &oCore.BridgeClaims{}

		ethTx := &core.EthTx{
			Metadata: relevantButNotFullMetadata,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, ethTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "origin chain not registered")
	})

	t.Run("ValidateAndAddClaim insufficient metadata with refund", func(t *testing.T) {
		relevantButNotFullMetadata, err := core.MarshalEthMetadata(core.BaseEthMetadata{
			BridgingTxType: common.BridgingTxTypeBridgingRequest,
		})
		require.NoError(t, err)
		require.NotNil(t, relevantButNotFullMetadata)

		claims := &oCore.BridgeClaims{}
		ethTx := &core.EthTx{
			Metadata: relevantButNotFullMetadata,
		}

		appConfig := getAppConfig(true)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, ethTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim origin chain not registered", func(t *testing.T) {
		metadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: "invalid",
			SenderAddr:         "addr1",
			Transactions:       []core.BridgingRequestEthMetadataTransaction{},
			BridgingFee:        big.NewInt(0),
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &oCore.BridgeClaims{}

		ethTx := &core.EthTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrPrime,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, ethTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "origin chain not registered")
	})

	t.Run("ValidateAndAddClaim destination chain not registered", func(t *testing.T) {
		destinationChainNonRegisteredMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: "invalid",
			SenderAddr:         "addr1",
			Transactions:       []core.BridgingRequestEthMetadataTransaction{},
			BridgingFee:        big.NewInt(0),
		})
		require.NoError(t, err)
		require.NotNil(t, destinationChainNonRegisteredMetadata)

		claims := &oCore.BridgeClaims{}

		ethTx := &core.EthTx{
			Metadata:      destinationChainNonRegisteredMetadata,
			OriginChainID: common.ChainIDStrNexus,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, ethTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "destination chain not registered")
	})

	t.Run("ValidateAndAddClaim forbidden transaction direction", func(t *testing.T) {
		destinationChainNonRegisteredMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         "addr1",
			Transactions:       []core.BridgingRequestEthMetadataTransaction{},
			BridgingFee:        big.NewInt(0),
		})
		require.NoError(t, err)
		require.NotNil(t, destinationChainNonRegisteredMetadata)

		claims := &oCore.BridgeClaims{}

		ethTx := &core.EthTx{
			Metadata:      destinationChainNonRegisteredMetadata,
			OriginChainID: common.ChainIDStrNexus,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, ethTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "transaction direction not allowed")
	})

	t.Run("ValidateAndAddClaim more than max receivers in metadata", func(t *testing.T) {
		moreThanMaxReceiversReceiversMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{Address: primeBridgingFeeAddr, Amount: big.NewInt(2)},
				{Address: primeBridgingFeeAddr, Amount: big.NewInt(2)},
				{Address: primeBridgingFeeAddr, Amount: big.NewInt(2)},
				{Address: primeBridgingFeeAddr, Amount: big.NewInt(2)},
			},
			BridgingFee: big.NewInt(0),
		})
		require.NoError(t, err)
		require.NotNil(t, moreThanMaxReceiversReceiversMetadata)

		claims := &oCore.BridgeClaims{}

		ethTx := &core.EthTx{
			Metadata:      moreThanMaxReceiversReceiversMetadata,
			OriginChainID: common.ChainIDStrNexus,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, ethTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "number of receivers in metadata greater than maximum allowed")
	})

	t.Run("ValidateAndAddClaim fee amount is too low", func(t *testing.T) {
		metadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{Address: validTestAddress, Amount: common.DfmToWei(new(big.Int).SetUint64(utxoMinValue))},
			},
			BridgingFee: common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging - 1)),
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &oCore.BridgeClaims{}

		ethTx := &core.EthTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrNexus,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, ethTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "bridging fee in metadata receivers is less than minimum")
	})

	t.Run("ValidateAndAddClaim fee amount is specified in receivers", func(t *testing.T) {
		metadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{Address: validTestAddress, Amount: common.DfmToWei(new(big.Int).SetUint64(utxoMinValue))},
				{Address: primeBridgingFeeAddr, Amount: common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging))},
			},
			BridgingFee: common.DfmToWei(new(big.Int).SetUint64(100)),
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &oCore.BridgeClaims{}

		ethTx := &core.EthTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrNexus,
			Value:         common.DfmToWei(new(big.Int).SetUint64(utxoMinValue + minFeeForBridging + 100)),
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, ethTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim utxo value below minimum in receivers in metadata", func(t *testing.T) {
		utxoValueBelowMinInReceiversMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{Address: validTestAddress, Amount: common.DfmToWei(new(big.Int).SetUint64(utxoMinValue))},
				{Address: primeBridgingFeeAddr, Amount: common.DfmToWei(new(big.Int).SetUint64(2))},
			},
			BridgingFee: big.NewInt(0),
		})
		require.NoError(t, err)
		require.NotNil(t, utxoValueBelowMinInReceiversMetadata)

		claims := &oCore.BridgeClaims{}

		ethTx := &core.EthTx{
			Metadata:      utxoValueBelowMinInReceiversMetadata,
			OriginChainID: common.ChainIDStrNexus,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, ethTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found a utxo value below minimum value in metadata receivers")
	})

	t.Run("ValidateAndAddClaim invalid receiver addr in metadata 1", func(t *testing.T) {
		invalidAddrInReceiversMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{Address: primeBridgingFeeAddr, Amount: common.DfmToWei(new(big.Int).SetUint64(utxoMinValue))},
				{Address: nexusBridgingAddr, Amount: common.DfmToWei(new(big.Int).SetUint64(utxoMinValue))},
			},
			BridgingFee: big.NewInt(0),
		})
		require.NoError(t, err)
		require.NotNil(t, invalidAddrInReceiversMetadata)

		claims := &oCore.BridgeClaims{}

		ethTx := &core.EthTx{
			Metadata:      invalidAddrInReceiversMetadata,
			OriginChainID: common.ChainIDStrNexus,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, ethTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found an invalid receiver addr in metadata")
	})

	t.Run("ValidateAndAddClaim invalid receiver addr in metadata 2", func(t *testing.T) {
		invalidAddrInReceiversMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{Address: primeBridgingFeeAddr, Amount: common.DfmToWei(new(big.Int).SetUint64(utxoMinValue))},
				{Address: "stake_test1urrzuuwrq6lfq82y9u642qzcwvkljshn0743hs0rpd5wz8s2pe23d", Amount: common.DfmToWei(new(big.Int).SetUint64(utxoMinValue))},
			},
			BridgingFee: big.NewInt(0),
		})
		require.NoError(t, err)
		require.NotNil(t, invalidAddrInReceiversMetadata)

		claims := &oCore.BridgeClaims{}

		ethTx := &core.EthTx{
			Metadata:      invalidAddrInReceiversMetadata,
			OriginChainID: common.ChainIDStrNexus,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, ethTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found an invalid receiver addr in metadata")
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim receivers amounts and tx value missmatch less", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrPrime

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []core.BridgingRequestEthMetadataTransaction{
			{Address: primeBridgingFeeAddr, Amount: common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging))},
			{Address: validTestAddress, Amount: common.DfmToWei(new(big.Int).SetUint64(utxoMinValue))},
		}

		validMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: destinationChainID,
			SenderAddr:         "addr1",
			Transactions:       receivers,
			BridgingFee:        big.NewInt(0),
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &oCore.BridgeClaims{}

		ethTx := &core.EthTx{
			Hash:          txHash,
			Metadata:      validMetadata,
			OriginChainID: common.ChainIDStrNexus,
			Value:         common.DfmToWei(new(big.Int).SetUint64(utxoMinValue + minFeeForBridging - 1)),
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, ethTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)

		require.Error(t, err)
		require.ErrorContains(t, err, "tx value is not equal to sum of receiver amounts + fee")
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim receivers amounts and tx value missmatch more", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrPrime

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []core.BridgingRequestEthMetadataTransaction{
			{Address: primeBridgingFeeAddr, Amount: common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging))},
			{Address: validTestAddress, Amount: common.DfmToWei(new(big.Int).SetUint64(utxoMinValue))},
		}

		validMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: destinationChainID,
			SenderAddr:         "addr1",
			Transactions:       receivers,
			BridgingFee:        big.NewInt(0),
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &oCore.BridgeClaims{}

		ethTx := &core.EthTx{
			Hash:          txHash,
			Metadata:      validMetadata,
			OriginChainID: common.ChainIDStrNexus,
			Value:         common.DfmToWei(new(big.Int).SetUint64(utxoMinValue + minFeeForBridging + 1)),
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, ethTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)

		require.Error(t, err)
		require.ErrorContains(t, err, "tx value is not equal to sum of receiver amounts + fee")
	})

	t.Run("ValidateAndAddClaim fee in receivers less than minimum", func(t *testing.T) {
		feeInReceiversLessThanMinMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{Address: primeBridgingFeeAddr, Amount: common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging - 1))},
			},
			BridgingFee: big.NewInt(0),
		})
		require.NoError(t, err)
		require.NotNil(t, feeInReceiversLessThanMinMetadata)

		claims := &oCore.BridgeClaims{}

		ethTx := &core.EthTx{
			Metadata:      feeInReceiversLessThanMinMetadata,
			OriginChainID: common.ChainIDStrNexus,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, ethTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "bridging fee in metadata receivers is less than minimum")
	})

	t.Run("ValidateAndAddClaim more than allowed", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrPrime

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []core.BridgingRequestEthMetadataTransaction{
			{Address: primeBridgingFeeAddr, Amount: common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging))},
			{Address: validTestAddress, Amount: common.DfmToWei(new(big.Int).Add(new(big.Int).SetUint64(1), maxAmountAllowedToBridge))},
		}

		validMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: destinationChainID,
			SenderAddr:         "addr1",
			Transactions:       receivers,
			BridgingFee:        big.NewInt(0),
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		claims := &oCore.BridgeClaims{}

		ethTx := &core.EthTx{
			Hash:          txHash,
			Metadata:      validMetadata,
			OriginChainID: common.ChainIDStrNexus,
			Value:         common.DfmToWei(new(big.Int).SetUint64(maxAmountAllowedToBridge.Uint64() + 1 + minFeeForBridging)),
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, ethTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "sum of receiver amounts + fee")
		require.ErrorContains(t, err, "greater than maximum allowed")
	})

	t.Run("ValidateAndAddClaim valid", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrPrime

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []core.BridgingRequestEthMetadataTransaction{
			{Address: primeBridgingFeeAddr, Amount: common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging))},
			{Address: validTestAddress, Amount: common.DfmToWei(new(big.Int).SetUint64(utxoMinValue))},
		}

		validMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: destinationChainID,
			SenderAddr:         "addr1",
			Transactions:       receivers,
			BridgingFee:        big.NewInt(0),
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		ethTx := &core.EthTx{
			Hash:          txHash,
			Metadata:      validMetadata,
			OriginChainID: common.ChainIDStrNexus,
			Value:         common.DfmToWei(new(big.Int).SetUint64(utxoMinValue + minFeeForBridging)),
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessor(refundRequestProcessorMock, hclog.NewNullLogger())

		claims := &oCore.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		require.Len(t, claims.BridgingRequestClaims, 1)
		require.Equal(t, txHash, claims.BridgingRequestClaims[0].ObservedTransactionHash)
		require.Equal(t, destinationChainID, common.ToStrChainID(claims.BridgingRequestClaims[0].DestinationChainId))
		require.Len(t, claims.BridgingRequestClaims[0].Receivers, len(receivers))
		require.Equal(t, receivers[1].Address,
			claims.BridgingRequestClaims[0].Receivers[0].DestinationAddress)
		require.Equal(t, common.WeiToDfm(receivers[1].Amount), claims.BridgingRequestClaims[0].Receivers[0].Amount)
		require.Equal(t, receivers[0].Address,
			claims.BridgingRequestClaims[0].Receivers[1].DestinationAddress)
		require.Equal(t, feeAddrBridgingAmount, claims.BridgingRequestClaims[0].Receivers[1].Amount.Uint64())
	})
}
