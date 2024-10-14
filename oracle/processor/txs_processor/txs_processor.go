package txsprocessor

import (
	"context"
	"sort"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	"github.com/hashicorp/go-hclog"
)

type TxsProcessorImpl struct {
	ctx                         context.Context
	appConfig                   *core.AppConfig
	chainType                   string
	stateProcessor              core.SpecificChainTxsProcessorState
	settings                    *txsProcessorSettings
	bridgeSubmitter             core.BridgeSubmitter
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater
	logger                      hclog.Logger
	TickTime                    time.Duration
}

var _ core.TxsProcessor = (*TxsProcessorImpl)(nil)

func NewTxsProcessorImpl(
	ctx context.Context,
	appConfig *core.AppConfig,
	chainType string,
	stateProcessor core.SpecificChainTxsProcessorState,
	bridgeSubmitter core.BridgeSubmitter,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
	logger hclog.Logger,
) *TxsProcessorImpl {
	return &TxsProcessorImpl{
		ctx:                         ctx,
		chainType:                   chainType,
		stateProcessor:              stateProcessor,
		appConfig:                   appConfig,
		settings:                    NewTxsProcessorSettings(appConfig),
		bridgeSubmitter:             bridgeSubmitter,
		bridgingRequestStateUpdater: bridgingRequestStateUpdater,
		logger:                      logger,
		TickTime:                    TickTimeMs,
	}
}

func (p *TxsProcessorImpl) Start() {
	p.logger.Debug("Starting TxsProcessor", "chainType", p.chainType)

	for {
		if !p.CheckShouldGenerateClaims() {
			return
		}
	}
}

func (p *TxsProcessorImpl) CheckShouldGenerateClaims() bool {
	// ensure always same order of iterating through bp.appConfig.CardanoChains
	keys := p.getSortedChainIDs()

	for _, key := range keys {
		select {
		case <-p.ctx.Done():
			return false
		case <-time.After(p.TickTime * time.Millisecond):
		}

		p.processAllStartingWithChain(key)
	}

	return true
}

func (p *TxsProcessorImpl) getSortedChainIDs() []string {
	keys := make([]string, 0)

	switch p.chainType {
	case common.ChainTypeCardanoStr:
		for k := range p.appConfig.CardanoChains {
			keys = append(keys, k)
		}
	case common.ChainTypeEVMStr:
		for k := range p.appConfig.EthChains {
			keys = append(keys, k)
		}
	default:
		p.logger.Error("Invalid chainType", "chainType", p.chainType)
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

	if bridgeClaims.Count() > 0 && !p.submitClaims(startChainID, bridgeClaims) {
		return
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

		p.processAllForChainAndPriority(
			bridgeClaims, chainID, maxClaimsToGroup, priority)
	}
}

func (p *TxsProcessorImpl) processAllForChainAndPriority(
	bridgeClaims *core.BridgeClaims,
	chainID string,
	maxClaimsToGroup int,
	priority uint8,
) {
	if !p.stateProcessor.PrepareFirstCheck(chainID, priority) {
		return
	}

	for {
		p.stateProcessor.RunCheck(bridgeClaims, maxClaimsToGroup)

		if !bridgeClaims.CanAddMore(maxClaimsToGroup) {
			return
		}

		if !p.stateProcessor.NextBlockInfo() {
			return
		}
	}
}

func (p *TxsProcessorImpl) submitClaims(startChainID string, bridgeClaims *core.BridgeClaims) bool {
	p.logger.Info("Submitting bridge claims", "claims", bridgeClaims)

	err := p.bridgeSubmitter.SubmitClaims(
		bridgeClaims, &eth.SubmitOpts{GasLimitMultiplier: p.settings.gasLimitMultiplier[startChainID]})
	if err != nil {
		p.logger.Error("Failed to submit claims", "err", err)

		p.settings.OnSubmitClaimsFailed(startChainID, bridgeClaims.Count())

		p.logger.Warn("Adjusted submit claims settings",
			"startChainID", startChainID,
			"maxBridgingClaimsToGroup", p.settings.maxBridgingClaimsToGroup[startChainID],
			"gasLimitMultiplier", p.settings.gasLimitMultiplier[startChainID],
		)

		return false
	}

	p.settings.ResetSubmitClaimsSettings(startChainID)

	telemetry.UpdateOracleClaimsSubmitCounter(bridgeClaims.Count()) // update telemetry

	return true
}
