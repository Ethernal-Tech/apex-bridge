package chain

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	paramsFromChain = []byte("protocol-params-fetched-from-chain")
	paramsFreshInDB = []byte("fresh-protocol-params-in-db")
	paramsStaleInDB = []byte("stale-protocol-params-in-db")
)

func newTestChainInfo(
	ctx context.Context, db *cCore.ProtocolParamsDBMock, txProvider *cardanotx.TxProviderTestMock,
) *CardanoChainInfo {
	return &CardanoChainInfo{
		config:     &cCore.CardanoChainConfig{ChainID: "prime"},
		ctx:        ctx,
		db:         db,
		logger:     hclog.NewNullLogger(),
		txProvider: txProvider,
	}
}

func freshExpiry() time.Time { return time.Now().UTC().Add(time.Hour) }
func staleExpiry() time.Time { return time.Now().UTC().Add(-time.Hour) }

func TestCardanoChainInfo_initialize(t *testing.T) {
	t.Run("db has fresh params - uses them without touching the chain", func(t *testing.T) {
		expiry := freshExpiry()
		db := &cCore.ProtocolParamsDBMock{Params: paramsFreshInDB, ExpiresAt: expiry}
		txProvider := &cardanotx.TxProviderTestMock{}

		info := newTestChainInfo(context.Background(), db, txProvider)

		require.NoError(t, info.initialize())

		txProvider.AssertNotCalled(t, "GetProtocolParameters")
		require.Equal(t, paramsFreshInDB, db.Params)
		require.Equal(t, expiry, db.ExpiresAt)
	})

	t.Run("db has stale params and fetch succeeds - refreshes and persists", func(t *testing.T) {
		db := &cCore.ProtocolParamsDBMock{Params: paramsStaleInDB, ExpiresAt: staleExpiry()}
		txProvider := &cardanotx.TxProviderTestMock{}
		txProvider.On("GetProtocolParameters", mock.Anything).Return(paramsFromChain, nil).Once()

		info := newTestChainInfo(context.Background(), db, txProvider)

		require.NoError(t, info.initialize())

		txProvider.AssertExpectations(t)
		require.Equal(t, paramsFromChain, db.Params)
		require.True(t, db.ExpiresAt.After(time.Now().UTC()))
	})

	t.Run("db has stale params and fetch fails - keeps stale params and continues", func(t *testing.T) {
		staleAt := staleExpiry()
		db := &cCore.ProtocolParamsDBMock{Params: paramsStaleInDB, ExpiresAt: staleAt}
		txProvider := &cardanotx.TxProviderTestMock{}
		txProvider.On("GetProtocolParameters", mock.Anything).Return(nil, errors.New("chain unreachable")).Once()

		info := newTestChainInfo(context.Background(), db, txProvider)

		// startup is not aborted by a failed refresh - the stale params are left untouched
		require.NoError(t, info.initialize())

		txProvider.AssertExpectations(t)
		require.Equal(t, paramsStaleInDB, db.Params)
		require.Equal(t, staleAt, db.ExpiresAt)
	})

	t.Run("db has stale params and context is done during refresh - returns error", func(t *testing.T) {
		staleAt := staleExpiry()
		db := &cCore.ProtocolParamsDBMock{Params: paramsStaleInDB, ExpiresAt: staleAt}
		txProvider := &cardanotx.TxProviderTestMock{}
		txProvider.On("GetProtocolParameters", mock.Anything).Return(nil, context.Canceled).Once()

		info := newTestChainInfo(context.Background(), db, txProvider)

		require.ErrorIs(t, info.initialize(), context.Canceled)
		require.Equal(t, paramsStaleInDB, db.Params)
		require.Equal(t, staleAt, db.ExpiresAt)
	})

	t.Run("db is empty and fetch succeeds - persists fetched params", func(t *testing.T) {
		db := &cCore.ProtocolParamsDBMock{}
		txProvider := &cardanotx.TxProviderTestMock{}
		txProvider.On("GetProtocolParameters", mock.Anything).Return(paramsFromChain, nil).Once()

		info := newTestChainInfo(context.Background(), db, txProvider)

		require.NoError(t, info.initialize())

		txProvider.AssertExpectations(t)
		require.Equal(t, paramsFromChain, db.Params)
		require.True(t, db.ExpiresAt.After(time.Now().UTC()))
	})

	t.Run("db is empty and fetch fails - keeps retrying, bounded only by the context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		db := &cCore.ProtocolParamsDBMock{}
		txProvider := &cardanotx.TxProviderTestMock{}
		// cancel on the first attempt so the retry-forever loop exits instead of blocking
		txProvider.On("GetProtocolParameters", mock.Anything).
			Run(func(mock.Arguments) { cancel() }).
			Return(nil, errors.New("chain unreachable"))

		info := newTestChainInfo(ctx, db, txProvider)

		require.Error(t, info.initialize())
		txProvider.AssertCalled(t, "GetProtocolParameters", mock.Anything)
		require.Empty(t, db.Params)
	})

	t.Run("db is empty and persisting fails - does not silently succeed", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// fetch works but the db write fails: this must be retried, never treated as success
		db := &cCore.ProtocolParamsDBMock{SaveErr: errors.New("db write failed")}
		txProvider := &cardanotx.TxProviderTestMock{}
		txProvider.On("GetProtocolParameters", mock.Anything).
			Run(func(mock.Arguments) { cancel() }).
			Return(paramsFromChain, nil)

		info := newTestChainInfo(ctx, db, txProvider)

		require.Error(t, info.initialize())
		txProvider.AssertCalled(t, "GetProtocolParameters", mock.Anything)
		require.Empty(t, db.Params)
	})

	t.Run("db read fails - returns error", func(t *testing.T) {
		db := &cCore.ProtocolParamsDBMock{GetErr: errors.New("db read failed")}
		txProvider := &cardanotx.TxProviderTestMock{}

		info := newTestChainInfo(context.Background(), db, txProvider)

		require.Error(t, info.initialize())
		txProvider.AssertNotCalled(t, "GetProtocolParameters")
	})
}

func TestCardanoChainInfo_GetProtocolParams(t *testing.T) {
	t.Run("db has fresh params - returns them without touching the chain", func(t *testing.T) {
		db := &cCore.ProtocolParamsDBMock{Params: paramsFreshInDB, ExpiresAt: freshExpiry()}
		txProvider := &cardanotx.TxProviderTestMock{}

		info := newTestChainInfo(context.Background(), db, txProvider)

		require.Equal(t, paramsFreshInDB, info.GetProtocolParams())
		txProvider.AssertNotCalled(t, "GetProtocolParameters")
	})

	t.Run("db has stale params and fetch succeeds - refreshes, persists and returns fresh", func(t *testing.T) {
		db := &cCore.ProtocolParamsDBMock{Params: paramsStaleInDB, ExpiresAt: staleExpiry()}
		txProvider := &cardanotx.TxProviderTestMock{}
		txProvider.On("GetProtocolParameters", mock.Anything).Return(paramsFromChain, nil).Once()

		info := newTestChainInfo(context.Background(), db, txProvider)

		require.Equal(t, paramsFromChain, info.GetProtocolParams())
		txProvider.AssertExpectations(t)
		require.Equal(t, paramsFromChain, db.Params)
		require.True(t, db.ExpiresAt.After(time.Now().UTC()))
	})

	t.Run("db has stale params and fetch fails - returns stale params without error", func(t *testing.T) {
		staleAt := staleExpiry()
		db := &cCore.ProtocolParamsDBMock{Params: paramsStaleInDB, ExpiresAt: staleAt}
		txProvider := &cardanotx.TxProviderTestMock{}
		txProvider.On("GetProtocolParameters", mock.Anything).Return(nil, errors.New("chain unreachable")).Once()

		info := newTestChainInfo(context.Background(), db, txProvider)

		require.Equal(t, paramsStaleInDB, info.GetProtocolParams())
		txProvider.AssertExpectations(t)
		require.Equal(t, paramsStaleInDB, db.Params)
		require.Equal(t, staleAt, db.ExpiresAt)
	})

	t.Run("db is unexpectedly empty - fetches, persists and returns", func(t *testing.T) {
		db := &cCore.ProtocolParamsDBMock{}
		txProvider := &cardanotx.TxProviderTestMock{}
		txProvider.On("GetProtocolParameters", mock.Anything).Return(paramsFromChain, nil).Once()

		info := newTestChainInfo(context.Background(), db, txProvider)

		require.Equal(t, paramsFromChain, info.GetProtocolParams())
		txProvider.AssertExpectations(t)
		require.Equal(t, paramsFromChain, db.Params)
	})

	t.Run("db read fails persistently - returns empty without error", func(t *testing.T) {
		// even with a completely broken db, the processors must get a value, never an error
		db := &cCore.ProtocolParamsDBMock{GetErr: errors.New("db read failed")}
		txProvider := &cardanotx.TxProviderTestMock{}
		txProvider.On("GetProtocolParameters", mock.Anything).Return(paramsFromChain, nil).Once()

		info := newTestChainInfo(context.Background(), db, txProvider)

		require.Empty(t, info.GetProtocolParams())
	})

	t.Run("concurrent callers are serialized and all get valid params", func(t *testing.T) {
		db := &cCore.ProtocolParamsDBMock{Params: paramsStaleInDB, ExpiresAt: staleExpiry()}
		txProvider := &cardanotx.TxProviderTestMock{}
		txProvider.On("GetProtocolParameters", mock.Anything).Return(paramsFromChain, nil)

		info := newTestChainInfo(context.Background(), db, txProvider)

		const callers = 20

		var (
			wg      sync.WaitGroup
			results = make([][]byte, callers)
		)

		wg.Add(callers)

		for i := range results {
			go func(idx int) {
				defer wg.Done()

				results[idx] = info.GetProtocolParams()
			}(i)
		}

		wg.Wait()

		for _, res := range results {
			require.Equal(t, paramsFromChain, res)
		}
	})
}

func TestNewCardanoChainInfo(t *testing.T) {
	t.Run("fails when a tx provider cannot be created", func(t *testing.T) {
		info, err := NewCardanoChainInfo(
			context.Background(), &cCore.CardanoChainConfig{}, &cCore.ProtocolParamsDBMock{}, hclog.NewNullLogger())

		require.Error(t, err)
		require.Nil(t, info)
	})
}
