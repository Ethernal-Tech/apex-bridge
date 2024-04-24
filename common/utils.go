package common

import (
	"context"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sethvargo/go-retry"
)

func IsValidURL(input string) bool {
	_, err := url.ParseRequestURI(input)
	return err == nil
}

func HexToAddress(s string) common.Address {
	return common.HexToAddress(s)
}

func DecodeHex(s string) ([]byte, error) {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
	}

	return hex.DecodeString(s)
}

func GetRequiredSignaturesForConsensus(cnt uint64) uint64 {
	return (cnt*2 + 2) / 3
}

// the context is cancelled or expired.
func RetryForever(ctx context.Context, interval time.Duration, fn func(context.Context) error) {
	_ = retry.Do(ctx, retry.NewConstant(interval), func(context.Context) error {
		// Execute function and end retries if no error or context done
		err := fn(ctx)
		if err == nil || IsContextDone(err) {
			return nil
		}

		// Retry on all other errors
		return retry.RetryableError(err)
	})
}

// IsContextDone returns true if the error is due to the context being cancelled
// or expired. This is useful for determining if a function should retry.
func IsContextDone(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
