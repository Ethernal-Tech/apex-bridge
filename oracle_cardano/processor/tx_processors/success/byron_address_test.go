package successtxprocessors

import (
	"hash/crc32"
	"testing"

	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/require"
)

// CIP-19 Byron testnet test vector
const byronTestnetAddr = "37btjrVyb4KDXBNC4haBVPCrro8AQPHwvCMp3RFhhSVWwfFmZ6wwzSK6JK1hY6wHNmtrpTf1kdbva8TCneM2YsiXT7mrzT21EacHnPpz5YyUdj64na"

// getByronAddresses returns a well formed Byron address and one with a too short address root,
// which panics cardano-infrastructure on GetInfo
func getByronAddresses(t *testing.T) []string {
	t.Helper()

	payload, err := cbor.Marshal([]any{make([]byte, 20), map[uint64][]byte{}, uint64(0)})
	require.NoError(t, err)

	raw, err := cbor.Marshal([]any{cbor.Tag{Number: 24, Content: payload}, crc32.ChecksumIEEE(payload)})
	require.NoError(t, err)

	addr, err := wallet.NewCardanoAddress(raw)
	require.NoError(t, err)

	return []string{byronTestnetAddr, addr.String()}
}
