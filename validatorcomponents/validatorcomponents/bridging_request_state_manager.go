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
	state := common.NewBridgingRequestState(sourceChainID, model.SourceTxHash, model.IsRefund)

	err := m.db.AddBridgingRequestState(state)
	if err != nil {
		return fmt.Errorf("failed to add new BridgingRequestState. err: %w", err)
	}

	m.logger.Debug("New BridgingRequestState", "srcChainID", state.SourceChainID,
		"srcTxHash", state.SourceTxHash, "Status", state.StatusStr())

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
	return m.updateStates([]common.BridgingRequestStateKey{key},
		func(stateKey common.BridgingRequestStateKey, state *common.BridgingRequestState) error {
			if err := state.IsTransitionPossible(common.BridgingRequestStatusInvalidRequest); err != nil {
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
	return m.updateStates([]common.BridgingRequestStateKey{key},
		func(stateKey common.BridgingRequestStateKey, state *common.BridgingRequestState) error {
			state.DestinationChainID = dstChainID
			state.IsRefund = stateKey.IsRefund

			if err := state.IsTransitionPossible(common.BridgingRequestStatusSubmittedToBridge); err != nil {
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
	return m.updateStates(txs,
		func(stateKey common.BridgingRequestStateKey, state *common.BridgingRequestState) error {
			if state.Status == common.BridgingRequestStatusSubmittedToDestination {
				return fmt.Errorf("%w: %s -> %s",
					errSkipTransition, state.StatusStr(), common.BridgingRequestStatusIncludedInBatch)
			}

			state.DestinationChainID = dstChainID
			state.IsRefund = stateKey.IsRefund

			if err := state.IsTransitionPossible(common.BridgingRequestStatusIncludedInBatch); err != nil {
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
	return m.updateStates(txs,
		func(stateKey common.BridgingRequestStateKey, state *common.BridgingRequestState) error {
			state.DestinationChainID = dstChainID
			state.IsRefund = stateKey.IsRefund

			if err := state.IsTransitionPossible(common.BridgingRequestStatusSubmittedToDestination); err != nil {
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
	return m.updateStates(txs,
		func(stateKey common.BridgingRequestStateKey, state *common.BridgingRequestState) error {
			state.DestinationChainID = dstChainID
			state.IsRefund = stateKey.IsRefund

			if err := state.IsTransitionPossible(common.BridgingRequestStatusFailedToExecuteOnDestination); err != nil {
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
	return m.updateStates(txs,
		func(stateKey common.BridgingRequestStateKey, state *common.BridgingRequestState) error {
			state.DestinationChainID = dstChainID
			state.IsRefund = stateKey.IsRefund

			if err := state.IsTransitionPossible(common.BridgingRequestStatusExecutedOnDestination); err != nil {
				return err
			}

			state.ToExecutedOnDestination(dstTxHash)

			return nil
		})
}

// Get implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) Get(
	srcChainID string, srcTxHash common.Hash,
) (*common.BridgingRequestState, error) {
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
) ([]*common.BridgingRequestState, error) {
	var (
		result = make([]*common.BridgingRequestState, 0, len(srcTxHashes))
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
	stateKeys []common.BridgingRequestStateKey,
	updateState func(stateKey common.BridgingRequestStateKey, state *common.BridgingRequestState) error,
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
			state = common.NewBridgingRequestState(stateKey.SourceChainID, stateKey.SourceTxHash, stateKey.IsRefund)

			err := m.db.AddBridgingRequestState(state)
			if err != nil {
				errs = append(errs, fmt.Errorf("BridgingRequestState does not exist (%s, %s) but failed to add: %w",
					stateKey.SourceChainID, stateKey.SourceTxHash, err))

				continue
			}
		}

		oldStatus := state.Status

		err = updateState(stateKey, state)
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
				"Old Status", oldStatus, "New Status", state.StatusStr())
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to update some BridgingRequestStates: %w", errors.Join(errs...))
	}

	return nil
}
