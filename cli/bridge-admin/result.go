package clibridgeadmin

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
)

type chainTokenQuantity struct {
	chainID       string
	amount        *big.Int
	wrappedAmount *big.Int
}
type chainTokenQuantityResult struct {
	results []chainTokenQuantity
}

func (r chainTokenQuantityResult) GetOutput() string {
	var buffer bytes.Buffer

	data := make([]string, 0, len(r.results)*2)

	for _, x := range r.results {
		data = append(data, fmt.Sprintf("chainID|%s", x.chainID),
			fmt.Sprintf("amount|%s", x.amount),
			fmt.Sprintf("wrappedAmount|%s", x.wrappedAmount))
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

type deployCardanoScriptResult struct {
	PlutusAddr       string `json:"plutusAddr"`
	PolicyID         string `json:"policyId"`
	TxHash           string `json:"txHash"`
	RefScriptUtxoIdx uint32 `json:"refScriptUtxoIdx"`
}

func (d deployCardanoScriptResult) GetOutput() string {
	var buffer bytes.Buffer

	buffer.WriteString(common.FormatKV(
		[]string{
			fmt.Sprintf("Plutus script address|%s", d.PlutusAddr),
			fmt.Sprintf("Policy Id|%s", d.PolicyID),
			fmt.Sprintf("Reference Script Utxo Hash|%s", d.TxHash),
			fmt.Sprintf("Reference Script Utxo Index|%d", d.RefScriptUtxoIdx),
		}))

	return buffer.String()
}

type registerGatewayTokenResult struct {
	tokenRegisteredEvent *contractbinding.GatewayTokenRegistered
}

func (r registerGatewayTokenResult) GetOutput() string {
	var buffer bytes.Buffer

	buffer.WriteString(common.FormatKV(
		[]string{
			fmt.Sprintf("contractAddr|%s", r.tokenRegisteredEvent.ContractAddress.Hex()),
			fmt.Sprintf("tokenID|%v", r.tokenRegisteredEvent.TokenId),
			fmt.Sprintf("tokenName|%s", r.tokenRegisteredEvent.Name),
			fmt.Sprintf("tokenSymbol|%s", r.tokenRegisteredEvent.Symbol),
			fmt.Sprintf("externallyOwned|%v", r.tokenRegisteredEvent.IsLockUnlockToken),
		}))

	return buffer.String()
}
