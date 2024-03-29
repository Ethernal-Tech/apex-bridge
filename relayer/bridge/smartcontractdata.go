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

func GetConfirmedBatch(ctx context.Context, ethTxHelper ethtxhelper.IEthTxHelper, destinationChain string, smartContractAddress string) (*ConfirmedBatch, error) {
	contract, err := contractbinding.NewBridgeContract(
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
		Id:                         confirmedBatch.Id.String(),
		RawTransaction:             rawTx,
		MultisigSignatures:         multisigSignatures,
		FeePayerMultisigSignatures: feePayerMultisigSignatures,
	}, nil
}
