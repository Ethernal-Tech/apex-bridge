package relayer

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/apex-bridge/relayer/bridge"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelayer(t *testing.T) {
	signedBatchId := big.NewInt(2)

	relayerConfig := &core.RelayerConfiguration{
		Bridge: core.BridgeConfig{
			NodeUrl:              "https://polygon-mumbai-pokt.nodies.app", // will be our node,
			SmartContractAddress: "0x816402271eE6D9078Fc8Cb537aDBDD58219485BA",
		},
		Base: core.BaseConfig{
			ChainId: "prime",
		},
		PullTimeMilis: 1000,
		Logger: logger.LoggerConfig{
			LogFilePath:   "./relayer_logs",
			LogLevel:      hclog.Debug,
			JSONLogFormat: false,
			AppendFile:    true,
		},
	}

	jsonData := []byte(`{
		"testnetMagic": 2,
		"blockfrostUrl": "https://cardano-preview.blockfrost.io/api/v0",
		"blockfrostApiKey": "preview7mGSjpyEKb24OxQ4cCxomxZ5axMs5PvE",
		"atLeastValidators": 0.6666666666666666,
		"potentialFee": 300000
		}`)

	chainSpecificConfig := core.ChainSpecific{
		ChainType: "Cardano",
		Config:    json.RawMessage(jsonData),
	}

	scAddress := relayerConfig.Bridge.SmartContractAddress

	wallet, err := ethtxhelper.NewEthTxWallet(dummyMumbaiAccPk)
	assert.NoError(t, err)

	ethClient, err := ethclient.Dial(relayerConfig.Bridge.NodeUrl)
	assert.NoError(t, err)
	txHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithClient(ethClient))
	assert.NoError(t, err)

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	contract, err := contractbinding.NewTestContract(common.HexToAddress(scAddress), txHelper.GetClient())
	assert.NoError(t, err)

	t.Run("check get data directly from contract", func(t *testing.T) {
		// Get value for comparison

		res, err := contract.GetConfirmedBatch(&bind.CallOpts{
			Context: ctx,
			From:    wallet.GetAddress(),
		}, relayerConfig.Base.ChainId)
		require.NoError(t, err)

		contractData, err := bridge.GetSmartContractData(ctx, txHelper, relayerConfig.Base.ChainId, relayerConfig.Bridge.SmartContractAddress)
		assert.NoError(t, err)

		assert.Equal(t, res.Id, contractData.Id)
		assert.Equal(t, res.RawTransaction, hex.EncodeToString(contractData.RawTransaction))
		for i, _ := range contractData.MultisigSignatures {
			assert.Equal(t, res.MultisigSignatures[i], hex.EncodeToString(contractData.MultisigSignatures[i]))
		}
		for i, _ := range contractData.FeePayerMultisigSignatures {
			assert.Equal(t, res.FeePayerMultisigSignatures[i], hex.EncodeToString(contractData.FeePayerMultisigSignatures[i]))
		}

	})

	assert.NoError(t, err)

	operations, err := GetChainSpecificOperations(chainSpecificConfig)
	assert.NoError(t, err)

	r := NewRelayer(relayerConfig, hclog.Default(), operations)

	t.Run("submit tx to cardano chain", func(t *testing.T) {
		cardanoChainConfig, err := core.ToCardanoChainConfig(chainSpecificConfig)
		assert.NoError(t, err)

		txRaw, txHash := createTxRawHelper(t, cardanoChainConfig, signedBatchId)
		_, witnessesBytes := generateWitnessesHelper(t, txHash)

		confirmedBatch := bridge.ConfirmedBatch{
			Id:                         signedBatchId.String(),
			RawTransaction:             txRaw,
			MultisigSignatures:         witnessesBytes[0:3],
			FeePayerMultisigSignatures: witnessesBytes[3:],
		}

		err = r.operations.SendTx(&confirmedBatch)
		assert.NoError(t, err)
	})
}

func createTxRawHelper(t *testing.T, config *core.CardanoChainConfig, signedBatchId *big.Int) (txRaw []byte, txHash string) {
	txInfos, err := cardanotx.NewTxInputInfos(
		dummyKeyHashes[0:3], dummyKeyHashes[3:], config.TestNetMagic)
	assert.NoError(t, err)

	txProvider, err := cardanowallet.NewTxProviderBlockFrost(config.BlockfrostUrl, config.BlockfrostAPIKey)
	assert.NoError(t, err)

	err = txInfos.CalculateWithRetriever(context.Background(), txProvider, cardanowallet.GetOutputsSum(dummyOutputs), config.PotentialFee)
	assert.NoError(t, err)

	metadata, err := cardanotx.CreateMetaData(signedBatchId)
	assert.NoError(t, err)

	protocolParams, err := txProvider.GetProtocolParameters(context.Background())
	assert.NoError(t, err)

	slotNumber, err := txProvider.GetSlot(context.Background())
	assert.NoError(t, err)

	txRaw, txHash, err = cardanotx.CreateTx(config.TestNetMagic, protocolParams, slotNumber+cardanotx.TTLSlotNumberInc,
		metadata, txInfos, dummyOutputs)
	assert.NoError(t, err)

	return
}

func generateWitnessesHelper(t *testing.T, txHash string) ([]string, [][]byte) {
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
