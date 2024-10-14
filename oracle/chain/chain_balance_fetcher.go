package chain

import (
	"context"
	"math/big"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
)

type ChainBalanceFetcherImpl struct {
	ctx        context.Context
	appConfig  *core.AppConfig
	balancesDB core.BalanceStatesDB
	logger     hclog.Logger
}

var _ core.ChainBalanceFetcher = (*ChainBalanceFetcherImpl)(nil)

func NewChainBalanceFetcher(
	ctx context.Context,
	config *core.AppConfig,
	balancesDB core.BalanceStatesDB,
	logger hclog.Logger,
) *ChainBalanceFetcherImpl {
	return &ChainBalanceFetcherImpl{
		ctx:        ctx,
		appConfig:  config,
		balancesDB: balancesDB,
		logger:     logger.Named("balance_fetcher"),
	}
}

func (cb *ChainBalanceFetcherImpl) NewUnprocessedTxs(originChainID string, txs []*indexer.Tx) error {
	var (
		supplyDelta  = new(big.Int).SetUint64(0)
		multisigAddr = cb.appConfig.CardanoChains[originChainID].BridgingAddresses.BridgingAddress
	)

	for _, tx := range txs {
		// INCREASE
		// if cardanoTx.Outputs[X].Address == multisigAddr
		// if cardanoTx.Outputs[X].Amount > 0
		for _, txOut := range tx.Outputs {
			if strings.Compare(txOut.Address, multisigAddr) != 0 {
				continue
			}

			supplyDelta.Add(supplyDelta, new(big.Int).SetUint64(txOut.Amount))
		}

		// DECREASE
		// cardanoTx.Inputs[0].Output.Address == multisigAddr
		// cardanoTx.Inputs[0].Output.Amount > 0
		for _, txIn := range tx.Inputs {
			if strings.Compare(txIn.Output.Address, multisigAddr) != 0 {
				continue
			}

			supplyDelta.Sub(supplyDelta, new(big.Int).SetUint64(txIn.Output.Amount))
		}

		// No new changes to supply
		if supplyDelta.Int64() == 0 {
			continue
		}

		chainBalance, err := cb.balancesDB.GetLastChainBalances(originChainID, 1)
		if err != nil {
			return err
		}

		balance, _ := new(big.Int).SetString(chainBalance[0].Amount, 0)
		balance.Add(balance, supplyDelta)

		if err = cb.balancesDB.AddChainBalance(&core.ChainBalance{
			ChainID: originChainID,
			Height:  tx.BlockSlot,
			Amount:  balance.String(),
		}); err != nil {
			return err
		}

		cb.logger.Info("available supply changed on chain", "originChainID", originChainID, "balance", balance)
	}

	return nil
}

func (cb *ChainBalanceFetcherImpl) Start() {
	cb.logger.Debug("Starting ChainBalanceFetcher")
}
