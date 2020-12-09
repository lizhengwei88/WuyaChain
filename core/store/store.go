package store

import (
	"WuyaChain/common"
	"WuyaChain/core/types"
	"math/big"
)

// BlockchainStore is the interface that wraps the atomic CRUD methods of blockchain.
type BlockchainStore interface {
	// DeleteBlockHash deletes the block hash of the specified canonical block height.
	DeleteBlockHash(height uint64) (bool, error)

	// PutBlockHash writes the height-to-blockHash entry in the canonical chain.
	PutBlockHash(height uint64, hash common.Hash) error

	// PutHeadBlockHash writes the HEAD block hash into the store.
	PutHeadBlockHash(hash common.Hash) error

	// GetBlockHash retrieves the block hash for the specified canonical block height.
	GetBlockHash(height uint64) (common.Hash, error)
	// PutBlockHeader serializes a block header with the specified total difficulty (td) into the store.
	// The input parameter isHead indicates if the header is a HEAD block header.
	PutBlockHeader(hash common.Hash, header *types.BlockHeader, td *big.Int, isHead bool) error

	GetBlock(hash common.Hash) (*types.Block,error)

	// GetHeadBlockHash retrieves the HEAD block hash.
	GetHeadBlockHash() (common.Hash, error)

	// GetBlockHeader retrieves the block header for the specified block hash.
	GetBlockHeader(hash common.Hash) (*types.BlockHeader, error)

	// GetBlockTotalDifficulty retrieves a block's total difficulty for the specified block hash.
	GetBlockTotalDifficulty(hash common.Hash) (*big.Int, error)

	// PutReceipts serializes given receipts for the specified block hash.
	PutReceipts(hash common.Hash, receipts []*types.Receipt) error

	// PutBlock serializes the given block with the given total difficulty (td) into the store.
	// The input parameter isHead indicates if the given block is a HEAD block.
	PutBlock(block *types.Block, td *big.Int, isHead bool) error
	// DeleteBlock deletes the block of the specified block hash.
	DeleteBlock(hash common.Hash) error
	// GetBlockByHeight retrieves the block for the specified block height.
	GetBlockByHeight(height uint64) (*types.Block, error)
	// RecoverHeightToBlockMap recover the height-to-block mapping
	RecoverHeightToBlockMap(block *types.Block) error

	// DeleteIndices deletes tx/debt indices of the specified block.
	DeleteIndices(block *types.Block) error

	// GetReceiptsByBlockHash retrieves the receipts for the specified block hash.
	GetReceiptsByBlockHash(hash common.Hash) ([]*types.Receipt, error)

	// GetReceiptByTxHash retrieves the receipt for the specified tx hash.
	GetReceiptByTxHash(txHash common.Hash) (*types.Receipt, error)

	// AddIndices addes tx/debt indices for the specified block.
	AddIndices(block *types.Block) error

	// GetTxIndex retrieves the tx index for the specified tx hash.
	GetTxIndex(txHash common.Hash) (*types.TxIndex, error)

}