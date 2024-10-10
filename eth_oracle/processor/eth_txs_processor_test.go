package processor

import (
	"context"
	"fmt"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethcore "github.com/Ethernal-Tech/apex-bridge/eth_oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/ethgo"

	databaseaccess "github.com/Ethernal-Tech/apex-bridge/eth_oracle/database_access"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/go-hclog"
)

func newValidProcessor(
	appConfig *core.AppConfig,
	oracleDB ethcore.EthTxsProcessorDB,
	txProcessor ethcore.EthTxProcessor,
	failedTxProcessor ethcore.EthTxFailedProcessor,
	bridgeSubmitter ethcore.BridgeSubmitter,
	indexerDbs map[string]eventTrackerStore.EventTrackerStore,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
) *EthTxsProcessorImpl {
	var txProcessors []ethcore.EthTxProcessor
	if txProcessor != nil {
		txProcessors = append(txProcessors, txProcessor)
	}

	var failedTxProcessors []ethcore.EthTxFailedProcessor
	if failedTxProcessor != nil {
		failedTxProcessors = append(failedTxProcessors, failedTxProcessor)
	}

	ethTxsProcessor := NewEthTxsProcessor(
		context.Background(),
		appConfig, oracleDB,
		txProcessors, failedTxProcessors,
		bridgeSubmitter, indexerDbs,
		bridgingRequestStateUpdater,
		hclog.NewNullLogger(),
	)

	return ethTxsProcessor
}

func TestEthTxsProcessor(t *testing.T) {
	appConfig := &core.AppConfig{
		EthChains: map[string]*core.EthChainConfig{
			common.ChainIDStrNexus: {},
		},
		BridgingSettings: core.BridgingSettings{
			MaxBridgingClaimsToGroup: 10,
		},
	}

	appConfig.FillOut()
	appConfig.EthChains[common.ChainIDStrNexus].NodeURL = "http://127.0.0.1"

	const (
		dbFilePath      = "temp_test_oracle.db"
		nexusDBFilePath = "temp_test_nexus.db"
	)

	dbCleanup := func() {
		common.RemoveDirOrFilePathIfExists(dbFilePath)      //nolint:errcheck
		common.RemoveDirOrFilePathIfExists(nexusDBFilePath) //nolint:errcheck
	}

	t.Cleanup(dbCleanup)

	t.Run("TestEthTxsProcessor", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		proc := NewEthTxsProcessor(context.Background(), appConfig, nil, nil, nil, nil, nil, nil, nil)
		require.NotNil(t, proc)

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{common.ChainIDStrNexus: &ethcore.EventStoreMock{}}

		proc = NewEthTxsProcessor(
			context.Background(),
			appConfig,
			&ethcore.EthTxsProcessorDBMock{},
			[]ethcore.EthTxProcessor{},
			[]ethcore.EthTxFailedProcessor{},
			&ethcore.BridgeSubmitterMock{}, indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
			hclog.NewNullLogger(),
		)
		require.NotNil(t, proc)
	})

	t.Run("NewUnprocessedTxs nil txs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		validTxProc := &ethcore.EthTxProcessorMock{}
		failedTxProc := &ethcore.EthTxFailedProcessorMock{}
		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{common.ChainIDStrNexus: &ethcore.EventStoreMock{}}

		oracleDB, err := databaseaccess.NewDatabase(dbFilePath)
		require.NoError(t, err)

		proc := newValidProcessor(
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		require.NoError(t, proc.NewUnprocessedLog(common.ChainIDStrNexus, nil))

		unprocessedTxs, err := oracleDB.GetAllUnprocessedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.Nil(t, unprocessedTxs)
	})

	t.Run("NewUnprocessedTxs no txs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{common.ChainIDStrNexus: &ethcore.EventStoreMock{}}

		oracleDB, err := databaseaccess.NewDatabase(dbFilePath)
		require.NoError(t, err)

		proc := newValidProcessor(
			appConfig, oracleDB,
			nil, nil, nil,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		require.NoError(t, proc.NewUnprocessedLog(common.ChainIDStrNexus, &ethgo.Log{}))

		unprocessedTxs, err := oracleDB.GetAllUnprocessedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.Nil(t, unprocessedTxs)
	})

	t.Run("NewUnprocessedTxs no relevant txs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{common.ChainIDStrNexus: &ethcore.EventStoreMock{}}

		oracleDB, err := databaseaccess.NewDatabase(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxProcessorMock{Type: "relevant"}

		proc := newValidProcessor(
			appConfig, oracleDB,
			validTxProc, nil, nil,
			indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		require.NoError(t, proc.NewUnprocessedLog(common.ChainIDStrNexus, &ethgo.Log{
			BlockHash: ethgo.Hash{1},
		}))

		unprocessedTxs, err := oracleDB.GetAllUnprocessedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.Nil(t, unprocessedTxs)
	})

	t.Run("NewUnprocessedTxs valid txs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{common.ChainIDStrNexus: &ethcore.EventStoreMock{}}

		oracleDB, err := databaseaccess.NewDatabase(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxProcessorMock{ShouldAddClaim: true, Type: "batch"}

		txHash := ethgo.Hash{1}

		proc := newValidProcessor(
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

		require.NoError(t, proc.NewUnprocessedLog(common.ChainIDStrNexus, log))

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

		oracleDB, err := databaseaccess.NewDatabase(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxProcessorMock{ShouldAddClaim: true, Type: "batch"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("test err"))

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil)

		txHash := ethgo.Hash{1}

		proc := newValidProcessor(
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

		require.NoError(t, proc.NewUnprocessedLog(common.ChainIDStrNexus, log))

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(originChainID, 0)
		require.Nil(t, unprocessedTxs)

		processedTx, _ := oracleDB.GetProcessedTx(originChainID, txHash)
		require.NotNil(t, processedTx)
		require.True(t, processedTx.IsInvalid)
	})

	t.Run("NewUnprocessedTxs - submit claims failed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		const (
			originChainID = common.ChainIDStrNexus
		)

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{originChainID: &ethcore.EventStoreMock{}}

		oracleDB, err := databaseaccess.NewDatabase(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxProcessorMock{ShouldAddClaim: true, Type: "batch"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("test err"))

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(fmt.Errorf("test err"))

		txHash := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910995")

		proc := newValidProcessor(
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

		require.NoError(t, proc.NewUnprocessedLog(common.ChainIDStrNexus, log))

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(originChainID, 0)
		require.Len(t, unprocessedTxs, 1)
		require.Equal(t, txHash, unprocessedTxs[0].Hash)
		require.Equal(t, originChainID, unprocessedTxs[0].OriginChainID)
		processedTx, _ := oracleDB.GetProcessedTx(originChainID, txHash)
		require.Nil(t, processedTx)
	})

	t.Run("Start - unprocessedTxs - valid", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		const (
			originChainID = common.ChainIDStrNexus
		)

		indexerDbs := map[string]eventTrackerStore.EventTrackerStore{originChainID: &ethcore.EventStoreMock{}}

		oracleDB, err := databaseaccess.NewDatabase(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxProcessorMock{ShouldAddClaim: true, Type: "batch"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil)

		txHash := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910995")

		proc := newValidProcessor(
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

		require.NoError(t, proc.NewUnprocessedLog(common.ChainIDStrNexus, log))

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(originChainID, 0)
		require.Nil(t, unprocessedTxs)

		processedTx, _ := oracleDB.GetProcessedTx(originChainID, txHash)
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

		oracleDB, err := databaseaccess.NewDatabase(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxProcessorMock{ShouldAddClaim: true, Type: "batch"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &ethcore.EthTxFailedProcessorMock{Type: "test"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("test err"))

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil)

		txHash := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910995")

		proc := newValidProcessor(
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

		require.NoError(t, proc.NewUnprocessedLog(originChainID, log))

		metadata, err := ethcore.MarshalEthMetadata(ethcore.BaseEthMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		err = oracleDB.AddExpectedTxs([]*ethcore.BridgeExpectedEthTx{
			{ChainID: originChainID, Hash: txHash, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

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

		oracleDB, err := databaseaccess.NewDatabase(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxProcessorMock{ShouldAddClaim: true, Type: "batch"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &ethcore.EthTxFailedProcessorMock{Type: "test"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("test err"))

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(fmt.Errorf("test err"))

		txHash := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910995")

		proc := newValidProcessor(
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

		require.NoError(t, proc.NewUnprocessedLog(originChainID, log))

		metadata, err := ethcore.MarshalEthMetadata(ethcore.BaseEthMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		err = oracleDB.AddExpectedTxs([]*ethcore.BridgeExpectedEthTx{
			{ChainID: originChainID, Hash: txHash, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

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

		oracleDB, err := databaseaccess.NewDatabase(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxProcessorMock{ShouldAddClaim: true, Type: "batch"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &ethcore.EthTxFailedProcessorMock{ShouldAddClaim: true, Type: "test"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*core.BridgeClaims

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *core.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil)

		txHash := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910995")

		proc := newValidProcessor(
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

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

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

		oracleDB, err := databaseaccess.NewDatabase(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxProcessorMock{}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &ethcore.EthTxFailedProcessorMock{ShouldAddClaim: true, Type: "batch"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*core.BridgeClaims

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *core.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil)

		txHash := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910995")

		proc := newValidProcessor(
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

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

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

		oracleDB, err := databaseaccess.NewDatabase(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxProcessorMock{ShouldAddClaim: true, Type: "batch"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &ethcore.EthTxFailedProcessorMock{ShouldAddClaim: true, Type: "batch"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*core.BridgeClaims

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *core.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil)

		txHash := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910995")
		txHash2 := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910996")

		proc := newValidProcessor(
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

		require.NoError(t, proc.NewUnprocessedLog(chainID, log))

		metadata, err := ethcore.MarshalEthMetadata(ethcore.BaseEthMetadata{BridgingTxType: "batch"})
		require.NoError(t, err)

		err = oracleDB.AddExpectedTxs([]*ethcore.BridgeExpectedEthTx{
			{ChainID: chainID, Hash: txHash2, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		store.On("GetLastProcessedBlock").Return(blockSlot, nil)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(chainID, 0)
		require.Nil(t, unprocessedTxs)

		processedTx, _ := oracleDB.GetProcessedTx(chainID, txHash)
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

		oracleDB, err := databaseaccess.NewDatabase(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxProcessorMock{ShouldAddClaim: true, Type: "batch"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &ethcore.EthTxFailedProcessorMock{ShouldAddClaim: true, Type: "batch"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*core.BridgeClaims

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *core.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil)

		txHash := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910995")
		txHash2 := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910996")

		proc := newValidProcessor(
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

		require.NoError(t, proc.NewUnprocessedLog(chainID, log))

		metadata, err := ethcore.MarshalEthMetadata(ethcore.BaseEthMetadata{BridgingTxType: "batch"})
		require.NoError(t, err)

		err = oracleDB.AddExpectedTxs([]*ethcore.BridgeExpectedEthTx{
			{ChainID: chainID, Hash: txHash2, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		store.On("GetLastProcessedBlock").Return(blockSlot, nil)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(chainID, 0)
		require.Nil(t, unprocessedTxs)

		processedTx, _ := oracleDB.GetProcessedTx(chainID, txHash)
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

		oracleDB, err := databaseaccess.NewDatabase(dbFilePath)
		require.NoError(t, err)

		validTxProc := &ethcore.EthTxProcessorMock{ShouldAddClaim: true, Type: common.BridgingTxTypeBatchExecution}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &ethcore.EthTxFailedProcessorMock{ShouldAddClaim: true, Type: common.BridgingTxTypeBatchExecution}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*core.BridgeClaims

		bridgeSubmitter := &ethcore.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *core.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil)

		txHash := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910995")
		txHash2 := ethgo.HexToHash("0xf62590f36f8b18f71bb343ad6e861ad62ac23bece85414772c7f06f1b1910996")

		proc := newValidProcessor(
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

		require.NoError(t, proc.NewUnprocessedLog(chainID, log))

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

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(chainID, 0)
		require.Nil(t, unprocessedTxs)

		processedTx, _ := oracleDB.GetProcessedTx(chainID, txHash)
		require.NotNil(t, processedTx)
		require.False(t, processedTx.IsInvalid)

		expectedTxs, _ := oracleDB.GetAllExpectedTxs(chainID, 0)
		require.Nil(t, expectedTxs)

		require.NotNil(t, submittedClaims)
		require.Len(t, submittedClaims, 1)
		require.Len(t, submittedClaims[0].BridgingRequestClaims, 1)
		require.Len(t, submittedClaims[0].BatchExecutionFailedClaims, 1)
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
