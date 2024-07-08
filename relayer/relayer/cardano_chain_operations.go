package relayer

import (
	"context"
	"encoding/json"
	"fmt"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

var _ core.ChainOperations = (*CardanoChainOperations)(nil)

type CardanoChainOperations struct {
	txProvider       cardanowallet.ITxProvider
	cardanoCliBinary string
	logger           hclog.Logger
}

func NewCardanoChainOperations(
	jsonConfig json.RawMessage,
	logger hclog.Logger,
) (*CardanoChainOperations, error) {
	config, err := cardanotx.NewCardanoChainConfig(jsonConfig)
	if err != nil {
		return nil, err
	}

	txProvider, err := config.CreateTxProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create tx provider: %w", err)
	}

	return &CardanoChainOperations{
		txProvider:       txProvider,
		cardanoCliBinary: cardanowallet.ResolveCardanoCliBinary(config.NetworkID),
		logger:           logger,
	}, nil
}

// SendTx implements core.ChainOperations.
func (cco *CardanoChainOperations) SendTx(
	ctx context.Context, smartContractData *eth.ConfirmedBatch,
) error {
	cco.logger.Debug("confirmed batch - sending tx", "batchID", smartContractData.ID, "binary", cco.cardanoCliBinary)

	witnesses := make(
		[][]byte, len(smartContractData.MultisigSignatures)+len(smartContractData.FeePayerMultisigSignatures))
	copy(witnesses, smartContractData.MultisigSignatures)
	copy(witnesses[len(smartContractData.MultisigSignatures):], smartContractData.FeePayerMultisigSignatures)

	txSigned, err := cardanotx.AssembleTxWitnesses(cco.cardanoCliBinary, smartContractData.RawTransaction, witnesses)
	if err != nil {
		return err
	}

	tip, err := cco.txProvider.GetTip(context.Background())
	if err == nil {
		cco.logger.Info("confirmed batch - sending tx current tip",
			"block", tip.Block, "slot", tip.Slot, "hash", tip.Hash)
	}

	info, err := indexer.ParseTxInfo(txSigned)
	if err == nil {
		cco.logger.Info("confirmed batch - sending tx",
			"hash", info.Hash, "ttl", info.TTL, "fee", info.Fee, "metadata", info.MetaData)
	} else {
		cco.logger.Error("confirmed batch - sending tx info error", "err", err)
	}

	return cco.txProvider.SubmitTx(context.Background(), txSigned)
}
