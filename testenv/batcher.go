//go:build !testenv
// +build !testenv

package testenv

type TTLFormatterFunc func(ttl uint64, batchID uint64) uint64

// getTTLFormatter returns formater for a test mode. By default it is just identity function
// 1 - first batch will fail
// 2 - first five batches will fail
// 3 - First batch 5 bathces fail in "random" predetermined sequence
func GetTTLFormatter(testMode uint8) TTLFormatterFunc {
	return func(ttl, batchID uint64) uint64 {
		return ttl
	}
}
