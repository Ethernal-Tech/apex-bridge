package databaseaccess

import (
	"fmt"
	"path/filepath"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/staking/core"
	"go.etcd.io/bbolt"
)

var (
	UnprocessedTxsBucket            = "UnprocessedTxs"
	PendingTxsBucket                = "PendingTxs"
	ProcessedTxsBucket              = "ProcessedTxs"
	ExpectedTxsBucket               = "ExpectedTxs"
	UnprocessedBatchEventsBucket    = "UnprocessedBatchEvents"
	ProcessedTxsByInnerActionBucket = "ProcessedTxsByInnerAction"
	StakingAddressesBucket          = "StakingAddresses"
	ExchangeRateBucket              = "ExchangeRate"
)

func NewDatabase(pathToFile string, smConfig *core.StakingManagerConfiguration) (*bbolt.DB, error) {
	if err := common.CreateDirectoryIfNotExists(filepath.Dir(pathToFile), 0770); err != nil {
		return nil, fmt.Errorf("failed to create directory for oracle database: %w", err)
	}

	return initDB(pathToFile, smConfig)
}

func ChainBucket(bucket string, chainID string) []byte {
	return fmt.Appendf(nil, "%s_%s", bucket, chainID)
}

func Bucket(bucket string) []byte {
	return []byte(bucket)
}

func initDB(filePath string, smConfig *core.StakingManagerConfiguration) (*bbolt.DB, error) {
	db, err := bbolt.Open(filePath, 0660, nil)
	if err != nil {
		return nil, fmt.Errorf("could not open db: %w", err)
	}

	allBuckets := [][]byte{Bucket(ExchangeRateBucket)}
	for _, chain := range smConfig.Chains {
		allBuckets = append(allBuckets, defaultChainBuckets(chain.ChainID)...)
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
		ChainBucket(StakingAddressesBucket, chainID),
	}
}
