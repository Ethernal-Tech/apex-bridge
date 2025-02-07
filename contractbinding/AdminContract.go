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
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_claimTransactionHash\",\"type\":\"bytes32\"}],\"name\":\"AlreadyConfirmed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_claimTransactionHash\",\"type\":\"uint8\"}],\"name\":\"AlreadyProposed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"}],\"name\":\"CanNotCreateBatchYet\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"}],\"name\":\"ChainAlreadyRegistered\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"}],\"name\":\"ChainIsNotRegistered\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"_availableAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_requestedAmount\",\"type\":\"uint256\"}],\"name\":\"DefundRequestTooHigh\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"data\",\"type\":\"string\"}],\"name\":\"InvalidData\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSignature\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_availableAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_decreaseAmount\",\"type\":\"uint256\"}],\"name\":\"NegativeChainTokenAmount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotAdminContract\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotBridge\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotClaims\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_claimTransactionHash\",\"type\":\"bytes32\"}],\"name\":\"NotEnoughBridgingTokensAvailable\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotFundAdmin\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSignedBatches\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSignedBatchesOrBridge\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSignedBatchesOrClaims\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotUpgradeAdmin\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotValidator\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"},{\"internalType\":\"uint64\",\"name\":\"_nonce\",\"type\":\"uint64\"}],\"name\":\"WrongBatchNonce\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroAddress\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"previousAdmin\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"AdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"beacon\",\"type\":\"address\"}],\"name\":\"BeaconUpgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"ChainDefunded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"DefundFailedAfterMultipleRetries\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_newFundAdmin\",\"type\":\"address\"}],\"name\":\"FundAdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"claimeType\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"availableAmount\",\"type\":\"uint256\"}],\"name\":\"NotEnoughFunds\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"isIncrement\",\"type\":\"bool\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"chainTokenQuantity\",\"type\":\"uint256\"}],\"name\":\"UpdatedChainTokenQuantity\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"isIncrement\",\"type\":\"bool\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"chainWrappedTokenQuantity\",\"type\":\"uint256\"}],\"name\":\"UpdatedChainWrappedTokenQuantity\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_maxNumberOfTransactions\",\"type\":\"uint256\"}],\"name\":\"UpdatedMaxNumberOfTransactions\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_timeoutBlocksNumber\",\"type\":\"uint256\"}],\"name\":\"UpdatedTimeoutBlocksNumber\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"Upgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"newChainProposal\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"}],\"name\":\"newChainRegistered\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_amountWrapped\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"_defundAddress\",\"type\":\"string\"}],\"name\":\"defund\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"fundAdmin\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"}],\"name\":\"getChainTokenQuantity\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"}],\"name\":\"getChainWrappedTokenQuantity\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_upgradeAdmin\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"proxiableUUID\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_claimsAddress\",\"type\":\"address\"}],\"name\":\"setDependencies\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_fundAdmin\",\"type\":\"address\"}],\"name\":\"setFundAdmin\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"},{\"internalType\":\"bool\",\"name\":\"_isIncrease\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"_chainTokenQuantity\",\"type\":\"uint256\"}],\"name\":\"updateChainTokenQuantity\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_chainId\",\"type\":\"uint8\"},{\"internalType\":\"bool\",\"name\":\"_isIncrease\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"_chainWrappedTokenQuantity\",\"type\":\"uint256\"}],\"name\":\"updateChainWrappedTokenQuantity\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"_maxNumberOfTransactions\",\"type\":\"uint16\"}],\"name\":\"updateMaxNumberOfTransactions\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_timeoutBlocksNumber\",\"type\":\"uint8\"}],\"name\":\"updateTimeoutBlocksNumber\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"}],\"name\":\"upgradeTo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"upgradeToAndCall\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
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

// FundAdmin is a free data retrieval call binding the contract method 0x9c456583.
//
// Solidity: function fundAdmin() view returns(address)
func (_AdminContract *AdminContractCaller) FundAdmin(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AdminContract.contract.Call(opts, &out, "fundAdmin")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// FundAdmin is a free data retrieval call binding the contract method 0x9c456583.
//
// Solidity: function fundAdmin() view returns(address)
func (_AdminContract *AdminContractSession) FundAdmin() (common.Address, error) {
	return _AdminContract.Contract.FundAdmin(&_AdminContract.CallOpts)
}

// FundAdmin is a free data retrieval call binding the contract method 0x9c456583.
//
// Solidity: function fundAdmin() view returns(address)
func (_AdminContract *AdminContractCallerSession) FundAdmin() (common.Address, error) {
	return _AdminContract.Contract.FundAdmin(&_AdminContract.CallOpts)
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

// GetChainWrappedTokenQuantity is a free data retrieval call binding the contract method 0x731cc65c.
//
// Solidity: function getChainWrappedTokenQuantity(uint8 _chainId) view returns(uint256)
func (_AdminContract *AdminContractCaller) GetChainWrappedTokenQuantity(opts *bind.CallOpts, _chainId uint8) (*big.Int, error) {
	var out []interface{}
	err := _AdminContract.contract.Call(opts, &out, "getChainWrappedTokenQuantity", _chainId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetChainWrappedTokenQuantity is a free data retrieval call binding the contract method 0x731cc65c.
//
// Solidity: function getChainWrappedTokenQuantity(uint8 _chainId) view returns(uint256)
func (_AdminContract *AdminContractSession) GetChainWrappedTokenQuantity(_chainId uint8) (*big.Int, error) {
	return _AdminContract.Contract.GetChainWrappedTokenQuantity(&_AdminContract.CallOpts, _chainId)
}

// GetChainWrappedTokenQuantity is a free data retrieval call binding the contract method 0x731cc65c.
//
// Solidity: function getChainWrappedTokenQuantity(uint8 _chainId) view returns(uint256)
func (_AdminContract *AdminContractCallerSession) GetChainWrappedTokenQuantity(_chainId uint8) (*big.Int, error) {
	return _AdminContract.Contract.GetChainWrappedTokenQuantity(&_AdminContract.CallOpts, _chainId)
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

// Defund is a paid mutator transaction binding the contract method 0x77aa493a.
//
// Solidity: function defund(uint8 _chainId, uint256 _amount, uint256 _amountWrapped, string _defundAddress) returns()
func (_AdminContract *AdminContractTransactor) Defund(opts *bind.TransactOpts, _chainId uint8, _amount *big.Int, _amountWrapped *big.Int, _defundAddress string) (*types.Transaction, error) {
	return _AdminContract.contract.Transact(opts, "defund", _chainId, _amount, _amountWrapped, _defundAddress)
}

// Defund is a paid mutator transaction binding the contract method 0x77aa493a.
//
// Solidity: function defund(uint8 _chainId, uint256 _amount, uint256 _amountWrapped, string _defundAddress) returns()
func (_AdminContract *AdminContractSession) Defund(_chainId uint8, _amount *big.Int, _amountWrapped *big.Int, _defundAddress string) (*types.Transaction, error) {
	return _AdminContract.Contract.Defund(&_AdminContract.TransactOpts, _chainId, _amount, _amountWrapped, _defundAddress)
}

// Defund is a paid mutator transaction binding the contract method 0x77aa493a.
//
// Solidity: function defund(uint8 _chainId, uint256 _amount, uint256 _amountWrapped, string _defundAddress) returns()
func (_AdminContract *AdminContractTransactorSession) Defund(_chainId uint8, _amount *big.Int, _amountWrapped *big.Int, _defundAddress string) (*types.Transaction, error) {
	return _AdminContract.Contract.Defund(&_AdminContract.TransactOpts, _chainId, _amount, _amountWrapped, _defundAddress)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address _owner, address _upgradeAdmin) returns()
func (_AdminContract *AdminContractTransactor) Initialize(opts *bind.TransactOpts, _owner common.Address, _upgradeAdmin common.Address) (*types.Transaction, error) {
	return _AdminContract.contract.Transact(opts, "initialize", _owner, _upgradeAdmin)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address _owner, address _upgradeAdmin) returns()
func (_AdminContract *AdminContractSession) Initialize(_owner common.Address, _upgradeAdmin common.Address) (*types.Transaction, error) {
	return _AdminContract.Contract.Initialize(&_AdminContract.TransactOpts, _owner, _upgradeAdmin)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address _owner, address _upgradeAdmin) returns()
func (_AdminContract *AdminContractTransactorSession) Initialize(_owner common.Address, _upgradeAdmin common.Address) (*types.Transaction, error) {
	return _AdminContract.Contract.Initialize(&_AdminContract.TransactOpts, _owner, _upgradeAdmin)
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

// SetFundAdmin is a paid mutator transaction binding the contract method 0xa96d1edc.
//
// Solidity: function setFundAdmin(address _fundAdmin) returns()
func (_AdminContract *AdminContractTransactor) SetFundAdmin(opts *bind.TransactOpts, _fundAdmin common.Address) (*types.Transaction, error) {
	return _AdminContract.contract.Transact(opts, "setFundAdmin", _fundAdmin)
}

// SetFundAdmin is a paid mutator transaction binding the contract method 0xa96d1edc.
//
// Solidity: function setFundAdmin(address _fundAdmin) returns()
func (_AdminContract *AdminContractSession) SetFundAdmin(_fundAdmin common.Address) (*types.Transaction, error) {
	return _AdminContract.Contract.SetFundAdmin(&_AdminContract.TransactOpts, _fundAdmin)
}

// SetFundAdmin is a paid mutator transaction binding the contract method 0xa96d1edc.
//
// Solidity: function setFundAdmin(address _fundAdmin) returns()
func (_AdminContract *AdminContractTransactorSession) SetFundAdmin(_fundAdmin common.Address) (*types.Transaction, error) {
	return _AdminContract.Contract.SetFundAdmin(&_AdminContract.TransactOpts, _fundAdmin)
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
// Solidity: function updateChainTokenQuantity(uint8 _chainId, bool _isIncrease, uint256 _chainTokenQuantity) returns()
func (_AdminContract *AdminContractTransactor) UpdateChainTokenQuantity(opts *bind.TransactOpts, _chainId uint8, _isIncrease bool, _chainTokenQuantity *big.Int) (*types.Transaction, error) {
	return _AdminContract.contract.Transact(opts, "updateChainTokenQuantity", _chainId, _isIncrease, _chainTokenQuantity)
}

// UpdateChainTokenQuantity is a paid mutator transaction binding the contract method 0x0504334f.
//
// Solidity: function updateChainTokenQuantity(uint8 _chainId, bool _isIncrease, uint256 _chainTokenQuantity) returns()
func (_AdminContract *AdminContractSession) UpdateChainTokenQuantity(_chainId uint8, _isIncrease bool, _chainTokenQuantity *big.Int) (*types.Transaction, error) {
	return _AdminContract.Contract.UpdateChainTokenQuantity(&_AdminContract.TransactOpts, _chainId, _isIncrease, _chainTokenQuantity)
}

// UpdateChainTokenQuantity is a paid mutator transaction binding the contract method 0x0504334f.
//
// Solidity: function updateChainTokenQuantity(uint8 _chainId, bool _isIncrease, uint256 _chainTokenQuantity) returns()
func (_AdminContract *AdminContractTransactorSession) UpdateChainTokenQuantity(_chainId uint8, _isIncrease bool, _chainTokenQuantity *big.Int) (*types.Transaction, error) {
	return _AdminContract.Contract.UpdateChainTokenQuantity(&_AdminContract.TransactOpts, _chainId, _isIncrease, _chainTokenQuantity)
}

// UpdateChainWrappedTokenQuantity is a paid mutator transaction binding the contract method 0x170eb3ff.
//
// Solidity: function updateChainWrappedTokenQuantity(uint8 _chainId, bool _isIncrease, uint256 _chainWrappedTokenQuantity) returns()
func (_AdminContract *AdminContractTransactor) UpdateChainWrappedTokenQuantity(opts *bind.TransactOpts, _chainId uint8, _isIncrease bool, _chainWrappedTokenQuantity *big.Int) (*types.Transaction, error) {
	return _AdminContract.contract.Transact(opts, "updateChainWrappedTokenQuantity", _chainId, _isIncrease, _chainWrappedTokenQuantity)
}

// UpdateChainWrappedTokenQuantity is a paid mutator transaction binding the contract method 0x170eb3ff.
//
// Solidity: function updateChainWrappedTokenQuantity(uint8 _chainId, bool _isIncrease, uint256 _chainWrappedTokenQuantity) returns()
func (_AdminContract *AdminContractSession) UpdateChainWrappedTokenQuantity(_chainId uint8, _isIncrease bool, _chainWrappedTokenQuantity *big.Int) (*types.Transaction, error) {
	return _AdminContract.Contract.UpdateChainWrappedTokenQuantity(&_AdminContract.TransactOpts, _chainId, _isIncrease, _chainWrappedTokenQuantity)
}

// UpdateChainWrappedTokenQuantity is a paid mutator transaction binding the contract method 0x170eb3ff.
//
// Solidity: function updateChainWrappedTokenQuantity(uint8 _chainId, bool _isIncrease, uint256 _chainWrappedTokenQuantity) returns()
func (_AdminContract *AdminContractTransactorSession) UpdateChainWrappedTokenQuantity(_chainId uint8, _isIncrease bool, _chainWrappedTokenQuantity *big.Int) (*types.Transaction, error) {
	return _AdminContract.Contract.UpdateChainWrappedTokenQuantity(&_AdminContract.TransactOpts, _chainId, _isIncrease, _chainWrappedTokenQuantity)
}

// UpdateMaxNumberOfTransactions is a paid mutator transaction binding the contract method 0x39588c20.
//
// Solidity: function updateMaxNumberOfTransactions(uint16 _maxNumberOfTransactions) returns()
func (_AdminContract *AdminContractTransactor) UpdateMaxNumberOfTransactions(opts *bind.TransactOpts, _maxNumberOfTransactions uint16) (*types.Transaction, error) {
	return _AdminContract.contract.Transact(opts, "updateMaxNumberOfTransactions", _maxNumberOfTransactions)
}

// UpdateMaxNumberOfTransactions is a paid mutator transaction binding the contract method 0x39588c20.
//
// Solidity: function updateMaxNumberOfTransactions(uint16 _maxNumberOfTransactions) returns()
func (_AdminContract *AdminContractSession) UpdateMaxNumberOfTransactions(_maxNumberOfTransactions uint16) (*types.Transaction, error) {
	return _AdminContract.Contract.UpdateMaxNumberOfTransactions(&_AdminContract.TransactOpts, _maxNumberOfTransactions)
}

// UpdateMaxNumberOfTransactions is a paid mutator transaction binding the contract method 0x39588c20.
//
// Solidity: function updateMaxNumberOfTransactions(uint16 _maxNumberOfTransactions) returns()
func (_AdminContract *AdminContractTransactorSession) UpdateMaxNumberOfTransactions(_maxNumberOfTransactions uint16) (*types.Transaction, error) {
	return _AdminContract.Contract.UpdateMaxNumberOfTransactions(&_AdminContract.TransactOpts, _maxNumberOfTransactions)
}

// UpdateTimeoutBlocksNumber is a paid mutator transaction binding the contract method 0xc378dd07.
//
// Solidity: function updateTimeoutBlocksNumber(uint8 _timeoutBlocksNumber) returns()
func (_AdminContract *AdminContractTransactor) UpdateTimeoutBlocksNumber(opts *bind.TransactOpts, _timeoutBlocksNumber uint8) (*types.Transaction, error) {
	return _AdminContract.contract.Transact(opts, "updateTimeoutBlocksNumber", _timeoutBlocksNumber)
}

// UpdateTimeoutBlocksNumber is a paid mutator transaction binding the contract method 0xc378dd07.
//
// Solidity: function updateTimeoutBlocksNumber(uint8 _timeoutBlocksNumber) returns()
func (_AdminContract *AdminContractSession) UpdateTimeoutBlocksNumber(_timeoutBlocksNumber uint8) (*types.Transaction, error) {
	return _AdminContract.Contract.UpdateTimeoutBlocksNumber(&_AdminContract.TransactOpts, _timeoutBlocksNumber)
}

// UpdateTimeoutBlocksNumber is a paid mutator transaction binding the contract method 0xc378dd07.
//
// Solidity: function updateTimeoutBlocksNumber(uint8 _timeoutBlocksNumber) returns()
func (_AdminContract *AdminContractTransactorSession) UpdateTimeoutBlocksNumber(_timeoutBlocksNumber uint8) (*types.Transaction, error) {
	return _AdminContract.Contract.UpdateTimeoutBlocksNumber(&_AdminContract.TransactOpts, _timeoutBlocksNumber)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_AdminContract *AdminContractTransactor) UpgradeTo(opts *bind.TransactOpts, newImplementation common.Address) (*types.Transaction, error) {
	return _AdminContract.contract.Transact(opts, "upgradeTo", newImplementation)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_AdminContract *AdminContractSession) UpgradeTo(newImplementation common.Address) (*types.Transaction, error) {
	return _AdminContract.Contract.UpgradeTo(&_AdminContract.TransactOpts, newImplementation)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_AdminContract *AdminContractTransactorSession) UpgradeTo(newImplementation common.Address) (*types.Transaction, error) {
	return _AdminContract.Contract.UpgradeTo(&_AdminContract.TransactOpts, newImplementation)
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

// AdminContractAdminChangedIterator is returned from FilterAdminChanged and is used to iterate over the raw logs and unpacked data for AdminChanged events raised by the AdminContract contract.
type AdminContractAdminChangedIterator struct {
	Event *AdminContractAdminChanged // Event containing the contract specifics and raw log

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
func (it *AdminContractAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdminContractAdminChanged)
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
		it.Event = new(AdminContractAdminChanged)
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
func (it *AdminContractAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdminContractAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdminContractAdminChanged represents a AdminChanged event raised by the AdminContract contract.
type AdminContractAdminChanged struct {
	PreviousAdmin common.Address
	NewAdmin      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterAdminChanged is a free log retrieval operation binding the contract event 0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f.
//
// Solidity: event AdminChanged(address previousAdmin, address newAdmin)
func (_AdminContract *AdminContractFilterer) FilterAdminChanged(opts *bind.FilterOpts) (*AdminContractAdminChangedIterator, error) {

	logs, sub, err := _AdminContract.contract.FilterLogs(opts, "AdminChanged")
	if err != nil {
		return nil, err
	}
	return &AdminContractAdminChangedIterator{contract: _AdminContract.contract, event: "AdminChanged", logs: logs, sub: sub}, nil
}

// WatchAdminChanged is a free log subscription operation binding the contract event 0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f.
//
// Solidity: event AdminChanged(address previousAdmin, address newAdmin)
func (_AdminContract *AdminContractFilterer) WatchAdminChanged(opts *bind.WatchOpts, sink chan<- *AdminContractAdminChanged) (event.Subscription, error) {

	logs, sub, err := _AdminContract.contract.WatchLogs(opts, "AdminChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdminContractAdminChanged)
				if err := _AdminContract.contract.UnpackLog(event, "AdminChanged", log); err != nil {
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
func (_AdminContract *AdminContractFilterer) ParseAdminChanged(log types.Log) (*AdminContractAdminChanged, error) {
	event := new(AdminContractAdminChanged)
	if err := _AdminContract.contract.UnpackLog(event, "AdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AdminContractBeaconUpgradedIterator is returned from FilterBeaconUpgraded and is used to iterate over the raw logs and unpacked data for BeaconUpgraded events raised by the AdminContract contract.
type AdminContractBeaconUpgradedIterator struct {
	Event *AdminContractBeaconUpgraded // Event containing the contract specifics and raw log

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
func (it *AdminContractBeaconUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdminContractBeaconUpgraded)
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
		it.Event = new(AdminContractBeaconUpgraded)
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
func (it *AdminContractBeaconUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdminContractBeaconUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdminContractBeaconUpgraded represents a BeaconUpgraded event raised by the AdminContract contract.
type AdminContractBeaconUpgraded struct {
	Beacon common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterBeaconUpgraded is a free log retrieval operation binding the contract event 0x1cf3b03a6cf19fa2baba4df148e9dcabedea7f8a5c07840e207e5c089be95d3e.
//
// Solidity: event BeaconUpgraded(address indexed beacon)
func (_AdminContract *AdminContractFilterer) FilterBeaconUpgraded(opts *bind.FilterOpts, beacon []common.Address) (*AdminContractBeaconUpgradedIterator, error) {

	var beaconRule []interface{}
	for _, beaconItem := range beacon {
		beaconRule = append(beaconRule, beaconItem)
	}

	logs, sub, err := _AdminContract.contract.FilterLogs(opts, "BeaconUpgraded", beaconRule)
	if err != nil {
		return nil, err
	}
	return &AdminContractBeaconUpgradedIterator{contract: _AdminContract.contract, event: "BeaconUpgraded", logs: logs, sub: sub}, nil
}

// WatchBeaconUpgraded is a free log subscription operation binding the contract event 0x1cf3b03a6cf19fa2baba4df148e9dcabedea7f8a5c07840e207e5c089be95d3e.
//
// Solidity: event BeaconUpgraded(address indexed beacon)
func (_AdminContract *AdminContractFilterer) WatchBeaconUpgraded(opts *bind.WatchOpts, sink chan<- *AdminContractBeaconUpgraded, beacon []common.Address) (event.Subscription, error) {

	var beaconRule []interface{}
	for _, beaconItem := range beacon {
		beaconRule = append(beaconRule, beaconItem)
	}

	logs, sub, err := _AdminContract.contract.WatchLogs(opts, "BeaconUpgraded", beaconRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdminContractBeaconUpgraded)
				if err := _AdminContract.contract.UnpackLog(event, "BeaconUpgraded", log); err != nil {
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
func (_AdminContract *AdminContractFilterer) ParseBeaconUpgraded(log types.Log) (*AdminContractBeaconUpgraded, error) {
	event := new(AdminContractBeaconUpgraded)
	if err := _AdminContract.contract.UnpackLog(event, "BeaconUpgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AdminContractChainDefundedIterator is returned from FilterChainDefunded and is used to iterate over the raw logs and unpacked data for ChainDefunded events raised by the AdminContract contract.
type AdminContractChainDefundedIterator struct {
	Event *AdminContractChainDefunded // Event containing the contract specifics and raw log

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
func (it *AdminContractChainDefundedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdminContractChainDefunded)
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
		it.Event = new(AdminContractChainDefunded)
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
func (it *AdminContractChainDefundedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdminContractChainDefundedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdminContractChainDefunded represents a ChainDefunded event raised by the AdminContract contract.
type AdminContractChainDefunded struct {
	ChainId uint8
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterChainDefunded is a free log retrieval operation binding the contract event 0xce8ca6f0aead771734dd729750170b73fbb3147d90cd4273c1bbcfe669bbf69d.
//
// Solidity: event ChainDefunded(uint8 _chainId, uint256 _amount)
func (_AdminContract *AdminContractFilterer) FilterChainDefunded(opts *bind.FilterOpts) (*AdminContractChainDefundedIterator, error) {

	logs, sub, err := _AdminContract.contract.FilterLogs(opts, "ChainDefunded")
	if err != nil {
		return nil, err
	}
	return &AdminContractChainDefundedIterator{contract: _AdminContract.contract, event: "ChainDefunded", logs: logs, sub: sub}, nil
}

// WatchChainDefunded is a free log subscription operation binding the contract event 0xce8ca6f0aead771734dd729750170b73fbb3147d90cd4273c1bbcfe669bbf69d.
//
// Solidity: event ChainDefunded(uint8 _chainId, uint256 _amount)
func (_AdminContract *AdminContractFilterer) WatchChainDefunded(opts *bind.WatchOpts, sink chan<- *AdminContractChainDefunded) (event.Subscription, error) {

	logs, sub, err := _AdminContract.contract.WatchLogs(opts, "ChainDefunded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdminContractChainDefunded)
				if err := _AdminContract.contract.UnpackLog(event, "ChainDefunded", log); err != nil {
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
func (_AdminContract *AdminContractFilterer) ParseChainDefunded(log types.Log) (*AdminContractChainDefunded, error) {
	event := new(AdminContractChainDefunded)
	if err := _AdminContract.contract.UnpackLog(event, "ChainDefunded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AdminContractDefundFailedAfterMultipleRetriesIterator is returned from FilterDefundFailedAfterMultipleRetries and is used to iterate over the raw logs and unpacked data for DefundFailedAfterMultipleRetries events raised by the AdminContract contract.
type AdminContractDefundFailedAfterMultipleRetriesIterator struct {
	Event *AdminContractDefundFailedAfterMultipleRetries // Event containing the contract specifics and raw log

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
func (it *AdminContractDefundFailedAfterMultipleRetriesIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdminContractDefundFailedAfterMultipleRetries)
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
		it.Event = new(AdminContractDefundFailedAfterMultipleRetries)
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
func (it *AdminContractDefundFailedAfterMultipleRetriesIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdminContractDefundFailedAfterMultipleRetriesIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdminContractDefundFailedAfterMultipleRetries represents a DefundFailedAfterMultipleRetries event raised by the AdminContract contract.
type AdminContractDefundFailedAfterMultipleRetries struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterDefundFailedAfterMultipleRetries is a free log retrieval operation binding the contract event 0xc53812079c8257d7e3e68b2e9d937f825484546404c4bbdfa79785d213976c6f.
//
// Solidity: event DefundFailedAfterMultipleRetries()
func (_AdminContract *AdminContractFilterer) FilterDefundFailedAfterMultipleRetries(opts *bind.FilterOpts) (*AdminContractDefundFailedAfterMultipleRetriesIterator, error) {

	logs, sub, err := _AdminContract.contract.FilterLogs(opts, "DefundFailedAfterMultipleRetries")
	if err != nil {
		return nil, err
	}
	return &AdminContractDefundFailedAfterMultipleRetriesIterator{contract: _AdminContract.contract, event: "DefundFailedAfterMultipleRetries", logs: logs, sub: sub}, nil
}

// WatchDefundFailedAfterMultipleRetries is a free log subscription operation binding the contract event 0xc53812079c8257d7e3e68b2e9d937f825484546404c4bbdfa79785d213976c6f.
//
// Solidity: event DefundFailedAfterMultipleRetries()
func (_AdminContract *AdminContractFilterer) WatchDefundFailedAfterMultipleRetries(opts *bind.WatchOpts, sink chan<- *AdminContractDefundFailedAfterMultipleRetries) (event.Subscription, error) {

	logs, sub, err := _AdminContract.contract.WatchLogs(opts, "DefundFailedAfterMultipleRetries")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdminContractDefundFailedAfterMultipleRetries)
				if err := _AdminContract.contract.UnpackLog(event, "DefundFailedAfterMultipleRetries", log); err != nil {
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
func (_AdminContract *AdminContractFilterer) ParseDefundFailedAfterMultipleRetries(log types.Log) (*AdminContractDefundFailedAfterMultipleRetries, error) {
	event := new(AdminContractDefundFailedAfterMultipleRetries)
	if err := _AdminContract.contract.UnpackLog(event, "DefundFailedAfterMultipleRetries", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AdminContractFundAdminChangedIterator is returned from FilterFundAdminChanged and is used to iterate over the raw logs and unpacked data for FundAdminChanged events raised by the AdminContract contract.
type AdminContractFundAdminChangedIterator struct {
	Event *AdminContractFundAdminChanged // Event containing the contract specifics and raw log

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
func (it *AdminContractFundAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdminContractFundAdminChanged)
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
		it.Event = new(AdminContractFundAdminChanged)
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
func (it *AdminContractFundAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdminContractFundAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdminContractFundAdminChanged represents a FundAdminChanged event raised by the AdminContract contract.
type AdminContractFundAdminChanged struct {
	NewFundAdmin common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterFundAdminChanged is a free log retrieval operation binding the contract event 0x217da1af647045bf1a50dc62120037c19f6963e96ae6bea7964bf241770fbe00.
//
// Solidity: event FundAdminChanged(address _newFundAdmin)
func (_AdminContract *AdminContractFilterer) FilterFundAdminChanged(opts *bind.FilterOpts) (*AdminContractFundAdminChangedIterator, error) {

	logs, sub, err := _AdminContract.contract.FilterLogs(opts, "FundAdminChanged")
	if err != nil {
		return nil, err
	}
	return &AdminContractFundAdminChangedIterator{contract: _AdminContract.contract, event: "FundAdminChanged", logs: logs, sub: sub}, nil
}

// WatchFundAdminChanged is a free log subscription operation binding the contract event 0x217da1af647045bf1a50dc62120037c19f6963e96ae6bea7964bf241770fbe00.
//
// Solidity: event FundAdminChanged(address _newFundAdmin)
func (_AdminContract *AdminContractFilterer) WatchFundAdminChanged(opts *bind.WatchOpts, sink chan<- *AdminContractFundAdminChanged) (event.Subscription, error) {

	logs, sub, err := _AdminContract.contract.WatchLogs(opts, "FundAdminChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdminContractFundAdminChanged)
				if err := _AdminContract.contract.UnpackLog(event, "FundAdminChanged", log); err != nil {
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
func (_AdminContract *AdminContractFilterer) ParseFundAdminChanged(log types.Log) (*AdminContractFundAdminChanged, error) {
	event := new(AdminContractFundAdminChanged)
	if err := _AdminContract.contract.UnpackLog(event, "FundAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
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
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_AdminContract *AdminContractFilterer) FilterInitialized(opts *bind.FilterOpts) (*AdminContractInitializedIterator, error) {

	logs, sub, err := _AdminContract.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &AdminContractInitializedIterator{contract: _AdminContract.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
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

// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_AdminContract *AdminContractFilterer) ParseInitialized(log types.Log) (*AdminContractInitialized, error) {
	event := new(AdminContractInitialized)
	if err := _AdminContract.contract.UnpackLog(event, "Initialized", log); err != nil {
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
	ChainId            *big.Int
	IsIncrement        bool
	ChainTokenQuantity *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterUpdatedChainTokenQuantity is a free log retrieval operation binding the contract event 0x63a8310d54d22ae170c9c99ec0101494848847baf8ba54b1f297456f4c01bd62.
//
// Solidity: event UpdatedChainTokenQuantity(uint256 indexed chainId, bool isIncrement, uint256 chainTokenQuantity)
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
// Solidity: event UpdatedChainTokenQuantity(uint256 indexed chainId, bool isIncrement, uint256 chainTokenQuantity)
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
// Solidity: event UpdatedChainTokenQuantity(uint256 indexed chainId, bool isIncrement, uint256 chainTokenQuantity)
func (_AdminContract *AdminContractFilterer) ParseUpdatedChainTokenQuantity(log types.Log) (*AdminContractUpdatedChainTokenQuantity, error) {
	event := new(AdminContractUpdatedChainTokenQuantity)
	if err := _AdminContract.contract.UnpackLog(event, "UpdatedChainTokenQuantity", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AdminContractUpdatedChainWrappedTokenQuantityIterator is returned from FilterUpdatedChainWrappedTokenQuantity and is used to iterate over the raw logs and unpacked data for UpdatedChainWrappedTokenQuantity events raised by the AdminContract contract.
type AdminContractUpdatedChainWrappedTokenQuantityIterator struct {
	Event *AdminContractUpdatedChainWrappedTokenQuantity // Event containing the contract specifics and raw log

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
func (it *AdminContractUpdatedChainWrappedTokenQuantityIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdminContractUpdatedChainWrappedTokenQuantity)
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
		it.Event = new(AdminContractUpdatedChainWrappedTokenQuantity)
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
func (it *AdminContractUpdatedChainWrappedTokenQuantityIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdminContractUpdatedChainWrappedTokenQuantityIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdminContractUpdatedChainWrappedTokenQuantity represents a UpdatedChainWrappedTokenQuantity event raised by the AdminContract contract.
type AdminContractUpdatedChainWrappedTokenQuantity struct {
	ChainId                   *big.Int
	IsIncrement               bool
	ChainWrappedTokenQuantity *big.Int
	Raw                       types.Log // Blockchain specific contextual infos
}

// FilterUpdatedChainWrappedTokenQuantity is a free log retrieval operation binding the contract event 0xd9cfc60422d337bdb03fa2a49851c73e6104f2c838b8ed8623dc48b282691b00.
//
// Solidity: event UpdatedChainWrappedTokenQuantity(uint256 indexed chainId, bool isIncrement, uint256 chainWrappedTokenQuantity)
func (_AdminContract *AdminContractFilterer) FilterUpdatedChainWrappedTokenQuantity(opts *bind.FilterOpts, chainId []*big.Int) (*AdminContractUpdatedChainWrappedTokenQuantityIterator, error) {

	var chainIdRule []interface{}
	for _, chainIdItem := range chainId {
		chainIdRule = append(chainIdRule, chainIdItem)
	}

	logs, sub, err := _AdminContract.contract.FilterLogs(opts, "UpdatedChainWrappedTokenQuantity", chainIdRule)
	if err != nil {
		return nil, err
	}
	return &AdminContractUpdatedChainWrappedTokenQuantityIterator{contract: _AdminContract.contract, event: "UpdatedChainWrappedTokenQuantity", logs: logs, sub: sub}, nil
}

// WatchUpdatedChainWrappedTokenQuantity is a free log subscription operation binding the contract event 0xd9cfc60422d337bdb03fa2a49851c73e6104f2c838b8ed8623dc48b282691b00.
//
// Solidity: event UpdatedChainWrappedTokenQuantity(uint256 indexed chainId, bool isIncrement, uint256 chainWrappedTokenQuantity)
func (_AdminContract *AdminContractFilterer) WatchUpdatedChainWrappedTokenQuantity(opts *bind.WatchOpts, sink chan<- *AdminContractUpdatedChainWrappedTokenQuantity, chainId []*big.Int) (event.Subscription, error) {

	var chainIdRule []interface{}
	for _, chainIdItem := range chainId {
		chainIdRule = append(chainIdRule, chainIdItem)
	}

	logs, sub, err := _AdminContract.contract.WatchLogs(opts, "UpdatedChainWrappedTokenQuantity", chainIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdminContractUpdatedChainWrappedTokenQuantity)
				if err := _AdminContract.contract.UnpackLog(event, "UpdatedChainWrappedTokenQuantity", log); err != nil {
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

// ParseUpdatedChainWrappedTokenQuantity is a log parse operation binding the contract event 0xd9cfc60422d337bdb03fa2a49851c73e6104f2c838b8ed8623dc48b282691b00.
//
// Solidity: event UpdatedChainWrappedTokenQuantity(uint256 indexed chainId, bool isIncrement, uint256 chainWrappedTokenQuantity)
func (_AdminContract *AdminContractFilterer) ParseUpdatedChainWrappedTokenQuantity(log types.Log) (*AdminContractUpdatedChainWrappedTokenQuantity, error) {
	event := new(AdminContractUpdatedChainWrappedTokenQuantity)
	if err := _AdminContract.contract.UnpackLog(event, "UpdatedChainWrappedTokenQuantity", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AdminContractUpdatedMaxNumberOfTransactionsIterator is returned from FilterUpdatedMaxNumberOfTransactions and is used to iterate over the raw logs and unpacked data for UpdatedMaxNumberOfTransactions events raised by the AdminContract contract.
type AdminContractUpdatedMaxNumberOfTransactionsIterator struct {
	Event *AdminContractUpdatedMaxNumberOfTransactions // Event containing the contract specifics and raw log

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
func (it *AdminContractUpdatedMaxNumberOfTransactionsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdminContractUpdatedMaxNumberOfTransactions)
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
		it.Event = new(AdminContractUpdatedMaxNumberOfTransactions)
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
func (it *AdminContractUpdatedMaxNumberOfTransactionsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdminContractUpdatedMaxNumberOfTransactionsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdminContractUpdatedMaxNumberOfTransactions represents a UpdatedMaxNumberOfTransactions event raised by the AdminContract contract.
type AdminContractUpdatedMaxNumberOfTransactions struct {
	MaxNumberOfTransactions *big.Int
	Raw                     types.Log // Blockchain specific contextual infos
}

// FilterUpdatedMaxNumberOfTransactions is a free log retrieval operation binding the contract event 0x0df74faf2972aea8bdd626cd6886a1fa4c9813a87981870c58f1e6c1ebd9f89a.
//
// Solidity: event UpdatedMaxNumberOfTransactions(uint256 _maxNumberOfTransactions)
func (_AdminContract *AdminContractFilterer) FilterUpdatedMaxNumberOfTransactions(opts *bind.FilterOpts) (*AdminContractUpdatedMaxNumberOfTransactionsIterator, error) {

	logs, sub, err := _AdminContract.contract.FilterLogs(opts, "UpdatedMaxNumberOfTransactions")
	if err != nil {
		return nil, err
	}
	return &AdminContractUpdatedMaxNumberOfTransactionsIterator{contract: _AdminContract.contract, event: "UpdatedMaxNumberOfTransactions", logs: logs, sub: sub}, nil
}

// WatchUpdatedMaxNumberOfTransactions is a free log subscription operation binding the contract event 0x0df74faf2972aea8bdd626cd6886a1fa4c9813a87981870c58f1e6c1ebd9f89a.
//
// Solidity: event UpdatedMaxNumberOfTransactions(uint256 _maxNumberOfTransactions)
func (_AdminContract *AdminContractFilterer) WatchUpdatedMaxNumberOfTransactions(opts *bind.WatchOpts, sink chan<- *AdminContractUpdatedMaxNumberOfTransactions) (event.Subscription, error) {

	logs, sub, err := _AdminContract.contract.WatchLogs(opts, "UpdatedMaxNumberOfTransactions")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdminContractUpdatedMaxNumberOfTransactions)
				if err := _AdminContract.contract.UnpackLog(event, "UpdatedMaxNumberOfTransactions", log); err != nil {
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
func (_AdminContract *AdminContractFilterer) ParseUpdatedMaxNumberOfTransactions(log types.Log) (*AdminContractUpdatedMaxNumberOfTransactions, error) {
	event := new(AdminContractUpdatedMaxNumberOfTransactions)
	if err := _AdminContract.contract.UnpackLog(event, "UpdatedMaxNumberOfTransactions", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AdminContractUpdatedTimeoutBlocksNumberIterator is returned from FilterUpdatedTimeoutBlocksNumber and is used to iterate over the raw logs and unpacked data for UpdatedTimeoutBlocksNumber events raised by the AdminContract contract.
type AdminContractUpdatedTimeoutBlocksNumberIterator struct {
	Event *AdminContractUpdatedTimeoutBlocksNumber // Event containing the contract specifics and raw log

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
func (it *AdminContractUpdatedTimeoutBlocksNumberIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdminContractUpdatedTimeoutBlocksNumber)
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
		it.Event = new(AdminContractUpdatedTimeoutBlocksNumber)
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
func (it *AdminContractUpdatedTimeoutBlocksNumberIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdminContractUpdatedTimeoutBlocksNumberIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdminContractUpdatedTimeoutBlocksNumber represents a UpdatedTimeoutBlocksNumber event raised by the AdminContract contract.
type AdminContractUpdatedTimeoutBlocksNumber struct {
	TimeoutBlocksNumber *big.Int
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterUpdatedTimeoutBlocksNumber is a free log retrieval operation binding the contract event 0x417143ffedb5f1c20f085be7f77b19ca8b7c93d98a93d5a133e3516ecc23b409.
//
// Solidity: event UpdatedTimeoutBlocksNumber(uint256 _timeoutBlocksNumber)
func (_AdminContract *AdminContractFilterer) FilterUpdatedTimeoutBlocksNumber(opts *bind.FilterOpts) (*AdminContractUpdatedTimeoutBlocksNumberIterator, error) {

	logs, sub, err := _AdminContract.contract.FilterLogs(opts, "UpdatedTimeoutBlocksNumber")
	if err != nil {
		return nil, err
	}
	return &AdminContractUpdatedTimeoutBlocksNumberIterator{contract: _AdminContract.contract, event: "UpdatedTimeoutBlocksNumber", logs: logs, sub: sub}, nil
}

// WatchUpdatedTimeoutBlocksNumber is a free log subscription operation binding the contract event 0x417143ffedb5f1c20f085be7f77b19ca8b7c93d98a93d5a133e3516ecc23b409.
//
// Solidity: event UpdatedTimeoutBlocksNumber(uint256 _timeoutBlocksNumber)
func (_AdminContract *AdminContractFilterer) WatchUpdatedTimeoutBlocksNumber(opts *bind.WatchOpts, sink chan<- *AdminContractUpdatedTimeoutBlocksNumber) (event.Subscription, error) {

	logs, sub, err := _AdminContract.contract.WatchLogs(opts, "UpdatedTimeoutBlocksNumber")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdminContractUpdatedTimeoutBlocksNumber)
				if err := _AdminContract.contract.UnpackLog(event, "UpdatedTimeoutBlocksNumber", log); err != nil {
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
func (_AdminContract *AdminContractFilterer) ParseUpdatedTimeoutBlocksNumber(log types.Log) (*AdminContractUpdatedTimeoutBlocksNumber, error) {
	event := new(AdminContractUpdatedTimeoutBlocksNumber)
	if err := _AdminContract.contract.UnpackLog(event, "UpdatedTimeoutBlocksNumber", log); err != nil {
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
