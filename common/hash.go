package common

import "WuyaChain/common/hexutil"

const (
	HashLength =32
)

// EmptyHash empty hash
var EmptyHash = Hash{}

type Hash [HashLength]byte

// BytesToHash converts bytes to hash value
func BytesToHash(b []byte) Hash {
	a := &Hash{}
	a.SetBytes(b)
	return *a
}

// SetBytes sets the hash to the value of b.
func (a *Hash) SetBytes(b []byte) {
	if len(b) > HashLength {
		b = b[len(b)-HashLength:]
	}

	copy(a[HashLength-len(b):], b)
}

// Bytes returns its actual bits
func (a Hash) Bytes() []byte {
	return a[:]
}


// String returns the string representation of the hash
func (a Hash) String() string {
	return a.Hex()
}

// Hex returns the hex form of the hash
func (a Hash) Hex() string {
	return hexutil.BytesToHex(a[:])
}