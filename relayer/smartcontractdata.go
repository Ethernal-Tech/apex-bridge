package relayer

import (
	"context"
	"encoding/hex"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type SmartContractData struct {
	id                         string
	rawTransaction             []byte
	multisigSignatures         [][]byte
	feePayerMultisigSignatures [][]byte
}

// TODO: update smart contract query with real parameter
func (r Relayer) getSmartContractData(ctx context.Context, ethTxHelper ethtxhelper.IEthTxHelper) (*SmartContractData, error) {
	contract, err := contractbinding.NewTestContract(
		common.HexToAddress(r.config.Bridge.SmartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return nil, err // TODO: recoverable error?
	}

	v, err := contract.GetConfirmedBatch(&bind.CallOpts{
		Context: ctx,
	}, "destinationChain")
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

	return &SmartContractData{
		id:                         v.Id,
		rawTransaction:             rawTx,
		multisigSignatures:         multisigSignatures,
		feePayerMultisigSignatures: feePayerMultisigSignatures,
	}, nil
}
