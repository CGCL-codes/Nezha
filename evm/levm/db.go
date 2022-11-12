// Package levm is a higher level wrapper for the EVM and the related
// StateDB.
// levm.go contains methods for creating the levm and deploying/calling
// contracts.
// db.go contains methods for interacting with the stateDB i.e. getting
// accounts and balances
package levm

import (
	"Nezha/ethereum/go-ethereum/common"
	"Nezha/ethereum/go-ethereum/core/state"
	"math/big"
)

// Account is a simple representation of an
// account in the stateDB, for further
// controls call DB() to retrieve the db
// reference itself.
// Almost as a snapshot
type Account struct {
	Balance *big.Int
	Nonce   uint64
	Code    []byte
}

// DB returns a pointer to the vm's state DB.
func (lvm *LEVM) DB() *state.StateDB {
	return lvm.stateDB
}

// GetAccount returns a copy of stateDB
func (lvm *LEVM) GetAccount(addr common.Address) Account {
	acc := Account{}
	lvm.stateDB.GetOrNewStateObject(addr)
	acc.Balance = lvm.stateDB.GetBalance(addr)
	acc.Nonce = lvm.stateDB.GetNonce(addr)
	acc.Code = lvm.stateDB.GetCode(addr)
	return acc
}

// SetAccount will update the statedb with
// related values contained in the Account
// snapshot.
func (lvm *LEVM) SetAccount(addr common.Address, acc Account) {
	lvm.stateDB.GetOrNewStateObject(addr)
	lvm.stateDB.SetBalance(addr, acc.Balance)
	lvm.stateDB.SetNonce(addr, acc.Nonce)
	lvm.stateDB.SetCode(addr, acc.Code)
}

// NewAccount Create a new Account and Set its balance
func (lvm *LEVM) NewAccount(addr common.Address, balance *big.Int) {
	lvm.stateDB.GetOrNewStateObject(addr)
	lvm.stateDB.SetBalance(addr, balance)
}
