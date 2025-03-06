package databaseaccess

import (
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBoltDatabase(t *testing.T) {
	testDir, err := os.MkdirTemp("", "boltdb-test")
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	filePath := filepath.Join(testDir, "temp_test.db")

	dbCleanup := func() {
		if _, err := os.Stat(filePath); err == nil {
			os.Remove(filePath)
		}
	}

	t.Run("Init", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)
	})

	t.Run("Init should fail", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init("")
		require.Error(t, err)
	})

	t.Run("Close", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		err = db.Close()
		require.NoError(t, err)
	})

	const primeChainID = "chainId"

	t.Run("AddLastSubmittedBatchID", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		batchID := big.NewInt(1)

		err = db.AddLastSubmittedBatchID(primeChainID, batchID)
		require.NoError(t, err)
	})

	t.Run("GetLastSubmittedBatchID", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		batchIDToInsert := big.NewInt(1)

		batchID, err := db.GetLastSubmittedBatchID(primeChainID)
		require.NoError(t, err)
		require.Nil(t, batchID)

		err = db.AddLastSubmittedBatchID(primeChainID, batchIDToInsert)
		require.NoError(t, err)

		batchID, err = db.GetLastSubmittedBatchID(primeChainID)
		require.NoError(t, err)
		require.NotNil(t, batchID)
		require.Equal(t, batchIDToInsert, batchID)
	})
}
