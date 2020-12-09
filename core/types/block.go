package types

import (
	"WuyaChain/common"
	"WuyaChain/crypto"
	"errors"
	"math/big"
)

var (
	// ErrBlockHeaderNil is returned when the block header is nil.
	ErrBlockHeaderNil = errors.New("block header is nil")
	// ErrBlockHashMismatch is returned when the block hash does not match the header hash.
	ErrBlockHashMismatch = errors.New("block header hash mismatch")

	// ErrBlockTxsHashMismatch is returned when the block transactions hash does not match
	// the transaction root hash in the header.
	ErrBlockTxsHashMismatch = errors.New("block transactions root hash mismatch")
)

type ConsensusType uint

const (
	PowConsensus ConsensusType = iota
	IstanbulConsensus
)

type BlockHeader struct {
	PreviousBlockHash common.Hash
	Creator           common.Address
	StateHash         common.Hash // StateHash is the root hash of the state trie
	TxHash            common.Hash
	ReceiptHash       common.Hash    // ReceiptHash is the root hash of the receipt merkle tree
	Difficulty        *big.Int
	Height            uint64
	CreateTimestamp   *big.Int
	Consensus         ConsensusType
	ExtraData         []byte // ExtraData stores the extra info of block header.
	Witness   []byte
	SecondWitness []byte
}


// Clone returns a clone of the block header.
func (header *BlockHeader) Clone() *BlockHeader {
	clone := *header

	if clone.Difficulty = new(big.Int); header.Difficulty != nil {
		clone.Difficulty.Set(header.Difficulty)
	}

	if clone.CreateTimestamp = new(big.Int); header.CreateTimestamp != nil {
		clone.CreateTimestamp.Set(header.CreateTimestamp)
	}

	clone.ExtraData = common.CopyBytes(header.ExtraData)
	clone.Witness = common.CopyBytes(header.Witness)

	return &clone
}


type Block struct {
	HeaderHash   common.Hash
	Header       *BlockHeader
	Transactions []*Transaction
}

func NewBlock(header *BlockHeader, txs []*Transaction, receipts []*Receipt) *Block {
	block := &Block{
		Header: header.Clone(),
	}

	// Copy the transactions and update the transaction trie root hash.
	block.Header.TxHash = MerkleRootHash(txs)
	if len(txs) > 0 {
		block.Transactions = make([]*Transaction, len(txs))
		copy(block.Transactions, txs)
	}

	block.Header.ReceiptHash = ReceiptMerkleRootHash(receipts)

	// Calculate the block header hash.
	block.HeaderHash = block.Header.Hash()

	return block
}

func (header *BlockHeader) Hash() common.Hash {
	return crypto.MustHash(header)
}

// Validate validates state independent fields in a block.
func (block *Block) Validate() error {
	// Block must have header
	if block.Header == nil {
		return ErrBlockHeaderNil
	}

	// Validate block header hash
	if !block.HeaderHash.Equal(block.Header.Hash()) {
		return ErrBlockHashMismatch
	}

	// Validate tx merkle root hash
	if h := MerkleRootHash(block.Transactions); !h.Equal(block.Header.TxHash) {
		return ErrBlockTxsHashMismatch
	}

	return nil
}

func (b *Block) Height() uint64 {
	return b.Header.Height
}

// GetShardNumber returns the shard number of the block, which means the shard number of the creator.
func (block *Block) GetShardNumber() uint {
	if block.Header == nil {
		return common.UndefinedShardNumber
	}

	return block.Header.Creator.Shard()
}

func (block *Block) WithSeal(header *BlockHeader) *Block {
	return &Block{
		HeaderHash:   header.Hash(),
		Header:       header.Clone(),
		Transactions: block.Transactions,
	}
}

// GetExcludeRewardTransactions returns all txs of a block except for the reward transaction
func (block *Block) GetExcludeRewardTransactions() []*Transaction {
	if len(block.Transactions) == 0 {
		return block.Transactions
	}

	return block.Transactions[1:]
}