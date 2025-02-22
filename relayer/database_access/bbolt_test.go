package databaseaccess

import (
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
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

	t.Run("AddLastSubmittedBatchID", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		err = db.AddLastSubmittedBatchID(common.ChainIDStrPrime, big.NewInt(0))
		require.NoError(t, err)
		err = db.AddLastSubmittedBatchID(common.ChainIDStrPrime, big.NewInt(123))
		require.NoError(t, err)
	})

	t.Run("GetLastSubmittedBatchID", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		expectedOutput := big.NewInt(1)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		res, err := db.GetLastSubmittedBatchID(common.ChainIDStrPrime)
		require.NoError(t, err)
		require.Nil(t, res)

		err = db.AddLastSubmittedBatchID(common.ChainIDStrPrime, expectedOutput)
		require.NoError(t, err)

		res, err = db.GetLastSubmittedBatchID(common.ChainIDStrPrime)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, 0, res.Cmp(expectedOutput))
	})
}
