package chain

import (
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/utils"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/stretchr/testify/require"
)

func TestCardanoChainObserver(t *testing.T) {
	settings := core.AppSettings{
		DbsPath:  "./tests_temp/",
		LogsPath: "./tests_temp/",
		LogLevel: 2,
	}

	foldersCleanup := func() {
		utils.RemoveDirOrFilePathIfExists(settings.DbsPath)
		utils.RemoveDirOrFilePathIfExists(settings.LogsPath)
	}

	t.Cleanup(foldersCleanup)

	chainConfig := core.CardanoChainConfig{
		ChainId:                "prime",
		NetworkAddress:         "backbone.cardano-mainnet.iohk.io:3001",
		NetworkMagic:           "764824073",
		StartBlockHash:         "df12b7a87cc041f238f400e3a410d1edb2f4a6fbcbedafff8fd9d507ef4947d7",
		StartSlot:              "76593517",
		StartBlockNumber:       "8000030",
		ConfirmationBlockCount: 10,
		OtherAddressesOfInterest: []string{
			"addr1v9kganeshgdqyhwnyn9stxxgl7r4y2ejfyqjn88n7ncapvs4sugsd",
		},
	}

	t.Run("check ErrorCh", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		chainObserver := NewCardanoChainObserver(settings, chainConfig, []*indexer.TxInputOutput{}, &core.CardanoTxsProcessorMock{})
		require.NotNil(t, chainObserver)

		errChan := chainObserver.ErrorCh()
		require.NotNil(t, errChan)
	})

	t.Run("check GetConfig", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		chainObserver := NewCardanoChainObserver(settings, chainConfig, []*indexer.TxInputOutput{}, &core.CardanoTxsProcessorMock{})
		require.NotNil(t, chainObserver)

		config := chainObserver.GetConfig()
		require.NotNil(t, config)
		require.Equal(t, chainConfig, config)
	})

	t.Run("check GetDb", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		chainObserver := NewCardanoChainObserver(settings, chainConfig, []*indexer.TxInputOutput{}, &core.CardanoTxsProcessorMock{})
		require.NotNil(t, chainObserver)

		db := chainObserver.GetDb()
		require.NotNil(t, db)
	})

	t.Run("check start stop", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		chainObserver := NewCardanoChainObserver(settings, chainConfig, []*indexer.TxInputOutput{}, &core.CardanoTxsProcessorMock{})
		require.NotNil(t, chainObserver)

		err := chainObserver.Start()
		require.NoError(t, err)

		err = chainObserver.Stop()
		require.NoError(t, err)
	})

	t.Run("check newConfirmedTxs called", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		txsProcessor := &core.CardanoTxsProcessorMock{}
		txsProcessor.On("NewUnprocessedTxs").Return()

		chainObserver := NewCardanoChainObserver(settings, chainConfig, []*indexer.TxInputOutput{}, txsProcessor)
		require.NotNil(t, chainObserver)

		doneCh := make(chan bool, 1)

		txsProcessor.NewUnprocessedTxsFn = func(originChainId string, txs []*indexer.Tx) error {
			t.Helper()
			close(doneCh)
			return nil
		}

		err := chainObserver.Start()
		require.NoError(t, err)

		defer chainObserver.Stop()

		timer := time.NewTimer(60 * time.Second)
		defer timer.Stop()

		select {
		case <-timer.C:
			t.Fail()
			close(doneCh)
		case <-doneCh:
		}
	})
}
