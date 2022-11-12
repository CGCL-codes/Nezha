package levm

import (
	cc "Nezha/core"
	"Nezha/ethereum/go-ethereum/accounts/abi"
	"Nezha/ethereum/go-ethereum/common"
	"Nezha/ethereum/go-ethereum/core"
	"Nezha/ethereum/go-ethereum/core/state"
	"Nezha/ethereum/go-ethereum/core/vm"
	"Nezha/ethereum/go-ethereum/ethdb"
	"Nezha/ethereum/go-ethereum/log"
	"Nezha/ethereum/go-ethereum/params"
	vmi "Nezha/evm/levm/vminterface"
	"fmt"
	"math/big"
)

// LEVM is a container for the go-ethereum EVM
// with methods to create and call contracts.
//
// LEVM contains the two most important objects
// for interacting with the EVM: stateDB and
// vm.EVM. The LEVM should be created with the
// LEVM.New() method, unless you know what you
// doing.
type LEVM struct {
	stateDB      *state.StateDB
	evm          *vm.EVM
	edb          ethdb.Database
	structLogger *vm.StructLogger
}

// New creates a new instace of the LEVM
func New(dbPath string, blockNumber *big.Int, origin common.Address) *LEVM {
	// create blank LEVM instance:
	lvm := LEVM{}

	// setup storage using dbpath
	lvm.stateDB, lvm.edb = vmi.NewStateDB(common.Hash{}, dbPath)

	// update the evm - creates new EVM
	lvm.NewEVM(blockNumber, origin)

	return &lvm
}

// NewEVM creates a fresh evm instance with
// new origin and blocknumber and time.
// This method recreates the contained EVM while
// keeping the stateDB the same.
func (lvm *LEVM) NewEVM(blockNumber *big.Int, origin common.Address) {

	// create contexted for the evm context
	chainContext := vmi.NewChainContext(origin)
	vmContext := vmi.NewVMContext(origin, origin, blockNumber, chainContext)

	// create vm config
	logConfig := vm.LogConfig{}
	structLogger := vm.NewStructLogger(&logConfig)
	vmConfig := vm.Config{Debug: true, Tracer: structLogger /*JumpTable: vm.NewByzantiumInstructionSet()*/}

	// create the evm
	lvm.evm = vm.NewEVM(vmContext, lvm.stateDB, params.MainnetChainConfig, vmConfig)
	lvm.structLogger = structLogger
}

// DeployContract will create and deploy a new
// contract from the contract data.
func (lvm *LEVM) DeployContract(fromAddr common.Address, contractData []byte) ([]byte, common.Address, uint64, error) {

	// Get reference to the transaction sender
	contractRef := vm.AccountRef(fromAddr)
	leftOver := big.NewInt(0)

	return lvm.evm.Create(
		contractRef,
		contractData,
		lvm.stateDB.GetBalance(fromAddr).Uint64(),
		leftOver,
	)
}

// CallContract - make a call to a Contract Method
// using prepacked Inputs. To use ABI directly try
// lvm.CallContractABI()
func (lvm *LEVM) CallContract(callerAddr, contractAddr common.Address, value *big.Int, inputs []byte) ([]byte, error) {
	// Get reference to the transaction sender
	callerRef := vm.AccountRef(callerAddr)
	output, gas, err := lvm.evm.Call(
		callerRef,
		contractAddr,
		inputs,
		lvm.stateDB.GetBalance(callerAddr).Uint64(),
		value,
	)
	lvm.stateDB.SetBalance(callerAddr, big.NewInt(0).SetUint64(gas))
	return output, err
}

// CallContractABI - make a call to a Contract Method
// using the ABI.
func (lvm *LEVM) CallContractABI(callerAddr, contractAddr common.Address, value *big.Int, abiObject abi.ABI, funcName string, args ...interface{}) ([]byte, error) {

	inputs, err := abiObject.Pack(funcName, args...)
	if err != nil {
		return nil, err
	}

	callerRef := vm.AccountRef(callerAddr)
	output, gas, err := lvm.evm.Call(
		callerRef,
		contractAddr,
		inputs,
		lvm.stateDB.GetBalance(callerAddr).Uint64(),
		value,
	)
	lvm.stateDB.SetBalance(callerAddr, big.NewInt(0).SetUint64(gas))
	return output, err
}

// CallContractABI2 - make a call to a Contract Method
// using the ABI, return the output and the read/write set
func (lvm *LEVM) CallContractABI2(callerAddr, contractAddr common.Address, value *big.Int, abiObject abi.ABI, funcName string, args ...interface{}) (vm.Storage, vm.Storage, []byte, error) {

	inputs, err := abiObject.Pack(funcName, args...)
	if err != nil {
		return nil, nil, nil, err
	}

	callerRef := vm.AccountRef(callerAddr)
	output, gas, err := lvm.evm.Call(
		callerRef,
		contractAddr,
		inputs,
		lvm.stateDB.GetBalance(callerAddr).Uint64(),
		value,
	)
	lvm.stateDB.SetBalance(callerAddr, big.NewInt(0).SetUint64(gas))

	rMap, wMap := lvm.structLogger.RWSetCapture(contractAddr)

	return rMap, wMap, output, err
}

// ReplayTransaction replay the real transaction dataset to obtain the read/write set
func (lvm *LEVM) ReplayTransaction(tx cc.EthTransaction, gp *core.GasPool) (vm.Storage, vm.Storage, []byte, error) {
	msg := tx.AsMessage()

	//st := vmi.NewStateTrans(msg, new(core.GasPool).AddGas(uint64(1000000)))
	st := vmi.NewStateTrans(msg, gp)
	sender := vm.AccountRef(msg.From())
	homestead := lvm.evm.ChainConfig().IsHomestead(lvm.evm.BlockNumber)
	istanbul := lvm.evm.ChainConfig().IsIstanbul(lvm.evm.BlockNumber)
	contractCreation := msg.To() == nil

	// Pay intrinsic gas
	gas, err := core.IntrinsicGas(st.Data, contractCreation, homestead, istanbul)
	if err != nil {
		return nil, nil, nil, err
	}
	if err = st.UseGas(gas); err != nil {
		return nil, nil, nil, err
	}

	var (
		output []byte
		vmerr  error
		rMap   vm.Storage
		wMap   vm.Storage
		addr   common.Address
	)

	if contractCreation {
		output, addr, st.Gas, vmerr = lvm.evm.Create(sender, st.Data, st.Gas, st.Value)
		fmt.Println(addr.Bytes())
	} else {
		// increment the nonce for the next transaction
		lvm.stateDB.SetNonce(msg.From(), lvm.stateDB.GetNonce(sender.Address())+1)
		output, st.Gas, vmerr = lvm.evm.Call(sender, *msg.To(), st.Data, st.Gas, st.Value)
	}

	if vmerr != nil {
		log.Debug("VM returned with error", "err", vmerr)
		if vmerr == vm.ErrInsufficientBalance {
			return nil, nil, nil, vmerr
		}
	}

	lvm.RefundGas(st)
	lvm.stateDB.AddBalance(lvm.evm.Coinbase, new(big.Int).Mul(new(big.Int).SetUint64(st.GasUsed()), st.GasPrice))

	// if tx is calling contract, capture its read/write set
	if !contractCreation {
		rMap, wMap = lvm.structLogger.RWSetCapture(*msg.To())
		return rMap, wMap, output, err
	}

	// TODO: Capture all the internal state (incurred by internal transactions)

	return nil, nil, output, err
}

// GetStateDB - obtain the current stateDB snapshot
func (lvm *LEVM) GetStateDB() *state.StateDB { return lvm.stateDB }

// RefundGas refund the remaining gas to the sender and gas pool
func (lvm *LEVM) RefundGas(st *vmi.StateTrans) {
	refund := st.GasUsed() / 2
	if refund > lvm.stateDB.GetRefund() {
		refund = lvm.stateDB.GetRefund()
	}
	st.Gas += refund

	remaining := new(big.Int).Mul(new(big.Int).SetUint64(st.Gas), st.GasPrice)
	lvm.stateDB.AddBalance(st.Msg.From(), remaining)

	st.Gp.AddGas(st.Gas)
}
