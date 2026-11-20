package cardanotx

import (
	"errors"
	"fmt"

	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

// ErrUnpayableUnknownTokensRefund marks a UTxO with unknown native tokens that cannot be refunded on
// its own: its amount does not cover the fee, or what is left after the fee is below the min UTxO
// the ledger requires for the refund output.
var ErrUnpayableUnknownTokensRefund = errors.New("unknown tokens utxo cannot be refunded")

// CreateUnknownTokensRefundTxOutput builds the output that returns a UTxO holding unknown native
// tokens to receiverAddr: all of the UTxO's tokens plus its amount minus fee. The fee is charged
// per UTxO, so the oracle and the batcher must both check every such UTxO with this function.
// txBuilder must have protocol parameters set. Errors other than ErrUnpayableUnknownTokensRefund
// come from the min UTxO calculation.
func CreateUnknownTokensRefundTxOutput(
	txBuilder *wallet.TxBuilder, receiverAddr string, utxo *indexer.TxOutput, fee uint64,
) (wallet.TxOutput, error) {
	if utxo.Amount < fee {
		return wallet.TxOutput{}, fmt.Errorf("%w: amount %d is less than fee %d",
			ErrUnpayableUnknownTokensRefund, utxo.Amount, fee)
	}

	tokens := make([]wallet.TokenAmount, len(utxo.Tokens))
	for i, t := range utxo.Tokens {
		tokens[i] = wallet.NewTokenAmount(wallet.NewToken(t.PolicyID, t.Name), t.Amount)
	}

	output := wallet.TxOutput{
		Addr:   receiverAddr,
		Amount: utxo.Amount - fee,
		Tokens: tokens,
	}

	minUtxo, err := txBuilder.CalculateMinUtxo(wallet.TxOutputWithRefScript{TxOutput: output})
	if err != nil {
		return wallet.TxOutput{}, fmt.Errorf("failed to calculate min utxo for refund output: %w", err)
	}

	if output.Amount < minUtxo {
		return wallet.TxOutput{}, fmt.Errorf("%w: amount after fee %d is less than min utxo %d",
			ErrUnpayableUnknownTokensRefund, output.Amount, minUtxo)
	}

	return output, nil
}
