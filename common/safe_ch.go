package common

import (
	"errors"
	"sync"
)

type SafeCh[T any] struct {
	ch     chan T
	closed bool
	m      sync.Mutex
}

func MakeSafeCh[T any](size int) *SafeCh[T] {
	return &SafeCh[T]{
		ch:     make(chan T, size),
		closed: false,
	}
}

func (sch *SafeCh[T]) Close() error {
	sch.m.Lock()
	defer sch.m.Unlock()

	if sch.ch == nil {
		return errors.New("channel not initialized. use MakeSafeCh")
	}

	if sch.closed {
		return errors.New("channel already closed")
	}

	close(sch.ch)
	sch.closed = true

	return nil
}

func (sch *SafeCh[T]) ReadCh() <-chan T {
	if sch.ch == nil {
		sch.ch = make(chan T, 1)
	}

	return sch.ch
}

func (sch *SafeCh[T]) Write(obj T) error {
	sch.m.Lock()
	defer sch.m.Unlock()

	if sch.ch == nil {
		return errors.New("channel not initialized. use MakeSafeCh")
	}

	if sch.closed {
		return errors.New("trying to write to a closed channel")
	}

	sch.ch <- obj

	return nil
}
