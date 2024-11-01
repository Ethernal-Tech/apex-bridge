package core

import (
	"reflect"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
)

const (
	LastProcessingPriority = uint8(1)

	RetryUnprocessedAfterSec = 15 * 60 // 15 min
)

type DBTxID struct {
	ChainID string
	DBKey   []byte
}

type TxTypeRegister struct {
	TTxTypes map[string]reflect.Type
}

func (r *TxTypeRegister) SetTTxTypes(appConfig *AppConfig, cardanoType reflect.Type, ethType reflect.Type) {
	r.TTxTypes = make(map[string]reflect.Type)

	if cardanoType != nil {
		for _, chain := range appConfig.CardanoChains {
			r.TTxTypes[chain.ChainID] = cardanoType
		}
	}

	if ethType != nil {
		for _, chain := range appConfig.EthChains {
			r.TTxTypes[chain.ChainID] = ethType
		}
	}
}

type NotEnoughFundsEvent = contractbinding.BridgeContractNotEnoughFunds
type BatchExecutionInfoEvent = contractbinding.BridgeContractBatchExecutionInfo

type SubmitClaimsEvents struct {
	NotEnoughFunds     []*NotEnoughFundsEvent
	BatchExecutionInfo []*BatchExecutionInfoEvent
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
}

func (d *UpdateTxsData[TTx, TProcessedTx, TExpectedTx]) Count() int {
	return len(d.ExpectedInvalid) +
		len(d.ExpectedProcessed) +
		len(d.UpdateUnprocessed) +
		len(d.MoveUnprocessedToPending) +
		len(d.MoveUnprocessedToProcessed) +
		len(d.MovePendingToUnprocessed) +
		len(d.MovePendingToProcessed)
}

type ProcessedTxByInnerAction struct {
	ChainID         string `json:"chain_id"`
	Hash            []byte `json:"hash"`
	InnerActionHash []byte `json:"ia_hash"`
}
