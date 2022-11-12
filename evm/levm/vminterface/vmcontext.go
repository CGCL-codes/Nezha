package vminterface

import (
	"Nezha/ethereum/go-ethereum/common"
	"Nezha/ethereum/go-ethereum/core"
	"Nezha/ethereum/go-ethereum/core/vm"
	"crypto/sha256"
	"math/big"
	"time"
)

// NewVMContext will construct a new EVM Context with default values.
// TODO: include gas price variable in params
func NewVMContext(origin common.Address, coinbase common.Address, blockNum *big.Int, chainContext ChainContext) vm.Context {
	return vm.Context{
		CanTransfer: CanTransfer,
		Transfer:    Transfer,
		GetHash:     core.GetHashFn(chainContext.GetHeader(sha256.Sum256([]byte("")), 0), chainContext),
		Origin:      origin,
		GasPrice:    big.NewInt(1),
		Coinbase:    coinbase,
		GasLimit:    uint64(1000000),
		BlockNumber: blockNum,
		Time:        big.NewInt(time.Now().Unix()),
		Difficulty:  big.NewInt(1),
	}
}

// CanTransfer checks whether there are enough funds in the address' account to make a transfer.
// This does not take the necessary gas in to account to make the transfer valid.
func CanTransfer(db vm.StateDB, addr common.Address, amount *big.Int) bool {
	return db.GetBalance(addr).Cmp(amount) >= 0
}

// Transfer subtracts amount from sender and adds amount to recipient using the given Db
func Transfer(db vm.StateDB, sender, recipient common.Address, amount *big.Int) {
	db.SubBalance(sender, amount)
	db.AddBalance(recipient, amount)
}
