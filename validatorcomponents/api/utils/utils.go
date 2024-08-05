package utils

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/model/response"
)

func WriteErrorResponse(w http.ResponseWriter, status int, err string) error {
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(response.ErrorResponse{Err: err})
}

func WriteUnauthorizedResponse(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusUnauthorized)

	return json.NewEncoder(w).Encode(response.ErrorResponse{Err: "Unauthorized"})
}

type IsRecoverableErrorFn func(err error) bool

var ErrExecutionTimeout = errors.New("timeout while trying to execute with retry")

// ExecuteWithRetry attempts to execute the provided executeFn function multiple times
// if the call fails with a recoverable error. It retries up to numRetries times.
func ExecuteWithRetry(ctx context.Context,
	numRetries int, waitTime time.Duration,
	executeFn func() (bool, error),
	isRecoverableError ...IsRecoverableErrorFn,
) error {
	for count := 0; count < numRetries; count++ {
		stop, err := executeFn()
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
