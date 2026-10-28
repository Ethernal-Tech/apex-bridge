package cligenerateconfigs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Ethernal-Tech/cardano-infrastructure/indexer/gouroboros"
	"github.com/stretchr/testify/require"
)

func validTLSChainParams() *cardanoChainGenerateConfigsParams {
	return &cardanoChainGenerateConfigsParams{
		chainIDString:  "prime",
		networkAddress: "10.0.0.1:3001",
		ogmiosURL:      "http://10.0.0.1:1337",
		networkMagic:   2,
		dbsPath:        "db",
	}
}

func newExecuteParams(t *testing.T) *cardanoChainGenerateConfigsParams {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "vc.json"), []byte(`{}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "relayer.json"), []byte(`{}`), 0o600))

	p := validTLSChainParams()
	p.outputDir = dir
	p.outputValidatorComponentsFileName = "vc.json"
	p.outputRelayerFileName = "relayer.json"

	return p
}

func readJSONFile(t *testing.T, path string) map[string]any {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var out map[string]any

	require.NoError(t, json.Unmarshal(data, &out))

	return out
}

func TestCardanoChainParams_NetworkTLSFlags(t *testing.T) {
	t.Parallel()

	t.Run("only ca file", func(t *testing.T) {
		t.Parallel()

		p := validTLSChainParams()
		p.networkTLSCAFile = "/c/ca.crt"

		err := p.validateFlags()
		require.ErrorContains(t, err, networkTLSCertFileFlag)
		require.ErrorContains(t, err, networkTLSKeyFileFlag)
	})

	t.Run("server name without files", func(t *testing.T) {
		t.Parallel()

		p := validTLSChainParams()
		p.networkTLSServerName = "node.example"

		require.ErrorContains(t, p.validateFlags(), networkTLSServerNameFlag)
	})

	t.Run("all flags valid and written", func(t *testing.T) {
		t.Parallel()

		p := newExecuteParams(t)
		p.networkTLSCAFile = "/c/ca.crt"
		p.networkTLSCertFile = "/c/v.crt"
		p.networkTLSKeyFile = "/c/v.key"
		p.networkTLSServerName = "node.example"

		require.NoError(t, p.validateFlags())

		_, err := p.Execute(nil)
		require.NoError(t, err)

		vc := readJSONFile(t, filepath.Join(p.outputDir, "vc.json"))

		chains, ok := vc["cardanoChains"].(map[string]any)
		require.True(t, ok)

		chain, ok := chains["prime"].(map[string]any)
		require.True(t, ok)
		require.Equal(t, map[string]any{
			"caFile": "/c/ca.crt", "certFile": "/c/v.crt", "keyFile": "/c/v.key", "serverName": "node.example",
		}, chain["networkTls"])

		// relayer chain entry has no networkTls
		relayer := readJSONFile(t, filepath.Join(p.outputDir, "relayer.json"))
		relayerChains, ok := relayer["chains"].(map[string]any)
		require.True(t, ok)
		require.NotContains(t, string(mustMarshal(t, relayerChains["prime"])), "networkTls")
		require.NotContains(t, string(mustMarshal(t, relayerChains["prime"])), "networkTLS")
	})

	t.Run("file flags only with ip address", func(t *testing.T) {
		t.Parallel()

		p := newExecuteParams(t)
		p.networkTLSCAFile = "/c/ca.crt"
		p.networkTLSCertFile = "/c/v.crt"
		p.networkTLSKeyFile = "/c/v.key"

		require.NoError(t, p.validateFlags())

		_, err := p.Execute(nil)
		require.NoError(t, err)
	})

	t.Run("unix socket address is rejected by network address validation", func(t *testing.T) {
		t.Parallel()

		p := validTLSChainParams()
		p.networkAddress = "/ipc/node.socket"
		p.networkTLSCAFile = "/c/ca.crt"
		p.networkTLSCertFile = "/c/v.crt"
		p.networkTLSKeyFile = "/c/v.key"

		require.ErrorContains(t, p.validateFlags(), "invalid "+networkAddressFlag)
	})

	t.Run("values stored verbatim", func(t *testing.T) {
		t.Parallel()

		want := gouroboros.TLSConfig{
			CAFile:     "--output=/tmp/x",
			CertFile:   "/p with spaces/v.crt",
			KeyFile:    "-$(touch /tmp/pwned)=k",
			ServerName: "node name.example",
		}

		p := newExecuteParams(t)
		p.networkTLSCAFile = want.CAFile
		p.networkTLSCertFile = want.CertFile
		p.networkTLSKeyFile = want.KeyFile
		p.networkTLSServerName = want.ServerName

		require.NoError(t, p.validateFlags())

		_, err := p.Execute(nil)
		require.NoError(t, err)

		raw, err := os.ReadFile(filepath.Join(p.outputDir, "vc.json"))
		require.NoError(t, err)

		var parsed struct {
			Chains map[string]struct {
				NetworkTLS *gouroboros.TLSConfig `json:"networkTls"`
			} `json:"cardanoChains"`
		}

		require.NoError(t, json.Unmarshal(raw, &parsed))
		require.Equal(t, &want, parsed.Chains["prime"].NetworkTLS)
	})

	t.Run("empty ca file flag counts as empty", func(t *testing.T) {
		t.Parallel()

		p := newExecuteParams(t)
		p.networkTLSCAFile = ""

		require.NoError(t, p.validateFlags())

		_, err := p.Execute(nil)
		require.NoError(t, err)

		require.NotContains(t, string(mustReadFile(t, filepath.Join(p.outputDir, "vc.json"))), "networkTls")
	})

	t.Run("rerun without flags drops the key", func(t *testing.T) {
		t.Parallel()

		p := newExecuteParams(t)
		p.networkTLSCAFile = "/c/ca.crt"
		p.networkTLSCertFile = "/c/v.crt"
		p.networkTLSKeyFile = "/c/v.key"

		_, err := p.Execute(nil)
		require.NoError(t, err)
		require.Contains(t, string(mustReadFile(t, filepath.Join(p.outputDir, "vc.json"))), "networkTls")

		p.networkTLSCAFile, p.networkTLSCertFile, p.networkTLSKeyFile = "", "", ""

		_, err = p.Execute(nil)
		require.NoError(t, err)
		require.NotContains(t, string(mustReadFile(t, filepath.Join(p.outputDir, "vc.json"))), "networkTls")
	})
}

func TestCardanoChainParams_networkTLSConfig(t *testing.T) {
	t.Parallel()

	p := validTLSChainParams()
	require.Nil(t, p.networkTLSConfig())

	p.networkTLSCertFile = "/c/v.crt"
	p.networkTLSKeyFile = "/c/v.key"

	require.Equal(t, &gouroboros.TLSConfig{CertFile: "/c/v.crt", KeyFile: "/c/v.key"}, p.networkTLSConfig())
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()

	data, err := json.Marshal(v)
	require.NoError(t, err)

	return data
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	return data
}
