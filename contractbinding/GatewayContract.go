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
}

// IGatewayStructsValidatorAddressChainData is an auto generated low-level Go binding around an user-defined struct.
type IGatewayStructsValidatorAddressChainData struct {
	Addr common.Address
	Data IGatewayStructsValidatorChainData
}

// IGatewayStructsValidatorChainData is an auto generated low-level Go binding around an user-defined struct.
type IGatewayStructsValidatorChainData struct {
	Key [4]*big.Int
}

// GatewayMetaData contains all meta data concerning the Gateway contract.
var GatewayMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"BatchAlreadyExecuted\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"BurnFailed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DecresedAllowenceBelowZero\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ExceedsMaxLength\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InsufficientAllowance\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"data\",\"type\":\"string\"}],\"name\":\"InvalidData\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidReceiver\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSignature\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotGateway\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotGatewayOrOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotPredicate\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotPredicateOrOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotRelayer\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"PrecompileCallFailed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ValidatorNotRegistered\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroAddress\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"previousAdmin\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"AdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"beacon\",\"type\":\"address\"}],\"name\":\"BeaconUpgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"TTLExpired\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"Upgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"destinationChainId\",\"type\":\"uint8\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"indexed\":false,\"internalType\":\"structIGatewayStructs.ReceiverWithdraw[]\",\"name\":\"receivers\",\"type\":\"tuple[]\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"feeAmount\",\"type\":\"uint256\"}],\"name\":\"Withdraw\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"MAX_LENGTH\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_addr\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"uint256[4]\",\"name\":\"key\",\"type\":\"uint256[4]\"}],\"internalType\":\"structIGatewayStructs.ValidatorChainData\",\"name\":\"_data\",\"type\":\"tuple\"}],\"name\":\"addValidatorChainData\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_signature\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"_bitmap\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"depositEvent\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"eRC20TokenPredicate\",\"outputs\":[{\"internalType\":\"contractERC20TokenPredicate\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"proxiableUUID\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"relayer\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_eRC20TokenPredicate\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_validators\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_relayer\",\"type\":\"address\"}],\"name\":\"setDependencies\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"uint256[4]\",\"name\":\"key\",\"type\":\"uint256[4]\"}],\"internalType\":\"structIGatewayStructs.ValidatorChainData\",\"name\":\"data\",\"type\":\"tuple\"}],\"internalType\":\"structIGatewayStructs.ValidatorAddressChainData[]\",\"name\":\"_chainDatas\",\"type\":\"tuple[]\"}],\"name\":\"setValidatorsChainData\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"ttlEvent\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"}],\"name\":\"upgradeTo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"upgradeToAndCall\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"validators\",\"outputs\":[{\"internalType\":\"contractValidators\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_destinationChainId\",\"type\":\"uint8\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIGatewayStructs.ReceiverWithdraw[]\",\"name\":\"_receivers\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256\",\"name\":\"_feeAmount\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_destinationChainId\",\"type\":\"uint8\"},{\"internalType\":\"address\",\"name\":\"_sender\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIGatewayStructs.ReceiverWithdraw[]\",\"name\":\"_receivers\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256\",\"name\":\"_feeAmount\",\"type\":\"uint256\"}],\"name\":\"withdrawEvent\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
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

// MAXLENGTH is a free data retrieval call binding the contract method 0xa6f9885c.
//
// Solidity: function MAX_LENGTH() view returns(uint256)
func (_Gateway *GatewayCaller) MAXLENGTH(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Gateway.contract.Call(opts, &out, "MAX_LENGTH")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXLENGTH is a free data retrieval call binding the contract method 0xa6f9885c.
//
// Solidity: function MAX_LENGTH() view returns(uint256)
func (_Gateway *GatewaySession) MAXLENGTH() (*big.Int, error) {
	return _Gateway.Contract.MAXLENGTH(&_Gateway.CallOpts)
}

// MAXLENGTH is a free data retrieval call binding the contract method 0xa6f9885c.
//
// Solidity: function MAX_LENGTH() view returns(uint256)
func (_Gateway *GatewayCallerSession) MAXLENGTH() (*big.Int, error) {
	return _Gateway.Contract.MAXLENGTH(&_Gateway.CallOpts)
}

// ERC20TokenPredicate is a free data retrieval call binding the contract method 0xb999908b.
//
// Solidity: function eRC20TokenPredicate() view returns(address)
func (_Gateway *GatewayCaller) ERC20TokenPredicate(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Gateway.contract.Call(opts, &out, "eRC20TokenPredicate")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ERC20TokenPredicate is a free data retrieval call binding the contract method 0xb999908b.
//
// Solidity: function eRC20TokenPredicate() view returns(address)
func (_Gateway *GatewaySession) ERC20TokenPredicate() (common.Address, error) {
	return _Gateway.Contract.ERC20TokenPredicate(&_Gateway.CallOpts)
}

// ERC20TokenPredicate is a free data retrieval call binding the contract method 0xb999908b.
//
// Solidity: function eRC20TokenPredicate() view returns(address)
func (_Gateway *GatewayCallerSession) ERC20TokenPredicate() (common.Address, error) {
	return _Gateway.Contract.ERC20TokenPredicate(&_Gateway.CallOpts)
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

// Relayer is a free data retrieval call binding the contract method 0x8406c079.
//
// Solidity: function relayer() view returns(address)
func (_Gateway *GatewayCaller) Relayer(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Gateway.contract.Call(opts, &out, "relayer")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Relayer is a free data retrieval call binding the contract method 0x8406c079.
//
// Solidity: function relayer() view returns(address)
func (_Gateway *GatewaySession) Relayer() (common.Address, error) {
	return _Gateway.Contract.Relayer(&_Gateway.CallOpts)
}

// Relayer is a free data retrieval call binding the contract method 0x8406c079.
//
// Solidity: function relayer() view returns(address)
func (_Gateway *GatewayCallerSession) Relayer() (common.Address, error) {
	return _Gateway.Contract.Relayer(&_Gateway.CallOpts)
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

// AddValidatorChainData is a paid mutator transaction binding the contract method 0x6fc99a94.
//
// Solidity: function addValidatorChainData(address _addr, (uint256[4]) _data) returns()
func (_Gateway *GatewayTransactor) AddValidatorChainData(opts *bind.TransactOpts, _addr common.Address, _data IGatewayStructsValidatorChainData) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "addValidatorChainData", _addr, _data)
}

// AddValidatorChainData is a paid mutator transaction binding the contract method 0x6fc99a94.
//
// Solidity: function addValidatorChainData(address _addr, (uint256[4]) _data) returns()
func (_Gateway *GatewaySession) AddValidatorChainData(_addr common.Address, _data IGatewayStructsValidatorChainData) (*types.Transaction, error) {
	return _Gateway.Contract.AddValidatorChainData(&_Gateway.TransactOpts, _addr, _data)
}

// AddValidatorChainData is a paid mutator transaction binding the contract method 0x6fc99a94.
//
// Solidity: function addValidatorChainData(address _addr, (uint256[4]) _data) returns()
func (_Gateway *GatewayTransactorSession) AddValidatorChainData(_addr common.Address, _data IGatewayStructsValidatorChainData) (*types.Transaction, error) {
	return _Gateway.Contract.AddValidatorChainData(&_Gateway.TransactOpts, _addr, _data)
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

// DepositEvent is a paid mutator transaction binding the contract method 0x355600f0.
//
// Solidity: function depositEvent(bytes _data) returns()
func (_Gateway *GatewayTransactor) DepositEvent(opts *bind.TransactOpts, _data []byte) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "depositEvent", _data)
}

// DepositEvent is a paid mutator transaction binding the contract method 0x355600f0.
//
// Solidity: function depositEvent(bytes _data) returns()
func (_Gateway *GatewaySession) DepositEvent(_data []byte) (*types.Transaction, error) {
	return _Gateway.Contract.DepositEvent(&_Gateway.TransactOpts, _data)
}

// DepositEvent is a paid mutator transaction binding the contract method 0x355600f0.
//
// Solidity: function depositEvent(bytes _data) returns()
func (_Gateway *GatewayTransactorSession) DepositEvent(_data []byte) (*types.Transaction, error) {
	return _Gateway.Contract.DepositEvent(&_Gateway.TransactOpts, _data)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_Gateway *GatewayTransactor) Initialize(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "initialize")
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_Gateway *GatewaySession) Initialize() (*types.Transaction, error) {
	return _Gateway.Contract.Initialize(&_Gateway.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_Gateway *GatewayTransactorSession) Initialize() (*types.Transaction, error) {
	return _Gateway.Contract.Initialize(&_Gateway.TransactOpts)
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
// Solidity: function setDependencies(address _eRC20TokenPredicate, address _validators, address _relayer) returns()
func (_Gateway *GatewayTransactor) SetDependencies(opts *bind.TransactOpts, _eRC20TokenPredicate common.Address, _validators common.Address, _relayer common.Address) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "setDependencies", _eRC20TokenPredicate, _validators, _relayer)
}

// SetDependencies is a paid mutator transaction binding the contract method 0xb52d326c.
//
// Solidity: function setDependencies(address _eRC20TokenPredicate, address _validators, address _relayer) returns()
func (_Gateway *GatewaySession) SetDependencies(_eRC20TokenPredicate common.Address, _validators common.Address, _relayer common.Address) (*types.Transaction, error) {
	return _Gateway.Contract.SetDependencies(&_Gateway.TransactOpts, _eRC20TokenPredicate, _validators, _relayer)
}

// SetDependencies is a paid mutator transaction binding the contract method 0xb52d326c.
//
// Solidity: function setDependencies(address _eRC20TokenPredicate, address _validators, address _relayer) returns()
func (_Gateway *GatewayTransactorSession) SetDependencies(_eRC20TokenPredicate common.Address, _validators common.Address, _relayer common.Address) (*types.Transaction, error) {
	return _Gateway.Contract.SetDependencies(&_Gateway.TransactOpts, _eRC20TokenPredicate, _validators, _relayer)
}

// SetValidatorsChainData is a paid mutator transaction binding the contract method 0x85015d5b.
//
// Solidity: function setValidatorsChainData((address,(uint256[4]))[] _chainDatas) returns()
func (_Gateway *GatewayTransactor) SetValidatorsChainData(opts *bind.TransactOpts, _chainDatas []IGatewayStructsValidatorAddressChainData) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "setValidatorsChainData", _chainDatas)
}

// SetValidatorsChainData is a paid mutator transaction binding the contract method 0x85015d5b.
//
// Solidity: function setValidatorsChainData((address,(uint256[4]))[] _chainDatas) returns()
func (_Gateway *GatewaySession) SetValidatorsChainData(_chainDatas []IGatewayStructsValidatorAddressChainData) (*types.Transaction, error) {
	return _Gateway.Contract.SetValidatorsChainData(&_Gateway.TransactOpts, _chainDatas)
}

// SetValidatorsChainData is a paid mutator transaction binding the contract method 0x85015d5b.
//
// Solidity: function setValidatorsChainData((address,(uint256[4]))[] _chainDatas) returns()
func (_Gateway *GatewayTransactorSession) SetValidatorsChainData(_chainDatas []IGatewayStructsValidatorAddressChainData) (*types.Transaction, error) {
	return _Gateway.Contract.SetValidatorsChainData(&_Gateway.TransactOpts, _chainDatas)
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

// TtlEvent is a paid mutator transaction binding the contract method 0xf4b53d53.
//
// Solidity: function ttlEvent(bytes _data) returns()
func (_Gateway *GatewayTransactor) TtlEvent(opts *bind.TransactOpts, _data []byte) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "ttlEvent", _data)
}

// TtlEvent is a paid mutator transaction binding the contract method 0xf4b53d53.
//
// Solidity: function ttlEvent(bytes _data) returns()
func (_Gateway *GatewaySession) TtlEvent(_data []byte) (*types.Transaction, error) {
	return _Gateway.Contract.TtlEvent(&_Gateway.TransactOpts, _data)
}

// TtlEvent is a paid mutator transaction binding the contract method 0xf4b53d53.
//
// Solidity: function ttlEvent(bytes _data) returns()
func (_Gateway *GatewayTransactorSession) TtlEvent(_data []byte) (*types.Transaction, error) {
	return _Gateway.Contract.TtlEvent(&_Gateway.TransactOpts, _data)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_Gateway *GatewayTransactor) UpgradeTo(opts *bind.TransactOpts, newImplementation common.Address) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "upgradeTo", newImplementation)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_Gateway *GatewaySession) UpgradeTo(newImplementation common.Address) (*types.Transaction, error) {
	return _Gateway.Contract.UpgradeTo(&_Gateway.TransactOpts, newImplementation)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_Gateway *GatewayTransactorSession) UpgradeTo(newImplementation common.Address) (*types.Transaction, error) {
	return _Gateway.Contract.UpgradeTo(&_Gateway.TransactOpts, newImplementation)
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

// Withdraw is a paid mutator transaction binding the contract method 0x26021e91.
//
// Solidity: function withdraw(uint8 _destinationChainId, (string,uint256)[] _receivers, uint256 _feeAmount) returns()
func (_Gateway *GatewayTransactor) Withdraw(opts *bind.TransactOpts, _destinationChainId uint8, _receivers []IGatewayStructsReceiverWithdraw, _feeAmount *big.Int) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "withdraw", _destinationChainId, _receivers, _feeAmount)
}

// Withdraw is a paid mutator transaction binding the contract method 0x26021e91.
//
// Solidity: function withdraw(uint8 _destinationChainId, (string,uint256)[] _receivers, uint256 _feeAmount) returns()
func (_Gateway *GatewaySession) Withdraw(_destinationChainId uint8, _receivers []IGatewayStructsReceiverWithdraw, _feeAmount *big.Int) (*types.Transaction, error) {
	return _Gateway.Contract.Withdraw(&_Gateway.TransactOpts, _destinationChainId, _receivers, _feeAmount)
}

// Withdraw is a paid mutator transaction binding the contract method 0x26021e91.
//
// Solidity: function withdraw(uint8 _destinationChainId, (string,uint256)[] _receivers, uint256 _feeAmount) returns()
func (_Gateway *GatewayTransactorSession) Withdraw(_destinationChainId uint8, _receivers []IGatewayStructsReceiverWithdraw, _feeAmount *big.Int) (*types.Transaction, error) {
	return _Gateway.Contract.Withdraw(&_Gateway.TransactOpts, _destinationChainId, _receivers, _feeAmount)
}

// WithdrawEvent is a paid mutator transaction binding the contract method 0xa8de9f9c.
//
// Solidity: function withdrawEvent(uint8 _destinationChainId, address _sender, (string,uint256)[] _receivers, uint256 _feeAmount) returns()
func (_Gateway *GatewayTransactor) WithdrawEvent(opts *bind.TransactOpts, _destinationChainId uint8, _sender common.Address, _receivers []IGatewayStructsReceiverWithdraw, _feeAmount *big.Int) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "withdrawEvent", _destinationChainId, _sender, _receivers, _feeAmount)
}

// WithdrawEvent is a paid mutator transaction binding the contract method 0xa8de9f9c.
//
// Solidity: function withdrawEvent(uint8 _destinationChainId, address _sender, (string,uint256)[] _receivers, uint256 _feeAmount) returns()
func (_Gateway *GatewaySession) WithdrawEvent(_destinationChainId uint8, _sender common.Address, _receivers []IGatewayStructsReceiverWithdraw, _feeAmount *big.Int) (*types.Transaction, error) {
	return _Gateway.Contract.WithdrawEvent(&_Gateway.TransactOpts, _destinationChainId, _sender, _receivers, _feeAmount)
}

// WithdrawEvent is a paid mutator transaction binding the contract method 0xa8de9f9c.
//
// Solidity: function withdrawEvent(uint8 _destinationChainId, address _sender, (string,uint256)[] _receivers, uint256 _feeAmount) returns()
func (_Gateway *GatewayTransactorSession) WithdrawEvent(_destinationChainId uint8, _sender common.Address, _receivers []IGatewayStructsReceiverWithdraw, _feeAmount *big.Int) (*types.Transaction, error) {
	return _Gateway.Contract.WithdrawEvent(&_Gateway.TransactOpts, _destinationChainId, _sender, _receivers, _feeAmount)
}

// GatewayAdminChangedIterator is returned from FilterAdminChanged and is used to iterate over the raw logs and unpacked data for AdminChanged events raised by the Gateway contract.
type GatewayAdminChangedIterator struct {
	Event *GatewayAdminChanged // Event containing the contract specifics and raw log

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
func (it *GatewayAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatewayAdminChanged)
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
		it.Event = new(GatewayAdminChanged)
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
func (it *GatewayAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatewayAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatewayAdminChanged represents a AdminChanged event raised by the Gateway contract.
type GatewayAdminChanged struct {
	PreviousAdmin common.Address
	NewAdmin      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterAdminChanged is a free log retrieval operation binding the contract event 0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f.
//
// Solidity: event AdminChanged(address previousAdmin, address newAdmin)
func (_Gateway *GatewayFilterer) FilterAdminChanged(opts *bind.FilterOpts) (*GatewayAdminChangedIterator, error) {

	logs, sub, err := _Gateway.contract.FilterLogs(opts, "AdminChanged")
	if err != nil {
		return nil, err
	}
	return &GatewayAdminChangedIterator{contract: _Gateway.contract, event: "AdminChanged", logs: logs, sub: sub}, nil
}

// WatchAdminChanged is a free log subscription operation binding the contract event 0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f.
//
// Solidity: event AdminChanged(address previousAdmin, address newAdmin)
func (_Gateway *GatewayFilterer) WatchAdminChanged(opts *bind.WatchOpts, sink chan<- *GatewayAdminChanged) (event.Subscription, error) {

	logs, sub, err := _Gateway.contract.WatchLogs(opts, "AdminChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatewayAdminChanged)
				if err := _Gateway.contract.UnpackLog(event, "AdminChanged", log); err != nil {
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
func (_Gateway *GatewayFilterer) ParseAdminChanged(log types.Log) (*GatewayAdminChanged, error) {
	event := new(GatewayAdminChanged)
	if err := _Gateway.contract.UnpackLog(event, "AdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GatewayBeaconUpgradedIterator is returned from FilterBeaconUpgraded and is used to iterate over the raw logs and unpacked data for BeaconUpgraded events raised by the Gateway contract.
type GatewayBeaconUpgradedIterator struct {
	Event *GatewayBeaconUpgraded // Event containing the contract specifics and raw log

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
func (it *GatewayBeaconUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatewayBeaconUpgraded)
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
		it.Event = new(GatewayBeaconUpgraded)
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
func (it *GatewayBeaconUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatewayBeaconUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatewayBeaconUpgraded represents a BeaconUpgraded event raised by the Gateway contract.
type GatewayBeaconUpgraded struct {
	Beacon common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterBeaconUpgraded is a free log retrieval operation binding the contract event 0x1cf3b03a6cf19fa2baba4df148e9dcabedea7f8a5c07840e207e5c089be95d3e.
//
// Solidity: event BeaconUpgraded(address indexed beacon)
func (_Gateway *GatewayFilterer) FilterBeaconUpgraded(opts *bind.FilterOpts, beacon []common.Address) (*GatewayBeaconUpgradedIterator, error) {

	var beaconRule []interface{}
	for _, beaconItem := range beacon {
		beaconRule = append(beaconRule, beaconItem)
	}

	logs, sub, err := _Gateway.contract.FilterLogs(opts, "BeaconUpgraded", beaconRule)
	if err != nil {
		return nil, err
	}
	return &GatewayBeaconUpgradedIterator{contract: _Gateway.contract, event: "BeaconUpgraded", logs: logs, sub: sub}, nil
}

// WatchBeaconUpgraded is a free log subscription operation binding the contract event 0x1cf3b03a6cf19fa2baba4df148e9dcabedea7f8a5c07840e207e5c089be95d3e.
//
// Solidity: event BeaconUpgraded(address indexed beacon)
func (_Gateway *GatewayFilterer) WatchBeaconUpgraded(opts *bind.WatchOpts, sink chan<- *GatewayBeaconUpgraded, beacon []common.Address) (event.Subscription, error) {

	var beaconRule []interface{}
	for _, beaconItem := range beacon {
		beaconRule = append(beaconRule, beaconItem)
	}

	logs, sub, err := _Gateway.contract.WatchLogs(opts, "BeaconUpgraded", beaconRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatewayBeaconUpgraded)
				if err := _Gateway.contract.UnpackLog(event, "BeaconUpgraded", log); err != nil {
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
func (_Gateway *GatewayFilterer) ParseBeaconUpgraded(log types.Log) (*GatewayBeaconUpgraded, error) {
	event := new(GatewayBeaconUpgraded)
	if err := _Gateway.contract.UnpackLog(event, "BeaconUpgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
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
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Gateway *GatewayFilterer) FilterInitialized(opts *bind.FilterOpts) (*GatewayInitializedIterator, error) {

	logs, sub, err := _Gateway.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &GatewayInitializedIterator{contract: _Gateway.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
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

// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Gateway *GatewayFilterer) ParseInitialized(log types.Log) (*GatewayInitialized, error) {
	event := new(GatewayInitialized)
	if err := _Gateway.contract.UnpackLog(event, "Initialized", log); err != nil {
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
	FeeAmount          *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterWithdraw is a free log retrieval operation binding the contract event 0x2b846d03da343b397a350d2e88aa5091d29b87dd95204dc125870a82860416c8.
//
// Solidity: event Withdraw(uint8 destinationChainId, address sender, (string,uint256)[] receivers, uint256 feeAmount)
func (_Gateway *GatewayFilterer) FilterWithdraw(opts *bind.FilterOpts) (*GatewayWithdrawIterator, error) {

	logs, sub, err := _Gateway.contract.FilterLogs(opts, "Withdraw")
	if err != nil {
		return nil, err
	}
	return &GatewayWithdrawIterator{contract: _Gateway.contract, event: "Withdraw", logs: logs, sub: sub}, nil
}

// WatchWithdraw is a free log subscription operation binding the contract event 0x2b846d03da343b397a350d2e88aa5091d29b87dd95204dc125870a82860416c8.
//
// Solidity: event Withdraw(uint8 destinationChainId, address sender, (string,uint256)[] receivers, uint256 feeAmount)
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

// ParseWithdraw is a log parse operation binding the contract event 0x2b846d03da343b397a350d2e88aa5091d29b87dd95204dc125870a82860416c8.
//
// Solidity: event Withdraw(uint8 destinationChainId, address sender, (string,uint256)[] receivers, uint256 feeAmount)
func (_Gateway *GatewayFilterer) ParseWithdraw(log types.Log) (*GatewayWithdraw, error) {
	event := new(GatewayWithdraw)
	if err := _Gateway.contract.UnpackLog(event, "Withdraw", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
