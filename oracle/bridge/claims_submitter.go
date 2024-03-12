package bridge

import (
	"context"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
)

type ClaimsSubmitterImpl struct {
	appConfig *core.AppConfig
	logger    hclog.Logger
	ctx       context.Context
	cancelCtx context.CancelFunc
	ethClient *ethclient.Client
}

var _ core.ClaimsSubmitter = (*ClaimsSubmitterImpl)(nil)

func NewClaimsSubmitter(
	appConfig *core.AppConfig,
	logger hclog.Logger,
) *ClaimsSubmitterImpl {
	ctx, cancelCtx := context.WithCancel(context.Background())
	return &ClaimsSubmitterImpl{
		appConfig: appConfig,
		logger:    logger,
		ctx:       ctx,
		cancelCtx: cancelCtx,
	}
}

func (cs *ClaimsSubmitterImpl) SubmitClaims(claims *core.BridgeClaims) error {
	if cs.ethClient == nil {
		ethClient, err := ethclient.Dial(cs.appConfig.Bridge.NodeUrl)
		if err != nil {
			cs.logger.Error("Failed to dial bridge", "err", err)
			return err
		}

		cs.ethClient = ethClient
	}

	ethTxHelper, _ := ethtxhelper.NewEThTxHelper(ethtxhelper.WithClient(cs.ethClient))
	receipt, err := cs.sendTx(ethTxHelper, claims)
	if err != nil {
		// ensure redial in case ethClient lost connection
		cs.ethClient = nil
		cs.logger.Error("failed to send tx", "err", err)
	}

	if receipt != nil {
		cs.logger.Info("Sent tx", "receipt", receipt)
	}

	return err
}

func (cs *ClaimsSubmitterImpl) Dispose() error {
	cs.cancelCtx()

	if cs.ethClient != nil {
		cs.ethClient.Close()
		cs.ethClient = nil
	}

	return nil
}

func (cs *ClaimsSubmitterImpl) sendTx(ethTxHelper ethtxhelper.IEthTxHelper, claims *core.BridgeClaims) (*types.Receipt, error) {
	// TODO: replace with real bridge contract
	contract, err := contractbinding.NewTestContract(
		common.HexToAddress(cs.appConfig.Bridge.SmartContractAddress), ethTxHelper.GetClient())
	if err != nil {
		return nil, err
	}

	wallet, err := ethtxhelper.NewEthTxWallet(string(cs.appConfig.Bridge.SigningKey))
	if err != nil {
		return nil, err
	}

	tx, err := ethTxHelper.SendTx(cs.ctx, wallet, bind.TransactOpts{}, true, func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
		// TODO: replace with real bridge contract call
		return contract.SetValue(txOpts, new(big.Int).SetUint64(
			uint64(len(claims.BatchExecuted)+len(claims.BridgingRequest)+len(claims.BatchExecutionFailed)),
		))
	})
	if err != nil {
		return nil, err
	}

	cs.logger.Info("tx has been sent", "tx hash", tx.Hash().String())

	receipt, err := ethTxHelper.WaitForReceipt(cs.ctx, tx.Hash().String(), true)
	if err != nil {
		return nil, err
	}

	cs.logger.Info("tx has been executed", "block", receipt.BlockHash.String(), "tx hash", receipt.TxHash.String())

	return receipt, nil
}
