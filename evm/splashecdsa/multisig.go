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
	"crypto/elliptic"

	"Nezha/evm/splashecdsa/ecmath"
	"Nezha/evm/splashmerkle"
)

// GenerateMultiSigKey - Create a new Random multiSigKey
func GenerateMultiSigKey(curve elliptic.Curve, order, partners uint8) (*MultiSigKey, error) {
	priv, err := GenerateKeys(curve)
	multiKey := MultiSigKey{
		PrivateKey: priv,
		Order:      order,
		Partners:   partners,
	}
	return &multiKey, err
}

// MultiSigKey Is a private key for an multi-signature
// ecdsa address. wraps PrivateKey with some extra MultiSignature
// parameters and values.
type MultiSigKey struct {
	PrivateKey PrivateKey
	Order      uint8
	Partners   uint8
}

// Sign - I a wrapper for `ecdsa.Sign` that will sign some bytes
// and return a reconstructable SplashSignature.
//
// The message should be less than 32 bytes long, for cases where
// the message is longer, hash the message and sign the result.
func (multi *MultiSigKey) Sign(data []byte) (SplashSignature, error) {
	sig, err := multi.PrivateKey.Sign(data)
	sig.O = multi.Order
	return sig, err
}

func (multi *MultiSigKey) GetPublicKey() PublicKey {
	return multi.PrivateKey.GetPublicKey()
}

// VerifyMutliSig verify mulitiple signstures to a single multi signature address
func VerifyMutliSig(sigs []SplashSignature, data []byte, addr []byte, C elliptic.Curve) bool {
	partners := make([]PublicKey, len(sigs))

	for _, sig := range sigs {
		pub := sig.ReconstructPublicKey(data, C)
		partners[sig.O] = pub
		if !pub.Verify(data, sig) {
			// QUIRK := tampered data will produce a valid public key that solves the equation
			// however will fail to match address. So this tests is mostly unnecessary, as invalid data
			// will pass, however the problem will be resolves by checking the addresses are
			// matching.
			return false
		}
	}

	return ecmath.CheckByteEq(GenerateMultiSigAddress(partners), addr)
}

// GenerateMultiSigAddress generates a merkle tree for an OrderedS List
// of PublicKeys.
func GenerateMultiSigAddress(keys []PublicKey) []byte {
	keySet := make([][]byte, 0)
	for _, pub := range keys {
		keySet = append(keySet, pub.CompressedBytes())
	}
	keyTree := splashmerkle.Tree{}
	keyTree.ConstructTree(keySet)

	v := byte(0x1)       // compression flag
	z := byte(len(keys)) // mutlisig flag

	return append([]byte{v, z}, keyTree.Root.Bytes()[:20]...)
}
