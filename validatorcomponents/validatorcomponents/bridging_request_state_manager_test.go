package validatorcomponents

import (
	"fmt"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	databaseaccess "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/database_access"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestBridgingRequestStateManager(t *testing.T) {
	t.Run("NewBridgingRequestStateManager", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}

		sm, err := NewBridgingRequestStateManager(db, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, sm)
	})

	t.Run("New 1", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("AddBridgingRequestState").Return(fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())
		err := sm.New(common.ChainIDStrPrime, &indexer.Tx{})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to add new BridgingRequestState")
	})

	t.Run("New 2", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("AddBridgingRequestState").Return(nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())
		err := sm.New(common.ChainIDStrPrime, &indexer.Tx{})
		require.NoError(t, err)
	})

	t.Run("NewMultiple 1", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())
		err := sm.NewMultiple(common.ChainIDStrPrime, []*indexer.Tx{})
		require.NoError(t, err)
	})

	t.Run("NewMultiple 2", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("AddBridgingRequestState").Return(fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())
		err := sm.NewMultiple(common.ChainIDStrPrime, []*indexer.Tx{{}})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to add some new BridgingRequestStates")
	})

	t.Run("NewMultiple 3", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("AddBridgingRequestState").Return(nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())
		err := sm.NewMultiple(common.ChainIDStrPrime, []*indexer.Tx{{}})
		require.NoError(t, err)
	})

	t.Run("Invalid 1", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState").Return(nil, fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.Invalid(common.BridgingRequestStateKey{SourceChainID: common.ChainIDStrPrime, SourceTxHash: "0xtest"})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get BridgingRequestState")
	})

	t.Run("Invalid 2", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState").Return(nil, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.Invalid(common.BridgingRequestStateKey{SourceChainID: common.ChainIDStrPrime, SourceTxHash: "0xtest"})
		require.Error(t, err)
		require.ErrorContains(t, err, "trying to get a non-existent BridgingRequestState")
	})

	t.Run("Invalid 3", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState").Return(&core.BridgingRequestState{}, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.Invalid(common.BridgingRequestStateKey{SourceChainID: common.ChainIDStrPrime, SourceTxHash: "0xtest"})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to update a BridgingRequestState")
	})

	t.Run("Invalid 4", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState").Return(core.NewBridgingRequestState(common.ChainIDStrPrime, "0xtest"), nil)
		db.On("UpdateBridgingRequestState").Return(fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.Invalid(common.BridgingRequestStateKey{SourceChainID: common.ChainIDStrPrime, SourceTxHash: "0xtest"})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to save updated BridgingRequestState")
	})

	t.Run("Invalid 5", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState").Return(core.NewBridgingRequestState(common.ChainIDStrPrime, "0xtest"), nil)
		db.On("UpdateBridgingRequestState").Return(nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.Invalid(common.BridgingRequestStateKey{SourceChainID: common.ChainIDStrPrime, SourceTxHash: "0xtest"})
		require.NoError(t, err)
	})

	t.Run("SubmittedToBridge 1", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState").Return(nil, fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToBridge(common.BridgingRequestStateKey{SourceChainID: common.ChainIDStrPrime, SourceTxHash: "0xtest"}, common.ChainIDStrVector)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get BridgingRequestState")
	})

	t.Run("SubmittedToBridge 2", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState").Return(nil, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToBridge(common.BridgingRequestStateKey{SourceChainID: common.ChainIDStrPrime, SourceTxHash: "0xtest"}, common.ChainIDStrVector)
		require.Error(t, err)
		require.ErrorContains(t, err, "trying to get a non-existent BridgingRequestState")
	})

	t.Run("SubmittedToBridge 3", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState").Return(&core.BridgingRequestState{}, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToBridge(common.BridgingRequestStateKey{SourceChainID: common.ChainIDStrPrime, SourceTxHash: "0xtest"}, common.ChainIDStrVector)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to update a BridgingRequestState")
	})

	t.Run("SubmittedToBridge 4", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState").Return(core.NewBridgingRequestState(common.ChainIDStrPrime, "0xtest"), nil)
		db.On("UpdateBridgingRequestState").Return(fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToBridge(common.BridgingRequestStateKey{SourceChainID: common.ChainIDStrPrime, SourceTxHash: "0xtest"}, common.ChainIDStrVector)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to save updated BridgingRequestState")
	})

	t.Run("SubmittedToBridge 5", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState").Return(core.NewBridgingRequestState(common.ChainIDStrPrime, "0xtest"), nil)
		db.On("UpdateBridgingRequestState").Return(nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToBridge(common.BridgingRequestStateKey{SourceChainID: common.ChainIDStrPrime, SourceTxHash: "0xtest"}, common.ChainIDStrVector)
		require.NoError(t, err)
	})

	t.Run("IncludedInBatch 1", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.IncludedInBatch(common.ChainIDStrVector, 1, []common.BridgingRequestStateKey{})
		require.NoError(t, err)
	})

	t.Run("IncludedInBatch 2", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState").Return(nil, fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.IncludedInBatch(common.ChainIDStrVector, 1, []common.BridgingRequestStateKey{{}})
		require.NoError(t, err)
	})

	t.Run("IncludedInBatch 3", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState").Return(&core.BridgingRequestState{}, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.IncludedInBatch(common.ChainIDStrVector, 1, []common.BridgingRequestStateKey{{}})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to update some BridgingRequestStates. errors: batch destinationChainId not equal to BridgingRequestState.DestinationChainId")
	})

	t.Run("IncludedInBatch 4", func(t *testing.T) {
		state := core.NewBridgingRequestState("", "")
		require.NoError(t, state.ToSubmittedToBridge(common.ChainIDStrVector))

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState").Return(state, nil)
		db.On("UpdateBridgingRequestState").Return(fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.IncludedInBatch(common.ChainIDStrVector, 1, []common.BridgingRequestStateKey{{}})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to update some BridgingRequestStates. errors: failed to save updated BridgingRequestState")
	})

	t.Run("IncludedInBatch 5", func(t *testing.T) {
		state := core.NewBridgingRequestState("", "")
		require.NoError(t, state.ToSubmittedToBridge(common.ChainIDStrVector))

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState").Return(state, nil)
		db.On("UpdateBridgingRequestState").Return(nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.IncludedInBatch(common.ChainIDStrVector, 1, []common.BridgingRequestStateKey{{}})
		require.NoError(t, err)
	})

	t.Run("SubmittedToDestination 1", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestStatesByBatchID").Return(nil, fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToDestination(common.ChainIDStrVector, 1)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get BridgingRequestStates")
	})

	t.Run("SubmittedToDestination 2", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestStatesByBatchID").Return(nil, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToDestination(common.ChainIDStrVector, 1)
		require.NoError(t, err)
	})

	t.Run("SubmittedToDestination 3", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestStatesByBatchID").Return([]*core.BridgingRequestState{{}}, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToDestination(common.ChainIDStrVector, 1)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to update some BridgingRequestStates. errors: failed to update a BridgingRequestState")
	})

	t.Run("SubmittedToDestination 4", func(t *testing.T) {
		state := core.NewBridgingRequestState("", "")
		require.NoError(t, state.ToSubmittedToBridge(common.ChainIDStrVector))
		require.NoError(t, state.ToIncludedInBatch(1))

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestStatesByBatchID").Return([]*core.BridgingRequestState{state}, nil)
		db.On("UpdateBridgingRequestState").Return(fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToDestination(common.ChainIDStrVector, 1)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to update some BridgingRequestStates. errors: failed to save updated BridgingRequestState")
	})

	t.Run("SubmittedToDestination 5", func(t *testing.T) {
		state := core.NewBridgingRequestState("", "")
		require.NoError(t, state.ToSubmittedToBridge(common.ChainIDStrVector))
		require.NoError(t, state.ToIncludedInBatch(1))

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestStatesByBatchID").Return([]*core.BridgingRequestState{state}, nil)
		db.On("UpdateBridgingRequestState").Return(nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToDestination(common.ChainIDStrVector, 1)
		require.NoError(t, err)
	})

	t.Run("FailedToExecuteOnDestination 1", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestStatesByBatchID").Return(nil, fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.FailedToExecuteOnDestination(common.ChainIDStrVector, 1)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get BridgingRequestStates")
	})

	t.Run("FailedToExecuteOnDestination 2", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestStatesByBatchID").Return(nil, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.FailedToExecuteOnDestination(common.ChainIDStrVector, 1)
		require.NoError(t, err)
	})

	t.Run("FailedToExecuteOnDestination 3", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestStatesByBatchID").Return([]*core.BridgingRequestState{{}}, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.FailedToExecuteOnDestination(common.ChainIDStrVector, 1)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to update some BridgingRequestStates. errors: failed to update a BridgingRequestState")
	})

	t.Run("FailedToExecuteOnDestination 4", func(t *testing.T) {
		state := core.NewBridgingRequestState("", "")
		require.NoError(t, state.ToSubmittedToBridge(common.ChainIDStrVector))
		require.NoError(t, state.ToIncludedInBatch(1))
		require.NoError(t, state.ToSubmittedToDestination())

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestStatesByBatchID").Return([]*core.BridgingRequestState{state}, nil)
		db.On("UpdateBridgingRequestState").Return(fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.FailedToExecuteOnDestination(common.ChainIDStrVector, 1)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to update some BridgingRequestStates. errors: failed to save updated BridgingRequestState")
	})

	t.Run("FailedToExecuteOnDestination 5", func(t *testing.T) {
		state := core.NewBridgingRequestState("", "")
		require.NoError(t, state.ToSubmittedToBridge(common.ChainIDStrVector))
		require.NoError(t, state.ToIncludedInBatch(1))
		require.NoError(t, state.ToSubmittedToDestination())

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestStatesByBatchID").Return([]*core.BridgingRequestState{state}, nil)
		db.On("UpdateBridgingRequestState").Return(nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.FailedToExecuteOnDestination(common.ChainIDStrVector, 1)
		require.NoError(t, err)
	})

	t.Run("ExecutedOnDestination 1", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestStatesByBatchID").Return(nil, fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.ExecutedOnDestination(common.ChainIDStrVector, 1, "")
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get BridgingRequestStates")
	})

	t.Run("ExecutedOnDestination 2", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestStatesByBatchID").Return(nil, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.ExecutedOnDestination(common.ChainIDStrVector, 1, "")
		require.NoError(t, err)
	})

	t.Run("ExecutedOnDestination 3", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestStatesByBatchID").Return([]*core.BridgingRequestState{{}}, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.ExecutedOnDestination(common.ChainIDStrVector, 1, "")
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to update some BridgingRequestStates. errors: failed to update a BridgingRequestState")
	})

	t.Run("ExecutedOnDestination 4", func(t *testing.T) {
		state := core.NewBridgingRequestState("", "")
		require.NoError(t, state.ToSubmittedToBridge(common.ChainIDStrVector))
		require.NoError(t, state.ToIncludedInBatch(1))
		require.NoError(t, state.ToSubmittedToDestination())

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestStatesByBatchID").Return([]*core.BridgingRequestState{state}, nil)
		db.On("UpdateBridgingRequestState").Return(fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.ExecutedOnDestination(common.ChainIDStrVector, 1, "")
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to update some BridgingRequestStates. errors: failed to save updated BridgingRequestState")
	})

	t.Run("ExecutedOnDestination 5", func(t *testing.T) {
		state := core.NewBridgingRequestState("", "")
		require.NoError(t, state.ToSubmittedToBridge(common.ChainIDStrVector))
		require.NoError(t, state.ToIncludedInBatch(1))
		require.NoError(t, state.ToSubmittedToDestination())

		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestStatesByBatchID").Return([]*core.BridgingRequestState{state}, nil)
		db.On("UpdateBridgingRequestState").Return(nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.ExecutedOnDestination(common.ChainIDStrVector, 1, "")
		require.NoError(t, err)
	})

	t.Run("Get 1", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState").Return(nil, fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		state, err := sm.Get(common.ChainIDStrPrime, "0xtest")
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get BridgingRequestState")
		require.Nil(t, state)
	})

	t.Run("Get 2", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState").Return(nil, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		state, err := sm.Get(common.ChainIDStrPrime, "0xtest")
		require.NoError(t, err)
		require.Nil(t, state)
	})

	t.Run("Get 3", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState").Return(&core.BridgingRequestState{}, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		state, err := sm.Get(common.ChainIDStrPrime, "0xtest")
		require.NoError(t, err)
		require.NotNil(t, state)
	})

	t.Run("GetMultiple 1", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState").Return(nil, fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		states, err := sm.GetMultiple(common.ChainIDStrPrime, []string{"0xtest"})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get some BridgingRequestStates")
		require.Nil(t, states)
	})

	t.Run("GetMultiple 2", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState").Return(nil, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		states, err := sm.GetMultiple(common.ChainIDStrPrime, []string{"0xtest"})
		require.NoError(t, err)
		require.Len(t, states, 0)
	})

	t.Run("GetMultiple 3", func(t *testing.T) {
		db := &databaseaccess.BridgingRequestStateDBMock{}
		db.On("GetBridgingRequestState").Return(&core.BridgingRequestState{}, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		states, err := sm.GetMultiple(common.ChainIDStrPrime, []string{"0xtest"})
		require.NoError(t, err)
		require.NotNil(t, states)
		require.Len(t, states, 1)
	})
}
