package database_access

import (
	"math/big"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBoltDatabase(t *testing.T) {
	const filePath = "temp_test.db"

	dbCleanup := func() {
		if _, err := os.Stat(filePath); err == nil {
			os.Remove(filePath)
		}
	}

	t.Cleanup(dbCleanup)

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

	t.Run("AddLastSubmittedBatchId", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		err = db.AddLastSubmittedBatchId("prime", big.NewInt(0))
		require.NoError(t, err)
		err = db.AddLastSubmittedBatchId("prime", big.NewInt(123))
		require.NoError(t, err)
	})

	t.Run("GetLastSubmittedBatchId", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		expectedOutput := big.NewInt(1)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		err = db.AddLastSubmittedBatchId("prime", expectedOutput)
		require.NoError(t, err)

		res, err := db.GetLastSubmittedBatchId("invalid")
		require.Error(t, err)
		require.Nil(t, res)

		res, err = db.GetLastSubmittedBatchId("prime")
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, 0, res.Cmp(expectedOutput))
	})
}
