package relayer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/apex-bridge/relayer/bridge"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
)

type RelayerImpl struct {
	config     *core.RelayerConfiguration
	logger     hclog.Logger
	ethClient  *ethclient.Client
	operations core.ChainOperations
}

var _ core.Relayer = (*RelayerImpl)(nil)

func NewRelayer(config *core.RelayerConfiguration, logger hclog.Logger, operations core.ChainOperations) *RelayerImpl {
	return &RelayerImpl{
		config:     config,
		logger:     logger,
		ethClient:  nil,
		operations: operations,
	}
}

func (r *RelayerImpl) Start(ctx context.Context) {
	var (
		timerTime = time.Millisecond * time.Duration(r.config.PullTimeMilis)
	)

	r.logger.Debug("Relayer started")

	timer := time.NewTimer(timerTime)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
		case <-ctx.Done():
			return
		}

		r.execute(ctx)

		timer.Reset(timerTime)
	}
}

func (r *RelayerImpl) execute(ctx context.Context) {
	var (
		err error
	)

	if r.ethClient == nil {
		r.ethClient, err = ethclient.Dial(r.config.Bridge.NodeUrl)
		if err != nil {
			r.logger.Error("Failed to dial bridge", "err", err)
			return
		}
	}

	ethTxHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithClient(r.ethClient))
	if err != nil {
		// In case of error, reset ethClient to nil to try again in the next iteration.
		r.ethClient = nil
		return
	}

	// invoke smart contract(s)
	smartContractData, err := bridge.GetConfirmedBatch(ctx, ethTxHelper, r.config.Base.ChainId, r.config.Bridge.SmartContractAddress)
	if err != nil {
		r.logger.Error("Failed to query bridge sc", "err", err)

		r.ethClient = nil
		return
	}
	r.logger.Info("Signed batch retrieved from contract")

	if err := r.operations.SendTx(smartContractData); err != nil {
		r.logger.Error("failed to send tx", "err", err)
		return
	}

	r.logger.Info("Transaction successfully submited")
}

// GetChainSpecificOperations returns the chain-specific operations based on the chain type
func GetChainSpecificOperations(config core.ChainSpecific) (core.ChainOperations, error) {
	var operations core.ChainOperations

	// Create the appropriate chain-specific configuration based on the chain type
	switch config.ChainType {
	case "Cardano":
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
