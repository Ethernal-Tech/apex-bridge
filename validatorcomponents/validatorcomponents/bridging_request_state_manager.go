package validatorcomponents

import (
	"errors"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/hashicorp/go-hclog"
)

type BridgingRequestStateManagerImpl struct {
	db     core.BridgingRequestStateDB
	logger hclog.Logger
}

var _ core.BridgingRequestStateManager = (*BridgingRequestStateManagerImpl)(nil)

func NewBridgingRequestStateManager(
	db core.BridgingRequestStateDB, logger hclog.Logger,
) (
	*BridgingRequestStateManagerImpl, error,
) {
	return &BridgingRequestStateManagerImpl{
		db:     db,
		logger: logger,
	}, nil
}

// New implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) New(sourceChainID string, model *common.NewBridgingRequestStateModel) error {
	state := core.NewBridgingRequestState(sourceChainID, model.SourceTxHash)

	err := m.db.AddBridgingRequestState(state)
	if err != nil {
		return fmt.Errorf("failed to add new BridgingRequestState. err: %w", err)
	}

	m.logger.Debug("New BridgingRequestState", "sourceChainID", state.SourceChainID,
		"sourceTxHash", state.SourceTxHash, "Status", state.Status)

	return nil
}

// NewMultiple implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) NewMultiple(
	sourceChainID string, models []*common.NewBridgingRequestStateModel,
) error {
	var errs []error

	for _, model := range models {
		err := m.New(sourceChainID, model)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to add some new BridgingRequestStates: %w", errors.Join(errs...))
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
func (m *BridgingRequestStateManagerImpl) SubmittedToBridge(
	key common.BridgingRequestStateKey, destinationChainID string,
) error {
	return m.updateStateByKey(key, func(state *core.BridgingRequestState) error {
		return state.ToSubmittedToBridge(destinationChainID)
	})
}

// IncludedInBatch implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) IncludedInBatch(
	destinationChainID string, batchID uint64, txs []common.BridgingRequestStateKey,
) error {
	var errs []error

	for _, key := range txs {
		state, err := m.db.GetBridgingRequestState(key.SourceChainID, key.SourceTxHash)
		if err != nil {
			errs = append(errs,
				fmt.Errorf("failed to get (%s, %s): %w", key.SourceChainID, key.SourceTxHash, err))

			continue
		}

		if state == nil {
			errs = append(errs,
				fmt.Errorf("state does not exist (%s, %s)", key.SourceChainID, key.SourceTxHash))

			continue
		}

		oldStatus := state.Status

		if state.DestinationChainID == "" {
			state.DestinationChainID = destinationChainID
		}

		if state.DestinationChainID != destinationChainID {
			errs = append(errs,
				fmt.Errorf(
					"destination chain not equal %s != %s for (%s, %s)",
					state.DestinationChainID, destinationChainID, key.SourceChainID, key.SourceTxHash))

			continue
		}

		err = state.ToIncludedInBatch(batchID)
		if err != nil {
			errs = append(errs,
				fmt.Errorf("failed to update (%s, %s): %w", key.SourceChainID, key.SourceTxHash, err))

			continue
		}

		err = m.db.UpdateBridgingRequestState(state)
		if err != nil {
			errs = append(errs,
				fmt.Errorf("failed to save (%s, %s): %w", key.SourceChainID, key.SourceTxHash, err))
		} else {
			m.logger.Debug("Updated BridgingRequestState", "sourceChainID", state.SourceChainID,
				"sourceTxHash", state.SourceTxHash, "Old Status", oldStatus, "New Status", state.Status)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to update some BridgingRequestStates for (%s, %d): %w",
			destinationChainID, batchID, errors.Join(errs...))
	}

	return nil
}

// SubmittedToDestination implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) SubmittedToDestination(destinationChainID string, batchID uint64) error {
	return m.updateStatesByBatchID(destinationChainID, batchID, func(state *core.BridgingRequestState) error {
		return state.ToSubmittedToDestination()
	})
}

// FailedToExecuteOnDestination implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) FailedToExecuteOnDestination(
	destinationChainID string, batchID uint64,
) error {
	return m.updateStatesByBatchID(destinationChainID, batchID, func(state *core.BridgingRequestState) error {
		return state.ToFailedToExecuteOnDestination()
	})
}

// ExecutedOnDestination implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) ExecutedOnDestination(
	destinationChainID string, batchID uint64, destinationTxHash common.Hash,
) error {
	return m.updateStatesByBatchID(destinationChainID, batchID, func(state *core.BridgingRequestState) error {
		return state.ToExecutedOnDestination(destinationTxHash)
	})
}

// Get implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) Get(
	sourceChainID string, sourceTxHash common.Hash,
) (
	*core.BridgingRequestState, error,
) {
	state, err := m.db.GetBridgingRequestState(sourceChainID, sourceTxHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get BridgingRequestState (%s, %s), err: %w",
			sourceChainID, sourceTxHash, err)
	}

	return state, nil
}

// GetAllForUser implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) GetMultiple(
	sourceChainID string, sourceTxHashes []common.Hash,
) ([]*core.BridgingRequestState, error) {
	var (
		result = make([]*core.BridgingRequestState, 0, len(sourceTxHashes))
		errs   []error
	)

	for _, sourceTxHash := range sourceTxHashes {
		state, err := m.db.GetBridgingRequestState(sourceChainID, sourceTxHash)
		if err != nil {
			errs = append(errs, err)
		} else if state != nil {
			result = append(result, state)
		}
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to get some BridgingRequestStates: %w", errors.Join(errs...))
	}

	return result, nil
}

func (m *BridgingRequestStateManagerImpl) updateStateByKey(
	key common.BridgingRequestStateKey, updateState func(state *core.BridgingRequestState) error,
) error {
	state, err := m.db.GetBridgingRequestState(key.SourceChainID, key.SourceTxHash)
	if err != nil {
		return fmt.Errorf("failed to get BridgingRequestState from db (%s, %s): %w",
			key.SourceChainID, key.SourceTxHash, err)
	}

	if state == nil {
		return fmt.Errorf("trying to get a non-existent BridgingRequestState (%s, %s)",
			key.SourceChainID, key.SourceTxHash)
	}

	oldStatus := state.Status

	err = updateState(state)
	if err != nil {
		return fmt.Errorf("failed to update a BridgingRequestState (%s, %s): %w",
			key.SourceChainID, key.SourceTxHash, err)
	}

	err = m.db.UpdateBridgingRequestState(state)
	if err != nil {
		return fmt.Errorf("failed to save updated BridgingRequestState (%s, %s): %w",
			key.SourceChainID, key.SourceTxHash, err)
	} else {
		m.logger.Debug("Updated BridgingRequestState", "sourceChainID", state.SourceChainID,
			"sourceTxHash", state.SourceTxHash, "Old Status", oldStatus, "New Status", state.Status)
	}

	return nil
}

func (m *BridgingRequestStateManagerImpl) updateStatesByBatchID(
	destinationChainID string, batchID uint64, updateState func(state *core.BridgingRequestState) error,
) error {
	states, err := m.db.GetBridgingRequestStatesByBatchID(destinationChainID, batchID)
	if err != nil {
		return fmt.Errorf("failed to get BridgingRequestStates. destinationChainId: %v, batchId: %v, err: %w",
			destinationChainID, batchID, err)
	}

	var errs []error

	for _, state := range states {
		oldStatus := state.Status

		err := updateState(state)
		if err != nil {
			errs = append(errs,
				fmt.Errorf("failed to update (%s, %s): %w",
					state.SourceChainID, state.SourceTxHash, err))

			continue
		}

		err = m.db.UpdateBridgingRequestState(state)
		if err != nil {
			errs = append(errs,
				fmt.Errorf("failed to save updated (%s, %s): %w",
					state.SourceChainID, state.SourceTxHash, err))
		} else {
			m.logger.Debug("Updated BridgingRequestState", "sourceChainID", state.SourceChainID,
				"sourceTxHash", state.SourceTxHash, "BatchID", state.BatchID,
				"Old Status", oldStatus, "New Status", state.Status)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to update some BridgingRequestStates: %w", errors.Join(errs...))
	}

	return nil
}
