package bridge

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	indexerDb "github.com/Ethernal-Tech/cardano-infrastructure/indexer/db"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestConfirmedBlocksSubmitter(t *testing.T) {
	testDir, err := os.MkdirTemp("", "block-submitter")
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	chainId := "prime"
	appConfig := &core.AppConfig{
		Bridge: core.BridgeConfig{
			SubmitConfig: core.SubmitConfig{
				ConfirmedBlocksThreshold:  10,
				ConfirmedBlocksSubmitTime: 10,
			},
		},
		Settings: core.AppSettings{
			DbsPath: testDir,
		},
	}

	foldersCleanup := func() {
		common.RemoveDirOrFilePathIfExists(appConfig.Settings.DbsPath)
	}

	initDb := func() (indexer.Database, error) {
		if err := common.CreateDirectoryIfNotExists(appConfig.Settings.DbsPath, 0750); err != nil {
			return nil, fmt.Errorf("failed to create db dir")
		}
		indexerDb, err := indexerDb.NewDatabaseInit("", path.Join(appConfig.Settings.DbsPath, chainId+".db"))
		if err != nil {
			return nil, fmt.Errorf("failed to open db")
		}

		return indexerDb, err
	}

	t.Run("start submit", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		bridgeSubmitter := core.BridgeSubmitterMock{}
		bridgeSubmitter.On("SubmitConfirmedBlocks").Return(nil)
		db := &core.CardanoTxsProcessorDbMock{}
		db.On("GetProcessedTx").Return(nil, nil)

		indexerDb, err := initDb()
		require.NoError(t, err)

		blocksSubmitter, err := NewConfirmedBlocksSubmitter(&bridgeSubmitter, appConfig, db, indexerDb, chainId, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, blocksSubmitter)

		blocksSubmitter.StartSubmit()

		time.Sleep(time.Millisecond * 100)

		blocksSubmitter.Dispose()
		require.NoError(t, <-blocksSubmitter.ErrorCh())
	})

	t.Run("dispose", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		indexerDb, err := initDb()
		require.NoError(t, err)

		blocksSubmitter, err := NewConfirmedBlocksSubmitter(nil, appConfig, &core.CardanoTxsProcessorDbMock{}, indexerDb, chainId, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, blocksSubmitter)

		err = blocksSubmitter.Dispose()
		require.NoError(t, err)
	})
}
