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
	config     core.CardanoChainConfig
	txProvider cardanowallet.ITxSubmitter
}

func NewCardanoChainOperations(
	txProvider cardanowallet.ITxSubmitter, config core.CardanoChainConfig,
) *CardanoChainOperations {
	return &CardanoChainOperations{
		txProvider: txProvider,
		config:     config,
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

	return cco.txProvider.SubmitTx(context.Background(), txSigned)
}
