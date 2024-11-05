package processor

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	oDatabaseaccess "github.com/Ethernal-Tech/apex-bridge/oracle_common/database_access"
	ethcore "github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	"github.com/Ethernal-Tech/ethgo"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethereum_common "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	txsprocessor "github.com/Ethernal-Tech/apex-bridge/oracle_common/processor/txs_processor"
	databaseaccess "github.com/Ethernal-Tech/apex-bridge/oracle_eth/database_access"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/go-hclog"
)

func newEthTxsProcessor(
	ctx context.Context,
	appConfig *oCore.AppConfig,
	db ethcore.EthTxsProcessorDB,
	successTxProcessors []ethcore.EthTxSuccessProcessor,
	failedTxProcessors []ethcore.EthTxFailedProcessor,
	bridgeSubmitter ethcore.BridgeSubmitter,
	indexerDbs map[string]eventTrackerStore.EventTrackerStore,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
) (*txsprocessor.TxsProcessorImpl, *EthTxsReceiverImpl) {
	txProcessors := NewTxProcessorsCollection(
		successTxProcessors, failedTxProcessors,
	)

	ethTxsReceiver := NewEthTxsReceiverImpl(appConfig, db, txProcessors, bridgingRequestStateUpdater, hclog.NewNullLogger())

	ethStateProcessor := NewEthStateProcessor(
		ctx, appConfig, db, txProcessors,
		indexerDbs, hclog.NewNullLogger(),
	)

	ethTxsProcessor := txsprocessor.NewTxsProcessorImpl(
		ctx, appConfig, ethStateProcessor, bridgeSubmitter, bridgingRequestStateUpdater,
		hclog.NewNullLogger(),
	)

	return ethTxsProcessor, ethTxsReceiver
}

func newValidProcessor(
	ctx context.Context,
	appConfig *oCore.AppConfig,
	oracleDB ethcore.Database,
	successTxProcessor ethcore.EthTxSuccessProcessor,
	failedTxProcessor ethcore.EthTxFailedProcessor,
	bridgeSubmitter ethcore.BridgeSubmitter,
	indexerDbs map[string]eventTrackerStore.EventTrackerStore,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
) (*txsprocessor.TxsProcessorImpl, *EthTxsReceiverImpl) {
	var successTxProcessors []ethcore.EthTxSuccessProcessor
	if successTxProcessor != nil {
		successTxProcessors = append(successTxProcessors, successTxProcessor)
	}

	var failedTxProcessors []ethcore.EthTxFailedProcessor
	if failedTxProcessor != nil {
		failedTxProcessors = append(failedTxProcessors, failedTxProcessor)
	}

	return newEthTxsProcessor(
		ctx, appConfig, oracleDB, successTxProcessors, failedTxProcessors,
		bridgeSubmitter, indexerDbs, bridgingRequestStateUpdater)
}

func TestEthTxsProcessor(t *testing.T) {
	appConfig := &oCore.AppConfig{
		EthChains: map[string]*oCore.EthChainConfig{
			common.ChainIDStrNexus: {},
		},
		BridgingSettings: oCore.BridgingSettings{
			MaxBridgingClaimsToGroup: 10,
		},
		RetryUnprocessedSettings: oCore.RetryUnprocessedSettings{
			BaseTimeout: time.Second * 60,
			MaxTimeout:  time.Second * 60,
		},
	}

	appConfig.FillOut()

	const (
		dbFilePath      = "temp_test_oracle.db"
		nexusDBFilePath = "temp_test_nexus.db"

		processingWaitTimeMs = 300
	)

	dbCleanup := func() {
		common.RemoveDirOrFilePathIfExists(dbFilePath)      //nolint:errcheck
		common.RemoveDirOrFilePathIfExists(nexusDBFilePath) //nolint:errcheck
	}

	createOracleDB := func(filePath string) (*databaseaccess.BBoltDatabase, error) {
		boltDB, err := oDatabaseaccess.NewDatabase(filePath, appConfig)
		if err != nil {
			return nil, err
		}

		typeRegister := oCore.NewTypeRegisterWithChains(appConfig, nil, reflect.TypeOf(ethcore.EthTx{}))

		oracleDB := &databaseaccess.BBoltDatabase{}
		oracleDB.Init(boltDB, appConfig, typeRegister)

		return oracleDB, nil
	}

	t.Cleanup(dbCleanup)

	t.Run("TestEthTxsProcessor", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		proc, rec := newEthTxsProcessor(context.Background(), appConfig, nil, nil, nil, nil, nil, nil)
		require.NotNil(t, proc)
		require.NotNil(t, rec)

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{common.ChainIDStrNexus: &ethcore.EventStoreMock{}}

		proc, rec = newEthTxsProcessor(
			context.Background(),
			appConfig,
			&ethcore.EthTxsProcessorDBMock{},
			[]ethcore.EthTxSuccessProcessor{},
			[]ethcore.EthTxFailedProcessor{},
			&ethcore.BridgeSubmitterMock{}, indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)
		require.NotNil(t, proc)
		require.NotNil(t, rec)
	})

	t.Run("NewUnprocessedTxs nil txs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		validTxProc := &ethcore.EthTxSuccessProcessorMock{}
		failedTxProc := &ethcore.EthTxFailedProcessorMock{}
		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{common.ChainIDStrNexus: &ethcore.EventStoreMock{}}

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		proc, rec := newValidProcessor(
			context.Background(),
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		require.NoError(t, rec.NewUnprocessedLog(common.ChainIDStrNexus, nil))

		unprocessedTxs, err := oracleDB.GetAllUnprocessedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.Nil(t, unprocessedTxs)
	})

	t.Run("NewUnprocessedTxs no txs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{common.ChainIDStrNexus: &ethcore.EventStoreMock{}}

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		proc, rec := newValidProcessor(
			context.Background(),
			appConfig, oracleDB,
			nil, nil, nil,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		require.NoError(t, rec.NewUnprocessedLog(common.ChainIDStrNexus, &ethgo.Log{}))

		unprocessedTxs, err := oracleDB.GetAllUnprocessedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.Nil(t, unprocessedTxs)
	})

	t.Run("NewUnprocessedTxs no relevant txs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{common.ChainIDStrNexus: &ethcore.EventStoreMock{}}

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxSuccessProcessorMock{Type: "relevant"}

		proc, rec := newValidProcessor(
			context.Background(),
			appConfig, oracleDB,
			validTxProc, nil, nil,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		require.NoError(t, rec.NewUnprocessedLog(common.ChainIDStrNexus, &ethgo.Log{
			BlockHash: ethgo.Hash{1},
		}))

		unprocessedTxs, err := oracleDB.GetAllUnprocessedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.Nil(t, unprocessedTxs)
	})

	t.Run("NewUnprocessedTxs valid txs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{common.ChainIDStrNexus: &ethcore.EventStoreMock{}}

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxSuccessProcessorMock{ShouldAddClaim: true, Type: "batch"}

		txHash := ethgo.Hash{1}

		proc, rec := newValidProcessor(
			context.Background(),
			appConfig, oracleDB,
			validTxProc, nil, nil,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		events, err := eth.GetNexusEventSignatures()
		require.NoError(t, err)

		depositEventSig := events[0]
		// withdrawEventSig := events[1]

		log := &ethgo.Log{
			BlockHash:       ethgo.Hash{1},
			TransactionHash: txHash,
			Data:            simulateRealData(),
			Topics: []ethgo.Hash{
				depositEventSig,
			},
		}

		require.NoError(t, rec.NewUnprocessedLog(common.ChainIDStrNexus, log))

		unprocessedTxs, err := oracleDB.GetAllUnprocessedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.Len(t, unprocessedTxs, 1)
		require.Equal(t, common.ChainIDStrNexus, unprocessedTxs[0].OriginChainID)
	})

	t.Run("NewUnprocessedTxs - tx validation err", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		const (
			originChainID = common.ChainIDStrNexus
		)

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{originChainID: &ethcore.EventStoreMock{}}

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxSuccessProcessorMock{ShouldAddClaim: true, Type: "batch"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("test err"))

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(&types.Receipt{}, nil)

		txHash := ethgo.Hash{1}

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, rec := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, nil, bridgeSubmitter,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		events, err := eth.GetNexusEventSignatures()
		require.NoError(t, err)

		depositEventSig := events[0]

		log := &ethgo.Log{
			BlockHash:       ethgo.Hash{1},
			TransactionHash: txHash,
			Data:            simulateRealData(),
			Topics: []ethgo.Hash{
				depositEventSig,
			},
		}

		require.NoError(t, rec.NewUnprocessedLog(common.ChainIDStrNexus, log))

		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(originChainID, 0)
		require.Nil(t, unprocessedTxs)

		processedTx, _ := oracleDB.GetProcessedTx(oCore.DBTxID{ChainID: originChainID, DBKey: txHash[:]})
		require.NotNil(t, processedTx)
		require.True(t, processedTx.IsInvalid)
	})

	t.Run("NewUnprocessedTxs - submit claims failed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		const (
			originChainID = common.ChainIDStrNexus
		)

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{originChainID: &ethcore.EventStoreMock{}}

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxSuccessProcessorMock{ShouldAddClaim: true, Type: "batch"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("test err"))

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("test err"))

		txHash := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910995")

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, rec := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, nil, bridgeSubmitter,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		events, err := eth.GetNexusEventSignatures()
		require.NoError(t, err)

		depositEventSig := events[0]

		log := &ethgo.Log{
			BlockHash:       ethgo.Hash{1},
			TransactionHash: txHash,
			Data:            simulateRealData(),
			Topics: []ethgo.Hash{
				depositEventSig,
			},
		}

		require.NoError(t, rec.NewUnprocessedLog(common.ChainIDStrNexus, log))

		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(originChainID, 0)
		require.Len(t, unprocessedTxs, 1)
		require.Equal(t, txHash, unprocessedTxs[0].Hash)
		require.Equal(t, originChainID, unprocessedTxs[0].OriginChainID)
		processedTx, _ := oracleDB.GetProcessedTx(oCore.DBTxID{ChainID: originChainID, DBKey: txHash[:]})
		require.Nil(t, processedTx)
	})

	t.Run("Start - unprocessedTxs - valid", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		const (
			originChainID = common.ChainIDStrNexus
		)

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{originChainID: &ethcore.EventStoreMock{}}

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxSuccessProcessorMock{ShouldAddClaim: true, Type: "batch"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(&types.Receipt{}, nil)

		txHash := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910995")

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, rec := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, nil, bridgeSubmitter,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		events, err := eth.GetNexusEventSignatures()
		require.NoError(t, err)

		depositEventSig := events[0]

		log := &ethgo.Log{
			BlockHash:       ethgo.Hash{1},
			TransactionHash: txHash,
			Data:            simulateRealData(),
			Topics: []ethgo.Hash{
				depositEventSig,
			},
		}

		require.NoError(t, rec.NewUnprocessedLog(common.ChainIDStrNexus, log))

		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(originChainID, 0)
		require.Nil(t, unprocessedTxs)

		processedTx, _ := oracleDB.GetProcessedTx(oCore.DBTxID{ChainID: originChainID, DBKey: txHash[:]})
		require.NotNil(t, processedTx)
		require.False(t, processedTx.IsInvalid)
	})

	t.Run("Start - expectedTxs - tx validation err", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		const (
			originChainID = common.ChainIDStrNexus
			ttl           = 2
		)

		store := &ethcore.EventStoreMock{}
		store.On("GetLastProcessedBlock").Return(uint64(6), nil)

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{originChainID: store}

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxSuccessProcessorMock{ShouldAddClaim: true, Type: "batch"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &ethcore.EthTxFailedProcessorMock{Type: "test"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("test err"))

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(&types.Receipt{}, nil)

		txHash := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910995")

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, rec := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		events, err := eth.GetNexusEventSignatures()
		require.NoError(t, err)

		depositEventSig := events[0]

		log := &ethgo.Log{
			BlockHash:       ethgo.Hash{1},
			TransactionHash: txHash,
			Data:            simulateRealData(),
			Topics: []ethgo.Hash{
				depositEventSig,
			},
		}

		require.NoError(t, rec.NewUnprocessedLog(originChainID, log))

		metadata, err := ethcore.MarshalEthMetadata(ethcore.BaseEthMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		err = oracleDB.AddExpectedTxs([]*ethcore.BridgeExpectedEthTx{
			{ChainID: originChainID, Hash: txHash, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		expectedTxs, _ := oracleDB.GetAllExpectedTxs(originChainID, 0)
		require.Nil(t, expectedTxs)
	})

	t.Run("Start - expectedTxs - submit claims failed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		const (
			originChainID = common.ChainIDStrNexus
			ttl           = 2
		)

		store := &ethcore.EventStoreMock{}
		store.On("GetLastProcessedBlock").Return(uint64(6), nil)

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{originChainID: store}

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxSuccessProcessorMock{ShouldAddClaim: true, Type: "batch"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &ethcore.EthTxFailedProcessorMock{Type: "test"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("test err"))

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("test err"))

		txHash := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910995")

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, rec := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		events, err := eth.GetNexusEventSignatures()
		require.NoError(t, err)

		depositEventSig := events[0]

		log := &ethgo.Log{
			BlockHash:       ethgo.Hash{1},
			TransactionHash: txHash,
			Data:            simulateRealData(),
			Topics: []ethgo.Hash{
				depositEventSig,
			},
		}

		require.NoError(t, rec.NewUnprocessedLog(originChainID, log))

		metadata, err := ethcore.MarshalEthMetadata(ethcore.BaseEthMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		err = oracleDB.AddExpectedTxs([]*ethcore.BridgeExpectedEthTx{
			{ChainID: originChainID, Hash: txHash, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		expectedTxs, _ := oracleDB.GetAllExpectedTxs(originChainID, 0)
		require.NotNil(t, expectedTxs)
	})

	t.Run("Start - expectedTxs - valid - tx not yet expired", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		const (
			originChainID = common.ChainIDStrNexus
			ttl           = 2
		)

		store := &ethcore.EventStoreMock{}
		store.On("GetLastProcessedBlock").Return(uint64(0), nil)

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{originChainID: store}

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxSuccessProcessorMock{ShouldAddClaim: true, Type: "batch"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &ethcore.EthTxFailedProcessorMock{ShouldAddClaim: true, Type: "test"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*oCore.BridgeClaims

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *oCore.BridgeClaims) (*types.Receipt, error) {
			submittedClaims = append(submittedClaims, claims)

			return &types.Receipt{}, nil
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return()

		txHash := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910995")

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, _ := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		metadata, err := ethcore.MarshalEthMetadata(ethcore.BaseEthMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		err = oracleDB.AddExpectedTxs([]*ethcore.BridgeExpectedEthTx{
			{ChainID: originChainID, Hash: txHash, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		expectedTxs, _ := oracleDB.GetAllExpectedTxs(originChainID, 0)
		require.NotNil(t, expectedTxs)
		require.Nil(t, submittedClaims)
	})

	t.Run("Start - expectedTxs - valid - expired tx", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		const (
			chainID = common.ChainIDStrNexus
			ttl     = 2
		)

		store := &ethcore.EventStoreMock{}
		store.On("GetLastProcessedBlock").Return(uint64(6), nil)

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{chainID: store}

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxSuccessProcessorMock{}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &ethcore.EthTxFailedProcessorMock{ShouldAddClaim: true, Type: "batch"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*oCore.BridgeClaims

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *oCore.BridgeClaims) (*types.Receipt, error) {
			submittedClaims = append(submittedClaims, claims)

			return &types.Receipt{}, nil
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return()
		bridgeSubmitter.On("GetBatchTransactions", "", uint64(0x1)).
			Return([]eth.TxDataInfo{}, error(nil))

		txHash := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910995")

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, _ := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		metadata, err := ethcore.MarshalEthMetadata(ethcore.BaseEthMetadata{BridgingTxType: "batch"})
		require.NoError(t, err)

		err = oracleDB.AddExpectedTxs([]*ethcore.BridgeExpectedEthTx{
			{ChainID: chainID, Hash: txHash, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		expectedTxs, _ := oracleDB.GetAllExpectedTxs(chainID, 0)
		require.Nil(t, expectedTxs)
		require.NotNil(t, submittedClaims)
		require.Len(t, submittedClaims, 1)
		require.Len(t, submittedClaims[0].BatchExecutionFailedClaims, 1)
	})

	t.Run("Start - unprocessedTxs, expectedTxs - single chain - valid 1", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		const (
			chainID   = common.ChainIDStrNexus
			ttl       = uint64(2)
			blockSlot = uint64(6)
		)

		store := &ethcore.EventStoreMock{}
		store.On("GetLastProcessedBlock").Return(uint64(6), nil).Once()

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{chainID: store}

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxSuccessProcessorMock{ShouldAddClaim: true, Type: "batch"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &ethcore.EthTxFailedProcessorMock{ShouldAddClaim: true, Type: "batch"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*oCore.BridgeClaims

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *oCore.BridgeClaims) (*types.Receipt, error) {
			submittedClaims = append(submittedClaims, claims)

			return &types.Receipt{}, nil
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return()
		bridgeSubmitter.On("GetBatchTransactions", "", uint64(0x1)).
			Return([]eth.TxDataInfo{}, error(nil))

		txHash := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910995")
		txHash2 := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910996")

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, rec := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		events, err := eth.GetNexusEventSignatures()
		require.NoError(t, err)

		depositEventSig := events[0]

		log := &ethgo.Log{
			BlockNumber:     blockSlot,
			BlockHash:       ethgo.Hash{1},
			TransactionHash: txHash,
			Data:            simulateRealData(),
			Topics: []ethgo.Hash{
				depositEventSig,
			},
		}

		require.NoError(t, rec.NewUnprocessedLog(chainID, log))

		metadata, err := ethcore.MarshalEthMetadata(ethcore.BaseEthMetadata{BridgingTxType: "batch"})
		require.NoError(t, err)

		err = oracleDB.AddExpectedTxs([]*ethcore.BridgeExpectedEthTx{
			{ChainID: chainID, Hash: txHash2, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		store.On("GetLastProcessedBlock").Return(blockSlot, nil)

		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(chainID, 0)
		require.Nil(t, unprocessedTxs)

		processedTx, _ := oracleDB.GetProcessedTx(oCore.DBTxID{ChainID: chainID, DBKey: txHash[:]})
		require.NotNil(t, processedTx)
		require.False(t, processedTx.IsInvalid)

		expectedTxs, _ := oracleDB.GetAllExpectedTxs(chainID, 0)
		require.Nil(t, expectedTxs)

		require.NotNil(t, submittedClaims)
		require.Len(t, submittedClaims, 1)
		require.Len(t, submittedClaims[0].BridgingRequestClaims, 1)
		require.Len(t, submittedClaims[0].BatchExecutionFailedClaims, 1)
	})

	t.Run("Start - unprocessedTxs, expectedTxs - single chain - valid 3", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		const (
			chainID   = common.ChainIDStrNexus
			ttl       = uint64(2)
			blockSlot = uint64(6)
		)

		store := &ethcore.EventStoreMock{}
		store.On("GetLastProcessedBlock").Return(uint64(6), nil)

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{chainID: store}

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxSuccessProcessorMock{ShouldAddClaim: true, Type: "batch"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &ethcore.EthTxFailedProcessorMock{ShouldAddClaim: true, Type: "batch"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*oCore.BridgeClaims

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *oCore.BridgeClaims) (*types.Receipt, error) {
			submittedClaims = append(submittedClaims, claims)

			return &types.Receipt{}, nil
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return()
		bridgeSubmitter.On("GetBatchTransactions", "", uint64(0x1)).
			Return([]eth.TxDataInfo{}, error(nil))

		txHash := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910995")
		txHash2 := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910996")

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, rec := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		events, err := eth.GetNexusEventSignatures()
		require.NoError(t, err)

		depositEventSig := events[0]

		log := &ethgo.Log{
			BlockNumber:     uint64(5),
			BlockHash:       ethgo.Hash{1},
			TransactionHash: txHash,
			Data:            simulateRealData(),
			Topics: []ethgo.Hash{
				depositEventSig,
			},
		}

		require.NoError(t, rec.NewUnprocessedLog(chainID, log))

		metadata, err := ethcore.MarshalEthMetadata(ethcore.BaseEthMetadata{BridgingTxType: "batch"})
		require.NoError(t, err)

		err = oracleDB.AddExpectedTxs([]*ethcore.BridgeExpectedEthTx{
			{ChainID: chainID, Hash: txHash2, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		store.On("GetLastProcessedBlock").Return(blockSlot, nil)

		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(chainID, 0)
		require.Nil(t, unprocessedTxs)

		processedTx, _ := oracleDB.GetProcessedTx(oCore.DBTxID{ChainID: chainID, DBKey: txHash[:]})
		require.NotNil(t, processedTx)
		require.False(t, processedTx.IsInvalid)

		expectedTxs, _ := oracleDB.GetAllExpectedTxs(chainID, 0)
		require.Nil(t, expectedTxs)

		require.NotNil(t, submittedClaims)
		require.Len(t, submittedClaims, 1)
		require.Len(t, submittedClaims[0].BridgingRequestClaims, 1)
		require.Len(t, submittedClaims[0].BatchExecutionFailedClaims, 1)
	})

	t.Run("Start - unprocessedTxs, expectedTxs - single chain - valid 4", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		const (
			chainID   = common.ChainIDStrNexus
			ttl       = uint64(2)
			blockSlot = uint64(6)
		)

		store := &ethcore.EventStoreMock{}
		store.On("GetLastProcessedBlock").Return(uint64(6), nil)

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{chainID: store}

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxSuccessProcessorMock{ShouldAddClaim: true, Type: common.BridgingTxTypeBatchExecution}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &ethcore.EthTxFailedProcessorMock{ShouldAddClaim: true, Type: common.BridgingTxTypeBatchExecution}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*oCore.BridgeClaims

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *oCore.BridgeClaims) (*types.Receipt, error) {
			submittedClaims = append(submittedClaims, claims)

			return &types.Receipt{}, nil
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return()
		bridgeSubmitter.On("GetBatchTransactions", "", uint64(0x1)).
			Return([]eth.TxDataInfo{}, error(nil))

		txHash := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910995")
		txHash2 := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910996")

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, rec := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		events, err := eth.GetNexusEventSignatures()
		require.NoError(t, err)

		depositEventSig := events[0]

		log := &ethgo.Log{
			BlockNumber:     blockSlot - 1,
			BlockHash:       ethgo.Hash{1},
			TransactionHash: txHash,
			Data:            simulateRealData(),
			Topics: []ethgo.Hash{
				depositEventSig,
			},
		}

		require.NoError(t, rec.NewUnprocessedLog(chainID, log))

		metadata, err := ethcore.MarshalEthMetadata(ethcore.BaseEthMetadata{BridgingTxType: common.BridgingTxTypeBatchExecution})
		require.NoError(t, err)

		err = oracleDB.AddExpectedTxs([]*ethcore.BridgeExpectedEthTx{
			{ChainID: chainID, Hash: txHash, TTL: uint64(8), Metadata: metadata},
		})
		require.NoError(t, err)

		err = oracleDB.AddExpectedTxs([]*ethcore.BridgeExpectedEthTx{
			{ChainID: chainID, Hash: txHash2, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		store.On("GetLastProcessedBlock").Return(blockSlot, nil)

		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(chainID, 0)
		require.Nil(t, unprocessedTxs)

		processedTx, _ := oracleDB.GetProcessedTx(oCore.DBTxID{ChainID: chainID, DBKey: txHash[:]})
		require.NotNil(t, processedTx)
		require.False(t, processedTx.IsInvalid)

		expectedTxs, _ := oracleDB.GetAllExpectedTxs(chainID, 0)
		require.Nil(t, expectedTxs)

		require.NotNil(t, submittedClaims)
		require.Len(t, submittedClaims, 1)
		require.Len(t, submittedClaims[0].BridgingRequestClaims, 1)
		require.Len(t, submittedClaims[0].BatchExecutionFailedClaims, 1)
	})

	t.Run("verify abi pack for withdraw", func(t *testing.T) {
		events, err := eth.GetNexusEventSignatures()
		require.NoError(t, err)

		withdrawEventSig := events[1]
		abi, err := contractbinding.GatewayMetaData.GetAbi()

		require.NoError(t, err)
		eventAbi, err := abi.EventByID(ethereum_common.Hash(withdrawEventSig))
		require.NoError(t, err)

		receiptData, err := eventAbi.Inputs.Pack(
			common.ChainIDIntPrime, ethereum_common.Address{}, []ReceiverWithdraw{{
				Receiver: "123",
				Amount:   big.NewInt(1),
			}},
			big.NewInt(1), big.NewInt(1),
		)
		require.NoError(t, err)

		gethLog := types.Log{
			Data:   receiptData,
			Topics: []ethereum_common.Hash{ethereum_common.Hash(withdrawEventSig)},
		}

		contract, err := contractbinding.NewGateway(ethereum_common.Address{}, nil)
		require.NoError(t, err)

		event, err := contract.ParseWithdraw(gethLog)
		require.NoError(t, err)
		require.NotNil(t, event)
	})

	t.Run("Start - unprocessedTxs - valid brc goes to pending", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		const (
			originChainID = common.ChainIDStrNexus
		)

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{originChainID: &ethcore.EventStoreMock{}}

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxSuccessProcessorMock{ShouldAddClaim: true, Type: common.BridgingTxTypeBridgingRequest}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(&types.Receipt{}, nil)
		bridgeSubmitter.On("GetBatchTransactions", "", uint64(0x1)).
			Return([]eth.TxDataInfo{}, error(nil))

		txHash := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910995")

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, rec := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, nil, bridgeSubmitter,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		events, err := eth.GetNexusEventSignatures()
		require.NoError(t, err)

		withdrawEventSig := events[1]
		abi, err := contractbinding.GatewayMetaData.GetAbi()

		require.NoError(t, err)
		eventAbi, err := abi.EventByID(ethereum_common.Hash(withdrawEventSig))
		require.NoError(t, err)

		receiptData, err := eventAbi.Inputs.Pack(
			common.ChainIDIntPrime, ethereum_common.Address{}, []ReceiverWithdraw{{
				Receiver: "123",
				Amount:   big.NewInt(1),
			}},
			big.NewInt(1), big.NewInt(1),
		)
		require.NoError(t, err)

		log := &ethgo.Log{
			BlockHash:       ethgo.Hash{1},
			TransactionHash: txHash,
			Data:            receiptData,
			Topics:          []ethgo.Hash{withdrawEventSig},
		}

		require.NoError(t, rec.NewUnprocessedLog(common.ChainIDStrNexus, log))

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(originChainID, 0)
		require.NotNil(t, unprocessedTxs)
		require.Len(t, unprocessedTxs, 1)

		tx := unprocessedTxs[0]

		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		unprocessedTxs, _ = oracleDB.GetAllUnprocessedTxs(originChainID, 0)
		require.Nil(t, unprocessedTxs)

		processedTx, _ := oracleDB.GetProcessedTx(oCore.DBTxID{ChainID: originChainID, DBKey: txHash[:]})
		require.Nil(t, processedTx)

		pendingTx, _ := oracleDB.GetPendingTx(oCore.DBTxID{ChainID: tx.GetChainID(), DBKey: tx.GetTxHash()})
		require.NotNil(t, pendingTx)
		require.Equal(t, originChainID, pendingTx.GetChainID())
		require.Equal(t, tx.Hash[:], pendingTx.GetTxHash())
	})

	t.Run("Start - unprocessedTxs - valid brc rejected and retry", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		const (
			originChainID = common.ChainIDStrNexus
		)

		txHash := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910995")

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{originChainID: &ethcore.EventStoreMock{}}

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxSuccessProcessorMock{
			AddClaimCallback: func(claims *oCore.BridgeClaims) {
				claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, oCore.BridgingRequestClaim{
					ObservedTransactionHash: txHash,
					SourceChainId:           common.ToNumChainID(originChainID),
				})
			},
			Type: common.BridgingTxTypeBridgingRequest,
		}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		eventSigs, err := eth.GetSubmitClaimsEventSignatures()
		require.NoError(t, err)

		receiptData, err := notEnoughFundsEventArguments.Pack("BRC", big.NewInt(0), big.NewInt(0))
		require.NoError(t, err)

		receipt := &types.Receipt{
			Logs: []*types.Log{
				{
					Topics: []ethereum_common.Hash{ethereum_common.Hash(eventSigs[0])},
					Data:   receiptData,
				},
			},
		}

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(receipt, nil)
		bridgeSubmitter.On("GetBatchTransactions", "", uint64(0x1)).
			Return([]eth.TxDataInfo{}, error(nil))

		contract, err := contractbinding.NewBridgeContract(ethereum_common.Address{}, nil)
		require.NoError(t, err)

		event, err := contract.ParseNotEnoughFunds(*receipt.Logs[0])
		require.NoError(t, err)
		require.NotNil(t, event)

		events, err := eth.GetNexusEventSignatures()
		require.NoError(t, err)

		withdrawEventSig := events[1]
		abi, err := contractbinding.GatewayMetaData.GetAbi()

		require.NoError(t, err)
		eventAbi, err := abi.EventByID(ethereum_common.Hash(withdrawEventSig))
		require.NoError(t, err)

		withdrawReceiptData, err := eventAbi.Inputs.Pack(
			common.ChainIDIntPrime, ethereum_common.Address{}, []ReceiverWithdraw{{
				Receiver: "123",
				Amount:   big.NewInt(1),
			}},
			big.NewInt(1), big.NewInt(1),
		)
		require.NoError(t, err)

		log := &ethgo.Log{
			BlockHash:       ethgo.Hash{1},
			TransactionHash: txHash,
			Data:            withdrawReceiptData,
			Topics:          []ethgo.Hash{withdrawEventSig},
		}

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, rec := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, nil, bridgeSubmitter,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		require.NoError(t, rec.NewUnprocessedLog(common.ChainIDStrNexus, log))

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(originChainID, 0)
		require.NotNil(t, unprocessedTxs)
		require.Len(t, unprocessedTxs, 1)

		tx := unprocessedTxs[0]

		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		processedTx, _ := oracleDB.GetProcessedTx(oCore.DBTxID{ChainID: originChainID, DBKey: txHash[:]})
		require.Nil(t, processedTx)

		pendingTx, _ := oracleDB.GetPendingTx(oCore.DBTxID{ChainID: tx.GetChainID(), DBKey: tx.GetTxHash()})
		require.Nil(t, pendingTx)

		// rejected
		unprocessedTxs, _ = oracleDB.GetAllUnprocessedTxs(originChainID, 0)
		require.NotNil(t, unprocessedTxs)
		require.Len(t, unprocessedTxs, 1)
		require.Equal(t, originChainID, unprocessedTxs[0].OriginChainID)
		require.Equal(t, tx.Hash, unprocessedTxs[0].Hash)
		require.Equal(t, uint32(1), unprocessedTxs[0].TryCount)
		require.False(t, unprocessedTxs[0].LastTimeTried.IsZero())

		// reset ctx to run again, and confirm by TryCount that this tx was skipped because of LastTimeTried
		ctx, cancelFunc = context.WithCancel(context.Background())
		proc, _ = newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, nil, bridgeSubmitter,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		unprocessedTxs, _ = oracleDB.GetAllUnprocessedTxs(originChainID, 0)
		require.NotNil(t, unprocessedTxs)
		require.Len(t, unprocessedTxs, 1)
		require.Equal(t, originChainID, unprocessedTxs[0].OriginChainID)
		require.Equal(t, tx.Hash, unprocessedTxs[0].Hash)
		require.Equal(t, uint32(1), unprocessedTxs[0].TryCount)
		require.False(t, unprocessedTxs[0].LastTimeTried.IsZero())

		newTx := unprocessedTxs[0]
		// set LastTimeTried to simulate time passing
		newTx.LastTimeTried = newTx.LastTimeTried.Add(-time.Second * 60)

		err = oracleDB.UpdateTxs(&ethcore.EthUpdateTxsData{
			UpdateUnprocessed: []*ethcore.EthTx{newTx},
		})
		require.NoError(t, err)

		// reset ctx to run again, and confirm by TryCount that this tx was tried again because we simulated time passing
		ctx, cancelFunc = context.WithCancel(context.Background())
		proc, _ = newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, nil, bridgeSubmitter,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		unprocessedTxs, _ = oracleDB.GetAllUnprocessedTxs(originChainID, 0)
		require.NotNil(t, unprocessedTxs)
		require.Len(t, unprocessedTxs, 1)
		require.Equal(t, originChainID, unprocessedTxs[0].OriginChainID)
		require.Equal(t, tx.Hash, unprocessedTxs[0].Hash)
		require.Equal(t, uint32(2), unprocessedTxs[0].TryCount)
		require.False(t, unprocessedTxs[0].LastTimeTried.IsZero())
	})

	t.Run("Start - BatchExecutionInfoEvent", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		originChainID := common.ChainIDStrNexus

		metadata, err := ethcore.MarshalEthMetadata(ethcore.BridgingRequestEthMetadata{
			BridgingTxType: common.BridgingTxTypeBridgingRequest,
		})
		require.NoError(t, err)

		txHash1 := ethgo.HexToHash("0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f61")
		ethTx1 := &ethcore.EthTx{
			Hash: txHash1, OriginChainID: originChainID, Address: ethgo.Address{},
			Metadata: metadata,
		}

		txHash2 := ethgo.HexToHash("0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62")
		ethTx2 := &ethcore.EthTx{Hash: txHash2, OriginChainID: originChainID, Address: ethgo.Address{}}

		txHashBatch := ethgo.HexToHash("0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f63")

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{originChainID: &ethcore.EventStoreMock{}}

		oracleDB, err := createOracleDB(dbFilePath)
		require.NoError(t, err)

		err = oracleDB.AddTxs([]*ethcore.ProcessedEthTx{}, []*ethcore.EthTx{ethTx1, ethTx2})
		require.NoError(t, err)

		err = oracleDB.UpdateTxs(&oCore.UpdateTxsData[*ethcore.EthTx, *ethcore.ProcessedEthTx, *ethcore.BridgeExpectedEthTx]{
			MoveUnprocessedToPending: []*ethcore.EthTx{ethTx1, ethTx2},
		})
		require.NoError(t, err)

		pendingTx1, _ := oracleDB.GetPendingTx(oCore.DBTxID{ChainID: ethTx1.GetChainID(), DBKey: ethTx1.GetTxHash()})
		require.NotNil(t, pendingTx1)

		pendingTx2, _ := oracleDB.GetPendingTx(oCore.DBTxID{ChainID: ethTx2.GetChainID(), DBKey: ethTx2.GetTxHash()})
		require.NotNil(t, pendingTx2)

		brcProc := &ethcore.EthTxSuccessProcessorMock{
			AddClaimCallback: func(claims *oCore.BridgeClaims) {
				claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, oCore.BridgingRequestClaim{
					ObservedTransactionHash: txHash1,
					SourceChainId:           common.ToNumChainID(originChainID),
				})
			},
			Type: common.BridgingTxTypeBridgingRequest,
		}
		brcProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		becProc := &ethcore.EthTxSuccessProcessorMock{
			AddClaimCallback: func(claims *oCore.BridgeClaims) {
				claims.BatchExecutedClaims = append(claims.BatchExecutedClaims, oCore.BatchExecutedClaim{
					ObservedTransactionHash: txHashBatch,
					BatchNonceId:            2,
					ChainId:                 common.ChainIDIntVector,
				})
			},
			Type: common.BridgingTxTypeBatchExecution,
		}
		becProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.OnSubmitClaims = func(claims *oCore.BridgeClaims) (*types.Receipt, error) {
			if len(claims.BatchExecutedClaims) == 0 && len(claims.BatchExecutionFailedClaims) == 0 {
				return &types.Receipt{}, nil
			}

			return &types.Receipt{}, nil
		}
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything, mock.Anything).Return()
		bridgeSubmitter.On("GetBatchTransactions", common.ChainIDStrVector, uint64(0x1)).
			Return([]eth.TxDataInfo{
				{
					SourceChainId:           common.ChainIDIntNexus,
					ObservedTransactionHash: txHash1,
				},
			}, error(nil))
		bridgeSubmitter.On("GetBatchTransactions", common.ChainIDStrVector, uint64(0x2)).
			Return([]eth.TxDataInfo{
				{
					SourceChainId:           common.ChainIDIntNexus,
					ObservedTransactionHash: txHash2,
				},
			}, error(nil))

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, rec := newEthTxsProcessor(
			ctx,
			appConfig, oracleDB,
			[]ethcore.EthTxSuccessProcessor{brcProc, becProc}, nil, bridgeSubmitter,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		events, err := eth.GetNexusEventSignatures()
		require.NoError(t, err)

		depositEventSig := events[0]

		data := simulateRealData()
		log := &ethgo.Log{
			BlockHash:       ethgo.Hash{1},
			TransactionHash: txHashBatch,
			Data:            data,
			Topics: []ethgo.Hash{
				depositEventSig,
			},
		}

		require.NoError(t, rec.NewUnprocessedLog(originChainID, log))

		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		pendingTx2, _ = oracleDB.GetPendingTx(oCore.DBTxID{ChainID: ethTx2.GetChainID(), DBKey: ethTx2.GetTxHash()})
		require.Nil(t, pendingTx2)

		pendingTx1, _ = oracleDB.GetPendingTx(oCore.DBTxID{ChainID: ethTx1.GetChainID(), DBKey: ethTx1.GetTxHash()})
		require.NotNil(t, pendingTx1)
		require.Equal(t, pendingTx1.GetTryCount(), uint32(1))

		unprocessedTxs, err := oracleDB.GetAllUnprocessedTxs(originChainID, 0)
		require.NoError(t, err)
		require.Len(t, unprocessedTxs, 0)

		processedTx1, err := oracleDB.GetProcessedTx(oCore.DBTxID{ChainID: originChainID, DBKey: ethTx1.Hash[:]})
		require.NoError(t, err)
		require.Nil(t, processedTx1)

		processedTx2, err := oracleDB.GetProcessedTx(oCore.DBTxID{ChainID: originChainID, DBKey: ethTx2.Hash[:]})
		require.NoError(t, err)
		require.NotNil(t, processedTx2)
		require.Equal(t, processedTx2.Hash, ethTx2.Hash)
	})
}

func simulateRealData() []byte {
	return []byte{
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 32,

		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 1, 0,

		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 32,

		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 1,

		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 90,

		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		15, 67, 252, 44, 4, 238, 0, 0,

		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 128,

		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 1,

		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 138, 7, 81, 200,
		52, 138, 167, 172, 216, 91, 182, 90,
		131, 25, 93, 99, 228, 141, 90, 141,

		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		13, 224, 182, 179, 167, 100, 0, 0,
	}
}

type ReceiverWithdraw struct {
	Receiver string   `json:"receiver" abi:"receiver"`
	Amount   *big.Int `json:"amount" abi:"amount"`
}

var (
	notEnoughFundsEventArguments = abi.Arguments{
		{Name: "claimeType", Type: abi.Type{T: abi.StringTy}},
		{Name: "index", Type: abi.Type{T: abi.UintTy, Size: 256}},
		{Name: "availableAmount", Type: abi.Type{T: abi.UintTy, Size: 256}},
	}
)
