package utils

import (
	"fmt"
	"path/filepath"
	"strings"

	loggerInfra "github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/hashicorp/go-hclog"
)

// NewIndexerLogger creates a logger which writes the indexer (block syncer, event tracker)
// output of a single chain into its own <chainID>-indexer.log file, placed next to the main
// log file. That way the noisy indexer output does not end up in the main log file.
func NewIndexerLogger(baseConfig loggerInfra.LoggerConfig, chainID string, mainLogger hclog.Logger,
) (hclog.Logger, error) {
	logFilePath := strings.TrimSpace(baseConfig.LogFilePath)
	if logFilePath == "" {
		return mainLogger, nil
	}

	indexerLoggerConfig := baseConfig
	indexerLoggerConfig.LogFilePath = filepath.Join(
		filepath.Dir(logFilePath), fmt.Sprintf("%s-indexer.log", chainID))

	indexerLogger, err := loggerInfra.NewLogger(indexerLoggerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create indexer logger for `%s`: %w", chainID, err)
	}

	return indexerLogger, nil
}
