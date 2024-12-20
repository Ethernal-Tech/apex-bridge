package validatorcomponents

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	relayerDb "github.com/Ethernal-Tech/apex-bridge/relayer/database_access"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestRelayerImitator(t *testing.T) {
	t.Run("NewRelayerImitator", func(t *testing.T) {
		brsUpdater := &common.BridgingRequestStateUpdaterMock{}
		bsc := &eth.BridgeSmartContractMock{}
		db := &relayerDb.DBMock{}

		ri, err := NewRelayerImitator(nil, brsUpdater, bsc, db, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, ri)
	})

	const chainID = common.ChainIDStrPrime

	t.Run("execute 1", func(t *testing.T) {
		ctx := context.Background()

		brsUpdater := &common.BridgingRequestStateUpdaterMock{}

		bsc := &eth.BridgeSmartContractMock{}
		bsc.On("GetConfirmedBatch", ctx, chainID).Return(nil, fmt.Errorf("test err"))

		db := &relayerDb.DBMock{}

		ri, _ := NewRelayerImitator(nil, brsUpdater, bsc, db, hclog.NewNullLogger())

		err := ri.execute(ctx, chainID)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to retrieve confirmed batch")
	})

	t.Run("execute 2", func(t *testing.T) {
		ctx := context.Background()

		brsUpdater := &common.BridgingRequestStateUpdaterMock{}

		bsc := &eth.BridgeSmartContractMock{}
		bsc.On("GetConfirmedBatch", ctx, chainID).Return(&eth.ConfirmedBatch{ID: 1}, nil)

		db := &relayerDb.DBMock{}
		db.On("GetLastSubmittedBatchID", chainID).Return(nil, fmt.Errorf("test err"))

		ri, _ := NewRelayerImitator(nil, brsUpdater, bsc, db, hclog.NewNullLogger())

		err := ri.execute(ctx, chainID)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get last submitted batch id from db")
	})

	t.Run("execute 3", func(t *testing.T) {
		ctx := context.Background()

		brsUpdater := &common.BridgingRequestStateUpdaterMock{}

		bsc := &eth.BridgeSmartContractMock{}
		bsc.On("GetConfirmedBatch", ctx, chainID).Return(&eth.ConfirmedBatch{ID: 1}, nil)

		db := &relayerDb.DBMock{}
		db.On("GetLastSubmittedBatchID", chainID).Return(big.NewInt(1), nil)

		ri, _ := NewRelayerImitator(nil, brsUpdater, bsc, db, hclog.NewNullLogger())

		err := ri.execute(ctx, chainID)
		require.NoError(t, err)
	})

	t.Run("execute 4", func(t *testing.T) {
		ctx := context.Background()

		brsUpdater := &common.BridgingRequestStateUpdaterMock{}

		bsc := &eth.BridgeSmartContractMock{}
		bsc.On("GetConfirmedBatch", ctx, chainID).Return(&eth.ConfirmedBatch{ID: 1}, nil)

		db := &relayerDb.DBMock{}
		db.On("GetLastSubmittedBatchID", chainID).Return(big.NewInt(2), nil)

		ri, _ := NewRelayerImitator(nil, brsUpdater, bsc, db, hclog.NewNullLogger())

		err := ri.execute(ctx, chainID)
		require.Error(t, err)
		require.ErrorContains(t, err, "last submitted batch id greater than received")
	})

	t.Run("execute 5", func(t *testing.T) {
		ctx := context.Background()

		brsUpdater := &common.BridgingRequestStateUpdaterMock{}
		brsUpdater.On("SubmittedToDestination").Return(nil)

		bsc := &eth.BridgeSmartContractMock{}
		bsc.On("GetConfirmedBatch", ctx, chainID).Return(&eth.ConfirmedBatch{ID: 2}, nil)

		db := &relayerDb.DBMock{}
		db.On("GetLastSubmittedBatchID", chainID).Return(nil, nil)
		db.On("AddLastSubmittedBatchID", chainID, big.NewInt(2)).Return(fmt.Errorf("test err"))

		ri, _ := NewRelayerImitator(nil, brsUpdater, bsc, db, hclog.NewNullLogger())

		err := ri.execute(ctx, chainID)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to insert last submitted batch id into db")
	})

	t.Run("execute 6", func(t *testing.T) {
		ctx := context.Background()

		brsUpdater := &common.BridgingRequestStateUpdaterMock{}
		brsUpdater.On("SubmittedToDestination").Return(nil)

		bsc := &eth.BridgeSmartContractMock{}
		bsc.On("GetConfirmedBatch", ctx, chainID).Return(&eth.ConfirmedBatch{ID: 2}, nil)

		db := &relayerDb.DBMock{}
		db.On("GetLastSubmittedBatchID", chainID).Return(big.NewInt(1), nil)
		db.On("AddLastSubmittedBatchID", chainID, big.NewInt(2)).Return(fmt.Errorf("test err"))

		ri, _ := NewRelayerImitator(nil, brsUpdater, bsc, db, hclog.NewNullLogger())

		err := ri.execute(ctx, chainID)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to insert last submitted batch id into db")
	})

	t.Run("execute 7", func(t *testing.T) {
		ctx := context.Background()

		brsUpdater := &common.BridgingRequestStateUpdaterMock{}
		brsUpdater.On("SubmittedToDestination").Return(nil)

		bsc := &eth.BridgeSmartContractMock{}
		bsc.On("GetConfirmedBatch", ctx, chainID).Return(&eth.ConfirmedBatch{ID: 2}, nil)

		db := &relayerDb.DBMock{}
		db.On("GetLastSubmittedBatchID", chainID).Return(nil, nil)
		db.On("AddLastSubmittedBatchID", chainID, big.NewInt(2)).Return(nil)

		ri, _ := NewRelayerImitator(nil, brsUpdater, bsc, db, hclog.NewNullLogger())

		err := ri.execute(ctx, chainID)
		require.NoError(t, err)
	})

	t.Run("execute 8", func(t *testing.T) {
		ctx := context.Background()

		brsUpdater := &common.BridgingRequestStateUpdaterMock{}
		brsUpdater.On("SubmittedToDestination").Return(nil)

		bsc := &eth.BridgeSmartContractMock{}
		bsc.On("GetConfirmedBatch", ctx, chainID).Return(&eth.ConfirmedBatch{ID: 2}, nil)

		db := &relayerDb.DBMock{}
		db.On("GetLastSubmittedBatchID", chainID).Return(big.NewInt(1), nil)
		db.On("AddLastSubmittedBatchID", chainID, big.NewInt(2)).Return(nil)

		ri, _ := NewRelayerImitator(nil, brsUpdater, bsc, db, hclog.NewNullLogger())

		err := ri.execute(ctx, chainID)
		require.NoError(t, err)
	})
}
