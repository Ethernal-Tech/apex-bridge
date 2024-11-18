package cardanotx

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

const (
	DefaultPotentialFee = 250_000
	splitStringLength   = 40

	retryCountForAmount    = 144
	retryWaitTimeForAMount = time.Second * 5
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

	outputsSum := cardanowallet.GetOutputsSum(receivers) + feeAmount
	outputs := []cardanowallet.TxOutput{
		{
			Addr:   bts.multiSigAddrSrc,
			Amount: outputsSum,
		},
		{
			Addr: senderAddr,
		},
	}

	desiredSum := outputsSum + bts.potentialFee + cardanowallet.MinUTxODefaultValue

	inputs, err := cardanowallet.GetUTXOsForAmount(ctx, bts.txProviderSrc, senderAddr, desiredSum, desiredSum)
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
		SetTestNetMagic(bts.testNetMagicSrc).
		AddInputs(inputs.Inputs...).
		AddOutputs(outputs...)

	fee, err := builder.CalculateFee(0)
	if err != nil {
		return nil, "", err
	}

	change := inputs.Sum - outputsSum - fee
	// handle overflow or insufficient amount
	if change > inputs.Sum || (change > 0 && change < cardanowallet.MinUTxODefaultValue) {
		return []byte{}, "", fmt.Errorf("insufficient amount %d for %d or min utxo not satisfied",
			inputs.Sum, outputsSum+fee)
	}

	if change == 0 {
		builder.RemoveOutput(-1)
	} else {
		builder.UpdateOutputAmount(-1, change)
	}

	builder.SetFee(fee)

	return builder.Build()
}

func (bts *BridgingTxSender) SendTx(
	ctx context.Context, txRaw []byte, txHash string, cardanoWallet cardanowallet.IWallet,
) error {
	builder, err := cardanowallet.NewTxBuilder(bts.cardanoCliBinary)
	if err != nil {
		return err
	}

	defer builder.Dispose()

	witness, err := cardanowallet.CreateTxWitness(txHash, cardanoWallet)
	if err != nil {
		return err
	}

	txSigned, err := builder.AssembleTxWitnesses(txRaw, [][]byte{witness})
	if err != nil {
		return err
	}

	_, err = infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (bool, error) {
		return true, bts.txProviderSrc.SubmitTx(ctx, txSigned)
	})

	return err
}

func (bts *BridgingTxSender) WaitForTx(
	ctx context.Context, receivers []cardanowallet.TxOutput,
) error {
	return WaitForTx(ctx, bts.txUtxoRetrieverDst, receivers)
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
	ctx context.Context, txUtxoRetriever cardanowallet.IUTxORetriever, receivers []cardanowallet.TxOutput,
) error {
	errs := make([]error, len(receivers))
	wg := sync.WaitGroup{}

	for i, x := range receivers {
		wg.Add(1)

		go func(idx int, recv cardanowallet.TxOutput) {
			defer wg.Done()

			originalAmount, err := infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (uint64, error) {
				utxos, err := txUtxoRetriever.GetUtxos(ctx, recv.Addr)
				if err != nil {
					return 0, err
				}

				return cardanowallet.GetUtxosSum(utxos), nil
			})
			if err != nil {
				errs[idx] = err

				return
			}

			_, errs[idx] = infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (bool, error) {
				utxos, err := txUtxoRetriever.GetUtxos(ctx, recv.Addr)
				if err != nil {
					return false, err
				}

				if cardanowallet.GetUtxosSum(utxos) < recv.Amount+originalAmount {
					return false, infracommon.ErrRetryTryAgain
				}

				return true, nil
			}, infracommon.WithRetryCount(retryCountForAmount), infracommon.WithRetryWaitTime(retryWaitTimeForAMount))
		}(i, x)
	}

	wg.Wait()

	return errors.Join(errs...)
}
