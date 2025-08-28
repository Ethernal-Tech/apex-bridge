package bridgingaddressscoordinator

import (
	"fmt"
	"testing"

	bam "github.com/Ethernal-Tech/apex-bridge/bridging_addresses_manager"
	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const tokenName = "b8b9cb79fc3317847e6aeee650093a738972e773c0702c7e5fe6e702.7465737431"

var cardanoChains = map[string]*oracleCore.CardanoChainConfig{
	common.ChainIDStrPrime: {
		ChainID: common.ChainIDStrPrime,
		CardanoChainConfig: cardanotx.CardanoChainConfig{
			NativeTokens: []sendtx.TokenExchangeConfig{
				{
					TokenName: tokenName,
				},
			},
		},
	},
}

func TestBridgingAddressesCoordinator(t *testing.T) {
	chainID := common.ChainIDIntPrime

	t.Run("GetAddressesAndAmountsForBatchForBatch 1 address currency", func(t *testing.T) {
		bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
		bridgingAddressesManagerMock.On("GetAllPaymentAddresses", mock.Anything).Return([]string{"addr1"}, nil)

		dbMock := &indexer.DatabaseMock{}
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 1_000_000_000,
					},
				},
			}, error(nil)).Once()

		coordinator := NewBridgingAddressesCoordinator(bridgingAddressesManagerMock, map[string]indexer.Database{
			"prime": dbMock,
		}, cardanoChains, hclog.NewNullLogger())

		amounts, err := coordinator.GetAddressesAndAmountsForBatch(chainID, "", false, []byte{}, &common.TxOutputs{
			Outputs: []cardanowallet.TxOutput{
				{
					Addr:   "addr1",
					Amount: 10_000_000,
				},
			},
			Sum: map[string]uint64{"lovelace": 10_000_000},
		})
		require.NoError(t, err)
		require.Equal(t, "addr1", amounts[0].Address)
		require.Equal(t, uint64(10_000_000), amounts[0].TokensAmounts[cardanowallet.AdaTokenName])
		require.Equal(t, uint64(1_000_000), amounts[0].IncludeChange)
		require.Equal(t, uint8(0), amounts[0].AddressIndex)
	})

	t.Run("GetAddressesAndAmountsForBatchForBatch 1 address currency not enough funds", func(t *testing.T) {
		bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
		bridgingAddressesManagerMock.On("GetAllPaymentAddresses", mock.Anything).Return([]string{"addr1"}, nil)

		dbMock := &indexer.DatabaseMock{}
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 1_000_000,
					},
				},
			}, error(nil)).Once()

		coordinator := NewBridgingAddressesCoordinator(bridgingAddressesManagerMock, map[string]indexer.Database{
			"prime": dbMock,
		}, cardanoChains, hclog.NewNullLogger())

		amounts, err := coordinator.GetAddressesAndAmountsForBatch(chainID, "", false, []byte{}, &common.TxOutputs{
			Outputs: []cardanowallet.TxOutput{
				{
					Addr:   "addr1",
					Amount: 10_000_000,
				},
			},
			Sum: map[string]uint64{"lovelace": 10_000_000},
		})
		require.Error(t, err)
		require.Nil(t, amounts)
	})

	t.Run("GetAddressesAndAmountsForBatchForBatch 1 address currency not enough funds", func(t *testing.T) {
		bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
		bridgingAddressesManagerMock.On("GetAllPaymentAddresses", mock.Anything).Return([]string{"addr1", "addr2"}, nil)

		dbMock := &indexer.DatabaseMock{}
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 10_000_000,
					},
				},
			}, error(nil)).Once()

		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 5_000_000,
					},
				},
			}, error(nil)).Once()

		coordinator := NewBridgingAddressesCoordinator(bridgingAddressesManagerMock, map[string]indexer.Database{
			"prime": dbMock,
		}, cardanoChains, hclog.NewNullLogger())

		amounts, err := coordinator.GetAddressesAndAmountsForBatch(chainID, "", false, []byte{}, &common.TxOutputs{
			Outputs: []cardanowallet.TxOutput{
				{
					Addr:   "addr1",
					Amount: 10_000_000,
				},
			},
			Sum: map[string]uint64{"lovelace": 10_000_000},
		})
		require.NoError(t, err)
		require.Equal(t, uint64(10_000_000), amounts[0].TokensAmounts[cardanowallet.AdaTokenName])
		require.Equal(t, uint64(0), amounts[0].IncludeChange)
	})

	token, err := cardanowallet.NewTokenWithFullName("b8b9cb79fc3317847e6aeee650093a738972e773c0702c7e5fe6e702.7465737431", true)
	require.NoError(t, err)

	protocolParams := []byte(`{"costModels":{"PlutusV1":[197209,0,1,1,396231,621,0,1,150000,1000,0,1,150000,32,2477736,29175,4,29773,100,29773,100,29773,100,29773,100,29773,100,29773,100,100,100,29773,100,150000,32,150000,32,150000,32,150000,1000,0,1,150000,32,150000,1000,0,8,148000,425507,118,0,1,1,150000,1000,0,8,150000,112536,247,1,150000,10000,1,136542,1326,1,1000,150000,1000,1,150000,32,150000,32,150000,32,1,1,150000,1,150000,4,103599,248,1,103599,248,1,145276,1366,1,179690,497,1,150000,32,150000,32,150000,32,150000,32,150000,32,150000,32,148000,425507,118,0,1,1,61516,11218,0,1,150000,32,148000,425507,118,0,1,1,148000,425507,118,0,1,1,2477736,29175,4,0,82363,4,150000,5000,0,1,150000,32,197209,0,1,1,150000,32,150000,32,150000,32,150000,32,150000,32,150000,32,150000,32,3345831,1,1],"PlutusV2":[205665,812,1,1,1000,571,0,1,1000,24177,4,1,1000,32,117366,10475,4,23000,100,23000,100,23000,100,23000,100,23000,100,23000,100,100,100,23000,100,19537,32,175354,32,46417,4,221973,511,0,1,89141,32,497525,14068,4,2,196500,453240,220,0,1,1,1000,28662,4,2,245000,216773,62,1,1060367,12586,1,208512,421,1,187000,1000,52998,1,80436,32,43249,32,1000,32,80556,1,57667,4,1000,10,197145,156,1,197145,156,1,204924,473,1,208896,511,1,52467,32,64832,32,65493,32,22558,32,16563,32,76511,32,196500,453240,220,0,1,1,69522,11687,0,1,60091,32,196500,453240,220,0,1,1,196500,453240,220,0,1,1,1159724,392670,0,2,806990,30482,4,1927926,82523,4,265318,0,4,0,85931,32,205665,812,1,1,41182,32,212342,32,31220,32,32696,32,43357,32,32247,32,38314,32,35892428,10,9462713,1021,10,38887044,32947,10]},"protocolVersion":{"major":7,"minor":0},"maxBlockHeaderSize":1100,"maxBlockBodySize":65536,"maxTxSize":16384,"txFeeFixed":155381,"txFeePerByte":44,"stakeAddressDeposit":0,"stakePoolDeposit":0,"minPoolCost":0,"poolRetireMaxEpoch":18,"stakePoolTargetNum":100,"poolPledgeInfluence":0,"monetaryExpansion":0.1,"treasuryCut":0.1,"collateralPercentage":150,"executionUnitPrices":{"priceMemory":0.0577,"priceSteps":0.0000721},"utxoCostPerByte":4310,"maxTxExecutionUnits":{"memory":16000000,"steps":10000000000},"maxBlockExecutionUnits":{"memory":80000000,"steps":40000000000},"maxCollateralInputs":3,"maxValueSize":5000,"extraPraosEntropy":null,"decentralization":null,"minUTxOValue":null}`)

	t.Run("GetAddressesAndAmountsForBatchForBatch 1 address native", func(t *testing.T) {
		bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
		bridgingAddressesManagerMock.On("GetAllPaymentAddresses", mock.Anything).Return([]string{"addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpr"}, nil)

		dbMock := &indexer.DatabaseMock{}
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 1_000_000_000,
						Tokens: []indexer.TokenAmount{
							{
								PolicyID: token.PolicyID,
								Name:     token.Name,
								Amount:   1000000,
							},
						},
					},
				},
			}, error(nil)).Once()

		coordinator := NewBridgingAddressesCoordinator(bridgingAddressesManagerMock, map[string]indexer.Database{
			"prime": dbMock,
		}, cardanoChains, hclog.NewNullLogger())

		amounts, err := coordinator.GetAddressesAndAmountsForBatch(chainID, cardanowallet.ResolveCardanoCliBinary(cardanowallet.TestNetNetwork), false, protocolParams, &common.TxOutputs{
			Outputs: []cardanowallet.TxOutput{
				{
					Addr:   "addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpr",
					Amount: 10_000_000,
					Tokens: []cardanowallet.TokenAmount{
						cardanowallet.NewTokenAmount(token, 1000000),
					},
				},
			},
			Sum: map[string]uint64{"lovelace": 10_000_000, token.String(): 1000000},
		})

		require.NoError(t, err)
		require.Equal(t, "addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpr", amounts[0].Address)
		require.Equal(t, uint64(10000000), amounts[0].TokensAmounts[cardanowallet.AdaTokenName])
		require.Equal(t, uint64(1000000), amounts[0].IncludeChange)
		require.Equal(t, uint64(1000000), amounts[0].TokensAmounts[token.String()])
		require.Equal(t, uint8(0), amounts[0].AddressIndex)
	})

	//nolint:dupl
	t.Run("GetAddressesAndAmountsForBatchForBatch 1 address native, not enough native funds", func(t *testing.T) {
		bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
		bridgingAddressesManagerMock.On("GetAllPaymentAddresses", mock.Anything).Return([]string{"addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpr"}, nil)

		dbMock := &indexer.DatabaseMock{}
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 1_000_000_000,
						Tokens: []indexer.TokenAmount{
							{
								PolicyID: token.PolicyID,
								Name:     token.Name,
								Amount:   1_000_000,
							},
						},
					},
				},
			}, error(nil)).Once()

		coordinator := NewBridgingAddressesCoordinator(bridgingAddressesManagerMock, map[string]indexer.Database{
			"prime": dbMock,
		}, cardanoChains, hclog.NewNullLogger())

		amounts, err := coordinator.GetAddressesAndAmountsForBatch(chainID, cardanowallet.ResolveCardanoCliBinary(cardanowallet.TestNetNetwork), false, protocolParams, &common.TxOutputs{
			Outputs: []cardanowallet.TxOutput{
				{
					Addr:   "addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpr",
					Amount: 10_000_000,
					Tokens: []cardanowallet.TokenAmount{
						cardanowallet.NewTokenAmount(token, 10_000_000),
					},
				},
			},
			Sum: map[string]uint64{"lovelace": 10_000_000, token.String(): 10_000_000},
		})
		require.Error(t, err)
		require.ErrorContains(t, err, "not enough b8b9cb79fc3317847e6aeee650093a738972e773c0702c7e5fe6e702.7465737431 token funds for batch")
		require.Nil(t, amounts)
	})

	//nolint:dupl
	t.Run("GetAddressesAndAmountsForBatchForBatch 1 address native, not enough currency funds", func(t *testing.T) {
		bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
		bridgingAddressesManagerMock.On("GetAllPaymentAddresses", mock.Anything).Return([]string{"addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpr"}, nil)

		dbMock := &indexer.DatabaseMock{}
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 1_000_000,
						Tokens: []indexer.TokenAmount{
							{
								PolicyID: token.PolicyID,
								Name:     token.Name,
								Amount:   1000000,
							},
						},
					},
				},
			}, error(nil)).Once()

		coordinator := NewBridgingAddressesCoordinator(bridgingAddressesManagerMock, map[string]indexer.Database{
			"prime": dbMock,
		}, cardanoChains, hclog.NewNullLogger())

		amounts, err := coordinator.GetAddressesAndAmountsForBatch(chainID, cardanowallet.ResolveCardanoCliBinary(cardanowallet.TestNetNetwork), false, protocolParams, &common.TxOutputs{
			Outputs: []cardanowallet.TxOutput{
				{
					Addr:   "addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpr",
					Amount: 10_000_000,
					Tokens: []cardanowallet.TokenAmount{
						cardanowallet.NewTokenAmount(token, 1000000),
					},
				},
			},
			Sum: map[string]uint64{"lovelace": 10_000_000, token.String(): 1000000},
		})

		require.Error(t, err)
		require.ErrorContains(t, err, "not enough lovelace token funds for batch")
		require.Nil(t, amounts)
	})

	t.Run("GetAddressesAndAmountsForBatchForBatch 1 address native 2 currency", func(t *testing.T) {
		bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
		bridgingAddressesManagerMock.On("GetAllPaymentAddresses", mock.Anything).Return([]string{"addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpr", "addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpp"}, nil)

		dbMock := &indexer.DatabaseMock{}
		dbMock.On("GetAllTxOutputs", "addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpr", true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 10_000_000,
						Tokens: []indexer.TokenAmount{
							{
								PolicyID: token.PolicyID,
								Name:     token.Name,
								Amount:   1_000_000,
							},
						},
					},
				},
			}, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", "addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpp", true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 9_000_000,
					},
				},
			}, error(nil)).Once()

		coordinator := NewBridgingAddressesCoordinator(bridgingAddressesManagerMock, map[string]indexer.Database{
			"prime": dbMock,
		}, cardanoChains, hclog.NewNullLogger())

		amounts, err := coordinator.GetAddressesAndAmountsForBatch(chainID, cardanowallet.ResolveCardanoCliBinary(cardanowallet.TestNetNetwork), false, protocolParams, &common.TxOutputs{
			Outputs: []cardanowallet.TxOutput{
				{
					Addr:   "addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpr",
					Amount: 10_000_000,
					Tokens: []cardanowallet.TokenAmount{
						cardanowallet.NewTokenAmount(token, 500_000),
					},
				},
			},
			Sum: map[string]uint64{"lovelace": 10_000_000, token.String(): 500_000},
		})

		require.NoError(t, err)
		require.Equal(t, uint64(500_000), amounts[0].TokensAmounts[token.String()])
		require.Equal(t, uint64(8_961_290), amounts[0].TokensAmounts[cardanowallet.AdaTokenName])
		require.Equal(t, uint64(1038710), amounts[0].IncludeChange)
		require.Equal(t, uint64(1038710), amounts[1].TokensAmounts[cardanowallet.AdaTokenName])
		require.Equal(t, uint64(1000000), amounts[1].IncludeChange)
	})

	t.Run("GetAddressesAndAmountsForBatchForBatch 2 address native 2 currency 2", func(t *testing.T) {
		bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
		bridgingAddressesManagerMock.On("GetAllPaymentAddresses", mock.Anything).Return([]string{"addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpr", "addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpp"}, nil)

		dbMock := &indexer.DatabaseMock{}
		dbMock.On("GetAllTxOutputs", "addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpr", true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 10_000_000,
						Tokens: []indexer.TokenAmount{
							{
								PolicyID: token.PolicyID,
								Name:     token.Name,
								Amount:   1_000_000,
							},
						},
					},
				},
			}, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", "addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpp", true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 9_000_000,
						Tokens: []indexer.TokenAmount{
							{
								PolicyID: token.PolicyID,
								Name:     token.Name,
								Amount:   1_000_000,
							},
						},
					},
				},
			}, error(nil)).Once()

		coordinator := NewBridgingAddressesCoordinator(bridgingAddressesManagerMock, map[string]indexer.Database{
			"prime": dbMock,
		}, cardanoChains, hclog.NewNullLogger())

		amounts, err := coordinator.GetAddressesAndAmountsForBatch(chainID, cardanowallet.ResolveCardanoCliBinary(cardanowallet.TestNetNetwork), false, protocolParams, &common.TxOutputs{
			Outputs: []cardanowallet.TxOutput{
				{
					Addr:   "addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpr",
					Amount: 10_000_000,
					Tokens: []cardanowallet.TokenAmount{
						cardanowallet.NewTokenAmount(token, 500_000),
					},
				},
			},
			Sum: map[string]uint64{"lovelace": 10_000_000, token.String(): 500_000},
		})

		require.NoError(t, err)
		require.Equal(t, uint64(500_000), amounts[0].TokensAmounts[token.String()])
		require.Equal(t, uint64(8961290), amounts[0].TokensAmounts[cardanowallet.AdaTokenName])
		require.Equal(t, uint64(1038710), amounts[0].IncludeChange)
		require.Equal(t, uint64(1038710), amounts[1].TokensAmounts[cardanowallet.AdaTokenName])
		require.Equal(t, uint64(0), amounts[1].TokensAmounts[token.String()])
		require.Equal(t, uint64(1000000), amounts[1].IncludeChange)
	})
}

func TestRedistributeTokens(t *testing.T) {
	chainID := common.ChainIDIntPrime

	t.Run("GetAddressesAndAmountsForBatch 1 address", func(t *testing.T) {
		bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
		bridgingAddressesManagerMock.On("GetAllPaymentAddresses", mock.Anything).Return([]string{"addr1"}, nil)

		token, err := cardanowallet.NewTokenWithFullName(tokenName, true)
		require.NoError(t, err)

		dbMock := &indexer.DatabaseMock{}
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 500_000_000,
						Tokens: []indexer.TokenAmount{
							{
								PolicyID: token.PolicyID,
								Name:     token.Name,
								Amount:   1000000,
							},
						},
					},
				},
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0013"),
					},
					Output: indexer.TxOutput{
						Amount: 500_000_000,
						Tokens: []indexer.TokenAmount{
							{
								PolicyID: token.PolicyID,
								Name:     token.Name,
								Amount:   1000000,
							},
						},
					},
				},
			}, error(nil)).Once()

		coordinator := NewBridgingAddressesCoordinator(bridgingAddressesManagerMock, map[string]indexer.Database{
			"prime": dbMock,
		}, cardanoChains, hclog.NewNullLogger())

		amounts, err := coordinator.GetAddressesAndAmountsForBatch(chainID, "", true, []byte{}, &common.TxOutputs{})
		require.NoError(t, err)
		require.Equal(t, "addr1", amounts[0].Address)
		require.Equal(t, uint64(1_000_000_000), amounts[0].TokensAmounts[cardanowallet.AdaTokenName])
		require.Equal(t, uint8(0), amounts[0].AddressIndex)
		require.Equal(t, uint64(2000000), amounts[0].TokensAmounts[tokenName])
	})

	t.Run("GetAddressesAndAmountsForBatch 2 addresses", func(t *testing.T) {
		addresses := []string{"addr1", "addr2"}
		bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
		bridgingAddressesManagerMock.On("GetAllPaymentAddresses", mock.Anything).Return(addresses, nil)

		dbMock := &indexer.DatabaseMock{}
		dbMock.On("GetAllTxOutputs", addresses[0], true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 200_000_000,
					},
				},
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0013"),
					},
					Output: indexer.TxOutput{
						Amount: 400_000_000,
					},
				},
			}, error(nil))
		dbMock.On("GetAllTxOutputs", addresses[1], true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0013"),
					},
					Output: indexer.TxOutput{
						Amount: 400_000_000,
					},
				},
			}, error(nil))

		coordinator := NewBridgingAddressesCoordinator(bridgingAddressesManagerMock, map[string]indexer.Database{
			"prime": dbMock,
		}, cardanoChains, hclog.NewNullLogger())

		amounts, err := coordinator.GetAddressesAndAmountsForBatch(chainID, "", true, []byte{}, &common.TxOutputs{})
		require.NoError(t, err)
		require.Equal(t, 2, len(amounts))

		for i, amount := range amounts {
			require.Equal(t, addresses[i], amount.Address)
			require.Equal(t, uint64(500_000_000), amount.TokensAmounts[cardanowallet.AdaTokenName])
			require.Equal(t, uint8(i), amount.AddressIndex)
		}
	})

	t.Run("GetAddressesAndAmountsForBatch 3 addresses", func(t *testing.T) {
		addresses := []string{"addr1", "addr2", "addr3"}
		expectedAmounts := []uint64{333333334, 333333333, 333333333}

		bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
		bridgingAddressesManagerMock.On("GetAllPaymentAddresses", mock.Anything).Return(addresses, nil)

		dbMock := &indexer.DatabaseMock{}
		dbMock.On("GetAllTxOutputs", addresses[0], true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 200_000_000,
					},
				},
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0013"),
					},
					Output: indexer.TxOutput{
						Amount: 400_000_000,
					},
				},
			}, error(nil))
		dbMock.On("GetAllTxOutputs", addresses[1], true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0013"),
					},
					Output: indexer.TxOutput{
						Amount: 400_000_000,
					},
				},
			}, error(nil))
		dbMock.On("GetAllTxOutputs", addresses[2], true).Return([]*indexer.TxInputOutput{}, error(nil))

		coordinator := NewBridgingAddressesCoordinator(bridgingAddressesManagerMock, map[string]indexer.Database{
			"prime": dbMock,
		}, cardanoChains, hclog.NewNullLogger())

		amounts, err := coordinator.GetAddressesAndAmountsForBatch(chainID, "", true, []byte{}, &common.TxOutputs{})
		require.NoError(t, err)
		require.Equal(t, 3, len(amounts))

		for i, amount := range amounts {
			require.Equal(t, addresses[i], amount.Address)
			require.Equal(t, expectedAmounts[i], amount.TokensAmounts[cardanowallet.AdaTokenName])
			require.Equal(t, uint8(i), amount.AddressIndex)
		}
	})

	t.Run("GetAddressesAndAmountsForBatch 3 addresses with output", func(t *testing.T) {
		addresses := []string{"addr1", "addr2", "addr3"}

		bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
		bridgingAddressesManagerMock.On("GetAllPaymentAddresses", mock.Anything).Return(addresses, nil)

		dbMock := &indexer.DatabaseMock{}
		dbMock.On("GetAllTxOutputs", addresses[0], true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 200_000_000,
					},
				},
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0013"),
					},
					Output: indexer.TxOutput{
						Amount: 400_000_000,
					},
				},
			}, error(nil))
		dbMock.On("GetAllTxOutputs", addresses[1], true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0013"),
					},
					Output: indexer.TxOutput{
						Amount: 400_000_000,
					},
				},
			}, error(nil))
		dbMock.On("GetAllTxOutputs", addresses[2], true).Return([]*indexer.TxInputOutput{}, error(nil))

		coordinator := NewBridgingAddressesCoordinator(bridgingAddressesManagerMock, map[string]indexer.Database{
			"prime": dbMock,
		}, cardanoChains, hclog.NewNullLogger())

		amounts, err := coordinator.GetAddressesAndAmountsForBatch(chainID, "", true, []byte{}, &common.TxOutputs{
			Outputs: []cardanowallet.TxOutput{
				{
					Amount: 100_000_000,
				},
			},
			Sum: map[string]uint64{"lovelace": 100_000_000},
		})
		require.NoError(t, err)
		require.Equal(t, 3, len(amounts))

		for i, amount := range amounts {
			require.Equal(t, addresses[i], amount.Address)
			require.Equal(t, uint64(300_000_000), amount.TokensAmounts[cardanowallet.AdaTokenName])
			require.Equal(t, uint8(i), amount.AddressIndex)
		}
	})

	t.Run("GetAddressesAndAmountsForBatch 3 addresses with outputs", func(t *testing.T) {
		addresses := []string{"addr1", "addr2", "addr3"}
		expectedTokens := []uint64{0, 1000000, 0}

		bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
		bridgingAddressesManagerMock.On("GetAllPaymentAddresses", mock.Anything).Return(addresses, nil)

		token, err := cardanowallet.NewTokenWithFullName(tokenName, true)
		require.NoError(t, err)

		dbMock := &indexer.DatabaseMock{}
		dbMock.On("GetAllTxOutputs", addresses[0], true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 200_000_000,
						Tokens: []indexer.TokenAmount{
							{
								PolicyID: token.PolicyID,
								Name:     token.Name,
								Amount:   1000000,
							},
						},
					},
				},
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0013"),
					},
					Output: indexer.TxOutput{
						Amount: 400_000_000,
					},
				},
			}, error(nil))
		dbMock.On("GetAllTxOutputs", addresses[1], true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0013"),
					},
					Output: indexer.TxOutput{
						Amount: 400_000_000,
						Tokens: []indexer.TokenAmount{
							{
								PolicyID: token.PolicyID,
								Name:     token.Name,
								Amount:   2000000,
							},
						},
					},
				},
			}, error(nil))
		dbMock.On("GetAllTxOutputs", addresses[2], true).Return([]*indexer.TxInputOutput{}, error(nil))

		coordinator := NewBridgingAddressesCoordinator(bridgingAddressesManagerMock, map[string]indexer.Database{
			"prime": dbMock,
		}, cardanoChains, hclog.NewNullLogger())

		amounts, err := coordinator.GetAddressesAndAmountsForBatch(chainID, "", true, []byte{}, &common.TxOutputs{
			Outputs: []cardanowallet.TxOutput{
				{
					Amount: 100_000_000,
					Tokens: []cardanowallet.TokenAmount{
						{
							Token: cardanowallet.Token{
								PolicyID: token.PolicyID,
								Name:     token.Name,
							},
							Amount: 2000000,
						},
					},
				},
				{
					Amount: 600_000_000,
				},
			},
			Sum: map[string]uint64{"lovelace": 700_000_000, token.String(): 2000000},
		})
		require.NoError(t, err)
		require.Equal(t, 3, len(amounts))

		for i, amount := range amounts {
			require.Equal(t, addresses[i], amount.Address)
			require.Equal(t, uint64(100_000_000), amount.TokensAmounts[cardanowallet.AdaTokenName])
			require.Equal(t, expectedTokens[i], amount.TokensAmounts[tokenName])
			require.Equal(t, uint8(i), amount.AddressIndex)
		}
	})

	t.Run("GetAddressesAndAmountsForBatch 1 address not enough funds", func(t *testing.T) {
		bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
		bridgingAddressesManagerMock.On("GetAllPaymentAddresses", mock.Anything).Return([]string{"addr1"}, nil)

		token, err := cardanowallet.NewTokenWithFullName(tokenName, true)
		require.NoError(t, err)

		dbMock := &indexer.DatabaseMock{}
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 500_000_000,
						Tokens: []indexer.TokenAmount{
							{
								PolicyID: token.PolicyID,
								Name:     token.Name,
								Amount:   1000000,
							},
						},
					},
				},
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0013"),
					},
					Output: indexer.TxOutput{
						Amount: 500_000_000,
						Tokens: []indexer.TokenAmount{
							{
								PolicyID: token.PolicyID,
								Name:     token.Name,
								Amount:   1000000,
							},
						},
					},
				},
			}, error(nil))

		coordinator := NewBridgingAddressesCoordinator(bridgingAddressesManagerMock, map[string]indexer.Database{
			"prime": dbMock,
		}, cardanoChains, hclog.NewNullLogger())

		_, err = coordinator.GetAddressesAndAmountsForBatch(chainID, "", true, []byte{}, &common.TxOutputs{
			Outputs: []cardanowallet.TxOutput{
				{
					Amount: 500_000_000,
				},
				{
					Amount: 600_000_000,
				},
			},
			Sum: map[string]uint64{"lovelace": 11_000_000_000},
		})
		require.Error(t, err)
		require.ErrorContains(t, err, "not enough lovelace token funds for batch")

		tok := cardanowallet.TokenAmount{
			Token: cardanowallet.Token{
				PolicyID: token.PolicyID,
				Name:     token.Name,
			},
			Amount: 3000000,
		}

		_, err = coordinator.GetAddressesAndAmountsForBatch(chainID, "", true, []byte{}, &common.TxOutputs{
			Outputs: []cardanowallet.TxOutput{
				{
					Amount: 500_000_000,
					Tokens: []cardanowallet.TokenAmount{
						tok,
					},
				},
			},
			Sum: map[string]uint64{"lovelace": 500_000_000, tok.String(): 3000000},
		})
		require.Error(t, err)
		require.ErrorContains(t, err, fmt.Sprintf("not enough %s token funds", tok.TokenName()))
	})
}
