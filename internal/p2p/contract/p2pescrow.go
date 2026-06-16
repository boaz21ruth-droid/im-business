// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

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

// P2PEscrowMetaData contains all meta data concerning the P2PEscrow contract.
var P2PEscrowMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_arbitrator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_feeRate\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_feeRecipient\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"CONFIRM_TIMEOUT\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MAX_LOCK_TIMEOUT\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MIN_LOCK_TIMEOUT\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"arbitrator\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"cancelExpiredLock\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"cancelOpenOrder\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"confirmRelease\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"createOrder\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"fiatAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"fiatCurrency\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"payMethod\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"lockTimeout\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"feeRate\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"feeRecipient\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"markPaid\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"orderCount\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"orders\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"seller\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"buyer\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"fiatAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"fiatCurrency\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"payMethod\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumP2PEscrow.OrderStatus\"},{\"name\":\"createdAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"lockedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"paidAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"lockTimeout\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"raiseDispute\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"resolveDispute\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"buyerWins\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setArbitrator\",\"inputs\":[{\"name\":\"_arbitrator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setFeeRate\",\"inputs\":[{\"name\":\"_feeRate\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setFeeRecipient\",\"inputs\":[{\"name\":\"_feeRecipient\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"takeOrder\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"DisputeRaised\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"raiser\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"DisputeResolved\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"buyerWins\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OrderCancelled\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"actor\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OrderCreated\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"seller\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OrderReleased\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"buyer\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"fee\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OrderTaken\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"buyer\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PaymentMarked\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"buyer\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ReentrancyGuardReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"SafeERC20FailedOperation\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}]}]",
}

// P2PEscrowABI is the input ABI used to generate the binding from.
// Deprecated: Use P2PEscrowMetaData.ABI instead.
var P2PEscrowABI = P2PEscrowMetaData.ABI

// P2PEscrow is an auto generated Go binding around an Ethereum contract.
type P2PEscrow struct {
	P2PEscrowCaller     // Read-only binding to the contract
	P2PEscrowTransactor // Write-only binding to the contract
	P2PEscrowFilterer   // Log filterer for contract events
}

// P2PEscrowCaller is an auto generated read-only Go binding around an Ethereum contract.
type P2PEscrowCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// P2PEscrowTransactor is an auto generated write-only Go binding around an Ethereum contract.
type P2PEscrowTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// P2PEscrowFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type P2PEscrowFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// P2PEscrowSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type P2PEscrowSession struct {
	Contract     *P2PEscrow        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// P2PEscrowCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type P2PEscrowCallerSession struct {
	Contract *P2PEscrowCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// P2PEscrowTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type P2PEscrowTransactorSession struct {
	Contract     *P2PEscrowTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// P2PEscrowRaw is an auto generated low-level Go binding around an Ethereum contract.
type P2PEscrowRaw struct {
	Contract *P2PEscrow // Generic contract binding to access the raw methods on
}

// P2PEscrowCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type P2PEscrowCallerRaw struct {
	Contract *P2PEscrowCaller // Generic read-only contract binding to access the raw methods on
}

// P2PEscrowTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type P2PEscrowTransactorRaw struct {
	Contract *P2PEscrowTransactor // Generic write-only contract binding to access the raw methods on
}

// NewP2PEscrow creates a new instance of P2PEscrow, bound to a specific deployed contract.
func NewP2PEscrow(address common.Address, backend bind.ContractBackend) (*P2PEscrow, error) {
	contract, err := bindP2PEscrow(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &P2PEscrow{P2PEscrowCaller: P2PEscrowCaller{contract: contract}, P2PEscrowTransactor: P2PEscrowTransactor{contract: contract}, P2PEscrowFilterer: P2PEscrowFilterer{contract: contract}}, nil
}

// NewP2PEscrowCaller creates a new read-only instance of P2PEscrow, bound to a specific deployed contract.
func NewP2PEscrowCaller(address common.Address, caller bind.ContractCaller) (*P2PEscrowCaller, error) {
	contract, err := bindP2PEscrow(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &P2PEscrowCaller{contract: contract}, nil
}

// NewP2PEscrowTransactor creates a new write-only instance of P2PEscrow, bound to a specific deployed contract.
func NewP2PEscrowTransactor(address common.Address, transactor bind.ContractTransactor) (*P2PEscrowTransactor, error) {
	contract, err := bindP2PEscrow(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &P2PEscrowTransactor{contract: contract}, nil
}

// NewP2PEscrowFilterer creates a new log filterer instance of P2PEscrow, bound to a specific deployed contract.
func NewP2PEscrowFilterer(address common.Address, filterer bind.ContractFilterer) (*P2PEscrowFilterer, error) {
	contract, err := bindP2PEscrow(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &P2PEscrowFilterer{contract: contract}, nil
}

// bindP2PEscrow binds a generic wrapper to an already deployed contract.
func bindP2PEscrow(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := P2PEscrowMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_P2PEscrow *P2PEscrowRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _P2PEscrow.Contract.P2PEscrowCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_P2PEscrow *P2PEscrowRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _P2PEscrow.Contract.P2PEscrowTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_P2PEscrow *P2PEscrowRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _P2PEscrow.Contract.P2PEscrowTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_P2PEscrow *P2PEscrowCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _P2PEscrow.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_P2PEscrow *P2PEscrowTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _P2PEscrow.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_P2PEscrow *P2PEscrowTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _P2PEscrow.Contract.contract.Transact(opts, method, params...)
}

// CONFIRMTIMEOUT is a free data retrieval call binding the contract method 0xd3161813.
//
// Solidity: function CONFIRM_TIMEOUT() view returns(uint256)
func (_P2PEscrow *P2PEscrowCaller) CONFIRMTIMEOUT(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _P2PEscrow.contract.Call(opts, &out, "CONFIRM_TIMEOUT")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CONFIRMTIMEOUT is a free data retrieval call binding the contract method 0xd3161813.
//
// Solidity: function CONFIRM_TIMEOUT() view returns(uint256)
func (_P2PEscrow *P2PEscrowSession) CONFIRMTIMEOUT() (*big.Int, error) {
	return _P2PEscrow.Contract.CONFIRMTIMEOUT(&_P2PEscrow.CallOpts)
}

// CONFIRMTIMEOUT is a free data retrieval call binding the contract method 0xd3161813.
//
// Solidity: function CONFIRM_TIMEOUT() view returns(uint256)
func (_P2PEscrow *P2PEscrowCallerSession) CONFIRMTIMEOUT() (*big.Int, error) {
	return _P2PEscrow.Contract.CONFIRMTIMEOUT(&_P2PEscrow.CallOpts)
}

// MAXLOCKTIMEOUT is a free data retrieval call binding the contract method 0xbb429bac.
//
// Solidity: function MAX_LOCK_TIMEOUT() view returns(uint256)
func (_P2PEscrow *P2PEscrowCaller) MAXLOCKTIMEOUT(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _P2PEscrow.contract.Call(opts, &out, "MAX_LOCK_TIMEOUT")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXLOCKTIMEOUT is a free data retrieval call binding the contract method 0xbb429bac.
//
// Solidity: function MAX_LOCK_TIMEOUT() view returns(uint256)
func (_P2PEscrow *P2PEscrowSession) MAXLOCKTIMEOUT() (*big.Int, error) {
	return _P2PEscrow.Contract.MAXLOCKTIMEOUT(&_P2PEscrow.CallOpts)
}

// MAXLOCKTIMEOUT is a free data retrieval call binding the contract method 0xbb429bac.
//
// Solidity: function MAX_LOCK_TIMEOUT() view returns(uint256)
func (_P2PEscrow *P2PEscrowCallerSession) MAXLOCKTIMEOUT() (*big.Int, error) {
	return _P2PEscrow.Contract.MAXLOCKTIMEOUT(&_P2PEscrow.CallOpts)
}

// MINLOCKTIMEOUT is a free data retrieval call binding the contract method 0x47d8dfa3.
//
// Solidity: function MIN_LOCK_TIMEOUT() view returns(uint256)
func (_P2PEscrow *P2PEscrowCaller) MINLOCKTIMEOUT(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _P2PEscrow.contract.Call(opts, &out, "MIN_LOCK_TIMEOUT")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MINLOCKTIMEOUT is a free data retrieval call binding the contract method 0x47d8dfa3.
//
// Solidity: function MIN_LOCK_TIMEOUT() view returns(uint256)
func (_P2PEscrow *P2PEscrowSession) MINLOCKTIMEOUT() (*big.Int, error) {
	return _P2PEscrow.Contract.MINLOCKTIMEOUT(&_P2PEscrow.CallOpts)
}

// MINLOCKTIMEOUT is a free data retrieval call binding the contract method 0x47d8dfa3.
//
// Solidity: function MIN_LOCK_TIMEOUT() view returns(uint256)
func (_P2PEscrow *P2PEscrowCallerSession) MINLOCKTIMEOUT() (*big.Int, error) {
	return _P2PEscrow.Contract.MINLOCKTIMEOUT(&_P2PEscrow.CallOpts)
}

// Arbitrator is a free data retrieval call binding the contract method 0x6cc6cde1.
//
// Solidity: function arbitrator() view returns(address)
func (_P2PEscrow *P2PEscrowCaller) Arbitrator(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _P2PEscrow.contract.Call(opts, &out, "arbitrator")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Arbitrator is a free data retrieval call binding the contract method 0x6cc6cde1.
//
// Solidity: function arbitrator() view returns(address)
func (_P2PEscrow *P2PEscrowSession) Arbitrator() (common.Address, error) {
	return _P2PEscrow.Contract.Arbitrator(&_P2PEscrow.CallOpts)
}

// Arbitrator is a free data retrieval call binding the contract method 0x6cc6cde1.
//
// Solidity: function arbitrator() view returns(address)
func (_P2PEscrow *P2PEscrowCallerSession) Arbitrator() (common.Address, error) {
	return _P2PEscrow.Contract.Arbitrator(&_P2PEscrow.CallOpts)
}

// FeeRate is a free data retrieval call binding the contract method 0x978bbdb9.
//
// Solidity: function feeRate() view returns(uint256)
func (_P2PEscrow *P2PEscrowCaller) FeeRate(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _P2PEscrow.contract.Call(opts, &out, "feeRate")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FeeRate is a free data retrieval call binding the contract method 0x978bbdb9.
//
// Solidity: function feeRate() view returns(uint256)
func (_P2PEscrow *P2PEscrowSession) FeeRate() (*big.Int, error) {
	return _P2PEscrow.Contract.FeeRate(&_P2PEscrow.CallOpts)
}

// FeeRate is a free data retrieval call binding the contract method 0x978bbdb9.
//
// Solidity: function feeRate() view returns(uint256)
func (_P2PEscrow *P2PEscrowCallerSession) FeeRate() (*big.Int, error) {
	return _P2PEscrow.Contract.FeeRate(&_P2PEscrow.CallOpts)
}

// FeeRecipient is a free data retrieval call binding the contract method 0x46904840.
//
// Solidity: function feeRecipient() view returns(address)
func (_P2PEscrow *P2PEscrowCaller) FeeRecipient(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _P2PEscrow.contract.Call(opts, &out, "feeRecipient")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// FeeRecipient is a free data retrieval call binding the contract method 0x46904840.
//
// Solidity: function feeRecipient() view returns(address)
func (_P2PEscrow *P2PEscrowSession) FeeRecipient() (common.Address, error) {
	return _P2PEscrow.Contract.FeeRecipient(&_P2PEscrow.CallOpts)
}

// FeeRecipient is a free data retrieval call binding the contract method 0x46904840.
//
// Solidity: function feeRecipient() view returns(address)
func (_P2PEscrow *P2PEscrowCallerSession) FeeRecipient() (common.Address, error) {
	return _P2PEscrow.Contract.FeeRecipient(&_P2PEscrow.CallOpts)
}

// OrderCount is a free data retrieval call binding the contract method 0x2453ffa8.
//
// Solidity: function orderCount() view returns(uint256)
func (_P2PEscrow *P2PEscrowCaller) OrderCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _P2PEscrow.contract.Call(opts, &out, "orderCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// OrderCount is a free data retrieval call binding the contract method 0x2453ffa8.
//
// Solidity: function orderCount() view returns(uint256)
func (_P2PEscrow *P2PEscrowSession) OrderCount() (*big.Int, error) {
	return _P2PEscrow.Contract.OrderCount(&_P2PEscrow.CallOpts)
}

// OrderCount is a free data retrieval call binding the contract method 0x2453ffa8.
//
// Solidity: function orderCount() view returns(uint256)
func (_P2PEscrow *P2PEscrowCallerSession) OrderCount() (*big.Int, error) {
	return _P2PEscrow.Contract.OrderCount(&_P2PEscrow.CallOpts)
}

// Orders is a free data retrieval call binding the contract method 0xa85c38ef.
//
// Solidity: function orders(uint256 ) view returns(address seller, address buyer, address token, uint256 amount, uint256 fiatAmount, string fiatCurrency, string payMethod, uint8 status, uint256 createdAt, uint256 lockedAt, uint256 paidAt, uint256 lockTimeout)
func (_P2PEscrow *P2PEscrowCaller) Orders(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Seller       common.Address
	Buyer        common.Address
	Token        common.Address
	Amount       *big.Int
	FiatAmount   *big.Int
	FiatCurrency string
	PayMethod    string
	Status       uint8
	CreatedAt    *big.Int
	LockedAt     *big.Int
	PaidAt       *big.Int
	LockTimeout  *big.Int
}, error) {
	var out []interface{}
	err := _P2PEscrow.contract.Call(opts, &out, "orders", arg0)

	outstruct := new(struct {
		Seller       common.Address
		Buyer        common.Address
		Token        common.Address
		Amount       *big.Int
		FiatAmount   *big.Int
		FiatCurrency string
		PayMethod    string
		Status       uint8
		CreatedAt    *big.Int
		LockedAt     *big.Int
		PaidAt       *big.Int
		LockTimeout  *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Seller = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Buyer = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)
	outstruct.Token = *abi.ConvertType(out[2], new(common.Address)).(*common.Address)
	outstruct.Amount = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.FiatAmount = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.FiatCurrency = *abi.ConvertType(out[5], new(string)).(*string)
	outstruct.PayMethod = *abi.ConvertType(out[6], new(string)).(*string)
	outstruct.Status = *abi.ConvertType(out[7], new(uint8)).(*uint8)
	outstruct.CreatedAt = *abi.ConvertType(out[8], new(*big.Int)).(**big.Int)
	outstruct.LockedAt = *abi.ConvertType(out[9], new(*big.Int)).(**big.Int)
	outstruct.PaidAt = *abi.ConvertType(out[10], new(*big.Int)).(**big.Int)
	outstruct.LockTimeout = *abi.ConvertType(out[11], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Orders is a free data retrieval call binding the contract method 0xa85c38ef.
//
// Solidity: function orders(uint256 ) view returns(address seller, address buyer, address token, uint256 amount, uint256 fiatAmount, string fiatCurrency, string payMethod, uint8 status, uint256 createdAt, uint256 lockedAt, uint256 paidAt, uint256 lockTimeout)
func (_P2PEscrow *P2PEscrowSession) Orders(arg0 *big.Int) (struct {
	Seller       common.Address
	Buyer        common.Address
	Token        common.Address
	Amount       *big.Int
	FiatAmount   *big.Int
	FiatCurrency string
	PayMethod    string
	Status       uint8
	CreatedAt    *big.Int
	LockedAt     *big.Int
	PaidAt       *big.Int
	LockTimeout  *big.Int
}, error) {
	return _P2PEscrow.Contract.Orders(&_P2PEscrow.CallOpts, arg0)
}

// Orders is a free data retrieval call binding the contract method 0xa85c38ef.
//
// Solidity: function orders(uint256 ) view returns(address seller, address buyer, address token, uint256 amount, uint256 fiatAmount, string fiatCurrency, string payMethod, uint8 status, uint256 createdAt, uint256 lockedAt, uint256 paidAt, uint256 lockTimeout)
func (_P2PEscrow *P2PEscrowCallerSession) Orders(arg0 *big.Int) (struct {
	Seller       common.Address
	Buyer        common.Address
	Token        common.Address
	Amount       *big.Int
	FiatAmount   *big.Int
	FiatCurrency string
	PayMethod    string
	Status       uint8
	CreatedAt    *big.Int
	LockedAt     *big.Int
	PaidAt       *big.Int
	LockTimeout  *big.Int
}, error) {
	return _P2PEscrow.Contract.Orders(&_P2PEscrow.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_P2PEscrow *P2PEscrowCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _P2PEscrow.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_P2PEscrow *P2PEscrowSession) Owner() (common.Address, error) {
	return _P2PEscrow.Contract.Owner(&_P2PEscrow.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_P2PEscrow *P2PEscrowCallerSession) Owner() (common.Address, error) {
	return _P2PEscrow.Contract.Owner(&_P2PEscrow.CallOpts)
}

// CancelExpiredLock is a paid mutator transaction binding the contract method 0x69fffccb.
//
// Solidity: function cancelExpiredLock(uint256 orderId) returns()
func (_P2PEscrow *P2PEscrowTransactor) CancelExpiredLock(opts *bind.TransactOpts, orderId *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.contract.Transact(opts, "cancelExpiredLock", orderId)
}

// CancelExpiredLock is a paid mutator transaction binding the contract method 0x69fffccb.
//
// Solidity: function cancelExpiredLock(uint256 orderId) returns()
func (_P2PEscrow *P2PEscrowSession) CancelExpiredLock(orderId *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.Contract.CancelExpiredLock(&_P2PEscrow.TransactOpts, orderId)
}

// CancelExpiredLock is a paid mutator transaction binding the contract method 0x69fffccb.
//
// Solidity: function cancelExpiredLock(uint256 orderId) returns()
func (_P2PEscrow *P2PEscrowTransactorSession) CancelExpiredLock(orderId *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.Contract.CancelExpiredLock(&_P2PEscrow.TransactOpts, orderId)
}

// CancelOpenOrder is a paid mutator transaction binding the contract method 0x812538e1.
//
// Solidity: function cancelOpenOrder(uint256 orderId) returns()
func (_P2PEscrow *P2PEscrowTransactor) CancelOpenOrder(opts *bind.TransactOpts, orderId *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.contract.Transact(opts, "cancelOpenOrder", orderId)
}

// CancelOpenOrder is a paid mutator transaction binding the contract method 0x812538e1.
//
// Solidity: function cancelOpenOrder(uint256 orderId) returns()
func (_P2PEscrow *P2PEscrowSession) CancelOpenOrder(orderId *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.Contract.CancelOpenOrder(&_P2PEscrow.TransactOpts, orderId)
}

// CancelOpenOrder is a paid mutator transaction binding the contract method 0x812538e1.
//
// Solidity: function cancelOpenOrder(uint256 orderId) returns()
func (_P2PEscrow *P2PEscrowTransactorSession) CancelOpenOrder(orderId *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.Contract.CancelOpenOrder(&_P2PEscrow.TransactOpts, orderId)
}

// ConfirmRelease is a paid mutator transaction binding the contract method 0xba6bc66e.
//
// Solidity: function confirmRelease(uint256 orderId) returns()
func (_P2PEscrow *P2PEscrowTransactor) ConfirmRelease(opts *bind.TransactOpts, orderId *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.contract.Transact(opts, "confirmRelease", orderId)
}

// ConfirmRelease is a paid mutator transaction binding the contract method 0xba6bc66e.
//
// Solidity: function confirmRelease(uint256 orderId) returns()
func (_P2PEscrow *P2PEscrowSession) ConfirmRelease(orderId *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.Contract.ConfirmRelease(&_P2PEscrow.TransactOpts, orderId)
}

// ConfirmRelease is a paid mutator transaction binding the contract method 0xba6bc66e.
//
// Solidity: function confirmRelease(uint256 orderId) returns()
func (_P2PEscrow *P2PEscrowTransactorSession) ConfirmRelease(orderId *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.Contract.ConfirmRelease(&_P2PEscrow.TransactOpts, orderId)
}

// CreateOrder is a paid mutator transaction binding the contract method 0x49d871b7.
//
// Solidity: function createOrder(address token, uint256 amount, uint256 fiatAmount, string fiatCurrency, string payMethod, uint256 lockTimeout) returns(uint256 orderId)
func (_P2PEscrow *P2PEscrowTransactor) CreateOrder(opts *bind.TransactOpts, token common.Address, amount *big.Int, fiatAmount *big.Int, fiatCurrency string, payMethod string, lockTimeout *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.contract.Transact(opts, "createOrder", token, amount, fiatAmount, fiatCurrency, payMethod, lockTimeout)
}

// CreateOrder is a paid mutator transaction binding the contract method 0x49d871b7.
//
// Solidity: function createOrder(address token, uint256 amount, uint256 fiatAmount, string fiatCurrency, string payMethod, uint256 lockTimeout) returns(uint256 orderId)
func (_P2PEscrow *P2PEscrowSession) CreateOrder(token common.Address, amount *big.Int, fiatAmount *big.Int, fiatCurrency string, payMethod string, lockTimeout *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.Contract.CreateOrder(&_P2PEscrow.TransactOpts, token, amount, fiatAmount, fiatCurrency, payMethod, lockTimeout)
}

// CreateOrder is a paid mutator transaction binding the contract method 0x49d871b7.
//
// Solidity: function createOrder(address token, uint256 amount, uint256 fiatAmount, string fiatCurrency, string payMethod, uint256 lockTimeout) returns(uint256 orderId)
func (_P2PEscrow *P2PEscrowTransactorSession) CreateOrder(token common.Address, amount *big.Int, fiatAmount *big.Int, fiatCurrency string, payMethod string, lockTimeout *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.Contract.CreateOrder(&_P2PEscrow.TransactOpts, token, amount, fiatAmount, fiatCurrency, payMethod, lockTimeout)
}

// MarkPaid is a paid mutator transaction binding the contract method 0xad207e1e.
//
// Solidity: function markPaid(uint256 orderId) returns()
func (_P2PEscrow *P2PEscrowTransactor) MarkPaid(opts *bind.TransactOpts, orderId *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.contract.Transact(opts, "markPaid", orderId)
}

// MarkPaid is a paid mutator transaction binding the contract method 0xad207e1e.
//
// Solidity: function markPaid(uint256 orderId) returns()
func (_P2PEscrow *P2PEscrowSession) MarkPaid(orderId *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.Contract.MarkPaid(&_P2PEscrow.TransactOpts, orderId)
}

// MarkPaid is a paid mutator transaction binding the contract method 0xad207e1e.
//
// Solidity: function markPaid(uint256 orderId) returns()
func (_P2PEscrow *P2PEscrowTransactorSession) MarkPaid(orderId *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.Contract.MarkPaid(&_P2PEscrow.TransactOpts, orderId)
}

// RaiseDispute is a paid mutator transaction binding the contract method 0xa5c1674e.
//
// Solidity: function raiseDispute(uint256 orderId) returns()
func (_P2PEscrow *P2PEscrowTransactor) RaiseDispute(opts *bind.TransactOpts, orderId *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.contract.Transact(opts, "raiseDispute", orderId)
}

// RaiseDispute is a paid mutator transaction binding the contract method 0xa5c1674e.
//
// Solidity: function raiseDispute(uint256 orderId) returns()
func (_P2PEscrow *P2PEscrowSession) RaiseDispute(orderId *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.Contract.RaiseDispute(&_P2PEscrow.TransactOpts, orderId)
}

// RaiseDispute is a paid mutator transaction binding the contract method 0xa5c1674e.
//
// Solidity: function raiseDispute(uint256 orderId) returns()
func (_P2PEscrow *P2PEscrowTransactorSession) RaiseDispute(orderId *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.Contract.RaiseDispute(&_P2PEscrow.TransactOpts, orderId)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_P2PEscrow *P2PEscrowTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _P2PEscrow.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_P2PEscrow *P2PEscrowSession) RenounceOwnership() (*types.Transaction, error) {
	return _P2PEscrow.Contract.RenounceOwnership(&_P2PEscrow.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_P2PEscrow *P2PEscrowTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _P2PEscrow.Contract.RenounceOwnership(&_P2PEscrow.TransactOpts)
}

// ResolveDispute is a paid mutator transaction binding the contract method 0x34b25ee2.
//
// Solidity: function resolveDispute(uint256 orderId, bool buyerWins) returns()
func (_P2PEscrow *P2PEscrowTransactor) ResolveDispute(opts *bind.TransactOpts, orderId *big.Int, buyerWins bool) (*types.Transaction, error) {
	return _P2PEscrow.contract.Transact(opts, "resolveDispute", orderId, buyerWins)
}

// ResolveDispute is a paid mutator transaction binding the contract method 0x34b25ee2.
//
// Solidity: function resolveDispute(uint256 orderId, bool buyerWins) returns()
func (_P2PEscrow *P2PEscrowSession) ResolveDispute(orderId *big.Int, buyerWins bool) (*types.Transaction, error) {
	return _P2PEscrow.Contract.ResolveDispute(&_P2PEscrow.TransactOpts, orderId, buyerWins)
}

// ResolveDispute is a paid mutator transaction binding the contract method 0x34b25ee2.
//
// Solidity: function resolveDispute(uint256 orderId, bool buyerWins) returns()
func (_P2PEscrow *P2PEscrowTransactorSession) ResolveDispute(orderId *big.Int, buyerWins bool) (*types.Transaction, error) {
	return _P2PEscrow.Contract.ResolveDispute(&_P2PEscrow.TransactOpts, orderId, buyerWins)
}

// SetArbitrator is a paid mutator transaction binding the contract method 0xb0eefabe.
//
// Solidity: function setArbitrator(address _arbitrator) returns()
func (_P2PEscrow *P2PEscrowTransactor) SetArbitrator(opts *bind.TransactOpts, _arbitrator common.Address) (*types.Transaction, error) {
	return _P2PEscrow.contract.Transact(opts, "setArbitrator", _arbitrator)
}

// SetArbitrator is a paid mutator transaction binding the contract method 0xb0eefabe.
//
// Solidity: function setArbitrator(address _arbitrator) returns()
func (_P2PEscrow *P2PEscrowSession) SetArbitrator(_arbitrator common.Address) (*types.Transaction, error) {
	return _P2PEscrow.Contract.SetArbitrator(&_P2PEscrow.TransactOpts, _arbitrator)
}

// SetArbitrator is a paid mutator transaction binding the contract method 0xb0eefabe.
//
// Solidity: function setArbitrator(address _arbitrator) returns()
func (_P2PEscrow *P2PEscrowTransactorSession) SetArbitrator(_arbitrator common.Address) (*types.Transaction, error) {
	return _P2PEscrow.Contract.SetArbitrator(&_P2PEscrow.TransactOpts, _arbitrator)
}

// SetFeeRate is a paid mutator transaction binding the contract method 0x45596e2e.
//
// Solidity: function setFeeRate(uint256 _feeRate) returns()
func (_P2PEscrow *P2PEscrowTransactor) SetFeeRate(opts *bind.TransactOpts, _feeRate *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.contract.Transact(opts, "setFeeRate", _feeRate)
}

// SetFeeRate is a paid mutator transaction binding the contract method 0x45596e2e.
//
// Solidity: function setFeeRate(uint256 _feeRate) returns()
func (_P2PEscrow *P2PEscrowSession) SetFeeRate(_feeRate *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.Contract.SetFeeRate(&_P2PEscrow.TransactOpts, _feeRate)
}

// SetFeeRate is a paid mutator transaction binding the contract method 0x45596e2e.
//
// Solidity: function setFeeRate(uint256 _feeRate) returns()
func (_P2PEscrow *P2PEscrowTransactorSession) SetFeeRate(_feeRate *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.Contract.SetFeeRate(&_P2PEscrow.TransactOpts, _feeRate)
}

// SetFeeRecipient is a paid mutator transaction binding the contract method 0xe74b981b.
//
// Solidity: function setFeeRecipient(address _feeRecipient) returns()
func (_P2PEscrow *P2PEscrowTransactor) SetFeeRecipient(opts *bind.TransactOpts, _feeRecipient common.Address) (*types.Transaction, error) {
	return _P2PEscrow.contract.Transact(opts, "setFeeRecipient", _feeRecipient)
}

// SetFeeRecipient is a paid mutator transaction binding the contract method 0xe74b981b.
//
// Solidity: function setFeeRecipient(address _feeRecipient) returns()
func (_P2PEscrow *P2PEscrowSession) SetFeeRecipient(_feeRecipient common.Address) (*types.Transaction, error) {
	return _P2PEscrow.Contract.SetFeeRecipient(&_P2PEscrow.TransactOpts, _feeRecipient)
}

// SetFeeRecipient is a paid mutator transaction binding the contract method 0xe74b981b.
//
// Solidity: function setFeeRecipient(address _feeRecipient) returns()
func (_P2PEscrow *P2PEscrowTransactorSession) SetFeeRecipient(_feeRecipient common.Address) (*types.Transaction, error) {
	return _P2PEscrow.Contract.SetFeeRecipient(&_P2PEscrow.TransactOpts, _feeRecipient)
}

// TakeOrder is a paid mutator transaction binding the contract method 0x3b5d1a24.
//
// Solidity: function takeOrder(uint256 orderId) returns()
func (_P2PEscrow *P2PEscrowTransactor) TakeOrder(opts *bind.TransactOpts, orderId *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.contract.Transact(opts, "takeOrder", orderId)
}

// TakeOrder is a paid mutator transaction binding the contract method 0x3b5d1a24.
//
// Solidity: function takeOrder(uint256 orderId) returns()
func (_P2PEscrow *P2PEscrowSession) TakeOrder(orderId *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.Contract.TakeOrder(&_P2PEscrow.TransactOpts, orderId)
}

// TakeOrder is a paid mutator transaction binding the contract method 0x3b5d1a24.
//
// Solidity: function takeOrder(uint256 orderId) returns()
func (_P2PEscrow *P2PEscrowTransactorSession) TakeOrder(orderId *big.Int) (*types.Transaction, error) {
	return _P2PEscrow.Contract.TakeOrder(&_P2PEscrow.TransactOpts, orderId)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_P2PEscrow *P2PEscrowTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _P2PEscrow.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_P2PEscrow *P2PEscrowSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _P2PEscrow.Contract.TransferOwnership(&_P2PEscrow.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_P2PEscrow *P2PEscrowTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _P2PEscrow.Contract.TransferOwnership(&_P2PEscrow.TransactOpts, newOwner)
}

// P2PEscrowDisputeRaisedIterator is returned from FilterDisputeRaised and is used to iterate over the raw logs and unpacked data for DisputeRaised events raised by the P2PEscrow contract.
type P2PEscrowDisputeRaisedIterator struct {
	Event *P2PEscrowDisputeRaised // Event containing the contract specifics and raw log

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
func (it *P2PEscrowDisputeRaisedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(P2PEscrowDisputeRaised)
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
		it.Event = new(P2PEscrowDisputeRaised)
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
func (it *P2PEscrowDisputeRaisedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *P2PEscrowDisputeRaisedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// P2PEscrowDisputeRaised represents a DisputeRaised event raised by the P2PEscrow contract.
type P2PEscrowDisputeRaised struct {
	OrderId *big.Int
	Raiser  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterDisputeRaised is a free log retrieval operation binding the contract event 0x84a477df8a28a4276ca6dee4458a06c3015f30c477d9c949ede4e13ff8a552b4.
//
// Solidity: event DisputeRaised(uint256 indexed orderId, address indexed raiser)
func (_P2PEscrow *P2PEscrowFilterer) FilterDisputeRaised(opts *bind.FilterOpts, orderId []*big.Int, raiser []common.Address) (*P2PEscrowDisputeRaisedIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var raiserRule []interface{}
	for _, raiserItem := range raiser {
		raiserRule = append(raiserRule, raiserItem)
	}

	logs, sub, err := _P2PEscrow.contract.FilterLogs(opts, "DisputeRaised", orderIdRule, raiserRule)
	if err != nil {
		return nil, err
	}
	return &P2PEscrowDisputeRaisedIterator{contract: _P2PEscrow.contract, event: "DisputeRaised", logs: logs, sub: sub}, nil
}

// WatchDisputeRaised is a free log subscription operation binding the contract event 0x84a477df8a28a4276ca6dee4458a06c3015f30c477d9c949ede4e13ff8a552b4.
//
// Solidity: event DisputeRaised(uint256 indexed orderId, address indexed raiser)
func (_P2PEscrow *P2PEscrowFilterer) WatchDisputeRaised(opts *bind.WatchOpts, sink chan<- *P2PEscrowDisputeRaised, orderId []*big.Int, raiser []common.Address) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var raiserRule []interface{}
	for _, raiserItem := range raiser {
		raiserRule = append(raiserRule, raiserItem)
	}

	logs, sub, err := _P2PEscrow.contract.WatchLogs(opts, "DisputeRaised", orderIdRule, raiserRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(P2PEscrowDisputeRaised)
				if err := _P2PEscrow.contract.UnpackLog(event, "DisputeRaised", log); err != nil {
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

// ParseDisputeRaised is a log parse operation binding the contract event 0x84a477df8a28a4276ca6dee4458a06c3015f30c477d9c949ede4e13ff8a552b4.
//
// Solidity: event DisputeRaised(uint256 indexed orderId, address indexed raiser)
func (_P2PEscrow *P2PEscrowFilterer) ParseDisputeRaised(log types.Log) (*P2PEscrowDisputeRaised, error) {
	event := new(P2PEscrowDisputeRaised)
	if err := _P2PEscrow.contract.UnpackLog(event, "DisputeRaised", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// P2PEscrowDisputeResolvedIterator is returned from FilterDisputeResolved and is used to iterate over the raw logs and unpacked data for DisputeResolved events raised by the P2PEscrow contract.
type P2PEscrowDisputeResolvedIterator struct {
	Event *P2PEscrowDisputeResolved // Event containing the contract specifics and raw log

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
func (it *P2PEscrowDisputeResolvedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(P2PEscrowDisputeResolved)
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
		it.Event = new(P2PEscrowDisputeResolved)
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
func (it *P2PEscrowDisputeResolvedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *P2PEscrowDisputeResolvedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// P2PEscrowDisputeResolved represents a DisputeResolved event raised by the P2PEscrow contract.
type P2PEscrowDisputeResolved struct {
	OrderId   *big.Int
	BuyerWins bool
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDisputeResolved is a free log retrieval operation binding the contract event 0x5a87909bff68caaaaf0b3fd9c74eeccc928832f879315e5c6fb7a73612f26c0c.
//
// Solidity: event DisputeResolved(uint256 indexed orderId, bool buyerWins)
func (_P2PEscrow *P2PEscrowFilterer) FilterDisputeResolved(opts *bind.FilterOpts, orderId []*big.Int) (*P2PEscrowDisputeResolvedIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _P2PEscrow.contract.FilterLogs(opts, "DisputeResolved", orderIdRule)
	if err != nil {
		return nil, err
	}
	return &P2PEscrowDisputeResolvedIterator{contract: _P2PEscrow.contract, event: "DisputeResolved", logs: logs, sub: sub}, nil
}

// WatchDisputeResolved is a free log subscription operation binding the contract event 0x5a87909bff68caaaaf0b3fd9c74eeccc928832f879315e5c6fb7a73612f26c0c.
//
// Solidity: event DisputeResolved(uint256 indexed orderId, bool buyerWins)
func (_P2PEscrow *P2PEscrowFilterer) WatchDisputeResolved(opts *bind.WatchOpts, sink chan<- *P2PEscrowDisputeResolved, orderId []*big.Int) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _P2PEscrow.contract.WatchLogs(opts, "DisputeResolved", orderIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(P2PEscrowDisputeResolved)
				if err := _P2PEscrow.contract.UnpackLog(event, "DisputeResolved", log); err != nil {
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

// ParseDisputeResolved is a log parse operation binding the contract event 0x5a87909bff68caaaaf0b3fd9c74eeccc928832f879315e5c6fb7a73612f26c0c.
//
// Solidity: event DisputeResolved(uint256 indexed orderId, bool buyerWins)
func (_P2PEscrow *P2PEscrowFilterer) ParseDisputeResolved(log types.Log) (*P2PEscrowDisputeResolved, error) {
	event := new(P2PEscrowDisputeResolved)
	if err := _P2PEscrow.contract.UnpackLog(event, "DisputeResolved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// P2PEscrowOrderCancelledIterator is returned from FilterOrderCancelled and is used to iterate over the raw logs and unpacked data for OrderCancelled events raised by the P2PEscrow contract.
type P2PEscrowOrderCancelledIterator struct {
	Event *P2PEscrowOrderCancelled // Event containing the contract specifics and raw log

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
func (it *P2PEscrowOrderCancelledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(P2PEscrowOrderCancelled)
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
		it.Event = new(P2PEscrowOrderCancelled)
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
func (it *P2PEscrowOrderCancelledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *P2PEscrowOrderCancelledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// P2PEscrowOrderCancelled represents a OrderCancelled event raised by the P2PEscrow contract.
type P2PEscrowOrderCancelled struct {
	OrderId *big.Int
	Actor   common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterOrderCancelled is a free log retrieval operation binding the contract event 0xc0362da6f2ff36b382b34aec0814f6b3cdf89f5ef282a1d1f114d0c0b036d596.
//
// Solidity: event OrderCancelled(uint256 indexed orderId, address indexed actor)
func (_P2PEscrow *P2PEscrowFilterer) FilterOrderCancelled(opts *bind.FilterOpts, orderId []*big.Int, actor []common.Address) (*P2PEscrowOrderCancelledIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var actorRule []interface{}
	for _, actorItem := range actor {
		actorRule = append(actorRule, actorItem)
	}

	logs, sub, err := _P2PEscrow.contract.FilterLogs(opts, "OrderCancelled", orderIdRule, actorRule)
	if err != nil {
		return nil, err
	}
	return &P2PEscrowOrderCancelledIterator{contract: _P2PEscrow.contract, event: "OrderCancelled", logs: logs, sub: sub}, nil
}

// WatchOrderCancelled is a free log subscription operation binding the contract event 0xc0362da6f2ff36b382b34aec0814f6b3cdf89f5ef282a1d1f114d0c0b036d596.
//
// Solidity: event OrderCancelled(uint256 indexed orderId, address indexed actor)
func (_P2PEscrow *P2PEscrowFilterer) WatchOrderCancelled(opts *bind.WatchOpts, sink chan<- *P2PEscrowOrderCancelled, orderId []*big.Int, actor []common.Address) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var actorRule []interface{}
	for _, actorItem := range actor {
		actorRule = append(actorRule, actorItem)
	}

	logs, sub, err := _P2PEscrow.contract.WatchLogs(opts, "OrderCancelled", orderIdRule, actorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(P2PEscrowOrderCancelled)
				if err := _P2PEscrow.contract.UnpackLog(event, "OrderCancelled", log); err != nil {
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

// ParseOrderCancelled is a log parse operation binding the contract event 0xc0362da6f2ff36b382b34aec0814f6b3cdf89f5ef282a1d1f114d0c0b036d596.
//
// Solidity: event OrderCancelled(uint256 indexed orderId, address indexed actor)
func (_P2PEscrow *P2PEscrowFilterer) ParseOrderCancelled(log types.Log) (*P2PEscrowOrderCancelled, error) {
	event := new(P2PEscrowOrderCancelled)
	if err := _P2PEscrow.contract.UnpackLog(event, "OrderCancelled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// P2PEscrowOrderCreatedIterator is returned from FilterOrderCreated and is used to iterate over the raw logs and unpacked data for OrderCreated events raised by the P2PEscrow contract.
type P2PEscrowOrderCreatedIterator struct {
	Event *P2PEscrowOrderCreated // Event containing the contract specifics and raw log

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
func (it *P2PEscrowOrderCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(P2PEscrowOrderCreated)
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
		it.Event = new(P2PEscrowOrderCreated)
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
func (it *P2PEscrowOrderCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *P2PEscrowOrderCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// P2PEscrowOrderCreated represents a OrderCreated event raised by the P2PEscrow contract.
type P2PEscrowOrderCreated struct {
	OrderId *big.Int
	Seller  common.Address
	Token   common.Address
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterOrderCreated is a free log retrieval operation binding the contract event 0xf9e17940e547201709f9aaacdec8ab3566601c60d8f5affb54f74b79b0b50b13.
//
// Solidity: event OrderCreated(uint256 indexed orderId, address indexed seller, address token, uint256 amount)
func (_P2PEscrow *P2PEscrowFilterer) FilterOrderCreated(opts *bind.FilterOpts, orderId []*big.Int, seller []common.Address) (*P2PEscrowOrderCreatedIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var sellerRule []interface{}
	for _, sellerItem := range seller {
		sellerRule = append(sellerRule, sellerItem)
	}

	logs, sub, err := _P2PEscrow.contract.FilterLogs(opts, "OrderCreated", orderIdRule, sellerRule)
	if err != nil {
		return nil, err
	}
	return &P2PEscrowOrderCreatedIterator{contract: _P2PEscrow.contract, event: "OrderCreated", logs: logs, sub: sub}, nil
}

// WatchOrderCreated is a free log subscription operation binding the contract event 0xf9e17940e547201709f9aaacdec8ab3566601c60d8f5affb54f74b79b0b50b13.
//
// Solidity: event OrderCreated(uint256 indexed orderId, address indexed seller, address token, uint256 amount)
func (_P2PEscrow *P2PEscrowFilterer) WatchOrderCreated(opts *bind.WatchOpts, sink chan<- *P2PEscrowOrderCreated, orderId []*big.Int, seller []common.Address) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var sellerRule []interface{}
	for _, sellerItem := range seller {
		sellerRule = append(sellerRule, sellerItem)
	}

	logs, sub, err := _P2PEscrow.contract.WatchLogs(opts, "OrderCreated", orderIdRule, sellerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(P2PEscrowOrderCreated)
				if err := _P2PEscrow.contract.UnpackLog(event, "OrderCreated", log); err != nil {
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

// ParseOrderCreated is a log parse operation binding the contract event 0xf9e17940e547201709f9aaacdec8ab3566601c60d8f5affb54f74b79b0b50b13.
//
// Solidity: event OrderCreated(uint256 indexed orderId, address indexed seller, address token, uint256 amount)
func (_P2PEscrow *P2PEscrowFilterer) ParseOrderCreated(log types.Log) (*P2PEscrowOrderCreated, error) {
	event := new(P2PEscrowOrderCreated)
	if err := _P2PEscrow.contract.UnpackLog(event, "OrderCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// P2PEscrowOrderReleasedIterator is returned from FilterOrderReleased and is used to iterate over the raw logs and unpacked data for OrderReleased events raised by the P2PEscrow contract.
type P2PEscrowOrderReleasedIterator struct {
	Event *P2PEscrowOrderReleased // Event containing the contract specifics and raw log

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
func (it *P2PEscrowOrderReleasedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(P2PEscrowOrderReleased)
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
		it.Event = new(P2PEscrowOrderReleased)
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
func (it *P2PEscrowOrderReleasedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *P2PEscrowOrderReleasedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// P2PEscrowOrderReleased represents a OrderReleased event raised by the P2PEscrow contract.
type P2PEscrowOrderReleased struct {
	OrderId *big.Int
	Buyer   common.Address
	Fee     *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterOrderReleased is a free log retrieval operation binding the contract event 0xd91f42113b172b52ce79098fcccffbdce809144244eae0a9aba15aa700714ee0.
//
// Solidity: event OrderReleased(uint256 indexed orderId, address indexed buyer, uint256 fee)
func (_P2PEscrow *P2PEscrowFilterer) FilterOrderReleased(opts *bind.FilterOpts, orderId []*big.Int, buyer []common.Address) (*P2PEscrowOrderReleasedIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var buyerRule []interface{}
	for _, buyerItem := range buyer {
		buyerRule = append(buyerRule, buyerItem)
	}

	logs, sub, err := _P2PEscrow.contract.FilterLogs(opts, "OrderReleased", orderIdRule, buyerRule)
	if err != nil {
		return nil, err
	}
	return &P2PEscrowOrderReleasedIterator{contract: _P2PEscrow.contract, event: "OrderReleased", logs: logs, sub: sub}, nil
}

// WatchOrderReleased is a free log subscription operation binding the contract event 0xd91f42113b172b52ce79098fcccffbdce809144244eae0a9aba15aa700714ee0.
//
// Solidity: event OrderReleased(uint256 indexed orderId, address indexed buyer, uint256 fee)
func (_P2PEscrow *P2PEscrowFilterer) WatchOrderReleased(opts *bind.WatchOpts, sink chan<- *P2PEscrowOrderReleased, orderId []*big.Int, buyer []common.Address) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var buyerRule []interface{}
	for _, buyerItem := range buyer {
		buyerRule = append(buyerRule, buyerItem)
	}

	logs, sub, err := _P2PEscrow.contract.WatchLogs(opts, "OrderReleased", orderIdRule, buyerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(P2PEscrowOrderReleased)
				if err := _P2PEscrow.contract.UnpackLog(event, "OrderReleased", log); err != nil {
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

// ParseOrderReleased is a log parse operation binding the contract event 0xd91f42113b172b52ce79098fcccffbdce809144244eae0a9aba15aa700714ee0.
//
// Solidity: event OrderReleased(uint256 indexed orderId, address indexed buyer, uint256 fee)
func (_P2PEscrow *P2PEscrowFilterer) ParseOrderReleased(log types.Log) (*P2PEscrowOrderReleased, error) {
	event := new(P2PEscrowOrderReleased)
	if err := _P2PEscrow.contract.UnpackLog(event, "OrderReleased", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// P2PEscrowOrderTakenIterator is returned from FilterOrderTaken and is used to iterate over the raw logs and unpacked data for OrderTaken events raised by the P2PEscrow contract.
type P2PEscrowOrderTakenIterator struct {
	Event *P2PEscrowOrderTaken // Event containing the contract specifics and raw log

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
func (it *P2PEscrowOrderTakenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(P2PEscrowOrderTaken)
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
		it.Event = new(P2PEscrowOrderTaken)
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
func (it *P2PEscrowOrderTakenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *P2PEscrowOrderTakenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// P2PEscrowOrderTaken represents a OrderTaken event raised by the P2PEscrow contract.
type P2PEscrowOrderTaken struct {
	OrderId *big.Int
	Buyer   common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterOrderTaken is a free log retrieval operation binding the contract event 0x9db2daab7847c7ac2c20246379b31e171fa4cc75c7b66394f27bf65d9f13a7bc.
//
// Solidity: event OrderTaken(uint256 indexed orderId, address indexed buyer)
func (_P2PEscrow *P2PEscrowFilterer) FilterOrderTaken(opts *bind.FilterOpts, orderId []*big.Int, buyer []common.Address) (*P2PEscrowOrderTakenIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var buyerRule []interface{}
	for _, buyerItem := range buyer {
		buyerRule = append(buyerRule, buyerItem)
	}

	logs, sub, err := _P2PEscrow.contract.FilterLogs(opts, "OrderTaken", orderIdRule, buyerRule)
	if err != nil {
		return nil, err
	}
	return &P2PEscrowOrderTakenIterator{contract: _P2PEscrow.contract, event: "OrderTaken", logs: logs, sub: sub}, nil
}

// WatchOrderTaken is a free log subscription operation binding the contract event 0x9db2daab7847c7ac2c20246379b31e171fa4cc75c7b66394f27bf65d9f13a7bc.
//
// Solidity: event OrderTaken(uint256 indexed orderId, address indexed buyer)
func (_P2PEscrow *P2PEscrowFilterer) WatchOrderTaken(opts *bind.WatchOpts, sink chan<- *P2PEscrowOrderTaken, orderId []*big.Int, buyer []common.Address) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var buyerRule []interface{}
	for _, buyerItem := range buyer {
		buyerRule = append(buyerRule, buyerItem)
	}

	logs, sub, err := _P2PEscrow.contract.WatchLogs(opts, "OrderTaken", orderIdRule, buyerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(P2PEscrowOrderTaken)
				if err := _P2PEscrow.contract.UnpackLog(event, "OrderTaken", log); err != nil {
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

// ParseOrderTaken is a log parse operation binding the contract event 0x9db2daab7847c7ac2c20246379b31e171fa4cc75c7b66394f27bf65d9f13a7bc.
//
// Solidity: event OrderTaken(uint256 indexed orderId, address indexed buyer)
func (_P2PEscrow *P2PEscrowFilterer) ParseOrderTaken(log types.Log) (*P2PEscrowOrderTaken, error) {
	event := new(P2PEscrowOrderTaken)
	if err := _P2PEscrow.contract.UnpackLog(event, "OrderTaken", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// P2PEscrowOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the P2PEscrow contract.
type P2PEscrowOwnershipTransferredIterator struct {
	Event *P2PEscrowOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *P2PEscrowOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(P2PEscrowOwnershipTransferred)
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
		it.Event = new(P2PEscrowOwnershipTransferred)
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
func (it *P2PEscrowOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *P2PEscrowOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// P2PEscrowOwnershipTransferred represents a OwnershipTransferred event raised by the P2PEscrow contract.
type P2PEscrowOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_P2PEscrow *P2PEscrowFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*P2PEscrowOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _P2PEscrow.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &P2PEscrowOwnershipTransferredIterator{contract: _P2PEscrow.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_P2PEscrow *P2PEscrowFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *P2PEscrowOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _P2PEscrow.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(P2PEscrowOwnershipTransferred)
				if err := _P2PEscrow.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_P2PEscrow *P2PEscrowFilterer) ParseOwnershipTransferred(log types.Log) (*P2PEscrowOwnershipTransferred, error) {
	event := new(P2PEscrowOwnershipTransferred)
	if err := _P2PEscrow.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// P2PEscrowPaymentMarkedIterator is returned from FilterPaymentMarked and is used to iterate over the raw logs and unpacked data for PaymentMarked events raised by the P2PEscrow contract.
type P2PEscrowPaymentMarkedIterator struct {
	Event *P2PEscrowPaymentMarked // Event containing the contract specifics and raw log

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
func (it *P2PEscrowPaymentMarkedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(P2PEscrowPaymentMarked)
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
		it.Event = new(P2PEscrowPaymentMarked)
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
func (it *P2PEscrowPaymentMarkedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *P2PEscrowPaymentMarkedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// P2PEscrowPaymentMarked represents a PaymentMarked event raised by the P2PEscrow contract.
type P2PEscrowPaymentMarked struct {
	OrderId *big.Int
	Buyer   common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterPaymentMarked is a free log retrieval operation binding the contract event 0x88529150c26608da00220e4736e92a7c1560e5e62480173a5bc911205ee46c84.
//
// Solidity: event PaymentMarked(uint256 indexed orderId, address indexed buyer)
func (_P2PEscrow *P2PEscrowFilterer) FilterPaymentMarked(opts *bind.FilterOpts, orderId []*big.Int, buyer []common.Address) (*P2PEscrowPaymentMarkedIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var buyerRule []interface{}
	for _, buyerItem := range buyer {
		buyerRule = append(buyerRule, buyerItem)
	}

	logs, sub, err := _P2PEscrow.contract.FilterLogs(opts, "PaymentMarked", orderIdRule, buyerRule)
	if err != nil {
		return nil, err
	}
	return &P2PEscrowPaymentMarkedIterator{contract: _P2PEscrow.contract, event: "PaymentMarked", logs: logs, sub: sub}, nil
}

// WatchPaymentMarked is a free log subscription operation binding the contract event 0x88529150c26608da00220e4736e92a7c1560e5e62480173a5bc911205ee46c84.
//
// Solidity: event PaymentMarked(uint256 indexed orderId, address indexed buyer)
func (_P2PEscrow *P2PEscrowFilterer) WatchPaymentMarked(opts *bind.WatchOpts, sink chan<- *P2PEscrowPaymentMarked, orderId []*big.Int, buyer []common.Address) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var buyerRule []interface{}
	for _, buyerItem := range buyer {
		buyerRule = append(buyerRule, buyerItem)
	}

	logs, sub, err := _P2PEscrow.contract.WatchLogs(opts, "PaymentMarked", orderIdRule, buyerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(P2PEscrowPaymentMarked)
				if err := _P2PEscrow.contract.UnpackLog(event, "PaymentMarked", log); err != nil {
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

// ParsePaymentMarked is a log parse operation binding the contract event 0x88529150c26608da00220e4736e92a7c1560e5e62480173a5bc911205ee46c84.
//
// Solidity: event PaymentMarked(uint256 indexed orderId, address indexed buyer)
func (_P2PEscrow *P2PEscrowFilterer) ParsePaymentMarked(log types.Log) (*P2PEscrowPaymentMarked, error) {
	event := new(P2PEscrowPaymentMarked)
	if err := _P2PEscrow.contract.UnpackLog(event, "PaymentMarked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
