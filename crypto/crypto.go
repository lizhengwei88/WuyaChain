package crypto

import (
	"WuyaChain/common"
	"WuyaChain/common/hexutil"
	"WuyaChain/crypto/secp256k1"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
)

func GetAddress(key *ecdsa.PublicKey) *common.Address  {
	addr:=common.PubKeyToAddress(key,MustHash)
	return &addr
}

// LoadECDSAFromString creates ecdsa private key from the given string.
// ecStr should start with 0x or 0X
func LoadECDSAFromString(ecStr string) (*ecdsa.PrivateKey, error) {
	if !hexutil.Has0xPrefix(ecStr) {
		return nil, errors.New("Input string not a valid ecdsa string")
	}
	key, err := hex.DecodeString(ecStr[2:])
	if err != nil {
		return nil, err
	}
	return ToECDSA(key)
}

// ToECDSA creates a private key with the given D value.
func ToECDSA(d []byte) (*ecdsa.PrivateKey, error) {
	return toECDSA(d, true)
}

// toECDSA creates a private key with the given D value. The strict parameter
// controls whether the key's length should be enforced at the curve size or
// it can also accept legacy encodings (0 prefixes).
func toECDSA(d []byte, strict bool) (*ecdsa.PrivateKey, error) {
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = S256()
	if strict && 8*len(d) != priv.Params().BitSize {
		return nil, fmt.Errorf("invalid length, need %d bits", priv.Params().BitSize)
	}
	priv.D = new(big.Int).SetBytes(d)
	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	if priv.PublicKey.X == nil {
		return nil, errors.New("invalid private key")
	}
	return priv, nil
}


// S256 returns an instance of the secp256k1 curve.
func S256() elliptic.Curve {
	return secp256k1.S256()
}

// PubkeyToAddress add this method for istanbul BFT integration
func PubkeyToAddress(key ecdsa.PublicKey) common.Address  {
	return *GetAddress(&key)
}

// Keccak512 calculates and returns the Keccak512 hash of the input data.
func Keccak512(data ...[]byte) []byte {
	d := NewKeccak512()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}