package cardanotx

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Run("Invalid config", func(t *testing.T) {
		config, err := NewCardanoChainConfig(json.RawMessage(""))
		require.Nil(t, config)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to unmarshal Cardano configuration")
	})

	t.Run("Valid config", func(t *testing.T) {
		config, err := NewCardanoChainConfig(json.RawMessage(
			[]byte(`{
				"testnetMagic": 2,
				"txProvider": {
					"blockfrostUrl": "pera",
					"blockfrostApiKey": "zdera"
				},
				"potentialFee": 300000
				}`),
		))
		require.NoError(t, err)
		require.NotNil(t, config)
		require.Equal(t, "cardano", config.GetChainType())
		require.Equal(t, uint32(2), config.NetworkMagic)
		require.Equal(t, "pera", config.TxProvider.BlockfrostURL)
		require.Equal(t, "zdera", config.TxProvider.BlockfrostAPIKey)
		require.Equal(t, uint64(300000), config.PotentialFee)
	})
}
