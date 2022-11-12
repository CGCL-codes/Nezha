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

package splashecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"

	"Nezha/evm/splashecdsa/ecmath"
)

// PublicKey is a wrapper for ecdsa.PublicKey which adds
// some useful methods like ToBytes(), FromBytes() and QuickCheck()
type PublicKey ecdsa.PublicKey

// Bytes concatonates the Public keys x and y values
// and returns the resulting byte array. For a smaller
// byte array see @CompressedBytes.
func (pub *PublicKey) Bytes() (buf []byte) {
	x := pub.X.Bytes()
	y := pub.Y.Bytes()
	buf = append(x, y...)
	return
}

// CompressedBytes returns a compress version of the
// public key that can be reconstructed using
// @SetCompressedBytes.
//
// CompressedBytes are typically of length (n/2)+1 where n
// is the length of normal Bytes(). It is more expensive to
// Generate compress/reconstruct from CompressedBytes.
func (pub *PublicKey) CompressedBytes() (buf []byte) {
	// get the two possible y values
	_, y1 := ecmath.GetY(pub.X, pub.Curve.Params())

	// set v flag
	v := byte(0x0)
	if y1.Cmp(pub.Y) == 0 {
		v = byte(0x1)
	}

	//get X bytes
	x := pub.X.Bytes()

	// append flag:
	buf = append([]byte{v}, x...)
	return
}

// SetCompressedBytes decodes and decompressed the buf then
// stores the values in the pub X and Y
func (pub *PublicKey) SetCompressedBytes(buf []byte) *PublicKey {
	v := buf[0]
	bigX := new(big.Int).SetBytes(buf[1:])
	bigY, y1 := ecmath.GetY(bigX, elliptic.P256().Params())
	if v == byte(0x1) {
		bigY = y1
	}
	pub.X = bigX
	pub.Y = bigY
	pub.Curve = elliptic.P256()
	return pub
}

// SetBytes decodes the buf and stores the values in the
// pub X and Y
func (pub *PublicKey) SetBytes(buf []byte) *PublicKey {
	bigX := new(big.Int)
	bigY := new(big.Int)
	bigX.SetBytes(buf[:32])
	bigY.SetBytes(buf[32:])

	pub.X = bigX
	pub.Y = bigY
	pub.Curve = elliptic.P256()
	return pub
}

// QuickCheck quickly checks that the public key is
// in accordance with splashs ecdsa curve
func (pub *PublicKey) QuickCheck(curve elliptic.Curve) bool {
	if pub.Curve != curve {
		return false
	}
	if !curve.IsOnCurve(pub.X, pub.Y) {
		return false
	}
	return true
}

// Verify verifies a SplashSignature of the hash belongs to this
// PublicKey
func (pub *PublicKey) Verify(hash []byte, sig SplashSignature) bool {
	ecPub := ecdsa.PublicKey(*pub)
	return ecdsa.Verify(&ecPub, hash, sig.R, sig.S)
}

// Matches will check if public key matches an address
func (pub *PublicKey) Matches(addr []byte) bool {
	if IsMultiSigAddress(addr) {
		return false
	}
	compression := IsAddressCompressed(addr)
	return ecmath.CheckByteEq(pub.GetAddress(compression), addr)
}
