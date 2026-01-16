package common

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
)

type VCRunMode string

const (
	HashSize = 32

	EthZeroAddr = "0x0000000000000000000000000000000000000000"

	PotentialFeeDefault           = 250_000
	MaxInputsPerBridgingTxDefault = 50

	ReactorMode VCRunMode = "reactor"
	SkylineMode VCRunMode = "skyline"

	BridgingConfirmedTxType       ConfirmedTxType = 0
	DefundConfirmedTxType         ConfirmedTxType = 1
	RefundConfirmedTxType         ConfirmedTxType = 2
	StakeConfirmedTxType          ConfirmedTxType = 3
	RedistributionConfirmedTxType ConfirmedTxType = 4

	StakeRegDelConfirmedTxSubType StakeConfirmedTxSubType = 0
	StakeDelConfirmedTxSubType    StakeConfirmedTxSubType = 1
	StakeDeregConfirmedTxSubType  StakeConfirmedTxSubType = 2
)

type Hash [HashSize]byte

type TxOutputIndex uint16

type BridgingRequestStateKey struct {
	SourceChainID string
	SourceTxHash  Hash
	IsRefund      bool
}

func NewBridgingRequestStateKey(sourceChainID string, sourceTxHash Hash, isRefund bool) BridgingRequestStateKey {
	return BridgingRequestStateKey{
		SourceChainID: sourceChainID,
		SourceTxHash:  sourceTxHash,
		IsRefund:      isRefund,
	}
}

type NewBridgingRequestStateModel struct {
	SourceTxHash Hash
	IsRefund     bool
}

type ConfirmedTxType uint8
type StakeConfirmedTxSubType uint8

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

func IsDirectlyConfirmedTransaction(txType uint8) bool {
	return txType == uint8(StakeConfirmedTxType) ||
		txType == uint8(DefundConfirmedTxType) ||
		txType == uint8(RedistributionConfirmedTxType)
}

type BigInt struct {
	*big.Int
}

func NewBigInt(v *big.Int) BigInt {
	return BigInt{Int: v}
}

func (b BigInt) MarshalJSON() ([]byte, error) {
	if b.Int == nil {
		return []byte("null"), nil
	}
	// encode as string to preserve precision
	return json.Marshal(b.String())
}

func (b *BigInt) UnmarshalJSON(data []byte) error {
	// allow null
	if string(data) == "null" {
		b.Int = nil
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("big.Int must be a JSON string: %w", err)
	}

	i, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return fmt.Errorf("invalid big.Int value: %s", s)
	}

	b.Int = i
	return nil
}
