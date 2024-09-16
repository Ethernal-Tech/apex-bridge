package ethcontracts

import (
	"context"
	"fmt"

	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type IEthContractUtils interface {
	DeployWithProxy(
		ctx context.Context,
		artifact *Artifact,
		proxyArtifact *Artifact,
		initParams ...interface{},
	) (ethcommon.Address, ethcommon.Address, error)
	ExecuteMethod(
		ctx context.Context,
		artifact *Artifact,
		address ethcommon.Address,
		method string,
		args ...interface{},
	) (ethcommon.Hash, error)
}

type ethContractUtils struct {
	txHelper           ethtxhelper.IEthTxHelper
	wallet             ethtxhelper.IEthTxWallet
	gasLimitMultipiler float64
	waitForTx          bool
}

func NewEthContractUtils(
	txHelper ethtxhelper.IEthTxHelper, wallet ethtxhelper.IEthTxWallet, gasLimitMultiplies float64, waitForTx bool,
) IEthContractUtils {
	return &ethContractUtils{
		txHelper:           txHelper,
		wallet:             wallet,
		gasLimitMultipiler: gasLimitMultiplies,
		waitForTx:          waitForTx,
	}
}

func (ecu *ethContractUtils) DeployWithProxy(
	ctx context.Context,
	artifact *Artifact,
	proxyArtifact *Artifact,
	initParams ...interface{},
) (ethcommon.Address, ethcommon.Address, error) {
	addrString, txHash, err := ecu.txHelper.Deploy(
		ctx, ecu.wallet, bind.TransactOpts{}, *artifact.Abi, artifact.Bytecode)
	if err != nil {
		return ethcommon.Address{}, ethcommon.Address{}, err
	}

	receipt, err := ecu.txHelper.WaitForReceipt(ctx, txHash, true)
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

	addrProxyStr, txHash, err := ecu.txHelper.Deploy(
		ctx, ecu.wallet, bind.TransactOpts{}, *proxyArtifact.Abi, proxyArtifact.Bytecode, addr, initializationData)
	if err != nil {
		return ethcommon.Address{}, ethcommon.Address{}, err
	}

	receipt, err = ecu.txHelper.WaitForReceipt(ctx, txHash, true)
	if err != nil {
		return ethcommon.Address{}, ethcommon.Address{}, err
	} else if receipt.Status != types.ReceiptStatusSuccessful {
		return ethcommon.Address{}, ethcommon.Address{},
			fmt.Errorf("receipt status for %s is %d", addrProxyStr, receipt.Status)
	}

	return ethcommon.HexToAddress(addrProxyStr), addr, nil
}

func (ecu *ethContractUtils) ExecuteMethod(
	ctx context.Context,
	artifact *Artifact,
	address ethcommon.Address,
	method string,
	args ...interface{},
) (ethcommon.Hash, error) {
	estimatedGas, _, err := ecu.txHelper.EstimateGas(
		ctx, ecu.wallet.GetAddress(), address, nil, ecu.gasLimitMultipiler, artifact.Abi, method, args...)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	bc := bind.NewBoundContract(
		address, *artifact.Abi, ecu.txHelper.GetClient(), ecu.txHelper.GetClient(), ecu.txHelper.GetClient())

	tx, err := ecu.txHelper.SendTx(ctx, ecu.wallet, bind.TransactOpts{},
		func(opts *bind.TransactOpts) (*types.Transaction, error) {
			opts.GasLimit = estimatedGas

			return bc.Transact(opts, method, args...)
		})
	if err != nil {
		return ethcommon.Hash{}, err
	}

	if ecu.waitForTx {
		receipt, err := ecu.txHelper.WaitForReceipt(ctx, tx.Hash().String(), true)
		if err != nil {
			return ethcommon.Hash{}, err
		} else if receipt.Status != types.ReceiptStatusSuccessful {
			return ethcommon.Hash{}, fmt.Errorf("receipt status for %s is %d", tx.Hash(), receipt.Status)
		}
	}

	return tx.Hash(), nil
}
