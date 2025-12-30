package clibridgeadmin

import (
	"encoding/hex"
	"fmt"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/bn256"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

const (
	configFlag = "config"

	configFlagDesc = "path to config json file"
)

type ValidatorChainData = contractbinding.IBridgeStructsValidatorChainData

type validatorsDataParams struct {
	bridgeNodeURL  string
	bridgeSCAddr   string
	config         string
	chainIDsConfig string
}

func (v *validatorsDataParams) ValidateFlags() error {
	if !common.IsValidHTTPURL(v.bridgeNodeURL) {
		return fmt.Errorf("invalid --%s flag", bridgeNodeURLFlag)
	}

	if !ethcommon.IsHexAddress(v.bridgeSCAddr) {
		return fmt.Errorf("invalid --%s flag", bridgeSCAddrFlag)
	}

	if err := validateConfigFilePath(v.config); err != nil {
		return err
	}

	if err := validateConfigFilePath(v.chainIDsConfig); err != nil {
		return err
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

	cmd.Flags().StringVar(
		&v.chainIDsConfig,
		chainIDsConfigFlag,
		"",
		chainIDsConfigFlagDesc,
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

	chainIDsConfig, err := common.LoadConfig[common.ChainIDsConfig](v.chainIDsConfig, "")
	if err != nil {
		return nil, fmt.Errorf("failed to load chain IDs config: %w", err)
	}

	config, err := loadConfig(v.config, chainIDsConfig)
	if err != nil {
		return nil, err
	}

	chainIDConverter := config.ChainIDConverter

	allRegisteredChains, err := contract.GetAllRegisteredChains(&bind.CallOpts{})
	if err != nil {
		return nil, err
	}

	for _, regChain := range allRegisteredChains {
		chainID := chainIDConverter.ToStrChainID(regChain.Id)

		switch regChain.ChainType {
		case common.ChainTypeCardano:
			chainConfig, exists := config.CardanoChains[chainID]
			if !exists {
				return nil, err
			}

			validatorsData, err := contract.GetValidatorsChainData(&bind.CallOpts{}, chainIDConverter.ToNumChainID(chainID))
			if err != nil {
				return nil, err
			}

			_, _ = outputter.Write([]byte(fmt.Sprintf("Validators data on %s chain: \n", chainID)))

			err = printChainValidatorsDataInfo(chainID, validatorsData, chainIDConverter, outputter)
			if err != nil {
				return nil, err
			}

			keyHashes, err := cardanotx.NewApexKeyHashes(validatorsData)
			if err != nil {
				return nil, err
			}

			addrCount, err := contract.GetBridgingAddressesCount(&bind.CallOpts{}, chainIDConverter.ToNumChainID(chainID))
			if err != nil {
				return nil, err
			}

			_, _ = outputter.Write([]byte(fmt.Sprintf("Addresses on %s chain (retrieved from validator data): \n", chainID)))
			_, _ = outputter.Write([]byte(fmt.Sprintf("Multisig address count: %d \n", addrCount)))

			for i := range addrCount {
				policyScripts := cardanotx.NewApexPolicyScripts(keyHashes, uint64(i))

				addrs, err := cardanotx.NewApexAddresses(
					wallet.ResolveCardanoCliBinary(chainConfig.NetworkID), uint(chainConfig.NetworkMagic), policyScripts)
				if err != nil {
					return nil, err
				}

				if i == 0 {
					_, _ = outputter.Write([]byte(fmt.Sprintf("Fee Payer Address = %s\n", addrs.Fee.Payment)))
				}

				_, _ = outputter.Write([]byte(fmt.Sprintf("Multisig Address %d =  %s\n", i, addrs.Multisig.Payment)))
			}

			_, _ = outputter.Write([]byte(fmt.Sprintf("Addresses on %s chain (retrieved from registered chains): \n", chainID)))
			_, _ = outputter.Write([]byte(fmt.Sprintf("Multisig Address =  %s\n", regChain.AddressMultisig)))
			_, _ = outputter.Write([]byte(fmt.Sprintf("Fee Payer Address = %s\n", regChain.AddressFeePayer)))

			outputter.WriteOutput()
		case common.ChainTypeEVM:
			validatorsData, err := contract.GetValidatorsChainData(&bind.CallOpts{}, chainIDConverter.ToNumChainID(chainID))
			if err != nil {
				return nil, err
			}

			_, _ = outputter.Write([]byte(fmt.Sprintf("Validators data on %s chain: \n", chainID)))

			err = printChainValidatorsDataInfo(chainID, validatorsData, chainIDConverter, outputter)
			if err != nil {
				return nil, err
			}

			_, _ = outputter.Write([]byte(fmt.Sprintf("Addresses on %s chain (retrieved from registered chains): \n", chainID)))
			_, _ = outputter.Write([]byte(fmt.Sprintf("Multisig Address =  %s\n", regChain.AddressMultisig)))
			outputter.WriteOutput()

		default:
		}
	}

	return nil, nil
}

func printChainValidatorsDataInfo(
	chainID string, data []ValidatorChainData,
	chainIDConverter *common.ChainIDConverter, outputter common.OutputFormatter,
) error {
	for _, x := range data {
		var formattedData string

		if chainIDConverter.IsEVMChainID(chainID) {
			pub, err := bn256.UnmarshalPublicKeyFromBigInt(x.Key)
			if err != nil {
				return err
			}

			formattedData = fmt.Sprintf("BLSKey=%s", hex.EncodeToString(pub.Marshal()))
		} else {
			formattedData = fmt.Sprintf(
				"MultisigKey=%s, FeeKey=%s",
				hex.EncodeToString(wallet.PadKeyToSize(x.Key[0].Bytes())),
				hex.EncodeToString(wallet.PadKeyToSize(x.Key[1].Bytes())),
			)
		}

		_, _ = outputter.Write([]byte(formattedData))
		outputter.WriteOutput()
	}

	return nil
}

var (
	_ common.CliCommandExecutor = (*validatorsDataParams)(nil)
)
