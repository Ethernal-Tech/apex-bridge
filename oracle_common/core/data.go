package core

const (
	LastProcessingPriority = uint8(1)
)

type SubmitClaimsEvents struct {
	// NotEnoughFunds event list
	// BatchExecutionInfo event list
}

type UpdateTxsData[
	TTx BaseTx,
	TProcessedTx BaseProcessedTx,
	TExpectedTx BaseExpectedTx,
] struct {
	ExpectedInvalid   []TExpectedTx
	ExpectedProcessed []TExpectedTx

	UpdateUnprocessed          []TTx          // if brc is rejected, need to update tryCount and lastTryTime
	MoveUnprocessedToPending   []TTx          // if brc is accepted, it moves to pending
	MoveUnprocessedToProcessed []TProcessedTx // if its bec or brc that is invalid
	MovePendingToUnprocessed   []TTx          // for befc txs, also update tryCount and set lastTryTime to nil
	MovePendingToProcessed     []TProcessedTx // for bec txs
}
