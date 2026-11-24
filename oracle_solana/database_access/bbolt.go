package databaseaccess

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	cDatabaseaccess "github.com/Ethernal-Tech/apex-bridge/oracle_common/database_access"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	"go.etcd.io/bbolt"
)

type BBoltDatabase struct {
	cDatabaseaccess.BBoltDBBase[
		*core.SolanaTx,
		*core.ProcessedSolanaTx,
		*core.BridgeExpectedSolanaTx,
	]
}

var _ core.Database = (*BBoltDatabase)(nil)

func (bd *BBoltDatabase) Init(db *bbolt.DB, appConfig *cCore.AppConfig, typeRegister common.TypeRegister) {
	bd.BBoltDBBase.DB = db
	bd.SupportedChains = make(map[string]bool, len(appConfig.SolanaChains))
	bd.TypeRegister = typeRegister

	for _, chain := range appConfig.SolanaChains {
		bd.SupportedChains[chain.ChainID] = true
	}
}

func (bd *BBoltDatabase) GetProcessedTxByInnerActionTxHash(
	chainID string, actionHash []byte,
) (*core.ProcessedSolanaTx, error) {
	if !bd.SupportedChains[chainID] {
		return nil, fmt.Errorf("unsupported chain: %s", chainID)
	}

	var result *core.ProcessedSolanaTx

	err := bd.DB.View(func(tx *bbolt.Tx) error {
		var err error

		result, err = findProcessedExecution(tx, chainID, actionHash)

		return err
	})

	return result, err
}

func (bd *BBoltDatabase) AddExpectedTxs(expected []*core.BridgeExpectedSolanaTx) error {
	return bd.AddExpectedTxsUnless(expected, func(tx *bbolt.Tx, expectedTx *core.BridgeExpectedSolanaTx) (bool, error) {
		executed, err := findProcessedExecution(tx, expectedTx.ChainID, expectedTx.Hash[:])

		return executed != nil && !executed.IsInvalid, err
	})
}

func findProcessedExecution(tx *bbolt.Tx, chainID string, actionHash []byte) (*core.ProcessedSolanaTx, error) {
	key, err := cDatabaseaccess.ReadInnerActionTxKey(tx, chainID, actionHash)
	if err != nil || key == nil {
		return nil, err
	}

	// A full signature identifies one row. Older links stored only the first
	// 32 bytes, so check every signature with that prefix for the requested action.
	if len(key) != sha256.Size && len(key) != ed25519.SignatureSize {
		return nil, fmt.Errorf("invalid processed transaction key length %d for inner action %x", len(key), actionHash)
	}

	bucket := tx.Bucket(cDatabaseaccess.ChainBucket(cDatabaseaccess.ProcessedTxsBucket, chainID))
	cursor := bucket.Cursor()

	var match *core.ProcessedSolanaTx

	for k, v := cursor.Seek(key); bytes.HasPrefix(k, key); k, v = cursor.Next() {
		if len(k) != ed25519.SignatureSize {
			continue
		}

		var candidate core.ProcessedSolanaTx
		if err := json.Unmarshal(v, &candidate); err != nil {
			return nil, err
		}

		if !bytes.Equal(candidate.InnerActionHash[:], actionHash) || !bytes.Equal(candidate.TxSignature[:], k) {
			continue
		}

		if match != nil {
			return nil, fmt.Errorf("ambiguous signature prefix %x for inner action %x", key, actionHash)
		}

		match = &candidate
	}

	if match == nil {
		return nil, fmt.Errorf("cannot resolve processed transaction key %x for inner action %x", key, actionHash)
	}

	return match, nil
}
