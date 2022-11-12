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
	"crypto/rand"
	"math/big"
)

// GenerateKeys is a wrapper for ecdsa.GenerateKey that returns
// splashkeys.PrivateKey instead ok *ecdsa.PrivateKey, this method will
// return an error key generation failed.
func GenerateKeys(curve elliptic.Curve) (PrivateKey, error) {
	k, err := ecdsa.GenerateKey(curve, rand.Reader)
	return PrivateKey(*k), err
}

// PrivateKey is a wrapper for ecdsa.PrivateKey which adds
// Some useful methogs like ToBytes(), FromBytes(), GetPublicKey()
type PrivateKey ecdsa.PrivateKey

// Bytes returns the Private Key's D values
// and returns the byte array
func (priv *PrivateKey) Bytes() (d []byte) {
	d = priv.D.Bytes()
	return
}

// SetBytes reconstructs the private key from D bytes
func (priv *PrivateKey) SetBytes(d []byte) *PrivateKey {
	bigD := new(big.Int)
	bigD.SetBytes(d)
	priv.D = bigD
	priv.Curve = elliptic.P256()
	if priv.PublicKey.X == nil {
		priv.PublicKey.Curve = elliptic.P256()
		priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(priv.D.Bytes())
	}
	return priv
}

// GetPublicKey returns the associated PublicKey for this privatekey,
// If the key is missing then one is generated. This is considered the safe way
// to get the public key
func (priv *PrivateKey) GetPublicKey() PublicKey {
	if priv.PublicKey.X == nil {
		priv.PublicKey.Curve = elliptic.P256()
		priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(priv.D.Bytes())
	}
	return PublicKey(priv.PublicKey)
}

// Sign - a wrapper for `ecdsa.Sign` that will sign some bytes
// and return a reconstructable SplashSignature.
//
// The message should be less than 32 bytes long, for cases where
// the message is longer, hash the message and sign the result.
func (priv *PrivateKey) Sign(msg []byte) (sig SplashSignature, err error) {
	ecPriv := ecdsa.PrivateKey(*priv)
	sig = SplashSignature{}
	sig.R, sig.S, err = ecdsa.Sign(rand.Reader, &ecPriv, msg)

	if sig.ReconstructPublicKey(msg, priv.Curve).X.Cmp(priv.GetPublicKey().X) == 0 {
		return
	}
	sig.V = 1

	return
}
