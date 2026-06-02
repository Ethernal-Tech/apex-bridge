package chain

import (
	"time"

	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	skyline "github.com/Ethernal-Tech/solana-infrastructure/sendtx/skyline_program"
	"github.com/Ethernal-Tech/solana-infrastructure/tracker"
	store "github.com/Ethernal-Tech/solana-infrastructure/tracker/store"
	"github.com/hashicorp/go-hclog"
)

type SolanaChainObserverImpl struct {
	config      *oCore.SolanaChainConfig
	indexerDB   store.StorageHandler
	txsReceiver core.SolanaTxsReceiver
	lastSlot    uint64
	closedCh    chan struct{}
	logger      hclog.Logger
}

var _ core.SolanaChainObserver = (*SolanaChainObserverImpl)(nil)

func NewSolanaChainObserver(
	config *oCore.SolanaChainConfig,
	txsReceiver core.SolanaTxsReceiver,
	indexerDB store.StorageHandler,
	logger hclog.Logger,
) (*SolanaChainObserverImpl, error) {
	return &SolanaChainObserverImpl{
		config:      config,
		indexerDB:   indexerDB,
		txsReceiver: txsReceiver,
		closedCh:    make(chan struct{}),
		logger:      logger,
	}, nil
}

func (so *SolanaChainObserverImpl) Start() error {
	so.logger.Debug("Starting solana chain observer", "endpoint", so.config.TxProviderEndpoint)

	trackerConfig, err := loadTrackerConfigs(so.config, so.txsReceiver, so.logger)
	if err != nil {
		so.logger.Error("Failed to load tracker configs", "error", err)

		return err
	}

	tracker, notifyClosedCh, err := newEventTrackerWrapper(trackerConfig, so.indexerDB)
	if err != nil {
		so.logger.Error("Failed to create event tracker", "error", err)

		return err
	}

	go tracker.Start()

	go func() {
		for {
			select {
			case <-so.closedCh:
				tracker.Close() // close old tracker

				return

			case <-time.After(so.config.RestartTrackerPullCheck):
				// restart tracker if it is not alive
				so.logger.Debug("Check if tracker is alive", "endpoint", trackerConfig.RPCEndpoint)

				if !so.updateIsTrackerAlive() {
					so.logger.Debug("Tracker is not alive anymore", "endpoint", trackerConfig.RPCEndpoint)

					tracker.Close() // close old tracker

					select {
					case <-so.closedCh:
					case <-notifyClosedCh:
						_ = so.Start()
					}

					return
				}
			}
		}
	}()

	return nil
}

func (so *SolanaChainObserverImpl) Dispose() error {
	close(so.closedCh)

	return nil
}

func (so *SolanaChainObserverImpl) GetConfig() *oCore.SolanaChainConfig {
	return so.config
}

func (so *SolanaChainObserverImpl) updateIsTrackerAlive() bool {
	blockPoint, err := so.indexerDB.GetLatestBlockPoint()
	if err != nil {
		so.logger.Warn("failed to retrieve last processed solana slot from solana tracker: %w")

		return true
	}

	// everything is ok, tracker slot is greater then previous saved
	if blockPoint != nil && blockPoint.BlockSlot > so.lastSlot {
		so.lastSlot = blockPoint.BlockSlot // update last slot number

		return true
	}

	return false
}

func loadTrackerConfigs(
	config *oCore.SolanaChainConfig,
	txsReceiver core.SolanaTxsReceiver,
	logger hclog.Logger,
) (*tracker.EventTrackerConfig, error) {
	specs := tracker.ProgramEventSpecs{}

	_, err := specs.AddEventSpec(&skyline.BridgeRequestEvent{}, core.BridgeRequestEvent)
	if err != nil {
		return nil, err
	}

	_, err = specs.AddEventSpec(&skyline.TransactionExecutedEvent{}, core.TransactionExecutedEvent)
	if err != nil {
		return nil, err
	}

	_, err = specs.AddEventSpec(&skyline.HotWalletIncrementEvent{}, core.HotWalletIncrementEvent)
	if err != nil {
		return nil, err
	}

	TrackedPrograms := map[string]tracker.ProgramEventSpecs{
		config.TrackedProgram: specs,
	}

	return &tracker.EventTrackerConfig{
		RPCEndpoint:            config.TxProviderEndpoint,
		TrackedPrograms:        TrackedPrograms,
		PollTime:               config.PoolIntervalMiliseconds,
		BlockRoundingThreshold: config.SlotRoundingThreshold,
		EventSubscriber: &confirmedEventHandler{
			ChainID:     config.ChainID,
			TxsReceiver: txsReceiver,
			Logger:      logger,
		},
		Logger: logger.Named(time.Now().UTC().String()),
	}, nil
}

type confirmedEventHandler struct {
	TxsReceiver core.SolanaTxsReceiver
	ChainID     string
	Logger      hclog.Logger
}

func (handler confirmedEventHandler) AddEvent(event tracker.EventNotification) error {
	handler.Logger.Info("Confirmed Event Handler invoked",
		"event", event)

	err := handler.TxsReceiver.NewUnprocessedEvent(handler.ChainID, event)
	if err != nil {
		handler.Logger.Error("Failed to process new event", "err", err, "event", event)

		return err
	}

	handler.Logger.Info("Event has been processed", "event", event)

	return nil
}
