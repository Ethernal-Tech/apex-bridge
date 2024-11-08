package ethcontracts

import (
	"context"
	"time"

	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	defaultNumRetries      = 10
	defaultRetriesWaitTime = time.Second * 4
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
	numRetries         int
	retriesWaitTime    time.Duration
}

func NewEthContractUtils(
	txHelper ethtxhelper.IEthTxHelper, wallet ethtxhelper.IEthTxWallet, gasLimitMultiplies float64,
) IEthContractUtils {
	return &ethContractUtils{
		txHelper:           txHelper,
		wallet:             wallet,
		gasLimitMultipiler: gasLimitMultiplies,
		numRetries:         defaultNumRetries,
		retriesWaitTime:    defaultRetriesWaitTime,
	}
}

func (ecu *ethContractUtils) DeployWithProxy(
	ctx context.Context,
	artifact *Artifact,
	proxyArtifact *Artifact,
	initParams ...interface{},
) (proxyAddr ethcommon.Address, proxyTxHash string, addr ethcommon.Address, txHash string, err error) {
	var addrString string

	err = wallet.ExecuteWithRetry(ctx, ecu.numRetries, ecu.retriesWaitTime, func() (bool, error) {
		addrString, txHash, err = ecu.txHelper.Deploy(
			ctx, ecu.wallet, bind.TransactOpts{}, *artifact.Abi, artifact.Bytecode)

		return err == nil, err
	})
	if err != nil {
		return proxyAddr, proxyTxHash, addr, txHash, err
	}

	addr = ethcommon.HexToAddress(addrString)

	err = wallet.ExecuteWithRetry(ctx, ecu.numRetries, ecu.retriesWaitTime, func() (bool, error) {
		initializationData, err := artifact.Abi.Pack("initialize", initParams...)
		if err != nil {
			return false, err
		}

		addrString, proxyTxHash, err = ecu.txHelper.Deploy(
			ctx, ecu.wallet, bind.TransactOpts{}, *proxyArtifact.Abi, proxyArtifact.Bytecode, addr, initializationData)

		return err == nil, err
	})
	if err != nil {
		return proxyAddr, proxyTxHash, addr, txHash, err
	}

	proxyAddr = ethcommon.HexToAddress(addrString)

	return proxyAddr, proxyTxHash, addr, txHash, nil
}

func (ecu *ethContractUtils) ExecuteMethod(
	ctx context.Context,
	artifact *Artifact,
	address ethcommon.Address,
	method string,
	args ...interface{},
) (string, error) {
	var tx *types.Transaction

	err := wallet.ExecuteWithRetry(ctx, ecu.numRetries, ecu.retriesWaitTime, func() (bool, error) {
		boundContract := bind.NewBoundContract(
			address, *artifact.Abi, ecu.txHelper.GetClient(), ecu.txHelper.GetClient(), ecu.txHelper.GetClient())

		estimatedGas, _, err := ecu.txHelper.EstimateGas(
			ctx, ecu.wallet.GetAddress(), address, nil, ecu.gasLimitMultipiler, artifact.Abi, method, args...)
		if err != nil {
			return false, err
		}

		tx, err = ecu.txHelper.SendTx(ctx, ecu.wallet, bind.TransactOpts{},
			func(opts *bind.TransactOpts) (*types.Transaction, error) {
				opts.GasLimit = estimatedGas

				return boundContract.Transact(opts, method, args...)
			})

		return err == nil, err
	})
	if err != nil {
		return "", err
	}

	return tx.Hash().String(), nil
}
