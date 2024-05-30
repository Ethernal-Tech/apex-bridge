package cardanotx

import (
	"os"
	"path"
	"testing"

	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	secretsHelper "github.com/Ethernal-Tech/cardano-infrastructure/secrets/helper"
	"github.com/stretchr/testify/require"
)

func TestWallet(t *testing.T) {
	testDir, err := os.MkdirTemp("", "test-cardano-wallet")
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	secretsMngr, err := secretsHelper.CreateSecretsManager(&secrets.SecretsManagerConfig{
		Path: path.Join(testDir, "w1"),
		Type: secrets.Local,
	})
	require.NoError(t, err)

	wallet, err := GenerateWallet(secretsMngr, "prime", false, true)
	require.NoError(t, err)

	walletStake, err := GenerateWallet(secretsMngr, "vector", true, true)
	require.NoError(t, err)

	t.Run("loading without stake", func(t *testing.T) {
		wallet2, err := LoadWallet(secretsMngr, "prime")
		require.NoError(t, err)

		require.Equal(t, wallet, wallet2)
	})

	t.Run("loading with stake", func(t *testing.T) {
		wallet2, err := LoadWallet(secretsMngr, "vector")
		require.NoError(t, err)

		require.Equal(t, walletStake, wallet2)
	})

	t.Run("idempotent without stake", func(t *testing.T) {
		wallet2, err := GenerateWallet(secretsMngr, "prime", false, false)
		require.NoError(t, err)

		require.Equal(t, wallet, wallet2)
	})

	t.Run("idempotent with stake", func(t *testing.T) {
		wallet2, err := GenerateWallet(secretsMngr, "vector", true, false)
		require.NoError(t, err)

		require.Equal(t, walletStake, wallet2)
	})
}
