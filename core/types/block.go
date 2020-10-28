package types

import (
	"WuyaChain/common"
	"WuyaChain/crypto"
	"math/big"
)

type BlockHeader struct {
    PreviousBlockHash common.Hash
    Creator common.Address
    TxHash common.Hash
    Difficulty *big.Int
    Height uint64
    CreateTimestamp *big.Int
}

type Block struct {
    HeaderHash common.Hash
    Header *BlockHeader
    Transactions []*Transaction
}

func NewBlock(header *BlockHeader,txs []*Transaction) *Block {
     block:=&Block{
     	Header: header,
     	Transactions: txs,
	 }
     block.HeaderHash=block.Header.Hash()
	 return block
}

func (header *BlockHeader) Hash() common.Hash {
	return crypto.MustHash(header)
}