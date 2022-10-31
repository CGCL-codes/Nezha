package core

import (
	"Nezha/ethereum/go-ethereum/common"
	"Nezha/ethereum/go-ethereum/core/types"
	"Nezha/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
	"math/big"
	"sync/atomic"
)

type EthTransaction struct {
	data txdata
	// caches
	hash atomic.Value
	size atomic.Value
	from atomic.Value
}

type txdata struct {
	AccountNonce uint64          `json:"nonce"    gencodec:"required"`
	Price        *big.Int        `json:"gasPrice" gencodec:"required"`
	GasLimit     uint64          `json:"gas"      gencodec:"required"`
	From         *common.Address `json:"from"     gencodec:"required"`
	Recipient    *common.Address `json:"to"       rlp:"nil"` // nil means contract creation
	Amount       *big.Int        `json:"value"    gencodec:"required"`
	Payload      []byte          `json:"input"    gencodec:"required"`

	// Signature values
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`
}

func NewEthTransaction(nonce uint64, from, to *common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *EthTransaction {
	return newEthTransaction(nonce, from, to, amount, gasLimit, gasPrice, data)
}

func NewContractCreation(nonce uint64, from *common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *EthTransaction {
	return newEthTransaction(nonce, from, nil, amount, gasLimit, gasPrice, data)
}

func newEthTransaction(nonce uint64, from, to *common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *EthTransaction {
	if len(data) > 0 {
		data = common.CopyBytes(data)
	}
	d := txdata{
		AccountNonce: nonce,
		From:         from,
		Recipient:    to,
		Payload:      data,
		Amount:       new(big.Int),
		GasLimit:     gasLimit,
		Price:        new(big.Int),
		V:            new(big.Int),
		R:            new(big.Int),
		S:            new(big.Int),
	}
	if amount != nil {
		d.Amount.Set(amount)
	}
	if gasPrice != nil {
		d.Price.Set(gasPrice)
	}

	return &EthTransaction{data: d}
}

func (tx *EthTransaction) Data() []byte         { return common.CopyBytes(tx.data.Payload) }
func (tx *EthTransaction) Gas() uint64          { return tx.data.GasLimit }
func (tx *EthTransaction) GasPrice() *big.Int   { return new(big.Int).Set(tx.data.Price) }
func (tx *EthTransaction) Value() *big.Int      { return new(big.Int).Set(tx.data.Amount) }
func (tx *EthTransaction) Nonce() uint64        { return tx.data.AccountNonce }
func (tx *EthTransaction) CheckNonce() bool     { return true }
func (tx *EthTransaction) From() common.Address { return *tx.data.From }

// To returns the recipient address of the transaction.
// It returns nil if the transaction is a contract creation.
func (tx *EthTransaction) To() *common.Address {
	if tx.data.Recipient == nil {
		return nil
	}
	to := *tx.data.Recipient
	return &to
}

// Hash hashes the RLP encoding of tx.
// It uniquely identifies the transaction.
func (tx *EthTransaction) Hash() common.Hash {
	if hash := tx.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}

	var v common.Hash
	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, tx)
	hw.Sum(v[:0])
	tx.hash.Store(v)
	return v
}

// Size returns the true RLP encoded storage size of the transaction, either by
// encoding and returning it, or returning a previsouly cached value.
func (tx *EthTransaction) Size() common.StorageSize {
	if size := tx.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	c := writeCounter(0)
	rlp.Encode(&c, &tx.data)
	tx.size.Store(common.StorageSize(c))
	return common.StorageSize(c)
}

type writeCounter common.StorageSize

func (c *writeCounter) Write(b []byte) (int, error) {
	*c += writeCounter(len(b))
	return len(b), nil
}

// AsMessage returns the transaction as a core.Message.
//
// AsMessage requires a signer to derive the sender.
//
// XXX Rename message to something less arbitrary?
func (tx *EthTransaction) AsMessage() types.Message {
	msg := types.NewMessage(*tx.data.From, tx.data.Recipient, tx.data.AccountNonce, tx.data.Amount,
		tx.data.GasLimit, new(big.Int).Set(tx.data.Price), tx.data.Payload, true)

	return msg
}
