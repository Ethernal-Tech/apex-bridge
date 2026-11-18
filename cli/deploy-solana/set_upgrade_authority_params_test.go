package clideploysolana

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

func newValidSetUpgradeParams(t *testing.T) *setUpgradeAuthorityParams {
	t.Helper()

	dir := t.TempDir()
	writeTestKeypair(t, filepath.Join(dir, "current.json"))
	writeTestKeypair(t, filepath.Join(dir, "new.json"))

	return &setUpgradeAuthorityParams{
		rpcURL:                     "http://127.0.0.1:8899",
		programID:                  solana.NewWallet().PublicKey().String(),
		upgradeAuthorityKeyPath:    filepath.Join(dir, "current.json"),
		newUpgradeAuthorityKeyPath: filepath.Join(dir, "new.json"),
		commitment:                 "confirmed",
	}
}

func TestSetUpgradeAuthorityArgs(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		p := newValidSetUpgradeParams(t)
		require.NoError(t, p.validateFlags())

		args := p.buildArgs(p.upgradeAuthorityKeyPath)

		require.Equal(t, []string{
			"program", "set-upgrade-authority", p.programID,
			"--url", p.rpcURL,
			"-k", p.upgradeAuthorityKeyPath,
			"--upgrade-authority", p.upgradeAuthorityKeyPath,
			"--new-upgrade-authority", p.newUpgradeAuthorityKeyPath,
			"--commitment", "confirmed",
		}, args)
		require.True(t, filepath.IsAbs(p.newUpgradeAuthorityKeyPath))
	})

	t.Run("fee payer", func(t *testing.T) {
		p := newValidSetUpgradeParams(t)
		p.feePayerKeyPath = filepath.Join(filepath.Dir(p.upgradeAuthorityKeyPath), "fee.json")
		writeTestKeypair(t, p.feePayerKeyPath)
		require.NoError(t, p.validateFlags())

		args := p.buildArgs(p.upgradeAuthorityKeyPath)
		require.Equal(t, p.feePayerKeyPath, argValue(t, args, "-k"))
		require.Equal(t, p.upgradeAuthorityKeyPath, argValue(t, args, "--upgrade-authority"))
	})

	t.Run("hostile paths", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)

		writeTestKeypair(t, filepath.Join(dir, "--final"))
		writeTestKeypair(t, filepath.Join(dir, "with space", "new key.json"))
		writeTestKeypair(t, filepath.Join(dir, "sub", "fee.json"))
		t.Chdir(filepath.Join(dir, "sub"))

		p := newValidSetUpgradeParams(t)
		p.upgradeAuthorityKeyPath = "../--final"
		p.newUpgradeAuthorityKeyPath = "../with space/new key.json"
		p.feePayerKeyPath = "../sub/fee.json"
		require.NoError(t, p.validateFlags())

		args := p.buildArgs(p.upgradeAuthorityKeyPath)

		for _, name := range []string{"-k", "--upgrade-authority", "--new-upgrade-authority"} {
			v := argValue(t, args, name)
			require.True(t, strings.HasPrefix(v, "/"), "%s=%s", name, v)
			require.NotContains(t, v, "..")
		}

		require.Equal(t, filepath.Join(dir, "with space", "new key.json"),
			argValue(t, args, "--new-upgrade-authority"))
		require.NotContains(t, args, "--final")
		require.NotContains(t, args, "--skip-new-upgrade-authority-signer-check")

		// the key file literally named --final must be a single absolute argument
		require.Equal(t, filepath.Join(dir, "--final"), argValue(t, args, "--upgrade-authority"))
	})
}

func TestSetUpgradeAuthorityValidateFlags(t *testing.T) {
	notAKeypair := func(content string) func(t *testing.T, p *setUpgradeAuthorityParams) {
		return func(t *testing.T, p *setUpgradeAuthorityParams) {
			t.Helper()

			path := filepath.Join(t.TempDir(), "bad.json")
			require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

			p.newUpgradeAuthorityKeyPath = path
		}
	}

	tests := []struct {
		name   string
		mutate func(t *testing.T, p *setUpgradeAuthorityParams)
		errHas string
	}{
		{"valid", func(*testing.T, *setUpgradeAuthorityParams) {}, ""},
		{"new equals current", func(_ *testing.T, p *setUpgradeAuthorityParams) {
			p.newUpgradeAuthorityKeyPath = p.upgradeAuthorityKeyPath
		}, "--" + newUpgradeAuthorityKeyFlag},
		{"neither key nor config", func(_ *testing.T, p *setUpgradeAuthorityParams) {
			p.upgradeAuthorityKeyPath = ""
		}, "--" + upgradeAuthorityKeyFlag},
		{"both key and config", func(_ *testing.T, p *setUpgradeAuthorityParams) {
			p.privateKeyConfig = "/some/config.json"
		}, "--" + privateKeyConfigFlag},
		{"bad program id", func(_ *testing.T, p *setUpgradeAuthorityParams) {
			p.programID = "not-a-key"
		}, "--" + programIDFlag},
		{"invalid commitment", func(_ *testing.T, p *setUpgradeAuthorityParams) {
			p.commitment = "--final"
		}, "--" + commitmentFlag},
		{"bad url", func(_ *testing.T, p *setUpgradeAuthorityParams) {
			p.rpcURL = "nope"
		}, "--" + rpcURLFlag},
		{"new key missing", func(_ *testing.T, p *setUpgradeAuthorityParams) {
			p.newUpgradeAuthorityKeyPath = ""
		}, "--" + newUpgradeAuthorityKeyFlag},
		{"new key empty file", notAKeypair(""), "--" + newUpgradeAuthorityKeyFlag},
		{"new key base58 pubkey", notAKeypair("11111111111111111111111111111111"), "--" + newUpgradeAuthorityKeyFlag},
		{"new key bad json", notAKeypair("{not json"), "--" + newUpgradeAuthorityKeyFlag},
		{"fee payer missing", func(t *testing.T, p *setUpgradeAuthorityParams) {
			t.Helper()

			p.feePayerKeyPath = filepath.Join(t.TempDir(), "nope.json")
		}, "--" + feePayerKeyFlag},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newValidSetUpgradeParams(t)
			tt.mutate(t, p)

			err := p.validateFlags()
			if tt.errHas == "" {
				require.NoError(t, err)

				return
			}

			require.ErrorContains(t, err, tt.errHas)
		})
	}
}

func TestDeploySolanaCommand(t *testing.T) {
	t.Run("set-upgrade-authority wired", func(t *testing.T) {
		root := GetDeploySolanaCommand()

		sub, _, err := root.Find([]string{setUpgradeAuthorityCommandUse})
		require.NoError(t, err)
		require.Equal(t, setUpgradeAuthorityCommandUse, sub.Use)

		upgrade, _, err := root.Find([]string{upgradeProgramCommandUse})
		require.NoError(t, err)
		require.Nil(t, upgrade.Flags().Lookup("key"))
		require.NotNil(t, upgrade.Flags().Lookup(upgradeAuthorityKeyFlag))

		// runPreRun dispatches to the command's own validateFlags: valid url and program id,
		// no key, so the error is about the upgrade authority key flag
		saved := *setUpgradeAuthorityParamsData

		t.Cleanup(func() { *setUpgradeAuthorityParamsData = saved })

		*setUpgradeAuthorityParamsData = setUpgradeAuthorityParams{
			rpcURL:     "http://127.0.0.1:8899",
			programID:  solana.NewWallet().PublicKey().String(),
			commitment: defaultCommitment,
		}

		require.ErrorContains(t, runPreRun(sub, nil), "--"+upgradeAuthorityKeyFlag)
	})
}

func TestSetUpgradeAuthorityConfigRPCFailure(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)

	p := newValidSetUpgradeParams(t)
	p.rpcURL = "http://127.0.0.1:1"
	require.NoError(t, p.validateFlags())

	// Use the in-memory signers as loaded by --key-config.
	p.currentPrivateKey = writeTestKeypair(t, filepath.Join(t.TempDir(), "cur.json"))
	p.upgradeAuthorityKeyPath = ""
	p.privateKeyConfig = "/cfg.json"

	_, err := p.Execute(fakeOutputter{})
	require.ErrorContains(t, err, "get latest blockhash")

	entries, err := os.ReadDir(tmp)
	require.NoError(t, err)
	require.Empty(t, entries)
}
