package core

import (
	ocCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
)

type StakingBridgingAddresses struct {
	StakingBridgingAddr string `json:"address"`
	FeeAddress          string `json:"feeAddress"`
}

type CardanoChainConfig struct {
	ocCore.BaseCardanoChainConfig
	ChainType           string                   `json:"type"`
	NetworkMagic        uint32                   `json:"testnetMagic"`
	StakingAddresses    []string                 `json:"stakingAddresses"`
	StakingBridgingAddr StakingBridgingAddresses `json:"stakingBridgingAddrs"`
}

type StakingConfiguration struct {
	Chain                  CardanoChainConfig `json:"chain"`
	UsersRewardsPercentage float64            `json:"usersRewardsPercentage"`
	PullTimeMilis          int64              `json:"pullTime"`
}

type StakingManagerConfiguration struct {
	Chains                 map[string]*CardanoChainConfig `json:"chains"`
	Logger                 logger.LoggerConfig            `json:"logger"`
	DbsPath                string                         `json:"dbsPath"`
	UsersRewardsPercentage float64                        `json:"usersRewardsPercentage"`
	PullTimeMilis          int64                          `json:"pullTime"`
}

func (c CardanoChainConfig) GetNetworkMagic() uint32 {
	return c.NetworkMagic
}

func (c CardanoChainConfig) GetAddressesOfInterest() []string {
	return append([]string{
		c.StakingBridgingAddr.StakingBridgingAddr,
		c.StakingBridgingAddr.FeeAddress,
	}, c.StakingAddresses...)
}

func (config *StakingManagerConfiguration) FillOut() {
	for chainID, chainConfig := range config.Chains {
		chainConfig.ChainID = chainID
	}
}
