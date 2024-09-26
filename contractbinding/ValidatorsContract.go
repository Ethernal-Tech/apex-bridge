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

// IGatewayStructsValidatorChainData is an auto generated low-level Go binding around an user-defined struct.
type IGatewayStructsValidatorChainData struct {
	Key [4]*big.Int
}

// ValidatorsMetaData contains all meta data concerning the Validators contract.
var ValidatorsMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"AddressEmptyCode\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"BatchAlreadyExecuted\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"ERC1967InvalidImplementation\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ERC1967NonPayable\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ExceedsMaxLength\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FailedInnerCall\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"data\",\"type\":\"string\"}],\"name\":\"InvalidData\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSignature\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotGateway\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotPredicate\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotPredicateOrOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TransferFailed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"UUPSUnauthorizedCallContext\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"slot\",\"type\":\"bytes32\"}],\"name\":\"UUPSUnsupportedProxiableUUID\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"expected\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"received\",\"type\":\"uint256\"}],\"name\":\"WrongValue\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroAddress\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"TTLExpired\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"Upgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"destinationChainId\",\"type\":\"uint8\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"indexed\":false,\"internalType\":\"structIGatewayStructs.ReceiverWithdraw[]\",\"name\":\"receivers\",\"type\":\"tuple[]\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"feeAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Withdraw\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"UPGRADE_INTERFACE_VERSION\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATOR_BLS_PRECOMPILE\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATOR_BLS_PRECOMPILE_GAS\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getValidatorsChainData\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256[4]\",\"name\":\"key\",\"type\":\"uint256[4]\"}],\"internalType\":\"structIGatewayStructs.ValidatorChainData[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_hash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"_signature\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"_bitmap\",\"type\":\"uint256\"}],\"name\":\"isBlsSignatureValid\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"proxiableUUID\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256[4]\",\"name\":\"key\",\"type\":\"uint256[4]\"}],\"internalType\":\"structIGatewayStructs.ValidatorChainData[]\",\"name\":\"_validatorsChainData\",\"type\":\"tuple[]\"}],\"name\":\"setValidatorsChainData\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"upgradeToAndCall\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
}

// ValidatorsABI is the input ABI used to generate the binding from.
// Deprecated: Use ValidatorsMetaData.ABI instead.
var ValidatorsABI = ValidatorsMetaData.ABI

// Validators is an auto generated Go binding around an Ethereum contract.
type Validators struct {
	ValidatorsCaller     // Read-only binding to the contract
	ValidatorsTransactor // Write-only binding to the contract
	ValidatorsFilterer   // Log filterer for contract events
}

// ValidatorsCaller is an auto generated read-only Go binding around an Ethereum contract.
type ValidatorsCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidatorsTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ValidatorsTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidatorsFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ValidatorsFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidatorsSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ValidatorsSession struct {
	Contract     *Validators       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ValidatorsCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ValidatorsCallerSession struct {
	Contract *ValidatorsCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// ValidatorsTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ValidatorsTransactorSession struct {
	Contract     *ValidatorsTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// ValidatorsRaw is an auto generated low-level Go binding around an Ethereum contract.
type ValidatorsRaw struct {
	Contract *Validators // Generic contract binding to access the raw methods on
}

// ValidatorsCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ValidatorsCallerRaw struct {
	Contract *ValidatorsCaller // Generic read-only contract binding to access the raw methods on
}

// ValidatorsTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ValidatorsTransactorRaw struct {
	Contract *ValidatorsTransactor // Generic write-only contract binding to access the raw methods on
}

// NewValidators creates a new instance of Validators, bound to a specific deployed contract.
func NewValidators(address common.Address, backend bind.ContractBackend) (*Validators, error) {
	contract, err := bindValidators(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Validators{ValidatorsCaller: ValidatorsCaller{contract: contract}, ValidatorsTransactor: ValidatorsTransactor{contract: contract}, ValidatorsFilterer: ValidatorsFilterer{contract: contract}}, nil
}

// NewValidatorsCaller creates a new read-only instance of Validators, bound to a specific deployed contract.
func NewValidatorsCaller(address common.Address, caller bind.ContractCaller) (*ValidatorsCaller, error) {
	contract, err := bindValidators(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ValidatorsCaller{contract: contract}, nil
}

// NewValidatorsTransactor creates a new write-only instance of Validators, bound to a specific deployed contract.
func NewValidatorsTransactor(address common.Address, transactor bind.ContractTransactor) (*ValidatorsTransactor, error) {
	contract, err := bindValidators(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ValidatorsTransactor{contract: contract}, nil
}

// NewValidatorsFilterer creates a new log filterer instance of Validators, bound to a specific deployed contract.
func NewValidatorsFilterer(address common.Address, filterer bind.ContractFilterer) (*ValidatorsFilterer, error) {
	contract, err := bindValidators(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ValidatorsFilterer{contract: contract}, nil
}

// bindValidators binds a generic wrapper to an already deployed contract.
func bindValidators(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ValidatorsMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Validators *ValidatorsRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Validators.Contract.ValidatorsCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Validators *ValidatorsRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Validators.Contract.ValidatorsTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Validators *ValidatorsRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Validators.Contract.ValidatorsTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Validators *ValidatorsCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Validators.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Validators *ValidatorsTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Validators.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Validators *ValidatorsTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Validators.Contract.contract.Transact(opts, method, params...)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_Validators *ValidatorsCaller) UPGRADEINTERFACEVERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Validators.contract.Call(opts, &out, "UPGRADE_INTERFACE_VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_Validators *ValidatorsSession) UPGRADEINTERFACEVERSION() (string, error) {
	return _Validators.Contract.UPGRADEINTERFACEVERSION(&_Validators.CallOpts)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_Validators *ValidatorsCallerSession) UPGRADEINTERFACEVERSION() (string, error) {
	return _Validators.Contract.UPGRADEINTERFACEVERSION(&_Validators.CallOpts)
}

// VALIDATORBLSPRECOMPILE is a free data retrieval call binding the contract method 0xd69c49ff.
//
// Solidity: function VALIDATOR_BLS_PRECOMPILE() view returns(address)
func (_Validators *ValidatorsCaller) VALIDATORBLSPRECOMPILE(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Validators.contract.Call(opts, &out, "VALIDATOR_BLS_PRECOMPILE")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// VALIDATORBLSPRECOMPILE is a free data retrieval call binding the contract method 0xd69c49ff.
//
// Solidity: function VALIDATOR_BLS_PRECOMPILE() view returns(address)
func (_Validators *ValidatorsSession) VALIDATORBLSPRECOMPILE() (common.Address, error) {
	return _Validators.Contract.VALIDATORBLSPRECOMPILE(&_Validators.CallOpts)
}

// VALIDATORBLSPRECOMPILE is a free data retrieval call binding the contract method 0xd69c49ff.
//
// Solidity: function VALIDATOR_BLS_PRECOMPILE() view returns(address)
func (_Validators *ValidatorsCallerSession) VALIDATORBLSPRECOMPILE() (common.Address, error) {
	return _Validators.Contract.VALIDATORBLSPRECOMPILE(&_Validators.CallOpts)
}

// VALIDATORBLSPRECOMPILEGAS is a free data retrieval call binding the contract method 0x4e58c78e.
//
// Solidity: function VALIDATOR_BLS_PRECOMPILE_GAS() view returns(uint256)
func (_Validators *ValidatorsCaller) VALIDATORBLSPRECOMPILEGAS(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Validators.contract.Call(opts, &out, "VALIDATOR_BLS_PRECOMPILE_GAS")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// VALIDATORBLSPRECOMPILEGAS is a free data retrieval call binding the contract method 0x4e58c78e.
//
// Solidity: function VALIDATOR_BLS_PRECOMPILE_GAS() view returns(uint256)
func (_Validators *ValidatorsSession) VALIDATORBLSPRECOMPILEGAS() (*big.Int, error) {
	return _Validators.Contract.VALIDATORBLSPRECOMPILEGAS(&_Validators.CallOpts)
}

// VALIDATORBLSPRECOMPILEGAS is a free data retrieval call binding the contract method 0x4e58c78e.
//
// Solidity: function VALIDATOR_BLS_PRECOMPILE_GAS() view returns(uint256)
func (_Validators *ValidatorsCallerSession) VALIDATORBLSPRECOMPILEGAS() (*big.Int, error) {
	return _Validators.Contract.VALIDATORBLSPRECOMPILEGAS(&_Validators.CallOpts)
}

// GetValidatorsChainData is a free data retrieval call binding the contract method 0x47f0daeb.
//
// Solidity: function getValidatorsChainData() view returns((uint256[4])[])
func (_Validators *ValidatorsCaller) GetValidatorsChainData(opts *bind.CallOpts) ([]IGatewayStructsValidatorChainData, error) {
	var out []interface{}
	err := _Validators.contract.Call(opts, &out, "getValidatorsChainData")

	if err != nil {
		return *new([]IGatewayStructsValidatorChainData), err
	}

	out0 := *abi.ConvertType(out[0], new([]IGatewayStructsValidatorChainData)).(*[]IGatewayStructsValidatorChainData)

	return out0, err

}

// GetValidatorsChainData is a free data retrieval call binding the contract method 0x47f0daeb.
//
// Solidity: function getValidatorsChainData() view returns((uint256[4])[])
func (_Validators *ValidatorsSession) GetValidatorsChainData() ([]IGatewayStructsValidatorChainData, error) {
	return _Validators.Contract.GetValidatorsChainData(&_Validators.CallOpts)
}

// GetValidatorsChainData is a free data retrieval call binding the contract method 0x47f0daeb.
//
// Solidity: function getValidatorsChainData() view returns((uint256[4])[])
func (_Validators *ValidatorsCallerSession) GetValidatorsChainData() ([]IGatewayStructsValidatorChainData, error) {
	return _Validators.Contract.GetValidatorsChainData(&_Validators.CallOpts)
}

// IsBlsSignatureValid is a free data retrieval call binding the contract method 0xde721b3a.
//
// Solidity: function isBlsSignatureValid(bytes32 _hash, bytes _signature, uint256 _bitmap) view returns(bool)
func (_Validators *ValidatorsCaller) IsBlsSignatureValid(opts *bind.CallOpts, _hash [32]byte, _signature []byte, _bitmap *big.Int) (bool, error) {
	var out []interface{}
	err := _Validators.contract.Call(opts, &out, "isBlsSignatureValid", _hash, _signature, _bitmap)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsBlsSignatureValid is a free data retrieval call binding the contract method 0xde721b3a.
//
// Solidity: function isBlsSignatureValid(bytes32 _hash, bytes _signature, uint256 _bitmap) view returns(bool)
func (_Validators *ValidatorsSession) IsBlsSignatureValid(_hash [32]byte, _signature []byte, _bitmap *big.Int) (bool, error) {
	return _Validators.Contract.IsBlsSignatureValid(&_Validators.CallOpts, _hash, _signature, _bitmap)
}

// IsBlsSignatureValid is a free data retrieval call binding the contract method 0xde721b3a.
//
// Solidity: function isBlsSignatureValid(bytes32 _hash, bytes _signature, uint256 _bitmap) view returns(bool)
func (_Validators *ValidatorsCallerSession) IsBlsSignatureValid(_hash [32]byte, _signature []byte, _bitmap *big.Int) (bool, error) {
	return _Validators.Contract.IsBlsSignatureValid(&_Validators.CallOpts, _hash, _signature, _bitmap)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Validators *ValidatorsCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Validators.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Validators *ValidatorsSession) Owner() (common.Address, error) {
	return _Validators.Contract.Owner(&_Validators.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Validators *ValidatorsCallerSession) Owner() (common.Address, error) {
	return _Validators.Contract.Owner(&_Validators.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_Validators *ValidatorsCaller) ProxiableUUID(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Validators.contract.Call(opts, &out, "proxiableUUID")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_Validators *ValidatorsSession) ProxiableUUID() ([32]byte, error) {
	return _Validators.Contract.ProxiableUUID(&_Validators.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_Validators *ValidatorsCallerSession) ProxiableUUID() ([32]byte, error) {
	return _Validators.Contract.ProxiableUUID(&_Validators.CallOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_Validators *ValidatorsTransactor) Initialize(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Validators.contract.Transact(opts, "initialize")
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_Validators *ValidatorsSession) Initialize() (*types.Transaction, error) {
	return _Validators.Contract.Initialize(&_Validators.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_Validators *ValidatorsTransactorSession) Initialize() (*types.Transaction, error) {
	return _Validators.Contract.Initialize(&_Validators.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Validators *ValidatorsTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Validators.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Validators *ValidatorsSession) RenounceOwnership() (*types.Transaction, error) {
	return _Validators.Contract.RenounceOwnership(&_Validators.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Validators *ValidatorsTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Validators.Contract.RenounceOwnership(&_Validators.TransactOpts)
}

// SetValidatorsChainData is a paid mutator transaction binding the contract method 0x420f231c.
//
// Solidity: function setValidatorsChainData((uint256[4])[] _validatorsChainData) returns()
func (_Validators *ValidatorsTransactor) SetValidatorsChainData(opts *bind.TransactOpts, _validatorsChainData []IGatewayStructsValidatorChainData) (*types.Transaction, error) {
	return _Validators.contract.Transact(opts, "setValidatorsChainData", _validatorsChainData)
}

// SetValidatorsChainData is a paid mutator transaction binding the contract method 0x420f231c.
//
// Solidity: function setValidatorsChainData((uint256[4])[] _validatorsChainData) returns()
func (_Validators *ValidatorsSession) SetValidatorsChainData(_validatorsChainData []IGatewayStructsValidatorChainData) (*types.Transaction, error) {
	return _Validators.Contract.SetValidatorsChainData(&_Validators.TransactOpts, _validatorsChainData)
}

// SetValidatorsChainData is a paid mutator transaction binding the contract method 0x420f231c.
//
// Solidity: function setValidatorsChainData((uint256[4])[] _validatorsChainData) returns()
func (_Validators *ValidatorsTransactorSession) SetValidatorsChainData(_validatorsChainData []IGatewayStructsValidatorChainData) (*types.Transaction, error) {
	return _Validators.Contract.SetValidatorsChainData(&_Validators.TransactOpts, _validatorsChainData)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Validators *ValidatorsTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Validators.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Validators *ValidatorsSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Validators.Contract.TransferOwnership(&_Validators.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Validators *ValidatorsTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Validators.Contract.TransferOwnership(&_Validators.TransactOpts, newOwner)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_Validators *ValidatorsTransactor) UpgradeToAndCall(opts *bind.TransactOpts, newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _Validators.contract.Transact(opts, "upgradeToAndCall", newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_Validators *ValidatorsSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _Validators.Contract.UpgradeToAndCall(&_Validators.TransactOpts, newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_Validators *ValidatorsTransactorSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _Validators.Contract.UpgradeToAndCall(&_Validators.TransactOpts, newImplementation, data)
}

// ValidatorsDepositIterator is returned from FilterDeposit and is used to iterate over the raw logs and unpacked data for Deposit events raised by the Validators contract.
type ValidatorsDepositIterator struct {
	Event *ValidatorsDeposit // Event containing the contract specifics and raw log

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
func (it *ValidatorsDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorsDeposit)
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
		it.Event = new(ValidatorsDeposit)
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
func (it *ValidatorsDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorsDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorsDeposit represents a Deposit event raised by the Validators contract.
type ValidatorsDeposit struct {
	Data []byte
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0x7adcde22575d10ee3d4e78ee24cc9f854ecc4ce2bc5fda5eadeb754384227db0.
//
// Solidity: event Deposit(bytes data)
func (_Validators *ValidatorsFilterer) FilterDeposit(opts *bind.FilterOpts) (*ValidatorsDepositIterator, error) {

	logs, sub, err := _Validators.contract.FilterLogs(opts, "Deposit")
	if err != nil {
		return nil, err
	}
	return &ValidatorsDepositIterator{contract: _Validators.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0x7adcde22575d10ee3d4e78ee24cc9f854ecc4ce2bc5fda5eadeb754384227db0.
//
// Solidity: event Deposit(bytes data)
func (_Validators *ValidatorsFilterer) WatchDeposit(opts *bind.WatchOpts, sink chan<- *ValidatorsDeposit) (event.Subscription, error) {

	logs, sub, err := _Validators.contract.WatchLogs(opts, "Deposit")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorsDeposit)
				if err := _Validators.contract.UnpackLog(event, "Deposit", log); err != nil {
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
func (_Validators *ValidatorsFilterer) ParseDeposit(log types.Log) (*ValidatorsDeposit, error) {
	event := new(ValidatorsDeposit)
	if err := _Validators.contract.UnpackLog(event, "Deposit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ValidatorsInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Validators contract.
type ValidatorsInitializedIterator struct {
	Event *ValidatorsInitialized // Event containing the contract specifics and raw log

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
func (it *ValidatorsInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorsInitialized)
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
		it.Event = new(ValidatorsInitialized)
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
func (it *ValidatorsInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorsInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorsInitialized represents a Initialized event raised by the Validators contract.
type ValidatorsInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Validators *ValidatorsFilterer) FilterInitialized(opts *bind.FilterOpts) (*ValidatorsInitializedIterator, error) {

	logs, sub, err := _Validators.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &ValidatorsInitializedIterator{contract: _Validators.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Validators *ValidatorsFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *ValidatorsInitialized) (event.Subscription, error) {

	logs, sub, err := _Validators.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorsInitialized)
				if err := _Validators.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_Validators *ValidatorsFilterer) ParseInitialized(log types.Log) (*ValidatorsInitialized, error) {
	event := new(ValidatorsInitialized)
	if err := _Validators.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ValidatorsOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Validators contract.
type ValidatorsOwnershipTransferredIterator struct {
	Event *ValidatorsOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *ValidatorsOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorsOwnershipTransferred)
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
		it.Event = new(ValidatorsOwnershipTransferred)
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
func (it *ValidatorsOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorsOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorsOwnershipTransferred represents a OwnershipTransferred event raised by the Validators contract.
type ValidatorsOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Validators *ValidatorsFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ValidatorsOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Validators.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ValidatorsOwnershipTransferredIterator{contract: _Validators.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Validators *ValidatorsFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ValidatorsOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Validators.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorsOwnershipTransferred)
				if err := _Validators.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_Validators *ValidatorsFilterer) ParseOwnershipTransferred(log types.Log) (*ValidatorsOwnershipTransferred, error) {
	event := new(ValidatorsOwnershipTransferred)
	if err := _Validators.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ValidatorsTTLExpiredIterator is returned from FilterTTLExpired and is used to iterate over the raw logs and unpacked data for TTLExpired events raised by the Validators contract.
type ValidatorsTTLExpiredIterator struct {
	Event *ValidatorsTTLExpired // Event containing the contract specifics and raw log

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
func (it *ValidatorsTTLExpiredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorsTTLExpired)
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
		it.Event = new(ValidatorsTTLExpired)
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
func (it *ValidatorsTTLExpiredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorsTTLExpiredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorsTTLExpired represents a TTLExpired event raised by the Validators contract.
type ValidatorsTTLExpired struct {
	Data []byte
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterTTLExpired is a free log retrieval operation binding the contract event 0x0ade0be4bc31d69de480eac67556041c790defee8b124e80c09859865d56af09.
//
// Solidity: event TTLExpired(bytes data)
func (_Validators *ValidatorsFilterer) FilterTTLExpired(opts *bind.FilterOpts) (*ValidatorsTTLExpiredIterator, error) {

	logs, sub, err := _Validators.contract.FilterLogs(opts, "TTLExpired")
	if err != nil {
		return nil, err
	}
	return &ValidatorsTTLExpiredIterator{contract: _Validators.contract, event: "TTLExpired", logs: logs, sub: sub}, nil
}

// WatchTTLExpired is a free log subscription operation binding the contract event 0x0ade0be4bc31d69de480eac67556041c790defee8b124e80c09859865d56af09.
//
// Solidity: event TTLExpired(bytes data)
func (_Validators *ValidatorsFilterer) WatchTTLExpired(opts *bind.WatchOpts, sink chan<- *ValidatorsTTLExpired) (event.Subscription, error) {

	logs, sub, err := _Validators.contract.WatchLogs(opts, "TTLExpired")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorsTTLExpired)
				if err := _Validators.contract.UnpackLog(event, "TTLExpired", log); err != nil {
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
func (_Validators *ValidatorsFilterer) ParseTTLExpired(log types.Log) (*ValidatorsTTLExpired, error) {
	event := new(ValidatorsTTLExpired)
	if err := _Validators.contract.UnpackLog(event, "TTLExpired", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ValidatorsUpgradedIterator is returned from FilterUpgraded and is used to iterate over the raw logs and unpacked data for Upgraded events raised by the Validators contract.
type ValidatorsUpgradedIterator struct {
	Event *ValidatorsUpgraded // Event containing the contract specifics and raw log

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
func (it *ValidatorsUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorsUpgraded)
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
		it.Event = new(ValidatorsUpgraded)
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
func (it *ValidatorsUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorsUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorsUpgraded represents a Upgraded event raised by the Validators contract.
type ValidatorsUpgraded struct {
	Implementation common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterUpgraded is a free log retrieval operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_Validators *ValidatorsFilterer) FilterUpgraded(opts *bind.FilterOpts, implementation []common.Address) (*ValidatorsUpgradedIterator, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _Validators.contract.FilterLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return &ValidatorsUpgradedIterator{contract: _Validators.contract, event: "Upgraded", logs: logs, sub: sub}, nil
}

// WatchUpgraded is a free log subscription operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_Validators *ValidatorsFilterer) WatchUpgraded(opts *bind.WatchOpts, sink chan<- *ValidatorsUpgraded, implementation []common.Address) (event.Subscription, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _Validators.contract.WatchLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorsUpgraded)
				if err := _Validators.contract.UnpackLog(event, "Upgraded", log); err != nil {
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
func (_Validators *ValidatorsFilterer) ParseUpgraded(log types.Log) (*ValidatorsUpgraded, error) {
	event := new(ValidatorsUpgraded)
	if err := _Validators.contract.UnpackLog(event, "Upgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ValidatorsWithdrawIterator is returned from FilterWithdraw and is used to iterate over the raw logs and unpacked data for Withdraw events raised by the Validators contract.
type ValidatorsWithdrawIterator struct {
	Event *ValidatorsWithdraw // Event containing the contract specifics and raw log

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
func (it *ValidatorsWithdrawIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorsWithdraw)
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
		it.Event = new(ValidatorsWithdraw)
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
func (it *ValidatorsWithdrawIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorsWithdrawIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorsWithdraw represents a Withdraw event raised by the Validators contract.
type ValidatorsWithdraw struct {
	DestinationChainId uint8
	Sender             common.Address
	Receivers          []IGatewayStructsReceiverWithdraw
	FeeAmount          *big.Int
	Value              *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterWithdraw is a free log retrieval operation binding the contract event 0x0f1583d54fa562a43b77e01026b8541efbe6a4bb8452e60302fe1140dd520281.
//
// Solidity: event Withdraw(uint8 destinationChainId, address sender, (string,uint256)[] receivers, uint256 feeAmount, uint256 value)
func (_Validators *ValidatorsFilterer) FilterWithdraw(opts *bind.FilterOpts) (*ValidatorsWithdrawIterator, error) {

	logs, sub, err := _Validators.contract.FilterLogs(opts, "Withdraw")
	if err != nil {
		return nil, err
	}
	return &ValidatorsWithdrawIterator{contract: _Validators.contract, event: "Withdraw", logs: logs, sub: sub}, nil
}

// WatchWithdraw is a free log subscription operation binding the contract event 0x0f1583d54fa562a43b77e01026b8541efbe6a4bb8452e60302fe1140dd520281.
//
// Solidity: event Withdraw(uint8 destinationChainId, address sender, (string,uint256)[] receivers, uint256 feeAmount, uint256 value)
func (_Validators *ValidatorsFilterer) WatchWithdraw(opts *bind.WatchOpts, sink chan<- *ValidatorsWithdraw) (event.Subscription, error) {

	logs, sub, err := _Validators.contract.WatchLogs(opts, "Withdraw")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorsWithdraw)
				if err := _Validators.contract.UnpackLog(event, "Withdraw", log); err != nil {
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

// ParseWithdraw is a log parse operation binding the contract event 0x0f1583d54fa562a43b77e01026b8541efbe6a4bb8452e60302fe1140dd520281.
//
// Solidity: event Withdraw(uint8 destinationChainId, address sender, (string,uint256)[] receivers, uint256 feeAmount, uint256 value)
func (_Validators *ValidatorsFilterer) ParseWithdraw(log types.Log) (*ValidatorsWithdraw, error) {
	event := new(ValidatorsWithdraw)
	if err := _Validators.contract.UnpackLog(event, "Withdraw", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
