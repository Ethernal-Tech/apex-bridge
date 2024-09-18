package ethcontracts

import (
	"context"

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
	) (ethcommon.Address, string, ethcommon.Address, string, error)
	ExecuteMethod(
		ctx context.Context,
		artifact *Artifact,
		address ethcommon.Address,
		method string,
		args ...interface{},
	) (string, error)
}

type ethContractUtils struct {
	txHelper           ethtxhelper.IEthTxHelper
	wallet             ethtxhelper.IEthTxWallet
	gasLimitMultipiler float64
}

func NewEthContractUtils(
	txHelper ethtxhelper.IEthTxHelper, wallet ethtxhelper.IEthTxWallet, gasLimitMultiplies float64,
) IEthContractUtils {
	return &ethContractUtils{
		txHelper:           txHelper,
		wallet:             wallet,
		gasLimitMultipiler: gasLimitMultiplies,
	}
}

func (ecu *ethContractUtils) DeployWithProxy(
	ctx context.Context,
	artifact *Artifact,
	proxyArtifact *Artifact,
	initParams ...interface{},
) (ethcommon.Address, string, ethcommon.Address, string, error) {
	addrString, txHash, err := ecu.txHelper.Deploy(
		ctx, ecu.wallet, bind.TransactOpts{}, *artifact.Abi, artifact.Bytecode)
	if err != nil {
		return ethcommon.Address{}, "", ethcommon.Address{}, "", err
	}

	addr := ethcommon.HexToAddress(addrString)

	initializationData, err := artifact.Abi.Pack("initialize", initParams...)
	if err != nil {
		return ethcommon.Address{}, "", ethcommon.Address{}, "", err
	}

	addrProxyStr, txHashProxy, err := ecu.txHelper.Deploy(
		ctx, ecu.wallet, bind.TransactOpts{}, *proxyArtifact.Abi, proxyArtifact.Bytecode, addr, initializationData)
	if err != nil {
		return ethcommon.Address{}, "", ethcommon.Address{}, "", err
	}

	return ethcommon.HexToAddress(addrProxyStr), txHashProxy, addr, txHash, nil
}

func (ecu *ethContractUtils) ExecuteMethod(
	ctx context.Context,
	artifact *Artifact,
	address ethcommon.Address,
	method string,
	args ...interface{},
) (string, error) {
	estimatedGas, _, err := ecu.txHelper.EstimateGas(
		ctx, ecu.wallet.GetAddress(), address, nil, ecu.gasLimitMultipiler, artifact.Abi, method, args...)
	if err != nil {
		return "", err
	}

	bc := bind.NewBoundContract(
		address, *artifact.Abi, ecu.txHelper.GetClient(), ecu.txHelper.GetClient(), ecu.txHelper.GetClient())

	tx, err := ecu.txHelper.SendTx(ctx, ecu.wallet, bind.TransactOpts{},
		func(opts *bind.TransactOpts) (*types.Transaction, error) {
			opts.GasLimit = estimatedGas

			return bc.Transact(opts, method, args...)
		})
	if err != nil {
		return "", err
	}

	return tx.Hash().String(), nil
}
