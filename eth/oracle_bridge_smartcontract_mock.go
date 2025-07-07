package eth

import (
	"context"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
)

type OracleBridgeSmartContractMock struct {
	mock.Mock
}

// GetBatchStatusAndTransactions implements IOracleBridgeSmartContract.
func (m *OracleBridgeSmartContractMock) GetBatchStatusAndTransactions(
	ctx context.Context, chainID string, batchID uint64,
) (uint8, []contractbinding.IBridgeStructsTxDataInfo, error) {
	args := m.Called(ctx, chainID, batchID)
	status, _ := args.Get(0).(uint8)
	txs, _ := args.Get(1).([]contractbinding.IBridgeStructsTxDataInfo)

	return status, txs, args.Error(2)
}

// GetLastObservedBlock implements IOracleBridgeSmartContract.
func (m *OracleBridgeSmartContractMock) GetLastObservedBlock(
	ctx context.Context, sourceChain string,
) (CardanoBlock, error) {
	args := m.Called()
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).(CardanoBlock)

		return arg0, args.Error(1)
	}

	return CardanoBlock{}, args.Error(1)
}

// GetRawTransactionFromLastBatch implements IOracleBridgeSmartContract.
func (m *OracleBridgeSmartContractMock) GetRawTransactionFromLastBatch(
	ctx context.Context, chainID string,
) ([]byte, error) {
	args := m.Called()
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).([]byte)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}

// SubmitClaims implements IOracleBridgeSmartContract.
func (m *OracleBridgeSmartContractMock) SubmitClaims(
	ctx context.Context, claims Claims, submitOpts *SubmitOpts) (*types.Receipt, error) {
	args := m.Called()
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).(*types.Receipt)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}

// SubmitLastObservedBlocks implements IOracleBridgeSmartContract.
func (m *OracleBridgeSmartContractMock) SubmitLastObservedBlocks(
	ctx context.Context, chainID string, blocks []CardanoBlock,
) error {
	args := m.Called()

	return args.Error(0)
}

var _ IOracleBridgeSmartContract = (*OracleBridgeSmartContractMock)(nil)
