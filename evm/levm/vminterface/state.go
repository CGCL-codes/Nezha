package vminterface

import (
	"Nezha/ethereum/go-ethereum/common"
	"Nezha/ethereum/go-ethereum/core/rawdb"
	"Nezha/ethereum/go-ethereum/core/state"
	"Nezha/ethereum/go-ethereum/ethdb"
	com "Nezha/evm/levm/common"
)

// NewStateDB - Create a new StateDB using levelDB instead of RAM
func NewStateDB(root common.Hash, dbPath string) (*state.StateDB, ethdb.Database) {

	// open ethdb
	/*edb, err := ethdb.NewLDBDatabase(dbPath, 100, 100)
	db := state.NewDatabase(edb)
	com.PanicErr(err)
	*/

	edb, _ := rawdb.NewLevelDBDatabase(dbPath, 100, 100, "")
	//edb := rawdb.NewMemoryDatabase()
	db := state.NewDatabase(edb)

	// make statedb
	stateDB, err := state.New(root, db)
	com.PanicErr(err)

	return stateDB, edb
}
