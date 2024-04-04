package bridge

import (
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/utils"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestConfirmedBlocksSubmitter(t *testing.T) {
	appConfig := &core.AppConfig{ // TODO
		Bridge: core.BridgeConfig{
			SubmitConfig: core.SubmitConfig{
				ConfirmedBlocksThreshhold: 10,
				ConfirmedBlocksSubmitTime: 10,
			},
		},
		Settings: core.AppSettings{
			DbsPath: "./tests_temp/",
		},
	}

	foldersCleanup := func() {
		utils.RemoveDirOrFilePathIfExists(appConfig.Settings.DbsPath)
	}

	t.Cleanup(foldersCleanup)

	t.Run("start submit", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		bridgeSubmitter := core.BridgeSubmitterMock{}
		bridgeSubmitter.On("SubmitConfirmedBlocks").Return(nil)
		db := &core.CardanoTxsProcessorDbMock{}
		db.On("GetProcessedTx").Return(nil, nil)

		blocksSubmitter, err := NewConfirmedBlocksSubmitter(&bridgeSubmitter, appConfig, db, "prime", hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, blocksSubmitter)

		blocksSubmitter.StartSubmit()

		time.Sleep(time.Millisecond * 100)

		blocksSubmitter.Dispose()
		require.NoError(t, <-blocksSubmitter.ErrorCh())
	})

	t.Run("dispose", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		blocksSubmitter, err := NewConfirmedBlocksSubmitter(nil, appConfig, &core.CardanoTxsProcessorDbMock{}, "prime", hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, blocksSubmitter)

		err = blocksSubmitter.Dispose()
		require.NoError(t, err)
	})
}
