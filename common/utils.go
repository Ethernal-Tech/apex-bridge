package common

import (
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"net/url"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sethvargo/go-retry"
	"golang.org/x/crypto/sha3"
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
func RetryForever(ctx context.Context, interval time.Duration, fn func(context.Context) error) error {
	err := retry.Do(ctx, retry.NewConstant(interval), func(context.Context) error {
		// Execute function and end retries if no error or context done
		err := fn(ctx)
		if IsContextDoneErr(err) {
			return err
		}

		if err == nil {
			return nil
		}

		// Retry on all other errors
		return retry.RetryableError(err)
	})

	return err
}

// IsContextDoneErr returns true if the error is due to the context being cancelled
// or expired. This is useful for determining if a function should retry.
func IsContextDoneErr(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// SplitString splits large string into slice of substrings
func SplitString(s string, mxlen int) (res []string) {
	for i := 0; i < len(s); i += mxlen {
		end := i + mxlen
		if end > len(s) {
			end = len(s)
		}

		res = append(res, s[i:end])
	}

	return res
}

// MulPercentage multuple value with percentage and divide with 100
func MulPercentage(value *big.Int, percentage uint64) *big.Int {
	res := new(big.Int).Mul(value, new(big.Int).SetUint64(percentage))

	return res.Div(res, big.NewInt(100))
}

// SafeSubtract subtracts safely two uint64 value and return default value if we have overflow
func SafeSubtract(a, b, def uint64) uint64 {
	if a >= b {
		return a - b
	}

	return def
}

// Keccak256 calculates the Keccak256
func Keccak256(v ...[]byte) ([]byte, error) {
	h := sha3.NewLegacyKeccak256()

	for _, i := range v {
		_, err := h.Write(i)
		if err != nil {
			return nil, err
		}
	}

	return h.Sum(nil), nil
}

const (
	DfmDecimals = 6
	WeiDecimals = 18
)

func DfmToWei(dfm *big.Int) *big.Int {
	wei, _ := new(big.Int).SetString(dfm.String(), 10)
	base := big.NewInt(10)

	return wei.Mul(wei, base.Exp(base, big.NewInt(WeiDecimals-DfmDecimals), nil))
}

func WeiToDfm(wei *big.Int) *big.Int {
	dfm, _ := new(big.Int).SetString(wei.String(), 10)
	base := big.NewInt(10)
	dfm.Div(dfm, base.Exp(base, big.NewInt(WeiDecimals-DfmDecimals), nil))

	return dfm
}

func WeiToDfmCeil(wei *big.Int) *big.Int {
	dfm, _ := new(big.Int).SetString(wei.String(), 10)
	base := big.NewInt(10)
	mod := big.NewInt(0)
	dfm.DivMod(dfm, base.Exp(base, big.NewInt(WeiDecimals-DfmDecimals), nil), mod)

	if mod.Cmp(big.NewInt(0)) != 0 {
		dfm.Add(dfm, big.NewInt(1))
	}

	return dfm
}
