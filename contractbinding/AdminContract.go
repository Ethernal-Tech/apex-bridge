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

// AdminContractMetaData contains all meta data concerning the AdminContract contract.
var AdminContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"AddressEmptyCode\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_claimTransactionHash\",\"type\":\"bytes32\"}],\"name\":\"AlreadyConfirmed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_claimTransactionHash\",\"type\":\"uint8\"}],\"name\":\"AlreadyProposed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"}],\"name\":\"CanNotCreateBatchYet\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"}],\"name\":\"ChainAlreadyRegistered\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"}],\"name\":\"ChainIsNotRegistered\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"ERC1967InvalidImplementation\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ERC1967NonPayable\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FailedInnerCall\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"data\",\"type\":\"string\"}],\"name\":\"InvalidData\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSignature\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_availableAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_decreaseAmount\",\"type\":\"uint256\"}],\"name\":\"NegativeChainTokenAmount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotAdminContract\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotBridge\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotClaims\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_claimTransactionHash\",\"type\":\"bytes32\"}],\"name\":\"NotEnoughBridgingTokensAvailable\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSignedBatches\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSignedBatchesOrBridge\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSignedBatchesOrClaims\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotValidator\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"UUPSUnauthorizedCallContext\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"slot\",\"type\":\"bytes32\"}],\"name\":\"UUPSUnsupportedProxiableUUID\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"},{\"internalType\":\"uint64\",\"name\":\"_nonce\",\"type\":\"uint64\"}],\"name\":\"WrongBatchNonce\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"availableAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"withdrawalAmount\",\"type\":\"uint256\"}],\"name\":\"InsufficientFunds\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"claimeType\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"availableAmount\",\"type\":\"uint256\"}],\"name\":\"NotEnoughFunds\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"isIncrement\",\"type\":\"bool\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenQuantity\",\"type\":\"uint256\"}],\"name\":\"UpdatedChainTokenQuantity\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"Upgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"newChainProposal\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"}],\"name\":\"newChainRegistered\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"UPGRADE_INTERFACE_VERSION\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"}],\"name\":\"getChainTokenQuantity\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"proxiableUUID\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_claimsAddress\",\"type\":\"address\"}],\"name\":\"setDependencies\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"},{\"internalType\":\"bool\",\"name\":\"_isIncrease\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"_quantity\",\"type\":\"uint256\"}],\"name\":\"updateChainTokenQuantity\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"upgradeToAndCall\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
}

// AdminContractABI is the input ABI used to generate the binding from.
// Deprecated: Use AdminContractMetaData.ABI instead.
var AdminContractABI = AdminContractMetaData.ABI

// AdminContract is an auto generated Go binding around an Ethereum contract.
type AdminContract struct {
	AdminContractCaller     // Read-only binding to the contract
	AdminContractTransactor // Write-only binding to the contract
	AdminContractFilterer   // Log filterer for contract events
}

// AdminContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type AdminContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AdminContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AdminContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AdminContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AdminContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AdminContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AdminContractSession struct {
	Contract     *AdminContract    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AdminContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AdminContractCallerSession struct {
	Contract *AdminContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// AdminContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AdminContractTransactorSession struct {
	Contract     *AdminContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// AdminContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type AdminContractRaw struct {
	Contract *AdminContract // Generic contract binding to access the raw methods on
}

// AdminContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AdminContractCallerRaw struct {
	Contract *AdminContractCaller // Generic read-only contract binding to access the raw methods on
}

// AdminContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AdminContractTransactorRaw struct {
	Contract *AdminContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAdminContract creates a new instance of AdminContract, bound to a specific deployed contract.
func NewAdminContract(address common.Address, backend bind.ContractBackend) (*AdminContract, error) {
	contract, err := bindAdminContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AdminContract{AdminContractCaller: AdminContractCaller{contract: contract}, AdminContractTransactor: AdminContractTransactor{contract: contract}, AdminContractFilterer: AdminContractFilterer{contract: contract}}, nil
}

// NewAdminContractCaller creates a new read-only instance of AdminContract, bound to a specific deployed contract.
func NewAdminContractCaller(address common.Address, caller bind.ContractCaller) (*AdminContractCaller, error) {
	contract, err := bindAdminContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AdminContractCaller{contract: contract}, nil
}

// NewAdminContractTransactor creates a new write-only instance of AdminContract, bound to a specific deployed contract.
func NewAdminContractTransactor(address common.Address, transactor bind.ContractTransactor) (*AdminContractTransactor, error) {
	contract, err := bindAdminContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AdminContractTransactor{contract: contract}, nil
}

// NewAdminContractFilterer creates a new log filterer instance of AdminContract, bound to a specific deployed contract.
func NewAdminContractFilterer(address common.Address, filterer bind.ContractFilterer) (*AdminContractFilterer, error) {
	contract, err := bindAdminContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AdminContractFilterer{contract: contract}, nil
}

// bindAdminContract binds a generic wrapper to an already deployed contract.
func bindAdminContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := AdminContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AdminContract *AdminContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AdminContract.Contract.AdminContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AdminContract *AdminContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AdminContract.Contract.AdminContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AdminContract *AdminContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AdminContract.Contract.AdminContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AdminContract *AdminContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AdminContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AdminContract *AdminContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AdminContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AdminContract *AdminContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AdminContract.Contract.contract.Transact(opts, method, params...)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_AdminContract *AdminContractCaller) UPGRADEINTERFACEVERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _AdminContract.contract.Call(opts, &out, "UPGRADE_INTERFACE_VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_AdminContract *AdminContractSession) UPGRADEINTERFACEVERSION() (string, error) {
	return _AdminContract.Contract.UPGRADEINTERFACEVERSION(&_AdminContract.CallOpts)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_AdminContract *AdminContractCallerSession) UPGRADEINTERFACEVERSION() (string, error) {
	return _AdminContract.Contract.UPGRADEINTERFACEVERSION(&_AdminContract.CallOpts)
}

// GetChainTokenQuantity is a free data retrieval call binding the contract method 0x14da8531.
//
// Solidity: function getChainTokenQuantity(uint8 _chainId) view returns(uint256)
func (_AdminContract *AdminContractCaller) GetChainTokenQuantity(opts *bind.CallOpts, _chainId uint8) (*big.Int, error) {
	var out []interface{}
	err := _AdminContract.contract.Call(opts, &out, "getChainTokenQuantity", _chainId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetChainTokenQuantity is a free data retrieval call binding the contract method 0x14da8531.
//
// Solidity: function getChainTokenQuantity(uint8 _chainId) view returns(uint256)
func (_AdminContract *AdminContractSession) GetChainTokenQuantity(_chainId uint8) (*big.Int, error) {
	return _AdminContract.Contract.GetChainTokenQuantity(&_AdminContract.CallOpts, _chainId)
}

// GetChainTokenQuantity is a free data retrieval call binding the contract method 0x14da8531.
//
// Solidity: function getChainTokenQuantity(uint8 _chainId) view returns(uint256)
func (_AdminContract *AdminContractCallerSession) GetChainTokenQuantity(_chainId uint8) (*big.Int, error) {
	return _AdminContract.Contract.GetChainTokenQuantity(&_AdminContract.CallOpts, _chainId)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AdminContract *AdminContractCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AdminContract.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AdminContract *AdminContractSession) Owner() (common.Address, error) {
	return _AdminContract.Contract.Owner(&_AdminContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AdminContract *AdminContractCallerSession) Owner() (common.Address, error) {
	return _AdminContract.Contract.Owner(&_AdminContract.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_AdminContract *AdminContractCaller) ProxiableUUID(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _AdminContract.contract.Call(opts, &out, "proxiableUUID")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_AdminContract *AdminContractSession) ProxiableUUID() ([32]byte, error) {
	return _AdminContract.Contract.ProxiableUUID(&_AdminContract.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_AdminContract *AdminContractCallerSession) ProxiableUUID() ([32]byte, error) {
	return _AdminContract.Contract.ProxiableUUID(&_AdminContract.CallOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _owner) returns()
func (_AdminContract *AdminContractTransactor) Initialize(opts *bind.TransactOpts, _owner common.Address) (*types.Transaction, error) {
	return _AdminContract.contract.Transact(opts, "initialize", _owner)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _owner) returns()
func (_AdminContract *AdminContractSession) Initialize(_owner common.Address) (*types.Transaction, error) {
	return _AdminContract.Contract.Initialize(&_AdminContract.TransactOpts, _owner)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _owner) returns()
func (_AdminContract *AdminContractTransactorSession) Initialize(_owner common.Address) (*types.Transaction, error) {
	return _AdminContract.Contract.Initialize(&_AdminContract.TransactOpts, _owner)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AdminContract *AdminContractTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AdminContract.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AdminContract *AdminContractSession) RenounceOwnership() (*types.Transaction, error) {
	return _AdminContract.Contract.RenounceOwnership(&_AdminContract.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AdminContract *AdminContractTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _AdminContract.Contract.RenounceOwnership(&_AdminContract.TransactOpts)
}

// SetDependencies is a paid mutator transaction binding the contract method 0x8389cb18.
//
// Solidity: function setDependencies(address _claimsAddress) returns()
func (_AdminContract *AdminContractTransactor) SetDependencies(opts *bind.TransactOpts, _claimsAddress common.Address) (*types.Transaction, error) {
	return _AdminContract.contract.Transact(opts, "setDependencies", _claimsAddress)
}

// SetDependencies is a paid mutator transaction binding the contract method 0x8389cb18.
//
// Solidity: function setDependencies(address _claimsAddress) returns()
func (_AdminContract *AdminContractSession) SetDependencies(_claimsAddress common.Address) (*types.Transaction, error) {
	return _AdminContract.Contract.SetDependencies(&_AdminContract.TransactOpts, _claimsAddress)
}

// SetDependencies is a paid mutator transaction binding the contract method 0x8389cb18.
//
// Solidity: function setDependencies(address _claimsAddress) returns()
func (_AdminContract *AdminContractTransactorSession) SetDependencies(_claimsAddress common.Address) (*types.Transaction, error) {
	return _AdminContract.Contract.SetDependencies(&_AdminContract.TransactOpts, _claimsAddress)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AdminContract *AdminContractTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _AdminContract.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AdminContract *AdminContractSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _AdminContract.Contract.TransferOwnership(&_AdminContract.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AdminContract *AdminContractTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _AdminContract.Contract.TransferOwnership(&_AdminContract.TransactOpts, newOwner)
}

// UpdateChainTokenQuantity is a paid mutator transaction binding the contract method 0x0504334f.
//
// Solidity: function updateChainTokenQuantity(uint8 _chainId, bool _isIncrease, uint256 _quantity) returns()
func (_AdminContract *AdminContractTransactor) UpdateChainTokenQuantity(opts *bind.TransactOpts, _chainId uint8, _isIncrease bool, _quantity *big.Int) (*types.Transaction, error) {
	return _AdminContract.contract.Transact(opts, "updateChainTokenQuantity", _chainId, _isIncrease, _quantity)
}

// UpdateChainTokenQuantity is a paid mutator transaction binding the contract method 0x0504334f.
//
// Solidity: function updateChainTokenQuantity(uint8 _chainId, bool _isIncrease, uint256 _quantity) returns()
func (_AdminContract *AdminContractSession) UpdateChainTokenQuantity(_chainId uint8, _isIncrease bool, _quantity *big.Int) (*types.Transaction, error) {
	return _AdminContract.Contract.UpdateChainTokenQuantity(&_AdminContract.TransactOpts, _chainId, _isIncrease, _quantity)
}

// UpdateChainTokenQuantity is a paid mutator transaction binding the contract method 0x0504334f.
//
// Solidity: function updateChainTokenQuantity(uint8 _chainId, bool _isIncrease, uint256 _quantity) returns()
func (_AdminContract *AdminContractTransactorSession) UpdateChainTokenQuantity(_chainId uint8, _isIncrease bool, _quantity *big.Int) (*types.Transaction, error) {
	return _AdminContract.Contract.UpdateChainTokenQuantity(&_AdminContract.TransactOpts, _chainId, _isIncrease, _quantity)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_AdminContract *AdminContractTransactor) UpgradeToAndCall(opts *bind.TransactOpts, newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _AdminContract.contract.Transact(opts, "upgradeToAndCall", newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_AdminContract *AdminContractSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _AdminContract.Contract.UpgradeToAndCall(&_AdminContract.TransactOpts, newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_AdminContract *AdminContractTransactorSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _AdminContract.Contract.UpgradeToAndCall(&_AdminContract.TransactOpts, newImplementation, data)
}

// AdminContractInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the AdminContract contract.
type AdminContractInitializedIterator struct {
	Event *AdminContractInitialized // Event containing the contract specifics and raw log

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
func (it *AdminContractInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdminContractInitialized)
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
		it.Event = new(AdminContractInitialized)
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
func (it *AdminContractInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdminContractInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdminContractInitialized represents a Initialized event raised by the AdminContract contract.
type AdminContractInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_AdminContract *AdminContractFilterer) FilterInitialized(opts *bind.FilterOpts) (*AdminContractInitializedIterator, error) {

	logs, sub, err := _AdminContract.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &AdminContractInitializedIterator{contract: _AdminContract.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_AdminContract *AdminContractFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *AdminContractInitialized) (event.Subscription, error) {

	logs, sub, err := _AdminContract.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdminContractInitialized)
				if err := _AdminContract.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_AdminContract *AdminContractFilterer) ParseInitialized(log types.Log) (*AdminContractInitialized, error) {
	event := new(AdminContractInitialized)
	if err := _AdminContract.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AdminContractInsufficientFundsIterator is returned from FilterInsufficientFunds and is used to iterate over the raw logs and unpacked data for InsufficientFunds events raised by the AdminContract contract.
type AdminContractInsufficientFundsIterator struct {
	Event *AdminContractInsufficientFunds // Event containing the contract specifics and raw log

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
func (it *AdminContractInsufficientFundsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdminContractInsufficientFunds)
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
		it.Event = new(AdminContractInsufficientFunds)
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
func (it *AdminContractInsufficientFundsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdminContractInsufficientFundsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdminContractInsufficientFunds represents a InsufficientFunds event raised by the AdminContract contract.
type AdminContractInsufficientFunds struct {
	AvailableAmount  *big.Int
	WithdrawalAmount *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterInsufficientFunds is a free log retrieval operation binding the contract event 0x03eb8b54a949acec2cd08fdb6d6bd4647a1f2c907d75d6900648effa92eb147f.
//
// Solidity: event InsufficientFunds(uint256 availableAmount, uint256 withdrawalAmount)
func (_AdminContract *AdminContractFilterer) FilterInsufficientFunds(opts *bind.FilterOpts) (*AdminContractInsufficientFundsIterator, error) {

	logs, sub, err := _AdminContract.contract.FilterLogs(opts, "InsufficientFunds")
	if err != nil {
		return nil, err
	}
	return &AdminContractInsufficientFundsIterator{contract: _AdminContract.contract, event: "InsufficientFunds", logs: logs, sub: sub}, nil
}

// WatchInsufficientFunds is a free log subscription operation binding the contract event 0x03eb8b54a949acec2cd08fdb6d6bd4647a1f2c907d75d6900648effa92eb147f.
//
// Solidity: event InsufficientFunds(uint256 availableAmount, uint256 withdrawalAmount)
func (_AdminContract *AdminContractFilterer) WatchInsufficientFunds(opts *bind.WatchOpts, sink chan<- *AdminContractInsufficientFunds) (event.Subscription, error) {

	logs, sub, err := _AdminContract.contract.WatchLogs(opts, "InsufficientFunds")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdminContractInsufficientFunds)
				if err := _AdminContract.contract.UnpackLog(event, "InsufficientFunds", log); err != nil {
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

// ParseInsufficientFunds is a log parse operation binding the contract event 0x03eb8b54a949acec2cd08fdb6d6bd4647a1f2c907d75d6900648effa92eb147f.
//
// Solidity: event InsufficientFunds(uint256 availableAmount, uint256 withdrawalAmount)
func (_AdminContract *AdminContractFilterer) ParseInsufficientFunds(log types.Log) (*AdminContractInsufficientFunds, error) {
	event := new(AdminContractInsufficientFunds)
	if err := _AdminContract.contract.UnpackLog(event, "InsufficientFunds", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AdminContractNotEnoughFundsIterator is returned from FilterNotEnoughFunds and is used to iterate over the raw logs and unpacked data for NotEnoughFunds events raised by the AdminContract contract.
type AdminContractNotEnoughFundsIterator struct {
	Event *AdminContractNotEnoughFunds // Event containing the contract specifics and raw log

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
func (it *AdminContractNotEnoughFundsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdminContractNotEnoughFunds)
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
		it.Event = new(AdminContractNotEnoughFunds)
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
func (it *AdminContractNotEnoughFundsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdminContractNotEnoughFundsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdminContractNotEnoughFunds represents a NotEnoughFunds event raised by the AdminContract contract.
type AdminContractNotEnoughFunds struct {
	ClaimeType      string
	Index           *big.Int
	AvailableAmount *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterNotEnoughFunds is a free log retrieval operation binding the contract event 0xfc533f3a2de16e1c5a6c03094c865069fabdd71ee6a9af63918823a35a55def5.
//
// Solidity: event NotEnoughFunds(string claimeType, uint256 index, uint256 availableAmount)
func (_AdminContract *AdminContractFilterer) FilterNotEnoughFunds(opts *bind.FilterOpts) (*AdminContractNotEnoughFundsIterator, error) {

	logs, sub, err := _AdminContract.contract.FilterLogs(opts, "NotEnoughFunds")
	if err != nil {
		return nil, err
	}
	return &AdminContractNotEnoughFundsIterator{contract: _AdminContract.contract, event: "NotEnoughFunds", logs: logs, sub: sub}, nil
}

// WatchNotEnoughFunds is a free log subscription operation binding the contract event 0xfc533f3a2de16e1c5a6c03094c865069fabdd71ee6a9af63918823a35a55def5.
//
// Solidity: event NotEnoughFunds(string claimeType, uint256 index, uint256 availableAmount)
func (_AdminContract *AdminContractFilterer) WatchNotEnoughFunds(opts *bind.WatchOpts, sink chan<- *AdminContractNotEnoughFunds) (event.Subscription, error) {

	logs, sub, err := _AdminContract.contract.WatchLogs(opts, "NotEnoughFunds")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdminContractNotEnoughFunds)
				if err := _AdminContract.contract.UnpackLog(event, "NotEnoughFunds", log); err != nil {
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
func (_AdminContract *AdminContractFilterer) ParseNotEnoughFunds(log types.Log) (*AdminContractNotEnoughFunds, error) {
	event := new(AdminContractNotEnoughFunds)
	if err := _AdminContract.contract.UnpackLog(event, "NotEnoughFunds", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AdminContractOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the AdminContract contract.
type AdminContractOwnershipTransferredIterator struct {
	Event *AdminContractOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *AdminContractOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdminContractOwnershipTransferred)
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
		it.Event = new(AdminContractOwnershipTransferred)
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
func (it *AdminContractOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdminContractOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdminContractOwnershipTransferred represents a OwnershipTransferred event raised by the AdminContract contract.
type AdminContractOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_AdminContract *AdminContractFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*AdminContractOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _AdminContract.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &AdminContractOwnershipTransferredIterator{contract: _AdminContract.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_AdminContract *AdminContractFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *AdminContractOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _AdminContract.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdminContractOwnershipTransferred)
				if err := _AdminContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_AdminContract *AdminContractFilterer) ParseOwnershipTransferred(log types.Log) (*AdminContractOwnershipTransferred, error) {
	event := new(AdminContractOwnershipTransferred)
	if err := _AdminContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AdminContractUpdatedChainTokenQuantityIterator is returned from FilterUpdatedChainTokenQuantity and is used to iterate over the raw logs and unpacked data for UpdatedChainTokenQuantity events raised by the AdminContract contract.
type AdminContractUpdatedChainTokenQuantityIterator struct {
	Event *AdminContractUpdatedChainTokenQuantity // Event containing the contract specifics and raw log

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
func (it *AdminContractUpdatedChainTokenQuantityIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdminContractUpdatedChainTokenQuantity)
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
		it.Event = new(AdminContractUpdatedChainTokenQuantity)
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
func (it *AdminContractUpdatedChainTokenQuantityIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdminContractUpdatedChainTokenQuantityIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdminContractUpdatedChainTokenQuantity represents a UpdatedChainTokenQuantity event raised by the AdminContract contract.
type AdminContractUpdatedChainTokenQuantity struct {
	ChainId       *big.Int
	IsIncrement   bool
	TokenQuantity *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterUpdatedChainTokenQuantity is a free log retrieval operation binding the contract event 0x63a8310d54d22ae170c9c99ec0101494848847baf8ba54b1f297456f4c01bd62.
//
// Solidity: event UpdatedChainTokenQuantity(uint256 indexed chainId, bool isIncrement, uint256 tokenQuantity)
func (_AdminContract *AdminContractFilterer) FilterUpdatedChainTokenQuantity(opts *bind.FilterOpts, chainId []*big.Int) (*AdminContractUpdatedChainTokenQuantityIterator, error) {

	var chainIdRule []interface{}
	for _, chainIdItem := range chainId {
		chainIdRule = append(chainIdRule, chainIdItem)
	}

	logs, sub, err := _AdminContract.contract.FilterLogs(opts, "UpdatedChainTokenQuantity", chainIdRule)
	if err != nil {
		return nil, err
	}
	return &AdminContractUpdatedChainTokenQuantityIterator{contract: _AdminContract.contract, event: "UpdatedChainTokenQuantity", logs: logs, sub: sub}, nil
}

// WatchUpdatedChainTokenQuantity is a free log subscription operation binding the contract event 0x63a8310d54d22ae170c9c99ec0101494848847baf8ba54b1f297456f4c01bd62.
//
// Solidity: event UpdatedChainTokenQuantity(uint256 indexed chainId, bool isIncrement, uint256 tokenQuantity)
func (_AdminContract *AdminContractFilterer) WatchUpdatedChainTokenQuantity(opts *bind.WatchOpts, sink chan<- *AdminContractUpdatedChainTokenQuantity, chainId []*big.Int) (event.Subscription, error) {

	var chainIdRule []interface{}
	for _, chainIdItem := range chainId {
		chainIdRule = append(chainIdRule, chainIdItem)
	}

	logs, sub, err := _AdminContract.contract.WatchLogs(opts, "UpdatedChainTokenQuantity", chainIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdminContractUpdatedChainTokenQuantity)
				if err := _AdminContract.contract.UnpackLog(event, "UpdatedChainTokenQuantity", log); err != nil {
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
func (_AdminContract *AdminContractFilterer) ParseUpdatedChainTokenQuantity(log types.Log) (*AdminContractUpdatedChainTokenQuantity, error) {
	event := new(AdminContractUpdatedChainTokenQuantity)
	if err := _AdminContract.contract.UnpackLog(event, "UpdatedChainTokenQuantity", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AdminContractUpgradedIterator is returned from FilterUpgraded and is used to iterate over the raw logs and unpacked data for Upgraded events raised by the AdminContract contract.
type AdminContractUpgradedIterator struct {
	Event *AdminContractUpgraded // Event containing the contract specifics and raw log

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
func (it *AdminContractUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdminContractUpgraded)
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
		it.Event = new(AdminContractUpgraded)
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
func (it *AdminContractUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdminContractUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdminContractUpgraded represents a Upgraded event raised by the AdminContract contract.
type AdminContractUpgraded struct {
	Implementation common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterUpgraded is a free log retrieval operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_AdminContract *AdminContractFilterer) FilterUpgraded(opts *bind.FilterOpts, implementation []common.Address) (*AdminContractUpgradedIterator, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _AdminContract.contract.FilterLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return &AdminContractUpgradedIterator{contract: _AdminContract.contract, event: "Upgraded", logs: logs, sub: sub}, nil
}

// WatchUpgraded is a free log subscription operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_AdminContract *AdminContractFilterer) WatchUpgraded(opts *bind.WatchOpts, sink chan<- *AdminContractUpgraded, implementation []common.Address) (event.Subscription, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _AdminContract.contract.WatchLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdminContractUpgraded)
				if err := _AdminContract.contract.UnpackLog(event, "Upgraded", log); err != nil {
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
func (_AdminContract *AdminContractFilterer) ParseUpgraded(log types.Log) (*AdminContractUpgraded, error) {
	event := new(AdminContractUpgraded)
	if err := _AdminContract.contract.UnpackLog(event, "Upgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AdminContractNewChainProposalIterator is returned from FilterNewChainProposal and is used to iterate over the raw logs and unpacked data for NewChainProposal events raised by the AdminContract contract.
type AdminContractNewChainProposalIterator struct {
	Event *AdminContractNewChainProposal // Event containing the contract specifics and raw log

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
func (it *AdminContractNewChainProposalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdminContractNewChainProposal)
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
		it.Event = new(AdminContractNewChainProposal)
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
func (it *AdminContractNewChainProposalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdminContractNewChainProposalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdminContractNewChainProposal represents a NewChainProposal event raised by the AdminContract contract.
type AdminContractNewChainProposal struct {
	ChainId uint8
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNewChainProposal is a free log retrieval operation binding the contract event 0xc546bc51d95705dd957ce30962375555dda4421e2004fbbf0b5e1527858f6c30.
//
// Solidity: event newChainProposal(uint8 indexed _chainId, address indexed sender)
func (_AdminContract *AdminContractFilterer) FilterNewChainProposal(opts *bind.FilterOpts, _chainId []uint8, sender []common.Address) (*AdminContractNewChainProposalIterator, error) {

	var _chainIdRule []interface{}
	for _, _chainIdItem := range _chainId {
		_chainIdRule = append(_chainIdRule, _chainIdItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _AdminContract.contract.FilterLogs(opts, "newChainProposal", _chainIdRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &AdminContractNewChainProposalIterator{contract: _AdminContract.contract, event: "newChainProposal", logs: logs, sub: sub}, nil
}

// WatchNewChainProposal is a free log subscription operation binding the contract event 0xc546bc51d95705dd957ce30962375555dda4421e2004fbbf0b5e1527858f6c30.
//
// Solidity: event newChainProposal(uint8 indexed _chainId, address indexed sender)
func (_AdminContract *AdminContractFilterer) WatchNewChainProposal(opts *bind.WatchOpts, sink chan<- *AdminContractNewChainProposal, _chainId []uint8, sender []common.Address) (event.Subscription, error) {

	var _chainIdRule []interface{}
	for _, _chainIdItem := range _chainId {
		_chainIdRule = append(_chainIdRule, _chainIdItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _AdminContract.contract.WatchLogs(opts, "newChainProposal", _chainIdRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdminContractNewChainProposal)
				if err := _AdminContract.contract.UnpackLog(event, "newChainProposal", log); err != nil {
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
func (_AdminContract *AdminContractFilterer) ParseNewChainProposal(log types.Log) (*AdminContractNewChainProposal, error) {
	event := new(AdminContractNewChainProposal)
	if err := _AdminContract.contract.UnpackLog(event, "newChainProposal", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AdminContractNewChainRegisteredIterator is returned from FilterNewChainRegistered and is used to iterate over the raw logs and unpacked data for NewChainRegistered events raised by the AdminContract contract.
type AdminContractNewChainRegisteredIterator struct {
	Event *AdminContractNewChainRegistered // Event containing the contract specifics and raw log

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
func (it *AdminContractNewChainRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdminContractNewChainRegistered)
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
		it.Event = new(AdminContractNewChainRegistered)
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
func (it *AdminContractNewChainRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdminContractNewChainRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdminContractNewChainRegistered represents a NewChainRegistered event raised by the AdminContract contract.
type AdminContractNewChainRegistered struct {
	ChainId uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNewChainRegistered is a free log retrieval operation binding the contract event 0x8541a0c729e909924d8678df4b2374f63c9514fcd5a430ac3de033d11d120256.
//
// Solidity: event newChainRegistered(uint8 indexed _chainId)
func (_AdminContract *AdminContractFilterer) FilterNewChainRegistered(opts *bind.FilterOpts, _chainId []uint8) (*AdminContractNewChainRegisteredIterator, error) {

	var _chainIdRule []interface{}
	for _, _chainIdItem := range _chainId {
		_chainIdRule = append(_chainIdRule, _chainIdItem)
	}

	logs, sub, err := _AdminContract.contract.FilterLogs(opts, "newChainRegistered", _chainIdRule)
	if err != nil {
		return nil, err
	}
	return &AdminContractNewChainRegisteredIterator{contract: _AdminContract.contract, event: "newChainRegistered", logs: logs, sub: sub}, nil
}

// WatchNewChainRegistered is a free log subscription operation binding the contract event 0x8541a0c729e909924d8678df4b2374f63c9514fcd5a430ac3de033d11d120256.
//
// Solidity: event newChainRegistered(uint8 indexed _chainId)
func (_AdminContract *AdminContractFilterer) WatchNewChainRegistered(opts *bind.WatchOpts, sink chan<- *AdminContractNewChainRegistered, _chainId []uint8) (event.Subscription, error) {

	var _chainIdRule []interface{}
	for _, _chainIdItem := range _chainId {
		_chainIdRule = append(_chainIdRule, _chainIdItem)
	}

	logs, sub, err := _AdminContract.contract.WatchLogs(opts, "newChainRegistered", _chainIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdminContractNewChainRegistered)
				if err := _AdminContract.contract.UnpackLog(event, "newChainRegistered", log); err != nil {
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
func (_AdminContract *AdminContractFilterer) ParseNewChainRegistered(log types.Log) (*AdminContractNewChainRegistered, error) {
	event := new(AdminContractNewChainRegistered)
	if err := _AdminContract.contract.UnpackLog(event, "newChainRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}