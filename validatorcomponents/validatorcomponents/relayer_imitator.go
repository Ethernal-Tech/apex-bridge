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
	config                      *core.AppConfig
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater
	bridgeSmartContract         eth.IBridgeSmartContract
	db                          relayerCore.Database
	logger                      hclog.Logger
	cancelCtx                   context.CancelFunc
}

var _ core.RelayerImitator = (*RelayerImitatorImpl)(nil)

func NewRelayerImitator(
	config *core.AppConfig,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
	bridgeSmartContract eth.IBridgeSmartContract,
	db relayerCore.Database,
	logger hclog.Logger,
) (*RelayerImitatorImpl, error) {
	return &RelayerImitatorImpl{
		config:                      config,
		bridgingRequestStateUpdater: bridgingRequestStateUpdater,
		bridgeSmartContract:         bridgeSmartContract,
		db:                          db,
		logger:                      logger,
	}, nil
}

// Start implements core.RelayerImitator.
func (ri *RelayerImitatorImpl) Start() error {
	ctx, cancelCtx := context.WithCancel(context.Background())
	ri.cancelCtx = cancelCtx

	ri.logger.Debug("Relayer imitator started")

	ticker := time.NewTicker(time.Millisecond * time.Duration(ri.config.RelayerImitatorPullTimeMilis))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}

		for chainId := range ri.config.CardanoChains {
			if err := ri.execute(ctx, chainId); err != nil {
				ri.logger.Error("execute failed", "err", err)
			}
		}
	}
}

func (ri *RelayerImitatorImpl) execute(ctx context.Context, chainId string) error {
	confirmedBatch, err := ri.bridgeSmartContract.GetConfirmedBatch(ctx, chainId)
	if err != nil {
		return fmt.Errorf("failed to retrieve confirmed batch: %v", err)
	}

	ri.logger.Info("Signed batch retrieved from contract")

	lastSubmittedBatchId, err := ri.db.GetLastSubmittedBatchId(chainId)
	if err != nil {
		return fmt.Errorf("failed to get last submitted batch id from db: %v", err)
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
			return fmt.Errorf("last submitted batch id greater than received: last submitted %v > received %v", lastSubmittedBatchId.String(), receivedBatchId.String())
		}
	}

	err = ri.bridgingRequestStateUpdater.SubmittedToDestination(chainId, receivedBatchId.Uint64())
	if err != nil {
		ri.logger.Error("error while updating bridging request states to SubmittedToDestination", "destinationChainId", chainId, "batchId", receivedBatchId.Uint64())
	}

	ri.logger.Info("Transaction successfully submitted")

	if err := ri.db.AddLastSubmittedBatchId(chainId, receivedBatchId); err != nil {
		return fmt.Errorf("failed to insert last submitted batch id into db: %v", err)
	}

	return nil
}

// Stop implements core.RelayerImitator.
func (ri *RelayerImitatorImpl) Stop() error {
	ri.cancelCtx()
	ri.logger.Debug("Relayer imitator stopped")

	return nil
}
