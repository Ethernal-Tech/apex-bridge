package core

import (
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

type BridgingAddress struct {
	ChainId    string `yaml:"chain_id"`
	Address    string `yaml:"address"`
	FeeAddress string `yaml:"fee_address"`
}

type CardanoChainConfig struct {
	ChainId                  string            `yaml:"chain_id"`
	NetworkAddress           string            `yaml:"network_address"`
	NetworkMagic             string            `yaml:"network_magic"`
	StartBlockHash           string            `yaml:"start_block_hash"`
	StartSlot                string            `yaml:"start_slot"`
	StartBlockNumber         string            `yaml:"start_block_number"`
	FeeAddress               string            `yaml:"fee_address"`
	BridgingAddresses        []BridgingAddress `yaml:"bridging_addresses"`
	OtherAddressesOfInterest []string          `yaml:"other_addresses_of_interest"`
	ConfirmationBlockCount   uint              `yaml:"confirmation_block_count"`
}

type AppSettings struct {
	DbsPath                  string `yaml:"dbs_path"`
	LogsPath                 string `yaml:"logs_path"`
	MaxBridgingClaimsToGroup int    `yaml:"max_bridging_claims_to_group"`
	LogLevel                 int32  `yaml:"log_level"`
}

type BridgingSettings struct {
	MinFeeForBridging uint64 `yaml:"min_fee_for_bridging"`
	UtxoMinValue      uint64 `yaml:"utxo_min_value"`
}

type AppConfig struct {
	CardanoChains    []CardanoChainConfig `yaml:"cardano_chains"`
	Settings         AppSettings          `yaml:"app_settings"`
	BridgingSettings BridgingSettings     `yaml:"bridging_settings"`
}

type InitialUtxos map[string][]*indexer.TxInputOutput

type CardanoTx struct {
	OriginChainId string `json:"origin_chain_id"`

	indexer.Tx
}

type UtxoTransaction struct {
}

type Utxo struct {
	Address string
	Amount  uint64
}

type BridgingRequestReceiver struct {
	Address string
	Amount  uint64
}

type BridgingRequestClaim struct {
	TxHash             string
	Receivers          []BridgingRequestReceiver
	OutputUtxos        []Utxo
	DestinationChainId string
}

type BatchExecutedClaim struct {
	TxHash       string
	BatchNonceId string
	OutputUtxos  []Utxo
}

type BatchExecutionFailedClaim struct {
	TxHash       string
	BatchNonceId string
}

type RefundRequestClaim struct {
	TxHash               string
	PreviousRefundTxHash string
	RetryCounter         int32
	RefundToAddress      string
	OutputUtxos          []Utxo
	DestinationChainId   string
	UtxoTransaction      UtxoTransaction
}

type RefundExecutedClaim struct {
	TxHash       string
	RefundTxHash string
	OutputUtxos  []Utxo
}

type BridgeClaims struct {
	BridgingRequest      []BridgingRequestClaim
	BatchExecuted        []BatchExecutedClaim
	BatchExecutionFailed []BatchExecutionFailedClaim
	RefundRequest        []RefundRequestClaim
	RefundExecuted       []RefundExecutedClaim
}

func (bc BridgeClaims) Any() bool {
	return len(bc.BridgingRequest) > 0 ||
		len(bc.BatchExecuted) > 0 ||
		len(bc.BatchExecutionFailed) > 0 ||
		len(bc.RefundRequest) > 0 ||
		len(bc.RefundExecuted) > 0
}

func (btx CardanoTx) Key() []byte {
	return []byte(btx.Hash)
}
