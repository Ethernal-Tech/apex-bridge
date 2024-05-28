package clisendtx

import (
	"bytes"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type CmdResult struct {
	SenderAddr string
	ChainID    string
	TxHash     string
	Receipts   []cardanowallet.TxOutput
}

func (r CmdResult) GetOutput() string {
	var buffer bytes.Buffer

	kvPairs := []string{
		fmt.Sprintf("Sender|%s", r.SenderAddr),
		fmt.Sprintf("Chain|%s", r.ChainID),
		fmt.Sprintf("Tx Hash|%s", r.TxHash),
	}

	for _, x := range r.Receipts {
		kvPairs = append(kvPairs, fmt.Sprintf("Receiver|%s", x.Addr))
		kvPairs = append(kvPairs, fmt.Sprintf("Amount|%d", x.Amount))
	}

	buffer.WriteString("Transaction has been bridged\n")
	buffer.WriteString(common.FormatKV(kvPairs))

	return buffer.String()
}