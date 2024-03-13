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

// TestContractKeyHashValidatorCardanoData is an auto generated low-level Go binding around an user-defined struct.
type TestContractKeyHashValidatorCardanoData struct {
	KeyHash         string
	KeyHashFee      string
	VerifyingKey    []byte
	VerifyingKeyFee []byte
}

// TestContractKeyHashMetaData contains all meta data concerning the TestContractKeyHash contract.
var TestContractKeyHashMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"chainID\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"KeyHash\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"KeyHashFee\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"VerifyingKey\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"VerifyingKeyFee\",\"type\":\"bytes\"}],\"internalType\":\"structTestContractKeyHash.ValidatorCardanoData\",\"name\":\"vd\",\"type\":\"tuple\"}],\"name\":\"setValidatorCardanoData\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"validatorCardanoDataMap\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"KeyHash\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"KeyHashFee\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"VerifyingKey\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"VerifyingKeyFee\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// TestContractKeyHashABI is the input ABI used to generate the binding from.
// Deprecated: Use TestContractKeyHashMetaData.ABI instead.
var TestContractKeyHashABI = TestContractKeyHashMetaData.ABI

// TestContractKeyHash is an auto generated Go binding around an Ethereum contract.
type TestContractKeyHash struct {
	TestContractKeyHashCaller     // Read-only binding to the contract
	TestContractKeyHashTransactor // Write-only binding to the contract
	TestContractKeyHashFilterer   // Log filterer for contract events
}

// TestContractKeyHashCaller is an auto generated read-only Go binding around an Ethereum contract.
type TestContractKeyHashCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestContractKeyHashTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TestContractKeyHashTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestContractKeyHashFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TestContractKeyHashFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestContractKeyHashSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TestContractKeyHashSession struct {
	Contract     *TestContractKeyHash // Generic contract binding to set the session for
	CallOpts     bind.CallOpts        // Call options to use throughout this session
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// TestContractKeyHashCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TestContractKeyHashCallerSession struct {
	Contract *TestContractKeyHashCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts              // Call options to use throughout this session
}

// TestContractKeyHashTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TestContractKeyHashTransactorSession struct {
	Contract     *TestContractKeyHashTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts              // Transaction auth options to use throughout this session
}

// TestContractKeyHashRaw is an auto generated low-level Go binding around an Ethereum contract.
type TestContractKeyHashRaw struct {
	Contract *TestContractKeyHash // Generic contract binding to access the raw methods on
}

// TestContractKeyHashCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TestContractKeyHashCallerRaw struct {
	Contract *TestContractKeyHashCaller // Generic read-only contract binding to access the raw methods on
}

// TestContractKeyHashTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TestContractKeyHashTransactorRaw struct {
	Contract *TestContractKeyHashTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTestContractKeyHash creates a new instance of TestContractKeyHash, bound to a specific deployed contract.
func NewTestContractKeyHash(address common.Address, backend bind.ContractBackend) (*TestContractKeyHash, error) {
	contract, err := bindTestContractKeyHash(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &TestContractKeyHash{TestContractKeyHashCaller: TestContractKeyHashCaller{contract: contract}, TestContractKeyHashTransactor: TestContractKeyHashTransactor{contract: contract}, TestContractKeyHashFilterer: TestContractKeyHashFilterer{contract: contract}}, nil
}

// NewTestContractKeyHashCaller creates a new read-only instance of TestContractKeyHash, bound to a specific deployed contract.
func NewTestContractKeyHashCaller(address common.Address, caller bind.ContractCaller) (*TestContractKeyHashCaller, error) {
	contract, err := bindTestContractKeyHash(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TestContractKeyHashCaller{contract: contract}, nil
}

// NewTestContractKeyHashTransactor creates a new write-only instance of TestContractKeyHash, bound to a specific deployed contract.
func NewTestContractKeyHashTransactor(address common.Address, transactor bind.ContractTransactor) (*TestContractKeyHashTransactor, error) {
	contract, err := bindTestContractKeyHash(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TestContractKeyHashTransactor{contract: contract}, nil
}

// NewTestContractKeyHashFilterer creates a new log filterer instance of TestContractKeyHash, bound to a specific deployed contract.
func NewTestContractKeyHashFilterer(address common.Address, filterer bind.ContractFilterer) (*TestContractKeyHashFilterer, error) {
	contract, err := bindTestContractKeyHash(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TestContractKeyHashFilterer{contract: contract}, nil
}

// bindTestContractKeyHash binds a generic wrapper to an already deployed contract.
func bindTestContractKeyHash(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := TestContractKeyHashMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TestContractKeyHash *TestContractKeyHashRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TestContractKeyHash.Contract.TestContractKeyHashCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TestContractKeyHash *TestContractKeyHashRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestContractKeyHash.Contract.TestContractKeyHashTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TestContractKeyHash *TestContractKeyHashRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TestContractKeyHash.Contract.TestContractKeyHashTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TestContractKeyHash *TestContractKeyHashCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TestContractKeyHash.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TestContractKeyHash *TestContractKeyHashTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestContractKeyHash.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TestContractKeyHash *TestContractKeyHashTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TestContractKeyHash.Contract.contract.Transact(opts, method, params...)
}

// ValidatorCardanoDataMap is a free data retrieval call binding the contract method 0x5834f23a.
//
// Solidity: function validatorCardanoDataMap(address , string ) view returns(string KeyHash, string KeyHashFee, bytes VerifyingKey, bytes VerifyingKeyFee)
func (_TestContractKeyHash *TestContractKeyHashCaller) ValidatorCardanoDataMap(opts *bind.CallOpts, arg0 common.Address, arg1 string) (struct {
	KeyHash         string
	KeyHashFee      string
	VerifyingKey    []byte
	VerifyingKeyFee []byte
}, error) {
	var out []interface{}
	err := _TestContractKeyHash.contract.Call(opts, &out, "validatorCardanoDataMap", arg0, arg1)

	outstruct := new(struct {
		KeyHash         string
		KeyHashFee      string
		VerifyingKey    []byte
		VerifyingKeyFee []byte
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.KeyHash = *abi.ConvertType(out[0], new(string)).(*string)
	outstruct.KeyHashFee = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.VerifyingKey = *abi.ConvertType(out[2], new([]byte)).(*[]byte)
	outstruct.VerifyingKeyFee = *abi.ConvertType(out[3], new([]byte)).(*[]byte)

	return *outstruct, err

}

// ValidatorCardanoDataMap is a free data retrieval call binding the contract method 0x5834f23a.
//
// Solidity: function validatorCardanoDataMap(address , string ) view returns(string KeyHash, string KeyHashFee, bytes VerifyingKey, bytes VerifyingKeyFee)
func (_TestContractKeyHash *TestContractKeyHashSession) ValidatorCardanoDataMap(arg0 common.Address, arg1 string) (struct {
	KeyHash         string
	KeyHashFee      string
	VerifyingKey    []byte
	VerifyingKeyFee []byte
}, error) {
	return _TestContractKeyHash.Contract.ValidatorCardanoDataMap(&_TestContractKeyHash.CallOpts, arg0, arg1)
}

// ValidatorCardanoDataMap is a free data retrieval call binding the contract method 0x5834f23a.
//
// Solidity: function validatorCardanoDataMap(address , string ) view returns(string KeyHash, string KeyHashFee, bytes VerifyingKey, bytes VerifyingKeyFee)
func (_TestContractKeyHash *TestContractKeyHashCallerSession) ValidatorCardanoDataMap(arg0 common.Address, arg1 string) (struct {
	KeyHash         string
	KeyHashFee      string
	VerifyingKey    []byte
	VerifyingKeyFee []byte
}, error) {
	return _TestContractKeyHash.Contract.ValidatorCardanoDataMap(&_TestContractKeyHash.CallOpts, arg0, arg1)
}

// SetValidatorCardanoData is a paid mutator transaction binding the contract method 0x6cd3e80c.
//
// Solidity: function setValidatorCardanoData(string chainID, (string,string,bytes,bytes) vd) returns()
func (_TestContractKeyHash *TestContractKeyHashTransactor) SetValidatorCardanoData(opts *bind.TransactOpts, chainID string, vd TestContractKeyHashValidatorCardanoData) (*types.Transaction, error) {
	return _TestContractKeyHash.contract.Transact(opts, "setValidatorCardanoData", chainID, vd)
}

// SetValidatorCardanoData is a paid mutator transaction binding the contract method 0x6cd3e80c.
//
// Solidity: function setValidatorCardanoData(string chainID, (string,string,bytes,bytes) vd) returns()
func (_TestContractKeyHash *TestContractKeyHashSession) SetValidatorCardanoData(chainID string, vd TestContractKeyHashValidatorCardanoData) (*types.Transaction, error) {
	return _TestContractKeyHash.Contract.SetValidatorCardanoData(&_TestContractKeyHash.TransactOpts, chainID, vd)
}

// SetValidatorCardanoData is a paid mutator transaction binding the contract method 0x6cd3e80c.
//
// Solidity: function setValidatorCardanoData(string chainID, (string,string,bytes,bytes) vd) returns()
func (_TestContractKeyHash *TestContractKeyHashTransactorSession) SetValidatorCardanoData(chainID string, vd TestContractKeyHashValidatorCardanoData) (*types.Transaction, error) {
	return _TestContractKeyHash.Contract.SetValidatorCardanoData(&_TestContractKeyHash.TransactOpts, chainID, vd)
}
