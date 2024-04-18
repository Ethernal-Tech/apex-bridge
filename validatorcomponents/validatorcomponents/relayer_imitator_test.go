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
		db := &relayerDb.DbMock{}

		ri, err := NewRelayerImitator(nil, brsUpdater, bsc, db, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, ri)
	})

	t.Run("execute 1", func(t *testing.T) {
		chainId := "prime"
		ctx := context.Background()

		brsUpdater := &common.BridgingRequestStateUpdaterMock{}

		bsc := &eth.BridgeSmartContractMock{}
		bsc.On("GetConfirmedBatch", ctx, chainId).Return(nil, fmt.Errorf("test err"))

		db := &relayerDb.DbMock{}

		ri, _ := NewRelayerImitator(nil, brsUpdater, bsc, db, hclog.NewNullLogger())

		err := ri.execute(ctx, chainId)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to retrieve confirmed batch")
	})

	t.Run("execute 2", func(t *testing.T) {
		chainId := "prime"
		ctx := context.Background()

		brsUpdater := &common.BridgingRequestStateUpdaterMock{}

		bsc := &eth.BridgeSmartContractMock{}
		bsc.On("GetConfirmedBatch", ctx, chainId).Return(&eth.ConfirmedBatch{Id: "1"}, nil)

		db := &relayerDb.DbMock{}
		db.On("GetLastSubmittedBatchId", chainId).Return(nil, fmt.Errorf("test err"))

		ri, _ := NewRelayerImitator(nil, brsUpdater, bsc, db, hclog.NewNullLogger())

		err := ri.execute(ctx, chainId)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get last submitted batch id from db")
	})

	t.Run("execute 3", func(t *testing.T) {
		chainId := "prime"
		ctx := context.Background()

		brsUpdater := &common.BridgingRequestStateUpdaterMock{}

		bsc := &eth.BridgeSmartContractMock{}
		bsc.On("GetConfirmedBatch", ctx, chainId).Return(&eth.ConfirmedBatch{Id: ""}, nil)

		db := &relayerDb.DbMock{}
		db.On("GetLastSubmittedBatchId", chainId).Return(nil, nil)

		ri, _ := NewRelayerImitator(nil, brsUpdater, bsc, db, hclog.NewNullLogger())

		err := ri.execute(ctx, chainId)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to convert confirmed batch id to big int")
	})

	t.Run("execute 4", func(t *testing.T) {
		chainId := "prime"
		ctx := context.Background()

		brsUpdater := &common.BridgingRequestStateUpdaterMock{}

		bsc := &eth.BridgeSmartContractMock{}
		bsc.On("GetConfirmedBatch", ctx, chainId).Return(&eth.ConfirmedBatch{Id: "1"}, nil)

		db := &relayerDb.DbMock{}
		db.On("GetLastSubmittedBatchId", chainId).Return(big.NewInt(1), nil)

		ri, _ := NewRelayerImitator(nil, brsUpdater, bsc, db, hclog.NewNullLogger())

		err := ri.execute(ctx, chainId)
		require.NoError(t, err)
	})

	t.Run("execute 5", func(t *testing.T) {
		chainId := "prime"
		ctx := context.Background()

		brsUpdater := &common.BridgingRequestStateUpdaterMock{}

		bsc := &eth.BridgeSmartContractMock{}
		bsc.On("GetConfirmedBatch", ctx, chainId).Return(&eth.ConfirmedBatch{Id: "1"}, nil)

		db := &relayerDb.DbMock{}
		db.On("GetLastSubmittedBatchId", chainId).Return(big.NewInt(2), nil)

		ri, _ := NewRelayerImitator(nil, brsUpdater, bsc, db, hclog.NewNullLogger())

		err := ri.execute(ctx, chainId)
		require.Error(t, err)
		require.ErrorContains(t, err, "last submitted batch id greater than received")
	})

	t.Run("execute 6", func(t *testing.T) {
		chainId := "prime"
		ctx := context.Background()

		brsUpdater := &common.BridgingRequestStateUpdaterMock{}
		brsUpdater.On("SubmittedToDestination").Return(nil)

		bsc := &eth.BridgeSmartContractMock{}
		bsc.On("GetConfirmedBatch", ctx, chainId).Return(&eth.ConfirmedBatch{Id: "2"}, nil)

		db := &relayerDb.DbMock{}
		db.On("GetLastSubmittedBatchId", chainId).Return(nil, nil)
		db.On("AddLastSubmittedBatchId", chainId, big.NewInt(2)).Return(fmt.Errorf("test err"))

		ri, _ := NewRelayerImitator(nil, brsUpdater, bsc, db, hclog.NewNullLogger())

		err := ri.execute(ctx, chainId)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to insert last submitted batch id into db")
	})

	t.Run("execute 7", func(t *testing.T) {
		chainId := "prime"
		ctx := context.Background()

		brsUpdater := &common.BridgingRequestStateUpdaterMock{}
		brsUpdater.On("SubmittedToDestination").Return(nil)

		bsc := &eth.BridgeSmartContractMock{}
		bsc.On("GetConfirmedBatch", ctx, chainId).Return(&eth.ConfirmedBatch{Id: "2"}, nil)

		db := &relayerDb.DbMock{}
		db.On("GetLastSubmittedBatchId", chainId).Return(big.NewInt(1), nil)
		db.On("AddLastSubmittedBatchId", chainId, big.NewInt(2)).Return(fmt.Errorf("test err"))

		ri, _ := NewRelayerImitator(nil, brsUpdater, bsc, db, hclog.NewNullLogger())

		err := ri.execute(ctx, chainId)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to insert last submitted batch id into db")
	})

	t.Run("execute 8", func(t *testing.T) {
		chainId := "prime"
		ctx := context.Background()

		brsUpdater := &common.BridgingRequestStateUpdaterMock{}
		brsUpdater.On("SubmittedToDestination").Return(nil)

		bsc := &eth.BridgeSmartContractMock{}
		bsc.On("GetConfirmedBatch", ctx, chainId).Return(&eth.ConfirmedBatch{Id: "2"}, nil)

		db := &relayerDb.DbMock{}
		db.On("GetLastSubmittedBatchId", chainId).Return(nil, nil)
		db.On("AddLastSubmittedBatchId", chainId, big.NewInt(2)).Return(nil)

		ri, _ := NewRelayerImitator(nil, brsUpdater, bsc, db, hclog.NewNullLogger())

		err := ri.execute(ctx, chainId)
		require.NoError(t, err)
	})

	t.Run("execute 9", func(t *testing.T) {
		chainId := "prime"
		ctx := context.Background()

		brsUpdater := &common.BridgingRequestStateUpdaterMock{}
		brsUpdater.On("SubmittedToDestination").Return(nil)

		bsc := &eth.BridgeSmartContractMock{}
		bsc.On("GetConfirmedBatch", ctx, chainId).Return(&eth.ConfirmedBatch{Id: "2"}, nil)

		db := &relayerDb.DbMock{}
		db.On("GetLastSubmittedBatchId", chainId).Return(big.NewInt(1), nil)
		db.On("AddLastSubmittedBatchId", chainId, big.NewInt(2)).Return(nil)

		ri, _ := NewRelayerImitator(nil, brsUpdater, bsc, db, hclog.NewNullLogger())

		err := ri.execute(ctx, chainId)
		require.NoError(t, err)
	})
}