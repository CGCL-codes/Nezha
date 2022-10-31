package core

import (
	"bytes"
	"encoding/gob"
	"log"
)

// StateAccount is the Ethereum consensus representation of accounts.
type StateAccount struct {
	Address []byte
	Value []byte
}

func CreateAccount(addr []byte, value []byte) *StateAccount {
	acc := StateAccount{addr, value}
	return &acc
}


// serialize Account
func (acc *StateAccount) Serialize() []byte {
	var encode bytes.Buffer

	enc := gob.NewEncoder(&encode)
	err := enc.Encode(acc)

	if err != nil {
		log.Panic("Account encode fail:", err)
	}

	return encode.Bytes()
}

// Deserialize Account
func DeserializeAcc(data []byte) *StateAccount {
	var acc StateAccount

	decode := gob.NewDecoder(bytes.NewReader(data))

	err := decode.Decode(&acc)
	if err != nil {
		log.Panic("Account decode fail:", err)
	}

	return &acc
}
