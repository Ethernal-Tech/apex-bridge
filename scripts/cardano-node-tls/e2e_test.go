//go:build e2e

package cardanonodetls

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer/gouroboros"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

const (
	serverName   = "node.test"
	networkMagic = 1
	pointSlot    = uint64(2607239)
	pointHash    = "34c36a9eb7228ca529e91babcf2215be29ce2a65b609540b483abc4520848d19"
	readTimeout  = 10 * time.Second
	syncTimeout  = 60 * time.Second
)

type env struct {
	addr       string
	dir        string
	foreignDir string
}

func loadEnv(t *testing.T) env {
	t.Helper()

	e := env{
		addr:       os.Getenv("E2E_NODE_ADDR"),
		dir:        os.Getenv("E2E_TLS_DIR"),
		foreignDir: os.Getenv("E2E_FOREIGN_TLS_DIR"),
	}
	if e.addr == "" || e.dir == "" || e.foreignDir == "" {
		t.Skip("E2E_NODE_ADDR, E2E_TLS_DIR and E2E_FOREIGN_TLS_DIR must be set (run e2e.sh)")
	}

	return e
}

func caPool(t *testing.T, dir string) *x509.CertPool {
	t.Helper()

	pem, err := os.ReadFile(filepath.Join(dir, "ca.crt"))
	require.NoError(t, err)

	pool := x509.NewCertPool()
	require.True(t, pool.AppendCertsFromPEM(pem))

	return pool
}

func loadPair(t *testing.T, dir string) tls.Certificate {
	t.Helper()

	pair, err := tls.LoadX509KeyPair(filepath.Join(dir, "validator1.crt"), filepath.Join(dir, "validator1.key"))
	require.NoError(t, err)

	return pair
}

// readAfterHandshake dials addr, runs the TLS handshake and reads once. In TLS 1.3 the server
// checks the client certificate after the client finished its handshake, so a rejection
// is visible on the first read. It returns the first error of handshake or read.
func readAfterHandshake(t *testing.T, addr string, cfg *tls.Config) error {
	t.Helper()

	raw, err := net.DialTimeout("tcp", addr, readTimeout)
	require.NoError(t, err)

	conn := tls.Client(raw, cfg)

	defer conn.Close()

	require.NoError(t, conn.SetDeadline(time.Now().UTC().Add(readTimeout)))

	if err := conn.Handshake(); err != nil {
		return err
	}

	_, err = conn.Read(make([]byte, 1))

	return err
}

func baseConfig(roots *x509.CertPool) *tls.Config {
	return &tls.Config{
		RootCAs:    roots,
		ServerName: serverName,
		MinVersion: tls.VersionTLS13,
	}
}

func TestNodeTLSE2E_SyncsBlock(t *testing.T) {
	e := loadEnv(t)

	blocks := make(chan indexer.BlockHeader, 1)
	handler := indexer.NewBlockSyncerHandlerMock(pointSlot, pointHash)
	handler.RollForwardFn = func(h indexer.BlockHeader, _ indexer.BlockTxsRetriever) error {
		select {
		case blocks <- h:
		default:
		}

		return nil
	}

	syncer := gouroboros.NewBlockSyncer(&gouroboros.BlockSyncerConfig{
		NetworkMagic: networkMagic,
		NodeAddress:  e.addr,
		RestartDelay: 100 * time.Millisecond,
		TLS: &gouroboros.TLSConfig{
			CAFile:     filepath.Join(e.dir, "ca.crt"),
			CertFile:   filepath.Join(e.dir, "validator1.crt"),
			KeyFile:    filepath.Join(e.dir, "validator1.key"),
			ServerName: serverName,
		},
	}, handler, hclog.NewNullLogger())

	t.Cleanup(func() { _ = syncer.Close() })

	require.NoError(t, syncer.Sync())

	select {
	case h := <-blocks:
		require.Greater(t, h.Slot, pointSlot)
	case err := <-syncer.ErrorCh():
		t.Fatalf("syncer error while waiting for a block from preprod node via %s: %v", e.addr, err)
	case <-time.After(syncTimeout):
		t.Fatalf("no block from the preprod node (preprod-node.play.dev.cardano.org:3001) via %s within %s",
			e.addr, syncTimeout)
	}
}

func TestNodeTLSE2E_RejectsNoClientCert(t *testing.T) {
	e := loadEnv(t)

	err := readAfterHandshake(t, e.addr, baseConfig(caPool(t, e.dir)))
	require.Error(t, err)
	t.Logf("rejected with: %v", err)
	require.NotErrorIs(t, err, os.ErrDeadlineExceeded)
}

func TestNodeTLSE2E_RejectsForeignClientCert(t *testing.T) {
	e := loadEnv(t)

	foreign := loadPair(t, e.foreignDir)
	cfg := baseConfig(caPool(t, e.dir))
	// Go does not send a certificate whose issuer is not in the server's CA list, force it
	cfg.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
		return &foreign, nil
	}

	err := readAfterHandshake(t, e.addr, cfg)
	require.Error(t, err)
	t.Logf("rejected with: %v", err)
	require.NotErrorIs(t, err, os.ErrDeadlineExceeded)
}

func TestNodeTLSE2E_RejectsServerFromOtherCA(t *testing.T) {
	e := loadEnv(t)

	cfg := baseConfig(caPool(t, e.foreignDir))
	pair := loadPair(t, e.dir)
	cfg.Certificates = []tls.Certificate{pair}

	err := readAfterHandshake(t, e.addr, cfg)

	var unknown x509.UnknownAuthorityError

	require.ErrorAs(t, err, &unknown)
}

func TestNodeTLSE2E_RejectsPlainTCP(t *testing.T) {
	e := loadEnv(t)

	conn, err := net.DialTimeout("tcp", e.addr, readTimeout)
	require.NoError(t, err)

	defer conn.Close()

	require.NoError(t, conn.SetDeadline(time.Now().UTC().Add(readTimeout)))

	_, _ = conn.Write([]byte("not a tls client hello\r\n"))

	buf := make([]byte, 512)

	for {
		_, err = conn.Read(buf)
		if err != nil {
			break
		}
	}

	require.Error(t, err)
	t.Logf("rejected with: %v", err)
	require.NotErrorIs(t, err, os.ErrDeadlineExceeded)
}
