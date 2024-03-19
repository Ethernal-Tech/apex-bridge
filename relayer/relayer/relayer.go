package relayer

import (
	"context"
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

	// TODO: Remove - added for testing
	shouldRetreive, err := bridge.ShouldRetreive(ctx, ethTxHelper, r.config.Bridge.SmartContractAddress)
	if err != nil {
		r.logger.Error("Failed to retrieve data from bridge", "err", err)
		// In case of error, reset ethClient to nil to try again in the next iteration.
		r.ethClient = nil
		return
	}

	if !shouldRetreive {
		r.logger.Info("Waiting for submited signed batch")
		return
	}

	// invoke smart contract(s)
	smartContractData, err := bridge.GetSmartContractData(ctx, ethTxHelper, r.config.CardanoChain.ChainId, r.config.Bridge.SmartContractAddress)
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

	// TODO: Remove - added for testing
	// Delay needed because of preview testnet wait time for tx to pass
	// Makes continuous testing possible
	time.Sleep(70 * time.Second)

	// TODO: Remove - added for testing
	err = bridge.ResetShouldRetreive(ctx, ethTxHelper, r.config.Bridge.SmartContractAddress)
	if err != nil {
		r.logger.Error("Failed to reset should retrieve", "err", err)

		r.ethClient = nil
		return
	}
}
