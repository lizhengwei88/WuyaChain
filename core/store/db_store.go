package store

import (
	"WuyaChain/common"
	"WuyaChain/core/types"
	"WuyaChain/database"
	"encoding/binary"
	"math/big"
)

var (
	keyHeadBlockHash = []byte("HeadBlockHash")

	keyPrefixHash   = []byte("H")
	keyPrefixHeader = []byte("h")
	keyPrefixTD     = []byte("t")
)

// blockchainDatabase wraps a database used for the blockchain
type blockchainDatabase struct {
	db database.Database
}

func NewBlockchainDatabase(db database.Database) BlockchainStore {
	return &blockchainDatabase{db}
}

func heightToHashKey(height uint64) []byte { return append(keyPrefixHash, encodeBlockHeight(height)...) }
func hashToHeaderKey(hash []byte) []byte   { return append(keyPrefixHeader, hash...) }
func hashToTDKey(hash []byte) []byte       { return append(keyPrefixTD, hash...) }

// encodeBlockHeight encodes a block height as big endian uint64
func encodeBlockHeight(height uint64) []byte {
	encoded := make([]byte, 8)
	binary.BigEndian.PutUint64(encoded, height)
	return encoded
}

// GetBlockHash gets the hash of the block with the specified height in the blockchain database
func (store *blockchainDatabase) GetBlockHash(height uint64) (common.Hash, error) {
 	hashBytes, err := store.db.Get(heightToHashKey(height))
 	if err != nil {
 		return common.EmptyHash, err
	}

	return common.BytesToHash(hashBytes), nil
}

func (store *blockchainDatabase) GetHeadBlockHash() (common.Hash,error){
	hashBytes,err:=store.db.Get(keyHeadBlockHash)
	if err!=nil{
     return common.EmptyHash, err
	}
	return common.BytesToHash(hashBytes),err
}

func (store *blockchainDatabase) GetBlockHead(hash common.Hash) (*types.BlockHeader, error) {
	headerBytes, err := store.db.Get(hashToHeaderKey(hash.Bytes()))
	if err != nil {
		return nil, err
	}
	header := new(types.BlockHeader)
	if err := common.Deserialize(headerBytes, header); err != nil {
		return nil, err
	}
	return header,err
}

func (store *blockchainDatabase) GetBlock(hash common.Hash) (*types.Block, error) {
	header,err:=store.GetBlockHead(hash)
	if err!=nil{
		return nil, err
	}
	Block:=&types.Block{}
	Block.HeaderHash=hash
	Block.Header=header
	return Block, err
}

// PutBlockHeader serializes the given block header of the block with the specified hash
// and total difficulty into the blockchain database.
// isHead indicates if the given header is the HEAD block header
func (store *blockchainDatabase) PutBlockHeader(hash common.Hash, header *types.BlockHeader, td *big.Int, isHead bool) error {
	return store.putBlockInternal(hash, header, td, isHead)
}

func (store *blockchainDatabase) putBlockInternal(hash common.Hash, header *types.BlockHeader, td *big.Int, isHead bool) error {
 	if header == nil {
		panic("header is nil")
	}
	headerBytes := common.SerializePanic(header)

	hashBytes := hash.Bytes()

	batch := store.db.NewBatch()

	batch.Put(hashToHeaderKey(hashBytes), headerBytes)
	batch.Put(hashToTDKey(hashBytes), common.SerializePanic(td))
	if isHead {
		batch.Put(heightToHashKey(header.Height), hashBytes)
		batch.Put(keyHeadBlockHash, hashBytes)
	}
 	return batch.Commit()
}
