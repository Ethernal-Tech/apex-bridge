package validatorobserver

import "github.com/Ethernal-Tech/apex-bridge/eth"

type IValidatorSetObserver interface {
	IsValidatorSetPending() bool
	GetValidatorSet(chainID string) []eth.ValidatorChainData
	GetValidatorSetReader() <-chan *ValidatorsPerChain
}
