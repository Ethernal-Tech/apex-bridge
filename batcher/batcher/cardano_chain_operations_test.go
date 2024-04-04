package batcher

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
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

		inputs := GenerateUTXOInputs(500, 1000000) // 500x 1Ada
		txInfos := GenerateTxInfos(42)
		outputs := GenerateUTXOOutputs(100, 1000000) // 100x 1Ada
		txCost := CalculateTxCost(outputs)

		metadata := []byte("")
		protocolParams, _ := GenerateProtocolParams()
		slotNumber := uint64(12345)

		txRaw, _, utxos, err := cco.CreateBatchTx(inputs, txCost, metadata, protocolParams, txInfos, outputs, slotNumber)
		require.NoError(t, err)
		require.Less(t, len(txRaw), 16000)
		require.Len(t, utxos.MultisigOwnedUTXOs, 100)
	})

	t.Run("CreateBatchTx_HalfInputs1Ada+Fill", func(t *testing.T) {
		cco := NewCardanoChainOperations(config, wallet)

		inputs := GenerateUTXOInputs(50, 1000000) // 500x 1Ada
		txInfos := GenerateTxInfos(42)
		outputs := GenerateUTXOOutputs(100, 1000000) // 100x 1Ada
		txCost := CalculateTxCost(outputs)

		metadata := []byte("")
		protocolParams, _ := GenerateProtocolParams()
		slotNumber := uint64(12345)

		txRaw, _, utxos, err := cco.CreateBatchTx(inputs, txCost, metadata, protocolParams, txInfos, outputs, slotNumber)
		require.NoError(t, err)
		require.Less(t, len(txRaw), 16000)
		require.Len(t, utxos.MultisigOwnedUTXOs, 51)
		require.Equal(t, utxos.MultisigOwnedUTXOs[50].Amount, big.NewInt(1000000000))
	})

	t.Run("CreateBatchTx_TxSizeTooBig_IncludeBig", func(t *testing.T) {
		cco := NewCardanoChainOperations(config, wallet)

		inputs := GenerateUTXOInputs(500, 1000000) // 500x 1Ada
		txInfos := GenerateTxInfos(42)
		outputs := GenerateUTXOOutputs(500, 1000000) // 500x 1Ada
		txCost := CalculateTxCost(outputs)

		metadata := []byte("")
		protocolParams, _ := GenerateProtocolParams()
		slotNumber := uint64(12345)

		txRaw, _, utxos, err := cco.CreateBatchTx(inputs, txCost, metadata, protocolParams, txInfos, outputs, slotNumber)
		require.NoError(t, err)
		require.Less(t, len(txRaw), 16000)
		require.Less(t, len(utxos.MultisigOwnedUTXOs), 500)
		require.Equal(t, utxos.MultisigOwnedUTXOs[len(utxos.MultisigOwnedUTXOs)].Amount, big.NewInt(1000000000))
	})

	t.Run("CreateBatchTx_TxSizeTooBig_IncludeBig2", func(t *testing.T) {
		cco := NewCardanoChainOperations(config, wallet)

		inputs := GenerateUTXOInputs(500, 1000000) // 500x 1Ada
		txInfos := GenerateTxInfos(42)
		outputs := GenerateUTXOOutputs(500, 2000000) // 500x 2Ada
		txCost := CalculateTxCost(outputs)

		metadata := []byte("")
		protocolParams, _ := GenerateProtocolParams()
		slotNumber := uint64(12345)

		txRaw, _, utxos, err := cco.CreateBatchTx(inputs, txCost, metadata, protocolParams, txInfos, outputs, slotNumber)
		require.NoError(t, err)
		require.Less(t, len(txRaw), 16000)
		require.Less(t, len(utxos.MultisigOwnedUTXOs), 500)
		require.Equal(t, utxos.MultisigOwnedUTXOs[len(utxos.MultisigOwnedUTXOs)].Amount, big.NewInt(2000000000))
	})

	t.Run("CreateBatchTx_TxSizeTooBig_BiggestUTXO", func(t *testing.T) {
		cco := NewCardanoChainOperations(config, wallet)

		inputs := GenerateUTXOInputs(10, 1000000) // 10x 1Ada
		txInfos := GenerateTxInfos(42)
		outputs := GenerateUTXOOutputs(500, 10000000) // 500x 10Ada
		txCost := CalculateTxCost(outputs)

		metadata := []byte("")
		protocolParams, _ := GenerateProtocolParams()
		slotNumber := uint64(12345)

		txRaw, _, utxos, err := cco.CreateBatchTx(inputs, txCost, metadata, protocolParams, txInfos, outputs, slotNumber)
		require.NoError(t, err)
		require.Less(t, len(txRaw), 16000)
		require.Equal(t, len(utxos.MultisigOwnedUTXOs), 1)
	})

}

func CalculateTxCost(outputs []cardanowallet.TxOutput) *big.Int {
	txCost := big.NewInt(0)
	for _, o := range outputs {
		txCost.Add(txCost, big.NewInt(int64(o.Amount)))
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

func GenerateTxInfos(testnetMagic uint) *cardano.TxInputInfos {
	keyHashes := []string{
		"ea0934197f2ae54fa8577d03e88ab3095058a3615f4d667aada1b5822c6140e7",
		"ea0934197f2ae54fa8577d03e88ab3095058a3615f4d667aada1b5822c6140e7",
		"ea0934197f2ae54fa8577d03e88ab3095058a3615f4d667aada1b5822c6140e7",
		"ea0934197f2ae54fa8577d03e88ab3095058a3615f4d667aada1b5822c6140e7",
		"ea0934197f2ae54fa8577d03e88ab3095058a3615f4d667aada1b5822c6140e7",
	}

	multisigPolicyScript, _ := cardanowallet.NewPolicyScript(keyHashes, 3)
	multisigFeePolicyScript, _ := cardanowallet.NewPolicyScript(keyHashes, 3)

	// multisigAddress, _ := multisigPolicyScript.CreateMultiSigAddress(testnetMagic)
	multisigAddress := "addr_test1vq7vsmgan2adwapu6r3xs5049s6dsf8hlgex68mwgxzraks4c0dpp"
	// multisigFeeAddress, _ := multisigFeePolicyScript.CreateMultiSigAddress(testnetMagic)
	multisigFeeAddress := "addr_test1vq7vsmgan2adwapu6r3xs5049s6dsf8hlgex68mwgxzraks4c0dpp"

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
