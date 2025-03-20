package cardanotx

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"sync"

	"github.com/Ethernal-Tech/apex-bridge/common"
	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

const (
	splitStringLength = 40
	maxInputs         = 40
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
	qtd, err := infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (cardanowallet.QueryTipData, error) {
		return bts.txProviderSrc.GetTip(ctx)
	})
	if err != nil {
		return nil, "", err
	}

	protocolParams := bts.protocolParameters
	if protocolParams == nil {
		protocolParams, err = infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) ([]byte, error) {
			return bts.txProviderSrc.GetProtocolParameters(ctx)
		})
		if err != nil {
			return nil, "", err
		}
	}

	allUtxos, err := infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) ([]cardanowallet.Utxo, error) {
		return bts.txProviderSrc.GetUtxos(ctx, senderAddr)
	})
	if err != nil {
		return nil, "", err
	}

	// utxos without tokens should come first
	sort.Slice(allUtxos, func(i, j int) bool {
		return len(allUtxos[i].Tokens) < len(allUtxos[j].Tokens)
	})

	metadata, err := bts.createMetadata(chain, senderAddr, receivers, feeBridgeAmount)
	if err != nil {
		return nil, "", err
	}

	builder, err := cardanowallet.NewTxBuilder(bts.cardanoCliBinary)
	if err != nil {
		return nil, "", err
	}

	defer builder.Dispose()

	builder.SetMetaData(metadata).
		SetProtocolParameters(protocolParams).
		SetTimeToLive(qtd.Slot + bts.ttlSlotNumberInc).
		SetTestNetMagic(bts.testNetMagicSrc)

	potentialTokenCost, err := cardanowallet.GetTokenCostSum(builder, senderAddr, allUtxos)
	if err != nil {
		return nil, "", fmt.Errorf("failed to retrieve token cost sum. err: %w", err)
	}

	minUtxoValue = max(minUtxoValue, potentialTokenCost)
	outputsSum := cardanowallet.GetOutputsSum(receivers)
	outputsSumLovelace := outputsSum[cardanowallet.AdaTokenName] + feeBridgeAmount
	desiredSumLovelace := outputsSumLovelace + bts.potentialFee + minUtxoValue

	inputs, err := cardanowallet.GetUTXOsForAmount(
		allUtxos, cardanowallet.AdaTokenName, desiredSumLovelace, maxInputs)
	if err != nil {
		return nil, "", err
	}

	tokens, err := cardanowallet.GetTokensFromSumMap(inputs.Sum)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create tokens from sum map. err: %w", err)
	}

	builder.AddInputs(inputs.Inputs...).AddOutputs(cardanowallet.TxOutput{
		Addr:   bts.multiSigAddrSrc,
		Amount: outputsSumLovelace,
	}, cardanowallet.TxOutput{
		Addr:   senderAddr,
		Tokens: tokens,
	})

	fee, err := builder.CalculateFee(1)
	if err != nil {
		return nil, "", err
	}

	// add bridging fee and calculated tx fee to lovelace output in order to calculate good change tx output
	outputsSum[cardanowallet.AdaTokenName] += fee + feeBridgeAmount

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
		SenderAddr:         addrToMetaDataAddr(senderAddr),
		Transactions:       make([]common.BridgingRequestMetadataTransaction, 0, len(receivers)+1),
		FeeAmount:          feeAmount,
	}

	for _, x := range receivers {
		metadataObj.Transactions = append(metadataObj.Transactions, common.BridgingRequestMetadataTransaction{
			Address: addrToMetaDataAddr(x.Addr),
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
