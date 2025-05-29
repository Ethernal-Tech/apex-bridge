package batcher

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"slices"
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

func Test_subtractTxOutputsFromSumMap(t *testing.T) {
	tok1, err := cardano.GetNativeTokenFromName("3.31")
	require.NoError(t, err)

	tok2, err := cardano.GetNativeTokenFromName("3.32")
	require.NoError(t, err)

	tok3, err := cardano.GetNativeTokenFromName("3.33")
	require.NoError(t, err)

	tok4, err := cardano.GetNativeTokenFromName("3.34")
	require.NoError(t, err)

	vals := subtractTxOutputsFromSumMap(map[string]uint64{
		cardanowallet.AdaTokenName: 200,
		tok1.String():              400,
		tok2.String():              500,
		tok4.String():              1000,
	}, []cardanowallet.TxOutput{
		cardanowallet.NewTxOutput("", 100, cardanowallet.NewTokenAmount(tok1, 200), cardanowallet.NewTokenAmount(tok2, 205)),
		cardanowallet.NewTxOutput("", 50, cardanowallet.NewTokenAmount(tok1, 150), cardanowallet.NewTokenAmount(tok3, 300)),
		cardanowallet.NewTxOutput("", 10, cardanowallet.NewTokenAmount(tok2, 300)),
	})

	require.Equal(t, map[string]uint64{
		cardanowallet.AdaTokenName: 40,
		tok1.String():              50,
		tok4.String():              1000,
	}, vals)
}

func Test_filterOutTokenUtxos(t *testing.T) {
	multisigUtxos := []*indexer.TxInputOutput{
		{
			Input: indexer.TxInput{Index: 0},
			Output: indexer.TxOutput{
				Amount: 30,
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
			Input: indexer.TxInput{Index: 1},
			Output: indexer.TxOutput{
				Amount: 40,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "3",
						Name:     "1",
						Amount:   30,
					},
					{
						PolicyID: "1",
						Name:     "2",
						Amount:   30,
					},
				},
			},
		},
		{
			Input: indexer.TxInput{Index: 2},
			Output: indexer.TxOutput{
				Amount: 50,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "1",
						Name:     "1",
						Amount:   51,
					},
					{
						PolicyID: "3",
						Name:     "1",
						Amount:   21,
					},
				},
			},
		},
		{
			Input: indexer.TxInput{Index: 3},
			Output: indexer.TxOutput{
				Amount: 2,
				Tokens: []indexer.TokenAmount{
					{
						PolicyID: "3",
						Name:     "1",
						Amount:   7,
					},
				},
			},
		},
	}

	t.Run("filter out all the tokens", func(t *testing.T) {
		resTxInputOutput := filterOutUtxosWithUnknownTokens(multisigUtxos)
		require.Len(t, resTxInputOutput, 0)
	})

	t.Run("filter out all the tokens except the one with specified token name", func(t *testing.T) {
		tok, err := cardano.GetNativeTokenFromName("1.31")
		require.NoError(t, err)

		resTxInputOutput := filterOutUtxosWithUnknownTokens(multisigUtxos, tok)
		require.Len(t, resTxInputOutput, 1)
		require.Equal(
			t,
			indexer.TxInput{Index: 0},
			resTxInputOutput[0].Input,
		)
	})

	t.Run("filter out InputOutput with invalid token even if it contains valid token as well", func(t *testing.T) {
		tok, err := cardano.GetNativeTokenFromName("3.31")
		require.NoError(t, err)

		resTxInputOutput := filterOutUtxosWithUnknownTokens(multisigUtxos, tok)
		require.Len(t, resTxInputOutput, 1)
		require.Equal(
			t,
			indexer.TxInput{Index: 3},
			resTxInputOutput[0].Input,
		)
	})

	t.Run("filter out all the tokens except those with specified token names", func(t *testing.T) {
		tok1, err := cardano.GetNativeTokenFromName("3.31")
		require.NoError(t, err)

		tok2, err := cardano.GetNativeTokenFromName("1.31")
		require.NoError(t, err)

		resTxInputOutput := filterOutUtxosWithUnknownTokens(multisigUtxos, tok1, tok2)
		require.Len(t, resTxInputOutput, 3)
		require.Equal(
			t,
			indexer.TxInput{Index: 0},
			resTxInputOutput[0].Input,
		)
		require.Equal(
			t,
			2,
			len(resTxInputOutput[1].Output.Tokens),
		)
	})
}

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
			"minUtxoAmount": 1000,
			"minFeeForBridging": 100
			}`))

	feeAddr := "0x002"

	cardanoConfig, err := cardano.NewCardanoChainConfig(configRaw)
	require.NoError(t, err)

	cco := &CardanoChainOperations{
		config: cardanoConfig,
	}
	cco.config.NetworkID = cardanowallet.MainNetNetwork

	//nolint:dupl
	txs := []eth.ConfirmedTransaction{
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

	t.Run("getOutputs pass", func(t *testing.T) {
		res := getOutputs(txs, cco.config,
			[][]*indexer.TxInputOutput{}, "", common.ChainIDStrPrime, hclog.NewNullLogger())

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
	})

	t.Run("getOutputs with refund pass", func(t *testing.T) {
		refundTxAmount := uint64(300)
		txs := append(slices.Clone(txs), eth.ConfirmedTransaction{
			TransactionType: uint8(common.RefundConfirmedTxType),
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
					Amount:             new(big.Int).SetUint64(refundTxAmount),
				},
			},
		})

		refundUtxos := make([][]*indexer.TxInputOutput, len(txs))
		refundUtxos[len(refundUtxos)-1] = []*indexer.TxInputOutput{
			{
				Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0x1"), Index: 3},
				Output: indexer.TxOutput{Amount: 250},
			},
			{
				Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0x2"), Index: 2},
				Output: indexer.TxOutput{Amount: 50},
			},
		}

		res := getOutputs(txs, cco.config,
			refundUtxos, feeAddr, common.ChainIDStrPrime, hclog.NewNullLogger())

		assert.Equal(t, uint64(7030), res.Sum[cardanowallet.AdaTokenName])
		assert.Equal(t, []cardanowallet.TxOutput{
			{
				Addr:   "addr128phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtupnz75xxcrtw79hu",
				Amount: 200,
			},
			{
				Addr:   "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
				Amount: 2100 + refundTxAmount - cco.config.MinFeeForBridging,
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
	})

	t.Run("getOutputs with refund pass with tokens", func(t *testing.T) {
		refundTxAmount := uint64(300)
		txs = append(slices.Clone(txs), eth.ConfirmedTransaction{
			TransactionType: uint8(common.RefundConfirmedTxType),
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
					Amount:             new(big.Int).SetUint64(refundTxAmount),
				},
			},
		})

		refundUtxos := make([][]*indexer.TxInputOutput, len(txs))
		refundUtxos[len(refundUtxos)-1] = []*indexer.TxInputOutput{
			{
				Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("0x1"), Index: 3},
				Output: indexer.TxOutput{
					Amount: 200,
					Tokens: []indexer.TokenAmount{
						{
							PolicyID: "1",
							Name:     "1",
							Amount:   15,
						},
					},
				},
			},
			{
				Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("0x21"), Index: 2},
				Output: indexer.TxOutput{
					Amount: 100,
					Tokens: []indexer.TokenAmount{
						{
							PolicyID: "1",
							Name:     "3",
							Amount:   15,
						},
					},
				},
			},
		}

		res := getOutputs(txs, cco.config,
			refundUtxos, feeAddr, common.ChainIDStrPrime, hclog.NewNullLogger())

		assert.Equal(t, uint64(7030), res.Sum[cardanowallet.AdaTokenName])
		assert.Equal(t, []cardanowallet.TxOutput{
			{
				Addr:   "addr128phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtupnz75xxcrtw79hu",
				Amount: 200,
			},
			{
				Addr:   "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
				Amount: 2100 + refundTxAmount - cco.config.MinFeeForBridging,
				Tokens: []cardanowallet.TokenAmount{
					cardanowallet.NewTokenAmount(cardanowallet.NewToken("1", "1"), 15),
					cardanowallet.NewTokenAmount(cardanowallet.NewToken("1", "3"), 15),
				},
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
	})
}

func Test_skylineGetOutputs(t *testing.T) {
	// prime -> cardano
	const (
		addr1 = "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k"
		addr2 = "addr128phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtupnz75xxcrtw79hu"
		addr3 = "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8"
		addr4 = "addr1qx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer3n0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgse35a3x"
		addr5 = "addr1w8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcyjy7wx"
	)

	feeAddr := "0x002"

	policyID := "584ffccecba8a7c6a18037152119907b6b5c2ed063798ee68b012c41"
	tokenName, _ := hex.DecodeString("526f75746533")
	token := cardanowallet.NewToken(policyID, string(tokenName))
	config := &cardano.CardanoChainConfig{
		NetworkID: cardanowallet.MainNetNetwork,
		NativeTokens: []sendtx.TokenExchangeConfig{
			{
				DstChainID: common.ChainIDStrCardano,
				TokenName:  token.String(),
			},
		},
		MinFeeForBridging: 100,
	}

	txs := []eth.ConfirmedTransaction{
		{
			SourceChainId: common.ChainIDIntPrime,
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
			SourceChainId: common.ChainIDIntPrime,
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

	t.Run("getOutputs pass", func(t *testing.T) {
		outputs := getOutputs(txs, config, [][]*indexer.TxInputOutput{}, addr1, common.ChainIDStrCardano, hclog.NewNullLogger())

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

	t.Run("getOutputs with refund pass", func(t *testing.T) {
		refundTxAmount := uint64(300)
		refundTxWrappedAmount := uint64(500)
		txs := append(slices.Clone(txs), eth.ConfirmedTransaction{
			TransactionType: uint8(common.RefundConfirmedTxType),
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: addr1,
					Amount:             new(big.Int).SetUint64(refundTxAmount),
					AmountWrapped:      new(big.Int).SetUint64(refundTxWrappedAmount),
				},
			},
		})

		refundUtxos := make([][]*indexer.TxInputOutput, len(txs))
		refundUtxos[len(refundUtxos)-1] = []*indexer.TxInputOutput{
			{
				Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("0x1"), Index: 3},
				Output: indexer.TxOutput{
					Amount: 250,
					Tokens: []indexer.TokenAmount{
						{
							PolicyID: token.PolicyID,
							Name:     token.Name,
							Amount:   300,
						},
					},
				},
			},
			{
				Input: indexer.TxInput{Hash: indexer.NewHashFromHexString("0x2"), Index: 2},
				Output: indexer.TxOutput{
					Amount: 50,
					Tokens: []indexer.TokenAmount{
						{
							PolicyID: token.PolicyID,
							Name:     token.Name,
							Amount:   200,
						},
					},
				},
			},
		}

		outputs := getOutputs(txs, config,
			refundUtxos, feeAddr, common.ChainIDStrCardano, hclog.NewNullLogger())

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
				Amount: 102 + refundTxAmount - config.MinFeeForBridging,
				Tokens: []cardanowallet.TokenAmount{
					cardanowallet.NewTokenAmount(token, 205+refundTxWrappedAmount),
				},
			},
			{
				Addr:   addr3,
				Amount: 8,
			},
		}, outputs.Outputs)
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
