package batcher

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"path"
	"testing"
	"time"

	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	secretsHelper "github.com/Ethernal-Tech/cardano-infrastructure/secrets/helper"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCardanoChainOperations_IsSynchronized(t *testing.T) {
	chainID := "prime"
	dbMock := &indexer.DatabaseMock{}
	bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
	ctx := context.Background()
	scBlock1 := &eth.CardanoBlock{
		BlockSlot: big.NewInt(15),
	}
	scBlock2 := &eth.CardanoBlock{
		BlockSlot: big.NewInt(20),
	}
	oracleBlock1 := &indexer.BlockPoint{
		BlockSlot: uint64(10),
	}
	oracleBlock2 := &indexer.BlockPoint{
		BlockSlot: uint64(20),
	}
	testErr1 := errors.New("test error 1")
	testErr2 := errors.New("test error 2")

	bridgeSmartContractMock.On("GetLastObservedBlock", ctx, chainID).Return((*eth.CardanoBlock)(nil), testErr1).Once()
	bridgeSmartContractMock.On("GetLastObservedBlock", ctx, chainID).Return((*eth.CardanoBlock)(nil), nil).Once()
	bridgeSmartContractMock.On("GetLastObservedBlock", ctx, chainID).Return(scBlock1, nil).Times(3)
	bridgeSmartContractMock.On("GetLastObservedBlock", ctx, chainID).Return(scBlock2, nil).Once()

	dbMock.On("GetLatestBlockPoint").Return((*indexer.BlockPoint)(nil), testErr2).Once()
	dbMock.On("GetLatestBlockPoint").Return(oracleBlock1, nil).Once()
	dbMock.On("GetLatestBlockPoint").Return(oracleBlock2, nil).Twice()

	cco := &CardanoChainOperations{
		db:     dbMock,
		logger: hclog.NewNullLogger(),
	}

	// sc error
	_, err := cco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.ErrorIs(t, err, testErr1)

	// sc point not set
	val, err := cco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.NoError(t, err)
	require.True(t, val)

	// database error
	_, err = cco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.ErrorIs(t, err, testErr2)

	// not in sync
	val, err = cco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.NoError(t, err)
	require.False(t, val)

	// in sync
	val, err = cco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.NoError(t, err)
	require.True(t, val)

	// in sync again
	val, err = cco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.NoError(t, err)
	require.True(t, val)
}

func TestCardanoChainOperations(t *testing.T) {
	t.Skip()
	// testDir, err := os.MkdirTemp("", "bat-chain-ops")
	// require.NoError(t, err)

	// defer func() {
	// 	os.RemoveAll(testDir)
	// 	os.Remove(testDir)
	// }()

	// secretsMngr, err := secretsHelper.CreateSecretsManager(&secrets.SecretsManagerConfig{
	// 	Path: path.Join(testDir, "stp"),
	// 	Type: secrets.Local,
	// })
	// require.NoError(t, err)

	// _, err = cardano.GenerateWallet(secretsMngr, "prime", true, false)
	// require.NoError(t, err)

	// _, err = cardano.GenerateWallet(secretsMngr, "vector", true, false)
	// require.NoError(t, err)

	// configRaw := json.RawMessage([]byte(`{
	// 		"socketPath": "./socket",
	// 		"testnetMagic": 2,
	// 		"potentialFee": 300000
	// 		}`))

	// 	t.Run("CreateBatchTx_AllInputs1Ada", func(t *testing.T) {
	// 		cco, err := NewCardanoChainOperations(configRaw, &indexer.DatabaseMock{}, secretsMngr, "prime", hclog.NewNullLogger())
	// 		require.NoError(t, err)

	// 		utxoCount := 10 // 10x 1Ada
	// 		inputs := generateUTXOInputs(utxoCount*2, 1000000)
	// 		outputs := calculateTxCost(generateUTXOOutputs(utxoCount, 1000000))
	// 		txInfos := generateTxInfos(t, cco.Config.TestNetMagic)

	// 		metadata, err := cardano.CreateBatchMetaData(100)
	// 		require.NoError(t, err)
	// 		protocolParams, err := generateProtocolParams()
	// 		require.NoError(t, err)

	// 		slotNumber := uint64(12345)

	// 		result, err := cco.createBatchTx(inputs, metadata, protocolParams, txInfos, outputs, slotNumber)
	// 		require.NoError(t, err)
	// 		require.Less(t, len(result.TxRaw), 16000)
	// 		require.Len(t, result.Utxos.MultisigOwnedUTXOs, utxoCount)
	// 		require.Equal(t, uint64(1000000), result.Utxos.MultisigOwnedUTXOs[0].Amount)
	// 		require.Equal(t, uint64(1000000), result.Utxos.MultisigOwnedUTXOs[len(result.Utxos.MultisigOwnedUTXOs)-1].Amount)
	// 	})

	// 	t.Run("CreateBatchTx_HalfInputs1Ada+Fill", func(t *testing.T) {
	// 		cco, err := NewCardanoChainOperations(configRaw, &indexer.DatabaseMock{}, secretsMngr, "prime", hclog.NewNullLogger())
	// 		require.NoError(t, err)

	// 		utxoCount := 10 // 10x 1Ada
	// 		inputs := generateUTXOInputs(utxoCount, 1000000)
	// 		outputs := calculateTxCost(generateUTXOOutputs(utxoCount*2, 1000000))
	// 		txInfos := generateTxInfos(t, cco.Config.TestNetMagic)

	// 		metadata, err := cardano.CreateBatchMetaData(100)
	// 		require.NoError(t, err)
	// 		protocolParams, err := generateProtocolParams()
	// 		require.NoError(t, err)

	// 		slotNumber := uint64(12345)

	// 		result, err := cco.createBatchTx(inputs, metadata, protocolParams, txInfos, outputs, slotNumber)
	// 		require.NoError(t, err)
	// 		require.Less(t, len(result.TxRaw), 16000)
	// 		require.Len(t, result.Utxos.FeePayerOwnedUTXOs, len(inputs.FeePayerOwnedUTXOs))
	// 		require.Len(t, result.Utxos.MultisigOwnedUTXOs, utxoCount+1) // 10 +1
	// 		require.Equal(t, uint64(1000000), result.Utxos.MultisigOwnedUTXOs[0].Amount)
	// 		require.Equal(t, uint64(1000000), result.Utxos.MultisigOwnedUTXOs[len(result.Utxos.MultisigOwnedUTXOs)-2].Amount)
	// 		require.Equal(t, uint64(1000000000), result.Utxos.MultisigOwnedUTXOs[len(result.Utxos.MultisigOwnedUTXOs)-1].Amount)
	// 	})

	// 	t.Run("CreateBatchTx_TxSizeTooBig_IncludeBig", func(t *testing.T) {
	// 		cco, err := NewCardanoChainOperations(configRaw, &indexer.DatabaseMock{}, secretsMngr, "vector", hclog.NewNullLogger())
	// 		require.NoError(t, err)

	// 		inputs := generateUTXOInputs(30, 1000000)
	// 		outputs := calculateTxCost(generateUTXOOutputs(400, 1000000))
	// 		txInfos := generateTxInfos(t, cco.Config.TestNetMagic)

	// 		metadata, err := cardano.CreateBatchMetaData(100)
	// 		require.NoError(t, err)
	// 		protocolParams, err := generateProtocolParams()
	// 		require.NoError(t, err)

	// 		slotNumber := uint64(12345)

	// 		result, err := cco.createBatchTx(inputs, metadata, protocolParams, txInfos, outputs, slotNumber)
	// 		require.NoError(t, err)
	// 		require.Less(t, len(result.TxRaw), 16000)
	// 		require.Less(t, len(result.Utxos.MultisigOwnedUTXOs), 30)
	// 		require.Len(t, result.Utxos.FeePayerOwnedUTXOs, len(inputs.FeePayerOwnedUTXOs))
	// 		require.Equal(t, uint64(1000000000), result.Utxos.MultisigOwnedUTXOs[0].Amount)
	// 		require.Equal(t, uint64(1000000), result.Utxos.MultisigOwnedUTXOs[len(result.Utxos.MultisigOwnedUTXOs)-2].Amount)
	// 		require.Equal(t, uint64(1000000), result.Utxos.MultisigOwnedUTXOs[len(result.Utxos.MultisigOwnedUTXOs)-1].Amount)
	// 	})

	// 	t.Run("CreateBatchTx_TxSizeTooBig_IncludeBig2", func(t *testing.T) {
	// 		cco, err := NewCardanoChainOperations(configRaw, &indexer.DatabaseMock{}, secretsMngr, "vector", hclog.NewNullLogger())
	// 		require.NoError(t, err)

	// 		inputs := generateUTXOInputs(30, 1000000)
	// 		outputs := calculateTxCost(generateUTXOOutputs(400, 10000000)) // 4000Ada
	// 		txInfos := generateTxInfos(t, cco.Config.TestNetMagic)

	// 		metadata, err := cardano.CreateBatchMetaData(100)
	// 		require.NoError(t, err)
	// 		protocolParams, err := generateProtocolParams()
	// 		require.NoError(t, err)

	// 		slotNumber := uint64(12345)

	// 		result, err := cco.createBatchTx(inputs, metadata, protocolParams, txInfos, outputs, slotNumber)
	// 		require.NoError(t, err)
	// 		require.Less(t, len(result.TxRaw), 16000)
	// 		require.Less(t, len(result.Utxos.MultisigOwnedUTXOs), 30)
	// 		require.Equal(t, uint64(1000000000), result.Utxos.MultisigOwnedUTXOs[0].Amount)
	// 		require.Equal(t, uint64(2000000000), result.Utxos.MultisigOwnedUTXOs[1].Amount)
	// 		require.Equal(t, uint64(3000000000), result.Utxos.MultisigOwnedUTXOs[2].Amount)
	// 		require.Equal(t, uint64(1000000), result.Utxos.MultisigOwnedUTXOs[len(result.Utxos.MultisigOwnedUTXOs)-2].Amount)
	// 		require.Equal(t, uint64(1000000), result.Utxos.MultisigOwnedUTXOs[len(result.Utxos.MultisigOwnedUTXOs)-1].Amount)
	// 	})

	// 	t.Run("CreateBatchTx_TxSizeTooBig_LargeInput", func(t *testing.T) {
	// 		cco, err := NewCardanoChainOperations(configRaw, &indexer.DatabaseMock{}, secretsMngr, "vector", hclog.NewNullLogger())
	// 		require.NoError(t, err)

	// 		count := 400
	// 		amount := 1000000
	// 		inputs := generateUTXOInputs(count, uint64(amount))
	// 		outputs := calculateTxCost(generateUTXOOutputs(count, uint64(amount)))
	// 		txInfos := generateTxInfos(t, cco.Config.TestNetMagic)

	// 		metadata, err := cardano.CreateBatchMetaData(100)
	// 		require.NoError(t, err)
	// 		protocolParams, err := generateProtocolParams()
	// 		require.NoError(t, err)

	// 		slotNumber := uint64(12345)

	// 		result, err := cco.createBatchTx(inputs, metadata, protocolParams, txInfos, outputs, slotNumber)
	// 		require.NoError(t, err)
	// 		require.Less(t, len(result.TxRaw), 16000)
	// 		require.Less(t, len(result.Utxos.MultisigOwnedUTXOs), 30)
	// 		require.Equal(t, uint64(1000000000), result.Utxos.MultisigOwnedUTXOs[0].Amount)
	// 		require.Equal(t, uint64(1000000), result.Utxos.MultisigOwnedUTXOs[len(result.Utxos.MultisigOwnedUTXOs)-2].Amount)
	// 		require.Equal(t, uint64(1000000), result.Utxos.MultisigOwnedUTXOs[len(result.Utxos.MultisigOwnedUTXOs)-1].Amount)
	// 	})

	// 	t.Run("CreateBatchTx_RandomInputs", func(t *testing.T) {
	// 		cco, err := NewCardanoChainOperations(configRaw, &indexer.DatabaseMock{}, secretsMngr, "vector", hclog.NewNullLogger())
	// 		require.NoError(t, err)

	// 		inputs := generateUTXORandomInputs(100, 1000000, 10000000)
	// 		outputs := calculateTxCost(generateUTXORandomOutputs(200, 1000000, 10000000))
	// 		txInfos := generateTxInfos(t, cco.Config.TestNetMagic)

	// 		metadata, err := cardano.CreateBatchMetaData(100)
	// 		require.NoError(t, err)
	// 		protocolParams, err := generateProtocolParams()
	// 		require.NoError(t, err)

	// 		slotNumber := uint64(12345)

	// 		result, err := cco.createBatchTx(inputs, metadata, protocolParams, txInfos, outputs, slotNumber)
	// 		require.NoError(t, err)
	// 		require.Less(t, len(result.TxRaw), 16000)
	// 		require.LessOrEqual(t, len(result.Utxos.MultisigOwnedUTXOs), 101)

	// 		utxoSum := calculateUTXOSum(result.Utxos.MultisigOwnedUTXOs)
	// 		require.Equal(t, utxoSum.Cmp(outputs.Sum), 1)
	// 	})

	// 	t.Run("CreateBatchTx_MinUtxoOrder", func(t *testing.T) {
	// 		cco, err := NewCardanoChainOperations(configRaw, &indexer.DatabaseMock{}, secretsMngr, "prime", hclog.NewNullLogger())
	// 		require.NoError(t, err)

	// 		inputs := generateUTXOInputsOrdered()                         // 50, 40, 30, 101, 102, 103, 104, 105
	// 		outputs := calculateTxCost(generateUTXOOutputs(403, 1000000)) // 403Ada
	// 		txInfos := generateTxInfos(t, cco.Config.TestNetMagic)

	// 		metadata, err := cardano.CreateBatchMetaData(100)
	// 		require.NoError(t, err)
	// 		protocolParams, err := generateProtocolParams()
	// 		require.NoError(t, err)

	// 		slotNumber := uint64(12345)

	// 		result, err := cco.createBatchTx(inputs, metadata, protocolParams, txInfos, outputs, slotNumber)
	// 		require.NoError(t, err)
	// 		require.Less(t, len(result.TxHash), 16000)
	// 		require.Len(t, result.Utxos.MultisigOwnedUTXOs, 5)
	// 		require.Equal(t, uint64(50000000), result.Utxos.MultisigOwnedUTXOs[0].Amount)
	// 		require.Equal(t, uint64(104000000), result.Utxos.MultisigOwnedUTXOs[1].Amount)
	// 		require.Equal(t, uint64(103000000), result.Utxos.MultisigOwnedUTXOs[2].Amount)
	// 		require.Equal(t, uint64(101000000), result.Utxos.MultisigOwnedUTXOs[3].Amount)
	// 		require.Equal(t, uint64(102000000), result.Utxos.MultisigOwnedUTXOs[4].Amount)
	// 	})
	// }
}

func TestGenerateBatchTransaction(t *testing.T) {
	testDir, err := os.MkdirTemp("", "bat-chain-ops-tx")
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	secretsMngr, err := secretsHelper.CreateSecretsManager(&secrets.SecretsManagerConfig{
		Path: path.Join(testDir, "stp"),
		Type: secrets.Local,
	})
	require.NoError(t, err)

	wallet, err := cardano.GenerateWallet(secretsMngr, "prime", true, false)
	require.NoError(t, err)

	configRaw := json.RawMessage([]byte(`{
			"socketPath": "./socket",
			"testnetMagic": 42
			}`))
	dbMock := &indexer.DatabaseMock{}
	txProviderMock := &cardano.TxProviderTestMock{
		ReturnDefaultParameters: true,
	}

	cco, err := NewCardanoChainOperations(configRaw, dbMock, secretsMngr, "prime", hclog.NewNullLogger())
	require.NoError(t, err)

	cco.txProvider = txProviderMock
	cco.config.SlotRoundingThreshold = 100

	testError := errors.New("test err")
	confirmedTransactions := make([]eth.ConfirmedTransaction, 1)
	confirmedTransactions[0] = eth.ConfirmedTransaction{
		Nonce:       1,
		BlockHeight: big.NewInt(1),
		Receivers: []contractbinding.IBridgeStructsReceiver{{
			DestinationAddress: "addr_test1vqeux7xwusdju9dvsj8h7mca9aup2k439kfmwy773xxc2hcu7zy99",
			Amount:             minUtxoAmount,
		}},
	}
	batchNonceID := uint64(1)
	destinationChain := common.ChainIDStrVector

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	t.Run("GetValidatorsCardanoData returns error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetValidatorsCardanoData", ctx, destinationChain).Return(nil, testError).Once()

		_, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("no vkey for multisig address error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		getValidatorsCardanoDataRet := []contractbinding.IBridgeStructsValidatorCardanoData{
			{
				VerifyingKey:    [32]byte{},
				VerifyingKeyFee: [32]byte{},
			},
		}

		bridgeSmartContractMock.On("GetValidatorsCardanoData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil).Once()

		_, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "verifying key of current batcher wasn't found in validators data queried from smart contract")
	})

	t.Run("no vkey for multisig fee address error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		getValidatorsCardanoDataRet := []contractbinding.IBridgeStructsValidatorCardanoData{
			{
				VerifyingKey:    [32]byte(wallet.MultiSig.GetVerificationKey()),
				VerifyingKeyFee: [32]byte{},
			},
		}

		bridgeSmartContractMock.On("GetValidatorsCardanoData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil).Once()

		_, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "verifying fee key of current batcher wasn't found in validators data queried from smart contract")
	})

	t.Run("GetTip return error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		getValidatorsCardanoDataRet := []contractbinding.IBridgeStructsValidatorCardanoData{
			{
				VerifyingKey:    [32]byte(wallet.MultiSig.GetVerificationKey()),
				VerifyingKeyFee: [32]byte(wallet.MultiSigFee.GetVerificationKey()),
			},
		}

		bridgeSmartContractMock.On("GetValidatorsCardanoData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil).Once()
		txProviderMock.On("GetTip", ctx).Return(nil, testError).Once()

		_, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GetAllTxOutputs return error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		getValidatorsCardanoDataRet := []contractbinding.IBridgeStructsValidatorCardanoData{
			{
				VerifyingKey:    [32]byte(wallet.MultiSig.GetVerificationKey()),
				VerifyingKeyFee: [32]byte(wallet.MultiSigFee.GetVerificationKey()),
			},
		}

		bridgeSmartContractMock.On("GetValidatorsCardanoData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil).Once()
		txProviderMock.On("GetTip", ctx).Return(cardanowallet.QueryTipData{Slot: 50}, nil).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput(nil), testError).Once()

		_, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GenerateBatchTransaction should pass", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		getValidatorsCardanoDataRet := []contractbinding.IBridgeStructsValidatorCardanoData{
			{
				VerifyingKey:    [32]byte(wallet.MultiSig.GetVerificationKey()),
				VerifyingKeyFee: [32]byte(wallet.MultiSigFee.GetVerificationKey()),
			},
		}

		bridgeSmartContractMock.On("GetValidatorsCardanoData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil).Once()
		txProviderMock.On("GetTip", ctx).Return(cardanowallet.QueryTipData{Slot: 50}, nil).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 2000000,
					},
				},
			}, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0xFF"),
					},
					Output: indexer.TxOutput{
						Amount: 20000,
					},
				},
			}, error(nil)).Once()

		result, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.NoError(t, err)
		require.NotNil(t, result.TxRaw)
		require.NotEqual(t, "", result.TxHash)
	})

	t.Run("Test SignBatchTransaction", func(t *testing.T) {
		witnessMultiSig, witnessMultiSigFee, err := cco.SignBatchTransaction(
			"26a9d1a894c7e3719a79342d0fc788989e5d55f076581327c54bcc0c7693905a")
		require.NoError(t, err)
		require.NotNil(t, witnessMultiSig)
		require.NotNil(t, witnessMultiSigFee)
	})
}

// func calculateTxCost(outputs []cardanowallet.TxOutput) cardano.TxOutputs {
// 	txCost := big.NewInt(0)
// 	for _, o := range outputs {
// 		txCost.Add(txCost, big.NewInt(int64(o.Amount)))
// 	}

// 	return cardano.TxOutputs{
// 		Outputs: outputs,
// 		Sum:     txCost,
// 	}
// }

// func calculateUTXOSum(inputs []eth.UTXO) *big.Int {
// 	txCost := big.NewInt(0)
// 	for _, i := range inputs {
// 		txCost.Add(txCost, new(big.Int).SetUint64(i.Amount))
// 	}

// 	return txCost
// }

// func generateUTXOInputs(count int, amount uint64) (inputs eth.UTXOs) {
// 	// Count x Input Ada, 1000Ada, 2000Ada, 3000Ada, 4000Ada, 5000Ada
// 	inputs = eth.UTXOs{
// 		MultisigOwnedUTXOs: make([]eth.UTXO, count+6),
// 		FeePayerOwnedUTXOs: []eth.UTXO{
// 			{
// 				Nonce:   1000,
// 				TxHash:  common.MustHashToBytes32("d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5"),
// 				TxIndex: 1000,
// 				Amount:  10000000,
// 			},
// 			{
// 				Nonce:   1001,
// 				TxHash:  common.MustHashToBytes32("d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5"),
// 				TxIndex: 1001,
// 				Amount:  10000000,
// 			},
// 		},
// 	}

// 	for i := 0; i < count; i++ {
// 		inputs.MultisigOwnedUTXOs[i] = eth.UTXO{
// 			Nonce:   uint64(i),
// 			TxHash:  common.MustHashToBytes32("d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5"),
// 			TxIndex: uint64(i),
// 			Amount:  amount,
// 		}
// 	}

// 	for i := 0; i < 5; i++ {
// 		inputs.MultisigOwnedUTXOs[count+i] = eth.UTXO{
// 			Nonce:   uint64(count + i),
// 			TxHash:  common.MustHashToBytes32("d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5"),
// 			TxIndex: uint64(count + i),
// 			Amount:  uint64(1000000000 * (i + 1)),
// 		}
// 	}

// 	inputs.MultisigOwnedUTXOs[count+5] = eth.UTXO{
// 		Nonce:   uint64(count + 5),
// 		TxHash:  common.MustHashToBytes32("d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5"),
// 		TxIndex: uint64(count + 5),
// 		Amount:  1000000000000,
// 	}

// 	return
// }

// func generateUTXORandomInputs(count int, min uint64, max uint64) (inputs eth.UTXOs) {
// 	// Count x [min-max] Ada, 1000000Ada
// 	inputs = eth.UTXOs{
// 		MultisigOwnedUTXOs: make([]eth.UTXO, count+1),
// 		FeePayerOwnedUTXOs: []eth.UTXO{
// 			{
// 				Nonce:   1000,
// 				TxHash:  common.MustHashToBytes32("d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5"),
// 				TxIndex: 1000,
// 				Amount:  10000000,
// 			},
// 			{
// 				Nonce:   1001,
// 				TxHash:  common.MustHashToBytes32("d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5"),
// 				TxIndex: 1001,
// 				Amount:  10000000,
// 			},
// 		},
// 	}

// 	for i := 0; i < count; i++ {
// 		randomAmount := rand.Uint64() % max
// 		if randomAmount < min {
// 			randomAmount += min
// 		}

// 		inputs.MultisigOwnedUTXOs[i] = eth.UTXO{
// 			Nonce:   uint64(i),
// 			TxHash:  common.MustHashToBytes32("d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5"),
// 			TxIndex: uint64(i),
// 			Amount:  randomAmount,
// 		}
// 	}

// 	inputs.MultisigOwnedUTXOs[count] = eth.UTXO{
// 		Nonce:   uint64(count + 5),
// 		TxHash:  common.MustHashToBytes32("d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5"),
// 		TxIndex: uint64(count + 5),
// 		Amount:  1000000000000,
// 	}

// 	return
// }

// func generateUTXOInputsOrdered() (inputs eth.UTXOs) {
// 	// Count x Input Ada, 1000Ada, 2000Ada, 3000Ada, 4000Ada, 5000Ada
// 	inputs = eth.UTXOs{
// 		MultisigOwnedUTXOs: make([]eth.UTXO, 8),
// 		FeePayerOwnedUTXOs: []eth.UTXO{
// 			{
// 				Nonce:   1000,
// 				TxHash:  common.MustHashToBytes32("d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5"),
// 				TxIndex: 1000,
// 				Amount:  10000000,
// 			},
// 			{
// 				Nonce:   1001,
// 				TxHash:  common.MustHashToBytes32("d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5"),
// 				TxIndex: 1001,
// 				Amount:  10000000,
// 			},
// 		},
// 	}
// 	inputs.MultisigOwnedUTXOs[0] = eth.UTXO{
// 		Nonce:   uint64(0),
// 		TxHash:  common.MustHashToBytes32("d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5"),
// 		TxIndex: 0,
// 		Amount:  50000000,
// 	}
// 	inputs.MultisigOwnedUTXOs[1] = eth.UTXO{
// 		Nonce:   uint64(1),
// 		TxHash:  common.MustHashToBytes32("d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5"),
// 		TxIndex: 0,
// 		Amount:  40000000,
// 	}
// 	inputs.MultisigOwnedUTXOs[2] = eth.UTXO{
// 		Nonce:   uint64(2),
// 		TxHash:  common.MustHashToBytes32("d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5"),
// 		TxIndex: 0,
// 		Amount:  30000000,
// 	}
// 	inputs.MultisigOwnedUTXOs[3] = eth.UTXO{
// 		Nonce:   uint64(3),
// 		TxHash:  common.MustHashToBytes32("d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5"),
// 		TxIndex: 0,
// 		Amount:  101000000,
// 	}
// 	inputs.MultisigOwnedUTXOs[4] = eth.UTXO{
// 		Nonce:   uint64(3),
// 		TxHash:  common.MustHashToBytes32("d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5"),
// 		TxIndex: 0,
// 		Amount:  102000000,
// 	}
// 	inputs.MultisigOwnedUTXOs[5] = eth.UTXO{
// 		Nonce:   uint64(5),
// 		TxHash:  common.MustHashToBytes32("d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5"),
// 		TxIndex: 0,
// 		Amount:  103000000,
// 	}
// 	inputs.MultisigOwnedUTXOs[6] = eth.UTXO{
// 		Nonce:   uint64(6),
// 		TxHash:  common.MustHashToBytes32("d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5"),
// 		TxIndex: 0,
// 		Amount:  104000000,
// 	}
// 	inputs.MultisigOwnedUTXOs[7] = eth.UTXO{
// 		Nonce:   uint64(7),
// 		TxHash:  common.MustHashToBytes32("d50577e2ff7b6df8e37beb178f86837284673a78977a45b065fec457995998b5"),
// 		TxIndex: 0,
// 		Amount:  105000000,
// 	}

// 	return
// }

// func generateUTXOOutputs(count int, amount uint64) (outputs []cardanowallet.TxOutput) {
// 	outputs = make([]cardanowallet.TxOutput, count)
// 	for i := 0; i < count; i++ {
// 		outputs[i] = cardanowallet.TxOutput{
// 			Addr:   "addr_test1vq7vsmgan2adwapu6r3xs5049s6dsf8hlgex68mwgxzraks4c0dpp",
// 			Amount: amount,
// 		}
// 	}

// 	return
// }

// func generateUTXORandomOutputs(count int, min uint64, max uint64) (outputs []cardanowallet.TxOutput) {
// 	outputs = make([]cardanowallet.TxOutput, count)

// 	for i := 0; i < count; i++ {
// 		randomAmount := rand.Uint64() % max
// 		if randomAmount < min {
// 			randomAmount += min
// 		}

// 		outputs[i] = cardanowallet.TxOutput{
// 			Addr:   "addr_test1vq7vsmgan2adwapu6r3xs5049s6dsf8hlgex68mwgxzraks4c0dpp",
// 			Amount: randomAmount,
// 		}
// 	}

// 	return
// }

// func generateTxInfos(t *testing.T, testnetMagic uint32) cardano.TxInputInfos {
// 	t.Helper()

// 	dummyKeyHashes := []string{
// 		"eff5e22355217ec6d770c3668010c2761fa0863afa12e96cff8a2205",
// 		"ad8e0ab92e1febfcaf44889d68c3ae78b59dc9c5fa9e05a272214c13",
// 		"bfd1c0eb0a453a7b7d668166ce5ca779c655e09e11487a6fac72dd6f",
// 		"b4689f2e8f37b406c5eb41b1fe2c9e9f4eec2597c3cc31b8dfee8f56",
// 		"39c196d28f804f70704b6dec5991fbb1112e648e067d17ca7abe614b",
// 		"adea661341df075349cbb2ad02905ce1828f8cf3e66f5012d48c3168",
// 	}

// 	multisigPolicyScript, err := cardanowallet.NewPolicyScript(dummyKeyHashes[0:3], 3)
// 	require.NoError(t, err)
// 	multisigFeePolicyScript, err := cardanowallet.NewPolicyScript(dummyKeyHashes[3:], 3)
// 	require.NoError(t, err)

// 	multisigAddress, err := multisigPolicyScript.CreateMultiSigAddress(uint(testnetMagic))
// 	require.NoError(t, err)
// 	multisigFeeAddress, err := multisigFeePolicyScript.CreateMultiSigAddress(uint(testnetMagic))
// 	require.NoError(t, err)

// 	return cardano.TxInputInfos{
// 		MultiSig: &cardano.TxInputInfo{
// 			PolicyScript: multisigPolicyScript,
// 			Address:      multisigAddress,
// 		},
// 		MultiSigFee: &cardano.TxInputInfo{
// 			PolicyScript: multisigFeePolicyScript,
// 			Address:      multisigFeeAddress,
// 		},
// 	}
// }

// func generateProtocolParams() ([]byte, error) {
// 	resultJSON := map[string]interface{}{
// 		"collateralPercentage": 150,
// 		"costModels":           nil,
// 		"decentralization":     nil,
// 		"executionUnitPrices": map[string]interface{}{
// 			"priceMemory": 5.77e-2,
// 			"priceSteps":  7.21e-5,
// 		},
// 		"extraPraosEntropy": nil,
// 		"maxBlockBodySize":  65536,
// 		"maxBlockExecutionUnits": map[string]interface{}{
// 			"memory": 80000000,
// 			"steps":  40000000000,
// 		},
// 		"maxBlockHeaderSize":  1100,
// 		"maxCollateralInputs": 3,
// 		"maxTxExecutionUnits": map[string]interface{}{
// 			"memory": 16000000,
// 			"steps":  10000000000,
// 		},
// 		"maxTxSize":           16384,
// 		"maxValueSize":        5000,
// 		"minPoolCost":         0,
// 		"minUTxOValue":        nil,
// 		"monetaryExpansion":   0.1,
// 		"poolPledgeInfluence": 0,
// 		"poolRetireMaxEpoch":  18,
// 		"protocolVersion": map[string]interface{}{
// 			"major": 7,
// 			"minor": 0,
// 		},
// 		"stakeAddressDeposit": 0,
// 		"stakePoolDeposit":    0,
// 		"stakePoolTargetNum":  100,
// 		"treasuryCut":         0.1,
// 		"txFeeFixed":          155381,
// 		"txFeePerByte":        44,
// 		"utxoCostPerByte":     4310,
// 	}

// return json.Marshal(resultJSON)}

func Test_getSlotNumberWithRoundingThreshold(t *testing.T) {
	_, err := getSlotNumberWithRoundingThreshold(66, 60, 0.125)
	assert.ErrorIs(t, err, errNonActiveBatchPeriod)

	_, err = getSlotNumberWithRoundingThreshold(12, 60, 0.2)
	assert.ErrorIs(t, err, errNonActiveBatchPeriod)

	_, err = getSlotNumberWithRoundingThreshold(115, 60, 0.125)
	assert.ErrorIs(t, err, errNonActiveBatchPeriod)

	_, err = getSlotNumberWithRoundingThreshold(224, 80, 0.2)
	assert.ErrorIs(t, err, errNonActiveBatchPeriod)

	_, err = getSlotNumberWithRoundingThreshold(336, 80, 0.2)
	assert.ErrorIs(t, err, errNonActiveBatchPeriod)

	_, err = getSlotNumberWithRoundingThreshold(0, 60, 0.125)
	assert.ErrorContains(t, err, "slot number is zero")

	val, err := getSlotNumberWithRoundingThreshold(75, 60, 0.125)
	assert.NoError(t, err)
	assert.Equal(t, uint64(120), val)

	val, err = getSlotNumberWithRoundingThreshold(105, 60, 0.125)
	assert.NoError(t, err)
	assert.Equal(t, uint64(120), val)

	val, err = getSlotNumberWithRoundingThreshold(40, 60, 0.125)
	assert.NoError(t, err)
	assert.Equal(t, uint64(60), val)

	val, err = getSlotNumberWithRoundingThreshold(270, 80, 0.125)
	assert.NoError(t, err)
	assert.Equal(t, uint64(320), val)

	val, err = getSlotNumberWithRoundingThreshold(223, 80, 0.2)
	assert.NoError(t, err)
	assert.Equal(t, uint64(240), val)

	val, err = getSlotNumberWithRoundingThreshold(337, 80, 0.2)
	assert.NoError(t, err)
	assert.Equal(t, uint64(400), val)
}

func Test_getNeededUtxos(t *testing.T) {
	inputs := []*indexer.TxInputOutput{
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0x1"), Index: 100},
			Output: indexer.TxOutput{Amount: 10},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0x1"), Index: 0},
			Output: indexer.TxOutput{Amount: 20},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0x2"), Index: 7},
			Output: indexer.TxOutput{Amount: 5},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0x4"), Index: 5},
			Output: indexer.TxOutput{Amount: 30},
		},
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0x4"), Index: 6},
			Output: indexer.TxOutput{Amount: 15},
		},
	}

	t.Run("pass", func(t *testing.T) {
		result, err := getNeededUtxos(inputs, 65, 5, 5, 30, 1)

		require.NoError(t, err)
		require.Equal(t, inputs[:len(inputs)-1], result)

		result, err = getNeededUtxos(inputs, 50, 6, 0, 2, 1)

		require.NoError(t, err)
		require.Equal(t, []*indexer.TxInputOutput{inputs[3], inputs[1]}, result)
	})

	t.Run("pass with change", func(t *testing.T) {
		result, err := getNeededUtxos(inputs, 67, 4, 5, 30, 1)

		require.NoError(t, err)
		require.Equal(t, inputs, result)
	})

	t.Run("pass with at least", func(t *testing.T) {
		result, err := getNeededUtxos(inputs, 10, 4, 5, 30, 3)

		require.NoError(t, err)
		require.Equal(t, inputs[:3], result)
	})

	t.Run("not enough sum", func(t *testing.T) {
		_, err := getNeededUtxos(inputs, 160, 5, 5, 30, 1)
		require.ErrorContains(t, err, "couldn't select UTXOs for sum")
	})
}

func Test_getOutputs(t *testing.T) {
	txs := []eth.ConfirmedTransaction{
		{
			Receivers: []contractbinding.IBridgeStructsReceiver{
				{
					DestinationAddress: "0x1",
					Amount:             100,
				},
				{
					DestinationAddress: "0x2",
					Amount:             200,
				},
				{
					DestinationAddress: "0x3",
					Amount:             400,
				},
			},
		},
		{
			Receivers: []contractbinding.IBridgeStructsReceiver{
				{
					DestinationAddress: "0x4",
					Amount:             50,
				},
				{
					DestinationAddress: "0x3",
					Amount:             900,
				},
				{
					DestinationAddress: "0x11",
					Amount:             0,
				},
			},
		},
		{
			Receivers: []contractbinding.IBridgeStructsReceiver{
				{
					DestinationAddress: "0x5",
					Amount:             3000,
				},
			},
		},
		{
			Receivers: []contractbinding.IBridgeStructsReceiver{
				{
					DestinationAddress: "0x1",
					Amount:             2000,
				},
				{
					DestinationAddress: "0x4",
					Amount:             170,
				},
				{
					DestinationAddress: "0x3",
					Amount:             10,
				},
			},
		},
	}

	res := getOutputs(txs)

	assert.Equal(t, big.NewInt(6830), res.Sum)
	assert.Equal(t, []cardanowallet.TxOutput{
		{
			Addr:   "0x1",
			Amount: 2100,
		},
		{
			Addr:   "0x2",
			Amount: 200,
		},
		{
			Addr:   "0x3",
			Amount: 1310,
		},
		{
			Addr:   "0x4",
			Amount: 220,
		},
		{
			Addr:   "0x5",
			Amount: 3000,
		},
	}, res.Outputs)
}
