package core

import (
	"encoding/binary"
	"math"
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
	SourceChainID           uint8    `json:"s_chain"`
	ObservedTransactionHash [32]byte `json:"s_tx_hash"`
}

type DBBatchInfoEvent struct {
	BatchID       uint64      `json:"batch"`
	ChainID       uint8       `json:"chain"`
	IsFailedClaim bool        `json:"failed"`
	TxHashes      []DBBatchTx `json:"txs"`
}

func NewDBBatchInfoEvent(
	chainID uint8, batchID uint64, isFailedClaim bool, txs []eth.TxDataInfo,
) *DBBatchInfoEvent {
	result := make([]DBBatchTx, len(txs))
	for i, tx := range txs {
		result[i] = DBBatchTx{
			SourceChainID:           tx.SourceChainId,
			ObservedTransactionHash: tx.ObservedTransactionHash,
		}
	}

	return &DBBatchInfoEvent{
		BatchID:       batchID,
		ChainID:       chainID,
		IsFailedClaim: isFailedClaim,
		TxHashes:      result,
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
