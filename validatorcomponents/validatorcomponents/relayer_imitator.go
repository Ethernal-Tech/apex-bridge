package validatorcomponents

import (
	"context"
	"math/big"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	relayerCore "github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/apex-bridge/relayer/relayer"
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

	waitTime := time.Millisecond * time.Duration(ri.config.RelayerImitatorPullTimeMilis)

	for {
		select {
		case <-ri.ctx.Done():
			return
		case <-time.After(waitTime):
		}

		for chainID := range ri.config.CardanoChains {
			if err := ri.execute(ri.ctx, chainID); err != nil {
				ri.logger.Error("execute failed", "err", err)
			}
		}
	}
}

func (ri *RelayerImitatorImpl) execute(ctx context.Context, chainID string) error {
	return relayer.RelayerExecute(
		ctx,
		chainID,
		ri.bridgeSmartContract,
		ri.db,
		func(ctx context.Context, confirmedBatch *eth.ConfirmedBatch) error {
			receivedBatchID := new(big.Int).SetUint64(confirmedBatch.ID)

			err := ri.bridgingRequestStateUpdater.SubmittedToDestination(chainID, receivedBatchID.Uint64())
			if err != nil {
				ri.logger.Error(
					"error while updating bridging request states to SubmittedToDestination",
					"destinationChainId", chainID, "batchId", receivedBatchID.Uint64())
			}

			return nil
		},
		ri.logger,
	)
}
