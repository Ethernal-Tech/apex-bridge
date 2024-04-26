package relayer

import (
	"context"
	"fmt"
	"math/big"
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
	confirmedBatch, err := r.bridgeSmartContract.GetConfirmedBatch(ctx, r.config.Chain.ChainID)
	if err != nil {
		return fmt.Errorf("failed to retrieve confirmed batch: %w", err)
	}

	r.logger.Info("Signed batch retrieved from contract")

	lastSubmittedBatchID, err := r.db.GetLastSubmittedBatchID(r.config.Chain.ChainID)
	if err != nil {
		return fmt.Errorf("failed to get last submitted batch id from db: %w", err)
	}

	receivedBatchID, ok := new(big.Int).SetString(confirmedBatch.ID, 0)
	if !ok {
		return fmt.Errorf("failed to convert confirmed batch id to big int")
	}

	if lastSubmittedBatchID != nil {
		if lastSubmittedBatchID.Cmp(receivedBatchID) == 0 {
			r.logger.Info("Waiting on new signed batch")

			return nil
		} else if lastSubmittedBatchID.Cmp(receivedBatchID) == 1 {
			return fmt.Errorf("last submitted batch id greater than received: last submitted %s > received %s",
				lastSubmittedBatchID, receivedBatchID)
		}
	}

	if err := r.operations.SendTx(confirmedBatch); err != nil {
		return fmt.Errorf("failed to send confirmed batch: %w", err)
	}

	r.logger.Info("Transaction successfully submitted")

	if err := r.db.AddLastSubmittedBatchID(r.config.Chain.ChainID, receivedBatchID); err != nil {
		return fmt.Errorf("failed to insert last submitted batch id into db: %w", err)
	}

	return nil
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
