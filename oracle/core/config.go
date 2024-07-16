package core

import (
	"time"

	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type BridgingAddresses struct {
	BridgingAddress string `json:"address"`
	FeeAddress      string `json:"feeAddress"`
}

type EthChainConfig struct {
	ChainID                 string
	BridgingAddresses       BridgingAddresses `json:"bridgingAddresses"`
	NodeURL                 string            `json:"nodeUrl"`
	SyncBatchSize           uint64            `json:"syncBatchSize"`
	NumBlockConfirmations   uint64            `json:"numBlockConfirmations"`
	StartBlockNumber        uint64            `json:"startBlockNumber"`
	PoolIntervalMiliseconds time.Duration     `json:"poolIntervalMs"`
	TTLBlockNumberInc       uint64            `json:"ttlBlockNumberInc"`
	BlockRoundingThreshold  uint64            `json:"blockRoundingThreshold"`
	NoBatchPeriodPercent    float64           `json:"noBatchPeriodPercent"`
}

type CardanoChainConfig struct {
	ChainID                  string
	NetworkAddress           string                           `json:"networkAddress"`
	NetworkMagic             uint32                           `json:"networkMagic"`
	NetworkID                cardanowallet.CardanoNetworkType `json:"networkID"`
	StartBlockHash           string                           `json:"startBlockHash"`
	StartSlot                uint64                           `json:"startSlot"`
	StartBlockNumber         uint64                           `json:"startBlockNumber"`
	ConfirmationBlockCount   uint                             `json:"confirmationBlockCount"`
	BridgingAddresses        BridgingAddresses                `json:"bridgingAddresses"`
	OtherAddressesOfInterest []string                         `json:"otherAddressesOfInterest"`
	InitialUtxos             []*indexer.TxInputOutput         `json:"initialUtxos"`
}

type SubmitConfig struct {
	ConfirmedBlocksThreshold  int `json:"confirmedBlocksThreshold"`
	ConfirmedBlocksSubmitTime int `json:"confirmedBlocksSubmitTime"`
}

type BridgeConfig struct {
	NodeURL              string       `json:"nodeUrl"`
	DynamicTx            bool         `json:"dynamicTx"`
	SmartContractAddress string       `json:"scAddress"`
	SubmitConfig         SubmitConfig `json:"submitConfig"`
}

type AppSettings struct {
	Logger  logger.LoggerConfig `json:"logger"`
	DbsPath string              `json:"dbsPath"`
}

type BridgingSettings struct {
	MinFeeForBridging              uint64 `json:"minFeeForBridging"`
	UtxoMinValue                   uint64 `json:"utxoMinValue"`
	MaxReceiversPerBridgingRequest int    `json:"maxReceiversPerBridgingRequest"`
	MaxBridgingClaimsToGroup       int    `json:"maxBridgingClaimsToGroup"`
}

type AppConfig struct {
	ValidatorDataDir    string                         `json:"validatorDataDir"`
	ValidatorConfigPath string                         `json:"validatorConfigPath"`
	CardanoChains       map[string]*CardanoChainConfig `json:"cardanoChains"`
	EthChains           map[string]*EthChainConfig     `json:"ethChains"`
	Bridge              BridgeConfig                   `json:"bridge"`
	Settings            AppSettings                    `json:"appSettings"`
	BridgingSettings    BridgingSettings               `json:"bridgingSettings"`
}

func (appConfig *AppConfig) FillOut() {
	for chainID, cardanoChainConfig := range appConfig.CardanoChains {
		cardanoChainConfig.ChainID = chainID
	}

	for chainID, ethChainConfig := range appConfig.EthChains {
		ethChainConfig.ChainID = chainID
	}
}
