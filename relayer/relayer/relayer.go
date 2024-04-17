package relayer

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
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

func NewRelayer(config *core.RelayerConfiguration,
	bridgeSmartContract eth.IBridgeSmartContract, logger hclog.Logger, operations core.ChainOperations, db core.Database) *RelayerImpl {
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
	confirmedBatch, err := r.bridgeSmartContract.GetConfirmedBatch(ctx, r.config.Base.ChainId)
	if err != nil {
		return fmt.Errorf("failed to retrieve confirmed batch: %v", err)
	}

	r.logger.Info("Signed batch retrieved from contract")

	lastSubmittedBatchId, err := r.db.GetLastSubmittedBatchId(r.config.Base.ChainId)
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
			r.logger.Info("Waiting on new signed batch")
			return nil
		} else if lastSubmittedBatchId.Cmp(receivedBatchId) == 1 {
			return fmt.Errorf("last submitted batch id greater than received: last submitted %v > received %v", lastSubmittedBatchId.String(), receivedBatchId.String())
		}
	}

	if err := r.operations.SendTx(confirmedBatch); err != nil {
		return fmt.Errorf("failed to send confirmed batch: %v", err)
	}

	r.logger.Info("Transaction successfully submitted")

	if err := r.db.AddLastSubmittedBatchId(r.config.Base.ChainId, receivedBatchId); err != nil {
		return fmt.Errorf("failed to insert last submitted batch id into db: %v", err)
	}

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

		txProvider, err := cardanotx.GetTxProvider(cardanoChainConfig.BlockfrostUrl, cardanoChainConfig.BlockfrostAPIKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create tx provider: %w", err)
		}

		operations = NewCardanoChainOperations(txProvider, cardanoChainConfig)
	default:
		return nil, fmt.Errorf("unknown chain type: %s", config.ChainType)
	}

	return operations, nil
}
