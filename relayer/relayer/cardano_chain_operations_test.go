package relayer

import (
	"encoding/json"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/stretchr/testify/require"
)

func TestToCardanoChainConfig(t *testing.T) {

	t.Run("Invalid chain type", func(t *testing.T) {
		chainSpecificConfig := core.ChainSpecific{
			ChainType: "NotCardano",
			Config:    json.RawMessage(""),
		}

		config, err := core.ToCardanoChainConfig(chainSpecificConfig)
		require.Nil(t, config)
		require.Error(t, err)
		require.ErrorContains(t, err, "chain type must be Cardano not NotCardano")
	})

	t.Run("Invalid config", func(t *testing.T) {
		chainSpecificConfig := core.ChainSpecific{
			ChainType: "Cardano",
			Config:    json.RawMessage(""),
		}

		config, err := core.ToCardanoChainConfig(chainSpecificConfig)
		require.Nil(t, config)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to unmarshal Cardano configuration")
	})

	jsonData := []byte(`{
		"testnetMagic": 2,
		"blockfrostUrl": "https://cardano-preview.blockfrost.io/api/v0",
		"blockfrostApiKey": "preview7mGSjpyEKb24OxQ4cCxomxZ5axMs5PvE",
		"atLeastValidators": 0.6666666666666666,
		"potentialFee": 300000
		}`)

	chainSpecificConfig := core.ChainSpecific{
		ChainType: "Cardano",
		Config:    json.RawMessage(jsonData),
	}

	config, err := core.ToCardanoChainConfig(chainSpecificConfig)
	require.NoError(t, err)
	require.NotNil(t, config)
	require.Equal(t, "Cardano", config.GetChainType())
	require.Equal(t, uint(2), config.TestNetMagic)
	require.Equal(t, "https://cardano-preview.blockfrost.io/api/v0", config.BlockfrostUrl)
	require.Equal(t, "preview7mGSjpyEKb24OxQ4cCxomxZ5axMs5PvE", config.BlockfrostAPIKey)
	require.Equal(t, uint64(300000), config.PotentialFee)
}
