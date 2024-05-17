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
