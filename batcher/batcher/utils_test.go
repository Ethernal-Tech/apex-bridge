package batcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getNumberWithRoundingThreshold(t *testing.T) {
	_, err := getNumberWithRoundingThreshold(66, 60, 0.125)
	assert.ErrorIs(t, err, errNonActiveBatchPeriod)

	_, err = getNumberWithRoundingThreshold(12, 60, 0.2)
	assert.ErrorIs(t, err, errNonActiveBatchPeriod)

	_, err = getNumberWithRoundingThreshold(115, 60, 0.125)
	assert.ErrorIs(t, err, errNonActiveBatchPeriod)

	_, err = getNumberWithRoundingThreshold(228, 80, 0.2)
	assert.ErrorIs(t, err, errNonActiveBatchPeriod)

	_, err = getNumberWithRoundingThreshold(336, 80, 0.2)
	assert.ErrorIs(t, err, errNonActiveBatchPeriod)

	_, err = getNumberWithRoundingThreshold(0, 60, 0.125)
	assert.ErrorContains(t, err, "cannot round a zero value")

	val, err := getNumberWithRoundingThreshold(75, 60, 0.125)
	assert.NoError(t, err)
	assert.Equal(t, uint64(120), val)

	val, err = getNumberWithRoundingThreshold(105, 60, 0.125)
	assert.NoError(t, err)
	assert.Equal(t, uint64(120), val)

	val, err = getNumberWithRoundingThreshold(40, 60, 0.125)
	assert.NoError(t, err)
	assert.Equal(t, uint64(60), val)

	val, err = getNumberWithRoundingThreshold(270, 80, 0.125)
	assert.NoError(t, err)
	assert.Equal(t, uint64(320), val)

	val, err = getNumberWithRoundingThreshold(223, 80, 0.2)
	assert.NoError(t, err)
	assert.Equal(t, uint64(240), val)

	val, err = getNumberWithRoundingThreshold(337, 80, 0.2)
	assert.NoError(t, err)
	assert.Equal(t, uint64(400), val)

	val, err = getNumberWithRoundingThreshold(5, 6, 0.09)
	assert.NoError(t, err)
	assert.Equal(t, uint64(6), val)
}
