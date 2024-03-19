package batcher

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/batcher/bridge"
	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

func TestBridgeMethods(t *testing.T) {
	config := &core.BatcherConfiguration{
		Base: core.BaseConfig{
			ChainId:               "prime",
			SigningKeyMultiSig:    "58201825bce09711e1563fc1702587da6892d1d869894386323bd4378ea5e3d6cba0",
			SigningKeyMultiSigFee: "58204cd84bf321e70ab223fbdbfe5eba249a5249bd9becbeb82109d45e56c9c610a9",
		},
		Bridge: core.BridgeConfig{
			NodeUrl:              "https://polygon-mumbai-pokt.nodies.app", // will be our node,
			SmartContractAddress: "0x4c7aBbe2c5A44d758b70BE5C3c07eEB573304Db4",
			SigningKey:           "3761f6deeb2e0b2aa8b843e804d880afa6e5fecf1631f411e267641a72d0ca20",
		},
		PullTimeMilis: 2500,
		Logger: logger.LoggerConfig{
			LogFilePath:   "./batcher_logs",
			LogLevel:      hclog.Debug,
			JSONLogFormat: false,
			AppendFile:    true,
		},
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	scAddress := config.Bridge.SmartContractAddress
	destinationChain := config.Base.ChainId

	ethClient, err := ethclient.Dial(config.Bridge.NodeUrl)
	assert.NoError(t, err)
	txHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithClient(ethClient))
	assert.NoError(t, err)

	t.Run("test ShouldCreateBatch", func(t *testing.T) {
		res, err := bridge.ShouldCreateBatch(ctx, txHelper, scAddress, destinationChain)
		assert.NoError(t, err)
		assert.Equal(t, false, res)
	})

	t.Run("test GetConfirmedTransactions", func(t *testing.T) {
		res, err := bridge.GetConfirmedTransactions(ctx, txHelper, scAddress, destinationChain)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(res))
	})

	t.Run("test GetAvailableUTXOs", func(t *testing.T) {
		res, err := bridge.GetAvailableUTXOs(ctx, txHelper, scAddress, destinationChain, big.NewInt(0))
		assert.NoError(t, err)
		assert.Equal(t, &contractbinding.TestContractUTXOs{
			MultisigOwnedUTXOs: []contractbinding.TestContractUTXO{},
			FeePayerOwnedUTXOs: []contractbinding.TestContractUTXO{},
		}, res)
	})

	t.Run("test SubmitSignedBatch", func(t *testing.T) {
		signedBatch := contractbinding.TestContractSignedBatch{
			Id:                        "",
			DestinationChainId:        destinationChain,
			RawTransaction:            "",
			MultisigSignature:         "",
			FeePayerMultisigSignature: "",
			IncludedTransactions:      []contractbinding.TestContractConfirmedTransaction{},
			UsedUTXOs:                 contractbinding.TestContractUTXOs{},
		}

		err := bridge.SubmitSignedBatch(ctx, txHelper, scAddress, signedBatch, config.Bridge.SigningKey)
		assert.NoError(t, err)
	})
}
