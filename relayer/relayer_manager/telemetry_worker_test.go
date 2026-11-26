package relayermanager

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

type chainOperationsMock struct{}

func (m *chainOperationsMock) SendTx(
	context.Context, eth.IBridgeSmartContract, *eth.ConfirmedBatch,
) error {
	return nil
}

type balanceReporterMock struct {
	chainOperationsMock
	addr    string
	balance *big.Int
	err     error
	calls   int
}

func (m *balanceReporterMock) GetRelayerAddress() string {
	return m.addr
}

func (m *balanceReporterMock) GetRelayerBalance(context.Context) (*big.Int, error) {
	m.calls++

	return m.balance, m.err
}

const testPullTime = 10 * time.Second

func newTestTelemetryWorker(t *testing.T, operations map[string]core.ChainOperations) *TelemetryWorker {
	t.Helper()

	return NewTelemetryWorker(operations, testPullTime, hclog.NewNullLogger())
}

func TestNewTelemetryWorker(t *testing.T) {
	t.Run("skips non reporters", func(t *testing.T) {
		reporter := &balanceReporterMock{addr: "0x1", balance: big.NewInt(1)}

		worker := newTestTelemetryWorker(t, map[string]core.ChainOperations{
			"prime": &chainOperationsMock{},
			"nexus": reporter,
		})

		require.Len(t, worker.reporters, 1)
		require.Contains(t, worker.reporters, "nexus")
	})

	t.Run("reports every chain that can report the balance", func(t *testing.T) {
		worker := newTestTelemetryWorker(t, map[string]core.ChainOperations{
			"nexus":    &balanceReporterMock{addr: "0x1", balance: big.NewInt(1)},
			"ethereum": &balanceReporterMock{addr: "0x2", balance: big.NewInt(2)},
		})

		require.Len(t, worker.reporters, 2)
		require.Contains(t, worker.reporters, "nexus")
		require.Contains(t, worker.reporters, "ethereum")
	})

	t.Run("uses the configured pull time", func(t *testing.T) {
		worker := NewTelemetryWorker(map[string]core.ChainOperations{}, 15*time.Second, hclog.NewNullLogger())
		require.Equal(t, 15*time.Second, worker.waitTime)
	})
}

func TestTelemetryWorkerExecute(t *testing.T) {
	oneEther, _ := new(big.Int).SetString("1000000000000000000", 10)

	t.Run("caches until the balance changes", func(t *testing.T) {
		reporter := &balanceReporterMock{addr: "0x1", balance: oneEther}
		worker := newTestTelemetryWorker(t, map[string]core.ChainOperations{"nexus": reporter})

		worker.execute(context.Background())
		require.Equal(t, oneEther, worker.latestBalances["nexus"])

		worker.execute(context.Background())
		require.Equal(t, 2, reporter.calls)
		require.Equal(t, oneEther, worker.latestBalances["nexus"])

		reporter.balance = big.NewInt(5)

		worker.execute(context.Background())
		require.Equal(t, big.NewInt(5), worker.latestBalances["nexus"])
	})

	t.Run("keeps the cached value on error", func(t *testing.T) {
		reporter := &balanceReporterMock{addr: "0x1", balance: oneEther}
		worker := newTestTelemetryWorker(t, map[string]core.ChainOperations{"nexus": reporter})

		worker.execute(context.Background())

		reporter.err = errors.New("node unreachable")

		worker.execute(context.Background())

		require.Equal(t, oneEther, worker.latestBalances["nexus"])
	})
}
