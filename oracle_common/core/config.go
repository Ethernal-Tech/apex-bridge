package core

import (
	"time"

	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type BridgingAddresses struct {
	BridgingAddress string `json:"address"`
	FeeAddress      string `json:"feeAddress"`
}

type EthBridgingAddresses struct {
	BridgingAddress string `json:"address"`
}

type CardanoChainConfigUtxo struct {
	Hash    [32]byte `json:"id"`
	Index   uint32   `json:"index"`
	Address string   `json:"address"`
	Amount  uint64   `json:"amount"`
	Slot    uint64   `json:"slot"`
}

type EthChainConfig struct {
	ChainID                 string                        `json:"-"`
	BridgingAddresses       EthBridgingAddresses          `json:"-"`
	NodeURL                 string                        `json:"nodeUrl"`
	SyncBatchSize           uint64                        `json:"syncBatchSize"`
	NumBlockConfirmations   uint64                        `json:"numBlockConfirmations"`
	StartBlockNumber        uint64                        `json:"startBlockNumber"`
	PoolIntervalMiliseconds time.Duration                 `json:"poolIntervalMs"`
	TTLBlockNumberInc       uint64                        `json:"ttlBlockNumberInc"`
	BlockRoundingThreshold  uint64                        `json:"blockRoundingThreshold"`
	NoBatchPeriodPercent    float64                       `json:"noBatchPeriodPercent"`
	DynamicTx               bool                          `json:"dynamicTx"`
	TestMode                uint8                         `json:"testMode"`
	NonceStrategy           ethtxhelper.NonceStrategyType `json:"nonceStrategy"`
}

type CardanoChainConfig struct {
	ChainID                  string                           `json:"-"`
	BridgingAddresses        BridgingAddresses                `json:"-"`
	NetworkAddress           string                           `json:"networkAddress"`
	NetworkMagic             uint32                           `json:"networkMagic"`
	NetworkID                cardanowallet.CardanoNetworkType `json:"networkID"`
	StartBlockHash           string                           `json:"startBlockHash"`
	StartSlot                uint64                           `json:"startSlot"`
	ConfirmationBlockCount   uint                             `json:"confirmationBlockCount"`
	OtherAddressesOfInterest []string                         `json:"otherAddressesOfInterest"`
	InitialUtxos             []CardanoChainConfigUtxo         `json:"initialUtxos"`

	OgmiosURL             string  `json:"ogmiosUrl"`
	BlockfrostURL         string  `json:"blockfrostUrl"`
	BlockfrostAPIKey      string  `json:"blockfrostApiKey"`
	SocketPath            string  `json:"socketPath"`
	PotentialFee          uint64  `json:"potentialFee"`
	SlotRoundingThreshold uint64  `json:"slotRoundingThreshold"`
	TTLSlotNumberInc      uint64  `json:"ttlSlotNumberIncrement"`
	NoBatchPeriodPercent  float64 `json:"noBatchPeriodPercent"`
	TakeAtLeastUtxoCount  int     `json:"takeAtLeastUtxoCount"`
}

type SubmitConfig struct {
	ConfirmedBlocksThreshold  int `json:"confirmedBlocksThreshold"`
	ConfirmedBlocksSubmitTime int `json:"confirmedBlocksSubmitTime"`
}

type BridgeConfig struct {
	NodeURL              string                        `json:"nodeUrl"`
	DynamicTx            bool                          `json:"dynamicTx"`
	SmartContractAddress string                        `json:"scAddress"`
	SubmitConfig         SubmitConfig                  `json:"submitConfig"`
	NonceStrategy        ethtxhelper.NonceStrategyType `json:"nonceStrategy"`
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

type RetryUnprocessedSettings struct {
	BaseTimeout time.Duration `json:"baseTimeout"`
	MaxTimeout  time.Duration `json:"maxTimeout"`
}

type AppConfig struct {
	ValidatorDataDir         string                         `json:"validatorDataDir"`
	ValidatorConfigPath      string                         `json:"validatorConfigPath"`
	CardanoChains            map[string]*CardanoChainConfig `json:"cardanoChains"`
	EthChains                map[string]*EthChainConfig     `json:"ethChains"`
	Bridge                   BridgeConfig                   `json:"bridge"`
	Settings                 AppSettings                    `json:"appSettings"`
	BridgingSettings         BridgingSettings               `json:"bridgingSettings"`
	RetryUnprocessedSettings RetryUnprocessedSettings       `json:"retryUnprocessedSettings"`
}

func (appConfig *AppConfig) FillOut() {
	for chainID, cardanoChainConfig := range appConfig.CardanoChains {
		cardanoChainConfig.ChainID = chainID
	}

	for chainID, ethChainConfig := range appConfig.EthChains {
		ethChainConfig.ChainID = chainID
	}
}
