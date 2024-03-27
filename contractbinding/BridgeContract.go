// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contractbinding

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// IBridgeContractStructsBatchExecutedClaim is an auto generated low-level Go binding around an user-defined struct.
type IBridgeContractStructsBatchExecutedClaim struct {
	ObservedTransactionHash string
	ChainID                 string
	BatchNonceID            *big.Int
	OutputUTXOs             IBridgeContractStructsUTXOs
}

// IBridgeContractStructsBatchExecutionFailedClaim is an auto generated low-level Go binding around an user-defined struct.
type IBridgeContractStructsBatchExecutionFailedClaim struct {
	ObservedTransactionHash string
	ChainID                 string
	BatchNonceID            *big.Int
}

// IBridgeContractStructsBridgingRequestClaim is an auto generated low-level Go binding around an user-defined struct.
type IBridgeContractStructsBridgingRequestClaim struct {
	ObservedTransactionHash string
	Receivers               []IBridgeContractStructsReceiver
	OutputUTXO              IBridgeContractStructsUTXO
	SourceChainID           string
	DestinationChainID      string
}

// IBridgeContractStructsCardanoBlock is an auto generated low-level Go binding around an user-defined struct.
type IBridgeContractStructsCardanoBlock struct {
	BlockHash string
	BlockSlot uint64
}

// IBridgeContractStructsChain is an auto generated low-level Go binding around an user-defined struct.
type IBridgeContractStructsChain struct {
	Id              string
	Utxos           IBridgeContractStructsUTXOs
	AddressMultisig string
	AddressFeePayer string
	KeyHashMultisig string
	KeyHashFeePayer string
	TokenQuantity   *big.Int
}

// IBridgeContractStructsConfirmedBatch is an auto generated low-level Go binding around an user-defined struct.
type IBridgeContractStructsConfirmedBatch struct {
	Id                         *big.Int
	RawTransaction             string
	MultisigSignatures         []string
	FeePayerMultisigSignatures []string
}

// IBridgeContractStructsConfirmedTransaction is an auto generated low-level Go binding around an user-defined struct.
type IBridgeContractStructsConfirmedTransaction struct {
	Nonce     *big.Int
	Receivers []IBridgeContractStructsReceiver
}

// IBridgeContractStructsReceiver is an auto generated low-level Go binding around an user-defined struct.
type IBridgeContractStructsReceiver struct {
	DestinationAddress string
	Amount             *big.Int
}

// IBridgeContractStructsRefundExecutedClaim is an auto generated low-level Go binding around an user-defined struct.
type IBridgeContractStructsRefundExecutedClaim struct {
	ObservedTransactionHash string
	ChainID                 string
	RefundTxHash            string
	Utxo                    IBridgeContractStructsUTXO
}

// IBridgeContractStructsRefundRequestClaim is an auto generated low-level Go binding around an user-defined struct.
type IBridgeContractStructsRefundRequestClaim struct {
	ObservedTransactionHash string
	PreviousRefundTxHash    string
	ChainID                 string
	Receiver                string
	Utxo                    IBridgeContractStructsUTXO
	RawTransaction          string
	MultisigSignature       string
	RetryCounter            *big.Int
}

// IBridgeContractStructsSignedBatch is an auto generated low-level Go binding around an user-defined struct.
type IBridgeContractStructsSignedBatch struct {
	Id                        *big.Int
	DestinationChainId        string
	RawTransaction            string
	MultisigSignature         string
	FeePayerMultisigSignature string
	IncludedTransactions      []IBridgeContractStructsConfirmedTransaction
	UsedUTXOs                 IBridgeContractStructsUTXOs
}

// IBridgeContractStructsUTXO is an auto generated low-level Go binding around an user-defined struct.
type IBridgeContractStructsUTXO struct {
	TxHash      string
	TxIndex     *big.Int
	AddressUTXO string
	Amount      *big.Int
}

// IBridgeContractStructsUTXOs is an auto generated low-level Go binding around an user-defined struct.
type IBridgeContractStructsUTXOs struct {
	MultisigOwnedUTXOs []IBridgeContractStructsUTXO
	FeePayerOwnedUTXOs []IBridgeContractStructsUTXO
}

// IBridgeContractStructsValidatorClaims is an auto generated low-level Go binding around an user-defined struct.
type IBridgeContractStructsValidatorClaims struct {
	BridgingRequestClaims      []IBridgeContractStructsBridgingRequestClaim
	BatchExecutedClaims        []IBridgeContractStructsBatchExecutedClaim
	BatchExecutionFailedClaims []IBridgeContractStructsBatchExecutionFailedClaim
	RefundRequestClaims        []IBridgeContractStructsRefundRequestClaim
	RefundExecutedClaims       []IBridgeContractStructsRefundExecutedClaim
}

// BridgeContractMetaData contains all meta data concerning the BridgeContract contract.
var BridgeContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_validators\",\"type\":\"address[]\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_claimTransactionHash\",\"type\":\"string\"}],\"name\":\"AlreadyConfirmed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_claimTransactionHash\",\"type\":\"string\"}],\"name\":\"AlreadyProposed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_blockchainID\",\"type\":\"string\"}],\"name\":\"CanNotCreateBatchYet\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_claimId\",\"type\":\"string\"}],\"name\":\"ChainAlreadyRegistered\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_slot\",\"type\":\"uint256\"}],\"name\":\"InvalidSlot\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotBridgeContract\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotClaimsHelper\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotClaimsManager\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotClaimsManagerOrBridgeContract\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_claimTransactionHash\",\"type\":\"string\"}],\"name\":\"NotEnoughBridgingTokensAwailable\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSignedBatchManagerOrBridgeContract\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotValidator\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"RefundRequestClaimNotYetSupporter\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"string\",\"name\":\"chainId\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"newChainProposal\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"string\",\"name\":\"chainId\",\"type\":\"string\"}],\"name\":\"newChainRegistered\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"MAX_NUMBER_OF_BLOCKS\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MAX_NUMBER_OF_TRANSACTIONS\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"confirmedTransactionNounce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAllRegisteredChains\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"id\",\"type\":\"string\"},{\"components\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"addressUTXO\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"multisigOwnedUTXOs\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"addressUTXO\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"feePayerOwnedUTXOs\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeContractStructs.UTXOs\",\"name\":\"utxos\",\"type\":\"tuple\"},{\"internalType\":\"string\",\"name\":\"addressMultisig\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"addressFeePayer\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"keyHashMultisig\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"keyHashFeePayer\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"tokenQuantity\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.Chain[]\",\"name\":\"_chains\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_destinationChain\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txCost\",\"type\":\"uint256\"}],\"name\":\"getAvailableUTXOs\",\"outputs\":[{\"components\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"addressUTXO\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"multisigOwnedUTXOs\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"addressUTXO\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"feePayerOwnedUTXOs\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeContractStructs.UTXOs\",\"name\":\"availableUTXOs\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_destinationChain\",\"type\":\"string\"}],\"name\":\"getConfirmedBatch\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"rawTransaction\",\"type\":\"string\"},{\"internalType\":\"string[]\",\"name\":\"multisigSignatures\",\"type\":\"string[]\"},{\"internalType\":\"string[]\",\"name\":\"feePayerMultisigSignatures\",\"type\":\"string[]\"}],\"internalType\":\"structIBridgeContractStructs.ConfirmedBatch\",\"name\":\"batch\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_destinationChain\",\"type\":\"string\"}],\"name\":\"getConfirmedTransactions\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"destinationAddress\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.Receiver[]\",\"name\":\"receivers\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeContractStructs.ConfirmedTransaction[]\",\"name\":\"_confirmedTransactions\",\"type\":\"tuple[]\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_sourceChain\",\"type\":\"string\"}],\"name\":\"getLastObservedBlock\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"blockHash\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"blockSlot\",\"type\":\"uint64\"}],\"internalType\":\"structIBridgeContractStructs.CardanoBlock\",\"name\":\"cblock\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_hash\",\"type\":\"bytes32\"}],\"name\":\"getNumberOfVotes\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getQuorumNumberOfValidators\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_chainId\",\"type\":\"string\"}],\"name\":\"isChainRegistered\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"lastBatchedClaim\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"lastClaimIncludedInBatch\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"nextTimeoutBlock\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_chainId\",\"type\":\"string\"},{\"components\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"addressUTXO\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"multisigOwnedUTXOs\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"addressUTXO\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"feePayerOwnedUTXOs\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeContractStructs.UTXOs\",\"name\":\"_initialUTXOs\",\"type\":\"tuple\"},{\"internalType\":\"string\",\"name\":\"_addressMultisig\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_addressFeePayer\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_keyHashMultisig\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_keyHashFeePayer\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"_tokenQuantity\",\"type\":\"uint256\"}],\"name\":\"registerChain\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_chainId\",\"type\":\"string\"},{\"components\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"addressUTXO\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"multisigOwnedUTXOs\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"addressUTXO\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"feePayerOwnedUTXOs\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeContractStructs.UTXOs\",\"name\":\"_initialUTXOs\",\"type\":\"tuple\"},{\"internalType\":\"string\",\"name\":\"_addressMultisig\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_addressFeePayer\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_keyHashMultisig\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_keyHashFeePayer\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"_tokenQuantity\",\"type\":\"uint256\"}],\"name\":\"registerChainGovernance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_claimsHelper\",\"type\":\"address\"}],\"name\":\"setClaimsHelper\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_claimsManager\",\"type\":\"address\"}],\"name\":\"setClaimsManager\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_chainId\",\"type\":\"string\"}],\"name\":\"setLastBatchedClaim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_chainId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"_blockNumber\",\"type\":\"uint256\"}],\"name\":\"setNextTimeoutBlock\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_signedBatchManager\",\"type\":\"address\"}],\"name\":\"setSignedBatchManager\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractSlotsManager\",\"name\":\"_slotsManager\",\"type\":\"address\"}],\"name\":\"setSlotsManager\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_utxosManager\",\"type\":\"address\"}],\"name\":\"setUTXOsManager\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_destinationChain\",\"type\":\"string\"}],\"name\":\"shouldCreateBatch\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"batch\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"observedTransactionHash\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"destinationAddress\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.Receiver[]\",\"name\":\"receivers\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"addressUTXO\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO\",\"name\":\"outputUTXO\",\"type\":\"tuple\"},{\"internalType\":\"string\",\"name\":\"sourceChainID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"destinationChainID\",\"type\":\"string\"}],\"internalType\":\"structIBridgeContractStructs.BridgingRequestClaim[]\",\"name\":\"bridgingRequestClaims\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"observedTransactionHash\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"chainID\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"batchNonceID\",\"type\":\"uint256\"},{\"components\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"addressUTXO\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"multisigOwnedUTXOs\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"addressUTXO\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"feePayerOwnedUTXOs\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeContractStructs.UTXOs\",\"name\":\"outputUTXOs\",\"type\":\"tuple\"}],\"internalType\":\"structIBridgeContractStructs.BatchExecutedClaim[]\",\"name\":\"batchExecutedClaims\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"observedTransactionHash\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"chainID\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"batchNonceID\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.BatchExecutionFailedClaim[]\",\"name\":\"batchExecutionFailedClaims\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"observedTransactionHash\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"previousRefundTxHash\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"chainID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"addressUTXO\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO\",\"name\":\"utxo\",\"type\":\"tuple\"},{\"internalType\":\"string\",\"name\":\"rawTransaction\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"multisigSignature\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"retryCounter\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.RefundRequestClaim[]\",\"name\":\"refundRequestClaims\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"observedTransactionHash\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"chainID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"refundTxHash\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"addressUTXO\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO\",\"name\":\"utxo\",\"type\":\"tuple\"}],\"internalType\":\"structIBridgeContractStructs.RefundExecutedClaim[]\",\"name\":\"refundExecutedClaims\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeContractStructs.ValidatorClaims\",\"name\":\"_claims\",\"type\":\"tuple\"}],\"name\":\"submitClaims\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"chainID\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"blockHash\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"blockSlot\",\"type\":\"uint64\"}],\"internalType\":\"structIBridgeContractStructs.CardanoBlock[]\",\"name\":\"blocks\",\"type\":\"tuple[]\"}],\"name\":\"submitLastObservableBlocks\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"destinationChainId\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"rawTransaction\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"multisigSignature\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"feePayerMultisigSignature\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"destinationAddress\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.Receiver[]\",\"name\":\"receivers\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeContractStructs.ConfirmedTransaction[]\",\"name\":\"includedTransactions\",\"type\":\"tuple[]\"},{\"components\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"addressUTXO\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"multisigOwnedUTXOs\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"addressUTXO\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"feePayerOwnedUTXOs\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeContractStructs.UTXOs\",\"name\":\"usedUTXOs\",\"type\":\"tuple\"}],\"internalType\":\"structIBridgeContractStructs.SignedBatch\",\"name\":\"_signedBatch\",\"type\":\"tuple\"}],\"name\":\"submitSignedBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"validatorsCount\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// BridgeContractABI is the input ABI used to generate the binding from.
// Deprecated: Use BridgeContractMetaData.ABI instead.
var BridgeContractABI = BridgeContractMetaData.ABI

// BridgeContract is an auto generated Go binding around an Ethereum contract.
type BridgeContract struct {
	BridgeContractCaller     // Read-only binding to the contract
	BridgeContractTransactor // Write-only binding to the contract
	BridgeContractFilterer   // Log filterer for contract events
}

// BridgeContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type BridgeContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BridgeContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BridgeContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BridgeContractSession struct {
	Contract     *BridgeContract   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BridgeContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BridgeContractCallerSession struct {
	Contract *BridgeContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// BridgeContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BridgeContractTransactorSession struct {
	Contract     *BridgeContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// BridgeContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type BridgeContractRaw struct {
	Contract *BridgeContract // Generic contract binding to access the raw methods on
}

// BridgeContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BridgeContractCallerRaw struct {
	Contract *BridgeContractCaller // Generic read-only contract binding to access the raw methods on
}

// BridgeContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BridgeContractTransactorRaw struct {
	Contract *BridgeContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBridgeContract creates a new instance of BridgeContract, bound to a specific deployed contract.
func NewBridgeContract(address common.Address, backend bind.ContractBackend) (*BridgeContract, error) {
	contract, err := bindBridgeContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BridgeContract{BridgeContractCaller: BridgeContractCaller{contract: contract}, BridgeContractTransactor: BridgeContractTransactor{contract: contract}, BridgeContractFilterer: BridgeContractFilterer{contract: contract}}, nil
}

// NewBridgeContractCaller creates a new read-only instance of BridgeContract, bound to a specific deployed contract.
func NewBridgeContractCaller(address common.Address, caller bind.ContractCaller) (*BridgeContractCaller, error) {
	contract, err := bindBridgeContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BridgeContractCaller{contract: contract}, nil
}

// NewBridgeContractTransactor creates a new write-only instance of BridgeContract, bound to a specific deployed contract.
func NewBridgeContractTransactor(address common.Address, transactor bind.ContractTransactor) (*BridgeContractTransactor, error) {
	contract, err := bindBridgeContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BridgeContractTransactor{contract: contract}, nil
}

// NewBridgeContractFilterer creates a new log filterer instance of BridgeContract, bound to a specific deployed contract.
func NewBridgeContractFilterer(address common.Address, filterer bind.ContractFilterer) (*BridgeContractFilterer, error) {
	contract, err := bindBridgeContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BridgeContractFilterer{contract: contract}, nil
}

// bindBridgeContract binds a generic wrapper to an already deployed contract.
func bindBridgeContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BridgeContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BridgeContract *BridgeContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BridgeContract.Contract.BridgeContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BridgeContract *BridgeContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BridgeContract.Contract.BridgeContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BridgeContract *BridgeContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BridgeContract.Contract.BridgeContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BridgeContract *BridgeContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BridgeContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BridgeContract *BridgeContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BridgeContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BridgeContract *BridgeContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BridgeContract.Contract.contract.Transact(opts, method, params...)
}

// MAXNUMBEROFBLOCKS is a free data retrieval call binding the contract method 0x0250ff37.
//
// Solidity: function MAX_NUMBER_OF_BLOCKS() view returns(uint8)
func (_BridgeContract *BridgeContractCaller) MAXNUMBEROFBLOCKS(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "MAX_NUMBER_OF_BLOCKS")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// MAXNUMBEROFBLOCKS is a free data retrieval call binding the contract method 0x0250ff37.
//
// Solidity: function MAX_NUMBER_OF_BLOCKS() view returns(uint8)
func (_BridgeContract *BridgeContractSession) MAXNUMBEROFBLOCKS() (uint8, error) {
	return _BridgeContract.Contract.MAXNUMBEROFBLOCKS(&_BridgeContract.CallOpts)
}

// MAXNUMBEROFBLOCKS is a free data retrieval call binding the contract method 0x0250ff37.
//
// Solidity: function MAX_NUMBER_OF_BLOCKS() view returns(uint8)
func (_BridgeContract *BridgeContractCallerSession) MAXNUMBEROFBLOCKS() (uint8, error) {
	return _BridgeContract.Contract.MAXNUMBEROFBLOCKS(&_BridgeContract.CallOpts)
}

// MAXNUMBEROFTRANSACTIONS is a free data retrieval call binding the contract method 0x9e579328.
//
// Solidity: function MAX_NUMBER_OF_TRANSACTIONS() view returns(uint16)
func (_BridgeContract *BridgeContractCaller) MAXNUMBEROFTRANSACTIONS(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "MAX_NUMBER_OF_TRANSACTIONS")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// MAXNUMBEROFTRANSACTIONS is a free data retrieval call binding the contract method 0x9e579328.
//
// Solidity: function MAX_NUMBER_OF_TRANSACTIONS() view returns(uint16)
func (_BridgeContract *BridgeContractSession) MAXNUMBEROFTRANSACTIONS() (uint16, error) {
	return _BridgeContract.Contract.MAXNUMBEROFTRANSACTIONS(&_BridgeContract.CallOpts)
}

// MAXNUMBEROFTRANSACTIONS is a free data retrieval call binding the contract method 0x9e579328.
//
// Solidity: function MAX_NUMBER_OF_TRANSACTIONS() view returns(uint16)
func (_BridgeContract *BridgeContractCallerSession) MAXNUMBEROFTRANSACTIONS() (uint16, error) {
	return _BridgeContract.Contract.MAXNUMBEROFTRANSACTIONS(&_BridgeContract.CallOpts)
}

// ConfirmedTransactionNounce is a free data retrieval call binding the contract method 0x5df2bb40.
//
// Solidity: function confirmedTransactionNounce(string ) view returns(uint256)
func (_BridgeContract *BridgeContractCaller) ConfirmedTransactionNounce(opts *bind.CallOpts, arg0 string) (*big.Int, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "confirmedTransactionNounce", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ConfirmedTransactionNounce is a free data retrieval call binding the contract method 0x5df2bb40.
//
// Solidity: function confirmedTransactionNounce(string ) view returns(uint256)
func (_BridgeContract *BridgeContractSession) ConfirmedTransactionNounce(arg0 string) (*big.Int, error) {
	return _BridgeContract.Contract.ConfirmedTransactionNounce(&_BridgeContract.CallOpts, arg0)
}

// ConfirmedTransactionNounce is a free data retrieval call binding the contract method 0x5df2bb40.
//
// Solidity: function confirmedTransactionNounce(string ) view returns(uint256)
func (_BridgeContract *BridgeContractCallerSession) ConfirmedTransactionNounce(arg0 string) (*big.Int, error) {
	return _BridgeContract.Contract.ConfirmedTransactionNounce(&_BridgeContract.CallOpts, arg0)
}

// GetAllRegisteredChains is a free data retrieval call binding the contract method 0x67f0cc44.
//
// Solidity: function getAllRegisteredChains() view returns((string,((string,uint256,string,uint256)[],(string,uint256,string,uint256)[]),string,string,string,string,uint256)[] _chains)
func (_BridgeContract *BridgeContractCaller) GetAllRegisteredChains(opts *bind.CallOpts) ([]IBridgeContractStructsChain, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getAllRegisteredChains")

	if err != nil {
		return *new([]IBridgeContractStructsChain), err
	}

	out0 := *abi.ConvertType(out[0], new([]IBridgeContractStructsChain)).(*[]IBridgeContractStructsChain)

	return out0, err

}

// GetAllRegisteredChains is a free data retrieval call binding the contract method 0x67f0cc44.
//
// Solidity: function getAllRegisteredChains() view returns((string,((string,uint256,string,uint256)[],(string,uint256,string,uint256)[]),string,string,string,string,uint256)[] _chains)
func (_BridgeContract *BridgeContractSession) GetAllRegisteredChains() ([]IBridgeContractStructsChain, error) {
	return _BridgeContract.Contract.GetAllRegisteredChains(&_BridgeContract.CallOpts)
}

// GetAllRegisteredChains is a free data retrieval call binding the contract method 0x67f0cc44.
//
// Solidity: function getAllRegisteredChains() view returns((string,((string,uint256,string,uint256)[],(string,uint256,string,uint256)[]),string,string,string,string,uint256)[] _chains)
func (_BridgeContract *BridgeContractCallerSession) GetAllRegisteredChains() ([]IBridgeContractStructsChain, error) {
	return _BridgeContract.Contract.GetAllRegisteredChains(&_BridgeContract.CallOpts)
}

// GetAvailableUTXOs is a free data retrieval call binding the contract method 0xa1cbdfb1.
//
// Solidity: function getAvailableUTXOs(string _destinationChain, uint256 txCost) view returns(((string,uint256,string,uint256)[],(string,uint256,string,uint256)[]) availableUTXOs)
func (_BridgeContract *BridgeContractCaller) GetAvailableUTXOs(opts *bind.CallOpts, _destinationChain string, txCost *big.Int) (IBridgeContractStructsUTXOs, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getAvailableUTXOs", _destinationChain, txCost)

	if err != nil {
		return *new(IBridgeContractStructsUTXOs), err
	}

	out0 := *abi.ConvertType(out[0], new(IBridgeContractStructsUTXOs)).(*IBridgeContractStructsUTXOs)

	return out0, err

}

// GetAvailableUTXOs is a free data retrieval call binding the contract method 0xa1cbdfb1.
//
// Solidity: function getAvailableUTXOs(string _destinationChain, uint256 txCost) view returns(((string,uint256,string,uint256)[],(string,uint256,string,uint256)[]) availableUTXOs)
func (_BridgeContract *BridgeContractSession) GetAvailableUTXOs(_destinationChain string, txCost *big.Int) (IBridgeContractStructsUTXOs, error) {
	return _BridgeContract.Contract.GetAvailableUTXOs(&_BridgeContract.CallOpts, _destinationChain, txCost)
}

// GetAvailableUTXOs is a free data retrieval call binding the contract method 0xa1cbdfb1.
//
// Solidity: function getAvailableUTXOs(string _destinationChain, uint256 txCost) view returns(((string,uint256,string,uint256)[],(string,uint256,string,uint256)[]) availableUTXOs)
func (_BridgeContract *BridgeContractCallerSession) GetAvailableUTXOs(_destinationChain string, txCost *big.Int) (IBridgeContractStructsUTXOs, error) {
	return _BridgeContract.Contract.GetAvailableUTXOs(&_BridgeContract.CallOpts, _destinationChain, txCost)
}

// GetConfirmedBatch is a free data retrieval call binding the contract method 0xd52c54c4.
//
// Solidity: function getConfirmedBatch(string _destinationChain) view returns((uint256,string,string[],string[]) batch)
func (_BridgeContract *BridgeContractCaller) GetConfirmedBatch(opts *bind.CallOpts, _destinationChain string) (IBridgeContractStructsConfirmedBatch, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getConfirmedBatch", _destinationChain)

	if err != nil {
		return *new(IBridgeContractStructsConfirmedBatch), err
	}

	out0 := *abi.ConvertType(out[0], new(IBridgeContractStructsConfirmedBatch)).(*IBridgeContractStructsConfirmedBatch)

	return out0, err

}

// GetConfirmedBatch is a free data retrieval call binding the contract method 0xd52c54c4.
//
// Solidity: function getConfirmedBatch(string _destinationChain) view returns((uint256,string,string[],string[]) batch)
func (_BridgeContract *BridgeContractSession) GetConfirmedBatch(_destinationChain string) (IBridgeContractStructsConfirmedBatch, error) {
	return _BridgeContract.Contract.GetConfirmedBatch(&_BridgeContract.CallOpts, _destinationChain)
}

// GetConfirmedBatch is a free data retrieval call binding the contract method 0xd52c54c4.
//
// Solidity: function getConfirmedBatch(string _destinationChain) view returns((uint256,string,string[],string[]) batch)
func (_BridgeContract *BridgeContractCallerSession) GetConfirmedBatch(_destinationChain string) (IBridgeContractStructsConfirmedBatch, error) {
	return _BridgeContract.Contract.GetConfirmedBatch(&_BridgeContract.CallOpts, _destinationChain)
}

// GetLastObservedBlock is a free data retrieval call binding the contract method 0x2175c3f7.
//
// Solidity: function getLastObservedBlock(string _sourceChain) view returns((string,uint64) cblock)
func (_BridgeContract *BridgeContractCaller) GetLastObservedBlock(opts *bind.CallOpts, _sourceChain string) (IBridgeContractStructsCardanoBlock, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getLastObservedBlock", _sourceChain)

	if err != nil {
		return *new(IBridgeContractStructsCardanoBlock), err
	}

	out0 := *abi.ConvertType(out[0], new(IBridgeContractStructsCardanoBlock)).(*IBridgeContractStructsCardanoBlock)

	return out0, err

}

// GetLastObservedBlock is a free data retrieval call binding the contract method 0x2175c3f7.
//
// Solidity: function getLastObservedBlock(string _sourceChain) view returns((string,uint64) cblock)
func (_BridgeContract *BridgeContractSession) GetLastObservedBlock(_sourceChain string) (IBridgeContractStructsCardanoBlock, error) {
	return _BridgeContract.Contract.GetLastObservedBlock(&_BridgeContract.CallOpts, _sourceChain)
}

// GetLastObservedBlock is a free data retrieval call binding the contract method 0x2175c3f7.
//
// Solidity: function getLastObservedBlock(string _sourceChain) view returns((string,uint64) cblock)
func (_BridgeContract *BridgeContractCallerSession) GetLastObservedBlock(_sourceChain string) (IBridgeContractStructsCardanoBlock, error) {
	return _BridgeContract.Contract.GetLastObservedBlock(&_BridgeContract.CallOpts, _sourceChain)
}

// GetNumberOfVotes is a free data retrieval call binding the contract method 0x8d2d814c.
//
// Solidity: function getNumberOfVotes(bytes32 _hash) view returns(uint8)
func (_BridgeContract *BridgeContractCaller) GetNumberOfVotes(opts *bind.CallOpts, _hash [32]byte) (uint8, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getNumberOfVotes", _hash)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetNumberOfVotes is a free data retrieval call binding the contract method 0x8d2d814c.
//
// Solidity: function getNumberOfVotes(bytes32 _hash) view returns(uint8)
func (_BridgeContract *BridgeContractSession) GetNumberOfVotes(_hash [32]byte) (uint8, error) {
	return _BridgeContract.Contract.GetNumberOfVotes(&_BridgeContract.CallOpts, _hash)
}

// GetNumberOfVotes is a free data retrieval call binding the contract method 0x8d2d814c.
//
// Solidity: function getNumberOfVotes(bytes32 _hash) view returns(uint8)
func (_BridgeContract *BridgeContractCallerSession) GetNumberOfVotes(_hash [32]byte) (uint8, error) {
	return _BridgeContract.Contract.GetNumberOfVotes(&_BridgeContract.CallOpts, _hash)
}

// GetQuorumNumberOfValidators is a free data retrieval call binding the contract method 0xd8718da0.
//
// Solidity: function getQuorumNumberOfValidators() view returns(uint8)
func (_BridgeContract *BridgeContractCaller) GetQuorumNumberOfValidators(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getQuorumNumberOfValidators")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetQuorumNumberOfValidators is a free data retrieval call binding the contract method 0xd8718da0.
//
// Solidity: function getQuorumNumberOfValidators() view returns(uint8)
func (_BridgeContract *BridgeContractSession) GetQuorumNumberOfValidators() (uint8, error) {
	return _BridgeContract.Contract.GetQuorumNumberOfValidators(&_BridgeContract.CallOpts)
}

// GetQuorumNumberOfValidators is a free data retrieval call binding the contract method 0xd8718da0.
//
// Solidity: function getQuorumNumberOfValidators() view returns(uint8)
func (_BridgeContract *BridgeContractCallerSession) GetQuorumNumberOfValidators() (uint8, error) {
	return _BridgeContract.Contract.GetQuorumNumberOfValidators(&_BridgeContract.CallOpts)
}

// IsChainRegistered is a free data retrieval call binding the contract method 0x18c586cd.
//
// Solidity: function isChainRegistered(string _chainId) view returns(bool)
func (_BridgeContract *BridgeContractCaller) IsChainRegistered(opts *bind.CallOpts, _chainId string) (bool, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "isChainRegistered", _chainId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsChainRegistered is a free data retrieval call binding the contract method 0x18c586cd.
//
// Solidity: function isChainRegistered(string _chainId) view returns(bool)
func (_BridgeContract *BridgeContractSession) IsChainRegistered(_chainId string) (bool, error) {
	return _BridgeContract.Contract.IsChainRegistered(&_BridgeContract.CallOpts, _chainId)
}

// IsChainRegistered is a free data retrieval call binding the contract method 0x18c586cd.
//
// Solidity: function isChainRegistered(string _chainId) view returns(bool)
func (_BridgeContract *BridgeContractCallerSession) IsChainRegistered(_chainId string) (bool, error) {
	return _BridgeContract.Contract.IsChainRegistered(&_BridgeContract.CallOpts, _chainId)
}

// LastBatchedClaim is a free data retrieval call binding the contract method 0xda49f330.
//
// Solidity: function lastBatchedClaim(string ) view returns(uint256)
func (_BridgeContract *BridgeContractCaller) LastBatchedClaim(opts *bind.CallOpts, arg0 string) (*big.Int, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "lastBatchedClaim", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastBatchedClaim is a free data retrieval call binding the contract method 0xda49f330.
//
// Solidity: function lastBatchedClaim(string ) view returns(uint256)
func (_BridgeContract *BridgeContractSession) LastBatchedClaim(arg0 string) (*big.Int, error) {
	return _BridgeContract.Contract.LastBatchedClaim(&_BridgeContract.CallOpts, arg0)
}

// LastBatchedClaim is a free data retrieval call binding the contract method 0xda49f330.
//
// Solidity: function lastBatchedClaim(string ) view returns(uint256)
func (_BridgeContract *BridgeContractCallerSession) LastBatchedClaim(arg0 string) (*big.Int, error) {
	return _BridgeContract.Contract.LastBatchedClaim(&_BridgeContract.CallOpts, arg0)
}

// LastClaimIncludedInBatch is a free data retrieval call binding the contract method 0x3489ad17.
//
// Solidity: function lastClaimIncludedInBatch(string ) view returns(uint256)
func (_BridgeContract *BridgeContractCaller) LastClaimIncludedInBatch(opts *bind.CallOpts, arg0 string) (*big.Int, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "lastClaimIncludedInBatch", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastClaimIncludedInBatch is a free data retrieval call binding the contract method 0x3489ad17.
//
// Solidity: function lastClaimIncludedInBatch(string ) view returns(uint256)
func (_BridgeContract *BridgeContractSession) LastClaimIncludedInBatch(arg0 string) (*big.Int, error) {
	return _BridgeContract.Contract.LastClaimIncludedInBatch(&_BridgeContract.CallOpts, arg0)
}

// LastClaimIncludedInBatch is a free data retrieval call binding the contract method 0x3489ad17.
//
// Solidity: function lastClaimIncludedInBatch(string ) view returns(uint256)
func (_BridgeContract *BridgeContractCallerSession) LastClaimIncludedInBatch(arg0 string) (*big.Int, error) {
	return _BridgeContract.Contract.LastClaimIncludedInBatch(&_BridgeContract.CallOpts, arg0)
}

// NextTimeoutBlock is a free data retrieval call binding the contract method 0x4ae4dd7c.
//
// Solidity: function nextTimeoutBlock(string ) view returns(uint256)
func (_BridgeContract *BridgeContractCaller) NextTimeoutBlock(opts *bind.CallOpts, arg0 string) (*big.Int, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "nextTimeoutBlock", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NextTimeoutBlock is a free data retrieval call binding the contract method 0x4ae4dd7c.
//
// Solidity: function nextTimeoutBlock(string ) view returns(uint256)
func (_BridgeContract *BridgeContractSession) NextTimeoutBlock(arg0 string) (*big.Int, error) {
	return _BridgeContract.Contract.NextTimeoutBlock(&_BridgeContract.CallOpts, arg0)
}

// NextTimeoutBlock is a free data retrieval call binding the contract method 0x4ae4dd7c.
//
// Solidity: function nextTimeoutBlock(string ) view returns(uint256)
func (_BridgeContract *BridgeContractCallerSession) NextTimeoutBlock(arg0 string) (*big.Int, error) {
	return _BridgeContract.Contract.NextTimeoutBlock(&_BridgeContract.CallOpts, arg0)
}

// ShouldCreateBatch is a free data retrieval call binding the contract method 0x77968b34.
//
// Solidity: function shouldCreateBatch(string _destinationChain) view returns(bool batch)
func (_BridgeContract *BridgeContractCaller) ShouldCreateBatch(opts *bind.CallOpts, _destinationChain string) (bool, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "shouldCreateBatch", _destinationChain)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// ShouldCreateBatch is a free data retrieval call binding the contract method 0x77968b34.
//
// Solidity: function shouldCreateBatch(string _destinationChain) view returns(bool batch)
func (_BridgeContract *BridgeContractSession) ShouldCreateBatch(_destinationChain string) (bool, error) {
	return _BridgeContract.Contract.ShouldCreateBatch(&_BridgeContract.CallOpts, _destinationChain)
}

// ShouldCreateBatch is a free data retrieval call binding the contract method 0x77968b34.
//
// Solidity: function shouldCreateBatch(string _destinationChain) view returns(bool batch)
func (_BridgeContract *BridgeContractCallerSession) ShouldCreateBatch(_destinationChain string) (bool, error) {
	return _BridgeContract.Contract.ShouldCreateBatch(&_BridgeContract.CallOpts, _destinationChain)
}

// ValidatorsCount is a free data retrieval call binding the contract method 0xed612f8c.
//
// Solidity: function validatorsCount() view returns(uint8)
func (_BridgeContract *BridgeContractCaller) ValidatorsCount(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "validatorsCount")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// ValidatorsCount is a free data retrieval call binding the contract method 0xed612f8c.
//
// Solidity: function validatorsCount() view returns(uint8)
func (_BridgeContract *BridgeContractSession) ValidatorsCount() (uint8, error) {
	return _BridgeContract.Contract.ValidatorsCount(&_BridgeContract.CallOpts)
}

// ValidatorsCount is a free data retrieval call binding the contract method 0xed612f8c.
//
// Solidity: function validatorsCount() view returns(uint8)
func (_BridgeContract *BridgeContractCallerSession) ValidatorsCount() (uint8, error) {
	return _BridgeContract.Contract.ValidatorsCount(&_BridgeContract.CallOpts)
}

// GetConfirmedTransactions is a paid mutator transaction binding the contract method 0x595051f9.
//
// Solidity: function getConfirmedTransactions(string _destinationChain) returns((uint256,(string,uint256)[])[] _confirmedTransactions)
func (_BridgeContract *BridgeContractTransactor) GetConfirmedTransactions(opts *bind.TransactOpts, _destinationChain string) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "getConfirmedTransactions", _destinationChain)
}

// GetConfirmedTransactions is a paid mutator transaction binding the contract method 0x595051f9.
//
// Solidity: function getConfirmedTransactions(string _destinationChain) returns((uint256,(string,uint256)[])[] _confirmedTransactions)
func (_BridgeContract *BridgeContractSession) GetConfirmedTransactions(_destinationChain string) (*types.Transaction, error) {
	return _BridgeContract.Contract.GetConfirmedTransactions(&_BridgeContract.TransactOpts, _destinationChain)
}

// GetConfirmedTransactions is a paid mutator transaction binding the contract method 0x595051f9.
//
// Solidity: function getConfirmedTransactions(string _destinationChain) returns((uint256,(string,uint256)[])[] _confirmedTransactions)
func (_BridgeContract *BridgeContractTransactorSession) GetConfirmedTransactions(_destinationChain string) (*types.Transaction, error) {
	return _BridgeContract.Contract.GetConfirmedTransactions(&_BridgeContract.TransactOpts, _destinationChain)
}

// RegisterChain is a paid mutator transaction binding the contract method 0xc3a6791b.
//
// Solidity: function registerChain(string _chainId, ((string,uint256,string,uint256)[],(string,uint256,string,uint256)[]) _initialUTXOs, string _addressMultisig, string _addressFeePayer, string _keyHashMultisig, string _keyHashFeePayer, uint256 _tokenQuantity) returns()
func (_BridgeContract *BridgeContractTransactor) RegisterChain(opts *bind.TransactOpts, _chainId string, _initialUTXOs IBridgeContractStructsUTXOs, _addressMultisig string, _addressFeePayer string, _keyHashMultisig string, _keyHashFeePayer string, _tokenQuantity *big.Int) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "registerChain", _chainId, _initialUTXOs, _addressMultisig, _addressFeePayer, _keyHashMultisig, _keyHashFeePayer, _tokenQuantity)
}

// RegisterChain is a paid mutator transaction binding the contract method 0xc3a6791b.
//
// Solidity: function registerChain(string _chainId, ((string,uint256,string,uint256)[],(string,uint256,string,uint256)[]) _initialUTXOs, string _addressMultisig, string _addressFeePayer, string _keyHashMultisig, string _keyHashFeePayer, uint256 _tokenQuantity) returns()
func (_BridgeContract *BridgeContractSession) RegisterChain(_chainId string, _initialUTXOs IBridgeContractStructsUTXOs, _addressMultisig string, _addressFeePayer string, _keyHashMultisig string, _keyHashFeePayer string, _tokenQuantity *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.RegisterChain(&_BridgeContract.TransactOpts, _chainId, _initialUTXOs, _addressMultisig, _addressFeePayer, _keyHashMultisig, _keyHashFeePayer, _tokenQuantity)
}

// RegisterChain is a paid mutator transaction binding the contract method 0xc3a6791b.
//
// Solidity: function registerChain(string _chainId, ((string,uint256,string,uint256)[],(string,uint256,string,uint256)[]) _initialUTXOs, string _addressMultisig, string _addressFeePayer, string _keyHashMultisig, string _keyHashFeePayer, uint256 _tokenQuantity) returns()
func (_BridgeContract *BridgeContractTransactorSession) RegisterChain(_chainId string, _initialUTXOs IBridgeContractStructsUTXOs, _addressMultisig string, _addressFeePayer string, _keyHashMultisig string, _keyHashFeePayer string, _tokenQuantity *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.RegisterChain(&_BridgeContract.TransactOpts, _chainId, _initialUTXOs, _addressMultisig, _addressFeePayer, _keyHashMultisig, _keyHashFeePayer, _tokenQuantity)
}

// RegisterChainGovernance is a paid mutator transaction binding the contract method 0xfe295440.
//
// Solidity: function registerChainGovernance(string _chainId, ((string,uint256,string,uint256)[],(string,uint256,string,uint256)[]) _initialUTXOs, string _addressMultisig, string _addressFeePayer, string _keyHashMultisig, string _keyHashFeePayer, uint256 _tokenQuantity) returns()
func (_BridgeContract *BridgeContractTransactor) RegisterChainGovernance(opts *bind.TransactOpts, _chainId string, _initialUTXOs IBridgeContractStructsUTXOs, _addressMultisig string, _addressFeePayer string, _keyHashMultisig string, _keyHashFeePayer string, _tokenQuantity *big.Int) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "registerChainGovernance", _chainId, _initialUTXOs, _addressMultisig, _addressFeePayer, _keyHashMultisig, _keyHashFeePayer, _tokenQuantity)
}

// RegisterChainGovernance is a paid mutator transaction binding the contract method 0xfe295440.
//
// Solidity: function registerChainGovernance(string _chainId, ((string,uint256,string,uint256)[],(string,uint256,string,uint256)[]) _initialUTXOs, string _addressMultisig, string _addressFeePayer, string _keyHashMultisig, string _keyHashFeePayer, uint256 _tokenQuantity) returns()
func (_BridgeContract *BridgeContractSession) RegisterChainGovernance(_chainId string, _initialUTXOs IBridgeContractStructsUTXOs, _addressMultisig string, _addressFeePayer string, _keyHashMultisig string, _keyHashFeePayer string, _tokenQuantity *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.RegisterChainGovernance(&_BridgeContract.TransactOpts, _chainId, _initialUTXOs, _addressMultisig, _addressFeePayer, _keyHashMultisig, _keyHashFeePayer, _tokenQuantity)
}

// RegisterChainGovernance is a paid mutator transaction binding the contract method 0xfe295440.
//
// Solidity: function registerChainGovernance(string _chainId, ((string,uint256,string,uint256)[],(string,uint256,string,uint256)[]) _initialUTXOs, string _addressMultisig, string _addressFeePayer, string _keyHashMultisig, string _keyHashFeePayer, uint256 _tokenQuantity) returns()
func (_BridgeContract *BridgeContractTransactorSession) RegisterChainGovernance(_chainId string, _initialUTXOs IBridgeContractStructsUTXOs, _addressMultisig string, _addressFeePayer string, _keyHashMultisig string, _keyHashFeePayer string, _tokenQuantity *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.RegisterChainGovernance(&_BridgeContract.TransactOpts, _chainId, _initialUTXOs, _addressMultisig, _addressFeePayer, _keyHashMultisig, _keyHashFeePayer, _tokenQuantity)
}

// SetClaimsHelper is a paid mutator transaction binding the contract method 0x3021bf78.
//
// Solidity: function setClaimsHelper(address _claimsHelper) returns()
func (_BridgeContract *BridgeContractTransactor) SetClaimsHelper(opts *bind.TransactOpts, _claimsHelper common.Address) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "setClaimsHelper", _claimsHelper)
}

// SetClaimsHelper is a paid mutator transaction binding the contract method 0x3021bf78.
//
// Solidity: function setClaimsHelper(address _claimsHelper) returns()
func (_BridgeContract *BridgeContractSession) SetClaimsHelper(_claimsHelper common.Address) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetClaimsHelper(&_BridgeContract.TransactOpts, _claimsHelper)
}

// SetClaimsHelper is a paid mutator transaction binding the contract method 0x3021bf78.
//
// Solidity: function setClaimsHelper(address _claimsHelper) returns()
func (_BridgeContract *BridgeContractTransactorSession) SetClaimsHelper(_claimsHelper common.Address) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetClaimsHelper(&_BridgeContract.TransactOpts, _claimsHelper)
}

// SetClaimsManager is a paid mutator transaction binding the contract method 0x73f195a9.
//
// Solidity: function setClaimsManager(address _claimsManager) returns()
func (_BridgeContract *BridgeContractTransactor) SetClaimsManager(opts *bind.TransactOpts, _claimsManager common.Address) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "setClaimsManager", _claimsManager)
}

// SetClaimsManager is a paid mutator transaction binding the contract method 0x73f195a9.
//
// Solidity: function setClaimsManager(address _claimsManager) returns()
func (_BridgeContract *BridgeContractSession) SetClaimsManager(_claimsManager common.Address) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetClaimsManager(&_BridgeContract.TransactOpts, _claimsManager)
}

// SetClaimsManager is a paid mutator transaction binding the contract method 0x73f195a9.
//
// Solidity: function setClaimsManager(address _claimsManager) returns()
func (_BridgeContract *BridgeContractTransactorSession) SetClaimsManager(_claimsManager common.Address) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetClaimsManager(&_BridgeContract.TransactOpts, _claimsManager)
}

// SetLastBatchedClaim is a paid mutator transaction binding the contract method 0x8489cef1.
//
// Solidity: function setLastBatchedClaim(string _chainId) returns()
func (_BridgeContract *BridgeContractTransactor) SetLastBatchedClaim(opts *bind.TransactOpts, _chainId string) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "setLastBatchedClaim", _chainId)
}

// SetLastBatchedClaim is a paid mutator transaction binding the contract method 0x8489cef1.
//
// Solidity: function setLastBatchedClaim(string _chainId) returns()
func (_BridgeContract *BridgeContractSession) SetLastBatchedClaim(_chainId string) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetLastBatchedClaim(&_BridgeContract.TransactOpts, _chainId)
}

// SetLastBatchedClaim is a paid mutator transaction binding the contract method 0x8489cef1.
//
// Solidity: function setLastBatchedClaim(string _chainId) returns()
func (_BridgeContract *BridgeContractTransactorSession) SetLastBatchedClaim(_chainId string) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetLastBatchedClaim(&_BridgeContract.TransactOpts, _chainId)
}

// SetNextTimeoutBlock is a paid mutator transaction binding the contract method 0x0d32b63e.
//
// Solidity: function setNextTimeoutBlock(string _chainId, uint256 _blockNumber) returns()
func (_BridgeContract *BridgeContractTransactor) SetNextTimeoutBlock(opts *bind.TransactOpts, _chainId string, _blockNumber *big.Int) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "setNextTimeoutBlock", _chainId, _blockNumber)
}

// SetNextTimeoutBlock is a paid mutator transaction binding the contract method 0x0d32b63e.
//
// Solidity: function setNextTimeoutBlock(string _chainId, uint256 _blockNumber) returns()
func (_BridgeContract *BridgeContractSession) SetNextTimeoutBlock(_chainId string, _blockNumber *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetNextTimeoutBlock(&_BridgeContract.TransactOpts, _chainId, _blockNumber)
}

// SetNextTimeoutBlock is a paid mutator transaction binding the contract method 0x0d32b63e.
//
// Solidity: function setNextTimeoutBlock(string _chainId, uint256 _blockNumber) returns()
func (_BridgeContract *BridgeContractTransactorSession) SetNextTimeoutBlock(_chainId string, _blockNumber *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetNextTimeoutBlock(&_BridgeContract.TransactOpts, _chainId, _blockNumber)
}

// SetSignedBatchManager is a paid mutator transaction binding the contract method 0x49ae9ba8.
//
// Solidity: function setSignedBatchManager(address _signedBatchManager) returns()
func (_BridgeContract *BridgeContractTransactor) SetSignedBatchManager(opts *bind.TransactOpts, _signedBatchManager common.Address) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "setSignedBatchManager", _signedBatchManager)
}

// SetSignedBatchManager is a paid mutator transaction binding the contract method 0x49ae9ba8.
//
// Solidity: function setSignedBatchManager(address _signedBatchManager) returns()
func (_BridgeContract *BridgeContractSession) SetSignedBatchManager(_signedBatchManager common.Address) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetSignedBatchManager(&_BridgeContract.TransactOpts, _signedBatchManager)
}

// SetSignedBatchManager is a paid mutator transaction binding the contract method 0x49ae9ba8.
//
// Solidity: function setSignedBatchManager(address _signedBatchManager) returns()
func (_BridgeContract *BridgeContractTransactorSession) SetSignedBatchManager(_signedBatchManager common.Address) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetSignedBatchManager(&_BridgeContract.TransactOpts, _signedBatchManager)
}

// SetSlotsManager is a paid mutator transaction binding the contract method 0x45793361.
//
// Solidity: function setSlotsManager(address _slotsManager) returns()
func (_BridgeContract *BridgeContractTransactor) SetSlotsManager(opts *bind.TransactOpts, _slotsManager common.Address) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "setSlotsManager", _slotsManager)
}

// SetSlotsManager is a paid mutator transaction binding the contract method 0x45793361.
//
// Solidity: function setSlotsManager(address _slotsManager) returns()
func (_BridgeContract *BridgeContractSession) SetSlotsManager(_slotsManager common.Address) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetSlotsManager(&_BridgeContract.TransactOpts, _slotsManager)
}

// SetSlotsManager is a paid mutator transaction binding the contract method 0x45793361.
//
// Solidity: function setSlotsManager(address _slotsManager) returns()
func (_BridgeContract *BridgeContractTransactorSession) SetSlotsManager(_slotsManager common.Address) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetSlotsManager(&_BridgeContract.TransactOpts, _slotsManager)
}

// SetUTXOsManager is a paid mutator transaction binding the contract method 0x82e82f14.
//
// Solidity: function setUTXOsManager(address _utxosManager) returns()
func (_BridgeContract *BridgeContractTransactor) SetUTXOsManager(opts *bind.TransactOpts, _utxosManager common.Address) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "setUTXOsManager", _utxosManager)
}

// SetUTXOsManager is a paid mutator transaction binding the contract method 0x82e82f14.
//
// Solidity: function setUTXOsManager(address _utxosManager) returns()
func (_BridgeContract *BridgeContractSession) SetUTXOsManager(_utxosManager common.Address) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetUTXOsManager(&_BridgeContract.TransactOpts, _utxosManager)
}

// SetUTXOsManager is a paid mutator transaction binding the contract method 0x82e82f14.
//
// Solidity: function setUTXOsManager(address _utxosManager) returns()
func (_BridgeContract *BridgeContractTransactorSession) SetUTXOsManager(_utxosManager common.Address) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetUTXOsManager(&_BridgeContract.TransactOpts, _utxosManager)
}

// SubmitClaims is a paid mutator transaction binding the contract method 0x08dc0ad1.
//
// Solidity: function submitClaims(((string,(string,uint256)[],(string,uint256,string,uint256),string,string)[],(string,string,uint256,((string,uint256,string,uint256)[],(string,uint256,string,uint256)[]))[],(string,string,uint256)[],(string,string,string,string,(string,uint256,string,uint256),string,string,uint256)[],(string,string,string,(string,uint256,string,uint256))[]) _claims) returns()
func (_BridgeContract *BridgeContractTransactor) SubmitClaims(opts *bind.TransactOpts, _claims IBridgeContractStructsValidatorClaims) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "submitClaims", _claims)
}

// SubmitClaims is a paid mutator transaction binding the contract method 0x08dc0ad1.
//
// Solidity: function submitClaims(((string,(string,uint256)[],(string,uint256,string,uint256),string,string)[],(string,string,uint256,((string,uint256,string,uint256)[],(string,uint256,string,uint256)[]))[],(string,string,uint256)[],(string,string,string,string,(string,uint256,string,uint256),string,string,uint256)[],(string,string,string,(string,uint256,string,uint256))[]) _claims) returns()
func (_BridgeContract *BridgeContractSession) SubmitClaims(_claims IBridgeContractStructsValidatorClaims) (*types.Transaction, error) {
	return _BridgeContract.Contract.SubmitClaims(&_BridgeContract.TransactOpts, _claims)
}

// SubmitClaims is a paid mutator transaction binding the contract method 0x08dc0ad1.
//
// Solidity: function submitClaims(((string,(string,uint256)[],(string,uint256,string,uint256),string,string)[],(string,string,uint256,((string,uint256,string,uint256)[],(string,uint256,string,uint256)[]))[],(string,string,uint256)[],(string,string,string,string,(string,uint256,string,uint256),string,string,uint256)[],(string,string,string,(string,uint256,string,uint256))[]) _claims) returns()
func (_BridgeContract *BridgeContractTransactorSession) SubmitClaims(_claims IBridgeContractStructsValidatorClaims) (*types.Transaction, error) {
	return _BridgeContract.Contract.SubmitClaims(&_BridgeContract.TransactOpts, _claims)
}

// SubmitLastObservableBlocks is a paid mutator transaction binding the contract method 0x7caea5b9.
//
// Solidity: function submitLastObservableBlocks(string chainID, (string,uint64)[] blocks) returns()
func (_BridgeContract *BridgeContractTransactor) SubmitLastObservableBlocks(opts *bind.TransactOpts, chainID string, blocks []IBridgeContractStructsCardanoBlock) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "submitLastObservableBlocks", chainID, blocks)
}

// SubmitLastObservableBlocks is a paid mutator transaction binding the contract method 0x7caea5b9.
//
// Solidity: function submitLastObservableBlocks(string chainID, (string,uint64)[] blocks) returns()
func (_BridgeContract *BridgeContractSession) SubmitLastObservableBlocks(chainID string, blocks []IBridgeContractStructsCardanoBlock) (*types.Transaction, error) {
	return _BridgeContract.Contract.SubmitLastObservableBlocks(&_BridgeContract.TransactOpts, chainID, blocks)
}

// SubmitLastObservableBlocks is a paid mutator transaction binding the contract method 0x7caea5b9.
//
// Solidity: function submitLastObservableBlocks(string chainID, (string,uint64)[] blocks) returns()
func (_BridgeContract *BridgeContractTransactorSession) SubmitLastObservableBlocks(chainID string, blocks []IBridgeContractStructsCardanoBlock) (*types.Transaction, error) {
	return _BridgeContract.Contract.SubmitLastObservableBlocks(&_BridgeContract.TransactOpts, chainID, blocks)
}

// SubmitSignedBatch is a paid mutator transaction binding the contract method 0xa5637dc5.
//
// Solidity: function submitSignedBatch((uint256,string,string,string,string,(uint256,(string,uint256)[])[],((string,uint256,string,uint256)[],(string,uint256,string,uint256)[])) _signedBatch) returns()
func (_BridgeContract *BridgeContractTransactor) SubmitSignedBatch(opts *bind.TransactOpts, _signedBatch IBridgeContractStructsSignedBatch) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "submitSignedBatch", _signedBatch)
}

// SubmitSignedBatch is a paid mutator transaction binding the contract method 0xa5637dc5.
//
// Solidity: function submitSignedBatch((uint256,string,string,string,string,(uint256,(string,uint256)[])[],((string,uint256,string,uint256)[],(string,uint256,string,uint256)[])) _signedBatch) returns()
func (_BridgeContract *BridgeContractSession) SubmitSignedBatch(_signedBatch IBridgeContractStructsSignedBatch) (*types.Transaction, error) {
	return _BridgeContract.Contract.SubmitSignedBatch(&_BridgeContract.TransactOpts, _signedBatch)
}

// SubmitSignedBatch is a paid mutator transaction binding the contract method 0xa5637dc5.
//
// Solidity: function submitSignedBatch((uint256,string,string,string,string,(uint256,(string,uint256)[])[],((string,uint256,string,uint256)[],(string,uint256,string,uint256)[])) _signedBatch) returns()
func (_BridgeContract *BridgeContractTransactorSession) SubmitSignedBatch(_signedBatch IBridgeContractStructsSignedBatch) (*types.Transaction, error) {
	return _BridgeContract.Contract.SubmitSignedBatch(&_BridgeContract.TransactOpts, _signedBatch)
}

// BridgeContractNewChainProposalIterator is returned from FilterNewChainProposal and is used to iterate over the raw logs and unpacked data for NewChainProposal events raised by the BridgeContract contract.
type BridgeContractNewChainProposalIterator struct {
	Event *BridgeContractNewChainProposal // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeContractNewChainProposalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractNewChainProposal)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeContractNewChainProposal)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeContractNewChainProposalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractNewChainProposalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractNewChainProposal represents a NewChainProposal event raised by the BridgeContract contract.
type BridgeContractNewChainProposal struct {
	ChainId common.Hash
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNewChainProposal is a free log retrieval operation binding the contract event 0x99960385426dfd945f1af41c805b3ce369f9f0585b1a7f48ed778e026d2caaae.
//
// Solidity: event newChainProposal(string indexed chainId, address indexed sender)
func (_BridgeContract *BridgeContractFilterer) FilterNewChainProposal(opts *bind.FilterOpts, chainId []string, sender []common.Address) (*BridgeContractNewChainProposalIterator, error) {

	var chainIdRule []interface{}
	for _, chainIdItem := range chainId {
		chainIdRule = append(chainIdRule, chainIdItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "newChainProposal", chainIdRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &BridgeContractNewChainProposalIterator{contract: _BridgeContract.contract, event: "newChainProposal", logs: logs, sub: sub}, nil
}

// WatchNewChainProposal is a free log subscription operation binding the contract event 0x99960385426dfd945f1af41c805b3ce369f9f0585b1a7f48ed778e026d2caaae.
//
// Solidity: event newChainProposal(string indexed chainId, address indexed sender)
func (_BridgeContract *BridgeContractFilterer) WatchNewChainProposal(opts *bind.WatchOpts, sink chan<- *BridgeContractNewChainProposal, chainId []string, sender []common.Address) (event.Subscription, error) {

	var chainIdRule []interface{}
	for _, chainIdItem := range chainId {
		chainIdRule = append(chainIdRule, chainIdItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "newChainProposal", chainIdRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractNewChainProposal)
				if err := _BridgeContract.contract.UnpackLog(event, "newChainProposal", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNewChainProposal is a log parse operation binding the contract event 0x99960385426dfd945f1af41c805b3ce369f9f0585b1a7f48ed778e026d2caaae.
//
// Solidity: event newChainProposal(string indexed chainId, address indexed sender)
func (_BridgeContract *BridgeContractFilterer) ParseNewChainProposal(log types.Log) (*BridgeContractNewChainProposal, error) {
	event := new(BridgeContractNewChainProposal)
	if err := _BridgeContract.contract.UnpackLog(event, "newChainProposal", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeContractNewChainRegisteredIterator is returned from FilterNewChainRegistered and is used to iterate over the raw logs and unpacked data for NewChainRegistered events raised by the BridgeContract contract.
type BridgeContractNewChainRegisteredIterator struct {
	Event *BridgeContractNewChainRegistered // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeContractNewChainRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractNewChainRegistered)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeContractNewChainRegistered)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeContractNewChainRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractNewChainRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractNewChainRegistered represents a NewChainRegistered event raised by the BridgeContract contract.
type BridgeContractNewChainRegistered struct {
	ChainId common.Hash
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNewChainRegistered is a free log retrieval operation binding the contract event 0x3cbe969d5c5f2c70c7cfb293cd355d3fcc80a852eac9bee1c1c317dc302f199d.
//
// Solidity: event newChainRegistered(string indexed chainId)
func (_BridgeContract *BridgeContractFilterer) FilterNewChainRegistered(opts *bind.FilterOpts, chainId []string) (*BridgeContractNewChainRegisteredIterator, error) {

	var chainIdRule []interface{}
	for _, chainIdItem := range chainId {
		chainIdRule = append(chainIdRule, chainIdItem)
	}

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "newChainRegistered", chainIdRule)
	if err != nil {
		return nil, err
	}
	return &BridgeContractNewChainRegisteredIterator{contract: _BridgeContract.contract, event: "newChainRegistered", logs: logs, sub: sub}, nil
}

// WatchNewChainRegistered is a free log subscription operation binding the contract event 0x3cbe969d5c5f2c70c7cfb293cd355d3fcc80a852eac9bee1c1c317dc302f199d.
//
// Solidity: event newChainRegistered(string indexed chainId)
func (_BridgeContract *BridgeContractFilterer) WatchNewChainRegistered(opts *bind.WatchOpts, sink chan<- *BridgeContractNewChainRegistered, chainId []string) (event.Subscription, error) {

	var chainIdRule []interface{}
	for _, chainIdItem := range chainId {
		chainIdRule = append(chainIdRule, chainIdItem)
	}

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "newChainRegistered", chainIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractNewChainRegistered)
				if err := _BridgeContract.contract.UnpackLog(event, "newChainRegistered", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNewChainRegistered is a log parse operation binding the contract event 0x3cbe969d5c5f2c70c7cfb293cd355d3fcc80a852eac9bee1c1c317dc302f199d.
//
// Solidity: event newChainRegistered(string indexed chainId)
func (_BridgeContract *BridgeContractFilterer) ParseNewChainRegistered(log types.Log) (*BridgeContractNewChainRegistered, error) {
	event := new(BridgeContractNewChainRegistered)
	if err := _BridgeContract.contract.UnpackLog(event, "newChainRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
