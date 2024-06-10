package processor

import (
	"context"
	"fmt"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	databaseaccess "github.com/Ethernal-Tech/apex-bridge/oracle/database_access"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	indexerDb "github.com/Ethernal-Tech/cardano-infrastructure/indexer/db"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newValidProcessor(
	appConfig *core.AppConfig,
	oracleDB core.Database,
	txProcessor core.CardanoTxProcessor,
	failedTxProcessor core.CardanoTxFailedProcessor,
	bridgeSubmitter core.BridgeSubmitter,
	ccoDbs map[string]indexer.Database,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
) *CardanoTxsProcessorImpl {
	var txProcessors []core.CardanoTxProcessor
	if txProcessor != nil {
		txProcessors = append(txProcessors, txProcessor)
	}

	var failedTxProcessors []core.CardanoTxFailedProcessor
	if failedTxProcessor != nil {
		failedTxProcessors = append(failedTxProcessors, failedTxProcessor)
	}

	cardanoTxsProcessor := NewCardanoTxsProcessor(
		context.Background(),
		appConfig, oracleDB,
		txProcessors, failedTxProcessors,
		bridgeSubmitter, ccoDbs,
		bridgingRequestStateUpdater,
		hclog.NewNullLogger(),
	)

	return cardanoTxsProcessor
}

func TestCardanoTxsProcessor(t *testing.T) {
	appConfig := &core.AppConfig{
		CardanoChains: map[string]*core.CardanoChainConfig{
			"prime":  {},
			"vector": {},
		},
		BridgingSettings: core.BridgingSettings{
			MaxBridgingClaimsToGroup: 10,
		},
	}

	appConfig.FillOut()

	const (
		dbFilePath       = "temp_test_oracle.db"
		primeDBFilePath  = "temp_test_prime.db"
		vectorDBFilePath = "temp_test_vector.db"
	)

	createDbs := func() (core.Database, indexer.Database, indexer.Database) {
		oracleDB, _ := databaseaccess.NewDatabase(dbFilePath)
		primeDB, _ := indexerDb.NewDatabaseInit("", primeDBFilePath)
		vectorDB, _ := indexerDb.NewDatabaseInit("", vectorDBFilePath)

		return oracleDB, primeDB, vectorDB
	}

	dbCleanup := func() {
		common.RemoveDirOrFilePathIfExists(dbFilePath)       //nolint:errcheck
		common.RemoveDirOrFilePathIfExists(primeDBFilePath)  //nolint:errcheck
		common.RemoveDirOrFilePathIfExists(vectorDBFilePath) //nolint:errcheck
	}

	t.Cleanup(dbCleanup)

	t.Run("NewCardanoTxsProcessor", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, primeDB, vectorDB := createDbs()

		proc := NewCardanoTxsProcessor(context.Background(), appConfig, nil, nil, nil, nil, nil, nil, nil)
		require.NotNil(t, proc)

		indexerDbs := map[string]indexer.Database{"prime": primeDB, "vector": vectorDB}

		proc = NewCardanoTxsProcessor(
			context.Background(),
			appConfig,
			oracleDB,
			[]core.CardanoTxProcessor{},
			[]core.CardanoTxFailedProcessor{},
			&core.BridgeSubmitterMock{}, indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
			hclog.NewNullLogger(),
		)
		require.NotNil(t, proc)
	})

	t.Run("NewUnprocessedTxs nil txs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, primeDB, vectorDB := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{}
		failedTxProc := &core.CardanoTxFailedProcessorMock{}
		bridgeSubmitter := &core.BridgeSubmitterMock{}

		proc := newValidProcessor(
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDB, "vector": vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		require.NoError(t, proc.NewUnprocessedTxs("prime", nil))

		unprocessedTxs, err := oracleDB.GetAllUnprocessedTxs("prime", 0)
		require.NoError(t, err)
		require.Nil(t, unprocessedTxs)
	})

	t.Run("NewUnprocessedTxs no txs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, primeDB, vectorDB := createDbs()

		proc := newValidProcessor(
			appConfig, oracleDB,
			nil, nil, nil,
			map[string]indexer.Database{"prime": primeDB, "vector": vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		require.NoError(t, proc.NewUnprocessedTxs("prime", []*indexer.Tx{}))

		unprocessedTxs, err := oracleDB.GetAllUnprocessedTxs("prime", 0)
		require.NoError(t, err)
		require.Nil(t, unprocessedTxs)
	})

	t.Run("NewUnprocessedTxs no relevant txs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, primeDB, vectorDB := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{Type: "relevant"}

		proc := newValidProcessor(
			appConfig, oracleDB,
			validTxProc, nil, nil,
			map[string]indexer.Database{"prime": primeDB, "vector": vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		require.NoError(t, proc.NewUnprocessedTxs("prime", []*indexer.Tx{
			{Hash: "test_hash"},
		}))

		unprocessedTxs, err := oracleDB.GetAllUnprocessedTxs("prime", 0)
		require.NoError(t, err)
		require.Nil(t, unprocessedTxs)
	})

	t.Run("NewUnprocessedTxs valid", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, primeDB, vectorDB := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{Type: "test"}

		proc := newValidProcessor(
			appConfig, oracleDB,
			validTxProc, nil, nil,
			map[string]indexer.Database{"prime": primeDB, "vector": vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			originChainID = "prime"
			txHash        = "test_hash"
		)

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		require.NoError(t, proc.NewUnprocessedTxs(originChainID, []*indexer.Tx{
			{Hash: txHash, Metadata: metadata},
		}))

		unprocessedTxs, err := oracleDB.GetAllUnprocessedTxs(originChainID, 0)
		require.NoError(t, err)
		require.Len(t, unprocessedTxs, 1)
		require.Equal(t, txHash, unprocessedTxs[0].Hash)
		require.Equal(t, originChainID, unprocessedTxs[0].OriginChainID)
	})

	t.Run("Start - unprocessedTxs - tx validation err", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, primeDB, vectorDB := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{Type: "test"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("test err"))

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything).Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDB,
			validTxProc, nil, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDB, "vector": vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			originChainID = "prime"
			txHash        = "test_hash"
		)

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		require.NoError(t, proc.NewUnprocessedTxs(originChainID, []*indexer.Tx{
			{Hash: txHash, Metadata: metadata},
		}))

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(5 * time.Second)
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

	t.Run("Start - unprocessedTxs - submit claims failed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, primeDB, vectorDB := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{ShouldAddClaim: true, Type: "test"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything).Return(fmt.Errorf("test err"))

		proc := newValidProcessor(
			appConfig, oracleDB,
			validTxProc, nil, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDB, "vector": vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			originChainID = "prime"
			txHash        = "test_hash"
		)

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		require.NoError(t, proc.NewUnprocessedTxs(originChainID, []*indexer.Tx{
			{Hash: txHash, Metadata: metadata},
		}))

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(5 * time.Second)

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

		oracleDB, primeDB, vectorDB := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{ShouldAddClaim: true, Type: "test"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything).Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDB,
			validTxProc, nil, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDB, "vector": vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			originChainID = "prime"
			txHash        = "test_hash"
		)

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		require.NoError(t, proc.NewUnprocessedTxs(originChainID, []*indexer.Tx{
			{Hash: txHash, Metadata: metadata},
		}))

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(5 * time.Second)

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

		oracleDB, primeDB, vectorDB := createDbs()

		failedTxProc := &core.CardanoTxFailedProcessorMock{Type: "test"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("test err"))

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything).Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDB,
			&core.CardanoTxProcessorMock{}, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDB, "vector": vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			chainID = "prime"
			txHash  = "test_hash"
			ttl     = 2
		)

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		err = oracleDB.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainID: chainID, Hash: txHash, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		require.NoError(t, primeDB.OpenTx().AddConfirmedBlock(&indexer.CardanoBlock{Slot: 6, Hash: "test_block_hash"}).Execute())

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(5 * time.Second)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		expectedTxs, _ := oracleDB.GetAllExpectedTxs(chainID, 0)
		require.Nil(t, expectedTxs)
	})

	t.Run("Start - expectedTxs - submit claims failed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, primeDB, vectorDB := createDbs()

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true, Type: "test"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything).Return(fmt.Errorf("test err"))

		proc := newValidProcessor(
			appConfig, oracleDB,
			&core.CardanoTxProcessorMock{}, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDB, "vector": vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			chainID = "prime"
			txHash  = "test_hash"
			ttl     = 2
		)

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		err = oracleDB.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainID: chainID, Hash: txHash, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		require.NoError(t, primeDB.OpenTx().AddConfirmedBlock(&indexer.CardanoBlock{Slot: 6, Hash: "test_block_hash"}).Execute())

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(5 * time.Second)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		expectedTxs, _ := oracleDB.GetAllExpectedTxs(chainID, 0)
		require.NotNil(t, expectedTxs)
	})

	t.Run("Start - expectedTxs - valid - tx not yet expired", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, primeDB, vectorDB := createDbs()

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true, Type: "test"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*core.BridgeClaims

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *core.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything).Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDB,
			&core.CardanoTxProcessorMock{}, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDB, "vector": vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			chainID = "prime"
			txHash  = "test_hash"
			ttl     = 2
		)

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		err = oracleDB.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainID: chainID, Hash: txHash, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(5 * time.Second)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		expectedTxs, _ := oracleDB.GetAllExpectedTxs(chainID, 0)
		require.NotNil(t, expectedTxs)
		require.Nil(t, submittedClaims)
	})

	t.Run("Start - expectedTxs - valid - expired tx", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, primeDB, vectorDB := createDbs()

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true, Type: "test"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*core.BridgeClaims

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *core.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything).Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDB,
			&core.CardanoTxProcessorMock{}, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDB, "vector": vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			chainID = "prime"
			txHash  = "test_hash"
			ttl     = 2
		)

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		err = oracleDB.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainID: chainID, Hash: txHash, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		require.NoError(t, primeDB.OpenTx().AddConfirmedBlock(&indexer.CardanoBlock{Slot: 6, Hash: "test_block_hash"}).Execute())

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(5 * time.Second)

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

		oracleDB, primeDB, vectorDB := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{ShouldAddClaim: true, Type: "test"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true, Type: "test"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*core.BridgeClaims

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *core.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything).Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDB, "vector": vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			chainID   = "prime"
			txHash1   = "test_hash_1"
			txHash2   = "test_hash_2"
			ttl       = 2
			blockSlot = 6
			blockHash = "test_block_hash"
		)

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		require.NoError(t, proc.NewUnprocessedTxs(chainID, []*indexer.Tx{
			{Hash: txHash1, BlockSlot: blockSlot, BlockHash: blockHash, Metadata: metadata},
		}))

		err = oracleDB.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainID: chainID, Hash: txHash2, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		require.NoError(t, primeDB.OpenTx().AddConfirmedBlock(&indexer.CardanoBlock{Slot: blockSlot, Hash: blockHash}).Execute())

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(5 * time.Second)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(chainID, 0)
		require.Nil(t, unprocessedTxs)

		processedTx, _ := oracleDB.GetProcessedTx(chainID, txHash1)
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

		oracleDB, primeDB, vectorDB := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{ShouldAddClaim: true, Type: "test"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true, Type: "test"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*core.BridgeClaims

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *core.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything).Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDB, "vector": vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			chainID   = "prime"
			txHash1   = "test_hash_1"
			txHash2   = "test_hash_2"
			ttl       = 2
			blockSlot = 6
			blockHash = "test_block_hash"
		)

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		require.NoError(t, proc.NewUnprocessedTxs(chainID, []*indexer.Tx{
			{Hash: txHash1, BlockSlot: blockSlot - 1, BlockHash: blockHash, Metadata: metadata},
		}))

		err = oracleDB.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainID: chainID, Hash: txHash2, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		require.NoError(t, primeDB.OpenTx().AddConfirmedBlock(&indexer.CardanoBlock{Slot: blockSlot, Hash: blockHash}).Execute())

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(12 * time.Second)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(chainID, 0)
		require.Nil(t, unprocessedTxs)

		processedTx, _ := oracleDB.GetProcessedTx(chainID, txHash1)
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

		oracleDB, primeDB, vectorDB := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{ShouldAddClaim: true, Type: common.BridgingTxTypeBatchExecution}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true, Type: common.BridgingTxTypeBatchExecution}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*core.BridgeClaims

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *core.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything).Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDB, "vector": vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			chainID   = "prime"
			txHash1   = "test_hash_1"
			txHash2   = "test_hash_2"
			ttl       = 2
			blockSlot = 6
			blockHash = "test_block_hash"
		)

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: common.BridgingTxTypeBatchExecution})
		require.NoError(t, err)

		require.NoError(t, proc.NewUnprocessedTxs(chainID, []*indexer.Tx{
			{Hash: txHash1, BlockSlot: blockSlot - 1, BlockHash: blockHash, Metadata: metadata},
		}))

		err = oracleDB.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainID: chainID, Hash: txHash1, TTL: blockSlot + 2, Metadata: metadata},
		})
		require.NoError(t, err)

		err = oracleDB.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainID: chainID, Hash: txHash2, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		require.NoError(t, primeDB.OpenTx().AddConfirmedBlock(&indexer.CardanoBlock{Slot: blockSlot, Hash: blockHash}).Execute())

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(12 * time.Second)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(chainID, 0)
		require.Nil(t, unprocessedTxs)

		processedTx, _ := oracleDB.GetProcessedTx(chainID, txHash1)
		require.NotNil(t, processedTx)
		require.False(t, processedTx.IsInvalid)

		expectedTxs, _ := oracleDB.GetAllExpectedTxs(chainID, 0)
		require.Nil(t, expectedTxs)

		require.NotNil(t, submittedClaims)
		require.Len(t, submittedClaims, 1)
		require.Len(t, submittedClaims[0].BridgingRequestClaims, 1)
		require.Len(t, submittedClaims[0].BatchExecutionFailedClaims, 1)
	})

	t.Run("Start - unprocessedTxs, expectedTxs - multiple chains - valid 1", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		dbCleanup()

		oracleDB, primeDB, vectorDB := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{ShouldAddClaim: true, Type: "test"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true, Type: "test"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*core.BridgeClaims

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *core.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything).Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDB, "vector": vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			chainID1  = "prime"
			chainID2  = "vector"
			txHash1   = "test_hash_1"
			txHash2   = "test_hash_2"
			ttl       = 2
			blockSlot = 6
			blockHash = "test_block_hash"
		)

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		require.NoError(t, proc.NewUnprocessedTxs(chainID1, []*indexer.Tx{
			{Hash: txHash1, BlockSlot: blockSlot - 1, BlockHash: blockHash, Metadata: metadata},
		}))

		err = oracleDB.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainID: chainID2, Hash: txHash2, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		require.NoError(t, vectorDB.OpenTx().AddConfirmedBlock(&indexer.CardanoBlock{Slot: blockSlot, Hash: blockHash}).Execute())

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(12 * time.Second)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(chainID1, 0)
		require.Nil(t, unprocessedTxs)

		processedTx, _ := oracleDB.GetProcessedTx(chainID1, txHash1)
		require.NotNil(t, processedTx)
		require.False(t, processedTx.IsInvalid)

		expectedTxs, _ := oracleDB.GetAllExpectedTxs(chainID2, 0)
		require.Nil(t, expectedTxs)

		require.NotNil(t, submittedClaims)
		require.Len(t, submittedClaims, 1)
		require.Len(t, submittedClaims[0].BridgingRequestClaims, 1)
		require.Len(t, submittedClaims[0].BatchExecutionFailedClaims, 1)
	})

	t.Run("Start - unprocessedTxs, expectedTxs - multiple chains - valid 2", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		dbCleanup()

		oracleDB, primeDB, vectorDB := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{ShouldAddClaim: true, Type: "test"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true, Type: "test"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*core.BridgeClaims

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *core.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything).Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDB, "vector": vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			chainID1  = "prime"
			chainID2  = "vector"
			txHash1   = "test_hash_1"
			txHash2   = "test_hash_2"
			ttl       = 2
			blockSlot = 6
			blockHash = "test_block_hash"
		)

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		require.NoError(t, proc.NewUnprocessedTxs(chainID1, []*indexer.Tx{
			{Hash: txHash1, BlockSlot: blockSlot - 1, BlockHash: blockHash, Metadata: metadata},
		}))

		require.NoError(t, proc.NewUnprocessedTxs(chainID1, []*indexer.Tx{
			{Hash: txHash2, BlockSlot: blockSlot - 1, BlockHash: blockHash, Metadata: metadata},
		}))

		err = oracleDB.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainID: chainID2, Hash: txHash1, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)
		err = oracleDB.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainID: chainID2, Hash: txHash2, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		require.NoError(t, vectorDB.OpenTx().AddConfirmedBlock(
			&indexer.CardanoBlock{Slot: blockSlot, Hash: blockHash},
		).Execute())

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(12 * time.Second)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(chainID1, 0)
		require.Nil(t, unprocessedTxs)

		processedTx, _ := oracleDB.GetProcessedTx(chainID1, txHash1)
		require.NotNil(t, processedTx)
		require.False(t, processedTx.IsInvalid)

		expectedTxs, _ := oracleDB.GetAllExpectedTxs(chainID2, 0)
		require.Nil(t, expectedTxs)

		require.NotNil(t, submittedClaims)
		require.Len(t, submittedClaims, 1)

		require.Len(t, submittedClaims[0].BridgingRequestClaims, 2)
		require.Len(t, submittedClaims[0].BatchExecutionFailedClaims, 2)
	})

	t.Run("Start - unprocessedTxs, expectedTxs - multiple chains - valid 3", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		dbCleanup()

		oracleDB, primeDB, vectorDB := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{ShouldAddClaim: true, Type: "test"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true, Type: "test"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*core.BridgeClaims

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *core.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything).Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDB, "vector": vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			chainID1   = "prime"
			chainID2   = "vector"
			txHash1    = "test_hash_1"
			txHash2    = "test_hash_2"
			ttl1       = 2
			blockSlot1 = 6
			ttl2       = 10
			blockSlot2 = 15
			blockHash  = "test_block_hash"
		)

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		require.NoError(t, proc.NewUnprocessedTxs(chainID1, []*indexer.Tx{
			{Hash: txHash1, BlockSlot: blockSlot1 - 1, BlockHash: blockHash, Metadata: metadata},
		}))

		require.NoError(t, proc.NewUnprocessedTxs(chainID1, []*indexer.Tx{
			{Hash: txHash2, BlockSlot: blockSlot1, BlockHash: blockHash, Metadata: metadata},
		}))

		err = oracleDB.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainID: chainID2, Hash: txHash1, TTL: ttl1, Metadata: metadata},
		})
		require.NoError(t, err)
		err = oracleDB.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainID: chainID2, Hash: txHash2, TTL: ttl2, Metadata: metadata},
		})
		require.NoError(t, err)

		require.NoError(t, vectorDB.OpenTx().AddConfirmedBlock(
			&indexer.CardanoBlock{Slot: blockSlot1, Hash: blockHash},
		).AddConfirmedBlock(
			&indexer.CardanoBlock{Slot: blockSlot2, Hash: blockHash},
		).Execute())

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(12 * time.Second)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(chainID1, 0)
		require.Nil(t, unprocessedTxs)

		processedTx, _ := oracleDB.GetProcessedTx(chainID1, txHash1)
		require.NotNil(t, processedTx)
		require.False(t, processedTx.IsInvalid)

		expectedTxs, _ := oracleDB.GetAllExpectedTxs(chainID2, 0)
		require.Nil(t, expectedTxs)

		require.NotNil(t, submittedClaims)
		require.Len(t, submittedClaims, 1)

		require.Len(t, submittedClaims[0].BridgingRequestClaims, 2)
		require.Len(t, submittedClaims[0].BatchExecutionFailedClaims, 2)
	})
}
