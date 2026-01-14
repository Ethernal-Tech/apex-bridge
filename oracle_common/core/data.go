package core

import (
	"encoding/binary"
	"math"
	"math/big"
	"reflect"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/ethgo"
)

const (
	LastProcessingPriority = uint8(1)
)

type DBTxID struct {
	ChainID string
	DBKey   []byte
}

func NewTypeRegisterWithChains(
	appConfig *AppConfig, cardanoType reflect.Type, ethType reflect.Type,
) common.TypeRegister {
	reg := common.NewTypeRegister()

	if cardanoType != nil {
		for _, chain := range appConfig.CardanoChains {
			reg.SetType(chain.ChainID, cardanoType)
		}
	}

	if ethType != nil {
		for _, chain := range appConfig.EthChains {
			reg.SetType(chain.ChainID, ethType)
		}
	}

	return reg
}

type NotEnoughFundsEvent = contractbinding.BridgeContractNotEnoughFunds

type SubmitClaimsEvents struct {
	NotEnoughFunds     []*NotEnoughFundsEvent
	BatchExecutionInfo []*DBBatchInfoEvent
}

type UpdateTxsData[
	TTx BaseTx,
	TProcessedTx BaseProcessedTx,
	TExpectedTx BaseExpectedTx,
] struct {
	ExpectedInvalid   []TExpectedTx
	ExpectedProcessed []TExpectedTx

	UpdateUnprocessed          []TTx             // if brc is rejected, need to update tryCount and lastTryTime
	MoveUnprocessedToPending   []TTx             // if brc is accepted, it moves to pending
	MoveUnprocessedToProcessed []TProcessedTx    // if its bec or brc that is invalid
	MovePendingToUnprocessed   []BaseTx          // for befc txs, also update tryCount and set lastTryTime to nil
	MovePendingToProcessed     []BaseProcessedTx // for bec txs
	AddBatchInfoEvents         []*DBBatchInfoEvent
	RemoveBatchInfoEvents      []*DBBatchInfoEvent
}

func (d *UpdateTxsData[TTx, TProcessedTx, TExpectedTx]) Count() int {
	return len(d.ExpectedInvalid) +
		len(d.ExpectedProcessed) +
		len(d.UpdateUnprocessed) +
		len(d.MoveUnprocessedToPending) +
		len(d.MoveUnprocessedToProcessed) +
		len(d.MovePendingToUnprocessed) +
		len(d.MovePendingToProcessed) +
		len(d.AddBatchInfoEvents) +
		len(d.RemoveBatchInfoEvents)
}

type DBBatchTx struct {
	SourceChainID           uint8       `json:"s_chain"`
	ObservedTransactionHash common.Hash `json:"s_tx_hash"`
	TransactionType         uint8       `json:"tx_type"`
}

type DBBatchInfoEvent struct {
	BatchID       uint64      `json:"batch"`
	DstChainID    uint8       `json:"chain"`
	DstTxHash     common.Hash `json:"tx_hash"`
	IsFailedClaim bool        `json:"failed"`
	TxHashes      []DBBatchTx `json:"txs"`
}

func NewDBBatchInfoEvent(
	batchID uint64, chainID uint8, txHash common.Hash, isFailedClaim bool, txs []eth.TxDataInfo,
) *DBBatchInfoEvent {
	dbBatchTxs := make([]DBBatchTx, len(txs))
	for i, tx := range txs {
		dbBatchTxs[i] = DBBatchTx{
			SourceChainID:           tx.SourceChainId,
			ObservedTransactionHash: tx.ObservedTransactionHash,
			TransactionType:         tx.TransactionType,
		}
	}

	return &DBBatchInfoEvent{
		BatchID:       batchID,
		DstChainID:    chainID,
		DstTxHash:     txHash,
		IsFailedClaim: isFailedClaim,
		TxHashes:      dbBatchTxs,
	}
}

func (e *DBBatchInfoEvent) DBKey() []byte {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, e.BatchID)

	return key
}

type ProcessedTxByInnerAction struct {
	ChainID         string     `json:"chain_id"`
	Hash            ethgo.Hash `json:"hash"`
	InnerActionHash ethgo.Hash `json:"ia_hash"`
}

func IsTxReady(triesCount uint32, lastTimeTried time.Time, settings RetryUnprocessedSettings) bool {
	if lastTimeTried.IsZero() || triesCount == 0 {
		return true
	}

	timeout := min(settings.BaseTimeout*time.Duration(math.Pow(2, float64(triesCount-1))),
		settings.MaxTimeout)

	return lastTimeTried.Add(timeout).Before(time.Now().UTC())
}

type BlocksSubmitterInfo struct {
	BlockNumOrSlot uint64 `json:"blockNumOrSlot"`
	CounterEmpty   int    `json:"counterEmpty"`
}

type ReceiverValidationContext struct {
	CardanoDestConfig *CardanoChainConfig
	EthDestConfig     *EthChainConfig

	BridgingSettings *BridgingSettings
	DestFeeAddress   string

	CurrencySrcID  uint16
	CurrencyDestID uint16

	AmountsSums map[uint16]*big.Int
}

type TotalTokensAmount struct {
	TotalAmountCurrencySrc *big.Int
	TotalAmountWrappedSrc  *big.Int
	TotalAmountCurrencyDst *big.Int
	TotalAmountWrappedDst  *big.Int
}

func NewTotalTokensAmount() *TotalTokensAmount {
	return &TotalTokensAmount{
		TotalAmountCurrencySrc: big.NewInt(0),
		TotalAmountWrappedSrc:  big.NewInt(0),
		TotalAmountCurrencyDst: big.NewInt(0),
		TotalAmountWrappedDst:  big.NewInt(0),
	}
}

func (totalTokensAmount *TotalTokensAmount) TrackSourceTokenAmount(
	sourceTokenID uint16,
	currencySrcID uint16,
	receiverAmountWei *big.Int,
	tokens map[uint16]common.Token,
) {
	if sourceTokenID == currencySrcID {
		// currency on source
		totalTokensAmount.TotalAmountCurrencySrc.Add(totalTokensAmount.TotalAmountCurrencySrc, receiverAmountWei)
	} else if tokens[sourceTokenID].IsWrappedCurrency {
		// source token is wrapped currency
		totalTokensAmount.TotalAmountWrappedSrc.Add(totalTokensAmount.TotalAmountWrappedSrc, receiverAmountWei)
	}
}

func (totalTokensAmount *TotalTokensAmount) TrackDestTokenAmount(
	receiverCurrencyWei *big.Int, receiverTokensWei *big.Int,
) {
	totalTokensAmount.TotalAmountCurrencyDst.Add(
		totalTokensAmount.TotalAmountCurrencyDst, receiverCurrencyWei)

	totalTokensAmount.TotalAmountWrappedDst.Add(
		totalTokensAmount.TotalAmountWrappedDst, receiverTokensWei)
}
