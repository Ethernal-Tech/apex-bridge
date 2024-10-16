package processor

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	databaseaccess "github.com/Ethernal-Tech/apex-bridge/oracle_cardano/database_access"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	txsprocessor "github.com/Ethernal-Tech/apex-bridge/oracle_common/processor/txs_processor"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	indexerDb "github.com/Ethernal-Tech/cardano-infrastructure/indexer/db"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newCardanoTxsProcessor(
	ctx context.Context,
	appConfig *cCore.AppConfig,
	db core.CardanoTxsProcessorDB,
	successTxProcessors []core.CardanoTxSuccessProcessor,
	failedTxProcessors []core.CardanoTxFailedProcessor,
	bridgeSubmitter core.BridgeSubmitter,
	indexerDbs map[string]indexer.Database,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
) (*txsprocessor.TxsProcessorImpl, *CardanoTxsReceiverImpl) {
	txProcessors := NewTxProcessorsCollection(
		successTxProcessors, failedTxProcessors,
	)

	cardanoTxsReceiver := NewCardanoTxsReceiverImpl(db, txProcessors, bridgingRequestStateUpdater, hclog.NewNullLogger())

	cardanoStateProcessor := NewCardanoStateProcessor(
		ctx, appConfig, db, txProcessors,
		indexerDbs, hclog.NewNullLogger(),
	)

	cardanoTxsProcessor := txsprocessor.NewTxsProcessorImpl(
		ctx, appConfig, cardanoStateProcessor, bridgeSubmitter, bridgingRequestStateUpdater,
		hclog.NewNullLogger(),
	)

	return cardanoTxsProcessor, cardanoTxsReceiver
}

func newValidProcessor(
	ctx context.Context,
	appConfig *cCore.AppConfig,
	oracleDB core.Database,
	successTxProcessor core.CardanoTxSuccessProcessor,
	failedTxProcessor core.CardanoTxFailedProcessor,
	bridgeSubmitter core.BridgeSubmitter,
	indexerDbs map[string]indexer.Database,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
) (*txsprocessor.TxsProcessorImpl, *CardanoTxsReceiverImpl) {
	var successTxProcessors []core.CardanoTxSuccessProcessor
	if successTxProcessor != nil {
		successTxProcessors = append(successTxProcessors, successTxProcessor)
	}

	var failedTxProcessors []core.CardanoTxFailedProcessor
	if failedTxProcessor != nil {
		failedTxProcessors = append(failedTxProcessors, failedTxProcessor)
	}

	return newCardanoTxsProcessor(
		ctx, appConfig, oracleDB, successTxProcessors, failedTxProcessors,
		bridgeSubmitter, indexerDbs, bridgingRequestStateUpdater)
}

func TestCardanoTxsProcessor(t *testing.T) {
	appConfig := &cCore.AppConfig{
		CardanoChains: map[string]*cCore.CardanoChainConfig{
			common.ChainIDStrPrime:  {},
			common.ChainIDStrVector: {},
		},
		BridgingSettings: cCore.BridgingSettings{
			MaxBridgingClaimsToGroup: 10,
		},
	}

	appConfig.FillOut()

	const (
		dbFilePath       = "temp_test_oracle.db"
		primeDBFilePath  = "temp_test_prime.db"
		vectorDBFilePath = "temp_test_vector.db"

		processingWaitTimeMs = 300
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

		proc, rec := newCardanoTxsProcessor(context.Background(), appConfig, nil, nil, nil, nil, nil, nil)
		require.NotNil(t, proc)
		require.NotNil(t, rec)

		indexerDbs := map[string]indexer.Database{common.ChainIDStrPrime: primeDB, common.ChainIDStrVector: vectorDB}

		proc, rec = newCardanoTxsProcessor(
			context.Background(),
			appConfig,
			oracleDB,
			[]core.CardanoTxSuccessProcessor{},
			[]core.CardanoTxFailedProcessor{},
			&core.BridgeSubmitterMock{}, indexerDbs,
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)
		require.NotNil(t, proc)
		require.NotNil(t, rec)
	})

	t.Run("NewUnprocessedTxs nil txs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, primeDB, vectorDB := createDbs()

		validTxProc := &core.CardanoTxSuccessProcessorMock{}
		failedTxProc := &core.CardanoTxFailedProcessorMock{}
		bridgeSubmitter := &core.BridgeSubmitterMock{}

		proc, rec := newValidProcessor(
			context.Background(),
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{common.ChainIDStrPrime: primeDB, common.ChainIDStrVector: vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		require.NoError(t, rec.NewUnprocessedTxs(common.ChainIDStrPrime, nil))

		unprocessedTxs, err := oracleDB.GetAllUnprocessedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.Nil(t, unprocessedTxs)
	})

	t.Run("NewUnprocessedTxs no txs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, primeDB, vectorDB := createDbs()

		proc, rec := newValidProcessor(
			context.Background(),
			appConfig, oracleDB,
			nil, nil, nil,
			map[string]indexer.Database{common.ChainIDStrPrime: primeDB, common.ChainIDStrVector: vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		require.NoError(t, rec.NewUnprocessedTxs(common.ChainIDStrPrime, []*indexer.Tx{}))

		unprocessedTxs, err := oracleDB.GetAllUnprocessedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.Nil(t, unprocessedTxs)
	})

	t.Run("NewUnprocessedTxs no relevant txs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, primeDB, vectorDB := createDbs()

		validTxProc := &core.CardanoTxSuccessProcessorMock{Type: "relevant"}

		proc, rec := newValidProcessor(
			context.Background(),
			appConfig, oracleDB,
			validTxProc, nil, nil,
			map[string]indexer.Database{common.ChainIDStrPrime: primeDB, common.ChainIDStrVector: vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		require.NoError(t, rec.NewUnprocessedTxs(common.ChainIDStrPrime, []*indexer.Tx{
			{Hash: indexer.Hash{1}},
		}))

		unprocessedTxs, err := oracleDB.GetAllUnprocessedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.Nil(t, unprocessedTxs)
	})

	t.Run("NewUnprocessedTxs valid", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, primeDB, vectorDB := createDbs()

		validTxProc := &core.CardanoTxSuccessProcessorMock{Type: "test"}

		proc, rec := newValidProcessor(
			context.Background(),
			appConfig, oracleDB,
			validTxProc, nil, nil,
			map[string]indexer.Database{common.ChainIDStrPrime: primeDB, common.ChainIDStrVector: vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			originChainID = common.ChainIDStrPrime
		)

		txHash := indexer.Hash{1}

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		require.NoError(t, rec.NewUnprocessedTxs(originChainID, []*indexer.Tx{
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

		validTxProc := &core.CardanoTxSuccessProcessorMock{Type: "test"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("test err"))

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil)

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, rec := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, nil, bridgeSubmitter,
			map[string]indexer.Database{common.ChainIDStrPrime: primeDB, common.ChainIDStrVector: vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			originChainID = common.ChainIDStrPrime
		)

		txHash := indexer.Hash(common.NewHashFromHexString("0x89FF"))

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		require.NoError(t, rec.NewUnprocessedTxs(originChainID, []*indexer.Tx{
			{Hash: txHash, Metadata: metadata},
		}))

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(5 * time.Second)
		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(originChainID, 0)
		require.Nil(t, unprocessedTxs)

		processedTx, _ := oracleDB.GetProcessedTx(originChainID, txHash)
		require.NotNil(t, processedTx)
		require.True(t, processedTx.IsInvalid)
	})

	t.Run("Start - unprocessedTxs - submit claims failed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, primeDB, vectorDB := createDbs()

		validTxProc := &core.CardanoTxSuccessProcessorMock{ShouldAddClaim: true, Type: "test"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(fmt.Errorf("test err"))

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, rec := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, nil, bridgeSubmitter,
			map[string]indexer.Database{common.ChainIDStrPrime: primeDB, common.ChainIDStrVector: vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			originChainID = common.ChainIDStrPrime
		)

		txHash := indexer.Hash(common.NewHashFromHexString("0xFFAA"))

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		require.NoError(t, rec.NewUnprocessedTxs(originChainID, []*indexer.Tx{
			{Hash: txHash, Metadata: metadata},
		}))

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(5 * time.Second)
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
		processedTx, _ := oracleDB.GetProcessedTx(originChainID, txHash)
		require.Nil(t, processedTx)
	})

	t.Run("Start - unprocessedTxs - valid", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, primeDB, vectorDB := createDbs()

		validTxProc := &core.CardanoTxSuccessProcessorMock{ShouldAddClaim: true, Type: "test"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil)

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, rec := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, nil, bridgeSubmitter,
			map[string]indexer.Database{common.ChainIDStrPrime: primeDB, common.ChainIDStrVector: vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			originChainID = common.ChainIDStrPrime
		)

		txHash := indexer.Hash(common.NewHashFromHexString("0xFFAABB"))

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		require.NoError(t, rec.NewUnprocessedTxs(originChainID, []*indexer.Tx{
			{Hash: txHash, Metadata: metadata},
		}))

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(5 * time.Second)
		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

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
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil)

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, _ := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			&core.CardanoTxSuccessProcessorMock{}, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{common.ChainIDStrPrime: primeDB, common.ChainIDStrVector: vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			chainID = common.ChainIDStrPrime
			ttl     = 2
		)

		txHash := indexer.Hash(common.NewHashFromHexString("0xFFAACC"))

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		err = oracleDB.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainID: chainID, Hash: txHash, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		require.NoError(t, primeDB.OpenTx().AddConfirmedBlock(&indexer.CardanoBlock{Slot: 6, Hash: indexer.Hash{1}}).Execute())

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(5 * time.Second)
		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

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
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(fmt.Errorf("test err"))

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, _ := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			&core.CardanoTxSuccessProcessorMock{}, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{common.ChainIDStrPrime: primeDB, common.ChainIDStrVector: vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			chainID = common.ChainIDStrPrime
			ttl     = 2
		)

		txHash := indexer.Hash(common.NewHashFromHexString("CC"))

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		err = oracleDB.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainID: chainID, Hash: txHash, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		require.NoError(t, primeDB.OpenTx().AddConfirmedBlock(&indexer.CardanoBlock{Slot: 6, Hash: indexer.Hash{3}}).Execute())

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(5 * time.Second)
		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		expectedTxs, _ := oracleDB.GetAllExpectedTxs(chainID, 0)
		require.NotNil(t, expectedTxs)
	})

	t.Run("Start - expectedTxs - valid - tx not yet expired", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, primeDB, vectorDB := createDbs()

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true, Type: "test"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*cCore.BridgeClaims

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *cCore.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil)

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, _ := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			&core.CardanoTxSuccessProcessorMock{}, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{common.ChainIDStrPrime: primeDB, common.ChainIDStrVector: vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			chainID = common.ChainIDStrPrime
			ttl     = 2
		)

		txHash := indexer.Hash(common.NewHashFromHexString("CCAA"))

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
		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		expectedTxs, _ := oracleDB.GetAllExpectedTxs(chainID, 0)
		require.NotNil(t, expectedTxs)
		require.Nil(t, submittedClaims)
	})

	t.Run("Start - expectedTxs - valid - expired tx", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, primeDB, vectorDB := createDbs()

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true, Type: "test"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*cCore.BridgeClaims

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *cCore.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil)

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, _ := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			&core.CardanoTxSuccessProcessorMock{}, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{common.ChainIDStrPrime: primeDB, common.ChainIDStrVector: vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			chainID = common.ChainIDStrPrime
			ttl     = 2
		)

		txHash := indexer.Hash(common.NewHashFromHexString("CCFF"))

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		err = oracleDB.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainID: chainID, Hash: txHash, TTL: ttl, Metadata: metadata},
		})
		require.NoError(t, err)

		require.NoError(t, primeDB.OpenTx().AddConfirmedBlock(&indexer.CardanoBlock{Slot: 6, Hash: indexer.Hash{3}}).Execute())

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(5 * time.Second)
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

		oracleDB, primeDB, vectorDB := createDbs()

		validTxProc := &core.CardanoTxSuccessProcessorMock{ShouldAddClaim: true, Type: "test"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true, Type: "test"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*cCore.BridgeClaims

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *cCore.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil)

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, rec := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{common.ChainIDStrPrime: primeDB, common.ChainIDStrVector: vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			chainID   = common.ChainIDStrPrime
			ttl       = 2
			blockSlot = 6
		)

		txHash1 := indexer.Hash(common.NewHashFromHexString("CCAA"))
		txHash2 := indexer.Hash(common.NewHashFromHexString("CCFF"))
		blockHash := indexer.Hash(common.NewHashFromHexString("1122"))

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		require.NoError(t, rec.NewUnprocessedTxs(chainID, []*indexer.Tx{
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
		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

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

		validTxProc := &core.CardanoTxSuccessProcessorMock{ShouldAddClaim: true, Type: "test"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true, Type: "test"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*cCore.BridgeClaims

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *cCore.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil)

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, rec := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{common.ChainIDStrPrime: primeDB, common.ChainIDStrVector: vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			chainID   = common.ChainIDStrPrime
			ttl       = 2
			blockSlot = 6
		)

		txHash1 := indexer.Hash(common.NewHashFromHexString("CCAA11"))
		txHash2 := indexer.Hash(common.NewHashFromHexString("CCFF22"))
		blockHash := indexer.Hash(common.NewHashFromHexString("112233"))

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		require.NoError(t, rec.NewUnprocessedTxs(chainID, []*indexer.Tx{
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
		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

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

		validTxProc := &core.CardanoTxSuccessProcessorMock{ShouldAddClaim: true, Type: common.BridgingTxTypeBatchExecution}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true, Type: common.BridgingTxTypeBatchExecution}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*cCore.BridgeClaims

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *cCore.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil)

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, rec := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{common.ChainIDStrPrime: primeDB, common.ChainIDStrVector: vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			chainID   = common.ChainIDStrPrime
			ttl       = 2
			blockSlot = 6
		)

		txHash1 := indexer.Hash(common.NewHashFromHexString("11CCAA"))
		txHash2 := indexer.Hash(common.NewHashFromHexString("11CCFF"))
		blockHash := indexer.Hash(common.NewHashFromHexString("221122"))

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: common.BridgingTxTypeBatchExecution})
		require.NoError(t, err)

		require.NoError(t, rec.NewUnprocessedTxs(chainID, []*indexer.Tx{
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
		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

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

		validTxProc := &core.CardanoTxSuccessProcessorMock{ShouldAddClaim: true, Type: "test"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true, Type: "test"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*cCore.BridgeClaims

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *cCore.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil)

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, rec := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{common.ChainIDStrPrime: primeDB, common.ChainIDStrVector: vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			chainID1  = common.ChainIDStrPrime
			chainID2  = common.ChainIDStrVector
			ttl       = 2
			blockSlot = 6
		)

		txHash1 := indexer.Hash(common.NewHashFromHexString("CCAABB"))
		txHash2 := indexer.Hash(common.NewHashFromHexString("CCFFAA"))
		blockHash := indexer.Hash(common.NewHashFromHexString("112233"))

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		require.NoError(t, rec.NewUnprocessedTxs(chainID1, []*indexer.Tx{
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
		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

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

		validTxProc := &core.CardanoTxSuccessProcessorMock{ShouldAddClaim: true, Type: "test"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true, Type: "test"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*cCore.BridgeClaims

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *cCore.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil)

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, rec := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{common.ChainIDStrPrime: primeDB, common.ChainIDStrVector: vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			chainID1  = common.ChainIDStrPrime
			chainID2  = common.ChainIDStrVector
			ttl       = 2
			blockSlot = 6
		)

		txHash1 := indexer.Hash(common.NewHashFromHexString("CCAABB"))
		txHash2 := indexer.Hash(common.NewHashFromHexString("CCFFAA"))
		blockHash := indexer.Hash(common.NewHashFromHexString("112233"))

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		require.NoError(t, rec.NewUnprocessedTxs(chainID1, []*indexer.Tx{
			{Hash: txHash1, BlockSlot: blockSlot - 1, BlockHash: blockHash, Metadata: metadata},
		}))

		require.NoError(t, rec.NewUnprocessedTxs(chainID1, []*indexer.Tx{
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
		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

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

		validTxProc := &core.CardanoTxSuccessProcessorMock{ShouldAddClaim: true, Type: "test"}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true, Type: "test"}
		failedTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var submittedClaims []*cCore.BridgeClaims

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *cCore.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil)

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, rec := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{common.ChainIDStrPrime: primeDB, common.ChainIDStrVector: vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		const (
			chainID1   = common.ChainIDStrPrime
			chainID2   = common.ChainIDStrVector
			ttl1       = 2
			blockSlot1 = 6
			ttl2       = 10
			blockSlot2 = 15
		)

		txHash1 := indexer.Hash(common.NewHashFromHexString("AACCAABB"))
		txHash2 := indexer.Hash(common.NewHashFromHexString("AACCFFAA"))
		blockHash := indexer.Hash(common.NewHashFromHexString("AA112233"))

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)

		require.NoError(t, rec.NewUnprocessedTxs(chainID1, []*indexer.Tx{
			{Hash: txHash1, BlockSlot: blockSlot1 - 1, BlockHash: blockHash, Metadata: metadata},
		}))

		require.NoError(t, rec.NewUnprocessedTxs(chainID1, []*indexer.Tx{
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
		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			cancelFunc()
		}()

		proc.TickTime = 1
		proc.Start()

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
