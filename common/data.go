package common

import (
	"encoding/hex"
)

type VCRunMode string

const (
	HashSize = 32

	EthZeroAddr = "0x0000000000000000000000000000000000000000"

	MinFeeForBridgingDefault   = uint64(1_000_010)
	MinFeeForBridgingToCardano = uint64(1_000_010)
	MinFeeForBridgingToPrime   = uint64(1_000_010)

	MinUtxoAmountDefault        = uint64(1_000_000)
	MinUtxoAmountDefaultCardano = uint64(1_000_000)
	MinUtxoAmountDefaultPrime   = uint64(1_000_000)

	PotentialFeeDefault           = 250_000
	MaxInputsPerBridgingTxDefault = 50

	ReactorMode VCRunMode = "reactor"
	SkylineMode VCRunMode = "skyline"
)

var (
	DefundTxHash, _ = hex.DecodeString("c74d0d70be942fd68984df57687b9f453f1321726e8db77762dee952a5c85b24")
)

type Hash [HashSize]byte

type BridgingRequestStateKey struct {
	SourceChainID string
	SourceTxHash  Hash
}

func NewBridgingRequestStateKey(sourceChainID string, sourceTxHash Hash) BridgingRequestStateKey {
	return BridgingRequestStateKey{
		SourceChainID: sourceChainID,
		SourceTxHash:  sourceTxHash,
	}
}

type NewBridgingRequestStateModel struct {
	SourceTxHash Hash
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

func NewHashFromHexString(hash string) Hash {
	v, _ := DecodeHex(hash)

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
