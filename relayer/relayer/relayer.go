package relayer

import (
	"context"
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
	db                  core.Database
}

var _ core.Relayer = (*RelayerImpl)(nil)

func NewRelayer(
	config *core.RelayerConfiguration, bridgeSmartContract eth.IBridgeSmartContract, logger hclog.Logger,
	operations core.ChainOperations, db core.Database,
) *RelayerImpl {
	return &RelayerImpl{
		config:              config,
		logger:              logger,
		bridgeSmartContract: bridgeSmartContract,
		operations:          operations,
		db:                  db,
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
	return RelayerExecute(
		ctx,
		r.config.Chain.ChainID,
		r.bridgeSmartContract,
		r.db,
		r.operations.SendTx,
		r.logger,
	)
}

// GetChainSpecificOperations returns the chain-specific operations based on the chain type
func GetChainSpecificOperations(config core.ChainConfig) (core.ChainOperations, error) {
	// Create the appropriate chain-specific configuration based on the chain type
	switch strings.ToLower(config.ChainType) {
	case "cardano":
		return NewCardanoChainOperations(config.ChainSpecific)
	default:
		return nil, fmt.Errorf("unknown chain type: %s", config.ChainType)
	}
}
