package core

import "context"

type RelayerManager interface {
	Start() error
	Stop() error
}

type Relayer interface {
	Start(ctx context.Context) error
}
