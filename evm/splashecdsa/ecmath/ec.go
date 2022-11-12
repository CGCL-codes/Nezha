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

package ecmath

import (
	"crypto/elliptic"
	"crypto/sha256"
	"math/big"
)

// This file contains some useful equations to extend golangs ECDSA library

// GetY calculates the Y coordinate of a given X coordinate
func GetY(x *big.Int, curve *elliptic.CurveParams) (*big.Int, *big.Int) {
	x3 := new(big.Int).Mul(x, x)
	x3.Mul(x3, x)

	threeX := new(big.Int).Lsh(x, 1)
	threeX.Add(threeX, x)

	x3.Sub(x3, threeX)
	x3.Add(x3, curve.B)
	x3.Mod(x3, curve.P)

	y1 := new(big.Int).ModSqrt(x3, curve.P)

	y2 := new(big.Int).Sub(curve.P, y1)
	y2.Mod(y2, curve.P)

	//return x3.ModSqrt(x3, curve.P), x3.Sub(curve.P, x3).Mod(x3, curve.P)
	return y1, y2
}

// InversePoint Returns the inverse of a point on a given curve.
func InversePoint(X, Y *big.Int, curve *elliptic.CurveParams) (*big.Int, *big.Int) {
	y2 := new(big.Int).Sub(curve.P, Y)
	y2.Mod(y2, curve.P)
	return X, y2
}

func numDr(S *big.Int, R *big.Int, curve *elliptic.CurveParams) *big.Int {
	RMod := new(big.Int).ModInverse(R, curve.N)
	NumDR := new(big.Int).Mul(S, RMod)
	return NumDR.Mod(NumDR, curve.N)
}

func getZ(msg []byte) *big.Int {
	hash := sha256.Sum256(msg)
	return new(big.Int).SetBytes(hash[:32])
}

func CheckByteEq(a, b []byte) bool {

	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
