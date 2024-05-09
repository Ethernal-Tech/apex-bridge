package cardanotx

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type CardanoChainConfig struct {
	TestNetMagic     uint32 `json:"testnetMagic"`
	OgmiosURL        string `json:"ogmiosUrl,omitempty"`
	BlockfrostURL    string `json:"blockfrostUrl,omitempty"`
	BlockfrostAPIKey string `json:"blockfrostApiKey,omitempty"`
	SocketPath       string `json:"socketPath,omitempty"`
	PotentialFee     uint64 `json:"potentialFee"`
	KeysDirPath      string `json:"keysDirPath,omitempty"`
}

var _ common.ChainSpecificConfig = (*CardanoChainConfig)(nil)

// GetChainType implements ChainSpecificConfig.
func (*CardanoChainConfig) GetChainType() string {
	return "Cardano"
}

func NewCardanoChainConfig(rawMessage json.RawMessage) (*CardanoChainConfig, error) {
	var cardanoChainConfig CardanoChainConfig
	if err := json.Unmarshal(rawMessage, &cardanoChainConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Cardano configuration: %w", err)
	}

	return &cardanoChainConfig, nil
}

func (config CardanoChainConfig) Serialize() ([]byte, error) {
	return json.Marshal(config)
}

func (config CardanoChainConfig) CreateTxProvider() (cardanowallet.ITxProvider, error) {
	//nolint:gocritic
	if config.OgmiosURL != "" {
		return cardanowallet.NewTxProviderOgmios(config.OgmiosURL), nil
	} else if config.SocketPath != "" {
		return cardanowallet.NewTxProviderCli(uint(config.TestNetMagic), config.SocketPath)
	} else if config.BlockfrostURL != "" {
		return cardanowallet.NewTxProviderBlockFrost(config.BlockfrostURL, config.BlockfrostAPIKey), nil
	}

	return nil, errors.New("neither a blockfrost nor a ogmios nor a socket path is specified")
}

func (config CardanoChainConfig) LoadWallet() (*CardanoWallet, error) {
	return LoadWallet(config.KeysDirPath, false)
}
