package cli

import (
	"fmt"
	"os"

	clicreateaddress "github.com/Ethernal-Tech/apex-bridge/cli/create-address"
	cligenerateconfigs "github.com/Ethernal-Tech/apex-bridge/cli/generateconfigs"
	cliregisterchain "github.com/Ethernal-Tech/apex-bridge/cli/registerchain"
	clirelayer "github.com/Ethernal-Tech/apex-bridge/cli/relayer"
	clivalidatorcomponents "github.com/Ethernal-Tech/apex-bridge/cli/validatorcomponents"
	cliwalletcreate "github.com/Ethernal-Tech/apex-bridge/cli/walletcreate"
	"github.com/spf13/cobra"
)

type RootCommand struct {
	baseCmd *cobra.Command
}

func NewRootCommand() *RootCommand {
	rootCommand := &RootCommand{
		baseCmd: &cobra.Command{
			Short: "cli commands for apex bridge",
		},
	}

	rootCommand.registerSubCommands()

	return rootCommand
}

func (rc *RootCommand) registerSubCommands() {
	rc.baseCmd.AddCommand(
		cliwalletcreate.GetWalletCreateCommand(),
		cliregisterchain.GetRegisterChainCommand(),
		clivalidatorcomponents.GetValidatorComponentsCommand(),
		clirelayer.GetRunRelayerCommand(),
		clicreateaddress.GetCreateAddressCommand(),
		cligenerateconfigs.GetGenerateConfigsCommand(),
	)
}

func (rc *RootCommand) Execute() {
	if err := rc.baseCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}
}
