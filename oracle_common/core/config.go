package core

import (
	"errors"
	"math/big"
	"time"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
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
	// Transaction hash
	Hash [32]byte `json:"id"`
	// Output index
	Index uint32 `json:"index"`
	// Output address
	Address string `json:"address"`
	// Amount of currency
	Amount uint64 `json:"amount"`
	// List of tokens including their name, policy ID, and amount
	Tokens []cardanowallet.TokenAmount `json:"tokens,omitempty"`
	// Output slot
	Slot uint64 `json:"slot"`
} // @name CardanoChainConfigUtxo

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
	FeeAddrBridgingAmount    uint64                   `json:"feeAddressBridgingAmount"`
	MinOperationFee          uint64                   `json:"minOperationFee"`
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
	MaxAmountAllowedToBridge       *big.Int `json:"maxAmountAllowedToBridge"`
	MaxTokenAmountAllowedToBridge  *big.Int `json:"maxTokenAmountAllowedToBridge"`
	MaxReceiversPerBridgingRequest int      `json:"maxReceiversPerBridgingRequest"`
	MaxBridgingClaimsToGroup       int      `json:"maxBridgingClaimsToGroup"`
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
	RunMode                  common.VCRunMode               `json:"-"`
	RefundEnabled            bool                           `json:"refundEnabled"`
	ValidatorDataDir         string                         `json:"validatorDataDir"`
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

	return sendtx.ChainConfig{
		CardanoCliBinary:      cardanowallet.ResolveCardanoCliBinary(config.NetworkID),
		TxProvider:            txProvider,
		MultiSigAddr:          config.BridgingAddresses.BridgingAddress,
		TestNetMagic:          uint(config.NetworkMagic),
		TTLSlotNumberInc:      config.TTLSlotNumberInc,
		MinUtxoValue:          config.UtxoMinAmount,
		NativeTokens:          config.NativeTokens,
		MinBridgingFeeAmount:  config.MinFeeForBridging,
		MinOperationFeeAmount: config.MinOperationFee,
		PotentialFee:          config.PotentialFee,
		ProtocolParameters:    nil,
	}, nil
}

func (config EthChainConfig) ToSendTxChainConfig() sendtx.ChainConfig {
	feeValue := new(big.Int).SetUint64(config.MinFeeForBridging)

	if len(feeValue.String()) == common.WeiDecimals {
		feeValue = common.WeiToDfm(feeValue)
	}

	return sendtx.ChainConfig{
		MultiSigAddr:         config.BridgingAddresses.BridgingAddress,
		MinBridgingFeeAmount: feeValue.Uint64(),
	}
}
