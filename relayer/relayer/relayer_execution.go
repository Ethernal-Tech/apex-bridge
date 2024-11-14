package relayer

import (
	"context"
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/hashicorp/go-hclog"
)

type SendTxFunc = func(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, confirmedBatch *eth.ConfirmedBatch,
) error

func RelayerExecute(
	ctx context.Context,
	chainID string,
	bridgeSmartContract eth.IBridgeSmartContract,
	db core.Database,
	sendTx SendTxFunc,
	logger hclog.Logger,
) error {
	confirmedBatch, err := bridgeSmartContract.GetConfirmedBatch(ctx, chainID)
	if err != nil {
		return fmt.Errorf("failed to retrieve confirmed batch for chainID: %s. err: %w", chainID, err)
	}

	lastSubmittedBatchID, err := db.GetLastSubmittedBatchID(chainID)
	if err != nil {
		return fmt.Errorf("failed to get last submitted batch id from db for chainID: %s. err: %w", chainID, err)
	}

	logger.Debug("Signed batch retrieved from contract", "batchID", confirmedBatch.ID, "chainID", chainID,
		"lastSubmittedBatchID", lastSubmittedBatchID)

	receivedBatchID := new(big.Int).SetUint64(confirmedBatch.ID)

	if lastSubmittedBatchID != nil {
		if lastSubmittedBatchID.Cmp(receivedBatchID) == 0 {
			logger.Info("Waiting on new signed batch")

			return nil
		} else if lastSubmittedBatchID.Cmp(receivedBatchID) == 1 {
			return fmt.Errorf("last submitted batch id greater than received: last submitted %s > received %s",
				lastSubmittedBatchID, receivedBatchID)
		}
	} else {
		if receivedBatchID.Sign() == 0 {
			logger.Info("Waiting on new signed batch", "chainID", chainID)

			return nil
		}
	}

	logger.Info("Submitting batch tx", "chainID", chainID, "confirmedBatch", confirmedBatch)

	if err := sendTx(ctx, bridgeSmartContract, confirmedBatch); err != nil {
		return fmt.Errorf("failed to send confirmed batch for chainID: %s. err: %w", chainID, err)
	}

	logger.Info("Transaction successfully submitted", "chainID", chainID, "batchID", confirmedBatch.ID)

	if err := db.AddLastSubmittedBatchID(chainID, receivedBatchID); err != nil {
		return fmt.Errorf("failed to insert last submitted batch id into db: %w", err)
	}

	return nil
}
