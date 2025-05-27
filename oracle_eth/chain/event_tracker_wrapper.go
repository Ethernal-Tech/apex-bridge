package chain

import (
	"context"
	"time"

	eventStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	"github.com/Ethernal-Tech/blockchain-event-tracker/tracker"
)

type eventTrackerWrapper struct {
	eventTracker   *tracker.EventTracker
	notifyClosedCh chan struct{}
	ctx            context.Context
	cancelFunc     context.CancelFunc
}

func newEventTrackerWrapper(
	config *tracker.EventTrackerConfig, store eventStore.EventTrackerStore,
) (*eventTrackerWrapper, <-chan struct{}, error) {
	ctx, cancel := context.WithCancel(context.Background())
	notifyClosedCh := make(chan struct{})
	// add timestamp to the logger to differentiate between multiple instances
	config.Logger = config.Logger.Named(time.Now().UTC().String())

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
		etw.eventTracker.Start(etw.ctx)
	}
}
