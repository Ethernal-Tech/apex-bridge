package eth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
)

type EthHelperWrapper struct {
	nodeURL     string
	wallet      ethtxhelper.IEthTxWallet
	ethTxHelper ethtxhelper.IEthTxHelper
	lock        sync.Mutex
}

func NewEthHelperWrapper(nodeURL string) *EthHelperWrapper {
	return &EthHelperWrapper{
		nodeURL: nodeURL,
	}
}

func NewEthHelperWrapperWithWallet(nodeURL string, wallet *ethtxhelper.EthTxWallet) (*EthHelperWrapper, error) {
	return &EthHelperWrapper{
		nodeURL: nodeURL,
		wallet:  wallet,
	}, nil
}

func (e *EthHelperWrapper) GetEthHelper(opts ...ethtxhelper.TxRelayerOption) (ethtxhelper.IEthTxHelper, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	if e.ethTxHelper != nil {
		return e.ethTxHelper, nil
	}

	finalOpts := append(
		append(
			make([]ethtxhelper.TxRelayerOption, 0, len(opts)+1),
			ethtxhelper.WithNodeURL(e.nodeURL),
		), opts...)

	ethTxHelper, err := ethtxhelper.NewEThTxHelper(finalOpts...)
	if err != nil {
		return nil, err
	}

	e.ethTxHelper = ethTxHelper

	return ethTxHelper, nil
}

func (e *EthHelperWrapper) ProcessError(err error) error {
	var netErr net.Error

	//nolint:godox
	// TODO: verify if these errors are the only ones we need to handle
	if errors.Is(err, net.ErrClosed) || errors.Is(err, context.DeadlineExceeded) {
		e.lock.Lock()
		e.ethTxHelper = nil
		e.lock.Unlock()
	} else if ok := errors.As(err, &netErr); ok && netErr.Timeout() {
		e.lock.Lock()
		e.ethTxHelper = nil
		e.lock.Unlock()
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

	//nolint:godox
	// TODO: enable logs bsc.logger.Info("tx has been sent", "tx hash", tx.Hash().String())

	receipt, err := ethTxHelper.WaitForReceipt(ctx, tx.Hash().String(), true)
	if err != nil {
		return "", e.ProcessError(err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return receipt.BlockHash.String(), fmt.Errorf("receipts status not successful: %v", receipt.Status)
	}

	//nolint:godox,lll
	// TODO: enable logs  bsc.logger.Info("tx has been executed", "block", receipt.BlockHash.String(), "tx hash", receipt.TxHash.String())

	return receipt.BlockHash.String(), nil
}
