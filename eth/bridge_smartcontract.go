package eth

import (
	"context"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
)

type IBridgeSmartContract interface {
	GetConfirmedBatch(
		ctx context.Context, destinationChain string) (*ConfirmedBatch, error)
	SubmitSignedBatch(ctx context.Context, signedBatch SignedBatch) error
}

type BridgeSmartContractImpl struct {
	smartContractAddress string
	ethHelper            *EthHelperWrapper
}

func NewBridgeSmartContract(nodeUrl, smartContractAddress string) *BridgeSmartContractImpl {
	return &BridgeSmartContractImpl{
		smartContractAddress: smartContractAddress,
		ethHelper:            NewEthHelperWrapper(nodeUrl),
	}
}

func NewBridgeSmartContractWithWallet(nodeUrl, smartContractAddress, signingKey string) (*BridgeSmartContractImpl, error) {
	ethHelper, err := NewEthHelperWrapperWithWallet(nodeUrl, signingKey)
	if err != nil {
		return nil, err
	}

	return &BridgeSmartContractImpl{
		smartContractAddress: smartContractAddress,
		ethHelper:            ethHelper,
	}, nil
}

func (bsc *BridgeSmartContractImpl) GetConfirmedBatch(ctx context.Context, destinationChain string) (*ConfirmedBatch, error) {
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

	contract, err := contractbinding.NewTestContract(
		common.HexToAddress(bsc.smartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return bsc.ethHelper.ProcessError(err)
	}

	newSignedBatch := contractbinding.TestContractSignedBatch{
		Id:                        signedBatch.Id.String(),
		DestinationChainId:        signedBatch.DestinationChainId,
		RawTransaction:            signedBatch.RawTransaction,
		MultisigSignature:         signedBatch.MultisigSignature,
		FeePayerMultisigSignature: signedBatch.FeePayerMultisigSignature,
		IncludedTransactions:      []contractbinding.TestContractConfirmedTransaction{},
		UsedUTXOs:                 contractbinding.TestContractUTXOs{},
	}

	_, err = bsc.ethHelper.SendTx(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return contract.SubmitSignedBatch(opts, newSignedBatch)
	})

	return bsc.ethHelper.ProcessError(err)
}
