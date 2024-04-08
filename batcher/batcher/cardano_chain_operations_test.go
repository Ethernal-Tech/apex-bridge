package batcher

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"math/rand"
	"testing"
	"time"

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
		require.Equal(t, utxos.MultisigOwnedUTXOs[0].Amount, big.NewInt(1000000000))
		require.Equal(t, utxos.MultisigOwnedUTXOs[len(utxos.MultisigOwnedUTXOs)-2].Amount, big.NewInt(1000000))
		require.Equal(t, utxos.MultisigOwnedUTXOs[len(utxos.MultisigOwnedUTXOs)-1].Amount, big.NewInt(1000000))
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
		require.Equal(t, utxos.MultisigOwnedUTXOs[0].Amount, big.NewInt(1000000000))
		require.Equal(t, utxos.MultisigOwnedUTXOs[1].Amount, big.NewInt(2000000000))
		require.Equal(t, utxos.MultisigOwnedUTXOs[2].Amount, big.NewInt(3000000000))
		require.Equal(t, utxos.MultisigOwnedUTXOs[len(utxos.MultisigOwnedUTXOs)-2].Amount, big.NewInt(1000000))
		require.Equal(t, utxos.MultisigOwnedUTXOs[len(utxos.MultisigOwnedUTXOs)-1].Amount, big.NewInt(1000000))
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
		require.Equal(t, utxos.MultisigOwnedUTXOs[0].Amount, big.NewInt(1000000000))
		require.Equal(t, utxos.MultisigOwnedUTXOs[len(utxos.MultisigOwnedUTXOs)-2].Amount, big.NewInt(1000000))
		require.Equal(t, utxos.MultisigOwnedUTXOs[len(utxos.MultisigOwnedUTXOs)-1].Amount, big.NewInt(1000000))
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

	t.Run("CreateBatchTx_MinUtxoOrder", func(t *testing.T) {
		cco := NewCardanoChainOperations(config, wallet)

		inputs := GenerateUTXOInputsOrdered()        // 50, 40, 30, 101, 102, 103, 104, 105
		outputs := GenerateUTXOOutputs(403, 1000000) // 403Ada
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
		require.Len(t, utxos.MultisigOwnedUTXOs, 5)
		require.Equal(t, utxos.MultisigOwnedUTXOs[0].Amount, big.NewInt(50000000))
		require.Equal(t, utxos.MultisigOwnedUTXOs[1].Amount, big.NewInt(104000000))
		require.Equal(t, utxos.MultisigOwnedUTXOs[2].Amount, big.NewInt(103000000))
		require.Equal(t, utxos.MultisigOwnedUTXOs[3].Amount, big.NewInt(101000000))
		require.Equal(t, utxos.MultisigOwnedUTXOs[4].Amount, big.NewInt(102000000))
	})
}

func TestGenerateBatchTransaction(t *testing.T) {
	config := core.CardanoChainConfig{
		TestNetMagic:      42,
		AtLeastValidators: 1,
		BlockfrostUrl:     "https://cardano-preview.blockfrost.io/api/v0",
		BlockfrostAPIKey:  "preview7mGSjpyEKb24OxQ4cCxomxZ5axMs5PvE",
	}

	multisigVkeyString := "68fc463c29900b00122423c7e6a39469987786314e07a5e7f5eae76a5fe671bf"
	multisigVkeyBytes, err := hex.DecodeString(multisigVkeyString)
	require.NoError(t, err)
	multisigSkeyString := "1825bce09711e1563fc1702587da6892d1d869894386323bd4378ea5e3d6cba0"
	multisigSkeyBytes, err := hex.DecodeString(multisigSkeyString)
	require.NoError(t, err)
	multisigKeyHash := "eff5e22355217ec6d770c3668010c2761fa0863afa12e96cff8a2205"
	multisigStakeVkeyString := "0a809d270f7017c54a63a85fff145733861e671cf423c30428afd0cd7c759ad6"
	multisigStakeVkeyBytes, err := hex.DecodeString(multisigStakeVkeyString)
	require.NoError(t, err)
	multisigStakeSkeyString := "016586a09a19122dddc92aa512fb6ff0f0c3dddfc561dea5e18438e4269d3e00"
	multisigStakeSkeyBytes, err := hex.DecodeString(multisigStakeSkeyString)
	require.NoError(t, err)

	multisigFeeVkeyString := "63e95162d952d2fbc5240457750e1c13bfb4a5e3d9a96bf048b90bfe08b13de6"
	multisigFeeVkeyBytes, err := hex.DecodeString(multisigFeeVkeyString)
	require.NoError(t, err)
	multisigFeeSkeyString := "4cd84bf321e70ab223fbdbfe5eba249a5249bd9becbeb82109d45e56c9c610a9"
	multisigFeeSkeyBytes, err := hex.DecodeString(multisigFeeSkeyString)
	require.NoError(t, err)
	multisigFeeKeyHash := "b4689f2e8f37b406c5eb41b1fe2c9e9f4eec2597c3cc31b8dfee8f56"
	multisigFeeStakeVkeyString := "549b55365bbcdcff3cc8f7c824ba920f28b99e3ca379b0db4cbb895ceefd2765"
	multisigFeeStakeVkeyBytes, err := hex.DecodeString(multisigFeeStakeVkeyString)
	require.NoError(t, err)
	multisigFeeStakeSkeyString := "6da69a3342177847927465c1b03569d8b46af9d274cbc11e35322ce0a86d449a"
	multisigFeeStakeSkeyBytes, err := hex.DecodeString(multisigFeeStakeSkeyString)
	require.NoError(t, err)

	wallet := cardano.CardanoWallet{
		MultiSig:    cardanowallet.NewStakeWallet(multisigVkeyBytes, multisigSkeyBytes, multisigKeyHash, multisigStakeVkeyBytes, multisigStakeSkeyBytes),
		MultiSigFee: cardanowallet.NewStakeWallet(multisigFeeVkeyBytes, multisigFeeSkeyBytes, multisigFeeKeyHash, multisigFeeStakeVkeyBytes, multisigFeeStakeSkeyBytes),
	}

	cco := NewCardanoChainOperations(config, wallet)
	testError := errors.New("test err")

	var confirmedTransactions []eth.ConfirmedTransaction = make([]contractbinding.IBridgeContractStructsConfirmedTransaction, 1)
	confirmedTransactions[0] = eth.ConfirmedTransaction{
		Nonce:       big.NewInt(1),
		BlockHeight: big.NewInt(1),
		Receivers: []contractbinding.IBridgeContractStructsReceiver{{
			DestinationAddress: "addr_test1vqeux7xwusdju9dvsj8h7mca9aup2k439kfmwy773xxc2hcu7zy99",
			Amount:             big.NewInt(int64(minUtxoAmount)),
		}},
	}
	batchNonceId := big.NewInt(1)
	destinationChain := "vector"
	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	t.Run("GetLastObservedBlock returns error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetLastObservedBlock", ctx, destinationChain).Return(nil, testError)

		_, _, _, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceId)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	getLastObservedBlockRet := eth.CardanoBlock{
		BlockHash: "hash",
		BlockSlot: 1,
	}

	t.Run("GetValidatorsCardanoData returns error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetLastObservedBlock", ctx, destinationChain).Return(&getLastObservedBlockRet, nil)
		bridgeSmartContractMock.On("GetValidatorsCardanoData", ctx, destinationChain).Return(nil, testError)

		_, _, _, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceId)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("no vkey for multisig address error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetLastObservedBlock", ctx, destinationChain).Return(&getLastObservedBlockRet, nil)

		var getValidatorsCardanoDataRet []eth.ValidatorCardanoData = make([]contractbinding.IBridgeContractStructsValidatorCardanoData, 1)
		getValidatorsCardanoDataRet[0] = eth.ValidatorCardanoData{
			KeyHash:         "",
			KeyHashFee:      "",
			VerifyingKey:    "",
			VerifyingKeyFee: "",
		}
		bridgeSmartContractMock.On("GetValidatorsCardanoData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil)

		_, _, _, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceId)
		require.Error(t, err)
		require.ErrorContains(t, err, "verifying key of current batcher wasn't found in validators data queried from smart contract")
	})

	t.Run("no vkey for multisig fee address error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetLastObservedBlock", ctx, destinationChain).Return(&getLastObservedBlockRet, nil)

		var getValidatorsCardanoDataRet []eth.ValidatorCardanoData = make([]contractbinding.IBridgeContractStructsValidatorCardanoData, 1)
		getValidatorsCardanoDataRet[0] = eth.ValidatorCardanoData{
			KeyHash:         wallet.MultiSig.GetKeyHash(),
			KeyHashFee:      "",
			VerifyingKey:    hex.EncodeToString(wallet.MultiSig.GetVerificationKey()),
			VerifyingKeyFee: "",
		}
		bridgeSmartContractMock.On("GetValidatorsCardanoData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil)

		_, _, _, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceId)
		require.Error(t, err)
		require.ErrorContains(t, err, "verifying fee key of current batcher wasn't found in validators data queried from smart contract")
	})

	t.Run("GetAvailableUTXOs return error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetLastObservedBlock", ctx, destinationChain).Return(&getLastObservedBlockRet, nil)

		var getValidatorsCardanoDataRet []eth.ValidatorCardanoData = make([]contractbinding.IBridgeContractStructsValidatorCardanoData, 1)
		getValidatorsCardanoDataRet[0] = eth.ValidatorCardanoData{
			KeyHash:         wallet.MultiSig.GetKeyHash(),
			KeyHashFee:      wallet.MultiSigFee.GetKeyHash(),
			VerifyingKey:    hex.EncodeToString(wallet.MultiSig.GetVerificationKey()),
			VerifyingKeyFee: hex.EncodeToString(wallet.MultiSigFee.GetVerificationKey()),
		}
		bridgeSmartContractMock.On("GetValidatorsCardanoData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil)
		bridgeSmartContractMock.On("GetAvailableUTXOs", ctx, destinationChain).Return(nil, testError)

		_, _, _, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceId)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GetAvailableUTXOs return error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetLastObservedBlock", ctx, destinationChain).Return(&getLastObservedBlockRet, nil)

		var getValidatorsCardanoDataRet []eth.ValidatorCardanoData = make([]contractbinding.IBridgeContractStructsValidatorCardanoData, 1)
		getValidatorsCardanoDataRet[0] = eth.ValidatorCardanoData{
			KeyHash:         wallet.MultiSig.GetKeyHash(),
			KeyHashFee:      wallet.MultiSigFee.GetKeyHash(),
			VerifyingKey:    hex.EncodeToString(wallet.MultiSig.GetVerificationKey()),
			VerifyingKeyFee: hex.EncodeToString(wallet.MultiSigFee.GetVerificationKey()),
		}
		bridgeSmartContractMock.On("GetValidatorsCardanoData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil)
		bridgeSmartContractMock.On("GetAvailableUTXOs", ctx, destinationChain).Return(nil, testError)

		_, _, _, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceId)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GenerateBatchTransaction should pass", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetLastObservedBlock", ctx, destinationChain).Return(&getLastObservedBlockRet, nil)

		var getValidatorsCardanoDataRet []eth.ValidatorCardanoData = make([]contractbinding.IBridgeContractStructsValidatorCardanoData, 1)
		getValidatorsCardanoDataRet[0] = eth.ValidatorCardanoData{
			KeyHash:         wallet.MultiSig.GetKeyHash(),
			KeyHashFee:      wallet.MultiSigFee.GetKeyHash(),
			VerifyingKey:    hex.EncodeToString(wallet.MultiSig.GetVerificationKey()),
			VerifyingKeyFee: hex.EncodeToString(wallet.MultiSigFee.GetVerificationKey()),
		}
		bridgeSmartContractMock.On("GetValidatorsCardanoData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil)

		getAvailableUTXOsRet := &eth.UTXOs{
			MultisigOwnedUTXOs: []contractbinding.IBridgeContractStructsUTXO{{
				Nonce:   0,
				TxHash:  "26a9d1a894c7e3719a79342d0fc788989e5d55f076581327c54bcc0c7693905a",
				TxIndex: big.NewInt(0),
				Amount:  big.NewInt(10000000000),
			}},
			FeePayerOwnedUTXOs: []contractbinding.IBridgeContractStructsUTXO{{
				Nonce:   0,
				TxHash:  "26a9d1a894c7e3719a79342d0fc788989e5d55f076581327c54bcc0c7693905a",
				TxIndex: big.NewInt(0),
				Amount:  big.NewInt(10000000000),
			}},
		}
		bridgeSmartContractMock.On("GetAvailableUTXOs", ctx, destinationChain).Return(getAvailableUTXOsRet, nil)

		rawTx, txHash, utxos, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceId)
		require.NoError(t, err)
		require.NotNil(t, rawTx)
		require.NotEqual(t, "", txHash)
		require.Equal(t, *utxos, *getAvailableUTXOsRet)
	})

	t.Run("Test SignBatchTransaction", func(t *testing.T) {
		witnessMultiSig, witnessMultiSigFee, err := cco.SignBatchTransaction("26a9d1a894c7e3719a79342d0fc788989e5d55f076581327c54bcc0c7693905a")
		require.NoError(t, err)
		require.NotNil(t, witnessMultiSig)
		require.NotNil(t, witnessMultiSigFee)
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

func GenerateUTXOInputsOrdered() (inputs *contractbinding.IBridgeContractStructsUTXOs) {
	// Count x Input Ada, 1000Ada, 2000Ada, 3000Ada, 4000Ada, 5000Ada
	inputs = &contractbinding.IBridgeContractStructsUTXOs{
		MultisigOwnedUTXOs: make([]contractbinding.IBridgeContractStructsUTXO, 8),
		FeePayerOwnedUTXOs: []contractbinding.IBridgeContractStructsUTXO{
			{Nonce: 1000, TxHash: "d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5", TxIndex: big.NewInt(1000), Amount: big.NewInt(10000000)},
			{Nonce: 1001, TxHash: "d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5", TxIndex: big.NewInt(1001), Amount: big.NewInt(10000000)},
		},
	}
	inputs.MultisigOwnedUTXOs[0] = contractbinding.IBridgeContractStructsUTXO{
		Nonce: uint64(0), TxHash: "d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5", TxIndex: big.NewInt(int64(0)), Amount: big.NewInt(int64(50000000)),
	}
	inputs.MultisigOwnedUTXOs[1] = contractbinding.IBridgeContractStructsUTXO{
		Nonce: uint64(1), TxHash: "d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5", TxIndex: big.NewInt(int64(0)), Amount: big.NewInt(int64(40000000)),
	}
	inputs.MultisigOwnedUTXOs[2] = contractbinding.IBridgeContractStructsUTXO{
		Nonce: uint64(2), TxHash: "d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5", TxIndex: big.NewInt(int64(0)), Amount: big.NewInt(int64(30000000)),
	}
	inputs.MultisigOwnedUTXOs[3] = contractbinding.IBridgeContractStructsUTXO{
		Nonce: uint64(3), TxHash: "d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5", TxIndex: big.NewInt(int64(0)), Amount: big.NewInt(int64(101000000)),
	}
	inputs.MultisigOwnedUTXOs[4] = contractbinding.IBridgeContractStructsUTXO{
		Nonce: uint64(3), TxHash: "d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5", TxIndex: big.NewInt(int64(0)), Amount: big.NewInt(int64(102000000)),
	}
	inputs.MultisigOwnedUTXOs[5] = contractbinding.IBridgeContractStructsUTXO{
		Nonce: uint64(5), TxHash: "d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5", TxIndex: big.NewInt(int64(0)), Amount: big.NewInt(int64(103000000)),
	}
	inputs.MultisigOwnedUTXOs[6] = contractbinding.IBridgeContractStructsUTXO{
		Nonce: uint64(6), TxHash: "d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5", TxIndex: big.NewInt(int64(0)), Amount: big.NewInt(int64(104000000)),
	}
	inputs.MultisigOwnedUTXOs[7] = contractbinding.IBridgeContractStructsUTXO{
		Nonce: uint64(7), TxHash: "d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5", TxIndex: big.NewInt(int64(0)), Amount: big.NewInt(int64(105000000)),
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
