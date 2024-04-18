package validatorcomponents

import (
	"fmt"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/database_access"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestBridgingRequestStateManager(t *testing.T) {
	t.Run("NewBridgingRequestStateManager", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}

		sm, err := NewBridgingRequestStateManager(db, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, sm)
	})

	t.Run("New 1", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("AddBridgingRequestState").Return(fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())
		err := sm.New("prime", &indexer.Tx{})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to add new BridgingRequestState")
	})

	t.Run("New 2", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("AddBridgingRequestState").Return(nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())
		err := sm.New("prime", &indexer.Tx{})
		require.NoError(t, err)
	})

	t.Run("NewMultiple 1", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())
		err := sm.NewMultiple("prime", []*indexer.Tx{})
		require.NoError(t, err)
	})

	t.Run("NewMultiple 2", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("AddBridgingRequestState").Return(fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())
		err := sm.NewMultiple("prime", []*indexer.Tx{{}})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to add some new BridgingRequestStates")
	})

	t.Run("NewMultiple 3", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("AddBridgingRequestState").Return(nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())
		err := sm.NewMultiple("prime", []*indexer.Tx{{}})
		require.NoError(t, err)
	})

	t.Run("Invalid 1", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestState").Return(nil, fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.Invalid(common.BridgingRequestStateKey{SourceChainId: "prime", SourceTxHash: "0xtest"})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get BridgingRequestState")
	})

	t.Run("Invalid 2", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestState").Return(nil, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.Invalid(common.BridgingRequestStateKey{SourceChainId: "prime", SourceTxHash: "0xtest"})
		require.Error(t, err)
		require.ErrorContains(t, err, "trying to get a non-existent BridgingRequestState")
	})

	t.Run("Invalid 3", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestState").Return(&core.BridgingRequestState{}, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.Invalid(common.BridgingRequestStateKey{SourceChainId: "prime", SourceTxHash: "0xtest"})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to update a BridgingRequestState")
	})

	t.Run("Invalid 4", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestState").Return(core.NewBridgingRequestState("prime", "0xtest", nil), nil)
		db.On("UpdateBridgingRequestState").Return(fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.Invalid(common.BridgingRequestStateKey{SourceChainId: "prime", SourceTxHash: "0xtest"})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to save updated BridgingRequestState")
	})

	t.Run("Invalid 5", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestState").Return(core.NewBridgingRequestState("prime", "0xtest", nil), nil)
		db.On("UpdateBridgingRequestState").Return(nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.Invalid(common.BridgingRequestStateKey{SourceChainId: "prime", SourceTxHash: "0xtest"})
		require.NoError(t, err)
	})

	t.Run("SubmittedToBridge 1", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestState").Return(nil, fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToBridge(common.BridgingRequestStateKey{SourceChainId: "prime", SourceTxHash: "0xtest"}, "vector")
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get BridgingRequestState")
	})

	t.Run("SubmittedToBridge 2", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestState").Return(nil, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToBridge(common.BridgingRequestStateKey{SourceChainId: "prime", SourceTxHash: "0xtest"}, "vector")
		require.Error(t, err)
		require.ErrorContains(t, err, "trying to get a non-existent BridgingRequestState")
	})

	t.Run("SubmittedToBridge 3", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestState").Return(&core.BridgingRequestState{}, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToBridge(common.BridgingRequestStateKey{SourceChainId: "prime", SourceTxHash: "0xtest"}, "vector")
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to update a BridgingRequestState")
	})

	t.Run("SubmittedToBridge 4", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestState").Return(core.NewBridgingRequestState("prime", "0xtest", nil), nil)
		db.On("UpdateBridgingRequestState").Return(fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToBridge(common.BridgingRequestStateKey{SourceChainId: "prime", SourceTxHash: "0xtest"}, "vector")
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to save updated BridgingRequestState")
	})

	t.Run("SubmittedToBridge 5", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestState").Return(core.NewBridgingRequestState("prime", "0xtest", nil), nil)
		db.On("UpdateBridgingRequestState").Return(nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToBridge(common.BridgingRequestStateKey{SourceChainId: "prime", SourceTxHash: "0xtest"}, "vector")
		require.NoError(t, err)
	})

	t.Run("IncludedInBatch 1", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.IncludedInBatch("vector", 1, []common.BridgingRequestStateKey{})
		require.NoError(t, err)
	})

	t.Run("IncludedInBatch 2", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestState").Return(nil, fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.IncludedInBatch("vector", 1, []common.BridgingRequestStateKey{{}})
		require.NoError(t, err)
	})

	t.Run("IncludedInBatch 3", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestState").Return(&core.BridgingRequestState{}, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.IncludedInBatch("vector", 1, []common.BridgingRequestStateKey{{}})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to update some BridgingRequestStates. errors: batch destinationChainId not equal to BridgingRequestState.DestinationChainId")
	})

	t.Run("IncludedInBatch 4", func(t *testing.T) {
		state := core.NewBridgingRequestState("", "", nil)
		state.ToSubmittedToBridge("vector")

		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestState").Return(state, nil)
		db.On("UpdateBridgingRequestState").Return(fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.IncludedInBatch("vector", 1, []common.BridgingRequestStateKey{{}})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to update some BridgingRequestStates. errors: failed to save updated BridgingRequestState")
	})

	t.Run("IncludedInBatch 5", func(t *testing.T) {
		state := core.NewBridgingRequestState("", "", nil)
		state.ToSubmittedToBridge("vector")

		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestState").Return(state, nil)
		db.On("UpdateBridgingRequestState").Return(nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.IncludedInBatch("vector", 1, []common.BridgingRequestStateKey{{}})
		require.NoError(t, err)
	})

	t.Run("SubmittedToDestination 1", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestStatesByBatchId").Return(nil, fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToDestination("vector", 1)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get BridgingRequestStates")
	})

	t.Run("SubmittedToDestination 2", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestStatesByBatchId").Return(nil, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToDestination("vector", 1)
		require.NoError(t, err)
	})

	t.Run("SubmittedToDestination 3", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestStatesByBatchId").Return([]*core.BridgingRequestState{{}}, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToDestination("vector", 1)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to update some BridgingRequestStates. errors: failed to update a BridgingRequestState")
	})

	t.Run("SubmittedToDestination 4", func(t *testing.T) {
		state := core.NewBridgingRequestState("", "", nil)
		state.ToSubmittedToBridge("vector")
		state.ToIncludedInBatch(1)

		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestStatesByBatchId").Return([]*core.BridgingRequestState{state}, nil)
		db.On("UpdateBridgingRequestState").Return(fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToDestination("vector", 1)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to update some BridgingRequestStates. errors: failed to save updated BridgingRequestState")
	})

	t.Run("SubmittedToDestination 5", func(t *testing.T) {
		state := core.NewBridgingRequestState("", "", nil)
		state.ToSubmittedToBridge("vector")
		state.ToIncludedInBatch(1)

		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestStatesByBatchId").Return([]*core.BridgingRequestState{state}, nil)
		db.On("UpdateBridgingRequestState").Return(nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.SubmittedToDestination("vector", 1)
		require.NoError(t, err)
	})

	t.Run("FailedToExecuteOnDestination 1", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestStatesByBatchId").Return(nil, fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.FailedToExecuteOnDestination("vector", 1)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get BridgingRequestStates")
	})

	t.Run("FailedToExecuteOnDestination 2", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestStatesByBatchId").Return(nil, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.FailedToExecuteOnDestination("vector", 1)
		require.NoError(t, err)
	})

	t.Run("FailedToExecuteOnDestination 3", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestStatesByBatchId").Return([]*core.BridgingRequestState{{}}, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.FailedToExecuteOnDestination("vector", 1)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to update some BridgingRequestStates. errors: failed to update a BridgingRequestState")
	})

	t.Run("FailedToExecuteOnDestination 4", func(t *testing.T) {
		state := core.NewBridgingRequestState("", "", nil)
		state.ToSubmittedToBridge("vector")
		state.ToIncludedInBatch(1)
		state.ToSubmittedToDestination()

		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestStatesByBatchId").Return([]*core.BridgingRequestState{state}, nil)
		db.On("UpdateBridgingRequestState").Return(fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.FailedToExecuteOnDestination("vector", 1)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to update some BridgingRequestStates. errors: failed to save updated BridgingRequestState")
	})

	t.Run("FailedToExecuteOnDestination 5", func(t *testing.T) {
		state := core.NewBridgingRequestState("", "", nil)
		state.ToSubmittedToBridge("vector")
		state.ToIncludedInBatch(1)
		state.ToSubmittedToDestination()

		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestStatesByBatchId").Return([]*core.BridgingRequestState{state}, nil)
		db.On("UpdateBridgingRequestState").Return(nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.FailedToExecuteOnDestination("vector", 1)
		require.NoError(t, err)
	})

	t.Run("ExecutedOnDestination 1", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestStatesByBatchId").Return(nil, fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.ExecutedOnDestination("vector", 1, "")
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get BridgingRequestStates")
	})

	t.Run("ExecutedOnDestination 2", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestStatesByBatchId").Return(nil, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.ExecutedOnDestination("vector", 1, "")
		require.NoError(t, err)
	})

	t.Run("ExecutedOnDestination 3", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestStatesByBatchId").Return([]*core.BridgingRequestState{{}}, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.ExecutedOnDestination("vector", 1, "")
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to update some BridgingRequestStates. errors: failed to update a BridgingRequestState")
	})

	t.Run("ExecutedOnDestination 4", func(t *testing.T) {
		state := core.NewBridgingRequestState("", "", nil)
		state.ToSubmittedToBridge("vector")
		state.ToIncludedInBatch(1)
		state.ToSubmittedToDestination()

		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestStatesByBatchId").Return([]*core.BridgingRequestState{state}, nil)
		db.On("UpdateBridgingRequestState").Return(fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.ExecutedOnDestination("vector", 1, "")
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to update some BridgingRequestStates. errors: failed to save updated BridgingRequestState")
	})

	t.Run("ExecutedOnDestination 5", func(t *testing.T) {
		state := core.NewBridgingRequestState("", "", nil)
		state.ToSubmittedToBridge("vector")
		state.ToIncludedInBatch(1)
		state.ToSubmittedToDestination()

		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestStatesByBatchId").Return([]*core.BridgingRequestState{state}, nil)
		db.On("UpdateBridgingRequestState").Return(nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		err := sm.ExecutedOnDestination("vector", 1, "")
		require.NoError(t, err)
	})

	t.Run("Get 1", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestState").Return(nil, fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		state, err := sm.Get("prime", "0xtest")
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get BridgingRequestState")
		require.Nil(t, state)
	})

	t.Run("Get 2", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestState").Return(nil, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		state, err := sm.Get("prime", "0xtest")
		require.NoError(t, err)
		require.Nil(t, state)
	})

	t.Run("Get 3", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetBridgingRequestState").Return(&core.BridgingRequestState{}, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		state, err := sm.Get("prime", "0xtest")
		require.NoError(t, err)
		require.NotNil(t, state)
	})

	t.Run("GetAllForUser 1", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetUserBridgingRequestStates").Return(nil, fmt.Errorf("test err"))

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		states, err := sm.GetAllForUser("prime", "0xtest")
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get all BridgingRequestStates for user")
		require.Nil(t, states)
	})

	t.Run("GetAllForUser 2", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetUserBridgingRequestStates").Return(nil, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		states, err := sm.GetAllForUser("prime", "0xtest")
		require.NoError(t, err)
		require.Nil(t, states)
	})

	t.Run("GetAllForUser 3", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetUserBridgingRequestStates").Return([]*core.BridgingRequestState{}, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		states, err := sm.GetAllForUser("prime", "0xtest")
		require.NoError(t, err)
		require.NotNil(t, states)
		require.Len(t, states, 0)
	})

	t.Run("GetAllForUser 4", func(t *testing.T) {
		db := &database_access.BridgingRequestStateDbMock{}
		db.On("GetUserBridgingRequestStates").Return([]*core.BridgingRequestState{{}}, nil)

		sm, _ := NewBridgingRequestStateManager(db, hclog.NewNullLogger())

		states, err := sm.GetAllForUser("prime", "0xtest")
		require.NoError(t, err)
		require.NotNil(t, states)
		require.Len(t, states, 1)
	})
}
