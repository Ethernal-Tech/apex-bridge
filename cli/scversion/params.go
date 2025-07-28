package cliscversion

import (
	"context"
	"fmt"
	"strings"

	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

const (
	nodeURLFlag = "node-url"
	scAddrFlag  = "addr"

	nodeURLFlagDesc = "node url"
	scAddrFlagDesc  = "list of smart contract addresses"
)

type scVersionParams struct {
	nodeURL     string
	scAddresses []string

	ethTxHelper ethtxhelper.IEthTxHelper
}

func (ip *scVersionParams) validateFlags() error {
	ethTxHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithNodeURL(ip.nodeURL))
	if err != nil {
		return fmt.Errorf("failed to connect to the bridge node: %w", err)
	}

	if len(ip.scAddresses) == 0 {
		return fmt.Errorf("no smart contract addresses specified: --%s", scAddrFlag)
	}

	ip.ethTxHelper = ethTxHelper

	return nil
}

func (ip *scVersionParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&ip.nodeURL,
		nodeURLFlag,
		"",
		nodeURLFlagDesc,
	)

	cmd.Flags().StringSliceVar(
		&ip.scAddresses,
		scAddrFlag,
		[]string{},
		scAddrFlagDesc,
	)
}

func (ip *scVersionParams) Execute(ctx context.Context, outputter common.OutputFormatter) (
	common.ICommandResult, error) {
	for _, scAddrStr := range ip.scAddresses {
		scParams := strings.SplitN(scAddrStr, ":", 2)

		if len(scParams) != 2 {
			_, _ = outputter.Write([]byte(scAddrStr + ": Invalid address flag\n"))
			outputter.WriteOutput()

			continue
		}

		addr := common.HexToAddress(scParams[0])
		scName := scParams[1]

		response, err := ip.ethTxHelper.GetClient().CallContract(context.Background(), ethereum.CallMsg{
			To: &addr,
			// bytes4(keccak256("version()"))
			Data: []byte{0x54, 0xfd, 0x4d, 0x50},
		}, nil)

		if err != nil || len(response) == 0 {
			_, _ = outputter.Write([]byte(scName + ": No version available\n"))
			outputter.WriteOutput()

			continue
		}

		_, _ = outputter.Write([]byte(scName + " Smart Contract version:\n"))
		_, _ = outputter.Write(response)
		_, _ = outputter.Write([]byte("\n"))
		outputter.WriteOutput()
	}

	return &CmdResult{}, nil
}
