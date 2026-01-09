package successtxprocessors

import (
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

func TestRefundRequestedProcessor(t *testing.T) {
	const (
		utxoMinValue         = 1000000
		minFeeForBridging    = 1000010
		primeBridgingAddr    = "addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj660"
		primeBridgingFeeAddr = "addr_test1vqqj5apwf5npsmudw0ranypkj9jw98t25wk4h83jy5mwypswekttt"
		nexusBridgingAddr    = "0xA4d1233A67776575425Ab185f6a9251aa00fEA25"
		validTestAddress     = "addr_test1vq6zkfat4rlmj2nd2sylpjjg5qhcg9mk92wykaw4m2dp2rqneafvl"
	)

	maxAmountAllowedToBridge := new(big.Int).SetUint64(100000000)

	getAppConfig := func(refundEnabled bool) *oCore.AppConfig {
		appConfig := &oCore.AppConfig{
			CardanoChains: map[string]*oCore.CardanoChainConfig{
				common.ChainIDStrPrime: {
					CardanoChainConfig: cardanotx.CardanoChainConfig{
						NetworkID:                wallet.TestNetNetwork,
						UtxoMinAmount:            utxoMinValue,
						DefaultMinFeeForBridging: minFeeForBridging,
						MinFeeForBridgingTokens:  minFeeForBridging,
					},
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
			RefundEnabled: refundEnabled,
			TryCountLimits: oCore.TryCountLimits{
				MaxRefundTryCount: 3,
			},
			ChainIDConverter: common.NewTestChainIDConverter(),
		}

		appConfig.FillOut()

		return appConfig
	}

	proc := NewRefundRequestProcessor(hclog.NewNullLogger())
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
		require.NoError(t, err)

		require.Len(t, claims.RefundRequestClaims, 1)
		require.Equal(t, common.ChainIDIntNexus, claims.RefundRequestClaims[0].OriginChainId)
		require.Equal(t, uint64(utxoMinValue+minFeeForBridging+100), claims.RefundRequestClaims[0].TokenAmounts[0].AmountCurrency.Uint64())
		require.Equal(t, uint64(0), claims.RefundRequestClaims[0].TokenAmounts[0].AmountTokens.Uint64())
		require.Equal(t, uint16(0), claims.RefundRequestClaims[0].TokenAmounts[0].TokenId)
		require.Equal(t, uint64(utxoMinValue+minFeeForBridging+100), claims.RefundRequestClaims[0].OriginAmount.Uint64())
		require.Equal(t, uint64(0), claims.RefundRequestClaims[0].OriginWrappedAmount.Uint64())
		require.Empty(t, claims.RefundRequestClaims[0].OutputIndexes)
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

	t.Run("ValidateAndAddClaim valid", func(t *testing.T) {
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
		require.NoError(t, err)

		require.Len(t, claims.RefundRequestClaims, 1)
		require.Equal(t, common.ChainIDIntNexus, claims.RefundRequestClaims[0].OriginChainId)
		require.Equal(t, common.ChainIDIntPrime, claims.RefundRequestClaims[0].DestinationChainId)
		require.Equal(t, uint64(utxoMinValue+minFeeForBridging+100), claims.RefundRequestClaims[0].TokenAmounts[0].AmountCurrency.Uint64())
		require.Equal(t, uint64(0), claims.RefundRequestClaims[0].TokenAmounts[0].AmountTokens.Uint64())
		require.Equal(t, uint16(0), claims.RefundRequestClaims[0].TokenAmounts[0].TokenId)
		require.Equal(t, uint64(utxoMinValue+minFeeForBridging+100), claims.RefundRequestClaims[0].OriginAmount.Uint64())
		require.Equal(t, uint64(0), claims.RefundRequestClaims[0].OriginWrappedAmount.Uint64())
		require.Empty(t, claims.RefundRequestClaims[0].OutputIndexes)
	})
}
