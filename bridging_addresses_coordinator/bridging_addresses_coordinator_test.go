package bridgingaddressscoordinator

import (
	"testing"

	bam "github.com/Ethernal-Tech/apex-bridge/bridging_addresses_manager"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBridgingAddressesCoordinator(t *testing.T) {
	chainID := common.ChainIDIntPrime

	t.Run("GetAddressesAndAmountsToPayFrom 1 address currency", func(t *testing.T) {
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
		}, hclog.NewNullLogger())

		amounts, err := coordinator.GetAddressesAndAmountsToPayFrom(chainID, "", []byte{}, &[]cardanowallet.TxOutput{
			{
				Addr:   "addr1",
				Amount: 10_000_000,
			},
		})
		require.NoError(t, err)
		require.Equal(t, "addr1", amounts[0].Address)
		require.Equal(t, uint64(10_000_000), amounts[0].TokensAmounts[cardanowallet.AdaTokenName])
		require.Equal(t, uint64(1_000_000), amounts[0].IncludeChnage)
		require.Equal(t, uint8(0), amounts[0].AddressIndex)
	})

	t.Run("GetAddressesAndAmountsToPayFrom 1 address currency not enough funds", func(t *testing.T) {
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
		}, hclog.NewNullLogger())

		amounts, err := coordinator.GetAddressesAndAmountsToPayFrom(chainID, "", []byte{}, &[]cardanowallet.TxOutput{
			{
				Addr:   "addr1",
				Amount: 10_000_000,
			},
		})
		require.Error(t, err)
		require.Nil(t, amounts)
	})

	t.Run("GetAddressesAndAmountsToPayFrom 1 address currency not enough funds", func(t *testing.T) {
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
		}, hclog.NewNullLogger())

		amounts, err := coordinator.GetAddressesAndAmountsToPayFrom(chainID, "", []byte{}, &[]cardanowallet.TxOutput{
			{
				Addr:   "addr1",
				Amount: 10_000_000,
			},
		})
		require.NoError(t, err)
		require.Equal(t, uint64(10_000_000), amounts[0].TokensAmounts[cardanowallet.AdaTokenName])
		require.Equal(t, uint64(0), amounts[0].IncludeChnage)
	})

	token, err := cardanowallet.NewTokenWithFullName("b8b9cb79fc3317847e6aeee650093a738972e773c0702c7e5fe6e702.7465737431", true)
	require.NoError(t, err)

	protocolParams := []byte(`{"costModels":{"PlutusV1":[197209,0,1,1,396231,621,0,1,150000,1000,0,1,150000,32,2477736,29175,4,29773,100,29773,100,29773,100,29773,100,29773,100,29773,100,100,100,29773,100,150000,32,150000,32,150000,32,150000,1000,0,1,150000,32,150000,1000,0,8,148000,425507,118,0,1,1,150000,1000,0,8,150000,112536,247,1,150000,10000,1,136542,1326,1,1000,150000,1000,1,150000,32,150000,32,150000,32,1,1,150000,1,150000,4,103599,248,1,103599,248,1,145276,1366,1,179690,497,1,150000,32,150000,32,150000,32,150000,32,150000,32,150000,32,148000,425507,118,0,1,1,61516,11218,0,1,150000,32,148000,425507,118,0,1,1,148000,425507,118,0,1,1,2477736,29175,4,0,82363,4,150000,5000,0,1,150000,32,197209,0,1,1,150000,32,150000,32,150000,32,150000,32,150000,32,150000,32,150000,32,3345831,1,1],"PlutusV2":[205665,812,1,1,1000,571,0,1,1000,24177,4,1,1000,32,117366,10475,4,23000,100,23000,100,23000,100,23000,100,23000,100,23000,100,100,100,23000,100,19537,32,175354,32,46417,4,221973,511,0,1,89141,32,497525,14068,4,2,196500,453240,220,0,1,1,1000,28662,4,2,245000,216773,62,1,1060367,12586,1,208512,421,1,187000,1000,52998,1,80436,32,43249,32,1000,32,80556,1,57667,4,1000,10,197145,156,1,197145,156,1,204924,473,1,208896,511,1,52467,32,64832,32,65493,32,22558,32,16563,32,76511,32,196500,453240,220,0,1,1,69522,11687,0,1,60091,32,196500,453240,220,0,1,1,196500,453240,220,0,1,1,1159724,392670,0,2,806990,30482,4,1927926,82523,4,265318,0,4,0,85931,32,205665,812,1,1,41182,32,212342,32,31220,32,32696,32,43357,32,32247,32,38314,32,35892428,10,9462713,1021,10,38887044,32947,10]},"protocolVersion":{"major":7,"minor":0},"maxBlockHeaderSize":1100,"maxBlockBodySize":65536,"maxTxSize":16384,"txFeeFixed":155381,"txFeePerByte":44,"stakeAddressDeposit":0,"stakePoolDeposit":0,"minPoolCost":0,"poolRetireMaxEpoch":18,"stakePoolTargetNum":100,"poolPledgeInfluence":0,"monetaryExpansion":0.1,"treasuryCut":0.1,"collateralPercentage":150,"executionUnitPrices":{"priceMemory":0.0577,"priceSteps":0.0000721},"utxoCostPerByte":4310,"maxTxExecutionUnits":{"memory":16000000,"steps":10000000000},"maxBlockExecutionUnits":{"memory":80000000,"steps":40000000000},"maxCollateralInputs":3,"maxValueSize":5000,"extraPraosEntropy":null,"decentralization":null,"minUTxOValue":null}`)

	t.Run("GetAddressesAndAmountsToPayFrom 1 address native", func(t *testing.T) {
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
		}, hclog.NewNullLogger())

		amounts, err := coordinator.GetAddressesAndAmountsToPayFrom(chainID, cardanowallet.ResolveCardanoCliBinary(cardanowallet.TestNetNetwork), protocolParams, &[]cardanowallet.TxOutput{
			{
				Addr:   "addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpr",
				Amount: 10_000_000,
				Tokens: []cardanowallet.TokenAmount{
					cardanowallet.NewTokenAmount(token, 1000000),
				},
			},
		})
		require.NoError(t, err)
		require.Equal(t, "addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpr", amounts[0].Address)
		require.Equal(t, uint64(10000000), amounts[0].TokensAmounts[cardanowallet.AdaTokenName])
		require.Equal(t, uint64(1000000), amounts[0].IncludeChnage)
		require.Equal(t, uint64(1000000), amounts[0].TokensAmounts[token.String()])
		require.Equal(t, uint8(0), amounts[0].AddressIndex)
	})

	t.Run("GetAddressesAndAmountsToPayFrom 1 address native, not enough native funds", func(t *testing.T) {
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
		}, hclog.NewNullLogger())

		amounts, err := coordinator.GetAddressesAndAmountsToPayFrom(chainID, cardanowallet.ResolveCardanoCliBinary(cardanowallet.TestNetNetwork), protocolParams, &[]cardanowallet.TxOutput{
			{
				Addr:   "addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpr",
				Amount: 10_000_000,
				Tokens: []cardanowallet.TokenAmount{
					cardanowallet.NewTokenAmount(token, 10_000_000),
				},
			},
		})
		require.Error(t, err)
		require.Nil(t, amounts)
	})

	t.Run("GetAddressesAndAmountsToPayFrom 1 address native, not enough currency funds", func(t *testing.T) {
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
		}, hclog.NewNullLogger())

		amounts, err := coordinator.GetAddressesAndAmountsToPayFrom(chainID, cardanowallet.ResolveCardanoCliBinary(cardanowallet.TestNetNetwork), protocolParams, &[]cardanowallet.TxOutput{
			{
				Addr:   "addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpr",
				Amount: 10_000_000,
				Tokens: []cardanowallet.TokenAmount{
					cardanowallet.NewTokenAmount(token, 1000000),
				},
			},
		})

		require.Error(t, err)
		require.Nil(t, amounts)
	})

	t.Run("GetAddressesAndAmountsToPayFrom 1 address native 2 currency", func(t *testing.T) {
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
		}, hclog.NewNullLogger())

		amounts, err := coordinator.GetAddressesAndAmountsToPayFrom(chainID, cardanowallet.ResolveCardanoCliBinary(cardanowallet.TestNetNetwork), protocolParams, &[]cardanowallet.TxOutput{
			{
				Addr:   "addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpr",
				Amount: 10_000_000,
				Tokens: []cardanowallet.TokenAmount{
					cardanowallet.NewTokenAmount(token, 500_000),
				},
			},
		})

		require.NoError(t, err)
		require.Equal(t, uint64(500_000), amounts[0].TokensAmounts[token.String()])
		require.Equal(t, uint64(8_961_290), amounts[0].TokensAmounts[cardanowallet.AdaTokenName])
		require.Equal(t, uint64(1038710), amounts[0].IncludeChnage)
		require.Equal(t, uint64(1038710), amounts[1].TokensAmounts[cardanowallet.AdaTokenName])
		require.Equal(t, uint64(1000000), amounts[1].IncludeChnage)
	})

	t.Run("GetAddressesAndAmountsToPayFrom 2 address native 2 currency 2", func(t *testing.T) {
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
		}, hclog.NewNullLogger())

		amounts, err := coordinator.GetAddressesAndAmountsToPayFrom(chainID, cardanowallet.ResolveCardanoCliBinary(cardanowallet.TestNetNetwork), protocolParams, &[]cardanowallet.TxOutput{
			{
				Addr:   "addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpr",
				Amount: 10_000_000,
				Tokens: []cardanowallet.TokenAmount{
					cardanowallet.NewTokenAmount(token, 500_000),
				},
			},
		})

		require.NoError(t, err)
		require.Equal(t, uint64(500_000), amounts[0].TokensAmounts[token.String()])
		require.Equal(t, uint64(8961290), amounts[0].TokensAmounts[cardanowallet.AdaTokenName])
		require.Equal(t, uint64(1038710), amounts[0].IncludeChnage)
		require.Equal(t, uint64(1038710), amounts[1].TokensAmounts[cardanowallet.AdaTokenName])
		require.Equal(t, uint64(0), amounts[1].TokensAmounts[token.String()])
		require.Equal(t, uint64(1000000), amounts[1].IncludeChnage)
	})
}
