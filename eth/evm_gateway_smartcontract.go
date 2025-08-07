package eth

import (
	"context"
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hashicorp/go-hclog"
)

const (
	depositGasLimitMultiplier   = 1.7
	updateVCDGasLimitMultiplier = depositGasLimitMultiplier // potentially set to an different value
)

type IEVMGatewaySmartContract interface {
	Deposit(ctx context.Context, signature []byte, bitmap *big.Int, data []byte) error
	UpdateValidatorsChainData(ctx context.Context, signature []byte, bitmap *big.Int, data []byte) error
}

type EVMGatewaySmartContractImpl struct {
	smartContractAddress ethcommon.Address
	ethHelper            *EthHelperWrapper
	depositGasLimit      uint64
	updateVCDGasLimit    uint64
	gasPrice             *big.Int
	gasFeeCap            *big.Int
	gasTipCap            *big.Int
}

var _ IEVMGatewaySmartContract = (*EVMGatewaySmartContractImpl)(nil)

func NewEVMGatewaySmartContract(
	smartContractAddress string, ethHelper *EthHelperWrapper, depositGasLimit uint64,
	gasPrice, gasFeeCap, gasTipCap *big.Int, logger hclog.Logger,
) (*EVMGatewaySmartContractImpl, error) {
	return &EVMGatewaySmartContractImpl{
		smartContractAddress: ethcommon.HexToAddress(smartContractAddress),
		ethHelper:            ethHelper,
		depositGasLimit:      depositGasLimit,
		updateVCDGasLimit:    depositGasLimit, // potentially set to different value
		gasPrice:             gasPrice,
		gasFeeCap:            gasFeeCap,
		gasTipCap:            gasTipCap,
	}, nil
}

func (bsc *EVMGatewaySmartContractImpl) Deposit(
	ctx context.Context,
	signature []byte,
	bitmap *big.Int,
	data []byte) error {
	return bsc.sendTx(ctx, signature, bitmap, data, depositToGatewayTxType)
}

func (bsc *EVMGatewaySmartContractImpl) UpdateValidatorsChainData(
	ctx context.Context,
	signature []byte,
	bitmap *big.Int,
	data []byte) error {
	return bsc.sendTx(ctx, signature, bitmap, data, updateVCDToGatewayTxType)
}

type toGatewayTxType uint8

const (
	depositToGatewayTxType = iota
	updateVCDToGatewayTxType
)

func (bsc *EVMGatewaySmartContractImpl) sendTx(
	ctx context.Context,
	signature []byte,
	bitmap *big.Int,
	data []byte,
	txType toGatewayTxType,
) error {
	parsedABI, err := contractbinding.GatewayMetaData.GetAbi()
	if err != nil {
		return fmt.Errorf("error while GatewayMetaData.GetAbi(): %w", err)
	}

	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return fmt.Errorf("error while GetEthHelper: %w", err)
	}

	contract, err := contractbinding.NewGateway(bsc.smartContractAddress, ethTxHelper.GetClient())
	if err != nil {
		return fmt.Errorf("error while NewGateway: %w", bsc.ethHelper.ProcessError(err))
	}

	var estimatedGas, estimatedGasOriginal uint64

	switch txType {
	case depositToGatewayTxType:
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
				return fmt.Errorf("error while EstimateGas: %w", bsc.ethHelper.ProcessError(err))
			}
		}

		bsc.ethHelper.logger.Debug("Estimated gas for deposit", "gas", estimatedGas, "original", estimatedGasOriginal,
			"wallet", bsc.ethHelper.wallet.GetAddress(), "contract", bsc.smartContractAddress)
	case updateVCDToGatewayTxType:
		if bsc.updateVCDGasLimit > 0 {
			estimatedGas = bsc.updateVCDGasLimit
		} else {
			bsc.ethHelper.logger.Debug("Estimating gas for update validators chain data",
				"wallet", bsc.ethHelper.wallet.GetAddress(),
				"contract", bsc.smartContractAddress)

			estimatedGas, estimatedGasOriginal, err = ethTxHelper.EstimateGas(
				ctx, bsc.ethHelper.wallet.GetAddress(), bsc.smartContractAddress, nil, updateVCDGasLimitMultiplier,
				parsedABI, "updateValidatorsChainData", signature, bitmap, data)
			if err != nil {
				return fmt.Errorf("error while EstimateGas: %w", bsc.ethHelper.ProcessError(err))
			}
		}

		bsc.ethHelper.logger.Debug("Estimated gas for update validators chain data", "gas", estimatedGas,
			"original", estimatedGasOriginal,
			"wallet", bsc.ethHelper.wallet.GetAddress(), "contract", bsc.smartContractAddress)
	default:
		return fmt.Errorf("unknown transaction type to be sent to gateway")
	}

	_, err = bsc.ethHelper.SendTx(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		opts.GasLimit = estimatedGas
		opts.GasPrice = bsc.gasPrice
		opts.GasFeeCap = bsc.gasFeeCap
		opts.GasTipCap = bsc.gasTipCap

		// Note: there's no need to check if txType is something other than deposit or updateVCD,
		// since any other value wouldn't reach this point — it would be handled by the switch above.
		if txType == depositToGatewayTxType {
			return contract.Deposit(opts, signature, bitmap, data)
		}

		return contract.UpdateValidatorsChainData(opts, signature, bitmap, data)
	})
	if err != nil {
		return fmt.Errorf("error while SendTx: %w", bsc.ethHelper.ProcessError(err))
	}

	return nil
}
