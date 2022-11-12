// MIT License

// Copyright (c) 2019 Kassius Barker <kasscrypto@gmail.com>

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package splashecdsa is a wrapper for the built in ecdsa, for the Splash ledger that
// adds:
//
// - Public key reconstruction from the signature
//
// - Mild support for multi signature applications
//
// - Generic address generation
//
// This lib depends on github.com/CryptoKass/splashmerkle.
//
// Example:
// 	C := elliptic.P256()
//
// 	key, _ := splashecdsa.GenerateKeys(C)
// 	fmt.Printf("key X: %#x\nkey Y: %#x\n", key.X, key.Y)
//
// 	hashedMessage := sha256.Sum256([]byte("hello"))
// 	sig, _ := key.Sign(hashedMessage[:])
//
// 	pub := sig.ReconstructPublicKey(hashedMessage[:], C)
// 	fmt.Printf("key X: %#x\nkey Y: %#x\n", pub.X, pub.Y)
//
// 	fmt.Println("reconstruction success:", pub.X.Cmp(key.X) == 0)
package splashecdsa

import (
	"crypto/sha256"
)

func (pub *PublicKey) GetAddress(compressed bool) []byte {
	buf := pub.Bytes()
	v := byte(0x0) // compression flag
	z := byte(0x0) // mutlisig flag
	if compressed {
		buf = pub.CompressedBytes()
		v = byte(0x1)
	}
	addrRaw := sha256.Sum256(buf)
	return append([]byte{v, z}, addrRaw[:20]...)
}

func (priv *PrivateKey) GetAddress(compressed bool) []byte {
	pub := priv.GetPublicKey()
	return pub.GetAddress(compressed)
}

func IsAddressCompressed(addr []byte) bool {
	v := addr[0]
	if v == 0x0 {
		return false
	}
	return true
}

func IsMultiSigAddress(addr []byte) bool {
	z := addr[1]
	if z == 0x0 {
		return false
	}
	return true
}

// IsAddressValid is a quick check to ensure
// address is of a reasonable length.
func IsAddressValid(addr []byte) bool {
	l := len(addr)
	return (l >= 20 && l <= 64)
}
