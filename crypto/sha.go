package crypto

import (
 "WuyaChain/common"
 "github.com/seeleteam/go-seele/crypto/sha3"
)

func HashBytes(data ...[]byte) common.Hash  {
 return common.BytesToHash(Keccak25(data...))
}

func Keccak25(data ...[]byte) []byte  {
    s:=sha3.NewKeccak256()
    for _,b:=range data{
     s.Write(b)
 }
 return s.Sum(nil)
}
