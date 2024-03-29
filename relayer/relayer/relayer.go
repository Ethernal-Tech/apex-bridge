package relayer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/hashicorp/go-hclog"
)

type RelayerImpl struct {
	config              *core.RelayerConfiguration
	logger              hclog.Logger
	operations          core.ChainOperations
	bridgeSmartContract eth.IBridgeSmartContract
}

var _ core.Relayer = (*RelayerImpl)(nil)

func NewRelayer(config *core.RelayerConfiguration,
	bridgeSmartContract eth.IBridgeSmartContract, logger hclog.Logger, operations core.ChainOperations) *RelayerImpl {
	return &RelayerImpl{
		config:              config,
		logger:              logger,
		bridgeSmartContract: bridgeSmartContract,
		operations:          operations,
	}
}

func (r *RelayerImpl) Start(ctx context.Context) {
	r.logger.Debug("Relayer started")

	ticker := time.NewTicker(time.Millisecond * time.Duration(r.config.PullTimeMilis))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		if err := r.execute(ctx); err != nil {
			r.logger.Error("execute failed", "err", err)
		}
	}
}

func (r *RelayerImpl) execute(ctx context.Context) error {
	confirmedBatch, err := r.bridgeSmartContract.GetConfirmedBatch(ctx, r.config.Base.ChainId)
	if err != nil {
		return fmt.Errorf("failed to retrieve confirmed batch: %v", err)
	}

	r.logger.Info("Signed batch retrieved from contract")

	if err := r.operations.SendTx(confirmedBatch); err != nil {
		return fmt.Errorf("failed to send confirmed batch: %v", err)
	}

	r.logger.Info("Transaction successfully submited")

	return nil
}

// GetChainSpecificOperations returns the chain-specific operations based on the chain type
func GetChainSpecificOperations(config core.ChainSpecific) (core.ChainOperations, error) {
	var operations core.ChainOperations

	// Create the appropriate chain-specific configuration based on the chain type
	switch strings.ToLower(config.ChainType) {
	case "cardano":
		var cardanoChainConfig core.CardanoChainConfig
		if err := json.Unmarshal(config.Config, &cardanoChainConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Cardano configuration: %v", err)
		}

		operations = NewCardanoChainOperations(cardanoChainConfig)
	default:
		return nil, fmt.Errorf("unknown chain type: %s", config.ChainType)
	}

	return operations, nil
}
