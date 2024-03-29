package eth

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
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
	nodeUrl              string
	ethTxHelper          ethtxhelper.IEthTxHelper
	wallet               ethtxhelper.IEthTxWallet
}

func NewBridgeSmartContract(nodeUrl, smartContractAddress string) *BridgeSmartContractImpl {
	return &BridgeSmartContractImpl{
		nodeUrl:              nodeUrl,
		smartContractAddress: smartContractAddress,
	}
}

func NewBridgeSmartContractWithWallet(nodeUrl, smartContractAddress, signingKey string) (*BridgeSmartContractImpl, error) {
	ethWallet, err := ethtxhelper.NewEthTxWallet(signingKey)
	if err != nil {
		return nil, err
	}

	return &BridgeSmartContractImpl{
		nodeUrl:              nodeUrl,
		smartContractAddress: smartContractAddress,
		wallet:               ethWallet,
	}, nil
}

func (bsc *BridgeSmartContractImpl) GetConfirmedBatch(ctx context.Context, destinationChain string) (*ConfirmedBatch, error) {
	ethTxHelper, err := bsc.getEthHelper()
	if err != nil {
		return nil, err
	}

	contract, err := contractbinding.NewBridgeContract(
		common.HexToAddress(bsc.smartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return nil, bsc.processError(err)
	}

	result, err := contract.GetConfirmedBatch(&bind.CallOpts{
		Context: ctx,
	}, destinationChain)
	if err != nil {
		return nil, bsc.processError(err)
	}

	return NewConfirmedBatch(result)
}

func (bsc *BridgeSmartContractImpl) SubmitSignedBatch(ctx context.Context, signedBatch SignedBatch) error {
	ethTxHelper, err := bsc.getEthHelper()
	if err != nil {
		return err
	}

	contract, err := contractbinding.NewTestContract(
		common.HexToAddress(bsc.smartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return err
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

	_, err = bsc.sendTx(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return contract.SubmitSignedBatch(opts, newSignedBatch)
	})

	return err
}

func (bsc *BridgeSmartContractImpl) getEthHelper(opts ...ethtxhelper.TxRelayerOption) (ethtxhelper.IEthTxHelper, error) {
	if bsc.ethTxHelper != nil {
		return bsc.ethTxHelper, nil
	}

	ethTxHelper, err := ethtxhelper.NewEThTxHelper(opts...)
	if err != nil {
		return nil, err
	}

	bsc.ethTxHelper = ethTxHelper

	return ethTxHelper, nil
}

func (bsc *BridgeSmartContractImpl) processError(err error) error {
	// TODO: handle connection lost error to trigger recreation of a eth tx helper/client
	if errors.Is(err, errors.New("connection lost")) {
		bsc.ethTxHelper = nil
	}

	return err
}

// sendTx should be called by all public methods that sends tx to the bridge
func (bsc *BridgeSmartContractImpl) sendTx(ctx context.Context, handler ethtxhelper.SendTxFunc) (string, error) {
	ethTxHelper, err := bsc.getEthHelper()
	if err != nil {
		return "", err
	}

	tx, err := ethTxHelper.SendTx(ctx, bsc.wallet, bind.TransactOpts{}, true, handler)
	if err != nil {
		return "", bsc.processError(err)
	}

	// TODO: enable logs bsc.logger.Info("tx has been sent", "tx hash", tx.Hash().String())

	receipt, err := ethTxHelper.WaitForReceipt(ctx, tx.Hash().String(), true)
	if err != nil {
		return "", bsc.processError(err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return receipt.BlockHash.String(), fmt.Errorf("receipts status not successful: %v", receipt.Status)
	}
	// TODO: enable logs  bsc.logger.Info("tx has been executed", "block", receipt.BlockHash.String(), "tx hash", receipt.TxHash.String())

	return receipt.BlockHash.String(), nil
}
