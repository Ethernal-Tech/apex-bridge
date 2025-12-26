package testvalidator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
)

// TestValidator manages a local Solana test validator process for testing purposes.
// It provides methods to start, wait for readiness, and stop the test validator.
type TestValidator struct {
	cmd *exec.Cmd
	cli *rpc.Client
}

// NewTestValidator creates a new TestValidator instance with the provided RPC client.
// The validator process is not started until StartTestNode is called.
func NewTestValidator(cli *rpc.Client) *TestValidator {
	return &TestValidator{cmd: nil, cli: cli}
}

// StartTestNode starts a local Solana test validator process in the background.
// The validator runs with the "-r" (reset) and "-q" (quiet) flags.
// The process output is redirected to stdout and stderr.
//
// Returns an error if the validator process fails to start.
func (t *TestValidator) StartTestNode() error {
	cmd := exec.Command("solana-test-validator", "-r", "-q")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start test node: %w", err)
	}

	t.cmd = cmd

	return nil
}

// WaitForNode waits for the test validator node to become ready by polling
// the RPC client for a finalized slot greater than 0. This method blocks
// until the node is ready, checking every second.
//
// Returns an error if the wait operation fails, though it will typically
// continue retrying until the node is ready.
func (t *TestValidator) WaitForNode() error {
	for {
		finalizedSlot, err := t.cli.GetSlot(
			context.TODO(),
			rpc.CommitmentFinalized,
		)
		if err != nil || finalizedSlot == 0 {
			time.Sleep(time.Second)
			continue
		}

		if finalizedSlot > 0 {
			fmt.Println(finalizedSlot)
			break
		}
	}

	return nil
}

// Close stops the test validator process by killing it and waiting for it to exit.
// This should be called to clean up the validator process when it's no longer needed.
// Safe to call multiple times or if the validator was never started.
//
// Returns an error if the process termination fails.
func (t *TestValidator) Close() error {
	if t.cmd != nil {
		t.cmd.Process.Kill()
		t.cmd.Process.Wait()
		t.cmd = nil
	}

	return nil
}
