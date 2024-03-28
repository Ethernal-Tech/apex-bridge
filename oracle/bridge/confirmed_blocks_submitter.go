package bridge

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/utils"
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
	indexerDb indexer.Database
	oracleDb  core.CardanoTxsDb
	logger    hclog.Logger
	ctx       context.Context
	cancelCtx context.CancelFunc
	ethClient *ethclient.Client

	latestConfirmedSlot uint64
	errorCh             chan error
}

var _ core.ConfirmedBlocksSubmitter = (*ConfirmedBlocksSubmitterImpl)(nil)

func NewConfirmedBlocksSubmitter(
	appConfig *core.AppConfig,
	oracleDb core.CardanoTxsDb,
	chainId string,
	logger hclog.Logger,
) (*ConfirmedBlocksSubmitterImpl, error) {
	if err := utils.CreateDirectoryIfNotExists(appConfig.Settings.DbsPath); err != nil {
		return nil, err
	}

	indexerDb, err := indexerDb.NewDatabaseInit("", appConfig.Settings.DbsPath+chainId+".db")
	if err != nil {
		return nil, err
	}

	latestBlockPoint, err := indexerDb.GetLatestBlockPoint()
	if err != nil {
		return nil, err
	}
	if latestBlockPoint == nil {
		latestBlockPoint = &indexer.BlockPoint{}
	}

	ctx, cancelCtx := context.WithCancel(context.Background())

	return &ConfirmedBlocksSubmitterImpl{
		appConfig: appConfig,
		chainId:   chainId,
		indexerDb: indexerDb,
		oracleDb:  oracleDb,
		logger:    logger,
		ctx:       ctx,
		cancelCtx: cancelCtx,

		latestConfirmedSlot: latestBlockPoint.BlockSlot,
		errorCh:             make(chan error, 1),
	}, nil
}

func (bs *ConfirmedBlocksSubmitterImpl) StartSubmit() error {
	go func() {
		for {
			select {
			case <-bs.ctx.Done():
				return
			default:
				time.Sleep(time.Millisecond * time.Duration(bs.appConfig.Bridge.SubmitConfig.ConfirmedBlocksSubmitTime))

				// Threshhold +1 because we will ignore first block
				blocks, err := bs.indexerDb.GetConfirmedBlocksFrom(
					bs.latestConfirmedSlot,
					bs.appConfig.Bridge.SubmitConfig.ConfirmedBlocksThreshhold+1)
				if err != nil {
					bs.errorCh <- fmt.Errorf("error getting latest confirmed blocks err: %v", err)
				}

				if len(blocks) == 0 {
					continue
				}

				var blockCounter = 0
				// Skip first block becuase it's already processed
				for _, block := range blocks[1:] {
					if bs.checkIfBlockIsProcessed(block) {
						blockCounter++
						continue
					}
					break
				}

				if blockCounter == 0 {
					continue
				}

				if bs.ethClient == nil {
					ethClient, err := ethclient.Dial(bs.appConfig.Bridge.NodeUrl)
					if err != nil {
						bs.errorCh <- fmt.Errorf("failed to dial bridge: %v", err)
						continue
					}

					bs.ethClient = ethClient
				}

				ethTxHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithClient(bs.ethClient))
				if err != nil {
					// ensure redial in case ethClient lost connection
					bs.ethClient = nil
					bs.errorCh <- fmt.Errorf("failed to create ethTxHelper: %v", err)
					continue
				}

				if _, err := bs.submitConfirmedBlocks(ethTxHelper, blocks); err != nil {
					bs.errorCh <- fmt.Errorf("error submitting confirmed blocks: %v", err)
					continue
				}
				bs.latestConfirmedSlot = blocks[blockCounter].Slot
			}
		}
	}()

	return nil
}

func (bs *ConfirmedBlocksSubmitterImpl) Dispose() error {
	bs.cancelCtx()
	close(bs.errorCh)

	if bs.ethClient != nil {
		bs.ethClient.Close()
		bs.ethClient = nil
	}

	return nil
}

func (bs *ConfirmedBlocksSubmitterImpl) ErrorCh() <-chan error {
	return bs.errorCh
}

func (bs *ConfirmedBlocksSubmitterImpl) GetChainId() string {
	return bs.chainId
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

func (bs *ConfirmedBlocksSubmitterImpl) checkIfBlockIsProcessed(block *indexer.CardanoBlock) bool {
	if len(block.Txs) == 0 {
		return true
	}

	txsProcessed := true
	for _, tx := range block.Txs {
		prTx, err := bs.oracleDb.GetProcessedTx(bs.chainId, tx)
		if err != nil {
			bs.errorCh <- fmt.Errorf("error getting processed tx for block err: %v", err)
		}

		if prTx == nil {
			txsProcessed = false
			break
		}
	}

	return txsProcessed
}
