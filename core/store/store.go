package store

import (
	"WuyaChain/common"
	"WuyaChain/core/types"
	"math/big"
)

// BlockchainStore is the interface that wraps the atomic CRUD methods of blockchain.
type BlockchainStore interface {
	// GetBlockHash retrieves the block hash for the specified canonical block height.
	GetBlockHash(height uint64) (common.Hash, error)
	// PutBlockHeader serializes a block header with the specified total difficulty (td) into the store.
	// The input parameter isHead indicates if the header is a HEAD block header.
	PutBlockHeader(hash common.Hash, header *types.BlockHeader, td *big.Int, isHead bool) error

	GetBlock(hash common.Hash) (*types.Block,error)

	// GetHeadBlockHash retrieves the HEAD block hash.
	GetHeadBlockHash() (common.Hash, error)



}