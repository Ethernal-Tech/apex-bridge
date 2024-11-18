package ethtxhelper

import (
	"errors"
	"net"
	"strings"

	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
)

func IsRetryableEthError(err error) bool {
	// Context was explicitly canceled or deadline exceeded; not retryable
	if infracommon.IsContextDoneErr(err) {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	if errors.Is(err, infracommon.ErrRetryTryAgain) {
		return true
	}

	retryableMessages := []string{
		"replacement tx underpriced",
		"nonce too low",
		"intrinsic gas too low",
		"tx with the same nonce is already present",
		"rejected future tx due to low slots",
	}
	errStr := err.Error()

	for _, msg := range retryableMessages {
		if strings.Contains(errStr, msg) {
			return true
		}
	}

	return false
}
