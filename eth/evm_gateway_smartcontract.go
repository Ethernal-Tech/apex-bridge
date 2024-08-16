package eth

import (
	"context"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hashicorp/go-hclog"
)

const depositGasLimitMultiplier = 1.7

type IEVMGatewaySmartContract interface {
	Deposit(ctx context.Context, signature []byte, bitmap *big.Int, data []byte) error
}

type EVMGatewaySmartContractImpl struct {
	smartContractAddress string
	ethHelper            *EthHelperWrapper
}

var _ IEVMGatewaySmartContract = (*EVMGatewaySmartContractImpl)(nil)

func NewEVMGatewaySmartContractWithWallet(
	nodeURL, smartContractAddress string, wallet *ethtxhelper.EthTxWallet, isDynamic bool, logger hclog.Logger,
) (*EVMGatewaySmartContractImpl, error) {
	ethHelper, err := NewEthHelperWrapperWithWallet(nodeURL, wallet, isDynamic, logger)
	if err != nil {
		return nil, err
	}

	return &EVMGatewaySmartContractImpl{
		smartContractAddress: smartContractAddress,
		ethHelper:            ethHelper,
	}, nil
}

func (bsc *EVMGatewaySmartContractImpl) Deposit(
	ctx context.Context, signature []byte, bitmap *big.Int, data []byte,
) error {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return err
	}

	toAddress := common.HexToAddress(bsc.smartContractAddress)

	contract, err := contractbinding.NewGateway(toAddress, ethTxHelper.GetClient())
	if err != nil {
		return bsc.ethHelper.ProcessError(err)
	}

	estimatedGas, estimatedGasOriginal, err := ethTxHelper.EstimateGas(
		context.Background(), bsc.ethHelper.wallet.GetAddress(), toAddress, nil, depositGasLimitMultiplier,
		contractbinding.GatewayMetaData, "deposit", signature, bitmap, data)
	if err != nil {
		return bsc.ethHelper.ProcessError(err)
	}

	bsc.ethHelper.logger.Debug("Estimated gas for deposit", "gas", estimatedGas, "original", estimatedGasOriginal)

	_, err = bsc.ethHelper.SendTx(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		opts.GasLimit = estimatedGas

		return contract.Deposit(opts, signature, bitmap, data)
	})

	return bsc.ethHelper.ProcessError(err)
}
