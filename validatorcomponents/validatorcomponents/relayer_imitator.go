package validatorcomponents

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	relayerCore "github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/hashicorp/go-hclog"
)

type RelayerImitatorImpl struct {
	ctx                         context.Context
	config                      *core.AppConfig
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater
	bridgeSmartContract         eth.IBridgeSmartContract
	db                          relayerCore.Database
	logger                      hclog.Logger
}

var _ core.RelayerImitator = (*RelayerImitatorImpl)(nil)

func NewRelayerImitator(
	ctx context.Context,
	config *core.AppConfig,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
	bridgeSmartContract eth.IBridgeSmartContract,
	db relayerCore.Database,
	logger hclog.Logger,
) (*RelayerImitatorImpl, error) {
	return &RelayerImitatorImpl{
		ctx:                         ctx,
		config:                      config,
		bridgingRequestStateUpdater: bridgingRequestStateUpdater,
		bridgeSmartContract:         bridgeSmartContract,
		db:                          db,
		logger:                      logger,
	}, nil
}

// Start implements core.RelayerImitator.
func (ri *RelayerImitatorImpl) Start() {

	ri.logger.Debug("Relayer imitator started")

	ticker := time.NewTicker(time.Millisecond * time.Duration(ri.config.RelayerImitatorPullTimeMilis))
	defer ticker.Stop()

	for {
		select {
		case <-ri.ctx.Done():
			return
		case <-ticker.C:
		}

		for chainId := range ri.config.CardanoChains {
			if err := ri.execute(ri.ctx, chainId); err != nil {
				ri.logger.Error("execute failed", "err", err)
			}
		}
	}
}

func (ri *RelayerImitatorImpl) execute(ctx context.Context, chainId string) error {
	confirmedBatch, err := ri.bridgeSmartContract.GetConfirmedBatch(ctx, chainId)
	if err != nil {
		return fmt.Errorf("failed to retrieve confirmed batch: %w", err)
	}

	ri.logger.Info("Signed batch retrieved from contract")

	lastSubmittedBatchId, err := ri.db.GetLastSubmittedBatchId(chainId)
	if err != nil {
		return fmt.Errorf("failed to get last submitted batch id from db: %w", err)
	}

	receivedBatchId := new(big.Int)
	receivedBatchId, ok := receivedBatchId.SetString(confirmedBatch.Id, 10)
	if !ok {
		return fmt.Errorf("failed to convert confirmed batch id to big int")
	}

	if lastSubmittedBatchId != nil {
		if lastSubmittedBatchId.Cmp(receivedBatchId) == 0 {
			ri.logger.Info("Waiting on new signed batch")
			return nil
		} else if lastSubmittedBatchId.Cmp(receivedBatchId) == 1 {
			return fmt.Errorf("last submitted batch id greater than received: last submitted %s > received %s",
				lastSubmittedBatchId, receivedBatchId)
		}
	}

	err = ri.bridgingRequestStateUpdater.SubmittedToDestination(chainId, receivedBatchId.Uint64())
	if err != nil {
		ri.logger.Error(
			"error while updating bridging request states to SubmittedToDestination",
			"destinationChainId", chainId, "batchId", receivedBatchId.Uint64())
	}

	ri.logger.Info("Transaction successfully submitted")

	if err := ri.db.AddLastSubmittedBatchId(chainId, receivedBatchId); err != nil {
		return fmt.Errorf("failed to insert last submitted batch id into db: %w", err)
	}

	return nil
}
