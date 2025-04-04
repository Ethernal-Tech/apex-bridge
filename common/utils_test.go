package common

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMulPercentage(t *testing.T) {
	t.Parallel()

	assert.Equal(t, big.NewInt(74777), MulPercentage(big.NewInt(43987), 170))
	assert.Equal(t, big.NewInt(258281956132), MulPercentage(big.NewInt(782672594341), 33))
}

func TestSafeSubtract(t *testing.T) {
	t.Parallel()

	assert.Equal(t, uint64(35), SafeSubtract(uint64(85), uint64(50), uint64(200)))
	assert.Equal(t, uint64(0), SafeSubtract(uint64(85), uint64(85), uint64(200)))
	assert.Equal(t, uint64(200), SafeSubtract(uint64(185), uint64(385), uint64(200)))
	assert.Equal(t, uint64(0), SafeSubtract(uint64(893), uint64(17833), uint64(0)))
}

func TestDfmToWei(t *testing.T) {
	t.Parallel()

	assert.Equal(t, new(big.Int).SetUint64(1000000000000), DfmToWei(big.NewInt(1)))
	assert.Equal(t, new(big.Int).SetUint64(100000000000000), DfmToWei(big.NewInt(100)))
}

func TestWeiToDfm(t *testing.T) {
	t.Parallel()

	assert.Equal(t, big.NewInt(0).Uint64(), WeiToDfm(big.NewInt(1)).Uint64())
	assert.Equal(t, big.NewInt(100), WeiToDfm(new(big.Int).SetUint64(100000000000000)))
}

func TestWeiToDfmCeil(t *testing.T) {
	t.Parallel()

	assert.Equal(t, big.NewInt(1), WeiToDfmCeil(big.NewInt(1)))
	assert.Equal(t, big.NewInt(100), WeiToDfmCeil(new(big.Int).SetUint64(100000000000000)))
}

func TestDecodeHex(t *testing.T) {
	t.Parallel()

	tst := func(s string, expected []byte) {
		t.Helper()

		v, err := DecodeHex(s)

		if expected != nil {
			require.NoError(t, err)
			assert.Equal(t, expected, v)
		} else {
			assert.Error(t, err)
		}
	}

	tst("0x10", []byte{16})
	tst("0X0F03", []byte{15, 3})
	tst("FE", []byte{254})
	tst("0XXX", nil)
	tst("0x", []byte{})
}

func TestRetryForever(t *testing.T) {
	t.Parallel()

	cnt := 0

	err := RetryForever(context.Background(), time.Millisecond*10, func(ctx context.Context) error {
		cnt++
		if cnt == 10 {
			return context.DeadlineExceeded
		}

		return errors.New("ferika")
	})
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Equal(t, cnt, 10)

	err = RetryForever(context.Background(), time.Millisecond*10, func(ctx context.Context) error {
		cnt++

		return nil
	})
	require.NoError(t, err)
	require.Equal(t, cnt, 11)
}

func TestSplitString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"a", "b", "c"}, SplitString("abc", 1))
	assert.Equal(t, []string{"ab", "cd"}, SplitString("abcd", 2))
	assert.Equal(t, []string{"ab", "cd", "e"}, SplitString("abcde", 2))
	assert.Equal(t, []string{"abc", "def", "gh"}, SplitString("abcdefgh", 3))
}

func TestGetRequiredSignaturesForConsensus(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 3, int(GetRequiredSignaturesForConsensus(4)))
	assert.Equal(t, 5, int(GetRequiredSignaturesForConsensus(6)))
	assert.Equal(t, 7, int(GetRequiredSignaturesForConsensus(10)))
	assert.Equal(t, 14, int(GetRequiredSignaturesForConsensus(20)))
}

func TestKeccak256(t *testing.T) {
	t.Parallel()

	v, err := Keccak256([]byte{1, 2, 3})
	require.NoError(t, err)
	assert.Equal(t, []byte{0xf1, 0x88, 0x5e, 0xda, 0x54, 0xb7, 0xa0, 0x53, 0x31, 0x8c, 0xd4, 0x1e, 0x20, 0x93, 0x22, 0xd, 0xab, 0x15, 0xd6, 0x53, 0x81, 0xb1, 0x15, 0x7a, 0x36, 0x33, 0xa8, 0x3b, 0xfd, 0x5c, 0x92, 0x39}, v)

	v, err = Keccak256([]byte{1, 2, 3}, []byte{8})
	require.NoError(t, err)
	assert.Equal(t, []byte{0x16, 0x8f, 0xd6, 0x4e, 0xe, 0x5b, 0xd7, 0x19, 0xf7, 0xaf, 0x8e, 0x9a, 0x67, 0xc6, 0x66, 0xc8, 0x87, 0x96, 0x7e, 0x8e, 0x2e, 0x8e, 0xf6, 0xbe, 0xc0, 0x19, 0xf7, 0xd1, 0x35, 0x5d, 0xa6, 0x88}, v)

	v, err = Keccak256(nil)
	require.NoError(t, err)
	assert.Equal(t, []byte{0xc5, 0xd2, 0x46, 0x1, 0x86, 0xf7, 0x23, 0x3c, 0x92, 0x7e, 0x7d, 0xb2, 0xdc, 0xc7, 0x3, 0xc0, 0xe5, 0x0, 0xb6, 0x53, 0xca, 0x82, 0x27, 0x3b, 0x7b, 0xfa, 0xd8, 0x4, 0x5d, 0x85, 0xa4, 0x70}, v)
}

func TestHexToAddress(t *testing.T) {
	t.Parallel()

	assert.Equal(t, append(make([]byte, 19), 0xAA), HexToAddress("0xAA").Bytes())
	assert.Equal(t, append(make([]byte, 17), 0xCC, 0xFF, 0xAA), HexToAddress("0xCCFFAA").Bytes())
}

func TestIsValidHTTPURL(t *testing.T) {
	t.Parallel()

	assert.False(t, IsValidHTTPURL(""))
	assert.False(t, IsValidHTTPURL("pera.com"))
	assert.False(t, IsValidHTTPURL("httpS://pera.com:aaa/fe"))
	assert.False(t, IsValidHTTPURL("/sevap"))
	assert.False(t, IsValidHTTPURL("https://sevap:90:90"))
	assert.False(t, IsValidHTTPURL("https://sevap:KRA"))
	assert.False(t, IsValidHTTPURL("https://sevap:-54"))
	assert.True(t, IsValidHTTPURL("http://pera.com"))
	assert.True(t, IsValidHTTPURL("https://pera.com:3345"))
	assert.True(t, IsValidHTTPURL("httpS://pera.com"))
	assert.True(t, IsValidHTTPURL("httpS://pera.com:8989/fe"))
	assert.True(t, IsValidHTTPURL("hTTp://pera.com/sto/pet?hello=123"))
}

func TestIsIsValidNetworkAddress(t *testing.T) {
	t.Parallel()

	assert.True(t, IsValidNetworkAddress("pera.com"))
	assert.True(t, IsValidNetworkAddress("pera.com:1898"))
	assert.True(t, IsValidNetworkAddress("192.8.0.1:1898"))
	assert.True(t, IsValidNetworkAddress("192.8.0.21"))
	assert.False(t, IsValidNetworkAddress("192.8.0"))
	assert.False(t, IsValidNetworkAddress("pera.com:-1"))
	assert.False(t, IsValidNetworkAddress("pera.com:2:0"))
	assert.False(t, IsValidNetworkAddress("pera.com/23232"))
	assert.False(t, IsValidNetworkAddress("http://pera.com:2"))
	assert.False(t, IsValidNetworkAddress(""))
}

func TestPackNumbersToBytes(t *testing.T) {
	type MyInts []int16

	input1 := []int32{4, 5, -10, 6, 20, 30, 17, 89893}
	input2 := []uint64{784834, 347834, 34893, 121, 0, 378273}
	input3 := []float32{0.3, 8.11, 89.8989, -189892.9}
	input4 := MyInts{-32768, 32767, 0, 20}

	bytes := PackNumbersToBytes(input1)

	result, err := UnpackNumbersToBytes[[]int32](bytes)
	require.NoError(t, err)
	assert.Equal(t, input1, result)

	bytes = PackNumbersToBytes(input2)

	result2, err := UnpackNumbersToBytes[[]uint64](bytes)
	require.NoError(t, err)
	assert.Equal(t, input2, result2)

	bytes = PackNumbersToBytes(input3)

	result3, err := UnpackNumbersToBytes[[]float32](bytes)
	require.NoError(t, err)
	assert.Equal(t, input3, result3)

	bytes = PackNumbersToBytes(input4)

	result4, err := UnpackNumbersToBytes[MyInts](bytes)
	require.NoError(t, err)
	assert.Equal(t, input4, result4)
}

func TestNumbersToString(t *testing.T) {
	assert.Equal(t, "1, 7, -3, 9090, 889", NumbersToString([]int{1, 7, -3, 9090, 889}))
	assert.Equal(t, "-1.01, 7, -3.56, 9090, 8.8", NumbersToString([]float32{-1.01, 7, -3.56, 9090, 8.8}))
}
