package batcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGenerateBatchTransaction(t *testing.T) {
	testDir, err := os.MkdirTemp("", "bat-chain-ops-tx")
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	slotThreshold := uint64(40)
	wallets := [3]*cardano.CardanoWallet{}
	multisigKeyHashes := [3]string{}
	feeKeyHashes := [3]string{}
	getValidatorsCardanoDataValidRet := make([]contractbinding.IBridgeStructsValidatorCardanoData, len(wallets))

	for i := range wallets {
		wallets[i], err = cardano.GenerateWallet(filepath.Join(testDir, fmt.Sprint(i)), false, false)
		require.NoError(t, err)

		multisigKeyHashes[i], err = cardanowallet.GetKeyHash(wallets[i].MultiSig.GetVerificationKey())
		require.NoError(t, err)

		feeKeyHashes[i], err = cardanowallet.GetKeyHash(wallets[i].MultiSigFee.GetVerificationKey())
		require.NoError(t, err)

		getValidatorsCardanoDataValidRet[i] = contractbinding.IBridgeStructsValidatorCardanoData{
			VerifyingKey:    [32]byte(wallets[i].MultiSig.GetVerificationKey()),
			VerifyingKeyFee: [32]byte(wallets[i].MultiSigFee.GetVerificationKey()),
		}
	}

	configRaw := json.RawMessage([]byte(fmt.Sprintf(`{
		"socketPath": "./socket",
		"slotRoundingThreshold": %d,
		"testnetMagic": 42,
		"keysDirPath": "%s"
		}`, slotThreshold, filepath.Join(testDir, "0"))))

	cco, err := NewCardanoChainOperations(configRaw, hclog.NewNullLogger())
	require.NoError(t, err)

	txProviderMock := &cardano.TxProviderTestMock{
		ReturnDefaultParameters: true,
	}
	cco.TxProvider = txProviderMock

	testError := errors.New("test err")
	confirmedTransactions := []eth.ConfirmedTransaction{
		{
			Nonce:       1,
			BlockHeight: big.NewInt(1),
			Receivers: []contractbinding.IBridgeStructsReceiver{{
				DestinationAddress: "addr_test1vqeux7xwusdju9dvsj8h7mca9aup2k439kfmwy773xxc2hcu7zy99",
				Amount:             minUtxoAmount,
			}},
		},
	}
	batchNonceID := uint64(1)
	destinationChain := common.ChainIDStrVector

	txInputInfos, err := cco.createTxInfos(multisigKeyHashes[:], feeKeyHashes[:])
	require.NoError(t, err)

	multisigAddr := txInputInfos.MultiSig.Address
	feeAddr := txInputInfos.MultiSigFee.Address

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	t.Run("GetBlockNumber returns error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetBlockNumber", ctx).Return(nil, testError).Once()
		bridgeSmartContractMock.On("GetValidatorsCardanoData", ctx, destinationChain).Return(getValidatorsCardanoDataValidRet, nil).Once()

		_, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.ErrorIs(t, err, testError)
	})

	t.Run("GetValidatorsCardanoData returns error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetValidatorsCardanoData", ctx, destinationChain).Return(nil, testError).Once()

		_, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.ErrorIs(t, err, testError)
	})

	t.Run("no vkey for multisig address error", func(t *testing.T) {
		getValidatorsCardanoDataRet := []contractbinding.IBridgeStructsValidatorCardanoData{
			{
				VerifyingKey:    [32]byte{},
				VerifyingKeyFee: [32]byte{},
			},
		}

		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetValidatorsCardanoData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil).Once()

		_, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.ErrorIs(t, err, errBatchKeyNotFound)
		require.ErrorContains(t, err, "multisig:")
	})

	t.Run("no vkey for fee address error", func(t *testing.T) {
		getValidatorsCardanoDataRet := []contractbinding.IBridgeStructsValidatorCardanoData{
			{
				VerifyingKey:    [32]byte(wallets[0].MultiSig.GetVerificationKey()),
				VerifyingKeyFee: [32]byte{},
			},
		}

		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetValidatorsCardanoData", ctx, destinationChain).Return(getValidatorsCardanoDataRet, nil).Once()

		_, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.ErrorIs(t, err, errBatchKeyNotFound)
		require.ErrorContains(t, err, "fee:")
	})

	t.Run("GetBatchProposerData return error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetBlockNumber", ctx).Return(uint64(10), error(nil)).Once()
		bridgeSmartContractMock.On("GetValidatorsCardanoData", ctx, destinationChain).Return(getValidatorsCardanoDataValidRet, nil).Once()
		bridgeSmartContractMock.On("GetBatchProposerData", ctx, destinationChain).Return(nil, testError).Once()

		_, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.ErrorIs(t, err, testError)
	})

	t.Run("GetSlot return error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetBlockNumber", ctx).Return(uint64(10), error(nil)).Once()
		bridgeSmartContractMock.On("GetValidatorsCardanoData", ctx, destinationChain).Return(getValidatorsCardanoDataValidRet, nil).Once()
		bridgeSmartContractMock.On("GetBatchProposerData", ctx, destinationChain).Return(eth.BatchProposerData{}, error(nil)).Once()

		txProviderMock.On("GetTip", ctx).Return(nil, testError).Once()

		_, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.ErrorIs(t, err, testError)
	})

	t.Run("GetAvailableUTXOs return error", func(t *testing.T) {
		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetBlockNumber", ctx).Return(uint64(3), error(nil)).Once()
		bridgeSmartContractMock.On("GetValidatorsCardanoData", ctx, destinationChain).Return(getValidatorsCardanoDataValidRet, nil).Once()
		bridgeSmartContractMock.On("GetBatchProposerData", ctx, destinationChain).Return(eth.BatchProposerData{}, error(nil)).Once()

		txProviderMock.On("GetTip", ctx).Return(cardanowallet.QueryTipData{
			Slot: uint64(1000),
		}, error(nil)).Once()
		txProviderMock.On("GetUtxos", ctx, mock.Anything).Return([]cardanowallet.Utxo{}, testError).Once()

		_, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.ErrorIs(t, err, testError)
	})

	t.Run("GenerateBatchTransaction should pass for proposer - proposal not set", func(t *testing.T) {
		const (
			blockNumber = uint64(10)
			slot        = uint64(38927)
			txHash1     = "26a9d1a894c7e3719a79342d0fc788989e5d55f076581327c54bcc0c7693905a"
			txHash2     = "26a9d1a894c7e3719a79342d0fc788989e5d55f076581327c54bcc0c7693905a"
			txHash3     = "Fa9d1a894c7e3719aF9342d0fc788ED9e5d5530765813AAc54bcc0c7693905aB"
			txIndex1    = uint32(0)
			txIndex2    = uint32(17)
			txIndex3    = uint32(45)
		)

		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetBlockNumber", ctx).Return(blockNumber, error(nil)).Once()
		bridgeSmartContractMock.On("GetValidatorsCardanoData", ctx, destinationChain).Return(getValidatorsCardanoDataValidRet, nil).Once()
		bridgeSmartContractMock.On("GetBatchProposerData", ctx, destinationChain).Return(eth.BatchProposerData{}, error(nil)).Once()

		txProviderMock.On("GetTip", ctx).Return(cardanowallet.QueryTipData{
			Slot: slot,
		}, error(nil)).Once()
		txProviderMock.On("GetUtxos", ctx, multisigAddr).Return([]cardanowallet.Utxo{
			{
				Hash:   txHash3,
				Index:  txIndex3,
				Amount: 50_000_245,
			},
			{
				Hash:   txHash1,
				Index:  txIndex1,
				Amount: 10_000_000_000,
			},
		}, error(nil)).Once()
		txProviderMock.On("GetUtxos", ctx, feeAddr).Return([]cardanowallet.Utxo{
			{
				Hash:   txHash2,
				Index:  txIndex2,
				Amount: 4_000_000,
			},
		}, error(nil)).Once()

		result, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.NoError(t, err)
		require.NotNil(t, result.TxRaw)
		require.NotEqual(t, "", result.TxHash)
		require.Equal(t, 0, result.ProposerIdx)
		require.Equal(t, 0, result.ValidatorIdx)
		require.Equal(t, blockNumber, result.BlockNumber)
		require.Equal(t, getRoundedSlot(slot, slotThreshold), result.Proposal.Slot)
		require.Equal(t, []eth.UTXO{
			{
				TxHash:  indexer.NewHashFromHexString(txHash1),
				TxIndex: uint64(txIndex1),
			},
			{
				TxHash:  indexer.NewHashFromHexString(txHash3),
				TxIndex: uint64(txIndex3),
			},
		}, result.Proposal.MultisigUTXOs)
		require.Equal(t, []eth.UTXO{
			{
				TxHash:  indexer.NewHashFromHexString(txHash2),
				TxIndex: uint64(txIndex2),
			},
		}, result.Proposal.FeePayerUTXOs)
	})

	t.Run("GenerateBatchTransaction should pass for proposer - invalid proposal already set", func(t *testing.T) {
		const (
			blockNumber = uint64(10)
			slot        = uint64(38927)
			txHash1     = "26a9d1a894c7e3719a79342d0fc788989e5d55f076581327c54bcc0c7693905a"
			txHash2     = "26a9d1a894c7e3719a79342d0fc788989e5d55f076581327c54bcc0c7693905a"
			txHash3     = "Fa9d1a894c7e3719aF9342d0fc788ED9e5d5530765813AAc54bcc0c7693905aB"
			txIndex1    = uint32(0)
			txIndex2    = uint32(17)
			txIndex3    = uint32(45)
		)

		batchProposerMultisigUtxos := []eth.UTXO{
			{
				TxHash:  indexer.NewHashFromHexString(txHash1),
				TxIndex: uint64(txIndex1),
			},
			{
				TxHash:  indexer.NewHashFromHexString(txHash3),
				TxIndex: uint64(txIndex3) + 1,
			},
		}
		batchProposerFeeUtxos := []eth.UTXO{
			{
				TxHash:  indexer.NewHashFromHexString(txHash2),
				TxIndex: uint64(txIndex2),
			},
		}

		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetBlockNumber", ctx).Return(blockNumber, error(nil)).Once()
		bridgeSmartContractMock.On("GetValidatorsCardanoData", ctx, destinationChain).Return(getValidatorsCardanoDataValidRet, nil).Once()
		bridgeSmartContractMock.On("GetBatchProposerData", ctx, destinationChain).Return(
			eth.BatchProposerData{
				Slot:          getRoundedSlot(slot, slotThreshold),
				MultisigUTXOs: batchProposerMultisigUtxos,
				FeePayerUTXOs: batchProposerFeeUtxos,
			},
			error(nil),
		).Once()

		txProviderMock.On("GetTip", ctx).Return(cardanowallet.QueryTipData{
			Slot: slot,
		}, error(nil)).Once()
		txProviderMock.On("GetUtxos", ctx, multisigAddr).Return([]cardanowallet.Utxo{
			{
				Hash:   txHash3,
				Index:  txIndex3,
				Amount: 50_000_245,
			},
			{
				Hash:   txHash1,
				Index:  txIndex1,
				Amount: 10_000_000_000,
			},
		}, error(nil)).Once()
		txProviderMock.On("GetUtxos", ctx, feeAddr).Return([]cardanowallet.Utxo{
			{
				Hash:   txHash2,
				Index:  txIndex2,
				Amount: 4_000_000,
			},
		}, error(nil)).Once()

		result, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.NoError(t, err)
		require.NotNil(t, result.TxRaw)
		require.NotEqual(t, "", result.TxHash)
		require.Equal(t, 0, result.ProposerIdx)
		require.Equal(t, 0, result.ValidatorIdx)
		require.Equal(t, blockNumber, result.BlockNumber)
		require.Equal(t, getRoundedSlot(slot, slotThreshold), result.Proposal.Slot)
		require.Equal(t, []eth.UTXO{
			{
				TxHash:  indexer.NewHashFromHexString(txHash1),
				TxIndex: uint64(txIndex1),
			},
			{
				TxHash:  indexer.NewHashFromHexString(txHash3),
				TxIndex: uint64(txIndex3),
			},
		}, result.Proposal.MultisigUTXOs)
		require.Equal(t, []eth.UTXO{
			{
				TxHash:  indexer.NewHashFromHexString(txHash2),
				TxIndex: uint64(txIndex2),
			},
		}, result.Proposal.FeePayerUTXOs)
	})

	t.Run("GenerateBatchTransaction should pass for non proposer", func(t *testing.T) {
		const (
			blockNumber = uint64(33)
			slot        = uint64(133497)
			txHash1     = "26a9d1a894c7e3719a79342d0fc788989e5d55f076581327c54bcc0c7693905a"
			txHash2     = "26a9d1a894c7e3719a79342d0fc788989e5d55f076581327c54bcc0c7693905a"
			txHash3     = "Fa9d1a894c7e3719aF9342d0fc788ED9e5d5530765813AAc54bcc0c7693905aB"
			txIndex1    = uint32(0)
			txIndex2    = uint32(17)
			txIndex3    = uint32(45)
		)

		batchProposerMultisigUtxos := []eth.UTXO{
			{
				TxHash:  indexer.NewHashFromHexString(txHash1),
				TxIndex: uint64(txIndex1),
			},
			{
				TxHash:  indexer.NewHashFromHexString(txHash3),
				TxIndex: uint64(txIndex3),
			},
		}
		batchProposerFeeUtxos := []eth.UTXO{
			{
				TxHash:  indexer.NewHashFromHexString(txHash2),
				TxIndex: uint64(txIndex2),
			},
		}

		bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
		bridgeSmartContractMock.On("GetBlockNumber", ctx).Return(blockNumber, error(nil)).Once()
		bridgeSmartContractMock.On("GetValidatorsCardanoData", ctx, destinationChain).Return(getValidatorsCardanoDataValidRet, nil).Once()
		bridgeSmartContractMock.On("GetBatchProposerData", ctx, destinationChain).Return(
			eth.BatchProposerData{
				Slot:          getRoundedSlot(slot, slotThreshold),
				MultisigUTXOs: batchProposerMultisigUtxos,
				FeePayerUTXOs: batchProposerFeeUtxos,
			},
			error(nil),
		).Once()

		txProviderMock.On("GetTip", ctx).Return(cardanowallet.QueryTipData{
			Slot: slot,
		}, error(nil)).Once()
		txProviderMock.On("GetUtxos", ctx, multisigAddr).Return([]cardanowallet.Utxo{
			{
				Hash:   txHash1,
				Index:  1003, // new tx
				Amount: 410_000_000_000,
			},
			{
				Hash:   txHash3,
				Index:  txIndex3,
				Amount: 50_000_245,
			},
			{
				Hash:   txHash1,
				Index:  txIndex1,
				Amount: 10_000_000_000,
			},
		}, error(nil)).Once()
		txProviderMock.On("GetUtxos", ctx, feeAddr).Return([]cardanowallet.Utxo{
			{
				Hash:   txHash2,
				Index:  1003, // new tx
				Amount: 410_000_000_000,
			},
			{
				Hash:   txHash2,
				Index:  txIndex2,
				Amount: 4_000_000,
			},
		}, error(nil)).Once()

		result, err := cco.GenerateBatchTransaction(ctx, bridgeSmartContractMock, destinationChain, confirmedTransactions, batchNonceID)
		require.NoError(t, err)
		require.NotNil(t, result.TxRaw)
		require.NotEqual(t, "", result.TxHash)
		require.Equal(t, 1, result.ProposerIdx)
		require.Equal(t, 0, result.ValidatorIdx)
		require.Equal(t, blockNumber, result.BlockNumber)
		require.Equal(t, getRoundedSlot(slot, slotThreshold), result.Proposal.Slot)
		require.Equal(t, batchProposerMultisigUtxos, result.Proposal.MultisigUTXOs)
		require.Equal(t, batchProposerFeeUtxos, result.Proposal.FeePayerUTXOs)
	})

	t.Run("Test SignBatchTransaction", func(t *testing.T) {
		witnessMultiSig, witnessMultiSigFee, err := cco.SignBatchTransaction("26a9d1a894c7e3719a79342d0fc788989e5d55f076581327c54bcc0c7693905a")
		require.NoError(t, err)
		require.NotNil(t, witnessMultiSig)
		require.NotNil(t, witnessMultiSigFee)
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

func Test_getRoundedSlot(t *testing.T) {
	assert.Equal(t, uint64(120), getRoundedSlot(66, 60))
	assert.Equal(t, uint64(320), getRoundedSlot(270, 80))
	assert.Equal(t, uint64(120), getRoundedSlot(105, 60))
	assert.Equal(t, uint64(32), getRoundedSlot(32, 32))
	assert.Equal(t, uint64(0), getRoundedSlot(0, 32))
}

func Test_getProposerIndex(t *testing.T) {
	assert.Equal(t, 1, getProposerIndex(66, 60, 5))
	assert.Equal(t, 2, getProposerIndex(426, 60, 5))
	assert.Equal(t, 0, getProposerIndex(180, 60, 3))
}

func Test_validateAndRetrieveUtxos(t *testing.T) {
	inputs := []cardanowallet.Utxo{
		{
			Hash:   "0x1",
			Index:  100,
			Amount: 20,
		},
		{
			Hash:   "0x2",
			Index:  7,
			Amount: 30,
		},
		{
			Hash:   "0x4",
			Index:  5,
			Amount: 10,
		},
		{
			Hash:   "0x1",
			Index:  0,
			Amount: 5,
		},
	}

	desired := []eth.UTXO{
		{
			TxHash:  indexer.NewHashFromHexString("0x1"),
			TxIndex: 0,
		},
		{
			TxHash:  indexer.NewHashFromHexString("0x2"),
			TxIndex: 7,
		},
	}

	t.Run("valid", func(t *testing.T) {
		res, err := validateAndRetrieveUtxos(inputs, desired, 35, 5)

		require.NoError(t, err)
		require.Len(t, res, 2)
		require.Equal(t, res[0], inputs[3])
		require.Equal(t, res[1], inputs[1])

		res, err = validateAndRetrieveUtxos(inputs, desired, 30, 5)

		require.NoError(t, err)
		require.Len(t, res, 2)
		require.Equal(t, res[0], inputs[3])
		require.Equal(t, res[1], inputs[1])
	})

	t.Run("invalid minimal change", func(t *testing.T) {
		_, err := validateAndRetrieveUtxos(inputs, desired, 32, 5)
		require.ErrorContains(t, err, "proposed utxos sum is not good")
	})

	t.Run("invalid sum", func(t *testing.T) {
		_, err := validateAndRetrieveUtxos(inputs, desired, 40, 5)
		require.ErrorContains(t, err, "proposed utxos sum is not good")
	})

	t.Run("invalid utxo", func(t *testing.T) {
		_, err := validateAndRetrieveUtxos(inputs, append([]eth.UTXO{
			{
				TxHash:  indexer.NewHashFromHexString("0x8888"),
				TxIndex: 20,
			},
		}, desired...), 30, 5)
		require.ErrorContains(t, err, "proposed utxo does not exists")

		_, err = validateAndRetrieveUtxos(inputs, append([]eth.UTXO{
			{
				TxHash:  indexer.NewHashFromHexString("0x4"),
				TxIndex: 1220,
			},
		}, desired...), 30, 5)
		require.ErrorContains(t, err, "proposed utxo does not exists")
	})
}

func Test_getNeededUtxos(t *testing.T) {
	inputs := []cardanowallet.Utxo{
		{
			Hash:   "0x1",
			Index:  100,
			Amount: 30,
		},
		{
			Hash:   "0x1",
			Index:  0,
			Amount: 20,
		},
		{
			Hash:   "0x2",
			Index:  7,
			Amount: 10,
		},
		{
			Hash:   "0x4",
			Index:  5,
			Amount: 5,
		},
	}

	t.Run("pass", func(t *testing.T) {
		result, err := getNeededUtxos(inputs, 65, 5, 5, 30, 1)

		require.NoError(t, err)
		require.Equal(t, inputs, result)

		result, err = getNeededUtxos(inputs, 50, 6, 5, 30, 1)

		require.NoError(t, err)
		require.Equal(t, inputs[:2], result)
	})

	t.Run("pass with change", func(t *testing.T) {
		result, err := getNeededUtxos(inputs, 61, 4, 5, 30, 1)

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
		require.ErrorContains(t, err, "could not select utxos for sum")
	})

	t.Run("max utxo count is reached", func(t *testing.T) {
		_, err := getNeededUtxos(inputs, 60, 5, 5, 7, 1)
		require.ErrorContains(t, err, "max utxo count is reached")
	})
}
