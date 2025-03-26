package clibridgeadmin

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/Ethernal-Tech/apex-bridge/common"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	vcCore "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	indexerDb "github.com/Ethernal-Tech/cardano-infrastructure/indexer/db"
	"github.com/spf13/cobra"
)

const (
	primeWalletAddressFlag  = "prime-wallet-addr"
	vectorWalletAddressFlag = "vector-wallet-addr"
	nexusWalletAddressFlag  = "nexus-wallet-addr"
	indexerDbsPathFlag      = "indexer-dbs-path"

	primeWalletAddressFlagDesc  = "prime wallet address"
	vectorWalletAddressFlagDesc = "vector wallet address"
	nexusWalletAddressFlagDesc  = "nexus wallet address"
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
		return fmt.Errorf("invalid config path: --%s", indexerDbsPathFlag)
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

	if b.indexerDbsPath != "" { // Retrieve Cardano balances from the database
		cardanoIndexerDbs := make(map[string]indexer.Database, len(appConfig.CardanoChains))

		// Open connections to the DB for Cardano chains
		for chainID := range appConfig.CardanoChains {
			indexerDB, err := indexerDb.NewDatabaseInit("",
				filepath.Join(b.indexerDbsPath, chainID+".db"))
			if err != nil {
				return nil, fmt.Errorf("failed to open oracle indexer db for `%s`: %w", chainID, err)
			}

			cardanoIndexerDbs[chainID] = indexerDB
		}

		// Retrieve UTXOs for Cardano BridgingAddresses from the DB
		for chainID, cardanoIndexerDB := range cardanoIndexerDbs {
			var bridgingAddress string
			if chainID == common.ChainIDStrPrime {
				bridgingAddress = b.primeWalletAddress
			} else if chainID == common.ChainIDStrVector {
				bridgingAddress = b.vectorWalletAddress
			}

			multisigUtxos, err := cardanoIndexerDB.GetAllTxOutputs(bridgingAddress, true)
			if err != nil {
				return nil, err
			}

			var multisigBalance uint64

			for _, utxo := range multisigUtxos {
				if len(utxo.Output.Tokens) == 0 {
					multisigBalance += utxo.Output.Amount
				}
			}

			_, _ = outputter.Write([]byte(fmt.Sprintf("Balances on %s chain: \n", chainID)))
			_, _ = outputter.Write([]byte(fmt.Sprintf("Bridging Address =  %s\n", bridgingAddress)))
			_, _ = outputter.Write([]byte(fmt.Sprintf("Balance =  %d\n", multisigBalance)))
			outputter.WriteOutput()
		}
	} else { // Retrieve Cardano balances via Ogmios

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
