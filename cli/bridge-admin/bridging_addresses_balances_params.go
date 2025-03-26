package clibridgeadmin

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/Ethernal-Tech/apex-bridge/common"
	vcCore "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/validatorcomponents"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	indexerDb "github.com/Ethernal-Tech/cardano-infrastructure/indexer/db"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
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
	if b.indexerDbsPath == "" && (b.primeWalletAddress == "" || b.vectorWalletAddress == "" || b.nexusWalletAddress == "") {
		return fmt.Errorf("either all wallet addresses --prime-wallet-addr, --vector-wallet-addr, and --nexus-wallet-addr must be set, or --indexer-dbs-path must be set")
	}

	if b.indexerDbsPath == "" {
		if !common.IsValidAddress(common.ChainIDStrPrime, b.primeWalletAddress) {
			return fmt.Errorf("invalid address: --%s", primeWalletAddressFlag)
		}

		if !common.IsValidAddress(common.ChainIDStrVector, b.vectorWalletAddress) {
			return fmt.Errorf("invalid address: --%s", vectorWalletAddressFlag)
		}

		if !common.IsValidAddress(common.ChainIDStrNexus, b.nexusWalletAddress) {
			return fmt.Errorf("invalid address: --%s", nexusWalletAddressFlag)
		}
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

	if b.indexerDbsPath != "" { //get UTXOs from database
		ctx := context.Background()

		ethHelper := eth.NewEthHelperWrapper(
			hclog.NewNullLogger(),
			ethtxhelper.WithNodeURL(appConfig.Bridge.NodeURL),
			ethtxhelper.WithInitClientAndChainIDFn(ctx),
			ethtxhelper.WithNonceStrategyType(appConfig.Bridge.NonceStrategy),
			ethtxhelper.WithDynamicTx(appConfig.Bridge.DynamicTx),
		)

		bridgeSmartContract := eth.NewBridgeSmartContract(
			appConfig.Bridge.SmartContractAddress, ethHelper)

		err = validatorcomponents.FixChainsAndAddresses(ctx, appConfig, bridgeSmartContract, hclog.NewNullLogger())
		if err != nil {
			return nil, err
		}

		oracleConfig, _ := appConfig.SeparateConfigs()

		cardanoIndexerDbs := make(map[string]indexer.Database, len(oracleConfig.CardanoChains))

		//Open connections to the DB for Cardano chains
		for _, cardanoChainConfig := range oracleConfig.CardanoChains {
			indexerDB, err := indexerDb.NewDatabaseInit("",
				filepath.Join(b.indexerDbsPath, cardanoChainConfig.ChainID+".db"))
			if err != nil {
				return nil, fmt.Errorf("failed to open oracle indexer db for `%s`: %w", cardanoChainConfig.ChainID, err)
			}

			cardanoIndexerDbs[cardanoChainConfig.ChainID] = indexerDB
		}

		//Retrieve UTXOs for Cardano BridgingAddresses from the DB
		for chainID, cardanoIndexerDB := range cardanoIndexerDbs {
			bridgingAddresses := oracleConfig.CardanoChains[chainID].BridgingAddresses

			multisigUtxos, err := cardanoIndexerDB.GetAllTxOutputs(bridgingAddresses.BridgingAddress, true)
			if err != nil {
				return nil, err
			}

			var multisigBalance uint64
			for _, utxo := range multisigUtxos {
				if len(utxo.Output.Tokens) == 0 {
					multisigBalance += utxo.Output.Amount
				}
			}

			feeUtxos, err := cardanoIndexerDB.GetAllTxOutputs(bridgingAddresses.FeeAddress, true)
			if err != nil {
				return nil, err
			}

			var feeBalance uint64
			for _, utxo := range feeUtxos {
				if len(utxo.Output.Tokens) == 0 {
					feeBalance += utxo.Output.Amount
				}
			}

			_, _ = outputter.Write([]byte(fmt.Sprintf("Balances on %s chain: \n", chainID)))
			_, _ = outputter.Write([]byte(fmt.Sprintf("Bridging Address =  %s\n", bridgingAddresses.BridgingAddress)))
			_, _ = outputter.Write([]byte(fmt.Sprintf("Balance =  %d\n", multisigBalance)))
			_, _ = outputter.Write([]byte(fmt.Sprintf("Fee Address =  %s\n", bridgingAddresses.FeeAddress)))
			_, _ = outputter.Write([]byte(fmt.Sprintf("Balance =  %d\n", feeBalance)))
			outputter.WriteOutput()
		}
	} else { //get UTXOs from ogmios

	}

	return nil, nil
}

var (
	_ common.CliCommandExecutor = (*bridgingAddressesBalancesParams)(nil)
)
