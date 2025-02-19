package chain

import (
	"context"

	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	infracommon "github.com/Ethernal-Tech/cardano-infrastructure/common"
)

type CardanoChainInfo struct {
	Config *cCore.CardanoChainConfig

	ProtocolParams []byte
}

func NewCardanoChainInfo(config *cCore.CardanoChainConfig) *CardanoChainInfo {
	return &CardanoChainInfo{
		Config: config,
	}
}

func (info *CardanoChainInfo) Populate(ctx context.Context) error {
	txProvider, err := info.Config.CreateTxProvider()
	if err != nil {
		return err
	}

	protocolParams, err := infracommon.ExecuteWithRetry(
		ctx, func(ctx context.Context) ([]byte, error) {
			return txProvider.GetProtocolParameters(ctx)
		},
	)
	if err != nil {
		return err
	}

	info.ProtocolParams = protocolParams

	return nil
}
