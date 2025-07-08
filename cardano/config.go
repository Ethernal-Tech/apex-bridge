package cardanotx

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type TxProviderConfig struct {
	OgmiosURL           string `json:"ogmiosUrl,omitempty"`
	BlockfrostURL       string `json:"blockfrostUrl,omitempty"`
	BlockfrostAPIKey    string `json:"blockfrostApiKey,omitempty"`
	DemeterSubmitAPIURL string `json:"demeterSubmitAPIURL,omitempty"`
	DemeterSubmitAPIKey string `json:"demeterSubmitAPIKey,omitempty"`
	SocketPath          string `json:"socketPath,omitempty"`
}

type CardanoChainConfig struct {
	NetworkID             cardanowallet.CardanoNetworkType `json:"networkID"`
	NetworkMagic          uint32                           `json:"testnetMagic"`
	TxProvider            *TxProviderConfig                `json:"txProvider,omitempty"`
	PotentialFee          uint64                           `json:"potentialFee"`
	TTLSlotNumberInc      uint64                           `json:"ttlSlotNumberIncrement"`
	SlotRoundingThreshold uint64                           `json:"slotRoundingThreshold"`
	NoBatchPeriodPercent  float64                          `json:"noBatchPeriodPercent"`
	UtxoMinAmount         uint64                           `json:"minUtxoAmount"`
	MaxFeeUtxoCount       uint                             `json:"maxFeeUtxoCount"`
	MaxUtxoCount          uint                             `json:"maxUtxoCount"`
	TakeAtLeastUtxoCount  uint                             `json:"takeAtLeastUtxoCount"`
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
	providerConfig := config.TxProvider
	if providerConfig == nil {
		return nil, errors.New("no tx provider specified")
	}

	if providerConfig.OgmiosURL != "" {
		return cardanowallet.NewTxProviderOgmios(providerConfig.OgmiosURL), nil
	}

	if providerConfig.SocketPath != "" {
		return cardanowallet.NewTxProviderCli(
			uint(config.NetworkMagic), providerConfig.SocketPath, cardanowallet.ResolveCardanoCliBinary(config.NetworkID))
	}

	if providerConfig.BlockfrostURL != "" {
		if providerConfig.DemeterSubmitAPIURL != "" {
			return cardanowallet.NewTxProviderDemeter(
				providerConfig.BlockfrostURL, providerConfig.BlockfrostAPIKey,
				providerConfig.DemeterSubmitAPIURL, providerConfig.DemeterSubmitAPIKey), nil
		}

		return cardanowallet.NewTxProviderBlockFrost(providerConfig.BlockfrostURL, providerConfig.BlockfrostAPIKey), nil
	}

	return nil, errors.New("no tx provider specified")
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

	token, err = GetNativeTokenFromName(tokenName)
	if err == nil {
		return token, nil
	}

	return token, fmt.Errorf("chainID: %s, err: %w", dstChainID, err)
}

func GetNativeTokenFromConfig(tokenConfig sendtx.TokenExchangeConfig) (token cardanowallet.Token, err error) {
	token, err = GetNativeTokenFromName(tokenConfig.TokenName)
	if err == nil {
		return token, nil
	}

	return token, fmt.Errorf("chainID: %s, err: %w", tokenConfig.DstChainID, err)
}

func GetNativeTokenFromName(tokenName string) (token cardanowallet.Token, err error) {
	return cardanowallet.NewTokenWithFullNameTry(tokenName)
}

var (
	_ common.ChainSpecificConfig = (*CardanoChainConfig)(nil)
	_ common.ChainSpecificConfig = (*RelayerEVMChainConfig)(nil)
)

type BatcherEVMChainConfig struct {
	TTLBlockNumberInc      uint64  `json:"ttlBlockNumberInc"`
	BlockRoundingThreshold uint64  `json:"blockRoundingThreshold"`
	NoBatchPeriodPercent   float64 `json:"noBatchPeriodPercent"`
	TestMode               uint8   `json:"testMode,omitempty"` // only functional in test mode (`-tags testenv`)
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
	NodeURL         string `json:"nodeUrl"`
	DynamicTx       bool   `json:"dynamicTx"`
	DataDir         string `json:"dataDir,omitempty"`
	ConfigPath      string `json:"configPath,omitempty"`
	DepositGasLimit uint64 `json:"depositGasLimit"`
	GasPrice        uint64 `json:"gasPrice"`
	GasFeeCap       uint64 `json:"gasFeeCap"`
	GasTipCap       uint64 `json:"gasTipCap"`
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
