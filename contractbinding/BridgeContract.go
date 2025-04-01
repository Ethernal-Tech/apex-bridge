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
	ObservedTransactionHash [32]byte
	BatchNonceId            uint64
	ChainId                 uint8
}

// IBridgeStructsBatchExecutionFailedClaim is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsBatchExecutionFailedClaim struct {
	ObservedTransactionHash [32]byte
	BatchNonceId            uint64
	ChainId                 uint8
}

// IBridgeStructsBridgingRequestClaim is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsBridgingRequestClaim struct {
	ObservedTransactionHash [32]byte
	Receivers               []IBridgeStructsReceiver
	TotalAmount             *big.Int
	RetryCounter            *big.Int
	SourceChainId           uint8
	DestinationChainId      uint8
}

// IBridgeStructsCardanoBlock is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsCardanoBlock struct {
	BlockSlot *big.Int
	BlockHash [32]byte
}

// IBridgeStructsChain is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsChain struct {
	Id              uint8
	ChainType       uint8
	AddressMultisig string
	AddressFeePayer string
}

// IBridgeStructsConfirmedBatch is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsConfirmedBatch struct {
	Signatures      [][]byte
	FeeSignatures   [][]byte
	Bitmap          *big.Int
	RawTransaction  []byte
	Id              uint64
	IsConsolidation bool
}

// IBridgeStructsConfirmedTransaction is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsConfirmedTransaction struct {
	BlockHeight             *big.Int
	TotalAmount             *big.Int
	RetryCounter            *big.Int
	ObservedTransactionHash [32]byte
	Nonce                   uint64
	SourceChainId           uint8
	TransactionType         uint8
	Receivers               []IBridgeStructsReceiver
	OutputIndexes           []byte
}

// IBridgeStructsHotWalletIncrementClaim is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsHotWalletIncrementClaim struct {
	ChainId uint8
	Amount  *big.Int
}

// IBridgeStructsReceiver is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsReceiver struct {
	Amount             *big.Int
	DestinationAddress string
}

// IBridgeStructsRefundRequestClaim is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsRefundRequestClaim struct {
	OriginTransactionHash    [32]byte
	RefundTransactionHash    [32]byte
	OriginAmount             *big.Int
	OutputIndexes            []byte
	OriginSenderAddress      string
	RetryCounter             uint64
	OriginChainId            uint8
	ShouldDecrementHotWallet bool
}

// IBridgeStructsSignedBatch is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsSignedBatch struct {
	Id                 uint64
	FirstTxNonceId     uint64
	LastTxNonceId      uint64
	DestinationChainId uint8
	Signature          []byte
	FeeSignature       []byte
	RawTransaction     []byte
	IsConsolidation    bool
}

// IBridgeStructsTxDataInfo is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsTxDataInfo struct {
	ObservedTransactionHash [32]byte
	SourceChainId           uint8
	TransactionType         uint8
}

// IBridgeStructsValidatorAddressChainData is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsValidatorAddressChainData struct {
	Addr common.Address
	Data IBridgeStructsValidatorChainData
}

// IBridgeStructsValidatorChainData is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsValidatorChainData struct {
	Key [4]*big.Int
}

// IBridgeStructsValidatorClaims is an auto generated low-level Go binding around an user-defined struct.
type IBridgeStructsValidatorClaims struct {
	BridgingRequestClaims      []IBridgeStructsBridgingRequestClaim
	BatchExecutedClaims        []IBridgeStructsBatchExecutedClaim
	BatchExecutionFailedClaims []IBridgeStructsBatchExecutionFailedClaim
	RefundRequestClaims        []IBridgeStructsRefundRequestClaim
	HotWalletIncrementClaims   []IBridgeStructsHotWalletIncrementClaim
}

// BridgeContractMetaData contains all meta data concerning the BridgeContract contract.
var BridgeContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_claimTransactionHash\",\"type\":\"bytes32\"}],\"name\":\"AlreadyConfirmed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_claimTransactionHash\",\"type\":\"uint8\"}],\"name\":\"AlreadyProposed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"}],\"name\":\"CanNotCreateBatchYet\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"}],\"name\":\"ChainAlreadyRegistered\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"}],\"name\":\"ChainIsNotRegistered\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"_availableAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_requestedAmount\",\"type\":\"uint256\"}],\"name\":\"DefundRequestTooHigh\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"data\",\"type\":\"string\"}],\"name\":\"InvalidData\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSignature\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_availableAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_decreaseAmount\",\"type\":\"uint256\"}],\"name\":\"NegativeChainTokenAmount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotAdminContract\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotBridge\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotClaims\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_claimTransactionHash\",\"type\":\"bytes32\"}],\"name\":\"NotEnoughBridgingTokensAvailable\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotFundAdmin\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSignedBatches\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSignedBatchesOrBridge\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSignedBatchesOrClaims\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotUpgradeAdmin\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotValidator\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"},{\"internalType\":\"uint64\",\"name\":\"_nonce\",\"type\":\"uint64\"}],\"name\":\"WrongBatchNonce\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroAddress\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"previousAdmin\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"AdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"beacon\",\"type\":\"address\"}],\"name\":\"BeaconUpgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"ChainDefunded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"DefundFailedAfterMultipleRetries\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_newFundAdmin\",\"type\":\"address\"}],\"name\":\"FundAdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"claimeType\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"availableAmount\",\"type\":\"uint256\"}],\"name\":\"NotEnoughFunds\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"isIncrement\",\"type\":\"bool\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenQuantity\",\"type\":\"uint256\"}],\"name\":\"UpdatedChainTokenQuantity\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_maxNumberOfTransactions\",\"type\":\"uint256\"}],\"name\":\"UpdatedMaxNumberOfTransactions\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_timeoutBlocksNumber\",\"type\":\"uint256\"}],\"name\":\"UpdatedTimeoutBlocksNumber\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"Upgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"newChainProposal\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"}],\"name\":\"newChainRegistered\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"getAllRegisteredChains\",\"outputs\":[{\"components\":[{\"internalType\":\"uint8\",\"name\":\"id\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"chainType\",\"type\":\"uint8\"},{\"internalType\":\"string\",\"name\":\"addressMultisig\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"addressFeePayer\",\"type\":\"string\"}],\"internalType\":\"structIBridgeStructs.Chain[]\",\"name\":\"_chains\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"},{\"internalType\":\"uint64\",\"name\":\"_batchId\",\"type\":\"uint64\"}],\"name\":\"getBatchTransactions\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"observedTransactionHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"sourceChainId\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"transactionType\",\"type\":\"uint8\"}],\"internalType\":\"structIBridgeStructs.TxDataInfo[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_destinationChain\",\"type\":\"uint8\"}],\"name\":\"getConfirmedBatch\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes[]\",\"name\":\"signatures\",\"type\":\"bytes[]\"},{\"internalType\":\"bytes[]\",\"name\":\"feeSignatures\",\"type\":\"bytes[]\"},{\"internalType\":\"uint256\",\"name\":\"bitmap\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"rawTransaction\",\"type\":\"bytes\"},{\"internalType\":\"uint64\",\"name\":\"id\",\"type\":\"uint64\"},{\"internalType\":\"bool\",\"name\":\"isConsolidation\",\"type\":\"bool\"}],\"internalType\":\"structIBridgeStructs.ConfirmedBatch\",\"name\":\"_batch\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_destinationChain\",\"type\":\"uint8\"}],\"name\":\"getConfirmedTransactions\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"blockHeight\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"retryCounter\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"observedTransactionHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"uint8\",\"name\":\"sourceChainId\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"transactionType\",\"type\":\"uint8\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"destinationAddress\",\"type\":\"string\"}],\"internalType\":\"structIBridgeStructs.Receiver[]\",\"name\":\"receivers\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes\",\"name\":\"outputIndexes\",\"type\":\"bytes\"}],\"internalType\":\"structIBridgeStructs.ConfirmedTransaction[]\",\"name\":\"_confirmedTransactions\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_sourceChain\",\"type\":\"uint8\"}],\"name\":\"getLastObservedBlock\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"blockSlot\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"blockHash\",\"type\":\"bytes32\"}],\"internalType\":\"structIBridgeStructs.CardanoBlock\",\"name\":\"_cblock\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_destinationChain\",\"type\":\"uint8\"}],\"name\":\"getNextBatchId\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"_result\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_destinationChain\",\"type\":\"uint8\"}],\"name\":\"getRawTransactionFromLastBatch\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"}],\"name\":\"getValidatorsChainData\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256[4]\",\"name\":\"key\",\"type\":\"uint256[4]\"}],\"internalType\":\"structIBridgeStructs.ValidatorChainData[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_upgradeAdmin\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"proxiableUUID\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint8\",\"name\":\"id\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"chainType\",\"type\":\"uint8\"},{\"internalType\":\"string\",\"name\":\"addressMultisig\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"addressFeePayer\",\"type\":\"string\"}],\"internalType\":\"structIBridgeStructs.Chain\",\"name\":\"_chain\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"_tokenQuantity\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"uint256[4]\",\"name\":\"key\",\"type\":\"uint256[4]\"}],\"internalType\":\"structIBridgeStructs.ValidatorChainData\",\"name\":\"data\",\"type\":\"tuple\"}],\"internalType\":\"structIBridgeStructs.ValidatorAddressChainData[]\",\"name\":\"_chainData\",\"type\":\"tuple[]\"}],\"name\":\"registerChain\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"_chainType\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"_tokenQuantity\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"uint256[4]\",\"name\":\"key\",\"type\":\"uint256[4]\"}],\"internalType\":\"structIBridgeStructs.ValidatorChainData\",\"name\":\"_validatorChainData\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"_keySignature\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"_keyFeeSignature\",\"type\":\"bytes\"}],\"name\":\"registerChainGovernance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"},{\"internalType\":\"string\",\"name\":\"addressMultisig\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"addressFeePayer\",\"type\":\"string\"}],\"name\":\"setChainAdditionalData\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_claimsAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_signedBatchesAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_slotsAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_validatorsAddress\",\"type\":\"address\"}],\"name\":\"setDependencies\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_destinationChain\",\"type\":\"uint8\"}],\"name\":\"shouldCreateBatch\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"_batch\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"observedTransactionHash\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"destinationAddress\",\"type\":\"string\"}],\"internalType\":\"structIBridgeStructs.Receiver[]\",\"name\":\"receivers\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256\",\"name\":\"totalAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"retryCounter\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"sourceChainId\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"destinationChainId\",\"type\":\"uint8\"}],\"internalType\":\"structIBridgeStructs.BridgingRequestClaim[]\",\"name\":\"bridgingRequestClaims\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"observedTransactionHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"batchNonceId\",\"type\":\"uint64\"},{\"internalType\":\"uint8\",\"name\":\"chainId\",\"type\":\"uint8\"}],\"internalType\":\"structIBridgeStructs.BatchExecutedClaim[]\",\"name\":\"batchExecutedClaims\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"observedTransactionHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"batchNonceId\",\"type\":\"uint64\"},{\"internalType\":\"uint8\",\"name\":\"chainId\",\"type\":\"uint8\"}],\"internalType\":\"structIBridgeStructs.BatchExecutionFailedClaim[]\",\"name\":\"batchExecutionFailedClaims\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"originTransactionHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"refundTransactionHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"originAmount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"outputIndexes\",\"type\":\"bytes\"},{\"internalType\":\"string\",\"name\":\"originSenderAddress\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"retryCounter\",\"type\":\"uint64\"},{\"internalType\":\"uint8\",\"name\":\"originChainId\",\"type\":\"uint8\"},{\"internalType\":\"bool\",\"name\":\"shouldDecrementHotWallet\",\"type\":\"bool\"}],\"internalType\":\"structIBridgeStructs.RefundRequestClaim[]\",\"name\":\"refundRequestClaims\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint8\",\"name\":\"chainId\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIBridgeStructs.HotWalletIncrementClaim[]\",\"name\":\"hotWalletIncrementClaims\",\"type\":\"tuple[]\"}],\"internalType\":\"structIBridgeStructs.ValidatorClaims\",\"name\":\"_claims\",\"type\":\"tuple\"}],\"name\":\"submitClaims\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"blockSlot\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"blockHash\",\"type\":\"bytes32\"}],\"internalType\":\"structIBridgeStructs.CardanoBlock[]\",\"name\":\"_blocks\",\"type\":\"tuple[]\"}],\"name\":\"submitLastObservedBlocks\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"id\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"firstTxNonceId\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"lastTxNonceId\",\"type\":\"uint64\"},{\"internalType\":\"uint8\",\"name\":\"destinationChainId\",\"type\":\"uint8\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"feeSignature\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"rawTransaction\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isConsolidation\",\"type\":\"bool\"}],\"internalType\":\"structIBridgeStructs.SignedBatch\",\"name\":\"_signedBatch\",\"type\":\"tuple\"}],\"name\":\"submitSignedBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"id\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"firstTxNonceId\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"lastTxNonceId\",\"type\":\"uint64\"},{\"internalType\":\"uint8\",\"name\":\"destinationChainId\",\"type\":\"uint8\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"feeSignature\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"rawTransaction\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isConsolidation\",\"type\":\"bool\"}],\"internalType\":\"structIBridgeStructs.SignedBatch\",\"name\":\"_signedBatch\",\"type\":\"tuple\"}],\"name\":\"submitSignedBatchEVM\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"}],\"name\":\"upgradeTo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"upgradeToAndCall\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
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
// Solidity: function getAllRegisteredChains() view returns((uint8,uint8,string,string)[] _chains)
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
// Solidity: function getAllRegisteredChains() view returns((uint8,uint8,string,string)[] _chains)
func (_BridgeContract *BridgeContractSession) GetAllRegisteredChains() ([]IBridgeStructsChain, error) {
	return _BridgeContract.Contract.GetAllRegisteredChains(&_BridgeContract.CallOpts)
}

// GetAllRegisteredChains is a free data retrieval call binding the contract method 0x67f0cc44.
//
// Solidity: function getAllRegisteredChains() view returns((uint8,uint8,string,string)[] _chains)
func (_BridgeContract *BridgeContractCallerSession) GetAllRegisteredChains() ([]IBridgeStructsChain, error) {
	return _BridgeContract.Contract.GetAllRegisteredChains(&_BridgeContract.CallOpts)
}

// GetBatchTransactions is a free data retrieval call binding the contract method 0x8615a765.
//
// Solidity: function getBatchTransactions(uint8 _chainId, uint64 _batchId) view returns((bytes32,uint8,uint8)[])
func (_BridgeContract *BridgeContractCaller) GetBatchTransactions(opts *bind.CallOpts, _chainId uint8, _batchId uint64) ([]IBridgeStructsTxDataInfo, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getBatchTransactions", _chainId, _batchId)

	if err != nil {
		return *new([]IBridgeStructsTxDataInfo), err
	}

	out0 := *abi.ConvertType(out[0], new([]IBridgeStructsTxDataInfo)).(*[]IBridgeStructsTxDataInfo)

	return out0, err

}

// GetBatchTransactions is a free data retrieval call binding the contract method 0x8615a765.
//
// Solidity: function getBatchTransactions(uint8 _chainId, uint64 _batchId) view returns((bytes32,uint8,uint8)[])
func (_BridgeContract *BridgeContractSession) GetBatchTransactions(_chainId uint8, _batchId uint64) ([]IBridgeStructsTxDataInfo, error) {
	return _BridgeContract.Contract.GetBatchTransactions(&_BridgeContract.CallOpts, _chainId, _batchId)
}

// GetBatchTransactions is a free data retrieval call binding the contract method 0x8615a765.
//
// Solidity: function getBatchTransactions(uint8 _chainId, uint64 _batchId) view returns((bytes32,uint8,uint8)[])
func (_BridgeContract *BridgeContractCallerSession) GetBatchTransactions(_chainId uint8, _batchId uint64) ([]IBridgeStructsTxDataInfo, error) {
	return _BridgeContract.Contract.GetBatchTransactions(&_BridgeContract.CallOpts, _chainId, _batchId)
}

// GetConfirmedBatch is a free data retrieval call binding the contract method 0x865e768e.
//
// Solidity: function getConfirmedBatch(uint8 _destinationChain) view returns((bytes[],bytes[],uint256,bytes,uint64,bool) _batch)
func (_BridgeContract *BridgeContractCaller) GetConfirmedBatch(opts *bind.CallOpts, _destinationChain uint8) (IBridgeStructsConfirmedBatch, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getConfirmedBatch", _destinationChain)

	if err != nil {
		return *new(IBridgeStructsConfirmedBatch), err
	}

	out0 := *abi.ConvertType(out[0], new(IBridgeStructsConfirmedBatch)).(*IBridgeStructsConfirmedBatch)

	return out0, err

}

// GetConfirmedBatch is a free data retrieval call binding the contract method 0x865e768e.
//
// Solidity: function getConfirmedBatch(uint8 _destinationChain) view returns((bytes[],bytes[],uint256,bytes,uint64,bool) _batch)
func (_BridgeContract *BridgeContractSession) GetConfirmedBatch(_destinationChain uint8) (IBridgeStructsConfirmedBatch, error) {
	return _BridgeContract.Contract.GetConfirmedBatch(&_BridgeContract.CallOpts, _destinationChain)
}

// GetConfirmedBatch is a free data retrieval call binding the contract method 0x865e768e.
//
// Solidity: function getConfirmedBatch(uint8 _destinationChain) view returns((bytes[],bytes[],uint256,bytes,uint64,bool) _batch)
func (_BridgeContract *BridgeContractCallerSession) GetConfirmedBatch(_destinationChain uint8) (IBridgeStructsConfirmedBatch, error) {
	return _BridgeContract.Contract.GetConfirmedBatch(&_BridgeContract.CallOpts, _destinationChain)
}

// GetConfirmedTransactions is a free data retrieval call binding the contract method 0x4cae8087.
//
// Solidity: function getConfirmedTransactions(uint8 _destinationChain) view returns((uint256,uint256,uint256,bytes32,uint64,uint8,uint8,(uint256,string)[],bytes)[] _confirmedTransactions)
func (_BridgeContract *BridgeContractCaller) GetConfirmedTransactions(opts *bind.CallOpts, _destinationChain uint8) ([]IBridgeStructsConfirmedTransaction, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getConfirmedTransactions", _destinationChain)

	if err != nil {
		return *new([]IBridgeStructsConfirmedTransaction), err
	}

	out0 := *abi.ConvertType(out[0], new([]IBridgeStructsConfirmedTransaction)).(*[]IBridgeStructsConfirmedTransaction)

	return out0, err

}

// GetConfirmedTransactions is a free data retrieval call binding the contract method 0x4cae8087.
//
// Solidity: function getConfirmedTransactions(uint8 _destinationChain) view returns((uint256,uint256,uint256,bytes32,uint64,uint8,uint8,(uint256,string)[],bytes)[] _confirmedTransactions)
func (_BridgeContract *BridgeContractSession) GetConfirmedTransactions(_destinationChain uint8) ([]IBridgeStructsConfirmedTransaction, error) {
	return _BridgeContract.Contract.GetConfirmedTransactions(&_BridgeContract.CallOpts, _destinationChain)
}

// GetConfirmedTransactions is a free data retrieval call binding the contract method 0x4cae8087.
//
// Solidity: function getConfirmedTransactions(uint8 _destinationChain) view returns((uint256,uint256,uint256,bytes32,uint64,uint8,uint8,(uint256,string)[],bytes)[] _confirmedTransactions)
func (_BridgeContract *BridgeContractCallerSession) GetConfirmedTransactions(_destinationChain uint8) ([]IBridgeStructsConfirmedTransaction, error) {
	return _BridgeContract.Contract.GetConfirmedTransactions(&_BridgeContract.CallOpts, _destinationChain)
}

// GetLastObservedBlock is a free data retrieval call binding the contract method 0xdf9a131f.
//
// Solidity: function getLastObservedBlock(uint8 _sourceChain) view returns((uint256,bytes32) _cblock)
func (_BridgeContract *BridgeContractCaller) GetLastObservedBlock(opts *bind.CallOpts, _sourceChain uint8) (IBridgeStructsCardanoBlock, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getLastObservedBlock", _sourceChain)

	if err != nil {
		return *new(IBridgeStructsCardanoBlock), err
	}

	out0 := *abi.ConvertType(out[0], new(IBridgeStructsCardanoBlock)).(*IBridgeStructsCardanoBlock)

	return out0, err

}

// GetLastObservedBlock is a free data retrieval call binding the contract method 0xdf9a131f.
//
// Solidity: function getLastObservedBlock(uint8 _sourceChain) view returns((uint256,bytes32) _cblock)
func (_BridgeContract *BridgeContractSession) GetLastObservedBlock(_sourceChain uint8) (IBridgeStructsCardanoBlock, error) {
	return _BridgeContract.Contract.GetLastObservedBlock(&_BridgeContract.CallOpts, _sourceChain)
}

// GetLastObservedBlock is a free data retrieval call binding the contract method 0xdf9a131f.
//
// Solidity: function getLastObservedBlock(uint8 _sourceChain) view returns((uint256,bytes32) _cblock)
func (_BridgeContract *BridgeContractCallerSession) GetLastObservedBlock(_sourceChain uint8) (IBridgeStructsCardanoBlock, error) {
	return _BridgeContract.Contract.GetLastObservedBlock(&_BridgeContract.CallOpts, _sourceChain)
}

// GetNextBatchId is a free data retrieval call binding the contract method 0x853609d6.
//
// Solidity: function getNextBatchId(uint8 _destinationChain) view returns(uint64 _result)
func (_BridgeContract *BridgeContractCaller) GetNextBatchId(opts *bind.CallOpts, _destinationChain uint8) (uint64, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getNextBatchId", _destinationChain)

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// GetNextBatchId is a free data retrieval call binding the contract method 0x853609d6.
//
// Solidity: function getNextBatchId(uint8 _destinationChain) view returns(uint64 _result)
func (_BridgeContract *BridgeContractSession) GetNextBatchId(_destinationChain uint8) (uint64, error) {
	return _BridgeContract.Contract.GetNextBatchId(&_BridgeContract.CallOpts, _destinationChain)
}

// GetNextBatchId is a free data retrieval call binding the contract method 0x853609d6.
//
// Solidity: function getNextBatchId(uint8 _destinationChain) view returns(uint64 _result)
func (_BridgeContract *BridgeContractCallerSession) GetNextBatchId(_destinationChain uint8) (uint64, error) {
	return _BridgeContract.Contract.GetNextBatchId(&_BridgeContract.CallOpts, _destinationChain)
}

// GetRawTransactionFromLastBatch is a free data retrieval call binding the contract method 0x9320dd41.
//
// Solidity: function getRawTransactionFromLastBatch(uint8 _destinationChain) view returns(bytes)
func (_BridgeContract *BridgeContractCaller) GetRawTransactionFromLastBatch(opts *bind.CallOpts, _destinationChain uint8) ([]byte, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getRawTransactionFromLastBatch", _destinationChain)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetRawTransactionFromLastBatch is a free data retrieval call binding the contract method 0x9320dd41.
//
// Solidity: function getRawTransactionFromLastBatch(uint8 _destinationChain) view returns(bytes)
func (_BridgeContract *BridgeContractSession) GetRawTransactionFromLastBatch(_destinationChain uint8) ([]byte, error) {
	return _BridgeContract.Contract.GetRawTransactionFromLastBatch(&_BridgeContract.CallOpts, _destinationChain)
}

// GetRawTransactionFromLastBatch is a free data retrieval call binding the contract method 0x9320dd41.
//
// Solidity: function getRawTransactionFromLastBatch(uint8 _destinationChain) view returns(bytes)
func (_BridgeContract *BridgeContractCallerSession) GetRawTransactionFromLastBatch(_destinationChain uint8) ([]byte, error) {
	return _BridgeContract.Contract.GetRawTransactionFromLastBatch(&_BridgeContract.CallOpts, _destinationChain)
}

// GetValidatorsChainData is a free data retrieval call binding the contract method 0x0141eafc.
//
// Solidity: function getValidatorsChainData(uint8 _chainId) view returns((uint256[4])[])
func (_BridgeContract *BridgeContractCaller) GetValidatorsChainData(opts *bind.CallOpts, _chainId uint8) ([]IBridgeStructsValidatorChainData, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "getValidatorsChainData", _chainId)

	if err != nil {
		return *new([]IBridgeStructsValidatorChainData), err
	}

	out0 := *abi.ConvertType(out[0], new([]IBridgeStructsValidatorChainData)).(*[]IBridgeStructsValidatorChainData)

	return out0, err

}

// GetValidatorsChainData is a free data retrieval call binding the contract method 0x0141eafc.
//
// Solidity: function getValidatorsChainData(uint8 _chainId) view returns((uint256[4])[])
func (_BridgeContract *BridgeContractSession) GetValidatorsChainData(_chainId uint8) ([]IBridgeStructsValidatorChainData, error) {
	return _BridgeContract.Contract.GetValidatorsChainData(&_BridgeContract.CallOpts, _chainId)
}

// GetValidatorsChainData is a free data retrieval call binding the contract method 0x0141eafc.
//
// Solidity: function getValidatorsChainData(uint8 _chainId) view returns((uint256[4])[])
func (_BridgeContract *BridgeContractCallerSession) GetValidatorsChainData(_chainId uint8) ([]IBridgeStructsValidatorChainData, error) {
	return _BridgeContract.Contract.GetValidatorsChainData(&_BridgeContract.CallOpts, _chainId)
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

// ShouldCreateBatch is a free data retrieval call binding the contract method 0x1dd28495.
//
// Solidity: function shouldCreateBatch(uint8 _destinationChain) view returns(bool _batch)
func (_BridgeContract *BridgeContractCaller) ShouldCreateBatch(opts *bind.CallOpts, _destinationChain uint8) (bool, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "shouldCreateBatch", _destinationChain)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// ShouldCreateBatch is a free data retrieval call binding the contract method 0x1dd28495.
//
// Solidity: function shouldCreateBatch(uint8 _destinationChain) view returns(bool _batch)
func (_BridgeContract *BridgeContractSession) ShouldCreateBatch(_destinationChain uint8) (bool, error) {
	return _BridgeContract.Contract.ShouldCreateBatch(&_BridgeContract.CallOpts, _destinationChain)
}

// ShouldCreateBatch is a free data retrieval call binding the contract method 0x1dd28495.
//
// Solidity: function shouldCreateBatch(uint8 _destinationChain) view returns(bool _batch)
func (_BridgeContract *BridgeContractCallerSession) ShouldCreateBatch(_destinationChain uint8) (bool, error) {
	return _BridgeContract.Contract.ShouldCreateBatch(&_BridgeContract.CallOpts, _destinationChain)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address _owner, address _upgradeAdmin) returns()
func (_BridgeContract *BridgeContractTransactor) Initialize(opts *bind.TransactOpts, _owner common.Address, _upgradeAdmin common.Address) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "initialize", _owner, _upgradeAdmin)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address _owner, address _upgradeAdmin) returns()
func (_BridgeContract *BridgeContractSession) Initialize(_owner common.Address, _upgradeAdmin common.Address) (*types.Transaction, error) {
	return _BridgeContract.Contract.Initialize(&_BridgeContract.TransactOpts, _owner, _upgradeAdmin)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address _owner, address _upgradeAdmin) returns()
func (_BridgeContract *BridgeContractTransactorSession) Initialize(_owner common.Address, _upgradeAdmin common.Address) (*types.Transaction, error) {
	return _BridgeContract.Contract.Initialize(&_BridgeContract.TransactOpts, _owner, _upgradeAdmin)
}

// RegisterChain is a paid mutator transaction binding the contract method 0xdb8f522e.
//
// Solidity: function registerChain((uint8,uint8,string,string) _chain, uint256 _tokenQuantity, (address,(uint256[4]))[] _chainData) returns()
func (_BridgeContract *BridgeContractTransactor) RegisterChain(opts *bind.TransactOpts, _chain IBridgeStructsChain, _tokenQuantity *big.Int, _chainData []IBridgeStructsValidatorAddressChainData) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "registerChain", _chain, _tokenQuantity, _chainData)
}

// RegisterChain is a paid mutator transaction binding the contract method 0xdb8f522e.
//
// Solidity: function registerChain((uint8,uint8,string,string) _chain, uint256 _tokenQuantity, (address,(uint256[4]))[] _chainData) returns()
func (_BridgeContract *BridgeContractSession) RegisterChain(_chain IBridgeStructsChain, _tokenQuantity *big.Int, _chainData []IBridgeStructsValidatorAddressChainData) (*types.Transaction, error) {
	return _BridgeContract.Contract.RegisterChain(&_BridgeContract.TransactOpts, _chain, _tokenQuantity, _chainData)
}

// RegisterChain is a paid mutator transaction binding the contract method 0xdb8f522e.
//
// Solidity: function registerChain((uint8,uint8,string,string) _chain, uint256 _tokenQuantity, (address,(uint256[4]))[] _chainData) returns()
func (_BridgeContract *BridgeContractTransactorSession) RegisterChain(_chain IBridgeStructsChain, _tokenQuantity *big.Int, _chainData []IBridgeStructsValidatorAddressChainData) (*types.Transaction, error) {
	return _BridgeContract.Contract.RegisterChain(&_BridgeContract.TransactOpts, _chain, _tokenQuantity, _chainData)
}

// RegisterChainGovernance is a paid mutator transaction binding the contract method 0x27f731e6.
//
// Solidity: function registerChainGovernance(uint8 _chainId, uint8 _chainType, uint256 _tokenQuantity, (uint256[4]) _validatorChainData, bytes _keySignature, bytes _keyFeeSignature) returns()
func (_BridgeContract *BridgeContractTransactor) RegisterChainGovernance(opts *bind.TransactOpts, _chainId uint8, _chainType uint8, _tokenQuantity *big.Int, _validatorChainData IBridgeStructsValidatorChainData, _keySignature []byte, _keyFeeSignature []byte) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "registerChainGovernance", _chainId, _chainType, _tokenQuantity, _validatorChainData, _keySignature, _keyFeeSignature)
}

// RegisterChainGovernance is a paid mutator transaction binding the contract method 0x27f731e6.
//
// Solidity: function registerChainGovernance(uint8 _chainId, uint8 _chainType, uint256 _tokenQuantity, (uint256[4]) _validatorChainData, bytes _keySignature, bytes _keyFeeSignature) returns()
func (_BridgeContract *BridgeContractSession) RegisterChainGovernance(_chainId uint8, _chainType uint8, _tokenQuantity *big.Int, _validatorChainData IBridgeStructsValidatorChainData, _keySignature []byte, _keyFeeSignature []byte) (*types.Transaction, error) {
	return _BridgeContract.Contract.RegisterChainGovernance(&_BridgeContract.TransactOpts, _chainId, _chainType, _tokenQuantity, _validatorChainData, _keySignature, _keyFeeSignature)
}

// RegisterChainGovernance is a paid mutator transaction binding the contract method 0x27f731e6.
//
// Solidity: function registerChainGovernance(uint8 _chainId, uint8 _chainType, uint256 _tokenQuantity, (uint256[4]) _validatorChainData, bytes _keySignature, bytes _keyFeeSignature) returns()
func (_BridgeContract *BridgeContractTransactorSession) RegisterChainGovernance(_chainId uint8, _chainType uint8, _tokenQuantity *big.Int, _validatorChainData IBridgeStructsValidatorChainData, _keySignature []byte, _keyFeeSignature []byte) (*types.Transaction, error) {
	return _BridgeContract.Contract.RegisterChainGovernance(&_BridgeContract.TransactOpts, _chainId, _chainType, _tokenQuantity, _validatorChainData, _keySignature, _keyFeeSignature)
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

// SetChainAdditionalData is a paid mutator transaction binding the contract method 0x8e297e82.
//
// Solidity: function setChainAdditionalData(uint8 _chainId, string addressMultisig, string addressFeePayer) returns()
func (_BridgeContract *BridgeContractTransactor) SetChainAdditionalData(opts *bind.TransactOpts, _chainId uint8, addressMultisig string, addressFeePayer string) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "setChainAdditionalData", _chainId, addressMultisig, addressFeePayer)
}

// SetChainAdditionalData is a paid mutator transaction binding the contract method 0x8e297e82.
//
// Solidity: function setChainAdditionalData(uint8 _chainId, string addressMultisig, string addressFeePayer) returns()
func (_BridgeContract *BridgeContractSession) SetChainAdditionalData(_chainId uint8, addressMultisig string, addressFeePayer string) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetChainAdditionalData(&_BridgeContract.TransactOpts, _chainId, addressMultisig, addressFeePayer)
}

// SetChainAdditionalData is a paid mutator transaction binding the contract method 0x8e297e82.
//
// Solidity: function setChainAdditionalData(uint8 _chainId, string addressMultisig, string addressFeePayer) returns()
func (_BridgeContract *BridgeContractTransactorSession) SetChainAdditionalData(_chainId uint8, addressMultisig string, addressFeePayer string) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetChainAdditionalData(&_BridgeContract.TransactOpts, _chainId, addressMultisig, addressFeePayer)
}

// SetDependencies is a paid mutator transaction binding the contract method 0x3bde7d2e.
//
// Solidity: function setDependencies(address _claimsAddress, address _signedBatchesAddress, address _slotsAddress, address _validatorsAddress) returns()
func (_BridgeContract *BridgeContractTransactor) SetDependencies(opts *bind.TransactOpts, _claimsAddress common.Address, _signedBatchesAddress common.Address, _slotsAddress common.Address, _validatorsAddress common.Address) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "setDependencies", _claimsAddress, _signedBatchesAddress, _slotsAddress, _validatorsAddress)
}

// SetDependencies is a paid mutator transaction binding the contract method 0x3bde7d2e.
//
// Solidity: function setDependencies(address _claimsAddress, address _signedBatchesAddress, address _slotsAddress, address _validatorsAddress) returns()
func (_BridgeContract *BridgeContractSession) SetDependencies(_claimsAddress common.Address, _signedBatchesAddress common.Address, _slotsAddress common.Address, _validatorsAddress common.Address) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetDependencies(&_BridgeContract.TransactOpts, _claimsAddress, _signedBatchesAddress, _slotsAddress, _validatorsAddress)
}

// SetDependencies is a paid mutator transaction binding the contract method 0x3bde7d2e.
//
// Solidity: function setDependencies(address _claimsAddress, address _signedBatchesAddress, address _slotsAddress, address _validatorsAddress) returns()
func (_BridgeContract *BridgeContractTransactorSession) SetDependencies(_claimsAddress common.Address, _signedBatchesAddress common.Address, _slotsAddress common.Address, _validatorsAddress common.Address) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetDependencies(&_BridgeContract.TransactOpts, _claimsAddress, _signedBatchesAddress, _slotsAddress, _validatorsAddress)
}

// SubmitClaims is a paid mutator transaction binding the contract method 0xfdb2d530.
//
// Solidity: function submitClaims(((bytes32,(uint256,string)[],uint256,uint256,uint8,uint8)[],(bytes32,uint64,uint8)[],(bytes32,uint64,uint8)[],(bytes32,bytes32,uint256,bytes,string,uint64,uint8,bool)[],(uint8,uint256)[]) _claims) returns()
func (_BridgeContract *BridgeContractTransactor) SubmitClaims(opts *bind.TransactOpts, _claims IBridgeStructsValidatorClaims) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "submitClaims", _claims)
}

// SubmitClaims is a paid mutator transaction binding the contract method 0xfdb2d530.
//
// Solidity: function submitClaims(((bytes32,(uint256,string)[],uint256,uint256,uint8,uint8)[],(bytes32,uint64,uint8)[],(bytes32,uint64,uint8)[],(bytes32,bytes32,uint256,bytes,string,uint64,uint8,bool)[],(uint8,uint256)[]) _claims) returns()
func (_BridgeContract *BridgeContractSession) SubmitClaims(_claims IBridgeStructsValidatorClaims) (*types.Transaction, error) {
	return _BridgeContract.Contract.SubmitClaims(&_BridgeContract.TransactOpts, _claims)
}

// SubmitClaims is a paid mutator transaction binding the contract method 0xfdb2d530.
//
// Solidity: function submitClaims(((bytes32,(uint256,string)[],uint256,uint256,uint8,uint8)[],(bytes32,uint64,uint8)[],(bytes32,uint64,uint8)[],(bytes32,bytes32,uint256,bytes,string,uint64,uint8,bool)[],(uint8,uint256)[]) _claims) returns()
func (_BridgeContract *BridgeContractTransactorSession) SubmitClaims(_claims IBridgeStructsValidatorClaims) (*types.Transaction, error) {
	return _BridgeContract.Contract.SubmitClaims(&_BridgeContract.TransactOpts, _claims)
}

// SubmitLastObservedBlocks is a paid mutator transaction binding the contract method 0x0019a66e.
//
// Solidity: function submitLastObservedBlocks(uint8 _chainId, (uint256,bytes32)[] _blocks) returns()
func (_BridgeContract *BridgeContractTransactor) SubmitLastObservedBlocks(opts *bind.TransactOpts, _chainId uint8, _blocks []IBridgeStructsCardanoBlock) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "submitLastObservedBlocks", _chainId, _blocks)
}

// SubmitLastObservedBlocks is a paid mutator transaction binding the contract method 0x0019a66e.
//
// Solidity: function submitLastObservedBlocks(uint8 _chainId, (uint256,bytes32)[] _blocks) returns()
func (_BridgeContract *BridgeContractSession) SubmitLastObservedBlocks(_chainId uint8, _blocks []IBridgeStructsCardanoBlock) (*types.Transaction, error) {
	return _BridgeContract.Contract.SubmitLastObservedBlocks(&_BridgeContract.TransactOpts, _chainId, _blocks)
}

// SubmitLastObservedBlocks is a paid mutator transaction binding the contract method 0x0019a66e.
//
// Solidity: function submitLastObservedBlocks(uint8 _chainId, (uint256,bytes32)[] _blocks) returns()
func (_BridgeContract *BridgeContractTransactorSession) SubmitLastObservedBlocks(_chainId uint8, _blocks []IBridgeStructsCardanoBlock) (*types.Transaction, error) {
	return _BridgeContract.Contract.SubmitLastObservedBlocks(&_BridgeContract.TransactOpts, _chainId, _blocks)
}

// SubmitSignedBatch is a paid mutator transaction binding the contract method 0x56a28c27.
//
// Solidity: function submitSignedBatch((uint64,uint64,uint64,uint8,bytes,bytes,bytes,bool) _signedBatch) returns()
func (_BridgeContract *BridgeContractTransactor) SubmitSignedBatch(opts *bind.TransactOpts, _signedBatch IBridgeStructsSignedBatch) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "submitSignedBatch", _signedBatch)
}

// SubmitSignedBatch is a paid mutator transaction binding the contract method 0x56a28c27.
//
// Solidity: function submitSignedBatch((uint64,uint64,uint64,uint8,bytes,bytes,bytes,bool) _signedBatch) returns()
func (_BridgeContract *BridgeContractSession) SubmitSignedBatch(_signedBatch IBridgeStructsSignedBatch) (*types.Transaction, error) {
	return _BridgeContract.Contract.SubmitSignedBatch(&_BridgeContract.TransactOpts, _signedBatch)
}

// SubmitSignedBatch is a paid mutator transaction binding the contract method 0x56a28c27.
//
// Solidity: function submitSignedBatch((uint64,uint64,uint64,uint8,bytes,bytes,bytes,bool) _signedBatch) returns()
func (_BridgeContract *BridgeContractTransactorSession) SubmitSignedBatch(_signedBatch IBridgeStructsSignedBatch) (*types.Transaction, error) {
	return _BridgeContract.Contract.SubmitSignedBatch(&_BridgeContract.TransactOpts, _signedBatch)
}

// SubmitSignedBatchEVM is a paid mutator transaction binding the contract method 0x15328e62.
//
// Solidity: function submitSignedBatchEVM((uint64,uint64,uint64,uint8,bytes,bytes,bytes,bool) _signedBatch) returns()
func (_BridgeContract *BridgeContractTransactor) SubmitSignedBatchEVM(opts *bind.TransactOpts, _signedBatch IBridgeStructsSignedBatch) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "submitSignedBatchEVM", _signedBatch)
}

// SubmitSignedBatchEVM is a paid mutator transaction binding the contract method 0x15328e62.
//
// Solidity: function submitSignedBatchEVM((uint64,uint64,uint64,uint8,bytes,bytes,bytes,bool) _signedBatch) returns()
func (_BridgeContract *BridgeContractSession) SubmitSignedBatchEVM(_signedBatch IBridgeStructsSignedBatch) (*types.Transaction, error) {
	return _BridgeContract.Contract.SubmitSignedBatchEVM(&_BridgeContract.TransactOpts, _signedBatch)
}

// SubmitSignedBatchEVM is a paid mutator transaction binding the contract method 0x15328e62.
//
// Solidity: function submitSignedBatchEVM((uint64,uint64,uint64,uint8,bytes,bytes,bytes,bool) _signedBatch) returns()
func (_BridgeContract *BridgeContractTransactorSession) SubmitSignedBatchEVM(_signedBatch IBridgeStructsSignedBatch) (*types.Transaction, error) {
	return _BridgeContract.Contract.SubmitSignedBatchEVM(&_BridgeContract.TransactOpts, _signedBatch)
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

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_BridgeContract *BridgeContractTransactor) UpgradeTo(opts *bind.TransactOpts, newImplementation common.Address) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "upgradeTo", newImplementation)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_BridgeContract *BridgeContractSession) UpgradeTo(newImplementation common.Address) (*types.Transaction, error) {
	return _BridgeContract.Contract.UpgradeTo(&_BridgeContract.TransactOpts, newImplementation)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_BridgeContract *BridgeContractTransactorSession) UpgradeTo(newImplementation common.Address) (*types.Transaction, error) {
	return _BridgeContract.Contract.UpgradeTo(&_BridgeContract.TransactOpts, newImplementation)
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

// BridgeContractAdminChangedIterator is returned from FilterAdminChanged and is used to iterate over the raw logs and unpacked data for AdminChanged events raised by the BridgeContract contract.
type BridgeContractAdminChangedIterator struct {
	Event *BridgeContractAdminChanged // Event containing the contract specifics and raw log

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
func (it *BridgeContractAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractAdminChanged)
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
		it.Event = new(BridgeContractAdminChanged)
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
func (it *BridgeContractAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractAdminChanged represents a AdminChanged event raised by the BridgeContract contract.
type BridgeContractAdminChanged struct {
	PreviousAdmin common.Address
	NewAdmin      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterAdminChanged is a free log retrieval operation binding the contract event 0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f.
//
// Solidity: event AdminChanged(address previousAdmin, address newAdmin)
func (_BridgeContract *BridgeContractFilterer) FilterAdminChanged(opts *bind.FilterOpts) (*BridgeContractAdminChangedIterator, error) {

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "AdminChanged")
	if err != nil {
		return nil, err
	}
	return &BridgeContractAdminChangedIterator{contract: _BridgeContract.contract, event: "AdminChanged", logs: logs, sub: sub}, nil
}

// WatchAdminChanged is a free log subscription operation binding the contract event 0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f.
//
// Solidity: event AdminChanged(address previousAdmin, address newAdmin)
func (_BridgeContract *BridgeContractFilterer) WatchAdminChanged(opts *bind.WatchOpts, sink chan<- *BridgeContractAdminChanged) (event.Subscription, error) {

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "AdminChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractAdminChanged)
				if err := _BridgeContract.contract.UnpackLog(event, "AdminChanged", log); err != nil {
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

// ParseAdminChanged is a log parse operation binding the contract event 0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f.
//
// Solidity: event AdminChanged(address previousAdmin, address newAdmin)
func (_BridgeContract *BridgeContractFilterer) ParseAdminChanged(log types.Log) (*BridgeContractAdminChanged, error) {
	event := new(BridgeContractAdminChanged)
	if err := _BridgeContract.contract.UnpackLog(event, "AdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeContractBeaconUpgradedIterator is returned from FilterBeaconUpgraded and is used to iterate over the raw logs and unpacked data for BeaconUpgraded events raised by the BridgeContract contract.
type BridgeContractBeaconUpgradedIterator struct {
	Event *BridgeContractBeaconUpgraded // Event containing the contract specifics and raw log

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
func (it *BridgeContractBeaconUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractBeaconUpgraded)
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
		it.Event = new(BridgeContractBeaconUpgraded)
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
func (it *BridgeContractBeaconUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractBeaconUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractBeaconUpgraded represents a BeaconUpgraded event raised by the BridgeContract contract.
type BridgeContractBeaconUpgraded struct {
	Beacon common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterBeaconUpgraded is a free log retrieval operation binding the contract event 0x1cf3b03a6cf19fa2baba4df148e9dcabedea7f8a5c07840e207e5c089be95d3e.
//
// Solidity: event BeaconUpgraded(address indexed beacon)
func (_BridgeContract *BridgeContractFilterer) FilterBeaconUpgraded(opts *bind.FilterOpts, beacon []common.Address) (*BridgeContractBeaconUpgradedIterator, error) {

	var beaconRule []interface{}
	for _, beaconItem := range beacon {
		beaconRule = append(beaconRule, beaconItem)
	}

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "BeaconUpgraded", beaconRule)
	if err != nil {
		return nil, err
	}
	return &BridgeContractBeaconUpgradedIterator{contract: _BridgeContract.contract, event: "BeaconUpgraded", logs: logs, sub: sub}, nil
}

// WatchBeaconUpgraded is a free log subscription operation binding the contract event 0x1cf3b03a6cf19fa2baba4df148e9dcabedea7f8a5c07840e207e5c089be95d3e.
//
// Solidity: event BeaconUpgraded(address indexed beacon)
func (_BridgeContract *BridgeContractFilterer) WatchBeaconUpgraded(opts *bind.WatchOpts, sink chan<- *BridgeContractBeaconUpgraded, beacon []common.Address) (event.Subscription, error) {

	var beaconRule []interface{}
	for _, beaconItem := range beacon {
		beaconRule = append(beaconRule, beaconItem)
	}

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "BeaconUpgraded", beaconRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractBeaconUpgraded)
				if err := _BridgeContract.contract.UnpackLog(event, "BeaconUpgraded", log); err != nil {
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

// ParseBeaconUpgraded is a log parse operation binding the contract event 0x1cf3b03a6cf19fa2baba4df148e9dcabedea7f8a5c07840e207e5c089be95d3e.
//
// Solidity: event BeaconUpgraded(address indexed beacon)
func (_BridgeContract *BridgeContractFilterer) ParseBeaconUpgraded(log types.Log) (*BridgeContractBeaconUpgraded, error) {
	event := new(BridgeContractBeaconUpgraded)
	if err := _BridgeContract.contract.UnpackLog(event, "BeaconUpgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeContractChainDefundedIterator is returned from FilterChainDefunded and is used to iterate over the raw logs and unpacked data for ChainDefunded events raised by the BridgeContract contract.
type BridgeContractChainDefundedIterator struct {
	Event *BridgeContractChainDefunded // Event containing the contract specifics and raw log

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
func (it *BridgeContractChainDefundedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractChainDefunded)
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
		it.Event = new(BridgeContractChainDefunded)
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
func (it *BridgeContractChainDefundedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractChainDefundedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractChainDefunded represents a ChainDefunded event raised by the BridgeContract contract.
type BridgeContractChainDefunded struct {
	ChainId uint8
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterChainDefunded is a free log retrieval operation binding the contract event 0xce8ca6f0aead771734dd729750170b73fbb3147d90cd4273c1bbcfe669bbf69d.
//
// Solidity: event ChainDefunded(uint8 _chainId, uint256 _amount)
func (_BridgeContract *BridgeContractFilterer) FilterChainDefunded(opts *bind.FilterOpts) (*BridgeContractChainDefundedIterator, error) {

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "ChainDefunded")
	if err != nil {
		return nil, err
	}
	return &BridgeContractChainDefundedIterator{contract: _BridgeContract.contract, event: "ChainDefunded", logs: logs, sub: sub}, nil
}

// WatchChainDefunded is a free log subscription operation binding the contract event 0xce8ca6f0aead771734dd729750170b73fbb3147d90cd4273c1bbcfe669bbf69d.
//
// Solidity: event ChainDefunded(uint8 _chainId, uint256 _amount)
func (_BridgeContract *BridgeContractFilterer) WatchChainDefunded(opts *bind.WatchOpts, sink chan<- *BridgeContractChainDefunded) (event.Subscription, error) {

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "ChainDefunded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractChainDefunded)
				if err := _BridgeContract.contract.UnpackLog(event, "ChainDefunded", log); err != nil {
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

// ParseChainDefunded is a log parse operation binding the contract event 0xce8ca6f0aead771734dd729750170b73fbb3147d90cd4273c1bbcfe669bbf69d.
//
// Solidity: event ChainDefunded(uint8 _chainId, uint256 _amount)
func (_BridgeContract *BridgeContractFilterer) ParseChainDefunded(log types.Log) (*BridgeContractChainDefunded, error) {
	event := new(BridgeContractChainDefunded)
	if err := _BridgeContract.contract.UnpackLog(event, "ChainDefunded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeContractDefundFailedAfterMultipleRetriesIterator is returned from FilterDefundFailedAfterMultipleRetries and is used to iterate over the raw logs and unpacked data for DefundFailedAfterMultipleRetries events raised by the BridgeContract contract.
type BridgeContractDefundFailedAfterMultipleRetriesIterator struct {
	Event *BridgeContractDefundFailedAfterMultipleRetries // Event containing the contract specifics and raw log

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
func (it *BridgeContractDefundFailedAfterMultipleRetriesIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractDefundFailedAfterMultipleRetries)
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
		it.Event = new(BridgeContractDefundFailedAfterMultipleRetries)
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
func (it *BridgeContractDefundFailedAfterMultipleRetriesIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractDefundFailedAfterMultipleRetriesIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractDefundFailedAfterMultipleRetries represents a DefundFailedAfterMultipleRetries event raised by the BridgeContract contract.
type BridgeContractDefundFailedAfterMultipleRetries struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterDefundFailedAfterMultipleRetries is a free log retrieval operation binding the contract event 0xc53812079c8257d7e3e68b2e9d937f825484546404c4bbdfa79785d213976c6f.
//
// Solidity: event DefundFailedAfterMultipleRetries()
func (_BridgeContract *BridgeContractFilterer) FilterDefundFailedAfterMultipleRetries(opts *bind.FilterOpts) (*BridgeContractDefundFailedAfterMultipleRetriesIterator, error) {

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "DefundFailedAfterMultipleRetries")
	if err != nil {
		return nil, err
	}
	return &BridgeContractDefundFailedAfterMultipleRetriesIterator{contract: _BridgeContract.contract, event: "DefundFailedAfterMultipleRetries", logs: logs, sub: sub}, nil
}

// WatchDefundFailedAfterMultipleRetries is a free log subscription operation binding the contract event 0xc53812079c8257d7e3e68b2e9d937f825484546404c4bbdfa79785d213976c6f.
//
// Solidity: event DefundFailedAfterMultipleRetries()
func (_BridgeContract *BridgeContractFilterer) WatchDefundFailedAfterMultipleRetries(opts *bind.WatchOpts, sink chan<- *BridgeContractDefundFailedAfterMultipleRetries) (event.Subscription, error) {

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "DefundFailedAfterMultipleRetries")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractDefundFailedAfterMultipleRetries)
				if err := _BridgeContract.contract.UnpackLog(event, "DefundFailedAfterMultipleRetries", log); err != nil {
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

// ParseDefundFailedAfterMultipleRetries is a log parse operation binding the contract event 0xc53812079c8257d7e3e68b2e9d937f825484546404c4bbdfa79785d213976c6f.
//
// Solidity: event DefundFailedAfterMultipleRetries()
func (_BridgeContract *BridgeContractFilterer) ParseDefundFailedAfterMultipleRetries(log types.Log) (*BridgeContractDefundFailedAfterMultipleRetries, error) {
	event := new(BridgeContractDefundFailedAfterMultipleRetries)
	if err := _BridgeContract.contract.UnpackLog(event, "DefundFailedAfterMultipleRetries", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeContractFundAdminChangedIterator is returned from FilterFundAdminChanged and is used to iterate over the raw logs and unpacked data for FundAdminChanged events raised by the BridgeContract contract.
type BridgeContractFundAdminChangedIterator struct {
	Event *BridgeContractFundAdminChanged // Event containing the contract specifics and raw log

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
func (it *BridgeContractFundAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractFundAdminChanged)
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
		it.Event = new(BridgeContractFundAdminChanged)
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
func (it *BridgeContractFundAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractFundAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractFundAdminChanged represents a FundAdminChanged event raised by the BridgeContract contract.
type BridgeContractFundAdminChanged struct {
	NewFundAdmin common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterFundAdminChanged is a free log retrieval operation binding the contract event 0x217da1af647045bf1a50dc62120037c19f6963e96ae6bea7964bf241770fbe00.
//
// Solidity: event FundAdminChanged(address _newFundAdmin)
func (_BridgeContract *BridgeContractFilterer) FilterFundAdminChanged(opts *bind.FilterOpts) (*BridgeContractFundAdminChangedIterator, error) {

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "FundAdminChanged")
	if err != nil {
		return nil, err
	}
	return &BridgeContractFundAdminChangedIterator{contract: _BridgeContract.contract, event: "FundAdminChanged", logs: logs, sub: sub}, nil
}

// WatchFundAdminChanged is a free log subscription operation binding the contract event 0x217da1af647045bf1a50dc62120037c19f6963e96ae6bea7964bf241770fbe00.
//
// Solidity: event FundAdminChanged(address _newFundAdmin)
func (_BridgeContract *BridgeContractFilterer) WatchFundAdminChanged(opts *bind.WatchOpts, sink chan<- *BridgeContractFundAdminChanged) (event.Subscription, error) {

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "FundAdminChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractFundAdminChanged)
				if err := _BridgeContract.contract.UnpackLog(event, "FundAdminChanged", log); err != nil {
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

// ParseFundAdminChanged is a log parse operation binding the contract event 0x217da1af647045bf1a50dc62120037c19f6963e96ae6bea7964bf241770fbe00.
//
// Solidity: event FundAdminChanged(address _newFundAdmin)
func (_BridgeContract *BridgeContractFilterer) ParseFundAdminChanged(log types.Log) (*BridgeContractFundAdminChanged, error) {
	event := new(BridgeContractFundAdminChanged)
	if err := _BridgeContract.contract.UnpackLog(event, "FundAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
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
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_BridgeContract *BridgeContractFilterer) FilterInitialized(opts *bind.FilterOpts) (*BridgeContractInitializedIterator, error) {

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &BridgeContractInitializedIterator{contract: _BridgeContract.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
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

// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_BridgeContract *BridgeContractFilterer) ParseInitialized(log types.Log) (*BridgeContractInitialized, error) {
	event := new(BridgeContractInitialized)
	if err := _BridgeContract.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeContractNotEnoughFundsIterator is returned from FilterNotEnoughFunds and is used to iterate over the raw logs and unpacked data for NotEnoughFunds events raised by the BridgeContract contract.
type BridgeContractNotEnoughFundsIterator struct {
	Event *BridgeContractNotEnoughFunds // Event containing the contract specifics and raw log

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
func (it *BridgeContractNotEnoughFundsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractNotEnoughFunds)
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
		it.Event = new(BridgeContractNotEnoughFunds)
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
func (it *BridgeContractNotEnoughFundsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractNotEnoughFundsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractNotEnoughFunds represents a NotEnoughFunds event raised by the BridgeContract contract.
type BridgeContractNotEnoughFunds struct {
	ClaimeType      string
	Index           *big.Int
	AvailableAmount *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterNotEnoughFunds is a free log retrieval operation binding the contract event 0xfc533f3a2de16e1c5a6c03094c865069fabdd71ee6a9af63918823a35a55def5.
//
// Solidity: event NotEnoughFunds(string claimeType, uint256 index, uint256 availableAmount)
func (_BridgeContract *BridgeContractFilterer) FilterNotEnoughFunds(opts *bind.FilterOpts) (*BridgeContractNotEnoughFundsIterator, error) {

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "NotEnoughFunds")
	if err != nil {
		return nil, err
	}
	return &BridgeContractNotEnoughFundsIterator{contract: _BridgeContract.contract, event: "NotEnoughFunds", logs: logs, sub: sub}, nil
}

// WatchNotEnoughFunds is a free log subscription operation binding the contract event 0xfc533f3a2de16e1c5a6c03094c865069fabdd71ee6a9af63918823a35a55def5.
//
// Solidity: event NotEnoughFunds(string claimeType, uint256 index, uint256 availableAmount)
func (_BridgeContract *BridgeContractFilterer) WatchNotEnoughFunds(opts *bind.WatchOpts, sink chan<- *BridgeContractNotEnoughFunds) (event.Subscription, error) {

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "NotEnoughFunds")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractNotEnoughFunds)
				if err := _BridgeContract.contract.UnpackLog(event, "NotEnoughFunds", log); err != nil {
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

// ParseNotEnoughFunds is a log parse operation binding the contract event 0xfc533f3a2de16e1c5a6c03094c865069fabdd71ee6a9af63918823a35a55def5.
//
// Solidity: event NotEnoughFunds(string claimeType, uint256 index, uint256 availableAmount)
func (_BridgeContract *BridgeContractFilterer) ParseNotEnoughFunds(log types.Log) (*BridgeContractNotEnoughFunds, error) {
	event := new(BridgeContractNotEnoughFunds)
	if err := _BridgeContract.contract.UnpackLog(event, "NotEnoughFunds", log); err != nil {
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

// BridgeContractUpdatedChainTokenQuantityIterator is returned from FilterUpdatedChainTokenQuantity and is used to iterate over the raw logs and unpacked data for UpdatedChainTokenQuantity events raised by the BridgeContract contract.
type BridgeContractUpdatedChainTokenQuantityIterator struct {
	Event *BridgeContractUpdatedChainTokenQuantity // Event containing the contract specifics and raw log

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
func (it *BridgeContractUpdatedChainTokenQuantityIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractUpdatedChainTokenQuantity)
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
		it.Event = new(BridgeContractUpdatedChainTokenQuantity)
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
func (it *BridgeContractUpdatedChainTokenQuantityIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractUpdatedChainTokenQuantityIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractUpdatedChainTokenQuantity represents a UpdatedChainTokenQuantity event raised by the BridgeContract contract.
type BridgeContractUpdatedChainTokenQuantity struct {
	ChainId       *big.Int
	IsIncrement   bool
	TokenQuantity *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterUpdatedChainTokenQuantity is a free log retrieval operation binding the contract event 0x63a8310d54d22ae170c9c99ec0101494848847baf8ba54b1f297456f4c01bd62.
//
// Solidity: event UpdatedChainTokenQuantity(uint256 indexed chainId, bool isIncrement, uint256 tokenQuantity)
func (_BridgeContract *BridgeContractFilterer) FilterUpdatedChainTokenQuantity(opts *bind.FilterOpts, chainId []*big.Int) (*BridgeContractUpdatedChainTokenQuantityIterator, error) {

	var chainIdRule []interface{}
	for _, chainIdItem := range chainId {
		chainIdRule = append(chainIdRule, chainIdItem)
	}

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "UpdatedChainTokenQuantity", chainIdRule)
	if err != nil {
		return nil, err
	}
	return &BridgeContractUpdatedChainTokenQuantityIterator{contract: _BridgeContract.contract, event: "UpdatedChainTokenQuantity", logs: logs, sub: sub}, nil
}

// WatchUpdatedChainTokenQuantity is a free log subscription operation binding the contract event 0x63a8310d54d22ae170c9c99ec0101494848847baf8ba54b1f297456f4c01bd62.
//
// Solidity: event UpdatedChainTokenQuantity(uint256 indexed chainId, bool isIncrement, uint256 tokenQuantity)
func (_BridgeContract *BridgeContractFilterer) WatchUpdatedChainTokenQuantity(opts *bind.WatchOpts, sink chan<- *BridgeContractUpdatedChainTokenQuantity, chainId []*big.Int) (event.Subscription, error) {

	var chainIdRule []interface{}
	for _, chainIdItem := range chainId {
		chainIdRule = append(chainIdRule, chainIdItem)
	}

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "UpdatedChainTokenQuantity", chainIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractUpdatedChainTokenQuantity)
				if err := _BridgeContract.contract.UnpackLog(event, "UpdatedChainTokenQuantity", log); err != nil {
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

// ParseUpdatedChainTokenQuantity is a log parse operation binding the contract event 0x63a8310d54d22ae170c9c99ec0101494848847baf8ba54b1f297456f4c01bd62.
//
// Solidity: event UpdatedChainTokenQuantity(uint256 indexed chainId, bool isIncrement, uint256 tokenQuantity)
func (_BridgeContract *BridgeContractFilterer) ParseUpdatedChainTokenQuantity(log types.Log) (*BridgeContractUpdatedChainTokenQuantity, error) {
	event := new(BridgeContractUpdatedChainTokenQuantity)
	if err := _BridgeContract.contract.UnpackLog(event, "UpdatedChainTokenQuantity", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeContractUpdatedMaxNumberOfTransactionsIterator is returned from FilterUpdatedMaxNumberOfTransactions and is used to iterate over the raw logs and unpacked data for UpdatedMaxNumberOfTransactions events raised by the BridgeContract contract.
type BridgeContractUpdatedMaxNumberOfTransactionsIterator struct {
	Event *BridgeContractUpdatedMaxNumberOfTransactions // Event containing the contract specifics and raw log

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
func (it *BridgeContractUpdatedMaxNumberOfTransactionsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractUpdatedMaxNumberOfTransactions)
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
		it.Event = new(BridgeContractUpdatedMaxNumberOfTransactions)
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
func (it *BridgeContractUpdatedMaxNumberOfTransactionsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractUpdatedMaxNumberOfTransactionsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractUpdatedMaxNumberOfTransactions represents a UpdatedMaxNumberOfTransactions event raised by the BridgeContract contract.
type BridgeContractUpdatedMaxNumberOfTransactions struct {
	MaxNumberOfTransactions *big.Int
	Raw                     types.Log // Blockchain specific contextual infos
}

// FilterUpdatedMaxNumberOfTransactions is a free log retrieval operation binding the contract event 0x0df74faf2972aea8bdd626cd6886a1fa4c9813a87981870c58f1e6c1ebd9f89a.
//
// Solidity: event UpdatedMaxNumberOfTransactions(uint256 _maxNumberOfTransactions)
func (_BridgeContract *BridgeContractFilterer) FilterUpdatedMaxNumberOfTransactions(opts *bind.FilterOpts) (*BridgeContractUpdatedMaxNumberOfTransactionsIterator, error) {

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "UpdatedMaxNumberOfTransactions")
	if err != nil {
		return nil, err
	}
	return &BridgeContractUpdatedMaxNumberOfTransactionsIterator{contract: _BridgeContract.contract, event: "UpdatedMaxNumberOfTransactions", logs: logs, sub: sub}, nil
}

// WatchUpdatedMaxNumberOfTransactions is a free log subscription operation binding the contract event 0x0df74faf2972aea8bdd626cd6886a1fa4c9813a87981870c58f1e6c1ebd9f89a.
//
// Solidity: event UpdatedMaxNumberOfTransactions(uint256 _maxNumberOfTransactions)
func (_BridgeContract *BridgeContractFilterer) WatchUpdatedMaxNumberOfTransactions(opts *bind.WatchOpts, sink chan<- *BridgeContractUpdatedMaxNumberOfTransactions) (event.Subscription, error) {

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "UpdatedMaxNumberOfTransactions")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractUpdatedMaxNumberOfTransactions)
				if err := _BridgeContract.contract.UnpackLog(event, "UpdatedMaxNumberOfTransactions", log); err != nil {
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

// ParseUpdatedMaxNumberOfTransactions is a log parse operation binding the contract event 0x0df74faf2972aea8bdd626cd6886a1fa4c9813a87981870c58f1e6c1ebd9f89a.
//
// Solidity: event UpdatedMaxNumberOfTransactions(uint256 _maxNumberOfTransactions)
func (_BridgeContract *BridgeContractFilterer) ParseUpdatedMaxNumberOfTransactions(log types.Log) (*BridgeContractUpdatedMaxNumberOfTransactions, error) {
	event := new(BridgeContractUpdatedMaxNumberOfTransactions)
	if err := _BridgeContract.contract.UnpackLog(event, "UpdatedMaxNumberOfTransactions", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeContractUpdatedTimeoutBlocksNumberIterator is returned from FilterUpdatedTimeoutBlocksNumber and is used to iterate over the raw logs and unpacked data for UpdatedTimeoutBlocksNumber events raised by the BridgeContract contract.
type BridgeContractUpdatedTimeoutBlocksNumberIterator struct {
	Event *BridgeContractUpdatedTimeoutBlocksNumber // Event containing the contract specifics and raw log

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
func (it *BridgeContractUpdatedTimeoutBlocksNumberIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractUpdatedTimeoutBlocksNumber)
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
		it.Event = new(BridgeContractUpdatedTimeoutBlocksNumber)
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
func (it *BridgeContractUpdatedTimeoutBlocksNumberIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractUpdatedTimeoutBlocksNumberIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractUpdatedTimeoutBlocksNumber represents a UpdatedTimeoutBlocksNumber event raised by the BridgeContract contract.
type BridgeContractUpdatedTimeoutBlocksNumber struct {
	TimeoutBlocksNumber *big.Int
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterUpdatedTimeoutBlocksNumber is a free log retrieval operation binding the contract event 0x417143ffedb5f1c20f085be7f77b19ca8b7c93d98a93d5a133e3516ecc23b409.
//
// Solidity: event UpdatedTimeoutBlocksNumber(uint256 _timeoutBlocksNumber)
func (_BridgeContract *BridgeContractFilterer) FilterUpdatedTimeoutBlocksNumber(opts *bind.FilterOpts) (*BridgeContractUpdatedTimeoutBlocksNumberIterator, error) {

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "UpdatedTimeoutBlocksNumber")
	if err != nil {
		return nil, err
	}
	return &BridgeContractUpdatedTimeoutBlocksNumberIterator{contract: _BridgeContract.contract, event: "UpdatedTimeoutBlocksNumber", logs: logs, sub: sub}, nil
}

// WatchUpdatedTimeoutBlocksNumber is a free log subscription operation binding the contract event 0x417143ffedb5f1c20f085be7f77b19ca8b7c93d98a93d5a133e3516ecc23b409.
//
// Solidity: event UpdatedTimeoutBlocksNumber(uint256 _timeoutBlocksNumber)
func (_BridgeContract *BridgeContractFilterer) WatchUpdatedTimeoutBlocksNumber(opts *bind.WatchOpts, sink chan<- *BridgeContractUpdatedTimeoutBlocksNumber) (event.Subscription, error) {

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "UpdatedTimeoutBlocksNumber")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractUpdatedTimeoutBlocksNumber)
				if err := _BridgeContract.contract.UnpackLog(event, "UpdatedTimeoutBlocksNumber", log); err != nil {
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

// ParseUpdatedTimeoutBlocksNumber is a log parse operation binding the contract event 0x417143ffedb5f1c20f085be7f77b19ca8b7c93d98a93d5a133e3516ecc23b409.
//
// Solidity: event UpdatedTimeoutBlocksNumber(uint256 _timeoutBlocksNumber)
func (_BridgeContract *BridgeContractFilterer) ParseUpdatedTimeoutBlocksNumber(log types.Log) (*BridgeContractUpdatedTimeoutBlocksNumber, error) {
	event := new(BridgeContractUpdatedTimeoutBlocksNumber)
	if err := _BridgeContract.contract.UnpackLog(event, "UpdatedTimeoutBlocksNumber", log); err != nil {
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
	ChainId uint8
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNewChainProposal is a free log retrieval operation binding the contract event 0xc546bc51d95705dd957ce30962375555dda4421e2004fbbf0b5e1527858f6c30.
//
// Solidity: event newChainProposal(uint8 indexed _chainId, address indexed sender)
func (_BridgeContract *BridgeContractFilterer) FilterNewChainProposal(opts *bind.FilterOpts, _chainId []uint8, sender []common.Address) (*BridgeContractNewChainProposalIterator, error) {

	var _chainIdRule []interface{}
	for _, _chainIdItem := range _chainId {
		_chainIdRule = append(_chainIdRule, _chainIdItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "newChainProposal", _chainIdRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &BridgeContractNewChainProposalIterator{contract: _BridgeContract.contract, event: "newChainProposal", logs: logs, sub: sub}, nil
}

// WatchNewChainProposal is a free log subscription operation binding the contract event 0xc546bc51d95705dd957ce30962375555dda4421e2004fbbf0b5e1527858f6c30.
//
// Solidity: event newChainProposal(uint8 indexed _chainId, address indexed sender)
func (_BridgeContract *BridgeContractFilterer) WatchNewChainProposal(opts *bind.WatchOpts, sink chan<- *BridgeContractNewChainProposal, _chainId []uint8, sender []common.Address) (event.Subscription, error) {

	var _chainIdRule []interface{}
	for _, _chainIdItem := range _chainId {
		_chainIdRule = append(_chainIdRule, _chainIdItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "newChainProposal", _chainIdRule, senderRule)
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

// ParseNewChainProposal is a log parse operation binding the contract event 0xc546bc51d95705dd957ce30962375555dda4421e2004fbbf0b5e1527858f6c30.
//
// Solidity: event newChainProposal(uint8 indexed _chainId, address indexed sender)
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
	ChainId uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNewChainRegistered is a free log retrieval operation binding the contract event 0x8541a0c729e909924d8678df4b2374f63c9514fcd5a430ac3de033d11d120256.
//
// Solidity: event newChainRegistered(uint8 indexed _chainId)
func (_BridgeContract *BridgeContractFilterer) FilterNewChainRegistered(opts *bind.FilterOpts, _chainId []uint8) (*BridgeContractNewChainRegisteredIterator, error) {

	var _chainIdRule []interface{}
	for _, _chainIdItem := range _chainId {
		_chainIdRule = append(_chainIdRule, _chainIdItem)
	}

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "newChainRegistered", _chainIdRule)
	if err != nil {
		return nil, err
	}
	return &BridgeContractNewChainRegisteredIterator{contract: _BridgeContract.contract, event: "newChainRegistered", logs: logs, sub: sub}, nil
}

// WatchNewChainRegistered is a free log subscription operation binding the contract event 0x8541a0c729e909924d8678df4b2374f63c9514fcd5a430ac3de033d11d120256.
//
// Solidity: event newChainRegistered(uint8 indexed _chainId)
func (_BridgeContract *BridgeContractFilterer) WatchNewChainRegistered(opts *bind.WatchOpts, sink chan<- *BridgeContractNewChainRegistered, _chainId []uint8) (event.Subscription, error) {

	var _chainIdRule []interface{}
	for _, _chainIdItem := range _chainId {
		_chainIdRule = append(_chainIdRule, _chainIdItem)
	}

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "newChainRegistered", _chainIdRule)
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

// ParseNewChainRegistered is a log parse operation binding the contract event 0x8541a0c729e909924d8678df4b2374f63c9514fcd5a430ac3de033d11d120256.
//
// Solidity: event newChainRegistered(uint8 indexed _chainId)
func (_BridgeContract *BridgeContractFilterer) ParseNewChainRegistered(log types.Log) (*BridgeContractNewChainRegistered, error) {
	event := new(BridgeContractNewChainRegistered)
	if err := _BridgeContract.contract.UnpackLog(event, "newChainRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
