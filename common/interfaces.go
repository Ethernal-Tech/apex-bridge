package common

import (
	"context"

	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type BridgingRequestStateUpdater interface {
	New(srcChainID string, model *NewBridgingRequestStateModel) error
	NewMultiple(srcChainID string, models []*NewBridgingRequestStateModel) error
	Invalid(key BridgingRequestStateKey) error
	SubmittedToBridge(key BridgingRequestStateKey, dstChainID string) error
	IncludedInBatch(txs []BridgingRequestStateKey, dstChainID string) error
	SubmittedToDestination(txs []BridgingRequestStateKey, dstChainID string) error
	FailedToExecuteOnDestination(txs []BridgingRequestStateKey, dstChainID string) error
	ExecutedOnDestination(txs []BridgingRequestStateKey, dstTxHash Hash, dstChainID string) error
}

// ChainSpecificConfig defines the interface for chain-specific configurations
type ChainSpecificConfig interface {
	GetChainType() string
}

type IStartable interface {
	Start(context.Context)
}

type BridgingAddressesManager interface {
	GetAllPaymentAddresses(chainID uint8) []string
	GetAllStakeAddresses(chainID uint8) []string
	GetPaymentAddressIndex(chainID uint8, address string) (uint8, bool)
	GetStakeAddressIndex(chainID uint8, address string) (uint8, bool)
	GetPaymentAddressFromIndex(chainID uint8, index uint8) (string, bool)
	GetStakeAddressFromIndex(chainID uint8, index uint8) (string, bool)
	GetPaymentPolicyScript(chainID uint8, index uint8) (*cardanowallet.PolicyScript, bool)
	GetStakePolicyScript(chainID uint8, index uint8) (*cardanowallet.PolicyScript, bool)
	GetFeeMultisigAddress(chainID uint8) string
	GetFeeMultisigPolicyScript(chainID uint8) (*cardanowallet.PolicyScript, bool)
}

type AddressAndAmount struct {
	AddressIndex      uint8
	Address           string
	TokensAmounts     map[string]uint64
	IncludeChange     uint64
	UtxoCount         int
	ShouldConsolidate bool
}

type TxOutputs struct {
	Outputs []cardanowallet.TxOutput
	Sum     map[string]uint64
}

type BridgingAddressesCoordinator interface {
	GetAddressesAndAmountsForBatch(
		chainID uint8,
		cardanoCliBinary string,
		isRedistribution bool,
		protocolParams []byte,
		txOutputs TxOutputs) ([]AddressAndAmount, bool, error)
	GetAddressToBridgeTo(chainID uint8, containsNativeTokens bool) (AddressAndAmount, error)
}
