package common

import (
	"encoding/hex"
	"strings"
)

const HashSize = 32

type Hash [HashSize]byte

type BridgingRequestStateKey struct {
	SourceChainID string
	SourceTxHash  Hash
}

type NewBridgingRequestStateModel struct {
	SourceTxHash Hash
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

func NewHashFromHexString(hash string) Hash {
	v, _ := hex.DecodeString(strings.TrimPrefix(hash, "0x"))

	return NewHashFromBytes(v)
}

func NewHashFromBytes(bytes []byte) Hash {
	if len(bytes) != HashSize {
		result := Hash{}
		size := min(HashSize, len(bytes))

		copy(result[HashSize-size:], bytes[:size])

		return result
	}

	return Hash(bytes)
}
