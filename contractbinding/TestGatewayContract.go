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

// TestGatewayReceiverWithdraw is an auto generated low-level Go binding around an user-defined struct.
type TestGatewayReceiverWithdraw struct {
	Receiver string
	Amount   *big.Int
}

// TestGatewayMetaData contains all meta data concerning the TestGateway contract.
var TestGatewayMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"destinationChainId\",\"type\":\"uint8\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"indexed\":false,\"internalType\":\"structTestGateway.ReceiverWithdraw[]\",\"name\":\"receivers\",\"type\":\"tuple[]\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"feeAmount\",\"type\":\"uint256\"}],\"name\":\"Withdraw\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"destinationChainId\",\"type\":\"uint8\"},{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structTestGateway.ReceiverWithdraw[]\",\"name\":\"receivers\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256\",\"name\":\"feeAmount\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// TestGatewayABI is the input ABI used to generate the binding from.
// Deprecated: Use TestGatewayMetaData.ABI instead.
var TestGatewayABI = TestGatewayMetaData.ABI

// TestGateway is an auto generated Go binding around an Ethereum contract.
type TestGateway struct {
	TestGatewayCaller     // Read-only binding to the contract
	TestGatewayTransactor // Write-only binding to the contract
	TestGatewayFilterer   // Log filterer for contract events
}

// TestGatewayCaller is an auto generated read-only Go binding around an Ethereum contract.
type TestGatewayCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestGatewayTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TestGatewayTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestGatewayFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TestGatewayFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestGatewaySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TestGatewaySession struct {
	Contract     *TestGateway      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// TestGatewayCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TestGatewayCallerSession struct {
	Contract *TestGatewayCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// TestGatewayTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TestGatewayTransactorSession struct {
	Contract     *TestGatewayTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// TestGatewayRaw is an auto generated low-level Go binding around an Ethereum contract.
type TestGatewayRaw struct {
	Contract *TestGateway // Generic contract binding to access the raw methods on
}

// TestGatewayCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TestGatewayCallerRaw struct {
	Contract *TestGatewayCaller // Generic read-only contract binding to access the raw methods on
}

// TestGatewayTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TestGatewayTransactorRaw struct {
	Contract *TestGatewayTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTestGateway creates a new instance of TestGateway, bound to a specific deployed contract.
func NewTestGateway(address common.Address, backend bind.ContractBackend) (*TestGateway, error) {
	contract, err := bindTestGateway(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &TestGateway{TestGatewayCaller: TestGatewayCaller{contract: contract}, TestGatewayTransactor: TestGatewayTransactor{contract: contract}, TestGatewayFilterer: TestGatewayFilterer{contract: contract}}, nil
}

// NewTestGatewayCaller creates a new read-only instance of TestGateway, bound to a specific deployed contract.
func NewTestGatewayCaller(address common.Address, caller bind.ContractCaller) (*TestGatewayCaller, error) {
	contract, err := bindTestGateway(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TestGatewayCaller{contract: contract}, nil
}

// NewTestGatewayTransactor creates a new write-only instance of TestGateway, bound to a specific deployed contract.
func NewTestGatewayTransactor(address common.Address, transactor bind.ContractTransactor) (*TestGatewayTransactor, error) {
	contract, err := bindTestGateway(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TestGatewayTransactor{contract: contract}, nil
}

// NewTestGatewayFilterer creates a new log filterer instance of TestGateway, bound to a specific deployed contract.
func NewTestGatewayFilterer(address common.Address, filterer bind.ContractFilterer) (*TestGatewayFilterer, error) {
	contract, err := bindTestGateway(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TestGatewayFilterer{contract: contract}, nil
}

// bindTestGateway binds a generic wrapper to an already deployed contract.
func bindTestGateway(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := TestGatewayMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TestGateway *TestGatewayRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TestGateway.Contract.TestGatewayCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TestGateway *TestGatewayRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestGateway.Contract.TestGatewayTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TestGateway *TestGatewayRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TestGateway.Contract.TestGatewayTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TestGateway *TestGatewayCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TestGateway.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TestGateway *TestGatewayTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestGateway.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TestGateway *TestGatewayTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TestGateway.Contract.contract.Transact(opts, method, params...)
}

// Deposit is a paid mutator transaction binding the contract method 0x98b1e06a.
//
// Solidity: function deposit(bytes _data) returns()
func (_TestGateway *TestGatewayTransactor) Deposit(opts *bind.TransactOpts, _data []byte) (*types.Transaction, error) {
	return _TestGateway.contract.Transact(opts, "deposit", _data)
}

// Deposit is a paid mutator transaction binding the contract method 0x98b1e06a.
//
// Solidity: function deposit(bytes _data) returns()
func (_TestGateway *TestGatewaySession) Deposit(_data []byte) (*types.Transaction, error) {
	return _TestGateway.Contract.Deposit(&_TestGateway.TransactOpts, _data)
}

// Deposit is a paid mutator transaction binding the contract method 0x98b1e06a.
//
// Solidity: function deposit(bytes _data) returns()
func (_TestGateway *TestGatewayTransactorSession) Deposit(_data []byte) (*types.Transaction, error) {
	return _TestGateway.Contract.Deposit(&_TestGateway.TransactOpts, _data)
}

// Withdraw is a paid mutator transaction binding the contract method 0xfa398db8.
//
// Solidity: function withdraw(uint8 destinationChainId, address sender, (string,uint256)[] receivers, uint256 feeAmount) returns()
func (_TestGateway *TestGatewayTransactor) Withdraw(opts *bind.TransactOpts, destinationChainId uint8, sender common.Address, receivers []TestGatewayReceiverWithdraw, feeAmount *big.Int) (*types.Transaction, error) {
	return _TestGateway.contract.Transact(opts, "withdraw", destinationChainId, sender, receivers, feeAmount)
}

// Withdraw is a paid mutator transaction binding the contract method 0xfa398db8.
//
// Solidity: function withdraw(uint8 destinationChainId, address sender, (string,uint256)[] receivers, uint256 feeAmount) returns()
func (_TestGateway *TestGatewaySession) Withdraw(destinationChainId uint8, sender common.Address, receivers []TestGatewayReceiverWithdraw, feeAmount *big.Int) (*types.Transaction, error) {
	return _TestGateway.Contract.Withdraw(&_TestGateway.TransactOpts, destinationChainId, sender, receivers, feeAmount)
}

// Withdraw is a paid mutator transaction binding the contract method 0xfa398db8.
//
// Solidity: function withdraw(uint8 destinationChainId, address sender, (string,uint256)[] receivers, uint256 feeAmount) returns()
func (_TestGateway *TestGatewayTransactorSession) Withdraw(destinationChainId uint8, sender common.Address, receivers []TestGatewayReceiverWithdraw, feeAmount *big.Int) (*types.Transaction, error) {
	return _TestGateway.Contract.Withdraw(&_TestGateway.TransactOpts, destinationChainId, sender, receivers, feeAmount)
}

// TestGatewayDepositIterator is returned from FilterDeposit and is used to iterate over the raw logs and unpacked data for Deposit events raised by the TestGateway contract.
type TestGatewayDepositIterator struct {
	Event *TestGatewayDeposit // Event containing the contract specifics and raw log

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
func (it *TestGatewayDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TestGatewayDeposit)
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
		it.Event = new(TestGatewayDeposit)
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
func (it *TestGatewayDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TestGatewayDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TestGatewayDeposit represents a Deposit event raised by the TestGateway contract.
type TestGatewayDeposit struct {
	Data []byte
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0x7adcde22575d10ee3d4e78ee24cc9f854ecc4ce2bc5fda5eadeb754384227db0.
//
// Solidity: event Deposit(bytes data)
func (_TestGateway *TestGatewayFilterer) FilterDeposit(opts *bind.FilterOpts) (*TestGatewayDepositIterator, error) {

	logs, sub, err := _TestGateway.contract.FilterLogs(opts, "Deposit")
	if err != nil {
		return nil, err
	}
	return &TestGatewayDepositIterator{contract: _TestGateway.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0x7adcde22575d10ee3d4e78ee24cc9f854ecc4ce2bc5fda5eadeb754384227db0.
//
// Solidity: event Deposit(bytes data)
func (_TestGateway *TestGatewayFilterer) WatchDeposit(opts *bind.WatchOpts, sink chan<- *TestGatewayDeposit) (event.Subscription, error) {

	logs, sub, err := _TestGateway.contract.WatchLogs(opts, "Deposit")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TestGatewayDeposit)
				if err := _TestGateway.contract.UnpackLog(event, "Deposit", log); err != nil {
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
func (_TestGateway *TestGatewayFilterer) ParseDeposit(log types.Log) (*TestGatewayDeposit, error) {
	event := new(TestGatewayDeposit)
	if err := _TestGateway.contract.UnpackLog(event, "Deposit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TestGatewayWithdrawIterator is returned from FilterWithdraw and is used to iterate over the raw logs and unpacked data for Withdraw events raised by the TestGateway contract.
type TestGatewayWithdrawIterator struct {
	Event *TestGatewayWithdraw // Event containing the contract specifics and raw log

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
func (it *TestGatewayWithdrawIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TestGatewayWithdraw)
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
		it.Event = new(TestGatewayWithdraw)
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
func (it *TestGatewayWithdrawIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TestGatewayWithdrawIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TestGatewayWithdraw represents a Withdraw event raised by the TestGateway contract.
type TestGatewayWithdraw struct {
	DestinationChainId uint8
	Sender             common.Address
	Receivers          []TestGatewayReceiverWithdraw
	FeeAmount          *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterWithdraw is a free log retrieval operation binding the contract event 0x2b846d03da343b397a350d2e88aa5091d29b87dd95204dc125870a82860416c8.
//
// Solidity: event Withdraw(uint8 destinationChainId, address sender, (string,uint256)[] receivers, uint256 feeAmount)
func (_TestGateway *TestGatewayFilterer) FilterWithdraw(opts *bind.FilterOpts) (*TestGatewayWithdrawIterator, error) {

	logs, sub, err := _TestGateway.contract.FilterLogs(opts, "Withdraw")
	if err != nil {
		return nil, err
	}
	return &TestGatewayWithdrawIterator{contract: _TestGateway.contract, event: "Withdraw", logs: logs, sub: sub}, nil
}

// WatchWithdraw is a free log subscription operation binding the contract event 0x2b846d03da343b397a350d2e88aa5091d29b87dd95204dc125870a82860416c8.
//
// Solidity: event Withdraw(uint8 destinationChainId, address sender, (string,uint256)[] receivers, uint256 feeAmount)
func (_TestGateway *TestGatewayFilterer) WatchWithdraw(opts *bind.WatchOpts, sink chan<- *TestGatewayWithdraw) (event.Subscription, error) {

	logs, sub, err := _TestGateway.contract.WatchLogs(opts, "Withdraw")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TestGatewayWithdraw)
				if err := _TestGateway.contract.UnpackLog(event, "Withdraw", log); err != nil {
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
func (_TestGateway *TestGatewayFilterer) ParseWithdraw(log types.Log) (*TestGatewayWithdraw, error) {
	event := new(TestGatewayWithdraw)
	if err := _TestGateway.contract.UnpackLog(event, "Withdraw", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
