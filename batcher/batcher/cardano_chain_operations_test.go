package batcher

import (
	"encoding/json"
	"math/big"
	"math/rand"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/stretchr/testify/require"
)

func TestCardanoChainOperations(t *testing.T) {
	config := core.CardanoChainConfig{
		TestNetMagic: 42,
	}

	wallet := cardano.CardanoWallet{}

	t.Run("CreateBatchTx_AllInputs1Ada", func(t *testing.T) {
		cco := NewCardanoChainOperations(config, wallet)

		utxoCount := 10 // 10x 1Ada
		inputs := GenerateUTXOInputs(utxoCount*2, 1000000)
		outputs := GenerateUTXOOutputs(utxoCount, 1000000)
		txCost := CalculateTxCost(outputs)
		txInfos := GenerateTxInfos(t, config.TestNetMagic)

		metadata, err := cardano.CreateBatchMetaData(big.NewInt(100))
		require.NoError(t, err)
		protocolParams, err := GenerateProtocolParams()
		require.NoError(t, err)
		slotNumber := uint64(12345)

		txRaw, _, utxos, err := cco.CreateBatchTx(inputs, txCost, metadata, protocolParams, txInfos, outputs, slotNumber)
		require.NoError(t, err)
		require.Less(t, len(txRaw), 16000)
		require.Len(t, utxos.MultisigOwnedUTXOs, utxoCount)
		require.Equal(t, utxos.MultisigOwnedUTXOs[0].Amount, big.NewInt(1000000))
		require.Equal(t, utxos.MultisigOwnedUTXOs[len(utxos.MultisigOwnedUTXOs)-1].Amount, big.NewInt(1000000))
	})

	t.Run("CreateBatchTx_HalfInputs1Ada+Fill", func(t *testing.T) {
		cco := NewCardanoChainOperations(config, wallet)

		utxoCount := 10 // 10x 1Ada
		inputs := GenerateUTXOInputs(utxoCount, 1000000)
		outputs := GenerateUTXOOutputs(utxoCount*2, 1000000)
		txCost := CalculateTxCost(outputs)
		txInfos := GenerateTxInfos(t, config.TestNetMagic)

		metadata, err := cardano.CreateBatchMetaData(big.NewInt(100))
		require.NoError(t, err)
		protocolParams, err := GenerateProtocolParams()
		require.NoError(t, err)
		slotNumber := uint64(12345)

		txRaw, _, utxos, err := cco.CreateBatchTx(inputs, txCost, metadata, protocolParams, txInfos, outputs, slotNumber)
		require.NoError(t, err)
		require.Less(t, len(txRaw), 16000)
		require.Len(t, utxos.MultisigOwnedUTXOs, utxoCount+1) // 10 +1
		require.Equal(t, utxos.MultisigOwnedUTXOs[0].Amount, big.NewInt(1000000))
		require.Equal(t, utxos.MultisigOwnedUTXOs[len(utxos.MultisigOwnedUTXOs)-2].Amount, big.NewInt(1000000))
		require.Equal(t, utxos.MultisigOwnedUTXOs[len(utxos.MultisigOwnedUTXOs)-1].Amount, big.NewInt(1000000000))
	})

	t.Run("CreateBatchTx_TxSizeTooBig_IncludeBig", func(t *testing.T) {
		cco := NewCardanoChainOperations(config, wallet)

		inputs := GenerateUTXOInputs(30, 1000000)
		outputs := GenerateUTXOOutputs(400, 1000000)
		txCost := CalculateTxCost(outputs)
		txInfos := GenerateTxInfos(t, config.TestNetMagic)

		metadata, err := cardano.CreateBatchMetaData(big.NewInt(100))
		require.NoError(t, err)
		protocolParams, err := GenerateProtocolParams()
		require.NoError(t, err)
		slotNumber := uint64(12345)

		txRaw, _, utxos, err := cco.CreateBatchTx(inputs, txCost, metadata, protocolParams, txInfos, outputs, slotNumber)
		require.NoError(t, err)
		require.Less(t, len(txRaw), 16000)
		require.Less(t, len(utxos.MultisigOwnedUTXOs), 30)
		require.Equal(t, utxos.MultisigOwnedUTXOs[0].Amount, big.NewInt(1000000))
		require.Equal(t, utxos.MultisigOwnedUTXOs[len(utxos.MultisigOwnedUTXOs)-2].Amount, big.NewInt(1000000))
		require.Equal(t, utxos.MultisigOwnedUTXOs[len(utxos.MultisigOwnedUTXOs)-1].Amount, big.NewInt(1000000000))
	})

	t.Run("CreateBatchTx_TxSizeTooBig_IncludeBig2", func(t *testing.T) {
		cco := NewCardanoChainOperations(config, wallet)

		inputs := GenerateUTXOInputs(30, 1000000)
		outputs := GenerateUTXOOutputs(400, 10000000) // 4000Ada
		txCost := CalculateTxCost(outputs)
		txInfos := GenerateTxInfos(t, config.TestNetMagic)

		metadata, err := cardano.CreateBatchMetaData(big.NewInt(100))
		require.NoError(t, err)
		protocolParams, err := GenerateProtocolParams()
		require.NoError(t, err)
		slotNumber := uint64(12345)

		txRaw, _, utxos, err := cco.CreateBatchTx(inputs, txCost, metadata, protocolParams, txInfos, outputs, slotNumber)
		require.NoError(t, err)
		require.Less(t, len(txRaw), 16000)
		require.Less(t, len(utxos.MultisigOwnedUTXOs), 30)
		require.Equal(t, utxos.MultisigOwnedUTXOs[0].Amount, big.NewInt(1000000))
		require.Equal(t, utxos.MultisigOwnedUTXOs[len(utxos.MultisigOwnedUTXOs)-2].Amount, big.NewInt(1000000))
		require.Equal(t, utxos.MultisigOwnedUTXOs[len(utxos.MultisigOwnedUTXOs)-1].Amount, big.NewInt(4000000000))
	})

	t.Run("CreateBatchTx_TxSizeTooBig_LargeInput", func(t *testing.T) {
		cco := NewCardanoChainOperations(config, wallet)

		inputs := GenerateUTXOInputs(400, 1000000)
		outputs := GenerateUTXOOutputs(400, 1000000)
		txCost := CalculateTxCost(outputs)
		txInfos := GenerateTxInfos(t, config.TestNetMagic)

		metadata, err := cardano.CreateBatchMetaData(big.NewInt(100))
		require.NoError(t, err)
		protocolParams, err := GenerateProtocolParams()
		require.NoError(t, err)
		slotNumber := uint64(12345)

		txRaw, _, utxos, err := cco.CreateBatchTx(inputs, txCost, metadata, protocolParams, txInfos, outputs, slotNumber)
		require.NoError(t, err)
		require.Less(t, len(txRaw), 16000)
		require.Less(t, len(utxos.MultisigOwnedUTXOs), 30)
		require.Equal(t, utxos.MultisigOwnedUTXOs[0].Amount, big.NewInt(1000000))
		require.Equal(t, utxos.MultisigOwnedUTXOs[len(utxos.MultisigOwnedUTXOs)-2].Amount, big.NewInt(1000000))
		require.Equal(t, utxos.MultisigOwnedUTXOs[len(utxos.MultisigOwnedUTXOs)-1].Amount, big.NewInt(1000000000))
	})

	t.Run("CreateBatchTx_RandomInputs", func(t *testing.T) {
		cco := NewCardanoChainOperations(config, wallet)

		inputs := GenerateUTXORandomInputs(100, 1000000, 10000000)
		outputs := GenerateUTXORandomOutputs(200, 1000000, 10000000)
		txCost := CalculateTxCost(outputs)
		txInfos := GenerateTxInfos(t, config.TestNetMagic)

		metadata, err := cardano.CreateBatchMetaData(big.NewInt(100))
		require.NoError(t, err)
		protocolParams, err := GenerateProtocolParams()
		require.NoError(t, err)
		slotNumber := uint64(12345)

		txRaw, _, utxos, err := cco.CreateBatchTx(inputs, txCost, metadata, protocolParams, txInfos, outputs, slotNumber)
		require.NoError(t, err)
		require.Less(t, len(txRaw), 16000)
		require.LessOrEqual(t, len(utxos.MultisigOwnedUTXOs), 101)

		utxoSum := CalculateUTXOSum(utxos.MultisigOwnedUTXOs)
		require.Equal(t, utxoSum.Cmp(txCost), 1)
	})
}

func CalculateTxCost(outputs []cardanowallet.TxOutput) *big.Int {
	txCost := big.NewInt(0)
	for _, o := range outputs {
		txCost.Add(txCost, big.NewInt(int64(o.Amount)))
	}
	return txCost
}

func CalculateUTXOSum(inputs []eth.UTXO) *big.Int {
	txCost := big.NewInt(0)
	for _, i := range inputs {
		txCost.Add(txCost, i.Amount)
	}
	return txCost
}

func GenerateUTXOInputs(count int, amount int64) (inputs *contractbinding.IBridgeContractStructsUTXOs) {
	// Count x Input Ada, 1000Ada, 2000Ada, 3000Ada, 4000Ada, 5000Ada
	inputs = &contractbinding.IBridgeContractStructsUTXOs{
		MultisigOwnedUTXOs: make([]contractbinding.IBridgeContractStructsUTXO, count+6),
		FeePayerOwnedUTXOs: []contractbinding.IBridgeContractStructsUTXO{
			{Nonce: 1000, TxHash: "d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5", TxIndex: big.NewInt(1000), Amount: big.NewInt(10000000)},
			{Nonce: 1001, TxHash: "d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5", TxIndex: big.NewInt(1001), Amount: big.NewInt(10000000)},
		},
	}
	for i := 0; i < count; i++ {
		inputs.MultisigOwnedUTXOs[i] = contractbinding.IBridgeContractStructsUTXO{
			Nonce: uint64(i), TxHash: "d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5", TxIndex: big.NewInt(int64(i)), Amount: big.NewInt(amount),
		}
	}
	for i := 0; i < 5; i++ {
		inputs.MultisigOwnedUTXOs[count+i] = contractbinding.IBridgeContractStructsUTXO{
			Nonce: uint64(count + i), TxHash: "d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5", TxIndex: big.NewInt(int64(count + i)), Amount: big.NewInt(int64(1000000000 * (i + 1))),
		}
	}
	inputs.MultisigOwnedUTXOs[count+5] = contractbinding.IBridgeContractStructsUTXO{
		Nonce: uint64(count + 5), TxHash: "d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5", TxIndex: big.NewInt(int64(count + 5)), Amount: big.NewInt(int64(1000000000000)),
	}

	return
}

func GenerateUTXORandomInputs(count int, min uint64, max uint64) (inputs *contractbinding.IBridgeContractStructsUTXOs) {
	// Count x [min-max] Ada, 1000000Ada
	inputs = &contractbinding.IBridgeContractStructsUTXOs{
		MultisigOwnedUTXOs: make([]contractbinding.IBridgeContractStructsUTXO, count+1),
		FeePayerOwnedUTXOs: []contractbinding.IBridgeContractStructsUTXO{
			{Nonce: 1000, TxHash: "d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5", TxIndex: big.NewInt(1000), Amount: big.NewInt(10000000)},
			{Nonce: 1001, TxHash: "d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5", TxIndex: big.NewInt(1001), Amount: big.NewInt(10000000)},
		},
	}
	for i := 0; i < count; i++ {
		randomAmount := rand.Uint64() % max
		if randomAmount < min {
			randomAmount += min
		}
		inputs.MultisigOwnedUTXOs[i] = contractbinding.IBridgeContractStructsUTXO{
			Nonce: uint64(i), TxHash: "d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5", TxIndex: big.NewInt(int64(i)), Amount: big.NewInt(int64(randomAmount)),
		}
	}
	inputs.MultisigOwnedUTXOs[count] = contractbinding.IBridgeContractStructsUTXO{
		Nonce: uint64(count + 5), TxHash: "d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5", TxIndex: big.NewInt(int64(count + 5)), Amount: big.NewInt(int64(1000000000000)),
	}

	return
}

func GenerateUTXOOutputs(count int, amount uint64) (outputs []cardanowallet.TxOutput) {
	outputs = make([]cardanowallet.TxOutput, count)
	for i := 0; i < count; i++ {
		outputs[i] = cardanowallet.TxOutput{
			Addr:   "addr_test1vq7vsmgan2adwapu6r3xs5049s6dsf8hlgex68mwgxzraks4c0dpp",
			Amount: amount,
		}
	}
	return
}

func GenerateUTXORandomOutputs(count int, min uint64, max uint64) (outputs []cardanowallet.TxOutput) {
	outputs = make([]cardanowallet.TxOutput, count)
	for i := 0; i < count; i++ {
		randomAmount := rand.Uint64() % max
		if randomAmount < min {
			randomAmount += min
		}
		outputs[i] = cardanowallet.TxOutput{
			Addr:   "addr_test1vq7vsmgan2adwapu6r3xs5049s6dsf8hlgex68mwgxzraks4c0dpp",
			Amount: randomAmount,
		}
	}
	return
}

func GenerateTxInfos(t *testing.T, testnetMagic uint) *cardano.TxInputInfos {
	dummyKeyHashes := []string{
		"eff5e22355217ec6d770c3668010c2761fa0863afa12e96cff8a2205",
		"ad8e0ab92e1febfcaf44889d68c3ae78b59dc9c5fa9e05a272214c13",
		"bfd1c0eb0a453a7b7d668166ce5ca779c655e09e11487a6fac72dd6f",
		"b4689f2e8f37b406c5eb41b1fe2c9e9f4eec2597c3cc31b8dfee8f56",
		"39c196d28f804f70704b6dec5991fbb1112e648e067d17ca7abe614b",
		"adea661341df075349cbb2ad02905ce1828f8cf3e66f5012d48c3168",
	}

	multisigPolicyScript, err := cardanowallet.NewPolicyScript(dummyKeyHashes[0:3], 3)
	require.NoError(t, err)
	multisigFeePolicyScript, err := cardanowallet.NewPolicyScript(dummyKeyHashes[3:], 3)
	require.NoError(t, err)

	multisigAddress, err := multisigPolicyScript.CreateMultiSigAddress(testnetMagic)
	require.NoError(t, err)
	multisigFeeAddress, err := multisigFeePolicyScript.CreateMultiSigAddress(testnetMagic)
	require.NoError(t, err)

	txInfos := &cardano.TxInputInfos{
		TestNetMagic: testnetMagic,
		MultiSig: &cardano.TxInputInfo{
			PolicyScript: multisigPolicyScript,
			Address:      multisigAddress,
		},
		MultiSigFee: &cardano.TxInputInfo{
			PolicyScript: multisigFeePolicyScript,
			Address:      multisigFeeAddress,
		},
	}

	return txInfos
}

func GenerateProtocolParams() ([]byte, error) {
	resultJson := map[string]interface{}{
		"collateralPercentage": 150,
		"costModels":           nil,
		"decentralization":     nil,
		"executionUnitPrices": map[string]interface{}{
			"priceMemory": 5.77e-2,
			"priceSteps":  7.21e-5,
		},
		"extraPraosEntropy": nil,
		"maxBlockBodySize":  65536,
		"maxBlockExecutionUnits": map[string]interface{}{
			"memory": 80000000,
			"steps":  40000000000,
		},
		"maxBlockHeaderSize":  1100,
		"maxCollateralInputs": 3,
		"maxTxExecutionUnits": map[string]interface{}{
			"memory": 16000000,
			"steps":  10000000000,
		},
		"maxTxSize":           16384,
		"maxValueSize":        5000,
		"minPoolCost":         0,
		"minUTxOValue":        nil,
		"monetaryExpansion":   0.1,
		"poolPledgeInfluence": 0,
		"poolRetireMaxEpoch":  18,
		"protocolVersion": map[string]interface{}{
			"major": 7,
			"minor": 0,
		},
		"stakeAddressDeposit": 0,
		"stakePoolDeposit":    0,
		"stakePoolTargetNum":  100,
		"treasuryCut":         0.1,
		"txFeeFixed":          155381,
		"txFeePerByte":        44,
		"utxoCostPerByte":     4310,
	}

	return json.Marshal(resultJson)
}
