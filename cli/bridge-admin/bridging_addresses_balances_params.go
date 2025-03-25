package clibridgeadmin

import (
	"fmt"
	"path/filepath"

	"github.com/Ethernal-Tech/apex-bridge/common"
	vcCore "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	indexerDb "github.com/Ethernal-Tech/cardano-infrastructure/indexer/db"
	"github.com/spf13/cobra"
)

type bridgingAddressesBalancesParams struct {
	config string
}

func (b *bridgingAddressesBalancesParams) ValidateFlags() error {
	return nil
}

func (b *bridgingAddressesBalancesParams) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&b.config,
		configFlag,
		"",
		configFlagDesc,
	)
}

func (b *bridgingAddressesBalancesParams) Execute(outputter common.OutputFormatter) (common.ICommandResult, error) {
	appConfig, err := common.LoadConfig[vcCore.AppConfig](b.config, "")
	if err != nil {
		return nil, err
	}

	appConfig.FillOut()

	oracleConfig, _ := appConfig.SeparateConfigs()

	cardanoIndexerDbs := make(map[string]indexer.Database, len(oracleConfig.CardanoChains))

	for _, cardanoChainConfig := range oracleConfig.CardanoChains {
		indexerDB, err := indexerDb.NewDatabaseInit("",
			filepath.Join(appConfig.Settings.DbsPath, cardanoChainConfig.ChainID+".db"))
		if err != nil {
			return nil, fmt.Errorf("failed to open oracle indexer db for `%s`: %w", cardanoChainConfig.ChainID, err)
		}

		cardanoIndexerDbs[cardanoChainConfig.ChainID] = indexerDB
	}

	ethIndexerDbs := make(map[string]eventTrackerStore.EventTrackerStore, len(oracleConfig.EthChains))

	for _, ethChainConfig := range oracleConfig.EthChains {
		indexerDB, err := eventTrackerStore.NewBoltDBEventTrackerStore(filepath.Join(
			appConfig.Settings.DbsPath, ethChainConfig.ChainID+".db"))
		if err != nil {
			return nil, fmt.Errorf("failed to open oracle indexer db for `%s`: %w", ethChainConfig.ChainID, err)
		}

		ethIndexerDbs[ethChainConfig.ChainID] = indexerDB
	}

	return nil, nil
}

var (
	_ common.CliCommandExecutor = (*bridgingAddressesBalancesParams)(nil)
)
