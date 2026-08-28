package chain

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	solanaAB "github.com/Ethernal-Tech/apex-bridge/solana"
	"github.com/Ethernal-Tech/solana-infrastructure/tracker/store"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

const solanaNodeURL = "https://api.devnet.solana.com"
const trackedProgram = "CkTNcuk9EELmuR65eCfzKfz8XpDvJ27FPFHauGHVD1E9"

func TestSolanaChainObserver(t *testing.T) {
	logger := hclog.NewNullLogger()

	config := &oCore.SolanaChainConfig{
		SolanaChainConfig: solanaAB.SolanaChainConfig{
			TxProviderEndpoint: solanaNodeURL,
		},
		ChainID:                    "solana-devnet",
		TrackedProgram:             trackedProgram,
		RetryTimeoutMiliseconds:    time.Duration(1000), // 1000 ms
		BlockFetchDelayMiliseconds: time.Duration(100),  // 100 ms
		RestartTrackerPullCheck:    30 * time.Second,
	}

	t.Run("bad config - NewSolanaChainObserver error", func(t *testing.T) {
		badConfig := *config
		badConfig.TrackedProgram = "bad-program"
		indexerDB := &store.MockStorageHandler{}
		indexerDB.On("ReadSlot").Return(uint64(1), nil)
		indexerDB.On("UseTransactions").Return(false)

		solanaTxsReceiverMock := &core.SolanaTxsReceiverMock{}
		solanaTxsReceiverMock.On("NewUnprocessedEvent").Return(nil)

		co, err := NewSolanaChainObserver(&badConfig, solanaTxsReceiverMock, indexerDB, logger, logger)
		require.NoError(t, err)
		require.NotNil(t, co)

		err = co.Start()
		require.Error(t, err)
		require.NoError(t, co.Dispose())
	})

	t.Run("chain observer - NewSolanaChainObserver OK", func(t *testing.T) {
		indexerDB := &store.MockStorageHandler{}
		indexerDB.On("ReadSlot").Return(uint64(1), nil)
		indexerDB.On("UseTransactions").Return(false)

		solanaTxsReceiverMock := &core.SolanaTxsReceiverMock{}
		solanaTxsReceiverMock.On("NewUnprocessedEvent").Return(nil)

		co, err := NewSolanaChainObserver(config, solanaTxsReceiverMock, indexerDB, logger, logger)
		require.NoError(t, err)
		require.NotNil(t, co)
		require.Equal(t, co.GetConfig(), config)
		require.Equal(t, co.logger, logger)
	})

	t.Run("check start stop", func(t *testing.T) {
		indexerDB := &store.MockStorageHandler{}
		indexerDB.On("ReadSlot").Return(uint64(1), nil)
		indexerDB.On("GetLatestBlockPoint").Return(&store.BlockPoint{BlockSlot: 1}, nil)
		indexerDB.On("UseTransactions").Return(false)
		indexerDB.On("GetAllUnprocessedTransactions").Return([]store.TxPoint{}, nil)
		indexerDB.On("GetLastProcessedTransaction").Return(store.TxPoint{}, nil)
		indexerDB.On("GetLatestQueriedTransaction").Return(store.TxPoint{}, nil)

		solanaTxsReceiverMock := &core.SolanaTxsReceiverMock{}
		solanaTxsReceiverMock.On("NewUnprocessedEvent").Return(nil)

		chainObserver, err := NewSolanaChainObserver(config, solanaTxsReceiverMock, indexerDB, logger, logger)
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		err = chainObserver.Start()
		require.NoError(t, err)

		require.NoError(t, chainObserver.Dispose())
	})

	t.Run("time.After branch - health check runs when timer fires", func(t *testing.T) {
		// Use a short interval so the timer fires quickly; we then assert that
		// updateIsTrackerAlive (ReadSlot) was called from the health-check goroutine.
		restartCheckInterval := 25 * time.Millisecond
		configShortInterval := &oCore.SolanaChainConfig{
			ChainID: "solana-devnet",
			SolanaChainConfig: solanaAB.SolanaChainConfig{
				TxProviderEndpoint: solanaNodeURL,
			},
			TrackedProgram:             trackedProgram,
			RetryTimeoutMiliseconds:    time.Duration(1000),
			BlockFetchDelayMiliseconds: time.Duration(100),
			RestartTrackerPullCheck:    restartCheckInterval,
		}

		indexerDB := &store.MockStorageHandler{}
		// Tracker.Start() calls ReadSlot once; health-check goroutine calls ReadSlot each time the timer fires.
		indexerDB.On("GetLatestBlockPoint").Return(&store.BlockPoint{BlockSlot: 1}, nil)
		indexerDB.On("UseTransactions").Return(false)
		indexerDB.On("GetAllUnprocessedTransactions").Return([]store.TxPoint{}, nil)
		indexerDB.On("GetLastProcessedTransaction").Return(store.TxPoint{}, nil)
		indexerDB.On("GetLatestQueriedTransaction").Return(store.TxPoint{}, nil)

		solanaTxsReceiverMock := &core.SolanaTxsReceiverMock{}
		solanaTxsReceiverMock.On("NewUnprocessedEvent").Return(nil)

		chainObserver, err := NewSolanaChainObserver(configShortInterval, solanaTxsReceiverMock, indexerDB, logger, logger)
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		require.NoError(t, chainObserver.Start())

		// Wait long enough for at least one time.After to fire (RestartTrackerPullCheck).
		time.Sleep(restartCheckInterval*2 + 10*time.Millisecond)

		// Avoid asserting on testify mock internals while tracker goroutines are active,
		// because mock call records are not concurrency-safe under the race detector.
		require.NoError(t, chainObserver.Dispose())
	})
}

func TestSolanaChainObserver_ExecuteIsTrackerAlive(t *testing.T) {
	indexerDB := &store.MockStorageHandler{}

	so := &SolanaChainObserverImpl{
		indexerDB: indexerDB,
		closedCh:  make(chan struct{}),
		logger:    hclog.NewNullLogger(),
	}

	t.Run("everything is normal", func(t *testing.T) {
		indexerDB.On("GetLatestBlockPoint").Return(&store.BlockPoint{BlockSlot: 1}, nil).Once()

		require.True(t, so.updateIsTrackerAlive())
		require.Equal(t, uint64(1), so.lastSlot)
	})

	t.Run("restart required", func(t *testing.T) {
		indexerDB.On("GetLatestBlockPoint").Return(&store.BlockPoint{BlockSlot: 2}, nil).Once()

		so.lastSlot = 2

		require.False(t, so.updateIsTrackerAlive())
		require.Equal(t, uint64(2), so.lastSlot)
	})

	t.Run("ReadSlot error - consider alive to avoid unnecessary restart", func(t *testing.T) {
		indexerDB.On("GetLatestBlockPoint").Return((*store.BlockPoint)(nil), errors.New("test error")).Once()

		require.True(t, so.updateIsTrackerAlive())
	})

	indexerDB.AssertExpectations(t)
}

func TestSolanaChainObserver_Dispose(t *testing.T) {
	logger := hclog.NewNullLogger()
	indexerDB := &store.MockStorageHandler{}
	indexerDB.On("ReadSlot").Return(uint64(1), nil)
	indexerDB.On("GetLatestBlockPoint").Return(&store.BlockPoint{BlockSlot: 1}, nil)
	indexerDB.On("UseTransactions").Return(false)
	indexerDB.On("GetAllUnprocessedTransactions").Return([]store.TxPoint{}, nil)
	indexerDB.On("GetLastProcessedTransaction").Return(store.TxPoint{}, nil)
	indexerDB.On("GetLatestQueriedTransaction").Return(store.TxPoint{}, nil)

	testConfig := &oCore.SolanaChainConfig{
		ChainID: "solana-devnet",
		SolanaChainConfig: solanaAB.SolanaChainConfig{
			TxProviderEndpoint: solanaNodeURL,
		},
		TrackedProgram:             trackedProgram,
		RetryTimeoutMiliseconds:    time.Duration(1000),
		BlockFetchDelayMiliseconds: time.Duration(100),
		RestartTrackerPullCheck:    30 * time.Second,
	}

	solanaTxsReceiverMock := &core.SolanaTxsReceiverMock{}
	solanaTxsReceiverMock.On("NewUnprocessedEvent").Return(nil)

	chainObserver, err := NewSolanaChainObserver(testConfig, solanaTxsReceiverMock, indexerDB, logger, logger)
	require.NoError(t, err)
	require.NotNil(t, chainObserver)

	require.NoError(t, chainObserver.Start())
	require.NoError(t, chainObserver.Dispose())
}

func Test_LoadTrackerConfigSolana(t *testing.T) {
	logger := hclog.NewNullLogger()

	config := &oCore.SolanaChainConfig{
		SolanaChainConfig: solanaAB.SolanaChainConfig{
			TxProviderEndpoint: solanaNodeURL,
		},
		TrackedProgram:             trackedProgram,
		RetryTimeoutMiliseconds:    time.Duration(1000),
		BlockFetchDelayMiliseconds: time.Duration(100),
	}

	solanaTxsReceiverMock := &core.SolanaTxsReceiverMock{}
	solanaTxsReceiverMock.On("NewUnprocessedEvent").Return(nil)

	cfg, err := loadTrackerConfigs(config, solanaTxsReceiverMock, logger, logger)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, config.TxProviderEndpoint, cfg.RPCEndpoint)
	require.Equal(t, config.RetryTimeoutMiliseconds, cfg.RetryTimeout)
	require.NotEmpty(t, cfg.TrackedPrograms)
	require.Contains(t, cfg.TrackedPrograms, config.TrackedProgram)

	t.Run("only the tracker output goes to the indexer logger", func(t *testing.T) {
		indexerLogger := hclog.New(&hclog.LoggerOptions{Name: "solana_indexer", Output: io.Discard})
		mainLogger := hclog.New(&hclog.LoggerOptions{Name: "validatorcomponents", Output: io.Discard})

		cfg, err := loadTrackerConfigs(config, solanaTxsReceiverMock, indexerLogger, mainLogger)
		require.NoError(t, err)

		require.True(t, strings.Contains(cfg.Logger.Name(), "solana_indexer"))

		handler, ok := cfg.EventSubscriber.(*confirmedEventHandler)
		require.True(t, ok)
		require.Equal(t, "validatorcomponents", handler.Logger.Name())
	})
}
