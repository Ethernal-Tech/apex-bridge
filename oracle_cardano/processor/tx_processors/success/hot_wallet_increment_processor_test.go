package successtxprocessors

import (
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestHotWalletIncrementProcessor(t *testing.T) {
	const (
		utxoMinValue          = 1000000
		minFeeForBridging     = 10000010
		primeBridgingAddr     = "addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj660"
		primeBridgingFeeAddr  = "addr_test1vqqj5apwf5npsmudw0ranypkj9jw98t25wk4h83jy5mwypswekttt"
		vectorBridgingAddr    = "addr_test1vr076kzqu8ejq22y4e3j0rpck54nlvryd8sjkewjxzsrjgq2lszpw"
		vectorBridgingFeeAddr = "addr_test1vpg5t5gv784rmlze9ye0r9nud706d2v5v94d5h7kpvllamgq6yfx4"
		validTestAddress      = "addr_test1vq6zkfat4rlmj2nd2sylpjjg5qhcg9mk92wykaw4m2dp2rqneafvl"
	)

	proc := NewHotWalletIncrementProcessor(hclog.NewNullLogger())
	appConfig := &cCore.AppConfig{
		CardanoChains: map[string]*cCore.CardanoChainConfig{
			common.ChainIDStrPrime: {
				BridgingAddresses: cCore.BridgingAddresses{
					BridgingAddress: primeBridgingAddr,
					FeeAddress:      primeBridgingFeeAddr,
				},
			},
			common.ChainIDStrVector: {
				BridgingAddresses: cCore.BridgingAddresses{
					BridgingAddress: vectorBridgingAddr,
					FeeAddress:      vectorBridgingFeeAddr,
				},
			},
		},
		BridgingSettings: cCore.BridgingSettings{
			MinFeeForBridging:              minFeeForBridging,
			UtxoMinValue:                   utxoMinValue,
			MaxReceiversPerBridgingRequest: 3,
		},
	}
	appConfig.FillOut()

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &cCore.BridgeClaims{}

		err := proc.ValidateAndAddClaim(claims, &core.CardanoTx{}, appConfig)
		require.Error(t, err)
	})

	t.Run("ValidateAndAddClaim random metadata", func(t *testing.T) {
		claims := &cCore.BridgeClaims{}

		err := proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: []byte{1, 2, 3},
			},
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "validation failed for tx")
	})

	t.Run("ValidateAndAddClaim no outputs", func(t *testing.T) {
		claims := &cCore.BridgeClaims{}
		err := proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: []byte{},
				Outputs:  []*indexer.TxOutput{},
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "validation failed for tx")
	})

	t.Run("ValidateAndAddClaim wrong hot wallet address", func(t *testing.T) {
		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: 1},
		}
		err := proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: []byte{},
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrVector,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "validation failed for tx")
	})

	t.Run("ValidateAndAddClaim multiple utxos", func(t *testing.T) {
		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: 1},
			{Address: primeBridgingAddr, Amount: 2},
		}
		err := proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: []byte{},
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "validation failed for tx")
	})

	t.Run("ValidateAndAddClaim valid", func(t *testing.T) {
		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: 1},
		}
		err := proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: []byte{},
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.NoError(t, err)
	})
}
