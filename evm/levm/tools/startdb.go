package tools

import (
	"Nezha/ethereum/go-ethereum/common"
	"Nezha/ethereum/go-ethereum/ethdb"
	"Nezha/ethereum/go-ethereum/trie"
	com "Nezha/evm/levm/common"
)

func StartTrieDB(edb ethdb.Database) *trie.Trie {
	tdb := trie.NewDatabase(edb)
	tr, err := trie.New(common.Hash{}, tdb)
	com.PanicErr(err)
	return tr
}
