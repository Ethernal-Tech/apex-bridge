package successtxprocessors

import (
	"testing"

	brAddrManager "github.com/Ethernal-Tech/apex-bridge/bridging_addresses_manager"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestHotWalletIncrementProcessor(t *testing.T) {
	const (
		primeBridgingAddr     = "addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj660"
		primeBridgingFeeAddr  = "addr_test1vqqj5apwf5npsmudw0ranypkj9jw98t25wk4h83jy5mwypswekttt"
		vectorBridgingAddr    = "addr_test1vr076kzqu8ejq22y4e3j0rpck54nlvryd8sjkewjxzsrjgq2lszpw"
		vectorBridgingFeeAddr = "addr_test1vpg5t5gv784rmlze9ye0r9nud706d2v5v94d5h7kpvllamgq6yfx4"
		validTestAddress      = "addr_test1vq6zkfat4rlmj2nd2sylpjjg5qhcg9mk92wykaw4m2dp2rqneafvl"
	)

	proc := NewHotWalletIncrementProcessor(hclog.NewNullLogger())

	brAddrManagerMock := &brAddrManager.BridgingAddressesManagerMock{}
	brAddrManagerMock.On("GetAllPaymentAddresses", common.ChainIDIntPrime).Return([]string{primeBridgingAddr}, nil)
	brAddrManagerMock.On("GetFeeMultisigAddress", common.ChainIDIntPrime).Return(primeBridgingFeeAddr)
	brAddrManagerMock.On("GetAllPaymentAddresses", common.ChainIDIntVector).Return([]string{vectorBridgingAddr}, nil)
	brAddrManagerMock.On("GetFeeMultisigAddress", common.ChainIDIntVector).Return(vectorBridgingFeeAddr)

	appConfig := &cCore.AppConfig{
		BridgingAddressesManager: brAddrManagerMock,
		CardanoChains: map[string]*cCore.CardanoChainConfig{
			common.ChainIDStrPrime: {
				/* BridgingAddresses: cCore.BridgingAddresses{
					BridgingAddress: primeBridgingAddr,
					FeeAddress:      primeBridgingFeeAddr,
				}, */
			},
			common.ChainIDStrVector: {
				/* BridgingAddresses: cCore.BridgingAddresses{
					BridgingAddress: vectorBridgingAddr,
					FeeAddress:      vectorBridgingFeeAddr,
				}, */
			},
		},
	}
	appConfig.FillOut()

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		err := proc.PreValidate(&core.CardanoTx{}, appConfig)
		require.Error(t, err)
	})

	t.Run("ValidateAndAddClaim random metadata", func(t *testing.T) {
		txOutputs := []*indexer.TxOutput{
			{Address: primeBridgingAddr, Amount: 1},
		}

		err := proc.PreValidate(&core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: []byte{1, 2, 3},
				Outputs:  txOutputs,
			},
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "validation failed for tx")
	})

	t.Run("ValidateAndAddClaim no outputs", func(t *testing.T) {
		err := proc.PreValidate(&core.CardanoTx{
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
		txOutputs := []*indexer.TxOutput{
			{Address: vectorBridgingAddr, Amount: 1},
		}

		err := proc.PreValidate(&core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: []byte{},
				Outputs:  txOutputs,
			},
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "validation failed for tx")
	})

	t.Run("ValidateAndAddClaim unknown tokens", func(t *testing.T) {
		claims := &cCore.BridgeClaims{}
		tx := &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: []byte{},
				Outputs: []*indexer.TxOutput{
					{Address: primeBridgingAddr, Amount: 1, Tokens: []indexer.TokenAmount{
						{
							PolicyID: "111",
							Name:     "222",
							Amount:   1_000_000,
						},
					}},
				},
			},
			OriginChainID: common.ChainIDStrPrime,
		}

		err := proc.PreValidate(tx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "with some unknown tokens")

		err = proc.ValidateAndAddClaim(claims, tx, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "with some unknown tokens")
	})

	t.Run("ValidateAndAddClaim valid", func(t *testing.T) {
		claims := &cCore.BridgeClaims{}
		tx := &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: []byte{},
				Outputs: []*indexer.TxOutput{
					{Address: primeBridgingAddr, Amount: 1},
				},
			},
			OriginChainID: common.ChainIDStrPrime,
		}

		err := proc.PreValidate(tx, appConfig)
		require.NoError(t, err)

		err = proc.ValidateAndAddClaim(claims, tx, appConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateAndAddClaim multiple utxos valid", func(t *testing.T) {
		claims := &cCore.BridgeClaims{}
		tx := &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: []byte{},
				Outputs: []*indexer.TxOutput{
					{Address: primeBridgingAddr, Amount: 1},
					{Address: primeBridgingAddr, Amount: 2},
					{Address: validTestAddress, Amount: 3},
				},
			},
			OriginChainID: common.ChainIDStrPrime,
		}

		err := proc.PreValidate(tx, appConfig)
		require.NoError(t, err)

		err = proc.ValidateAndAddClaim(claims, tx, appConfig)
		require.NoError(t, err)
	})
}
