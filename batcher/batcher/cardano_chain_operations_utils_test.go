package batcher

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getTxOutputFromSumMap(t *testing.T) {
	const addr = "addr1_stokturist"

	token1 := cardanowallet.NewToken("34", "dju")
	token2 := cardanowallet.NewToken("35", "dju")

	result, err := getTxOutputFromSumMap(addr, map[string]uint64{
		cardanowallet.AdaTokenName: 151,
		token1.String():            10,
		token2.String():            348,
	})

	require.NoError(t, err)
	require.Equal(t, cardanowallet.TxOutput{
		Addr:   addr,
		Amount: 151,
		Tokens: []cardanowallet.TokenAmount{
			cardanowallet.NewTokenAmount(token1, 10),
			cardanowallet.NewTokenAmount(token2, 348),
		},
	}, result)
}

func Test_getNeededUtxos(t *testing.T) {
	const minUtxoAmount = 5

	desiredAmounts := map[string]uint64{
		cardanowallet.AdaTokenName: 0,
	}
	inputs := []*indexer.TxInputOutput{
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("01"), Index: 100},
			Output: indexer.TxOutput{Amount: 100},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("02"), Index: 0},
			Output: indexer.TxOutput{Amount: 50},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("03"), Index: 7},
			Output: indexer.TxOutput{Amount: 150},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("04"), Index: 5},
			Output: indexer.TxOutput{Amount: 200},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("05"), Index: 6},
			Output: indexer.TxOutput{Amount: 160},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("06"), Index: 8},
			Output: indexer.TxOutput{Amount: 400},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("07"), Index: 10},
			Output: indexer.TxOutput{Amount: 200},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("08"), Index: 9},
			Output: indexer.TxOutput{Amount: 50},
		},
	}

	t.Run("exact amount", func(t *testing.T) {
		desiredAmounts[cardanowallet.AdaTokenName] = 605
		result, err := getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 4, 1)

		require.NoError(t, err)
		require.Equal(t, []*indexer.TxInputOutput{inputs[0], inputs[2], inputs[3], inputs[4]}, result)

		desiredAmounts[cardanowallet.AdaTokenName] = 245
		result, err = getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 2, 1)

		require.NoError(t, err)
		require.Equal(t, []*indexer.TxInputOutput{inputs[0], inputs[2]}, result)
	})

	t.Run("pass with change", func(t *testing.T) {
		desiredAmounts[cardanowallet.AdaTokenName] = 706
		result, err := getNeededUtxos(inputs, desiredAmounts, 4, 3, 1)

		require.NoError(t, err)
		require.Equal(t, inputs[3:6], result)
	})

	t.Run("pass with at least", func(t *testing.T) {
		desiredAmounts[cardanowallet.AdaTokenName] = 5
		result, err := getNeededUtxos(inputs, desiredAmounts, 4, 30, 3)

		require.NoError(t, err)
		require.Equal(t, inputs[:3], result)
	})

	t.Run("not enough sum", func(t *testing.T) {
		desiredAmounts[cardanowallet.AdaTokenName] = 1550
		_, err := getNeededUtxos(inputs, desiredAmounts, 5, 30, 1)
		require.ErrorIs(t, err, cardanowallet.ErrUTXOsCouldNotSelect)
	})
}

func Test_getNeededSkylineUtxos(t *testing.T) {
	inputs := []*indexer.TxInputOutput{
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("01"), Index: 100},
			Output: indexer.TxOutput{Amount: 100},
		},
		{
			Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("02"), Index: 0},
			Output: indexer.TxOutput{
				Amount: 50,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "1",
						Name:     "1",
						Amount:   100,
					},
				},
			},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("03"), Index: 7},
			Output: indexer.TxOutput{Amount: 150},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("04"), Index: 5},
			Output: indexer.TxOutput{Amount: 200},
		},
		{
			Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("05"), Index: 6},
			Output: indexer.TxOutput{
				Amount: 160,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "1",
						Name:     "1",
						Amount:   50,
					},
				},
			},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("06"), Index: 8},
			Output: indexer.TxOutput{Amount: 400},
		},
		{
			Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("07"), Index: 10},
			Output: indexer.TxOutput{
				Amount: 200,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "1",
						Name:     "1",
						Amount:   400,
					},
				},
			},
		},
		{
			Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("08"), Index: 9},
			Output: indexer.TxOutput{
				Amount: 50,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "1",
						Name:     "1",
						Amount:   200,
					},
				},
			},
		},
	}

	t.Run("pass", func(t *testing.T) {
		const minUtxoAmount = 5

		desiredAmounts := map[string]uint64{
			cardanowallet.AdaTokenName: 590,
		}

		result, err := getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 4, 1)

		require.NoError(t, err)
		require.Equal(t, []*indexer.TxInputOutput{inputs[0], inputs[2], inputs[3], inputs[4]}, result)

		desiredAmounts = map[string]uint64{
			cardanowallet.AdaTokenName: minUtxoAmount,
			"1.31":                     100,
		}

		result, err = getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 1, 1)

		require.NoError(t, err)
		require.Equal(t, []*indexer.TxInputOutput{inputs[1]}, result)
	})

	t.Run("pass with change", func(t *testing.T) {
		const minUtxoAmount = 4

		desiredAmounts := map[string]uint64{
			cardanowallet.AdaTokenName: minUtxoAmount,
			"1.31":                     350,
		}
		result, err := getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 2, 1)

		require.NoError(t, err)
		require.Equal(t, []*indexer.TxInputOutput{inputs[1], inputs[6]}, result)
	})

	t.Run("pass with at least", func(t *testing.T) {
		const minUtxoAmount = 4

		desiredAmounts := map[string]uint64{
			cardanowallet.AdaTokenName: minUtxoAmount,
			"1.31":                     20,
		}
		result, err := getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 30, 3)

		require.NoError(t, err)
		require.Equal(t, inputs[:3], result)

		desiredAmounts = map[string]uint64{
			cardanowallet.AdaTokenName: 12,
		}
		result, err = getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 30, 3)

		require.NoError(t, err)
		require.Equal(t, inputs[:3], result)
	})

	t.Run("not enough sum", func(t *testing.T) {
		const minUtxoAmount = 5

		desiredAmounts := map[string]uint64{
			cardanowallet.AdaTokenName: 1600,
		}
		_, err := getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 30, 1)
		require.ErrorIs(t, err, cardanowallet.ErrUTXOsCouldNotSelect)

		desiredAmounts = map[string]uint64{
			cardanowallet.AdaTokenName: minUtxoAmount,
			"1.31":                     2500,
		}
		_, err = getNeededUtxos(inputs, desiredAmounts, minUtxoAmount, 30, 1)
		require.ErrorIs(t, err, cardanowallet.ErrUTXOsCouldNotSelect)
	})
}

func Test_reactorGetOutputs(t *testing.T) {
	configRaw := json.RawMessage([]byte(`{
			"socketPath": "./socket",
			"testnetMagic": 42,
			"minUtxoAmount": 1000
			}`))

	cardanoConfig, err := cardano.NewCardanoChainConfig(configRaw)
	require.NoError(t, err)

	cco := &CardanoChainOperations{
		config: cardanoConfig,
	}
	cco.config.NetworkID = cardanowallet.MainNetNetwork

	txs := []eth.ConfirmedTransaction{
		{
			TransactionType:    uint8(common.StakeConfirmedTxType),
			TransactionSubType: uint8(common.StakeRegDelConfirmedTxSubType),
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
					Amount:             big.NewInt(100),
				},
				{
					DestinationAddress: "addr128phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtupnz75xxcrtw79hu",
					Amount:             big.NewInt(200),
				},
				{
					DestinationAddress: "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8",
					Amount:             big.NewInt(400),
				},
			},
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "addr1w8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcyjy7wx",
					Amount:             big.NewInt(50),
				},
				{
					DestinationAddress: "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8",
					Amount:             big.NewInt(900),
				},
				{
					DestinationAddress: "addr1z8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gten0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgs9yc0hh",
					Amount:             big.NewInt(0),
				},
			},
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "addr1qx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer3n0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgse35a3x",
					Amount:             big.NewInt(3000),
				},
				{
					// this one will be skipped
					DestinationAddress: "stake178phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcccycj5",
					Amount:             big.NewInt(3000),
				},
			},
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
					Amount:             big.NewInt(2000),
				},
				{
					DestinationAddress: "addr1w8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcyjy7wx",
					Amount:             big.NewInt(170),
				},
				{
					DestinationAddress: "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8",
					Amount:             big.NewInt(10),
				},
			},
		},
	}

	res, isRedistribution, err := getOutputs(txs, cco.config, hclog.NewNullLogger())
	require.NoError(t, err)

	assert.False(t, isRedistribution)
	assert.Equal(t, uint64(6830), res.Sum[cardanowallet.AdaTokenName])
	assert.Equal(t, []cardanowallet.TxOutput{
		{
			Addr:   "addr128phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtupnz75xxcrtw79hu",
			Amount: 200,
		},
		{
			Addr:   "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
			Amount: 2100,
		},
		{
			Addr:   "addr1qx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer3n0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgse35a3x",
			Amount: 3000,
		},
		{
			Addr:   "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8",
			Amount: 1310,
		},
		{
			Addr:   "addr1w8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcyjy7wx",
			Amount: 220,
		},
	}, res.Outputs)
}

func Test_skylineGetOutputs(t *testing.T) {
	// cardano -> prime
	const (
		addr1 = "addr_test1yz2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzerkr0vd4msrxnuwnccdxlhdjar77j6lg0wypcc9uar5d2shsf5r8qx"
		addr2 = "addr_test1qz2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer3n0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgs68faae"
		addr3 = "addr_test1zrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gten0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgsxj90mg"
	)

	policyID := "584ffccecba8a7c6a18037152119907b6b5c2ed063798ee68b012c41"
	tokenName, _ := hex.DecodeString("526f75746533")
	token := cardanowallet.NewToken(policyID, string(tokenName))
	config := &cardano.CardanoChainConfig{
		NetworkID: cardanowallet.TestNetNetwork,
		NativeTokens: []sendtx.TokenExchangeConfig{
			{
				DstChainID: common.ChainIDStrCardano,
				TokenName:  token.String(),
			},
		},
	}

	txs := []eth.ConfirmedTransaction{
		{
			TransactionType:    uint8(common.StakeConfirmedTxType),
			TransactionSubType: uint8(common.StakeRegDelConfirmedTxSubType),
		},
		{
			SourceChainId: common.ChainIDIntCardano,
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: addr1,
					Amount:             big.NewInt(100),
					AmountWrapped:      big.NewInt(200),
				},
				{
					DestinationAddress: addr2,
					Amount:             big.NewInt(51),
					AmountWrapped:      big.NewInt(102),
				},
			},
		},
		{
			SourceChainId: common.ChainIDIntCardano,
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: addr3,
					Amount:             big.NewInt(8),
				},
				{
					DestinationAddress: addr1,
					Amount:             big.NewInt(2),
					AmountWrapped:      big.NewInt(5),
				},
			},
		},
	}

	outputs, isRedistribution, err := getOutputs(txs, config, hclog.NewNullLogger())
	require.NoError(t, err)

	assert.False(t, isRedistribution)
	require.Equal(t, []cardanowallet.TxOutput{
		{
			Addr:   addr2,
			Amount: 51,
			Tokens: []cardanowallet.TokenAmount{
				cardanowallet.NewTokenAmount(token, 102),
			},
		},
		{
			Addr:   addr1,
			Amount: 102,
			Tokens: []cardanowallet.TokenAmount{
				cardanowallet.NewTokenAmount(token, 205),
			},
		},
		{
			Addr:   addr3,
			Amount: 8,
		},
	}, outputs.Outputs)
	require.Len(t, outputs.Sum, 2)
	require.Equal(t, uint64(307), outputs.Sum[token.String()])
	require.Equal(t, uint64(161), outputs.Sum[cardanowallet.AdaTokenName])

	t.Run("GetOutputs with redistribute tokens transaction", func(t *testing.T) {
		txs = append(txs, eth.ConfirmedTransaction{
			TransactionType: uint8(common.RedistributionConfirmedTxType),
		})

		outputs, isRedistribution, err := getOutputs(txs, config, hclog.NewNullLogger())
		require.NoError(t, err)

		assert.True(t, isRedistribution)
		require.Equal(t, []cardanowallet.TxOutput{
			{
				Addr:   addr2,
				Amount: 51,
				Tokens: []cardanowallet.TokenAmount{
					cardanowallet.NewTokenAmount(token, 102),
				},
			},
			{
				Addr:   addr1,
				Amount: 102,
				Tokens: []cardanowallet.TokenAmount{
					cardanowallet.NewTokenAmount(token, 205),
				},
			},
			{
				Addr:   addr3,
				Amount: 8,
			},
		}, outputs.Outputs)
		require.Len(t, outputs.Sum, 2)
		require.Equal(t, uint64(307), outputs.Sum[token.String()])
		require.Equal(t, uint64(161), outputs.Sum[cardanowallet.AdaTokenName])
	})
}

func Test_getSkylineUTXOs(t *testing.T) {
	sumMap := map[string]uint64{
		cardanowallet.AdaTokenName: 60,
		"1.31":                     60,
	}

	t.Run("pass", func(t *testing.T) {
		expectedUtxos := []*indexer.TxInputOutput{
			{
				Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("01"), Index: 2},
				Output: indexer.TxOutput{
					Amount: 30,
					Slot:   80,
					Tokens: []indexer.TokenAmount{
						{
							PolicyID: "1",
							Name:     "1",
							Amount:   40,
						},
					},
				},
			},
			{
				Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("01"), Index: 3},
				Output: indexer.TxOutput{
					Amount: 40,
					Slot:   1900,
					Tokens: []indexer.TokenAmount{
						{
							PolicyID: "1",
							Name:     "1",
							Amount:   30,
						},
					},
				},
			},
			{
				Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("AA"), Index: 100},
				Output: indexer.TxOutput{
					Amount: 10,
				},
			},
		}

		multisigUtxos, err := getNeededUtxos(
			expectedUtxos, sumMap, 0, 50, 1)

		require.NoError(t, err)
		require.Equal(t, expectedUtxos[0:2], multisigUtxos)
	})
}

func Test_extractStakeKeyDepositAmount(t *testing.T) {
	protocolParams := []byte(`{"costModels":{"PlutusV1":[197209,0,1,1,396231,621,0,1,150000,1000,0,1,150000,32,2477736,29175,4,29773,100,29773,100,29773,100,29773,100,29773,100,29773,100,100,100,29773,100,150000,32,150000,32,150000,32,150000,1000,0,1,150000,32,150000,1000,0,8,148000,425507,118,0,1,1,150000,1000,0,8,150000,112536,247,1,150000,10000,1,136542,1326,1,1000,150000,1000,1,150000,32,150000,32,150000,32,1,1,150000,1,150000,4,103599,248,1,103599,248,1,145276,1366,1,179690,497,1,150000,32,150000,32,150000,32,150000,32,150000,32,150000,32,148000,425507,118,0,1,1,61516,11218,0,1,150000,32,148000,425507,118,0,1,1,148000,425507,118,0,1,1,2477736,29175,4,0,82363,4,150000,5000,0,1,150000,32,197209,0,1,1,150000,32,150000,32,150000,32,150000,32,150000,32,150000,32,150000,32,3345831,1,1],"PlutusV2":[205665,812,1,1,1000,571,0,1,1000,24177,4,1,1000,32,117366,10475,4,23000,100,23000,100,23000,100,23000,100,23000,100,23000,100,100,100,23000,100,19537,32,175354,32,46417,4,221973,511,0,1,89141,32,497525,14068,4,2,196500,453240,220,0,1,1,1000,28662,4,2,245000,216773,62,1,1060367,12586,1,208512,421,1,187000,1000,52998,1,80436,32,43249,32,1000,32,80556,1,57667,4,1000,10,197145,156,1,197145,156,1,204924,473,1,208896,511,1,52467,32,64832,32,65493,32,22558,32,16563,32,76511,32,196500,453240,220,0,1,1,69522,11687,0,1,60091,32,196500,453240,220,0,1,1,196500,453240,220,0,1,1,1159724,392670,0,2,806990,30482,4,1927926,82523,4,265318,0,4,0,85931,32,205665,812,1,1,41182,32,212342,32,31220,32,32696,32,43357,32,32247,32,38314,32,35892428,10,9462713,1021,10,38887044,32947,10]},"protocolVersion":{"major":7,"minor":0},"maxBlockHeaderSize":1100,"maxBlockBodySize":65536,"maxTxSize":16384,"txFeeFixed":155381,"txFeePerByte":44,"stakeAddressDeposit":0,"stakePoolDeposit":0,"minPoolCost":0,"poolRetireMaxEpoch":18,"stakePoolTargetNum":100,"poolPledgeInfluence":0,"monetaryExpansion":0.1,"treasuryCut":0.1,"collateralPercentage":150,"executionUnitPrices":{"priceMemory":0.0577,"priceSteps":0.0000721},"utxoCostPerByte":4310,"maxTxExecutionUnits":{"memory":16000000,"steps":10000000000},"maxBlockExecutionUnits":{"memory":80000000,"steps":40000000000},"maxCollateralInputs":3,"maxValueSize":5000,"extraPraosEntropy":null,"decentralization":null,"minUTxOValue":null}`)

	amount, err := extractStakeKeyDepositAmount(protocolParams)
	require.NoError(t, err)
	require.Equal(t, uint64(0), amount)
}

func Test_allocateInputsForConsolidation(t *testing.T) {
	getUtxos := func(count int) []*indexer.TxInputOutput {
		outputs, _ := generateSmallUtxoOutputs(10, uint64(count))

		return outputs
	}

	t.Run("total < max", func(t *testing.T) {
		inputs := []AddressConsolidationData{
			{Address: "addr1", AddressIndex: 0, Utxos: getUtxos(5), UtxoCount: 5},
			{Address: "addr2", AddressIndex: 1, Utxos: getUtxos(10), UtxoCount: 10},
			{Address: "addr3", AddressIndex: 2, Utxos: getUtxos(20), UtxoCount: 20},
		}

		alloc := allocateInputsForConsolidation(inputs, 50, 35)
		require.Equal(t, inputs, alloc)
	})

	t.Run("total == max", func(t *testing.T) {
		inputs := []AddressConsolidationData{
			{Address: "addr1", AddressIndex: 0, UtxoCount: 10, Utxos: getUtxos(10)},
			{Address: "addr2", AddressIndex: 1, UtxoCount: 20, Utxos: getUtxos(20)},
			{Address: "addr3", AddressIndex: 2, UtxoCount: 20, Utxos: getUtxos(20)},
		}

		alloc := allocateInputsForConsolidation(inputs, 50, 50)
		require.Equal(t, inputs, alloc)
	})

	t.Run("total > max", func(t *testing.T) {
		inputs := []AddressConsolidationData{
			{Address: "addr1", AddressIndex: 0, UtxoCount: 10, Utxos: getUtxos(10)},
			{Address: "addr2", AddressIndex: 1, UtxoCount: 20, Utxos: getUtxos(20)},
			{Address: "addr3", AddressIndex: 2, UtxoCount: 30, Utxos: getUtxos(30)},
		}

		alloc := allocateInputsForConsolidation(inputs, 50, 60)
		require.Equal(t, []AddressConsolidationData{
			{Address: "addr2", AddressIndex: 1, UtxoCount: 17, Utxos: inputs[1].Utxos[:17]},
			{Address: "addr1", AddressIndex: 0, UtxoCount: 8, Utxos: inputs[0].Utxos[:8]},
			{Address: "addr3", AddressIndex: 2, UtxoCount: 25, Utxos: inputs[2].Utxos[:25]},
		}, alloc)
	})

	t.Run("total >> max", func(t *testing.T) {
		inputs := []AddressConsolidationData{
			{Address: "addr1", AddressIndex: 0, UtxoCount: 1, Utxos: getUtxos(1)},
			{Address: "addr2", AddressIndex: 1, UtxoCount: 9, Utxos: getUtxos(9)},
		}

		alloc := allocateInputsForConsolidation(inputs, 2, 9)
		require.Equal(t, []AddressConsolidationData{
			{Address: "addr2", AddressIndex: 1, UtxoCount: 2, Utxos: inputs[1].Utxos[:2]},
		}, alloc)
	})

	t.Run("1 utxo in output fix", func(t *testing.T) {
		inputs := []AddressConsolidationData{
			{Address: "addr1", AddressIndex: 0, UtxoCount: 2, Utxos: getUtxos(1)},
			{Address: "addr2", AddressIndex: 1, UtxoCount: 3, Utxos: getUtxos(3)},
		}

		alloc := allocateInputsForConsolidation(inputs, 3, 5)
		require.Equal(t, []AddressConsolidationData{
			{Address: "addr2", AddressIndex: 1, UtxoCount: 3, Utxos: inputs[1].Utxos},
		}, alloc)
	})
}
