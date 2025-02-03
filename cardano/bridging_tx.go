package cardanotx

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/Ethernal-Tech/apex-bridge/common"
	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

const (
	DefaultPotentialFee = 450_000
	splitStringLength   = 40
	maxInputs           = 40
)

type BridgingTxSender struct {
	cardanoCliBinary   string
	txProviderSrc      cardanowallet.ITxProvider
	txUtxoRetrieverDst cardanowallet.IUTxORetriever
	multiSigAddrSrc    string
	testNetMagicSrc    uint
	potentialFee       uint64
	ttlSlotNumberInc   uint64
	protocolParameters []byte
}

func NewBridgingTxSender(
	cardanoCliBinary string,
	txProvider cardanowallet.ITxProvider,
	txUtxoRetriever cardanowallet.IUTxORetriever,
	testNetMagic uint,
	multiSigAddr string,
	ttlSlotNumberInc uint64,
	potentialFee uint64,
) *BridgingTxSender {
	return &BridgingTxSender{
		cardanoCliBinary:   cardanoCliBinary,
		txProviderSrc:      txProvider,
		txUtxoRetrieverDst: txUtxoRetriever,
		testNetMagicSrc:    testNetMagic,
		multiSigAddrSrc:    multiSigAddr,
		potentialFee:       potentialFee,
		ttlSlotNumberInc:   ttlSlotNumberInc,
	}
}

// CreateTx creates tx and returns cbor of raw transaction data, tx hash and error
func (bts *BridgingTxSender) CreateTx(
	ctx context.Context,
	chain string,
	senderAddr string,
	receivers []cardanowallet.TxOutput,
	feeBridgeAmount uint64,
	minUtxoValue uint64,
) ([]byte, string, error) {
	qtd, err := bts.txProviderSrc.GetTip(ctx)
	if err != nil {
		return nil, "", err
	}

	protocolParams := bts.protocolParameters
	if protocolParams == nil {
		protocolParams, err = bts.txProviderSrc.GetProtocolParameters(ctx)
		if err != nil {
			return nil, "", err
		}
	}

	builder, err := cardanowallet.NewTxBuilder(bts.cardanoCliBinary)
	if err != nil {
		return nil, "", err
	}

	defer builder.Dispose()

	metadata, err := bts.createMetadata(chain, senderAddr, receivers, feeBridgeAmount)
	if err != nil {
		return nil, "", err
	}

	allUtxos, err := bts.txProviderSrc.GetUtxos(ctx, senderAddr)
	if err != nil {
		return nil, "", err
	}

	potentialTokenCost, err := cardanowallet.GetTokenCostSum(builder, senderAddr, allUtxos)
	if err != nil {
		return nil, "", fmt.Errorf("failed to retrieve token cost sum. err: %w", err)
	}

	minUtxoValue = max(minUtxoValue, potentialTokenCost)
	outputsSum := cardanowallet.GetOutputsSum(receivers)
	outputsSumLovelace := outputsSum[cardanowallet.AdaTokenName] + feeBridgeAmount
	desiredSumLovelace := outputsSumLovelace + bts.potentialFee + minUtxoValue

	inputs, err := getUTXOsForAmount(allUtxos, desiredSumLovelace, maxInputs)
	if err != nil {
		return nil, "", err
	}

	outputs := []cardanowallet.TxOutput{
		{
			Addr:   bts.multiSigAddrSrc,
			Amount: outputsSumLovelace,
		},
		{
			Addr: senderAddr,
		},
	}

	builder.SetMetaData(metadata).
		SetProtocolParameters(protocolParams).
		SetTimeToLive(qtd.Slot + bts.ttlSlotNumberInc).
		SetTestNetMagic(bts.testNetMagicSrc).
		AddInputs(inputs.Inputs...).
		AddOutputs(outputs...)

	fee, err := builder.CalculateFee(0)
	if err != nil {
		return nil, "", err
	}

	outputsSum[cardanowallet.AdaTokenName] += fee

	changeTxOutput, err := cardanowallet.CreateTxOutputChange(cardanowallet.TxOutput{
		Addr: senderAddr,
	}, inputs.Sum, outputsSum)
	if err != nil {
		return nil, "", err
	}

	if changeTxOutput.Amount > 0 || len(changeTxOutput.Tokens) > 0 {
		builder.ReplaceOutput(-1, changeTxOutput)
	} else {
		builder.RemoveOutput(-1)
	}

	builder.SetFee(fee)

	return builder.Build()
}

func (bts *BridgingTxSender) SendTx(
	ctx context.Context, txRaw []byte, cardanoWallet cardanowallet.ITxSigner,
) error {
	builder, err := cardanowallet.NewTxBuilder(bts.cardanoCliBinary)
	if err != nil {
		return err
	}

	defer builder.Dispose()

	txSigned, err := builder.SignTx(txRaw, []cardanowallet.ITxSigner{cardanoWallet})
	if err != nil {
		return err
	}

	_, err = infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (bool, error) {
		return true, bts.txProviderSrc.SubmitTx(ctx, txSigned)
	})

	return err
}

func (bts *BridgingTxSender) WaitForTx(
	ctx context.Context, receivers []cardanowallet.TxOutput, tokenName string,
) error {
	return WaitForTx(ctx, bts.txUtxoRetrieverDst, receivers, tokenName)
}

func (bts *BridgingTxSender) createMetadata(
	chain, senderAddr string, receivers []cardanowallet.TxOutput, feeAmount uint64,
) ([]byte, error) {
	metadataObj := common.BridgingRequestMetadata{
		BridgingTxType:     common.BridgingTxTypeBridgingRequest,
		DestinationChainID: chain,
		SenderAddr:         common.SplitString(senderAddr, splitStringLength),
		Transactions:       make([]common.BridgingRequestMetadataTransaction, 0, len(receivers)+1),
		FeeAmount:          feeAmount,
	}

	for _, x := range receivers {
		metadataObj.Transactions = append(metadataObj.Transactions, common.BridgingRequestMetadataTransaction{
			Address: common.SplitString(x.Addr, splitStringLength),
			Amount:  x.Amount,
		})
	}

	return common.MarshalMetadata(common.MetadataEncodingTypeJSON, metadataObj)
}

func WaitForTx(
	ctx context.Context, txUtxoRetriever cardanowallet.IUTxORetriever,
	receivers []cardanowallet.TxOutput, tokenName string,
) error {
	errs := make([]error, len(receivers))
	wg := sync.WaitGroup{}

	for i, x := range receivers {
		wg.Add(1)

		go func(idx int, recv cardanowallet.TxOutput) {
			defer wg.Done()

			_, errs[idx] = common.WaitForAmount(
				ctx, new(big.Int).SetUint64(recv.Amount), func(ctx context.Context) (*big.Int, error) {
					utxos, err := txUtxoRetriever.GetUtxos(ctx, recv.Addr)
					if err != nil {
						return nil, err
					}

					sum := cardanowallet.GetUtxosSum(utxos)

					return new(big.Int).SetUint64(sum[tokenName]), nil
				})
		}(i, x)
	}

	wg.Wait()

	return errors.Join(errs...)
}

func getUTXOsForAmount(
	utxos []cardanowallet.Utxo,
	desiredSumLovelace uint64,
	maxInputs int,
) (cardanowallet.TxInputs, error) {
	findMinUtxo := func(utxos []cardanowallet.Utxo) (cardanowallet.Utxo, int) {
		minUtxo := utxos[0]
		idx := 0

		for i, utxo := range utxos[1:] {
			if utxo.Amount < minUtxo.Amount {
				minUtxo = utxo
				idx = i + 1
			}
		}

		return minUtxo, idx
	}

	utxos2TxInputs := func(utxos []cardanowallet.Utxo) []cardanowallet.TxInput {
		inputs := make([]cardanowallet.TxInput, len(utxos))
		for i, x := range utxos {
			inputs[i] = cardanowallet.TxInput{
				Hash:  x.Hash,
				Index: x.Index,
			}
		}

		return inputs
	}

	// Loop through utxos to find first input with enough tokens
	// If we don't have this UTXO we need to use more of them
	//nolint:prealloc
	var (
		currentSum  = map[string]uint64{}
		chosenUTXOs []cardanowallet.Utxo
		tokenName   = cardanowallet.AdaTokenName
	)

	for _, utxo := range utxos {
		currentSum[tokenName] += utxo.Amount

		for _, token := range utxo.Tokens {
			currentSum[token.TokenName()] += token.Amount
		}

		chosenUTXOs = append(chosenUTXOs, utxo)

		if len(chosenUTXOs) > maxInputs {
			lastIdx := len(chosenUTXOs) - 1
			minChosenUTXO, minChosenUTXOIdx := findMinUtxo(chosenUTXOs)

			chosenUTXOs[minChosenUTXOIdx] = chosenUTXOs[lastIdx]
			chosenUTXOs = chosenUTXOs[:lastIdx]
			currentSum[tokenName] -= minChosenUTXO.Amount

			for _, token := range minChosenUTXO.Tokens {
				currentSum[token.TokenName()] -= token.Amount
			}
		}

		if currentSum[tokenName] >= desiredSumLovelace {
			return cardanowallet.TxInputs{
				Inputs: utxos2TxInputs(chosenUTXOs),
				Sum:    currentSum,
			}, nil
		}
	}

	return cardanowallet.TxInputs{}, fmt.Errorf(
		"not enough funds for the transaction: (available, desired) = (%d, %d)",
		currentSum[tokenName], desiredSumLovelace)
}
