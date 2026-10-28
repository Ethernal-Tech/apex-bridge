package chain

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	indexerDb "github.com/Ethernal-Tech/cardano-infrastructure/indexer/db"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer/gouroboros"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockSyncer struct {
	mock.Mock
	errCh chan error
}

func (m *MockSyncer) Sync() error {
	return m.Called().Error(0)
}

func (m *MockSyncer) Close() error {
	return m.Called().Error(0)
}

func (m *MockSyncer) ErrorCh() <-chan error {
	return m.errCh
}

type MockIndexerDB struct {
	mock.Mock
	indexer.Database
}

func (m *MockIndexerDB) GetLatestBlockPoint() (*indexer.BlockPoint, error) {
	args := m.Called()

	return args.Get(0).(*indexer.BlockPoint), args.Error(1)
}

func (m *MockIndexerDB) Close() error {
	return m.Called().Error(0)
}

func TestCardanoChainObserver(t *testing.T) {
	testDir, err := os.MkdirTemp("", "boltdb-test")
	require.NoError(t, err)

	foldersCleanup := func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}

	defer foldersCleanup()

	settings := cCore.AppSettings{
		DbsPath: testDir,
		Logger: logger.LoggerConfig{
			LogFilePath: testDir,
		},
	}

	chainConfig := &cCore.CardanoChainConfig{
		ChainID:        common.ChainIDStrPrime,
		NetworkAddress: "backbone.cardano.iog.io:3001",
		CardanoChainConfig: cardanotx.CardanoChainConfig{
			NetworkMagic: 764824073,
		},
		StartBlockHash:         "335ac2d90bc37906c1264cfdc5769a652293cf64fa42b0c74d323473938b8ff1",
		StartSlot:              127933773,
		ConfirmationBlockCount: 10,
		OtherAddressesOfInterest: []string{
			"addr1v9kganeshgdqyhwnyn9stxxgl7r4y2ejfyqjn88n7ncapvs4sugsd",
		},
	}

	txsReceiverMock := &core.CardanoTxsReceiverMock{}
	txsReceiverMock.On("NewUnprocessedTxs", mock.Anything, mock.Anything).Return(error(nil))

	initDB := func(t *testing.T) indexer.Database {
		t.Helper()

		require.NoError(t, common.CreateDirectoryIfNotExists(settings.DbsPath, 0750))

		indexerDB, err := indexerDb.NewDatabaseInit("", filepath.Join(settings.DbsPath, chainConfig.ChainID+".db"))
		require.NoError(t, err)

		return indexerDB
	}

	t.Run("check ErrorCh", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		db := &core.CardanoTxsProcessorDBMock{}
		db.On("ClearAllTxs", mock.Anything).Return(error(nil))
		db.On("MoveProcessedExpectedTxs", mock.Anything).Return(error(nil))

		indexerDB := initDB(t)

		chainObserver, err := NewCardanoChainObserver(context.Background(), chainConfig, txsReceiverMock, db, indexerDB, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		defer chainObserver.Dispose() //nolint:errcheck

		errChan := chainObserver.ErrorCh()
		require.NotNil(t, errChan)
	})

	t.Run("check GetConfig", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		db := &core.CardanoTxsProcessorDBMock{}
		db.On("ClearAllTxs", mock.Anything).Return(error(nil))
		db.On("MoveProcessedExpectedTxs", mock.Anything).Return(error(nil))

		indexerDB := initDB(t)

		chainObserver, err := NewCardanoChainObserver(context.Background(), chainConfig, txsReceiverMock, db, indexerDB, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		config := chainObserver.GetConfig()
		require.NotNil(t, config)
		require.Equal(t, chainConfig, config)
	})

	t.Run("check start stop", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		db := &core.CardanoTxsProcessorDBMock{}
		db.On("ClearAllTxs", mock.Anything).Return(error(nil))
		db.On("MoveProcessedExpectedTxs", mock.Anything).Return(error(nil))

		indexerDB := initDB(t)

		var logBuffer bytes.Buffer
		logger := hclog.New(&hclog.LoggerOptions{
			Name:       "test",
			Level:      hclog.Debug,
			Output:     &logBuffer,
			JSONFormat: false,
		})

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		chainObserver, err := NewCardanoChainObserver(ctx, chainConfig, txsReceiverMock, db, indexerDB, logger)
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		err = chainObserver.Start()
		require.NoError(t, err)

		_ = chainObserver.Dispose()

		output := logBuffer.String()
		require.NotContains(t, output, "Failed to close syncer")
		require.NotContains(t, output, "Failed to close indexerDB")
	})

	t.Run("check newConfirmedTxs called", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		db := &core.CardanoTxsProcessorDBMock{}
		db.On("ClearAllTxs", mock.Anything).Return(error(nil))
		db.On("MoveProcessedExpectedTxs", mock.Anything).Return(error(nil))

		indexerDB := initDB(t)

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		chainObserver, err := NewCardanoChainObserver(ctx, chainConfig, txsReceiverMock, db, indexerDB, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		doneCh := make(chan bool, 1)
		closed := uint32(0)

		txsReceiverMock.NewUnprocessedTxsFn = func(originChainId string, txs []*indexer.Tx) error {
			if atomic.CompareAndSwapUint32(&closed, 0, 1) {
				close(doneCh)
			}

			return nil
		}

		err = chainObserver.Start()
		require.NoError(t, err)

		timer := time.NewTimer(60 * time.Second)
		defer timer.Stop()

		select {
		case <-timer.C:
			t.Fatal("timeout")
		case <-doneCh:
		}
	})

	t.Run("errorCh fatal error", func(t *testing.T) {
		syncerMock := &MockSyncer{}
		syncerMock.errCh = make(chan error, 1)

		syncerMock.errCh <- indexer.ErrBlockIndexerFatal

		select {
		case err := <-syncerMock.ErrorCh():
			require.Error(t, err)
			require.ErrorIs(t, err, indexer.ErrBlockIndexerFatal)
		case <-time.After(time.Second):
			t.Fatal("Expected error not received from ErrorCh")
		}
	})

	t.Run("check close called", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		db := &core.CardanoTxsProcessorDBMock{}
		db.On("ClearAllTxs", mock.Anything).Return(error(nil))
		db.On("MoveProcessedExpectedTxs", mock.Anything).Return(error(nil))

		syncer := &MockSyncer{}
		indexerDB := &MockIndexerDB{}

		testErr := indexer.ErrBlockIndexerFatal

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		syncer.errCh = make(chan error, 1)
		syncerDisposeCalled := make(chan struct{})
		indexerDBDisposeCalled := make(chan struct{})

		syncer.On("Sync").Return(nil)
		syncer.On("Close").Run(func(args mock.Arguments) {
			close(syncerDisposeCalled)
		}).Return(nil)

		indexerDB.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{}, nil).Once()
		indexerDB.On("Close").Run(func(args mock.Arguments) {
			close(indexerDBDisposeCalled)
		}).Return(nil)

		chainObserver := &CardanoChainObserverImpl{
			ctx:       ctx,
			indexerDB: indexerDB,
			syncer:    syncer,
			logger:    hclog.NewNullLogger(),
			config:    &cCore.CardanoChainConfig{ChainID: "test"},
		}

		err := chainObserver.Start()
		require.NoError(t, err)

		syncer.errCh <- testErr

		select {
		case <-syncerDisposeCalled:
			syncer.AssertCalled(t, "Close")
		case <-time.After(time.Second):
			t.Fatal("Syncer was not closed")
		}

		select {
		case <-indexerDBDisposeCalled:
			indexerDB.AssertCalled(t, "Close")
		case <-time.After(time.Second):
			t.Fatal("IndexerDB was not closed")
		}
	})
}

func Test_initOracleState(t *testing.T) {
	dbMock := &indexer.DatabaseMock{}
	dbWriterMock := &indexer.DBTransactionWriterMock{}
	utxos := []cCore.CardanoChainConfigUtxo{
		{
			Hash:    indexer.NewHashFromHexString("0x1"),
			Index:   2,
			Address: "0xffaa",
			Amount:  uint64(200),
		},
		{
			Hash:    indexer.NewHashFromHexString("0x2"),
			Index:   1,
			Address: "0xff03",
			Amount:  uint64(500),
		},
	}
	blockSlot := uint64(100)
	oracleDBMock := &core.CardanoTxsProcessorDBMock{}
	oracleDBMock.On("MoveProcessedExpectedTxs", mock.Anything).Return(error(nil))
	blockHash := "0xA1"
	chainID := "prime"

	t.Run("error", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{}, errors.New("test error")).Once()

		require.Error(t, initOracleState(dbMock, oracleDBMock, blockHash, blockSlot, utxos, chainID, hclog.NewNullLogger()))
	})

	t.Run("empty hash", func(t *testing.T) {
		require.NoError(t, initOracleState(dbMock, oracleDBMock, "", blockSlot, utxos, chainID, hclog.NewNullLogger()))
	})

	t.Run("updated in db smaller slot", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return((*indexer.BlockPoint)(nil), error(nil)).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{
			BlockSlot: uint64(50),
		}, error(nil)).Once()

		dbMock.On("OpenTx").Return(dbWriterMock).Twice()
		dbWriterMock.On("DeleteAllTxOutputsPhysically").Return(dbMock).Twice()
		dbWriterMock.On("AddTxOutputs", convertUtxos(utxos)).Return(dbMock).Twice()
		dbWriterMock.On("SetLatestBlockPoint", &indexer.BlockPoint{
			BlockSlot: blockSlot,
			BlockHash: indexer.NewHashFromHexString(blockHash),
		}).Return(dbMock).Twice()
		dbWriterMock.On("Execute").Return(error(nil)).Twice()
		oracleDBMock.On("ClearAllTxs", chainID).Return(error(nil)).Twice()

		// empty
		require.NoError(t, initOracleState(dbMock, oracleDBMock, blockHash, blockSlot, utxos, chainID, hclog.NewNullLogger()))
		// smaller
		require.NoError(t, initOracleState(dbMock, oracleDBMock, blockHash, blockSlot, utxos, chainID, hclog.NewNullLogger()))
	})

	t.Run("skipping - has equal slot", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{
			BlockSlot: uint64(100),
		}, error(nil)).Once()

		require.NoError(t, initOracleState(dbMock, oracleDBMock, blockHash, blockSlot, utxos, chainID, hclog.NewNullLogger()))
	})

	t.Run("skipping - has greater slot", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{
			BlockSlot: uint64(150),
		}, error(nil)).Once()

		require.NoError(t, initOracleState(dbMock, oracleDBMock, blockHash, blockSlot, utxos, chainID, hclog.NewNullLogger()))
	})

	t.Run("ClearAllTxs error", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return((*indexer.BlockPoint)(nil), error(nil)).Once()
		oracleDBMock.On("ClearAllTxs", chainID).Return(errors.New("test error")).Once()

		require.Error(t, initOracleState(dbMock, oracleDBMock, blockHash, blockSlot, utxos, chainID, hclog.NewNullLogger()))
	})
}

func Test_convertUtxos(t *testing.T) {
	utxos := []cCore.CardanoChainConfigUtxo{
		{
			Hash:    indexer.NewHashFromHexString("0x2"),
			Index:   2,
			Address: "0xffaa",
			Amount:  uint64(200),
			Slot:    34,
		},
		{
			Hash:    indexer.NewHashFromHexString("0x1"),
			Index:   1,
			Address: "0xff03",
			Amount:  uint64(500),
			Slot:    196,
		},
	}

	require.Equal(t, []*indexer.TxInputOutput{
		{
			Input: indexer.TxInput{
				Hash:  indexer.NewHashFromHexString("0x2"),
				Index: 2,
			},
			Output: indexer.TxOutput{
				Address: "0xffaa",
				Amount:  200,
				Slot:    34,
			},
		},
		{
			Input: indexer.TxInput{
				Hash:  indexer.NewHashFromHexString("0x1"),
				Index: 1,
			},
			Output: indexer.TxOutput{
				Address: "0xff03",
				Amount:  500,
				Slot:    196,
			},
		},
	}, convertUtxos(utxos))
}

type testTLSFiles struct {
	caFile, certFile, keyFile string
}

func newTestCert(t *testing.T, isCA bool, parent *x509.Certificate, parentKey *ecdsa.PrivateKey) (
	[]byte, []byte, *x509.Certificate, *ecdsa.PrivateKey,
) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	serial, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(serial.Int64()),
		Subject:               pkix.Name{CommonName: "test"},
		NotBefore:             time.Now().UTC().Add(-time.Hour),
		NotAfter:              time.Now().UTC().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  isCA,
	}
	if isCA {
		tmpl.KeyUsage |= x509.KeyUsageCertSign
	}

	signer, signerKey := tmpl, key
	if parent != nil {
		signer, signerKey = parent, parentKey
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, signer, &key.PublicKey, signerKey)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}), cert, key
}

func writeTestFile(t *testing.T, dir, name string, data []byte) string {
	t.Helper()

	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, data, 0o600))

	return path
}

// newTestTLSFiles generates a CA and a client certificate/key signed by it in a temp dir.
func newTestTLSFiles(t *testing.T) testTLSFiles {
	t.Helper()

	dir := t.TempDir()
	caPEM, _, caCert, caKey := newTestCert(t, true, nil, nil)
	certPEM, keyPEM, _, _ := newTestCert(t, false, caCert, caKey)

	return testTLSFiles{
		caFile:   writeTestFile(t, dir, "ca.crt", caPEM),
		certFile: writeTestFile(t, dir, "client.crt", certPEM),
		keyFile:  writeTestFile(t, dir, "client.key", keyPEM),
	}
}

func Test_loadSyncerConfigs_NetworkTLS(t *testing.T) {
	t.Parallel()

	t.Run("networkTls is passed to the syncer", func(t *testing.T) {
		t.Parallel()

		tlsCfg := &gouroboros.TLSConfig{
			CAFile: "/a/ca.crt", CertFile: "/a/c.crt", KeyFile: "/a/c.key", ServerName: "node.example",
		}
		_, syncerConfig := loadSyncerConfigs(&cCore.CardanoChainConfig{
			NetworkAddress: "10.0.0.1:3001",
			NetworkTLS:     tlsCfg,
		})

		require.Equal(t, tlsCfg, syncerConfig.TLS)
	})

	t.Run("no networkTls", func(t *testing.T) {
		t.Parallel()

		_, syncerConfig := loadSyncerConfigs(&cCore.CardanoChainConfig{NetworkAddress: "10.0.0.1:3001"})

		require.Nil(t, syncerConfig.TLS)
	})
}

func TestCardanoChainConfig_NetworkTLSOmitted(t *testing.T) {
	t.Parallel()

	t.Run("old config round trips without the key", func(t *testing.T) {
		t.Parallel()

		var cfg cCore.CardanoChainConfig

		require.NoError(t, json.Unmarshal([]byte(`{"networkAddress":"10.0.0.1:3001"}`), &cfg))
		require.Nil(t, cfg.NetworkTLS)

		out, err := json.Marshal(cfg)
		require.NoError(t, err)
		require.NotContains(t, string(out), "networkTls")
	})

	t.Run("null means no tls", func(t *testing.T) {
		t.Parallel()

		var cfg cCore.CardanoChainConfig

		require.NoError(t, json.Unmarshal([]byte(`{"networkAddress":"10.0.0.1:3001","networkTls":null}`), &cfg))
		require.Nil(t, cfg.NetworkTLS)
	})

	t.Run("empty object fails at startup", func(t *testing.T) {
		t.Parallel()

		var cfg cCore.CardanoChainConfig

		require.NoError(t, json.Unmarshal([]byte(`{"networkAddress":"10.0.0.1:3001","networkTls":{}}`), &cfg))
		require.NotNil(t, cfg.NetworkTLS)

		cfg.ChainID = common.ChainIDStrPrime
		db := &core.CardanoTxsProcessorDBMock{}

		_, err := NewCardanoChainObserver(context.Background(), &cfg, &core.CardanoTxsReceiverMock{}, db,
			&MockIndexerDB{}, hclog.NewNullLogger())
		require.ErrorContains(t, err, "tls caFile")
		db.AssertNotCalled(t, "MoveProcessedExpectedTxs", mock.Anything)
	})
}

func Test_networkTLSDefaultServerName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		address string
		tlsName string
		want    string
		wantErr string
	}{
		{name: "https prefix", address: "https://10.0.0.1:3001", want: "10.0.0.1"},
		{name: "http prefix", address: "http://10.0.0.1:3001", want: "10.0.0.1"},
		{name: "ipv6", address: "[::1]:3001", want: "::1"},
		{name: "host name", address: "node.example:3001", want: "node.example"},
		{name: "explicit server name is not parsed", address: "node.example", tlsName: "x.example", want: ""},
		{name: "no port", address: "node.example", wantErr: "networkAddress"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := networkTLSDefaultServerName(&cCore.CardanoChainConfig{
				NetworkAddress: tc.address,
				NetworkTLS:     &gouroboros.TLSConfig{ServerName: tc.tlsName},
			})
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestCardanoChainObserver_InvalidNetworkTLS(t *testing.T) {
	t.Parallel()

	newObserver := func(t *testing.T, address string, tlsCfg *gouroboros.TLSConfig) (
		*CardanoChainObserverImpl, *core.CardanoTxsProcessorDBMock, error,
	) {
		t.Helper()

		dir := t.TempDir()
		indexerDB, err := indexerDb.NewDatabaseInit("", filepath.Join(dir, "prime.db"))
		require.NoError(t, err)

		t.Cleanup(func() { _ = indexerDB.Close() })

		db := &core.CardanoTxsProcessorDBMock{}
		db.On("MoveProcessedExpectedTxs", mock.Anything).Return(error(nil))

		obs, err := NewCardanoChainObserver(context.Background(), &cCore.CardanoChainConfig{
			ChainID:        common.ChainIDStrPrime,
			NetworkAddress: address,
			NetworkTLS:     tlsCfg,
		}, &core.CardanoTxsReceiverMock{}, db, indexerDB, hclog.NewNullLogger())

		return obs, db, err
	}

	t.Run("missing ca file", func(t *testing.T) {
		t.Parallel()

		files := newTestTLSFiles(t)
		_, db, err := newObserver(t, "10.0.0.1:3001", &gouroboros.TLSConfig{
			CAFile: filepath.Join(t.TempDir(), "nope.crt"), CertFile: files.certFile, KeyFile: files.keyFile,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid networkTls for chain prime: ")
		require.Contains(t, err.Error(), "tls caFile")
		db.AssertNotCalled(t, "MoveProcessedExpectedTxs", mock.Anything)
	})

	for _, tc := range []struct{ name, address string }{
		{"unix socket address", "/ipc/node.socket"},
		{"unix socket address after scheme trimming", "http:///ipc/node.socket"},
	} {
		for _, serverName := range []string{"node.example", ""} {
			t.Run(tc.name+" serverName="+serverName, func(t *testing.T) {
				t.Parallel()

				files := newTestTLSFiles(t)
				_, db, err := newObserver(t, tc.address, &gouroboros.TLSConfig{
					CAFile: files.caFile, CertFile: files.certFile, KeyFile: files.keyFile, ServerName: serverName,
				})
				require.ErrorContains(t, err, "invalid networkTls for chain prime")

				if serverName != "" {
					require.ErrorContains(t, err, "unix socket")
				}

				db.AssertNotCalled(t, "MoveProcessedExpectedTxs", mock.Anything)
			})
		}
	}

	t.Run("no port and no server name", func(t *testing.T) {
		t.Parallel()

		files := newTestTLSFiles(t)
		_, db, err := newObserver(t, "node.example", &gouroboros.TLSConfig{
			CAFile: files.caFile, CertFile: files.certFile, KeyFile: files.keyFile,
		})
		require.ErrorContains(t, err, "invalid networkTls for chain prime: ")
		db.AssertNotCalled(t, "MoveProcessedExpectedTxs", mock.Anything)
	})

	t.Run("explicit server name and no port passes", func(t *testing.T) {
		t.Parallel()

		files := newTestTLSFiles(t)
		obs, db, err := newObserver(t, "node.example", &gouroboros.TLSConfig{
			CAFile: files.caFile, CertFile: files.certFile, KeyFile: files.keyFile, ServerName: "node.example",
		})
		require.NoError(t, err)
		require.NotNil(t, obs)
		db.AssertCalled(t, "MoveProcessedExpectedTxs", common.ChainIDStrPrime)
	})

	t.Run("valid files with ip address pass", func(t *testing.T) {
		t.Parallel()

		files := newTestTLSFiles(t)
		obs, _, err := newObserver(t, "https://10.0.0.1:3001", &gouroboros.TLSConfig{
			CAFile: files.caFile, CertFile: files.certFile, KeyFile: files.keyFile,
		})
		require.NoError(t, err)
		require.NotNil(t, obs)
	})

	t.Run("empty certFile", func(t *testing.T) {
		t.Parallel()

		files := newTestTLSFiles(t)
		_, db, err := newObserver(t, "10.0.0.1:3001", &gouroboros.TLSConfig{
			CAFile: files.caFile, KeyFile: files.keyFile,
		})
		require.ErrorContains(t, err, "tls certFile")
		db.AssertNotCalled(t, "MoveProcessedExpectedTxs", mock.Anything)
	})

	t.Run("ca file holds a leaf certificate", func(t *testing.T) {
		t.Parallel()

		files := newTestTLSFiles(t)
		_, db, err := newObserver(t, "10.0.0.1:3001", &gouroboros.TLSConfig{
			CAFile: files.certFile, CertFile: files.certFile, KeyFile: files.keyFile,
		})
		require.ErrorContains(t, err, "invalid networkTls for chain prime")
		db.AssertNotCalled(t, "MoveProcessedExpectedTxs", mock.Anything)
	})

	t.Run("ca file holds three CA certificates", func(t *testing.T) {
		t.Parallel()

		files := newTestTLSFiles(t)

		var bundle []byte

		for i := 0; i < 3; i++ {
			caPEM, _, _, _ := newTestCert(t, true, nil, nil)
			bundle = append(bundle, caPEM...)
		}

		_, db, err := newObserver(t, "10.0.0.1:3001", &gouroboros.TLSConfig{
			CAFile:   writeTestFile(t, t.TempDir(), "bundle.crt", bundle),
			CertFile: files.certFile, KeyFile: files.keyFile,
		})
		require.ErrorContains(t, err, "invalid networkTls for chain prime")
		db.AssertNotCalled(t, "MoveProcessedExpectedTxs", mock.Anything)
	})

	t.Run("mismatched cert and key", func(t *testing.T) {
		t.Parallel()

		files := newTestTLSFiles(t)
		other := newTestTLSFiles(t)
		_, db, err := newObserver(t, "10.0.0.1:3001", &gouroboros.TLSConfig{
			CAFile: files.caFile, CertFile: files.certFile, KeyFile: other.keyFile,
		})
		require.ErrorContains(t, err, "tls certFile/keyFile")
		db.AssertNotCalled(t, "MoveProcessedExpectedTxs", mock.Anything)
	})
}
