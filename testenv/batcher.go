//go:build !testenv
// +build !testenv

package testenv

type TTLFormatterFunc func(ttl uint64, batchID uint64) uint64

// GetTTLFormatter returns a TTLFormatterFunc that always returns the provided TTL value unchanged.
func GetTTLFormatter(testMode uint8) TTLFormatterFunc {
	return func(ttl, batchID uint64) uint64 {
		return ttl
	}
}
