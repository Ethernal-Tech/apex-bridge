package cardanotx

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type CardanoChainConfig struct {
	NetworkID             cardanowallet.CardanoNetworkType `json:"networkID"`
	TestNetMagic          uint32                           `json:"testnetMagic"`
	OgmiosURL             string                           `json:"ogmiosUrl,omitempty"`
	BlockfrostURL         string                           `json:"blockfrostUrl,omitempty"`
	BlockfrostAPIKey      string                           `json:"blockfrostApiKey,omitempty"`
	SocketPath            string                           `json:"socketPath,omitempty"`
	PotentialFee          uint64                           `json:"potentialFee"`
	TTLSlotNumberInc      uint64                           `json:"ttlSlotNumberIncrement"`
	SlotRoundingThreshold uint64                           `json:"slotRoundingThreshold"`
}

// GetChainType implements ChainSpecificConfig.
func (*CardanoChainConfig) GetChainType() string {
	return common.ChainTypeCardanoStr
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
	if config.OgmiosURL != "" {
		return cardanowallet.NewTxProviderOgmios(config.OgmiosURL), nil
	}

	if config.SocketPath != "" {
		return cardanowallet.NewTxProviderCli(
			uint(config.TestNetMagic), config.SocketPath, cardanowallet.ResolveCardanoCliBinary(config.NetworkID))
	}

	if config.BlockfrostURL != "" {
		return cardanowallet.NewTxProviderBlockFrost(config.BlockfrostURL, config.BlockfrostAPIKey), nil
	}

	return nil, errors.New("neither a blockfrost nor a ogmios nor a socket path is specified")
}

var (
	_ common.ChainSpecificConfig = (*CardanoChainConfig)(nil)
	_ common.ChainSpecificConfig = (*EVMChainConfig)(nil)
)

type EVMChainConfig struct {
	NodeURL      string `json:"nodeUrl"`
	BridgingAddr string `json:"bridgingAddr"`
	FeePayerAddr string `json:"feePayerAddr"`
}

func NewEVMChainConfig(rawMessage json.RawMessage) (*EVMChainConfig, error) {
	var config EVMChainConfig
	if err := json.Unmarshal(rawMessage, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal EVM configuration: %w", err)
	}

	return &config, nil
}

// GetChainType implements ChainSpecificConfig.
func (*EVMChainConfig) GetChainType() string {
	return common.ChainTypeEVMStr
}

func (config EVMChainConfig) Serialize() ([]byte, error) {
	return json.Marshal(config)
}
