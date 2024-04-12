package ethtxhelper

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

const (
	mumbaiNodeUrl    = "https://polygon-mumbai.blockpi.network/v1/rpc/public"
	dummyMumbaiAccPk = "61deed8dda92a396e8e9dbcbb5a058bee274de1adc57b2067975691dacdd55c7"
)

var (
	scBytecode, _ = hex.DecodeString("6080604052348015600e575f80fd5b506101438061001c5f395ff3fe608060405234801561000f575f80fd5b5060043610610034575f3560e01c806320965255146100385780635524107714610056575b5f80fd5b610040610072565b60405161004d919061009b565b60405180910390f35b610070600480360381019061006b91906100e2565b61007a565b005b5f8054905090565b805f8190555050565b5f819050919050565b61009581610083565b82525050565b5f6020820190506100ae5f83018461008c565b92915050565b5f80fd5b6100c181610083565b81146100cb575f80fd5b50565b5f813590506100dc816100b8565b92915050565b5f602082840312156100f7576100f66100b4565b5b5f610104848285016100ce565b9150509291505056fea264697066735822122033aed3a54af0b72781ebf6e2338f64b95cabcd37258eeef0474b43dbf54f99fd64736f6c63430008190033")
)

func TestTxHelper(t *testing.T) {
	// this test works with real live node, so it should be run only on special occasions
	t.Skip()

	scAddress := "0xA4B44930686aC7179FC63688749EC60e0C4A96C3"

	wallet, err := NewEthTxWallet(dummyMumbaiAccPk)
	require.NoError(t, err)

	txHelper, err := NewEThTxHelper(WithNodeUrl(mumbaiNodeUrl), WithGasFeeMultiplier(1.3))
	require.NoError(t, err)

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelCtx()

	t.Run("deploy smart contract", func(t *testing.T) {
		abiData, err := contractbinding.TestContractMetaData.GetAbi()
		require.NoError(t, err)

		nonce, err := txHelper.GetNonce(ctx, wallet.GetAddress().String(), false)
		require.NoError(t, err)

		addr, hash, err := txHelper.Deploy(ctx, new(big.Int).SetUint64(nonce),
			uint64(300000), false, *abiData, scBytecode, wallet)
		require.NoError(t, err)
		require.NotEqual(t, common.Address{}, addr)

		receipt, err := txHelper.WaitForReceipt(ctx, hash, true)
		require.NoError(t, err)
		require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
		require.Equal(t, common.HexToAddress(addr), receipt.ContractAddress)

		scAddress = addr
	})

	t.Run("sending smart contract transaction and query smart contract", func(t *testing.T) {
		valueToSet := uint64(time.Now().UTC().UnixNano())

		contract, err := contractbinding.NewTestContract(common.HexToAddress(scAddress), txHelper.GetClient())
		require.NoError(t, err)

		res, err := contract.GetValue(&bind.CallOpts{
			Context: ctx,
			From:    wallet.GetAddress(),
		})
		require.NoError(t, err)
		require.False(t, new(big.Int).SetUint64(valueToSet).Cmp(res) == 0)

		// first call is just for creating tx
		tx, err := txHelper.SendTx(ctx, wallet, bind.TransactOpts{}, true, func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
			return contract.SetValue(txOpts, new(big.Int).SetUint64(valueToSet))
		})
		require.NoError(t, err)

		receipt, err := txHelper.WaitForReceipt(ctx, tx.Hash().String(), true)
		require.NoError(t, err)
		require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)

		// check value again
		res, err = contract.GetValue(&bind.CallOpts{
			Context: ctx,
			From:    wallet.GetAddress(),
		})
		require.NoError(t, err)
		require.True(t, new(big.Int).SetUint64(valueToSet).Cmp(res) == 0)
	})

	t.Run("send transfer transaction legacy", func(t *testing.T) {
		const (
			ethAddr  = "0xBa65B75FDA35561626A455b1aF806A7C58A57DdE"
			ethValue = uint64(2001)
		)

		client, ok := txHelper.GetClient().(*ethclient.Client)
		require.True(t, ok)

		chainID, err := client.ChainID(ctx)
		require.NoError(t, err)

		oldVal, err := client.BalanceAt(ctx, common.HexToAddress(ethAddr), nil)
		require.NoError(t, err)

		// first call is just for creating tx
		txOpts := bind.TransactOpts{
			Value:    new(big.Int).SetUint64(ethValue),
			GasLimit: 21000, // default value for transfer
		}

		err = txHelper.PopulateTxOpts(ctx, wallet.GetAddress(), false, &txOpts)
		require.NoError(t, err)

		tx := TxOpts2LegacyTx(ethAddr, []byte{}, &txOpts)

		signedTx, err := wallet.SignTx(chainID, tx)
		require.NoError(t, err)

		err = client.SendTransaction(ctx, signedTx)
		require.NoError(t, err)

		receipt, err := txHelper.WaitForReceipt(ctx, signedTx.Hash().String(), true)
		require.NoError(t, err)
		require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)

		// check value again
		newVal, err := client.BalanceAt(ctx, common.HexToAddress(ethAddr), nil)
		require.NoError(t, err)

		desiredVal := new(big.Int).Add(new(big.Int).SetUint64(ethValue), oldVal)
		require.True(t, desiredVal.Cmp(newVal) == 0)
	})

	t.Run("send transfer transaction dynamicfee", func(t *testing.T) {
		const (
			ethAddr  = "0xBa65B75FDA35561626A455b1aF806A7C58A57DdE"
			ethValue = uint64(2001)
		)

		client, ok := txHelper.GetClient().(*ethclient.Client)
		require.True(t, ok)

		chainID, err := client.ChainID(ctx)
		require.NoError(t, err)

		oldVal, err := client.BalanceAt(ctx, common.HexToAddress(ethAddr), nil)
		require.NoError(t, err)

		// first call is just for creating tx
		txOpts := bind.TransactOpts{
			Value:    new(big.Int).SetUint64(ethValue),
			GasLimit: 21000, // default value for transfer
		}

		err = txHelper.PopulateTxOpts(ctx, wallet.GetAddress(), true, &txOpts)
		require.NoError(t, err)

		tx := TxOpts2DynamicFeeTx(ethAddr, chainID, []byte{}, &txOpts)

		signedTx, err := wallet.SignTx(chainID, tx)
		require.NoError(t, err)

		err = client.SendTransaction(ctx, signedTx)
		require.NoError(t, err)

		receipt, err := txHelper.WaitForReceipt(ctx, signedTx.Hash().String(), true)
		require.NoError(t, err)
		require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)

		// check value again
		newVal, err := client.BalanceAt(ctx, common.HexToAddress(ethAddr), nil)
		require.NoError(t, err)

		desiredVal := new(big.Int).Add(new(big.Int).SetUint64(ethValue), oldVal)
		require.True(t, desiredVal.Cmp(newVal) == 0)
	})
}
