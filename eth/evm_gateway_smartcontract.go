package eth

import (
	"context"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hashicorp/go-hclog"
)

const depositGasLimit = uint64(8_000_000)

type IEVMGatewaySmartContract interface {
	Deposit(ctx context.Context, signature []byte, bitmap []byte, data []byte) error
}

type EVMGatewaySmartContractImpl struct {
	smartContractAddress string
	ethHelper            *EthHelperWrapper
}

var _ IEVMGatewaySmartContract = (*EVMGatewaySmartContractImpl)(nil)

func NewEVMGatewaySmartContract(
	nodeURL, smartContractAddress string, isDynamic bool, logger hclog.Logger,
) *EVMGatewaySmartContractImpl {
	return &EVMGatewaySmartContractImpl{
		smartContractAddress: smartContractAddress,
		ethHelper:            NewEthHelperWrapper(nodeURL, isDynamic, logger),
	}
}

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
	ctx context.Context, signature []byte, bitmap []byte, data []byte,
) error {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return err
	}

	contract, err := contractbinding.NewGateway(
		common.HexToAddress(bsc.smartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return bsc.ethHelper.ProcessError(err)
	}

	_, err = bsc.ethHelper.SendTx(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		opts.GasLimit = depositGasLimit

		return contract.Deposit(opts, signature, bitmap, data)
	})

	return bsc.ethHelper.ProcessError(err)
}
