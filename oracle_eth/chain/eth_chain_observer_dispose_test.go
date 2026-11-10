package chain

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const zeroHash = "0x0000000000000000000000000000000000000000000000000000000000000000"

// testBlockJSON is a block, as a node returns it without transactions.
func testBlockJSON(number uint64) string {
	return fmt.Sprintf(`{
		"number": "0x%x", "hash": "%s", "parentHash": "%s", "sha3Uncles": "%s",
		"transactionsRoot": "%s", "stateRoot": "%s", "receiptsRoot": "%s", "mixHash": "%s",
		"miner": "0x0000000000000000000000000000000000000000", "difficulty": "0x0",
		"gasLimit": "0x1c9c380", "gasUsed": "0x0", "timestamp": "0x6553f100",
		"extraData": "0x", "nonce": "0x0000000000000000", "baseFeePerGas": "0x7",
		"transactions": [], "uncles": []
	}`, number, zeroHash, zeroHash, zeroHash, zeroHash, zeroHash, zeroHash, zeroHash)
}

// newTestRPCServer answers eth_chainId, and hands every other request to handleOther.
// It counts the requests other than eth_chainId.
func newTestRPCServer(t *testing.T, handleOther func(w http.ResponseWriter, method string)) (
	*httptest.Server, *atomic.Int64,
) {
	t.Helper()

	var requests atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		var request struct {
			Method string `json:"method"`
		}

		_ = json.Unmarshal(body, &request)

		if request.Method == "eth_chainId" {
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x1"}`))

			return
		}

		requests.Add(1)
		handleOther(w, request.Method)
	}))

	return server, &requests
}

func newDisposeTestObserver(t *testing.T, nodeURL string) *EthChainObserverImpl {
	t.Helper()

	config := &oCore.EthChainConfig{
		ChainID:                 "nexus",
		NodeURL:                 nodeURL,
		PoolIntervalMiliseconds: 50,
		SyncBatchSize:           10,
		StartBlockNumber:        1,
		RestartTrackerPullCheck: time.Hour,
	}

	indexerDB, err := eventTrackerStore.NewBoltDBEventTrackerStore(filepath.Join(t.TempDir(), "nexus.db"))
	require.NoError(t, err)

	oracleDB := &core.EthTxsProcessorDBMock{}
	oracleDB.On("ClearAllTxs", mock.Anything).Return(error(nil))

	chainObserver, err := NewEthChainObserver(
		config, &core.EthTxsReceiverMock{}, oracleDB, indexerDB, hclog.NewNullLogger(), hclog.NewNullLogger())
	require.NoError(t, err)

	return chainObserver
}

func TestEthChainObserver_DisposeWaitsForTracker(t *testing.T) {
	t.Run("waits until the tracker has stopped", func(t *testing.T) {
		server, requests := newTestRPCServer(t, func(w http.ResponseWriter, method string) {
			switch method {
			case "eth_getBlockByNumber":
				_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":` + testBlockJSON(10) + `}`))
			default:
				_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":[]}`))
			}
		})
		defer server.Close()

		chainObserver := newDisposeTestObserver(t, server.URL)
		require.NoError(t, chainObserver.Start())

		require.Eventually(t, func() bool { return requests.Load() > 0 }, 5*time.Second, 10*time.Millisecond)

		require.NoError(t, chainObserver.Dispose())

		select {
		case <-chainObserver.stoppedCh:
		default:
			t.Fatal("Dispose returned before the tracker stopped")
		}

		// the tracker sends no more requests once Dispose has returned
		requestsAfterDispose := requests.Load()

		time.Sleep(300 * time.Millisecond)
		require.Equal(t, requestsAfterDispose, requests.Load())
	})

	t.Run("returns an error when the tracker does not stop in time", func(t *testing.T) {
		release := make(chan struct{})
		requestSeen := make(chan struct{}, 1)

		server, _ := newTestRPCServer(t, func(http.ResponseWriter, string) {
			select {
			case requestSeen <- struct{}{}:
			default:
			}

			<-release
		})

		defer func() {
			close(release)
			server.Close()
		}()

		chainObserver := newDisposeTestObserver(t, server.URL)
		chainObserver.stopTimeout = 300 * time.Millisecond

		require.NoError(t, chainObserver.Start())

		select {
		case <-requestSeen:
		case <-time.After(5 * time.Second):
			t.Fatal("tracker never sent a request")
		}

		started := time.Now().UTC()

		require.ErrorContains(t, chainObserver.Dispose(), "event tracker for chain nexus did not stop within 300ms")
		require.Less(t, time.Since(started), 5*time.Second)
	})

	t.Run("returns right away when the observer was never started", func(t *testing.T) {
		chainObserver := newDisposeTestObserver(t, "http://127.0.0.1:1")

		require.NoError(t, chainObserver.Dispose())
	})
}
