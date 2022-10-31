package core

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/tv42/base58"
	"log"
	"math/big"
	"reflect"
	"time"
)

// Transaction represents a Ethereum transaction
type Transaction struct {
	Header    TransactionHeader
	Signature []byte
	Payload   []byte
	ID   []byte
}

type TransactionHeader struct {
	From          []byte
	To            []byte
	Timestamp     uint32
	PayloadHash   [32]byte
	PayloadLength uint32
	Nonce         uint32
}


// Serialize returns a serialized Transaction
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// Hash returns the hash of the Transaction
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

// Sign transaction
func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey) {

	dataToSign := fmt.Sprintf("%x\n", *tx)

	r, s, err := ecdsa.Sign(rand.Reader, &privateKey, []byte(dataToSign))
	if err != nil {
		log.Panic(err)
	}
	signature := append(r.Bytes(), s.Bytes()...)

	tx.Signature = signature
}

// String returns a human-readable representation of a transaction
func (tx Transaction) String() string {
	var lines string
	lines = fmt.Sprintf("--- Transaction %x:", tx.ID)
	return lines
}

// Verify transaction
func (tx *Transaction) VerifyTransaction() bool {

	txHash := tx.Hash()
	payloadHash := sha256.Sum256(tx.Payload)

	return reflect.DeepEqual(payloadHash, tx.Header.PayloadHash) && SignatureVerify(tx.Header.From, tx.Signature, txHash)
}

func (tx *Transaction) GenerateNonce(prefix []byte) uint32 {

	newT := tx
	newTHash := newT.Hash()
	for {

		if reflect.DeepEqual(prefix, newTHash[:len(prefix)]) {
			break
		}

		newT.Header.Nonce++
	}

	return newT.Header.Nonce
}

// Verify signatures
func SignatureVerify(publicKey, sig, hash []byte) bool {

	b, _ := base58.DecodeToBig(publicKey)
	pubL := splitBig(b, 2)
	x, y := pubL[0], pubL[1]

	b, _ = base58.DecodeToBig(sig)
	sigL := splitBig(b, 2)
	r, s := sigL[0], sigL[1]

	pub := ecdsa.PublicKey{elliptic.P256(), x, y}

	return ecdsa.Verify(&pub, hash, r, s)
}

func splitBig(b *big.Int, parts int) []*big.Int {

	bs := b.Bytes()
	if len(bs)%2 != 0 {
		bs = append([]byte{0}, bs...)
	}

	l := len(bs) / parts
	as := make([]*big.Int, parts)

	for i, _ := range as {

		as[i] = new(big.Int).SetBytes(bs[i*l : (i+1)*l])
	}

	return as
}

// NewTransaction creates a new transaction
func NewTransaction(from, to, payload []byte) *Transaction {

	tx := Transaction{Header: TransactionHeader{From: from, To: to}, Payload: payload}

	tx.Header.Timestamp = uint32(time.Now().Unix())
	tx.Header.PayloadHash = sha256.Sum256(tx.Payload)
	tx.Header.PayloadLength = uint32(len(tx.Payload))

	return &tx
}

// NewCoinbaseTX creates a new coinbase transaction
func NewCoinbaseTX() *Transaction {
	randData := make([]byte, 20)
	_, err := rand.Read(randData)
	if err != nil {
		log.Panic(err)
	}
	data := fmt.Sprintf("%x", randData)
	tx := Transaction{TransactionHeader{}, nil, []byte(data), nil}
	tx.ID = tx.Hash()

	return &tx
}

// DeserializeTransaction deserializes a transaction
func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	if err != nil {
		log.Panic(err)
	}

	return transaction
}

