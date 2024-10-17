package txsprocessor

import (
	"context"
	"sort"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hashicorp/go-hclog"
)

type TxsProcessorImpl struct {
	ctx                         context.Context
	appConfig                   *core.AppConfig
	stateProcessor              core.SpecificChainTxsProcessorState
	settings                    *txsProcessorSettings
	bridgeSubmitter             core.BridgeClaimsSubmitter
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater
	logger                      hclog.Logger
	TickTime                    time.Duration
}

var _ core.TxsProcessor = (*TxsProcessorImpl)(nil)

func NewTxsProcessorImpl(
	ctx context.Context,
	appConfig *core.AppConfig,
	stateProcessor core.SpecificChainTxsProcessorState,
	bridgeSubmitter core.BridgeClaimsSubmitter,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
	logger hclog.Logger,
) *TxsProcessorImpl {
	return &TxsProcessorImpl{
		ctx:                         ctx,
		stateProcessor:              stateProcessor,
		appConfig:                   appConfig,
		settings:                    NewTxsProcessorSettings(appConfig, stateProcessor.GetChainType()),
		bridgeSubmitter:             bridgeSubmitter,
		bridgingRequestStateUpdater: bridgingRequestStateUpdater,
		logger:                      logger,
		TickTime:                    TickTimeMs,
	}
}

func (p *TxsProcessorImpl) Start() {
	p.logger.Debug("Starting TxsProcessor", "chainType", p.stateProcessor.GetChainType())

	// ensure always same order of iterating through bp.appConfig.CardanoChains or .EthChains
	keys := p.getSortedChainIDs()

	for {
		for _, key := range keys {
			select {
			case <-p.ctx.Done():
				return
			case <-time.After(p.TickTime * time.Millisecond):
			}

			p.processAllStartingWithChain(key)
		}
	}
}

func (p *TxsProcessorImpl) getSortedChainIDs() []string {
	keys := make([]string, 0)

	switch p.stateProcessor.GetChainType() {
	case common.ChainTypeCardanoStr:
		for k := range p.appConfig.CardanoChains {
			keys = append(keys, k)
		}
	case common.ChainTypeEVMStr:
		for k := range p.appConfig.EthChains {
			keys = append(keys, k)
		}
	default:
		p.logger.Error("Invalid chainType", "chainType", p.stateProcessor.GetChainType())
	}

	sort.Strings(keys)

	return keys
}

// first process for a specific chainID, to give every chainID the chance
// and then, if max claims not reached, rest of the chains can be processed too
func (p *TxsProcessorImpl) processAllStartingWithChain(
	startChainID string,
) {
	p.stateProcessor.Reset()

	var (
		bridgeClaims     = &core.BridgeClaims{}
		maxClaimsToGroup = p.settings.maxBridgingClaimsToGroup[startChainID]
	)

	p.processAllForChain(bridgeClaims, startChainID, maxClaimsToGroup)

	keys := p.getSortedChainIDs()

	for _, key := range keys {
		if key != startChainID {
			p.processAllForChain(bridgeClaims, key, maxClaimsToGroup)
		}
	}

	if bridgeClaims.Count() > 0 {
		receipt, ok := p.submitClaims(startChainID, bridgeClaims)
		if !ok {
			return
		}

		events := p.processSubmitClaimsReceipt(receipt)
		p.stateProcessor.ProcessSubmitClaimsEvents(events, bridgeClaims)
	}

	p.stateProcessor.PersistNew(bridgeClaims, p.bridgingRequestStateUpdater)
}

func (p *TxsProcessorImpl) processAllForChain(
	bridgeClaims *core.BridgeClaims,
	chainID string,
	maxClaimsToGroup int,
) {
	for priority := uint8(0); priority <= core.LastProcessingPriority; priority++ {
		if !bridgeClaims.CanAddMore(maxClaimsToGroup) {
			break
		}

		p.stateProcessor.RunChecks(
			bridgeClaims, chainID, maxClaimsToGroup, priority)
	}
}

func (p *TxsProcessorImpl) submitClaims(
	startChainID string, bridgeClaims *core.BridgeClaims) (*types.Receipt, bool) {
	p.logger.Info("Submitting bridge claims", "claims", bridgeClaims)

	receipt, err := p.bridgeSubmitter.SubmitClaims(
		bridgeClaims, &eth.SubmitOpts{GasLimitMultiplier: p.settings.gasLimitMultiplier[startChainID]})
	if err != nil {
		p.logger.Error("Failed to submit claims", "err", err)

		p.settings.OnSubmitClaimsFailed(startChainID, bridgeClaims.Count())

		p.logger.Warn("Adjusted submit claims settings",
			"startChainID", startChainID,
			"maxBridgingClaimsToGroup", p.settings.maxBridgingClaimsToGroup[startChainID],
			"gasLimitMultiplier", p.settings.gasLimitMultiplier[startChainID],
		)

		return nil, false
	}

	p.settings.ResetSubmitClaimsSettings(startChainID)

	telemetry.UpdateOracleClaimsSubmitCounter(bridgeClaims.Count()) // update telemetry

	return receipt, true
}

func (p *TxsProcessorImpl) processSubmitClaimsReceipt(_ *types.Receipt) *core.SubmitClaimsEvents {
	// parse receipt for events, and put them in SubmitClaimsEvents structure
	// events: emit NotEnoughFunds(_type, _index, _chainTokenQuantity); where type="BRC"
	// BatchExecutionInfo(
	// 		batchID uint64, claimIdx int, isFailedClaim bool,
	//		txHashes []tuple(bytes32 txSourceHash, txSourceChainID byte))
	return &core.SubmitClaimsEvents{}
}
