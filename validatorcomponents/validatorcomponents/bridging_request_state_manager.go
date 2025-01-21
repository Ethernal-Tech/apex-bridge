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
		return state.ToInvalidRequest()
	})
}

// SubmittedToBridge implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) SubmittedToBridge(
	key common.BridgingRequestStateKey, dstChainID string,
) error {
	return m.updateStates([]common.BridgingRequestStateKey{key}, func(state *core.BridgingRequestState) error {
		return state.ToSubmittedToBridge(dstChainID)
	})
}

// IncludedInBatch implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) IncludedInBatch(
	txs []common.BridgingRequestStateKey, dstChainID string,
) error {
	return m.updateStates(txs, func(state *core.BridgingRequestState) error {
		if state.DestinationChainID == "" {
			state.DestinationChainID = dstChainID
		}

		if state.DestinationChainID != dstChainID {
			return fmt.Errorf("destination chain not equal %s != %s for (%s, %s)",
				state.DestinationChainID, dstChainID, state.SourceChainID, state.SourceTxHash)
		}

		return state.ToIncludedInBatch()
	})
}

// SubmittedToDestination implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) SubmittedToDestination(
	txs []common.BridgingRequestStateKey,
) error {
	return m.updateStates(txs, func(state *core.BridgingRequestState) error {
		return state.ToSubmittedToDestination()
	})
}

// FailedToExecuteOnDestination implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) FailedToExecuteOnDestination(
	txs []common.BridgingRequestStateKey,
) error {
	return m.updateStates(txs, func(state *core.BridgingRequestState) error {
		return state.ToFailedToExecuteOnDestination()
	})
}

// ExecutedOnDestination implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) ExecutedOnDestination(
	txs []common.BridgingRequestStateKey, dstTxHash common.Hash,
) error {
	return m.updateStates(txs, func(state *core.BridgingRequestState) error {
		return state.ToExecutedOnDestination(dstTxHash)
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
			errs = append(errs, fmt.Errorf("BridgingRequestState does not exist (%s, %s)",
				stateKey.SourceChainID, stateKey.SourceTxHash))

			continue
		}

		oldStatus := state.Status

		err = updateState(state)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to update BridgingRequestState (%s, %s) with status %s: %w",
				state.SourceChainID, state.SourceTxHash, oldStatus, err))

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
