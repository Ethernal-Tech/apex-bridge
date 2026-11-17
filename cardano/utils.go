package cardanotx

import (
	"bytes"

	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

const splitStringLength = 40

// IsValidReceiverAddress validates an address that a batch pays to. The oracles use it on user
// supplied metadata and the batcher uses it again before building outputs, so both apply the same rules.
//
// Byron addresses are rejected by header before GetInfo is called: the cardano-infrastructure
// Byron parser panics on a malformed address root.
//
// The address must also be in canonical form (see isCanonicalShelleyAddress), because the
// cardano-infrastructure parser accepts encodings that the ledger's strict address decoding rejects.
func IsValidReceiverAddress(addr string, networkID wallet.CardanoNetworkType) bool {
	cardAddr, err := wallet.NewCardanoAddressFromString(addr)
	if err != nil || wallet.GetAddressTypeFromHeader(cardAddr.GetBytes()[0]) == wallet.ByronAddress {
		return false
	}

	info := cardAddr.GetInfo()
	if info.AddressType == wallet.RewardAddress || info.Network != networkID {
		return false
	}

	return isCanonicalShelleyAddress(addr, cardAddr.GetBytes(), info)
}

// isCanonicalShelleyAddress reports whether raw is exactly the encoding of its own decoded parts
// and addr is the bech32 string of raw. This rejects base addresses with trailing bytes (the
// parser only checks a minimum length), pointer addresses with non-minimal varints, and strings
// whose prefix does not match the network or that are not lower case.
func isCanonicalShelleyAddress(addr string, raw []byte, info wallet.CardanoAddressInfo) bool {
	if len(info.Extra) > 0 {
		return false
	}

	rebuilt, err := info.ToCardanoAddress()
	if err != nil || !bytes.Equal(rebuilt.GetBytes(), raw) {
		return false
	}

	// rebuilt has no cached string, so String encodes it again from the bytes
	return rebuilt.String() == addr
}
