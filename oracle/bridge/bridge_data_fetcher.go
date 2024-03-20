package bridge

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
)

const (
	TickTimeMs = 5000
)

type BridgeDataFetcherImpl struct {
	appConfig *core.AppConfig
	db        core.BridgeExpectedCardanoTxsDb
	ctx       context.Context
	cancelCtx context.CancelFunc
	ethClient *ethclient.Client
	logger    hclog.Logger
}

var _ core.BridgeDataFetcher = (*BridgeDataFetcherImpl)(nil)

func NewBridgeDataFetcher(
	appConfig *core.AppConfig,
	db core.BridgeExpectedCardanoTxsDb,
	logger hclog.Logger,
) *BridgeDataFetcherImpl {
	ctx, cancelCtx := context.WithCancel(context.Background())
	return &BridgeDataFetcherImpl{
		appConfig: appConfig,
		db:        db,
		ctx:       ctx,
		cancelCtx: cancelCtx,
		logger:    logger,
	}
}

func (df *BridgeDataFetcherImpl) Start() error {
	df.logger.Debug("Starting BridgeDataFetcher")

	timerTime := TickTimeMs * time.Millisecond
	timer := time.NewTimer(timerTime)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			df.fetchData()
		case <-df.ctx.Done():
			return nil
		}

		timer.Reset(timerTime)
	}
}

func (df *BridgeDataFetcherImpl) Stop() error {
	df.logger.Debug("Stopping BridgeDataFetcher")
	df.cancelCtx()
	return nil
}

func (df *BridgeDataFetcherImpl) fetchData() {
	if df.ethClient == nil {
		ethClient, err := ethclient.Dial(df.appConfig.Bridge.NodeUrl)
		if err != nil {
			df.logger.Error("Failed to dial bridge", "err", err)
			return
		}

		df.ethClient = ethClient
	}

	ethTxHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithClient(df.ethClient))
	if err != nil {
		// ensure redial in case ethClient lost connection
		df.ethClient = nil
		df.logger.Error("Failed to create ethTxHelper", "err", err)
		return
	}

	expectedTxs, err := df.fetchExpectedTxs(ethTxHelper)
	if err != nil {
		// ensure redial in case ethClient lost connection
		df.ethClient = nil
		df.logger.Error("Failed to fetch expected txs from bridge", "err", err)
		return
	}

	if len(expectedTxs) > 0 {
		err = df.db.AddExpectedTxs(expectedTxs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add expected txs. error: %v\n", err)
			df.logger.Error("Failed to add expected txs", "err", err)
			return
		}
	}
}

func (df *BridgeDataFetcherImpl) fetchExpectedTxs(ethTxHelper ethtxhelper.IEthTxHelper) ([]*core.BridgeExpectedCardanoTx, error) {
	// TODO: replace with real bridge contract
	contract, err := contractbinding.NewTestContract(
		common.HexToAddress(df.appConfig.Bridge.SmartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return nil, err
	}

	// TODO: replace with real bridge contract call
	_, err = contract.GetValue(&bind.CallOpts{
		Context: df.ctx,
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (df *BridgeDataFetcherImpl) FetchLatestBlockPoint(chainId string) (*indexer.BlockPoint, error) {
	ethTxHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithClient(df.ethClient))
	if err != nil {
		// ensure redial in case ethClient lost connection
		df.ethClient = nil
		df.logger.Error("Failed to create ethTxHelper", "err", err)
		return nil, err
	}

	// TODO: replace with real bridge contract
	contract, err := contractbinding.NewTestContract(
		common.HexToAddress(df.appConfig.Bridge.SmartContractAddress),
		ethTxHelper.GetClient())
	if err != nil {
		return nil, err
	}

	for retries := 0; retries < 5; retries++ {
		// TODO: replace with real bridge contract call
		_, err = contract.GetValue(&bind.CallOpts{
			Context: df.ctx,
		})
		if err == nil {
			break
		}
		time.Sleep(time.Millisecond * 500)
	}

	return nil, nil
}
