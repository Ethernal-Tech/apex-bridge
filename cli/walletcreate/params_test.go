package cliwalletcreate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	solanatx "github.com/Ethernal-Tech/apex-bridge/solana"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	"github.com/gagliardetto/solana-go"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

type nopOutputter struct{}

func (nopOutputter) SetError(error)                           {}
func (nopOutputter) SetCommandResult(common.ICommandResult)   {}
func (nopOutputter) WriteOutput()                             {}
func (nopOutputter) WriteCommandResult(common.ICommandResult) {}
func (nopOutputter) Write(p []byte) (int, error)              { return len(p), nil }

func newSolanaParams(t *testing.T, dataDir, walletType string) *walletCreateParams {
	t.Helper()

	return &walletCreateParams{
		validatorDataDir: dataDir,
		chainID:          "solana",
		walletType:       walletType,
	}
}

func openSecrets(t *testing.T, dataDir string) secrets.SecretsManager {
	t.Helper()

	mngr, err := common.GetSecretsManager(dataDir, "", true)
	require.NoError(t, err)

	return mngr
}

func storedKey(t *testing.T, dataDir, keyName string) solana.PublicKey {
	t.Helper()

	bytes, err := openSecrets(t, dataDir).GetSecret(keyName)
	require.NoError(t, err)

	key, err := solanatx.PrivateKeyFromWalletString(string(bytes))
	require.NoError(t, err)

	return key.PublicKey()
}

func writeKeygenFile(t *testing.T, key solana.PrivateKey) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "keypair.json")
	require.NoError(t, solanatx.WriteSolanaKeypairFile(path, key))

	return path
}

func runSolana(t *testing.T, p *walletCreateParams) (*evmCmdResult, error) {
	t.Helper()

	if err := p.validateFlags(); err != nil {
		return nil, err
	}

	res, err := p.Execute(nopOutputter{})
	if err != nil {
		return nil, err
	}

	evmRes, ok := res.(*evmCmdResult)
	require.True(t, ok)

	return evmRes, nil
}

func TestWalletCreateImportSolanaKey(t *testing.T) {
	keyNames := map[string]string{
		relayerSolanaKeyType: solanatx.GetKeyNameForRelayerSolana("solana"),
		batcherSolanaKeyType: solanatx.GetKeyNameForBatcherSolana("solana"),
		solanaAdminKeyType:   solanatx.GetKeyNameForSolanaAdmin(),
		solanaProgramKeyType: solanatx.GetKeyNameForSolanaProgram(),
	}

	for walletType, keyName := range keyNames {
		t.Run(walletType+" from --key", func(t *testing.T) {
			dataDir := t.TempDir()
			key := solana.NewWallet().PrivateKey

			p := newSolanaParams(t, dataDir, walletType)
			p.key = key.String()

			res, err := runSolana(t, p)
			require.NoError(t, err)
			require.Equal(t, key.PublicKey().String(), res.Address)
			require.Equal(t, key.PublicKey(), storedKey(t, dataDir, keyName))
		})

		t.Run(walletType+" from --key-file", func(t *testing.T) {
			dataDir := t.TempDir()
			key := solana.NewWallet().PrivateKey

			p := newSolanaParams(t, dataDir, walletType)
			p.keyFile = writeKeygenFile(t, key)

			res, err := runSolana(t, p)
			require.NoError(t, err)
			require.Equal(t, key.PublicKey().String(), res.Address)
			require.Equal(t, key.PublicKey(), storedKey(t, dataDir, keyName))
		})
	}
}

func TestWalletCreateImportSolanaKeyExisting(t *testing.T) {
	keyName := solanatx.GetKeyNameForBatcherSolana("solana")

	setup := func(t *testing.T) (string, solana.PrivateKey) {
		t.Helper()

		dataDir := t.TempDir()
		original := solana.NewWallet().PrivateKey

		p := newSolanaParams(t, dataDir, batcherSolanaKeyType)
		p.key = original.String()

		_, err := runSolana(t, p)
		require.NoError(t, err)

		return dataDir, original
	}

	t.Run("same key again is a no-op", func(t *testing.T) {
		dataDir, original := setup(t)

		p := newSolanaParams(t, dataDir, batcherSolanaKeyType)
		p.keyFile = writeKeygenFile(t, original)

		_, err := runSolana(t, p)
		require.NoError(t, err)
		require.Equal(t, original.PublicKey(), storedKey(t, dataDir, keyName))
	})

	t.Run("different key without force is refused", func(t *testing.T) {
		dataDir, original := setup(t)

		p := newSolanaParams(t, dataDir, batcherSolanaKeyType)
		p.key = solana.NewWallet().PrivateKey.String()

		_, err := runSolana(t, p)
		require.ErrorContains(t, err, "already holds a different key")
		require.Equal(t, original.PublicKey(), storedKey(t, dataDir, keyName))
	})

	t.Run("different key with force replaces", func(t *testing.T) {
		dataDir, _ := setup(t)
		replacement := solana.NewWallet().PrivateKey

		p := newSolanaParams(t, dataDir, batcherSolanaKeyType)
		p.key = replacement.String()
		p.forceRegenerate = true

		_, err := runSolana(t, p)
		require.NoError(t, err)
		require.Equal(t, replacement.PublicKey(), storedKey(t, dataDir, keyName))
	})

	t.Run("generate without key keeps the imported key", func(t *testing.T) {
		dataDir, original := setup(t)

		res, err := runSolana(t, newSolanaParams(t, dataDir, batcherSolanaKeyType))
		require.NoError(t, err)
		require.Equal(t, original.PublicKey().String(), res.Address)
	})
}

func TestWalletCreateGenerateSolanaKey(t *testing.T) {
	dataDir := t.TempDir()

	res, err := runSolana(t, newSolanaParams(t, dataDir, relayerSolanaKeyType))
	require.NoError(t, err)
	require.Equal(t, res.Address, storedKey(t, dataDir, solanatx.GetKeyNameForRelayerSolana("solana")).String())
}

func TestWalletCreateImportKeyValidation(t *testing.T) {
	t.Run("non-solana type", func(t *testing.T) {
		p := newSolanaParams(t, t.TempDir(), "relayer-evm")
		p.key = solana.NewWallet().PrivateKey.String()

		require.ErrorContains(t, p.validateFlags(), "only supported for")
	})

	t.Run("both key and key-file", func(t *testing.T) {
		key := solana.NewWallet().PrivateKey
		p := newSolanaParams(t, t.TempDir(), batcherSolanaKeyType)
		p.key = key.String()
		p.keyFile = writeKeygenFile(t, key)

		require.ErrorContains(t, p.validateFlags(), "only one of")
	})

	t.Run("invalid key", func(t *testing.T) {
		p := newSolanaParams(t, t.TempDir(), batcherSolanaKeyType)
		p.key = "not-a-key"

		require.ErrorContains(t, p.validateFlags(), "invalid --key")
	})

	t.Run("public key instead of keypair file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "pub.txt")
		require.NoError(t, os.WriteFile(path, []byte(solana.NewWallet().PublicKey().String()), 0o600))

		p := newSolanaParams(t, t.TempDir(), batcherSolanaKeyType)
		p.keyFile = path

		require.ErrorContains(t, p.validateFlags(), "not a valid solana keypair file")
	})

	t.Run("missing key file", func(t *testing.T) {
		p := newSolanaParams(t, t.TempDir(), batcherSolanaKeyType)
		p.keyFile = filepath.Join(t.TempDir(), "missing.json")

		require.ErrorContains(t, p.validateFlags(), "--key-file")
	})

	t.Run("cobra rejects key with key-file", func(t *testing.T) {
		cmd := &cobra.Command{Use: "x"}
		(&walletCreateParams{}).setFlags(cmd)

		require.NoError(t, cmd.ParseFlags([]string{"--key", "a", "--key-file", "b"}))
		require.Error(t, cmd.ValidateFlagGroups())
	})
}
