package cardanotx

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

const (
	splitStringLength = 40
	potentialFee      = 250_000

	retryWait       = time.Millisecond * 1000
	retriesMaxCount = 10

	retriesTxHashInUtxosCount = 60
	retriesTxHashInUtxosWait  = time.Millisecond * 4000
)

type BridgingTxSender struct {
	TxProviderSrc      cardanowallet.ITxProvider
	TxUtxoRetrieverDst cardanowallet.IUTxORetriever
	MultiSigAddrSrc    string
	TestNetMagicSrc    uint
	PotentialFee       uint64
	TTLSlotNumberInc   uint64
	ProtocolParameters []byte
}

func NewBridgingTxSender(
	txProvider cardanowallet.ITxProvider,
	txUtxoRetriever cardanowallet.IUTxORetriever,
	testNetMagic uint,
	multiSigAddr string,
	ttlSlotNumberInc uint64,
) *BridgingTxSender {
	return &BridgingTxSender{
		TxProviderSrc:      txProvider,
		TxUtxoRetrieverDst: txUtxoRetriever,
		TestNetMagicSrc:    testNetMagic,
		MultiSigAddrSrc:    multiSigAddr,
		PotentialFee:       potentialFee,
		TTLSlotNumberInc:   ttlSlotNumberInc,
	}
}

// CreateTx creates tx and returns cbor of raw transaction data, tx hash and error
func (bts *BridgingTxSender) CreateTx(
	ctx context.Context,
	chain string,
	senderAddr string,
	receivers []cardanowallet.TxOutput,
) ([]byte, string, error) {
	qtd, err := bts.TxProviderSrc.GetTip(ctx)
	if err != nil {
		return nil, "", err
	}

	protocolParams := bts.ProtocolParameters
	if protocolParams == nil {
		protocolParams, err = bts.TxProviderSrc.GetProtocolParameters(ctx)
		if err != nil {
			return nil, "", err
		}
	}

	metadata, err := bts.createMetadata(chain, senderAddr, receivers)
	if err != nil {
		return nil, "", err
	}

	outputsSum := cardanowallet.GetOutputsSum(receivers)
	outputs := []cardanowallet.TxOutput{
		{
			Addr:   bts.MultiSigAddrSrc,
			Amount: outputsSum,
		},
		{
			Addr: senderAddr,
		},
	}

	inputs, err := cardanowallet.GetUTXOsForAmount(
		ctx, bts.TxProviderSrc, senderAddr, outputsSum+bts.PotentialFee, cardanowallet.MinUTxODefaultValue)
	if err != nil {
		return nil, "", err
	}

	builder, err := cardanowallet.NewTxBuilder()
	if err != nil {
		return nil, "", err
	}

	defer builder.Dispose()

	builder.SetMetaData(metadata).
		SetProtocolParameters(protocolParams).
		SetTimeToLive(qtd.Slot + bts.TTLSlotNumberInc).
		SetTestNetMagic(bts.TestNetMagicSrc).
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
	ctx context.Context, cardanoWallet cardanowallet.IWallet, txRaw []byte, txHash string,
) error {
	txSigned, err := cardanowallet.SignTx(txRaw, txHash, cardanoWallet)
	if err != nil {
		return err
	}

	return cardanowallet.ExecuteWithRetry(ctx, retriesMaxCount, retryWait, func() (bool, error) {
		err := bts.TxProviderSrc.SubmitTx(ctx, txSigned)

		return err == nil, err
	}, isRecoverableError)
}

func (bts *BridgingTxSender) WaitForTx(
	ctx context.Context, receivers []cardanowallet.TxOutput,
) error {
	errs := make([]error, len(receivers))
	wg := sync.WaitGroup{}

	for i, x := range receivers {
		wg.Add(1)

		go func(idx int, recv cardanowallet.TxOutput) {
			defer wg.Done()

			var expectedAmount *big.Int

			errs[idx] = cardanowallet.ExecuteWithRetry(ctx, retriesMaxCount, retryWait, func() (bool, error) {
				utxos, err := bts.TxUtxoRetrieverDst.GetUtxos(ctx, recv.Addr)
				expectedAmount = cardanowallet.GetUtxosSum(utxos)

				return err == nil, err
			}, isRecoverableError)

			if errs[idx] != nil {
				return
			}

			expectedAmount.Add(expectedAmount, new(big.Int).SetUint64(recv.Amount))

			errs[idx] = cardanowallet.WaitForAmount(
				ctx, bts.TxUtxoRetrieverDst, recv.Addr, func(newAmount *big.Int) bool {
					return newAmount.Cmp(expectedAmount) >= 0
				},
				retriesTxHashInUtxosCount, retriesTxHashInUtxosWait, isRecoverableError)
		}(i, x)
	}

	wg.Wait()

	return errors.Join(errs...)
}

func (bts *BridgingTxSender) createMetadata(
	chain, senderAddr string, receivers []cardanowallet.TxOutput,
) ([]byte, error) {
	metadataObj := common.BridgingRequestMetadata{
		BridgingTxType:     common.BridgingTxTypeBridgingRequest,
		DestinationChainID: chain,
		SenderAddr:         common.SplitString(senderAddr, splitStringLength),
		Transactions:       make([]common.BridgingRequestMetadataTransaction, 0, len(receivers)+1),
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

func isRecoverableError(err error) bool {
	return strings.Contains(err.Error(), "status code 500") // retry if error is ogmios "status code 500"
}
