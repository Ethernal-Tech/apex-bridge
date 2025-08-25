package batcher

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	bac "github.com/Ethernal-Tech/apex-bridge/bridging_addresses_coordinator"
	bam "github.com/Ethernal-Tech/apex-bridge/bridging_addresses_manager"
	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	secretsHelper "github.com/Ethernal-Tech/cardano-infrastructure/secrets/helper"
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
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
	scBlock1 := eth.CardanoBlock{
		BlockSlot: big.NewInt(15),
	}
	scBlock2 := eth.CardanoBlock{
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

	bridgeSmartContractMock.On("GetLastObservedBlock", ctx, chainID).Return(eth.CardanoBlock{}, testErr1).Once()
	bridgeSmartContractMock.On("GetLastObservedBlock", ctx, chainID).Return(eth.CardanoBlock{}, nil).Once()
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

	// database error
	_, err = cco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
	require.ErrorIs(t, err, testErr2)

	// not in sync
	val, err := cco.IsSynchronized(ctx, bridgeSmartContractMock, chainID)
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

func TestGenerateBatchTransaction(t *testing.T) {
	testDir, err := os.MkdirTemp("", "bat-chain-ops-tx")
	require.NoError(t, err)

	minUtxoAmount := new(big.Int).SetUint64(1_000)

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	secretsMngr, err := secretsHelper.CreateSecretsManager(&secrets.SecretsManagerConfig{
		Path: filepath.Join(testDir, "stp"),
		Type: secrets.Local,
	})
	require.NoError(t, err)

	_, err = cardano.GenerateWallet(secretsMngr, "prime", true, false)
	require.NoError(t, err)

	configRaw := json.RawMessage([]byte(`{
			"socketPath": "./socket",
			"testnetMagic": 42,
			"minUtxoAmount": 1000
			}`))
	dbMock := &indexer.DatabaseMock{}
	txProviderMock := &cardano.TxProviderTestMock{
		ReturnDefaultParameters: true,
	}

	bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
	bridgingAddressCoordinatorMock := &bac.BridgingAddressesCoordinatorMock{}

	cco, err := NewCardanoChainOperations(configRaw, dbMock, secretsMngr, "prime", bridgingAddressesManagerMock, bridgingAddressCoordinatorMock, hclog.NewNullLogger())
	require.NoError(t, err)

	cco.txProvider = txProviderMock
	cco.config.SlotRoundingThreshold = 100
	cco.config.MaxFeeUtxoCount = 4
	cco.config.MaxUtxoCount = 50

	testError := errors.New("test err")
	confirmedTransactions := make([]eth.ConfirmedTransaction, 1)
	confirmedTransactions[0] = eth.ConfirmedTransaction{
		Nonce:       1,
		BlockHeight: big.NewInt(1),
		Receivers: []eth.BridgeReceiver{{
			DestinationAddress: "addr_test1vqeux7xwusdju9dvsj8h7mca9aup2k439kfmwy773xxc2hcu7zy99",
			Amount:             minUtxoAmount,
		}},
		TransactionType: uint8(common.BridgingConfirmedTxType),
	}
	batchNonceID := uint64(1)
	destinationChain := common.ChainIDStrVector
	cliUtils := cardanowallet.NewCliUtils(cardanowallet.ResolveCardanoCliBinary(cardanowallet.MainNetNetwork))
	keys := []string{
		"846d5cb85238b2f433e3a35f1df61a4fbc2a69a705e5bbcb626ce9ae",
		"5a12073b94bfbdfcbb7cb985eed1f35f5eeafaf57912905f654f41bf",
		"adefefdd0dec7d8044285998b51d6cf39b4d8fb613c695045d021ac4",
	}
	script := cardanowallet.NewPolicyScript(keys, 2, cardanowallet.WithAfter(0))
	bridgingAddr, err := cliUtils.GetPolicyScriptBaseAddress(cardanowallet.TestNetProtocolMagic, script, script)
	require.NoError(t, err)

	keys = []string{
		"846d5cb85238b2f433e3a35f1df61a4fbc2a69a705e5bbcb626ce9ae",
		"adefefdd0dec7d8044285998b51d6cf39b4d8fb613c695045d021ac4",
		"5a12073b94bfbdfcbb7cb985eed1f35f5eeafaf57912905f654f41bf",
	}
	script = cardanowallet.NewPolicyScript(keys, 2, cardanowallet.WithAfter(0))
	feeAddr, err := cliUtils.GetPolicyScriptBaseAddress(cardanowallet.TestNetProtocolMagic, script, script)
	require.NoError(t, err)

	addressAndAmountRet := []common.AddressAndAmount{
		{
			AddressIndex:  0,
			Address:       bridgingAddr,
			TokensAmounts: map[string]uint64{"lovelace": minUtxoAmount.Uint64()},
		},
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	t.Run("GetAddressesAndAmountsToPayFrom returns error", func(t *testing.T) {
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]common.AddressAndAmount(nil), testError).Once()

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GetAllTxOutputs return error", func(t *testing.T) {
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput(nil), testError).Once()

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GenerateBatchTransaction fee multisig does not have any utxo", func(t *testing.T) {
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).
			Return([]*indexer.TxInputOutput{}, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", bridgingAddr, true).
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

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.ErrorContains(t, err, "fee multisig does not have any utxo")
	})

	t.Run("GetLatestBlockPoint return error", func(t *testing.T) {
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr).Once()
		dbMock.On("GetAllTxOutputs", bridgingAddr, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 4000000,
					},
				},
			}, error(nil))
		dbMock.On("GetAllTxOutputs", feeAddr, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 2000000,
					},
				},
			}, error(nil))
		dbMock.On("GetLatestBlockPoint").Return((*indexer.BlockPoint)(nil), testError).Once()

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GetPaymentPolicyScript return false", func(t *testing.T) {
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		bridgingAddressesManagerMock.On("GetPaymentPolicyScript", mock.Anything, mock.Anything).Return(nil, false).Once()

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get payment policy script for address")
	})

	t.Run("GetPaymentAddressFromIndex return false", func(t *testing.T) {
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		bridgingAddressesManagerMock.On("GetPaymentPolicyScript", mock.Anything, mock.Anything).Return(script, true).Once()
		bridgingAddressesManagerMock.On("GetPaymentAddressFromIndex", mock.Anything, mock.Anything).Return("", false).Once()

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get payment address for address index")
	})

	t.Run("GetFeePolicyScript return false", func(t *testing.T) {
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		bridgingAddressesManagerMock.On("GetPaymentPolicyScript", mock.Anything, mock.Anything).Return(script, true).Once()
		bridgingAddressesManagerMock.On("GetPaymentAddressFromIndex", mock.Anything, mock.Anything).Return(bridgingAddr, true).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigPolicyScript", mock.Anything, mock.Anything).Return(nil, false).Once()

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get fee policy script for chain")
	})

	t.Run("should pass", func(t *testing.T) {
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		bridgingAddressesManagerMock.On("GetPaymentPolicyScript", mock.Anything, mock.Anything).Return(script, true).Once()
		bridgingAddressesManagerMock.On("GetPaymentAddressFromIndex", mock.Anything, mock.Anything).Return(bridgingAddr, true).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigPolicyScript", mock.Anything, mock.Anything).Return(script, true).Once()

		result, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.NoError(t, err)
		require.NotNil(t, result.TxRaw)
		require.NotEqual(t, "", result.TxHash)
	})

	t.Run("should pass with redistribution", func(t *testing.T) {
		redistributionConfirmedTx := make([]eth.ConfirmedTransaction, 1)
		redistributionConfirmedTx[0] = eth.ConfirmedTransaction{
			Nonce:           1,
			BlockHeight:     big.NewInt(1),
			TransactionType: uint8(common.RedistributionConfirmedTxType),
		}

		bridgingAddressCoordinatorMock.On("GetAddressesAndAmounts", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		bridgingAddressesManagerMock.On("GetPaymentPolicyScript", mock.Anything, mock.Anything).Return(script, true).Once()
		bridgingAddressesManagerMock.On("GetPaymentAddressFromIndex", mock.Anything, mock.Anything).Return(bridgingAddr, true).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigPolicyScript", mock.Anything, mock.Anything).Return(script, true).Once()

		dbMock.On("GetAllTxOutputs", bridgingAddr, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 4000000,
					},
				},
			}, error(nil))
		dbMock.On("GetAllTxOutputs", feeAddr, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 2000000,
					},
				},
			}, error(nil))

		result, err := cco.GenerateBatchTransaction(ctx, destinationChain, redistributionConfirmedTx, batchNonceID)
		require.NoError(t, err)
		require.NotNil(t, result.TxRaw)
		require.NotEqual(t, "", result.TxHash)
	})

	t.Run("Test SignBatchTransaction", func(t *testing.T) {
		txRaw, err := hex.DecodeString("84a5008282582000000000000000000000000000000000000000000000000000000000000000120082582000000000000000000000000000000000000000000000000000000000000000ff00018282581d6033c378cee41b2e15ac848f7f6f1d2f78155ab12d93b713de898d855f1903e882581d702b5398fcb481e94163a6b5cca889c54bcd9d340fb71c5eaa9f2c8d441a001e8098021a0002e76d031864075820c5e403ad2ee72ff4eb1ab7e988c1e1b4cb34df699cb9112d6bded8e8f3195f34a10182830301818200581ce67d6de92a4abb3712e887fe2cf0f07693028fad13a3e510dbe73394830301818200581c31a31e2f2cd4e1d66fc25f400aa02ab0fe6ca5a3d735c2974e842a89f5d90103a100a101a2616e016174656261746368")
		require.NoError(t, err)

		signatures, err := cco.SignBatchTransaction(&core.GeneratedBatchTxData{TxRaw: txRaw, IsPaymentSignNeeded: true})
		require.NoError(t, err)
		require.NotNil(t, signatures.Multisig)
		require.NotNil(t, signatures.Fee)
		require.Nil(t, signatures.MultsigStake)
	})
}

func TestGenerateBatchTransactionOnlyStaking(t *testing.T) {
	testDir, err := os.MkdirTemp("", "bat-chain-ops-tx")
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	secretsMngr, err := secretsHelper.CreateSecretsManager(&secrets.SecretsManagerConfig{
		Path: filepath.Join(testDir, "stp"),
		Type: secrets.Local,
	})
	require.NoError(t, err)

	_, err = cardano.GenerateWallet(secretsMngr, "prime", true, false)
	require.NoError(t, err)

	configRaw := json.RawMessage([]byte(`{
			"socketPath": "./socket",
			"testnetMagic": 42,
			"minUtxoAmount": 1000
			}`))
	dbMock := &indexer.DatabaseMock{}
	txProviderMock := &cardano.TxProviderTestMock{
		ReturnDefaultParameters: true,
	}

	bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
	bridgingAddressCoordinatorMock := &bac.BridgingAddressesCoordinatorMock{}

	cco, err := NewCardanoChainOperations(configRaw, dbMock, secretsMngr, "prime", bridgingAddressesManagerMock, bridgingAddressCoordinatorMock, hclog.NewNullLogger())
	require.NoError(t, err)

	cco.txProvider = txProviderMock
	cco.config.SlotRoundingThreshold = 100
	cco.config.MaxFeeUtxoCount = 4
	cco.config.MaxUtxoCount = 50

	testError := errors.New("test err")
	confirmedTransactions := make([]eth.ConfirmedTransaction, 1)
	confirmedTransactions[0] = eth.ConfirmedTransaction{
		Nonce:              1,
		BlockHeight:        big.NewInt(1),
		SourceChainId:      common.ChainIDIntPrime,
		DestinationChainId: common.ChainIDIntPrime,
		TransactionType:    uint8(common.StakeDelConfirmedTxType),
		StakePoolId:        "pool1f0drqjkgfhqcdeyvfuvgv9hsss59hpfj5rrrk9hlg7tm29tmkjr",
	}
	batchNonceID := uint64(1)
	destinationChain := common.ChainIDStrVector
	cliUtils := cardanowallet.NewCliUtils(cardanowallet.ResolveCardanoCliBinary(cardanowallet.MainNetNetwork))
	keys := []string{
		"846d5cb85238b2f433e3a35f1df61a4fbc2a69a705e5bbcb626ce9ae",
		"5a12073b94bfbdfcbb7cb985eed1f35f5eeafaf57912905f654f41bf",
		"adefefdd0dec7d8044285998b51d6cf39b4d8fb613c695045d021ac4",
	}
	script := cardanowallet.NewPolicyScript(keys, 2, cardanowallet.WithAfter(0))
	bridgingAddr, err := cliUtils.GetPolicyScriptBaseAddress(cardanowallet.TestNetProtocolMagic, script, script)
	require.NoError(t, err)

	keys = []string{
		"846d5cb85238b2f433e3a35f1df61a4fbc2a69a705e5bbcb626ce9ae",
		"adefefdd0dec7d8044285998b51d6cf39b4d8fb613c695045d021ac4",
		"5a12073b94bfbdfcbb7cb985eed1f35f5eeafaf57912905f654f41bf",
	}
	script = cardanowallet.NewPolicyScript(keys, 2, cardanowallet.WithAfter(0))
	feeAddr, err := cliUtils.GetPolicyScriptBaseAddress(cardanowallet.TestNetProtocolMagic, script, script)
	require.NoError(t, err)

	addressAndAmountRet := []common.AddressAndAmount{
		{
			AddressIndex:  0,
			Address:       bridgingAddr,
			TokensAmounts: map[string]uint64{"lovelace": 1_000_000},
		},
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	t.Run("GetStakePolicyScript returns false", func(t *testing.T) {
		bridgingAddressesManagerMock.On("GetStakePolicyScript", mock.Anything, mock.Anything).Return(nil, false).Once()

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get stake policy script for address")
	})

	t.Run("GetStakeAddressFromIndex returns false", func(t *testing.T) {
		bridgingAddressesManagerMock.On("GetStakePolicyScript", mock.Anything, mock.Anything).Return(script, true).Once()
		bridgingAddressesManagerMock.On("GetStakeAddressFromIndex", mock.Anything, mock.Anything).Return("", false).Once()

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get stake address from index")
	})

	t.Run("GetAddressesAndAmountsToPayFrom returns error", func(t *testing.T) {
		bridgingAddressesManagerMock.On("GetStakePolicyScript", mock.Anything, mock.Anything).Return(script, true).Once()
		bridgingAddressesManagerMock.On("GetStakeAddressFromIndex", mock.Anything, mock.Anything).Return("stake_test1uqehkck0lajq8gr28t9uxnuvgcqrc6070x3k9r8048z8y5gssrtvn", true).Once()
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]common.AddressAndAmount(nil), testError).Once()

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GetAllTxOutputs return error", func(t *testing.T) {
		bridgingAddressesManagerMock.On("GetStakePolicyScript", mock.Anything, mock.Anything).Return(script, true)
		bridgingAddressesManagerMock.On("GetStakeAddressFromIndex", mock.Anything, mock.Anything).Return("stake_test1uqehkck0lajq8gr28t9uxnuvgcqrc6070x3k9r8048z8y5gssrtvn", true)
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput(nil), testError).Once()

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GenerateBatchTransaction fee multisig does not have any utxo", func(t *testing.T) {
		bridgingAddressesManagerMock.On("GetStakePolicyScript", mock.Anything, mock.Anything).Return(script, true)
		bridgingAddressesManagerMock.On("GetStakeAddressFromIndex", mock.Anything, mock.Anything).Return("stake_test1uqehkck0lajq8gr28t9uxnuvgcqrc6070x3k9r8048z8y5gssrtvn", true)
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).
			Return([]*indexer.TxInputOutput{}, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", bridgingAddr, true).
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

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.ErrorContains(t, err, "fee multisig does not have any utxo")
	})

	t.Run("GetLatestBlockPoint return error", func(t *testing.T) {
		bridgingAddressesManagerMock.On("GetStakePolicyScript", mock.Anything, mock.Anything).Return(script, true)
		bridgingAddressesManagerMock.On("GetStakeAddressFromIndex", mock.Anything, mock.Anything).Return("stake_test1uqehkck0lajq8gr28t9uxnuvgcqrc6070x3k9r8048z8y5gssrtvn", true)
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr).Once()
		dbMock.On("GetAllTxOutputs", bridgingAddr, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 5000000,
					},
				},
			}, error(nil))
		dbMock.On("GetAllTxOutputs", feeAddr, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 2000000,
					},
				},
			}, error(nil))
		dbMock.On("GetLatestBlockPoint").Return((*indexer.BlockPoint)(nil), testError).Once()

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GetPaymentPolicyScript return false", func(t *testing.T) {
		bridgingAddressesManagerMock.On("GetStakePolicyScript", mock.Anything, mock.Anything).Return(script, true)
		bridgingAddressesManagerMock.On("GetStakeAddressFromIndex", mock.Anything, mock.Anything).Return("stake_test1uqehkck0lajq8gr28t9uxnuvgcqrc6070x3k9r8048z8y5gssrtvn", true)
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		bridgingAddressesManagerMock.On("GetPaymentPolicyScript", mock.Anything, mock.Anything).Return(nil, false).Once()

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get payment policy script for address")
	})

	t.Run("GetPaymentAddressFromIndex return false", func(t *testing.T) {
		bridgingAddressesManagerMock.On("GetStakePolicyScript", mock.Anything, mock.Anything).Return(script, true)
		bridgingAddressesManagerMock.On("GetStakeAddressFromIndex", mock.Anything, mock.Anything).Return("stake_test1uqehkck0lajq8gr28t9uxnuvgcqrc6070x3k9r8048z8y5gssrtvn", true)
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		bridgingAddressesManagerMock.On("GetPaymentPolicyScript", mock.Anything, mock.Anything).Return(script, true).Once()
		bridgingAddressesManagerMock.On("GetPaymentAddressFromIndex", mock.Anything, mock.Anything).Return("", false).Once()

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get payment address for address index")
	})

	t.Run("GetFeePolicyScript return false", func(t *testing.T) {
		bridgingAddressesManagerMock.On("GetStakePolicyScript", mock.Anything, mock.Anything).Return(script, true)
		bridgingAddressesManagerMock.On("GetStakeAddressFromIndex", mock.Anything, mock.Anything).Return("stake_test1uqehkck0lajq8gr28t9uxnuvgcqrc6070x3k9r8048z8y5gssrtvn", true)
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		bridgingAddressesManagerMock.On("GetPaymentPolicyScript", mock.Anything, mock.Anything).Return(script, true).Once()
		bridgingAddressesManagerMock.On("GetPaymentAddressFromIndex", mock.Anything, mock.Anything).Return(bridgingAddr, true).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigPolicyScript", mock.Anything, mock.Anything).Return(nil, false).Once()

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get fee policy script for chain")
	})

	t.Run("should pass", func(t *testing.T) {
		bridgingAddressesManagerMock.On("GetStakePolicyScript", mock.Anything, mock.Anything).Return(script, true)
		bridgingAddressesManagerMock.On("GetStakeAddressFromIndex", mock.Anything, mock.Anything).Return("stake_test1uqehkck0lajq8gr28t9uxnuvgcqrc6070x3k9r8048z8y5gssrtvn", true)
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		bridgingAddressesManagerMock.On("GetPaymentPolicyScript", mock.Anything, mock.Anything).Return(script, true).Once()
		bridgingAddressesManagerMock.On("GetPaymentAddressFromIndex", mock.Anything, mock.Anything).Return(bridgingAddr, true).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigPolicyScript", mock.Anything, mock.Anything).Return(script, true).Once()

		result, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.NoError(t, err)
		require.NotNil(t, result.TxRaw)
		require.NotEqual(t, "", result.TxHash)
	})

	t.Run("Test SignBatchTransaction", func(t *testing.T) {
		txRaw, err := hex.DecodeString("84a5008282582000000000000000000000000000000000000000000000000000000000000000120082582000000000000000000000000000000000000000000000000000000000000000ff00018282581d6033c378cee41b2e15ac848f7f6f1d2f78155ab12d93b713de898d855f1903e882581d702b5398fcb481e94163a6b5cca889c54bcd9d340fb71c5eaa9f2c8d441a001e8098021a0002e76d031864075820c5e403ad2ee72ff4eb1ab7e988c1e1b4cb34df699cb9112d6bded8e8f3195f34a10182830301818200581ce67d6de92a4abb3712e887fe2cf0f07693028fad13a3e510dbe73394830301818200581c31a31e2f2cd4e1d66fc25f400aa02ab0fe6ca5a3d735c2974e842a89f5d90103a100a101a2616e016174656261746368")
		require.NoError(t, err)

		signatures, err := cco.SignBatchTransaction(
			&core.GeneratedBatchTxData{TxRaw: txRaw, IsStakeSignNeeded: true})
		require.NoError(t, err)
		require.Nil(t, signatures.Multisig)
		require.NotNil(t, signatures.Fee)
		require.NotNil(t, signatures.MultsigStake)
	})
}

func TestGenerateBatchTransactionWithStaking(t *testing.T) {
	testDir, err := os.MkdirTemp("", "bat-chain-ops-tx")
	require.NoError(t, err)

	minUtxoAmount := new(big.Int).SetUint64(1_000)

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	secretsMngr, err := secretsHelper.CreateSecretsManager(&secrets.SecretsManagerConfig{
		Path: filepath.Join(testDir, "stp"),
		Type: secrets.Local,
	})
	require.NoError(t, err)

	_, err = cardano.GenerateWallet(secretsMngr, "prime", true, false)
	require.NoError(t, err)

	configRaw := json.RawMessage([]byte(`{
				"socketPath": "./socket",
				"testnetMagic": 42,
				"minUtxoAmount": 1000
				}`))
	dbMock := &indexer.DatabaseMock{}
	txProviderMock := &cardano.TxProviderTestMock{
		ReturnDefaultParameters: true,
	}

	bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
	bridgingAddressCoordinatorMock := &bac.BridgingAddressesCoordinatorMock{}

	cco, err := NewCardanoChainOperations(configRaw, dbMock, secretsMngr, "prime", bridgingAddressesManagerMock, bridgingAddressCoordinatorMock, hclog.NewNullLogger())
	require.NoError(t, err)

	cco.txProvider = txProviderMock
	cco.config.SlotRoundingThreshold = 100
	cco.config.MaxFeeUtxoCount = 4
	cco.config.MaxUtxoCount = 50

	testError := errors.New("test err")
	confirmedTransactions := make([]eth.ConfirmedTransaction, 2)
	confirmedTransactions[0] = eth.ConfirmedTransaction{
		Nonce:       1,
		BlockHeight: big.NewInt(1),
		Receivers: []eth.BridgeReceiver{{
			DestinationAddress: "addr_test1vqeux7xwusdju9dvsj8h7mca9aup2k439kfmwy773xxc2hcu7zy99",
			Amount:             minUtxoAmount,
		}},
		TransactionType: uint8(common.BridgingConfirmedTxType),
	}
	confirmedTransactions[1] = eth.ConfirmedTransaction{
		Nonce:              2,
		BlockHeight:        big.NewInt(2),
		SourceChainId:      common.ChainIDIntPrime,
		DestinationChainId: common.ChainIDIntPrime,
		TransactionType:    uint8(common.StakeDelConfirmedTxType),
		StakePoolId:        "pool1f0drqjkgfhqcdeyvfuvgv9hsss59hpfj5rrrk9hlg7tm29tmkjr",
	}
	batchNonceID := uint64(1)
	destinationChain := common.ChainIDStrVector
	cliUtils := cardanowallet.NewCliUtils(cardanowallet.ResolveCardanoCliBinary(cardanowallet.MainNetNetwork))
	keys := []string{
		"846d5cb85238b2f433e3a35f1df61a4fbc2a69a705e5bbcb626ce9ae",
		"5a12073b94bfbdfcbb7cb985eed1f35f5eeafaf57912905f654f41bf",
		"adefefdd0dec7d8044285998b51d6cf39b4d8fb613c695045d021ac4",
	}
	script := cardanowallet.NewPolicyScript(keys, 2, cardanowallet.WithAfter(0))
	bridgingAddr, err := cliUtils.GetPolicyScriptBaseAddress(cardanowallet.TestNetProtocolMagic, script, script)
	require.NoError(t, err)

	keys = []string{
		"846d5cb85238b2f433e3a35f1df61a4fbc2a69a705e5bbcb626ce9ae",
		"adefefdd0dec7d8044285998b51d6cf39b4d8fb613c695045d021ac4",
		"5a12073b94bfbdfcbb7cb985eed1f35f5eeafaf57912905f654f41bf",
	}
	script = cardanowallet.NewPolicyScript(keys, 2, cardanowallet.WithAfter(0))
	feeAddr, err := cliUtils.GetPolicyScriptBaseAddress(cardanowallet.TestNetProtocolMagic, script, script)
	require.NoError(t, err)

	addressAndAmountRet := []common.AddressAndAmount{
		{
			AddressIndex:  0,
			Address:       bridgingAddr,
			TokensAmounts: map[string]uint64{"lovelace": 1_000_000},
		},
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	t.Run("GetStakePolicyScript returns false", func(t *testing.T) {
		bridgingAddressesManagerMock.On("GetStakePolicyScript", mock.Anything, mock.Anything).Return(nil, false).Once()

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get stake policy script for address")
	})

	t.Run("GetStakeAddressFromIndex returns false", func(t *testing.T) {
		bridgingAddressesManagerMock.On("GetStakePolicyScript", mock.Anything, mock.Anything).Return(script, true).Once()
		bridgingAddressesManagerMock.On("GetStakeAddressFromIndex", mock.Anything, mock.Anything).Return("", false).Once()

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get stake address from index")
	})

	t.Run("GetAddressesAndAmountsToPayFrom returns error", func(t *testing.T) {
		bridgingAddressesManagerMock.On("GetStakePolicyScript", mock.Anything, mock.Anything).Return(script, true).Once()
		bridgingAddressesManagerMock.On("GetStakeAddressFromIndex", mock.Anything, mock.Anything).Return("stake_test1uqehkck0lajq8gr28t9uxnuvgcqrc6070x3k9r8048z8y5gssrtvn", true).Once()
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]common.AddressAndAmount(nil), testError).Once()

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GetAllTxOutputs return error", func(t *testing.T) {
		bridgingAddressesManagerMock.On("GetStakePolicyScript", mock.Anything, mock.Anything).Return(script, true)
		bridgingAddressesManagerMock.On("GetStakeAddressFromIndex", mock.Anything, mock.Anything).Return("stake_test1uqehkck0lajq8gr28t9uxnuvgcqrc6070x3k9r8048z8y5gssrtvn", true)
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput(nil), testError).Once()

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GenerateBatchTransaction fee multisig does not have any utxo", func(t *testing.T) {
		bridgingAddressesManagerMock.On("GetStakePolicyScript", mock.Anything, mock.Anything).Return(script, true)
		bridgingAddressesManagerMock.On("GetStakeAddressFromIndex", mock.Anything, mock.Anything).Return("stake_test1uqehkck0lajq8gr28t9uxnuvgcqrc6070x3k9r8048z8y5gssrtvn", true)
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).
			Return([]*indexer.TxInputOutput{}, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", bridgingAddr, true).
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

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.ErrorContains(t, err, "fee multisig does not have any utxo")
	})

	t.Run("GetLatestBlockPoint return error", func(t *testing.T) {
		bridgingAddressesManagerMock.On("GetStakePolicyScript", mock.Anything, mock.Anything).Return(script, true)
		bridgingAddressesManagerMock.On("GetStakeAddressFromIndex", mock.Anything, mock.Anything).Return("stake_test1uqehkck0lajq8gr28t9uxnuvgcqrc6070x3k9r8048z8y5gssrtvn", true)
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr).Once()
		dbMock.On("GetAllTxOutputs", bridgingAddr, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 5000000,
					},
				},
			}, error(nil))
		dbMock.On("GetAllTxOutputs", feeAddr, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 2000000,
					},
				},
			}, error(nil))
		dbMock.On("GetLatestBlockPoint").Return((*indexer.BlockPoint)(nil), testError).Once()

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GetPaymentPolicyScript return false", func(t *testing.T) {
		bridgingAddressesManagerMock.On("GetStakePolicyScript", mock.Anything, mock.Anything).Return(script, true)
		bridgingAddressesManagerMock.On("GetStakeAddressFromIndex", mock.Anything, mock.Anything).Return("stake_test1uqehkck0lajq8gr28t9uxnuvgcqrc6070x3k9r8048z8y5gssrtvn", true)
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		bridgingAddressesManagerMock.On("GetPaymentPolicyScript", mock.Anything, mock.Anything).Return(nil, false).Once()

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get payment policy script for address")
	})

	t.Run("GetFeePolicyScript return false", func(t *testing.T) {
		bridgingAddressesManagerMock.On("GetStakePolicyScript", mock.Anything, mock.Anything).Return(script, true)
		bridgingAddressesManagerMock.On("GetStakeAddressFromIndex", mock.Anything, mock.Anything).Return("stake_test1uqehkck0lajq8gr28t9uxnuvgcqrc6070x3k9r8048z8y5gssrtvn", true)
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		bridgingAddressesManagerMock.On("GetPaymentPolicyScript", mock.Anything, mock.Anything).Return(script, true).Once()
		bridgingAddressesManagerMock.On("GetPaymentAddressFromIndex", mock.Anything, mock.Anything).Return("", false).Once()

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get payment address for address index")
	})

	t.Run("GetFeePolicyScript return false", func(t *testing.T) {
		bridgingAddressesManagerMock.On("GetStakePolicyScript", mock.Anything, mock.Anything).Return(script, true)
		bridgingAddressesManagerMock.On("GetStakeAddressFromIndex", mock.Anything, mock.Anything).Return("stake_test1uqehkck0lajq8gr28t9uxnuvgcqrc6070x3k9r8048z8y5gssrtvn", true)
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		bridgingAddressesManagerMock.On("GetPaymentPolicyScript", mock.Anything, mock.Anything).Return(script, true).Once()
		bridgingAddressesManagerMock.On("GetPaymentAddressFromIndex", mock.Anything, mock.Anything).Return(bridgingAddr, true).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigPolicyScript", mock.Anything, mock.Anything).Return(nil, false).Once()

		_, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get fee policy script for chain")
	})

	t.Run("should pass", func(t *testing.T) {
		bridgingAddressesManagerMock.On("GetStakePolicyScript", mock.Anything, mock.Anything).Return(script, true)
		bridgingAddressesManagerMock.On("GetStakeAddressFromIndex", mock.Anything, mock.Anything).Return("stake_test1uqehkck0lajq8gr28t9uxnuvgcqrc6070x3k9r8048z8y5gssrtvn", true)
		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		bridgingAddressesManagerMock.On("GetPaymentPolicyScript", mock.Anything, mock.Anything).Return(script, true).Once()
		bridgingAddressesManagerMock.On("GetPaymentAddressFromIndex", mock.Anything, mock.Anything).Return(bridgingAddr, true).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigPolicyScript", mock.Anything, mock.Anything).Return(script, true).Once()

		result, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.NoError(t, err)
		require.NotNil(t, result.TxRaw)
		require.NotEqual(t, "", result.TxHash)
	})

	t.Run("Test SignBatchTransaction", func(t *testing.T) {
		txRaw, err := hex.DecodeString("84a5008282582000000000000000000000000000000000000000000000000000000000000000120082582000000000000000000000000000000000000000000000000000000000000000ff00018282581d6033c378cee41b2e15ac848f7f6f1d2f78155ab12d93b713de898d855f1903e882581d702b5398fcb481e94163a6b5cca889c54bcd9d340fb71c5eaa9f2c8d441a001e8098021a0002e76d031864075820c5e403ad2ee72ff4eb1ab7e988c1e1b4cb34df699cb9112d6bded8e8f3195f34a10182830301818200581ce67d6de92a4abb3712e887fe2cf0f07693028fad13a3e510dbe73394830301818200581c31a31e2f2cd4e1d66fc25f400aa02ab0fe6ca5a3d735c2974e842a89f5d90103a100a101a2616e016174656261746368")
		require.NoError(t, err)

		signatures, err := cco.SignBatchTransaction(
			&core.GeneratedBatchTxData{TxRaw: txRaw, IsPaymentSignNeeded: true, IsStakeSignNeeded: true})
		require.NoError(t, err)
		require.NotNil(t, signatures.Multisig)
		require.NotNil(t, signatures.Fee)
		require.NotNil(t, signatures.MultsigStake)
		require.NotEqual(t, signatures.Multisig, signatures.MultsigStake)
	})
}

func Test_createBatchInitialData(t *testing.T) {
	testDir, err := os.MkdirTemp("", "bat-chain-ops-tx")
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	secretsMngr, err := secretsHelper.CreateSecretsManager(&secrets.SecretsManagerConfig{
		Path: filepath.Join(testDir, "stp"),
		Type: secrets.Local,
	})
	require.NoError(t, err)

	_, err = cardano.GenerateWallet(secretsMngr, "prime", true, false)
	require.NoError(t, err)

	configRaw := json.RawMessage([]byte(`{
			"socketPath": "./socket",
			"testnetMagic": 42,
			"minUtxoAmount": 1000
			}`))
	dbMock := &indexer.DatabaseMock{}
	txProviderMock := &cardano.TxProviderTestMock{
		ReturnDefaultParameters: true,
	}

	bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
	bridgingAddressCoordinatorMock := &bac.BridgingAddressesCoordinatorMock{}

	cco, err := NewCardanoChainOperations(configRaw, dbMock, secretsMngr, "prime", bridgingAddressesManagerMock, bridgingAddressCoordinatorMock, hclog.NewNullLogger())
	require.NoError(t, err)

	cco.txProvider = txProviderMock

	batchID := uint64(1)
	destinationChain := common.ChainIDStrVector

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	t.Run("GetProtocolParameters error", func(t *testing.T) {
		desiredErr := errors.New("hello")

		txProviderMock.On("GetProtocolParameters", ctx).Return(nil, desiredErr)

		txProviderMock.ReturnDefaultParameters = false
		defer func() {
			txProviderMock.ReturnDefaultParameters = true
		}()

		_, err := cco.createBatchInitialData(ctx, destinationChain, batchID)
		require.ErrorIs(t, err, desiredErr)
	})

	t.Run("good", func(t *testing.T) {
		data, err := cco.createBatchInitialData(ctx, destinationChain, batchID)
		require.NoError(t, err)
		assert.Equal(t, batchID, data.BatchNonceID)
		assert.Greater(t, len(data.Metadata), 5)
	})
}

func TestGenerateConsolidationTransaction(t *testing.T) {
	testDir, err := os.MkdirTemp("", "bat-chain-ops-tx")
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	secretsMngr, err := secretsHelper.CreateSecretsManager(&secrets.SecretsManagerConfig{
		Path: filepath.Join(testDir, "stp"),
		Type: secrets.Local,
	})
	require.NoError(t, err)

	_, err = cardano.GenerateWallet(secretsMngr, "prime", false, false)
	require.NoError(t, err)

	configRaw := json.RawMessage([]byte(`{
			"socketPath": "./socket",
			"testnetMagic": 42,
			"minUtxoAmount": 1000
			}`))
	dbMock := &indexer.DatabaseMock{}
	txProviderMock := &cardano.TxProviderTestMock{
		ReturnDefaultParameters: true,
	}

	bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
	bridgingAddressCoordinatorMock := &bac.BridgingAddressesCoordinatorMock{}

	cco, err := NewCardanoChainOperations(configRaw, dbMock, secretsMngr, "prime", bridgingAddressesManagerMock, bridgingAddressCoordinatorMock, hclog.NewNullLogger())
	require.NoError(t, err)

	cco.txProvider = txProviderMock
	cco.config.SlotRoundingThreshold = 100
	cco.config.MaxFeeUtxoCount = 4
	cco.config.MaxUtxoCount = 50

	testError := errors.New("test err")
	batchID := uint64(1)
	destinationChain := common.ChainIDStrVector
	cliUtils := cardanowallet.NewCliUtils(cardanowallet.ResolveCardanoCliBinary(cardanowallet.MainNetNetwork))
	keys := []string{
		"846d5cb85238b2f433e3a35f1df61a4fbc2a69a705e5bbcb626ce9ae",
		"5a12073b94bfbdfcbb7cb985eed1f35f5eeafaf57912905f654f41bf",
		"adefefdd0dec7d8044285998b51d6cf39b4d8fb613c695045d021ac4",
	}
	script := cardanowallet.NewPolicyScript(keys, 2, cardanowallet.WithAfter(0))
	bridgingAddr, err := cliUtils.GetPolicyScriptBaseAddress(cardanowallet.TestNetProtocolMagic, script, script)
	require.NoError(t, err)

	keys = []string{
		"846d5cb85238b2f433e3a35f1df61a4fbc2a69a705e5bbcb626ce9ae",
		"adefefdd0dec7d8044285998b51d6cf39b4d8fb613c695045d021ac4",
		"5a12073b94bfbdfcbb7cb985eed1f35f5eeafaf57912905f654f41bf",
	}
	script = cardanowallet.NewPolicyScript(keys, 2, cardanowallet.WithAfter(1))
	feeAddr, err := cliUtils.GetPolicyScriptBaseAddress(cardanowallet.TestNetProtocolMagic, script, script)
	require.NoError(t, err)

	addressAndAmountRet := []common.AddressAndAmount{
		{
			AddressIndex:  0,
			Address:       bridgingAddr,
			TokensAmounts: map[string]uint64{"lovelace": 1_000_000},
			IncludeChange: 1_000_000,
		},
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr)

	t.Run("GetLatestBlockPoint return error", func(t *testing.T) {
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
			}, error(nil)).Twice()
		dbMock.On("GetLatestBlockPoint").Return((*indexer.BlockPoint)(nil), testError).Once()

		data, err := cco.createBatchInitialData(ctx, destinationChain, batchID)
		require.NoError(t, err)

		_, err = cco.generateConsolidationTransaction(data, addressAndAmountRet)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GetAllTxOutputs return error", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput(nil), testError).Once()

		data, err := cco.createBatchInitialData(ctx, destinationChain, batchID)
		require.NoError(t, err)

		_, err = cco.generateConsolidationTransaction(data, addressAndAmountRet)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GetFeeMultisigPolicyScript returns false", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
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
			Return([]*indexer.TxInputOutput{}, error(nil)).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigPolicyScript", mock.Anything).Return(nil, false).Once()

		data, err := cco.createBatchInitialData(ctx, destinationChain, batchID)
		require.NoError(t, err)

		_, err = cco.generateConsolidationTransaction(data, addressAndAmountRet)
		require.ErrorContains(t, err, "failed to get fee policy script")
	})

	t.Run("GenerateConsolidationTransaction fee multisig does not have any utxo", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{}, error(nil)).Once()
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

		bridgingAddressesManagerMock.On("GetFeeMultisigPolicyScript", mock.Anything).Return(script, true).Once()

		data, err := cco.createBatchInitialData(ctx, destinationChain, batchID)
		require.NoError(t, err)

		_, err = cco.generateConsolidationTransaction(data, addressAndAmountRet)
		require.ErrorContains(t, err, "fee multisig does not have any utxo")
	})

	t.Run("GenerateConsolidationTransaction should pass", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0xFF"),
					},
					Output: indexer.TxOutput{
						Amount: 300_000,
					},
				},
			}, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", bridgingAddr, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 2_000_000,
					},
				},
			}, error(nil)).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigPolicyScript", mock.Anything).Return(script, true).Once()
		bridgingAddressesManagerMock.On("GetPaymentPolicyScript", mock.Anything, mock.Anything).Return(script, true).Once()

		data, err := cco.createBatchInitialData(ctx, destinationChain, batchID)
		require.NoError(t, err)

		result, err := cco.generateConsolidationTransaction(data, addressAndAmountRet)

		require.NoError(t, err)
		require.NotNil(t, result.TxRaw)
		require.NotEqual(t, "", result.TxHash)
	})

	t.Run("getUTXOsForConsolidation should pass when there is more utxos than maxUtxo", func(t *testing.T) {
		dbMock.ExpectedCalls = nil
		dbMock.Calls = nil
		multisigUtxoOutputs, _ := generateSmallUtxoOutputs(10, 50)
		feePayerUtxoOutputs, _ := generateSmallUtxoOutputs(10, 10)

		dbMock.On("GetAllTxOutputs", feeAddr, true).
			Return(feePayerUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", bridgingAddr, true).
			Return(multisigUtxoOutputs, error(nil)).Once()

		multisigUtxos, feeUtxos, err := cco.getUTXOsForConsolidation(addressAndAmountRet, feeAddr)
		require.NoError(t, err)
		require.Equal(t, 4, len(feeUtxos))
		require.Equal(t, 46, len(multisigUtxos[0]))
	})

	t.Run("getUTXOsForConsolidation should pass when there is les utxos than maxUtxo", func(t *testing.T) {
		dbMock.ExpectedCalls = nil
		dbMock.Calls = nil
		multisigUtxoOutputs, _ := generateSmallUtxoOutputs(10, 30)
		feePayerUtxoOutputs, _ := generateSmallUtxoOutputs(10, 3)

		dbMock.On("GetAllTxOutputs", bridgingAddr, true).
			Return(multisigUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).
			Return(feePayerUtxoOutputs, error(nil))

		multisigUtxos, feeUtxos, err := cco.getUTXOsForConsolidation(addressAndAmountRet, feeAddr)
		require.NoError(t, err)
		require.Equal(t, 30, len(multisigUtxos[0]))
		require.Equal(t, 3, len(feeUtxos))
	})

	t.Run("GenerateBatchTransaction execute consolidation and pass", func(t *testing.T) {
		dbMock.ExpectedCalls = nil
		dbMock.Calls = nil
		batchNonceID := uint64(1)

		multisigUtxoOutputs, _ := generateSmallUtxoOutputs(20_010, 1000)
		feePayerUtxoOutputs, _ := generateSmallUtxoOutputs(1_000_000, 10)

		confirmedTransactions := make([]eth.ConfirmedTransaction, 1)
		confirmedTransactions[0] = eth.ConfirmedTransaction{
			Nonce:       1,
			BlockHeight: big.NewInt(1),
			Receivers: []eth.BridgeReceiver{{
				DestinationAddress: "addr_test1vqeux7xwusdju9dvsj8h7mca9aup2k439kfmwy773xxc2hcu7zy99",
				Amount:             new(big.Int).SetUint64(3_000_000),
			}},
			TransactionType: uint8(common.BridgingConfirmedTxType),
		}

		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).
			Return(feePayerUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return(multisigUtxoOutputs, error(nil)).Once()

		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 55}, nil).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).
			Return(feePayerUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return(multisigUtxoOutputs, error(nil)).Once()

		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil)
		bridgingAddressesManagerMock.On("GetPaymentPolicyScript", mock.Anything, mock.Anything).Return(script, true)
		bridgingAddressesManagerMock.On("GetPaymentAddressFromIndex", mock.Anything, mock.Anything).Return(bridgingAddr, true)
		bridgingAddressesManagerMock.On("GetFeeMultisigPolicyScript", mock.Anything, mock.Anything).Return(script, true)

		result, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.NoError(t, err)
		require.True(t, result.IsConsolidation())
		require.NotNil(t, result.TxRaw)
		require.NotEqual(t, "", result.TxHash)
	})
}

func TestSkylineConsolidation(t *testing.T) {
	testDir, err := os.MkdirTemp("", "bat-chain-ops-tx")
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	secretsMngr, err := secretsHelper.CreateSecretsManager(&secrets.SecretsManagerConfig{
		Path: filepath.Join(testDir, "stp"),
		Type: secrets.Local,
	})
	require.NoError(t, err)

	_, err = cardano.GenerateWallet(secretsMngr, "prime", false, false)
	require.NoError(t, err)

	token1, _ := cardanowallet.NewTokenWithFullName("29f8873beb52e126f207a2dfd50f7cff556806b5b4cba9834a7b26a8.4b6173685f546f6b656e", true)
	token2, _ := cardanowallet.NewTokenWithFullName("29f8873beb52e126f207a2dfd50f7cff556806b5b4cba9834a7b26a8.526f75746533", true)

	configRaw := json.RawMessage([]byte(`{
			"socketPath": "./socket",
			"testnetMagic": 42,
			"minUtxoAmount": 1000
			}`))
	dbMock := &indexer.DatabaseMock{}
	txProviderMock := &cardano.TxProviderTestMock{
		ReturnDefaultParameters: true,
	}

	bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
	bridgingAddressCoordinatorMock := &bac.BridgingAddressesCoordinatorMock{}

	cco, err := NewCardanoChainOperations(
		configRaw, dbMock, secretsMngr, common.ChainIDStrPrime, bridgingAddressesManagerMock, bridgingAddressCoordinatorMock, hclog.NewNullLogger())
	require.NoError(t, err)

	destinationChain := common.ChainIDStrCardano
	cco.txProvider = txProviderMock
	cco.config.SlotRoundingThreshold = 100
	cco.config.MaxFeeUtxoCount = 4
	cco.config.MaxUtxoCount = 40
	cco.config.NativeTokens = []sendtx.TokenExchangeConfig{
		{
			DstChainID: destinationChain,
			TokenName:  token1.String(),
		},
	}
	cliUtils := cardanowallet.NewCliUtils(cardanowallet.ResolveCardanoCliBinary(cardanowallet.MainNetNetwork))
	keys := []string{
		"846d5cb85238b2f433e3a35f1df61a4fbc2a69a705e5bbcb626ce9ae",
		"5a12073b94bfbdfcbb7cb985eed1f35f5eeafaf57912905f654f41bf",
		"adefefdd0dec7d8044285998b51d6cf39b4d8fb613c695045d021ac4",
	}
	script := cardanowallet.NewPolicyScript(keys, 2, cardanowallet.WithAfter(0))
	bridgingAddr, err := cliUtils.GetPolicyScriptBaseAddress(cardanowallet.TestNetProtocolMagic, script, script)
	require.NoError(t, err)

	keys = []string{
		"846d5cb85238b2f433e3a35f1df61a4fbc2a69a705e5bbcb626ce9ae",
		"adefefdd0dec7d8044285998b51d6cf39b4d8fb613c695045d021ac4",
		"5a12073b94bfbdfcbb7cb985eed1f35f5eeafaf57912905f654f41bf",
	}
	script = cardanowallet.NewPolicyScript(keys, 2, cardanowallet.WithAfter(1))
	feeAddr, err := cliUtils.GetPolicyScriptBaseAddress(cardanowallet.TestNetProtocolMagic, script, script)
	require.NoError(t, err)

	addressAndAmountRet := []common.AddressAndAmount{
		{
			AddressIndex:  0,
			Address:       bridgingAddr,
			TokensAmounts: map[string]uint64{"lovelace": 1_000_000, token1.String(): 10},
		},
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	t.Run("getUTXOsForConsolidation should pass when there is more utxos than maxUtxo", func(t *testing.T) {
		multisigUtxoOutputs, _ := generateSmallUtxoOutputs(10, 100, token1, token2)
		feePayerUtxoOutputs, _ := generateSmallUtxoOutputs(10, 10)

		dbMock.On("GetAllTxOutputs", bridgingAddr, true).
			Return(multisigUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).
			Return(feePayerUtxoOutputs, error(nil)).Once()

		multisigUtxos, feeUtxos, err := cco.getUTXOsForConsolidation(addressAndAmountRet, feeAddr)
		require.NoError(t, err)
		require.Equal(t, 37, len(multisigUtxos[0]))
		require.Equal(t, 3, len(feeUtxos))
	})

	t.Run("getUTXOsForConsolidation should pass when there is les utxos than maxUtxo", func(t *testing.T) {
		dbMock.ExpectedCalls = nil
		dbMock.Calls = nil
		multisigUtxoOutputs, _ := generateSmallUtxoOutputs(10, 30, token1, token2)
		feePayerUtxoOutputs, _ := generateSmallUtxoOutputs(10, 3)

		dbMock.On("GetAllTxOutputs", bridgingAddr, true).
			Return(multisigUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).
			Return(feePayerUtxoOutputs, error(nil)).Once()

		multisigUtxos, feeUtxos, err := cco.getUTXOsForConsolidation(addressAndAmountRet, feeAddr)
		require.NoError(t, err)
		require.Equal(t, 15, len(multisigUtxos[0]))
		require.Equal(t, 3, len(feeUtxos))
	})

	t.Run("GenerateBatchTransaction execute consolidation and pass", func(t *testing.T) {
		dbMock.ExpectedCalls = nil
		dbMock.Calls = nil

		batchNonceID := uint64(1)

		multisigUtxoOutputs, _ := generateSmallUtxoOutputs(20_000, 100, token1, token2)
		feePayerUtxoOutputs, _ := generateSmallUtxoOutputs(1_000_000, 10)

		confirmedTransactions := make([]eth.ConfirmedTransaction, 1)
		confirmedTransactions[0] = eth.ConfirmedTransaction{
			Nonce:         1,
			BlockHeight:   big.NewInt(1),
			SourceChainId: common.ChainIDIntCardano,
			Receivers: []eth.BridgeReceiver{{
				DestinationAddress: "addr_test1vqeux7xwusdju9dvsj8h7mca9aup2k439kfmwy773xxc2hcu7zy99",
				Amount:             big.NewInt(2_000_000),
				AmountWrapped:      big.NewInt(1_500_000),
			}},
			TransactionType: uint8(common.BridgingConfirmedTxType),
		}

		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		dbMock.On("GetAllTxOutputs", bridgingAddr, true).
			Return(multisigUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).
			Return(feePayerUtxoOutputs, error(nil)).Once()

		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 55}, nil).Once()
		dbMock.On("GetAllTxOutputs", bridgingAddr, true).
			Return(multisigUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).
			Return(feePayerUtxoOutputs, error(nil)).Once()

		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil)
		bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr)
		bridgingAddressesManagerMock.On("GetFeeMultisigPolicyScript", mock.Anything, mock.Anything).Return(script, true)
		bridgingAddressesManagerMock.On("GetPaymentPolicyScript", mock.Anything, mock.Anything).Return(script, true)
		bridgingAddressesManagerMock.On("GetPaymentAddressFromIndex", mock.Anything, mock.Anything).Return(bridgingAddr, true)

		result, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.NoError(t, err)
		require.True(t, result.IsConsolidation())
		require.NotNil(t, result.TxRaw)
		require.NotEqual(t, "", result.TxHash)
	})
}

func TestGenerateConsolidationTransactionWithMultipleAddresses(t *testing.T) {
	testDir, err := os.MkdirTemp("", "bat-chain-ops-tx")
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	secretsMngr, err := secretsHelper.CreateSecretsManager(&secrets.SecretsManagerConfig{
		Path: filepath.Join(testDir, "stp"),
		Type: secrets.Local,
	})
	require.NoError(t, err)

	_, err = cardano.GenerateWallet(secretsMngr, "prime", false, false)
	require.NoError(t, err)

	configRaw := json.RawMessage([]byte(`{
			"socketPath": "./socket",
			"testnetMagic": 42,
			"minUtxoAmount": 1000
			}`))
	dbMock := &indexer.DatabaseMock{}
	txProviderMock := &cardano.TxProviderTestMock{
		ReturnDefaultParameters: true,
	}

	bridgingAddressesManagerMock := &bam.BridgingAddressesManagerMock{}
	bridgingAddressCoordinatorMock := &bac.BridgingAddressesCoordinatorMock{}

	cco, err := NewCardanoChainOperations(configRaw, dbMock, secretsMngr, "prime", bridgingAddressesManagerMock, bridgingAddressCoordinatorMock, hclog.NewNullLogger())
	require.NoError(t, err)

	cco.txProvider = txProviderMock
	cco.config.SlotRoundingThreshold = 100
	cco.config.MaxFeeUtxoCount = 4
	cco.config.MaxUtxoCount = 50

	testError := errors.New("test err")
	batchID := uint64(1)
	destinationChain := common.ChainIDStrVector
	cliUtils := cardanowallet.NewCliUtils(cardanowallet.ResolveCardanoCliBinary(cardanowallet.MainNetNetwork))
	keys := []string{
		"846d5cb85238b2f433e3a35f1df61a4fbc2a69a705e5bbcb626ce9ae",
		"5a12073b94bfbdfcbb7cb985eed1f35f5eeafaf57912905f654f41bf",
		"adefefdd0dec7d8044285998b51d6cf39b4d8fb613c695045d021ac4",
	}
	script := cardanowallet.NewPolicyScript(keys, 2, cardanowallet.WithAfter(0))
	bridgingAddr1, err := cliUtils.GetPolicyScriptBaseAddress(cardanowallet.TestNetProtocolMagic, script, script)
	require.NoError(t, err)

	script = cardanowallet.NewPolicyScript(keys, 2, cardanowallet.WithAfter(1))
	bridgingAddr2, err := cliUtils.GetPolicyScriptBaseAddress(cardanowallet.TestNetProtocolMagic, script, script)
	require.NoError(t, err)

	script = cardanowallet.NewPolicyScript(keys, 2, cardanowallet.WithAfter(2))
	feeAddr, err := cliUtils.GetPolicyScriptBaseAddress(cardanowallet.TestNetProtocolMagic, script, script)
	require.NoError(t, err)

	addressAndAmountRet := []common.AddressAndAmount{
		{
			AddressIndex:  0,
			Address:       bridgingAddr1,
			TokensAmounts: map[string]uint64{"lovelace": 1_000_000},
		},
		{
			AddressIndex:  1,
			Address:       bridgingAddr2,
			TokensAmounts: map[string]uint64{"lovelace": 1_000_000},
		},
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	bridgingAddressesManagerMock.On("GetFeeMultisigAddress", mock.Anything).Return(feeAddr)

	t.Run("GetLatestBlockPoint return error", func(t *testing.T) {
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
			}, error(nil)).Times(3)
		dbMock.On("GetLatestBlockPoint").Return((*indexer.BlockPoint)(nil), testError).Once()

		data, err := cco.createBatchInitialData(ctx, destinationChain, batchID)
		require.NoError(t, err)

		_, err = cco.generateConsolidationTransaction(data, addressAndAmountRet)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GetAllTxOutputs return error", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput(nil), testError).Once()

		data, err := cco.createBatchInitialData(ctx, destinationChain, batchID)
		require.NoError(t, err)

		_, err = cco.generateConsolidationTransaction(data, addressAndAmountRet)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GetFeeMultisigPolicyScript returns false", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
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
			}, error(nil)).Twice()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{}, error(nil)).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigPolicyScript", mock.Anything).Return(nil, false).Once()

		data, err := cco.createBatchInitialData(ctx, destinationChain, batchID)
		require.NoError(t, err)

		_, err = cco.generateConsolidationTransaction(data, addressAndAmountRet)
		require.ErrorContains(t, err, "failed to get fee policy script")
	})

	t.Run("GenerateConsolidationTransaction fee multisig does not have any utxo", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		dbMock.On("GetAllTxOutputs", mock.Anything, true).
			Return([]*indexer.TxInputOutput{}, error(nil)).Once()
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
			}, error(nil)).Twice()

		bridgingAddressesManagerMock.On("GetFeeMultisigPolicyScript", mock.Anything).Return(script, true).Once()

		data, err := cco.createBatchInitialData(ctx, destinationChain, batchID)
		require.NoError(t, err)

		_, err = cco.generateConsolidationTransaction(data, addressAndAmountRet)
		require.ErrorContains(t, err, "fee multisig does not have any utxo")
	})

	t.Run("GenerateConsolidationTransaction should pass", func(t *testing.T) {
		dbMock.ExpectedCalls = nil
		dbMock.Calls = nil
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0xFF"),
					},
					Output: indexer.TxOutput{
						Amount: 300_000,
					},
				},
			}, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", bridgingAddr1, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0012"),
					},
					Output: indexer.TxOutput{
						Amount: 3_000_000,
					},
				},
			}, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", bridgingAddr2, true).
			Return([]*indexer.TxInputOutput{
				{
					Input: indexer.TxInput{
						Hash: indexer.NewHashFromHexString("0x0013"),
					},
					Output: indexer.TxOutput{
						Amount: 3_000_000,
					},
				},
			}, error(nil)).Once()
		bridgingAddressesManagerMock.On("GetFeeMultisigPolicyScript", mock.Anything).Return(script, true).Once()
		bridgingAddressesManagerMock.On("GetPaymentPolicyScript", mock.Anything, mock.Anything).Return(script, true).Twice()

		data, err := cco.createBatchInitialData(ctx, destinationChain, batchID)
		require.NoError(t, err)

		result, err := cco.generateConsolidationTransaction(data, addressAndAmountRet)

		require.NoError(t, err)
		require.NotNil(t, result.TxRaw)
		require.NotEqual(t, "", result.TxHash)
	})

	t.Run("getUTXOsForConsolidation should pass when there is more utxos than maxUtxo", func(t *testing.T) {
		dbMock.ExpectedCalls = nil
		dbMock.Calls = nil
		multisigUtxoOutputs, _ := generateSmallUtxoOutputs(10, 50)
		feePayerUtxoOutputs, _ := generateSmallUtxoOutputs(10, 10)

		dbMock.On("GetAllTxOutputs", feeAddr, true).
			Return(feePayerUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", bridgingAddr1, true).
			Return(multisigUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", bridgingAddr2, true).
			Return(multisigUtxoOutputs, error(nil)).Once()

		multisigUtxos, feeUtxos, err := cco.getUTXOsForConsolidation(addressAndAmountRet, feeAddr)
		require.NoError(t, err)
		require.Equal(t, 2, len(feeUtxos))
		require.Equal(t, 24, len(multisigUtxos[0]))
		require.Equal(t, 24, len(multisigUtxos[1]))
	})

	t.Run("getUTXOsForConsolidation should pass when there is les utxos than maxUtxo", func(t *testing.T) {
		dbMock.ExpectedCalls = nil
		dbMock.Calls = nil
		multisigUtxoOutputs, _ := generateSmallUtxoOutputs(10, 30)
		feePayerUtxoOutputs, _ := generateSmallUtxoOutputs(10, 3)

		dbMock.On("GetAllTxOutputs", bridgingAddr1, true).
			Return(multisigUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", bridgingAddr2, true).
			Return(multisigUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).
			Return(feePayerUtxoOutputs, error(nil))

		multisigUtxos, feeUtxos, err := cco.getUTXOsForConsolidation(addressAndAmountRet, feeAddr)
		require.NoError(t, err)
		require.Equal(t, 24, len(multisigUtxos[0]))
		require.Equal(t, 24, len(multisigUtxos[1]))
		require.Equal(t, 2, len(feeUtxos))
	})

	t.Run("GenerateBatchTransaction execute consolidation and pass", func(t *testing.T) {
		dbMock.ExpectedCalls = nil
		dbMock.Calls = nil
		batchNonceID := uint64(1)

		multisigUtxoOutputs, _ := generateSmallUtxoOutputs(20_010, 1000)
		feePayerUtxoOutputs, _ := generateSmallUtxoOutputs(1000000, 10)

		confirmedTransactions := make([]eth.ConfirmedTransaction, 1)
		confirmedTransactions[0] = eth.ConfirmedTransaction{
			Nonce:       1,
			BlockHeight: big.NewInt(1),
			Receivers: []eth.BridgeReceiver{{
				DestinationAddress: "addr_test1vqeux7xwusdju9dvsj8h7mca9aup2k439kfmwy773xxc2hcu7zy99",
				Amount:             new(big.Int).SetUint64(3_000_000),
			}},
			TransactionType: uint8(common.BridgingConfirmedTxType),
		}

		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 50}, nil).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).
			Return(feePayerUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", bridgingAddr1, true).
			Return(multisigUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", bridgingAddr2, true).
			Return(multisigUtxoOutputs, error(nil)).Once()

		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: 55}, nil).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).
			Return(feePayerUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", bridgingAddr1, true).
			Return(multisigUtxoOutputs, error(nil)).Once()
		dbMock.On("GetAllTxOutputs", bridgingAddr2, true).
			Return(multisigUtxoOutputs, error(nil)).Once()

		bridgingAddressCoordinatorMock.On("GetAddressesAndAmountsToPayFrom", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(addressAndAmountRet, nil)
		bridgingAddressesManagerMock.On("GetPaymentPolicyScript", mock.Anything, mock.Anything).Return(script, true)
		bridgingAddressesManagerMock.On("GetPaymentAddressFromIndex", uint8(2), uint8(0)).Return(bridgingAddr1, true)
		bridgingAddressesManagerMock.On("GetPaymentAddressFromIndex", uint8(2), uint8(1)).Return(bridgingAddr2, true)
		bridgingAddressesManagerMock.On("GetFeeMultisigPolicyScript", mock.Anything, mock.Anything).Return(script, true)

		result, err := cco.GenerateBatchTransaction(ctx, destinationChain, confirmedTransactions, batchNonceID)
		require.NoError(t, err)
		require.True(t, result.IsConsolidation())
		require.NotNil(t, result.TxRaw)
		require.NotEqual(t, "", result.TxHash)
	})
}

func Test_getUTXOsForNormalBatch(t *testing.T) {
	token, _ := cardanowallet.NewTokenWithFullName("29f8873beb52e126f207a2dfd50f7cff556806b5b4cba9834a7b26a8.4b6173685f546f6b656e", true)
	dbMock := &indexer.DatabaseMock{}
	txProviderMock := &cardano.TxProviderTestMock{
		ReturnDefaultParameters: true,
	}
	multisigAddr := "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k"
	feeAddr := "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8"
	cco := &CardanoChainOperations{
		config: &cardano.CardanoChainConfig{
			MaxFeeUtxoCount: 1,
			MaxUtxoCount:    3,
			NativeTokens: []sendtx.TokenExchangeConfig{
				{
					DstChainID: common.ChainIDStrPrime,
					TokenName:  token.String(),
				},
			},
		},
		cardanoCliBinary: cardanowallet.ResolveCardanoCliBinary(cardanowallet.MainNetNetwork),
		db:               dbMock,
		txProvider:       txProviderMock,
		logger:           hclog.NewNullLogger(),
	}

	t.Run("empty fee", func(t *testing.T) {
		dbMock.On("GetAllTxOutputs", multisigAddr, true).Return([]*indexer.TxInputOutput{}, nil).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).Return([]*indexer.TxInputOutput{}, nil).Once()

		_, err := cco.getUTXOsForNormalBatch([]common.AddressAndAmount{}, multisigAddr, false)
		require.ErrorContains(t, err, "fee")
	})

	t.Run("pass", func(t *testing.T) {
		dbMock.ExpectedCalls = nil
		dbMock.Calls = nil
		multisigUtxos := []*indexer.TxInputOutput{
			{
				Input: indexer.TxInput{
					Hash: indexer.Hash{1, 2},
				},
				Output: indexer.TxOutput{
					Address: multisigAddr,
					Amount:  50,
					Tokens: []indexer.TokenAmount{
						{
							PolicyID: token.PolicyID,
							Name:     token.Name,
							Amount:   50,
						},
					},
				},
			},
			{
				Input: indexer.TxInput{
					Hash: indexer.Hash{1, 2, 3},
				},
				Output: indexer.TxOutput{
					Address: multisigAddr,
					Amount:  10,
					Tokens: []indexer.TokenAmount{
						{
							PolicyID: token.PolicyID,
							Name:     token.Name,
							Amount:   160,
						},
					},
				},
			},
			{
				Input: indexer.TxInput{
					Hash: indexer.Hash{1, 2, 9},
				},
				Output: indexer.TxOutput{
					Address: multisigAddr,
					Amount:  2073290,
					Tokens: []indexer.TokenAmount{
						{
							PolicyID: token.PolicyID,
							Name:     token.Name,
							Amount:   160,
						},
					},
				},
			},
		}
		feeUtxos := []*indexer.TxInputOutput{
			{
				Output: indexer.TxOutput{
					Amount: 260,
				},
			},
			{
				Output: indexer.TxOutput{
					Amount: 6260,
				},
			},
		}

		dbMock.On("GetAllTxOutputs", multisigAddr, true).Return(multisigUtxos, nil).Once()
		dbMock.On("GetAllTxOutputs", feeAddr, true).Return(feeUtxos, nil).Once()

		utxoSelectionResult, err := cco.getUTXOsForNormalBatch(
			[]common.AddressAndAmount{{
				AddressIndex: 0,
				Address:      multisigAddr,
				TokensAmounts: map[string]uint64{
					cardanowallet.AdaTokenName: 2_000_000,
				},
			}}, feeAddr, false)

		require.NoError(t, err)
		require.Equal(t, []*indexer.TxInputOutput{
			multisigUtxos[0], multisigUtxos[2],
		}, utxoSelectionResult.multisigUtxos[0])
		require.Equal(t, feeUtxos[:cco.config.MaxFeeUtxoCount], utxoSelectionResult.feeUtxos)
	})
}

// if tokens are passed as parameters, two of them are required
func generateSmallUtxoOutputs(value, n uint64, tokens ...cardanowallet.Token) ([]*indexer.TxInputOutput, uint64) {
	utxoOutput := make([]*indexer.TxInputOutput, 0, n)
	returnCurrencySum := uint64(0)

	for i := uint64(0); i < n; i++ {
		tx := &indexer.TxInputOutput{
			Input: indexer.TxInput{
				Hash: indexer.NewHashFromHexString(fmt.Sprintf("0x00%d", i)),
			},
			Output: indexer.TxOutput{
				Amount: value,
			},
		}

		if len(tokens) > 0 {
			token := tokens[int(i)%len(tokens)]

			tx.Output.Tokens = append(tx.Output.Tokens, indexer.TokenAmount{
				PolicyID: token.PolicyID,
				Name:     token.Name,
				Amount:   value,
			})
		}

		utxoOutput = append(utxoOutput, tx)
		returnCurrencySum += value
	}

	return utxoOutput, returnCurrencySum
}
