package validatorcomponents

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
)

type BridgingRequestStateManagerImpl struct {
	db     core.BridgingRequestStateDb
	logger hclog.Logger
}

var _ core.BridgingRequestStateManager = (*BridgingRequestStateManagerImpl)(nil)

func NewBridgingRequestStateManager(db core.BridgingRequestStateDb, logger hclog.Logger) (*BridgingRequestStateManagerImpl, error) {
	return &BridgingRequestStateManagerImpl{
		db:     db,
		logger: logger,
	}, nil
}

// New implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) New(sourceChainId string, tx *indexer.Tx) error {
	inputAddrs := make([]string, 0, len(tx.Inputs))
	for _, input := range tx.Inputs {
		inputAddrs = append(inputAddrs, input.Output.Address)
	}

	err := m.db.AddBridgingRequestState(core.NewBridgingRequestState(sourceChainId, tx.Hash, inputAddrs))
	if err != nil {
		m.logger.Error("failed to add new BridgingRequestState", "err", err)
		return fmt.Errorf("failed to add new BridgingRequestState. err: %w", err)
	}

	return nil
}

// NewMultiple implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) NewMultiple(sourceChainId string, txs []*indexer.Tx) error {
	hasErrors := false
	for _, tx := range txs {
		err := m.New(sourceChainId, tx)
		if err != nil {
			hasErrors = true
		}
	}

	if hasErrors {
		m.logger.Error("failed to add some new BridgingRequestStates")
		return fmt.Errorf("failed to add some new BridgingRequestStates")
	}

	return nil
}

// Invalid implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) Invalid(key common.BridgingRequestStateKey) error {
	return m.updateStateByKey(key, func(state *core.BridgingRequestState) error {
		return state.ToInvalidRequest()
	})
}

// SubmittedToBridge implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) SubmittedToBridge(key common.BridgingRequestStateKey, destinationChainId string) error {
	return m.updateStateByKey(key, func(state *core.BridgingRequestState) error {
		return state.ToSubmittedToBridge(destinationChainId)
	})
}

// IncludedInBatch implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) IncludedInBatch(destinationChainId string, batchId uint64, txs []common.BridgingRequestStateKey) error {
	hasErrors := false
	for _, key := range txs {
		state, err := m.db.GetBridgingRequestState(key.SourceChainId, key.SourceTxHash)
		if state == nil || err != nil {
			continue
		}

		if state.DestinationChainId != destinationChainId {
			m.logger.Error("batch destinationChainId not equal to BridgingRequestState.DestinationChainId", "sourceChainId", key.SourceChainId, "sourceTxHash", key.SourceTxHash)
			hasErrors = true
			continue
		}

		err = state.ToIncludedInBatch(batchId)
		if err != nil {
			m.logger.Error("failed to update a BridgingRequestState", "sourceChainId", key.SourceChainId, "sourceTxHash", key.SourceTxHash, "err", err)
			hasErrors = true
			continue
		}

		err = m.db.UpdateBridgingRequestState(state)
		if err != nil {
			m.logger.Error("failed to save updated BridgingRequestState", "sourceChainId", key.SourceChainId, "sourceTxHash", key.SourceTxHash, "err", err)
			hasErrors = true
		}
	}

	if hasErrors {
		m.logger.Error("failed to update some BridgingRequestStates")
		return fmt.Errorf("failed to update some BridgingRequestStates")
	}

	return nil
}

// SubmittedToDestination implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) SubmittedToDestination(destinationChainId string, batchId uint64) error {
	return m.updateStatesByBatchId(destinationChainId, batchId, func(state *core.BridgingRequestState) error {
		return state.ToSubmittedToDestination()
	})
}

// FailedToExecuteOnDestination implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) FailedToExecuteOnDestination(destinationChainId string, batchId uint64) error {
	return m.updateStatesByBatchId(destinationChainId, batchId, func(state *core.BridgingRequestState) error {
		return state.ToFailedToExecuteOnDestination()
	})
}

// ExecutedOnDestination implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) ExecutedOnDestination(destinationChainId string, batchId uint64, destinationTxHash string) error {
	return m.updateStatesByBatchId(destinationChainId, batchId, func(state *core.BridgingRequestState) error {
		return state.ToExecutedOnDestination(destinationTxHash)
	})
}

// Get implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) Get(sourceChainId string, sourceTxHash string) (*core.BridgingRequestState, error) {
	state, err := m.db.GetBridgingRequestState(sourceChainId, sourceTxHash)
	if err != nil {
		m.logger.Error("failed to get BridgingRequestState", "sourceChainId", sourceChainId, "sourceTxHash", sourceTxHash, "err", err)
		return nil, fmt.Errorf("failed to get BridgingRequestState. sourceChainId: %v, sourceTxHash: %v, err: %w", sourceChainId, sourceTxHash, err)
	}

	return state, nil
}

// GetAllForUser implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) GetAllForUser(sourceChainId string, userAddr string) ([]*core.BridgingRequestState, error) {
	states, err := m.db.GetUserBridgingRequestStates(sourceChainId, userAddr)
	if err != nil {
		m.logger.Error("failed to get all BridgingRequestStates for user", "sourceChainId", sourceChainId, "userAddr", userAddr, "err", err)
		return nil, fmt.Errorf("failed to get all BridgingRequestStates for user. sourceChainId: %v, userAddr: %v, err: %w", sourceChainId, userAddr, err)
	}

	return states, nil
}

func (m *BridgingRequestStateManagerImpl) updateStateByKey(key common.BridgingRequestStateKey, updateState func(state *core.BridgingRequestState) error) error {
	state, err := m.db.GetBridgingRequestState(key.SourceChainId, key.SourceTxHash)
	if err != nil {
		m.logger.Error("failed to get BridgingRequestState", "sourceChainId", key.SourceChainId, "sourceTxHash", key.SourceTxHash, "err", err)
		return fmt.Errorf("failed to get BridgingRequestState. sourceChainId: %v, sourceTxHash: %v, err: %w", key.SourceChainId, key.SourceTxHash, err)
	}

	if state == nil {
		m.logger.Error("trying to get a non-existent BridgingRequestState", "sourceChainId", key.SourceChainId, "sourceTxHash", key.SourceTxHash)
		return fmt.Errorf("trying to get a non-existent BridgingRequestState. sourceChainId: %v, sourceTxHash: %v", key.SourceChainId, key.SourceTxHash)
	}

	err = updateState(state)
	if err != nil {
		m.logger.Error("failed to update a BridgingRequestState", "sourceChainId", key.SourceChainId, "sourceTxHash", key.SourceTxHash, "err", err)
		return fmt.Errorf("failed to update a BridgingRequestState. sourceChainId: %v, sourceTxHash: %v, err: %w", key.SourceChainId, key.SourceTxHash, err)
	}

	err = m.db.UpdateBridgingRequestState(state)
	if err != nil {
		m.logger.Error("failed to save updated BridgingRequestState", "sourceChainId", key.SourceChainId, "sourceTxHash", key.SourceTxHash, "err", err)
		return fmt.Errorf("failed to save updated BridgingRequestState. sourceChainId: %v, sourceTxHash: %v, err: %w", key.SourceChainId, key.SourceTxHash, err)
	}

	return nil
}

func (m *BridgingRequestStateManagerImpl) updateStatesByBatchId(destinationChainId string, batchId uint64, updateState func(state *core.BridgingRequestState) error) error {
	states, err := m.db.GetBridgingRequestStatesByBatchId(destinationChainId, batchId)
	if err != nil {
		m.logger.Error("failed to get BridgingRequestStates", "destinationChainId", destinationChainId, "batchId", batchId, "err", err)
		return fmt.Errorf("failed to get BridgingRequestStates. destinationChainId: %v, batchId: %v, err: %w", destinationChainId, batchId, err)
	}

	hasErrors := false
	for _, state := range states {
		err := updateState(state)
		if err != nil {
			m.logger.Error("failed to update a BridgingRequestState", "sourceChainId", state.SourceChainId, "sourceTxHash", state.SourceTxHash, "err", err)
			hasErrors = true
			continue
		}

		err = m.db.UpdateBridgingRequestState(state)
		if err != nil {
			m.logger.Error("failed to save updated BridgingRequestState", "sourceChainId", state.SourceChainId, "sourceTxHash", state.SourceTxHash, "err", err)
			hasErrors = true
		}
	}

	if hasErrors {
		m.logger.Error("failed to update some BridgingRequestStates")
		return fmt.Errorf("failed to update some BridgingRequestStates")
	}

	return nil
}
