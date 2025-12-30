package successtxprocessors

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	brAddrManager "github.com/Ethernal-Tech/apex-bridge/bridging_addresses_manager"
	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	oChain "github.com/Ethernal-Tech/apex-bridge/oracle_common/chain"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

var (
	protocolParameters = []byte(`{"costModels":{"PlutusV1":[197209,0,1,1,396231,621,0,1,150000,1000,0,1,150000,32,2477736,29175,4,29773,100,29773,100,29773,100,29773,100,29773,100,29773,100,100,100,29773,100,150000,32,150000,32,150000,32,150000,1000,0,1,150000,32,150000,1000,0,8,148000,425507,118,0,1,1,150000,1000,0,8,150000,112536,247,1,150000,10000,1,136542,1326,1,1000,150000,1000,1,150000,32,150000,32,150000,32,1,1,150000,1,150000,4,103599,248,1,103599,248,1,145276,1366,1,179690,497,1,150000,32,150000,32,150000,32,150000,32,150000,32,150000,32,148000,425507,118,0,1,1,61516,11218,0,1,150000,32,148000,425507,118,0,1,1,148000,425507,118,0,1,1,2477736,29175,4,0,82363,4,150000,5000,0,1,150000,32,197209,0,1,1,150000,32,150000,32,150000,32,150000,32,150000,32,150000,32,150000,32,3345831,1,1],"PlutusV2":[205665,812,1,1,1000,571,0,1,1000,24177,4,1,1000,32,117366,10475,4,23000,100,23000,100,23000,100,23000,100,23000,100,23000,100,100,100,23000,100,19537,32,175354,32,46417,4,221973,511,0,1,89141,32,497525,14068,4,2,196500,453240,220,0,1,1,1000,28662,4,2,245000,216773,62,1,1060367,12586,1,208512,421,1,187000,1000,52998,1,80436,32,43249,32,1000,32,80556,1,57667,4,1000,10,197145,156,1,197145,156,1,204924,473,1,208896,511,1,52467,32,64832,32,65493,32,22558,32,16563,32,76511,32,196500,453240,220,0,1,1,69522,11687,0,1,60091,32,196500,453240,220,0,1,1,196500,453240,220,0,1,1,1159724,392670,0,2,806990,30482,4,1927926,82523,4,265318,0,4,0,85931,32,205665,812,1,1,41182,32,212342,32,31220,32,32696,32,43357,32,32247,32,38314,32,35892428,10,9462713,1021,10,38887044,32947,10]},"protocolVersion":{"major":7,"minor":0},"maxBlockHeaderSize":1100,"maxBlockBodySize":65536,"maxTxSize":16384,"txFeeFixed":155381,"txFeePerByte":44,"stakeAddressDeposit":0,"stakePoolDeposit":0,"minPoolCost":0,"poolRetireMaxEpoch":18,"stakePoolTargetNum":100,"poolPledgeInfluence":0,"monetaryExpansion":0.1,"treasuryCut":0.1,"collateralPercentage":150,"executionUnitPrices":{"priceMemory":0.0577,"priceSteps":0.0000721},"utxoCostPerByte":4310,"maxTxExecutionUnits":{"memory":16000000,"steps":10000000000},"maxBlockExecutionUnits":{"memory":80000000,"steps":40000000000},"maxCollateralInputs":3,"maxValueSize":5000,"extraPraosEntropy":null,"decentralization":null,"minUTxOValue":null}`)
)

func TestBridgingRequestedProcessorSkyline(t *testing.T) {
	const (
		utxoMinValue               = 1000000
		minFeeForBridging          = 1000010
		minOperationFee            = 1000010
		minColCoinsAllowedToBridge = 100000
		feeAddrBridgingAmount      = uint64(1000005)
		feeAddrBridgingAmountEvm   = uint64(1000006)
		primeBridgingAddr          = "addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj660"
		primeBridgingFeeAddr       = "addr_test1vqqj5apwf5npsmudw0ranypkj9jw98t25wk4h83jy5mwypswekttt"
		validTestAddress           = "addr_test1vq6zkfat4rlmj2nd2sylpjjg5qhcg9mk92wykaw4m2dp2rqneafvl"

		nexusBridgingAddr   = "0xA4d1233A67776575425Ab185f6a9251aa00fEA25"
		polygonBridgingAddr = "0xA4d1233A67776575425Ab185f6a9251aa00fEA26"
		validEvmAddress     = "0xB4d1233A67776575425Ab185f6a9251aa00fEA27"
		evmZeroAddr         = common.EthZeroAddr

		policyID = "29f8873beb52e126f207a2dfd50f7cff556806b5b4cba9834a7b26a8"

		primeCurrencyID       = uint16(1)
		cardanoCurrencyID     = uint16(2)
		nexusCurrencyID       = uint16(3)
		polygonCurrencyID     = uint16(4)
		primeWrappedTokenID   = uint16(5)
		cardanoWrappedTokenID = uint16(6)
		polygonWrappedTokenID = uint16(7)
		usdtTokenID           = uint16(8)
		ccTokenID             = uint16(9)
	)

	maxAmountAllowedToBridge := new(big.Int).SetUint64(100000000)
	maxTokenAmountAllowedToBridge := new(big.Int).SetUint64(100000000)
	testChainID := "test"

	wrappedTokenPrime, err := wallet.NewTokenWithFullName(
		fmt.Sprintf("%s.%s",
			policyID,
			hex.EncodeToString([]byte("wrappedAda"))), true,
	)
	require.NoError(t, err)

	ccToken, err := wallet.NewTokenWithFullName(
		fmt.Sprintf("%s.%s",
			policyID,
			hex.EncodeToString([]byte("ccToken"))), true,
	)
	require.NoError(t, err)

	brAddrManagerMock := &brAddrManager.BridgingAddressesManagerMock{}
	brAddrManagerMock.On("GetAllPaymentAddresses", common.ChainIDIntPrime).Return([]string{primeBridgingAddr}, nil)
	brAddrManagerMock.On("GetFeeMultisigAddress", common.ChainIDIntPrime).Return(primeBridgingFeeAddr)

	getAppConfig := func(refundEnabled bool) *oCore.AppConfig {
		config := &oCore.AppConfig{
			BridgingAddressesManager: brAddrManagerMock,
			CardanoChains: map[string]*oCore.CardanoChainConfig{
				common.ChainIDStrPrime: {
					CardanoChainConfig: cardanotx.CardanoChainConfig{
						NetworkID:                wallet.TestNetNetwork,
						UtxoMinAmount:            utxoMinValue,
						DefaultMinFeeForBridging: minFeeForBridging,
						MinFeeForBridgingTokens:  minFeeForBridging,
						DestinationChains: map[string]common.TokenPairs{
							common.ChainIDStrCardano: []common.TokenPair{
								{SourceTokenID: primeCurrencyID, DestinationTokenID: cardanoWrappedTokenID, TrackSourceToken: true, TrackDestinationToken: true},
								{SourceTokenID: primeWrappedTokenID, DestinationTokenID: cardanoCurrencyID, TrackSourceToken: true, TrackDestinationToken: true},
							},
						},
						Tokens: map[uint16]common.Token{
							primeCurrencyID:     {ChainSpecific: wallet.AdaTokenName, LockUnlock: true},
							primeWrappedTokenID: {ChainSpecific: wrappedTokenPrime.String(), LockUnlock: true, IsWrappedCurrency: true},
							ccTokenID:           {ChainSpecific: ccToken.String(), LockUnlock: false, IsWrappedCurrency: false},
						},
					},
					FeeAddrBridgingAmount: feeAddrBridgingAmount,
				},
			},
			EthChains: map[string]*oCore.EthChainConfig{
				common.ChainIDStrNexus: {
					BridgingAddresses: oCore.EthBridgingAddresses{
						BridgingAddress: nexusBridgingAddr,
					},
					MinFeeForBridging: minFeeForBridging,
					MinOperationFee:   minOperationFee,
					DestinationChains: map[string]common.TokenPairs{
						common.ChainIDStrPrime: []common.TokenPair{
							{SourceTokenID: nexusCurrencyID, DestinationTokenID: primeWrappedTokenID, TrackSourceToken: true, TrackDestinationToken: true},
							{SourceTokenID: usdtTokenID, DestinationTokenID: primeCurrencyID, TrackSourceToken: true, TrackDestinationToken: true},
							{SourceTokenID: ccTokenID, DestinationTokenID: ccTokenID, TrackSourceToken: true, TrackDestinationToken: true},
						},
						common.ChainIDStrPolygon: []common.TokenPair{
							{SourceTokenID: nexusCurrencyID, DestinationTokenID: polygonWrappedTokenID, TrackSourceToken: true, TrackDestinationToken: true},
							{SourceTokenID: usdtTokenID, DestinationTokenID: usdtTokenID, TrackSourceToken: true, TrackDestinationToken: true},
						},
					},
					Tokens: map[uint16]common.Token{
						nexusCurrencyID: {ChainSpecific: wallet.AdaTokenName, LockUnlock: true},
						usdtTokenID:     {ChainSpecific: "0x11", LockUnlock: false, IsWrappedCurrency: false},
						ccTokenID:       {ChainSpecific: "0x22", LockUnlock: false, IsWrappedCurrency: false},
					},
					FeeAddrBridgingAmount: feeAddrBridgingAmountEvm,
				},
				common.ChainIDStrPolygon: {
					BridgingAddresses: oCore.EthBridgingAddresses{
						BridgingAddress: polygonBridgingAddr,
					},
					MinFeeForBridging: minFeeForBridging,
					MinOperationFee:   minOperationFee,
					DestinationChains: map[string]common.TokenPairs{
						common.ChainIDStrNexus: []common.TokenPair{
							{SourceTokenID: polygonWrappedTokenID, DestinationTokenID: nexusCurrencyID, TrackSourceToken: true, TrackDestinationToken: true},
						},
					},
					Tokens: map[uint16]common.Token{
						polygonCurrencyID:     {ChainSpecific: wallet.AdaTokenName, LockUnlock: true},
						polygonWrappedTokenID: {ChainSpecific: "0x33", LockUnlock: false, IsWrappedCurrency: true},
					},
					FeeAddrBridgingAmount: feeAddrBridgingAmountEvm,
				},
			},
			BridgingSettings: oCore.BridgingSettings{
				MaxReceiversPerBridgingRequest: 3,
				MaxAmountAllowedToBridge:       maxAmountAllowedToBridge,
				MaxTokenAmountAllowedToBridge:  maxTokenAmountAllowedToBridge,
				MinColCoinsAllowedToBridge:     minColCoinsAllowedToBridge,
			},
			RefundEnabled:    refundEnabled,
			ChainIDConverter: common.NewChainIDConverterForTest(),
		}
		config.FillOut()

		return config
	}

	getChainInfos := func() map[string]*oChain.CardanoChainInfo {
		appConfig := getAppConfig(true)
		chainInfos := make(map[string]*oChain.CardanoChainInfo, len(appConfig.CardanoChains))

		for _, cc := range appConfig.CardanoChains {
			info := oChain.NewCardanoChainInfo(cc)

			info.ProtocolParams = protocolParameters

			chainInfos[cc.ChainID] = info
		}

		return chainInfos
	}

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &oCore.BridgeClaims{}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, &core.EthTx{}, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

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

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		err := proc.ValidateAndAddClaim(claims, &core.EthTx{}, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim empty tx with refund err", func(t *testing.T) {
		claims := &oCore.BridgeClaims{}

		appConfig := getAppConfig(true)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, &core.EthTx{}, appConfig).Return(fmt.Errorf("test err"))

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

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

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

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

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.NoError(t, err)
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

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim transaction direction not allowed - invalid destination chain", func(t *testing.T) {
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
			OriginChainID: common.ChainIDStrNexus,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, ethTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "destination chain not registered")
	})

	t.Run("ValidateAndAddClaim destination chain not registered", func(t *testing.T) {
		destinationChainNonRegisteredMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: testChainID,
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

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "destination chain not registered")
	})

	t.Run("ValidateAndAddClaim - origin chain not registered", func(t *testing.T) {
		metadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions:       []core.BridgingRequestEthMetadataTransaction{},
			BridgingFee:        big.NewInt(0),
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &oCore.BridgeClaims{}

		ethTx := &core.EthTx{
			Metadata:      metadata,
			OriginChainID: testChainID,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, ethTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "origin chain not registered")
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim transaction direction not allowed", func(t *testing.T) {
		transactionDirectionNotSupportedMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{Address: primeBridgingAddr, Amount: big.NewInt(2), TokenID: primeCurrencyID},
			},
			BridgingFee: big.NewInt(0),
		})
		require.NoError(t, err)
		require.NotNil(t, transactionDirectionNotSupportedMetadata)

		claims := &oCore.BridgeClaims{}

		ethTx := &core.EthTx{
			Metadata:      transactionDirectionNotSupportedMetadata,
			OriginChainID: common.ChainIDStrNexus,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, ethTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "no bridging path from source chain")
		require.ErrorContains(t, err, "to destination chain")
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim transaction direction not allowed - evm receiver", func(t *testing.T) {
		transactionDirectionNotSupportedMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPolygon,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{Address: validEvmAddress, Amount: big.NewInt(2), TokenID: ccTokenID},
			},
			BridgingFee: big.NewInt(0),
		})
		require.NoError(t, err)
		require.NotNil(t, transactionDirectionNotSupportedMetadata)

		claims := &oCore.BridgeClaims{}

		ethTx := &core.EthTx{
			Metadata:      transactionDirectionNotSupportedMetadata,
			OriginChainID: common.ChainIDStrNexus,
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, ethTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "no bridging path from source chain")
		require.ErrorContains(t, err, "to destination chain")
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

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

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
				{
					Address: validTestAddress,
					Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue)),
					TokenID: nexusCurrencyID,
				},
			},
			BridgingFee:  common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging - 1)),
			OperationFee: common.DfmToWei(new(big.Int).SetUint64(minOperationFee)),
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

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "bridging fee in metadata is less than minimum")
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim fee receiver metadata invalid", func(t *testing.T) {
		metadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{
					Address: validTestAddress,
					Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue)),
					TokenID: nexusCurrencyID,
				},
				{
					Address: primeBridgingFeeAddr,
					Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue)),
					TokenID: usdtTokenID,
				},
			},
			BridgingFee:  common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging)),
			OperationFee: common.DfmToWei(new(big.Int).SetUint64(minOperationFee)),
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

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "fee receiver metadata invalid")
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim fee receiver metadata invalid - evm receiver", func(t *testing.T) {
		metadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPolygon,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{
					Address: validEvmAddress,
					Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue)),
					TokenID: usdtTokenID,
				},
				{
					Address: evmZeroAddr,
					Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue)),
					TokenID: usdtTokenID,
				},
			},
			BridgingFee:  common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging)),
			OperationFee: common.DfmToWei(new(big.Int).SetUint64(minOperationFee)),
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

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "fee receiver metadata invalid")
	})

	t.Run("ValidateAndAddClaim fee amount is specified in receivers", func(t *testing.T) {
		metadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{
					Address: validTestAddress,
					Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue)),
					TokenID: nexusCurrencyID,
				},
				{
					Address: primeBridgingFeeAddr,
					Amount:  common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging)),
					TokenID: nexusCurrencyID,
				},
			},
			BridgingFee:  common.DfmToWei(new(big.Int).SetUint64(100)),
			OperationFee: common.DfmToWei(new(big.Int).SetUint64(minOperationFee)),
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &oCore.BridgeClaims{}

		ethTx := &core.EthTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrNexus,
			Value:         common.DfmToWei(new(big.Int).SetUint64(utxoMinValue + minFeeForBridging + 100 + minOperationFee)),
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorError", claims, ethTx, appConfig).Return(nil)
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.NoError(t, err)
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim utxo value below minimum in receivers in metadata", func(t *testing.T) {
		utxoValueBelowMinInReceiversMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{
					Address: validTestAddress,
					Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue - 1)),
					TokenID: usdtTokenID,
				},
				{
					Address: primeBridgingFeeAddr,
					Amount:  common.DfmToWei(new(big.Int).SetUint64(2)),
					TokenID: nexusCurrencyID,
				},
			},
			BridgingFee:  common.DfmToWei(new(big.Int).SetUint64(100)),
			OperationFee: common.DfmToWei(new(big.Int).SetUint64(minOperationFee)),
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

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found a utxo value below minimum value in metadata receivers")
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim token amount below minimum allowed", func(t *testing.T) {
		utxoValueBelowMinInReceiversMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{
					Address: validTestAddress,
					Amount:  common.DfmToWei(new(big.Int).SetUint64(minColCoinsAllowedToBridge - 1)),
					TokenID: ccTokenID,
				},
				{
					Address: primeBridgingFeeAddr,
					Amount:  common.DfmToWei(new(big.Int).SetUint64(2)),
					TokenID: nexusCurrencyID,
				},
			},
			BridgingFee:  common.DfmToWei(new(big.Int).SetUint64(100)),
			OperationFee: common.DfmToWei(new(big.Int).SetUint64(minOperationFee)),
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

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "token amount below minimum allowed")
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim token amount below minimum allowed - evm receiver", func(t *testing.T) {
		utxoValueBelowMinInReceiversMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPolygon,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{
					Address: validEvmAddress,
					Amount:  common.DfmToWei(new(big.Int).SetUint64(minColCoinsAllowedToBridge - 1)),
					TokenID: usdtTokenID,
				},
				{
					Address: evmZeroAddr,
					Amount:  common.DfmToWei(new(big.Int).SetUint64(2)),
					TokenID: nexusCurrencyID,
				},
			},
			BridgingFee:  common.DfmToWei(new(big.Int).SetUint64(100)),
			OperationFee: common.DfmToWei(new(big.Int).SetUint64(minOperationFee)),
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

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "token amount below minimum allowed")
	})

	t.Run("ValidateAndAddClaim amount of tokens for receivers too high", func(t *testing.T) {
		utxoValueBelowMinInReceiversMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{
					Address: validTestAddress,
					Amount:  common.DfmToWei(new(big.Int).Add(maxTokenAmountAllowedToBridge, big.NewInt(1))),
					TokenID: ccTokenID,
				},
				{
					Address: primeBridgingFeeAddr,
					Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue)),
					TokenID: nexusCurrencyID,
				},
			},
			BridgingFee:  common.DfmToWei(new(big.Int).SetUint64(100)),
			OperationFee: common.DfmToWei(new(big.Int).SetUint64(minOperationFee)),
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

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "amount of tokens for receivers too high")
	})

	t.Run("ValidateAndAddClaim operation fee in metadata is less than minimum", func(t *testing.T) {
		utxoValueBelowMinInReceiversMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{
					Address: validTestAddress,
					Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue)),
					TokenID: ccTokenID,
				},
				{
					Address: primeBridgingFeeAddr,
					Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue)),
					TokenID: nexusCurrencyID,
				},
			},
			BridgingFee:  common.DfmToWei(new(big.Int).SetUint64(100)),
			OperationFee: common.DfmToWei(new(big.Int).SetUint64(minOperationFee - 1)),
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

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "operation fee in metadata is less than minimum")
	})

	t.Run("ValidateAndAddClaim invalid receiver addr in metadata 1", func(t *testing.T) {
		invalidAddrInReceiversMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{
					Address: primeBridgingFeeAddr,
					Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue)),
					TokenID: nexusCurrencyID,
				},
				{
					Address: nexusBridgingAddr,
					Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue)),
					TokenID: nexusCurrencyID,
				},
			},
			BridgingFee:  common.DfmToWei(new(big.Int).SetUint64(100)),
			OperationFee: common.DfmToWei(new(big.Int).SetUint64(minOperationFee)),
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

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

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
				{
					Address: primeBridgingFeeAddr,
					Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue)),
					TokenID: nexusCurrencyID,
				},
				{
					Address: "stake_test1urrzuuwrq6lfq82y9u642qzcwvkljshn0743hs0rpd5wz8s2pe23d",
					Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue)),
					TokenID: nexusCurrencyID,
				},
			},
			BridgingFee:  common.DfmToWei(new(big.Int).SetUint64(100)),
			OperationFee: common.DfmToWei(new(big.Int).SetUint64(minOperationFee)),
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

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found an invalid receiver addr in metadata")
	})

	t.Run("ValidateAndAddClaim invalid receiver addr in metadata - evm receiver", func(t *testing.T) {
		invalidAddrInReceiversMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrPolygon,
			SenderAddr:         "addr1",
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{
					Address: evmZeroAddr,
					Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue)),
					TokenID: nexusCurrencyID,
				},
				{
					Address: validEvmAddress[:len(validEvmAddress)-1], // invalid address
					Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue)),
					TokenID: usdtTokenID,
				},
			},
			BridgingFee:  common.DfmToWei(new(big.Int).SetUint64(100)),
			OperationFee: common.DfmToWei(new(big.Int).SetUint64(minOperationFee)),
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

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "found an invalid eth receiver addr in metadata")
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim receivers amounts and tx value missmatch less", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrPrime

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []core.BridgingRequestEthMetadataTransaction{
			{
				Address: primeBridgingFeeAddr,
				Amount:  common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging)),
				TokenID: nexusCurrencyID,
			},
			{
				Address: validTestAddress,
				Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue)),
				TokenID: nexusCurrencyID,
			},
		}

		validMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: destinationChainID,
			SenderAddr:         "addr1",
			Transactions:       receivers,
			BridgingFee:        common.DfmToWei(new(big.Int).SetUint64(100)),
			OperationFee:       common.DfmToWei(new(big.Int).SetUint64(minOperationFee)),
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

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)

		require.Error(t, err)
		require.ErrorContains(t, err, "tx value is not equal to sum of receiver amounts + fee")
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim receivers amounts and tx value missmatch more", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrPrime

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []core.BridgingRequestEthMetadataTransaction{
			{
				Address: primeBridgingFeeAddr,
				Amount:  common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging)),
				TokenID: nexusCurrencyID,
			},
			{
				Address: validTestAddress,
				Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue)),
				TokenID: nexusCurrencyID,
			},
		}

		validMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: destinationChainID,
			SenderAddr:         "addr1",
			Transactions:       receivers,
			BridgingFee:        common.DfmToWei(new(big.Int).SetUint64(100)),
			OperationFee:       common.DfmToWei(new(big.Int).SetUint64(minOperationFee)),
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

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

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
				{
					Address: primeBridgingFeeAddr,
					Amount:  common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging - 1)),
					TokenID: nexusCurrencyID,
				},
			},
			BridgingFee:  big.NewInt(0),
			OperationFee: common.DfmToWei(new(big.Int).SetUint64(minOperationFee)),
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

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "bridging fee in metadata is less than minimum")
	})

	t.Run("ValidateAndAddClaim more than allowed", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrPrime

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []core.BridgingRequestEthMetadataTransaction{
			{
				Address: primeBridgingFeeAddr,
				Amount:  common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging)),
				TokenID: nexusCurrencyID,
			},
			{
				Address: validTestAddress,
				Amount:  common.DfmToWei(new(big.Int).Add(new(big.Int).SetUint64(1), maxAmountAllowedToBridge)),
				TokenID: nexusCurrencyID,
			},
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

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "sum of receiver amounts")
		require.ErrorContains(t, err, "greater than maximum allowed")
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim valid", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrPrime

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []core.BridgingRequestEthMetadataTransaction{
			{
				Address: primeBridgingFeeAddr,
				Amount:  common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging)),
				TokenID: nexusCurrencyID,
			},
			{
				Address: validTestAddress,
				Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue)),
				TokenID: nexusCurrencyID,
			},
		}

		validMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: destinationChainID,
			SenderAddr:         "addr1",
			Transactions:       receivers,
			BridgingFee:        common.DfmToWei(new(big.Int).SetUint64(100)),
			OperationFee:       common.DfmToWei(new(big.Int).SetUint64(minOperationFee)),
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		ethTx := &core.EthTx{
			Hash:          txHash,
			Metadata:      validMetadata,
			OriginChainID: common.ChainIDStrNexus,
			Value:         common.DfmToWei(new(big.Int).SetUint64(utxoMinValue + minFeeForBridging + 100 + minOperationFee)),
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		claims := &oCore.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		require.Len(t, claims.BridgingRequestClaims, 1)
		require.Equal(t, txHash, claims.BridgingRequestClaims[0].ObservedTransactionHash)
		require.Equal(t, destinationChainID, appConfig.ChainIDConverter.ToStrChainID(claims.BridgingRequestClaims[0].DestinationChainId))
		require.Len(t, claims.BridgingRequestClaims[0].Receivers, len(receivers))
		require.Equal(t, receivers[1].Address,
			claims.BridgingRequestClaims[0].Receivers[0].DestinationAddress)
		require.Equal(t, common.WeiToDfm(receivers[1].Amount), claims.BridgingRequestClaims[0].Receivers[0].AmountWrapped)
		require.Equal(t, receivers[0].Address,
			claims.BridgingRequestClaims[0].Receivers[1].DestinationAddress)
		require.Equal(t, feeAddrBridgingAmount, claims.BridgingRequestClaims[0].Receivers[1].Amount.Uint64())
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim valid - currency on destination", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrPrime

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []core.BridgingRequestEthMetadataTransaction{
			{
				Address: primeBridgingFeeAddr,
				Amount:  common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging)),
				TokenID: nexusCurrencyID,
			},
			{
				Address: validTestAddress,
				Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue)),
				TokenID: usdtTokenID,
			},
		}

		validMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: destinationChainID,
			SenderAddr:         "addr1",
			Transactions:       receivers,
			BridgingFee:        common.DfmToWei(new(big.Int).SetUint64(100)),
			OperationFee:       common.DfmToWei(new(big.Int).SetUint64(minOperationFee)),
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		ethTx := &core.EthTx{
			Hash:          txHash,
			Metadata:      validMetadata,
			OriginChainID: common.ChainIDStrNexus,
			Value:         common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging + 100 + minOperationFee)),
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		claims := &oCore.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		require.Len(t, claims.BridgingRequestClaims, 1)
		require.Equal(t, txHash, claims.BridgingRequestClaims[0].ObservedTransactionHash)
		require.Equal(t, destinationChainID, appConfig.ChainIDConverter.ToStrChainID(claims.BridgingRequestClaims[0].DestinationChainId))
		require.Len(t, claims.BridgingRequestClaims[0].Receivers, len(receivers))
		require.Equal(t, receivers[1].Address,
			claims.BridgingRequestClaims[0].Receivers[0].DestinationAddress)
		require.Equal(t, common.WeiToDfm(receivers[1].Amount), claims.BridgingRequestClaims[0].Receivers[0].Amount)
		require.Equal(t, receivers[0].Address,
			claims.BridgingRequestClaims[0].Receivers[1].DestinationAddress)
		require.Equal(t, feeAddrBridgingAmount, claims.BridgingRequestClaims[0].Receivers[1].Amount.Uint64())
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim valid - evm receiver", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrPolygon

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []core.BridgingRequestEthMetadataTransaction{
			{
				Address: evmZeroAddr,
				Amount:  common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging)),
				TokenID: nexusCurrencyID,
			},
			{
				Address: validEvmAddress,
				Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue)),
				TokenID: nexusCurrencyID,
			},
		}

		validMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: destinationChainID,
			SenderAddr:         "addr1",
			Transactions:       receivers,
			BridgingFee:        common.DfmToWei(new(big.Int).SetUint64(100)),
			OperationFee:       common.DfmToWei(new(big.Int).SetUint64(minOperationFee)),
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		ethTx := &core.EthTx{
			Hash:          txHash,
			Metadata:      validMetadata,
			OriginChainID: common.ChainIDStrNexus,
			Value:         common.DfmToWei(new(big.Int).SetUint64(utxoMinValue + minFeeForBridging + 100 + minOperationFee)),
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		claims := &oCore.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		require.Len(t, claims.BridgingRequestClaims, 1)
		require.Equal(t, txHash, claims.BridgingRequestClaims[0].ObservedTransactionHash)
		require.Equal(t, destinationChainID, appConfig.ChainIDConverter.ToStrChainID(claims.BridgingRequestClaims[0].DestinationChainId))
		require.Len(t, claims.BridgingRequestClaims[0].Receivers, len(receivers))
		require.Equal(t, receivers[1].Address,
			claims.BridgingRequestClaims[0].Receivers[0].DestinationAddress)
		require.Equal(t, common.WeiToDfm(receivers[1].Amount), claims.BridgingRequestClaims[0].Receivers[0].AmountWrapped)
		require.Equal(t, receivers[0].Address,
			claims.BridgingRequestClaims[0].Receivers[1].DestinationAddress)
		require.Equal(t, feeAddrBridgingAmountEvm, claims.BridgingRequestClaims[0].Receivers[1].Amount.Uint64())
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim valid - evm receiver and usdt", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrPolygon

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []core.BridgingRequestEthMetadataTransaction{
			{
				Address: evmZeroAddr,
				Amount:  common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging)),
				TokenID: nexusCurrencyID,
			},
			{
				Address: validEvmAddress,
				Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue)),
				TokenID: usdtTokenID,
			},
		}

		validMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: destinationChainID,
			SenderAddr:         "addr1",
			Transactions:       receivers,
			BridgingFee:        common.DfmToWei(new(big.Int).SetUint64(100)),
			OperationFee:       common.DfmToWei(new(big.Int).SetUint64(minOperationFee)),
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		ethTx := &core.EthTx{
			Hash:          txHash,
			Metadata:      validMetadata,
			OriginChainID: common.ChainIDStrNexus,
			Value:         common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging + 100 + minOperationFee)),
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		claims := &oCore.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		require.Len(t, claims.BridgingRequestClaims, 1)
		require.Equal(t, txHash, claims.BridgingRequestClaims[0].ObservedTransactionHash)
		require.Equal(t, destinationChainID, appConfig.ChainIDConverter.ToStrChainID(claims.BridgingRequestClaims[0].DestinationChainId))
		require.Len(t, claims.BridgingRequestClaims[0].Receivers, len(receivers))
		require.Equal(t, receivers[1].Address,
			claims.BridgingRequestClaims[0].Receivers[0].DestinationAddress)
		require.Equal(t, common.WeiToDfm(receivers[1].Amount), claims.BridgingRequestClaims[0].Receivers[0].AmountWrapped)
		require.Equal(t, receivers[0].Address,
			claims.BridgingRequestClaims[0].Receivers[1].DestinationAddress)
		require.Equal(t, feeAddrBridgingAmountEvm, claims.BridgingRequestClaims[0].Receivers[1].Amount.Uint64())
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim valid - evm receiver and currency on destination", func(t *testing.T) {
		const destinationChainID = common.ChainIDStrNexus

		txHash := [32]byte(common.NewHashFromHexString("0x2244FF"))
		receivers := []core.BridgingRequestEthMetadataTransaction{
			{
				Address: evmZeroAddr,
				Amount:  common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging)),
				TokenID: polygonCurrencyID,
			},
			{
				Address: validEvmAddress,
				Amount:  common.DfmToWei(new(big.Int).SetUint64(utxoMinValue)),
				TokenID: polygonWrappedTokenID,
			},
		}

		validMetadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: destinationChainID,
			SenderAddr:         "addr1",
			Transactions:       receivers,
			BridgingFee:        common.DfmToWei(new(big.Int).SetUint64(100)),
			OperationFee:       common.DfmToWei(new(big.Int).SetUint64(minOperationFee)),
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		ethTx := &core.EthTx{
			Hash:          txHash,
			Metadata:      validMetadata,
			OriginChainID: common.ChainIDStrPolygon,
			Value:         common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging + 100 + minOperationFee)),
		}

		appConfig := getAppConfig(false)
		refundRequestProcessorMock := &core.EthTxSuccessRefundProcessorMock{}
		refundRequestProcessorMock.On(
			"HandleBridgingProcessorPreValidate", ethTx, appConfig).Return(nil)

		proc := NewEthBridgingRequestedProcessorSkyline(refundRequestProcessorMock, hclog.NewNullLogger(), getChainInfos())

		claims := &oCore.BridgeClaims{}
		err = proc.ValidateAndAddClaim(claims, ethTx, appConfig)
		require.NoError(t, err)
		require.True(t, claims.Count() == 1)
		require.Len(t, claims.BridgingRequestClaims, 1)
		require.Equal(t, txHash, claims.BridgingRequestClaims[0].ObservedTransactionHash)
		require.Equal(t, destinationChainID, appConfig.ChainIDConverter.ToStrChainID(claims.BridgingRequestClaims[0].DestinationChainId))
		require.Len(t, claims.BridgingRequestClaims[0].Receivers, len(receivers))
		require.Equal(t, receivers[1].Address,
			claims.BridgingRequestClaims[0].Receivers[0].DestinationAddress)
		require.Equal(t, common.WeiToDfm(receivers[1].Amount), claims.BridgingRequestClaims[0].Receivers[0].Amount)
		require.Equal(t, receivers[0].Address,
			claims.BridgingRequestClaims[0].Receivers[1].DestinationAddress)
		require.Equal(t, feeAddrBridgingAmountEvm, claims.BridgingRequestClaims[0].Receivers[1].Amount.Uint64())
	})
}
