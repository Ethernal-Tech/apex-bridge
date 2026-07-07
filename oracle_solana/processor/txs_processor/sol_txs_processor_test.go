package processor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	oDatabaseaccess "github.com/Ethernal-Tech/apex-bridge/oracle_common/database_access"
	txsprocessor "github.com/Ethernal-Tech/apex-bridge/oracle_common/processor/txs_processor"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	databaseaccess "github.com/Ethernal-Tech/apex-bridge/oracle_solana/database_access"
	"github.com/Ethernal-Tech/solana-infrastructure/tracker"
	solanaTxsStore "github.com/Ethernal-Tech/solana-infrastructure/tracker/store"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newSolTxsProcessor(
	ctx context.Context,
	appConfig *oCore.AppConfig,
	db core.SolanaTxsProcessorDB,
	successTxProcessors []core.SolanaTxSuccessProcessor,
	failedTxProcessors []core.SolanaTxFailedProcessor,
	bridgeDataFetcher core.SolanaBridgeDataFetcher,
	bridgeSubmitter oCore.BridgeClaimsSubmitter,
	indexerDbs map[string]solanaTxsStore.StorageHandler,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
) (*txsprocessor.TxsProcessorImpl, *SolEventReceiverImpl) {
	txProcessors := NewTxProcessorsCollection(successTxProcessors, failedTxProcessors)

	solTxsReceiver := NewSolTxsReceiver(appConfig, db, txProcessors, bridgingRequestStateUpdater, hclog.NewNullLogger())

	solStateProcessor := NewSolStateProcessor(
		ctx, appConfig, db, txProcessors, indexerDbs, hclog.NewNullLogger(),
	)

	solTxsProcessor := txsprocessor.NewTxsProcessorImpl(
		ctx, appConfig, solStateProcessor, bridgeDataFetcher, bridgeSubmitter, bridgingRequestStateUpdater,
		hclog.NewNullLogger(),
	)

	return solTxsProcessor, solTxsReceiver
}

func newValidSolProcessor(
	ctx context.Context,
	appConfig *oCore.AppConfig,
	oracleDB core.Database,
	successTxProcessor core.SolanaTxSuccessProcessor,
	failedTxProcessor core.SolanaTxFailedProcessor,
	bridgeDataFetcher core.SolanaBridgeDataFetcher,
	bridgeSubmitter oCore.BridgeClaimsSubmitter,
	indexerDbs map[string]solanaTxsStore.StorageHandler,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
) (*txsprocessor.TxsProcessorImpl, *SolEventReceiverImpl) {
	var successTxProcessors []core.SolanaTxSuccessProcessor
	if successTxProcessor != nil {
		successTxProcessors = append(successTxProcessors, successTxProcessor)
	}

	var failedTxProcessors []core.SolanaTxFailedProcessor
	if failedTxProcessor != nil {
		failedTxProcessors = append(failedTxProcessors, failedTxProcessor)
	}

	return newSolTxsProcessor(
		ctx, appConfig, oracleDB, successTxProcessors, failedTxProcessors,
		bridgeDataFetcher, bridgeSubmitter, indexerDbs, bridgingRequestStateUpdater)
}

func TestSolTxsProcessor(t *testing.T) {
	appConfig := &oCore.AppConfig{
		SolanaChains: map[string]*oCore.SolanaChainConfig{
			common.ChainIDStrSolana: {},
		},
		BridgingSettings: oCore.BridgingSettings{
			MaxBridgingClaimsToGroup: 10,
		},
		RetryUnprocessedSettings: oCore.RetryUnprocessedSettings{
			BaseTimeout: time.Second * 60,
			MaxTimeout:  time.Second * 60,
		},
		ChainIDConverter: common.NewTestChainIDConverter(),
	}

	appConfig.FillOut()

	testDir, err := os.MkdirTemp("", "sol-txs-proc-test")
	require.NoError(t, err)

	defer os.RemoveAll(testDir)

	dbFilePath := filepath.Join(testDir, "temp_test_oracle.db")

	const processingWaitTimeMs = 300

	dbCleanup := func() {
		if _, err := os.Stat(dbFilePath); err == nil {
			os.Remove(dbFilePath)
		}
	}

	createOracleDB := func(filePath string) (*databaseaccess.BBoltDatabase, error) {
		boltDB, err := oDatabaseaccess.NewDatabase(filePath, appConfig)
		if err != nil {
			return nil, err
		}

		typeRegister := common.NewTypeRegister()
		typeRegister.SetType(common.ChainIDStrSolana, reflect.TypeOf(core.SolanaTx{}))

		oracleDB := &databaseaccess.BBoltDatabase{}
		oracleDB.Init(boltDB, appConfig, typeRegister)

		return oracleDB, nil
	}

	t.Run("NewSolTxsProcessor", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		proc, rec := newSolTxsProcessor(
			context.Background(), appConfig, nil, nil, nil, nil, nil, nil, nil)
		require.NotNil(t, proc)
		require.NotNil(t, rec)

		indexerMock := &solanaTxsStore.MockStorageHandler{}
		indexerMock.On("ReadSlot").Return(uint64(0), nil)
		indexerDbs := map[string]solanaTxsStore.StorageHandler{common.ChainIDStrSolana: indexerMock}

		proc, rec = newSolTxsProcessor(
			context.Background(),
			appConfig,
			&core.SolanaTxsProcessorDBMock{},
			[]core.SolanaTxSuccessProcessor{},
			[]core.SolanaTxFailedProcessor{},
			&core.SolanaBridgeDataFetcherMock{},
			&core.SolanaBridgeSubmitterMock{},
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)
		require.NotNil(t, proc)
		require.NotNil(t, rec)
	})

	t.Run("NewUnprocessedEvent unregistered chain", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		indexerMock := &solanaTxsStore.MockStorageHandler{}
		indexerMock.On("ReadSlot").Return(uint64(0), nil)
		indexerDbs := map[string]solanaTxsStore.StorageHandler{common.ChainIDStrSolana: indexerMock}

		_, rec := newValidSolProcessor(
			context.Background(),
			appConfig, oracleDB,
			nil, nil, nil, nil,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		err = rec.NewUnprocessedEvent("unregistered", emptyEventNotification())
		require.Error(t, err)
		require.ErrorContains(t, err, "originChainID not registered")
	})

	t.Run("NewUnprocessedEvent empty event", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		indexerMock := &solanaTxsStore.MockStorageHandler{}
		indexerMock.On("ReadSlot").Return(uint64(0), nil)
		indexerDbs := map[string]solanaTxsStore.StorageHandler{common.ChainIDStrSolana: indexerMock}

		_, rec := newValidSolProcessor(
			context.Background(),
			appConfig, oracleDB,
			nil, nil, nil, nil,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		err = rec.NewUnprocessedEvent(common.ChainIDStrSolana, emptyEventNotification())
		require.NoError(t, err)
	})

	t.Run("NewUnprocessedEvent unknown event name", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		indexerMock := &solanaTxsStore.MockStorageHandler{}
		indexerMock.On("ReadSlot").Return(uint64(0), nil)
		indexerDbs := map[string]solanaTxsStore.StorageHandler{common.ChainIDStrSolana: indexerMock}

		_, rec := newValidSolProcessor(
			context.Background(),
			appConfig, oracleDB,
			nil, nil, nil, nil,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		err = rec.NewUnprocessedEvent(common.ChainIDStrSolana, unknownEventNotification())
		require.Error(t, err)
		require.ErrorContains(t, err, "unknown event name")
	})

	t.Run("Start with unprocessed tx validation error", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		validTxProc := &core.SolanaTxSuccessProcessorMock{ShouldAddClaim: true, Type: common.BridgingTxTypeBridgingRequest}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("test validation err"))

		bridgeDataFetcher := &core.SolanaBridgeDataFetcherMock{}
		bridgeDataFetcher.On("GetBatchTransactions", mock.Anything, mock.Anything).Return([]eth.TxDataInfo{}, nil)

		bridgeSubmitter := &core.SolanaBridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(&types.Receipt{}, nil)

		indexerMock := &solanaTxsStore.MockStorageHandler{}
		indexerMock.On("ReadSlot").Return(uint64(1000), nil)
		indexerDbs := map[string]solanaTxsStore.StorageHandler{common.ChainIDStrSolana: indexerMock}

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, _ := newValidSolProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, nil, bridgeDataFetcher, bridgeSubmitter,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)
		require.NotNil(t, proc)

		metadata, err := core.MarshalSolMetadata(core.BaseSolMetadata{
			BridgingTxType: common.BridgingTxTypeBridgingRequest,
		})
		require.NoError(t, err)

		unprocessedTx := &core.SolanaTx{
			OriginChainID: common.ChainIDStrSolana,
			Priority:      0,
			SlotNumber:    100,
			Metadata:      metadata,
		}

		err = oracleDB.AddTxs(nil, []*core.SolanaTx{unprocessedTx})
		require.NoError(t, err)

		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(common.ChainIDStrSolana, 0)
		require.Empty(t, unprocessedTxs)

		processedTx, _ := oracleDB.GetProcessedTx(oCore.DBTxID{
			ChainID: common.ChainIDStrSolana,
			DBKey:   unprocessedTx.GetTxHash(),
		})
		require.NotNil(t, processedTx)
		require.True(t, processedTx.IsInvalid)
	})

	t.Run("Start with unprocessed txs valid", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		validTxProc := &core.SolanaTxSuccessProcessorMock{
			ShouldAddClaim: true,
			Type:           common.BridgingTxTypeBridgingRequest,
		}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		bridgeDataFetcher := &core.SolanaBridgeDataFetcherMock{}
		bridgeDataFetcher.On("GetBatchTransactions", mock.Anything, mock.Anything).Return([]eth.TxDataInfo{}, nil)

		bridgeSubmitter := &core.SolanaBridgeSubmitterMock{}
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(&types.Receipt{}, nil)

		indexerMock := &solanaTxsStore.MockStorageHandler{}
		indexerMock.On("ReadSlot").Return(uint64(1000), nil)
		indexerDbs := map[string]solanaTxsStore.StorageHandler{common.ChainIDStrSolana: indexerMock}

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, _ := newValidSolProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, nil, bridgeDataFetcher, bridgeSubmitter,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)
		require.NotNil(t, proc)

		metadata, err := core.MarshalSolMetadata(core.BaseSolMetadata{
			BridgingTxType: common.BridgingTxTypeBridgingRequest,
		})
		require.NoError(t, err)

		unprocessedTx := &core.SolanaTx{
			OriginChainID: common.ChainIDStrSolana,
			Priority:      0,
			SlotNumber:    100,
			Metadata:      metadata,
		}

		err = oracleDB.AddTxs(nil, []*core.SolanaTx{unprocessedTx})
		require.NoError(t, err)

		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(common.ChainIDStrSolana, 0)
		require.Empty(t, unprocessedTxs)
	})

	t.Run("Start with submit claims failed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		validTxProc := &core.SolanaTxSuccessProcessorMock{
			ShouldAddClaim: true,
			Type:           common.BridgingTxTypeBridgingRequest,
		}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		bridgeDataFetcher := &core.SolanaBridgeDataFetcherMock{}
		bridgeDataFetcher.On("GetBatchTransactions", mock.Anything, mock.Anything).Return([]eth.TxDataInfo{}, nil)

		bridgeSubmitter := &core.SolanaBridgeSubmitterMock{}
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("submit failed"))

		indexerMock := &solanaTxsStore.MockStorageHandler{}
		indexerMock.On("ReadSlot").Return(uint64(1000), nil)
		indexerDbs := map[string]solanaTxsStore.StorageHandler{common.ChainIDStrSolana: indexerMock}

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, _ := newValidSolProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, nil, bridgeDataFetcher, bridgeSubmitter,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)
		require.NotNil(t, proc)

		metadata, err := core.MarshalSolMetadata(core.BaseSolMetadata{
			BridgingTxType: common.BridgingTxTypeBridgingRequest,
		})
		require.NoError(t, err)

		unprocessedTx := &core.SolanaTx{
			OriginChainID: common.ChainIDStrSolana,
			Priority:      0,
			SlotNumber:    100,
			Metadata:      metadata,
		}

		err = oracleDB.AddTxs(nil, []*core.SolanaTx{unprocessedTx})
		require.NoError(t, err)

		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(common.ChainIDStrSolana, 0)
		require.Len(t, unprocessedTxs, 1)
	})

	t.Run("Start with expected txs - TTL not expired", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		failedTxProc := &core.SolanaTxFailedProcessorMock{
			ShouldAddClaim: false,
			Type:           common.BridgingTxTypeBatchExecution,
		}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		bridgeDataFetcher := &core.SolanaBridgeDataFetcherMock{}
		bridgeDataFetcher.On("GetBatchTransactions", mock.Anything, mock.Anything).Return([]eth.TxDataInfo{}, nil)

		bridgeSubmitter := &core.SolanaBridgeSubmitterMock{}
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(&types.Receipt{}, nil)

		indexerMock := &solanaTxsStore.MockStorageHandler{}
		indexerMock.On("GetLatestFinalizedBlockNumber").Return(uint64(100), nil)
		indexerDbs := map[string]solanaTxsStore.StorageHandler{common.ChainIDStrSolana: indexerMock}

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, _ := newValidSolProcessor(
			ctx,
			appConfig, oracleDB,
			nil, failedTxProc, bridgeDataFetcher, bridgeSubmitter,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)
		require.NotNil(t, proc)

		metadata, err := core.MarshalSolMetadata(core.BaseSolMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
		})
		require.NoError(t, err)

		expectedTx := &core.BridgeExpectedSolanaTx{
			ChainID:  common.ChainIDStrSolana,
			Metadata: metadata,
			TTL:      500,
			Priority: 0,
		}

		err = oracleDB.AddExpectedTxs([]*core.BridgeExpectedSolanaTx{expectedTx})
		require.NoError(t, err)

		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		expectedTxs, _ := oracleDB.GetAllExpectedTxs(common.ChainIDStrSolana, 0)
		require.Len(t, expectedTxs, 1)
	})

	t.Run("Start with expected txs - expired", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		failedTxProc := &core.SolanaTxFailedProcessorMock{
			ShouldAddClaim: false,
			Type:           common.BridgingTxTypeBatchExecution,
		}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		bridgeDataFetcher := &core.SolanaBridgeDataFetcherMock{}
		bridgeDataFetcher.On("GetBatchTransactions", mock.Anything, mock.Anything).Return([]eth.TxDataInfo{}, nil)

		bridgeSubmitter := &core.SolanaBridgeSubmitterMock{}
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(&types.Receipt{}, nil)

		indexerMock := &solanaTxsStore.MockStorageHandler{}
		indexerMock.On("GetLatestFinalizedBlockNumber").Return(uint64(1000), nil)
		indexerDbs := map[string]solanaTxsStore.StorageHandler{common.ChainIDStrSolana: indexerMock}

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, _ := newValidSolProcessor(
			ctx,
			appConfig, oracleDB,
			nil, failedTxProc, bridgeDataFetcher, bridgeSubmitter,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)
		require.NotNil(t, proc)

		metadata, err := core.MarshalSolMetadata(core.BaseSolMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
		})
		require.NoError(t, err)

		expectedTx := &core.BridgeExpectedSolanaTx{
			ChainID:  common.ChainIDStrSolana,
			Metadata: metadata,
			TTL:      10,
			Priority: 0,
		}

		err = oracleDB.AddExpectedTxs([]*core.BridgeExpectedSolanaTx{expectedTx})
		require.NoError(t, err)

		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		expectedTxs, _ := oracleDB.GetAllExpectedTxs(common.ChainIDStrSolana, 0)
		require.Empty(t, expectedTxs)
	})
}

func emptyEventNotification() tracker.EventNotification { //nolint:unused
	return tracker.EventNotification{}
}

func unknownEventNotification() tracker.EventNotification { //nolint:unused
	return tracker.EventNotification{
		EventName: "UnknownEvent",
		EventData: "data",
	}
}
