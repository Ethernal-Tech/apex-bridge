package utils

import (
	"path/filepath"
	"testing"

	loggerInfra "github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestNewIndexerLogger(t *testing.T) {
	mainLogger := hclog.NewNullLogger()

	t.Run("writes next to the main log file", func(t *testing.T) {
		logsDir := t.TempDir()

		indexerLogger, err := NewIndexerLogger(loggerInfra.LoggerConfig{
			LogFilePath: filepath.Join(logsDir, "validator-components.log"),
			LogLevel:    hclog.Info,
			AppendFile:  true,
		}, "prime", mainLogger)
		require.NoError(t, err)
		require.NotNil(t, indexerLogger)

		indexerLogger.Info("indexer is alive")

		require.FileExists(t, filepath.Join(logsDir, "prime-indexer.log"))
		require.NoFileExists(t, filepath.Join(logsDir, "validator-components.log"))
	})

	t.Run("every chain gets its own file", func(t *testing.T) {
		logsDir := t.TempDir()
		baseConfig := loggerInfra.LoggerConfig{
			LogFilePath: filepath.Join(logsDir, "validator-components.log"),
			LogLevel:    hclog.Info,
			AppendFile:  true,
		}

		for _, chainID := range []string{"prime", "vector", "nexus", "polygon", "solana"} {
			indexerLogger, err := NewIndexerLogger(baseConfig, chainID, mainLogger)
			require.NoError(t, err)

			indexerLogger.Info("indexer is alive")

			require.FileExists(t, filepath.Join(logsDir, chainID+"-indexer.log"))
		}
	})

	t.Run("main logger is reused when it has no log file", func(t *testing.T) {
		indexerLogger, err := NewIndexerLogger(loggerInfra.LoggerConfig{
			LogFilePath: "  ",
			LogLevel:    hclog.Info,
		}, "prime", mainLogger)
		require.NoError(t, err)
		require.Equal(t, mainLogger, indexerLogger)
	})
}
