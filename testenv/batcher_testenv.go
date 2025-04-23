//go:build testenv
// +build testenv

package testenv

type TTLFormatterFunc func(ttl uint64, batchID uint64) uint64

// GetTTLFormatter returns formater for a test mode. By default it is just identity function
// 1 - first batch will fail
// 2 - first five batches will fail
// 3 - First batch 5 bathces fail in "random" predetermined sequence
func GetTTLFormatter(testMode uint8) TTLFormatterFunc {
	switch testMode {
	default:
		return func(ttl, batchID uint64) uint64 {
			return ttl
		}
	case 1:
		return func(ttl, batchID uint64) uint64 {
			if batchID > 1 {
				return ttl
			}

			return 0
		}
	case 2:
		return func(ttl, batchID uint64) uint64 {
			if batchID > 5 {
				return ttl
			}

			return 0
		}
	case 3:
		return func(ttl, batchID uint64) uint64 {
			if batchID%2 == 1 && batchID <= 10 {
				return 0
			}

			return ttl
		}
	case 4:
		return func(ttl, batchID uint64) uint64 {
			if batchID%3 == 1 && batchID <= 15 {
				return 0
			}

			return ttl
		}
	}
}
