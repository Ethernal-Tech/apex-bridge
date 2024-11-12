package ethtxhelper

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type NonceStrategyType int

const (
	NonceNodePendingStrategy NonceStrategyType = iota
	NonceInMemoryStrategy
	NonceCombinedStrategy
)

type NonceStrategy interface {
	GetNextNonce(ctx context.Context, client *ethclient.Client, addr common.Address) (uint64, error)
	UpdateNonce(addr common.Address, value uint64, success bool)
}

func NonceStrategyFactory(strategy NonceStrategyType) NonceStrategy {
	switch strategy {
	case NonceInMemoryStrategy:
		return &nonceInMemoryStrategyImpl{
			lastNonceMap: map[common.Address]uint64{},
		}
	case NonceCombinedStrategy:
		return &nonceCombinedStrategyImpl{
			lastNonceMap: map[common.Address]uint64{},
		}
	default:
		return &nonceNodePendingStrategyImpl{}
	}
}

type nonceNodePendingStrategyImpl struct{}

func (a *nonceNodePendingStrategyImpl) GetNextNonce(
	ctx context.Context, client *ethclient.Client, addr common.Address,
) (uint64, error) {
	return client.PendingNonceAt(ctx, addr)
}

func (a *nonceNodePendingStrategyImpl) UpdateNonce(addr common.Address, value uint64, success bool) {
}

type nonceInMemoryStrategyImpl struct {
	lastNonceMap map[common.Address]uint64
}

func (a *nonceInMemoryStrategyImpl) GetNextNonce(
	ctx context.Context, client *ethclient.Client, addr common.Address,
) (nextNonce uint64, err error) {
	if value, exists := a.lastNonceMap[addr]; !exists {
		nextNonce, err = client.PendingNonceAt(ctx, addr)
		if err != nil {
			return 0, fmt.Errorf("error while getting next nonce: %w", err)
		}
	} else {
		nextNonce = value + 1
	}

	return nextNonce, nil
}

func (a *nonceInMemoryStrategyImpl) UpdateNonce(addr common.Address, value uint64, success bool) {
	if success {
		a.lastNonceMap[addr] = value
	} else {
		delete(a.lastNonceMap, addr)
	}
}

type nonceCombinedStrategyImpl struct {
	lastNonceMap map[common.Address]uint64
}

func (a *nonceCombinedStrategyImpl) GetNextNonce(
	ctx context.Context, client *ethclient.Client, addr common.Address,
) (nextNonce uint64, err error) {
	nextNonce, err = client.PendingNonceAt(ctx, addr)
	if err != nil {
		return 0, fmt.Errorf("error while PendingNonceAt: %w", err)
	}
	// if pending txpool nonce is less than saved noce => next nonce should be taken
	// from previous value + 1
	if prevValue, exists := a.lastNonceMap[addr]; exists && prevValue >= nextNonce {
		nextNonce = prevValue + 1
	}

	return nextNonce, nil
}

func (a *nonceCombinedStrategyImpl) UpdateNonce(addr common.Address, value uint64, success bool) {
	if success {
		a.lastNonceMap[addr] = value
	}
}
