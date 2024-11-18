package ethcontracts

import (
	"context"
	"time"

	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
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
	) (ethtxhelper.TxDeployInfo, ethtxhelper.TxDeployInfo, error)
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
) (proxyTx ethtxhelper.TxDeployInfo, tx ethtxhelper.TxDeployInfo, err error) {
	tx, err = infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (ethtxhelper.TxDeployInfo, error) {
		return ecu.txHelper.Deploy(
			ctx, ecu.wallet, bind.TransactOpts{}, *artifact.Abi, artifact.Bytecode)
	}, infracommon.WithRetryCount(ecu.numRetries), infracommon.WithRetryWaitTime(ecu.retriesWaitTime))
	if err != nil {
		return proxyTx, tx, err
	}

	initializationData, err := artifact.Abi.Pack("initialize", initParams...)
	if err != nil {
		return proxyTx, tx, err
	}

	proxyTx, err = infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (ethtxhelper.TxDeployInfo, error) {
		return ecu.txHelper.Deploy(
			ctx, ecu.wallet, bind.TransactOpts{}, *proxyArtifact.Abi, proxyArtifact.Bytecode,
			tx.Address, initializationData)
	}, infracommon.WithRetryCount(ecu.numRetries), infracommon.WithRetryWaitTime(ecu.retriesWaitTime))

	return proxyTx, tx, err
}

func (ecu *ethContractUtils) ExecuteMethod(
	ctx context.Context,
	artifact *Artifact,
	address ethcommon.Address,
	method string,
	args ...interface{},
) (string, error) {
	tx, err := infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (*types.Transaction, error) {
		boundContract := bind.NewBoundContract(
			address, *artifact.Abi, ecu.txHelper.GetClient(), ecu.txHelper.GetClient(), ecu.txHelper.GetClient())

		estimatedGas, _, err := ecu.txHelper.EstimateGas(
			ctx, ecu.wallet.GetAddress(), address, nil, ecu.gasLimitMultipiler, artifact.Abi, method, args...)
		if err != nil {
			return nil, err
		}

		return ecu.txHelper.SendTx(ctx, ecu.wallet, bind.TransactOpts{},
			func(opts *bind.TransactOpts) (*types.Transaction, error) {
				opts.GasLimit = estimatedGas

				return boundContract.Transact(opts, method, args...)
			})
	}, infracommon.WithRetryCount(ecu.numRetries), infracommon.WithRetryWaitTime(ecu.retriesWaitTime))
	if err != nil {
		return "", err
	}

	return tx.Hash().String(), nil
}
