package clibridgeadmin

import (
	"fmt"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	vcCore "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

const (
	configFlag = "config"

	configFlagDesc = "path to config json file"
)

type validatorsDataParams struct {
	bridgeNodeURL string
	bridgeSCAddr  string
	config        string
}

func (v *validatorsDataParams) ValidateFlags() error {
	if !common.IsValidHTTPURL(v.bridgeNodeURL) {
		return fmt.Errorf("invalid --%s flag", bridgeNodeURLFlag)
	}

	if !ethcommon.IsHexAddress(v.bridgeSCAddr) {
		return fmt.Errorf("invalid --%s flag", bridgeSCAddrFlag)
	}

	return nil
}

func (v *validatorsDataParams) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&v.bridgeNodeURL,
		bridgeNodeURLFlag,
		"",
		bridgeNodeURLFlagDesc,
	)

	cmd.Flags().StringVar(
		&v.bridgeSCAddr,
		bridgeSCAddrFlag,
		apexBridgeScAddress.String(),
		bridgeSCAddrFlagDesc,
	)

	cmd.Flags().StringVar(
		&v.config,
		configFlag,
		"",
		configFlagDesc,
	)
}

func (v *validatorsDataParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	txHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithNodeURL(v.bridgeNodeURL))
	if err != nil {
		return nil, err
	}

	contract, err := contractbinding.NewBridgeContract(common.HexToAddress(v.bridgeSCAddr), txHelper.GetClient())
	if err != nil {
		return nil, err
	}

	config, err := common.LoadConfig[vcCore.AppConfig](v.config, "")
	if err != nil {
		return nil, err
	}

	allRegisteredChains, err := contract.GetAllRegisteredChains(&bind.CallOpts{})
	if err != nil {
		return nil, err
	}

	for _, regChain := range allRegisteredChains {
		chainID := common.ToStrChainID(regChain.Id)

		switch regChain.ChainType {
		case common.ChainTypeCardano:
			chainConfig, exists := config.CardanoChains[chainID]
			if !exists {
				return nil, err
			}

			validatorsData, err := contract.GetValidatorsChainData(&bind.CallOpts{}, common.ToNumChainID(chainID))
			if err != nil {
				return nil, err
			}

			_, _ = outputter.Write([]byte(fmt.Sprintf("Validators data on %s chain: \n", chainID)))
			_, _ = outputter.Write([]byte(eth.GetChainValidatorsDataInfoString(chainID, validatorsData)))
			outputter.WriteOutput()

			multisigPolicyScript, multisigFeePolicyScript, err := cardanotx.GetPolicyScripts(validatorsData)
			if err != nil {
				return nil, err
			}

			multisigAddr, feeAddr, err := cardanotx.GetMultisigAddresses(
				wallet.ResolveCardanoCliBinary(chainConfig.NetworkID), uint(chainConfig.NetworkMagic),
				multisigPolicyScript, multisigFeePolicyScript)
			if err != nil {
				return nil, err
			}

			_, _ = outputter.Write([]byte(fmt.Sprintf("Addresses on %s chain (retrieved from validator data): \n", chainID)))
			_, _ = outputter.Write([]byte(fmt.Sprintf("Multisig Address =  %s\n", multisigAddr)))
			_, _ = outputter.Write([]byte(fmt.Sprintf("Fee Payer Address = %s\n", feeAddr)))

			_, _ = outputter.Write([]byte(fmt.Sprintf("Addresses on %s chain (retrieved from registered chains): \n", chainID)))
			_, _ = outputter.Write([]byte(fmt.Sprintf("Multisig Address =  %s\n", regChain.AddressMultisig)))
			_, _ = outputter.Write([]byte(fmt.Sprintf("Fee Payer Address = %s\n", regChain.AddressFeePayer)))
			outputter.WriteOutput()
		case common.ChainTypeEVM:
			validatorsData, err := contract.GetValidatorsChainData(&bind.CallOpts{}, common.ToNumChainID(chainID))
			if err != nil {
				return nil, err
			}

			_, _ = outputter.Write([]byte(fmt.Sprintf("Validators data on %s chain: \n", chainID)))
			_, _ = outputter.Write([]byte(eth.GetChainValidatorsDataInfoString(chainID, validatorsData)))
			outputter.WriteOutput()

			_, _ = outputter.Write([]byte(fmt.Sprintf("Addresses on %s chain (retrieved from registered chains): \n", chainID)))
			_, _ = outputter.Write([]byte(fmt.Sprintf("Multisig Address =  %s\n", regChain.AddressMultisig)))
			outputter.WriteOutput()

		default:
		}
	}

	return nil, nil
}

var (
	_ common.CliCommandExecutor = (*validatorsDataParams)(nil)
)
