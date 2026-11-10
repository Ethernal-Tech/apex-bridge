package successtxprocessors

import (
	"hash/crc32"
	"testing"

	"github.com/blinklabs-io/gouroboros/base58"
	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/require"
)

// CIP-19 Byron testnet test vector
const byronTestnetAddr = "37btjrVyb4KDXBNC4haBVPCrro8AQPHwvCMp3RFhhSVWwfFmZ6wwzSK6JK1hY6wHNmtrpTf1kdbva8TCneM2YsiXT7mrzT21EacHnPpz5YyUdj64na"

// getByronAddresses returns a well formed Byron address and a checksum-valid encoding
// with a too short address root.
func getByronAddresses(t *testing.T) []string {
	t.Helper()

	payload, err := cbor.Marshal([]any{make([]byte, 20), map[uint64][]byte{}, uint64(0)})
	require.NoError(t, err)

	raw, err := cbor.Marshal([]any{cbor.Tag{Number: 24, Content: payload}, crc32.ChecksumIEEE(payload)})
	require.NoError(t, err)

	// Encode directly so the validating constructor does not reject the test fixture.
	return []string{byronTestnetAddr, base58.Encode(raw)}
}
