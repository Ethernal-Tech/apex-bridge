package eth

import (
	"context"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hashicorp/go-hclog"
)

const depositGasLimitMultiplier = 1.7

type IEVMGatewaySmartContract interface {
	Deposit(ctx context.Context, signature []byte, bitmap *big.Int, data []byte) error
}

type EVMGatewaySmartContractImpl struct {
	smartContractAddress ethcommon.Address
	ethHelper            *EthHelperWrapper
	depositGasLimit      uint64
	gasPrice             *big.Int
	gasFeeCap            *big.Int
	gasTipCap            *big.Int
}

var _ IEVMGatewaySmartContract = (*EVMGatewaySmartContractImpl)(nil)

func NewEVMGatewaySmartContractWithWallet(
	nodeURL, smartContractAddress string, wallet *ethtxhelper.EthTxWallet, isDynamic bool, depositGasLimit uint64,
	gasPrice, gasFeeCap, gasTipCap *big.Int, logger hclog.Logger,
) (*EVMGatewaySmartContractImpl, error) {
	return &EVMGatewaySmartContractImpl{
		smartContractAddress: ethcommon.HexToAddress(smartContractAddress),
		ethHelper:            NewEthHelperWrapperWithWallet(nodeURL, wallet, isDynamic, logger),
		depositGasLimit:      depositGasLimit,
		gasPrice:             gasPrice,
		gasFeeCap:            gasFeeCap,
		gasTipCap:            gasTipCap,
	}, nil
}

func (bsc *EVMGatewaySmartContractImpl) Deposit(
	ctx context.Context, signature []byte, bitmap *big.Int, data []byte,
) error {
	parsedABI, err := contractbinding.GatewayMetaData.GetAbi()
	if err != nil {
		return err
	}

	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return err
	}

	contract, err := contractbinding.NewGateway(bsc.smartContractAddress, ethTxHelper.GetClient())
	if err != nil {
		return bsc.ethHelper.ProcessError(err)
	}

	var estimatedGas, estimatedGasOriginal uint64

	if bsc.depositGasLimit > 0 {
		estimatedGas = bsc.depositGasLimit
	} else {
		bsc.ethHelper.logger.Debug("Estimating gas for deposit",
			"wallet", bsc.ethHelper.wallet.GetAddress(),
			"contract", bsc.smartContractAddress)

		estimatedGas, estimatedGasOriginal, err = ethTxHelper.EstimateGas(
			ctx, bsc.ethHelper.wallet.GetAddress(), bsc.smartContractAddress, nil, depositGasLimitMultiplier,
			parsedABI, "deposit", signature, bitmap, data)
		if err != nil {
			return bsc.ethHelper.ProcessError(err)
		}
	}

	bsc.ethHelper.logger.Debug("Estimated gas for deposit", "gas", estimatedGas, "original", estimatedGasOriginal,
		"wallet", bsc.ethHelper.wallet.GetAddress(), "contract", bsc.smartContractAddress)

	_, err = bsc.ethHelper.SendTx(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		opts.GasLimit = estimatedGas
		opts.GasPrice = bsc.gasPrice
		opts.GasFeeCap = bsc.gasFeeCap
		opts.GasTipCap = bsc.gasTipCap

		return contract.Deposit(opts, signature, bitmap, data)
	})

	return bsc.ethHelper.ProcessError(err)
}
