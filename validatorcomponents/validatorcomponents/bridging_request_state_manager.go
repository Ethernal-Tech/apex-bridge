package validatorcomponents

import (
	"errors"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
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
func (m *BridgingRequestStateManagerImpl) New(sourceChainID string, tx *indexer.Tx) error {
	inputAddrs := make([]string, 0, len(tx.Inputs))
	for _, input := range tx.Inputs {
		inputAddrs = append(inputAddrs, input.Output.Address)
	}

	state := core.NewBridgingRequestState(sourceChainID, tx.Hash, inputAddrs)

	err := m.db.AddBridgingRequestState(state)
	if err != nil {
		m.logger.Error("failed to add new BridgingRequestState", "err", err)

		return fmt.Errorf("failed to add new BridgingRequestState. err: %w", err)
	}

	m.logger.Debug("New BridgingRequestState", "sourceChainID", state.SourceChainID,
		"sourceTxHash", state.SourceTxHash, "Status", state.Status)

	return nil
}

// NewMultiple implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) NewMultiple(sourceChainID string, txs []*indexer.Tx) error {
	errs := make([]error, 0)

	for _, tx := range txs {
		err := m.New(sourceChainID, tx)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		m.logger.Error("failed to add some new BridgingRequestStates", "errors", errors.Join(errs...))

		return fmt.Errorf("failed to add some new BridgingRequestStates. errors: %w", errors.Join(errs...))
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
	errs := make([]error, 0)

	for _, key := range txs {
		state, err := m.db.GetBridgingRequestState(key.SourceChainID, key.SourceTxHash)
		if state == nil || err != nil {
			continue
		}

		if state.DestinationChainID != destinationChainID {
			m.logger.Error(
				"batch destinationChainId not equal to BridgingRequestState.DestinationChainId",
				"sourceChainId", key.SourceChainID, "sourceTxHash", key.SourceTxHash)

			errs = append(errs,
				fmt.Errorf(
					"batch destinationChainId not equal to BridgingRequestState.DestinationChainId. sourceChainId: %v, sourceTxHash: %v", //nolint:lll
					key.SourceChainID, key.SourceTxHash))

			continue
		}

		oldStatus := state.Status

		err = state.ToIncludedInBatch(batchID)
		if err != nil {
			m.logger.Error(
				"failed to update a BridgingRequestState",
				"sourceChainId", key.SourceChainID, "sourceTxHash", key.SourceTxHash, "err", err)

			errs = append(errs,
				fmt.Errorf("failed to update a BridgingRequestState. sourceChainId: %v, sourceTxHash: %v, err: %w",
					key.SourceChainID, key.SourceTxHash, err))

			continue
		}

		err = m.db.UpdateBridgingRequestState(state)
		if err != nil {
			m.logger.Error(
				"failed to save updated BridgingRequestState",
				"sourceChainId", key.SourceChainID, "sourceTxHash", key.SourceTxHash, "err", err)

			errs = append(errs,
				fmt.Errorf("failed to save updated BridgingRequestState. sourceChainId: %v, sourceTxHash: %v, err: %w",
					key.SourceChainID, key.SourceTxHash, err))
		}

		m.logger.Debug("Updated BridgingRequestState", "sourceChainID", state.SourceChainID,
			"sourceTxHash", state.SourceTxHash, "Old Status", oldStatus, "New Status", state.Status)
	}

	if len(errs) > 0 {
		m.logger.Error("failed to update some BridgingRequestStates", "errors", errors.Join(errs...))

		return fmt.Errorf("failed to update some BridgingRequestStates. errors: %w", errors.Join(errs...))
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
	destinationChainID string, batchID uint64, destinationTxHash string,
) error {
	return m.updateStatesByBatchID(destinationChainID, batchID, func(state *core.BridgingRequestState) error {
		return state.ToExecutedOnDestination(destinationTxHash)
	})
}

// Get implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) Get(
	sourceChainID string, sourceTxHash string,
) (
	*core.BridgingRequestState, error,
) {
	state, err := m.db.GetBridgingRequestState(sourceChainID, sourceTxHash)
	if err != nil {
		m.logger.Error(
			"failed to get BridgingRequestState",
			"sourceChainId", sourceChainID, "sourceTxHash", sourceTxHash, "err", err)

		return nil, fmt.Errorf("failed to get BridgingRequestState. sourceChainId: %v, sourceTxHash: %v, err: %w",
			sourceChainID, sourceTxHash, err)
	}

	return state, nil
}

// GetAllForUser implements core.BridgingRequestStateManager.
func (m *BridgingRequestStateManagerImpl) GetAllForUser(
	sourceChainID string, userAddr string,
) (
	[]*core.BridgingRequestState, error,
) {
	states, err := m.db.GetUserBridgingRequestStates(sourceChainID, userAddr)
	if err != nil {
		m.logger.Error(
			"failed to get all BridgingRequestStates for user",
			"sourceChainId", sourceChainID, "userAddr", userAddr, "err", err)

		return nil, fmt.Errorf("failed to get all BridgingRequestStates for user. sourceChainId: %v, userAddr: %v, err: %w",
			sourceChainID, userAddr, err)
	}

	return states, nil
}

func (m *BridgingRequestStateManagerImpl) updateStateByKey(
	key common.BridgingRequestStateKey, updateState func(state *core.BridgingRequestState) error,
) error {
	state, err := m.db.GetBridgingRequestState(key.SourceChainID, key.SourceTxHash)
	if err != nil {
		m.logger.Error(
			"failed to get BridgingRequestState",
			"sourceChainId", key.SourceChainID, "sourceTxHash", key.SourceTxHash, "err", err)

		return fmt.Errorf("failed to get BridgingRequestState. sourceChainId: %v, sourceTxHash: %v, err: %w",
			key.SourceChainID, key.SourceTxHash, err)
	}

	if state == nil {
		m.logger.Error(
			"trying to get a non-existent BridgingRequestState", "sourceChainId",
			key.SourceChainID, "sourceTxHash", key.SourceTxHash)

		return fmt.Errorf("trying to get a non-existent BridgingRequestState. sourceChainId: %v, sourceTxHash: %v",
			key.SourceChainID, key.SourceTxHash)
	}

	oldStatus := state.Status

	err = updateState(state)
	if err != nil {
		m.logger.Error(
			"failed to update a BridgingRequestState",
			"sourceChainId", key.SourceChainID, "sourceTxHash", key.SourceTxHash, "err", err)

		return fmt.Errorf("failed to update a BridgingRequestState. sourceChainId: %v, sourceTxHash: %v, err: %w",
			key.SourceChainID, key.SourceTxHash, err)
	}

	err = m.db.UpdateBridgingRequestState(state)
	if err != nil {
		m.logger.Error(
			"failed to save updated BridgingRequestState",
			"sourceChainId", key.SourceChainID, "sourceTxHash", key.SourceTxHash, "err", err)

		return fmt.Errorf("failed to save updated BridgingRequestState. sourceChainId: %v, sourceTxHash: %v, err: %w",
			key.SourceChainID, key.SourceTxHash, err)
	}

	m.logger.Debug("Updated BridgingRequestState", "sourceChainID", state.SourceChainID,
		"sourceTxHash", state.SourceTxHash, "Old Status", oldStatus, "New Status", state.Status)

	return nil
}

func (m *BridgingRequestStateManagerImpl) updateStatesByBatchID(
	destinationChainID string, batchID uint64, updateState func(state *core.BridgingRequestState) error,
) error {
	states, err := m.db.GetBridgingRequestStatesByBatchID(destinationChainID, batchID)
	if err != nil {
		m.logger.Error(
			"failed to get BridgingRequestStates",
			"destinationChainId", destinationChainID, "batchId", batchID, "err", err)

		return fmt.Errorf("failed to get BridgingRequestStates. destinationChainId: %v, batchId: %v, err: %w",
			destinationChainID, batchID, err)
	}

	errs := make([]error, 0)

	for _, state := range states {
		oldStatus := state.Status

		err := updateState(state)
		if err != nil {
			m.logger.Error(
				"failed to update a BridgingRequestState",
				"sourceChainId", state.SourceChainID, "sourceTxHash", state.SourceTxHash, "err", err)

			errs = append(errs,
				fmt.Errorf("failed to update a BridgingRequestState. sourceChainId: %v, sourceTxHash: %v, err: %w",
					state.SourceChainID, state.SourceTxHash, err))

			continue
		}

		err = m.db.UpdateBridgingRequestState(state)
		if err != nil {
			m.logger.Error(
				"failed to save updated BridgingRequestState",
				"sourceChainId", state.SourceChainID, "sourceTxHash", state.SourceTxHash, "err", err)

			errs = append(errs,
				fmt.Errorf("failed to save updated BridgingRequestState. sourceChainId: %v, sourceTxHash: %v, err: %w",
					state.SourceChainID, state.SourceTxHash, err))
		}

		m.logger.Debug("Updated BridgingRequestState", "sourceChainID", state.SourceChainID,
			"sourceTxHash", state.SourceTxHash, "BatchID", state.BatchID,
			"Old Status", oldStatus, "New Status", state.Status)
	}

	if len(errs) > 0 {
		m.logger.Error("failed to update some BridgingRequestStates", "errors", errors.Join(errs...))

		return fmt.Errorf("failed to update some BridgingRequestStates. errors: %w", errors.Join(errs...))
	}

	return nil
}
