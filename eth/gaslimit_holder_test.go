package eth

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGasLimitHolder(t *testing.T) {
	a := NewGasLimitHolder(10, 21, 3)
	err := errors.New("F")

	assert.Equal(t, uint64(10), a.GetGasLimit())

	a.Update(err)

	assert.Equal(t, uint64(14), a.GetGasLimit())

	a.Update(err)

	assert.Equal(t, uint64(18), a.GetGasLimit())

	a.Update(err)

	assert.Equal(t, uint64(21), a.GetGasLimit())

	a.Update(err)

	assert.Equal(t, uint64(21), a.GetGasLimit())

	a.Update(nil)

	assert.Equal(t, uint64(10), a.GetGasLimit())
}
