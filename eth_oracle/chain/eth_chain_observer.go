package chain

import (
	"context"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/eth_oracle/core"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	eventTracker "github.com/Ethernal-Tech/blockchain-event-tracker/tracker"
	"github.com/hashicorp/go-hclog"
)

type EthChainObserverImpl struct {
	ctx     context.Context
	logger  hclog.Logger
	config  *oracleCore.EthChainConfig
	tracker *eventTracker.EventTracker
}

var _ core.EthChainObserver = (*EthChainObserverImpl)(nil)

func NewEthChainObserver(
	ctx context.Context,
	logger hclog.Logger,
	config *oracleCore.EthChainConfig,
) (*EthChainObserverImpl, error) {
	trackerConfig := &eventTracker.EventTrackerConfig{
		SyncBatchSize: 10,
		RPCEndpoint:   "https://some-url.com",
		PollInterval:  10 * time.Second,
	}

	trackerStore, err := eventTrackerStore.NewBoltDBEventTrackerStore("/path/to/my.db")
	if err != nil {
		return nil, err
	}

	ethTracker, err := eventTracker.NewEventTracker(trackerConfig, trackerStore, 0)
	if err != nil {
		return nil, err
	}

	return &EthChainObserverImpl{
		ctx:     ctx,
		logger:  logger,
		config:  config,
		tracker: ethTracker,
	}, nil
}

func (co *EthChainObserverImpl) Start() error {
	if err := co.tracker.Start(); err != nil {
		return err
	}

	return nil
}

func (co *EthChainObserverImpl) Dispose() error {
	return nil
}

func (co *EthChainObserverImpl) GetConfig() *oracleCore.EthChainConfig {
	return nil
}

func (co *EthChainObserverImpl) ErrorCh() <-chan error {
	return nil
}
