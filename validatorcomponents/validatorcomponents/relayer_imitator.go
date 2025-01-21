package validatorcomponents

import (
	"context"
	"fmt"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	relayerCore "github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/apex-bridge/relayer/relayer"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/hashicorp/go-hclog"
)

type RelayerImitatorImpl struct {
	config                      *core.AppConfig
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater
	bridgeSmartContract         eth.IBridgeSmartContract
	db                          relayerCore.Database
	logger                      hclog.Logger
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
func (ri *RelayerImitatorImpl) Start(ctx context.Context) {
	ri.logger.Debug("Relayer imitator started")

	waitTime := time.Millisecond * time.Duration(ri.config.RelayerImitatorPullTimeMilis)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(waitTime):
		}

		for chainID := range ri.config.CardanoChains {
			if err := ri.execute(ctx, chainID); err != nil {
				ri.logger.Error("execute failed", "err", err)
			}
		}

		for chainID := range ri.config.EthChains {
			if err := ri.execute(ctx, chainID); err != nil {
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
		func(ctx context.Context, sc eth.IBridgeSmartContract, confirmedBatch *eth.ConfirmedBatch) error {
			txs, err := sc.GetConfirmedTransactions(ctx, chainID)
			if err != nil {
				return fmt.Errorf("failed to retrieve confirmed txs for (%s, %d): %w",
					chainID, confirmedBatch.ID, err)
			}

			txsKeys := make([]common.BridgingRequestStateKey, len(txs))
			for i, tx := range txs {
				txsKeys[i] = common.NewBridgingRequestStateKey(
					common.ToStrChainID(tx.SourceChainId), tx.ObservedTransactionHash)
			}

			err = ri.bridgingRequestStateUpdater.SubmittedToDestination(txsKeys)
			if err != nil {
				return fmt.Errorf("failed to update bridging request states (%s, %d) to SubmittedToDestination: %w",
					chainID, confirmedBatch.ID, err)
			}

			return nil
		},
		ri.logger,
	)
}
