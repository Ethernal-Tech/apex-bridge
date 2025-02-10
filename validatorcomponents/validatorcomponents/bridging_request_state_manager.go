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

var (
	errSkipTransition                                  = errors.New("skip transition")
	_                 core.BridgingRequestStateManager = (*BridgingRequestStateManagerImpl)(nil)
)

func NewBridgingRequestStateManager(
	db core.BridgingRequestStateDB, logger hclog.Logger,
) *BridgingRequestStateManagerImpl {
	return &BridgingRequestStateManagerImpl{
		db:     db,
		logger: logger,
	}
}

// New implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) New(sourceChainID string, model *common.NewBridgingRequestStateModel) error {
	state := core.NewBridgingRequestState(sourceChainID, model.SourceTxHash)

	err := m.db.AddBridgingRequestState(state)
	if err != nil {
		return fmt.Errorf("failed to add new BridgingRequestState. err: %w", err)
	}

	m.logger.Debug("New BridgingRequestState", "srcChainID", state.SourceChainID,
		"srcTxHash", state.SourceTxHash, "Status", state.Status)

	return nil
}

// NewMultiple implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) NewMultiple(
	srcChainID string, models []*common.NewBridgingRequestStateModel,
) error {
	var errs []error

	for _, model := range models {
		err := m.New(srcChainID, model)
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
	return m.updateStates([]common.BridgingRequestStateKey{key}, func(state *core.BridgingRequestState) error {
		if err := state.IsTransitionPossible(core.BridgingRequestStatusInvalidRequest); err != nil {
			return err
		}
		state.ToInvalidRequest()
		return nil
	})
}

// SubmittedToBridge implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) SubmittedToBridge(
	key common.BridgingRequestStateKey, dstChainID string,
) error {
	return m.updateStates([]common.BridgingRequestStateKey{key}, func(state *core.BridgingRequestState) error {
		if err := state.UpdateDestChainID(dstChainID); err != nil {
			return err
		}
		if err := state.IsTransitionPossible(core.BridgingRequestStatusSubmittedToBridge); err != nil {
			return err
		}
		state.ToSubmittedToBridge()
		return nil
	})
}

// IncludedInBatch implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) IncludedInBatch(
	txs []common.BridgingRequestStateKey, dstChainID string,
) error {
	return m.updateStates(txs, func(state *core.BridgingRequestState) error {
		if state.Status == core.BridgingRequestStatusSubmittedToDestination {
			return fmt.Errorf("%w: %s -> %s",
				errSkipTransition, state.Status, core.BridgingRequestStatusIncludedInBatch)
		}

		if err := state.UpdateDestChainID(dstChainID); err != nil {
			return err
		}

		if err := state.IsTransitionPossible(core.BridgingRequestStatusIncludedInBatch); err != nil {
			return err
		}

		state.ToIncludedInBatch()

		return nil
	})
}

// SubmittedToDestination implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) SubmittedToDestination(
	txs []common.BridgingRequestStateKey, dstChainID string,
) error {
	return m.updateStates(txs, func(state *core.BridgingRequestState) error {
		if err := state.UpdateDestChainID(dstChainID); err != nil {
			return err
		}

		if err := state.IsTransitionPossible(core.BridgingRequestStatusSubmittedToDestination); err != nil {
			return err
		}

		state.ToSubmittedToDestination()

		return nil
	})
}

// FailedToExecuteOnDestination implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) FailedToExecuteOnDestination(
	txs []common.BridgingRequestStateKey, dstChainID string,
) error {
	return m.updateStates(txs, func(state *core.BridgingRequestState) error {
		if err := state.UpdateDestChainID(dstChainID); err != nil {
			return err
		}

		if err := state.IsTransitionPossible(core.BridgingRequestStatusFailedToExecuteOnDestination); err != nil {
			return err
		}

		state.ToFailedToExecuteOnDestination()

		return nil
	})
}

// ExecutedOnDestination implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) ExecutedOnDestination(
	txs []common.BridgingRequestStateKey, dstTxHash common.Hash, dstChainID string,
) error {
	return m.updateStates(txs, func(state *core.BridgingRequestState) error {
		if err := state.UpdateDestChainID(dstChainID); err != nil {
			return err
		}

		if err := state.IsTransitionPossible(core.BridgingRequestStatusExecutedOnDestination); err != nil {
			return err
		}

		state.ToExecutedOnDestination(dstTxHash)

		return nil
	})
}

// Get implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) Get(
	srcChainID string, srcTxHash common.Hash,
) (*core.BridgingRequestState, error) {
	state, err := m.db.GetBridgingRequestState(srcChainID, srcTxHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get BridgingRequestState (%s, %s), err: %w",
			srcChainID, srcTxHash, err)
	}

	return state, nil
}

// GetAllForUser implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) GetMultiple(
	srcChainID string, srcTxHashes []common.Hash,
) ([]*core.BridgingRequestState, error) {
	var (
		result = make([]*core.BridgingRequestState, 0, len(srcTxHashes))
		errs   []error
	)

	for _, sourceTxHash := range srcTxHashes {
		state, err := m.db.GetBridgingRequestState(srcChainID, sourceTxHash)
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

func (m *BridgingRequestStateManagerImpl) updateStates(
	stateKeys []common.BridgingRequestStateKey, updateState func(state *core.BridgingRequestState) error,
) error {
	var errs []error

	for _, stateKey := range stateKeys {
		state, err := m.db.GetBridgingRequestState(stateKey.SourceChainID, stateKey.SourceTxHash)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get BridgingRequestState from db (%s, %s): %w",
				stateKey.SourceChainID, stateKey.SourceTxHash, err))

			continue
		}

		if state == nil {
			// insert bridging request state if not exists in db
			state = core.NewBridgingRequestState(stateKey.SourceChainID, stateKey.SourceTxHash)

			err := m.db.AddBridgingRequestState(state)
			if err != nil {
				errs = append(errs, fmt.Errorf("BridgingRequestState does not exist (%s, %s) but failed to add: %w",
					stateKey.SourceChainID, stateKey.SourceTxHash, err))

				continue
			}
		}

		oldStatus := state.Status

		err = updateState(state)
		if err != nil {
			// Some transitions should not be considered errors:
			// For example, a relayer imitator might be faster than the batch submitter
			// and may set the status to SubmittedToBridge before the batch submitter updates it to IncludedInBatch.
			// In that case, we simply need to skip the transition.
			if !errors.Is(err, errSkipTransition) {
				errs = append(errs, fmt.Errorf("failed to update BridgingRequestState (%s, %s) with status %s: %w",
					state.SourceChainID, state.SourceTxHash, oldStatus, err))
			}

			continue
		}

		err = m.db.UpdateBridgingRequestState(state)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to save updated BridgingRequestState (%s, %s) with status %s: %w",
				state.SourceChainID, state.SourceTxHash, oldStatus, err))
		} else {
			m.logger.Debug("Updated BridgingRequestState",
				"srcChainID", state.SourceChainID, "srcTxHash", state.SourceTxHash,
				"Old Status", oldStatus, "New Status", state.Status)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to update some BridgingRequestStates: %w", errors.Join(errs...))
	}

	return nil
}
