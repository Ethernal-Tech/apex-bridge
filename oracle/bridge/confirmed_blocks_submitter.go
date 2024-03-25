package bridge

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	indexerDb "github.com/Ethernal-Tech/cardano-infrastructure/indexer/db"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
)

type ConfirmedBlocksSubmitterImpl struct {
	appConfig *core.AppConfig
	chainId   string
	db        indexer.Database
	logger    hclog.Logger
	ctx       context.Context
	cancelCtx context.CancelFunc
	ethClient *ethclient.Client
}

var _ core.ConfirmedBlocksSubmitter = (*ConfirmedBlocksSubmitterImpl)(nil)

func NewConfirmedBlocksSubmitter(
	appConfig *core.AppConfig,
	chainId string,
	logger hclog.Logger,
) *ConfirmedBlocksSubmitterImpl {
	indexerDb, err := indexerDb.NewDatabaseInit("", appConfig.Settings.DbsPath+chainId+".db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		logger.Error("Open database failed", "err", err)
		return nil
	}

	ctx, cancelCtx := context.WithCancel(context.Background())

	return &ConfirmedBlocksSubmitterImpl{
		appConfig: appConfig,
		chainId:   chainId,
		db:        indexerDb,
		logger:    logger,
		ctx:       ctx,
		cancelCtx: cancelCtx,
	}
}

func (bs *ConfirmedBlocksSubmitterImpl) StartSubmit() error {
	for {
		select {
		case <-bs.ctx.Done():
			return nil
		default:
			blocks, err := bs.db.GetLatestConfirmedBlocks(bs.appConfig.Bridge.SubmitConfig.ConfirmedBlocksThreshhold)
			if err != nil {
				bs.logger.Error("Error submiting confirmed blocks", "err:", err)
				return err
			}

			if bs.ethClient == nil {
				ethClient, err := ethclient.Dial(bs.appConfig.Bridge.NodeUrl)
				if err != nil {
					bs.logger.Error("Failed to dial bridge", "err", err)
					return err
				}

				bs.ethClient = ethClient
			}

			ethTxHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithClient(bs.ethClient))
			if err != nil {
				// ensure redial in case ethClient lost connection
				bs.ethClient = nil
				bs.logger.Error("Failed to create ethTxHelper", "err", err)
				return err
			}

			bs.submitConfirmedBlocks(ethTxHelper, blocks)

			time.Sleep(time.Millisecond * time.Duration(bs.appConfig.Bridge.SubmitConfig.ConfirmedBlocksSubmitTime))
		}
	}
}

func (bs *ConfirmedBlocksSubmitterImpl) Dispose() error {
	bs.cancelCtx()

	if bs.ethClient != nil {
		bs.ethClient.Close()
		bs.ethClient = nil
	}

	return nil
}

func (bs *ConfirmedBlocksSubmitterImpl) submitConfirmedBlocks(ethTxHelper *ethtxhelper.EthTxHelperImpl, blocks []*indexer.CardanoBlock) (*types.Receipt, error) {
	// TODO: replace with real bridge contract
	contract, err := contractbinding.NewTestContract(
		common.HexToAddress(bs.appConfig.Bridge.SmartContractAddress), ethTxHelper.GetClient())
	if err != nil {
		return nil, err
	}

	wallet, err := ethtxhelper.NewEthTxWallet(string(bs.appConfig.Bridge.SigningKey))
	if err != nil {
		return nil, err
	}

	tx, err := ethTxHelper.SendTx(bs.ctx, wallet, bind.TransactOpts{}, true, func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
		// TODO: replace with real bridge contract call
		return contract.SetValue(txOpts, new(big.Int).SetUint64(
			uint64(len(blocks)+len(bs.chainId)),
		))
	})
	if err != nil {
		return nil, err
	}

	bs.logger.Info("tx has been sent", "tx hash", tx.Hash().String())

	receipt, err := ethTxHelper.WaitForReceipt(bs.ctx, tx.Hash().String(), true)
	if err != nil {
		return nil, err
	}

	bs.logger.Info("tx has been executed", "block", receipt.BlockHash.String(), "tx hash", receipt.TxHash.String())

	return receipt, nil
}
