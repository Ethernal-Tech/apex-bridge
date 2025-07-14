package validatorobserver

import (
	"context"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/hashicorp/go-hclog"
)

type ValidatorSetObserver struct {
	validatorSetPending bool
	bridgeSmartContract eth.IBridgeSmartContract
	logger              hclog.Logger
}

const (
	timeout = 1 * time.Second
)

func NewValidatorSetObserver(
	bridgeSmartContract eth.IBridgeSmartContract,
	logger hclog.Logger,
) (*ValidatorSetObserver, error) {
	validatorSetPending, err := bridgeSmartContract.IsNewValidatorSetPending()
	if err != nil {
		// TODO: no need to err at this point?
		validatorSetPending = false
	}

	return &ValidatorSetObserver{
		validatorSetPending: validatorSetPending,
		bridgeSmartContract: bridgeSmartContract,
		logger:              logger.Named("validator_set_observer"),
	}, nil
}

func (vs *ValidatorSetObserver) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(timeout):
				if err := vs.execute(); err != nil {
					vs.logger.Error("error while executing", "err", err)
				}
			}
		}
	}()
}

func (vs *ValidatorSetObserver) execute() error {
	var err error

	vs.validatorSetPending, err = vs.bridgeSmartContract.IsNewValidatorSetPending()

	return err
}

func (vs *ValidatorSetObserver) IsValidatorSetPending() bool {
	return vs.validatorSetPending
}
