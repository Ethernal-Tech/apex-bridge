package queue

import (
	"sync"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/hashicorp/go-hclog"
)

type ExecutableQueueFN func() error

type RecoverableErrorFN func(err error) bool

type ExecutableQueue struct {
	queue                *ConsumerQueue[ExecutableQueueFN]
	isRecoverableErrorFN RecoverableErrorFN
	logger               hclog.Logger
}

func NewExecutableQueue(
	isRecoverableErrorFN RecoverableErrorFN, logger hclog.Logger,
) *ExecutableQueue {
	return &ExecutableQueue{
		queue:                NewCosumerQueue[ExecutableQueueFN](),
		isRecoverableErrorFN: isRecoverableErrorFN,
		logger:               logger,
	}
}

func (eq *ExecutableQueue) Execute() {
	for {
		items := eq.queue.WaitForItems()
		if items == nil {
			return
		}

		for _, fn := range items {
			err := fn()
			if err != nil {
				if eq.isRecoverableErrorFN(err) {
					eq.logger.Info("queue error", "err", err)

					eq.queue.AddWithoutSignal(fn)
				} else if !common.IsContextDoneErr(err) {
					eq.logger.Error("uncrecoverable queue error", "err", err)
				}
			}
		}
	}
}

func (eq *ExecutableQueue) Add(fn ExecutableQueueFN) {
	eq.queue.Add(fn)
}

func (eq *ExecutableQueue) Stop() {
	eq.queue.Stop()
}

type ConsumerQueue[T any] struct {
	lock    *sync.Cond
	data    []T
	stopped bool
}

func NewCosumerQueue[T any]() *ConsumerQueue[T] {
	return &ConsumerQueue[T]{
		lock: sync.NewCond(&sync.Mutex{}),
		data: []T{},
	}
}

func (cq *ConsumerQueue[T]) Add(item T) {
	cq.lock.L.Lock()
	cq.data = append(cq.data, item)
	cq.lock.Signal()
	cq.lock.L.Unlock()
}

func (cq *ConsumerQueue[T]) AddWithoutSignal(item T) {
	cq.lock.L.Lock()
	cq.data = append(cq.data, item)
	cq.lock.L.Unlock()
}

func (cq *ConsumerQueue[T]) WaitForItems() (result []T) {
	cq.lock.L.Lock()

	for len(cq.data) == 0 && !cq.stopped {
		cq.lock.Wait()
	}

	defer cq.lock.L.Unlock()

	if cq.stopped {
		return nil
	}

	result = append([]T{}, cq.data...)
	cq.data = cq.data[:0]

	return result
}

func (cq *ConsumerQueue[T]) Stop() {
	cq.lock.L.Lock()
	cq.stopped = true
	cq.lock.Broadcast()
	cq.lock.L.Unlock()
}
