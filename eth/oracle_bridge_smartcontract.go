package eth

import (
	"context"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
)

type CardanoBlock = contractbinding.IBridgeContractStructsCardanoBlock
type Claims = contractbinding.IBridgeContractStructsValidatorClaims

type IOracleBridgeSmartContract interface {
	GetLastObservedBlock(ctx context.Context, sourceChain string) (*CardanoBlock, error)
	GetExpectedTx(ctx context.Context, chainID string) (string, error) // TODO: replace with real when implemented
	SubmitClaims(ctx context.Context, claims Claims) error
	SubmitLastObservableBlocks(ctx context.Context, chainID string, blocks []CardanoBlock) error
}

type OracleBridgeSmartContractImpl struct {
	smartContractAddress string
	ethHelper            *EthHelperWrapper
}

var _ IOracleBridgeSmartContract = (*OracleBridgeSmartContractImpl)(nil)

func NewOracleBridgeSmartContract(nodeUrl, smartContractAddress string) *OracleBridgeSmartContractImpl {
	return &OracleBridgeSmartContractImpl{
		smartContractAddress: smartContractAddress,
		ethHelper:            NewEthHelperWrapper(nodeUrl),
	}
}

func NewOracleBridgeSmartContractWithWallet(nodeUrl, smartContractAddress, signingKey string) (*OracleBridgeSmartContractImpl, error) {
	ethHelper, err := NewEthHelperWrapperWithWallet(nodeUrl, signingKey)
	if err != nil {
		return nil, err
	}

	return &OracleBridgeSmartContractImpl{
		smartContractAddress: smartContractAddress,
		ethHelper:            ethHelper,
	}, nil
}

func (bsc *OracleBridgeSmartContractImpl) GetLastObservedBlock(ctx context.Context, sourceChain string) (*CardanoBlock, error) {
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

	result, err := contract.GetLastObservedBlock(&bind.CallOpts{
		Context: ctx,
	}, sourceChain)
	if err != nil {
		return nil, bsc.ethHelper.ProcessError(err)
	}

	return &result, nil
}

func (bsc *OracleBridgeSmartContractImpl) GetExpectedTx(ctx context.Context, chainID string) (string, error) {
	return "", nil
	// TODO: implement when done on SC
	/*
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

		result, err := contract.GetExpectedTxs(&bind.CallOpts{
			Context: ctx,
		})
		if err != nil {
			return nil, bsc.ethHelper.ProcessError(err)
		}

		return &result, nil
	*/
}

func (bsc *OracleBridgeSmartContractImpl) SubmitClaims(ctx context.Context, claims Claims) error {
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

	_, err = bsc.ethHelper.SendTx(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return contract.SubmitClaims(opts, claims)
	})

	return bsc.ethHelper.ProcessError(err)
}

func (bsc *OracleBridgeSmartContractImpl) SubmitLastObservableBlocks(ctx context.Context, chainID string, blocks []CardanoBlock) error {
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

	_, err = bsc.ethHelper.SendTx(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return contract.SubmitLastObservableBlocks(opts, chainID, blocks)
	})

	return bsc.ethHelper.ProcessError(err)
}
