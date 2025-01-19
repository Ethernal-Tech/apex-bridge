package clibridgeadmin

import (
	"context"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
)

const (
	bridgeSCAddrFlag     = "bridge-addr"
	bridgePrivateKeyFlag = "bridge-key"
	bridgingAddrFlag     = "bridging-addr"
	feeAddrFlag          = "fee-addr"

	bridgeSCAddrFlagDesc     = "bridge smart contract address"
	bridgePrivateKeyFlagDesc = "private key for bridge wallet"
	bridgingAddrFlagDesc     = "bridging address string"
	feeAddrFlagDesc          = "fee address string"
)

type setAdditionalDataParams struct {
	chainID          string
	bridgeNodeURL    string
	bridgeSCAddr     string
	bridgePrivateKey string
	bridgingAddr     string
	feeAddr          string
}

func (ip *setAdditionalDataParams) ValidateFlags() error {
	if !common.IsExistingChainID(ip.chainID) {
		return fmt.Errorf("invalid --%s flag", chainIDFlag)
	}

	if !common.IsValidHTTPURL(ip.bridgeNodeURL) {
		return fmt.Errorf("invalid --%s flag", bridgeNodeURLFlag)
	}

	if !ethcommon.IsHexAddress(ip.bridgeSCAddr) {
		return fmt.Errorf("invalid --%s flag", bridgeSCAddrFlag)
	}

	if ip.bridgingAddr == "" || !common.IsValidAddress(ip.chainID, ip.bridgingAddr) {
		return fmt.Errorf("invalid --%s flag", bridgingAddrFlag)
	}

	if common.IsEVMChainID(ip.chainID) {
		ip.feeAddr = ""
	} else if ip.feeAddr == "" || !common.IsValidAddress(ip.chainID, ip.feeAddr) {
		return fmt.Errorf("invalid --%s flag", feeAddrFlag)
	}

	return nil
}

func (ip *setAdditionalDataParams) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&ip.chainID,
		chainIDFlag,
		"",
		chainIDFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.bridgeNodeURL,
		bridgeNodeURLFlag,
		"",
		bridgeNodeURLFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.bridgeSCAddr,
		bridgeSCAddrFlag,
		apexBridgeScAddress.String(),
		bridgeSCAddrFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.bridgePrivateKey,
		bridgePrivateKeyFlag,
		"",
		bridgePrivateKeyFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.bridgingAddr,
		bridgingAddrFlag,
		"",
		bridgingAddrFlagDesc,
	)

	cmd.Flags().StringVar(
		&ip.feeAddr,
		feeAddrFlag,
		"",
		feeAddrFlagDesc,
	)
}

func (ip *setAdditionalDataParams) Execute(
	outputter common.OutputFormatter,
) (common.ICommandResult, error) {
	ctx := context.Background()

	wallet, err := ethtxhelper.NewEthTxWallet(ip.bridgePrivateKey)
	if err != nil {
		return nil, err
	}

	txHelperWrapper := eth.NewEthHelperWrapperWithWallet(
		wallet, hclog.NewNullLogger(),
		ethtxhelper.WithNodeURL(ip.bridgeNodeURL),
		ethtxhelper.WithInitClientAndChainIDFn(ctx),
		ethtxhelper.WithNonceStrategyType(ethtxhelper.NonceInMemoryStrategy),
		ethtxhelper.WithDynamicTx(false))
	smartContract := eth.NewBridgeSmartContract(ip.bridgeSCAddr, txHelperWrapper)

	_, _ = outputter.Write([]byte("Sending transactions..."))
	outputter.WriteOutput()

	_, err = infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (bool, error) {
		err := smartContract.SetChainAdditionalData(ctx, ip.chainID, ip.bridgingAddr, ip.feeAddr)

		return err == nil, err
	})
	if err != nil {
		return nil, err
	}

	return successResult{}, nil
}
