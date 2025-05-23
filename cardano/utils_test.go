package cardanotx

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetPolicyScripts_And_GetMultisigAddresses(t *testing.T) {
	keys := []string{
		"5d767d06a9426bafd31eae25122b586fb6cac32efcee60c94bf8f43faddb8f5b",
		"f5c69c8a0bb63016068d4683dab19af0b0833158ed7a5ed91bd328d0939f3173",
		"2983addc84a6032feeb8870a9f74308e4a5446779bf44a8a790be3fb266e1abd",
		"001c4e9e6493675a3f380d1749d203a5aabb92b217b28b918a1b0aea6b8981b0",
	}
	feeKeys := []string{
		"07e522c4ddf84b7bd0f6e005cfd101c56fd6f361268327899d2eb132c480",
		"100f649faa1661922873cb05caed4daf4e4fa0e870d9ae4dd1c30ec0a00a9a16",
		"afcb5befaeeab56bbf731ee3bebb143331854ce394b9e061bfc5764ad62c07cf",
		"2b31319bc86a77f72a6e140618116861f2093faa7516ba6b34db4abcb3cbbf5d",
	}
	validatorsData := make([]eth.ValidatorChainData, len(keys))

	for i := range validatorsData {
		bytes, err := hex.DecodeString(keys[i])
		require.NoError(t, err)

		bytesFee, err := hex.DecodeString(feeKeys[i])
		require.NoError(t, err)

		validatorsData[i] = eth.ValidatorChainData{
			Key: [4]*big.Int{
				new(big.Int).SetBytes(bytes), new(big.Int).SetBytes(bytesFee),
			},
		}
	}

	ps, psFee, err := GetPolicyScripts(validatorsData)
	require.NoError(t, err)

	addr, feeAddr, err := GetMultisigAddresses(
		wallet.ResolveCardanoCliBinary(wallet.TestNetNetwork), wallet.TestNetProtocolMagic, ps, psFee)
	require.NoError(t, err)

	require.Equal(t, "addr_test1wp8ylty98278gsgmxdm90uq8338maed4hnp3up23560dpvs76xwds", addr)
	require.Equal(t, "addr_test1wpqcqpc58msz3gkcev0ecl067077cdtkjys7mae6cd0jxqgkfs4cm", feeAddr)
}

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

func Test_BigIntToKey(t *testing.T) {
	t.Run("less than 32 bytes", func(t *testing.T) {
		b := BigIntToKey(big.NewInt(1))

		require.Equal(t, 32, len(b))
		require.Equal(t, append(make([]byte, 31), 1), b)
	})

	t.Run("exactly 32 bytes", func(t *testing.T) {
		bytes := make([]byte, 32)
		bytes[31] = 1
		bytes[0] = 0xFF

		b := BigIntToKey(new(big.Int).SetBytes(bytes))

		require.Equal(t, 32, len(b))
		require.Equal(t, bytes, b)
	})

	t.Run("more than 32 bytes", func(t *testing.T) {
		bytes := make([]byte, 34)
		bytes[31] = 1
		bytes[0] = 0xFF

		b := BigIntToKey(new(big.Int).SetBytes(bytes))

		require.Equal(t, 32, len(b))
		require.Equal(t, bytes[:32], b)
	})
}
