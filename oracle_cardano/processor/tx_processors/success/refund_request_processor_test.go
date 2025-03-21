package successtxprocessors

import (
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/chain"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestRefundRequestedProcessor(t *testing.T) {
	const (
		utxoMinValue           = 1000000
		minFeeForBridging      = 1000010
		primeBridgingAddr      = "addr_test1vq6xsx99frfepnsjuhzac48vl9s2lc9awkvfknkgs89srqqslj660"
		primeBridgingFeeAddr   = "addr_test1vqqj5apwf5npsmudw0ranypkj9jw98t25wk4h83jy5mwypswekttt"
		vectorBridgingAddr     = "vector_test1w2h482rf4gf44ek0rekamxksulazkr64yf2fhmm7f5gxjpsdm4zsg"
		vectorBridgingFeeAddr  = "vector_test1wtyslvqxffyppmzhs7ecwunsnpq6g2p6kf9r4aa8ntfzc4qj925fr"
		validPrimeTestAddress  = "addr_test1wrz24vv4tvfqsywkxn36rv5zagys2d7euafcgt50gmpgqpq4ju9uv"
		validVectorTestAddress = "vector_test1vgrgxh4s35a5pdv0dc4zgq33crn34emnk2e7vnensf4tezq3tkm9m"
	)

	maxAmountAllowedToBridge := new(big.Int).SetUint64(100000000)

	appConfig := &cCore.AppConfig{
		CardanoChains: map[string]*cCore.CardanoChainConfig{
			common.ChainIDStrPrime: {
				NetworkID: wallet.TestNetNetwork,
				OgmiosURL: "http://ogmios.prime.testnet.apexfusion.org:1337",
				BridgingAddresses: cCore.BridgingAddresses{
					BridgingAddress: primeBridgingAddr,
					FeeAddress:      primeBridgingFeeAddr,
				},
				UtxoMinAmount:     utxoMinValue,
				MinFeeForBridging: minFeeForBridging,
			},
			common.ChainIDStrVector: {
				NetworkID:     wallet.VectorTestNetNetwork,
				BlockfrostURL: "http://ogmios.vector.testnet.apexfusion.org:1337",
				BridgingAddresses: cCore.BridgingAddresses{
					BridgingAddress: vectorBridgingAddr,
					FeeAddress:      vectorBridgingFeeAddr,
				},
				UtxoMinAmount:     utxoMinValue,
				MinFeeForBridging: minFeeForBridging,
			},
		},
		BridgingSettings: cCore.BridgingSettings{
			MaxReceiversPerBridgingRequest: 3,
			MaxAmountAllowedToBridge:       maxAmountAllowedToBridge,
		},
	}
	appConfig.FillOut()

	chainInfos := make(map[string]*chain.CardanoChainInfo, len(appConfig.CardanoChains))

	for _, cc := range appConfig.CardanoChains {
		info := chain.NewCardanoChainInfo(cc)

		// err := info.Populate(ctx)
		// require.NoError(t, err)

		chainInfos[cc.ChainID] = info
	}

	proc := NewRefundRequestProcessor(hclog.NewNullLogger(), chainInfos)

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &cCore.BridgeClaims{}

		err := proc.ValidateAndAddClaim(claims, &core.CardanoTx{}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to unmarshal metadata")
	})

	t.Run("ValidateAndAddClaim insufficient metadata", func(t *testing.T) {
		relevantButNotFullMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BaseMetadata{
			BridgingTxType: common.BridgingTxTypeBridgingRequest,
		})
		require.NoError(t, err)
		require.NotNil(t, relevantButNotFullMetadata)

		claims := &cCore.BridgeClaims{}

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: relevantButNotFullMetadata,
			},
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "unsupported chain id found in tx")
	})

	t.Run("ValidateAndAddClaim unsuported sender chainID", func(t *testing.T) {
		destinationChainNonRegisteredMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{validPrimeTestAddress},
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

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: "invalid",
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "unsupported chain id found in tx")
	})

	t.Run("ValidateAndAddClaim invalid sender address", func(t *testing.T) {
		destinationChainNonRegisteredMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"invalid_address"},
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

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "invalid sender addr")
	})

	t.Run("ValidateAndAddClaim outputs contains more unknown tokens than allowed", func(t *testing.T) {
		destinationChainNonRegisteredMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     common.BridgingTxTypeBridgingRequest,
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{validPrimeTestAddress},
			Transactions:       []common.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, destinationChainNonRegisteredMetadata)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{
				Address: primeBridgingAddr,
				Amount:  1,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "1",
						Name:     "1",
						Amount:   1_000_000,
					},
				},
			},
			{
				Address: primeBridgingAddr,
				Amount:  1_000_000,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "3",
						Name:     "1",
						Amount:   2_000_000,
					},
				},
			},
			{
				Address: primeBridgingAddr,
				Amount:  1_000_000,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "3",
						Name:     "3",
						Amount:   3_000_000,
					},
				},
			},
			{
				Address: primeBridgingAddr,
				Amount:  1_000_000,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "3",
						Name:     "1",
						Amount:   2_000_000,
					},
				},
			},
			{Address: "addr2", Amount: 2_000_000},
			{
				Address: primeBridgingAddr,
				Amount:  3_000_000,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "1",
						Name:     "3",
						Amount:   100_000,
					},
				},
			},
			{Address: primeBridgingFeeAddr, Amount: 4},
		}

		tx := indexer.Tx{
			Metadata: destinationChainNonRegisteredMetadata,
			Outputs:  txOutputs,
		}

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "more UTxOs with unknown tokens than allowed")
	})
}
