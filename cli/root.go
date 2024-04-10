package cli

import (
	"fmt"
	"os"

	cliregisterchain "github.com/Ethernal-Tech/apex-bridge/cli/registerchain"
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
	)
}

func (rc *RootCommand) Execute() {
	if err := rc.baseCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}
}