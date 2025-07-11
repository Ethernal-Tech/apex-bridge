package cardanotx

import (
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

const splitStringLength = 40

func IsValidOutputAddress(addr string, networkID wallet.CardanoNetworkType) bool {
	cardAddr, err := wallet.NewCardanoAddressFromString(addr)

	return err == nil && cardAddr.GetInfo().AddressType != wallet.RewardAddress &&
		cardAddr.GetInfo().Network == networkID
}

func AddrToMetaDataAddr(addr string) []string {
	addr = strings.TrimPrefix(strings.TrimPrefix(addr, "0x"), "0X")

	return common.SplitString(addr, splitStringLength)
}
