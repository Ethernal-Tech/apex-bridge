package cardanotx

import (
	"context"
	"fmt"

	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type TxInputInfos struct {
	TestNetMagic uint
	MultiSig     *TxInputInfo
	MultiSigFee  *TxInputInfo
}

func NewTxInputInfos(keyHashesMultiSig []string, keyHashesMultiSigFee []string, testNetMagic uint) (*TxInputInfos, error) {
	result := [2]*TxInputInfo{}

	for i, keyHashes := range [][]string{keyHashesMultiSig, keyHashesMultiSigFee} {
		ps, err := cardanowallet.NewPolicyScript(keyHashes, len(keyHashes)*2/3+1)
		if err != nil {
			return nil, err
		}

		addr, err := ps.CreateMultiSigAddress(testNetMagic)
		if err != nil {
			return nil, err
		}

		result[i] = &TxInputInfo{
			PolicyScript: ps,
			Address:      addr,
		}
	}

	return &TxInputInfos{
		TestNetMagic: testNetMagic,
		MultiSig:     result[0],
		MultiSigFee:  result[1],
	}, nil
}

func (txinfos *TxInputInfos) Calculate(utxos, utxosFee []cardanowallet.Utxo, desired, desiredFee uint64) error {
	if err := txinfos.MultiSig.Calculate(utxos, desired); err != nil {
		return err
	}

	return txinfos.MultiSigFee.Calculate(utxosFee, desiredFee)
}

func (txinfos *TxInputInfos) CalculateWithRetriever(
	ctx context.Context, retriever cardanowallet.IUTxORetriever, desired, desiredFee uint64,
) error {
	if err := txinfos.MultiSig.CalculateWithRetriever(ctx, retriever, desired); err != nil {
		return err
	}

	return txinfos.MultiSigFee.CalculateWithRetriever(ctx, retriever, desiredFee)
}

type TxInputUTXOs struct {
	Inputs    []cardanowallet.TxInput
	InputsSum uint64
}

type TxInputInfo struct {
	TxInputUTXOs
	PolicyScript *cardanowallet.PolicyScript
	Address      string
}

func (txinfo *TxInputInfo) Calculate(utxos []cardanowallet.Utxo, desired uint64) error {
	// Loop through utxos to find first input with enough tokens
	// If we don't have this UTXO we need to use more of them
	var amountSum = uint64(0)
	chosenUTXOs := make([]cardanowallet.TxInput, 0, len(utxos))

	for _, utxo := range utxos {
		if utxo.Amount >= desired {
			txinfo.Inputs = []cardanowallet.TxInput{
				{
					Hash:  utxo.Hash,
					Index: utxo.Index,
				},
			}
			txinfo.InputsSum = utxo.Amount

			return nil
		}

		amountSum += utxo.Amount
		chosenUTXOs = append(chosenUTXOs, cardanowallet.TxInput{
			Hash:  utxo.Hash,
			Index: utxo.Index,
		})

		if amountSum >= desired {
			txinfo.Inputs = chosenUTXOs
			txinfo.InputsSum = amountSum

			return nil
		}
	}

	return fmt.Errorf("not enough funds to generate the transaction: %d available vs %d required", amountSum, desired)
}

func (txinfo *TxInputInfo) CalculateWithRetriever(
	ctx context.Context, retriever cardanowallet.IUTxORetriever, desired uint64,
) error {
	utxos, err := retriever.GetUtxos(ctx, txinfo.Address)
	if err != nil {
		return err
	}

	return txinfo.Calculate(utxos, desired)
}
