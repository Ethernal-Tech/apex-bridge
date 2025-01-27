package cardanotx

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

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
	NativeTokens          []sendtx.TokenExchangeConfig     `json:"nativeTokens"`
}

// GetChainType implements ChainSpecificConfig.
func (CardanoChainConfig) GetChainType() string {
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

func (config CardanoChainConfig) GetNativeTokenName(dstChainID string) string {
	for _, dst := range config.NativeTokens {
		if dst.DstChainID != dstChainID {
			continue
		}

		return dst.TokenName
	}

	return ""
}

func (config CardanoChainConfig) GetNativeToken(dstChainID string) (token cardanowallet.Token, err error) {
	tokenName := config.GetNativeTokenName(dstChainID)
	if tokenName == "" {
		return token, fmt.Errorf("no native token specified for destination: %s", dstChainID)
	}

	token, err = cardanowallet.NewTokenWithFullName(tokenName, true)
	if err == nil {
		return token, nil
	}

	token, err = cardanowallet.NewTokenWithFullName(tokenName, false)
	if err == nil {
		return token, nil
	}

	return token, fmt.Errorf("invalid token name %s for destination: %s", tokenName, dstChainID)
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
