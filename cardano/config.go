package cardanotx

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

// CardanoConfigTokenExchange holds src and dst token to exchange with destionation chain
// full token name[policyID.hex(name)or lovelace or "" for eth
type CardanoConfigTokenExchange struct {
	Chain        string
	SrcTokenName string
	DstTokenName string
}

type CardanoChainConfig struct {
	NetworkID             cardanowallet.CardanoNetworkType `json:"networkID"`
	NetworkMagic          uint32                           `json:"testnetMagic"`
	OgmiosURL             string                           `json:"ogmiosUrl,omitempty"`
	BlockfrostURL         string                           `json:"blockfrostUrl,omitempty"`
	BlockfrostAPIKey      string                           `json:"blockfrostApiKey,omitempty"`
	SocketPath            string                           `json:"socketPath,omitempty"`
	PotentialFee          uint64                           `json:"potentialFee"`
	TTLSlotNumberInc      uint64                           `json:"ttlSlotNumberIncrement"`
	SlotRoundingThreshold uint64                           `json:"slotRoundingThreshold"`
	NoBatchPeriodPercent  float64                          `json:"noBatchPeriodPercent"`
	UtxoMinAmount         uint64                           `json:"minUtxoAmount"`
	TakeAtLeastUtxoCount  int                              `json:"takeAtLeastUtxoCount"`
	Destinations          []CardanoConfigTokenExchange     `json:"destinations"`
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
			uint(config.NetworkMagic), config.SocketPath, cardanowallet.ResolveCardanoCliBinary(config.NetworkID))
	}

	if config.BlockfrostURL != "" {
		return cardanowallet.NewTxProviderBlockFrost(config.BlockfrostURL, config.BlockfrostAPIKey), nil
	}

	return nil, errors.New("neither a blockfrost nor a ogmios nor a socket path is specified")
}

var (
	_ common.ChainSpecificConfig = (*CardanoChainConfig)(nil)
	_ common.ChainSpecificConfig = (*RelayerEVMChainConfig)(nil)
)

type BatcherEVMChainConfig struct {
	TTLBlockNumberInc      uint64                        `json:"ttlBlockNumberInc"`
	BlockRoundingThreshold uint64                        `json:"blockRoundingThreshold"`
	NoBatchPeriodPercent   float64                       `json:"noBatchPeriodPercent"`
	TestMode               uint8                         `json:"testMode"`
	NonceStrategy          ethtxhelper.NonceStrategyType `json:"nonceStrategy"`
}

func NewBatcherEVMChainConfig(rawMessage json.RawMessage) (*BatcherEVMChainConfig, error) {
	var config BatcherEVMChainConfig
	if err := json.Unmarshal(rawMessage, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal EVM configuration: %w", err)
	}

	return &config, nil
}

// GetChainType implements ChainSpecificConfig.
func (*BatcherEVMChainConfig) GetChainType() string {
	return common.ChainTypeEVMStr
}

func (config BatcherEVMChainConfig) Serialize() ([]byte, error) {
	return json.Marshal(config)
}

type RelayerEVMChainConfig struct {
	NodeURL         string                        `json:"nodeUrl"`
	DynamicTx       bool                          `json:"dynamicTx"`
	DataDir         string                        `json:"dataDir,omitempty"`
	ConfigPath      string                        `json:"configPath,omitempty"`
	DepositGasLimit uint64                        `json:"depositGasLimit"`
	GasPrice        uint64                        `json:"gasPrice"`
	GasFeeCap       uint64                        `json:"gasFeeCap"`
	GasTipCap       uint64                        `json:"gasTipCap"`
	NonceStrategy   ethtxhelper.NonceStrategyType `json:"nonceStrategy"`
}

func NewRelayerEVMChainConfig(rawMessage json.RawMessage) (*RelayerEVMChainConfig, error) {
	var config RelayerEVMChainConfig
	if err := json.Unmarshal(rawMessage, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal EVM configuration: %w", err)
	}

	return &config, nil
}

// GetChainType implements ChainSpecificConfig.
func (*RelayerEVMChainConfig) GetChainType() string {
	return common.ChainTypeEVMStr
}

func (config RelayerEVMChainConfig) Serialize() ([]byte, error) {
	return json.Marshal(config)
}
