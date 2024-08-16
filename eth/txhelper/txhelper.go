package ethtxhelper

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/big"
	"sync"
	"time"

	apexCommon "github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type SendTxFunc func(*bind.TransactOpts) (*types.Transaction, error)

const (
	defaultGasLimit         = uint64(5_242_880) // 0x500000
	defaultNumRetries       = 1000
	defaultGasFeeMultiplier = 170 // 170%
)

type IEthTxHelper interface {
	GetClient() *ethclient.Client
	GetNonce(ctx context.Context, addr string, pending bool) (uint64, error)
	Deploy(ctx context.Context, wallet IEthTxWallet,
		txOptsParam bind.TransactOpts, abiData abi.ABI, bytecode []byte) (string, string, error)
	WaitForReceipt(ctx context.Context, hash string, skipNotFound bool) (*types.Receipt, error)
	SendTx(ctx context.Context, wallet IEthTxWallet,
		txOpts bind.TransactOpts, sendTxHandler SendTxFunc) (*types.Transaction, error)
	EstimateGas(
		ctx context.Context, from, to common.Address, value *big.Int, gasLimitMultiplier float64,
		bindMetadata *bind.MetaData, method string, args ...interface{},
	) (uint64, uint64, error)
	PopulateTxOpts(ctx context.Context, from common.Address, txOpts *bind.TransactOpts) error
}

type EthTxHelperImpl struct {
	client           *ethclient.Client
	nodeURL          string
	writer           io.Writer
	numRetries       int
	receiptWaitTime  time.Duration
	gasFeeMultiplier uint64
	isDynamic        bool
	zeroGasPrice     bool
	defaultGasLimit  uint64
	chainID          *big.Int
	mutex            sync.Mutex
}

var _ IEthTxHelper = (*EthTxHelperImpl)(nil)

func NewEThTxHelper(opts ...TxRelayerOption) (*EthTxHelperImpl, error) {
	t := &EthTxHelperImpl{
		receiptWaitTime:  50 * time.Millisecond,
		numRetries:       defaultNumRetries,
		gasFeeMultiplier: defaultGasFeeMultiplier,
		zeroGasPrice:     true,
		defaultGasLimit:  defaultGasLimit,
	}
	for _, opt := range opts {
		opt(t)
	}

	if t.client == nil {
		client, err := ethclient.Dial(t.nodeURL)
		if err != nil {
			return nil, err
		}

		t.client = client
	}

	return t, nil
}

func (t *EthTxHelperImpl) GetClient() *ethclient.Client {
	return t.client
}

func (t *EthTxHelperImpl) GetNonce(ctx context.Context, addr string, pending bool) (uint64, error) {
	if pending {
		return t.client.PendingNonceAt(ctx, common.HexToAddress(addr))
	}

	return t.client.NonceAt(ctx, common.HexToAddress(addr), nil)
}

func (t *EthTxHelperImpl) Deploy(
	ctx context.Context, wallet IEthTxWallet, txOptsParam bind.TransactOpts, abiData abi.ABI, bytecode []byte,
) (string, string, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	chainID := t.chainID
	if chainID == nil {
		// chainID retrieval
		retChainID, err := t.client.ChainID(ctx)
		if err != nil {
			return "", "", err
		}

		chainID = retChainID
	}

	// Create contract deployment transaction
	txOptsRes, err := wallet.GetTransactOpts(chainID)
	if err != nil {
		return "", "", err
	}

	copyTxOpts(txOptsRes, &txOptsParam)

	if err := t.PopulateTxOpts(ctx, wallet.GetAddress(), txOptsRes); err != nil {
		return "", "", err
	}

	// Deploy the contract
	contractAddress, tx, _, err := bind.DeployContract(txOptsRes, abiData, bytecode, t.client)
	if err != nil {
		return "", "", err
	}

	return contractAddress.String(), tx.Hash().String(), nil
}

func (t *EthTxHelperImpl) WaitForReceipt(
	ctx context.Context, hash string, skipNotFound bool,
) (*types.Receipt, error) {
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

func (t *EthTxHelperImpl) SendTx(
	ctx context.Context, wallet IEthTxWallet, txOptsParam bind.TransactOpts, sendTxHandler SendTxFunc,
) (*types.Transaction, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	chainID := t.chainID
	if chainID == nil {
		// chainID retrieval
		retChainID, err := t.client.ChainID(ctx)
		if err != nil {
			return nil, err
		}

		chainID = retChainID
	}

	txOptsRes, err := wallet.GetTransactOpts(chainID)
	if err != nil {
		return nil, err
	}

	copyTxOpts(txOptsRes, &txOptsParam)

	if err := t.PopulateTxOpts(ctx, wallet.GetAddress(), txOptsRes); err != nil {
		return nil, err
	}

	return sendTxHandler(txOptsRes)
}

func (t *EthTxHelperImpl) EstimateGas(
	ctx context.Context, from, to common.Address, value *big.Int, gasLimitMultiplier float64,
	bindMetadata *bind.MetaData, method string, args ...interface{},
) (uint64, uint64, error) {
	parsed, err := bindMetadata.GetAbi()
	if err != nil {
		return 0, 0, err
	}

	input, err := parsed.Pack(method, args...)
	if err != nil {
		return 0, 0, err
	}

	estimatedGas, err := t.GetClient().EstimateGas(ctx, ethereum.CallMsg{
		From:  from,
		To:    &to,
		Value: value,
		Data:  input,
	})
	if err != nil {
		return 0, 0, err
	}

	return uint64(float64(estimatedGas) * gasLimitMultiplier), estimatedGas, nil
}

func (t *EthTxHelperImpl) PopulateTxOpts(
	ctx context.Context, from common.Address, txOpts *bind.TransactOpts,
) error {
	txOpts.Context = ctx
	txOpts.From = from

	// Nonce retrieval
	if txOpts.Nonce == nil {
		nonce, err := t.client.PendingNonceAt(ctx, txOpts.From)
		if err != nil {
			return err
		}

		txOpts.Nonce = new(big.Int).SetUint64(nonce)
	}

	if txOpts.GasLimit == 0 {
		txOpts.GasLimit = t.defaultGasLimit
	}

	// Gas price
	if !t.isDynamic {
		if txOpts.GasPrice == nil {
			if t.zeroGasPrice {
				txOpts.GasPrice = big.NewInt(0)
			} else {
				gasPrice, err := t.client.SuggestGasPrice(ctx)
				if err != nil {
					return err
				}

				txOpts.GasPrice = apexCommon.MulPercentage(gasPrice, t.gasFeeMultiplier)
			}
		}
	} else if txOpts.GasFeeCap == nil || txOpts.GasTipCap == nil {
		gasTipCap, err := t.client.SuggestGasTipCap(ctx)
		if err != nil {
			return err
		}

		txOpts.GasTipCap = apexCommon.MulPercentage(gasTipCap, t.gasFeeMultiplier)

		hs, err := t.client.FeeHistory(ctx, 1, nil, nil)
		if err != nil {
			return err
		}

		gasFeeCap := hs.BaseFee[len(hs.BaseFee)-1]
		gasFeeCap = gasFeeCap.Add(gasFeeCap, gasTipCap)

		txOpts.GasFeeCap = apexCommon.MulPercentage(gasFeeCap, t.gasFeeMultiplier)
	}

	return nil
}

type TxRelayerOption func(*EthTxHelperImpl)

func WithDynamicTx(value bool) TxRelayerOption {
	return func(t *EthTxHelperImpl) {
		t.isDynamic = value
	}
}

func WithClient(client *ethclient.Client) TxRelayerOption {
	return func(t *EthTxHelperImpl) {
		t.client = client
	}
}

func WithNodeURL(nodeURL string) TxRelayerOption {
	return func(t *EthTxHelperImpl) {
		t.nodeURL = nodeURL
	}
}

func WithReceiptWaitTime(receiptWaitTime time.Duration) TxRelayerOption {
	return func(t *EthTxHelperImpl) {
		t.receiptWaitTime = receiptWaitTime
	}
}

func WithWriter(writer io.Writer) TxRelayerOption {
	return func(t *EthTxHelperImpl) {
		t.writer = writer
	}
}

// WithNumRetries sets the maximum number of eth_getTransactionReceipt retries
// before considering the transaction sending as timed out. Set to -1 to disable
// waitForReceipt and not wait for the transaction receipt
func WithNumRetries(numRetries int) TxRelayerOption {
	return func(t *EthTxHelperImpl) {
		t.numRetries = numRetries
	}
}

func WithGasFeeMultiplier(gasFeeMultiplier uint64) TxRelayerOption {
	return func(t *EthTxHelperImpl) {
		t.gasFeeMultiplier = gasFeeMultiplier
	}
}

func WithZeroGasPrice(zeroGasPrice bool) TxRelayerOption {
	return func(t *EthTxHelperImpl) {
		t.zeroGasPrice = zeroGasPrice
	}
}

func WithDefaultGasLimit(gasLimit uint64) TxRelayerOption {
	return func(t *EthTxHelperImpl) {
		t.defaultGasLimit = gasLimit
	}
}

func WithChainID(chainID *big.Int) TxRelayerOption {
	return func(t *EthTxHelperImpl) {
		t.chainID = chainID
	}
}

func copyTxOpts(dst, src *bind.TransactOpts) {
	dst.NoSend = src.NoSend
	dst.GasPrice = src.GasPrice
	dst.GasFeeCap = src.GasFeeCap
	dst.GasTipCap = src.GasTipCap
	dst.GasLimit = src.GasLimit
	dst.Nonce = src.Nonce
	dst.Value = src.Value
}
