package chain

import (
	"context"

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
