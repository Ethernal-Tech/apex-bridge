package batcher

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func splitTokenAmount(name string, isNameEncoded bool) (string, string, error) {
	parts := strings.Split(name, ".")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid full token name: %s", name)
	}

	if !isNameEncoded {
		name = parts[1]
	} else {
		decodedName, err := hex.DecodeString(parts[1])
		if err != nil {
			return "", "", fmt.Errorf("invalid full token name: %s", name)
		}

		name = string(decodedName)
	}

	return parts[0], name, nil
}

func TestSkylineCardanoChainOperations_IsSynchronized(t *testing.T) {
	chainID := "prime"
	dbMock := &indexer.DatabaseMock{}
	bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
	ctx := context.Background()
	scBlock1 := eth.CardanoBlock{
		BlockSlot: big.NewInt(15),
	}
	scBlock2 := eth.CardanoBlock{
		BlockSlot: big.NewInt(20),
	}
	oracleBlock1 := &indexer.BlockPoint{
		BlockSlot: uint64(10),
	}
	oracleBlock2 := &indexer.BlockPoint{
		BlockSlot: uint64(20),
	}
	testErr1 := errors.New("test error 1")
	testErr2 := errors.New("test error 2")

	bridgeSmartContractMock.On("GetLastObservedBlock", ctx, chainID).Return(eth.CardanoBlock{}, testErr1).Once()
	bridgeSmartContractMock.On("GetLastObservedBlock", ctx, chainID).Return(eth.CardanoBlock{}, nil).Once()
	bridgeSmartContractMock.On("GetLastObservedBlock", ctx, chainID).Return(scBlock1, nil).Times(3)
	bridgeSmartContractMock.On("GetLastObservedBlock", ctx, chainID).Return(scBlock2, nil).Once()

	dbMock.On("GetLatestBlockPoint").Return((*indexer.BlockPoint)(nil), testErr2).Once()
	dbMock.On("GetLatestBlockPoint").Return(oracleBlock1, nil).Once()
	dbMock.On("GetLatestBlockPoint").Return(oracleBlock2, nil).Twice()

	scco := &SkylineCardanoChainOperations{
		db:     dbMock,
		logger: hclog.NewNullLogger(),
	}

	// sc error
	_, err := scco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.ErrorIs(t, err, testErr1)

	// database error
	_, err = scco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.ErrorIs(t, err, testErr2)

	// not in sync
	val, err := scco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.NoError(t, err)
	require.False(t, val)

	// in sync
	val, err = scco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.NoError(t, err)
	require.True(t, val)

	// in sync again
	val, err = scco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.NoError(t, err)
	require.True(t, val)
}

func Test_getSkylineOutputs(t *testing.T) {
	cardanoPrimeWrappedTokenName := "72f3d1e6c885e4d0bdcf5250513778dbaa851c0b4bfe3ed4e1bcceb0.4b6173685f546f6b656e"
	primeCardanoWrappedTokenName := "29f8873beb52e126f207a2dfd50f7cff556806b5b4cba9834a7b26a8.526f75746533"
	ccPrimeTokenExchange := []cardanotx.CardanoConfigTokenExchange{
		{
			Chain:        common.ChainIDStrCardano,
			SrcTokenName: cardanowallet.AdaTokenName,
			DstTokenName: cardanoPrimeWrappedTokenName,
		},
		{
			Chain:        common.ChainIDStrCardano,
			SrcTokenName: primeCardanoWrappedTokenName,
			DstTokenName: cardanowallet.AdaTokenName,
		},
	}

	_ = ccPrimeTokenExchange

	ccCardanoTokenExchange := []cardanotx.CardanoConfigTokenExchange{
		{
			Chain:        common.ChainIDStrPrime,
			SrcTokenName: cardanowallet.AdaTokenName,
			DstTokenName: primeCardanoWrappedTokenName,
		},
		{
			Chain:        common.ChainIDStrPrime,
			SrcTokenName: cardanoPrimeWrappedTokenName,
			DstTokenName: cardanowallet.AdaTokenName,
		},
	}
	txs := []eth.ConfirmedTransaction{
		{
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
					Amount:             big.NewInt(100),
					AmountWrapped:      big.NewInt(10),
				},
				{
					DestinationAddress: "addr128phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtupnz75xxcrtw79hu",
					Amount:             big.NewInt(200),
					AmountWrapped:      big.NewInt(20),
				},
				{
					DestinationAddress: "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8",
					Amount:             big.NewInt(400),
					AmountWrapped:      big.NewInt(0),
				},
			},
			SourceChainId: 4,
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8",
					Amount:             big.NewInt(900),
					AmountWrapped:      big.NewInt(80),
				},
			},
			SourceChainId: 4,
		},
	}

	polID, tName, _ := splitTokenAmount(primeCardanoWrappedTokenName, true)
	res, err := getSkylineOutputs(txs, ccCardanoTokenExchange, common.ChainIDStrPrime, cardanowallet.MainNetNetwork, hclog.NewNullLogger())
	assert.NoError(t, err)

	assert.Equal(t, map[string]uint64{
		fmt.Sprintf("%s", cardanowallet.AdaTokenName):   1600,
		fmt.Sprintf("%s", primeCardanoWrappedTokenName): 110,
	}, res.Sum)

	assert.Equal(t, []cardanowallet.TxOutput{
		{
			Addr:   "addr128phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtupnz75xxcrtw79hu",
			Amount: 200,
			Tokens: []cardanowallet.TokenAmount{
				{
					PolicyID: polID,
					Name:     tName,
					Amount:   20,
				},
			},
		},
		{
			Addr:   "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
			Amount: 100,
			Tokens: []cardanowallet.TokenAmount{
				{
					PolicyID: polID,
					Name:     tName,
					Amount:   10,
				},
			},
		},
		{
			Addr:   "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8",
			Amount: 1300,
			Tokens: []cardanowallet.TokenAmount{
				{
					PolicyID: polID,
					Name:     tName,
					Amount:   80,
				},
			},
		},
	}, res.Outputs)

	txs = []eth.ConfirmedTransaction{
		{
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "addr1qx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer3n0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgse35a3x",
					Amount:             big.NewInt(3000),
					AmountWrapped:      big.NewInt(200),
				},
				{
					DestinationAddress: "addr1z8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gten0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgs9yc0hh",
					Amount:             big.NewInt(0),
					AmountWrapped:      big.NewInt(0),
				},
				{
					// this one will be skipped
					DestinationAddress: "stake178phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcccycj5",
					Amount:             big.NewInt(3000),
					AmountWrapped:      big.NewInt(1300),
				},
			},
			SourceChainId: 1,
		},
		{
			Receivers: []eth.BridgeReceiver{
				{
					DestinationAddress: "addr1w8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcyjy7wx",
					Amount:             big.NewInt(170),
					AmountWrapped:      big.NewInt(50),
				},
			},
			SourceChainId: 1,
		},
	}

	polID, tName, _ = splitTokenAmount(cardanoPrimeWrappedTokenName, true)
	res, err = getSkylineOutputs(txs, ccPrimeTokenExchange, common.ChainIDStrCardano, cardanowallet.MainNetNetwork, hclog.NewNullLogger())
	assert.NoError(t, err)
	assert.Equal(t, map[string]uint64{
		fmt.Sprintf("%s", cardanowallet.AdaTokenName):   3170,
		fmt.Sprintf("%s", cardanoPrimeWrappedTokenName): 250,
	}, res.Sum)

	assert.Equal(t, []cardanowallet.TxOutput{
		{
			Addr:   "addr1qx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer3n0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgse35a3x",
			Amount: 3000,
			Tokens: []cardanowallet.TokenAmount{
				{
					PolicyID: polID,
					Name:     tName,
					Amount:   200,
				},
			},
		},
		{
			Addr:   "addr1w8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcyjy7wx",
			Amount: 170,
			Tokens: []cardanowallet.TokenAmount{
				{
					PolicyID: polID,
					Name:     tName,
					Amount:   50,
				},
			},
		},
	}, res.Outputs)
}
