package databaseaccess

import (
	"encoding/json"
	"fmt"

	cDatabaseaccess "github.com/Ethernal-Tech/apex-bridge/oracle_common/database_access"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	"github.com/Ethernal-Tech/ethgo"
	"go.etcd.io/bbolt"
)

type BBoltDatabase struct {
	cDatabaseaccess.BBoltDBBase[
		*core.EthTx,
		*core.ProcessedEthTx,
		*core.BridgeExpectedEthTx,
	]
}

var (
	processedTxsByInnerActionBucket = []byte("ProcessedTxsByInnerAction")
)

var _ core.Database = (*BBoltDatabase)(nil)

func (bd *BBoltDatabase) Init(filePath string) error {
	return bd.BBoltDBBase.Init(filePath, [][]byte{processedTxsByInnerActionBucket})
}

func (bd *BBoltDatabase) GetProcessedTx(
	chainID string, txHash ethgo.Hash,
) (result *core.ProcessedEthTx, err error) {
	return bd.BBoltDBBase.GetProcessedTx(chainID, core.ToEthTxKey(chainID, txHash))
}

func (bd *BBoltDatabase) GetProcessedTxByInnerActionTxHash(
	chainID string, innerActionTxHash ethgo.Hash,
) (result *core.ProcessedEthTx, err error) {
	err = bd.DB.View(func(tx *bbolt.Tx) error {
		var processedTxByInnerAction *core.ProcessedEthTxByInnerAction
		if data := tx.Bucket(processedTxsByInnerActionBucket).Get(
			core.ToEthTxKey(chainID, innerActionTxHash)); len(data) > 0 {
			if err := json.Unmarshal(data, &processedTxByInnerAction); err != nil {
				return err
			}

			if data := tx.Bucket(cDatabaseaccess.ProcessedTxsBucket).Get(
				core.ToEthTxKey(chainID, processedTxByInnerAction.Hash)); len(data) > 0 {
				return json.Unmarshal(data, &result)
			}
		}

		return nil
	})

	return result, err
}

func (bd *BBoltDatabase) UpdateTxs(data *core.EthUpdateTxsData) error {
	return bd.BBoltDBBase.UpdateTxs(
		data,
		func(tx *bbolt.Tx) error {
			var newProcessed []*core.ProcessedEthTx
			newProcessed = append(newProcessed, data.MoveUnprocessedToProcessed...)
			newProcessed = append(newProcessed, data.MovePendingToProcessed...)

			for _, processedTx := range newProcessed {
				innerActionTxBytes, err := json.Marshal(processedTx.ToProcessedTxByInnerAction())
				if err != nil {
					return fmt.Errorf("could not marshal processed tx by inner action: %w", err)
				}

				if err = tx.Bucket(processedTxsByInnerActionBucket).Put(
					processedTx.KeyByInnerAction(), innerActionTxBytes); err != nil {
					return fmt.Errorf("processed tx by inner action write error: %w", err)
				}
			}

			return nil
		},
	)
}
