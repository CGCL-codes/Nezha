package utils

import (
	"Nezha/core"
	"bytes"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
)

// StoreState store state to db
func StoreState(db *leveldb.DB, acc *core.StateAccount) error {
	address := acc.Address
	key := bytes.Join([][]byte{[]byte("a"), address},[]byte{})
	value := acc.Serialize()
	// log.Println("Store state key: ", key)
	if err := db.Put(key ,value,nil); err != nil {
		return err
	}
	return nil
}

// RemoveState remove account state from db
func RemoveState(db *leveldb.DB, addr []byte) error {
	key := bytes.Join([][]byte{[]byte("a"), addr},[]byte{})
	if err := db.Delete(key,nil); err != nil {
		return err
	}
	return nil
}

// FetchState get state via account address from db
func FetchState(db *leveldb.DB, addr []byte) (*core.StateAccount, error) {
	key := bytes.Join([][]byte{[]byte("a"), addr},[]byte{})
	log.Println("Fetch state key: ", key)
	data, err := db.Get(key,nil)
	if err != nil {
		return nil, err
	}
	acc := core.DeserializeAcc(data)
	return acc, nil
}

// open db instance or create a db if not exist
func LoadDB(blkFile string) (*leveldb.DB, error) {
	// log.Println("LoadDB() function is called")
	db, err := leveldb.OpenFile(blkFile,nil)
	if err != nil {
		return nil, err
	}
	return db, nil
}
