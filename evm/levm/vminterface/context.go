package vminterface

import (
	"Nezha/ethereum/go-ethereum/common"
	"Nezha/ethereum/go-ethereum/consensus"
	"Nezha/ethereum/go-ethereum/core/types"
	"crypto/sha256"
	"math/big"
	"time"
)

// ChainContext needed by the evm
type ChainContext struct {
	coinbase common.Address
}

// NewChainContext constructs a new context needed by EVMContext
func NewChainContext(coinbase common.Address) ChainContext {
	return ChainContext{coinbase}
}

// GetHeader returns a header needed by the EVMContext, it also satisfies
// the ChainContext Interface.
// TODO: increase gaslimit ?
func (cc ChainContext) GetHeader(hash common.Hash, number uint64) *types.Header {
	return &types.Header{
		Coinbase:   cc.coinbase,
		Difficulty: big.NewInt(1),
		Number:     big.NewInt(0),
		GasLimit:   uint64(1000000),
		GasUsed:    uint64(0),
		Time:       uint64(time.Now().Unix()),
		Extra:      nil,
	}
}

// Engine is only here to satisfy the chaincontext interface
func (cc ChainContext) Engine() consensus.Engine {
	return nil
}

// NewHeader calls GetHeader() on the passed chain context
// with arbitrary arguments
func NewHeader(cc ChainContext) *types.Header {
	return cc.GetHeader(sha256.Sum256([]byte("test")), 0)
}
