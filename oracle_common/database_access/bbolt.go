package databaseaccess

import (
	"encoding/json"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/ethgo"
	"go.etcd.io/bbolt"
)

const defaultKey = "defaultKey"

type BBoltDBBase[
	TTx core.BaseTx,
	TProcessedTx core.BaseProcessedTx,
	TExpectedTx core.BaseExpectedTx,
] struct {
	DB              *bbolt.DB
	SupportedChains map[string]bool
	TypeRegister    common.TypeRegister
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) GetUnprocessedTxs(
	chainID string, priority uint8, threshold int,
) ([]TTx, error) {
	var result []TTx

	if supported := bd.SupportedChains[chainID]; !supported {
		return nil, fmt.Errorf("unsupported chain: %s", chainID)
	}

	err := bd.DB.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(ChainBucket(UnprocessedTxsBucket, chainID)).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var chainTx TTx

			if err := json.Unmarshal(v, &chainTx); err != nil {
				return err
			}

			if chainTx.GetPriority() == priority {
				result = append(result, chainTx)
				if threshold > 0 && len(result) == threshold {
					break
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) GetAllUnprocessedTxs(
	chainID string, threshold int,
) ([]TTx, error) {
	var result []TTx

	if supported := bd.SupportedChains[chainID]; !supported {
		return nil, fmt.Errorf("unsupported chain: %s", chainID)
	}

	err := bd.DB.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(ChainBucket(UnprocessedTxsBucket, chainID)).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var chainTx TTx

			if err := json.Unmarshal(v, &chainTx); err != nil {
				return err
			}

			result = append(result, chainTx)
			if threshold > 0 && len(result) == threshold {
				break
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) GetPendingTx(
	entityID core.DBTxID,
) (result core.BaseTx, err error) {
	err = bd.DB.View(func(tx *bbolt.Tx) (err error) {
		data := tx.Bucket(ChainBucket(PendingTxsBucket, entityID.ChainID)).Get(entityID.DBKey)
		if len(data) == 0 {
			return fmt.Errorf("couldn't get pending tx for entityID: %v", entityID)
		}

		result, err = common.GetRegisteredTypeInstance[core.BaseTx](bd.TypeRegister, entityID.ChainID)
		if err != nil {
			return err
		}

		return json.Unmarshal(data, &result)
	})

	return result, err
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) GetProcessedTx(
	entityID core.DBTxID,
) (result TProcessedTx, err error) {
	if supported := bd.SupportedChains[entityID.ChainID]; !supported {
		return result, fmt.Errorf("unsupported chain: %s", entityID.ChainID)
	}

	err = bd.DB.View(func(tx *bbolt.Tx) error {
		data := tx.Bucket(ChainBucket(ProcessedTxsBucket, entityID.ChainID)).Get(entityID.DBKey)
		if len(data) > 0 {
			return json.Unmarshal(data, &result)
		}

		return nil
	})

	return result, err
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) GetProcessedTxByInnerActionTxHash(
	chainID string, innerActionTxHash []byte,
) (result TProcessedTx, err error) {
	if supported := bd.SupportedChains[chainID]; !supported {
		return result, fmt.Errorf("unsupported chain: %s", chainID)
	}

	err = bd.DB.View(func(tx *bbolt.Tx) error {
		var processedTxByInnerAction *core.ProcessedTxByInnerAction

		midBucket := tx.Bucket(ChainBucket(ProcessedTxsByInnerActionBucket, chainID))
		if data := midBucket.Get(innerActionTxHash); len(data) > 0 {
			if err := json.Unmarshal(data, &processedTxByInnerAction); err != nil {
				return err
			}

			bucket := tx.Bucket(ChainBucket(ProcessedTxsBucket, chainID))
			if data := bucket.Get(processedTxByInnerAction.Hash[:]); len(data) > 0 {
				return json.Unmarshal(data, &result)
			}
		}

		return nil
	})

	return result, err
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) AddTxs(
	processedTxs []TProcessedTx, unprocessedTxs []TTx,
) error {
	return bd.DB.Update(func(tx *bbolt.Tx) error {
		for _, processedTx := range processedTxs {
			if supported := bd.SupportedChains[processedTx.GetChainID()]; !supported {
				return fmt.Errorf("unsupported chain: %s", processedTx.GetChainID())
			}

			processedBucket := tx.Bucket(ChainBucket(ProcessedTxsBucket, processedTx.GetChainID()))

			bytes, err := json.Marshal(processedTx)
			if err != nil {
				return fmt.Errorf("could not marshal processed tx: %w", err)
			}

			if err = processedBucket.Put(processedTx.GetTxHash(), bytes); err != nil {
				return fmt.Errorf("processed tx write error: %w", err)
			}
		}

		for _, unprocessedTx := range unprocessedTxs {
			if supported := bd.SupportedChains[unprocessedTx.GetChainID()]; !supported {
				return fmt.Errorf("unsupported chain: %s", unprocessedTx.GetChainID())
			}

			unprocessedBucket := tx.Bucket(ChainBucket(UnprocessedTxsBucket, unprocessedTx.GetChainID()))
			bytes, err := json.Marshal(unprocessedTx)
			if err != nil {
				return fmt.Errorf("could not marshal unprocessed tx: %w", err)
			}

			if err = unprocessedBucket.Put(unprocessedTx.UnprocessedDBKey(), bytes); err != nil {
				return fmt.Errorf("unprocessed tx write error: %w", err)
			}
		}

		return nil
	})
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) ClearAllTxs(chainID string) error {
	if supported := bd.SupportedChains[chainID]; !supported {
		return fmt.Errorf("unsupported chain: %s", chainID)
	}

	return bd.DB.Update(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(ChainBucket(UnprocessedTxsBucket, chainID)).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var unprocessedTx TTx

			if err := json.Unmarshal(v, &unprocessedTx); err != nil {
				return err
			}

			err := cursor.Bucket().Delete(unprocessedTx.UnprocessedDBKey())
			if err != nil {
				return err
			}
		}

		cursor = tx.Bucket(ChainBucket(PendingTxsBucket, chainID)).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var pendingTx TTx

			if err := json.Unmarshal(v, &pendingTx); err != nil {
				return err
			}

			err := cursor.Bucket().Delete(pendingTx.GetTxHash())
			if err != nil {
				return err
			}
		}

		cursor = tx.Bucket(ChainBucket(ExpectedTxsBucket, chainID)).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var expectedTx TExpectedTx

			if err := json.Unmarshal(v, &expectedTx); err != nil {
				return err
			}

			if !expectedTx.GetIsInvalid() && !expectedTx.GetIsProcessed() {
				if err := cursor.Bucket().Delete(expectedTx.DBKey()); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) AddExpectedTxs(expectedTxs []TExpectedTx) error {
	return bd.DB.Update(func(tx *bbolt.Tx) error {
		for _, expectedTx := range expectedTxs {
			if supported := bd.SupportedChains[expectedTx.GetChainID()]; !supported {
				return fmt.Errorf("unsupported chain: %s", expectedTx.GetChainID())
			}

			bucket := tx.Bucket(ChainBucket(ExpectedTxsBucket, expectedTx.GetChainID()))
			key := expectedTx.DBKey()

			if data := bucket.Get(key); len(data) == 0 {
				bytes, err := json.Marshal(expectedTx)
				if err != nil {
					return fmt.Errorf("could not marshal expected tx: %w", err)
				}

				if err = bucket.Put(key, bytes); err != nil {
					return fmt.Errorf("expected tx write error: %w", err)
				}
			}
		}

		return nil
	})
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) GetExpectedTxs(
	chainID string, priority uint8, threshold int,
) ([]TExpectedTx, error) {
	if supported := bd.SupportedChains[chainID]; !supported {
		return nil, fmt.Errorf("unsupported chain: %s", chainID)
	}

	var result []TExpectedTx

	err := bd.DB.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(ChainBucket(ExpectedTxsBucket, chainID)).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var expectedTx TExpectedTx

			if err := json.Unmarshal(v, &expectedTx); err != nil {
				return err
			}

			if expectedTx.GetPriority() == priority &&
				!expectedTx.GetIsProcessed() && !expectedTx.GetIsInvalid() {
				result = append(result, expectedTx)
				if threshold > 0 && len(result) == threshold {
					break
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) GetAllExpectedTxs(
	chainID string, threshold int,
) ([]TExpectedTx, error) {
	if supported := bd.SupportedChains[chainID]; !supported {
		return nil, fmt.Errorf("unsupported chain: %s", chainID)
	}

	var result []TExpectedTx

	err := bd.DB.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(ChainBucket(ExpectedTxsBucket, chainID)).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var expectedTx TExpectedTx

			if err := json.Unmarshal(v, &expectedTx); err != nil {
				return err
			}

			if !expectedTx.GetIsProcessed() && !expectedTx.GetIsInvalid() {
				result = append(result, expectedTx)
				if threshold > 0 && len(result) == threshold {
					break
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) GetUnprocessedBatchEvents(
	chainID string,
) ([]*core.DBBatchInfoEvent, error) {
	if supported := bd.SupportedChains[chainID]; !supported {
		return nil, fmt.Errorf("unsupported chain: %s", chainID)
	}

	var result []*core.DBBatchInfoEvent

	err := bd.DB.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(ChainBucket(UnprocessedBatchEventsBucket, chainID)).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var batchInfo *core.DBBatchInfoEvent

			if err := json.Unmarshal(v, &batchInfo); err != nil {
				return err
			}

			result = append(result, batchInfo)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) GetBlocksSubmitterInfo(
	chainID string,
) (result core.BlocksSubmitterInfo, err error) {
	if supported := bd.SupportedChains[chainID]; !supported {
		return result, fmt.Errorf("unsupported chain: %s", chainID)
	}

	err = bd.DB.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(ChainBucket(BlocksSubmitterBucket, chainID))

		if data := bucket.Get([]byte(defaultKey)); len(data) > 0 {
			if err := json.Unmarshal(data, &result); err != nil {
				return fmt.Errorf("could not read blocks submitter info for chain %s: %w", chainID, err)
			}
		}

		return nil
	})

	return result, err
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) SetBlocksSubmitterInfo(
	chainID string, info core.BlocksSubmitterInfo,
) error {
	if supported := bd.SupportedChains[chainID]; !supported {
		return fmt.Errorf("unsupported chain: %s", chainID)
	}

	bytes, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("could not serialize blocks submitter info for chain %s: %w", chainID, err)
	}

	return bd.DB.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(ChainBucket(BlocksSubmitterBucket, chainID))

		if err := bucket.Put([]byte(defaultKey), bytes); err != nil {
			return fmt.Errorf("could not save block submitter info for chain %s: %w", chainID, err)
		}

		return nil
	})
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) UpdateTxs(
	data *core.UpdateTxsData[TTx, TProcessedTx, TExpectedTx], chainIDConverter *common.ChainIDConverter,
) error {
	return bd.DB.Update(func(tx *bbolt.Tx) error {
		err := bd.markExpectedTxs(tx, data.ExpectedInvalid, func(expectedTx TExpectedTx) {
			expectedTx.SetInvalid()
		})
		if err != nil {
			return err
		}

		err = bd.markExpectedTxs(tx, data.ExpectedProcessed, func(expectedTx TExpectedTx) {
			expectedTx.SetProcessed()
		})
		if err != nil {
			return err
		}

		if err := bd.updateUnprocessed(tx, data.UpdateUnprocessed); err != nil {
			return err
		}

		if err := bd.moveUnprocessedToPending(tx, data.MoveUnprocessedToPending); err != nil {
			return err
		}

		if err := bd.moveUnprocessedToProcessed(tx, data.MoveUnprocessedToProcessed); err != nil {
			return err
		}

		if err := bd.movePendingToUnprocessed(tx, data.MovePendingToUnprocessed); err != nil {
			return err
		}

		if err := bd.movePendingToProcessed(tx, data.MovePendingToProcessed); err != nil {
			return err
		}

		if err := bd.removeBatchInfoEvents(tx, data.RemoveBatchInfoEvents, chainIDConverter); err != nil {
			return err
		}

		if err := bd.addBatchInfoEvents(tx, data.AddBatchInfoEvents, chainIDConverter); err != nil {
			return err
		}

		err = bd.handleInnerActionLink(tx, data.MoveUnprocessedToProcessed, data.MovePendingToProcessed)
		if err != nil {
			return err
		}

		return nil
	})
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) markExpectedTxs(
	tx *bbolt.Tx, expectedTxs []TExpectedTx, markFunc func(expectedTx TExpectedTx),
) error {
	for _, expectedTx := range expectedTxs {
		if supported := bd.SupportedChains[expectedTx.GetChainID()]; !supported {
			return fmt.Errorf("unsupported chain: %s", expectedTx.GetChainID())
		}

		bucket := tx.Bucket(ChainBucket(ExpectedTxsBucket, expectedTx.GetChainID()))
		key := expectedTx.DBKey()

		if data := bucket.Get(key); len(data) > 0 {
			var dbExpectedTx TExpectedTx

			if err := json.Unmarshal(data, &dbExpectedTx); err != nil {
				return err
			}

			markFunc(dbExpectedTx)

			bytes, err := json.Marshal(dbExpectedTx)
			if err != nil {
				return fmt.Errorf("could not marshal db expected tx: %w", err)
			}

			if err := bucket.Put(key, bytes); err != nil {
				return fmt.Errorf("db expected tx write error: %w", err)
			}
		}
	}

	return nil
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) updateUnprocessed(
	tx *bbolt.Tx, unprocessedTxs []TTx,
) error {
	for _, unprocessedTx := range unprocessedTxs {
		if supported := bd.SupportedChains[unprocessedTx.GetChainID()]; !supported {
			return fmt.Errorf("unsupported chain: %s", unprocessedTx.GetChainID())
		}

		unprocessedBucket := tx.Bucket(ChainBucket(UnprocessedTxsBucket, unprocessedTx.GetChainID()))

		bytes, err := json.Marshal(unprocessedTx)
		if err != nil {
			return fmt.Errorf("could not marshal unprocessed tx: %w", err)
		}

		if err = unprocessedBucket.Put(unprocessedTx.UnprocessedDBKey(), bytes); err != nil {
			return fmt.Errorf("unprocessed tx write error: %w", err)
		}
	}

	return nil
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) moveUnprocessedToPending(
	tx *bbolt.Tx, unprocessedTxs []TTx,
) error {
	for _, unprocessedTx := range unprocessedTxs {
		if supported := bd.SupportedChains[unprocessedTx.GetChainID()]; !supported {
			return fmt.Errorf("unsupported chain: %s", unprocessedTx.GetChainID())
		}

		unprocessedBucket := tx.Bucket(ChainBucket(UnprocessedTxsBucket, unprocessedTx.GetChainID()))
		pendingBucket := tx.Bucket(ChainBucket(PendingTxsBucket, unprocessedTx.GetChainID()))

		bytes, err := json.Marshal(unprocessedTx)
		if err != nil {
			return fmt.Errorf("could not marshal pending tx: %w", err)
		}

		if err = pendingBucket.Put(unprocessedTx.GetTxHash(), bytes); err != nil {
			return fmt.Errorf("pending tx write error: %w", err)
		}

		if err := unprocessedBucket.Delete(unprocessedTx.UnprocessedDBKey()); err != nil {
			return fmt.Errorf("could not remove from unprocessed txs: %w", err)
		}
	}

	return nil
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) moveUnprocessedToProcessed(
	tx *bbolt.Tx, unprocessedTxs []TProcessedTx,
) error {
	for _, unprocessedTx := range unprocessedTxs {
		if supported := bd.SupportedChains[unprocessedTx.GetChainID()]; !supported {
			return fmt.Errorf("unsupported chain: %s", unprocessedTx.GetChainID())
		}

		unprocessedBucket := tx.Bucket(ChainBucket(UnprocessedTxsBucket, unprocessedTx.GetChainID()))
		processedBucket := tx.Bucket(ChainBucket(ProcessedTxsBucket, unprocessedTx.GetChainID()))

		bytes, err := json.Marshal(unprocessedTx)
		if err != nil {
			return fmt.Errorf("could not marshal processed tx: %w", err)
		}

		if err = processedBucket.Put(unprocessedTx.GetTxHash(), bytes); err != nil {
			return fmt.Errorf("processed tx write error: %w", err)
		}

		if err := unprocessedBucket.Delete(unprocessedTx.UnprocessedDBKey()); err != nil {
			return fmt.Errorf("could not remove from unprocessed txs: %w", err)
		}
	}

	return nil
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) movePendingToUnprocessed(
	tx *bbolt.Tx, pendingTxs []core.BaseTx,
) error {
	for _, pendingTx := range pendingTxs {
		unprocessedBucket := tx.Bucket(ChainBucket(UnprocessedTxsBucket, pendingTx.GetChainID()))
		pendingBucket := tx.Bucket(ChainBucket(PendingTxsBucket, pendingTx.GetChainID()))

		bytes, err := json.Marshal(pendingTx)
		if err != nil {
			return fmt.Errorf("could not marshal unprocessed tx: %w", err)
		}

		if err = unprocessedBucket.Put(pendingTx.UnprocessedDBKey(), bytes); err != nil {
			return fmt.Errorf("unprocessed tx write error: %w", err)
		}

		if err := pendingBucket.Delete(pendingTx.GetTxHash()); err != nil {
			return fmt.Errorf("could not remove from pending txs: %w", err)
		}
	}

	return nil
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) movePendingToProcessed(
	tx *bbolt.Tx, pendingTxs []core.BaseProcessedTx,
) error {
	for _, pendingTx := range pendingTxs {
		processedBucket := tx.Bucket(ChainBucket(ProcessedTxsBucket, pendingTx.GetChainID()))
		pendingBucket := tx.Bucket(ChainBucket(PendingTxsBucket, pendingTx.GetChainID()))

		bytes, err := json.Marshal(pendingTx)
		if err != nil {
			return fmt.Errorf("could not marshal processed tx: %w", err)
		}

		if err = processedBucket.Put(pendingTx.GetTxHash(), bytes); err != nil {
			return fmt.Errorf("processed tx write error: %w", err)
		}

		if err := pendingBucket.Delete(pendingTx.GetTxHash()); err != nil {
			return fmt.Errorf("could not remove from pending txs: %w", err)
		}
	}

	return nil
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) addBatchInfoEvents(
	tx *bbolt.Tx, batchInfoEvents []*core.DBBatchInfoEvent, chainIDConverter *common.ChainIDConverter,
) error {
	for _, evt := range batchInfoEvents {
		chainID := chainIDConverter.ToStrChainID(evt.DstChainID)
		if supported := bd.SupportedChains[chainID]; !supported {
			return fmt.Errorf("unsupported chain: %s", chainID)
		}

		unprocessedBatchEventsBucket := tx.Bucket(ChainBucket(
			UnprocessedBatchEventsBucket, chainID))

		bytes, err := json.Marshal(evt)
		if err != nil {
			return fmt.Errorf("could not marshal unprocessed batch event: %w", err)
		}

		if err = unprocessedBatchEventsBucket.Put(evt.DBKey(), bytes); err != nil {
			return fmt.Errorf("unprocessed batch event write error: %w", err)
		}
	}

	return nil
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) removeBatchInfoEvents(
	tx *bbolt.Tx, batchInfoEvents []*core.DBBatchInfoEvent, chainIDConverter *common.ChainIDConverter,
) error {
	for _, evt := range batchInfoEvents {
		chainID := chainIDConverter.ToStrChainID(evt.DstChainID)
		if supported := bd.SupportedChains[chainID]; !supported {
			return fmt.Errorf("unsupported chain: %s", chainID)
		}

		unprocessedBatchEventsBucket := tx.Bucket(ChainBucket(
			UnprocessedBatchEventsBucket, chainID))

		if err := unprocessedBatchEventsBucket.Delete(evt.DBKey()); err != nil {
			return fmt.Errorf("could not remove unprocessed batch event: %w", err)
		}
	}

	return nil
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) handleInnerActionLink(
	tx *bbolt.Tx, moveUnprocessedToProcessed []TProcessedTx,
	movePendingToProcessed []core.BaseProcessedTx,
) error {
	links := make([]*core.ProcessedTxByInnerAction, 0, len(moveUnprocessedToProcessed)+len(movePendingToProcessed))

	for _, tx := range moveUnprocessedToProcessed {
		if tx.HasInnerActionTxHash() {
			links = append(links, &core.ProcessedTxByInnerAction{
				ChainID:         tx.GetChainID(),
				Hash:            ethgo.Hash(tx.GetTxHash()),
				InnerActionHash: ethgo.Hash(tx.GetInnerActionTxHash()),
			})
		}
	}

	for _, tx := range movePendingToProcessed {
		if tx.HasInnerActionTxHash() {
			links = append(links, &core.ProcessedTxByInnerAction{
				ChainID:         tx.GetChainID(),
				Hash:            ethgo.Hash(tx.GetTxHash()),
				InnerActionHash: ethgo.Hash(tx.GetInnerActionTxHash()),
			})
		}
	}

	for _, link := range links {
		innerActionTxBytes, err := json.Marshal(link)
		if err != nil {
			return fmt.Errorf("could not marshal processed tx by inner action: %w", err)
		}

		bucket := tx.Bucket(ChainBucket(ProcessedTxsByInnerActionBucket, link.ChainID))
		if err = bucket.Put(link.InnerActionHash[:], innerActionTxBytes); err != nil {
			return fmt.Errorf("processed tx by inner action write error: %w", err)
		}
	}

	return nil
}
