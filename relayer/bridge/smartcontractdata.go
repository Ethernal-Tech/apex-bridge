package bridge

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type ConfirmedBatch struct {
	Id                         string
	RawTransaction             []byte
	MultisigSignatures         [][]byte
	FeePayerMultisigSignatures [][]byte
}

func GetSmartContractData(ctx context.Context, ethTxHelper ethtxhelper.IEthTxHelper, destinationChain string, smartContractAddress string) (*ConfirmedBatch, error) {
	contract, err := contractbinding.NewTestContract(
		common.HexToAddress(smartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return nil, err
	}

	confirmedBatch, err := contract.GetConfirmedBatch(&bind.CallOpts{
		Context: ctx,
	}, destinationChain)
	if err != nil {
		return nil, err
	}

	// Convert string arrays to byte arrays
	var multisigSignatures [][]byte
	for _, sig := range confirmedBatch.MultisigSignatures {
		sigBytes, err := hex.DecodeString(sig)
		if err != nil {
			return nil, err
		}
		multisigSignatures = append(multisigSignatures, sigBytes)
	}

	var feePayerMultisigSignatures [][]byte
	for _, sig := range confirmedBatch.FeePayerMultisigSignatures {
		sigBytes, err := hex.DecodeString(sig)
		if err != nil {
			return nil, err
		}
		feePayerMultisigSignatures = append(feePayerMultisigSignatures, sigBytes)
	}

	// Convert rawTransaction from string to byte array
	rawTx, err := hex.DecodeString(confirmedBatch.RawTransaction)
	if err != nil {
		return nil, err
	}

	return &ConfirmedBatch{
		Id:                         confirmedBatch.Id,
		RawTransaction:             rawTx,
		MultisigSignatures:         multisigSignatures,
		FeePayerMultisigSignatures: feePayerMultisigSignatures,
	}, nil
}

// TODO: Remove - added for testing
func ShouldRetreive(ctx context.Context, ethTxHelper ethtxhelper.IEthTxHelper, smartContractAddress string) (bool, error) {
	contract, err := contractbinding.NewTestContract(
		common.HexToAddress(smartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return false, err
	}

	return contract.ShouldRelayerRetrieve(&bind.CallOpts{
		Context: ctx,
	})
}

// TODO: Remove - added for testing
func ResetShouldRetreive(ctx context.Context, ethTxHelper ethtxhelper.IEthTxHelper, smartContractAddress string) error {
	contract, err := contractbinding.NewTestContract(
		common.HexToAddress(smartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return err
	}

	wallet, err := ethtxhelper.NewEthTxWallet("3761f6deeb2e0b2aa8b843e804d880afa6e5fecf1631f411e267641a72d0ca20")

	tx, err := ethTxHelper.SendTx(ctx, wallet, bind.TransactOpts{}, true, func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
		return contract.ResetShouldRetrieve(txOpts)
	})
	if err != nil {
		return err
	}

	receipt, err := ethTxHelper.WaitForReceipt(ctx, tx.Hash().String(), true)
	if err != nil {
		return err
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("Not successfull")
	}

	return nil
}
