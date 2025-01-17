package core

import (
	"math/big"
	"time"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
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
	MinFeeForBridging       uint64                        `json:"minFeeForBridging"`
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
	MaxAmountAllowedToBridge       *big.Int `json:"maxAmountAllowedToBridge"`
	MaxReceiversPerBridgingRequest int      `json:"maxReceiversPerBridgingRequest"`
	MaxBridgingClaimsToGroup       int      `json:"maxBridgingClaimsToGroup"`
}

type RetryUnprocessedSettings struct {
	BaseTimeout time.Duration `json:"baseTimeout"`
	MaxTimeout  time.Duration `json:"maxTimeout"`
}

type AppConfig struct {
	RunMode                  common.VCRunMode               `json:"-"`
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

func (appConfig AppConfig) ToSendTxChainConfigs() (map[string]sendtx.ChainConfig, error) {
	result := make(map[string]sendtx.ChainConfig, len(appConfig.CardanoChains)+len(appConfig.EthChains))

	for chainID, cardanoConfig := range appConfig.CardanoChains {
		cfg, err := cardanoConfig.ToSendTxChainConfig()
		if err != nil {
			return nil, err
		}

		result[chainID] = cfg
	}

	for chainID, config := range appConfig.EthChains {
		result[chainID] = config.ToSendTxChainConfig()
	}

	return result, nil
}

func (config CardanoChainConfig) ToSendTxChainConfig() (res sendtx.ChainConfig, err error) {
	txProvider, err := config.CreateTxProvider()
	if err != nil {
		return res, err
	}

	var tokens []sendtx.TokenExchangeConfig

	for _, tdst := range config.Destinations {
		if tdst.SrcTokenName != cardanowallet.AdaTokenName {
			tokens = append(tokens, sendtx.TokenExchangeConfig{
				DstChainID: tdst.DstChainID,
				TokenName:  tdst.SrcTokenName,
			})
		}
	}

	return sendtx.ChainConfig{
		CardanoCliBinary: cardanowallet.ResolveCardanoCliBinary(config.NetworkID),
		TxProvider:       txProvider,
		MultiSigAddr:     config.BridgingAddresses.BridgingAddress,
		TestNetMagic:     uint(config.NetworkMagic),
		TTLSlotNumberInc: config.TTLSlotNumberInc,
		MinUtxoValue:     config.UtxoMinAmount,
		PotentialFee:     config.PotentialFee,
		NativeTokens:     tokens,
	}, nil
}

func (config EthChainConfig) ToSendTxChainConfig() sendtx.ChainConfig {
	return sendtx.ChainConfig{
		MultiSigAddr: config.BridgingAddresses.BridgingAddress,
	}
}
