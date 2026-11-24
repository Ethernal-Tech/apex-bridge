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
	failed "github.com/Ethernal-Tech/apex-bridge/oracle_solana/processor/tx_processors/failed"
	success "github.com/Ethernal-Tech/apex-bridge/oracle_solana/processor/tx_processors/success"
	"github.com/Ethernal-Tech/solana-infrastructure/tracker"
	solanaTxsStore "github.com/Ethernal-Tech/solana-infrastructure/tracker/store"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gagliardetto/solana-go"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
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

	for _, scenario := range []string{"full signature", "legacy prefix", "dangling link", "ambiguous prefix"} {
		t.Run("late expectation/"+scenario, func(t *testing.T) {
			oracleDB, err := createOracleDB(filepath.Join(t.TempDir(), "oracle.db"))
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, oracleDB.DB.Close()) })

			chainID := common.ChainIDStrSolana
			action := [32]byte{4, 5}
			sig := solana.Signature{0: 1, 63: 2}
			metadata := []byte(`{"t":"batch","n":1}`)
			event := &core.SolanaTx{
				OriginChainID: chainID, SlotNumber: 12, TxSignature: sig, InnerActionHash: action, Metadata: metadata,
			}
			expected := &core.BridgeExpectedSolanaTx{ChainID: chainID, Hash: action, TTL: 100, Metadata: metadata}
			logger := hclog.NewNullLogger()
			store := &solanaTxsStore.MockStorageHandler{}
			store.On("GetLatestFinalizedBlockNumber").Return(uint64(102), nil)

			newProcessor := func() *SolStateProcessor {
				processors := NewTxProcessorsCollection(
					[]core.SolanaTxSuccessProcessor{success.NewSolanaBatchExecutedProcessor(logger)},
					[]core.SolanaTxFailedProcessor{failed.NewSolanaBatchExecutionFailedProcessor(logger)},
				)
				processor := NewSolStateProcessor(t.Context(), appConfig, oracleDB, processors,
					map[string]solanaTxsStore.StorageHandler{chainID: store}, logger)
				processor.Reset()

				return processor
			}

			require.NoError(t, oracleDB.AddTxs(nil, []*core.SolanaTx{event}))

			processor := newProcessor()
			claims := &oCore.BridgeClaims{ChainIDConverter: appConfig.ChainIDConverter}
			processor.RunChecks(claims, chainID, 10, 0)
			require.Len(t, claims.BatchExecutedClaims, 1)

			// The expected record arrives after the success check, before the claim is persisted.
			require.NoError(t, oracleDB.AddExpectedTxs([]*core.BridgeExpectedSolanaTx{expected}))
			processor.PersistNew()
			require.NoError(t, oracleDB.DB.View(func(tx *bbolt.Tx) error {
				key, err := oDatabaseaccess.ReadInnerActionTxKey(tx, chainID, action[:])
				require.NoError(t, err)
				require.Equal(t, sig[:], key)

				return nil
			}))

			if scenario == "legacy prefix" || scenario == "ambiguous prefix" {
				// A second row shares the prefix. Only matching inner actions count.
				other := event.ToProcessedSolanaTx(false)

				other.TxSignature[63] = 1
				if scenario == "legacy prefix" {
					other.InnerActionHash[0]++
				}

				require.NoError(t, oracleDB.AddTxs([]*core.ProcessedSolanaTx{other}, nil))
				require.NoError(t, oracleDB.DB.Update(func(tx *bbolt.Tx) error {
					legacy := fmt.Sprintf(`{"chain_id":%q,"hash":"0x%x","ia_hash":"0x%x"}`, chainID, sig[:32], action)

					return tx.Bucket(oDatabaseaccess.ChainBucket(oDatabaseaccess.ProcessedTxsByInnerActionBucket, chainID)).
						Put(action[:], []byte(legacy))
				}))
			}

			if scenario == "dangling link" {
				require.NoError(t, oracleDB.DB.Update(func(tx *bbolt.Tx) error {
					return tx.Bucket(oDatabaseaccess.ChainBucket(oDatabaseaccess.ProcessedTxsBucket, chainID)).Delete(sig[:])
				}))
			}

			expiry := newProcessor()
			claims = &oCore.BridgeClaims{ChainIDConverter: appConfig.ChainIDConverter}
			expiry.RunChecks(claims, chainID, 10, 0)
			require.Empty(t, claims.BatchExecutionFailedClaims)
			expiry.PersistNew()

			queued, err := oracleDB.GetAllExpectedTxs(chainID, 0)
			require.NoError(t, err)

			if scenario == "dangling link" || scenario == "ambiguous prefix" {
				require.Len(t, queued, 1, "lookup errors must leave the expectation available for retry")

				return
			}

			require.Empty(t, queued)

			// A different TTL makes this a fresh expectation key for an already committed execution.
			expected.TTL++
			require.NoError(t, oracleDB.AddExpectedTxs([]*core.BridgeExpectedSolanaTx{expected}))
			queued, err = oracleDB.GetAllExpectedTxs(chainID, 0)
			require.NoError(t, err)
			require.Empty(t, queued)
		})
	}

	t.Run("absent and invalid executions can expire", func(t *testing.T) {
		oracleDB, err := createOracleDB(filepath.Join(t.TempDir(), "oracle.db"))
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, oracleDB.DB.Close()) })

		chainID := common.ChainIDStrSolana
		invalid := &core.ProcessedSolanaTx{
			OriginChainID: chainID, TxSignature: solana.Signature{63: 1}, IsInvalid: true, InnerActionHash: [32]byte{1},
		}
		require.NoError(t, oracleDB.UpdateTxs(&core.SolanaUpdateTxsData{
			MoveUnprocessedToProcessed: []*core.ProcessedSolanaTx{invalid},
		}, appConfig.ChainIDConverter))

		metadata := []byte(`{"t":"batch","n":1}`)
		require.NoError(t, oracleDB.AddExpectedTxs([]*core.BridgeExpectedSolanaTx{
			{ChainID: chainID, Hash: invalid.InnerActionHash, TTL: 100, Metadata: metadata},
			{ChainID: chainID, Hash: [32]byte{2}, TTL: 100, Metadata: metadata},
		}))

		store := &solanaTxsStore.MockStorageHandler{}
		store.On("GetLatestFinalizedBlockNumber").Return(uint64(102), nil)

		logger := hclog.NewNullLogger()
		processors := NewTxProcessorsCollection(nil, []core.SolanaTxFailedProcessor{
			failed.NewSolanaBatchExecutionFailedProcessor(logger),
		})
		processor := NewSolStateProcessor(t.Context(), appConfig, oracleDB, processors,
			map[string]solanaTxsStore.StorageHandler{chainID: store}, logger)
		processor.Reset()

		claims := &oCore.BridgeClaims{ChainIDConverter: appConfig.ChainIDConverter}
		processor.RunChecks(claims, chainID, 10, 0)
		require.Len(t, claims.BatchExecutionFailedClaims, 2)
	})

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
