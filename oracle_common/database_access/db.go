package databaseaccess

import (
	"fmt"
	"path/filepath"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"go.etcd.io/bbolt"
)

var (
	UnprocessedTxsBucket            = "UnprocessedTxs"
	PendingTxsBucket                = "PendingTxs"
	ProcessedTxsBucket              = "ProcessedTxs"
	ExpectedTxsBucket               = "ExpectedTxs"
	UnprocessedBatchEventsBucket    = "UnprocessedBatchEvents"
	ProcessedTxsByInnerActionBucket = "ProcessedTxsByInnerAction"
	BlocksSubmitterBucket           = "BlocksSubmitterBucket"
)

func NewDatabase(pathToFile string, appConfig *core.AppConfig) (*bbolt.DB, error) {
	if err := common.CreateDirectoryIfNotExists(filepath.Dir(pathToFile), 0770); err != nil {
		return nil, fmt.Errorf("failed to create directory for oracle database: %w", err)
	}

	return initDB(pathToFile, appConfig)
}

func ChainBucket(bucket string, chainID string) []byte {
	return fmt.Appendf(nil, "%s_%s", bucket, chainID)
}

func initDB(filePath string, appConfig *core.AppConfig) (*bbolt.DB, error) {
	db, err := bbolt.Open(filePath, 0660, nil)
	if err != nil {
		return nil, fmt.Errorf("could not open db: %w", err)
	}

	var allBuckets [][]byte
	for _, chain := range appConfig.CardanoChains {
		allBuckets = append(allBuckets, defaultChainBuckets(chain.ChainID)...)
	}

	for _, chain := range appConfig.EthChains {
		allBuckets = append(
			append(allBuckets, defaultChainBuckets(chain.ChainID)...),
			ChainBucket(ProcessedTxsByInnerActionBucket, chain.ChainID),
		)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		for _, bn := range allBuckets {
			_, err := tx.CreateBucketIfNotExists(bn)
			if err != nil {
				return fmt.Errorf("could not bucket: %s, err: %w", string(bn), err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return db, nil
}

func defaultChainBuckets(chainID string) [][]byte {
	return [][]byte{
		ChainBucket(UnprocessedTxsBucket, chainID),
		ChainBucket(PendingTxsBucket, chainID),
		ChainBucket(ProcessedTxsBucket, chainID),
		ChainBucket(ExpectedTxsBucket, chainID),
		ChainBucket(UnprocessedBatchEventsBucket, chainID),
		ChainBucket(BlocksSubmitterBucket, chainID),
	}
}
