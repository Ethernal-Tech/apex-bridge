package validatorcomponents

import (
	"fmt"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	databaseaccess "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/database_access"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBridgingRequestStateManager(t *testing.T) {
	txHash := common.NewHashFromHexString("0xFF")
	srcChainID := common.ChainIDStrPrime
	dstChainID := common.ChainIDStrVector
	srcTxHash := common.NewHashFromHexString("0xff")

	t.Run("New 1", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("AddBridgingRequestState", mock.Anything).Return(fmt.Errorf("test err"))

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())
		err := sm.New(common.ChainIDStrPrime, &common.NewBridgingRequestStateModel{})
		require.ErrorContains(t, err, "failed to add new BridgingRequestState")

		db.AssertExpectations(t)
	})

	t.Run("New 2", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("AddBridgingRequestState", mock.Anything).Return(nil)

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.New(common.ChainIDStrPrime, &common.NewBridgingRequestStateModel{})
		require.NoError(t, err)

		db.AssertExpectations(t)
	})

	t.Run("NewMultiple 1", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.NewMultiple(common.ChainIDStrPrime, []*common.NewBridgingRequestStateModel{})
		require.NoError(t, err)

		db.AssertExpectations(t)
	})

	t.Run("NewMultiple 2", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("AddBridgingRequestState", mock.Anything).Return(fmt.Errorf("test err"))

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.NewMultiple(common.ChainIDStrPrime, []*common.NewBridgingRequestStateModel{{}})
		require.ErrorContains(t, err, "failed to add some new BridgingRequestStates")

		db.AssertExpectations(t)
	})

	t.Run("NewMultiple 3", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("AddBridgingRequestState", mock.Anything).Return(nil)

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.NewMultiple(common.ChainIDStrPrime, []*common.NewBridgingRequestStateModel{{}})
		require.NoError(t, err)

		db.AssertExpectations(t)
	})

	t.Run("Invalid 1", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(nil, fmt.Errorf("test err"))

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.Invalid(common.BridgingRequestStateKey{
			SourceChainID: srcChainID, SourceTxHash: srcTxHash,
		})
		require.ErrorContains(t, err, "failed to get BridgingRequestState")

		db.AssertExpectations(t)
	})

	t.Run("Invalid 2", func(t *testing.T) {
		state := core.NewBridgingRequestState(srcChainID, srcTxHash)

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(nil, nil)
		db.On("AddBridgingRequestState", state).Return(fmt.Errorf("test err"))

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.Invalid(common.BridgingRequestStateKey{
			SourceChainID: srcChainID, SourceTxHash: srcTxHash,
		})
		require.ErrorContains(t, err, "BridgingRequestState does not exist")

		db.AssertExpectations(t)
	})

	t.Run("Invalid 3", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(&core.BridgingRequestState{
			Status: core.BridgingRequestStatusExecutedOnDestination,
		}, nil)

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.Invalid(common.BridgingRequestStateKey{
			SourceChainID: srcChainID, SourceTxHash: srcTxHash,
		})
		require.ErrorContains(t, err, "invalid transition")

		db.AssertExpectations(t)
	})

	t.Run("Invalid 4", func(t *testing.T) {
		state := core.NewBridgingRequestState(srcChainID, srcTxHash)

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(state, nil)
		db.On("UpdateBridgingRequestState", mock.Anything).Return(fmt.Errorf("test err"))

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.Invalid(common.BridgingRequestStateKey{
			SourceChainID: srcChainID, SourceTxHash: srcTxHash,
		})
		require.ErrorContains(t, err, "failed to update some BridgingRequestStates")

		db.AssertExpectations(t)
	})

	t.Run("Invalid 5", func(t *testing.T) {
		state := core.NewBridgingRequestState(srcChainID, srcTxHash)

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(state, nil)
		db.On("UpdateBridgingRequestState", mock.Anything).Return(nil)

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.Invalid(common.BridgingRequestStateKey{
			SourceChainID: srcChainID, SourceTxHash: srcTxHash,
		})
		require.NoError(t, err)

		db.AssertExpectations(t)
	})

	t.Run("SubmittedToBridge 1", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(nil, fmt.Errorf("test err"))

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToBridge(common.BridgingRequestStateKey{
			SourceChainID: srcChainID, SourceTxHash: srcTxHash,
		}, common.ChainIDStrVector)
		require.ErrorContains(t, err, "failed to get BridgingRequestState")

		db.AssertExpectations(t)
	})

	t.Run("SubmittedToBridge 2", func(t *testing.T) {
		state := core.NewBridgingRequestState(srcChainID, srcTxHash)

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(nil, nil)
		db.On("AddBridgingRequestState", state).Return(fmt.Errorf("test err"))

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToBridge(common.BridgingRequestStateKey{
			SourceChainID: srcChainID, SourceTxHash: srcTxHash,
		}, common.ChainIDStrVector)
		require.ErrorContains(t, err, "BridgingRequestState does not exist")

		db.AssertExpectations(t)
	})

	t.Run("SubmittedToBridge 3", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(&core.BridgingRequestState{
			Status: core.BridgingRequestStatusExecutedOnDestination,
		}, nil)

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToBridge(common.BridgingRequestStateKey{
			SourceChainID: srcChainID, SourceTxHash: srcTxHash,
		}, common.ChainIDStrVector)
		require.ErrorContains(t, err, "invalid transition")

		db.AssertExpectations(t)
	})

	t.Run("SubmittedToBridge 4", func(t *testing.T) {
		state := core.NewBridgingRequestState(srcChainID, srcTxHash)

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(state, nil)
		db.On("UpdateBridgingRequestState", mock.Anything).Return(fmt.Errorf("test err"))

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToBridge(common.BridgingRequestStateKey{
			SourceChainID: srcChainID, SourceTxHash: srcTxHash,
		}, common.ChainIDStrVector)
		require.ErrorContains(t, err, "failed to save updated")

		db.AssertExpectations(t)
	})

	t.Run("SubmittedToBridge 5", func(t *testing.T) {
		state := core.NewBridgingRequestState(srcChainID, srcTxHash)

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(state, nil)
		db.On("UpdateBridgingRequestState", mock.Anything).Return(nil)

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToBridge(common.BridgingRequestStateKey{
			SourceChainID: srcChainID, SourceTxHash: srcTxHash,
		}, common.ChainIDStrVector)
		require.NoError(t, err)

		db.AssertExpectations(t)
	})

	t.Run("IncludedInBatch 1", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(nil, fmt.Errorf("test err"))

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.IncludedInBatch([]common.BridgingRequestStateKey{
			common.NewBridgingRequestStateKey(srcChainID, srcTxHash),
		}, common.ChainIDStrVector)

		require.ErrorContains(t, err, "failed to get")

		db.AssertExpectations(t)
	})

	t.Run("IncludedInBatch 2", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.IncludedInBatch([]common.BridgingRequestStateKey{}, common.ChainIDStrVector)
		require.NoError(t, err)

		db.AssertExpectations(t)
	})

	t.Run("IncludedInBatch 3", func(t *testing.T) {
		state := core.NewBridgingRequestState(srcChainID, srcTxHash)
		state.DestinationChainID = "nonsense"

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(state, nil)

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.IncludedInBatch([]common.BridgingRequestStateKey{
			common.NewBridgingRequestStateKey(srcChainID, srcTxHash),
		}, common.ChainIDStrVector)
		require.ErrorContains(t, err, "failed to update BridgingRequestState")

		db.AssertExpectations(t)
	})

	t.Run("IncludedInBatch 4", func(t *testing.T) {
		state := core.NewBridgingRequestState(srcChainID, srcTxHash)

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(state, nil)
		db.On("UpdateBridgingRequestState", mock.Anything).Return(fmt.Errorf("test err"))

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.IncludedInBatch([]common.BridgingRequestStateKey{
			common.NewBridgingRequestStateKey(srcChainID, srcTxHash),
		}, common.ChainIDStrVector)

		require.ErrorContains(t, err, "failed to save updated")

		db.AssertExpectations(t)
	})

	t.Run("IncludedInBatch 5", func(t *testing.T) {
		state := core.NewBridgingRequestState(srcChainID, srcTxHash)

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(state, nil)
		db.On("UpdateBridgingRequestState", mock.Anything).Return(nil)

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.IncludedInBatch([]common.BridgingRequestStateKey{
			common.NewBridgingRequestStateKey(srcChainID, srcTxHash),
		}, common.ChainIDStrVector)
		require.NoError(t, err)

		db.AssertExpectations(t)
	})

	t.Run("IncludedInBatch 6", func(t *testing.T) {
		state := core.NewBridgingRequestState(srcChainID, srcTxHash)
		state.ToSubmittedToDestination()

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(state, nil)

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.IncludedInBatch([]common.BridgingRequestStateKey{
			common.NewBridgingRequestStateKey(srcChainID, srcTxHash),
		}, common.ChainIDStrVector)
		require.NoError(t, err)

		db.AssertExpectations(t)
	})

	t.Run("SubmittedToDestination 1", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(nil, fmt.Errorf("test err"))

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToDestination([]common.BridgingRequestStateKey{
			common.NewBridgingRequestStateKey(srcChainID, srcTxHash),
		}, dstChainID)
		require.ErrorContains(t, err, "failed to get BridgingRequestState")

		db.AssertExpectations(t)
	})

	t.Run("SubmittedToDestination 2", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToDestination(nil, dstChainID)
		require.NoError(t, err)

		db.AssertExpectations(t)
	})

	t.Run("SubmittedToDestination 3", func(t *testing.T) {
		state := core.NewBridgingRequestState(srcChainID, srcTxHash)
		state.ToExecutedOnDestination(txHash)

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(state, nil)

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToDestination([]common.BridgingRequestStateKey{
			common.NewBridgingRequestStateKey(srcChainID, srcTxHash),
		}, dstChainID)
		require.ErrorContains(t, err, "failed to update")

		db.AssertExpectations(t)
	})

	t.Run("SubmittedToDestination 4", func(t *testing.T) {
		state := core.NewBridgingRequestState(srcChainID, srcTxHash)

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(state, nil)
		db.On("UpdateBridgingRequestState", mock.Anything).Return(fmt.Errorf("test err"))

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToDestination([]common.BridgingRequestStateKey{
			common.NewBridgingRequestStateKey(srcChainID, srcTxHash),
		}, dstChainID)
		require.ErrorContains(t, err, "failed to save updated")

		db.AssertExpectations(t)
	})

	t.Run("SubmittedToDestination 5", func(t *testing.T) {
		state := core.NewBridgingRequestState(srcChainID, srcTxHash)

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(state, nil)
		db.On("UpdateBridgingRequestState", mock.Anything).Return(nil)

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToDestination([]common.BridgingRequestStateKey{
			common.NewBridgingRequestStateKey(srcChainID, srcTxHash),
		}, dstChainID)
		require.NoError(t, err)

		db.AssertExpectations(t)
	})

	t.Run("SubmittedToDestination 6", func(t *testing.T) {
		state := core.NewBridgingRequestState(srcChainID, srcTxHash)

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("AddBridgingRequestState", state).Return(nil)
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(nil, nil)
		db.On("UpdateBridgingRequestState", mock.Anything).Return(nil)

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToDestination([]common.BridgingRequestStateKey{
			common.NewBridgingRequestStateKey(srcChainID, srcTxHash),
		}, dstChainID)
		require.NoError(t, err)

		db.AssertExpectations(t)
	})

	t.Run("FailedToExecuteOnDestination 1", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(nil, fmt.Errorf("test err"))

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.FailedToExecuteOnDestination([]common.BridgingRequestStateKey{
			common.NewBridgingRequestStateKey(srcChainID, srcTxHash),
		}, dstChainID)
		require.ErrorContains(t, err, "failed to get BridgingRequestState from db")

		db.AssertExpectations(t)
	})

	t.Run("FailedToExecuteOnDestination 2", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.FailedToExecuteOnDestination(nil, dstChainID)
		require.NoError(t, err)

		db.AssertExpectations(t)
	})

	t.Run("FailedToExecuteOnDestination 3", func(t *testing.T) {
		state := core.NewBridgingRequestState(srcChainID, srcTxHash)

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(state, nil)
		db.On("UpdateBridgingRequestState", mock.Anything).Return(fmt.Errorf("test err"))

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.FailedToExecuteOnDestination([]common.BridgingRequestStateKey{
			common.NewBridgingRequestStateKey(srcChainID, srcTxHash),
		}, dstChainID)
		require.ErrorContains(t, err, "failed to update")

		db.AssertExpectations(t)
	})

	t.Run("FailedToExecuteOnDestination 4", func(t *testing.T) {
		state := core.NewBridgingRequestState(srcChainID, srcTxHash)

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(state, nil)
		db.On("UpdateBridgingRequestState", mock.Anything).Return(fmt.Errorf("test err"))

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.FailedToExecuteOnDestination([]common.BridgingRequestStateKey{
			common.NewBridgingRequestStateKey(srcChainID, srcTxHash),
		}, dstChainID)
		require.ErrorContains(t, err, "failed to save updated")

		db.AssertExpectations(t)
	})

	t.Run("FailedToExecuteOnDestination 5", func(t *testing.T) {
		state := core.NewBridgingRequestState(srcChainID, srcTxHash)

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(state, nil)
		db.On("UpdateBridgingRequestState", mock.Anything).Return(nil)

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.FailedToExecuteOnDestination([]common.BridgingRequestStateKey{
			common.NewBridgingRequestStateKey(srcChainID, srcTxHash),
		}, dstChainID)
		require.NoError(t, err)

		db.AssertExpectations(t)
	})

	t.Run("FailedToExecuteOnDestination 6", func(t *testing.T) {
		state := core.NewBridgingRequestState(srcChainID, srcTxHash)

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("AddBridgingRequestState", state).Return(nil)
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(nil, nil)
		db.On("UpdateBridgingRequestState", mock.Anything).Return(nil)

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.FailedToExecuteOnDestination([]common.BridgingRequestStateKey{
			common.NewBridgingRequestStateKey(srcChainID, srcTxHash),
		}, dstChainID)
		require.NoError(t, err)

		db.AssertExpectations(t)
	})

	t.Run("ExecutedOnDestination 1", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(nil, fmt.Errorf("test err"))

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.ExecutedOnDestination([]common.BridgingRequestStateKey{
			common.NewBridgingRequestStateKey(srcChainID, srcTxHash),
		}, txHash, dstChainID)
		require.ErrorContains(t, err, "failed to get BridgingRequestState from db")

		db.AssertExpectations(t)
	})

	t.Run("ExecutedOnDestination 2", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.ExecutedOnDestination([]common.BridgingRequestStateKey{}, common.Hash{}, dstChainID)
		require.NoError(t, err)

		db.AssertExpectations(t)
	})

	t.Run("ExecutedOnDestination 3", func(t *testing.T) {
		state := core.NewBridgingRequestState(srcChainID, srcTxHash)
		state.ToInvalidRequest()

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(state, nil)

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.ExecutedOnDestination([]common.BridgingRequestStateKey{
			common.NewBridgingRequestStateKey(srcChainID, srcTxHash),
		}, txHash, dstChainID)
		require.ErrorContains(t, err, "failed to update some BridgingRequestStates: failed to update")

		db.AssertExpectations(t)
	})

	t.Run("ExecutedOnDestination 4", func(t *testing.T) {
		state := core.NewBridgingRequestState(srcChainID, srcTxHash)

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(state, nil)
		db.On("UpdateBridgingRequestState", mock.Anything).Return(fmt.Errorf("test err"))

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.ExecutedOnDestination([]common.BridgingRequestStateKey{
			common.NewBridgingRequestStateKey(srcChainID, srcTxHash),
		}, txHash, dstChainID)
		require.ErrorContains(t, err, "failed to update some BridgingRequestStates: failed to save updated")

		db.AssertExpectations(t)
	})

	t.Run("ExecutedOnDestination 5", func(t *testing.T) {
		state := core.NewBridgingRequestState(srcChainID, srcTxHash)

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(state, nil)
		db.On("UpdateBridgingRequestState", mock.Anything).Return(nil)

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.ExecutedOnDestination([]common.BridgingRequestStateKey{
			common.NewBridgingRequestStateKey(srcChainID, srcTxHash),
		}, txHash, dstChainID)
		require.NoError(t, err)

		db.AssertExpectations(t)
	})

	t.Run("ExecutedOnDestination 6", func(t *testing.T) {
		state := core.NewBridgingRequestState(srcChainID, srcTxHash)

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("AddBridgingRequestState", state).Return(nil)
		db.On("GetBridgingRequestState", srcChainID, srcTxHash).Return(nil, nil)
		db.On("UpdateBridgingRequestState", mock.Anything).Return(nil)

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.ExecutedOnDestination([]common.BridgingRequestStateKey{
			common.NewBridgingRequestStateKey(srcChainID, srcTxHash),
		}, txHash, dstChainID)
		require.NoError(t, err)

		db.AssertExpectations(t)
	})

	t.Run("Get 1", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", common.ChainIDStrPrime, txHash).Return(nil, fmt.Errorf("test err"))

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		state, err := sm.Get(common.ChainIDStrPrime, txHash)
		require.ErrorContains(t, err, "failed to get BridgingRequestState")
		require.Nil(t, state)

		db.AssertExpectations(t)
	})

	t.Run("Get 2", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", common.ChainIDStrPrime, txHash).Return(nil, nil)

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		state, err := sm.Get(common.ChainIDStrPrime, txHash)
		require.NoError(t, err)
		require.Nil(t, state)

		db.AssertExpectations(t)
	})

	t.Run("Get 3", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", common.ChainIDStrPrime, txHash).Return(&core.BridgingRequestState{}, nil)

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		state, err := sm.Get(common.ChainIDStrPrime, txHash)
		require.NoError(t, err)
		require.NotNil(t, state)

		db.AssertExpectations(t)
	})

	t.Run("GetMultiple 1", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", common.ChainIDStrPrime, txHash).Return(nil, fmt.Errorf("test err"))

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		states, err := sm.GetMultiple(common.ChainIDStrPrime, []common.Hash{
			txHash,
		})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get some BridgingRequestStates")
		require.Nil(t, states)

		db.AssertExpectations(t)
	})

	t.Run("GetMultiple 2", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", common.ChainIDStrPrime, txHash).Return(nil, nil)

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		states, err := sm.GetMultiple(common.ChainIDStrPrime, []common.Hash{
			txHash,
		})
		require.NoError(t, err)
		require.Len(t, states, 0)

		db.AssertExpectations(t)
	})

	t.Run("GetMultiple 3", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState", common.ChainIDStrPrime, txHash).Return(&core.BridgingRequestState{}, nil)

		sm := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		states, err := sm.GetMultiple(common.ChainIDStrPrime, []common.Hash{
			txHash,
		})
		require.NoError(t, err)
		require.NotNil(t, states)
		require.Len(t, states, 1)

		db.AssertExpectations(t)
	})
}
