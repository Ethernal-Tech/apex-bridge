package helper

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func WaitFor(t *testing.T, timeout time.Duration, tick time.Duration, fn func() bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(tick)

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for condition")
		case <-ticker.C:
			if fn() {
				return nil
			}
		}
	}
}
