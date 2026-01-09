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

// IGatewayStructsReceiverWithdraw is an auto generated low-level Go binding around an user-defined struct.
type IGatewayStructsReceiverWithdraw struct {
	Receiver string
	Amount   *big.Int
	TokenId  uint16
}

// GatewayMetaData contains all meta data concerning the Gateway contract.
var GatewayMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"AddressEmptyCode\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"BatchAlreadyExecuted\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"CurrencyTokenId\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"ERC1967InvalidImplementation\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ERC1967NonPayable\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FailedInnerCall\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"minFeeAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"feeAmount\",\"type\":\"uint256\"}],\"name\":\"InsufficientFee\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"minBridgingAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"bridgingAmount\",\"type\":\"uint256\"}],\"name\":\"InvalidBridgingAmount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSignature\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"NotContractAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotGateway\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotPredicate\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"TokenAddressAlreadyRegistered\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"tokenId\",\"type\":\"uint16\"}],\"name\":\"TokenIdAlreadyRegistered\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"tokenId\",\"type\":\"uint16\"}],\"name\":\"TokenNotRegistered\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TransferFailed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"UUPSUnauthorizedCallContext\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"slot\",\"type\":\"bytes32\"}],\"name\":\"UUPSUnsupportedProxiableUUID\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"WrongValidatorsSetValue\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"expected\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"received\",\"type\":\"uint256\"}],\"name\":\"WrongValue\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"DepositedToken\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"FundsDeposited\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"minFee\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"minAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"minTokenAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"minOperationFee\",\"type\":\"uint256\"}],\"name\":\"MinAmountsUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"TTLExpired\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"tokenId\",\"type\":\"uint16\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"contractAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"isLockUnlockToken\",\"type\":\"bool\"}],\"name\":\"TokenRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"treasuryAddress\",\"type\":\"address\"}],\"name\":\"TreasuryAddressUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"Upgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"ValidatorSetUpdatedGW\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"ValidatorsSetUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"destinationChainId\",\"type\":\"uint8\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint16\",\"name\":\"tokenId\",\"type\":\"uint16\"}],\"indexed\":false,\"internalType\":\"structIGatewayStructs.ReceiverWithdraw[]\",\"name\":\"receivers\",\"type\":\"tuple[]\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"operationFee\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Withdraw\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"WithdrawToken\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"UPGRADE_INTERFACE_VERSION\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"currencyTokenId\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_signature\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"_bitmap\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"_tokenId\",\"type\":\"uint16\"}],\"name\":\"getTokenAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_minFee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_minBridgingAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_minTokenBridgingAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_minOperationFee\",\"type\":\"uint256\"},{\"internalType\":\"uint16\",\"name\":\"_currencyTokenId\",\"type\":\"uint16\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minBridgingAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minOperationFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minTokenBridgingAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"nativeTokenPredicate\",\"outputs\":[{\"internalType\":\"contractINativeTokenPredicate\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"proxiableUUID\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_lockUnlockSCAddress\",\"type\":\"address\"},{\"internalType\":\"uint16\",\"name\":\"_tokenId\",\"type\":\"uint16\"},{\"internalType\":\"string\",\"name\":\"_name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"registerToken\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_nativeTokenPredicateAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_tokenFactoryAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_validatorsAddress\",\"type\":\"address\"}],\"name\":\"setDependencies\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_minFee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_minBridgingAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_minTokenBridgingAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_minOperationFee\",\"type\":\"uint256\"}],\"name\":\"setMinAmounts\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_treasuryAddress\",\"type\":\"address\"}],\"name\":\"setTreasuryAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"tokenFactory\",\"outputs\":[{\"internalType\":\"contractTokenFactory\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"treasuryAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_signature\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"_bitmap\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"updateValidatorsChainData\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"upgradeToAndCall\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"validators\",\"outputs\":[{\"internalType\":\"contractIValidators\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"version\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_destinationChainId\",\"type\":\"uint8\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint16\",\"name\":\"tokenId\",\"type\":\"uint16\"}],\"internalType\":\"structIGatewayStructs.ReceiverWithdraw[]\",\"name\":\"_receivers\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256\",\"name\":\"_fee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_operationFee\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
}

// GatewayABI is the input ABI used to generate the binding from.
// Deprecated: Use GatewayMetaData.ABI instead.
var GatewayABI = GatewayMetaData.ABI

// Gateway is an auto generated Go binding around an Ethereum contract.
type Gateway struct {
	GatewayCaller     // Read-only binding to the contract
	GatewayTransactor // Write-only binding to the contract
	GatewayFilterer   // Log filterer for contract events
}

// GatewayCaller is an auto generated read-only Go binding around an Ethereum contract.
type GatewayCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GatewayTransactor is an auto generated write-only Go binding around an Ethereum contract.
type GatewayTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GatewayFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type GatewayFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GatewaySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type GatewaySession struct {
	Contract     *Gateway          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GatewayCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type GatewayCallerSession struct {
	Contract *GatewayCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// GatewayTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type GatewayTransactorSession struct {
	Contract     *GatewayTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// GatewayRaw is an auto generated low-level Go binding around an Ethereum contract.
type GatewayRaw struct {
	Contract *Gateway // Generic contract binding to access the raw methods on
}

// GatewayCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type GatewayCallerRaw struct {
	Contract *GatewayCaller // Generic read-only contract binding to access the raw methods on
}

// GatewayTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type GatewayTransactorRaw struct {
	Contract *GatewayTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGateway creates a new instance of Gateway, bound to a specific deployed contract.
func NewGateway(address common.Address, backend bind.ContractBackend) (*Gateway, error) {
	contract, err := bindGateway(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Gateway{GatewayCaller: GatewayCaller{contract: contract}, GatewayTransactor: GatewayTransactor{contract: contract}, GatewayFilterer: GatewayFilterer{contract: contract}}, nil
}

// NewGatewayCaller creates a new read-only instance of Gateway, bound to a specific deployed contract.
func NewGatewayCaller(address common.Address, caller bind.ContractCaller) (*GatewayCaller, error) {
	contract, err := bindGateway(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GatewayCaller{contract: contract}, nil
}

// NewGatewayTransactor creates a new write-only instance of Gateway, bound to a specific deployed contract.
func NewGatewayTransactor(address common.Address, transactor bind.ContractTransactor) (*GatewayTransactor, error) {
	contract, err := bindGateway(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GatewayTransactor{contract: contract}, nil
}

// NewGatewayFilterer creates a new log filterer instance of Gateway, bound to a specific deployed contract.
func NewGatewayFilterer(address common.Address, filterer bind.ContractFilterer) (*GatewayFilterer, error) {
	contract, err := bindGateway(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GatewayFilterer{contract: contract}, nil
}

// bindGateway binds a generic wrapper to an already deployed contract.
func bindGateway(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := GatewayMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Gateway *GatewayRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Gateway.Contract.GatewayCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Gateway *GatewayRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Gateway.Contract.GatewayTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Gateway *GatewayRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Gateway.Contract.GatewayTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Gateway *GatewayCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Gateway.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Gateway *GatewayTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Gateway.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Gateway *GatewayTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Gateway.Contract.contract.Transact(opts, method, params...)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_Gateway *GatewayCaller) UPGRADEINTERFACEVERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Gateway.contract.Call(opts, &out, "UPGRADE_INTERFACE_VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_Gateway *GatewaySession) UPGRADEINTERFACEVERSION() (string, error) {
	return _Gateway.Contract.UPGRADEINTERFACEVERSION(&_Gateway.CallOpts)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_Gateway *GatewayCallerSession) UPGRADEINTERFACEVERSION() (string, error) {
	return _Gateway.Contract.UPGRADEINTERFACEVERSION(&_Gateway.CallOpts)
}

// CurrencyTokenId is a free data retrieval call binding the contract method 0x4ad45416.
//
// Solidity: function currencyTokenId() view returns(uint16)
func (_Gateway *GatewayCaller) CurrencyTokenId(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _Gateway.contract.Call(opts, &out, "currencyTokenId")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// CurrencyTokenId is a free data retrieval call binding the contract method 0x4ad45416.
//
// Solidity: function currencyTokenId() view returns(uint16)
func (_Gateway *GatewaySession) CurrencyTokenId() (uint16, error) {
	return _Gateway.Contract.CurrencyTokenId(&_Gateway.CallOpts)
}

// CurrencyTokenId is a free data retrieval call binding the contract method 0x4ad45416.
//
// Solidity: function currencyTokenId() view returns(uint16)
func (_Gateway *GatewayCallerSession) CurrencyTokenId() (uint16, error) {
	return _Gateway.Contract.CurrencyTokenId(&_Gateway.CallOpts)
}

// GetTokenAddress is a free data retrieval call binding the contract method 0xef365218.
//
// Solidity: function getTokenAddress(uint16 _tokenId) view returns(address)
func (_Gateway *GatewayCaller) GetTokenAddress(opts *bind.CallOpts, _tokenId uint16) (common.Address, error) {
	var out []interface{}
	err := _Gateway.contract.Call(opts, &out, "getTokenAddress", _tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetTokenAddress is a free data retrieval call binding the contract method 0xef365218.
//
// Solidity: function getTokenAddress(uint16 _tokenId) view returns(address)
func (_Gateway *GatewaySession) GetTokenAddress(_tokenId uint16) (common.Address, error) {
	return _Gateway.Contract.GetTokenAddress(&_Gateway.CallOpts, _tokenId)
}

// GetTokenAddress is a free data retrieval call binding the contract method 0xef365218.
//
// Solidity: function getTokenAddress(uint16 _tokenId) view returns(address)
func (_Gateway *GatewayCallerSession) GetTokenAddress(_tokenId uint16) (common.Address, error) {
	return _Gateway.Contract.GetTokenAddress(&_Gateway.CallOpts, _tokenId)
}

// MinBridgingAmount is a free data retrieval call binding the contract method 0x7ceb0eaa.
//
// Solidity: function minBridgingAmount() view returns(uint256)
func (_Gateway *GatewayCaller) MinBridgingAmount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Gateway.contract.Call(opts, &out, "minBridgingAmount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinBridgingAmount is a free data retrieval call binding the contract method 0x7ceb0eaa.
//
// Solidity: function minBridgingAmount() view returns(uint256)
func (_Gateway *GatewaySession) MinBridgingAmount() (*big.Int, error) {
	return _Gateway.Contract.MinBridgingAmount(&_Gateway.CallOpts)
}

// MinBridgingAmount is a free data retrieval call binding the contract method 0x7ceb0eaa.
//
// Solidity: function minBridgingAmount() view returns(uint256)
func (_Gateway *GatewayCallerSession) MinBridgingAmount() (*big.Int, error) {
	return _Gateway.Contract.MinBridgingAmount(&_Gateway.CallOpts)
}

// MinFee is a free data retrieval call binding the contract method 0x24ec7590.
//
// Solidity: function minFee() view returns(uint256)
func (_Gateway *GatewayCaller) MinFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Gateway.contract.Call(opts, &out, "minFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinFee is a free data retrieval call binding the contract method 0x24ec7590.
//
// Solidity: function minFee() view returns(uint256)
func (_Gateway *GatewaySession) MinFee() (*big.Int, error) {
	return _Gateway.Contract.MinFee(&_Gateway.CallOpts)
}

// MinFee is a free data retrieval call binding the contract method 0x24ec7590.
//
// Solidity: function minFee() view returns(uint256)
func (_Gateway *GatewayCallerSession) MinFee() (*big.Int, error) {
	return _Gateway.Contract.MinFee(&_Gateway.CallOpts)
}

// MinOperationFee is a free data retrieval call binding the contract method 0x3ad25471.
//
// Solidity: function minOperationFee() view returns(uint256)
func (_Gateway *GatewayCaller) MinOperationFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Gateway.contract.Call(opts, &out, "minOperationFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinOperationFee is a free data retrieval call binding the contract method 0x3ad25471.
//
// Solidity: function minOperationFee() view returns(uint256)
func (_Gateway *GatewaySession) MinOperationFee() (*big.Int, error) {
	return _Gateway.Contract.MinOperationFee(&_Gateway.CallOpts)
}

// MinOperationFee is a free data retrieval call binding the contract method 0x3ad25471.
//
// Solidity: function minOperationFee() view returns(uint256)
func (_Gateway *GatewayCallerSession) MinOperationFee() (*big.Int, error) {
	return _Gateway.Contract.MinOperationFee(&_Gateway.CallOpts)
}

// MinTokenBridgingAmount is a free data retrieval call binding the contract method 0xa0fbfc66.
//
// Solidity: function minTokenBridgingAmount() view returns(uint256)
func (_Gateway *GatewayCaller) MinTokenBridgingAmount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Gateway.contract.Call(opts, &out, "minTokenBridgingAmount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinTokenBridgingAmount is a free data retrieval call binding the contract method 0xa0fbfc66.
//
// Solidity: function minTokenBridgingAmount() view returns(uint256)
func (_Gateway *GatewaySession) MinTokenBridgingAmount() (*big.Int, error) {
	return _Gateway.Contract.MinTokenBridgingAmount(&_Gateway.CallOpts)
}

// MinTokenBridgingAmount is a free data retrieval call binding the contract method 0xa0fbfc66.
//
// Solidity: function minTokenBridgingAmount() view returns(uint256)
func (_Gateway *GatewayCallerSession) MinTokenBridgingAmount() (*big.Int, error) {
	return _Gateway.Contract.MinTokenBridgingAmount(&_Gateway.CallOpts)
}

// NativeTokenPredicate is a free data retrieval call binding the contract method 0xd4945a2c.
//
// Solidity: function nativeTokenPredicate() view returns(address)
func (_Gateway *GatewayCaller) NativeTokenPredicate(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Gateway.contract.Call(opts, &out, "nativeTokenPredicate")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// NativeTokenPredicate is a free data retrieval call binding the contract method 0xd4945a2c.
//
// Solidity: function nativeTokenPredicate() view returns(address)
func (_Gateway *GatewaySession) NativeTokenPredicate() (common.Address, error) {
	return _Gateway.Contract.NativeTokenPredicate(&_Gateway.CallOpts)
}

// NativeTokenPredicate is a free data retrieval call binding the contract method 0xd4945a2c.
//
// Solidity: function nativeTokenPredicate() view returns(address)
func (_Gateway *GatewayCallerSession) NativeTokenPredicate() (common.Address, error) {
	return _Gateway.Contract.NativeTokenPredicate(&_Gateway.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Gateway *GatewayCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Gateway.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Gateway *GatewaySession) Owner() (common.Address, error) {
	return _Gateway.Contract.Owner(&_Gateway.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Gateway *GatewayCallerSession) Owner() (common.Address, error) {
	return _Gateway.Contract.Owner(&_Gateway.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_Gateway *GatewayCaller) ProxiableUUID(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Gateway.contract.Call(opts, &out, "proxiableUUID")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_Gateway *GatewaySession) ProxiableUUID() ([32]byte, error) {
	return _Gateway.Contract.ProxiableUUID(&_Gateway.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_Gateway *GatewayCallerSession) ProxiableUUID() ([32]byte, error) {
	return _Gateway.Contract.ProxiableUUID(&_Gateway.CallOpts)
}

// TokenFactory is a free data retrieval call binding the contract method 0xe77772fe.
//
// Solidity: function tokenFactory() view returns(address)
func (_Gateway *GatewayCaller) TokenFactory(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Gateway.contract.Call(opts, &out, "tokenFactory")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TokenFactory is a free data retrieval call binding the contract method 0xe77772fe.
//
// Solidity: function tokenFactory() view returns(address)
func (_Gateway *GatewaySession) TokenFactory() (common.Address, error) {
	return _Gateway.Contract.TokenFactory(&_Gateway.CallOpts)
}

// TokenFactory is a free data retrieval call binding the contract method 0xe77772fe.
//
// Solidity: function tokenFactory() view returns(address)
func (_Gateway *GatewayCallerSession) TokenFactory() (common.Address, error) {
	return _Gateway.Contract.TokenFactory(&_Gateway.CallOpts)
}

// TreasuryAddress is a free data retrieval call binding the contract method 0xc5f956af.
//
// Solidity: function treasuryAddress() view returns(address)
func (_Gateway *GatewayCaller) TreasuryAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Gateway.contract.Call(opts, &out, "treasuryAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TreasuryAddress is a free data retrieval call binding the contract method 0xc5f956af.
//
// Solidity: function treasuryAddress() view returns(address)
func (_Gateway *GatewaySession) TreasuryAddress() (common.Address, error) {
	return _Gateway.Contract.TreasuryAddress(&_Gateway.CallOpts)
}

// TreasuryAddress is a free data retrieval call binding the contract method 0xc5f956af.
//
// Solidity: function treasuryAddress() view returns(address)
func (_Gateway *GatewayCallerSession) TreasuryAddress() (common.Address, error) {
	return _Gateway.Contract.TreasuryAddress(&_Gateway.CallOpts)
}

// Validators is a free data retrieval call binding the contract method 0xca1e7819.
//
// Solidity: function validators() view returns(address)
func (_Gateway *GatewayCaller) Validators(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Gateway.contract.Call(opts, &out, "validators")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Validators is a free data retrieval call binding the contract method 0xca1e7819.
//
// Solidity: function validators() view returns(address)
func (_Gateway *GatewaySession) Validators() (common.Address, error) {
	return _Gateway.Contract.Validators(&_Gateway.CallOpts)
}

// Validators is a free data retrieval call binding the contract method 0xca1e7819.
//
// Solidity: function validators() view returns(address)
func (_Gateway *GatewayCallerSession) Validators() (common.Address, error) {
	return _Gateway.Contract.Validators(&_Gateway.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(string)
func (_Gateway *GatewayCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Gateway.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(string)
func (_Gateway *GatewaySession) Version() (string, error) {
	return _Gateway.Contract.Version(&_Gateway.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(string)
func (_Gateway *GatewayCallerSession) Version() (string, error) {
	return _Gateway.Contract.Version(&_Gateway.CallOpts)
}

// Deposit is a paid mutator transaction binding the contract method 0x6e4c8d8a.
//
// Solidity: function deposit(bytes _signature, uint256 _bitmap, bytes _data) returns()
func (_Gateway *GatewayTransactor) Deposit(opts *bind.TransactOpts, _signature []byte, _bitmap *big.Int, _data []byte) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "deposit", _signature, _bitmap, _data)
}

// Deposit is a paid mutator transaction binding the contract method 0x6e4c8d8a.
//
// Solidity: function deposit(bytes _signature, uint256 _bitmap, bytes _data) returns()
func (_Gateway *GatewaySession) Deposit(_signature []byte, _bitmap *big.Int, _data []byte) (*types.Transaction, error) {
	return _Gateway.Contract.Deposit(&_Gateway.TransactOpts, _signature, _bitmap, _data)
}

// Deposit is a paid mutator transaction binding the contract method 0x6e4c8d8a.
//
// Solidity: function deposit(bytes _signature, uint256 _bitmap, bytes _data) returns()
func (_Gateway *GatewayTransactorSession) Deposit(_signature []byte, _bitmap *big.Int, _data []byte) (*types.Transaction, error) {
	return _Gateway.Contract.Deposit(&_Gateway.TransactOpts, _signature, _bitmap, _data)
}

// Initialize is a paid mutator transaction binding the contract method 0xb9102f7f.
//
// Solidity: function initialize(uint256 _minFee, uint256 _minBridgingAmount, uint256 _minTokenBridgingAmount, uint256 _minOperationFee, uint16 _currencyTokenId) returns()
func (_Gateway *GatewayTransactor) Initialize(opts *bind.TransactOpts, _minFee *big.Int, _minBridgingAmount *big.Int, _minTokenBridgingAmount *big.Int, _minOperationFee *big.Int, _currencyTokenId uint16) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "initialize", _minFee, _minBridgingAmount, _minTokenBridgingAmount, _minOperationFee, _currencyTokenId)
}

// Initialize is a paid mutator transaction binding the contract method 0xb9102f7f.
//
// Solidity: function initialize(uint256 _minFee, uint256 _minBridgingAmount, uint256 _minTokenBridgingAmount, uint256 _minOperationFee, uint16 _currencyTokenId) returns()
func (_Gateway *GatewaySession) Initialize(_minFee *big.Int, _minBridgingAmount *big.Int, _minTokenBridgingAmount *big.Int, _minOperationFee *big.Int, _currencyTokenId uint16) (*types.Transaction, error) {
	return _Gateway.Contract.Initialize(&_Gateway.TransactOpts, _minFee, _minBridgingAmount, _minTokenBridgingAmount, _minOperationFee, _currencyTokenId)
}

// Initialize is a paid mutator transaction binding the contract method 0xb9102f7f.
//
// Solidity: function initialize(uint256 _minFee, uint256 _minBridgingAmount, uint256 _minTokenBridgingAmount, uint256 _minOperationFee, uint16 _currencyTokenId) returns()
func (_Gateway *GatewayTransactorSession) Initialize(_minFee *big.Int, _minBridgingAmount *big.Int, _minTokenBridgingAmount *big.Int, _minOperationFee *big.Int, _currencyTokenId uint16) (*types.Transaction, error) {
	return _Gateway.Contract.Initialize(&_Gateway.TransactOpts, _minFee, _minBridgingAmount, _minTokenBridgingAmount, _minOperationFee, _currencyTokenId)
}

// RegisterToken is a paid mutator transaction binding the contract method 0xc1bd443e.
//
// Solidity: function registerToken(address _lockUnlockSCAddress, uint16 _tokenId, string _name, string _symbol) returns()
func (_Gateway *GatewayTransactor) RegisterToken(opts *bind.TransactOpts, _lockUnlockSCAddress common.Address, _tokenId uint16, _name string, _symbol string) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "registerToken", _lockUnlockSCAddress, _tokenId, _name, _symbol)
}

// RegisterToken is a paid mutator transaction binding the contract method 0xc1bd443e.
//
// Solidity: function registerToken(address _lockUnlockSCAddress, uint16 _tokenId, string _name, string _symbol) returns()
func (_Gateway *GatewaySession) RegisterToken(_lockUnlockSCAddress common.Address, _tokenId uint16, _name string, _symbol string) (*types.Transaction, error) {
	return _Gateway.Contract.RegisterToken(&_Gateway.TransactOpts, _lockUnlockSCAddress, _tokenId, _name, _symbol)
}

// RegisterToken is a paid mutator transaction binding the contract method 0xc1bd443e.
//
// Solidity: function registerToken(address _lockUnlockSCAddress, uint16 _tokenId, string _name, string _symbol) returns()
func (_Gateway *GatewayTransactorSession) RegisterToken(_lockUnlockSCAddress common.Address, _tokenId uint16, _name string, _symbol string) (*types.Transaction, error) {
	return _Gateway.Contract.RegisterToken(&_Gateway.TransactOpts, _lockUnlockSCAddress, _tokenId, _name, _symbol)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Gateway *GatewayTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Gateway *GatewaySession) RenounceOwnership() (*types.Transaction, error) {
	return _Gateway.Contract.RenounceOwnership(&_Gateway.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Gateway *GatewayTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Gateway.Contract.RenounceOwnership(&_Gateway.TransactOpts)
}

// SetDependencies is a paid mutator transaction binding the contract method 0xb52d326c.
//
// Solidity: function setDependencies(address _nativeTokenPredicateAddress, address _tokenFactoryAddress, address _validatorsAddress) returns()
func (_Gateway *GatewayTransactor) SetDependencies(opts *bind.TransactOpts, _nativeTokenPredicateAddress common.Address, _tokenFactoryAddress common.Address, _validatorsAddress common.Address) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "setDependencies", _nativeTokenPredicateAddress, _tokenFactoryAddress, _validatorsAddress)
}

// SetDependencies is a paid mutator transaction binding the contract method 0xb52d326c.
//
// Solidity: function setDependencies(address _nativeTokenPredicateAddress, address _tokenFactoryAddress, address _validatorsAddress) returns()
func (_Gateway *GatewaySession) SetDependencies(_nativeTokenPredicateAddress common.Address, _tokenFactoryAddress common.Address, _validatorsAddress common.Address) (*types.Transaction, error) {
	return _Gateway.Contract.SetDependencies(&_Gateway.TransactOpts, _nativeTokenPredicateAddress, _tokenFactoryAddress, _validatorsAddress)
}

// SetDependencies is a paid mutator transaction binding the contract method 0xb52d326c.
//
// Solidity: function setDependencies(address _nativeTokenPredicateAddress, address _tokenFactoryAddress, address _validatorsAddress) returns()
func (_Gateway *GatewayTransactorSession) SetDependencies(_nativeTokenPredicateAddress common.Address, _tokenFactoryAddress common.Address, _validatorsAddress common.Address) (*types.Transaction, error) {
	return _Gateway.Contract.SetDependencies(&_Gateway.TransactOpts, _nativeTokenPredicateAddress, _tokenFactoryAddress, _validatorsAddress)
}

// SetMinAmounts is a paid mutator transaction binding the contract method 0xa2045745.
//
// Solidity: function setMinAmounts(uint256 _minFee, uint256 _minBridgingAmount, uint256 _minTokenBridgingAmount, uint256 _minOperationFee) returns()
func (_Gateway *GatewayTransactor) SetMinAmounts(opts *bind.TransactOpts, _minFee *big.Int, _minBridgingAmount *big.Int, _minTokenBridgingAmount *big.Int, _minOperationFee *big.Int) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "setMinAmounts", _minFee, _minBridgingAmount, _minTokenBridgingAmount, _minOperationFee)
}

// SetMinAmounts is a paid mutator transaction binding the contract method 0xa2045745.
//
// Solidity: function setMinAmounts(uint256 _minFee, uint256 _minBridgingAmount, uint256 _minTokenBridgingAmount, uint256 _minOperationFee) returns()
func (_Gateway *GatewaySession) SetMinAmounts(_minFee *big.Int, _minBridgingAmount *big.Int, _minTokenBridgingAmount *big.Int, _minOperationFee *big.Int) (*types.Transaction, error) {
	return _Gateway.Contract.SetMinAmounts(&_Gateway.TransactOpts, _minFee, _minBridgingAmount, _minTokenBridgingAmount, _minOperationFee)
}

// SetMinAmounts is a paid mutator transaction binding the contract method 0xa2045745.
//
// Solidity: function setMinAmounts(uint256 _minFee, uint256 _minBridgingAmount, uint256 _minTokenBridgingAmount, uint256 _minOperationFee) returns()
func (_Gateway *GatewayTransactorSession) SetMinAmounts(_minFee *big.Int, _minBridgingAmount *big.Int, _minTokenBridgingAmount *big.Int, _minOperationFee *big.Int) (*types.Transaction, error) {
	return _Gateway.Contract.SetMinAmounts(&_Gateway.TransactOpts, _minFee, _minBridgingAmount, _minTokenBridgingAmount, _minOperationFee)
}

// SetTreasuryAddress is a paid mutator transaction binding the contract method 0x6605bfda.
//
// Solidity: function setTreasuryAddress(address _treasuryAddress) returns()
func (_Gateway *GatewayTransactor) SetTreasuryAddress(opts *bind.TransactOpts, _treasuryAddress common.Address) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "setTreasuryAddress", _treasuryAddress)
}

// SetTreasuryAddress is a paid mutator transaction binding the contract method 0x6605bfda.
//
// Solidity: function setTreasuryAddress(address _treasuryAddress) returns()
func (_Gateway *GatewaySession) SetTreasuryAddress(_treasuryAddress common.Address) (*types.Transaction, error) {
	return _Gateway.Contract.SetTreasuryAddress(&_Gateway.TransactOpts, _treasuryAddress)
}

// SetTreasuryAddress is a paid mutator transaction binding the contract method 0x6605bfda.
//
// Solidity: function setTreasuryAddress(address _treasuryAddress) returns()
func (_Gateway *GatewayTransactorSession) SetTreasuryAddress(_treasuryAddress common.Address) (*types.Transaction, error) {
	return _Gateway.Contract.SetTreasuryAddress(&_Gateway.TransactOpts, _treasuryAddress)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Gateway *GatewayTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Gateway *GatewaySession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Gateway.Contract.TransferOwnership(&_Gateway.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Gateway *GatewayTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Gateway.Contract.TransferOwnership(&_Gateway.TransactOpts, newOwner)
}

// UpdateValidatorsChainData is a paid mutator transaction binding the contract method 0x188febdd.
//
// Solidity: function updateValidatorsChainData(bytes _signature, uint256 _bitmap, bytes _data) returns()
func (_Gateway *GatewayTransactor) UpdateValidatorsChainData(opts *bind.TransactOpts, _signature []byte, _bitmap *big.Int, _data []byte) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "updateValidatorsChainData", _signature, _bitmap, _data)
}

// UpdateValidatorsChainData is a paid mutator transaction binding the contract method 0x188febdd.
//
// Solidity: function updateValidatorsChainData(bytes _signature, uint256 _bitmap, bytes _data) returns()
func (_Gateway *GatewaySession) UpdateValidatorsChainData(_signature []byte, _bitmap *big.Int, _data []byte) (*types.Transaction, error) {
	return _Gateway.Contract.UpdateValidatorsChainData(&_Gateway.TransactOpts, _signature, _bitmap, _data)
}

// UpdateValidatorsChainData is a paid mutator transaction binding the contract method 0x188febdd.
//
// Solidity: function updateValidatorsChainData(bytes _signature, uint256 _bitmap, bytes _data) returns()
func (_Gateway *GatewayTransactorSession) UpdateValidatorsChainData(_signature []byte, _bitmap *big.Int, _data []byte) (*types.Transaction, error) {
	return _Gateway.Contract.UpdateValidatorsChainData(&_Gateway.TransactOpts, _signature, _bitmap, _data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_Gateway *GatewayTransactor) UpgradeToAndCall(opts *bind.TransactOpts, newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "upgradeToAndCall", newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_Gateway *GatewaySession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _Gateway.Contract.UpgradeToAndCall(&_Gateway.TransactOpts, newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_Gateway *GatewayTransactorSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _Gateway.Contract.UpgradeToAndCall(&_Gateway.TransactOpts, newImplementation, data)
}

// Withdraw is a paid mutator transaction binding the contract method 0x086af24c.
//
// Solidity: function withdraw(uint8 _destinationChainId, (string,uint256,uint16)[] _receivers, uint256 _fee, uint256 _operationFee) payable returns()
func (_Gateway *GatewayTransactor) Withdraw(opts *bind.TransactOpts, _destinationChainId uint8, _receivers []IGatewayStructsReceiverWithdraw, _fee *big.Int, _operationFee *big.Int) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "withdraw", _destinationChainId, _receivers, _fee, _operationFee)
}

// Withdraw is a paid mutator transaction binding the contract method 0x086af24c.
//
// Solidity: function withdraw(uint8 _destinationChainId, (string,uint256,uint16)[] _receivers, uint256 _fee, uint256 _operationFee) payable returns()
func (_Gateway *GatewaySession) Withdraw(_destinationChainId uint8, _receivers []IGatewayStructsReceiverWithdraw, _fee *big.Int, _operationFee *big.Int) (*types.Transaction, error) {
	return _Gateway.Contract.Withdraw(&_Gateway.TransactOpts, _destinationChainId, _receivers, _fee, _operationFee)
}

// Withdraw is a paid mutator transaction binding the contract method 0x086af24c.
//
// Solidity: function withdraw(uint8 _destinationChainId, (string,uint256,uint16)[] _receivers, uint256 _fee, uint256 _operationFee) payable returns()
func (_Gateway *GatewayTransactorSession) Withdraw(_destinationChainId uint8, _receivers []IGatewayStructsReceiverWithdraw, _fee *big.Int, _operationFee *big.Int) (*types.Transaction, error) {
	return _Gateway.Contract.Withdraw(&_Gateway.TransactOpts, _destinationChainId, _receivers, _fee, _operationFee)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Gateway *GatewayTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Gateway.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Gateway *GatewaySession) Receive() (*types.Transaction, error) {
	return _Gateway.Contract.Receive(&_Gateway.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Gateway *GatewayTransactorSession) Receive() (*types.Transaction, error) {
	return _Gateway.Contract.Receive(&_Gateway.TransactOpts)
}

// GatewayDepositIterator is returned from FilterDeposit and is used to iterate over the raw logs and unpacked data for Deposit events raised by the Gateway contract.
type GatewayDepositIterator struct {
	Event *GatewayDeposit // Event containing the contract specifics and raw log

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
func (it *GatewayDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatewayDeposit)
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
		it.Event = new(GatewayDeposit)
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
func (it *GatewayDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatewayDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatewayDeposit represents a Deposit event raised by the Gateway contract.
type GatewayDeposit struct {
	Data []byte
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0x7adcde22575d10ee3d4e78ee24cc9f854ecc4ce2bc5fda5eadeb754384227db0.
//
// Solidity: event Deposit(bytes data)
func (_Gateway *GatewayFilterer) FilterDeposit(opts *bind.FilterOpts) (*GatewayDepositIterator, error) {

	logs, sub, err := _Gateway.contract.FilterLogs(opts, "Deposit")
	if err != nil {
		return nil, err
	}
	return &GatewayDepositIterator{contract: _Gateway.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0x7adcde22575d10ee3d4e78ee24cc9f854ecc4ce2bc5fda5eadeb754384227db0.
//
// Solidity: event Deposit(bytes data)
func (_Gateway *GatewayFilterer) WatchDeposit(opts *bind.WatchOpts, sink chan<- *GatewayDeposit) (event.Subscription, error) {

	logs, sub, err := _Gateway.contract.WatchLogs(opts, "Deposit")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatewayDeposit)
				if err := _Gateway.contract.UnpackLog(event, "Deposit", log); err != nil {
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

// ParseDeposit is a log parse operation binding the contract event 0x7adcde22575d10ee3d4e78ee24cc9f854ecc4ce2bc5fda5eadeb754384227db0.
//
// Solidity: event Deposit(bytes data)
func (_Gateway *GatewayFilterer) ParseDeposit(log types.Log) (*GatewayDeposit, error) {
	event := new(GatewayDeposit)
	if err := _Gateway.contract.UnpackLog(event, "Deposit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GatewayDepositedTokenIterator is returned from FilterDepositedToken and is used to iterate over the raw logs and unpacked data for DepositedToken events raised by the Gateway contract.
type GatewayDepositedTokenIterator struct {
	Event *GatewayDepositedToken // Event containing the contract specifics and raw log

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
func (it *GatewayDepositedTokenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatewayDepositedToken)
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
		it.Event = new(GatewayDepositedToken)
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
func (it *GatewayDepositedTokenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatewayDepositedTokenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatewayDepositedToken represents a DepositedToken event raised by the Gateway contract.
type GatewayDepositedToken struct {
	Data []byte
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterDepositedToken is a free log retrieval operation binding the contract event 0x82c6a6a514dcc64474d22bee3ca6c8968021d85c8325eb4a306503fcafc3501c.
//
// Solidity: event DepositedToken(bytes data)
func (_Gateway *GatewayFilterer) FilterDepositedToken(opts *bind.FilterOpts) (*GatewayDepositedTokenIterator, error) {

	logs, sub, err := _Gateway.contract.FilterLogs(opts, "DepositedToken")
	if err != nil {
		return nil, err
	}
	return &GatewayDepositedTokenIterator{contract: _Gateway.contract, event: "DepositedToken", logs: logs, sub: sub}, nil
}

// WatchDepositedToken is a free log subscription operation binding the contract event 0x82c6a6a514dcc64474d22bee3ca6c8968021d85c8325eb4a306503fcafc3501c.
//
// Solidity: event DepositedToken(bytes data)
func (_Gateway *GatewayFilterer) WatchDepositedToken(opts *bind.WatchOpts, sink chan<- *GatewayDepositedToken) (event.Subscription, error) {

	logs, sub, err := _Gateway.contract.WatchLogs(opts, "DepositedToken")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatewayDepositedToken)
				if err := _Gateway.contract.UnpackLog(event, "DepositedToken", log); err != nil {
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

// ParseDepositedToken is a log parse operation binding the contract event 0x82c6a6a514dcc64474d22bee3ca6c8968021d85c8325eb4a306503fcafc3501c.
//
// Solidity: event DepositedToken(bytes data)
func (_Gateway *GatewayFilterer) ParseDepositedToken(log types.Log) (*GatewayDepositedToken, error) {
	event := new(GatewayDepositedToken)
	if err := _Gateway.contract.UnpackLog(event, "DepositedToken", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GatewayFundsDepositedIterator is returned from FilterFundsDeposited and is used to iterate over the raw logs and unpacked data for FundsDeposited events raised by the Gateway contract.
type GatewayFundsDepositedIterator struct {
	Event *GatewayFundsDeposited // Event containing the contract specifics and raw log

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
func (it *GatewayFundsDepositedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatewayFundsDeposited)
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
		it.Event = new(GatewayFundsDeposited)
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
func (it *GatewayFundsDepositedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatewayFundsDepositedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatewayFundsDeposited represents a FundsDeposited event raised by the Gateway contract.
type GatewayFundsDeposited struct {
	Sender common.Address
	Value  *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterFundsDeposited is a free log retrieval operation binding the contract event 0x543ba50a5eec5e6178218e364b1d0f396157b3c8fa278522c2cb7fd99407d474.
//
// Solidity: event FundsDeposited(address indexed sender, uint256 value)
func (_Gateway *GatewayFilterer) FilterFundsDeposited(opts *bind.FilterOpts, sender []common.Address) (*GatewayFundsDepositedIterator, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Gateway.contract.FilterLogs(opts, "FundsDeposited", senderRule)
	if err != nil {
		return nil, err
	}
	return &GatewayFundsDepositedIterator{contract: _Gateway.contract, event: "FundsDeposited", logs: logs, sub: sub}, nil
}

// WatchFundsDeposited is a free log subscription operation binding the contract event 0x543ba50a5eec5e6178218e364b1d0f396157b3c8fa278522c2cb7fd99407d474.
//
// Solidity: event FundsDeposited(address indexed sender, uint256 value)
func (_Gateway *GatewayFilterer) WatchFundsDeposited(opts *bind.WatchOpts, sink chan<- *GatewayFundsDeposited, sender []common.Address) (event.Subscription, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Gateway.contract.WatchLogs(opts, "FundsDeposited", senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatewayFundsDeposited)
				if err := _Gateway.contract.UnpackLog(event, "FundsDeposited", log); err != nil {
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

// ParseFundsDeposited is a log parse operation binding the contract event 0x543ba50a5eec5e6178218e364b1d0f396157b3c8fa278522c2cb7fd99407d474.
//
// Solidity: event FundsDeposited(address indexed sender, uint256 value)
func (_Gateway *GatewayFilterer) ParseFundsDeposited(log types.Log) (*GatewayFundsDeposited, error) {
	event := new(GatewayFundsDeposited)
	if err := _Gateway.contract.UnpackLog(event, "FundsDeposited", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GatewayInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Gateway contract.
type GatewayInitializedIterator struct {
	Event *GatewayInitialized // Event containing the contract specifics and raw log

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
func (it *GatewayInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatewayInitialized)
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
		it.Event = new(GatewayInitialized)
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
func (it *GatewayInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatewayInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatewayInitialized represents a Initialized event raised by the Gateway contract.
type GatewayInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Gateway *GatewayFilterer) FilterInitialized(opts *bind.FilterOpts) (*GatewayInitializedIterator, error) {

	logs, sub, err := _Gateway.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &GatewayInitializedIterator{contract: _Gateway.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Gateway *GatewayFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *GatewayInitialized) (event.Subscription, error) {

	logs, sub, err := _Gateway.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatewayInitialized)
				if err := _Gateway.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_Gateway *GatewayFilterer) ParseInitialized(log types.Log) (*GatewayInitialized, error) {
	event := new(GatewayInitialized)
	if err := _Gateway.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GatewayMinAmountsUpdatedIterator is returned from FilterMinAmountsUpdated and is used to iterate over the raw logs and unpacked data for MinAmountsUpdated events raised by the Gateway contract.
type GatewayMinAmountsUpdatedIterator struct {
	Event *GatewayMinAmountsUpdated // Event containing the contract specifics and raw log

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
func (it *GatewayMinAmountsUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatewayMinAmountsUpdated)
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
		it.Event = new(GatewayMinAmountsUpdated)
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
func (it *GatewayMinAmountsUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatewayMinAmountsUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatewayMinAmountsUpdated represents a MinAmountsUpdated event raised by the Gateway contract.
type GatewayMinAmountsUpdated struct {
	MinFee          *big.Int
	MinAmount       *big.Int
	MinTokenAmount  *big.Int
	MinOperationFee *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterMinAmountsUpdated is a free log retrieval operation binding the contract event 0x618870fd286934102806f0b698965599ebc0d3e3af247d88f2f09513b30a5f28.
//
// Solidity: event MinAmountsUpdated(uint256 minFee, uint256 minAmount, uint256 minTokenAmount, uint256 minOperationFee)
func (_Gateway *GatewayFilterer) FilterMinAmountsUpdated(opts *bind.FilterOpts) (*GatewayMinAmountsUpdatedIterator, error) {

	logs, sub, err := _Gateway.contract.FilterLogs(opts, "MinAmountsUpdated")
	if err != nil {
		return nil, err
	}
	return &GatewayMinAmountsUpdatedIterator{contract: _Gateway.contract, event: "MinAmountsUpdated", logs: logs, sub: sub}, nil
}

// WatchMinAmountsUpdated is a free log subscription operation binding the contract event 0x618870fd286934102806f0b698965599ebc0d3e3af247d88f2f09513b30a5f28.
//
// Solidity: event MinAmountsUpdated(uint256 minFee, uint256 minAmount, uint256 minTokenAmount, uint256 minOperationFee)
func (_Gateway *GatewayFilterer) WatchMinAmountsUpdated(opts *bind.WatchOpts, sink chan<- *GatewayMinAmountsUpdated) (event.Subscription, error) {

	logs, sub, err := _Gateway.contract.WatchLogs(opts, "MinAmountsUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatewayMinAmountsUpdated)
				if err := _Gateway.contract.UnpackLog(event, "MinAmountsUpdated", log); err != nil {
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

// ParseMinAmountsUpdated is a log parse operation binding the contract event 0x618870fd286934102806f0b698965599ebc0d3e3af247d88f2f09513b30a5f28.
//
// Solidity: event MinAmountsUpdated(uint256 minFee, uint256 minAmount, uint256 minTokenAmount, uint256 minOperationFee)
func (_Gateway *GatewayFilterer) ParseMinAmountsUpdated(log types.Log) (*GatewayMinAmountsUpdated, error) {
	event := new(GatewayMinAmountsUpdated)
	if err := _Gateway.contract.UnpackLog(event, "MinAmountsUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GatewayOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Gateway contract.
type GatewayOwnershipTransferredIterator struct {
	Event *GatewayOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *GatewayOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatewayOwnershipTransferred)
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
		it.Event = new(GatewayOwnershipTransferred)
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
func (it *GatewayOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatewayOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatewayOwnershipTransferred represents a OwnershipTransferred event raised by the Gateway contract.
type GatewayOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Gateway *GatewayFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*GatewayOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Gateway.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &GatewayOwnershipTransferredIterator{contract: _Gateway.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Gateway *GatewayFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *GatewayOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Gateway.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatewayOwnershipTransferred)
				if err := _Gateway.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_Gateway *GatewayFilterer) ParseOwnershipTransferred(log types.Log) (*GatewayOwnershipTransferred, error) {
	event := new(GatewayOwnershipTransferred)
	if err := _Gateway.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GatewayTTLExpiredIterator is returned from FilterTTLExpired and is used to iterate over the raw logs and unpacked data for TTLExpired events raised by the Gateway contract.
type GatewayTTLExpiredIterator struct {
	Event *GatewayTTLExpired // Event containing the contract specifics and raw log

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
func (it *GatewayTTLExpiredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatewayTTLExpired)
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
		it.Event = new(GatewayTTLExpired)
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
func (it *GatewayTTLExpiredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatewayTTLExpiredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatewayTTLExpired represents a TTLExpired event raised by the Gateway contract.
type GatewayTTLExpired struct {
	Data []byte
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterTTLExpired is a free log retrieval operation binding the contract event 0x0ade0be4bc31d69de480eac67556041c790defee8b124e80c09859865d56af09.
//
// Solidity: event TTLExpired(bytes data)
func (_Gateway *GatewayFilterer) FilterTTLExpired(opts *bind.FilterOpts) (*GatewayTTLExpiredIterator, error) {

	logs, sub, err := _Gateway.contract.FilterLogs(opts, "TTLExpired")
	if err != nil {
		return nil, err
	}
	return &GatewayTTLExpiredIterator{contract: _Gateway.contract, event: "TTLExpired", logs: logs, sub: sub}, nil
}

// WatchTTLExpired is a free log subscription operation binding the contract event 0x0ade0be4bc31d69de480eac67556041c790defee8b124e80c09859865d56af09.
//
// Solidity: event TTLExpired(bytes data)
func (_Gateway *GatewayFilterer) WatchTTLExpired(opts *bind.WatchOpts, sink chan<- *GatewayTTLExpired) (event.Subscription, error) {

	logs, sub, err := _Gateway.contract.WatchLogs(opts, "TTLExpired")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatewayTTLExpired)
				if err := _Gateway.contract.UnpackLog(event, "TTLExpired", log); err != nil {
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

// ParseTTLExpired is a log parse operation binding the contract event 0x0ade0be4bc31d69de480eac67556041c790defee8b124e80c09859865d56af09.
//
// Solidity: event TTLExpired(bytes data)
func (_Gateway *GatewayFilterer) ParseTTLExpired(log types.Log) (*GatewayTTLExpired, error) {
	event := new(GatewayTTLExpired)
	if err := _Gateway.contract.UnpackLog(event, "TTLExpired", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GatewayTokenRegisteredIterator is returned from FilterTokenRegistered and is used to iterate over the raw logs and unpacked data for TokenRegistered events raised by the Gateway contract.
type GatewayTokenRegisteredIterator struct {
	Event *GatewayTokenRegistered // Event containing the contract specifics and raw log

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
func (it *GatewayTokenRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatewayTokenRegistered)
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
		it.Event = new(GatewayTokenRegistered)
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
func (it *GatewayTokenRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatewayTokenRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatewayTokenRegistered represents a TokenRegistered event raised by the Gateway contract.
type GatewayTokenRegistered struct {
	Name              string
	Symbol            string
	TokenId           uint16
	ContractAddress   common.Address
	IsLockUnlockToken bool
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterTokenRegistered is a free log retrieval operation binding the contract event 0xc1c275bf6f344563f8ca02cc3c8032d7d7e44c1f11ef2ea648b6c5be9e40b680.
//
// Solidity: event TokenRegistered(string name, string symbol, uint16 tokenId, address contractAddress, bool isLockUnlockToken)
func (_Gateway *GatewayFilterer) FilterTokenRegistered(opts *bind.FilterOpts) (*GatewayTokenRegisteredIterator, error) {

	logs, sub, err := _Gateway.contract.FilterLogs(opts, "TokenRegistered")
	if err != nil {
		return nil, err
	}
	return &GatewayTokenRegisteredIterator{contract: _Gateway.contract, event: "TokenRegistered", logs: logs, sub: sub}, nil
}

// WatchTokenRegistered is a free log subscription operation binding the contract event 0xc1c275bf6f344563f8ca02cc3c8032d7d7e44c1f11ef2ea648b6c5be9e40b680.
//
// Solidity: event TokenRegistered(string name, string symbol, uint16 tokenId, address contractAddress, bool isLockUnlockToken)
func (_Gateway *GatewayFilterer) WatchTokenRegistered(opts *bind.WatchOpts, sink chan<- *GatewayTokenRegistered) (event.Subscription, error) {

	logs, sub, err := _Gateway.contract.WatchLogs(opts, "TokenRegistered")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatewayTokenRegistered)
				if err := _Gateway.contract.UnpackLog(event, "TokenRegistered", log); err != nil {
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

// ParseTokenRegistered is a log parse operation binding the contract event 0xc1c275bf6f344563f8ca02cc3c8032d7d7e44c1f11ef2ea648b6c5be9e40b680.
//
// Solidity: event TokenRegistered(string name, string symbol, uint16 tokenId, address contractAddress, bool isLockUnlockToken)
func (_Gateway *GatewayFilterer) ParseTokenRegistered(log types.Log) (*GatewayTokenRegistered, error) {
	event := new(GatewayTokenRegistered)
	if err := _Gateway.contract.UnpackLog(event, "TokenRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GatewayTreasuryAddressUpdatedIterator is returned from FilterTreasuryAddressUpdated and is used to iterate over the raw logs and unpacked data for TreasuryAddressUpdated events raised by the Gateway contract.
type GatewayTreasuryAddressUpdatedIterator struct {
	Event *GatewayTreasuryAddressUpdated // Event containing the contract specifics and raw log

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
func (it *GatewayTreasuryAddressUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatewayTreasuryAddressUpdated)
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
		it.Event = new(GatewayTreasuryAddressUpdated)
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
func (it *GatewayTreasuryAddressUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatewayTreasuryAddressUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatewayTreasuryAddressUpdated represents a TreasuryAddressUpdated event raised by the Gateway contract.
type GatewayTreasuryAddressUpdated struct {
	TreasuryAddress common.Address
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterTreasuryAddressUpdated is a free log retrieval operation binding the contract event 0xb6a5e89655cf506139085f051af608195ed056f8dc550b180a1c38d401e2b6c4.
//
// Solidity: event TreasuryAddressUpdated(address treasuryAddress)
func (_Gateway *GatewayFilterer) FilterTreasuryAddressUpdated(opts *bind.FilterOpts) (*GatewayTreasuryAddressUpdatedIterator, error) {

	logs, sub, err := _Gateway.contract.FilterLogs(opts, "TreasuryAddressUpdated")
	if err != nil {
		return nil, err
	}
	return &GatewayTreasuryAddressUpdatedIterator{contract: _Gateway.contract, event: "TreasuryAddressUpdated", logs: logs, sub: sub}, nil
}

// WatchTreasuryAddressUpdated is a free log subscription operation binding the contract event 0xb6a5e89655cf506139085f051af608195ed056f8dc550b180a1c38d401e2b6c4.
//
// Solidity: event TreasuryAddressUpdated(address treasuryAddress)
func (_Gateway *GatewayFilterer) WatchTreasuryAddressUpdated(opts *bind.WatchOpts, sink chan<- *GatewayTreasuryAddressUpdated) (event.Subscription, error) {

	logs, sub, err := _Gateway.contract.WatchLogs(opts, "TreasuryAddressUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatewayTreasuryAddressUpdated)
				if err := _Gateway.contract.UnpackLog(event, "TreasuryAddressUpdated", log); err != nil {
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

// ParseTreasuryAddressUpdated is a log parse operation binding the contract event 0xb6a5e89655cf506139085f051af608195ed056f8dc550b180a1c38d401e2b6c4.
//
// Solidity: event TreasuryAddressUpdated(address treasuryAddress)
func (_Gateway *GatewayFilterer) ParseTreasuryAddressUpdated(log types.Log) (*GatewayTreasuryAddressUpdated, error) {
	event := new(GatewayTreasuryAddressUpdated)
	if err := _Gateway.contract.UnpackLog(event, "TreasuryAddressUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GatewayUpgradedIterator is returned from FilterUpgraded and is used to iterate over the raw logs and unpacked data for Upgraded events raised by the Gateway contract.
type GatewayUpgradedIterator struct {
	Event *GatewayUpgraded // Event containing the contract specifics and raw log

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
func (it *GatewayUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatewayUpgraded)
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
		it.Event = new(GatewayUpgraded)
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
func (it *GatewayUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatewayUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatewayUpgraded represents a Upgraded event raised by the Gateway contract.
type GatewayUpgraded struct {
	Implementation common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterUpgraded is a free log retrieval operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_Gateway *GatewayFilterer) FilterUpgraded(opts *bind.FilterOpts, implementation []common.Address) (*GatewayUpgradedIterator, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _Gateway.contract.FilterLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return &GatewayUpgradedIterator{contract: _Gateway.contract, event: "Upgraded", logs: logs, sub: sub}, nil
}

// WatchUpgraded is a free log subscription operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_Gateway *GatewayFilterer) WatchUpgraded(opts *bind.WatchOpts, sink chan<- *GatewayUpgraded, implementation []common.Address) (event.Subscription, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _Gateway.contract.WatchLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatewayUpgraded)
				if err := _Gateway.contract.UnpackLog(event, "Upgraded", log); err != nil {
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
func (_Gateway *GatewayFilterer) ParseUpgraded(log types.Log) (*GatewayUpgraded, error) {
	event := new(GatewayUpgraded)
	if err := _Gateway.contract.UnpackLog(event, "Upgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GatewayValidatorSetUpdatedGWIterator is returned from FilterValidatorSetUpdatedGW and is used to iterate over the raw logs and unpacked data for ValidatorSetUpdatedGW events raised by the Gateway contract.
type GatewayValidatorSetUpdatedGWIterator struct {
	Event *GatewayValidatorSetUpdatedGW // Event containing the contract specifics and raw log

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
func (it *GatewayValidatorSetUpdatedGWIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatewayValidatorSetUpdatedGW)
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
		it.Event = new(GatewayValidatorSetUpdatedGW)
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
func (it *GatewayValidatorSetUpdatedGWIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatewayValidatorSetUpdatedGWIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatewayValidatorSetUpdatedGW represents a ValidatorSetUpdatedGW event raised by the Gateway contract.
type GatewayValidatorSetUpdatedGW struct {
	Data []byte
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterValidatorSetUpdatedGW is a free log retrieval operation binding the contract event 0xe18a7671675ff9ab74c96c1fca4e7ef4985c60f743b44c2952d9233bd9069f15.
//
// Solidity: event ValidatorSetUpdatedGW(bytes data)
func (_Gateway *GatewayFilterer) FilterValidatorSetUpdatedGW(opts *bind.FilterOpts) (*GatewayValidatorSetUpdatedGWIterator, error) {

	logs, sub, err := _Gateway.contract.FilterLogs(opts, "ValidatorSetUpdatedGW")
	if err != nil {
		return nil, err
	}
	return &GatewayValidatorSetUpdatedGWIterator{contract: _Gateway.contract, event: "ValidatorSetUpdatedGW", logs: logs, sub: sub}, nil
}

// WatchValidatorSetUpdatedGW is a free log subscription operation binding the contract event 0xe18a7671675ff9ab74c96c1fca4e7ef4985c60f743b44c2952d9233bd9069f15.
//
// Solidity: event ValidatorSetUpdatedGW(bytes data)
func (_Gateway *GatewayFilterer) WatchValidatorSetUpdatedGW(opts *bind.WatchOpts, sink chan<- *GatewayValidatorSetUpdatedGW) (event.Subscription, error) {

	logs, sub, err := _Gateway.contract.WatchLogs(opts, "ValidatorSetUpdatedGW")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatewayValidatorSetUpdatedGW)
				if err := _Gateway.contract.UnpackLog(event, "ValidatorSetUpdatedGW", log); err != nil {
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

// ParseValidatorSetUpdatedGW is a log parse operation binding the contract event 0xe18a7671675ff9ab74c96c1fca4e7ef4985c60f743b44c2952d9233bd9069f15.
//
// Solidity: event ValidatorSetUpdatedGW(bytes data)
func (_Gateway *GatewayFilterer) ParseValidatorSetUpdatedGW(log types.Log) (*GatewayValidatorSetUpdatedGW, error) {
	event := new(GatewayValidatorSetUpdatedGW)
	if err := _Gateway.contract.UnpackLog(event, "ValidatorSetUpdatedGW", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GatewayValidatorsSetUpdatedIterator is returned from FilterValidatorsSetUpdated and is used to iterate over the raw logs and unpacked data for ValidatorsSetUpdated events raised by the Gateway contract.
type GatewayValidatorsSetUpdatedIterator struct {
	Event *GatewayValidatorsSetUpdated // Event containing the contract specifics and raw log

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
func (it *GatewayValidatorsSetUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatewayValidatorsSetUpdated)
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
		it.Event = new(GatewayValidatorsSetUpdated)
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
func (it *GatewayValidatorsSetUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatewayValidatorsSetUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatewayValidatorsSetUpdated represents a ValidatorsSetUpdated event raised by the Gateway contract.
type GatewayValidatorsSetUpdated struct {
	Data []byte
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterValidatorsSetUpdated is a free log retrieval operation binding the contract event 0x77255af89c7a2379efc8c5869c76a372b4f2fd22756ba669541eb0d2380cc936.
//
// Solidity: event ValidatorsSetUpdated(bytes data)
func (_Gateway *GatewayFilterer) FilterValidatorsSetUpdated(opts *bind.FilterOpts) (*GatewayValidatorsSetUpdatedIterator, error) {

	logs, sub, err := _Gateway.contract.FilterLogs(opts, "ValidatorsSetUpdated")
	if err != nil {
		return nil, err
	}
	return &GatewayValidatorsSetUpdatedIterator{contract: _Gateway.contract, event: "ValidatorsSetUpdated", logs: logs, sub: sub}, nil
}

// WatchValidatorsSetUpdated is a free log subscription operation binding the contract event 0x77255af89c7a2379efc8c5869c76a372b4f2fd22756ba669541eb0d2380cc936.
//
// Solidity: event ValidatorsSetUpdated(bytes data)
func (_Gateway *GatewayFilterer) WatchValidatorsSetUpdated(opts *bind.WatchOpts, sink chan<- *GatewayValidatorsSetUpdated) (event.Subscription, error) {

	logs, sub, err := _Gateway.contract.WatchLogs(opts, "ValidatorsSetUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatewayValidatorsSetUpdated)
				if err := _Gateway.contract.UnpackLog(event, "ValidatorsSetUpdated", log); err != nil {
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

// ParseValidatorsSetUpdated is a log parse operation binding the contract event 0x77255af89c7a2379efc8c5869c76a372b4f2fd22756ba669541eb0d2380cc936.
//
// Solidity: event ValidatorsSetUpdated(bytes data)
func (_Gateway *GatewayFilterer) ParseValidatorsSetUpdated(log types.Log) (*GatewayValidatorsSetUpdated, error) {
	event := new(GatewayValidatorsSetUpdated)
	if err := _Gateway.contract.UnpackLog(event, "ValidatorsSetUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GatewayWithdrawIterator is returned from FilterWithdraw and is used to iterate over the raw logs and unpacked data for Withdraw events raised by the Gateway contract.
type GatewayWithdrawIterator struct {
	Event *GatewayWithdraw // Event containing the contract specifics and raw log

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
func (it *GatewayWithdrawIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatewayWithdraw)
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
		it.Event = new(GatewayWithdraw)
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
func (it *GatewayWithdrawIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatewayWithdrawIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatewayWithdraw represents a Withdraw event raised by the Gateway contract.
type GatewayWithdraw struct {
	DestinationChainId uint8
	Sender             common.Address
	Receivers          []IGatewayStructsReceiverWithdraw
	Fee                *big.Int
	OperationFee       *big.Int
	Value              *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterWithdraw is a free log retrieval operation binding the contract event 0x5346f1615d0f5d79989c2d9c7deb07d6a9e52196a209ec7abfbebabb8d346a69.
//
// Solidity: event Withdraw(uint8 destinationChainId, address sender, (string,uint256,uint16)[] receivers, uint256 fee, uint256 operationFee, uint256 value)
func (_Gateway *GatewayFilterer) FilterWithdraw(opts *bind.FilterOpts) (*GatewayWithdrawIterator, error) {

	logs, sub, err := _Gateway.contract.FilterLogs(opts, "Withdraw")
	if err != nil {
		return nil, err
	}
	return &GatewayWithdrawIterator{contract: _Gateway.contract, event: "Withdraw", logs: logs, sub: sub}, nil
}

// WatchWithdraw is a free log subscription operation binding the contract event 0x5346f1615d0f5d79989c2d9c7deb07d6a9e52196a209ec7abfbebabb8d346a69.
//
// Solidity: event Withdraw(uint8 destinationChainId, address sender, (string,uint256,uint16)[] receivers, uint256 fee, uint256 operationFee, uint256 value)
func (_Gateway *GatewayFilterer) WatchWithdraw(opts *bind.WatchOpts, sink chan<- *GatewayWithdraw) (event.Subscription, error) {

	logs, sub, err := _Gateway.contract.WatchLogs(opts, "Withdraw")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatewayWithdraw)
				if err := _Gateway.contract.UnpackLog(event, "Withdraw", log); err != nil {
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

// ParseWithdraw is a log parse operation binding the contract event 0x5346f1615d0f5d79989c2d9c7deb07d6a9e52196a209ec7abfbebabb8d346a69.
//
// Solidity: event Withdraw(uint8 destinationChainId, address sender, (string,uint256,uint16)[] receivers, uint256 fee, uint256 operationFee, uint256 value)
func (_Gateway *GatewayFilterer) ParseWithdraw(log types.Log) (*GatewayWithdraw, error) {
	event := new(GatewayWithdraw)
	if err := _Gateway.contract.UnpackLog(event, "Withdraw", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GatewayWithdrawTokenIterator is returned from FilterWithdrawToken and is used to iterate over the raw logs and unpacked data for WithdrawToken events raised by the Gateway contract.
type GatewayWithdrawTokenIterator struct {
	Event *GatewayWithdrawToken // Event containing the contract specifics and raw log

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
func (it *GatewayWithdrawTokenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatewayWithdrawToken)
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
		it.Event = new(GatewayWithdrawToken)
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
func (it *GatewayWithdrawTokenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatewayWithdrawTokenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatewayWithdrawToken represents a WithdrawToken event raised by the Gateway contract.
type GatewayWithdrawToken struct {
	Sender   common.Address
	Receiver common.Hash
	Amount   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterWithdrawToken is a free log retrieval operation binding the contract event 0x5f9265f1efffb7bef5c495c7bba35558ec47a79213e7e8aa5d3010308c0fda43.
//
// Solidity: event WithdrawToken(address sender, string indexed receiver, uint256 amount)
func (_Gateway *GatewayFilterer) FilterWithdrawToken(opts *bind.FilterOpts, receiver []string) (*GatewayWithdrawTokenIterator, error) {

	var receiverRule []interface{}
	for _, receiverItem := range receiver {
		receiverRule = append(receiverRule, receiverItem)
	}

	logs, sub, err := _Gateway.contract.FilterLogs(opts, "WithdrawToken", receiverRule)
	if err != nil {
		return nil, err
	}
	return &GatewayWithdrawTokenIterator{contract: _Gateway.contract, event: "WithdrawToken", logs: logs, sub: sub}, nil
}

// WatchWithdrawToken is a free log subscription operation binding the contract event 0x5f9265f1efffb7bef5c495c7bba35558ec47a79213e7e8aa5d3010308c0fda43.
//
// Solidity: event WithdrawToken(address sender, string indexed receiver, uint256 amount)
func (_Gateway *GatewayFilterer) WatchWithdrawToken(opts *bind.WatchOpts, sink chan<- *GatewayWithdrawToken, receiver []string) (event.Subscription, error) {

	var receiverRule []interface{}
	for _, receiverItem := range receiver {
		receiverRule = append(receiverRule, receiverItem)
	}

	logs, sub, err := _Gateway.contract.WatchLogs(opts, "WithdrawToken", receiverRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatewayWithdrawToken)
				if err := _Gateway.contract.UnpackLog(event, "WithdrawToken", log); err != nil {
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

// ParseWithdrawToken is a log parse operation binding the contract event 0x5f9265f1efffb7bef5c495c7bba35558ec47a79213e7e8aa5d3010308c0fda43.
//
// Solidity: event WithdrawToken(address sender, string indexed receiver, uint256 amount)
func (_Gateway *GatewayFilterer) ParseWithdrawToken(log types.Log) (*GatewayWithdrawToken, error) {
	event := new(GatewayWithdrawToken)
	if err := _Gateway.contract.UnpackLog(event, "WithdrawToken", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
