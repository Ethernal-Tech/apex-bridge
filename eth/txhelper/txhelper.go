package ethtxhelper

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type SendTxFunc func(*bind.TransactOpts) (*types.Transaction, error)

const (
	defaultGasPrice         = 1879048192 // 0x70000000
	DefaultGasLimit         = 5242880    // 0x500000
	defaultNumRetries       = 1000
	defaultGasFeeMultiplier = 1.6
)

type IEthTxHelper interface {
	GetClient() bind.ContractBackend
	GetNonce(ctx context.Context, addr string, pending bool) (uint64, error)
	Deploy(ctx context.Context, nonce *big.Int, gasLimit uint64, isDynamic bool,
		abiData abi.ABI, bytecode []byte, wallet *EthTxWallet) (string, string, error)
	WaitForReceipt(ctx context.Context, hash string, skipNotFound bool) (*types.Receipt, error)
	SendTx(ctx context.Context,
		wallet *EthTxWallet, txOpts bind.TransactOpts, isDynamic bool, sendTxHandler SendTxFunc) (*types.Transaction, error)
	PopulateTxOpts(ctx context.Context, from string, isDynamic bool, txOpts *bind.TransactOpts) error
}

type EThTxHelper struct {
	client           *ethclient.Client
	nodeUrl          string
	writer           io.Writer
	numRetries       int
	receiptWaitTime  time.Duration
	gasFeeMultiplier float64
}

var _ IEthTxHelper = (*EThTxHelper)(nil)

func NewEThTxHelper(opts ...TxRelayerOption) (*EThTxHelper, error) {
	t := &EThTxHelper{
		receiptWaitTime:  50 * time.Millisecond,
		numRetries:       defaultNumRetries,
		gasFeeMultiplier: defaultGasFeeMultiplier,
	}
	for _, opt := range opts {
		opt(t)
	}

	if t.client == nil {
		client, err := ethclient.Dial(t.nodeUrl)
		if err != nil {
			return nil, err
		}

		t.client = client
	}

	return t, nil
}

func (t EThTxHelper) GetClient() bind.ContractBackend {
	return t.client
}

func (t EThTxHelper) GetNonce(ctx context.Context, addr string, pending bool) (uint64, error) {
	if pending {
		return t.client.PendingNonceAt(ctx, common.HexToAddress(addr))
	}

	return t.client.NonceAt(ctx, common.HexToAddress(addr), nil)
}

func (t EThTxHelper) Deploy(ctx context.Context, nonce *big.Int, gasLimit uint64, isDynamic bool,
	abiData abi.ABI, bytecode []byte, wallet *EthTxWallet) (string, string, error) {
	// chainID retrieval
	chainID, err := t.client.ChainID(ctx)
	if err != nil {
		return "", "", err
	}

	// Create contract deployment transaction
	auth, err := wallet.GetTransactOpts(chainID)
	if err != nil {
		return "", "", err
	}

	auth.Nonce = nonce
	auth.GasLimit = gasLimit

	if err := t.PopulateTxOpts(ctx, wallet.GetAddressHex(), isDynamic, auth); err != nil {
		return "", "", err
	}

	// Deploy the contract
	contractAddress, tx, _, err := bind.DeployContract(auth, abiData, bytecode, t.client)
	if err != nil {
		return "", "", err
	}

	return contractAddress.String(), tx.Hash().String(), nil
}

func (t EThTxHelper) WaitForReceipt(ctx context.Context, hash string, skipNotFound bool) (*types.Receipt, error) {
	for count := 0; count < t.numRetries; count++ {
		receipt, err := t.client.TransactionReceipt(ctx, common.HexToHash(hash))
		if err != nil {
			if !skipNotFound && errors.Is(err, ethereum.NotFound) {
				return nil, err
			}
		} else if receipt != nil {
			return receipt, nil
		}

		select {
		case <-time.After(t.receiptWaitTime):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("timeout while waiting for transaction %s to be processed", hash)
}

func (t EThTxHelper) SendTx(ctx context.Context, wallet *EthTxWallet, txOptsParam bind.TransactOpts,
	isDynamic bool, sendTxHandler SendTxFunc) (*types.Transaction, error) {
	// chainID retrieval
	chainID, err := t.client.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	txOptsRes, err := wallet.GetTransactOpts(chainID)
	if err != nil {
		return nil, err
	}

	txOptsRes.NoSend = txOptsParam.NoSend
	txOptsRes.GasPrice = txOptsParam.GasPrice
	txOptsRes.GasFeeCap = txOptsParam.GasFeeCap
	txOptsRes.GasTipCap = txOptsParam.GasTipCap
	txOptsRes.GasLimit = txOptsParam.GasLimit
	txOptsRes.Nonce = txOptsParam.Nonce
	txOptsRes.Value = txOptsParam.Value

	if err := t.PopulateTxOpts(ctx, wallet.GetAddressHex(), isDynamic, txOptsRes); err != nil {
		return nil, err
	}

	// first call sendTx with noSend (that will return tx but not send it on the node)
	if txOptsRes.GasLimit == 0 {
		txOptsRes.NoSend = true

		tx, err := sendTxHandler(txOptsRes)
		if err != nil {
			return nil, err
		}

		// estimate gas
		gas, err := t.client.EstimateGas(ctx, ethereum.CallMsg{
			From:     wallet.GetAddress(),
			To:       tx.To(),
			Gas:      tx.Gas(),
			GasPrice: tx.GasPrice(),
			Data:     tx.Data(),
			Value:    tx.Value(),
		})
		if err != nil {
			return nil, err
		}

		txOptsRes.GasLimit = gas
		txOptsRes.NoSend = false
	}

	return sendTxHandler(txOptsRes)
}

func (t EThTxHelper) PopulateTxOpts(ctx context.Context, from string, isDynamic bool, txOpts *bind.TransactOpts) error {
	txOpts.Context = ctx
	txOpts.From = common.HexToAddress(from)

	// Nonce retrieval
	if txOpts.Nonce == nil {
		nonce, err := t.client.PendingNonceAt(ctx, txOpts.From)
		if err != nil {
			return err
		}

		txOpts.Nonce = new(big.Int).SetUint64(nonce)
	}

	// Gas price
	if !isDynamic {
		if txOpts.GasPrice == nil {
			gasPrice, err := t.client.SuggestGasPrice(ctx)
			if err != nil {
				return err
			}

			txOpts.GasPrice = new(big.Int).SetUint64(uint64(float64(gasPrice.Uint64()) * t.gasFeeMultiplier))
		}
	} else if txOpts.GasFeeCap == nil || txOpts.GasTipCap == nil {
		gasTipCap, err := t.client.SuggestGasTipCap(ctx)
		if err != nil {
			return err
		}

		txOpts.GasTipCap = new(big.Int).SetUint64(uint64(float64(gasTipCap.Uint64()) * t.gasFeeMultiplier))

		hs, err := t.client.FeeHistory(ctx, 1, nil, nil)
		if err != nil {
			return err
		}

		gasFeeCap := hs.BaseFee[len(hs.BaseFee)-1]
		gasFeeCap = gasFeeCap.Add(gasFeeCap, txOpts.GasTipCap)

		txOpts.GasFeeCap = new(big.Int).SetUint64(uint64(float64(gasFeeCap.Uint64()) * t.gasFeeMultiplier))
	}

	return nil
}

type TxRelayerOption func(*EThTxHelper)

func WithClient(client *ethclient.Client) TxRelayerOption {
	return func(t *EThTxHelper) {
		t.client = client
	}
}

func WithNodeUrl(nodeUrl string) TxRelayerOption {
	return func(t *EThTxHelper) {
		t.nodeUrl = nodeUrl
	}
}

func WithReceiptWaitTime(receiptWaitTime time.Duration) TxRelayerOption {
	return func(t *EThTxHelper) {
		t.receiptWaitTime = receiptWaitTime
	}
}

func WithWriter(writer io.Writer) TxRelayerOption {
	return func(t *EThTxHelper) {
		t.writer = writer
	}
}

// WithNumRetries sets the maximum number of eth_getTransactionReceipt retries
// before considering the transaction sending as timed out. Set to -1 to disable
// waitForReceipt and not wait for the transaction receipt
func WithNumRetries(numRetries int) TxRelayerOption {
	return func(t *EThTxHelper) {
		t.numRetries = numRetries
	}
}

func WithGasFeeMultiplier(gasFeeMultiplier float64) TxRelayerOption {
	return func(t *EThTxHelper) {
		t.gasFeeMultiplier = gasFeeMultiplier
	}
}
