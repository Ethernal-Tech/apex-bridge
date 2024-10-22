package successtxprocessors

import (
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	"github.com/Ethernal-Tech/ethgo"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestHotWalletIncrementProcessor(t *testing.T) {
	const (
		nexusBridgingAddr = "0xA4d1233A67776575425Ab185f6a9251aa00fEA25"
	)

	proc := NewHotWalletIncrementProcessor(hclog.NewNullLogger())
	appConfig := &oCore.AppConfig{
		EthChains: map[string]*oCore.EthChainConfig{
			common.ChainIDStrNexus: {
				BridgingAddresses: oCore.EthBridgingAddresses{
					BridgingAddress: nexusBridgingAddr,
				},
			},
		},
	}
	appConfig.FillOut()

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &oCore.BridgeClaims{}

		err := proc.ValidateAndAddClaim(claims, &core.EthTx{}, appConfig)
		require.Error(t, err)
	})

	t.Run("ValidateAndAddClaim random metadata", func(t *testing.T) {
		claims := &oCore.BridgeClaims{}

		err := proc.ValidateAndAddClaim(claims, &core.EthTx{
			Metadata: []byte{1, 2, 3},
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "validation failed for tx")
	})

	t.Run("ValidateAndAddClaim wrong hot wallet address", func(t *testing.T) {
		claims := &oCore.BridgeClaims{}
		err := proc.ValidateAndAddClaim(claims, &core.EthTx{
			Address:       ethgo.HexToAddress("0xBadBadBad7776575425Ab185f6a9251aa00fEA25"),
			Metadata:      []byte{},
			OriginChainID: common.ToStrChainID(common.ChainIDIntNexus),
			Value:         new(big.Int).SetUint64(1),
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "validation failed for tx")
	})

	t.Run("ValidateAndAddClaim wrong origin chain", func(t *testing.T) {
		claims := &oCore.BridgeClaims{}
		err := proc.ValidateAndAddClaim(claims, &core.EthTx{
			Address:       ethgo.HexToAddress("0xA4d1233A67776575425Ab185f6a9251aa00fEA25"),
			Metadata:      []byte{},
			OriginChainID: common.ChainIDStrPrime,
			Value:         new(big.Int).SetUint64(1),
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "validation failed for tx")
	})

	t.Run("ValidateAndAddClaim valid", func(t *testing.T) {
		claims := &oCore.BridgeClaims{}
		err := proc.ValidateAndAddClaim(claims, &core.EthTx{
			Address:       ethgo.HexToAddress("0xA4d1233A67776575425Ab185f6a9251aa00fEA25"),
			Metadata:      []byte{},
			OriginChainID: common.ChainIDStrNexus,
			Value:         new(big.Int).SetUint64(1),
		}, appConfig)
		require.NoError(t, err)
	})
}
