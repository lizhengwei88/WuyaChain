package crypto

import (
	"WuyaChain/common"
	"crypto/ecdsa"
)

func GetAddress(key *ecdsa.PublicKey) *common.Address  {
	addr:=common.PubKeyToAddress(key,MustHash)
	return &addr
}
