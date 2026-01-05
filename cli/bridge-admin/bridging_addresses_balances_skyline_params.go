package clibridgeadmin

import (
	"fmt"
	"os"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	vcCore "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/spf13/cobra"
)

const (
	cardanoWalletAddressFlag = "cardano-wallet-addr"

	cardanoWalletAddressFlagDesc = "cardano hot wallet/bridging/multisig address"
)

type bridgingAddressesBalancesSkylineParams struct {
	config               string
	chainIDsConfig       string
	primeWalletAddress   string
	cardanoWalletAddress string
	indexerDbsPath       string

	appConfig *vcCore.AppConfig
}

func (b *bridgingAddressesBalancesSkylineParams) ValidateFlags() error {
	if b.primeWalletAddress == "" || b.cardanoWalletAddress == "" {
		return fmt.Errorf("all wallet addresses --%s and --%s must be set",
			primeWalletAddressFlag, cardanoWalletAddressFlag)
	}

	if err := validateConfigFilePath(b.config); err != nil {
		return err
	}

	if err := validateConfigFilePath(b.chainIDsConfig); err != nil {
		return err
	}

	chainIDsConfig, err := common.LoadConfig[common.ChainIDsConfigFile](b.chainIDsConfig, "")
	if err != nil {
		return err
	}

	appConfig, err := loadConfig(b.config, chainIDsConfig)
	if err != nil {
		return fmt.Errorf("failed to load config file: %w", err)
	}

	b.appConfig = appConfig

	if b.indexerDbsPath != "" {
		if _, err := os.Stat(b.indexerDbsPath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("indexer database path does not exist: %s", b.indexerDbsPath)
			}

			return fmt.Errorf("failed to check indexer database path: %s. err: %w", b.indexerDbsPath, err)
		}
	}

	if !common.IsValidAddress(common.ChainIDStrPrime, b.primeWalletAddress, appConfig.ChainIDConverter) {
		return fmt.Errorf("invalid address: --%s", primeWalletAddressFlag)
	}

	if !common.IsValidAddress(common.ChainIDStrCardano, b.cardanoWalletAddress, appConfig.ChainIDConverter) {
		return fmt.Errorf("invalid address: --%s", cardanoWalletAddressFlag)
	}

	return nil
}

func (b *bridgingAddressesBalancesSkylineParams) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&b.config,
		configFlag,
		"",
		configFlagDesc,
	)
	cmd.Flags().StringVar(
		&b.chainIDsConfig,
		chainIDsConfigFlag,
		"",
		chainIDsConfigFlagDesc,
	)
	cmd.Flags().StringVar(
		&b.primeWalletAddress,
		primeWalletAddressFlag,
		"",
		primeWalletAddressFlagDesc,
	)
	cmd.Flags().StringVar(
		&b.cardanoWalletAddress,
		cardanoWalletAddressFlag,
		"",
		cardanoWalletAddressFlagDesc,
	)
	cmd.Flags().StringVar(
		&b.indexerDbsPath,
		indexerDbsPathFlag,
		"",
		indexerDbsPathFlagDesc,
	)
}

func (b *bridgingAddressesBalancesSkylineParams) Execute(
	outputter common.OutputFormatter) (common.ICommandResult, error) {
	if b.appConfig.RunMode != common.SkylineMode {
		return nil, fmt.Errorf("running command for the wrong run mode: %s", b.appConfig.RunMode)
	}

	chainWalletAddr := map[string]string{
		common.ChainIDStrPrime:   b.primeWalletAddress,
		common.ChainIDStrCardano: b.cardanoWalletAddress,
	}

	multisigUtxos, err := getAllUtxos(b.appConfig, chainWalletAddr, b.indexerDbsPath)
	if err != nil {
		return nil, err
	}

	for chainID, utxos := range multisigUtxos {
		var (
			lovelaceBalance     = uint64(0)
			wrappedTokenBalance = uint64(0)
			filteredCount       int
		)

		chainConfig := b.appConfig.CardanoChains[chainID]

		knownTokens, err := cardanotx.GetKnownTokens(&chainConfig.CardanoChainConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to get known tokens: %w", err)
		}

		for _, utxo := range utxos {
			if !cardanotx.UtxoContainsUnknownTokens(utxo, knownTokens...) {
				filteredCount++

				lovelaceBalance += utxo.Amount

				if len(chainConfig.CardanoChainConfig.Tokens) == 0 {
					continue
				}

				nativeToken, err := chainConfig.CardanoChainConfig.GetWrappedToken()
				if err != nil {
					return nil, err
				}

				multisigWrappedTokenAmount := cardanotx.GetTokenAmount(
					&utxo, nativeToken.String())

				wrappedTokenBalance += multisigWrappedTokenAmount
			}
		}

		_, _ = outputter.Write([]byte(fmt.Sprintf("Balances on %s chain: \n", chainID)))
		_, _ = outputter.Write([]byte(fmt.Sprintf("Bridging Address = %s\n", chainWalletAddr[chainID])))
		_, _ = outputter.Write([]byte(fmt.Sprintf("Lovelace Balance = %d\n", lovelaceBalance)))
		_, _ = outputter.Write([]byte(fmt.Sprintf("Wrapped Token Balance = %d\n", wrappedTokenBalance)))
		_, _ = outputter.Write([]byte(fmt.Sprintf("All UTXOs = %d\n", len(utxos))))
		_, _ = outputter.Write([]byte(fmt.Sprintf("Filtered UTXOs = %d\n", filteredCount)))
		outputter.WriteOutput()
	}

	return nil, nil
}

var (
	_ common.CliCommandExecutor = (*bridgingAddressesBalancesParams)(nil)
)
