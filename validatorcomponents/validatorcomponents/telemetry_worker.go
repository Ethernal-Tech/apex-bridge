package validatorcomponents

import (
	"context"
	"math/big"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	oracleCommonCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/hashicorp/go-hclog"
)

const (
	feeMetricName         = "fee"
	multisigMetricName    = "multisig"
	nativeTokenMetricName = "nativeToken"
)

var apexBridgeAdminScAddress = common.HexToAddress("0xABEF000000000000000000000000000000000006")

type TelemetryWorker struct {
	etxHelperWrapper                   *eth.EthHelperWrapper
	cardanoDBs                         map[string]indexer.Database
	ethDBs                             map[string]eventTrackerStore.EventTrackerStore
	config                             *oracleCommonCore.AppConfig
	waitTime                           time.Duration
	latestBlock                        map[string]*indexer.BlockPoint
	latestHotWalletState               map[string]*big.Int
	latestHotWalletStateForNativeToken map[string]*big.Int
	latestFeeMultisigState             map[string]uint64
	logger                             hclog.Logger
}

func NewTelemetryWorker(
	txHelper *eth.EthHelperWrapper,
	cardanoDBs map[string]indexer.Database,
	ethDBs map[string]eventTrackerStore.EventTrackerStore,
	config *oracleCommonCore.AppConfig,
	waitTime time.Duration,
	logger hclog.Logger,
) *TelemetryWorker {
	return &TelemetryWorker{
		etxHelperWrapper:                   txHelper,
		cardanoDBs:                         cardanoDBs,
		ethDBs:                             ethDBs,
		config:                             config,
		latestBlock:                        map[string]*indexer.BlockPoint{},
		latestHotWalletState:               map[string]*big.Int{},
		latestHotWalletStateForNativeToken: map[string]*big.Int{},
		latestFeeMultisigState:             map[string]uint64{},
		logger:                             logger,
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

		ti.updateFeeHotWalletState(db, chainID)
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
		if val := ti.getHotWalletState(contract, chainID); val != nil {
			telemetry.UpdateHotWalletState(chainID, multisigMetricName, val.Uint64())
		}
	}

	for chainID := range ti.ethDBs {
		if val := ti.getHotWalletState(contract, chainID); val != nil {
			telemetry.UpdateHotWalletState(chainID, multisigMetricName, val.Uint64())
		}
	}

	if ti.config.RunMode == common.SkylineMode {
		for chainID := range ti.cardanoDBs {
			if val := ti.getHotWalletStateForNativeToken(contract, chainID); val != nil {
				telemetry.UpdateHotWalletState(chainID, nativeTokenMetricName, val.Uint64())
			}
		}
	}
}

func (ti *TelemetryWorker) updateFeeHotWalletState(db indexer.Database, chainID string) {
	txInOuts, err := db.GetAllTxOutputs(ti.config.CardanoChains[chainID].BridgingAddresses.FeeAddress, true)
	if err != nil {
		ti.logger.Warn("failed to retrieve utxos for fee multisig", "chain", chainID, "err", err)
	} else {
		stateSum := uint64(0)

		for _, x := range txInOuts {
			// do not count utxos with tokens - reactor only
			if len(x.Output.Tokens) == 0 {
				stateSum += x.Output.Amount
			}
		}

		if cache := ti.latestFeeMultisigState[chainID]; cache != stateSum {
			telemetry.UpdateHotWalletState(chainID, feeMetricName, stateSum)
		}
	}
}

func (ti *TelemetryWorker) getHotWalletState(
	contract *contractbinding.AdminContract, chainID string,
) (value *big.Int) {
	val, err := contract.GetChainTokenQuantity(&bind.CallOpts{}, common.ToNumChainID(chainID))
	if err != nil {
		ti.logger.Warn("failed to retrieve hot wallet state", "chain", chainID, "err", err)
	} else if cache := ti.latestHotWalletState[chainID]; cache == nil || cache.Cmp(val) != 0 {
		ti.latestHotWalletState[chainID] = val

		value = val
	}

	return value
}

func (ti *TelemetryWorker) getHotWalletStateForNativeToken(contract *contractbinding.AdminContract, chainID string) (value *big.Int) {
	val, err := contract.GetChainWrappedTokenQuantity(&bind.CallOpts{}, common.ToNumChainID(chainID))
	if err != nil {
		ti.logger.Warn("failed to retrieve hot wallet state for native token", "chain", chainID, "err", err)
	} else if cache := ti.latestHotWalletStateForNativeToken[chainID]; cache == nil || cache.Cmp(val) != 0 {
		ti.latestHotWalletStateForNativeToken[chainID] = val
		value = val
	}

	return value
}
