package relayer

import (
	"context"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

var _ core.ChainOperations = (*CardanoChainOperations)(nil)

type CardanoChainOperations struct {
	config core.CardanoChainConfig
}

func NewCardanoChainOperations(config core.CardanoChainConfig) *CardanoChainOperations {
	return &CardanoChainOperations{
		config: config,
	}
}

// SendTx implements core.ChainOperations.
func (cco *CardanoChainOperations) SendTx(smartContractData *eth.ConfirmedBatch) error {
	witnesses := make([][]byte, len(smartContractData.MultisigSignatures)+len(smartContractData.FeePayerMultisigSignatures))
	copy(witnesses, smartContractData.MultisigSignatures)
	copy(witnesses[len(smartContractData.MultisigSignatures):], smartContractData.FeePayerMultisigSignatures)

	txSigned, err := cardanotx.AssembleTxWitnesses(smartContractData.RawTransaction, witnesses)
	if err != nil {
		return err
	}

	txProvider, err := cardanowallet.NewTxProviderBlockFrost(cco.config.BlockfrostUrl, cco.config.BlockfrostAPIKey)
	if err != nil {
		return err
	}

	return txProvider.SubmitTx(context.Background(), txSigned)
}
