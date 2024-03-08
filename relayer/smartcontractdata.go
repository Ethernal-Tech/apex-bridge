package relayer

import (
	"context"
	"encoding/hex"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ConfirmedBatch struct {
	id                         string
	rawTransaction             []byte
	multisigSignatures         [][]byte
	feePayerMultisigSignatures [][]byte
}

type IChainOperations interface {
	GetConfirmedBatch(ctx context.Context, ethClient *ethclient.Client, smartContractAddress string) (*ConfirmedBatch, error)
}

type PrimeChainOperations struct {
	chainId string
}

func NewPrimeChainOperations(chainId string) *PrimeChainOperations {
	return &PrimeChainOperations{
		chainId: chainId,
	}
}

func (c *PrimeChainOperations) GetConfirmedBatch(ctx context.Context, ethClient *ethclient.Client, smartContractAddress string) (*ConfirmedBatch, error) {
	return getSmartContractData(ctx, ethClient, smartContractAddress, c.chainId)
}

type VectorChainOperations struct {
	chainId string
}

func NewVectorChainOperations(chainId string) *VectorChainOperations {
	return &VectorChainOperations{
		chainId: chainId,
	}
}

func (c *VectorChainOperations) GetConfirmedBatch(ctx context.Context, ethClient *ethclient.Client, smartContractAddress string) (*ConfirmedBatch, error) {
	return getSmartContractData(ctx, ethClient, smartContractAddress, c.chainId)
}

func GetOperations(testnetMagic uint) IChainOperations {
	switch testnetMagic {
	case 1:
		return NewPrimeChainOperations("prime")
	case 2:
		return NewVectorChainOperations("vector")
	}

	return nil
}

func getSmartContractData(ctx context.Context, ethClient *ethclient.Client, smartContractAddress string, destinationChain string) (*ConfirmedBatch, error) {
	ethTxHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithClient(ethClient))
	if err != nil {
		// In case of error, reset ethClient to nil to try again in the next iteration.
		ethClient = nil
		return nil, err
	}

	contract, err := contractbinding.NewTestContract(
		common.HexToAddress(smartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		// In case of error, reset ethClient to nil to try again in the next iteration.
		ethClient = nil
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
		id:                         v.Id,
		rawTransaction:             rawTx,
		multisigSignatures:         multisigSignatures,
		feePayerMultisigSignatures: feePayerMultisigSignatures,
	}, nil
}
