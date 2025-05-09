package chain

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	indexerDb "github.com/Ethernal-Tech/cardano-infrastructure/indexer/db"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockSyncer struct {
	mock.Mock
	indexer.BlockSyncer
}

func (m *MockSyncer) Close() error {
	args := m.Called()

	return args.Error(0)
}

func (m *MockSyncer) ErrorCh() <-chan error {
	args := m.Called()

	return args.Get(0).(chan error)
}

type MockIndexerDB struct {
	mock.Mock
	indexer.Database
}

func (m *MockIndexerDB) Close() error {
	args := m.Called()

	return args.Error(0)
}

// Implement other required indexer.Database methods
func (m *MockIndexerDB) GetLatestBlockPoint() (*indexer.BlockPoint, error) {
	args := m.Called()

	return args.Get(0).(*indexer.BlockPoint), args.Error(1)
}

func TestCardanoChainObserver(t *testing.T) {
	testDir, err := os.MkdirTemp("", "boltdb-test")
	require.NoError(t, err)

	foldersCleanup := func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}

	defer foldersCleanup()

	settings := cCore.AppSettings{
		DbsPath: testDir,
		Logger: logger.LoggerConfig{
			LogFilePath: testDir,
		},
	}

	chainConfig := &cCore.CardanoChainConfig{
		ChainID:                common.ChainIDStrPrime,
		NetworkAddress:         "backbone.cardano.iog.io:3001",
		NetworkMagic:           764824073,
		StartBlockHash:         "335ac2d90bc37906c1264cfdc5769a652293cf64fa42b0c74d323473938b8ff1",
		StartSlot:              127933773,
		ConfirmationBlockCount: 10,
		OtherAddressesOfInterest: []string{
			"addr1v9kganeshgdqyhwnyn9stxxgl7r4y2ejfyqjn88n7ncapvs4sugsd",
		},
	}

	txsReceiverMock := &core.CardanoTxsReceiverMock{}
	txsReceiverMock.On("NewUnprocessedTxs", mock.Anything, mock.Anything).Return(error(nil))

	initDB := func(t *testing.T) indexer.Database {
		t.Helper()

		require.NoError(t, common.CreateDirectoryIfNotExists(settings.DbsPath, 0750))

		indexerDB, err := indexerDb.NewDatabaseInit("", filepath.Join(settings.DbsPath, chainConfig.ChainID+".db"))
		require.NoError(t, err)

		return indexerDB
	}

	t.Run("check ErrorCh", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		db := &core.CardanoTxsProcessorDBMock{}
		db.On("ClearAllTxs", mock.Anything).Return(error(nil))

		indexerDB := initDB(t)

		chainObserver, err := NewCardanoChainObserver(context.Background(), chainConfig, txsReceiverMock, db, indexerDB, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		defer chainObserver.Dispose() //nolint:errcheck

		errChan := chainObserver.ErrorCh()
		require.NotNil(t, errChan)
	})

	t.Run("check GetConfig", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		db := &core.CardanoTxsProcessorDBMock{}
		db.On("ClearAllTxs", mock.Anything).Return(error(nil))

		indexerDB := initDB(t)

		chainObserver, err := NewCardanoChainObserver(context.Background(), chainConfig, txsReceiverMock, db, indexerDB, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		config := chainObserver.GetConfig()
		require.NotNil(t, config)
		require.Equal(t, chainConfig, config)
	})

	t.Run("check start stop", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		db := &core.CardanoTxsProcessorDBMock{}
		db.On("ClearAllTxs", mock.Anything).Return(error(nil))

		indexerDB := initDB(t)

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		chainObserver, err := NewCardanoChainObserver(ctx, chainConfig, txsReceiverMock, db, indexerDB, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		err = chainObserver.Start()
		require.NoError(t, err)

		err = chainObserver.Dispose()
		require.NoError(t, err)
	})

	t.Run("check newConfirmedTxs called", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		db := &core.CardanoTxsProcessorDBMock{}
		db.On("ClearAllTxs", mock.Anything).Return(error(nil))

		indexerDB := initDB(t)

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		chainObserver, err := NewCardanoChainObserver(ctx, chainConfig, txsReceiverMock, db, indexerDB, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		doneCh := make(chan bool, 1)
		closed := uint32(0)

		txsReceiverMock.NewUnprocessedTxsFn = func(originChainId string, txs []*indexer.Tx) error {
			if atomic.CompareAndSwapUint32(&closed, 0, 1) {
				close(doneCh)
			}

			return nil
		}

		err = chainObserver.Start()
		require.NoError(t, err)

		timer := time.NewTimer(60 * time.Second)
		defer timer.Stop()

		select {
		case <-timer.C:
			t.Fatal("timeout")
		case <-doneCh:
		}
	})

	t.Run("errorCh fatal error", func(t *testing.T) {
		syncerMock := &MockSyncer{}
		errCh := make(chan error, 1)
		syncerMock.On("ErrorCh").Return(errCh)

		errCh <- indexer.ErrBlockIndexerFatal

		select {
		case err := <-syncerMock.ErrorCh():
			require.Error(t, err)
			require.ErrorIs(t, err, indexer.ErrBlockIndexerFatal)
		case <-time.After(time.Second):
			t.Fatal("Expected error not received from ErrorCh")
		}
	})

	t.Run("check close called", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		db := &core.CardanoTxsProcessorDBMock{}
		db.On("ClearAllTxs", mock.Anything).Return(error(nil))

		// setup syncer mocks
		syncerMock := &MockSyncer{}

		errorCh := make(chan error, 1)
		syncerMock.On("ErrorCh").Return(errorCh)
		syncerMock.On("Close").Return(nil)

		testErr := indexer.ErrBlockIndexerFatal

		// setup indexerDB mocks
		indexerDBMock := &MockIndexerDB{
			Database: initDB(t),
		}

		indexerDBMock.On("Close").Return(nil)
		indexerDBMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{}, nil)

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		chainObserver, err := NewCardanoChainObserver(ctx, chainConfig, txsReceiverMock, db, indexerDBMock, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, chainObserver)
		require.Same(t, indexerDBMock, chainObserver.indexerDB)

		err = chainObserver.Start()
		require.NoError(t, err)

		errorCh <- testErr

		select {
		case err := <-syncerMock.ErrorCh():
			require.Error(t, err)
			require.ErrorIs(t, err, indexer.ErrBlockIndexerFatal)
		case <-time.After(time.Second):
			t.Fatal("Expected error not received from ErrorCh")
		}
	})
}

func Test_initOracleState(t *testing.T) {
	dbMock := &indexer.DatabaseMock{}
	dbWriterMock := &indexer.DBTransactionWriterMock{}
	utxos := []cCore.CardanoChainConfigUtxo{
		{
			Hash:    indexer.NewHashFromHexString("0x1"),
			Index:   2,
			Address: "0xffaa",
			Amount:  uint64(200),
		},
		{
			Hash:    indexer.NewHashFromHexString("0x2"),
			Index:   1,
			Address: "0xff03",
			Amount:  uint64(500),
		},
	}
	blockSlot := uint64(100)
	oracleDBMock := &core.CardanoTxsProcessorDBMock{}
	blockHash := "0xA1"
	chainID := "prime"

	t.Run("error", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{}, errors.New("test error")).Once()

		require.Error(t, initOracleState(dbMock, oracleDBMock, blockHash, blockSlot, utxos, chainID, hclog.NewNullLogger()))
	})

	t.Run("empty hash", func(t *testing.T) {
		require.NoError(t, initOracleState(dbMock, oracleDBMock, "", blockSlot, utxos, chainID, hclog.NewNullLogger()))
	})

	t.Run("updated in db smaller slot", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return((*indexer.BlockPoint)(nil), error(nil)).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{
			BlockSlot: uint64(50),
		}, error(nil)).Once()

		dbMock.On("OpenTx").Return(dbWriterMock).Twice()
		dbWriterMock.On("DeleteAllTxOutputsPhysically").Return(dbMock).Twice()
		dbWriterMock.On("AddTxOutputs", convertUtxos(utxos)).Return(dbMock).Twice()
		dbWriterMock.On("SetLatestBlockPoint", &indexer.BlockPoint{
			BlockSlot: blockSlot,
			BlockHash: indexer.NewHashFromHexString(blockHash),
		}).Return(dbMock).Twice()
		dbWriterMock.On("Execute").Return(error(nil)).Twice()
		oracleDBMock.On("ClearAllTxs", chainID).Return(error(nil)).Twice()

		// empty
		require.NoError(t, initOracleState(dbMock, oracleDBMock, blockHash, blockSlot, utxos, chainID, hclog.NewNullLogger()))
		// smaller
		require.NoError(t, initOracleState(dbMock, oracleDBMock, blockHash, blockSlot, utxos, chainID, hclog.NewNullLogger()))
	})

	t.Run("skipping - has equal slot", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{
			BlockSlot: uint64(100),
		}, error(nil)).Once()

		require.NoError(t, initOracleState(dbMock, oracleDBMock, blockHash, blockSlot, utxos, chainID, hclog.NewNullLogger()))
	})

	t.Run("skipping - has greater slot", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{
			BlockSlot: uint64(150),
		}, error(nil)).Once()

		require.NoError(t, initOracleState(dbMock, oracleDBMock, blockHash, blockSlot, utxos, chainID, hclog.NewNullLogger()))
	})

	t.Run("ClearAllTxs error", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return((*indexer.BlockPoint)(nil), error(nil)).Once()
		oracleDBMock.On("ClearAllTxs", chainID).Return(errors.New("test error")).Once()

		require.Error(t, initOracleState(dbMock, oracleDBMock, blockHash, blockSlot, utxos, chainID, hclog.NewNullLogger()))
	})
}

func Test_convertUtxos(t *testing.T) {
	utxos := []cCore.CardanoChainConfigUtxo{
		{
			Hash:    indexer.NewHashFromHexString("0x2"),
			Index:   2,
			Address: "0xffaa",
			Amount:  uint64(200),
			Slot:    34,
		},
		{
			Hash:    indexer.NewHashFromHexString("0x1"),
			Index:   1,
			Address: "0xff03",
			Amount:  uint64(500),
			Slot:    196,
		},
	}

	require.Equal(t, []*indexer.TxInputOutput{
		{
			Input: indexer.TxInput{
				Hash:  indexer.NewHashFromHexString("0x2"),
				Index: 2,
			},
			Output: indexer.TxOutput{
				Address: "0xffaa",
				Amount:  200,
				Slot:    34,
			},
		},
		{
			Input: indexer.TxInput{
				Hash:  indexer.NewHashFromHexString("0x1"),
				Index: 1,
			},
			Output: indexer.TxOutput{
				Address: "0xff03",
				Amount:  500,
				Slot:    196,
			},
		},
	}, convertUtxos(utxos))
}
