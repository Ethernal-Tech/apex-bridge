package chain

import (
	"context"
	"time"

	"github.com/Ethernal-Tech/solana-infrastructure/tracker"
	store "github.com/Ethernal-Tech/solana-infrastructure/tracker/store"
)

type eventTrackerWrapper struct {
	eventTracker   *tracker.EventTracker
	notifyClosedCh chan struct{}
	ctx            context.Context
	cancelFunc     context.CancelFunc
}

func newEventTrackerWrapper(
	config *tracker.EventTrackerConfig, store store.StorageHandler,
) (*eventTrackerWrapper, <-chan struct{}, error) {
	ctx, cancel := context.WithCancel(context.Background())
	notifyClosedCh := make(chan struct{})

	et, err := tracker.NewEventTracker(config, store)

	return &eventTrackerWrapper{
		eventTracker:   et,
		ctx:            ctx,
		cancelFunc:     cancel,
		notifyClosedCh: notifyClosedCh,
	}, notifyClosedCh, err
}

func (etw *eventTrackerWrapper) Close() {
	etw.cancelFunc()
}

func (etw *eventTrackerWrapper) Start() {
	defer close(etw.notifyClosedCh)

	if etw.eventTracker != nil {
		// add delay for rpc endpoint cooldown
		time.Sleep(10 * time.Second)

		etw.eventTracker.Start(etw.ctx)
	}
}
