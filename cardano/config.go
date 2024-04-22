package cardanotx

import (
	"encoding/json"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type CardanoChainConfig struct {
	TestNetMagic      uint32  `json:"testnetMagic"`
	BlockfrostUrl     string  `json:"blockfrostUrl"`
	BlockfrostAPIKey  string  `json:"blockfrostApiKey"`
	SocketPath        string  `json:"socketPath"`
	AtLeastValidators float64 `json:"atLeastValidators"`
	PotentialFee      uint64  `json:"potentialFee"`
	KeysDirPath       string  `json:"keysDirPath"`
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
	if config.SocketPath != "" {
		return cardanowallet.NewTxProviderCli(uint(config.TestNetMagic), config.SocketPath)
	} else if config.BlockfrostUrl != "" {
		return cardanowallet.NewTxProviderBlockFrost(config.BlockfrostUrl, config.BlockfrostAPIKey)
	}

	return &TxProviderTestMock{
		ReturnDefaultParameters: true,
	}, nil
}

func (config CardanoChainConfig) LoadWallet() (*CardanoWallet, error) {
	return LoadWallet(config.KeysDirPath, false)
}
