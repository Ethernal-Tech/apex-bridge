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
	enabled := map[string]core.ChainConfig{
		"prime":    {Telemetry: true},
		"nexus":    {Telemetry: true},
		"ethereum": {Telemetry: true},
	}

	t.Run("skips non reporters", func(t *testing.T) {
		reporter := &balanceReporterMock{addr: "0x1", balance: big.NewInt(1)}

		worker := NewTelemetryWorker(map[string]core.ChainOperations{
			"prime": &chainOperationsMock{},
			"nexus": reporter,
		}, enabled, 0, hclog.NewNullLogger())

		require.Len(t, worker.reporters, 1)
		require.Contains(t, worker.reporters, "nexus")
	})

	t.Run("skips chains not enabled in the configuration", func(t *testing.T) {
		worker := NewTelemetryWorker(map[string]core.ChainOperations{
			"nexus":    &balanceReporterMock{addr: "0x1", balance: big.NewInt(1)},
			"ethereum": &balanceReporterMock{addr: "0x2", balance: big.NewInt(2)},
		}, map[string]core.ChainConfig{
			"nexus":    {Telemetry: false},
			"ethereum": {Telemetry: true},
		}, 0, hclog.NewNullLogger())

		require.Len(t, worker.reporters, 1)
		require.Contains(t, worker.reporters, "ethereum")
	})

	t.Run("skips chains without a configuration", func(t *testing.T) {
		worker := NewTelemetryWorker(map[string]core.ChainOperations{
			"nexus": &balanceReporterMock{addr: "0x1", balance: big.NewInt(1)},
		}, nil, 0, hclog.NewNullLogger())

		require.Empty(t, worker.reporters)
	})
}

func TestTelemetryWorkerExecute(t *testing.T) {
	oneEther, _ := new(big.Int).SetString("1000000000000000000", 10)

	t.Run("caches until the balance changes", func(t *testing.T) {
		reporter := &balanceReporterMock{addr: "0x1", balance: oneEther}
		worker := NewTelemetryWorker(
			map[string]core.ChainOperations{"nexus": reporter},
			map[string]core.ChainConfig{"nexus": {Telemetry: true}}, 0, hclog.NewNullLogger())

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
			map[string]core.ChainOperations{"nexus": reporter},
			map[string]core.ChainConfig{"nexus": {Telemetry: true}}, 0, hclog.NewNullLogger())

		worker.execute(context.Background())

		reporter.err = errors.New("node unreachable")

		worker.execute(context.Background())

		require.Equal(t, oneEther, worker.latestBalances["nexus"])
	})
}
