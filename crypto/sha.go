package crypto

import (
    "WuyaChain/common"
    "github.com/ethereum/go-ethereum/rlp"
)

func HashBytes(data ...[]byte) common.Hash  {
 return common.BytesToHash(Keccak25(data...))
}

func Keccak25(data ...[]byte) []byte  {
    s:=NewKeccak256()
    for _,b:=range data{
     s.Write(b)
 }
 return s.Sum(nil)
}

func MustHash(v interface{}) common.Hash {
    return HashBytes(SerializePanic(v))
}

// SerializePanic serialize the input data to byte array.
// Panics on error, e.g. unsupported data type for RLP encoding.
func SerializePanic(in interface{}) []byte {
    bytes, err := Serialize(in)
    if err != nil {
        panic(err)
    }

    return bytes
}

// Serialize wrapper encode
func Serialize(in interface{}) ([]byte, error) {
    return rlp.EncodeToBytes(in)
}