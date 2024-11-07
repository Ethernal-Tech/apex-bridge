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
type NonceRetrieveFunc func(ctx context.Context, client *ethclient.Client, addr common.Address) (uint64, error)
type NonceUpdateFunc func(addr common.Address, value uint64)

const (
	defaultGasLimit          = uint64(5_242_880) // 0x500000
	defaultGasFeeMultiplier  = 170               // 170%
	defaultReceiptRetriesCnt = 1000
	defaultReceiptWaitTime   = 300 * time.Millisecond
)

var (
	errGasPriceSetWhileDynamicTx   = errors.New("gasPrice cannot be set while dynamicTx is true")
	errGasCapsSetWhileNotDynamicTx = errors.New("gasFeeCap and gasTipCap cannot be set while dynamicTx is false")
)

type IEthTxHelper interface {
	GetClient() *ethclient.Client
	GetNonce(ctx context.Context, addr string, pending bool) (uint64, error)
	Deploy(ctx context.Context, wallet IEthTxWallet, txOptsParam bind.TransactOpts,
		abiData abi.ABI, bytecode []byte, params ...interface{}) (string, string, error)
	WaitForReceipt(ctx context.Context, hash string, skipNotFound bool) (*types.Receipt, error)
	SendTx(ctx context.Context, wallet IEthTxWallet,
		txOpts bind.TransactOpts, sendTxHandler SendTxFunc) (*types.Transaction, error)
	EstimateGas(
		ctx context.Context, from, to common.Address, value *big.Int, gasLimitMultiplier float64,
		abi *abi.ABI, method string, args ...interface{},
	) (uint64, uint64, error)
	PopulateTxOpts(ctx context.Context, from common.Address, txOpts *bind.TransactOpts) error
}

type EthTxHelperImpl struct {
	client            *ethclient.Client
	nodeURL           string
	writer            io.Writer
	receiptRetriesCnt int
	receiptWaitTime   time.Duration
	gasFeeMultiplier  uint64
	isDynamic         bool
	zeroGasPrice      bool
	defaultGasLimit   uint64
	chainID           *big.Int
	initFn            func(*EthTxHelperImpl) error
	nonceRetrieveFn   NonceRetrieveFunc
	nonceUpdateFn     NonceUpdateFunc
	mutex             sync.Mutex
}

var _ IEthTxHelper = (*EthTxHelperImpl)(nil)

func NewEThTxHelper(opts ...TxRelayerOption) (*EthTxHelperImpl, error) {
	t := &EthTxHelperImpl{
		receiptWaitTime:   defaultReceiptWaitTime,
		receiptRetriesCnt: defaultReceiptRetriesCnt,
		gasFeeMultiplier:  defaultGasFeeMultiplier,
		zeroGasPrice:      true,
		defaultGasLimit:   defaultGasLimit,
		nonceRetrieveFn: func(ctx context.Context, client *ethclient.Client, addr common.Address) (uint64, error) {
			return client.PendingNonceAt(ctx, addr)
		},
		nonceUpdateFn: func(addr common.Address, value uint64) {},
		initFn: func(t *EthTxHelperImpl) error {
			if t.client == nil {
				client, err := ethclient.Dial(t.nodeURL)
				if err != nil {
					return err
				}

				t.client = client
			}

			return nil
		},
	}
	for _, opt := range opts {
		opt(t)
	}

	if err := t.initFn(t); err != nil {
		return nil, err
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
	ctx context.Context, wallet IEthTxWallet, txOptsParam bind.TransactOpts,
	abiData abi.ABI, bytecode []byte, params ...interface{},
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
	contractAddress, tx, _, err := bind.DeployContract(txOptsRes, abiData, bytecode, t.client, params...)
	if err != nil {
		return "", "", err
	}

	t.nonceUpdateFn(wallet.GetAddress(), txOptsRes.Nonce.Uint64())

	return contractAddress.String(), tx.Hash().String(), nil
}

func (t *EthTxHelperImpl) WaitForReceipt(
	ctx context.Context, hash string, skipNotFound bool,
) (*types.Receipt, error) {
	for count := 0; count < t.receiptRetriesCnt; count++ {
		receipt, err := t.client.TransactionReceipt(ctx, common.HexToHash(hash))
		if err != nil {
			if !skipNotFound && errors.Is(err, ethereum.NotFound) {
				return nil, fmt.Errorf("transaction %s not found", hash)
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

	tx, err := sendTxHandler(txOptsRes)
	if err != nil {
		return nil, err
	}

	t.nonceUpdateFn(wallet.GetAddress(), txOptsRes.Nonce.Uint64())

	return tx, nil
}

func (t *EthTxHelperImpl) EstimateGas(
	ctx context.Context, from, to common.Address, value *big.Int, gasLimitMultiplier float64,
	abi *abi.ABI, method string, args ...interface{},
) (uint64, uint64, error) {
	input, err := abi.Pack(method, args...)
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
		nonce, err := t.nonceRetrieveFn(ctx, t.client, from)
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
		if txOpts.GasFeeCap != nil || txOpts.GasTipCap != nil {
			return errGasCapsSetWhileNotDynamicTx
		}

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
		if txOpts.GasPrice != nil {
			return errGasPriceSetWhileDynamicTx
		}

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

// WithReceiptRetriesCnt sets the maximum number of eth_getTransactionReceipt retries
// before considering the transaction sending as timed out.
func WithReceiptRetriesCnt(receiptRetriesCnt int) TxRelayerOption {
	return func(t *EthTxHelperImpl) {
		t.receiptRetriesCnt = receiptRetriesCnt
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

func WithInitFn(fn func(*EthTxHelperImpl) error) TxRelayerOption {
	return func(t *EthTxHelperImpl) {
		t.initFn = fn
	}
}

func WithInitClientAndChainIDFn(ctx context.Context) TxRelayerOption {
	return func(t *EthTxHelperImpl) {
		t.initFn = func(ethi *EthTxHelperImpl) error {
			client, err := ethclient.DialContext(ctx, t.nodeURL)
			if err != nil {
				return err
			}

			chainID, err := client.ChainID(ctx)
			if err != nil {
				return err
			}

			t.client = client
			t.chainID = chainID

			return nil
		}
	}
}

func WithNonceRetrieveFunc(fn NonceRetrieveFunc) TxRelayerOption {
	return func(t *EthTxHelperImpl) {
		t.nonceRetrieveFn = fn
	}
}

func WithNonceRetrieveCounterFunc() TxRelayerOption {
	return func(t *EthTxHelperImpl) {
		counterMap := map[common.Address]uint64{}

		t.nonceRetrieveFn = func(
			ctx context.Context, client *ethclient.Client, addr common.Address,
		) (result uint64, err error) {
			if value, exists := counterMap[addr]; !exists {
				result, err = client.PendingNonceAt(ctx, addr)
				if err != nil {
					return 0, err
				}
			} else {
				result = value + 1
			}

			return result, nil
		}
		t.nonceUpdateFn = func(addr common.Address, nonce uint64) {
			counterMap[addr] = nonce
		}
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

func WaitForTransactions(
	ctx context.Context, txHelper IEthTxHelper, txHashes ...string,
) ([]*types.Receipt, error) {
	receipts := make([]*types.Receipt, len(txHashes))
	errs := make([]error, len(txHashes))
	sg := sync.WaitGroup{}

	for i, txHash := range txHashes {
		sg.Add(1)

		go func(idx int, txHash string) {
			defer sg.Done()

			rec, err := txHelper.WaitForReceipt(ctx, txHash, true)
			if err == nil && rec.Status != types.ReceiptStatusSuccessful {
				err = fmt.Errorf("receipt status for %s is unsuccessful", txHash)
			}

			receipts[idx], errs[idx] = rec, err
		}(i, txHash)
	}

	sg.Wait()

	return receipts, errors.Join(errs...)
}
