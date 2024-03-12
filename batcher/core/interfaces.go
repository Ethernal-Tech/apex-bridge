package core

import "context"

type BatcherManager interface {
	Start() error
	Stop() error
}

type Batcher interface {
	Start(ctx context.Context) error
}
