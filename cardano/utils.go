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
