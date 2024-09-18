package clideployevm

import (
	"bytes"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
)

type CmdResult struct {
	gatewayProxyAddr              string
	gatewayAddr                   string
	nativeTokenPredicateProxyAddr string
	nativeTokenPredicateAddr      string
	nativeTokenWalletProxyAddr    string
	nativeTokenWalletAddr         string
	validatorsProxyAddr           string
	validatorsAddr                string
}

func (r CmdResult) GetOutput() string {
	var buffer bytes.Buffer

	buffer.WriteString(common.FormatKV(
		[]string{
			fmt.Sprintf("Gateway Proxy Address|%s", r.gatewayProxyAddr),
			fmt.Sprintf("Gateway Address|%s", r.gatewayAddr),
			fmt.Sprintf("NativeTokenPredicate Proxy Address|%s", r.nativeTokenPredicateProxyAddr),
			fmt.Sprintf("NativeTokenPredicate Address|%s", r.nativeTokenPredicateAddr),
			fmt.Sprintf("NativeTokenWallet Proxy Address|%s", r.nativeTokenWalletProxyAddr),
			fmt.Sprintf("NativeTokenWallet Address|%s", r.nativeTokenWalletAddr),
			fmt.Sprintf("Validators Proxy Address|%s", r.validatorsProxyAddr),
			fmt.Sprintf("Validators Address|%s", r.validatorsAddr),
		}))

	return buffer.String()
}
