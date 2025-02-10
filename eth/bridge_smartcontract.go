package eth

import (
	"context"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	setChainAdditionalDataGasLimitMultipiler = 1.2
)

type Chain = contractbinding.IBridgeStructsChain

type IBridgeSmartContract interface {
	GetConfirmedBatch(
		ctx context.Context, destinationChain string) (*ConfirmedBatch, error)
	SubmitSignedBatch(ctx context.Context, signedBatch SignedBatch, gasLimit uint64) error
	SubmitSignedBatchEVM(ctx context.Context, signedBatch SignedBatch, gasLimit uint64) error
	ShouldCreateBatch(ctx context.Context, destinationChain string) (bool, error)
	GetConfirmedTransactions(ctx context.Context, destinationChain string) ([]ConfirmedTransaction, error)
	GetLastObservedBlock(ctx context.Context, destinationChain string) (CardanoBlock, error)
	GetValidatorsChainData(ctx context.Context, destinationChain string) ([]ValidatorChainData, error)
	GetNextBatchID(ctx context.Context, destinationChain string) (uint64, error)
	GetAllRegisteredChains(ctx context.Context) ([]Chain, error)
	GetBlockNumber(ctx context.Context) (uint64, error)
	SetChainAdditionalData(ctx context.Context, chainID, multisigAddr, feeAddr string) error
	GetBatchTransactions(ctx context.Context, chainID string, batchID uint64) ([]TxDataInfo, error)
}

type BridgeSmartContractImpl struct {
	smartContractAddress ethcommon.Address
	ethHelper            *EthHelperWrapper
}

var _ IBridgeSmartContract = (*BridgeSmartContractImpl)(nil)

func NewBridgeSmartContract(
	smartContractAddress string, ethHelper *EthHelperWrapper,
) *BridgeSmartContractImpl {
	return &BridgeSmartContractImpl{
		smartContractAddress: common.HexToAddress(smartContractAddress),
		ethHelper:            ethHelper,
	}
}

func (bsc *BridgeSmartContractImpl) GetConfirmedBatch(
	ctx context.Context, destinationChain string,
) (*ConfirmedBatch, error) {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return nil, fmt.Errorf("error while GetEthHelper: %w", err)
	}

	contract, err := contractbinding.NewBridgeContract(
		bsc.smartContractAddress,
		ethTxHelper.GetClient())
	if err != nil {
		return nil, fmt.Errorf("error while NewBridgeContract: %w", bsc.ethHelper.ProcessError(err))
	}

	result, err := contract.GetConfirmedBatch(&bind.CallOpts{
		Context: ctx,
	}, common.ToNumChainID(destinationChain))
	if err != nil {
		return nil, fmt.Errorf("error while GetConfirmedBatch: %w", bsc.ethHelper.ProcessError(err))
	}

	return NewConfirmedBatch(result), nil
}

func (bsc *BridgeSmartContractImpl) SubmitSignedBatch(
	ctx context.Context, signedBatch SignedBatch, gasLimit uint64,
) error {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return fmt.Errorf("error while GetEthHelper: %w", err)
	}

	contract, err := contractbinding.NewBridgeContract(
		bsc.smartContractAddress,
		ethTxHelper.GetClient())
	if err != nil {
		return fmt.Errorf("error while NewBridgeContract: %w", bsc.ethHelper.ProcessError(err))
	}

	_, err = bsc.ethHelper.SendTx(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		opts.GasLimit = gasLimit

		return contract.SubmitSignedBatch(opts, signedBatch)
	})
	if err != nil {
		return fmt.Errorf("error while SendTx SubmitSignedBatch: %w", bsc.ethHelper.ProcessError(err))
	}

	return nil
}

func (bsc *BridgeSmartContractImpl) SubmitSignedBatchEVM(
	ctx context.Context, signedBatch SignedBatch, gasLimit uint64,
) error {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return fmt.Errorf("error while GetEthHelper: %w", err)
	}

	contract, err := contractbinding.NewBridgeContract(
		bsc.smartContractAddress,
		ethTxHelper.GetClient())
	if err != nil {
		return fmt.Errorf("error while NewBridgeContract: %w", bsc.ethHelper.ProcessError(err))
	}

	_, err = bsc.ethHelper.SendTx(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		opts.GasLimit = gasLimit

		return contract.SubmitSignedBatchEVM(opts, signedBatch)
	})
	if err != nil {
		return fmt.Errorf("error while SendTx SubmitSignedBatchEVM: %w", bsc.ethHelper.ProcessError(err))
	}

	return nil
}

func (bsc *BridgeSmartContractImpl) ShouldCreateBatch(ctx context.Context, destinationChain string) (bool, error) {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return false, fmt.Errorf("error while GetEthHelper: %w", err)
	}

	contract, err := contractbinding.NewBridgeContract(
		bsc.smartContractAddress,
		ethTxHelper.GetClient())
	if err != nil {
		return false, fmt.Errorf("error while NewBridgeContract: %w", bsc.ethHelper.ProcessError(err))
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
		return nil, fmt.Errorf("error while GetEthHelper: %w", err)
	}

	contract, err := contractbinding.NewBridgeContract(
		bsc.smartContractAddress,
		ethTxHelper.GetClient())
	if err != nil {
		return nil, fmt.Errorf("error while NewBridgeContract: %w", bsc.ethHelper.ProcessError(err))
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
		return CardanoBlock{}, fmt.Errorf("error while GetEthHelper: %w", err)
	}

	contract, err := contractbinding.NewBridgeContract(
		bsc.smartContractAddress,
		ethTxHelper.GetClient())
	if err != nil {
		return CardanoBlock{}, fmt.Errorf("error while NewBridgeContract: %w", bsc.ethHelper.ProcessError(err))
	}

	cardanoBlock, err := contract.GetLastObservedBlock(&bind.CallOpts{
		Context: ctx,
	}, common.ToNumChainID(destinationChain))
	if err != nil {
		return CardanoBlock{}, fmt.Errorf("error while GetLastObservedBlock: %w", bsc.ethHelper.ProcessError(err))
	}

	return cardanoBlock, nil
}

func (bsc *BridgeSmartContractImpl) GetValidatorsChainData(
	ctx context.Context, destinationChain string,
) ([]ValidatorChainData, error) {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return nil, fmt.Errorf("error while GetEthHelper: %w", err)
	}

	contract, err := contractbinding.NewBridgeContract(
		bsc.smartContractAddress,
		ethTxHelper.GetClient())
	if err != nil {
		return nil, fmt.Errorf("error while NewBridgeContract: %w", bsc.ethHelper.ProcessError(err))
	}

	return contract.GetValidatorsChainData(&bind.CallOpts{
		Context: ctx,
	}, common.ToNumChainID(destinationChain))
}

func (bsc *BridgeSmartContractImpl) GetNextBatchID(ctx context.Context, destinationChain string) (uint64, error) {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return 0, fmt.Errorf("error while GetEthHelper: %w", err)
	}

	contract, err := contractbinding.NewBridgeContract(
		bsc.smartContractAddress,
		ethTxHelper.GetClient())
	if err != nil {
		return 0, fmt.Errorf("error while NewBridgeContract: %w", bsc.ethHelper.ProcessError(err))
	}

	return contract.GetNextBatchId(&bind.CallOpts{
		Context: ctx,
	}, common.ToNumChainID(destinationChain))
}

func (bsc *BridgeSmartContractImpl) GetAllRegisteredChains(ctx context.Context) ([]Chain, error) {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return nil, fmt.Errorf("error while GetEthHelper: %w", err)
	}

	contract, err := contractbinding.NewBridgeContract(
		bsc.smartContractAddress,
		ethTxHelper.GetClient())
	if err != nil {
		return nil, fmt.Errorf("error while NewBridgeContract: %w", bsc.ethHelper.ProcessError(err))
	}

	result, err := contract.GetAllRegisteredChains(&bind.CallOpts{
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("error while GetAllRegisteredChains: %w", bsc.ethHelper.ProcessError(err))
	}

	return result, nil
}

func (bsc *BridgeSmartContractImpl) GetBlockNumber(ctx context.Context) (uint64, error) {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return 0, fmt.Errorf("error while GetEthHelper: %w", err)
	}

	return ethTxHelper.GetClient().BlockNumber(ctx)
}

func (bsc *BridgeSmartContractImpl) SetChainAdditionalData(
	ctx context.Context, chainID, multisigAddr, feeAddr string,
) error {
	parsedABI, err := contractbinding.BridgeContractMetaData.GetAbi()
	if err != nil {
		return err
	}

	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return fmt.Errorf("error while GetEthHelper: %w", err)
	}

	contract, err := contractbinding.NewBridgeContract(
		bsc.smartContractAddress,
		ethTxHelper.GetClient())
	if err != nil {
		return fmt.Errorf("error while NewBridgeContract: %w", bsc.ethHelper.ProcessError(err))
	}

	chainIDNum := common.ToNumChainID(chainID)

	estimatedGas, _, err := ethTxHelper.EstimateGas(
		ctx, bsc.ethHelper.wallet.GetAddress(),
		bsc.smartContractAddress, nil, setChainAdditionalDataGasLimitMultipiler,
		parsedABI, "setChainAdditionalData", chainIDNum, multisigAddr, feeAddr)
	if err != nil {
		return err
	}

	_, err = bsc.ethHelper.SendTx(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		opts.GasLimit = estimatedGas

		return contract.SetChainAdditionalData(opts, chainIDNum, multisigAddr, feeAddr)
	})
	if err != nil {
		return fmt.Errorf("error while SendTx SetChainAdditionalData: %w", bsc.ethHelper.ProcessError(err))
	}

	return nil
}

func (bsc *BridgeSmartContractImpl) GetBatchTransactions(
	ctx context.Context, chainID string, batchID uint64,
) ([]TxDataInfo, error) {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return nil, fmt.Errorf("error while GetEthHelper: %w", err)
	}

	contract, err := contractbinding.NewBridgeContract(
		bsc.smartContractAddress,
		ethTxHelper.GetClient())
	if err != nil {
		return nil, fmt.Errorf("error while NewBridgeContract: %w", bsc.ethHelper.ProcessError(err))
	}

	result, err := contract.GetBatchTransactions(&bind.CallOpts{
		Context: ctx,
	}, common.ToNumChainID(chainID), batchID)
	if err != nil {
		return nil, fmt.Errorf("error while GetBatchTransactions: %w", bsc.ethHelper.ProcessError(err))
	}

	return result, nil
}
