package validatorobserver

import (
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/stretchr/testify/mock"
)

type ValidatorSetObserverMock struct {
	mock.Mock
}

var _ IValidatorSetObserver = (*ValidatorSetObserverMock)(nil)

func (vso *ValidatorSetObserverMock) IsValidatorSetPending() bool {
	args := vso.Called()

	return args.Bool(0)
}

func (vso *ValidatorSetObserverMock) GetValidatorSet(chainID string) []eth.ValidatorChainData {
	args := vso.Called(chainID)

	if args.Get(0) != nil {
		arg0, _ := args.Get(0).([]eth.ValidatorChainData)

		return arg0
	}

	return nil
}

func (vso *ValidatorSetObserverMock) GetValidatorSetReader() <-chan *ValidatorsPerChain {
	args := vso.Called()

	if args.Get(0) != nil {
		arg0, _ := args.Get(0).(<-chan *ValidatorsPerChain)

		return arg0
	}

	return nil
}
