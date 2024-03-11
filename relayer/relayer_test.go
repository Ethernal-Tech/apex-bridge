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
		res, err := contract.GetConfirmedBatch(&bind.CallOpts{
			Context: ctx,
			From:    wallet.GetAddress(),
		}, "prime")
		require.NoError(t, err)

		assert.Equal(t, valueToSet.Id, res.Id)
		assert.Equal(t, valueToSet.RawTransaction, res.RawTransaction)
		assert.Equal(t, valueToSet.MultisigSignatures, res.MultisigSignatures)
		assert.Equal(t, valueToSet.FeePayerMultisigSignatures, res.FeePayerMultisigSignatures)
	})

	var contractData *ConfirmedBatch
	logger, err := logger.NewLogger(config.Logger)
	assert.NoError(t, err)

	operations := GetOperations(config.Cardano.TestNetMagic)
	relayer := NewRelayer(config, logger, operations)
	t.Run("check data from relayer.getSmartContractData()", func(t *testing.T) {
		expectedReturn := ConfirmedBatch{
			id:                         signedBatchId.String(),
			rawTransaction:             txRaw,
			multisigSignatures:         witnessesBytes[0:3],
			feePayerMultisigSignatures: witnessesBytes[3:],
		}

		contractData, err = relayer.operations.GetConfirmedBatch(ctx, ethClient, relayer.config.Bridge.SmartContractAddress)
		assert.NoError(t, err)

		assert.Equal(t, expectedReturn.id, contractData.id)
		assert.Equal(t, expectedReturn.rawTransaction, contractData.rawTransaction)
		assert.Equal(t, expectedReturn.multisigSignatures, contractData.multisigSignatures)
		assert.Equal(t, expectedReturn.feePayerMultisigSignatures, contractData.feePayerMultisigSignatures)
	})

	t.Run("submit tx to cardano chain", func(t *testing.T) {
		err = relayer.SendTx(contractData)
		assert.NoError(t, err)
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
		"eff5e22355217ec6d770c3668010c2761fa0863afa12e96cff8a2205",
		"ad8e0ab92e1febfcaf44889d68c3ae78b59dc9c5fa9e05a272214c13",
		"bfd1c0eb0a453a7b7d668166ce5ca779c655e09e11487a6fac72dd6f",
		"b4689f2e8f37b406c5eb41b1fe2c9e9f4eec2597c3cc31b8dfee8f56",
		"39c196d28f804f70704b6dec5991fbb1112e648e067d17ca7abe614b",
		"adea661341df075349cbb2ad02905ce1828f8cf3e66f5012d48c3168",
	}
	dummySigningKeys = []string{
		"58201825bce09711e1563fc1702587da6892d1d869894386323bd4378ea5e3d6cba0",
		"5820ccdae0d1cd3fa9be16a497941acff33b9aa20bdbf2f9aa5715942d152988e083",
		"582094bfc7d65a5d936e7b527c93ea6bf75de51029290b1ef8c8877bffe070398b40",
		"58204cd84bf321e70ab223fbdbfe5eba249a5249bd9becbeb82109d45e56c9c610a9",
		"58208fcc8cac6b7fedf4c30aed170633df487642cb22f7e8615684e2b98e367fcaa3",
		"582058fb35da120c65855ad691dadf5681a2e4fc62e9dcda0d0774ff6fdc463a679a",
	}
	dummyOutputs = []cardanowallet.TxOutput{
		{
			Addr:   "addr_test1vqjysa7p4mhu0l25qknwznvj0kghtr29ud7zp732ezwtzec0w8g3u",
			Amount: cardanowallet.MinUTxODefaultValue,
		},
	}
)
