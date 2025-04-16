package clibridgeadmin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ethernal-Tech/apex-bridge/common"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	vcCore "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	indexerDb "github.com/Ethernal-Tech/cardano-infrastructure/indexer/db"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/spf13/cobra"
)

const (
	primeWalletAddressFlag  = "prime-wallet-addr"
	vectorWalletAddressFlag = "vector-wallet-addr"
	nexusWalletAddressFlag  = "nexus-wallet-addr"
	indexerDbsPathFlag      = "indexer-dbs-path"

	primeWalletAddressFlagDesc  = "prime hot wallet/bridging/multisig address"
	vectorWalletAddressFlagDesc = "vector hot wallet/bridging/multisig address"
	nexusWalletAddressFlagDesc  = "nexus NativeTokenWallet Proxy sc address"
	indexerDbsPathFlagDesc      = "path to the indexer database"
)

type bridgingAddressesBalancesParams struct {
	config              string
	primeWalletAddress  string
	vectorWalletAddress string
	nexusWalletAddress  string
	indexerDbsPath      string
}

func (b *bridgingAddressesBalancesParams) ValidateFlags() error {
	if b.primeWalletAddress == "" || b.vectorWalletAddress == "" || b.nexusWalletAddress == "" {
		return fmt.Errorf("all wallet addresses --%s, --%s and --%s must be set",
			primeWalletAddressFlag, vectorWalletAddressFlag, nexusWalletAddressFlag)
	}

	if !common.IsValidAddress(common.ChainIDStrPrime, b.primeWalletAddress) {
		return fmt.Errorf("invalid address: --%s", primeWalletAddressFlag)
	}

	if !common.IsValidAddress(common.ChainIDStrVector, b.vectorWalletAddress) {
		return fmt.Errorf("invalid address: --%s", vectorWalletAddressFlag)
	}

	if !common.IsValidAddress(common.ChainIDStrNexus, b.nexusWalletAddress) {
		return fmt.Errorf("invalid address: --%s", nexusWalletAddressFlag)
	}

	if b.config == "" {
		return fmt.Errorf("--%s flag not specified", configFlag)
	}

	if _, err := os.Stat(b.config); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config file does not exist: %s", b.config)
		}

		return fmt.Errorf("failed to check config file: %s. err: %w", b.config, err)
	}

	if b.indexerDbsPath != "" {
		if _, err := os.Stat(b.indexerDbsPath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("indexer database path does not exist: %s", b.indexerDbsPath)
			}

			return fmt.Errorf("failed to check indexer database path: %s. err: %w", b.indexerDbsPath, err)
		}
	}

	return nil
}

func (b *bridgingAddressesBalancesParams) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&b.config,
		configFlag,
		"",
		configFlagDesc,
	)
	cmd.Flags().StringVar(
		&b.primeWalletAddress,
		primeWalletAddressFlag,
		"",
		primeWalletAddressFlagDesc,
	)
	cmd.Flags().StringVar(
		&b.vectorWalletAddress,
		vectorWalletAddressFlag,
		"",
		vectorWalletAddressFlagDesc,
	)
	cmd.Flags().StringVar(
		&b.nexusWalletAddress,
		nexusWalletAddressFlag,
		"",
		nexusWalletAddressFlagDesc,
	)
	cmd.Flags().StringVar(
		&b.indexerDbsPath,
		indexerDbsPathFlag,
		"",
		indexerDbsPathFlagDesc,
	)
}

func (b *bridgingAddressesBalancesParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	appConfig, err := common.LoadConfig[vcCore.AppConfig](b.config, "")
	if err != nil {
		return nil, err
	}

	if appConfig.RunMode != common.ReactorMode {
		return nil, fmt.Errorf("running command for the wrong run mode: %s", appConfig.RunMode)
	}

	chainWalletAddr := map[string]string{
		common.ChainIDStrPrime:  b.primeWalletAddress,
		common.ChainIDStrVector: b.vectorWalletAddress,
		common.ChainIDStrNexus:  b.nexusWalletAddress,
	}

	multisigUtxos, err := getAllUtxos(appConfig, chainWalletAddr, b.indexerDbsPath)
	if err != nil {
		return nil, err
	}

	for chainID, utxos := range multisigUtxos {
		var (
			lovelaceBalance = uint64(0)
			filteredCount   int
		)

		for _, utxo := range utxos {
			if len(utxo.Tokens) == 0 {
				lovelaceBalance += utxo.Amount
				filteredCount++
			}
		}

		_, _ = outputter.Write([]byte(fmt.Sprintf("Balances on %s chain: \n", chainID)))
		_, _ = outputter.Write([]byte(fmt.Sprintf("Bridging Address = %s\n", chainWalletAddr[chainID])))
		_, _ = outputter.Write([]byte(fmt.Sprintf("Balance = %d\n", lovelaceBalance)))
		_, _ = outputter.Write([]byte(fmt.Sprintf("All UTXOs = %d\n", len(utxos))))
		_, _ = outputter.Write([]byte(fmt.Sprintf("Filtered UTXOs = %d\n", filteredCount)))
		outputter.WriteOutput()
	}

	// Retrieve balances for Ethereum chains
	for chainID, ethChainConfig := range appConfig.EthChains {
		ethHelper, err := ethtxhelper.NewEThTxHelper(
			ethtxhelper.WithNodeURL(ethChainConfig.NodeURL))
		if err != nil {
			return nil, err
		}

		address := common.HexToAddress(b.nexusWalletAddress)

		balance, err := ethHelper.GetClient().BalanceAt(context.Background(), address, nil)
		if err != nil {
			return nil, err
		}

		_, _ = outputter.Write([]byte(fmt.Sprintf("Balances on %s chain: \n", chainID)))
		_, _ = outputter.Write([]byte(fmt.Sprintf("Bridging Address = %s\n", b.nexusWalletAddress)))
		_, _ = outputter.Write([]byte(fmt.Sprintf("Balance =  %d\n", balance)))
		outputter.WriteOutput()
	}

	return nil, nil
}

var (
	_ common.CliCommandExecutor = (*bridgingAddressesBalancesParams)(nil)
)

func getAllUtxos(
	appConfig *vcCore.AppConfig, chainWalletAddr map[string]string, indexerDbsPath string,
) (map[string][]indexer.TxOutput, error) {
	multisigUtxos := make(map[string][]indexer.TxOutput)

	if indexerDbsPath != "" { // Retrieve Cardano balances from the database
		cardanoIndexerDbs := make(map[string]indexer.Database, len(appConfig.CardanoChains))

		// Open connections to the DB for Cardano chains
		for chainID := range appConfig.CardanoChains {
			indexerDB, err := indexerDb.NewDatabaseInit("",
				filepath.Join(indexerDbsPath, chainID+".db"))
			if err != nil {
				return nil, fmt.Errorf("failed to open oracle indexer db for `%s`: %w", chainID, err)
			}

			cardanoIndexerDbs[chainID] = indexerDB
		}

		// Retrieve UTXOs for Cardano BridgingAddresses from the DB
		for chainID, cardanoIndexerDB := range cardanoIndexerDbs {
			bridgingAddress := chainWalletAddr[chainID]

			indexerUtxos, err := cardanoIndexerDB.GetAllTxOutputs(bridgingAddress, true)
			if err != nil {
				return nil, err
			}

			for _, txOut := range indexerUtxos {
				multisigUtxos[chainID] = append(multisigUtxos[chainID], txOut.Output)
			}
		}

		for chainID, indexerDB := range cardanoIndexerDbs {
			err := indexerDB.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to close the indexer db for chain: %s. err: %w", chainID, err)
			}
		}
	} else { // Retrieve Cardano balances via Ogmios
		for chainID, cardanoConfig := range appConfig.CardanoChains {
			txProvider := cardanowallet.NewTxProviderOgmios(cardanoConfig.OgmiosURL)

			allUtxos, err := infracommon.ExecuteWithRetry(context.Background(),
				func(ctx context.Context) ([]cardanowallet.Utxo, error) {
					return txProvider.GetUtxos(ctx, chainWalletAddr[chainID])
				})
			if err != nil {
				return nil, err
			}

			for _, utxo := range allUtxos {
				walletUtxo := indexer.TxOutput{
					Amount: utxo.Amount,
					Tokens: make([]indexer.TokenAmount, len(utxo.Tokens)),
				}
				for i, token := range utxo.Tokens {
					walletUtxo.Tokens[i] = indexer.TokenAmount{
						PolicyID: token.PolicyID,
						Name:     token.Name,
						Amount:   token.Amount,
					}
				}

				multisigUtxos[chainID] = append(multisigUtxos[chainID], walletUtxo)
			}
		}
	}

	return multisigUtxos, nil
}
