package clisendtx

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

const sendSkylineTxCommandUse = "skyline"

var (
	sendtxParamsData = &sendTxParams{}
)

func GetSendTxCommand() *cobra.Command {
	cmdSendTx := &cobra.Command{
		Use:   "sendtx",
		Short: "sends apex bridging transaction",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return sendtxParamsData.validateFlags()
		},
		Run: common.GetCliRunCommand(sendtxParamsData),
	}
	cmdSendSkylineTx := getSendSkylineTxCommand()

	sendtxParamsData.setFlags(cmdSendTx)

	cmdSendTx.AddCommand(cmdSendSkylineTx)

	return cmdSendTx
}
