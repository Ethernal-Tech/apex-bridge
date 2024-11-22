package clideployevm

import (
	"bytes"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

type contractInfo struct {
	Name    string
	Addr    ethcommon.Address
	IsProxy bool
}

type cmdResult struct {
	Contracts []contractInfo
}

func (r cmdResult) GetOutput() string {
	var (
		buffer  bytes.Buffer
		columns []string
	)

	for _, x := range r.Contracts {
		if x.IsProxy {
			columns = append(columns, fmt.Sprintf("%s Proxy Address|%s", x.Name, x.Addr))
		} else {
			columns = append(columns, fmt.Sprintf("%s Address|%s", x.Name, x.Addr))
		}
	}

	buffer.WriteString(common.FormatKV(columns))

	return buffer.String()
}
