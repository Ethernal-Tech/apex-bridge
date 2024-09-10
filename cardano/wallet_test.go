package cardanotx

import (
	"os"
	"path/filepath"
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
		Path: filepath.Join(testDir, "w1"),
		Type: secrets.Local,
	})
	require.NoError(t, err)

	wallet, err := GenerateWallet(secretsMngr, "prime", false, true)
	require.NoError(t, err)

	walletStake, err := GenerateWallet(secretsMngr, "vector", true, true)
	require.NoError(t, err)

	t.Run("loading without stake", func(t *testing.T) {
		walletWithoutStake, err := LoadWallet(secretsMngr, "prime")
		require.NoError(t, err)

		require.Equal(t, wallet, walletWithoutStake)
	})

	t.Run("loading with stake", func(t *testing.T) {
		walletWithStake, err := LoadWallet(secretsMngr, "vector")
		require.NoError(t, err)

		require.Equal(t, walletStake, walletWithStake)
	})

	t.Run("idempotent without stake", func(t *testing.T) {
		walletWithoutStake, err := GenerateWallet(secretsMngr, "prime", false, false)
		require.NoError(t, err)

		require.Equal(t, wallet, walletWithoutStake)
	})

	t.Run("idempotent with stake", func(t *testing.T) {
		walletWithStake, err := GenerateWallet(secretsMngr, "vector", true, false)
		require.NoError(t, err)

		require.Equal(t, walletStake, walletWithStake)
	})
}
