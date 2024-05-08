package eth

import (
	"context"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
)

type Chain = contractbinding.IBridgeContractStructsChain

type IBridgeSmartContract interface {
	GetConfirmedBatch(
		ctx context.Context, destinationChain string) (*ConfirmedBatch, error)
	SubmitSignedBatch(ctx context.Context, signedBatch SignedBatch) error
	ShouldCreateBatch(ctx context.Context, destinationChain string) (bool, error)
	GetConfirmedTransactions(ctx context.Context, destinationChain string) ([]ConfirmedTransaction, error)
	GetAvailableUTXOs(ctx context.Context, destinationChain string) (*UTXOs, error)
	GetLastObservedBlock(ctx context.Context, destinationChain string) (*CardanoBlock, error)
	GetValidatorsCardanoData(ctx context.Context, destinationChain string) ([]ValidatorCardanoData, error)
	GetNextBatchID(ctx context.Context, destinationChain string) (*big.Int, error)
	GetAllRegisteredChains(ctx context.Context) ([]Chain, error)
}

type BridgeSmartContractImpl struct {
	smartContractAddress string
	ethHelper            *EthHelperWrapper
}

var _ IBridgeSmartContract = (*BridgeSmartContractImpl)(nil)

func NewBridgeSmartContract(nodeURL, smartContractAddress string, isDynamic bool) *BridgeSmartContractImpl {
	return &BridgeSmartContractImpl{
		smartContractAddress: smartContractAddress,
		ethHelper:            NewEthHelperWrapper(nodeURL, isDynamic),
	}
}

func NewBridgeSmartContractWithWallet(
	nodeURL, smartContractAddress string, wallet *ethtxhelper.EthTxWallet, isDynamic bool,
) (*BridgeSmartContractImpl, error) {
	ethHelper, err := NewEthHelperWrapperWithWallet(nodeURL, wallet, isDynamic)
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
) (
	*ConfirmedBatch, error,
) {
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
	}, destinationChain)
	if err != nil {
		return nil, bsc.ethHelper.ProcessError(err)
	}

	return NewConfirmedBatch(result)
}

func (bsc *BridgeSmartContractImpl) SubmitSignedBatch(ctx context.Context, signedBatch SignedBatch) error {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return err
	}

	contract, err := contractbinding.NewBridgeContract(
		common.HexToAddress(bsc.smartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return bsc.ethHelper.ProcessError(err)
	}

	newSignedBatch := SignedBatch{
		Id:                        signedBatch.Id,
		DestinationChainId:        signedBatch.DestinationChainId,
		RawTransaction:            signedBatch.RawTransaction,
		MultisigSignature:         signedBatch.MultisigSignature,
		FeePayerMultisigSignature: signedBatch.FeePayerMultisigSignature,
		IncludedTransactions:      []*big.Int{},
		UsedUTXOs:                 UTXOs{},
	}

	_, err = bsc.ethHelper.SendTx(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return contract.SubmitSignedBatch(opts, newSignedBatch)
	})

	return bsc.ethHelper.ProcessError(err)
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
	}, destinationChain)
}

func (bsc *BridgeSmartContractImpl) GetConfirmedTransactions(
	ctx context.Context, destinationChain string,
) (
	[]ConfirmedTransaction, error,
) {
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
	}, destinationChain)
}

func (bsc *BridgeSmartContractImpl) GetAvailableUTXOs(ctx context.Context, destinationChain string) (*UTXOs, error) {
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

	availableUtxos, err := contract.GetAvailableUTXOs(&bind.CallOpts{
		Context: ctx,
	}, destinationChain)
	if err != nil {
		return nil, bsc.ethHelper.ProcessError(err)
	}

	return &availableUtxos, nil
}

// GetLastObservedBlock implements IBridgeSmartContract.
func (bsc *BridgeSmartContractImpl) GetLastObservedBlock(
	ctx context.Context, destinationChain string,
) (
	*CardanoBlock, error,
) {
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

	cardanoBlock, err := contract.GetLastObservedBlock(&bind.CallOpts{
		Context: ctx,
	}, destinationChain)
	if err != nil {
		return nil, bsc.ethHelper.ProcessError(err)
	}

	return &cardanoBlock, nil
}

func (bsc *BridgeSmartContractImpl) GetValidatorsCardanoData(
	ctx context.Context, destinationChain string,
) (
	[]ValidatorCardanoData, error,
) {
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

	return contract.GetValidatorsCardanoData(&bind.CallOpts{
		Context: ctx,
	}, destinationChain)
}

func (bsc *BridgeSmartContractImpl) GetNextBatchID(ctx context.Context, destinationChain string) (*big.Int, error) {
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

	return contract.GetNextBatchId(&bind.CallOpts{
		Context: ctx,
	}, destinationChain)
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
