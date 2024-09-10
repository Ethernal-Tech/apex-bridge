package batcher

import (
	"errors"
	"fmt"
)

var (
	errNonActiveBatchPeriod = errors.New("non active batch period")
)

func getNumberWithRoundingThreshold(
	number, threshold uint64, noBatchPeriodPercent float64,
) (uint64, error) {
	if number == 0 {
		return 0, errors.New("cannot round a zero value")
	}

	newNumber := ((number + threshold - 1) / threshold) * threshold
	diffFromPrevious := number - (newNumber - threshold)

	if diffFromPrevious <= uint64(float64(threshold)*noBatchPeriodPercent) ||
		diffFromPrevious > uint64(float64(threshold)*(1.0-noBatchPeriodPercent)) {
		return 0, fmt.Errorf("%w: (number, rounded) = (%d, %d)", errNonActiveBatchPeriod, number, newNumber)
	}

	return newNumber, nil
}
