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

// TestContractMetaData contains all meta data concerning the TestContract contract.
var TestContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"getValue\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"number\",\"type\":\"uint256\"}],\"name\":\"setValue\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// TestContractABI is the input ABI used to generate the binding from.
// Deprecated: Use TestContractMetaData.ABI instead.
var TestContractABI = TestContractMetaData.ABI

// TestContract is an auto generated Go binding around an Ethereum contract.
type TestContract struct {
	TestContractCaller     // Read-only binding to the contract
	TestContractTransactor // Write-only binding to the contract
	TestContractFilterer   // Log filterer for contract events
}

// TestContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type TestContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TestContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TestContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TestContractSession struct {
	Contract     *TestContract     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// TestContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TestContractCallerSession struct {
	Contract *TestContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// TestContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TestContractTransactorSession struct {
	Contract     *TestContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// TestContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type TestContractRaw struct {
	Contract *TestContract // Generic contract binding to access the raw methods on
}

// TestContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TestContractCallerRaw struct {
	Contract *TestContractCaller // Generic read-only contract binding to access the raw methods on
}

// TestContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TestContractTransactorRaw struct {
	Contract *TestContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTestContract creates a new instance of TestContract, bound to a specific deployed contract.
func NewTestContract(address common.Address, backend bind.ContractBackend) (*TestContract, error) {
	contract, err := bindTestContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &TestContract{TestContractCaller: TestContractCaller{contract: contract}, TestContractTransactor: TestContractTransactor{contract: contract}, TestContractFilterer: TestContractFilterer{contract: contract}}, nil
}

// NewTestContractCaller creates a new read-only instance of TestContract, bound to a specific deployed contract.
func NewTestContractCaller(address common.Address, caller bind.ContractCaller) (*TestContractCaller, error) {
	contract, err := bindTestContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TestContractCaller{contract: contract}, nil
}

// NewTestContractTransactor creates a new write-only instance of TestContract, bound to a specific deployed contract.
func NewTestContractTransactor(address common.Address, transactor bind.ContractTransactor) (*TestContractTransactor, error) {
	contract, err := bindTestContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TestContractTransactor{contract: contract}, nil
}

// NewTestContractFilterer creates a new log filterer instance of TestContract, bound to a specific deployed contract.
func NewTestContractFilterer(address common.Address, filterer bind.ContractFilterer) (*TestContractFilterer, error) {
	contract, err := bindTestContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TestContractFilterer{contract: contract}, nil
}

// bindTestContract binds a generic wrapper to an already deployed contract.
func bindTestContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := TestContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TestContract *TestContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TestContract.Contract.TestContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TestContract *TestContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestContract.Contract.TestContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TestContract *TestContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TestContract.Contract.TestContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TestContract *TestContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TestContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TestContract *TestContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TestContract *TestContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TestContract.Contract.contract.Transact(opts, method, params...)
}

// GetValue is a free data retrieval call binding the contract method 0x20965255.
//
// Solidity: function getValue() view returns(uint256)
func (_TestContract *TestContractCaller) GetValue(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _TestContract.contract.Call(opts, &out, "getValue")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetValue is a free data retrieval call binding the contract method 0x20965255.
//
// Solidity: function getValue() view returns(uint256)
func (_TestContract *TestContractSession) GetValue() (*big.Int, error) {
	return _TestContract.Contract.GetValue(&_TestContract.CallOpts)
}

// GetValue is a free data retrieval call binding the contract method 0x20965255.
//
// Solidity: function getValue() view returns(uint256)
func (_TestContract *TestContractCallerSession) GetValue() (*big.Int, error) {
	return _TestContract.Contract.GetValue(&_TestContract.CallOpts)
}

// SetValue is a paid mutator transaction binding the contract method 0x55241077.
//
// Solidity: function setValue(uint256 number) returns()
func (_TestContract *TestContractTransactor) SetValue(opts *bind.TransactOpts, number *big.Int) (*types.Transaction, error) {
	return _TestContract.contract.Transact(opts, "setValue", number)
}

// SetValue is a paid mutator transaction binding the contract method 0x55241077.
//
// Solidity: function setValue(uint256 number) returns()
func (_TestContract *TestContractSession) SetValue(number *big.Int) (*types.Transaction, error) {
	return _TestContract.Contract.SetValue(&_TestContract.TransactOpts, number)
}

// SetValue is a paid mutator transaction binding the contract method 0x55241077.
//
// Solidity: function setValue(uint256 number) returns()
func (_TestContract *TestContractTransactorSession) SetValue(number *big.Int) (*types.Transaction, error) {
	return _TestContract.Contract.SetValue(&_TestContract.TransactOpts, number)
}
