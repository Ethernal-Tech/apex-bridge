package batcher

import (
	"fmt"
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

func Test_getNumberWithRoundingThresholdRoundDown(t *testing.T) {
	_, err := getNumberWithRoundingThresholdRoundDown(66, 60, 0.125)
	assert.ErrorIs(t, err, errNonActiveBatchPeriod)

	_, err = getNumberWithRoundingThresholdRoundDown(115, 60, 0.125)
	assert.ErrorIs(t, err, errNonActiveBatchPeriod)

	_, err = getNumberWithRoundingThresholdRoundDown(228, 80, 0.2)
	assert.ErrorIs(t, err, errNonActiveBatchPeriod)

	_, err = getNumberWithRoundingThresholdRoundDown(0, 60, 0.125)
	assert.ErrorContains(t, err, "cannot round a zero value")

	val, err := getNumberWithRoundingThresholdRoundDown(75, 60, 0.125)
	assert.NoError(t, err)
	assert.Equal(t, uint64(60), val)

	val, err = getNumberWithRoundingThresholdRoundDown(105, 60, 0.125)
	assert.NoError(t, err)
	assert.Equal(t, uint64(60), val)

	val, err = getNumberWithRoundingThresholdRoundDown(40, 60, 0.125)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), val)

	val, err = getNumberWithRoundingThresholdRoundDown(270, 80, 0.125)
	assert.NoError(t, err)
	assert.Equal(t, uint64(240), val)

	val, err = getNumberWithRoundingThresholdRoundDown(223, 80, 0.2)
	assert.NoError(t, err)
	assert.Equal(t, uint64(160), val)

	val, err = getNumberWithRoundingThresholdRoundDown(337, 80, 0.2)
	assert.NoError(t, err)
	assert.Equal(t, uint64(320), val)

	val, err = getNumberWithRoundingThresholdRoundDown(5, 6, 0.09)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), val)
}

func Test_getNumberWithRoundingThresholdRoundDown_OurConfig_WindowBehavior(t *testing.T) {
	threshold := uint64(10)
	noBatchPeriodPercent := 0.01
	lowCut := uint64(float64(threshold) * noBatchPeriodPercent)          // 1
	highCut := uint64(float64(threshold) * (1.0 - noBatchPeriodPercent)) // 9

	fmt.Println("lowCut", lowCut, "highCut", highCut)

	for n := uint64(1); n <= 1000; n++ {
		got, err := getNumberWithRoundingThresholdRoundDown(n, threshold, noBatchPeriodPercent)
		rem := n % threshold

		assert.NoError(t, err, "n=%d rem=%d", n, rem)
		assert.Equal(t, (n/threshold)*threshold, got, "n=%d rem=%d", n, rem)
	}
}
