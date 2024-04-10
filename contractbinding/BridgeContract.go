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
	Nonce       *big.Int
	BlockHeight *big.Int
	Receivers   []IBridgeContractStructsReceiver
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
	IncludedTransactions      []*big.Int
	UsedUTXOs                 IBridgeContractStructsUTXOs
}

// IBridgeContractStructsUTXO is an auto generated low-level Go binding around an user-defined struct.
type IBridgeContractStructsUTXO struct {
	Nonce   uint64
	TxHash  string
	TxIndex *big.Int
	Amount  *big.Int
}

// IBridgeContractStructsUTXOs is an auto generated low-level Go binding around an user-defined struct.
type IBridgeContractStructsUTXOs struct {
	MultisigOwnedUTXOs []IBridgeContractStructsUTXO
	FeePayerOwnedUTXOs []IBridgeContractStructsUTXO
}

// IBridgeContractStructsValidatorAddressCardanoData is an auto generated low-level Go binding around an user-defined struct.
type IBridgeContractStructsValidatorAddressCardanoData struct {
	Addr common.Address
	Data IBridgeContractStructsValidatorCardanoData
}

// IBridgeContractStructsValidatorCardanoData is an auto generated low-level Go binding around an user-defined struct.
type IBridgeContractStructsValidatorCardanoData struct {
	VerifyingKey    string
	VerifyingKeyFee string
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
	ABI: "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_claimTransactionHash\",\"type\":\"string\"}],\"name\":\"AlreadyConfirmed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_claimTransactionHash\",\"type\":\"string\"}],\"name\":\"AlreadyProposed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_blockchainID\",\"type\":\"string\"}],\"name\":\"CanNotCreateBatchYet\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_claimId\",\"type\":\"string\"}],\"name\":\"ChainAlreadyRegistered\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_chainId\",\"type\":\"string\"}],\"name\":\"ChainIsNotRegistered\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"data\",\"type\":\"string\"}],\"name\":\"InvalidData\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSignature\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_slot\",\"type\":\"uint256\"}],\"name\":\"InvalidSlot\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotBridgeContract\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotClaimsHelper\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotClaimsManager\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotClaimsManagerOrBridgeContract\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_claimTransactionHash\",\"type\":\"string\"}],\"name\":\"NotEnoughBridgingTokensAwailable\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSignedBatchManagerOrBridgeContract\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotValidator\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"RefundRequestClaimNotYetSupporter\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_chainId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"WrongBatchNonce\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"string\",\"name\":\"chainId\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"newChainProposal\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"string\",\"name\":\"chainId\",\"type\":\"string\"}],\"name\":\"newChainRegistered\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"getAllRegisteredChains\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"id\",\"type\":\"string\"},{\"components\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"multisigOwnedUTXOs\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"feePayerOwnedUTXOs\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeContractStructs.UTXOs\",\"name\":\"utxos\",\"type\":\"tuple\"},{\"internalType\":\"string\",\"name\":\"addressMultisig\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"addressFeePayer\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"tokenQuantity\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.Chain[]\",\"name\":\"_chains\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_destinationChain\",\"type\":\"string\"}],\"name\":\"getAvailableUTXOs\",\"outputs\":[{\"components\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"multisigOwnedUTXOs\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"feePayerOwnedUTXOs\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeContractStructs.UTXOs\",\"name\":\"availableUTXOs\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_destinationChain\",\"type\":\"string\"}],\"name\":\"getConfirmedBatch\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"rawTransaction\",\"type\":\"string\"},{\"internalType\":\"string[]\",\"name\":\"multisigSignatures\",\"type\":\"string[]\"},{\"internalType\":\"string[]\",\"name\":\"feePayerMultisigSignatures\",\"type\":\"string[]\"}],\"internalType\":\"structIBridgeContractStructs.ConfirmedBatch\",\"name\":\"batch\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_destinationChain\",\"type\":\"string\"}],\"name\":\"getConfirmedTransactions\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"blockHeight\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"destinationAddress\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.Receiver[]\",\"name\":\"receivers\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeContractStructs.ConfirmedTransaction[]\",\"name\":\"confirmedTransactions\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_sourceChain\",\"type\":\"string\"}],\"name\":\"getLastObservedBlock\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"blockHash\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"blockSlot\",\"type\":\"uint64\"}],\"internalType\":\"structIBridgeContractStructs.CardanoBlock\",\"name\":\"cblock\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_destinationChain\",\"type\":\"string\"}],\"name\":\"getNextBatchId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_destinationChain\",\"type\":\"string\"}],\"name\":\"getRawTransactionFromLastBatch\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_chainId\",\"type\":\"string\"}],\"name\":\"getValidatorsCardanoData\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"verifyingKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"verifyingKeyFee\",\"type\":\"string\"}],\"internalType\":\"structIBridgeContractStructs.ValidatorCardanoData[]\",\"name\":\"validators\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_chainId\",\"type\":\"string\"},{\"components\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"multisigOwnedUTXOs\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"feePayerOwnedUTXOs\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeContractStructs.UTXOs\",\"name\":\"_initialUTXOs\",\"type\":\"tuple\"},{\"internalType\":\"string\",\"name\":\"_addressMultisig\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_addressFeePayer\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"verifyingKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"verifyingKeyFee\",\"type\":\"string\"}],\"internalType\":\"structIBridgeContractStructs.ValidatorCardanoData\",\"name\":\"data\",\"type\":\"tuple\"}],\"internalType\":\"structIBridgeContractStructs.ValidatorAddressCardanoData[]\",\"name\":\"_validatorData\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256\",\"name\":\"_tokenQuantity\",\"type\":\"uint256\"}],\"name\":\"registerChain\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_chainId\",\"type\":\"string\"},{\"components\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"multisigOwnedUTXOs\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"feePayerOwnedUTXOs\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeContractStructs.UTXOs\",\"name\":\"_initialUTXOs\",\"type\":\"tuple\"},{\"internalType\":\"string\",\"name\":\"_addressMultisig\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_addressFeePayer\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"verifyingKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"verifyingKeyFee\",\"type\":\"string\"}],\"internalType\":\"structIBridgeContractStructs.ValidatorCardanoData\",\"name\":\"_validatorData\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"_tokenQuantity\",\"type\":\"uint256\"}],\"name\":\"registerChainGovernance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_destinationChain\",\"type\":\"string\"}],\"name\":\"shouldCreateBatch\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"observedTransactionHash\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"destinationAddress\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.Receiver[]\",\"name\":\"receivers\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO\",\"name\":\"outputUTXO\",\"type\":\"tuple\"},{\"internalType\":\"string\",\"name\":\"sourceChainID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"destinationChainID\",\"type\":\"string\"}],\"internalType\":\"structIBridgeContractStructs.BridgingRequestClaim[]\",\"name\":\"bridgingRequestClaims\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"observedTransactionHash\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"chainID\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"batchNonceID\",\"type\":\"uint256\"},{\"components\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"multisigOwnedUTXOs\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"feePayerOwnedUTXOs\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeContractStructs.UTXOs\",\"name\":\"outputUTXOs\",\"type\":\"tuple\"}],\"internalType\":\"structIBridgeContractStructs.BatchExecutedClaim[]\",\"name\":\"batchExecutedClaims\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"observedTransactionHash\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"chainID\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"batchNonceID\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.BatchExecutionFailedClaim[]\",\"name\":\"batchExecutionFailedClaims\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"observedTransactionHash\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"previousRefundTxHash\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"chainID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO\",\"name\":\"utxo\",\"type\":\"tuple\"},{\"internalType\":\"string\",\"name\":\"rawTransaction\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"multisigSignature\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"retryCounter\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.RefundRequestClaim[]\",\"name\":\"refundRequestClaims\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"observedTransactionHash\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"chainID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"refundTxHash\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO\",\"name\":\"utxo\",\"type\":\"tuple\"}],\"internalType\":\"structIBridgeContractStructs.RefundExecutedClaim[]\",\"name\":\"refundExecutedClaims\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeContractStructs.ValidatorClaims\",\"name\":\"_claims\",\"type\":\"tuple\"}],\"name\":\"submitClaims\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"chainID\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"blockHash\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"blockSlot\",\"type\":\"uint64\"}],\"internalType\":\"structIBridgeContractStructs.CardanoBlock[]\",\"name\":\"blocks\",\"type\":\"tuple[]\"}],\"name\":\"submitLastObservableBlocks\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"destinationChainId\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"rawTransaction\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"multisigSignature\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"feePayerMultisigSignature\",\"type\":\"string\"},{\"internalType\":\"uint256[]\",\"name\":\"includedTransactions\",\"type\":\"uint256[]\"},{\"components\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"multisigOwnedUTXOs\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeContractStructs.UTXO[]\",\"name\":\"feePayerOwnedUTXOs\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeContractStructs.UTXOs\",\"name\":\"usedUTXOs\",\"type\":\"tuple\"}],\"internalType\":\"structIBridgeContractStructs.SignedBatch\",\"name\":\"_signedBatch\",\"type\":\"tuple\"}],\"name\":\"submitSignedBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
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

// GetAllRegisteredChains is a free data retrieval call binding the contract method 0x67f0cc44.
//
// Solidity: function getAllRegisteredChains() view returns((string,((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]),string,string,uint256)[] _chains)
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
// Solidity: function getAllRegisteredChains() view returns((string,((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]),string,string,uint256)[] _chains)
func (_BridgeContract *BridgeContractSession) GetAllRegisteredChains() ([]IBridgeContractStructsChain, error) {
	return _BridgeContract.Contract.GetAllRegisteredChains(&_BridgeContract.CallOpts)
}

// GetAllRegisteredChains is a free data retrieval call binding the contract method 0x67f0cc44.
//
// Solidity: function getAllRegisteredChains() view returns((string,((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]),string,string,uint256)[] _chains)
func (_BridgeContract *BridgeContractCallerSession) GetAllRegisteredChains() ([]IBridgeContractStructsChain, error) {
	return _BridgeContract.Contract.GetAllRegisteredChains(&_BridgeContract.CallOpts)
}

// GetAvailableUTXOs is a free data retrieval call binding the contract method 0x03fe69ae.
//
// Solidity: function getAvailableUTXOs(string _destinationChain) view returns(((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]) availableUTXOs)
func (_BridgeContract *BridgeContractCaller) GetAvailableUTXOs(opts *bind.CallOpts, _destinationChain string) (IBridgeContractStructsUTXOs, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getAvailableUTXOs", _destinationChain)

	if err != nil {
		return *new(IBridgeContractStructsUTXOs), err
	}

	out0 := *abi.ConvertType(out[0], new(IBridgeContractStructsUTXOs)).(*IBridgeContractStructsUTXOs)

	return out0, err

}

// GetAvailableUTXOs is a free data retrieval call binding the contract method 0x03fe69ae.
//
// Solidity: function getAvailableUTXOs(string _destinationChain) view returns(((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]) availableUTXOs)
func (_BridgeContract *BridgeContractSession) GetAvailableUTXOs(_destinationChain string) (IBridgeContractStructsUTXOs, error) {
	return _BridgeContract.Contract.GetAvailableUTXOs(&_BridgeContract.CallOpts, _destinationChain)
}

// GetAvailableUTXOs is a free data retrieval call binding the contract method 0x03fe69ae.
//
// Solidity: function getAvailableUTXOs(string _destinationChain) view returns(((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]) availableUTXOs)
func (_BridgeContract *BridgeContractCallerSession) GetAvailableUTXOs(_destinationChain string) (IBridgeContractStructsUTXOs, error) {
	return _BridgeContract.Contract.GetAvailableUTXOs(&_BridgeContract.CallOpts, _destinationChain)
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

// GetConfirmedTransactions is a free data retrieval call binding the contract method 0x595051f9.
//
// Solidity: function getConfirmedTransactions(string _destinationChain) view returns((uint256,uint256,(string,uint256)[])[] confirmedTransactions)
func (_BridgeContract *BridgeContractCaller) GetConfirmedTransactions(opts *bind.CallOpts, _destinationChain string) ([]IBridgeContractStructsConfirmedTransaction, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getConfirmedTransactions", _destinationChain)

	if err != nil {
		return *new([]IBridgeContractStructsConfirmedTransaction), err
	}

	out0 := *abi.ConvertType(out[0], new([]IBridgeContractStructsConfirmedTransaction)).(*[]IBridgeContractStructsConfirmedTransaction)

	return out0, err

}

// GetConfirmedTransactions is a free data retrieval call binding the contract method 0x595051f9.
//
// Solidity: function getConfirmedTransactions(string _destinationChain) view returns((uint256,uint256,(string,uint256)[])[] confirmedTransactions)
func (_BridgeContract *BridgeContractSession) GetConfirmedTransactions(_destinationChain string) ([]IBridgeContractStructsConfirmedTransaction, error) {
	return _BridgeContract.Contract.GetConfirmedTransactions(&_BridgeContract.CallOpts, _destinationChain)
}

// GetConfirmedTransactions is a free data retrieval call binding the contract method 0x595051f9.
//
// Solidity: function getConfirmedTransactions(string _destinationChain) view returns((uint256,uint256,(string,uint256)[])[] confirmedTransactions)
func (_BridgeContract *BridgeContractCallerSession) GetConfirmedTransactions(_destinationChain string) ([]IBridgeContractStructsConfirmedTransaction, error) {
	return _BridgeContract.Contract.GetConfirmedTransactions(&_BridgeContract.CallOpts, _destinationChain)
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

// GetNextBatchId is a free data retrieval call binding the contract method 0x3cd9ae3e.
//
// Solidity: function getNextBatchId(string _destinationChain) view returns(uint256)
func (_BridgeContract *BridgeContractCaller) GetNextBatchId(opts *bind.CallOpts, _destinationChain string) (*big.Int, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getNextBatchId", _destinationChain)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNextBatchId is a free data retrieval call binding the contract method 0x3cd9ae3e.
//
// Solidity: function getNextBatchId(string _destinationChain) view returns(uint256)
func (_BridgeContract *BridgeContractSession) GetNextBatchId(_destinationChain string) (*big.Int, error) {
	return _BridgeContract.Contract.GetNextBatchId(&_BridgeContract.CallOpts, _destinationChain)
}

// GetNextBatchId is a free data retrieval call binding the contract method 0x3cd9ae3e.
//
// Solidity: function getNextBatchId(string _destinationChain) view returns(uint256)
func (_BridgeContract *BridgeContractCallerSession) GetNextBatchId(_destinationChain string) (*big.Int, error) {
	return _BridgeContract.Contract.GetNextBatchId(&_BridgeContract.CallOpts, _destinationChain)
}

// GetRawTransactionFromLastBatch is a free data retrieval call binding the contract method 0x49187cd9.
//
// Solidity: function getRawTransactionFromLastBatch(string _destinationChain) view returns(string)
func (_BridgeContract *BridgeContractCaller) GetRawTransactionFromLastBatch(opts *bind.CallOpts, _destinationChain string) (string, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getRawTransactionFromLastBatch", _destinationChain)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetRawTransactionFromLastBatch is a free data retrieval call binding the contract method 0x49187cd9.
//
// Solidity: function getRawTransactionFromLastBatch(string _destinationChain) view returns(string)
func (_BridgeContract *BridgeContractSession) GetRawTransactionFromLastBatch(_destinationChain string) (string, error) {
	return _BridgeContract.Contract.GetRawTransactionFromLastBatch(&_BridgeContract.CallOpts, _destinationChain)
}

// GetRawTransactionFromLastBatch is a free data retrieval call binding the contract method 0x49187cd9.
//
// Solidity: function getRawTransactionFromLastBatch(string _destinationChain) view returns(string)
func (_BridgeContract *BridgeContractCallerSession) GetRawTransactionFromLastBatch(_destinationChain string) (string, error) {
	return _BridgeContract.Contract.GetRawTransactionFromLastBatch(&_BridgeContract.CallOpts, _destinationChain)
}

// GetValidatorsCardanoData is a free data retrieval call binding the contract method 0x636b8a0d.
//
// Solidity: function getValidatorsCardanoData(string _chainId) view returns((string,string)[] validators)
func (_BridgeContract *BridgeContractCaller) GetValidatorsCardanoData(opts *bind.CallOpts, _chainId string) ([]IBridgeContractStructsValidatorCardanoData, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getValidatorsCardanoData", _chainId)

	if err != nil {
		return *new([]IBridgeContractStructsValidatorCardanoData), err
	}

	out0 := *abi.ConvertType(out[0], new([]IBridgeContractStructsValidatorCardanoData)).(*[]IBridgeContractStructsValidatorCardanoData)

	return out0, err

}

// GetValidatorsCardanoData is a free data retrieval call binding the contract method 0x636b8a0d.
//
// Solidity: function getValidatorsCardanoData(string _chainId) view returns((string,string)[] validators)
func (_BridgeContract *BridgeContractSession) GetValidatorsCardanoData(_chainId string) ([]IBridgeContractStructsValidatorCardanoData, error) {
	return _BridgeContract.Contract.GetValidatorsCardanoData(&_BridgeContract.CallOpts, _chainId)
}

// GetValidatorsCardanoData is a free data retrieval call binding the contract method 0x636b8a0d.
//
// Solidity: function getValidatorsCardanoData(string _chainId) view returns((string,string)[] validators)
func (_BridgeContract *BridgeContractCallerSession) GetValidatorsCardanoData(_chainId string) ([]IBridgeContractStructsValidatorCardanoData, error) {
	return _BridgeContract.Contract.GetValidatorsCardanoData(&_BridgeContract.CallOpts, _chainId)
}

// ShouldCreateBatch is a free data retrieval call binding the contract method 0x77968b34.
//
// Solidity: function shouldCreateBatch(string _destinationChain) view returns(bool)
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
// Solidity: function shouldCreateBatch(string _destinationChain) view returns(bool)
func (_BridgeContract *BridgeContractSession) ShouldCreateBatch(_destinationChain string) (bool, error) {
	return _BridgeContract.Contract.ShouldCreateBatch(&_BridgeContract.CallOpts, _destinationChain)
}

// ShouldCreateBatch is a free data retrieval call binding the contract method 0x77968b34.
//
// Solidity: function shouldCreateBatch(string _destinationChain) view returns(bool)
func (_BridgeContract *BridgeContractCallerSession) ShouldCreateBatch(_destinationChain string) (bool, error) {
	return _BridgeContract.Contract.ShouldCreateBatch(&_BridgeContract.CallOpts, _destinationChain)
}

// RegisterChain is a paid mutator transaction binding the contract method 0xf75b143f.
//
// Solidity: function registerChain(string _chainId, ((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]) _initialUTXOs, string _addressMultisig, string _addressFeePayer, (address,(string,string))[] _validatorData, uint256 _tokenQuantity) returns()
func (_BridgeContract *BridgeContractTransactor) RegisterChain(opts *bind.TransactOpts, _chainId string, _initialUTXOs IBridgeContractStructsUTXOs, _addressMultisig string, _addressFeePayer string, _validatorData []IBridgeContractStructsValidatorAddressCardanoData, _tokenQuantity *big.Int) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "registerChain", _chainId, _initialUTXOs, _addressMultisig, _addressFeePayer, _validatorData, _tokenQuantity)
}

// RegisterChain is a paid mutator transaction binding the contract method 0xf75b143f.
//
// Solidity: function registerChain(string _chainId, ((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]) _initialUTXOs, string _addressMultisig, string _addressFeePayer, (address,(string,string))[] _validatorData, uint256 _tokenQuantity) returns()
func (_BridgeContract *BridgeContractSession) RegisterChain(_chainId string, _initialUTXOs IBridgeContractStructsUTXOs, _addressMultisig string, _addressFeePayer string, _validatorData []IBridgeContractStructsValidatorAddressCardanoData, _tokenQuantity *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.RegisterChain(&_BridgeContract.TransactOpts, _chainId, _initialUTXOs, _addressMultisig, _addressFeePayer, _validatorData, _tokenQuantity)
}

// RegisterChain is a paid mutator transaction binding the contract method 0xf75b143f.
//
// Solidity: function registerChain(string _chainId, ((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]) _initialUTXOs, string _addressMultisig, string _addressFeePayer, (address,(string,string))[] _validatorData, uint256 _tokenQuantity) returns()
func (_BridgeContract *BridgeContractTransactorSession) RegisterChain(_chainId string, _initialUTXOs IBridgeContractStructsUTXOs, _addressMultisig string, _addressFeePayer string, _validatorData []IBridgeContractStructsValidatorAddressCardanoData, _tokenQuantity *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.RegisterChain(&_BridgeContract.TransactOpts, _chainId, _initialUTXOs, _addressMultisig, _addressFeePayer, _validatorData, _tokenQuantity)
}

// RegisterChainGovernance is a paid mutator transaction binding the contract method 0xbf522790.
//
// Solidity: function registerChainGovernance(string _chainId, ((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]) _initialUTXOs, string _addressMultisig, string _addressFeePayer, (string,string) _validatorData, uint256 _tokenQuantity) returns()
func (_BridgeContract *BridgeContractTransactor) RegisterChainGovernance(opts *bind.TransactOpts, _chainId string, _initialUTXOs IBridgeContractStructsUTXOs, _addressMultisig string, _addressFeePayer string, _validatorData IBridgeContractStructsValidatorCardanoData, _tokenQuantity *big.Int) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "registerChainGovernance", _chainId, _initialUTXOs, _addressMultisig, _addressFeePayer, _validatorData, _tokenQuantity)
}

// RegisterChainGovernance is a paid mutator transaction binding the contract method 0xbf522790.
//
// Solidity: function registerChainGovernance(string _chainId, ((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]) _initialUTXOs, string _addressMultisig, string _addressFeePayer, (string,string) _validatorData, uint256 _tokenQuantity) returns()
func (_BridgeContract *BridgeContractSession) RegisterChainGovernance(_chainId string, _initialUTXOs IBridgeContractStructsUTXOs, _addressMultisig string, _addressFeePayer string, _validatorData IBridgeContractStructsValidatorCardanoData, _tokenQuantity *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.RegisterChainGovernance(&_BridgeContract.TransactOpts, _chainId, _initialUTXOs, _addressMultisig, _addressFeePayer, _validatorData, _tokenQuantity)
}

// RegisterChainGovernance is a paid mutator transaction binding the contract method 0xbf522790.
//
// Solidity: function registerChainGovernance(string _chainId, ((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]) _initialUTXOs, string _addressMultisig, string _addressFeePayer, (string,string) _validatorData, uint256 _tokenQuantity) returns()
func (_BridgeContract *BridgeContractTransactorSession) RegisterChainGovernance(_chainId string, _initialUTXOs IBridgeContractStructsUTXOs, _addressMultisig string, _addressFeePayer string, _validatorData IBridgeContractStructsValidatorCardanoData, _tokenQuantity *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.RegisterChainGovernance(&_BridgeContract.TransactOpts, _chainId, _initialUTXOs, _addressMultisig, _addressFeePayer, _validatorData, _tokenQuantity)
}

// SubmitClaims is a paid mutator transaction binding the contract method 0xb95a432c.
//
// Solidity: function submitClaims(((string,(string,uint256)[],(uint64,string,uint256,uint256),string,string)[],(string,string,uint256,((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]))[],(string,string,uint256)[],(string,string,string,string,(uint64,string,uint256,uint256),string,string,uint256)[],(string,string,string,(uint64,string,uint256,uint256))[]) _claims) returns()
func (_BridgeContract *BridgeContractTransactor) SubmitClaims(opts *bind.TransactOpts, _claims IBridgeContractStructsValidatorClaims) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "submitClaims", _claims)
}

// SubmitClaims is a paid mutator transaction binding the contract method 0xb95a432c.
//
// Solidity: function submitClaims(((string,(string,uint256)[],(uint64,string,uint256,uint256),string,string)[],(string,string,uint256,((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]))[],(string,string,uint256)[],(string,string,string,string,(uint64,string,uint256,uint256),string,string,uint256)[],(string,string,string,(uint64,string,uint256,uint256))[]) _claims) returns()
func (_BridgeContract *BridgeContractSession) SubmitClaims(_claims IBridgeContractStructsValidatorClaims) (*types.Transaction, error) {
	return _BridgeContract.Contract.SubmitClaims(&_BridgeContract.TransactOpts, _claims)
}

// SubmitClaims is a paid mutator transaction binding the contract method 0xb95a432c.
//
// Solidity: function submitClaims(((string,(string,uint256)[],(uint64,string,uint256,uint256),string,string)[],(string,string,uint256,((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]))[],(string,string,uint256)[],(string,string,string,string,(uint64,string,uint256,uint256),string,string,uint256)[],(string,string,string,(uint64,string,uint256,uint256))[]) _claims) returns()
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

// SubmitSignedBatch is a paid mutator transaction binding the contract method 0xf39b5f49.
//
// Solidity: function submitSignedBatch((uint256,string,string,string,string,uint256[],((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[])) _signedBatch) returns()
func (_BridgeContract *BridgeContractTransactor) SubmitSignedBatch(opts *bind.TransactOpts, _signedBatch IBridgeContractStructsSignedBatch) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "submitSignedBatch", _signedBatch)
}

// SubmitSignedBatch is a paid mutator transaction binding the contract method 0xf39b5f49.
//
// Solidity: function submitSignedBatch((uint256,string,string,string,string,uint256[],((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[])) _signedBatch) returns()
func (_BridgeContract *BridgeContractSession) SubmitSignedBatch(_signedBatch IBridgeContractStructsSignedBatch) (*types.Transaction, error) {
	return _BridgeContract.Contract.SubmitSignedBatch(&_BridgeContract.TransactOpts, _signedBatch)
}

// SubmitSignedBatch is a paid mutator transaction binding the contract method 0xf39b5f49.
//
// Solidity: function submitSignedBatch((uint256,string,string,string,string,uint256[],((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[])) _signedBatch) returns()
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