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
	"github.com/hashicorp/go-hclog"
)

type EthHelperWrapper struct {
	nodeURL     string
	wallet      ethtxhelper.IEthTxWallet
	ethTxHelper ethtxhelper.IEthTxHelper
	isDynamic   bool
	logger      hclog.Logger
	lock        sync.Mutex
}

func NewEthHelperWrapper(
	nodeURL string, isDynamic bool, logger hclog.Logger,
) *EthHelperWrapper {
	return &EthHelperWrapper{
		nodeURL:   nodeURL,
		isDynamic: isDynamic,
		logger:    logger,
	}
}

func NewEthHelperWrapperWithWallet(
	nodeURL string, wallet *ethtxhelper.EthTxWallet, isDynamic bool, logger hclog.Logger,
) (*EthHelperWrapper, error) {
	return &EthHelperWrapper{
		nodeURL:   nodeURL,
		wallet:    wallet,
		isDynamic: isDynamic,
		logger:    logger,
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
			make([]ethtxhelper.TxRelayerOption, 0, len(opts)+2),
			ethtxhelper.WithNodeURL(e.nodeURL),
			ethtxhelper.WithDynamicTx(e.isDynamic),
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

	tx, err := ethTxHelper.SendTx(ctx, e.wallet, bind.TransactOpts{}, handler)
	if err != nil {
		return "", e.ProcessError(err)
	}

	e.logger.Info("tx has been sent", "tx hash", tx.Hash().String())

	receipt, err := ethTxHelper.WaitForReceipt(ctx, tx.Hash().String(), true)
	if err != nil {
		return "", e.ProcessError(err)
	}

	e.logger.Info("tx has been included in block", "tx hash", tx.Hash().String(),
		"status", receipt.Status, "block", receipt.BlockNumber, "block hash", receipt.BlockHash.String())

	if receipt.Status != types.ReceiptStatusSuccessful {
		return receipt.BlockHash.String(), fmt.Errorf("receipts status not successful: %v", receipt.Status)
	}

	return receipt.BlockHash.String(), nil
}
