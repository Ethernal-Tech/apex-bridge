package ethcontracts

import (
	"context"
	"fmt"

	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func DeployContractWithProxy(
	ctx context.Context,
	txHelper ethtxhelper.IEthTxHelper,
	wallet ethtxhelper.IEthTxWallet,
	artifact *Artifact,
	proxyArtifact *Artifact,
	initParams ...interface{},
) (ethcommon.Address, ethcommon.Address, error) {
	addrString, txHash, err := txHelper.Deploy(ctx, wallet, bind.TransactOpts{}, *artifact.Abi, artifact.Bytecode)
	if err != nil {
		return ethcommon.Address{}, ethcommon.Address{}, err
	}

	receipt, err := txHelper.WaitForReceipt(ctx, txHash, true)
	if err != nil {
		return ethcommon.Address{}, ethcommon.Address{}, err
	} else if receipt.Status != types.ReceiptStatusSuccessful {
		return ethcommon.Address{}, ethcommon.Address{},
			fmt.Errorf("receipt status for %s is %d", addrString, receipt.Status)
	}

	addr := ethcommon.HexToAddress(addrString)

	initializationData, err := artifact.Abi.Pack("initialize", initParams...)
	if err != nil {
		return ethcommon.Address{}, ethcommon.Address{}, err
	}

	addrProxyStr, txHash, err := txHelper.Deploy(
		ctx, wallet, bind.TransactOpts{}, *proxyArtifact.Abi, proxyArtifact.Bytecode, addr, initializationData)
	if err != nil {
		return ethcommon.Address{}, ethcommon.Address{}, err
	}

	receipt, err = txHelper.WaitForReceipt(ctx, txHash, true)
	if err != nil {
		return ethcommon.Address{}, ethcommon.Address{}, err
	} else if receipt.Status != types.ReceiptStatusSuccessful {
		return ethcommon.Address{}, ethcommon.Address{},
			fmt.Errorf("receipt status for %s is %d", addrProxyStr, receipt.Status)
	}

	return ethcommon.HexToAddress(addrProxyStr), addr, nil
}

func ExecuteContractMethod(
	ctx context.Context,
	txHelper ethtxhelper.IEthTxHelper,
	wallet ethtxhelper.IEthTxWallet,
	artifact *Artifact,
	gasLimitMultipiler float64,
	waitForTx bool,
	address ethcommon.Address,
	method string,
	args ...interface{},
) (string, error) {
	estimatedGas, _, err := txHelper.EstimateGas(
		ctx, wallet.GetAddress(), address, nil, gasLimitMultipiler,
		artifact.Abi, method, args...)
	if err != nil {
		return "", err
	}

	bc := bind.NewBoundContract(
		address, *artifact.Abi, txHelper.GetClient(), txHelper.GetClient(), txHelper.GetClient())

	tx, err := txHelper.SendTx(ctx, wallet, bind.TransactOpts{},
		func(opts *bind.TransactOpts) (*types.Transaction, error) {
			opts.GasLimit = estimatedGas

			return bc.Transact(opts, method, args...)
		})
	if err != nil {
		return "", err
	}

	txHash := tx.Hash().String()

	if waitForTx {
		receipt, err := txHelper.WaitForReceipt(ctx, txHash, true)
		if err != nil {
			return "", err
		} else if receipt.Status != types.ReceiptStatusSuccessful {
			return "", fmt.Errorf("receipt status for %s is %d", txHash, receipt.Status)
		}
	}

	return txHash, nil
}
