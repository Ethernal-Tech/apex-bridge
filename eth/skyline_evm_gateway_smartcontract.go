package eth

import (
	"context"
	"fmt"
	"math/big"

	skylinegatewaycontractbinding "github.com/Ethernal-Tech/apex-bridge/contractbinding/gateway/skyline"
	"github.com/Ethernal-Tech/ethgo"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hashicorp/go-hclog"
)

const registerTokenGasLimitMultiplier = 1.7

type ISkylineEVMGatewaySmartContract interface {
	IEVMGatewaySmartContract
	RegisterToken(
		ctx context.Context, lockUnlockSCAddress ethcommon.Address,
		tokenID uint16, name string, symbol string,
	) (*skylinegatewaycontractbinding.GatewayTokenRegistered, error)
}

type SkylineEVMGatewaySmartContractImpl struct {
	smartContractAddress ethcommon.Address
	ethHelper            *EthHelperWrapper
	depositGasLimit      uint64
	gasPrice             *big.Int
	gasFeeCap            *big.Int
	gasTipCap            *big.Int
}

var _ ISkylineEVMGatewaySmartContract = (*SkylineEVMGatewaySmartContractImpl)(nil)

func NewSkylineEVMGatewaySmartContract(
	smartContractAddress string, ethHelper *EthHelperWrapper, depositGasLimit uint64,
	gasPrice, gasFeeCap, gasTipCap *big.Int, logger hclog.Logger,
) (*SkylineEVMGatewaySmartContractImpl, error) {
	return &SkylineEVMGatewaySmartContractImpl{
		smartContractAddress: ethcommon.HexToAddress(smartContractAddress),
		ethHelper:            ethHelper,
		depositGasLimit:      depositGasLimit,
		gasPrice:             gasPrice,
		gasFeeCap:            gasFeeCap,
		gasTipCap:            gasTipCap,
	}, nil
}

func NewSimpleSkylineEVMGatewaySmartContract(
	smartContractAddress string, ethHelper *EthHelperWrapper, logger hclog.Logger,
) (*SkylineEVMGatewaySmartContractImpl, error) {
	return &SkylineEVMGatewaySmartContractImpl{
		smartContractAddress: ethcommon.HexToAddress(smartContractAddress),
		ethHelper:            ethHelper,
	}, nil
}

//nolint:dupl
func (bsc *SkylineEVMGatewaySmartContractImpl) Deposit(
	ctx context.Context, signature []byte, bitmap *big.Int, data []byte,
) error {
	parsedABI, err := skylinegatewaycontractbinding.GatewayMetaData.GetAbi()
	if err != nil {
		return fmt.Errorf("error while GatewayMetaData.GetAbi(): %w", err)
	}

	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return fmt.Errorf("error while GetEthHelper: %w", err)
	}

	contract, err := skylinegatewaycontractbinding.NewGateway(bsc.smartContractAddress, ethTxHelper.GetClient())
	if err != nil {
		return fmt.Errorf("error while NewGateway: %w", bsc.ethHelper.ProcessError(err))
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
			return fmt.Errorf("error while EstimateGas: %w", bsc.ethHelper.ProcessError(err))
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
	if err != nil {
		return fmt.Errorf("error while SendTx Deposit: %w", bsc.ethHelper.ProcessError(err))
	}

	return nil
}

func (bsc *SkylineEVMGatewaySmartContractImpl) RegisterToken(
	ctx context.Context, lockUnlockSCAddress ethcommon.Address, tokenID uint16, name string, symbol string,
) (*skylinegatewaycontractbinding.GatewayTokenRegistered, error) {
	parsedABI, err := skylinegatewaycontractbinding.GatewayMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("error while GatewayMetaData.GetAbi(): %w", err)
	}

	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return nil, fmt.Errorf("error while GetEthHelper: %w", err)
	}

	contract, err := skylinegatewaycontractbinding.NewGateway(bsc.smartContractAddress, ethTxHelper.GetClient())
	if err != nil {
		return nil, fmt.Errorf("error while NewGateway: %w", bsc.ethHelper.ProcessError(err))
	}

	bsc.ethHelper.logger.Debug("Estimating gas for RegisterToken",
		"wallet", bsc.ethHelper.wallet.GetAddress(),
		"contract", bsc.smartContractAddress)

	estimatedGas, estimatedGasOriginal, err := ethTxHelper.EstimateGas(
		ctx, bsc.ethHelper.wallet.GetAddress(), bsc.smartContractAddress, nil, registerTokenGasLimitMultiplier,
		parsedABI, "registerToken", lockUnlockSCAddress, tokenID, name, symbol)
	if err != nil {
		return nil, fmt.Errorf("error while EstimateGas: %w", bsc.ethHelper.ProcessError(err))
	}

	bsc.ethHelper.logger.Debug("Estimated gas for RegisterToken", "gas", estimatedGas, "original", estimatedGasOriginal,
		"wallet", bsc.ethHelper.wallet.GetAddress(), "contract", bsc.smartContractAddress)

	receipt, err := bsc.ethHelper.SendTx(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		opts.GasLimit = estimatedGas

		return contract.RegisterToken(opts, lockUnlockSCAddress, tokenID, name, symbol)
	})
	if err != nil {
		return nil, fmt.Errorf("error while SendTx RegisterToken: %w", bsc.ethHelper.ProcessError(err))
	}

	event, err := extractTokenRegisteredEvent(contract, receipt)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func extractTokenRegisteredEvent(contract *skylinegatewaycontractbinding.Gateway, receipt *types.Receipt) (
	*skylinegatewaycontractbinding.GatewayTokenRegistered, error,
) {
	eventSigs, err := GetGatewayRegisterTokenEventSignatures()
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway register token event signatures. err: %w", err)
	}

	var (
		tokenRegisteredEventSig = eventSigs[0]
		tokenRegisteredEvent    *skylinegatewaycontractbinding.GatewayTokenRegistered
	)

	for _, log := range receipt.Logs {
		if len(log.Topics) == 0 {
			continue
		}

		if eventSig := ethgo.Hash(log.Topics[0]); eventSig != tokenRegisteredEventSig {
			continue
		}

		tokenRegisteredEvent, err = contract.GatewayFilterer.ParseTokenRegistered(*log)
		if err != nil {
			return nil, fmt.Errorf("failed parsing tokenRegistered log. err: %w", err)
		}
	}

	if tokenRegisteredEvent == nil {
		return nil, fmt.Errorf("no tokenRegistered event found")
	}

	return tokenRegisteredEvent, nil
}
