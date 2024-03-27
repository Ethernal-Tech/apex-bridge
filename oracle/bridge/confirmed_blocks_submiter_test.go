package bridge

import (
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/utils"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestConfirmedBlocksSubmitter(t *testing.T) {
	appConfig := &core.AppConfig{ // TODO
		Bridge: core.BridgeConfig{
			NodeUrl:              "https://polygon-mumbai-pokt.nodies.app",
			SmartContractAddress: "0xb2B87f7e652Aa847F98Cc05e130d030b91c7B37d",
			SigningKey:           "93c91e490bfd3736d17d04f53a10093e9cf2435309f4be1f5751381c8e201d23",
			SubmitConfig: core.SubmitConfig{
				ConfirmedBlocksThreshhold: 10,
				ConfirmedBlocksSubmitTime: 3000,
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

		blocksSubmitter, err := NewConfirmedBlocksSubmitter(appConfig, "prime", hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, blocksSubmitter)

		err = blocksSubmitter.StartSubmit()

		time.Sleep(time.Second * 3)
		// ethTxHelper.SendTx#132 is returning error when sending Tx, can be ignored for now
		require.NoError(t, <-blocksSubmitter.ErrorCh())

		blocksSubmitter.Dispose()
		require.NoError(t, err)
	})

	t.Run("dispose", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		blocksSubmitter, err := NewConfirmedBlocksSubmitter(appConfig, "prime", hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, blocksSubmitter)

		client, err := ethclient.Dial(appConfig.Bridge.NodeUrl)
		require.NoError(t, err)

		blocksSubmitter.ethClient = client

		err = blocksSubmitter.Dispose()
		require.NoError(t, err)
		require.Nil(t, blocksSubmitter.ethClient)
	})
}
