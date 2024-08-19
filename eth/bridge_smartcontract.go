package eth

import (
	"context"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hashicorp/go-hclog"
)

const submitBatchGasLimitMultiplier = 1.7

type Chain = contractbinding.IBridgeStructsChain

type IBridgeSmartContract interface {
	GetConfirmedBatch(
		ctx context.Context, destinationChain string) (*ConfirmedBatch, error)
	SubmitSignedBatch(ctx context.Context, signedBatch SignedBatch) error
	SubmitSignedBatchEVM(ctx context.Context, signedBatch SignedBatch) error
	ShouldCreateBatch(ctx context.Context, destinationChain string) (bool, error)
	GetConfirmedTransactions(ctx context.Context, destinationChain string) ([]ConfirmedTransaction, error)
	GetLastObservedBlock(ctx context.Context, destinationChain string) (CardanoBlock, error)
	GetValidatorsChainData(ctx context.Context, destinationChain string) ([]ValidatorChainData, error)
	GetNextBatchID(ctx context.Context, destinationChain string) (uint64, error)
	GetAllRegisteredChains(ctx context.Context) ([]Chain, error)
	GetBlockNumber(ctx context.Context) (uint64, error)
}

type BridgeSmartContractImpl struct {
	smartContractAddress string
	ethHelper            *EthHelperWrapper
}

var _ IBridgeSmartContract = (*BridgeSmartContractImpl)(nil)

func NewBridgeSmartContract(
	nodeURL, smartContractAddress string, isDynamic bool, logger hclog.Logger,
) *BridgeSmartContractImpl {
	return &BridgeSmartContractImpl{
		smartContractAddress: smartContractAddress,
		ethHelper:            NewEthHelperWrapper(nodeURL, isDynamic, logger),
	}
}

func NewBridgeSmartContractWithWallet(
	nodeURL, smartContractAddress string, wallet *ethtxhelper.EthTxWallet, isDynamic bool, logger hclog.Logger,
) (*BridgeSmartContractImpl, error) {
	ethHelper, err := NewEthHelperWrapperWithWallet(nodeURL, wallet, isDynamic, logger)
	if err != nil {
		return nil, err
	}

	return &BridgeSmartContractImpl{
		smartContractAddress: smartContractAddress,
		ethHelper:            ethHelper,
	}, nil
}

func (bsc *BridgeSmartContractImpl) GetConfirmedBatch(
	ctx context.Context, destinationChain string,
) (*ConfirmedBatch, error) {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return nil, err
	}

	contract, err := contractbinding.NewBridgeContract(
		common.HexToAddress(bsc.smartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return nil, bsc.ethHelper.ProcessError(err)
	}

	result, err := contract.GetConfirmedBatch(&bind.CallOpts{
		Context: ctx,
	}, common.ToNumChainID(destinationChain))
	if err != nil {
		return nil, bsc.ethHelper.ProcessError(err)
	}

	return NewConfirmedBatch(result)
}

func (bsc *BridgeSmartContractImpl) SubmitSignedBatch(ctx context.Context, signedBatch SignedBatch) error {
	return bsc.submitSignedBatch(ctx, signedBatch, false)
}

func (bsc *BridgeSmartContractImpl) SubmitSignedBatchEVM(ctx context.Context, signedBatch SignedBatch) error {
	return bsc.submitSignedBatch(ctx, signedBatch, true)
}

func (bsc *BridgeSmartContractImpl) ShouldCreateBatch(ctx context.Context, destinationChain string) (bool, error) {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return false, err
	}

	contract, err := contractbinding.NewBridgeContract(
		common.HexToAddress(bsc.smartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return false, bsc.ethHelper.ProcessError(err)
	}

	return contract.ShouldCreateBatch(&bind.CallOpts{
		Context: ctx,
	}, common.ToNumChainID(destinationChain))
}

func (bsc *BridgeSmartContractImpl) GetConfirmedTransactions(
	ctx context.Context, destinationChain string,
) ([]ConfirmedTransaction, error) {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return nil, err
	}

	contract, err := contractbinding.NewBridgeContract(
		common.HexToAddress(bsc.smartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return nil, bsc.ethHelper.ProcessError(err)
	}

	return contract.GetConfirmedTransactions(&bind.CallOpts{
		Context: ctx,
	}, common.ToNumChainID(destinationChain))
}

// GetLastObservedBlock implements IBridgeSmartContract.
func (bsc *BridgeSmartContractImpl) GetLastObservedBlock(
	ctx context.Context, destinationChain string,
) (CardanoBlock, error) {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return CardanoBlock{}, err
	}

	contract, err := contractbinding.NewBridgeContract(
		common.HexToAddress(bsc.smartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return CardanoBlock{}, bsc.ethHelper.ProcessError(err)
	}

	cardanoBlock, err := contract.GetLastObservedBlock(&bind.CallOpts{
		Context: ctx,
	}, common.ToNumChainID(destinationChain))
	if err != nil {
		return CardanoBlock{}, bsc.ethHelper.ProcessError(err)
	}

	return cardanoBlock, nil
}

func (bsc *BridgeSmartContractImpl) GetValidatorsChainData(
	ctx context.Context, destinationChain string,
) ([]ValidatorChainData, error) {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return nil, err
	}

	contract, err := contractbinding.NewBridgeContract(
		common.HexToAddress(bsc.smartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return nil, bsc.ethHelper.ProcessError(err)
	}

	return contract.GetValidatorsChainData(&bind.CallOpts{
		Context: ctx,
	}, common.ToNumChainID(destinationChain))
}

func (bsc *BridgeSmartContractImpl) GetNextBatchID(ctx context.Context, destinationChain string) (uint64, error) {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return 0, err
	}

	contract, err := contractbinding.NewBridgeContract(
		common.HexToAddress(bsc.smartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return 0, bsc.ethHelper.ProcessError(err)
	}

	return contract.GetNextBatchId(&bind.CallOpts{
		Context: ctx,
	}, common.ToNumChainID(destinationChain))
}

func (bsc *BridgeSmartContractImpl) GetAllRegisteredChains(ctx context.Context) ([]Chain, error) {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return nil, err
	}

	contract, err := contractbinding.NewBridgeContract(
		common.HexToAddress(bsc.smartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return nil, bsc.ethHelper.ProcessError(err)
	}

	result, err := contract.GetAllRegisteredChains(&bind.CallOpts{
		Context: ctx,
	})
	if err != nil {
		return nil, bsc.ethHelper.ProcessError(err)
	}

	return result, nil
}

func (bsc *BridgeSmartContractImpl) GetBlockNumber(ctx context.Context) (uint64, error) {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return 0, err
	}

	return ethTxHelper.GetClient().BlockNumber(ctx)
}

func (bsc *BridgeSmartContractImpl) submitSignedBatch(
	ctx context.Context, signedBatch SignedBatch, isEVM bool,
) error {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return err
	}

	var (
		method      string
		fromAddress = bsc.ethHelper.wallet.GetAddress()
		toAddress   = common.HexToAddress(bsc.smartContractAddress)
	)

	contract, err := contractbinding.NewBridgeContract(toAddress, ethTxHelper.GetClient())
	if err != nil {
		return bsc.ethHelper.ProcessError(err)
	}

	if isEVM {
		method = "submitSignedBatchEVM"
	} else {
		method = "submitSignedBatch"
	}

	estimatedGas, estimatedGasOriginal, err := ethTxHelper.EstimateGas(
		ctx, fromAddress, toAddress, nil, submitBatchGasLimitMultiplier,
		contractbinding.BridgeContractMetaData, method, signedBatch)
	if err != nil {
		return bsc.ethHelper.ProcessError(err)
	}

	bsc.ethHelper.logger.Debug(fmt.Sprintf("Estimated gas for %s", method),
		"gas", estimatedGas, "original", estimatedGasOriginal)

	_, err = bsc.ethHelper.SendTx(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		opts.GasLimit = estimatedGas

		if isEVM {
			return contract.SubmitSignedBatchEVM(opts, signedBatch)
		}

		return contract.SubmitSignedBatch(opts, signedBatch)
	})

	return bsc.ethHelper.ProcessError(err)
}
