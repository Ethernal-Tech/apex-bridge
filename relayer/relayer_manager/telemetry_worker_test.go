package relayermanager

import (
	"context"
	"errors"
	"math/big"
	"testing"

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

func TestNewTelemetryWorker(t *testing.T) {
	t.Run("skips non reporters", func(t *testing.T) {
		reporter := &balanceReporterMock{addr: "0x1", balance: big.NewInt(1)}

		worker := NewTelemetryWorker(map[string]core.ChainOperations{
			"prime": &chainOperationsMock{},
			"nexus": reporter,
		}, 0, hclog.NewNullLogger())

		require.Len(t, worker.reporters, 1)
		require.Contains(t, worker.reporters, "nexus")
	})

	t.Run("reports every chain that can report the balance", func(t *testing.T) {
		worker := NewTelemetryWorker(map[string]core.ChainOperations{
			"nexus":    &balanceReporterMock{addr: "0x1", balance: big.NewInt(1)},
			"ethereum": &balanceReporterMock{addr: "0x2", balance: big.NewInt(2)},
		}, 0, hclog.NewNullLogger())

		require.Len(t, worker.reporters, 2)
		require.Contains(t, worker.reporters, "nexus")
		require.Contains(t, worker.reporters, "ethereum")
	})
}

func TestTelemetryWorkerExecute(t *testing.T) {
	oneEther, _ := new(big.Int).SetString("1000000000000000000", 10)

	t.Run("caches until the balance changes", func(t *testing.T) {
		reporter := &balanceReporterMock{addr: "0x1", balance: oneEther}
		worker := NewTelemetryWorker(
			map[string]core.ChainOperations{"nexus": reporter}, 0, hclog.NewNullLogger())

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
		worker := NewTelemetryWorker(
			map[string]core.ChainOperations{"nexus": reporter}, 0, hclog.NewNullLogger())

		worker.execute(context.Background())

		reporter.err = errors.New("node unreachable")

		worker.execute(context.Background())

		require.Equal(t, oneEther, worker.latestBalances["nexus"])
	})
}
