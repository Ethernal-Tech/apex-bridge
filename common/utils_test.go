package common

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMulPercentage(t *testing.T) {
	assert.Equal(t, big.NewInt(74777), MulPercentage(big.NewInt(43987), 170))
	assert.Equal(t, big.NewInt(258281956132), MulPercentage(big.NewInt(782672594341), 33))
}

func TestSafeSubtract(t *testing.T) {
	assert.Equal(t, uint64(35), SafeSubtract(uint64(85), uint64(50), uint64(200)))
	assert.Equal(t, uint64(0), SafeSubtract(uint64(85), uint64(85), uint64(200)))
	assert.Equal(t, uint64(200), SafeSubtract(uint64(185), uint64(385), uint64(200)))
	assert.Equal(t, uint64(0), SafeSubtract(uint64(893), uint64(17833), uint64(0)))
}

func TestDfmToWei(t *testing.T) {
	assert.Equal(t, new(big.Int).SetUint64(1000000000000), DfmToWei(big.NewInt(1)))
	assert.Equal(t, new(big.Int).SetUint64(100000000000000), DfmToWei(big.NewInt(100)))
}

func TestWeiToDfm(t *testing.T) {
	assert.Equal(t, big.NewInt(0).Uint64(), WeiToDfm(big.NewInt(1)).Uint64())
	assert.Equal(t, big.NewInt(100), WeiToDfm(new(big.Int).SetUint64(100000000000000)))
}

func TestWeiToDfmCeil(t *testing.T) {
	assert.Equal(t, big.NewInt(1), WeiToDfmCeil(big.NewInt(1)))
	assert.Equal(t, big.NewInt(100), WeiToDfmCeil(new(big.Int).SetUint64(100000000000000)))
}
