package solanatx

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
)

type SolanaChainConfig struct {
	ConfirmationTimeout                   time.Duration                `json:"confirmationTimeout"`
	TTLNumberInc                          uint64                       `json:"ttlNumberIncrement"`
	SlotRoundingThreshold                 uint64                       `json:"slotRoundingThreshold"`
	NoBatchPeriodPercent                  float64                      `json:"noBatchPeriodPercent"`
	MinFeeForBridging                     *big.Int                     `json:"minFeeForBridging"`
	DestinationChains                     map[string]common.TokenPairs `json:"destChain"`
	Tokens                                map[uint16]common.Token      `json:"tokens"`
	AlwaysTrackCurrencyAndWrappedCurrency bool                         `json:"alwaysTrackCurrencyAndWrappedCurrency"`
	TxProviderEndpoint                    string                       `json:"txProviderEndpoint"`
	ALTPublicKey                          string                       `json:"altPublicKey"`
	ProgramID                             string                       `json:"trackedProgram"`
}

func NewSolanaChainConfig(rawMessage json.RawMessage) (*SolanaChainConfig, error) {
	var solanaChainConfig SolanaChainConfig
	if err := json.Unmarshal(rawMessage, &solanaChainConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Solana configuration: %w", err)
	}

	return &solanaChainConfig, nil
}

func (config SolanaChainConfig) Serialize() ([]byte, error) {
	return json.Marshal(config)
}

func (config SolanaChainConfig) GetTokenID(tokenMint string) (uint16, error) {
	for id, token := range config.Tokens {
		if token.ChainSpecific == tokenMint {
			return id, nil
		}
	}

	return 0, fmt.Errorf("token not found in chain config")
}

func (config SolanaChainConfig) GetTokenMint(tokenID uint16) (string, error) {
	token, ok := config.Tokens[tokenID]
	if !ok {
		return "", fmt.Errorf("token not found in chain config")
	}

	return token.ChainSpecific, nil
}

func (config SolanaChainConfig) GetWrappedTokenID() (uint16, error) {
	for id, token := range config.Tokens {
		if token.IsWrappedCurrency {
			return id, nil
		}
	}

	return 0, fmt.Errorf("wrapped token id not found")
}
