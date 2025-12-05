package successtxprocessors

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestRefundRequestedProcessorSkyline(t *testing.T) {
	const (
		utxoMinValue         = 1000000
		minFeeForBridging    = 1000010
		primeBridgingAddr    = "addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj660"
		primeBridgingFeeAddr = "addr_test1vqqj5apwf5npsmudw0ranypkj9jw98t25wk4h83jy5mwypswekttt"
		nexusBridgingAddr    = "0xA4d1233A67776575425Ab185f6a9251aa00fEA25"
		validTestAddress     = "addr_test1vq6zkfat4rlmj2nd2sylpjjg5qhcg9mk92wykaw4m2dp2rqneafvl"

		primeCurrencyID     = uint16(1)
		nexusCurrencyID     = uint16(3)
		primeWrappedTokenID = uint16(4)
		nexusWrappedTokenID = uint16(5)
		usdtTokenID         = uint16(6)

		policyID = "29f8873beb52e126f207a2dfd50f7cff556806b5b4cba9834a7b26a8"
	)

	maxAmountAllowedToBridge := new(big.Int).SetUint64(100000000)

	wrappedTokenPrime, err := wallet.NewTokenWithFullName(
		fmt.Sprintf("%s.%s",
			policyID,
			hex.EncodeToString([]byte("wrappedAda"))), true,
	)
	require.NoError(t, err)

	getAppConfig := func(refundEnabled bool) *oCore.AppConfig {
		appConfig := &oCore.AppConfig{
			CardanoChains: map[string]*oCore.CardanoChainConfig{
				common.ChainIDStrPrime: {
					CardanoChainConfig: cardanotx.CardanoChainConfig{
						NetworkID:                wallet.TestNetNetwork,
						UtxoMinAmount:            utxoMinValue,
						DefaultMinFeeForBridging: minFeeForBridging,
						MinFeeForBridgingTokens:  minFeeForBridging,
						Tokens: map[uint16]common.Token{
							primeCurrencyID:     {ChainSpecific: wallet.AdaTokenName, LockUnlock: true},
							primeWrappedTokenID: {ChainSpecific: wrappedTokenPrime.String(), LockUnlock: true, IsWrappedCurrency: true},
						},
					},
				},
			},
			EthChains: map[string]*oCore.EthChainConfig{
				common.ChainIDStrNexus: {
					BridgingAddresses: oCore.EthBridgingAddresses{
						BridgingAddress: nexusBridgingAddr,
					},
					MinFeeForBridging: minFeeForBridging,
					DestinationChain: map[string]common.TokenPairs{
						common.ChainIDStrPrime: []common.TokenPair{
							{SourceTokenID: nexusCurrencyID, DestinationTokenID: primeWrappedTokenID, TrackSourceToken: true, TrackDestinationToken: true},
							{SourceTokenID: nexusWrappedTokenID, DestinationTokenID: primeCurrencyID, TrackSourceToken: true, TrackDestinationToken: true},
							{SourceTokenID: usdtTokenID, DestinationTokenID: primeCurrencyID, TrackSourceToken: true, TrackDestinationToken: true},
						},
					},
					Tokens: map[uint16]common.Token{
						nexusCurrencyID:     {ChainSpecific: wallet.AdaTokenName, LockUnlock: true},
						nexusWrappedTokenID: {ChainSpecific: "0x11", LockUnlock: false, IsWrappedCurrency: true},
						usdtTokenID:         {ChainSpecific: "0x12", LockUnlock: false, IsWrappedCurrency: false},
					},
				},
			},
			BridgingSettings: oCore.BridgingSettings{
				MaxReceiversPerBridgingRequest: 3,
				MaxAmountAllowedToBridge:       maxAmountAllowedToBridge,
			},
			RefundEnabled: refundEnabled,
			TryCountLimits: oCore.TryCountLimits{
				MaxRefundTryCount: 3,
			},
		}

		appConfig.FillOut()

		return appConfig
	}

	proc := NewRefundRequestProcessorSkyline(hclog.NewNullLogger())
	disabledProc := NewRefundDisabledProcessor()

	t.Run("Refund disabled - HandleBridgingProcessorPreValidate", func(t *testing.T) {
		appConfig := getAppConfig(false)

		err := disabledProc.HandleBridgingProcessorPreValidate(&core.EthTx{}, appConfig)
		require.NoError(t, err)
	})

	t.Run("Refund disabled - HandleBridgingProcessorError", func(t *testing.T) {
		appConfig := getAppConfig(false)

		err := disabledProc.HandleBridgingProcessorError(
			&oCore.BridgeClaims{}, &core.EthTx{}, appConfig, fmt.Errorf("test err"), "")
		require.ErrorContains(t, err, "test err")
	})

	t.Run("Refund disabled - ValidateAndAddClaim", func(t *testing.T) {
		appConfig := getAppConfig(false)

		err := disabledProc.ValidateAndAddClaim(&oCore.BridgeClaims{}, &core.EthTx{}, appConfig)
		require.ErrorContains(t, err, "refund is not enabled")
	})

	t.Run("HandleBridgingProcessorPreValidate - empty tx", func(t *testing.T) {
		appConfig := getAppConfig(false)

		err := proc.HandleBridgingProcessorPreValidate(&core.EthTx{}, appConfig)
		require.NoError(t, err)
	})

	t.Run("HandleBridgingProcessorPreValidate - batchTryCount over", func(t *testing.T) {
		appConfig := getAppConfig(false)

		err := proc.HandleBridgingProcessorPreValidate(&core.EthTx{BatchTryCount: 1}, appConfig)
		require.ErrorContains(t, err, "try count exceeded")
	})

	t.Run("HandleBridgingProcessorPreValidate - submitTryCount over", func(t *testing.T) {
		appConfig := getAppConfig(false)

		err := proc.HandleBridgingProcessorPreValidate(&core.EthTx{SubmitTryCount: 1}, appConfig)
		require.ErrorContains(t, err, "try count exceeded")
	})

	t.Run("HandleBridgingProcessorError - empty ty", func(t *testing.T) {
		appConfig := getAppConfig(false)

		err := proc.HandleBridgingProcessorError(
			&oCore.BridgeClaims{}, &core.EthTx{}, appConfig, nil, "")
		require.ErrorContains(t, err, "unexpected end of JSON input")
	})

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &oCore.BridgeClaims{}

		appConfig := getAppConfig(true)

		err := proc.ValidateAndAddClaim(claims, &core.EthTx{}, appConfig)
		require.ErrorContains(t, err, "failed to unmarshal metadata")
	})

	t.Run("ValidateAndAddClaim origin chain not registered", func(t *testing.T) {
		metadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.TxTypeRefundRequest,
			DestinationChainID: "invalid",
			SenderAddr:         "addr1",
			Transactions:       []core.BridgingRequestEthMetadataTransaction{},
			BridgingFee:        big.NewInt(0),
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &oCore.BridgeClaims{}

		appConfig := getAppConfig(true)

		err = proc.ValidateAndAddClaim(claims, &core.EthTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.ErrorContains(t, err, "unsupported chain id found in tx")
	})

	t.Run("ValidateAndAddClaim invalid sender address", func(t *testing.T) {
		metadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.TxTypeRefundRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "invalid_address",
			Transactions:       []core.BridgingRequestEthMetadataTransaction{},
			BridgingFee:        common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging - 1)),
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &oCore.BridgeClaims{}

		appConfig := getAppConfig(true)

		err = proc.ValidateAndAddClaim(claims, &core.EthTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrNexus,
		}, appConfig)
		require.ErrorContains(t, err, "invalid sender addr")
	})

	t.Run("ValidateAndAddClaim bridging amount is too low", func(t *testing.T) {
		metadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.TxTypeRefundRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         nexusBridgingAddr,
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{Address: validTestAddress, Amount: new(big.Int).SetUint64(utxoMinValue - 1)},
			},
			BridgingFee: common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging - 1)),
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &oCore.BridgeClaims{}

		appConfig := getAppConfig(true)

		err = proc.ValidateAndAddClaim(claims, &core.EthTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrNexus,
			Value:         new(big.Int).SetUint64(utxoMinValue - 1),
		}, appConfig)
		require.ErrorContains(t, err, "less than the minimum required for refund")
	})

	t.Run("ValidateAndAddClaim unsupported destination chain id found in metadata", func(t *testing.T) {
		metadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.TxTypeRefundRequest,
			DestinationChainID: "",
			SenderAddr:         nexusBridgingAddr,
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
			BridgingFee: common.DfmToWei(new(big.Int).SetUint64(100)),
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &oCore.BridgeClaims{}

		appConfig := getAppConfig(true)

		err = proc.ValidateAndAddClaim(claims, &core.EthTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrNexus,
			Value:         common.DfmToWei(new(big.Int).SetUint64(utxoMinValue + minFeeForBridging + 100)),
		}, appConfig)
		require.ErrorContains(t, err, "unsupported destination chain id found in metadata")
	})

	//nolint:dupl
	t.Run("ValidateAndAddClaim try count exceeded", func(t *testing.T) {
		metadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.TxTypeRefundRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         nexusBridgingAddr,
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{Address: validTestAddress, Amount: common.DfmToWei(new(big.Int).SetUint64(utxoMinValue))},
				{Address: primeBridgingFeeAddr, Amount: common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging))},
			},
			BridgingFee: common.DfmToWei(new(big.Int).SetUint64(100)),
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &oCore.BridgeClaims{}

		appConfig := getAppConfig(true)

		err = proc.ValidateAndAddClaim(claims, &core.EthTx{
			Metadata:       metadata,
			OriginChainID:  common.ChainIDStrNexus,
			Value:          common.DfmToWei(new(big.Int).SetUint64(utxoMinValue + minFeeForBridging + 100)),
			RefundTryCount: 5,
		}, appConfig)
		require.ErrorContains(t, err, "try count exceeded")
	})

	t.Run("ValidateAndAddClaim token not registered", func(t *testing.T) {
		metadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.TxTypeRefundRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         nexusBridgingAddr,
			Transactions: []core.BridgingRequestEthMetadataTransaction{
				{Address: validTestAddress, Amount: common.DfmToWei(new(big.Int).SetUint64(utxoMinValue))},
				{Address: primeBridgingFeeAddr, Amount: common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging))},
			},
			BridgingFee: common.DfmToWei(new(big.Int).SetUint64(100)),
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &oCore.BridgeClaims{}

		appConfig := getAppConfig(true)

		err = proc.ValidateAndAddClaim(claims, &core.EthTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrNexus,
			Value:         common.DfmToWei(new(big.Int).SetUint64(utxoMinValue + minFeeForBridging + 100)),
		}, appConfig)
		require.ErrorContains(t, err, "token with ID 0 is not registered in chain")
	})

	t.Run("ValidateAndAddClaim valid", func(t *testing.T) {
		metadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.TxTypeRefundRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         nexusBridgingAddr,
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
			BridgingFee: common.DfmToWei(new(big.Int).SetUint64(100)),
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &oCore.BridgeClaims{}

		appConfig := getAppConfig(true)

		txValue := new(big.Int).SetUint64(utxoMinValue + minFeeForBridging + 100)

		err = proc.ValidateAndAddClaim(claims, &core.EthTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrNexus,
			Value:         common.DfmToWei(txValue),
		}, appConfig)
		require.NoError(t, err)

		require.True(t, claims.Count() == 1)
		require.Len(t, claims.RefundRequestClaims, 1)
		require.Equal(t, common.ChainIDStrPrime, common.ToStrChainID(claims.RefundRequestClaims[0].DestinationChainId))
		require.Equal(t, nexusBridgingAddr, claims.RefundRequestClaims[0].OriginSenderAddress)
		require.Equal(t, big.NewInt(minFeeForBridging+utxoMinValue), claims.RefundRequestClaims[0].OriginAmount)
		require.Equal(t, big.NewInt(0), claims.RefundRequestClaims[0].OriginWrappedAmount)
		require.Len(t, claims.RefundRequestClaims[0].TokenAmounts, 1)
		require.Equal(t, nexusCurrencyID, claims.RefundRequestClaims[0].TokenAmounts[0].TokenId)
		require.Equal(t, txValue, claims.RefundRequestClaims[0].TokenAmounts[0].AmountCurrency)
		require.Equal(t, big.NewInt(0), claims.RefundRequestClaims[0].TokenAmounts[0].AmountTokens)
	})

	t.Run("ValidateAndAddClaim valid - wrapped on source", func(t *testing.T) {
		amountWrapped := new(big.Int).SetUint64(utxoMinValue)

		receivers := []core.BridgingRequestEthMetadataTransaction{
			{
				Address: validTestAddress,
				Amount:  common.DfmToWei(amountWrapped),
				TokenID: nexusWrappedTokenID,
			},
			{
				Address: primeBridgingFeeAddr,
				Amount:  common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging)),
				TokenID: nexusCurrencyID,
			},
		}
		metadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.TxTypeRefundRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         nexusBridgingAddr,
			Transactions:       receivers,
			BridgingFee:        common.DfmToWei(new(big.Int).SetUint64(100)),
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &oCore.BridgeClaims{}

		appConfig := getAppConfig(true)

		txValue := new(big.Int).SetUint64(minFeeForBridging + 100)

		err = proc.ValidateAndAddClaim(claims, &core.EthTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrNexus,
			Value:         common.DfmToWei(txValue),
		}, appConfig)
		require.NoError(t, err)

		require.True(t, claims.Count() == 1)
		require.Len(t, claims.RefundRequestClaims, 1)
		require.Equal(t, common.ChainIDStrPrime, common.ToStrChainID(claims.RefundRequestClaims[0].DestinationChainId))
		require.Equal(t, nexusBridgingAddr, claims.RefundRequestClaims[0].OriginSenderAddress)
		require.Equal(t, big.NewInt(minFeeForBridging), claims.RefundRequestClaims[0].OriginAmount)
		require.Equal(t, big.NewInt(utxoMinValue), claims.RefundRequestClaims[0].OriginWrappedAmount)
		require.Len(t, claims.RefundRequestClaims[0].TokenAmounts, 1)
		require.Equal(t, nexusWrappedTokenID, claims.RefundRequestClaims[0].TokenAmounts[0].TokenId)
		require.Equal(t, txValue, claims.RefundRequestClaims[0].TokenAmounts[0].AmountCurrency)
		require.Equal(t, amountWrapped, claims.RefundRequestClaims[0].TokenAmounts[0].AmountTokens)
	})

	t.Run("ValidateAndAddClaim valid - non-wrapped token on source", func(t *testing.T) {
		amountToken := new(big.Int).SetUint64(utxoMinValue)

		receivers := []core.BridgingRequestEthMetadataTransaction{
			{
				Address: validTestAddress,
				Amount:  common.DfmToWei(amountToken),
				TokenID: usdtTokenID,
			},
			{
				Address: primeBridgingFeeAddr,
				Amount:  common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging)),
				TokenID: nexusCurrencyID,
			},
		}
		metadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.TxTypeRefundRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         nexusBridgingAddr,
			Transactions:       receivers,
			BridgingFee:        common.DfmToWei(new(big.Int).SetUint64(100)),
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &oCore.BridgeClaims{}

		appConfig := getAppConfig(true)

		txValue := new(big.Int).SetUint64(minFeeForBridging + 100)

		err = proc.ValidateAndAddClaim(claims, &core.EthTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrNexus,
			Value:         common.DfmToWei(txValue),
		}, appConfig)
		require.NoError(t, err)

		require.True(t, claims.Count() == 1)
		require.Len(t, claims.RefundRequestClaims, 1)
		require.Equal(t, common.ChainIDStrPrime, common.ToStrChainID(claims.RefundRequestClaims[0].DestinationChainId))
		require.Equal(t, nexusBridgingAddr, claims.RefundRequestClaims[0].OriginSenderAddress)
		require.Equal(t, big.NewInt(minFeeForBridging), claims.RefundRequestClaims[0].OriginAmount)
		require.Equal(t, big.NewInt(0), claims.RefundRequestClaims[0].OriginWrappedAmount)
		require.Len(t, claims.RefundRequestClaims[0].TokenAmounts, 1)
		require.Equal(t, usdtTokenID, claims.RefundRequestClaims[0].TokenAmounts[0].TokenId)
		require.Equal(t, txValue, claims.RefundRequestClaims[0].TokenAmounts[0].AmountCurrency)
		require.Equal(t, amountToken, claims.RefundRequestClaims[0].TokenAmounts[0].AmountTokens)
	})
}
