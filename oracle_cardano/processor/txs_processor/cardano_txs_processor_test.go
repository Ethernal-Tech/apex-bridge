package processor

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	databaseaccess "github.com/Ethernal-Tech/apex-bridge/oracle_cardano/database_access"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	txsprocessor "github.com/Ethernal-Tech/apex-bridge/oracle_common/processor/txs_processor"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	indexerDb "github.com/Ethernal-Tech/cardano-infrastructure/indexer/db"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethereum_common "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

	cardanoTxsReceiver := NewCardanoTxsReceiverImpl(appConfig, db, txProcessors, bridgingRequestStateUpdater, hclog.NewNullLogger())

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
			{
				Hash:     indexer.Hash{1},
				Metadata: []byte{1, 2, 3},
			},
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
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(&types.Receipt{}, nil)

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
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("test err"))

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
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(&types.Receipt{}, nil)

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
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(&types.Receipt{}, nil)

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
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("test err"))

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
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(&types.Receipt{}, nil)

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
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(&types.Receipt{}, nil)

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
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(&types.Receipt{}, nil)

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
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(&types.Receipt{}, nil)

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
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(&types.Receipt{}, nil)

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
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(&types.Receipt{}, nil)

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
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(&types.Receipt{}, nil)

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
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(&types.Receipt{}, nil)

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

	t.Run("Start - unprocessedTxs - valid brc goes to pending", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, primeDB, vectorDB := createDbs()

		validTxProc := &core.CardanoTxSuccessProcessorMock{ShouldAddClaim: true, Type: common.BridgingTxTypeBridgingRequest}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(&types.Receipt{}, nil)

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
			common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
				BridgingTxType: common.BridgingTxTypeBridgingRequest,
			})
		require.NoError(t, err)

		indexerTx := &indexer.Tx{Hash: txHash, Metadata: metadata}

		require.NoError(t, rec.NewUnprocessedTxs(originChainID, []*indexer.Tx{indexerTx}))

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

		processedTx, _ := oracleDB.GetProcessedTx(originChainID, txHash)
		require.Nil(t, processedTx)

		pendingTxs, _ := oracleDB.GetPendingTxs([][]byte{tx.Key()})
		require.NotNil(t, pendingTxs)
		require.Len(t, pendingTxs, 1)
		require.Equal(t, originChainID, pendingTxs[0].OriginChainID)
		require.Equal(t, tx.Hash, pendingTxs[0].Hash)
	})

	t.Run("Start - unprocessedTxs - valid brc rejected and retry", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, primeDB, vectorDB := createDbs()

		const (
			originChainID = common.ChainIDStrPrime
		)

		txHash := indexer.Hash(common.NewHashFromHexString("0xFFAABB"))

		validTxProc := &core.CardanoTxSuccessProcessorMock{
			AddClaimCallback: func(claims *cCore.BridgeClaims) {
				claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, cCore.BridgingRequestClaim{
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

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything).Return(receipt, nil)

		contract, err := contractbinding.NewBridgeContract(ethereum_common.Address{}, nil)
		require.NoError(t, err)

		event, err := contract.ParseNotEnoughFunds(*receipt.Logs[0])
		require.NoError(t, err)
		require.NotNil(t, event)

		metadata, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
				BridgingTxType: common.BridgingTxTypeBridgingRequest,
			})
		require.NoError(t, err)

		indexerTx := &indexer.Tx{Hash: txHash, Metadata: metadata}

		ctx, cancelFunc := context.WithCancel(context.Background())
		proc, rec := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, nil, bridgeSubmitter,
			map[string]indexer.Database{common.ChainIDStrPrime: primeDB, common.ChainIDStrVector: vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		require.NoError(t, rec.NewUnprocessedTxs(originChainID, []*indexer.Tx{indexerTx}))

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

		processedTx, _ := oracleDB.GetProcessedTx(originChainID, txHash)
		require.Nil(t, processedTx)

		pendingTxs, _ := oracleDB.GetPendingTxs([][]byte{tx.Key()})
		require.Nil(t, pendingTxs)

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
			map[string]indexer.Database{common.ChainIDStrPrime: primeDB, common.ChainIDStrVector: vectorDB},
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
		newTx.LastTimeTried = newTx.LastTimeTried.Add(-cCore.RetryUnprocessedAfterSec * time.Second)

		err = oracleDB.UpdateTxs(&core.CardanoUpdateTxsData{
			UpdateUnprocessed: []*core.CardanoTx{newTx},
		})
		require.NoError(t, err)

		// reset ctx to run again, and confirm by TryCount that this tx was tried again because we simulated time passing
		ctx, cancelFunc = context.WithCancel(context.Background())
		proc, _ = newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, nil, bridgeSubmitter,
			map[string]indexer.Database{common.ChainIDStrPrime: primeDB, common.ChainIDStrVector: vectorDB},
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

	t.Run("Start - BatchExecutionInfoEvent - invalid tx goes to unprocessed/pending", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		oracleDB, primeDB, vectorDB := createDbs()

		originChainID := common.ChainIDStrPrime

		metadata1, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
				BridgingTxType: common.BridgingTxTypeBridgingRequest,
			},
		)
		require.NoError(t, err)

		tx1 := &indexer.Tx{Hash: indexer.Hash(common.NewHashFromHexString("0xFF11223341")), Metadata: metadata1}
		cardanoTx1 := core.CardanoTx{OriginChainID: originChainID, Tx: *tx1}

		metadata2, err := common.SimulateRealMetadata(
			common.MetadataEncodingTypeCbor, common.BridgingRequestMetadata{
				BridgingTxType: common.BridgingTxTypeBridgingRequest,
			},
		)
		require.NoError(t, err)

		tx2 := &indexer.Tx{Hash: indexer.Hash(common.NewHashFromHexString("0xFF11223342")), Metadata: metadata2}
		cardanoTx2 := core.CardanoTx{OriginChainID: originChainID, Tx: *tx2}

		err = oracleDB.AddTxs([]*core.ProcessedCardanoTx{}, []*core.CardanoTx{&cardanoTx1, &cardanoTx2})
		require.NoError(t, err)

		err = oracleDB.UpdateTxs(&cCore.UpdateTxsData[*core.CardanoTx, *core.ProcessedCardanoTx, *core.BridgeExpectedCardanoTx]{
			MoveUnprocessedToPending: []*core.CardanoTx{&cardanoTx1, &cardanoTx2},
		})
		require.NoError(t, err)

		pendingTxs, _ := oracleDB.GetPendingTxs([][]byte{cardanoTx1.Key(), cardanoTx2.Key()})
		require.Len(t, pendingTxs, 2)

		validTxProc := &core.CardanoTxSuccessProcessorMock{
			AddClaimCallback: func(claims *cCore.BridgeClaims) {
				claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, cCore.BridgingRequestClaim{
					ObservedTransactionHash: tx1.Hash,
					SourceChainId:           common.ToNumChainID(originChainID),
				})
				claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, cCore.BridgingRequestClaim{
					ObservedTransactionHash: tx2.Hash,
					SourceChainId:           common.ToNumChainID(originChainID),
				})
			},
			Type: common.BridgingTxTypeBridgingRequest,
		}
		validTxProc.On("ValidateAndAddClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		eventSigs, err := eth.GetSubmitClaimsEventSignatures()
		require.NoError(t, err)

		batchExecFailed := getBatchExecutionReceipt(t, 1, true, common.ChainIDIntPrime,
			[]*contractbinding.IBridgeStructsTxDataInfo{
				{
					SourceChainId:           common.ChainIDIntPrime,
					ObservedTransactionHash: tx1.Hash,
				},
			})

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims", mock.Anything, mock.Anything, mock.Anything).Return(&types.Receipt{
			Logs: []*types.Log{
				{
					Topics: []ethereum_common.Hash{ethereum_common.Hash(eventSigs[1])},
					Data:   batchExecFailed,
				},
			},
		}, nil)

		ctx, canceFunc := context.WithCancel(context.Background())
		proc, rec := newValidProcessor(
			ctx,
			appConfig, oracleDB,
			validTxProc, nil, bridgeSubmitter,
			map[string]indexer.Database{common.ChainIDStrPrime: primeDB, common.ChainIDStrVector: vectorDB},
			&common.BridgingRequestStateUpdaterMock{ReturnNil: true},
		)

		require.NotNil(t, proc)

		require.NoError(t, rec.NewUnprocessedTxs(originChainID, []*indexer.Tx{tx1}))

		go func() {
			<-time.After(time.Millisecond * processingWaitTimeMs)
			canceFunc()
		}()

		proc.TickTime = 1
		proc.Start()

		pendingTxs, _ = oracleDB.GetPendingTxs([][]byte{cardanoTx1.Key(), cardanoTx2.Key()})
		require.Len(t, pendingTxs, 0)

		processedTx1, err := oracleDB.GetProcessedTx(originChainID, cardanoTx1.Hash)
		require.NoError(t, err)
		require.Nil(t, processedTx1)

		processedTx2, err := oracleDB.GetProcessedTx(originChainID, cardanoTx2.Hash)
		require.NoError(t, err)
		require.NotNil(t, processedTx2)
		require.Equal(t, processedTx2.Hash, cardanoTx2.Hash)

		unprocessedTxs, _ := oracleDB.GetAllUnprocessedTxs(originChainID, 0)
		require.Len(t, unprocessedTxs, 1)
		require.Equal(t, unprocessedTxs[0].TryCount, uint32(1))
	})
}

func getBatchExecutionReceipt(
	t *testing.T,
	batchID uint64,
	isFailedTx bool,
	chainID uint8,
	txHashes []*contractbinding.IBridgeStructsTxDataInfo,
) []byte {
	t.Helper()

	events, err := eth.GetSubmitClaimsEventSignatures()
	require.NoError(t, err)

	batchExecInfo := events[1]
	abi, err := contractbinding.BridgeContractMetaData.GetAbi()
	require.NoError(t, err)

	eventAbi, err := abi.EventByID(ethereum_common.Hash(batchExecInfo))
	require.NoError(t, err)

	type TxDataInfo struct {
		SourceChainID           uint8    `json:"sourceChainId" abi:"sourceChainId"`
		ObservedTransactionHash [32]byte `json:"observedTransactionHash" abi:"observedTransactionHash"`
	}

	txDataInfo := make([]TxDataInfo, len(txHashes))

	for idx, info := range txHashes {
		txDataInfo[idx] = TxDataInfo{
			SourceChainID:           info.SourceChainId,
			ObservedTransactionHash: info.ObservedTransactionHash,
		}
	}

	receiptData, err := eventAbi.Inputs.Pack(
		batchID,
		chainID,
		isFailedTx,
		txDataInfo,
	)
	require.NoError(t, err)

	return receiptData
}

var (
	notEnoughFundsEventArguments = abi.Arguments{
		{Name: "claimeType", Type: abi.Type{T: abi.StringTy}},
		{Name: "index", Type: abi.Type{T: abi.UintTy, Size: 256}},
		{Name: "availableAmount", Type: abi.Type{T: abi.UintTy, Size: 256}},
	}
)
