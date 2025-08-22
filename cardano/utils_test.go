package cardanotx

import (
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_IsValidOutputAddress(t *testing.T) {
	listValidMain := []string{
		"addr1qx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer3n0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgse35a3x",
		"addr1z8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gten0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgs9yc0hh",
		"addr1yx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzerkr0vd4msrxnuwnccdxlhdjar77j6lg0wypcc9uar5d2shs2z78ve",
		"addr1x8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gt7r0vd4msrxnuwnccdxlhdjar77j6lg0wypcc9uar5d2shskhj42g",
		"addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
		"addr128phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtupnz75xxcrtw79hu",
		"addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8",
		"addr1w8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcyjy7wx",
	}
	listValidTest := []string{
		"addr_test1qz2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer3n0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgs68faae",
		"addr_test1zrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gten0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgsxj90mg",
		"addr_test1yz2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzerkr0vd4msrxnuwnccdxlhdjar77j6lg0wypcc9uar5d2shsf5r8qx",
		"addr_test1xrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gt7r0vd4msrxnuwnccdxlhdjar77j6lg0wypcc9uar5d2shs4p04xh",
		"addr_test1gz2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrdw5vky",
		"addr_test12rphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtupnz75xxcryqrvmw",
		"addr_test1vz2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzerspjrlsz",
		"addr_test1wrphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcl6szpr",
	}
	listInvalid := []string{
		"stake1uyehkck0lajq8gr28t9uxnuvgcqrc6070x3k9r8048z8y5gh6ffgw",
		"stake178phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcccycj5",
		"stake_test1uqehkck0lajq8gr28t9uxnuvgcqrc6070x3k9r8048z8y5gssrtvn",
		"stake_test17rphkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcljw6kf",
		"addr1dummy",
	}

	// mainnet
	for _, x := range listValidMain {
		assert.True(t, IsValidOutputAddress(x, wallet.MainNetNetwork))
	}

	for _, x := range listValidTest {
		assert.False(t, IsValidOutputAddress(x, wallet.MainNetNetwork))
	}

	for _, x := range listInvalid {
		assert.False(t, IsValidOutputAddress(x, wallet.MainNetNetwork))
	}

	// test
	for _, x := range listValidMain {
		assert.False(t, IsValidOutputAddress(x, wallet.TestNetNetwork))
	}

	for _, x := range listValidTest {
		assert.True(t, IsValidOutputAddress(x, wallet.TestNetNetwork))
	}

	for _, x := range listInvalid {
		assert.False(t, IsValidOutputAddress(x, wallet.TestNetNetwork))
	}
}

func Test_GetKnownTokens(t *testing.T) {
	token1, _ := wallet.NewTokenWithFullName("29f8873beb52e126f207a2dfd50f7cff556806b5b4cba9834a7b26a8.4b6173685f546f6b656e", true)
	token2, _ := wallet.NewTokenWithFullName("29f8873beb52e126f207a2dfd50f7cff556806b5b4cba9834a7b26a8.526f75746533", true)

	config := &CardanoChainConfig{
		NativeTokens: []sendtx.TokenExchangeConfig{
			{
				DstChainID: common.ChainIDStrVector,
				TokenName:  token1.String(),
			},
		},
	}

	retTokens, err := GetKnownTokens(config)
	require.NoError(t, err)
	require.Equal(t, 1, len(retTokens))
	require.Equal(t, token1, retTokens[0])

	config.NativeTokens = append(config.NativeTokens,
		sendtx.TokenExchangeConfig{
			DstChainID: common.ChainIDStrVector,
			TokenName:  token2.String(),
		},
	)

	retTokens, err = GetKnownTokens(config)
	require.NoError(t, err)
	require.Equal(t, 2, len(retTokens))
	require.Equal(t, token1, retTokens[0])
	require.Equal(t, token2, retTokens[1])

	config.NativeTokens = config.NativeTokens[1:]

	retTokens, err = GetKnownTokens(config)
	require.NoError(t, err)
	require.Equal(t, 1, len(retTokens))
	require.Equal(t, token2, retTokens[0])
}

func Test_subtractTxOutputsFromSumMap(t *testing.T) {
	tok1, err := GetNativeTokenFromName("3.31")
	require.NoError(t, err)

	tok2, err := GetNativeTokenFromName("3.32")
	require.NoError(t, err)

	tok3, err := GetNativeTokenFromName("3.33")
	require.NoError(t, err)

	tok4, err := GetNativeTokenFromName("3.34")
	require.NoError(t, err)

	vals := subtractTxOutputsFromSumMap(map[string]uint64{
		wallet.AdaTokenName: 200,
		tok1.String():       400,
		tok2.String():       500,
		tok4.String():       1000,
	}, []wallet.TxOutput{
		wallet.NewTxOutput("", 100, wallet.NewTokenAmount(tok1, 200), wallet.NewTokenAmount(tok2, 205)),
		wallet.NewTxOutput("", 50, wallet.NewTokenAmount(tok1, 150), wallet.NewTokenAmount(tok3, 300)),
		wallet.NewTxOutput("", 10, wallet.NewTokenAmount(tok2, 300)),
	})

	require.Equal(t, map[string]uint64{
		wallet.AdaTokenName: 40,
		tok1.String():       50,
		tok4.String():       1000,
	}, vals)
}
