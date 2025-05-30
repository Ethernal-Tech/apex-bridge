package core

import "context"

type StakingManager interface {
	Start()
}

type StakingComponent interface {
	Start(ctx context.Context)
}
