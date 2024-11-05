package txsprocessor

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	"github.com/Ethernal-Tech/ethgo"
	ethereum_common "github.com/ethereum/go-ethereum/common"
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
	p.stateProcessor.ProcessSavedEvents()
	p.stateProcessor.PersistNew()

	var (
		bridgeClaims     = &core.BridgeClaims{}
		maxClaimsToGroup = p.settings.maxBridgingClaimsToGroup[startChainID]
	)

	p.stateProcessor.Reset()

	p.processAllForChain(bridgeClaims, startChainID, maxClaimsToGroup)

	keys := p.getSortedChainIDs()

	for _, key := range keys {
		if key != startChainID {
			p.processAllForChain(bridgeClaims, key, maxClaimsToGroup)
		}
	}

	if bridgeClaims.Count() > 0 {
		batchTxs, err := p.retrieveTxsForEachBatchFromClaims(bridgeClaims)
		if err != nil {
			p.logger.Error("retrieving txs for submitted batches", "err", err)

			return
		}

		receipt, ok := p.submitClaims(startChainID, bridgeClaims)
		if !ok {
			return
		}

		events, err := p.extractEventsFromReceipt(receipt)
		if err != nil {
			p.logger.Error("extracting events from submit claims receipt", "err", err)
		} else {
			events.BatchExecutionInfo = batchTxs
			p.stateProcessor.ProcessSubmitClaimsEvents(events, bridgeClaims)
		}
	}

	p.stateProcessor.UpdateBridgingRequestStates(bridgeClaims, p.bridgingRequestStateUpdater)
	p.stateProcessor.PersistNew()
}

func (p *TxsProcessorImpl) retrieveTxsForEachBatchFromClaims(
	claims *core.BridgeClaims,
) (result []*core.DBBatchInfoEvent, err error) {
	addInfo := func(chainIDInt uint8, batchID uint64, isFailedClaim bool) error {
		chainID := common.ToStrChainID(chainIDInt)

		txs, err := p.bridgeSubmitter.GetBatchTransactions(chainID, batchID)
		if err != nil {
			return fmt.Errorf("failed to retrieve txs for batch: chainID = %s, batchID = %d, failed = %v, err = %w",
				chainID, batchID, isFailedClaim, err)
		}

		result = append(result, core.NewDBBatchInfoEvent(chainIDInt, batchID, isFailedClaim, txs))

		return nil
	}

	for _, x := range claims.BatchExecutedClaims {
		if err := addInfo(x.ChainId, x.BatchNonceId, false); err != nil {
			return nil, err
		}
	}

	for _, x := range claims.BatchExecutionFailedClaims {
		if err := addInfo(x.ChainId, x.BatchNonceId, true); err != nil {
			return nil, err
		}
	}

	return result, nil
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

func (p *TxsProcessorImpl) extractEventsFromReceipt(receipt *types.Receipt) (*core.SubmitClaimsEvents, error) {
	eventSigs, err := eth.GetSubmitClaimsEventSignatures()
	if err != nil {
		p.logger.Error("failed to get submit claims event signatures", "err", err)

		return nil, err
	}

	notEnoughFundsEventSig := eventSigs[0]

	contract, err := contractbinding.NewBridgeContract(ethereum_common.Address{}, nil)
	if err != nil {
		p.logger.Error("failed to get contractbinding brdge contract", "err", err)

		return nil, err
	}

	events := &core.SubmitClaimsEvents{}

	for _, log := range receipt.Logs {
		if len(log.Topics) == 0 {
			continue
		}

		switch eventSig := ethgo.Hash(log.Topics[0]); eventSig {
		case notEnoughFundsEventSig:
			notEnoughFunds, err := contract.BridgeContractFilterer.ParseNotEnoughFunds(*log)
			if err != nil {
				return nil, fmt.Errorf("failed parsing notEnoughFunds log. err: %w", err)
			}

			events.NotEnoughFunds = append(events.NotEnoughFunds, notEnoughFunds)

			p.logger.Info("NotEnoughFunds event found in submit claims receipt", "event", notEnoughFunds)
		default:
			p.logger.Debug("unsupported event signature", "eventSig", eventSig)
		}
	}

	return events, nil
}
