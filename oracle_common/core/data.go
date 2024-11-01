package core

import (
	"reflect"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/ethgo"
)

const (
	LastProcessingPriority = uint8(1)

	RetryUnprocessedAfterSec = 1 * 60 // 1 min
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
	ChainID         string     `json:"chain_id"`
	Hash            ethgo.Hash `json:"hash"`
	InnerActionHash ethgo.Hash `json:"ia_hash"`
}
