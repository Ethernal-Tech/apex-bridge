package successtxprocessors

import (
	"math/big"
	"testing"

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

	appConfig := &oCore.AppConfig{
		CardanoChains: map[string]*oCore.CardanoChainConfig{
			common.ChainIDStrPrime: {
				NetworkID: wallet.TestNetNetwork,
				BridgingAddresses: oCore.BridgingAddresses{
					BridgingAddress: primeBridgingAddr,
					FeeAddress:      primeBridgingFeeAddr,
				},
				UtxoMinAmount:     utxoMinValue,
				MinFeeForBridging: minFeeForBridging,
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
	appConfig.FillOut()

	proc := NewRefundRequestProcessor(hclog.NewNullLogger())

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &oCore.BridgeClaims{}

		err := proc.ValidateAndAddClaim(claims, &core.EthTx{}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to unmarshal metadata")
	})

	t.Run("ValidateAndAddClaim origin chain not registered", func(t *testing.T) {
		metadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.TxTypeRefundRequest,
			DestinationChainID: "invalid",
			SenderAddr:         "addr1",
			Transactions:       []core.BridgingRequestEthMetadataTransaction{},
			FeeAmount:          big.NewInt(0),
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &oCore.BridgeClaims{}

		err = proc.ValidateAndAddClaim(claims, &core.EthTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "unsupported chain id found in tx")
	})

	t.Run("ValidateAndAddClaim invalid sender address", func(t *testing.T) {
		metadata, err := core.MarshalEthMetadata(core.BridgingRequestEthMetadata{
			BridgingTxType:     common.TxTypeRefundRequest,
			DestinationChainID: common.ChainIDStrPrime,
			SenderAddr:         "invalid_address",
			Transactions:       []core.BridgingRequestEthMetadataTransaction{},
			FeeAmount:          common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging - 1)),
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &oCore.BridgeClaims{}

		err = proc.ValidateAndAddClaim(claims, &core.EthTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrNexus,
		}, appConfig)
		require.Error(t, err)
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
			FeeAmount: common.DfmToWei(new(big.Int).SetUint64(minFeeForBridging - 1)),
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &oCore.BridgeClaims{}

		err = proc.ValidateAndAddClaim(claims, &core.EthTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrNexus,
			Value:         new(big.Int).SetUint64(utxoMinValue - 1),
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "less than the minimum required for refund")
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
			FeeAmount: common.DfmToWei(new(big.Int).SetUint64(100)),
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		claims := &oCore.BridgeClaims{}

		err = proc.ValidateAndAddClaim(claims, &core.EthTx{
			Metadata:      metadata,
			OriginChainID: common.ChainIDStrNexus,
			Value:         common.DfmToWei(new(big.Int).SetUint64(utxoMinValue + minFeeForBridging + 100)),
		}, appConfig)
		require.NoError(t, err)
	})
}
