package successtxprocessors

import (
	"fmt"
	"math/big"
	"testing"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/chain"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

var (
	protocolParameters = []byte(`{"costModels":{"PlutusV1":[197209,0,1,1,396231,621,0,1,150000,1000,0,1,150000,32,2477736,29175,4,29773,100,29773,100,29773,100,29773,100,29773,100,29773,100,100,100,29773,100,150000,32,150000,32,150000,32,150000,1000,0,1,150000,32,150000,1000,0,8,148000,425507,118,0,1,1,150000,1000,0,8,150000,112536,247,1,150000,10000,1,136542,1326,1,1000,150000,1000,1,150000,32,150000,32,150000,32,1,1,150000,1,150000,4,103599,248,1,103599,248,1,145276,1366,1,179690,497,1,150000,32,150000,32,150000,32,150000,32,150000,32,150000,32,148000,425507,118,0,1,1,61516,11218,0,1,150000,32,148000,425507,118,0,1,1,148000,425507,118,0,1,1,2477736,29175,4,0,82363,4,150000,5000,0,1,150000,32,197209,0,1,1,150000,32,150000,32,150000,32,150000,32,150000,32,150000,32,150000,32,3345831,1,1],"PlutusV2":[205665,812,1,1,1000,571,0,1,1000,24177,4,1,1000,32,117366,10475,4,23000,100,23000,100,23000,100,23000,100,23000,100,23000,100,100,100,23000,100,19537,32,175354,32,46417,4,221973,511,0,1,89141,32,497525,14068,4,2,196500,453240,220,0,1,1,1000,28662,4,2,245000,216773,62,1,1060367,12586,1,208512,421,1,187000,1000,52998,1,80436,32,43249,32,1000,32,80556,1,57667,4,1000,10,197145,156,1,197145,156,1,204924,473,1,208896,511,1,52467,32,64832,32,65493,32,22558,32,16563,32,76511,32,196500,453240,220,0,1,1,69522,11687,0,1,60091,32,196500,453240,220,0,1,1,196500,453240,220,0,1,1,1159724,392670,0,2,806990,30482,4,1927926,82523,4,265318,0,4,0,85931,32,205665,812,1,1,41182,32,212342,32,31220,32,32696,32,43357,32,32247,32,38314,32,35892428,10,9462713,1021,10,38887044,32947,10]},"protocolVersion":{"major":7,"minor":0},"maxBlockHeaderSize":1100,"maxBlockBodySize":65536,"maxTxSize":16384,"txFeeFixed":155381,"txFeePerByte":44,"stakeAddressDeposit":0,"stakePoolDeposit":0,"minPoolCost":0,"poolRetireMaxEpoch":18,"stakePoolTargetNum":100,"poolPledgeInfluence":0,"monetaryExpansion":0.1,"treasuryCut":0.1,"collateralPercentage":150,"executionUnitPrices":{"priceMemory":0.0577,"priceSteps":0.0000721},"utxoCostPerByte":4310,"maxTxExecutionUnits":{"memory":16000000,"steps":10000000000},"maxBlockExecutionUnits":{"memory":80000000,"steps":40000000000},"maxCollateralInputs":3,"maxValueSize":5000,"extraPraosEntropy":null,"decentralization":null,"minUTxOValue":null}`)
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

	token, _ := wallet.NewTokenWithFullNameTry("29f8873beb52e126f207a2dfd50f7cff556806b5b4cba9834a7b26a8.4b6173685f546f6b656e")
	tokenAmount := wallet.NewTokenAmount(token, 2_000_000)

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
					MinFeeForBridging: minFeeForBridging,
				},
				common.ChainIDStrVector: {
					BridgingAddresses: cCore.BridgingAddresses{
						BridgingAddress: vectorBridgingAddr,
						FeeAddress:      vectorBridgingFeeAddr,
					},
					CardanoChainConfig: cardanotx.CardanoChainConfig{
						NetworkID:     wallet.TestNetNetwork,
						OgmiosURL:     "http://ogmios.vector.testnet.apexfusion.org:1337",
						UtxoMinAmount: utxoMinValue,
					},
					MinFeeForBridging: minFeeForBridging,
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

	getChainInfos := func() map[string]*chain.CardanoChainInfo {
		appConfig := getAppConfig(true)
		chainInfos := make(map[string]*chain.CardanoChainInfo, len(appConfig.CardanoChains))

		for _, cc := range appConfig.CardanoChains {
			info := chain.NewCardanoChainInfo(cc)

			info.ProtocolParams = protocolParameters

			chainInfos[cc.ChainID] = info
		}

		return chainInfos
	}

	proc := NewRefundRequestProcessor(hclog.NewNullLogger(), getChainInfos())
	disabledProc := NewRefundDisabledProcessor()

	t.Run("Refund disabled - HandleBridgingProcessorPreValidate", func(t *testing.T) {
		appConfig := getAppConfig(false)

		err := disabledProc.HandleBridgingProcessorPreValidate(&core.CardanoTx{}, appConfig)
		require.NoError(t, err)
	})

	t.Run("Refund disabled - HandleBridgingProcessorError", func(t *testing.T) {
		appConfig := getAppConfig(false)

		err := disabledProc.HandleBridgingProcessorError(
			&cCore.BridgeClaims{}, &core.CardanoTx{}, appConfig, fmt.Errorf("test err"), "")
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("Refund disabled - ValidateAndAddClaim", func(t *testing.T) {
		appConfig := getAppConfig(false)

		err := disabledProc.ValidateAndAddClaim(&cCore.BridgeClaims{}, &core.CardanoTx{}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "refund is not enabled")
	})

	t.Run("HandleBridgingProcessorPreValidate - empty tx", func(t *testing.T) {
		appConfig := getAppConfig(false)

		err := proc.HandleBridgingProcessorPreValidate(&core.CardanoTx{}, appConfig)
		require.NoError(t, err)
	})

	t.Run("HandleBridgingProcessorPreValidate - batchTryCount over", func(t *testing.T) {
		appConfig := getAppConfig(false)

		err := proc.HandleBridgingProcessorPreValidate(&core.CardanoTx{BatchTryCount: 1}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "try count exceeded")
	})

	t.Run("HandleBridgingProcessorPreValidate - submitTryCount over", func(t *testing.T) {
		appConfig := getAppConfig(false)

		err := proc.HandleBridgingProcessorPreValidate(&core.CardanoTx{SubmitTryCount: 1}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "try count exceeded")
	})

	t.Run("HandleBridgingProcessorError - empty ty", func(t *testing.T) {
		appConfig := getAppConfig(false)

		err := proc.HandleBridgingProcessorError(
			&cCore.BridgeClaims{}, &core.CardanoTx{}, appConfig, nil, "")
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to unmarshal metadata, err: EOF")
	})

	t.Run("ValidateAndAddClaim empty tx", func(t *testing.T) {
		claims := &cCore.BridgeClaims{}

		appConfig := getAppConfig(true)

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

		appConfig := getAppConfig(true)

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx: indexer.Tx{
				Metadata: relevantButNotFullMetadata,
			},
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "unsupported chain id found in tx")
	})

	t.Run("ValidateAndAddClaim unsuported sender chainID", func(t *testing.T) {
		metadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{validPrimeTestAddress},
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		appConfig := getAppConfig(true)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: "addr1", Amount: 1},
			{Address: "addr2", Amount: 2},
			{Address: primeBridgingAddr, Amount: 3},
			{Address: primeBridgingFeeAddr, Amount: 4},
		}

		tx := indexer.Tx{
			Metadata: metadata,
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
		metadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{"invalid_address"},
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		appConfig := getAppConfig(true)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: "addr1", Amount: 1},
			{Address: "addr2", Amount: 2},
			{Address: primeBridgingAddr, Amount: 3},
			{Address: primeBridgingFeeAddr, Amount: 4},
		}

		tx := indexer.Tx{
			Metadata: metadata,
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
		metadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{validPrimeTestAddress},
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		appConfig := getAppConfig(true)

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
			Metadata: metadata,
			Outputs:  txOutputs,
		}

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "more UTxOs with unknown tokens than allowed")
	})

	t.Run("ValidateAndAddClaim sum of amounts less than the minimum required", func(t *testing.T) {
		metadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{validPrimeTestAddress},
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, metadata)

		appConfig := getAppConfig(true)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: "addr1", Amount: 500_000},
			{Address: "addr2", Amount: 500_000},
			{
				Address: primeBridgingAddr,
				Amount:  500_000,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: token.PolicyID,
						Name:     token.Name,
						Amount:   tokenAmount.Amount,
					},
				},
			},
			{Address: primeBridgingFeeAddr, Amount: 600_000},
		}

		tx := indexer.Tx{
			Metadata: metadata,
			Outputs:  txOutputs,
		}

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.Error(t, err)
		require.ErrorContains(t, err, "less than the minimum required for refund")
	})

	t.Run("ValidateAndAddClaim valid", func(t *testing.T) {
		validMetadata, err := common.SimulateRealMetadata(common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
			BridgingTxType:     sendtx.BridgingRequestType(common.BridgingTxTypeBridgingRequest),
			DestinationChainID: common.ChainIDStrVector,
			SenderAddr:         []string{validPrimeTestAddress},
			Transactions:       []sendtx.BridgingRequestMetadataTransaction{},
		})
		require.NoError(t, err)
		require.NotNil(t, validMetadata)

		appConfig := getAppConfig(true)

		claims := &cCore.BridgeClaims{}
		txOutputs := []*indexer.TxOutput{
			{Address: "addr1", Amount: 500_000},
			{Address: "addr2", Amount: 500_000},
			{
				Address: primeBridgingAddr,
				Amount:  2_500_000,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: token.PolicyID,
						Name:     token.Name,
						Amount:   tokenAmount.Amount,
					},
				},
			},
			{Address: primeBridgingFeeAddr, Amount: 1_000_000},
		}

		tx := indexer.Tx{
			Metadata: validMetadata,
			Outputs:  txOutputs,
		}

		err = proc.ValidateAndAddClaim(claims, &core.CardanoTx{
			Tx:            tx,
			OriginChainID: common.ChainIDStrPrime,
		}, appConfig)
		require.NoError(t, err)
	})
}
