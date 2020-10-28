package core

import (
	"WuyaChain/common"
	"WuyaChain/core/store"
	"WuyaChain/core/types"
	"WuyaChain/log"
	"sync/atomic"
	"time"
)

type Blockchain struct {
	bcStore       store.BlockchainStore
	genesisBlock  *types.Block
	currentBlock  atomic.Value
	log           *log.WuyaLog
	lastBlockTime time.Time // last sucessful written block time.
}

func (bc *Blockchain) CurrentBlock() *types.Block {
	return bc.currentBlock.Load().(*types.Block)
}

func (bc *Blockchain) GetCurrentInfo() *types.Block {
	block := bc.CurrentBlock()
	return block
}

func NewBlockchain(bcStore store.BlockchainStore, startHeight int) (*Blockchain, error) {
	bc := &Blockchain{
		bcStore:       bcStore,
		log:           log.GetLogger("blockchain"),
		lastBlockTime: time.Now(),
	}
	// Get the genesis block from store
	genesisHash, err := bcStore.GetBlockHash(genesisBlockHeight)
 	if err != nil {
	 return nil, err
	}

	//取出创世hash，用hash去库里取block
	bc.genesisBlock, err = bcStore.GetBlock(genesisHash)
	if err != nil {
		return nil, err
	}
	var currentHeadHash common.Hash
	if startHeight == -1 {
		currentHeadHash, err = bcStore.GetHeadBlockHash()
		if err != nil {
 			return nil, err
		}
	} else {
		currentHeight := uint64(startHeight)
		currentHeadHash, err = bcStore.GetBlockHash(currentHeight)
		if err != nil {
			return nil, err
		}
	}
    currentBlock,err:=bcStore.GetBlock(currentHeadHash)
	if err != nil {
 		return nil, err
	}
	bc.currentBlock.Store(currentBlock)
 	return bc, nil
}
