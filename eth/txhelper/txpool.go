package ethtxhelper

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

type TxPoolContent struct {
	Pending map[uint64]*types.Transaction `json:"pending"`
	Queued  map[uint64]*types.Transaction `json:"queued"`
}

func (txPoolContent TxPoolContent) IsTxInTxPoolContent(txHash common.Hash) bool {
	for _, tx := range txPoolContent.Pending {
		if tx.Hash() == txHash {
			return true
		}
	}

	for _, tx := range txPoolContent.Queued {
		if tx.Hash() == txHash {
			return true
		}
	}

	return false
}

func GetTxPoolStateForAddr(
	ctx context.Context, rpcClient *rpc.Client, addr common.Address,
) (result TxPoolContent, err error) {
	err = rpcClient.CallContext(ctx, &result, "txpool_contentFrom", addr)

	return result, err
}

func IsTxInTxPool(
	ctx context.Context, rpcClient *rpc.Client, addr common.Address, txHashStr string,
) (bool, error) {
	txPoolContent, err := GetTxPoolStateForAddr(ctx, rpcClient, addr)
	if err != nil {
		return false, err
	}

	return txPoolContent.IsTxInTxPoolContent(common.HexToHash(txHashStr)), nil
}
