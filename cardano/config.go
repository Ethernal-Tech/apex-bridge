package cardanotx

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type CardanoChainConfig struct {
	NetworkID                cardanowallet.CardanoNetworkType `json:"networkID"`
	NetworkMagic             uint32                           `json:"testnetMagic"`
	OgmiosURL                string                           `json:"ogmiosUrl,omitempty"`
	BlockfrostURL            string                           `json:"blockfrostUrl,omitempty"`
	BlockfrostAPIKey         string                           `json:"blockfrostApiKey,omitempty"`
	SocketPath               string                           `json:"socketPath,omitempty"`
	PotentialFee             uint64                           `json:"potentialFee"`
	TTLSlotNumberInc         uint64                           `json:"ttlSlotNumberIncrement"`
	SlotRoundingThreshold    uint64                           `json:"slotRoundingThreshold"`
	NoBatchPeriodPercent     float64                          `json:"noBatchPeriodPercent"`
	UtxoMinAmount            uint64                           `json:"minUtxoAmount"`
	MaxFeeUtxoCount          uint                             `json:"maxFeeUtxoCount"`
	MaxUtxoCount             uint                             `json:"maxUtxoCount"`
	DefaultMinFeeForBridging uint64                           `json:"defaultMinFeeForBridging"`
	MinFeeForBridgingTokens  uint64                           `json:"minFeeForBridgingTokens"`
	TakeAtLeastUtxoCount     uint                             `json:"takeAtLeastUtxoCount"`
	DestinationChains        map[string]common.TokenPairs     `json:"destChains"`
	Tokens                   map[uint16]common.Token          `json:"tokens"`
	MintingScriptTxInput     *cardanowallet.TxInput           `json:"mintingScriptTxInput,omitempty"`
	CustodialNft             *cardanowallet.Token             `json:"custodialNft,omitempty"`
	RelayerAddress           string                           `json:"relayerAddress,omitempty"`
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

func (config CardanoChainConfig) GetMinBridgingFee(hasNativeTokens bool) uint64 {
	if hasNativeTokens {
		return config.MinFeeForBridgingTokens
	}

	return config.DefaultMinFeeForBridging
}

func (config CardanoChainConfig) GetCurrencyID() (uint16, error) {
	for id, token := range config.Tokens {
		if token.ChainSpecific == cardanowallet.AdaTokenName {
			return id, nil
		}
	}

	return 0, fmt.Errorf("currency not found in chain config")
}

func (config CardanoChainConfig) GetWrappedTokenID() (uint16, bool) {
	for tokenID, token := range config.Tokens {
		if token.IsWrappedCurrency {
			return tokenID, true
		}
	}

	return 0, false
}

func (config CardanoChainConfig) GetTokenByID(tokenID uint16) (token cardanowallet.Token, err error) {
	tokenConfig, ok := config.Tokens[tokenID]
	if !ok {
		return token, fmt.Errorf("token not found in chain config")
	}

	return cardanowallet.NewTokenWithFullNameTry(tokenConfig.ChainSpecific)
}

func (config CardanoChainConfig) GetWrappedToken() (token cardanowallet.Token, err error) {
	for _, tokenConfig := range config.Tokens {
		if tokenConfig.IsWrappedCurrency {
			return cardanowallet.NewTokenWithFullNameTry(tokenConfig.ChainSpecific)
		}
	}

	return token, fmt.Errorf("wrapped token not found in chain config")
}

func (config CardanoChainConfig) GetTokenDataForTokenID(
	tokenID uint16,
) (token cardanowallet.Token, shouldMint bool, err error) {
	if tokenID == 0 {
		token, err = config.GetWrappedToken()

		return token, false, err
	}

	tokenConfig, ok := config.Tokens[tokenID]
	if !ok {
		return token, false, fmt.Errorf("token not found in chain config: %d", tokenID)
	}

	token, err = cardanowallet.NewTokenWithFullNameTry(tokenConfig.ChainSpecific)
	if err != nil {
		return token, false, fmt.Errorf("failed to get token from name: %w", err)
	}

	return token, !tokenConfig.LockUnlock, nil
}

func (config CardanoChainConfig) GetFullTokenNamesAndIds() (map[string]uint16, error) {
	tokens := make(map[string]uint16, len(config.Tokens))

	for tokenID, token := range config.Tokens {
		if token.ChainSpecific == cardanowallet.AdaTokenName {
			continue
		}

		confToken, err := cardanowallet.NewTokenWithFullNameTry(token.ChainSpecific)
		if err != nil {
			return nil, fmt.Errorf("failed to get token with ID %d from config. err: %w", tokenID, err)
		}

		tokens[confToken.String()] = tokenID
	}

	return tokens, nil
}

var (
	_ common.ChainSpecificConfig = (*CardanoChainConfig)(nil)
	_ common.ChainSpecificConfig = (*RelayerEVMChainConfig)(nil)
)

type BatcherEVMChainConfig struct {
	TTLBlockNumberInc      uint64                       `json:"ttlBlockNumberInc"`
	BlockRoundingThreshold uint64                       `json:"blockRoundingThreshold"`
	NoBatchPeriodPercent   float64                      `json:"noBatchPeriodPercent"`
	TestMode               uint8                        `json:"testMode,omitempty"` // only test mode (`-tags testenv`)
	MinFeeForBridging      common.BigInt                `json:"minFeeForBridging"`
	DestinationChains      map[string]common.TokenPairs `json:"destChains"`
	Tokens                 map[uint16]common.Token      `json:"tokens"`
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

func (config BatcherEVMChainConfig) GetCurrencyID() (uint16, error) {
	for id, token := range config.Tokens {
		if token.ChainSpecific == cardanowallet.AdaTokenName {
			return id, nil
		}
	}

	return 0, fmt.Errorf("currency id not found")
}

func (config BatcherEVMChainConfig) GetWrappedTokenID() (uint16, error) {
	for id, token := range config.Tokens {
		if token.IsWrappedCurrency {
			return id, nil
		}
	}

	return 0, fmt.Errorf("wrapped token id not found")
}

type RelayerEVMChainConfig struct {
	NodeURL          string `json:"nodeUrl"`
	DynamicTx        bool   `json:"dynamicTx"`
	DataDir          string `json:"dataDir,omitempty"`
	ConfigPath       string `json:"configPath,omitempty"`
	DepositGasLimit  uint64 `json:"depositGasLimit"`
	GasPrice         uint64 `json:"gasPrice"`
	GasFeeCap        uint64 `json:"gasFeeCap"`
	GasTipCap        uint64 `json:"gasTipCap"`
	GasFeeMultiplier uint64 `json:"gasFeeMultiplier"`
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
