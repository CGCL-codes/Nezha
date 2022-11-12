package splashecdsa_test

import (
	"crypto/elliptic"
	"crypto/sha256"
	"math/big"
	"testing"

	"Nezha/evm/splashecdsa"
	"Nezha/evm/splashecdsa/ecmath"
)

func TestSigningP256(t *testing.T) {
	// initial params
	curve := elliptic.P256()
	data := sha256.Sum256([]byte("test data"))

	// generate key
	key, err := splashecdsa.GenerateKeys(curve)
	if err != nil {
		t.Error("failed to generate key \n", err)
	}

	// generate sig
	sig, err := key.Sign(data[:])
	if err != nil {
		t.Error("failed to generate signature \n", err)
	}

	// verify sig
	pub := key.GetPublicKey()
	if !pub.Verify(data[:], sig) {
		t.Error("failed to verify signature with public key")
	}

	// aslo quick check the public key
	if !pub.QuickCheck(curve) {
		t.Error("public key failed to pass QuickCheck")
	}

	// attempt verification on invalid data
	if pub.Verify(data[1:4], sig) {
		t.Error("incorrect data was verified")
	}

	badData := data[:]
	badData[0] = 0x0
	if pub.Verify(badData, sig) && !pub.Verify(data[:], sig) {
		t.Error("incorrect data was verified")
	}

	// attempt verification with invalid signature
	invalidSig := sig
	invalidSig.R.Sub(invalidSig.R, big.NewInt(5))
	if pub.Verify(data[:], invalidSig) {
		t.Error("invalid signature was verified")
	}

	// attempt verification with invalid publickey
	altkey, _ := splashecdsa.GenerateKeys(curve)
	altpub := altkey.GetPublicKey()
	if altpub.Verify(data[:], sig) {
		t.Error("incorrect public key was able to verify the signature")
	}

	//reconstruct tampered public key
	key.X = nil
	tamperedPub := key.GetPublicKey()
	if tamperedPub.X == nil {
		t.Error("unable to reconstruct tampered public key")
	}

	// attempt quick check with invalid public key
	tamperedPub.Y = big.NewInt(54)
	if tamperedPub.QuickCheck(elliptic.P256()) {
		t.Error("invalid public key passed quick check")
	}
	tamperedPub.Curve = elliptic.P224()
	if tamperedPub.QuickCheck(elliptic.P256()) {
		t.Error("invalid public key passed quick check")
	}

	//test were successful

}

func TestSigningP224(t *testing.T) {
	// initial params
	curve := elliptic.P224()
	data := sha256.Sum256([]byte("test data"))

	// generate key
	key, err := splashecdsa.GenerateKeys(curve)
	if err != nil {
		t.Error("failed to generate key \n", err)
	}

	// generate sig
	sig, err := key.Sign(data[:])
	if err != nil {
		t.Error("failed to generate signature \n", err)
	}

	// verify sig
	pub := key.GetPublicKey()
	if !pub.Verify(data[:], sig) {
		t.Error("failed to verify signature with public key")
	}

	// attempt verification on invalid data
	if pub.Verify(data[1:4], sig) {
		t.Error("incorrect data was verified")
	}

	// attempt verification with invalid signature
	invalidSig := sig
	invalidSig.R.Sub(invalidSig.R, big.NewInt(5))
	if pub.Verify(data[:], invalidSig) {
		t.Error("invalid signature was verified")
	}

	// attempt verification with invalid publickey
	altkey, _ := splashecdsa.GenerateKeys(curve)
	altpub := altkey.GetPublicKey()
	if altpub.Verify(data[:], sig) {
		t.Error("incorrect public key was able to verify the signature")
	}

	//tets were successful
}

func TestReconstruction(t *testing.T) {
	curve := elliptic.P256()
	data := sha256.Sum256([]byte("test data"))

	// generate key
	key, err := splashecdsa.GenerateKeys(curve)
	if err != nil {
		t.Error("failed to generate key \n", err)
	}

	// generate sig
	sig, err := key.Sign(data[:])
	if err != nil {
		t.Error("failed to generate signature \n", err)
	}

	// reconstruct public key
	pub := sig.ReconstructPublicKey(data[:], curve.Params())

	// test X value is correct
	if key.X.Cmp(pub.X) != 0 {
		t.Error("reconstructed public key has incorrect X value")
	}

	// test Y value is correct
	if key.Y.Cmp(pub.Y) != 0 {
		t.Error("reconstructed public key has incorrect Y value")
	}

	//Check that value lives on the curve
	if !curve.IsOnCurve(pub.X, pub.Y) {
		t.Error("reconstructed public key's ecpoint does not exist on the curve")
	}
}

func TestMultiSig(t *testing.T) {
	//initial vairables
	curve := elliptic.P256()
	data := sha256.Sum256([]byte("test data"))
	t.Logf("correct data: %#x", data)

	// generate 2 test keys
	key1, err := splashecdsa.GenerateMultiSigKey(curve, 0, 2)
	if err != nil {
		t.Error("failed to generate key 1 \n", err)
	}
	key2, err := splashecdsa.GenerateMultiSigKey(curve, 1, 2)
	if err != nil {
		t.Error("failed to generate key 2 \n", err)
	}

	// generate 2 signatures
	sigs := make([]splashecdsa.SplashSignature, 2)
	sigs[0], _ = key1.Sign(data[:])
	sigs[1], _ = key2.Sign(data[:])

	// generate composite multisig address
	addr := splashecdsa.GenerateMultiSigAddress(
		[]splashecdsa.PublicKey{
			key1.GetPublicKey(),
			key2.GetPublicKey(),
		},
	)

	// verify is valid multi sig address
	if !splashecdsa.IsMultiSigAddress(addr) {
		t.Error("multi sig address failed to pass IsMultiSigAddress")
	}

	// verify address is valid multi sig address
	if !splashecdsa.IsAddressValid(addr) {
		t.Error("multi sig address failed to pass isAddressValid")
	}

	// verify signatures
	if !splashecdsa.VerifyMutliSig(sigs, data[:], addr, curve) {
		t.Error("valid multi sig failed to verify")
	}

	// attempt verification with invalid data
	data[31] = 0x0
	t.Logf("bad     data: %#x", data[:])
	if splashecdsa.VerifyMutliSig(sigs, data[:], addr, curve) {
		t.Error("multi sign passed verifcation on incorrect data.")
	}

	//otherwise everything went well
}

func TestEncodingToBytes(t *testing.T) {
	curve := elliptic.P256()
	key, _ := splashecdsa.GenerateKeys(curve)

	// encode key to bytes
	enc := key.Bytes()

	// decode key and compare
	dec := new(splashecdsa.PrivateKey).SetBytes(enc)
	if dec.D.Cmp(key.D) != 0 {
		t.Errorf("private key decoding failed, had: %d, expected: %d", dec.D, key.D)
	}

	// encode public key
	pub := key.GetPublicKey()
	encpub := pub.Bytes()

	// decode public key and compare
	decpub := new(splashecdsa.PublicKey).SetBytes(encpub)
	if decpub.X.Cmp(pub.X) != 0 {
		t.Error("public key to decode, X value was incorrect")
	}
	if decpub.Y.Cmp(pub.Y) != 0 {
		t.Error("public key to decode, Y value was incorrect")
	}

	// encode compress public key
	enccomp := pub.CompressedBytes()

	// decode compressed bytes and compare
	deccomp := new(splashecdsa.PublicKey).SetCompressedBytes(enccomp)
	if deccomp.X.Cmp(pub.X) != 0 {
		t.Error("public key failed to decoded compressed, X value was incorrect")
	}
	if deccomp.Y.Cmp(pub.Y) != 0 {
		t.Error("public key failed to decoded compressed, Y value was incorrect")
	}
}

func TestAddresses(t *testing.T) {
	curve := elliptic.P256()
	key, _ := splashecdsa.GenerateKeys(curve)

	// generate uncompressed address
	addr := key.GetAddress(false)

	// check address is of standard length of 22 bytes
	if len(addr) != 22 {
		t.Errorf("address of incorrect length, got %d, expected 22", len(addr))
	}

	// check address is not of compressed type
	if splashecdsa.IsAddressCompressed(addr) {
		t.Error("uncompressed address falsely reported as compressed by IsAddressCompressed.")
	}

	// check address is not of compressed type
	if splashecdsa.IsMultiSigAddress(addr) {
		t.Error("non multisig address falsely reported as multisig by IsMultiSigAddress.")
	}

	// check address is not of compressed type
	if !splashecdsa.IsAddressValid(addr) {
		t.Error("Address failed to pass IsAddressValid")
	}

	// check address matches
	pub := key.GetPublicKey()
	if !pub.Matches(addr) {
		t.Error("public key doesnt match address")
	}

	//check compressed address
	compAddr := key.GetAddress(true)
	if !splashecdsa.IsAddressCompressed(compAddr) {
		t.Error("compressed address failed to pass IsAddressCompressed")
	}

	// check address matches
	comppub := key.GetPublicKey()
	if !comppub.Matches(addr) {
		t.Error("compressed public key doesnt match address")
	}

	//check invalid address
	invalidAddr := make([]byte, 10)
	if splashecdsa.IsAddressValid(invalidAddr) {
		t.Error("invalid address passed IsAddressValid")
	}

	// check address matches
	if pub.Matches(invalidAddr) {
		t.Error("public key matches invalid address")
	}

}

func TestECMath(t *testing.T) {
	if ecmath.CheckByteEq([]byte{0x0}, nil) {
		t.Error("Byte Equivilancy check passed on nil value")
	}

	if ecmath.CheckByteEq([]byte{0x0}, []byte{0x0, 0x1}) {
		t.Error("Byte Equivilency check passed on unequal pair")
	}

	if ecmath.CheckByteEq([]byte{0x0, 0x8}, []byte{0x0, 0x1}) {
		t.Error("Byte Equivilency check passed on unequal pair")
	}
}
