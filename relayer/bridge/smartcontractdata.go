package bridge

import (
	"context"
	"encoding/hex"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
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

	v, err := contract.GetConfirmedBatch(&bind.CallOpts{
		Context: ctx,
	}, destinationChain)
	if err != nil {
		return nil, err
	}

	// Convert string arrays to byte arrays
	var multisigSignatures [][]byte
	for _, sig := range v.MultisigSignatures {
		sigBytes, err := hex.DecodeString(sig)
		if err != nil {
			return nil, err
		}
		multisigSignatures = append(multisigSignatures, sigBytes)
	}

	var feePayerMultisigSignatures [][]byte
	for _, sig := range v.FeePayerMultisigSignatures {
		sigBytes, err := hex.DecodeString(sig)
		if err != nil {
			return nil, err
		}
		feePayerMultisigSignatures = append(feePayerMultisigSignatures, sigBytes)
	}

	// Convert rawTransaction from string to byte array
	rawTx, err := hex.DecodeString(v.RawTransaction)
	if err != nil {
		return nil, err
	}

	return &ConfirmedBatch{
		Id:                         v.Id,
		RawTransaction:             rawTx,
		MultisigSignatures:         multisigSignatures,
		FeePayerMultisigSignatures: feePayerMultisigSignatures,
	}, nil
}
