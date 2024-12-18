package validatorcomponents

import (
	"context"
	"math/big"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/hashicorp/go-hclog"
)

var apexBridgeAdminScAddress = common.HexToAddress("0xABEF000000000000000000000000000000000006")

type TelemetryWorker struct {
	etxHelperWrapper     *eth.EthHelperWrapper
	cardanoDBs           map[string]indexer.Database
	ethDBs               map[string]eventTrackerStore.EventTrackerStore
	waitTime             time.Duration
	latestBlock          map[string]*indexer.BlockPoint
	latestHotWalletState map[string]*big.Int
	logger               hclog.Logger
}

func NewTelemetryWorker(
	txHelper *eth.EthHelperWrapper,
	cardanoDBs map[string]indexer.Database,
	ethDBs map[string]eventTrackerStore.EventTrackerStore,
	waitTime time.Duration,
	logger hclog.Logger,
) *TelemetryWorker {
	return &TelemetryWorker{
		etxHelperWrapper:     txHelper,
		cardanoDBs:           cardanoDBs,
		ethDBs:               ethDBs,
		latestBlock:          map[string]*indexer.BlockPoint{},
		latestHotWalletState: map[string]*big.Int{},
		logger:               logger,
	}
}

func (ti *TelemetryWorker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(ti.waitTime):
			ti.execute()
		}
	}
}

func (ti *TelemetryWorker) execute() {
	for chainID, db := range ti.cardanoDBs {
		bp, err := db.GetLatestBlockPoint()
		if err != nil {
			ti.logger.Warn("failed to retrieve block point", "chain", chainID, "err", err)
		} else if cache := ti.latestBlock[chainID]; cache == nil ||
			cache.BlockHash != bp.BlockHash || cache.BlockSlot != bp.BlockSlot {
			ti.latestBlock[chainID] = bp

			telemetry.UpdateIndexersBlockCounter(chainID, 1)
		}
	}

	for chainID, db := range ti.ethDBs {
		blockNumber, err := db.GetLastProcessedBlock()
		if err != nil {
			ti.logger.Warn("failed to retrieve latest processed block", "chain", chainID, "err", err)
		} else if cache := ti.latestBlock[chainID]; cache == nil || cache.BlockNumber != blockNumber {
			ti.latestBlock[chainID] = &indexer.BlockPoint{
				BlockNumber: blockNumber,
			}

			telemetry.UpdateIndexersBlockCounter(chainID, 1)
		}
	}

	ethTxHelper, err := ti.etxHelperWrapper.GetEthHelper()
	if err != nil {
		ti.logger.Warn("failed to create eth helper", "err", err)

		return
	}

	contract, err := contractbinding.NewAdminContract(
		apexBridgeAdminScAddress,
		ethTxHelper.GetClient())
	if err != nil {
		ti.logger.Warn("failed to create contract", "err", err)

		return
	}

	for chainID := range ti.cardanoDBs {
		val, err := contract.GetChainTokenQuantity(&bind.CallOpts{}, common.ToNumChainID(chainID))
		if err != nil {
			ti.logger.Warn("failed to retrieve hot wallet state", "chain", chainID, "err", err)
		} else if cache := ti.latestHotWalletState[chainID]; cache == nil || cache.Cmp(val) != 0 {
			ti.latestHotWalletState[chainID] = val

			telemetry.UpdateHotWalletState(chainID, val)
		}
	}

	for chainID := range ti.ethDBs {
		val, err := contract.GetChainTokenQuantity(&bind.CallOpts{}, common.ToNumChainID(chainID))
		if err != nil {
			ti.logger.Warn("failed to retrieve hot wallet state", "chain", chainID, "err", err)
		} else if cache := ti.latestHotWalletState[chainID]; cache == nil || cache.Cmp(val) != 0 {
			ti.latestHotWalletState[chainID] = val

			telemetry.UpdateHotWalletState(chainID, val)
		}
	}
}
