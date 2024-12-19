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
	DefaultPotentialFee = 250_000
	splitStringLength   = 40
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
	feeAmount uint64,
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

	metadata, err := bts.createMetadata(chain, senderAddr, receivers, feeAmount)
	if err != nil {
		return nil, "", err
	}

	outputsSum := cardanowallet.GetOutputsSum(receivers)[cardanowallet.AdaTokenName] + feeAmount
	desiredSum := outputsSum + bts.potentialFee + minUtxoValue

	inputs, err := bts.GetUTXOsForAmount(
		ctx, bts.txProviderSrc, senderAddr, cardanowallet.AdaTokenName, desiredSum, desiredSum)
	if err != nil {
		return nil, "", err
	}

	tokens, err := cardanowallet.GetTokensFromSumMap(inputs.Sum)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create tokens from sum map. err: %w", err)
	}

	outputs := []cardanowallet.TxOutput{
		{
			Addr:   bts.multiSigAddrSrc,
			Amount: outputsSum,
		},
		{
			Addr:   senderAddr,
			Tokens: tokens,
		},
	}

	builder, err := cardanowallet.NewTxBuilder(bts.cardanoCliBinary)
	if err != nil {
		return nil, "", err
	}

	defer builder.Dispose()

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

	inputsAdaSum := inputs.Sum[cardanowallet.AdaTokenName]
	change := inputsAdaSum - outputsSum - fee
	// handle overflow or insufficient amount
	if change > inputsAdaSum || change < minUtxoValue {
		return []byte{}, "", fmt.Errorf("insufficient amount %d for %d or min utxo not satisfied",
			inputsAdaSum, outputsSum+fee)
	}

	builder.UpdateOutputAmount(-1, change)

	builder.SetFee(fee)

	return builder.Build()
}

func (bts *BridgingTxSender) GetUTXOsForAmount(
	ctx context.Context, retriever cardanowallet.IUTxORetriever, addr string,
	tokenName string, exactSum uint64, atLeastSum uint64,
) (cardanowallet.TxInputs, error) {
	utxos, err := retriever.GetUtxos(ctx, addr)
	if err != nil {
		return cardanowallet.TxInputs{}, err
	}

	// Loop through utxos to find first input with enough tokens
	// If we don't have this UTXO we need to use more of them
	//nolint:prealloc
	var (
		currentSum  = map[string]uint64{}
		chosenUTXOs []cardanowallet.TxInput
	)

	for _, utxo := range utxos {
		currentSum[cardanowallet.AdaTokenName] += utxo.Amount

		for _, token := range utxo.Tokens {
			currentSum[token.TokenName()] += token.Amount
		}

		chosenUTXOs = append(chosenUTXOs, cardanowallet.TxInput{
			Hash:  utxo.Hash,
			Index: utxo.Index,
		})

		if currentSum[tokenName] == exactSum || currentSum[tokenName] >= atLeastSum {
			return cardanowallet.TxInputs{
				Inputs: chosenUTXOs,
				Sum:    currentSum,
			}, nil
		}
	}

	return cardanowallet.TxInputs{}, fmt.Errorf(
		"not enough funds for the transaction: (available, exact, at least) = (%d, %d, %d)",
		currentSum[tokenName], exactSum, atLeastSum)
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
		FeeAmount: common.BridgingRequestMetadataAmount{
			DestinationCurrencyAmount: feeAmount,
		},
	}

	for _, x := range receivers {
		metadataObj.Transactions = append(metadataObj.Transactions, common.BridgingRequestMetadataTransaction{
			Address: common.SplitString(x.Addr, splitStringLength),
			Amount:  x.Amount,
		})
	}

	return common.MarshalMetadata(common.MetadataEncodingTypeJSON, metadataObj)
}

func IsAddressInOutputs(
	receivers []cardanowallet.TxOutput, addr string,
) bool {
	for _, x := range receivers {
		if x.Addr == addr {
			return true
		}
	}

	return false
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
