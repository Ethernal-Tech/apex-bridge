package eth

type GasLimitHolder struct {
	currentGasLimit uint64
	minGasLimit     uint64
	maxGasLimit     uint64
	increment       uint64
}

func NewGasLimitHolder(minGL uint64, maxGL uint64, steps uint64) GasLimitHolder {
	return GasLimitHolder{
		currentGasLimit: minGL,
		minGasLimit:     minGL,
		maxGasLimit:     maxGL,
		increment:       (maxGL - minGL + steps - 1) / steps,
	}
}

func (glh *GasLimitHolder) Update(err error) {
	if err != nil {
		glh.currentGasLimit = min(glh.maxGasLimit, glh.currentGasLimit+glh.increment)
	} else {
		glh.currentGasLimit = glh.minGasLimit
	}
}

func (glh GasLimitHolder) GetGasLimit() uint64 {
	return glh.currentGasLimit
}
