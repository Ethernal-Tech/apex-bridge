package core

import (
	"math/big"
	"time"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
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
	ChainID                 string               `json:"-"`
	BridgingAddresses       EthBridgingAddresses `json:"-"`
	NodeURL                 string               `json:"nodeUrl"`
	SyncBatchSize           uint64               `json:"syncBatchSize"`
	NumBlockConfirmations   uint64               `json:"numBlockConfirmations"`
	StartBlockNumber        uint64               `json:"startBlockNumber"`
	PoolIntervalMiliseconds time.Duration        `json:"poolIntervalMs"`
	TTLBlockNumberInc       uint64               `json:"ttlBlockNumberInc"`
	BlockRoundingThreshold  uint64               `json:"blockRoundingThreshold"`
	NoBatchPeriodPercent    float64              `json:"noBatchPeriodPercent"`
	DynamicTx               bool                 `json:"dynamicTx"`
	TestMode                uint8                `json:"testMode"`
	MinFeeForBridging       uint64               `json:"minFeeForBridging"`
	RestartTrackerPullCheck time.Duration        `json:"restartTrackerPullCheck"`
	FeeAddrBridgingAmount   uint64               `json:"feeAddressBridgingAmount"`
}

type CardanoChainConfig struct {
	cardanotx.CardanoChainConfig
	ChainID                  string                   `json:"-"`
	BridgingAddresses        BridgingAddresses        `json:"-"`
	NetworkAddress           string                   `json:"networkAddress"`
	StartBlockHash           string                   `json:"startBlockHash"`
	StartSlot                uint64                   `json:"startSlot"`
	ConfirmationBlockCount   uint                     `json:"confirmationBlockCount"`
	OtherAddressesOfInterest []string                 `json:"otherAddressesOfInterest"`
	InitialUtxos             []CardanoChainConfigUtxo `json:"initialUtxos"`
	MinFeeForBridging        uint64                   `json:"minFeeForBridging"`
	FeeAddrBridgingAmount    uint64                   `json:"feeAddressBridgingAmount"`
}

type SubmitConfig struct {
	ConfirmedBlocksThreshold  int             `json:"confirmedBlocksThreshold"`
	ConfirmedBlocksSubmitTime int             `json:"confirmedBlocksSubmitTime"`
	EmptyBlocksThreshold      map[string]uint `json:"emptyBlocksThreshold"`
	UpdateFromIndexerDB       bool            `json:"updateFromIndexerDb"`
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
	MaxAmountAllowedToBridge       *big.Int            `json:"maxAmountAllowedToBridge"`
	MaxReceiversPerBridgingRequest int                 `json:"maxReceiversPerBridgingRequest"`
	MaxBridgingClaimsToGroup       int                 `json:"maxBridgingClaimsToGroup"`
	AllowedDirections              map[string][]string `json:"allowedDirections"`
}

type RetryUnprocessedSettings struct {
	BaseTimeout time.Duration `json:"baseTimeout"`
	MaxTimeout  time.Duration `json:"maxTimeout"`
}

type TryCountLimits struct {
	MaxBatchTryCount  uint32 `json:"maxBatchTryCount"`
	MaxSubmitTryCount uint32 `json:"maxSubmitTryCount"`
	MaxRefundTryCount uint32 `json:"maxRefundTryCount"`
}

type AppConfig struct {
	ValidatorDataDir         string                         `json:"validatorDataDir"`
	RefundEnabled            bool                           `json:"refundEnabled"`
	ValidatorConfigPath      string                         `json:"validatorConfigPath"`
	CardanoChains            map[string]*CardanoChainConfig `json:"cardanoChains"`
	EthChains                map[string]*EthChainConfig     `json:"ethChains"`
	Bridge                   BridgeConfig                   `json:"bridge"`
	Settings                 AppSettings                    `json:"appSettings"`
	BridgingSettings         BridgingSettings               `json:"bridgingSettings"`
	RetryUnprocessedSettings RetryUnprocessedSettings       `json:"retryUnprocessedSettings"`
	TryCountLimits           TryCountLimits                 `json:"tryCountLimits"`
}

func (appConfig *AppConfig) FillOut() {
	for chainID, cardanoChainConfig := range appConfig.CardanoChains {
		cardanoChainConfig.ChainID = chainID
	}

	for chainID, ethChainConfig := range appConfig.EthChains {
		ethChainConfig.ChainID = chainID
	}
}
