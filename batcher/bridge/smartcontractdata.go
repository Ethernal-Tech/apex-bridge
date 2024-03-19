package bridge

import (
	"context"
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func ShouldCreateBatch(ctx context.Context, ethTxHelper ethtxhelper.IEthTxHelper, smartContractAddress string, destinationChain string) (bool, error) {
	contract, err := contractbinding.NewTestContract(
		common.HexToAddress(smartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return false, err
	}

	return contract.ShouldCreateBatch(&bind.CallOpts{
		Context: ctx,
	}, destinationChain)
}

func GetConfirmedTransactions(ctx context.Context, ethTxHelper ethtxhelper.IEthTxHelper, smartContractAddress string, destinationChain string) ([]contractbinding.TestContractConfirmedTransaction, error) {
	contract, err := contractbinding.NewTestContract(
		common.HexToAddress(smartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return nil, err
	}

	return contract.GetConfirmedTransactions(&bind.CallOpts{
		Context: ctx,
	}, destinationChain)
}

func GetAvailableUTXOs(ctx context.Context, ethTxHelper ethtxhelper.IEthTxHelper, smartContractAddress string, destinationChain string, txCost *big.Int) (*contractbinding.TestContractUTXOs, error) {
	contract, err := contractbinding.NewTestContract(
		common.HexToAddress(smartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return nil, err
	}

	availableUtxos, err := contract.GetAvailableUTXOs(&bind.CallOpts{
		Context: ctx,
	}, destinationChain, txCost)
	if err != nil {
		return nil, err
	}

	return &availableUtxos, nil
}

func SubmitSignedBatch(
	ctx context.Context,
	ethTxHelper ethtxhelper.IEthTxHelper,
	smartContractAddress string,
	signedBatch contractbinding.TestContractSignedBatch,
	signingKey string) error {
	contract, err := contractbinding.NewTestContract(
		common.HexToAddress(smartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return err
	}

	newSignedBatch := contractbinding.TestContractSignedBatch{
		Id:                        signedBatch.Id,
		DestinationChainId:        signedBatch.DestinationChainId,
		RawTransaction:            signedBatch.RawTransaction,
		MultisigSignature:         signedBatch.MultisigSignature,
		FeePayerMultisigSignature: signedBatch.FeePayerMultisigSignature,
		IncludedTransactions:      []contractbinding.TestContractConfirmedTransaction{},
		UsedUTXOs:                 contractbinding.TestContractUTXOs{},
	}

	wallet, err := ethtxhelper.NewEthTxWallet(signingKey)
	if err != nil {
		return err
	}

	tx, err := ethTxHelper.SendTx(ctx, wallet, bind.TransactOpts{}, true, func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
		return contract.SubmitSignedBatch(txOpts, newSignedBatch)
	})
	if err != nil {
		return err
	}

	receipt, err := ethTxHelper.WaitForReceipt(ctx, tx.Hash().String(), true)
	if err != nil {
		return err
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("Receipts status not successful: %v", receipt.Status)
	}
	return nil
}

var (
	dummyKeyHashes = []string{
		"eff5e22355217ec6d770c3668010c2761fa0863afa12e96cff8a2205",
		"ad8e0ab92e1febfcaf44889d68c3ae78b59dc9c5fa9e05a272214c13",
		"bfd1c0eb0a453a7b7d668166ce5ca779c655e09e11487a6fac72dd6f",
		"b4689f2e8f37b406c5eb41b1fe2c9e9f4eec2597c3cc31b8dfee8f56",
		"39c196d28f804f70704b6dec5991fbb1112e648e067d17ca7abe614b",
		"adea661341df075349cbb2ad02905ce1828f8cf3e66f5012d48c3168",
	}
)
