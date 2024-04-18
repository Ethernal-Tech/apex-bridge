package database_access

import (
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/stretchr/testify/require"
)

func TestDatabase(t *testing.T) {
	const filePath = "temp_test.db"

	dbCleanup := func() {
		common.RemoveDirOrFilePathIfExists(filePath)
	}

	t.Cleanup(dbCleanup)

	t.Run("NewDatabase", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := NewDatabase(filePath)
		require.NoError(t, err)
		require.NotNil(t, db)
	})

	t.Run("NewDatabase should fail", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := NewDatabase("")
		require.Error(t, err)
		require.Nil(t, db)
	})
}
