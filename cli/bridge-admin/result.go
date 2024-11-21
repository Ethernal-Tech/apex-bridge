package clibridgeadmin

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
)

type chainTokenQuantity struct {
	chainID string
	amount  *big.Int
}
type chainTokenQuantityResult struct {
	results []chainTokenQuantity
}

func (r chainTokenQuantityResult) GetOutput() string {
	var buffer bytes.Buffer

	data := make([]string, 0, len(r.results)*2)

	for _, x := range r.results {
		data = append(data, fmt.Sprintf("chainID|%s", x.chainID),
			fmt.Sprintf("amount|%s", x.amount))
	}

	buffer.WriteString(common.FormatKV(data))

	return buffer.String()
}

type successResult struct {
}

func (r successResult) GetOutput() string {
	var buffer bytes.Buffer

	buffer.WriteString("command execution has been finished\n")

	return buffer.String()
}
