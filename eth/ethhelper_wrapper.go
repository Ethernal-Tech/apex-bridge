package eth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	apexCommon "github.com/Ethernal-Tech/apex-bridge/common"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hashicorp/go-hclog"
)

const (
	defaultReceiptRetriesCnt = 1000
	defaultReceiptWaitTime   = 300 * time.Millisecond
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

	option := ethtxhelper.WithReceiptRetryConfig(
		defaultReceiptRetriesCnt, defaultReceiptWaitTime, func(err error) bool {
			return err == nil || errors.Is(err, ethereum.NotFound)
		})

	ethTxHelper, err := ethtxhelper.NewEThTxHelper(
		append([]ethtxhelper.TxRelayerOption{option}, e.opts...)...)
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
	if errors.Is(err, net.ErrClosed) || apexCommon.IsContextDoneErr(err) {
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

	tx, foundInTxPool, err := e.sendTx(ctx, ethTxHelper, handler)
	if err != nil {
		return nil, fmt.Errorf("error while SendTx: %w", e.ProcessError(err))
	}

	txHashStr := tx.Hash().String()

	e.logger.Info("tx has been sent", "hash", txHashStr,
		"gas limit", tx.Gas(), "gas price", tx.GasPrice(), "foundInTxPool", foundInTxPool)

	// If the transaction is not included in the transaction pool, we should continue waiting for the receipt
	// This prevents the oracle/batcher from getting stuck due to missing txpool inclusion
	if foundInTxPool {
		if err = ethTxHelper.WaitForTxExitTxPool(ctx, e.wallet, txHashStr); err != nil {
			return nil, fmt.Errorf("gas limit = %d, gas price = %s: %w",
				tx.Gas(), tx.GasPrice(), e.ProcessError(err))
		}

		e.logger.Info("tx has exited tx pool",
			"hash", txHashStr, "gas limit", tx.Gas(), "gas price", tx.GasPrice())
	}

	receipt, err := ethTxHelper.WaitForReceipt(ctx, txHashStr)
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

func (e *EthHelperWrapper) sendTx(
	ctx context.Context, ethTxHelper ethtxhelper.IEthTxHelper, handler ethtxhelper.SendTxFunc,
) (*types.Transaction, bool, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	opts, err := ethTxHelper.PrepareSendTx(ctx, e.wallet, bind.TransactOpts{})
	if err != nil {
		return nil, false, err
	}

	tx, err := handler(opts)
	if err != nil {
		return nil, false, err
	}

	foundInTxPool, err := ethTxHelper.WaitForTxEnterTxPool(ctx, e.wallet, tx.Hash().String())

	return tx, foundInTxPool, err
}
