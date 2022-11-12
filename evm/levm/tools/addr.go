package tools

import (
	"Nezha/ethereum/go-ethereum/common"
	"Nezha/evm/splashecdsa"
	"crypto/elliptic"
)

func NewRandomAddress() common.Address {
	curve := elliptic.P256()
	priv, _ := splashecdsa.GenerateKeys(curve)
	addrBytes := priv.GetAddress(true)
	return common.BytesToAddress(addrBytes)
}
