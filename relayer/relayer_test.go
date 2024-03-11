package relayer

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	contractbinding "github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelayConfig(t *testing.T) {
	config := &RelayerConfiguration{
		Cardano: CardanoConfig{
			TestNetMagic:      uint(2),
			BlockfrostUrl:     "https://cardano-preview.blockfrost.io/api/v0",
			BlockfrostAPIKey:  "preview7mGSjpyEKb24OxQ4cCxomxZ5axMs5PvE",
			AtLeastValidators: 2.0 / 3.0,
			PotentialFee:      300_000,
		},
		Bridge: BridgeConfig{
			NodeUrl:              "https://polygon-mumbai-pokt.nodies.app", // will be our node,
			SmartContractAddress: "0xF146ba6fAF3741df932a5d4074f414A15a621797",
		},
		PullTimeMilis: 1000,
		Logger: logger.LoggerConfig{
			LogFilePath:   "./relayer_logs",
			LogLevel:      hclog.Debug,
			JSONLogFormat: false,
			AppendFile:    true,
		},
	}

	loadedConfig, err := LoadConfig()
	assert.NoError(t, err)

	assert.Equal(t, config.Cardano, loadedConfig.Cardano)
	assert.Equal(t, config.Bridge, loadedConfig.Bridge)
	assert.Equal(t, config.PullTimeMilis, loadedConfig.PullTimeMilis)
	assert.Equal(t, config.Logger, loadedConfig.Logger)
}

func TestBatchSubmissionContract(t *testing.T) {

	config, err := LoadConfig()
	assert.NoError(t, err)

	signedBatchId := big.NewInt(1)

	txRaw, txHash := createTxRawHelper(t, config, signedBatchId)
	witnessesString, witnessesBytes := generateWitnesses(t, txHash)

	valueToSet := contractbinding.TestContractConfirmedBatch{
		Id:                         signedBatchId.String(),
		RawTransaction:             hex.EncodeToString(txRaw),
		MultisigSignatures:         witnessesString[0:3],
		FeePayerMultisigSignatures: witnessesString[3:],
	}

	scAddress := config.Bridge.SmartContractAddress

	wallet, err := ethtxhelper.NewEthTxWallet(dummyMumbaiAccPk)
	require.NoError(t, err)

	ethClient, err := ethclient.Dial(config.Bridge.NodeUrl)
	require.NoError(t, err)
	txHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithClient(ethClient))
	require.NoError(t, err)

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	contract, err := contractbinding.NewTestContract(common.HexToAddress(scAddress), txHelper.GetClient())
	require.NoError(t, err)

	// Set confirmed batch value
	tx, err := txHelper.SendTx(ctx, wallet, bind.TransactOpts{}, true, func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
		return contract.SetConfirmedBatch(txOpts, valueToSet)
	})
	require.NoError(t, err)

	receipt, err := txHelper.WaitForReceipt(ctx, tx.Hash().String(), true)
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)

	t.Run("check get data directly from contract", func(t *testing.T) {
		// Get value for comparison
		// TODO: Update with real parameter
		res, err := contract.GetConfirmedBatch(&bind.CallOpts{
			Context: ctx,
			From:    wallet.GetAddress(),
		}, "destinationChain")
		require.NoError(t, err)

		assert.Equal(t, valueToSet.Id, res.Id)
		assert.Equal(t, valueToSet.RawTransaction, res.RawTransaction)
		assert.Equal(t, valueToSet.MultisigSignatures, res.MultisigSignatures)
		assert.Equal(t, valueToSet.FeePayerMultisigSignatures, res.FeePayerMultisigSignatures)
	})

	t.Run("check data from relayer.getSmartContractData()", func(t *testing.T) {
		expectedReturn := ConfirmedBatch{
			id:                         signedBatchId.String(),
			rawTransaction:             txRaw,
			multisigSignatures:         witnessesBytes[0:3],
			feePayerMultisigSignatures: witnessesBytes[3:],
		}

		logger, err := logger.NewLogger(config.Logger)
		assert.NoError(t, err)

		operations := GetOperations(config.Cardano.TestNetMagic)
		r := NewRelayer(config, logger, operations)

		res, err := r.operations.GetConfirmedBatch(ctx, ethClient, r.config.Bridge.SmartContractAddress)
		assert.NoError(t, err)

		assert.Equal(t, expectedReturn.id, res.id)
		assert.Equal(t, expectedReturn.rawTransaction, res.rawTransaction)
		assert.Equal(t, expectedReturn.multisigSignatures, res.multisigSignatures)
		assert.Equal(t, expectedReturn.feePayerMultisigSignatures, res.feePayerMultisigSignatures)
	})
}

func createTxRawHelper(t *testing.T, config *RelayerConfiguration, signedBatchId *big.Int) (txRaw []byte, txHash string) {
	txInfos, err := cardanotx.NewTxInputInfos(
		dummyKeyHashes[0:3], dummyKeyHashes[3:], config.Cardano.TestNetMagic)
	assert.NoError(t, err)

	txProvider, err := cardanowallet.NewTxProviderBlockFrost(config.Cardano.BlockfrostUrl, config.Cardano.BlockfrostAPIKey)
	assert.NoError(t, err)

	err = txInfos.CalculateWithRetriever(txProvider, cardanowallet.GetOutputsSum(dummyOutputs), config.Cardano.PotentialFee)
	assert.NoError(t, err)

	metadata, err := cardanotx.CreateMetaData(signedBatchId)
	assert.NoError(t, err)

	protocolParams, err := txProvider.GetProtocolParameters()
	assert.NoError(t, err)

	slotNumber, err := txProvider.GetSlot()
	assert.NoError(t, err)

	txRaw, txHash, err = cardanotx.CreateTx(config.Cardano.TestNetMagic, protocolParams, slotNumber+cardanotx.TTLSlotNumberInc,
		metadata, txInfos, dummyOutputs)
	assert.NoError(t, err)

	return
}

func generateWitnesses(t *testing.T, txHash string) ([]string, [][]byte) {
	var witnessesString []string
	var witnessesBytes [][]byte
	for _, key := range dummySigningKeys {
		witness, err := cardanotx.CreateTxWitness(txHash, cardanotx.NewSigningKey(key))
		assert.NoError(t, err)
		witnessesBytes = append(witnessesBytes, witness)
		witnessesString = append(witnessesString, hex.EncodeToString(witness))
	}

	return witnessesString, witnessesBytes
}

var (
	dummyMumbaiAccPk = "3761f6deeb2e0b2aa8b843e804d880afa6e5fecf1631f411e267641a72d0ca20"
	dummyKeyHashes   = []string{
		"089732e4f6fc248b599c6b24b75187c39842f515733c833e0f09795b",
		"474187985a19732d1abbe1114c1af4cf084d58511884800ddfca3a82",
		"d92df0aff3bf46f084c5744ef25ef33f34318621027a66790b66da31",
		"cd0f2d9b43edb2cfa501f4d7c64413ed57c9147ce0c3aac520bfc565",
		"f8dd5736c4bc7b0d07bff7f018948838f87c703c01b368a38f2cf234",
		"004ee443c6b1a1aa59699545b7bfdf25db64c4d3a64fd1fe10d20829",
	}
	dummySigningKeys = []string{
		"58201217236ac24d8ac12684b308cf9468f68ef5283096896dc1c5c3caf8351e2847",
		"58207e62090b7c574dd71423d4d1d089675bcde049fb2c677fea7add2d94120f01de",
		"582060d76923536885313a7a9dc5a8ed68a22a5e0edee88ca5eb8b10f1e162c57530",
		"5820f2c3b9527ec2f0d70e6ee2db5752e27066fe63f5c84d1aa5bf20a5fc4d2411e6",
		"58202bf1bed17d19f44f53ac64fa4621c879f8295af52080cffb2a8d9d10117ae772",
		"58202cdf4d3b56f3d9ea7b7c9424d841273e2adb1bd11a98a4370ad22f3bac9104e2",
	}
	dummyOutputs = []cardanowallet.TxOutput{
		{
			Addr:   "addr_test1vqjysa7p4mhu0l25qknwznvj0kghtr29ud7zp732ezwtzec0w8g3u",
			Amount: cardanowallet.MinUTxODefaultValue,
		},
	}
)
