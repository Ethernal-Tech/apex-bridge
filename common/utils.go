package common

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unsafe"

	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer/gouroboros"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/sethvargo/go-retry"
	"golang.org/x/crypto/sha3"
	"golang.org/x/exp/constraints"
)

const (
	waitForAmountRetryCount = 144 // 144 * 5 = 12 min
	waitForAmountWaitTime   = time.Second * 5
)

func IsValidHTTPURL(input string) bool {
	u, err := url.Parse(input)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return false
	}

	return IsValidNetworkAddress(u.Host)
}

func IsValidNetworkAddress(input string) bool {
	host, port, err := net.SplitHostPort(input)
	if err != nil {
		// If there's an error, it might be because the port is not included, so treat the input as the host
		if !strings.Contains(err.Error(), "missing port in address") {
			return false
		}

		host = input
	} else if portVal, err := strconv.ParseInt(port, 10, 32); err != nil || portVal < 0 {
		return false
	}

	// Check if host is a valid IP address
	if net.ParseIP(host) != nil {
		return true
	}

	// Check if the host is a valid domain name by trying to resolve it
	_, err = net.LookupHost(host)

	return err == nil
}

func HexToAddress(s string) ethcommon.Address {
	return ethcommon.HexToAddress(s)
}

func DecodeHex(s string) ([]byte, error) {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
	}

	return hex.DecodeString(s)
}

func GetRequiredSignaturesForConsensus(cnt uint64) uint64 {
	return cnt*2/3 + 1
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
		res = append(res, s[i:min(i+mxlen, len(s))])
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
	wei := new(big.Int).Set(dfm)
	base := big.NewInt(10)

	return wei.Mul(wei, base.Exp(base, big.NewInt(WeiDecimals-DfmDecimals), nil))
}

func WeiToDfm(wei *big.Int) *big.Int {
	dfm := new(big.Int).Set(wei)
	base := big.NewInt(10)
	dfm.Div(dfm, base.Exp(base, big.NewInt(WeiDecimals-DfmDecimals), nil))

	return dfm
}

func WeiToDfmCeil(wei *big.Int) *big.Int {
	dfm := new(big.Int).Set(wei)
	base := big.NewInt(10)
	mod := new(big.Int)
	dfm.DivMod(dfm, base.Exp(base, big.NewInt(WeiDecimals-DfmDecimals), nil), mod)

	if mod.BitLen() > 0 { // for zero big.Int BitLen() == 0
		dfm.Add(dfm, big.NewInt(1))
	}

	return dfm
}

type IsRecoverableErrorFn func(err error) bool

var ErrExecutionTimeout = errors.New("timeout while trying to execute with retry")

// ExecuteWithRetry attempts to execute the provided executeFn function multiple times
// if the call fails with a recoverable error. It retries up to numRetries times.
func ExecuteWithRetry(ctx context.Context,
	numRetries int, waitTime time.Duration,
	executeFn func(ctx context.Context) (bool, error),
	isRecoverableError ...IsRecoverableErrorFn,
) error {
	for count := 0; count < numRetries; count++ {
		stop, err := executeFn(ctx)
		if err != nil {
			if len(isRecoverableError) == 0 || !isRecoverableError[0](err) {
				return err
			}
		} else if stop {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
		}
	}

	return ErrExecutionTimeout
}

func ExecuteCLICommand(binary string, args []string, workingDir string) (string, error) {
	var (
		stdErrBuffer bytes.Buffer
		stdOutBuffer bytes.Buffer
	)

	cmd := exec.Command(binary, args...)
	cmd.Stderr = &stdErrBuffer
	cmd.Stdout = &stdOutBuffer
	cmd.Dir = workingDir

	err := cmd.Run()

	if stdErrBuffer.Len() > 0 {
		return "", fmt.Errorf("error while executing command: %s", stdErrBuffer.String())
	} else if err != nil {
		return "", err
	}

	return stdOutBuffer.String(), nil
}

func WaitForAmount(
	ctx context.Context, receivedAmount *big.Int, getBalanceFn func(ctx context.Context) (*big.Int, error),
) (*big.Int, error) {
	originalAmount, err := infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (*big.Int, error) {
		return getBalanceFn(ctx)
	})
	if err != nil {
		return nil, err
	}

	expectedBalance := originalAmount.Add(originalAmount, receivedAmount)

	return infracommon.ExecuteWithRetry(ctx, func(ctx context.Context) (*big.Int, error) {
		balance, err := getBalanceFn(ctx)
		if err != nil {
			return nil, err
		}

		if balance.Cmp(expectedBalance) < 0 {
			return balance, infracommon.ErrRetryTryAgain
		}

		return balance, nil
	}, infracommon.WithRetryCount(waitForAmountRetryCount), infracommon.WithRetryWaitTime(waitForAmountWaitTime))
}

func IsValidAddress(chainID string, addr string) bool {
	switch chainID {
	case ChainIDStrNexus:
		return ethcommon.IsHexAddress(addr)
	default:
		addr, err := cardanowallet.NewCardanoAddressFromString(addr)

		return err == nil && addr.GetInfo().AddressType != cardanowallet.RewardAddress
	}
}

func GetDfmAmount(chainID string, amount *big.Int) *big.Int {
	switch chainID {
	case ChainIDStrNexus:
		return WeiToDfm(amount)
	default:
		return amount
	}
}

func PackNumbersToBytes[Slice ~[]T, T constraints.Integer | constraints.Float](nums Slice) []byte {
	var zero T

	buf := new(bytes.Buffer)
	// Preallocate needed bytes
	buf.Grow(len(nums) * int(unsafe.Sizeof(zero)))

	for _, v := range nums {
		_ = binary.Write(buf, binary.LittleEndian, v) // error can not be thrown
	}

	return buf.Bytes()
}

func UnpackNumbersToBytes[Slice ~[]T, T constraints.Integer | constraints.Float](packedBytes []byte) (Slice, error) {
	var value T

	buf := bytes.NewReader(packedBytes)
	sizeInBytes := int(unsafe.Sizeof(value))
	result := make([]T, len(packedBytes)/sizeInBytes)

	for i := range result {
		if err := binary.Read(buf, binary.LittleEndian, &value); err != nil {
			return nil, err
		}

		result[i] = value
	}

	return result, nil
}

func NumbersToString[Slice ~[]T, T constraints.Integer | constraints.Float](nums Slice) string {
	var sb strings.Builder

	for i, x := range nums {
		if i > 0 {
			sb.WriteString(", ")
		}

		sb.WriteString(fmt.Sprint(x))
	}

	return sb.String()
}

func ParseTxInfo(txRaw []byte, full bool) (indexer.TxInfo, error) {
	return gouroboros.ParseTxInfo(txRaw, full)
}
