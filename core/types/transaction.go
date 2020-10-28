package types

import (
	"WuyaChain/common"
	"WuyaChain/crypto"
	"math/big"
)

type Transaction struct {
	Hash      common.Hash
	Data      TransactionData
	Signature crypto.Signature
}

type TransactionData struct {
	From common.Address
	To common.Address
    Amount *big.Int
	AccountNonce uint64
	Timestamp uint64
}