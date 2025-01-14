package clisendtx

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

const sendSkylineTxCommandUse = "skyline-tx"

var sendtxParamsData = &sendTxParams{}
var sendSkylineTxParamsData = &sendSkylineTxParams{}

func GetSendTxCommand() *cobra.Command {
	cmdSendTx := &cobra.Command{
		Use:     "sendtx",
		Short:   "sends apex bridging transaction",
		PreRunE: runPreRun,
		Run:     common.GetCliRunCommand(sendtxParamsData),
	}
	cmdSendSkylineTx := &cobra.Command{
		Use:     sendSkylineTxCommandUse,
		Short:   "sends the transaction in skyline mode",
		PreRunE: runPreRun,
		Run:     common.GetCliRunCommand(sendSkylineTxParamsData),
	}

	sendtxParamsData.setFlags(cmdSendTx)
	sendSkylineTxParamsData.setFlags(cmdSendSkylineTx)

	cmdSendTx.AddCommand(cmdSendSkylineTx)

	return cmdSendTx
}

func runPreRun(cb *cobra.Command, _ []string) error {
	if cb.Use == sendSkylineTxCommandUse {
		return sendSkylineTxParamsData.validateFlags()
	}

	return sendtxParamsData.validateFlags()
}
