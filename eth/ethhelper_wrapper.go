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
	wallet      ethtxhelper.IEthTxWallet
	ethTxHelper ethtxhelper.IEthTxHelper
	opts        []ethtxhelper.TxRelayerOption
	lock        sync.Mutex
	logger      hclog.Logger
}

func NewEthHelperWrapper(
	logger hclog.Logger,
	opts ...ethtxhelper.TxRelayerOption,
) *EthHelperWrapper {
	return &EthHelperWrapper{
		opts:   append([]ethtxhelper.TxRelayerOption(nil), opts...),
		logger: logger,
	}
}

func NewEthHelperWrapperWithWallet(
	wallet *ethtxhelper.EthTxWallet, logger hclog.Logger,
	opts ...ethtxhelper.TxRelayerOption,
) *EthHelperWrapper {
	return &EthHelperWrapper{
		wallet: wallet,
		opts:   append([]ethtxhelper.TxRelayerOption(nil), opts...),
		logger: logger,
	}
}

func (e *EthHelperWrapper) GetEthHelper() (ethtxhelper.IEthTxHelper, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	if e.ethTxHelper != nil {
		return e.ethTxHelper, nil
	}

	ethTxHelper, err := ethtxhelper.NewEThTxHelper(e.opts...)
	if err != nil {
		return nil, fmt.Errorf("error while NewEThTxHelper: %w", err)
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
func (e *EthHelperWrapper) SendTx(ctx context.Context, handler ethtxhelper.SendTxFunc) (*types.Receipt, error) {
	ethTxHelper, err := e.GetEthHelper()
	if err != nil {
		return nil, fmt.Errorf("error while GetEthHelper: %w", err)
	}

	tx, err := ethTxHelper.SendTx(ctx, e.wallet, bind.TransactOpts{}, handler)
	if err != nil {
		// tx is not available here to pick hash/gas/gasprice
		return nil, fmt.Errorf("error while SendTx: %w", e.ProcessError(err))
	}

	txHashStr := tx.Hash().String()

	e.logger.Info("tx has been sent", "hash", txHashStr, "gas limit", tx.Gas(), "gas price", tx.GasPrice())

	err = ethTxHelper.WaitForTxPool(ctx, e.wallet, txHashStr)
	if err != nil {
		if !errors.Is(err, ethtxhelper.ErrTxNotIncludedInTxPool) {
			return nil, fmt.Errorf("gas limit = %d, gas price = %s: %w",
				tx.Gas(), tx.GasPrice(), e.ProcessError(err))
		}

		// If the transaction is not included in the transaction pool, we should continue waiting for the receipt
		// This prevents the oracle/batcher from getting stuck due to missing txpool inclusion
		e.logger.Info("tx has not been included in tx pool", "hash", txHashStr, "gas limit", tx.Gas(), "gas price", tx.GasPrice())
	}

	receipt, err := ethTxHelper.WaitForReceipt(ctx, txHashStr, true)
	if err != nil {
		return nil, fmt.Errorf("failed to receive receipt for tx %s, gas limit = %d, gas price = %s: %w",
			txHashStr, tx.Gas(), tx.GasPrice(), e.ProcessError(err))
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return receipt,
			fmt.Errorf("tx receipt status is unsuccessful for %s, gas limit = %d, gas price = %s",
				txHashStr, tx.Gas(), tx.GasPrice())
	}

	e.logger.Info("tx has been included in block", "hash", txHashStr,
		"block", receipt.BlockNumber, "block hash", receipt.BlockHash, "gas used", receipt.GasUsed)

	return receipt, nil
}
