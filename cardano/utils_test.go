package cardanotx

import (
	"hash/crc32"
	"testing"

	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/fxamacker/cbor/v2"
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

func Test_IsValidReceiverAddress(t *testing.T) {
	const (
		validMain = "addr1qx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer3n0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgse35a3x"
		validTest = "addr_test1vz2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzerspjrlsz"
		// CIP-19 Byron test vectors
		byronMain = "Ae2tdPwUPEZFRbyhz3cpfC2CumGzNkFBN2L42rcUc2yjQpEkxDbkPodpMAi"
		byronTest = "37btjrVyb4KDXBNC4haBVPCrro8AQPHwvCMp3RFhhSVWwfFmZ6wwzSK6JK1hY6wHNmtrpTf1kdbva8TCneM2YsiXT7mrzT21EacHnPpz5YyUdj64na"
	)

	t.Run("shelley addresses", func(t *testing.T) {
		require.True(t, IsValidReceiverAddress(validMain, wallet.MainNetNetwork))
		require.False(t, IsValidReceiverAddress(validMain, wallet.TestNetNetwork))
		require.True(t, IsValidReceiverAddress(validTest, wallet.TestNetNetwork))
		require.False(t, IsValidReceiverAddress(validTest, wallet.MainNetNetwork))
		require.False(t, IsValidReceiverAddress(
			"stake_test1uqehkck0lajq8gr28t9uxnuvgcqrc6070x3k9r8048z8y5gssrtvn", wallet.TestNetNetwork))
		require.False(t, IsValidReceiverAddress("addr1dummy", wallet.MainNetNetwork))
		require.False(t, IsValidReceiverAddress("", wallet.MainNetNetwork))
	})

	t.Run("well formed byron addresses", func(t *testing.T) {
		// IsValidOutputAddress accepts these, so only the Byron check rejects them
		require.True(t, IsValidOutputAddress(byronMain, wallet.MainNetNetwork))
		require.True(t, IsValidOutputAddress(byronTest, wallet.TestNetNetwork))
		require.True(t, IsValidOutputAddress(newByronAddress(t, 28), wallet.MainNetNetwork))

		for _, addr := range []string{byronMain, byronTest, newByronAddress(t, 28)} {
			require.False(t, IsValidReceiverAddress(addr, wallet.MainNetNetwork))
			require.False(t, IsValidReceiverAddress(addr, wallet.TestNetNetwork))
		}
	})

	t.Run("byron addresses with malformed root", func(t *testing.T) {
		for _, rootLen := range []int{0, 1, 27} {
			addr := newByronAddress(t, rootLen)

			for _, networkID := range []wallet.CardanoNetworkType{wallet.MainNetNetwork, wallet.TestNetNetwork} {
				valid := true

				require.NotPanics(t, func() {
					valid = IsValidReceiverAddress(addr, networkID)
				})
				require.False(t, valid)
			}
		}
	})
}

// newByronAddress builds a Byron address around a zeroed address root of rootLen bytes.
// Real Byron roots are 28 bytes, but cardano-infrastructure parses any length with a valid checksum.
func newByronAddress(t *testing.T, rootLen int) string {
	t.Helper()

	payload, err := cbor.Marshal([]any{make([]byte, rootLen), map[uint64][]byte{}, uint64(0)})
	require.NoError(t, err)

	raw, err := cbor.Marshal([]any{cbor.Tag{Number: 24, Content: payload}, crc32.ChecksumIEEE(payload)})
	require.NoError(t, err)

	addr, err := wallet.NewCardanoAddress(raw)
	require.NoError(t, err)

	return addr.String()
}
