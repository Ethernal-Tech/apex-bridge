package core

import (
	"encoding/json"
	"testing"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/stretchr/testify/require"
)

func TestSeparateConfigs_EVMDestination(t *testing.T) {
	const (
		evmChainID = uint64(9070)
		gateway    = "0x2222222222222222222222222222222222222222"
	)

	appConfig := &AppConfig{
		EthChains: map[string]*oracleCore.EthChainConfig{
			"nexus": {
				ChainID:           "nexus",
				EVMChainID:        evmChainID,
				BridgingAddresses: oracleCore.EthBridgingAddresses{BridgingAddress: gateway},
			},
		},
	}

	_, batcherConfig := appConfig.SeparateConfigs()

	var found bool

	for _, chain := range batcherConfig.Chains {
		if chain.ChainID != "nexus" {
			continue
		}

		found = true

		var cfg cardanotx.BatcherEVMChainConfig

		require.NoError(t, json.Unmarshal(chain.ChainSpecific, &cfg))
		require.Equal(t, evmChainID, cfg.EVMChainID)
		require.Equal(t, gateway, cfg.GatewayAddress)
	}

	require.True(t, found)
}
