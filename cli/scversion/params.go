package cliscversion

import (
	"context"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/Ethernal-Tech/apex-bridge/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

const (
	bridgeURLFlag      = "bridge-url"
	bridgeSCAddrFlag   = "bridge-addr"
	gatewayURLFlag     = "gateway-url"
	gatewayAddressFlag = "gateway-addr"

	bridgeURLFlagDesc      = "bridge node url"
	bridgeSCAddrFlagDesc   = "bridge smart contract address"
	gatewayURLFlagDesc     = "gateway url"
	gatewayAddressFlagDesc = "gateway smart contract address"
)

type scVersionParams struct {
	bridgeURL    string
	bridgeSCAddr string

	gatewayURL  string
	gatewayAddr string

	ethTxHelper ethtxhelper.IEthTxHelper
}

func (ip *scVersionParams) validateFlags() error {
	ethTxHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithNodeURL(ip.bridgeURL))
	if err != nil {
		return fmt.Errorf("failed to connect to the bridge node: %w", err)
	}

	ip.ethTxHelper = ethTxHelper

	addrDecoded, err := common.DecodeHex(ip.bridgeSCAddr)
	if err != nil || len(addrDecoded) == 0 || len(addrDecoded) > 20 {
		return fmt.Errorf("invalid bridge smart contract address: %s", ip.bridgeSCAddr)
	}

	return nil
}

func (ip *scVersionParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&ip.bridgeURL,
		bridgeURLFlag,
		"",
		bridgeURLFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.bridgeSCAddr,
		bridgeSCAddrFlag,
		"",
		bridgeSCAddrFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.gatewayURL,
		gatewayURLFlag,
		"",
		gatewayURLFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.gatewayAddr,
		gatewayAddressFlag,
		"",
		gatewayAddressFlagDesc,
	)
}

func (ip *scVersionParams) Execute(ctx context.Context, outputter common.OutputFormatter) (
	common.ICommandResult, error) {
	contract, err := contractbinding.NewBridgeContract(
		common.HexToAddress(ip.bridgeSCAddr), ip.ethTxHelper.GetClient())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to bridge smart contract: %w", err)
	}

	version, err := contract.Version(&bind.CallOpts{
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get smart : %w", err)
	}

	_, _ = outputter.Write([]byte("Smart Contract version retrieved:\n"))
	_, _ = outputter.Write([]byte(version))
	_, _ = outputter.Write([]byte("\n"))
	outputter.WriteOutput()

	ethTxHelper, err := ethtxhelper.NewEThTxHelper(
		[]ethtxhelper.TxRelayerOption{
			ethtxhelper.WithNodeURL(ip.gatewayURL),
		}...)
	if err != nil {
		return nil, fmt.Errorf("error while NewEThTxHelper: %w", err)
	}

	smartContractAddress := ethcommon.HexToAddress(ip.gatewayAddr)

	evmContract, err := contractbinding.NewGateway(smartContractAddress, ethTxHelper.GetClient())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gateway smart contract : %w", err)
	}

	gatewayVersion, err := evmContract.Version(&bind.CallOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway version: %w", err)
	}

	_, _ = outputter.Write([]byte("Gateway Contract version retrieved:\n"))
	_, _ = outputter.Write([]byte(gatewayVersion))
	_, _ = outputter.Write([]byte("\n"))
	outputter.WriteOutput()

	return &CmdResult{}, nil
}
