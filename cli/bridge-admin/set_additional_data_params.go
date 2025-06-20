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
	bridgeSCAddrFlag = "bridge-addr"

	bridgingAddrFlag = "bridging-addr"
	feeAddrFlag      = "fee-addr"

	bridgeSCAddrFlagDesc = "bridge smart contract address"
	bridgingAddrFlagDesc = "bridging address string"
	feeAddrFlagDesc      = "fee address string"
)

type setAdditionalDataParams struct {
	chainID                string
	bridgeNodeURL          string
	bridgeSCAddr           string
	bridgePrivateKey       string
	bridgePrivateKeyConfig string
	bridgingAddr           string
	feeAddr                string
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

	if ip.bridgePrivateKey == "" || ip.bridgePrivateKeyConfig == "" {
		return fmt.Errorf("specify at least one: --%s or --%s", bridgePrivateKeyFlag, bridgePrivateKeyConfigFlag)
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
		&ip.bridgePrivateKeyConfig,
		bridgePrivateKeyConfigFlag,
		"",
		bridgePrivateKeyConfigFlagDesc,
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

	cmd.MarkFlagsMutuallyExclusive(bridgePrivateKeyConfigFlag, bridgePrivateKeyFlag)
}

func (ip *setAdditionalDataParams) Execute(
	outputter common.OutputFormatter,
) (common.ICommandResult, error) {
	ctx := context.Background()

	wallet, err := eth.GetEthWalletForBladeAdmin(false, ip.bridgePrivateKey, ip.bridgePrivateKeyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create bridge admin wallet: %w", err)
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
