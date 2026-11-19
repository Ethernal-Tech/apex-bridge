package clideploysolana

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/gagliardetto/solana-go"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

// writeTestKeypair writes a fresh random keypair as a solana-keygen JSON byte array.
func writeTestKeypair(t *testing.T, path string) solana.PrivateKey {
	t.Helper()

	pk, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)

	values := make([]int, len(pk))
	for i, b := range pk {
		values[i] = int(b)
	}

	content, err := json.Marshal(values)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	require.NoError(t, os.WriteFile(path, content, 0o600))

	return pk
}

func argValue(t *testing.T, args []string, name string) string {
	t.Helper()

	for i, a := range args {
		if a == name {
			require.Less(t, i+1, len(args))

			return args[i+1]
		}
	}

	require.Failf(t, "flag not found", "%s in %v", name, args)

	return ""
}

func newValidUpgradeParams(t *testing.T) *upgradeProgramParams {
	t.Helper()

	dir := t.TempDir()
	programID := solana.NewWallet().PublicKey().String()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "prog.so"), []byte("x"), 0o600))
	writeTestKeypair(t, filepath.Join(dir, "admin.json"))

	return &upgradeProgramParams{
		rpcURL:                     "http://127.0.0.1:8899",
		programID:                  programID,
		buildPath:                  filepath.Join(dir, "prog.so"),
		commitment:                 "confirmed",
		upgradeProgramVersion:      "0.5.0",
		adminKeyPath:               filepath.Join(dir, "admin.json"),
		confirmationTimeoutSeconds: 10,
	}
}

func TestBuildUpgradeDeployArgs(t *testing.T) {
	t.Run("admin is upgrade authority", func(t *testing.T) {
		p := newValidUpgradeParams(t)
		require.NoError(t, p.validateFlags())

		args := p.deployArgs(p.adminKeyPath)

		require.Equal(t, p.programID, argValue(t, args, "--program-id"))
		require.Equal(t, p.adminKeyPath, argValue(t, args, "--upgrade-authority"))
		require.Equal(t, p.adminKeyPath, argValue(t, args, "--fee-payer"))
		require.Equal(t, p.rpcURL, argValue(t, args, "--url"))
		require.Equal(t, "confirmed", argValue(t, args, "--commitment"))
		// the solana CLI needs -k even with --fee-payer, or it looks for ~/.config/solana/id.json
		require.Equal(t, p.adminKeyPath, argValue(t, args, "-k"))
		require.Equal(t, p.buildPath, args[len(args)-1])
	})

	t.Run("separate upgrade authority", func(t *testing.T) {
		p := newValidUpgradeParams(t)
		p.upgradeAuthorityKeyPath = filepath.Join(filepath.Dir(p.adminKeyPath), "upgrade.json")
		writeTestKeypair(t, p.upgradeAuthorityKeyPath)
		require.NoError(t, p.validateFlags())

		args := p.deployArgs(p.adminKeyPath)

		require.Equal(t, p.upgradeAuthorityKeyPath, argValue(t, args, "--upgrade-authority"))
		require.NotEqual(t, p.adminKeyPath, argValue(t, args, "--upgrade-authority"))
	})

	t.Run("hostile paths", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)

		writeTestKeypair(t, filepath.Join(dir, "--final"))
		writeTestKeypair(t, filepath.Join(dir, "with space", "fee payer.json"))
		writeTestKeypair(t, filepath.Join(dir, "sub", "admin.json"))

		p := newValidUpgradeParams(t)
		p.buildPath = "--final"
		p.feePayerKeyPath = "with space/fee payer.json"
		p.upgradeAuthorityKeyPath = "--final"
		p.adminKeyPath = "sub/../sub/admin.json"

		// a relative build path literally named "--final" must be found relative to cwd
		require.NoError(t, p.validateFlags())

		args := p.deployArgs(p.adminKeyPath)

		for _, name := range []string{"-k", "--fee-payer", "--upgrade-authority"} {
			v := argValue(t, args, name)
			require.True(t, strings.HasPrefix(v, "/"), "%s=%s", name, v)
		}

		require.Equal(t, filepath.Join(dir, "with space", "fee payer.json"), argValue(t, args, "--fee-payer"))
		require.Equal(t, argValue(t, args, "--fee-payer"), argValue(t, args, "-k"))
		require.True(t, strings.HasPrefix(args[len(args)-1], "/"))
		require.NotContains(t, args, "--final")

		// ".." relative path is resolved to an absolute, clean path
		p2 := newValidUpgradeParams(t)
		t.Chdir(filepath.Join(dir, "sub"))

		p2.buildPath = "../--final"
		require.NoError(t, p2.validateFlags())
		require.Equal(t, filepath.Join(dir, "--final"), p2.buildPath)
		require.True(t, strings.HasPrefix(p2.deployArgs(p2.adminKeyPath)[len(p2.deployArgs(p2.adminKeyPath))-1], "/"))
	})
}

func TestUpgradeProgramFlags(t *testing.T) {
	t.Run("key flag removed", func(t *testing.T) {
		cmd := &cobra.Command{Use: "x"}
		(&upgradeProgramParams{}).setFlags(cmd)

		require.Error(t, cmd.ParseFlags([]string{"--key", "x"}))
		require.ErrorContains(t, cmd.ParseFlags([]string{"--key", "x"}), "unknown flag: --key")
		require.NoError(t, cmd.ParseFlags([]string{"--" + upgradeAuthorityKeyFlag, "x"}))
	})
}

func TestUpgradeProgramValidateFlags(t *testing.T) {
	t.Run("missing upgrade authority file", func(t *testing.T) {
		p := newValidUpgradeParams(t)
		p.upgradeAuthorityKeyPath = filepath.Join(t.TempDir(), "missing-upgrade.json")

		err := p.validateFlags()
		require.ErrorContains(t, err, p.upgradeAuthorityKeyPath)
	})

	t.Run("invalid commitment", func(t *testing.T) {
		for _, c := range []string{"--final", "bogus"} {
			p := newValidUpgradeParams(t)
			p.commitment = c

			err := p.validateFlags()
			require.ErrorContains(t, err, "--"+commitmentFlag, c)
		}
	})

	t.Run("valid", func(t *testing.T) {
		require.NoError(t, newValidUpgradeParams(t).validateFlags())
	})
}

var _ common.OutputFormatter = (*fakeOutputter)(nil)

type fakeOutputter struct{}

func (fakeOutputter) SetError(error)                           {}
func (fakeOutputter) SetCommandResult(common.ICommandResult)   {}
func (fakeOutputter) WriteOutput()                             {}
func (fakeOutputter) WriteCommandResult(common.ICommandResult) {}
func (fakeOutputter) Write(p []byte) (int, error)              { return len(p), nil }
