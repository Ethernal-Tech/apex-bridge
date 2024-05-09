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

// IBridgeStructsBatchExecutedClaim is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsBatchExecutedClaim struct {
	ObservedTransactionHash string
	ChainID                 string
	BatchNonceID            *big.Int
	OutputUTXOs             IBridgeStructsUTXOs
}

// IBridgeStructsBatchExecutionFailedClaim is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsBatchExecutionFailedClaim struct {
	ObservedTransactionHash string
	ChainID                 string
	BatchNonceID            *big.Int
}

// IBridgeStructsBridgingRequestClaim is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsBridgingRequestClaim struct {
	ObservedTransactionHash string
	Receivers               []IBridgeStructsReceiver
	OutputUTXO              IBridgeStructsUTXO
	SourceChainID           string
	DestinationChainID      string
}

// IBridgeStructsCardanoBlock is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsCardanoBlock struct {
	BlockHash string
	BlockSlot uint64
}

// IBridgeStructsChain is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsChain struct {
	Id              string
	Utxos           IBridgeStructsUTXOs
	AddressMultisig string
	AddressFeePayer string
}

// IBridgeStructsConfirmedBatch is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsConfirmedBatch struct {
	Id                         *big.Int
	RawTransaction             string
	MultisigSignatures         []string
	FeePayerMultisigSignatures []string
}

// IBridgeStructsConfirmedTransaction is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsConfirmedTransaction struct {
	ObservedTransactionHash string
	Nonce                   *big.Int
	BlockHeight             *big.Int
	SourceChainID           string
	Receivers               []IBridgeStructsReceiver
}

// IBridgeStructsReceiver is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsReceiver struct {
	DestinationAddress string
	Amount             *big.Int
}

// IBridgeStructsRefundExecutedClaim is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsRefundExecutedClaim struct {
	ObservedTransactionHash string
	ChainID                 string
	RefundTxHash            string
	Utxo                    IBridgeStructsUTXO
}

// IBridgeStructsRefundRequestClaim is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsRefundRequestClaim struct {
	ObservedTransactionHash string
	PreviousRefundTxHash    string
	ChainID                 string
	Receiver                string
	Utxo                    IBridgeStructsUTXO
	RawTransaction          string
	MultisigSignature       string
	RetryCounter            *big.Int
}

// IBridgeStructsSignedBatch is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsSignedBatch struct {
	Id                        *big.Int
	DestinationChainId        string
	RawTransaction            string
	MultisigSignature         string
	FeePayerMultisigSignature string
	IncludedTransactions      []*big.Int
	UsedUTXOs                 IBridgeStructsUTXOs
}

// IBridgeStructsUTXO is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsUTXO struct {
	Nonce   uint64
	TxHash  string
	TxIndex *big.Int
	Amount  *big.Int
}

// IBridgeStructsUTXOs is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsUTXOs struct {
	MultisigOwnedUTXOs []IBridgeStructsUTXO
	FeePayerOwnedUTXOs []IBridgeStructsUTXO
}

// IBridgeStructsValidatorAddressCardanoData is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsValidatorAddressCardanoData struct {
	Addr common.Address
	Data IBridgeStructsValidatorCardanoData
}

// IBridgeStructsValidatorCardanoData is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsValidatorCardanoData struct {
	VerifyingKey    string
	VerifyingKeyFee string
}

// IBridgeStructsValidatorClaims is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsValidatorClaims struct {
	BridgingRequestClaims      []IBridgeStructsBridgingRequestClaim
	BatchExecutedClaims        []IBridgeStructsBatchExecutedClaim
	BatchExecutionFailedClaims []IBridgeStructsBatchExecutionFailedClaim
	RefundRequestClaims        []IBridgeStructsRefundRequestClaim
	RefundExecutedClaims       []IBridgeStructsRefundExecutedClaim
}

// BridgeContractMetaData contains all meta data concerning the BridgeContract contract.
var BridgeContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"AddressEmptyCode\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_claimTransactionHash\",\"type\":\"string\"}],\"name\":\"AlreadyConfirmed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_claimTransactionHash\",\"type\":\"string\"}],\"name\":\"AlreadyProposed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_blockchainID\",\"type\":\"string\"}],\"name\":\"CanNotCreateBatchYet\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_claimId\",\"type\":\"string\"}],\"name\":\"ChainAlreadyRegistered\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_chainId\",\"type\":\"string\"}],\"name\":\"ChainIsNotRegistered\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"ERC1967InvalidImplementation\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ERC1967NonPayable\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FailedInnerCall\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"data\",\"type\":\"string\"}],\"name\":\"InvalidData\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSignature\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotBridge\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotClaims\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_claimTransactionHash\",\"type\":\"string\"}],\"name\":\"NotEnoughBridgingTokensAwailable\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSignedBatches\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSignedBatchesOrBridge\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotValidator\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"UUPSUnauthorizedCallContext\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"slot\",\"type\":\"bytes32\"}],\"name\":\"UUPSUnsupportedProxiableUUID\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_chainId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"WrongBatchNonce\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"Upgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"string\",\"name\":\"chainId\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"newChainProposal\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"string\",\"name\":\"chainId\",\"type\":\"string\"}],\"name\":\"newChainRegistered\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"UPGRADE_INTERFACE_VERSION\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAllRegisteredChains\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"id\",\"type\":\"string\"},{\"components\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeStructs.UTXO[]\",\"name\":\"multisigOwnedUTXOs\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeStructs.UTXO[]\",\"name\":\"feePayerOwnedUTXOs\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeStructs.UTXOs\",\"name\":\"utxos\",\"type\":\"tuple\"},{\"internalType\":\"string\",\"name\":\"addressMultisig\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"addressFeePayer\",\"type\":\"string\"}],\"internalType\":\"structIBridgeStructs.Chain[]\",\"name\":\"_chains\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_destinationChain\",\"type\":\"string\"}],\"name\":\"getAvailableUTXOs\",\"outputs\":[{\"components\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeStructs.UTXO[]\",\"name\":\"multisigOwnedUTXOs\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeStructs.UTXO[]\",\"name\":\"feePayerOwnedUTXOs\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeStructs.UTXOs\",\"name\":\"availableUTXOs\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_destinationChain\",\"type\":\"string\"}],\"name\":\"getConfirmedBatch\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"rawTransaction\",\"type\":\"string\"},{\"internalType\":\"string[]\",\"name\":\"multisigSignatures\",\"type\":\"string[]\"},{\"internalType\":\"string[]\",\"name\":\"feePayerMultisigSignatures\",\"type\":\"string[]\"}],\"internalType\":\"structIBridgeStructs.ConfirmedBatch\",\"name\":\"batch\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_destinationChain\",\"type\":\"string\"}],\"name\":\"getConfirmedTransactions\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"observedTransactionHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"blockHeight\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"sourceChainID\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"destinationAddress\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeStructs.Receiver[]\",\"name\":\"receivers\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeStructs.ConfirmedTransaction[]\",\"name\":\"_confirmedTransactions\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_sourceChain\",\"type\":\"string\"}],\"name\":\"getLastObservedBlock\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"blockHash\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"blockSlot\",\"type\":\"uint64\"}],\"internalType\":\"structIBridgeStructs.CardanoBlock\",\"name\":\"cblock\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_destinationChain\",\"type\":\"string\"}],\"name\":\"getNextBatchId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"result\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_destinationChain\",\"type\":\"string\"}],\"name\":\"getRawTransactionFromLastBatch\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_chainId\",\"type\":\"string\"}],\"name\":\"getValidatorsCardanoData\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"verifyingKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"verifyingKeyFee\",\"type\":\"string\"}],\"internalType\":\"structIBridgeStructs.ValidatorCardanoData[]\",\"name\":\"validatorCardanoData\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"proxiableUUID\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_chainId\",\"type\":\"string\"},{\"components\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeStructs.UTXO[]\",\"name\":\"multisigOwnedUTXOs\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeStructs.UTXO[]\",\"name\":\"feePayerOwnedUTXOs\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeStructs.UTXOs\",\"name\":\"_initialUTXOs\",\"type\":\"tuple\"},{\"internalType\":\"string\",\"name\":\"_addressMultisig\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_addressFeePayer\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"verifyingKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"verifyingKeyFee\",\"type\":\"string\"}],\"internalType\":\"structIBridgeStructs.ValidatorCardanoData\",\"name\":\"data\",\"type\":\"tuple\"}],\"internalType\":\"structIBridgeStructs.ValidatorAddressCardanoData[]\",\"name\":\"_validatorsAddressCardanoData\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256\",\"name\":\"_tokenQuantity\",\"type\":\"uint256\"}],\"name\":\"registerChain\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_chainId\",\"type\":\"string\"},{\"components\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeStructs.UTXO[]\",\"name\":\"multisigOwnedUTXOs\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeStructs.UTXO[]\",\"name\":\"feePayerOwnedUTXOs\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeStructs.UTXOs\",\"name\":\"_initialUTXOs\",\"type\":\"tuple\"},{\"internalType\":\"string\",\"name\":\"_addressMultisig\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_addressFeePayer\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"verifyingKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"verifyingKeyFee\",\"type\":\"string\"}],\"internalType\":\"structIBridgeStructs.ValidatorCardanoData\",\"name\":\"_validatorCardanoData\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"_tokenQuantity\",\"type\":\"uint256\"}],\"name\":\"registerChainGovernance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_claimsAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_signedBatchesAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_slotsAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_utxoscAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_validatorsAddress\",\"type\":\"address\"}],\"name\":\"setDependencies\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_destinationChain\",\"type\":\"string\"}],\"name\":\"shouldCreateBatch\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"batch\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"observedTransactionHash\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"destinationAddress\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeStructs.Receiver[]\",\"name\":\"receivers\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeStructs.UTXO\",\"name\":\"outputUTXO\",\"type\":\"tuple\"},{\"internalType\":\"string\",\"name\":\"sourceChainID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"destinationChainID\",\"type\":\"string\"}],\"internalType\":\"structIBridgeStructs.BridgingRequestClaim[]\",\"name\":\"bridgingRequestClaims\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"observedTransactionHash\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"chainID\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"batchNonceID\",\"type\":\"uint256\"},{\"components\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeStructs.UTXO[]\",\"name\":\"multisigOwnedUTXOs\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeStructs.UTXO[]\",\"name\":\"feePayerOwnedUTXOs\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeStructs.UTXOs\",\"name\":\"outputUTXOs\",\"type\":\"tuple\"}],\"internalType\":\"structIBridgeStructs.BatchExecutedClaim[]\",\"name\":\"batchExecutedClaims\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"observedTransactionHash\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"chainID\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"batchNonceID\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeStructs.BatchExecutionFailedClaim[]\",\"name\":\"batchExecutionFailedClaims\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"observedTransactionHash\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"previousRefundTxHash\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"chainID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeStructs.UTXO\",\"name\":\"utxo\",\"type\":\"tuple\"},{\"internalType\":\"string\",\"name\":\"rawTransaction\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"multisigSignature\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"retryCounter\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeStructs.RefundRequestClaim[]\",\"name\":\"refundRequestClaims\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"observedTransactionHash\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"chainID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"refundTxHash\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeStructs.UTXO\",\"name\":\"utxo\",\"type\":\"tuple\"}],\"internalType\":\"structIBridgeStructs.RefundExecutedClaim[]\",\"name\":\"refundExecutedClaims\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeStructs.ValidatorClaims\",\"name\":\"_claims\",\"type\":\"tuple\"}],\"name\":\"submitClaims\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"chainID\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"blockHash\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"blockSlot\",\"type\":\"uint64\"}],\"internalType\":\"structIBridgeStructs.CardanoBlock[]\",\"name\":\"blocks\",\"type\":\"tuple[]\"}],\"name\":\"submitLastObservedBlocks\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"destinationChainId\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"rawTransaction\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"multisigSignature\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"feePayerMultisigSignature\",\"type\":\"string\"},{\"internalType\":\"uint256[]\",\"name\":\"includedTransactions\",\"type\":\"uint256[]\"},{\"components\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeStructs.UTXO[]\",\"name\":\"multisigOwnedUTXOs\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"txHash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"txIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeStructs.UTXO[]\",\"name\":\"feePayerOwnedUTXOs\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeStructs.UTXOs\",\"name\":\"usedUTXOs\",\"type\":\"tuple\"}],\"internalType\":\"structIBridgeStructs.SignedBatch\",\"name\":\"_signedBatch\",\"type\":\"tuple\"}],\"name\":\"submitSignedBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"upgradeToAndCall\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
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

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_BridgeContract *BridgeContractCaller) UPGRADEINTERFACEVERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "UPGRADE_INTERFACE_VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_BridgeContract *BridgeContractSession) UPGRADEINTERFACEVERSION() (string, error) {
	return _BridgeContract.Contract.UPGRADEINTERFACEVERSION(&_BridgeContract.CallOpts)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_BridgeContract *BridgeContractCallerSession) UPGRADEINTERFACEVERSION() (string, error) {
	return _BridgeContract.Contract.UPGRADEINTERFACEVERSION(&_BridgeContract.CallOpts)
}

// GetAllRegisteredChains is a free data retrieval call binding the contract method 0x67f0cc44.
//
// Solidity: function getAllRegisteredChains() view returns((string,((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]),string,string)[] _chains)
func (_BridgeContract *BridgeContractCaller) GetAllRegisteredChains(opts *bind.CallOpts) ([]IBridgeStructsChain, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getAllRegisteredChains")

	if err != nil {
		return *new([]IBridgeStructsChain), err
	}

	out0 := *abi.ConvertType(out[0], new([]IBridgeStructsChain)).(*[]IBridgeStructsChain)

	return out0, err

}

// GetAllRegisteredChains is a free data retrieval call binding the contract method 0x67f0cc44.
//
// Solidity: function getAllRegisteredChains() view returns((string,((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]),string,string)[] _chains)
func (_BridgeContract *BridgeContractSession) GetAllRegisteredChains() ([]IBridgeStructsChain, error) {
	return _BridgeContract.Contract.GetAllRegisteredChains(&_BridgeContract.CallOpts)
}

// GetAllRegisteredChains is a free data retrieval call binding the contract method 0x67f0cc44.
//
// Solidity: function getAllRegisteredChains() view returns((string,((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]),string,string)[] _chains)
func (_BridgeContract *BridgeContractCallerSession) GetAllRegisteredChains() ([]IBridgeStructsChain, error) {
	return _BridgeContract.Contract.GetAllRegisteredChains(&_BridgeContract.CallOpts)
}

// GetAvailableUTXOs is a free data retrieval call binding the contract method 0x03fe69ae.
//
// Solidity: function getAvailableUTXOs(string _destinationChain) view returns(((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]) availableUTXOs)
func (_BridgeContract *BridgeContractCaller) GetAvailableUTXOs(opts *bind.CallOpts, _destinationChain string) (IBridgeStructsUTXOs, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getAvailableUTXOs", _destinationChain)

	if err != nil {
		return *new(IBridgeStructsUTXOs), err
	}

	out0 := *abi.ConvertType(out[0], new(IBridgeStructsUTXOs)).(*IBridgeStructsUTXOs)

	return out0, err

}

// GetAvailableUTXOs is a free data retrieval call binding the contract method 0x03fe69ae.
//
// Solidity: function getAvailableUTXOs(string _destinationChain) view returns(((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]) availableUTXOs)
func (_BridgeContract *BridgeContractSession) GetAvailableUTXOs(_destinationChain string) (IBridgeStructsUTXOs, error) {
	return _BridgeContract.Contract.GetAvailableUTXOs(&_BridgeContract.CallOpts, _destinationChain)
}

// GetAvailableUTXOs is a free data retrieval call binding the contract method 0x03fe69ae.
//
// Solidity: function getAvailableUTXOs(string _destinationChain) view returns(((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]) availableUTXOs)
func (_BridgeContract *BridgeContractCallerSession) GetAvailableUTXOs(_destinationChain string) (IBridgeStructsUTXOs, error) {
	return _BridgeContract.Contract.GetAvailableUTXOs(&_BridgeContract.CallOpts, _destinationChain)
}

// GetConfirmedBatch is a free data retrieval call binding the contract method 0xd52c54c4.
//
// Solidity: function getConfirmedBatch(string _destinationChain) view returns((uint256,string,string[],string[]) batch)
func (_BridgeContract *BridgeContractCaller) GetConfirmedBatch(opts *bind.CallOpts, _destinationChain string) (IBridgeStructsConfirmedBatch, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getConfirmedBatch", _destinationChain)

	if err != nil {
		return *new(IBridgeStructsConfirmedBatch), err
	}

	out0 := *abi.ConvertType(out[0], new(IBridgeStructsConfirmedBatch)).(*IBridgeStructsConfirmedBatch)

	return out0, err

}

// GetConfirmedBatch is a free data retrieval call binding the contract method 0xd52c54c4.
//
// Solidity: function getConfirmedBatch(string _destinationChain) view returns((uint256,string,string[],string[]) batch)
func (_BridgeContract *BridgeContractSession) GetConfirmedBatch(_destinationChain string) (IBridgeStructsConfirmedBatch, error) {
	return _BridgeContract.Contract.GetConfirmedBatch(&_BridgeContract.CallOpts, _destinationChain)
}

// GetConfirmedBatch is a free data retrieval call binding the contract method 0xd52c54c4.
//
// Solidity: function getConfirmedBatch(string _destinationChain) view returns((uint256,string,string[],string[]) batch)
func (_BridgeContract *BridgeContractCallerSession) GetConfirmedBatch(_destinationChain string) (IBridgeStructsConfirmedBatch, error) {
	return _BridgeContract.Contract.GetConfirmedBatch(&_BridgeContract.CallOpts, _destinationChain)
}

// GetConfirmedTransactions is a free data retrieval call binding the contract method 0x595051f9.
//
// Solidity: function getConfirmedTransactions(string _destinationChain) view returns((string,uint256,uint256,string,(string,uint256)[])[] _confirmedTransactions)
func (_BridgeContract *BridgeContractCaller) GetConfirmedTransactions(opts *bind.CallOpts, _destinationChain string) ([]IBridgeStructsConfirmedTransaction, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getConfirmedTransactions", _destinationChain)

	if err != nil {
		return *new([]IBridgeStructsConfirmedTransaction), err
	}

	out0 := *abi.ConvertType(out[0], new([]IBridgeStructsConfirmedTransaction)).(*[]IBridgeStructsConfirmedTransaction)

	return out0, err

}

// GetConfirmedTransactions is a free data retrieval call binding the contract method 0x595051f9.
//
// Solidity: function getConfirmedTransactions(string _destinationChain) view returns((string,uint256,uint256,string,(string,uint256)[])[] _confirmedTransactions)
func (_BridgeContract *BridgeContractSession) GetConfirmedTransactions(_destinationChain string) ([]IBridgeStructsConfirmedTransaction, error) {
	return _BridgeContract.Contract.GetConfirmedTransactions(&_BridgeContract.CallOpts, _destinationChain)
}

// GetConfirmedTransactions is a free data retrieval call binding the contract method 0x595051f9.
//
// Solidity: function getConfirmedTransactions(string _destinationChain) view returns((string,uint256,uint256,string,(string,uint256)[])[] _confirmedTransactions)
func (_BridgeContract *BridgeContractCallerSession) GetConfirmedTransactions(_destinationChain string) ([]IBridgeStructsConfirmedTransaction, error) {
	return _BridgeContract.Contract.GetConfirmedTransactions(&_BridgeContract.CallOpts, _destinationChain)
}

// GetLastObservedBlock is a free data retrieval call binding the contract method 0x2175c3f7.
//
// Solidity: function getLastObservedBlock(string _sourceChain) view returns((string,uint64) cblock)
func (_BridgeContract *BridgeContractCaller) GetLastObservedBlock(opts *bind.CallOpts, _sourceChain string) (IBridgeStructsCardanoBlock, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getLastObservedBlock", _sourceChain)

	if err != nil {
		return *new(IBridgeStructsCardanoBlock), err
	}

	out0 := *abi.ConvertType(out[0], new(IBridgeStructsCardanoBlock)).(*IBridgeStructsCardanoBlock)

	return out0, err

}

// GetLastObservedBlock is a free data retrieval call binding the contract method 0x2175c3f7.
//
// Solidity: function getLastObservedBlock(string _sourceChain) view returns((string,uint64) cblock)
func (_BridgeContract *BridgeContractSession) GetLastObservedBlock(_sourceChain string) (IBridgeStructsCardanoBlock, error) {
	return _BridgeContract.Contract.GetLastObservedBlock(&_BridgeContract.CallOpts, _sourceChain)
}

// GetLastObservedBlock is a free data retrieval call binding the contract method 0x2175c3f7.
//
// Solidity: function getLastObservedBlock(string _sourceChain) view returns((string,uint64) cblock)
func (_BridgeContract *BridgeContractCallerSession) GetLastObservedBlock(_sourceChain string) (IBridgeStructsCardanoBlock, error) {
	return _BridgeContract.Contract.GetLastObservedBlock(&_BridgeContract.CallOpts, _sourceChain)
}

// GetNextBatchId is a free data retrieval call binding the contract method 0x3cd9ae3e.
//
// Solidity: function getNextBatchId(string _destinationChain) view returns(uint256 result)
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
// Solidity: function getNextBatchId(string _destinationChain) view returns(uint256 result)
func (_BridgeContract *BridgeContractSession) GetNextBatchId(_destinationChain string) (*big.Int, error) {
	return _BridgeContract.Contract.GetNextBatchId(&_BridgeContract.CallOpts, _destinationChain)
}

// GetNextBatchId is a free data retrieval call binding the contract method 0x3cd9ae3e.
//
// Solidity: function getNextBatchId(string _destinationChain) view returns(uint256 result)
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
// Solidity: function getValidatorsCardanoData(string _chainId) view returns((string,string)[] validatorCardanoData)
func (_BridgeContract *BridgeContractCaller) GetValidatorsCardanoData(opts *bind.CallOpts, _chainId string) ([]IBridgeStructsValidatorCardanoData, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getValidatorsCardanoData", _chainId)

	if err != nil {
		return *new([]IBridgeStructsValidatorCardanoData), err
	}

	out0 := *abi.ConvertType(out[0], new([]IBridgeStructsValidatorCardanoData)).(*[]IBridgeStructsValidatorCardanoData)

	return out0, err

}

// GetValidatorsCardanoData is a free data retrieval call binding the contract method 0x636b8a0d.
//
// Solidity: function getValidatorsCardanoData(string _chainId) view returns((string,string)[] validatorCardanoData)
func (_BridgeContract *BridgeContractSession) GetValidatorsCardanoData(_chainId string) ([]IBridgeStructsValidatorCardanoData, error) {
	return _BridgeContract.Contract.GetValidatorsCardanoData(&_BridgeContract.CallOpts, _chainId)
}

// GetValidatorsCardanoData is a free data retrieval call binding the contract method 0x636b8a0d.
//
// Solidity: function getValidatorsCardanoData(string _chainId) view returns((string,string)[] validatorCardanoData)
func (_BridgeContract *BridgeContractCallerSession) GetValidatorsCardanoData(_chainId string) ([]IBridgeStructsValidatorCardanoData, error) {
	return _BridgeContract.Contract.GetValidatorsCardanoData(&_BridgeContract.CallOpts, _chainId)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BridgeContract *BridgeContractCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BridgeContract *BridgeContractSession) Owner() (common.Address, error) {
	return _BridgeContract.Contract.Owner(&_BridgeContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BridgeContract *BridgeContractCallerSession) Owner() (common.Address, error) {
	return _BridgeContract.Contract.Owner(&_BridgeContract.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_BridgeContract *BridgeContractCaller) ProxiableUUID(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "proxiableUUID")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_BridgeContract *BridgeContractSession) ProxiableUUID() ([32]byte, error) {
	return _BridgeContract.Contract.ProxiableUUID(&_BridgeContract.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_BridgeContract *BridgeContractCallerSession) ProxiableUUID() ([32]byte, error) {
	return _BridgeContract.Contract.ProxiableUUID(&_BridgeContract.CallOpts)
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

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_BridgeContract *BridgeContractTransactor) Initialize(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "initialize")
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_BridgeContract *BridgeContractSession) Initialize() (*types.Transaction, error) {
	return _BridgeContract.Contract.Initialize(&_BridgeContract.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_BridgeContract *BridgeContractTransactorSession) Initialize() (*types.Transaction, error) {
	return _BridgeContract.Contract.Initialize(&_BridgeContract.TransactOpts)
}

// RegisterChain is a paid mutator transaction binding the contract method 0xf75b143f.
//
// Solidity: function registerChain(string _chainId, ((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]) _initialUTXOs, string _addressMultisig, string _addressFeePayer, (address,(string,string))[] _validatorsAddressCardanoData, uint256 _tokenQuantity) returns()
func (_BridgeContract *BridgeContractTransactor) RegisterChain(opts *bind.TransactOpts, _chainId string, _initialUTXOs IBridgeStructsUTXOs, _addressMultisig string, _addressFeePayer string, _validatorsAddressCardanoData []IBridgeStructsValidatorAddressCardanoData, _tokenQuantity *big.Int) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "registerChain", _chainId, _initialUTXOs, _addressMultisig, _addressFeePayer, _validatorsAddressCardanoData, _tokenQuantity)
}

// RegisterChain is a paid mutator transaction binding the contract method 0xf75b143f.
//
// Solidity: function registerChain(string _chainId, ((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]) _initialUTXOs, string _addressMultisig, string _addressFeePayer, (address,(string,string))[] _validatorsAddressCardanoData, uint256 _tokenQuantity) returns()
func (_BridgeContract *BridgeContractSession) RegisterChain(_chainId string, _initialUTXOs IBridgeStructsUTXOs, _addressMultisig string, _addressFeePayer string, _validatorsAddressCardanoData []IBridgeStructsValidatorAddressCardanoData, _tokenQuantity *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.RegisterChain(&_BridgeContract.TransactOpts, _chainId, _initialUTXOs, _addressMultisig, _addressFeePayer, _validatorsAddressCardanoData, _tokenQuantity)
}

// RegisterChain is a paid mutator transaction binding the contract method 0xf75b143f.
//
// Solidity: function registerChain(string _chainId, ((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]) _initialUTXOs, string _addressMultisig, string _addressFeePayer, (address,(string,string))[] _validatorsAddressCardanoData, uint256 _tokenQuantity) returns()
func (_BridgeContract *BridgeContractTransactorSession) RegisterChain(_chainId string, _initialUTXOs IBridgeStructsUTXOs, _addressMultisig string, _addressFeePayer string, _validatorsAddressCardanoData []IBridgeStructsValidatorAddressCardanoData, _tokenQuantity *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.RegisterChain(&_BridgeContract.TransactOpts, _chainId, _initialUTXOs, _addressMultisig, _addressFeePayer, _validatorsAddressCardanoData, _tokenQuantity)
}

// RegisterChainGovernance is a paid mutator transaction binding the contract method 0xbf522790.
//
// Solidity: function registerChainGovernance(string _chainId, ((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]) _initialUTXOs, string _addressMultisig, string _addressFeePayer, (string,string) _validatorCardanoData, uint256 _tokenQuantity) returns()
func (_BridgeContract *BridgeContractTransactor) RegisterChainGovernance(opts *bind.TransactOpts, _chainId string, _initialUTXOs IBridgeStructsUTXOs, _addressMultisig string, _addressFeePayer string, _validatorCardanoData IBridgeStructsValidatorCardanoData, _tokenQuantity *big.Int) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "registerChainGovernance", _chainId, _initialUTXOs, _addressMultisig, _addressFeePayer, _validatorCardanoData, _tokenQuantity)
}

// RegisterChainGovernance is a paid mutator transaction binding the contract method 0xbf522790.
//
// Solidity: function registerChainGovernance(string _chainId, ((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]) _initialUTXOs, string _addressMultisig, string _addressFeePayer, (string,string) _validatorCardanoData, uint256 _tokenQuantity) returns()
func (_BridgeContract *BridgeContractSession) RegisterChainGovernance(_chainId string, _initialUTXOs IBridgeStructsUTXOs, _addressMultisig string, _addressFeePayer string, _validatorCardanoData IBridgeStructsValidatorCardanoData, _tokenQuantity *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.RegisterChainGovernance(&_BridgeContract.TransactOpts, _chainId, _initialUTXOs, _addressMultisig, _addressFeePayer, _validatorCardanoData, _tokenQuantity)
}

// RegisterChainGovernance is a paid mutator transaction binding the contract method 0xbf522790.
//
// Solidity: function registerChainGovernance(string _chainId, ((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]) _initialUTXOs, string _addressMultisig, string _addressFeePayer, (string,string) _validatorCardanoData, uint256 _tokenQuantity) returns()
func (_BridgeContract *BridgeContractTransactorSession) RegisterChainGovernance(_chainId string, _initialUTXOs IBridgeStructsUTXOs, _addressMultisig string, _addressFeePayer string, _validatorCardanoData IBridgeStructsValidatorCardanoData, _tokenQuantity *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.RegisterChainGovernance(&_BridgeContract.TransactOpts, _chainId, _initialUTXOs, _addressMultisig, _addressFeePayer, _validatorCardanoData, _tokenQuantity)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BridgeContract *BridgeContractTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BridgeContract *BridgeContractSession) RenounceOwnership() (*types.Transaction, error) {
	return _BridgeContract.Contract.RenounceOwnership(&_BridgeContract.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BridgeContract *BridgeContractTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _BridgeContract.Contract.RenounceOwnership(&_BridgeContract.TransactOpts)
}

// SetDependencies is a paid mutator transaction binding the contract method 0x997ef201.
//
// Solidity: function setDependencies(address _claimsAddress, address _signedBatchesAddress, address _slotsAddress, address _utxoscAddress, address _validatorsAddress) returns()
func (_BridgeContract *BridgeContractTransactor) SetDependencies(opts *bind.TransactOpts, _claimsAddress common.Address, _signedBatchesAddress common.Address, _slotsAddress common.Address, _utxoscAddress common.Address, _validatorsAddress common.Address) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "setDependencies", _claimsAddress, _signedBatchesAddress, _slotsAddress, _utxoscAddress, _validatorsAddress)
}

// SetDependencies is a paid mutator transaction binding the contract method 0x997ef201.
//
// Solidity: function setDependencies(address _claimsAddress, address _signedBatchesAddress, address _slotsAddress, address _utxoscAddress, address _validatorsAddress) returns()
func (_BridgeContract *BridgeContractSession) SetDependencies(_claimsAddress common.Address, _signedBatchesAddress common.Address, _slotsAddress common.Address, _utxoscAddress common.Address, _validatorsAddress common.Address) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetDependencies(&_BridgeContract.TransactOpts, _claimsAddress, _signedBatchesAddress, _slotsAddress, _utxoscAddress, _validatorsAddress)
}

// SetDependencies is a paid mutator transaction binding the contract method 0x997ef201.
//
// Solidity: function setDependencies(address _claimsAddress, address _signedBatchesAddress, address _slotsAddress, address _utxoscAddress, address _validatorsAddress) returns()
func (_BridgeContract *BridgeContractTransactorSession) SetDependencies(_claimsAddress common.Address, _signedBatchesAddress common.Address, _slotsAddress common.Address, _utxoscAddress common.Address, _validatorsAddress common.Address) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetDependencies(&_BridgeContract.TransactOpts, _claimsAddress, _signedBatchesAddress, _slotsAddress, _utxoscAddress, _validatorsAddress)
}

// SubmitClaims is a paid mutator transaction binding the contract method 0xb95a432c.
//
// Solidity: function submitClaims(((string,(string,uint256)[],(uint64,string,uint256,uint256),string,string)[],(string,string,uint256,((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]))[],(string,string,uint256)[],(string,string,string,string,(uint64,string,uint256,uint256),string,string,uint256)[],(string,string,string,(uint64,string,uint256,uint256))[]) _claims) returns()
func (_BridgeContract *BridgeContractTransactor) SubmitClaims(opts *bind.TransactOpts, _claims IBridgeStructsValidatorClaims) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "submitClaims", _claims)
}

// SubmitClaims is a paid mutator transaction binding the contract method 0xb95a432c.
//
// Solidity: function submitClaims(((string,(string,uint256)[],(uint64,string,uint256,uint256),string,string)[],(string,string,uint256,((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]))[],(string,string,uint256)[],(string,string,string,string,(uint64,string,uint256,uint256),string,string,uint256)[],(string,string,string,(uint64,string,uint256,uint256))[]) _claims) returns()
func (_BridgeContract *BridgeContractSession) SubmitClaims(_claims IBridgeStructsValidatorClaims) (*types.Transaction, error) {
	return _BridgeContract.Contract.SubmitClaims(&_BridgeContract.TransactOpts, _claims)
}

// SubmitClaims is a paid mutator transaction binding the contract method 0xb95a432c.
//
// Solidity: function submitClaims(((string,(string,uint256)[],(uint64,string,uint256,uint256),string,string)[],(string,string,uint256,((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[]))[],(string,string,uint256)[],(string,string,string,string,(uint64,string,uint256,uint256),string,string,uint256)[],(string,string,string,(uint64,string,uint256,uint256))[]) _claims) returns()
func (_BridgeContract *BridgeContractTransactorSession) SubmitClaims(_claims IBridgeStructsValidatorClaims) (*types.Transaction, error) {
	return _BridgeContract.Contract.SubmitClaims(&_BridgeContract.TransactOpts, _claims)
}

// SubmitLastObservedBlocks is a paid mutator transaction binding the contract method 0x406f8f04.
//
// Solidity: function submitLastObservedBlocks(string chainID, (string,uint64)[] blocks) returns()
func (_BridgeContract *BridgeContractTransactor) SubmitLastObservedBlocks(opts *bind.TransactOpts, chainID string, blocks []IBridgeStructsCardanoBlock) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "submitLastObservedBlocks", chainID, blocks)
}

// SubmitLastObservedBlocks is a paid mutator transaction binding the contract method 0x406f8f04.
//
// Solidity: function submitLastObservedBlocks(string chainID, (string,uint64)[] blocks) returns()
func (_BridgeContract *BridgeContractSession) SubmitLastObservedBlocks(chainID string, blocks []IBridgeStructsCardanoBlock) (*types.Transaction, error) {
	return _BridgeContract.Contract.SubmitLastObservedBlocks(&_BridgeContract.TransactOpts, chainID, blocks)
}

// SubmitLastObservedBlocks is a paid mutator transaction binding the contract method 0x406f8f04.
//
// Solidity: function submitLastObservedBlocks(string chainID, (string,uint64)[] blocks) returns()
func (_BridgeContract *BridgeContractTransactorSession) SubmitLastObservedBlocks(chainID string, blocks []IBridgeStructsCardanoBlock) (*types.Transaction, error) {
	return _BridgeContract.Contract.SubmitLastObservedBlocks(&_BridgeContract.TransactOpts, chainID, blocks)
}

// SubmitSignedBatch is a paid mutator transaction binding the contract method 0xf39b5f49.
//
// Solidity: function submitSignedBatch((uint256,string,string,string,string,uint256[],((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[])) _signedBatch) returns()
func (_BridgeContract *BridgeContractTransactor) SubmitSignedBatch(opts *bind.TransactOpts, _signedBatch IBridgeStructsSignedBatch) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "submitSignedBatch", _signedBatch)
}

// SubmitSignedBatch is a paid mutator transaction binding the contract method 0xf39b5f49.
//
// Solidity: function submitSignedBatch((uint256,string,string,string,string,uint256[],((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[])) _signedBatch) returns()
func (_BridgeContract *BridgeContractSession) SubmitSignedBatch(_signedBatch IBridgeStructsSignedBatch) (*types.Transaction, error) {
	return _BridgeContract.Contract.SubmitSignedBatch(&_BridgeContract.TransactOpts, _signedBatch)
}

// SubmitSignedBatch is a paid mutator transaction binding the contract method 0xf39b5f49.
//
// Solidity: function submitSignedBatch((uint256,string,string,string,string,uint256[],((uint64,string,uint256,uint256)[],(uint64,string,uint256,uint256)[])) _signedBatch) returns()
func (_BridgeContract *BridgeContractTransactorSession) SubmitSignedBatch(_signedBatch IBridgeStructsSignedBatch) (*types.Transaction, error) {
	return _BridgeContract.Contract.SubmitSignedBatch(&_BridgeContract.TransactOpts, _signedBatch)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BridgeContract *BridgeContractTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BridgeContract *BridgeContractSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BridgeContract.Contract.TransferOwnership(&_BridgeContract.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BridgeContract *BridgeContractTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BridgeContract.Contract.TransferOwnership(&_BridgeContract.TransactOpts, newOwner)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_BridgeContract *BridgeContractTransactor) UpgradeToAndCall(opts *bind.TransactOpts, newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "upgradeToAndCall", newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_BridgeContract *BridgeContractSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _BridgeContract.Contract.UpgradeToAndCall(&_BridgeContract.TransactOpts, newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_BridgeContract *BridgeContractTransactorSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _BridgeContract.Contract.UpgradeToAndCall(&_BridgeContract.TransactOpts, newImplementation, data)
}

// BridgeContractInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the BridgeContract contract.
type BridgeContractInitializedIterator struct {
	Event *BridgeContractInitialized // Event containing the contract specifics and raw log

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
func (it *BridgeContractInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractInitialized)
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
		it.Event = new(BridgeContractInitialized)
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
func (it *BridgeContractInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractInitialized represents a Initialized event raised by the BridgeContract contract.
type BridgeContractInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_BridgeContract *BridgeContractFilterer) FilterInitialized(opts *bind.FilterOpts) (*BridgeContractInitializedIterator, error) {

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &BridgeContractInitializedIterator{contract: _BridgeContract.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_BridgeContract *BridgeContractFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *BridgeContractInitialized) (event.Subscription, error) {

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractInitialized)
				if err := _BridgeContract.contract.UnpackLog(event, "Initialized", log); err != nil {
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

// ParseInitialized is a log parse operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_BridgeContract *BridgeContractFilterer) ParseInitialized(log types.Log) (*BridgeContractInitialized, error) {
	event := new(BridgeContractInitialized)
	if err := _BridgeContract.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeContractOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the BridgeContract contract.
type BridgeContractOwnershipTransferredIterator struct {
	Event *BridgeContractOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *BridgeContractOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractOwnershipTransferred)
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
		it.Event = new(BridgeContractOwnershipTransferred)
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
func (it *BridgeContractOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractOwnershipTransferred represents a OwnershipTransferred event raised by the BridgeContract contract.
type BridgeContractOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BridgeContract *BridgeContractFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BridgeContractOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BridgeContractOwnershipTransferredIterator{contract: _BridgeContract.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BridgeContract *BridgeContractFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BridgeContractOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractOwnershipTransferred)
				if err := _BridgeContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BridgeContract *BridgeContractFilterer) ParseOwnershipTransferred(log types.Log) (*BridgeContractOwnershipTransferred, error) {
	event := new(BridgeContractOwnershipTransferred)
	if err := _BridgeContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeContractUpgradedIterator is returned from FilterUpgraded and is used to iterate over the raw logs and unpacked data for Upgraded events raised by the BridgeContract contract.
type BridgeContractUpgradedIterator struct {
	Event *BridgeContractUpgraded // Event containing the contract specifics and raw log

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
func (it *BridgeContractUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractUpgraded)
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
		it.Event = new(BridgeContractUpgraded)
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
func (it *BridgeContractUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractUpgraded represents a Upgraded event raised by the BridgeContract contract.
type BridgeContractUpgraded struct {
	Implementation common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterUpgraded is a free log retrieval operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_BridgeContract *BridgeContractFilterer) FilterUpgraded(opts *bind.FilterOpts, implementation []common.Address) (*BridgeContractUpgradedIterator, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return &BridgeContractUpgradedIterator{contract: _BridgeContract.contract, event: "Upgraded", logs: logs, sub: sub}, nil
}

// WatchUpgraded is a free log subscription operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_BridgeContract *BridgeContractFilterer) WatchUpgraded(opts *bind.WatchOpts, sink chan<- *BridgeContractUpgraded, implementation []common.Address) (event.Subscription, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractUpgraded)
				if err := _BridgeContract.contract.UnpackLog(event, "Upgraded", log); err != nil {
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

// ParseUpgraded is a log parse operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_BridgeContract *BridgeContractFilterer) ParseUpgraded(log types.Log) (*BridgeContractUpgraded, error) {
	event := new(BridgeContractUpgraded)
	if err := _BridgeContract.contract.UnpackLog(event, "Upgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
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
