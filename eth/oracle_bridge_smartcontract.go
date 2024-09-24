package eth

import (
	"context"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hashicorp/go-hclog"
)

const submitClaimsGasLimit = uint64(10_000_000)

type CardanoBlock = contractbinding.IBridgeStructsCardanoBlock
type Claims = contractbinding.IBridgeStructsValidatorClaims

type SubmitOpts struct {
	GasLimitMultiplier float32
}

type IOracleBridgeSmartContract interface {
	GetLastObservedBlock(ctx context.Context, sourceChain string) (CardanoBlock, error)
	GetRawTransactionFromLastBatch(ctx context.Context, chainID string) ([]byte, error)
	SubmitClaims(ctx context.Context, claims Claims, submitOpts *SubmitOpts) error
	SubmitLastObservedBlocks(ctx context.Context, chainID string, blocks []CardanoBlock) error
}

type OracleBridgeSmartContractImpl struct {
	smartContractAddress ethcommon.Address
	ethHelper            *EthHelperWrapper
}

var _ IOracleBridgeSmartContract = (*OracleBridgeSmartContractImpl)(nil)

func NewOracleBridgeSmartContract(
	nodeURL, smartContractAddress string, isDynamic bool, logger hclog.Logger,
) *OracleBridgeSmartContractImpl {
	return &OracleBridgeSmartContractImpl{
		smartContractAddress: ethcommon.HexToAddress(smartContractAddress),
		ethHelper:            NewEthHelperWrapper(nodeURL, isDynamic, logger),
	}
}

func NewOracleBridgeSmartContractWithWallet(
	nodeURL, smartContractAddress string, wallet *ethtxhelper.EthTxWallet, isDynamic bool, logger hclog.Logger,
) (*OracleBridgeSmartContractImpl, error) {
	ethHelper, err := NewEthHelperWrapperWithWallet(nodeURL, wallet, isDynamic, logger)
	if err != nil {
		return nil, err
	}

	return &OracleBridgeSmartContractImpl{
		smartContractAddress: ethcommon.HexToAddress(smartContractAddress),
		ethHelper:            ethHelper,
	}, nil
}

func (bsc *OracleBridgeSmartContractImpl) GetLastObservedBlock(
	ctx context.Context, sourceChain string,
) (CardanoBlock, error) {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return CardanoBlock{}, err
	}

	contract, err := contractbinding.NewBridgeContract(
		bsc.smartContractAddress,
		ethTxHelper.GetClient())
	if err != nil {
		return CardanoBlock{}, bsc.ethHelper.ProcessError(err)
	}

	result, err := contract.GetLastObservedBlock(&bind.CallOpts{
		Context: ctx,
	}, common.ToNumChainID(sourceChain))
	if err != nil {
		return CardanoBlock{}, bsc.ethHelper.ProcessError(err)
	}

	return result, nil
}

func (bsc *OracleBridgeSmartContractImpl) GetRawTransactionFromLastBatch(
	ctx context.Context, chainID string,
) ([]byte, error) {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return nil, err
	}

	contract, err := contractbinding.NewBridgeContract(
		bsc.smartContractAddress,
		ethTxHelper.GetClient())
	if err != nil {
		return nil, bsc.ethHelper.ProcessError(err)
	}

	result, err := contract.GetRawTransactionFromLastBatch(&bind.CallOpts{
		Context: ctx,
	}, common.ToNumChainID(chainID))
	if err != nil {
		return nil, bsc.ethHelper.ProcessError(err)
	}

	return result, nil
}

func (bsc *OracleBridgeSmartContractImpl) SubmitClaims(
	ctx context.Context, claims Claims, submitOpts *SubmitOpts,
) error {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return err
	}

	contract, err := contractbinding.NewBridgeContract(
		bsc.smartContractAddress,
		ethTxHelper.GetClient())
	if err != nil {
		return bsc.ethHelper.ProcessError(err)
	}

	_, err = bsc.ethHelper.SendTx(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		opts.GasLimit = submitClaimsGasLimit
		if submitOpts != nil && submitOpts.GasLimitMultiplier != 0 {
			opts.GasLimit = uint64(float32(opts.GasLimit) * submitOpts.GasLimitMultiplier)
		}

		return contract.SubmitClaims(opts, claims)
	})

	return bsc.ethHelper.ProcessError(err)
}

func (bsc *OracleBridgeSmartContractImpl) SubmitLastObservedBlocks(
	ctx context.Context, chainID string, blocks []CardanoBlock,
) error {
	ethTxHelper, err := bsc.ethHelper.GetEthHelper()
	if err != nil {
		return err
	}

	contract, err := contractbinding.NewBridgeContract(
		bsc.smartContractAddress,
		ethTxHelper.GetClient())
	if err != nil {
		return bsc.ethHelper.ProcessError(err)
	}

	_, err = bsc.ethHelper.SendTx(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return contract.SubmitLastObservedBlocks(opts, common.ToNumChainID(chainID), blocks)
	})

	return bsc.ethHelper.ProcessError(err)
}
