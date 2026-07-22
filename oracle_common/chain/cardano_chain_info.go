package chain

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

const (
	protocolParamsTTL     = 30 * time.Minute
	populateRetryInterval = 5 * time.Second
)

type CardanoChainInfo struct {
	config *cCore.CardanoChainConfig

	ctx        context.Context
	db         cCore.ProtocolParamsDB
	logger     hclog.Logger
	txProvider wallet.ITxProvider

	lock sync.Mutex
}

func NewCardanoChainInfo(
	ctx context.Context, config *cCore.CardanoChainConfig,
	db cCore.ProtocolParamsDB, logger hclog.Logger,
) (*CardanoChainInfo, error) {
	txProvider, err := config.CreateTxProvider()
	if err != nil {
		return nil, err
	}

	info := &CardanoChainInfo{
		config:     config,
		ctx:        ctx,
		db:         db,
		logger:     logger,
		txProvider: txProvider,
	}

	if err := info.initialize(); err != nil {
		return nil, err
	}

	return info, nil
}

func (info *CardanoChainInfo) initialize() error {
	info.logger.Debug("getting protocol params from db", "chainID", info.config.ChainID)

	persisted, expiresAt, err := info.db.GetProtocolParams(info.config.ChainID)
	if err != nil {
		return fmt.Errorf(
			"failed to read protocol parameters from the database for chain %s. err: %w", info.config.ChainID, err)
	}

	if len(persisted) > 0 {
		info.logger.Debug("db contains protocol params", "chainID", info.config.ChainID)

		if time.Now().UTC().Before(expiresAt) {
			return nil
		}

		info.logger.Debug("db protocol params stale. trying to fetch fresh",
			"chainID", info.config.ChainID, "expiresAt", expiresAt)

		pp, err := info.txProvider.GetProtocolParameters(info.ctx)
		if err != nil {
			if common.IsContextDoneErr(err) {
				return err
			} else {
				info.logger.Error("failed to fetch protocol parameters", "chainID", info.config.ChainID, "err", err)

				return nil
			}
		}

		newExpiresAt := time.Now().UTC().Add(protocolParamsTTL)

		info.logger.Debug("saving fresh protocol params to db", "chainID", info.config.ChainID, "newExpiresAt", newExpiresAt)

		if err := info.db.SaveProtocolParams(
			info.config.ChainID, pp, newExpiresAt); err != nil {
			info.logger.Error("failed to persist protocol parameters",
				"chainID", info.config.ChainID, "err", err)
		}

		return nil
	}

	if err := info.ensureProtocolParamsPersisted(); err != nil {
		return fmt.Errorf("failed to ensure protocol parameters for chain %s. err: %w", info.config.ChainID, err)
	}

	return nil
}

func (info *CardanoChainInfo) GetProtocolParams() []byte {
	info.lock.Lock()
	defer info.lock.Unlock()

	info.logger.Debug("getting protocol params from db", "chainID", info.config.ChainID)

	persisted, expiresAt, err := info.db.GetProtocolParams(info.config.ChainID)
	if err != nil || len(persisted) == 0 {
		info.logger.Error("failed to read protocol parameters from the database",
			"chainID", info.config.ChainID, "err", err)

		if err = info.ensureProtocolParamsPersisted(); err != nil {
			return []byte{}
		}

		persisted, _, err = info.db.GetProtocolParams(info.config.ChainID)
		if err != nil {
			info.logger.Error("failed to read protocol parameters from the database after ensuring they are persisted",
				"chainID", info.config.ChainID, "err", err)

			return []byte{}
		}

		return persisted
	}

	if time.Now().UTC().Before(expiresAt) {
		return persisted
	}

	info.logger.Debug("db protocol params stale. trying to fetch fresh",
		"chainID", info.config.ChainID, "expiresAt", expiresAt)

	pp, err := info.txProvider.GetProtocolParameters(info.ctx)
	if err != nil {
		info.logger.Error("failed to fetch protocol parameters",
			"chainID", info.config.ChainID, "err", err)

		return persisted
	}

	newExpiresAt := time.Now().UTC().Add(protocolParamsTTL)

	info.logger.Debug("saving fresh protocol params to db", "chainID", info.config.ChainID, "newExpiresAt", newExpiresAt)

	if err := info.db.SaveProtocolParams(
		info.config.ChainID, pp, newExpiresAt); err != nil {
		info.logger.Error("failed to persist protocol parameters",
			"chainID", info.config.ChainID, "err", err)
	}

	return pp
}

func (info *CardanoChainInfo) ensureProtocolParamsPersisted() error {
	return common.RetryForever(info.ctx, populateRetryInterval, func(ctx context.Context) error {
		info.logger.Debug("ensuring db protocol params - fetching", "chainID", info.config.ChainID)

		pp, err := info.txProvider.GetProtocolParameters(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch protocol parameters for chain %s. err: %w", info.config.ChainID, err)
		}

		newExpiresAt := time.Now().UTC().Add(protocolParamsTTL)

		info.logger.Debug("ensuring db protocol params - persisting",
			"chainID", info.config.ChainID, "newExpiresAt", newExpiresAt)

		if err := info.db.SaveProtocolParams(
			info.config.ChainID, pp, newExpiresAt); err != nil {
			return fmt.Errorf("failed to persist protocol parameters for chain %s. err: %w", info.config.ChainID, err)
		}

		return nil
	})
}
