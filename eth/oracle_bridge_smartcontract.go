package eth

import (
	"context"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const submitClaimsGasLimit = uint64(10_000_000)

type CardanoBlock = contractbinding.IBridgeStructsCardanoBlock
type Claims = contractbinding.IBridgeStructsValidatorClaims
type TxDataInfo = contractbinding.IBridgeStructsTxDataInfo

type SubmitOpts struct {
	GasLimitMultiplier float32
}

type IOracleBridgeSmartContract interface {
	GetLastObservedBlock(ctx context.Context, sourceChain string) (CardanoBlock, error)
	GetRawTransactionFromLastBatch(ctx context.Context, chainID string) ([]byte, error)
	SubmitClaims(ctx context.Context, claims Claims, submitOpts *SubmitOpts) (*types.Receipt, error)
	SubmitLastObservedBlocks(ctx context.Context, chainID string, blocks []CardanoBlock) error
	GetBatchStatusAndTransactions(ctx context.Context, chainID string, batchID uint64) (uint8, []TxDataInfo, error)
}

type OracleBridgeSmartContractImpl struct {
	smartContractAddress ethcommon.Address
	ethHelper            *EthHelperWrapper
	chainIDConverter     *common.ChainIDConverter
}

var _ IOracleBridgeSmartContract = (*OracleBridgeSmartContractImpl)(nil)

func NewOracleBridgeSmartContract(
	smartContractAddress string, ethHelper *EthHelperWrapper, chainIDConverter *common.ChainIDConverter,
) *OracleBridgeSmartContractImpl {
	return &OracleBridgeSmartContractImpl{
		smartContractAddress: ethcommon.HexToAddress(smartContractAddress),
		ethHelper:            ethHelper,
		chainIDConverter:     chainIDConverter,
	}
}

func (bsc *OracleBridgeSmartContractImpl) GetLastObservedBlock(
	ctx context.Context, sourceChain string,
) (CardanoBlock, error) {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return CardanoBlock{}, fmt.Errorf("error while GetEthHelper: %w", err)
	}

	contract, err := contractbinding.NewBridgeContract(
		bsc.smartContractAddress,
		ethTxHelper.GetClient())
	if err != nil {
		return CardanoBlock{}, fmt.Errorf("error while NewBridgeContract: %w", bsc.ethHelper.ProcessError(err))
	}

	result, err := contract.GetLastObservedBlock(&bind.CallOpts{
		Context: ctx,
	}, bsc.chainIDConverter.ToChainIDNum(sourceChain))
	if err != nil {
		return CardanoBlock{}, fmt.Errorf("error while GetLastObservedBlock: %w", bsc.ethHelper.ProcessError(err))
	}

	return result, nil
}

func (bsc *OracleBridgeSmartContractImpl) GetRawTransactionFromLastBatch(
	ctx context.Context, chainID string,
) ([]byte, error) {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return nil, fmt.Errorf("error while GetEthHelper: %w", err)
	}

	contract, err := contractbinding.NewBridgeContract(
		bsc.smartContractAddress,
		ethTxHelper.GetClient())
	if err != nil {
		return nil, fmt.Errorf("error while NewBridgeContract: %w", bsc.ethHelper.ProcessError(err))
	}

	result, err := contract.GetRawTransactionFromLastBatch(&bind.CallOpts{
		Context: ctx,
	}, bsc.chainIDConverter.ToChainIDNum(chainID))
	if err != nil {
		return nil, fmt.Errorf("error while GetRawTransactionFromLastBatch: %w", bsc.ethHelper.ProcessError(err))
	}

	return result, nil
}

func (bsc *OracleBridgeSmartContractImpl) SubmitClaims(
	ctx context.Context, claims Claims, submitOpts *SubmitOpts,
) (*types.Receipt, error) {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return nil, fmt.Errorf("error while GetEthHelper: %w", err)
	}

	contract, err := contractbinding.NewBridgeContract(
		bsc.smartContractAddress,
		ethTxHelper.GetClient())
	if err != nil {
		return nil, fmt.Errorf("error while NewBridgeContract: %w", bsc.ethHelper.ProcessError(err))
	}

	receipt, err := bsc.ethHelper.SendTx(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		opts.GasLimit = submitClaimsGasLimit
		if submitOpts != nil && submitOpts.GasLimitMultiplier != 0 {
			opts.GasLimit = uint64(float32(opts.GasLimit) * submitOpts.GasLimitMultiplier)
		}

		return contract.SubmitClaims(opts, claims)
	})
	if err != nil {
		return nil, fmt.Errorf("error while SendTx SubmitClaims: %w", bsc.ethHelper.ProcessError(err))
	}

	return receipt, nil
}

func (bsc *OracleBridgeSmartContractImpl) SubmitLastObservedBlocks(
	ctx context.Context, chainID string, blocks []CardanoBlock,
) error {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return fmt.Errorf("error while GetEthHelper: %w", err)
	}

	contract, err := contractbinding.NewBridgeContract(
		bsc.smartContractAddress,
		ethTxHelper.GetClient())
	if err != nil {
		return fmt.Errorf("error while NewBridgeContract: %w", bsc.ethHelper.ProcessError(err))
	}

	_, err = bsc.ethHelper.SendTx(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return contract.SubmitLastObservedBlocks(opts, bsc.chainIDConverter.ToChainIDNum(chainID), blocks)
	})
	if err != nil {
		return fmt.Errorf("error while SendTx SubmitLastObservedBlocks: %w", bsc.ethHelper.ProcessError(err))
	}

	return nil
}

func (bsc *OracleBridgeSmartContractImpl) GetBatchStatusAndTransactions(
	ctx context.Context, chainID string, batchID uint64,
) (uint8, []TxDataInfo, error) {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return 0, nil, fmt.Errorf("error while GetEthHelper: %w", err)
	}

	contract, err := contractbinding.NewBridgeContract(
		bsc.smartContractAddress,
		ethTxHelper.GetClient())
	if err != nil {
		return 0, nil, fmt.Errorf("error while NewBridgeContract: %w", bsc.ethHelper.ProcessError(err))
	}

	result, err := contract.GetBatchStatusAndTransactions(&bind.CallOpts{
		Context: ctx,
	}, bsc.chainIDConverter.ToChainIDNum(chainID), batchID)
	if err != nil {
		return 0, nil, fmt.Errorf("error while GetBatchStatusAndTransactions: %w", bsc.ethHelper.ProcessError(err))
	}

	return result.Status, result.Txs, nil
}
