package processor

import (
	"fmt"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/database_access"
	"github.com/Ethernal-Tech/apex-bridge/oracle/utils"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	indexerDb "github.com/Ethernal-Tech/cardano-infrastructure/indexer/db"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func newValidProcessor(
	appConfig *core.AppConfig,
	oracleDb core.Database,
	txProcessor core.CardanoTxProcessor,
	failedTxProcessor core.CardanoTxFailedProcessor,
	bridgeSubmitter core.BridgeSubmitter,
	ccoDbs map[string]indexer.Database,
) *CardanoTxsProcessorImpl {

	txProcessors := []core.CardanoTxProcessor{txProcessor}
	failedTxProcessors := []core.CardanoTxFailedProcessor{failedTxProcessor}

	getCardanoChainObserverDb := func(chainId string) indexer.Database {
		return ccoDbs[chainId]
	}

	cardanoTxsProcessor := NewCardanoTxsProcessor(
		appConfig, oracleDb,
		txProcessors, failedTxProcessors,
		bridgeSubmitter, getCardanoChainObserverDb,
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
		Settings: core.AppSettings{
			MaxBridgingClaimsToGroup: 10,
		},
	}

	appConfig.FillOut()

	const dbFilePath = "temp_test_oracle.db"
	const primeDbFilePath = "temp_test_prime.db"
	const vectorDbFilePath = "temp_test_vector.db"

	createDbs := func() (core.Database, indexer.Database, indexer.Database) {
		oracleDb, _ := database_access.NewDatabase(dbFilePath)
		primeDb, _ := indexerDb.NewDatabaseInit("", primeDbFilePath)
		vectorDb, _ := indexerDb.NewDatabaseInit("", vectorDbFilePath)

		return oracleDb, primeDb, vectorDb
	}

	dbCleanup := func() {
		utils.RemoveDirOrFilePathIfExists(dbFilePath)
		utils.RemoveDirOrFilePathIfExists(primeDbFilePath)
		utils.RemoveDirOrFilePathIfExists(vectorDbFilePath)
	}

	t.Cleanup(dbCleanup)

	t.Run("NewCardanoTxsProcessor", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		oracleDb, primeDb, vectorDb := createDbs()

		proc := NewCardanoTxsProcessor(nil, nil, nil, nil, nil, nil, nil)
		require.NotNil(t, proc)

		getCardanoChainObserverDb := func(chainId string) indexer.Database {
			return (map[string]indexer.Database{"prime": primeDb, "vector": vectorDb})[chainId]
		}

		proc = NewCardanoTxsProcessor(
			appConfig,
			oracleDb,
			[]core.CardanoTxProcessor{},
			[]core.CardanoTxFailedProcessor{},
			&core.BridgeSubmitterMock{}, getCardanoChainObserverDb, hclog.NewNullLogger(),
		)
		require.NotNil(t, proc)
	})

	t.Run("NewUnprocessedTxs nil txs", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		oracleDb, primeDb, vectorDb := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{}
		failedTxProc := &core.CardanoTxFailedProcessorMock{}
		bridgeSubmitter := &core.BridgeSubmitterMock{}

		proc := newValidProcessor(
			appConfig, oracleDb,
			validTxProc, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDb, "vector": vectorDb},
		)

		require.NotNil(t, proc)

		proc.NewUnprocessedTxs("prime", nil)

		unprocessedTxs, err := oracleDb.GetUnprocessedTxs("prime", 0)
		require.NoError(t, err)
		require.Nil(t, unprocessedTxs)
	})

	t.Run("NewUnprocessedTxs no txs", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		oracleDb, primeDb, vectorDb := createDbs()

		proc := newValidProcessor(
			appConfig, oracleDb,
			nil, nil, nil,
			map[string]indexer.Database{"prime": primeDb, "vector": vectorDb},
		)

		require.NotNil(t, proc)

		proc.NewUnprocessedTxs("prime", []*indexer.Tx{})

		unprocessedTxs, err := oracleDb.GetUnprocessedTxs("prime", 0)
		require.NoError(t, err)
		require.Nil(t, unprocessedTxs)
	})

	t.Run("NewUnprocessedTxs no relevant txs", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		oracleDb, primeDb, vectorDb := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{}
		validTxProc.On("IsTxRelevant").Return(false, nil)

		proc := newValidProcessor(
			appConfig, oracleDb,
			validTxProc, nil, nil,
			map[string]indexer.Database{"prime": primeDb, "vector": vectorDb},
		)

		require.NotNil(t, proc)

		proc.NewUnprocessedTxs("prime", []*indexer.Tx{
			{Hash: "test_hash"},
		})

		unprocessedTxs, err := oracleDb.GetUnprocessedTxs("prime", 0)
		require.NoError(t, err)
		require.Nil(t, unprocessedTxs)
	})

	t.Run("NewUnprocessedTxs invalid txs", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		oracleDb, primeDb, vectorDb := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{}
		validTxProc.On("IsTxRelevant").Return(false, fmt.Errorf("test err"))

		proc := newValidProcessor(
			appConfig, oracleDb,
			validTxProc, nil, nil,
			map[string]indexer.Database{"prime": primeDb, "vector": vectorDb},
		)

		require.NotNil(t, proc)

		proc.NewUnprocessedTxs("prime", []*indexer.Tx{
			{Hash: "test_hash"},
		})

		unprocessedTxs, err := oracleDb.GetUnprocessedTxs("prime", 0)
		require.NoError(t, err)
		require.Nil(t, unprocessedTxs)
	})

	t.Run("NewUnprocessedTxs valid", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		oracleDb, primeDb, vectorDb := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{}
		validTxProc.On("IsTxRelevant").Return(true, nil)

		proc := newValidProcessor(
			appConfig, oracleDb,
			validTxProc, nil, nil,
			map[string]indexer.Database{"prime": primeDb, "vector": vectorDb},
		)

		require.NotNil(t, proc)

		const originChainId = "prime"
		const txHash = "test_hash"
		proc.NewUnprocessedTxs(originChainId, []*indexer.Tx{
			{Hash: txHash},
		})

		unprocessedTxs, err := oracleDb.GetUnprocessedTxs(originChainId, 0)
		require.NoError(t, err)
		require.Len(t, unprocessedTxs, 1)
		require.Equal(t, txHash, unprocessedTxs[0].Hash)
		require.Equal(t, originChainId, unprocessedTxs[0].OriginChainId)
	})

	t.Run("Start - unprocessedTxs - tx validation err", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		oracleDb, primeDb, vectorDb := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{}
		validTxProc.On("IsTxRelevant").Return(true, nil)
		validTxProc.On("ValidateAndAddClaim").Return(fmt.Errorf("test err"))

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims").Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDb,
			validTxProc, nil, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDb, "vector": vectorDb},
		)

		require.NotNil(t, proc)

		const originChainId = "prime"
		const txHash = "test_hash"
		proc.NewUnprocessedTxs(originChainId, []*indexer.Tx{
			{Hash: txHash},
		})

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(5 * time.Second)
		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		unprocessedTxs, _ := oracleDb.GetUnprocessedTxs(originChainId, 0)
		require.Nil(t, unprocessedTxs)
		processedTx, _ := oracleDb.GetProcessedTx(originChainId, txHash)
		require.NotNil(t, processedTx)
		require.True(t, processedTx.IsInvalid)
	})

	t.Run("Start - unprocessedTxs - submit claims failed", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		oracleDb, primeDb, vectorDb := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{ShouldAddClaim: true}
		validTxProc.On("IsTxRelevant").Return(true, nil)
		validTxProc.On("ValidateAndAddClaim").Return(nil)

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims").Return(fmt.Errorf("test err"))

		proc := newValidProcessor(
			appConfig, oracleDb,
			validTxProc, nil, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDb, "vector": vectorDb},
		)

		require.NotNil(t, proc)

		const originChainId = "prime"
		const txHash = "test_hash"
		proc.NewUnprocessedTxs(originChainId, []*indexer.Tx{
			{Hash: txHash},
		})

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(5 * time.Second)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		unprocessedTxs, _ := oracleDb.GetUnprocessedTxs(originChainId, 0)
		require.Len(t, unprocessedTxs, 1)
		require.Equal(t, txHash, unprocessedTxs[0].Hash)
		require.Equal(t, originChainId, unprocessedTxs[0].OriginChainId)
		processedTx, _ := oracleDb.GetProcessedTx(originChainId, txHash)
		require.Nil(t, processedTx)
	})

	t.Run("Start - unprocessedTxs - valid", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		oracleDb, primeDb, vectorDb := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{ShouldAddClaim: true}
		validTxProc.On("IsTxRelevant").Return(true, nil)
		validTxProc.On("ValidateAndAddClaim").Return(nil)

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims").Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDb,
			validTxProc, nil, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDb, "vector": vectorDb},
		)

		require.NotNil(t, proc)

		const originChainId = "prime"
		const txHash = "test_hash"
		proc.NewUnprocessedTxs(originChainId, []*indexer.Tx{
			{Hash: txHash},
		})

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(5 * time.Second)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		unprocessedTxs, _ := oracleDb.GetUnprocessedTxs(originChainId, 0)
		require.Nil(t, unprocessedTxs)
		processedTx, _ := oracleDb.GetProcessedTx(originChainId, txHash)
		require.NotNil(t, processedTx)
		require.False(t, processedTx.IsInvalid)
	})

	t.Run("Start - expectedTxs - tx validation err", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		oracleDb, primeDb, vectorDb := createDbs()

		failedTxProc := &core.CardanoTxFailedProcessorMock{}
		failedTxProc.On("IsTxRelevant").Return(true, nil)
		failedTxProc.On("ValidateAndAddClaim").Return(fmt.Errorf("test err"))

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims").Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDb,
			nil, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDb, "vector": vectorDb},
		)

		require.NotNil(t, proc)

		const chainId = "prime"
		const txHash = "test_hash"
		const ttl = 2

		err := oracleDb.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainId: chainId, Hash: txHash, Ttl: ttl},
		})
		require.NoError(t, err)

		primeDb.OpenTx().AddConfirmedBlock(&indexer.CardanoBlock{Slot: 6, Hash: "test_block_hash"}).Execute()

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(5 * time.Second)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		expectedTxs, _ := oracleDb.GetExpectedTxs(chainId, 0)
		require.Nil(t, expectedTxs)
	})

	t.Run("Start - expectedTxs - submit claims failed", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		oracleDb, primeDb, vectorDb := createDbs()

		failedTxProc := &core.CardanoTxFailedProcessorMock{}
		failedTxProc.On("IsTxRelevant").Return(true, nil)
		failedTxProc.On("ValidateAndAddClaim").Return(nil)

		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims").Return(fmt.Errorf("test err"))

		proc := newValidProcessor(
			appConfig, oracleDb,
			nil, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDb, "vector": vectorDb},
		)

		require.NotNil(t, proc)

		const chainId = "prime"
		const txHash = "test_hash"
		const ttl = 2

		err := oracleDb.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainId: chainId, Hash: txHash, Ttl: ttl},
		})
		require.NoError(t, err)

		primeDb.OpenTx().AddConfirmedBlock(&indexer.CardanoBlock{Slot: 6, Hash: "test_block_hash"}).Execute()

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(5 * time.Second)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		expectedTxs, _ := oracleDb.GetExpectedTxs(chainId, 0)
		require.NotNil(t, expectedTxs)
	})

	t.Run("Start - expectedTxs - valid - tx not yet expired", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		oracleDb, primeDb, vectorDb := createDbs()

		failedTxProc := &core.CardanoTxFailedProcessorMock{}
		failedTxProc.On("IsTxRelevant").Return(true, nil)
		failedTxProc.On("ValidateAndAddClaim").Return(nil)

		var submittedClaims []*core.BridgeClaims
		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *core.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims").Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDb,
			nil, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDb, "vector": vectorDb},
		)

		require.NotNil(t, proc)

		const chainId = "prime"
		const txHash = "test_hash"
		const ttl = 2

		err := oracleDb.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainId: chainId, Hash: txHash, Ttl: ttl},
		})
		require.NoError(t, err)

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(5 * time.Second)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		expectedTxs, _ := oracleDb.GetExpectedTxs(chainId, 0)
		require.NotNil(t, expectedTxs)
		require.Nil(t, submittedClaims)
	})

	t.Run("Start - expectedTxs - valid - expired tx", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		oracleDb, primeDb, vectorDb := createDbs()

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true}
		failedTxProc.On("IsTxRelevant").Return(true, nil)
		failedTxProc.On("ValidateAndAddClaim").Return(nil)

		var submittedClaims []*core.BridgeClaims
		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *core.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims").Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDb,
			nil, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDb, "vector": vectorDb},
		)

		require.NotNil(t, proc)

		const chainId = "prime"
		const txHash = "test_hash"
		const ttl = 2

		err := oracleDb.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainId: chainId, Hash: txHash, Ttl: ttl},
		})
		require.NoError(t, err)

		primeDb.OpenTx().AddConfirmedBlock(&indexer.CardanoBlock{Slot: 6, Hash: "test_block_hash"}).Execute()

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(5 * time.Second)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		expectedTxs, _ := oracleDb.GetExpectedTxs(chainId, 0)
		require.Nil(t, expectedTxs)
		require.NotNil(t, submittedClaims)
		require.Len(t, submittedClaims, 1)
		require.Len(t, submittedClaims[0].BatchExecutionFailedClaims, 1)
	})

	t.Run("Start - unprocessedTxs, expectedTxs - single chain - valid 1", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		oracleDb, primeDb, vectorDb := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{ShouldAddClaim: true}
		validTxProc.On("IsTxRelevant").Return(true, nil)
		validTxProc.On("ValidateAndAddClaim").Return(nil)

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true}
		failedTxProc.On("IsTxRelevant").Return(true, nil)
		failedTxProc.On("ValidateAndAddClaim").Return(nil)

		var submittedClaims []*core.BridgeClaims
		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *core.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims").Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDb,
			validTxProc, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDb, "vector": vectorDb},
		)

		require.NotNil(t, proc)

		const chainId = "prime"
		const txHash1 = "test_hash_1"
		const txHash2 = "test_hash_2"
		const ttl = 2
		const blockSlot = 6
		const blockHash = "test_block_hash"

		proc.NewUnprocessedTxs(chainId, []*indexer.Tx{
			{Hash: txHash1, BlockSlot: blockSlot, BlockHash: blockHash},
		})

		err := oracleDb.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainId: chainId, Hash: txHash2, Ttl: ttl},
		})
		require.NoError(t, err)

		primeDb.OpenTx().AddConfirmedBlock(&indexer.CardanoBlock{Slot: blockSlot, Hash: blockHash}).Execute()

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(5 * time.Second)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		unprocessedTxs, _ := oracleDb.GetUnprocessedTxs(chainId, 0)
		require.Nil(t, unprocessedTxs)
		processedTx, _ := oracleDb.GetProcessedTx(chainId, txHash1)
		require.NotNil(t, processedTx)
		require.False(t, processedTx.IsInvalid)

		expectedTxs, _ := oracleDb.GetExpectedTxs(chainId, 0)
		require.Nil(t, expectedTxs)

		require.NotNil(t, submittedClaims)
		require.Len(t, submittedClaims, 1)
		require.Len(t, submittedClaims[0].BridgingRequestClaims, 1)
		require.Len(t, submittedClaims[0].BatchExecutionFailedClaims, 1)
	})

	t.Run("Start - unprocessedTxs, expectedTxs - single chain - valid 2", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		oracleDb, primeDb, vectorDb := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{ShouldAddClaim: true}
		validTxProc.On("IsTxRelevant").Return(true, nil)
		validTxProc.On("ValidateAndAddClaim").Return(nil)

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true}
		failedTxProc.On("IsTxRelevant").Return(true, nil)
		failedTxProc.On("ValidateAndAddClaim").Return(nil)

		var submittedClaims []*core.BridgeClaims
		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *core.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims").Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDb,
			validTxProc, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDb, "vector": vectorDb},
		)

		require.NotNil(t, proc)

		const chainId = "prime"
		const txHash1 = "test_hash_1"
		const txHash2 = "test_hash_2"
		const ttl = 2
		const blockSlot = 6
		const blockHash = "test_block_hash"

		proc.NewUnprocessedTxs(chainId, []*indexer.Tx{
			{Hash: txHash1, BlockSlot: blockSlot + 1, BlockHash: blockHash},
		})

		err := oracleDb.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainId: chainId, Hash: txHash2, Ttl: ttl},
		})
		require.NoError(t, err)

		primeDb.OpenTx().AddConfirmedBlock(&indexer.CardanoBlock{Slot: blockSlot, Hash: blockHash}).Execute()

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(10 * time.Second)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		unprocessedTxs, _ := oracleDb.GetUnprocessedTxs(chainId, 0)
		require.Nil(t, unprocessedTxs)
		processedTx, _ := oracleDb.GetProcessedTx(chainId, txHash1)
		require.NotNil(t, processedTx)
		require.False(t, processedTx.IsInvalid)

		expectedTxs, _ := oracleDb.GetExpectedTxs(chainId, 0)
		require.Nil(t, expectedTxs)

		require.NotNil(t, submittedClaims)
		require.Len(t, submittedClaims, 2)
		require.Len(t, submittedClaims[0].BatchExecutionFailedClaims, 1)
		require.Len(t, submittedClaims[1].BridgingRequestClaims, 1)
	})

	t.Run("Start - unprocessedTxs, expectedTxs - single chain - valid 3", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		oracleDb, primeDb, vectorDb := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{ShouldAddClaim: true}
		validTxProc.On("IsTxRelevant").Return(true, nil)
		validTxProc.On("ValidateAndAddClaim").Return(nil)

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true}
		failedTxProc.On("IsTxRelevant").Return(true, nil)
		failedTxProc.On("ValidateAndAddClaim").Return(nil)

		var submittedClaims []*core.BridgeClaims
		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *core.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims").Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDb,
			validTxProc, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDb, "vector": vectorDb},
		)

		require.NotNil(t, proc)

		const chainId = "prime"
		const txHash1 = "test_hash_1"
		const txHash2 = "test_hash_2"
		const ttl = 2
		const blockSlot = 6
		const blockHash = "test_block_hash"

		proc.NewUnprocessedTxs(chainId, []*indexer.Tx{
			{Hash: txHash1, BlockSlot: blockSlot - 1, BlockHash: blockHash},
		})

		err := oracleDb.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainId: chainId, Hash: txHash2, Ttl: ttl},
		})
		require.NoError(t, err)

		primeDb.OpenTx().AddConfirmedBlock(&indexer.CardanoBlock{Slot: blockSlot, Hash: blockHash}).Execute()

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(12 * time.Second)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		unprocessedTxs, _ := oracleDb.GetUnprocessedTxs(chainId, 0)
		require.Nil(t, unprocessedTxs)
		processedTx, _ := oracleDb.GetProcessedTx(chainId, txHash1)
		require.NotNil(t, processedTx)
		require.False(t, processedTx.IsInvalid)

		expectedTxs, _ := oracleDb.GetExpectedTxs(chainId, 0)
		require.Nil(t, expectedTxs)

		require.NotNil(t, submittedClaims)
		require.Len(t, submittedClaims, 2)
		require.Len(t, submittedClaims[0].BridgingRequestClaims, 1)
		require.Len(t, submittedClaims[1].BatchExecutionFailedClaims, 1)
	})

	t.Run("Start - unprocessedTxs, expectedTxs - single chain - valid 4", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		oracleDb, primeDb, vectorDb := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{ShouldAddClaim: true}
		validTxProc.On("IsTxRelevant").Return(true, nil)
		validTxProc.On("ValidateAndAddClaim").Return(nil)

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true}
		failedTxProc.On("IsTxRelevant").Return(true, nil)
		failedTxProc.On("ValidateAndAddClaim").Return(nil)

		var submittedClaims []*core.BridgeClaims
		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *core.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims").Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDb,
			validTxProc, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDb, "vector": vectorDb},
		)

		require.NotNil(t, proc)

		const chainId = "prime"
		const txHash1 = "test_hash_1"
		const txHash2 = "test_hash_2"
		const ttl = 2
		const blockSlot = 6
		const blockHash = "test_block_hash"

		proc.NewUnprocessedTxs(chainId, []*indexer.Tx{
			{Hash: txHash1, BlockSlot: blockSlot - 1, BlockHash: blockHash},
		})

		err := oracleDb.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainId: chainId, Hash: txHash1, Ttl: blockSlot + 2},
		})
		require.NoError(t, err)

		err = oracleDb.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainId: chainId, Hash: txHash2, Ttl: ttl},
		})
		require.NoError(t, err)

		primeDb.OpenTx().AddConfirmedBlock(&indexer.CardanoBlock{Slot: blockSlot, Hash: blockHash}).Execute()

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(12 * time.Second)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		unprocessedTxs, _ := oracleDb.GetUnprocessedTxs(chainId, 0)
		require.Nil(t, unprocessedTxs)
		processedTx, _ := oracleDb.GetProcessedTx(chainId, txHash1)
		require.NotNil(t, processedTx)
		require.False(t, processedTx.IsInvalid)

		expectedTxs, _ := oracleDb.GetExpectedTxs(chainId, 0)
		require.Nil(t, expectedTxs)

		require.NotNil(t, submittedClaims)
		require.Len(t, submittedClaims, 2)
		require.Len(t, submittedClaims[0].BridgingRequestClaims, 1)
		require.Len(t, submittedClaims[1].BatchExecutionFailedClaims, 1)
	})

	t.Run("Start - unprocessedTxs, expectedTxs - multiple chains - valid 1", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		dbCleanup()
		oracleDb, primeDb, vectorDb := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{ShouldAddClaim: true}
		validTxProc.On("IsTxRelevant").Return(true, nil)
		validTxProc.On("ValidateAndAddClaim").Return(nil)

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true}
		failedTxProc.On("IsTxRelevant").Return(true, nil)
		failedTxProc.On("ValidateAndAddClaim").Return(nil)

		var submittedClaims []*core.BridgeClaims
		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *core.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims").Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDb,
			validTxProc, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDb, "vector": vectorDb},
		)

		require.NotNil(t, proc)

		const chainId1 = "prime"
		const chainId2 = "vector"
		const txHash1 = "test_hash_1"
		const txHash2 = "test_hash_2"
		const ttl = 2
		const blockSlot = 6
		const blockHash = "test_block_hash"

		proc.NewUnprocessedTxs(chainId1, []*indexer.Tx{
			{Hash: txHash1, BlockSlot: blockSlot - 1, BlockHash: blockHash},
		})

		err := oracleDb.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainId: chainId2, Hash: txHash2, Ttl: ttl},
		})
		require.NoError(t, err)

		vectorDb.OpenTx().AddConfirmedBlock(&indexer.CardanoBlock{Slot: blockSlot, Hash: blockHash}).Execute()

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(12 * time.Second)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		unprocessedTxs, _ := oracleDb.GetUnprocessedTxs(chainId1, 0)
		require.Nil(t, unprocessedTxs)
		processedTx, _ := oracleDb.GetProcessedTx(chainId1, txHash1)
		require.NotNil(t, processedTx)
		require.False(t, processedTx.IsInvalid)

		expectedTxs, _ := oracleDb.GetExpectedTxs(chainId2, 0)
		require.Nil(t, expectedTxs)

		require.NotNil(t, submittedClaims)
		require.Len(t, submittedClaims, 2)
		require.Len(t, submittedClaims[0].BridgingRequestClaims, 1)
		require.Len(t, submittedClaims[1].BatchExecutionFailedClaims, 1)
	})

	t.Run("Start - unprocessedTxs, expectedTxs - multiple chains - valid 2", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		dbCleanup()
		oracleDb, primeDb, vectorDb := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{ShouldAddClaim: true}
		validTxProc.On("IsTxRelevant").Return(true, nil)
		validTxProc.On("ValidateAndAddClaim").Return(nil)

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true}
		failedTxProc.On("IsTxRelevant").Return(true, nil)
		failedTxProc.On("ValidateAndAddClaim").Return(nil)

		var submittedClaims []*core.BridgeClaims
		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *core.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims").Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDb,
			validTxProc, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDb, "vector": vectorDb},
		)

		require.NotNil(t, proc)

		const chainId1 = "prime"
		const chainId2 = "vector"
		const txHash1 = "test_hash_1"
		const txHash2 = "test_hash_2"
		const ttl = 2
		const blockSlot = 6
		const blockHash = "test_block_hash"

		proc.NewUnprocessedTxs(chainId1, []*indexer.Tx{
			{Hash: txHash1, BlockSlot: blockSlot - 1, BlockHash: blockHash},
		})
		proc.NewUnprocessedTxs(chainId1, []*indexer.Tx{
			{Hash: txHash2, BlockSlot: blockSlot - 1, BlockHash: blockHash},
		})

		err := oracleDb.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainId: chainId2, Hash: txHash1, Ttl: ttl},
		})
		require.NoError(t, err)
		err = oracleDb.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainId: chainId2, Hash: txHash2, Ttl: ttl},
		})
		require.NoError(t, err)

		vectorDb.OpenTx().AddConfirmedBlock(
			&indexer.CardanoBlock{Slot: blockSlot, Hash: blockHash},
		).Execute()

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(12 * time.Second)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		unprocessedTxs, _ := oracleDb.GetUnprocessedTxs(chainId1, 0)
		require.Nil(t, unprocessedTxs)
		processedTx, _ := oracleDb.GetProcessedTx(chainId1, txHash1)
		require.NotNil(t, processedTx)
		require.False(t, processedTx.IsInvalid)

		expectedTxs, _ := oracleDb.GetExpectedTxs(chainId2, 0)
		require.Nil(t, expectedTxs)

		require.NotNil(t, submittedClaims)
		require.Len(t, submittedClaims, 2)

		require.Len(t, submittedClaims[0].BridgingRequestClaims, 2)
		require.Len(t, submittedClaims[1].BatchExecutionFailedClaims, 2)
	})

	t.Run("Start - unprocessedTxs, expectedTxs - multiple chains - valid 3", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		dbCleanup()
		oracleDb, primeDb, vectorDb := createDbs()

		validTxProc := &core.CardanoTxProcessorMock{ShouldAddClaim: true}
		validTxProc.On("IsTxRelevant").Return(true, nil)
		validTxProc.On("ValidateAndAddClaim").Return(nil)

		failedTxProc := &core.CardanoTxFailedProcessorMock{ShouldAddClaim: true}
		failedTxProc.On("IsTxRelevant").Return(true, nil)
		failedTxProc.On("ValidateAndAddClaim").Return(nil)

		var submittedClaims []*core.BridgeClaims
		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.OnSubmitClaims = func(claims *core.BridgeClaims) {
			submittedClaims = append(submittedClaims, claims)
		}
		bridgeSubmitter.On("Dispose").Return(nil)
		bridgeSubmitter.On("SubmitClaims").Return(nil)

		proc := newValidProcessor(
			appConfig, oracleDb,
			validTxProc, failedTxProc, bridgeSubmitter,
			map[string]indexer.Database{"prime": primeDb, "vector": vectorDb},
		)

		require.NotNil(t, proc)

		const chainId1 = "prime"
		const chainId2 = "vector"
		const txHash1 = "test_hash_1"
		const txHash2 = "test_hash_2"
		const ttl1 = 2
		const blockSlot1 = 6
		const ttl2 = 10
		const blockSlot2 = 15
		const blockHash = "test_block_hash"

		proc.NewUnprocessedTxs(chainId1, []*indexer.Tx{
			{Hash: txHash1, BlockSlot: blockSlot1 - 1, BlockHash: blockHash},
		})
		proc.NewUnprocessedTxs(chainId1, []*indexer.Tx{
			{Hash: txHash2, BlockSlot: blockSlot1, BlockHash: blockHash},
		})

		err := oracleDb.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainId: chainId2, Hash: txHash1, Ttl: ttl1},
		})
		require.NoError(t, err)
		err = oracleDb.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainId: chainId2, Hash: txHash2, Ttl: ttl2},
		})
		require.NoError(t, err)

		vectorDb.OpenTx().AddConfirmedBlock(
			&indexer.CardanoBlock{Slot: blockSlot1, Hash: blockHash},
		).AddConfirmedBlock(
			&indexer.CardanoBlock{Slot: blockSlot2, Hash: blockHash},
		).Execute()

		// go proc.Start()
		// defer proc.Stop()
		// time.Sleep(12 * time.Second)

		proc.tickTime = 1
		for i := 0; i < 5; i++ {
			proc.checkShouldGenerateClaims()
		}

		unprocessedTxs, _ := oracleDb.GetUnprocessedTxs(chainId1, 0)
		require.Nil(t, unprocessedTxs)
		processedTx, _ := oracleDb.GetProcessedTx(chainId1, txHash1)
		require.NotNil(t, processedTx)
		require.False(t, processedTx.IsInvalid)

		expectedTxs, _ := oracleDb.GetExpectedTxs(chainId2, 0)
		require.Nil(t, expectedTxs)

		require.NotNil(t, submittedClaims)
		require.Len(t, submittedClaims, 4)

		require.Len(t, submittedClaims[0].BridgingRequestClaims, 1)
		require.Len(t, submittedClaims[1].BatchExecutionFailedClaims, 1)
		require.Len(t, submittedClaims[2].BridgingRequestClaims, 1)
		require.Len(t, submittedClaims[3].BatchExecutionFailedClaims, 1)
	})
}
