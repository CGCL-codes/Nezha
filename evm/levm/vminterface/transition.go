package vminterface

import (
	"Nezha/ethereum/go-ethereum/core"
	"Nezha/ethereum/go-ethereum/core/types"
	"Nezha/ethereum/go-ethereum/core/vm"
	"math/big"
)

type StateTrans struct {
	Gp         *core.GasPool
	Msg        types.Message
	Gas        uint64
	GasPrice   *big.Int
	InitialGas uint64
	Value      *big.Int
	Data       []byte
}

func NewStateTrans(msg types.Message, gp *core.GasPool) *StateTrans {
	return &StateTrans{
		Gp:         gp,
		Msg:        msg,
		Gas:        msg.Gas(),
		GasPrice:   msg.GasPrice(),
		InitialGas: msg.Gas(),
		Value:      msg.Value(),
		Data:       msg.Data(),
	}
}

func (st *StateTrans) UseGas(amount uint64) error {
	if st.Gas < amount {
		return vm.ErrOutOfGas
	}
	st.Gas -= amount

	return nil
}

// GasUsed returns the amount of gas used up by the state transition.
func (st *StateTrans) GasUsed() uint64 {
	return st.InitialGas - st.Gas
}
