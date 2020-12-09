package crypto

import (
	"WuyaChain/common"
	"WuyaChain/crypto/secp256k1"
	"crypto/ecdsa"
	"crypto/elliptic"
)

type Signature struct {
	Sig []byte
}

// Verify verifies the signature against the specified hash.
// Return true if the signature is valid, otherwise false.
func (s Signature) Verify(signer common.Address, hash []byte) bool {
	if len(s.Sig) != 65 {
		return false
	}

	pubKey, err := SigToPub(hash, s.Sig)
	if err != nil {
		return false // Signature was modified
	}

	if !GetAddress(pubKey).Equal(signer) {
		return false
	}

	compressed := secp256k1.CompressPubkey(pubKey.X, pubKey.Y)
	return secp256k1.VerifySignature(compressed, hash, s.Sig[:64])
}

func SigToPub(hash, sig []byte) (*ecdsa.PublicKey, error) {
	s, err := Ecrecover(hash, sig)
	if err != nil {
		return nil, err
	}

	x, y := elliptic.Unmarshal(S256(), s)
	return &ecdsa.PublicKey{Curve: S256(), X: x, Y: y}, nil
}

func Ecrecover(hash, sig []byte) ([]byte, error) {
	return secp256k1.RecoverPubkey(hash, sig)
}