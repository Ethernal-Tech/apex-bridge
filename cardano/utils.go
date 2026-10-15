package cardanotx

import (
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

const splitStringLength = 40

func IsValidOutputAddress(addr string, networkID wallet.CardanoNetworkType) bool {
	cardAddr, err := wallet.NewCardanoAddressFromString(addr)

	return err == nil && cardAddr.GetInfo().AddressType != wallet.RewardAddress &&
		cardAddr.GetInfo().Network == networkID
}

// IsValidReceiverAddress validates a payout address taken from user supplied metadata.
// Byron addresses are rejected by header before GetInfo is called: the cardano-infrastructure
// Byron parser panics on a malformed address root.
func IsValidReceiverAddress(addr string, networkID wallet.CardanoNetworkType) bool {
	cardAddr, err := wallet.NewCardanoAddressFromString(addr)
	if err != nil || wallet.GetAddressTypeFromHeader(cardAddr.GetBytes()[0]) == wallet.ByronAddress {
		return false
	}

	return cardAddr.GetInfo().AddressType != wallet.RewardAddress && cardAddr.GetInfo().Network == networkID
}
