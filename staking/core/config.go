package core

import (
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type StakingBridgingAddresses struct {
	StakingBridgingAddr string `json:"address"`
	FeeAddress          string `json:"feeAddress"`
}

type CardanoChainConfigUtxo struct {
	Hash    [32]byte                    `json:"id"`
	Index   uint32                      `json:"index"`
	Address string                      `json:"address"`
	Amount  uint64                      `json:"amount"`
	Tokens  []cardanowallet.TokenAmount `json:"tokens,omitempty"`
	Slot    uint64                      `json:"slot"`
}

type ChainConfig struct {
	ChainID                string                   `json:"-"`
	ChainType              string                   `json:"type"`
	NetworkAddress         string                   `json:"networkAddress"`
	NetworkMagic           uint32                   `json:"testnetMagic"`
	StartBlockHash         string                   `json:"startBlockHash"`
	StartSlot              uint64                   `json:"startSlot"`
	InitialUtxos           []CardanoChainConfigUtxo `json:"initialUtxos"`
	ConfirmationBlockCount uint                     `json:"confirmationBlockCount"`
	StakingAddresses       []string                 `json:"stakingAddresses"`
	StakingBridgingAddr    StakingBridgingAddresses `json:"stakingBridgingAddrs"`
}

type StakingConfiguration struct {
	Chain         ChainConfig `json:"chain"`
	PullTimeMilis int64       `json:"pullTime"`
}

type StakingManagerConfiguration struct {
	Chains        map[string]*ChainConfig `json:"chains"`
	Logger        logger.LoggerConfig     `json:"logger"`
	DbsPath       string                  `json:"dbsPath"`
	PullTimeMilis int64                   `json:"pullTime"`
}

func (config *StakingManagerConfiguration) FillOut() {
	for chainID, chainConfig := range config.Chains {
		chainConfig.ChainID = chainID
	}
}
