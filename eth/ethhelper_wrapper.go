package eth

import (
	"context"
	"errors"
	"fmt"
	"net"

	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
)

type EthHelperWrapper struct {
	nodeUrl     string
	wallet      ethtxhelper.IEthTxWallet
	ethTxHelper ethtxhelper.IEthTxHelper
}

func NewEthHelperWrapper(nodeUrl string) *EthHelperWrapper {
	return &EthHelperWrapper{
		nodeUrl: nodeUrl,
	}
}

func NewEthHelperWrapperWithWallet(nodeUrl, signingKey string) (*EthHelperWrapper, error) {
	ethWallet, err := ethtxhelper.NewEthTxWallet(signingKey)
	if err != nil {
		return nil, err
	}

	return &EthHelperWrapper{
		nodeUrl: nodeUrl,
		wallet:  ethWallet,
	}, nil
}

func (e *EthHelperWrapper) GetEthHelper(opts ...ethtxhelper.TxRelayerOption) (ethtxhelper.IEthTxHelper, error) {
	if e.ethTxHelper != nil {
		return e.ethTxHelper, nil
	}

	ethTxHelper, err := ethtxhelper.NewEThTxHelper(opts...)
	if err != nil {
		return nil, err
	}

	e.ethTxHelper = ethTxHelper

	return ethTxHelper, nil
}

func (e *EthHelperWrapper) ProcessError(err error) error {
	// TODO: verify if these errors are the only ones we need to handle
	if errors.Is(err, net.ErrClosed) || errors.Is(err, context.DeadlineExceeded) {
		e.ethTxHelper = nil
	} else if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
		e.ethTxHelper = nil
	}

	return err
}

// sendTx should be called by all public methods that sends tx to the bridge
func (e *EthHelperWrapper) SendTx(ctx context.Context, handler ethtxhelper.SendTxFunc) (string, error) {
	ethTxHelper, err := e.GetEthHelper()
	if err != nil {
		return "", err
	}

	tx, err := ethTxHelper.SendTx(ctx, e.wallet, bind.TransactOpts{}, true, handler)
	if err != nil {
		return "", e.ProcessError(err)
	}

	// TODO: enable logs bsc.logger.Info("tx has been sent", "tx hash", tx.Hash().String())

	receipt, err := ethTxHelper.WaitForReceipt(ctx, tx.Hash().String(), true)
	if err != nil {
		return "", e.ProcessError(err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return receipt.BlockHash.String(), fmt.Errorf("receipts status not successful: %v", receipt.Status)
	}
	// TODO: enable logs  bsc.logger.Info("tx has been executed", "block", receipt.BlockHash.String(), "tx hash", receipt.TxHash.String())

	return receipt.BlockHash.String(), nil
}
