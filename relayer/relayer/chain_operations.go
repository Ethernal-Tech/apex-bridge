package relayer

import (
	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/relayer/bridge"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

var _ core.ChainOperations = (*CardanoChainOperations)(nil)

type CardanoChainOperations struct {
	txProvider *cardanowallet.TxProviderBlockFrost
}

func NewCardanoChainOperations(txProvider *cardanowallet.TxProviderBlockFrost) *CardanoChainOperations {
	return &CardanoChainOperations{
		txProvider: txProvider,
	}
}

// SendTx implements core.ChainOperations.
func (cco *CardanoChainOperations) SendTx(smartContractData *bridge.ConfirmedBatch) error {
	witnesses := make([][]byte, len(smartContractData.MultisigSignatures)+len(smartContractData.FeePayerMultisigSignatures))
	copy(witnesses, smartContractData.MultisigSignatures)
	copy(witnesses[len(smartContractData.MultisigSignatures):], smartContractData.FeePayerMultisigSignatures)

	txSigned, err := cardanotx.AssembleTxWitnesses(smartContractData.RawTransaction, witnesses)
	if err != nil {
		return err
	}

	if err := cco.txProvider.SubmitTx(txSigned); err != nil {
		return err
	}

	return nil
}
