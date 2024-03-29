package eth

import (
	"context"
	"errors"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type IBridgeSmartContract interface {
	GetConfirmedBatch(
		ctx context.Context, destinationChain string) (*ConfirmedBatch, error)
}

type BridgeSmartContractImpl struct {
	smartContractAddress string
	nodeUrl              string
	ethTxHelper          ethtxhelper.IEthTxHelper
}

func NewBridgeSmartContract(nodeUrl, smartContractAddress string) *BridgeSmartContractImpl {
	return &BridgeSmartContractImpl{
		nodeUrl:              nodeUrl,
		smartContractAddress: smartContractAddress,
	}
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
		return nil, err
	}

	result, err := contract.GetConfirmedBatch(&bind.CallOpts{
		Context: ctx,
	}, destinationChain)
	if err != nil {
		return nil, bsc.processError(err)
	}

	return NewConfirmedBatch(result)
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
