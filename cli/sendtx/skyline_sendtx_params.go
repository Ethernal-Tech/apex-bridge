package clisendtx

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

type sendSkylineTxParams struct {
}

func (p *sendSkylineTxParams) validateFlags() error {
	return fmt.Errorf("unimplemented")
}

func (p *sendSkylineTxParams) setFlags(cmd *cobra.Command) {

}

func (p *sendSkylineTxParams) Execute(
	outputter common.OutputFormatter,
) (common.ICommandResult, error) {
	return nil, fmt.Errorf("unimplemented")
}
