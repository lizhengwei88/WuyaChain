package core

import (
	"WuyaChain/common"
	"WuyaChain/common/errors"
	"WuyaChain/consensus"
	"WuyaChain/core/state"
	"WuyaChain/core/store"
	"WuyaChain/core/svm"
	"WuyaChain/core/txs"
	"WuyaChain/core/types"
	"WuyaChain/database"
	"WuyaChain/event"
	"WuyaChain/log"
	"WuyaChain/metrics"
	"fmt"
	leveldbErrors "github.com/syndtr/goleveldb/leveldb/errors"
	"math/big"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// limit block should not be ahead of 10 seconds of current time
	futureBlockLimit int64 = 10
	// BlockByteLimit is the limit of size in bytes
	BlockByteLimit = 1024 * 1024
)

var (
	// ErrBlockStateHashMismatch is returned when the calculated account state hash of block
	// does not match the state root hash in block header.
	ErrBlockStateHashMismatch = errors.New("block state hash mismatch")
	// ErrBlockCreateTimeNull is returned when block create time is nil
	ErrBlockCreateTimeNull = errors.New("block must have create time")
	// ErrBlockReceiptHashMismatch is returned when the calculated receipts hash of block
	// does not match the receipts root hash in block header.
	ErrBlockReceiptHashMismatch = errors.New("block receipts hash mismatch")
	// ErrBlockEmptyTxs is returned when writing a block with empty transactions.
	ErrBlockEmptyTxs = errors.New("empty transactions in block")
	// ErrBlockTooManyTxs is returned when block have too many txs
	ErrBlockTooManyTxs = errors.New("block have too many transactions")
	// ErrBlockCreateTimeInFuture is returned when block create time is ahead of 10 seconds of now
	ErrBlockCreateTimeInFuture = errors.New("future block. block time is ahead 10 seconds of now")
	// ErrBlockExtraDataNotEmpty is returned when the block extra data is not empty.
	ErrBlockExtraDataNotEmpty = errors.New("block extra data is not empty")
)

type Blockchain struct {
	bcStore        store.BlockchainStore
	accountStateDB database.Database
	engine         consensus.Engine
	genesisBlock   *types.Block
	lock           sync.RWMutex // lock for update blockchain info. for example write block
	blockLeaves    *BlockLeaves
	currentBlock   atomic.Value
	log            *log.WuyaLog
	rp             *recoveryPoint // used to recover blockchain in case of program crashed when write a block

	lastBlockTime time.Time // last sucessful written block time.
}

func NewBlockchain(bcStore store.BlockchainStore, accountStateDB database.Database, recoveryPointFile string, engine consensus.Engine,
	startHeight int) (*Blockchain, error) {
	bc := &Blockchain{
		bcStore:        bcStore,
		accountStateDB: accountStateDB,
		engine:         engine,
		log:            log.GetLogger("blockchain"),
		lastBlockTime:  time.Now(),
	}

	var err error

	// recover from program crash
	bc.rp, err = loadRecoveryPoint(recoveryPointFile)
	if err != nil {
		return nil, errors.NewStackedErrorf(err, "failed to load recovery point info from file %v", recoveryPointFile)
	}

	if err = bc.rp.recover(bcStore); err != nil {
		return nil, errors.NewStackedErrorf(err, "failed to recover blockchain with RP %+v", *bc.rp)
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
	currentBlock, err := bcStore.GetBlock(currentHeadHash)
	if err != nil {
		return nil, err
	}
	bc.currentBlock.Store(currentBlock)

	//blockIndex := NewBlockIndex(currentHeaderHash, currentBlock.Header.Height, td)
	bc.blockLeaves = NewBlockLeaves()
	//bc.blockLeaves.Add(blockIndex)
	return bc, nil
}

// AccountDB returns the account state database in blockchain.
func (bc *Blockchain) AccountDB() database.Database {
	return bc.accountStateDB
}

// CurrentBlock returns the HEAD block of the blockchain.
func (bc *Blockchain) CurrentBlock() *types.Block {
	return bc.currentBlock.Load().(*types.Block)
}

// UpdateCurrentBlock updates the HEAD block of the blockchain.
func (bc *Blockchain) UpdateCurrentBlock(block *types.Block) {
	bc.currentBlock.Store(block)
}

func (bc *Blockchain) AddBlockLeaves(blockIndex *BlockIndex) {
	bc.blockLeaves.Add(blockIndex)
}

func (bc *Blockchain) RemoveBlockLeaves(hash common.Hash) {
	bc.blockLeaves.Remove(hash)
}

// CurrentHeader returns the HEAD block header of the blockchain.
func (bc *Blockchain) CurrentHeader() *types.BlockHeader {
	return bc.CurrentBlock().Header
}

// GetCurrentState returns the state DB of the current block.
func (bc *Blockchain) GetCurrentState() (*state.Statedb, error) {
	block := bc.CurrentBlock()
	return state.NewStatedb(block.Header.StateHash, bc.accountStateDB)
}

// GetHeaderByHeight retrieves a block header by height.
func (bc *Blockchain) GetHeaderByHeight(height uint64) *types.BlockHeader {
	hash, err := bc.bcStore.GetBlockHash(height)
	if err != nil {
		bc.log.Debug("get block header by height failed, err %s. height %d", err, height)
		return nil
	}

	return bc.GetHeaderByHash(hash)
}

// GetHeaderByHash retrieves a block header by hash.
func (bc *Blockchain) GetHeaderByHash(hash common.Hash) *types.BlockHeader {
	header, err := bc.bcStore.GetBlockHeader(hash)
	if err != nil {
		bc.log.Warn("get block header by hash failed, err %s, hash: %v", err, hash)
		return nil
	}

	return header
}

// GetBlockByHash retrieves a block by hash.
func (bc *Blockchain) GetBlockByHash(hash common.Hash) *types.Block {
	block, err := bc.bcStore.GetBlock(hash)
	if err != nil {
		bc.log.Warn("get block by hash failed, err %s", err)
		return nil
	}

	return block
}

// GetState returns the state DB of the specified root hash.
func (bc *Blockchain) GetState(root common.Hash) (*state.Statedb, error) {
	return state.NewStatedb(root, bc.accountStateDB)
}

// GetStateByRootAndBlockHash will panic, since not supported
func (bc *Blockchain) GetStateByRootAndBlockHash(root, blockHash common.Hash) (*state.Statedb, error) {
	panic("unsupported")
}

// Genesis returns the genesis block of blockchain.
func (bc *Blockchain) Genesis() *types.Block {
	return bc.genesisBlock
}

// GetCurrentInfo return the current block and current state info
func (bc *Blockchain) GetCurrentInfo() (*types.Block, *state.Statedb, error) {
	block := bc.CurrentBlock()
	statedb, err := state.NewStatedb(block.Header.StateHash, bc.accountStateDB)
	return block, statedb, err
}

// WriteBlock writes the specified block to the blockchain store.
func (bc *Blockchain) WriteBlock(block *types.Block, txPool *Pool) error {
	startWriteBlockTime := time.Now()
	if err := bc.doWriteBlock(block, txPool); err != nil {
		return err
	}
	markTime := time.Since(startWriteBlockTime)
	metrics.MetricsWriteBlockMeter.Mark(markTime.Nanoseconds())
	return nil
}

func (bc *Blockchain) doWriteBlock(block *types.Block, pool *Pool) error {
	bc.lock.Lock()
	defer bc.lock.Unlock()
	// validate block
	if err := bc.validateBlock(block); err != nil {
		return errors.NewStackedError(err, "failed to validate block")
	}

	preHeader, err := bc.bcStore.GetBlockHeader(block.Header.PreviousBlockHash)
	if err != nil {
		return errors.NewStackedErrorf(err, "failed to get block header by hash %v", block.Header.PreviousBlockHash)
	}
	// Process the txs in the block and check the state root hash.
	var blockStatedb *state.Statedb
	var receipts []*types.Receipt
	if blockStatedb, receipts, err = bc.applyTxs(block, preHeader.StateHash); err != nil {
		return errors.NewStackedError(err, "failed to apply block txs")
	}
	// Validate receipts root hash.
	if receiptsRootHash := types.ReceiptMerkleRootHash(receipts); !receiptsRootHash.Equal(block.Header.ReceiptHash) {
		return ErrBlockReceiptHashMismatch
	}
	// Validate state root hash.
	batch := bc.accountStateDB.NewBatch()
	committed := false
	defer func() {
		if !committed {
			batch.Rollback()
		}
	}()
	var stateRootHash common.Hash
	if stateRootHash, err = blockStatedb.Commit(batch); err != nil {
		return errors.NewStackedError(err, "failed to commit statedb changes to database batch")
	}

	if !stateRootHash.Equal(block.Header.StateHash) {
		return ErrBlockStateHashMismatch
	}

	// Update block leaves and write the block into store.
	currentBlock := &types.Block{
		HeaderHash:   block.HeaderHash,
		Header:       block.Header.Clone(),
		Transactions: make([]*types.Transaction, len(block.Transactions)),
	}
	copy(currentBlock.Transactions, block.Transactions)
	for i, tx := range block.Transactions { // for 1st tx is reward tx, no need to check the duplicate
		if i == 0 {
			continue
		}
		if !pool.cachedTxs.has(tx.Hash) {
			bc.log.Debug("[CachedTxs] add tx %+v from synced block", tx.Hash)
			pool.cachedTxs.add(tx)
		}
	}
	var previousTd *big.Int
	if previousTd, err = bc.bcStore.GetBlockTotalDifficulty(block.Header.PreviousBlockHash); err != nil {
		return errors.NewStackedErrorf(err, "failed to get block TD by hash %v", block.Header.PreviousBlockHash)
	}

	currentTd := new(big.Int).Add(previousTd, block.Header.Difficulty)
	blockIndex := NewBlockIndex(currentBlock.HeaderHash, currentBlock.Header.Height, currentTd)
	isHead := true

	if err = batch.Commit(); err != nil {
		return errors.NewStackedError(err, "failed to batch commit statedb changes to database")
	}

	if err = bc.rp.onPutBlockStart(block, bc.bcStore, isHead); err != nil {
		return errors.NewStackedErrorf(err, "failed to set recovery point before put block into store, isNewHead = %v", isHead)
	}

	if err = bc.bcStore.PutReceipts(block.HeaderHash, receipts); err != nil {
		return errors.NewStackedErrorf(err, "failed to save receipts into store, blockHash = %v, receipts count = %v", block.HeaderHash, len(receipts))
	}

	if err = bc.bcStore.PutBlock(block, currentTd, isHead); err != nil {
		return errors.NewStackedErrorf(err, "failed to save block into store, blockHash = %v, newTD = %v, isNewHead = %v", block.HeaderHash, currentTd, isHead)
	}
	// If the new block has larger TD, the canonical chain will be changed.
	// In this case, need to update the height-to-blockHash mapping for the new canonical chain.
	if isHead {
		largerHeight := block.Header.Height + 1
		if err = DeleteLargerHeightBlocks(bc.bcStore, largerHeight, bc.rp); err != nil {
			bc.log.Error(errors.NewStackedErrorf(err, "failed to delete larger height blocks, height = %v", largerHeight).Error())
		}

		previousHash := block.Header.PreviousBlockHash
		if err = OverwriteStaleBlocks(bc.bcStore, previousHash, bc.rp); err != nil {
			bc.log.Error(errors.NewStackedErrorf(err, "failed to overwrite stale blocks, hash = %v", previousHash).Error())
		}
	}

	// update block header after meta info updated
 	bc.blockLeaves.Add(blockIndex)
 	bc.blockLeaves.Remove(block.Header.PreviousBlockHash)

	committed = true
	if isHead {
		//fmt.Printf("store currentBlock: %d", currentBlock.Header.Height)
		bc.currentBlock.Store(currentBlock)

		bc.blockLeaves.PurgeAsync(bc.bcStore, func(err error) {
			if err != nil {
				bc.log.Error(errors.NewStackedError(err, "failed to purge block").Error())
			}
		})

		event.ChainHeaderChangedEventMananger.Fire(block)
	}

	bc.lastBlockTime = time.Now()

	return nil
}

// validateBlock validates all blockhain independent fields in the block.
func (bc *Blockchain) validateBlock(block *types.Block) error {
	if block == nil {
		return types.ErrBlockHeaderNil
	}

	if err := ValidateBlockHeader(block.Header, bc.engine, bc.bcStore, bc); err != nil {
		return errors.NewStackedError(err, "failed to validate block header")
	}

	if err := block.Validate(); err != nil {
		return errors.NewStackedError(err, "failed to validate block")
	}

	if len(block.Transactions) == 0 {
		return ErrBlockEmptyTxs
	}

	if (types.GetTransactionsSize(block.Transactions[1:])) > BlockByteLimit {
		return ErrBlockTooManyTxs
	}

	// Validate miner shard
	if common.IsShardEnabled() {
		if shard := block.GetShardNumber(); shard != common.LocalShardNumber {
			return fmt.Errorf("invalid shard number. block shard number is [%v], but local shard number is [%v]", shard, common.LocalShardNumber)
		}
	}

	return nil
}

// ValidateBlockHeader validates the specified header.
func ValidateBlockHeader(header *types.BlockHeader, engine consensus.Engine, bcStore store.BlockchainStore, chainReader consensus.ChainReader) error {
	if header == nil {
		return types.ErrBlockHeaderNil
	}

	// Validate timestamp
	if header.CreateTimestamp == nil {
		return ErrBlockCreateTimeNull
	}

	future := new(big.Int).SetInt64(time.Now().Unix() + futureBlockLimit)
	if header.CreateTimestamp.Cmp(future) > 0 {
		return ErrBlockCreateTimeInFuture
	}

	// Now, the extra data in block header should be empty except the genesis block.
	if header.Consensus != types.IstanbulConsensus && len(header.ExtraData) > 0 {
		return ErrBlockExtraDataNotEmpty
	}

	if err := engine.VerifyHeader(chainReader, header); err != nil {
		return errors.NewStackedError(err, "failed to verify header by consensus engine")
	}

	return nil
}

// applyTxs processes the txs in the specified block and returns the new state DB of the block.
// This method supposes the specified block is validated.
func (bc *Blockchain) applyTxs(block *types.Block, root common.Hash) (*state.Statedb, []*types.Receipt, error) {

	statedb, err := state.NewStatedb(root, bc.accountStateDB)
	if err != nil {
		return nil, nil, errors.NewStackedErrorf(err, "failed to create statedb by root hash %v", root)
	}

	// apply txs
	receipts, err := bc.applyRewardAndRegularTxs(statedb, block.Transactions[0], block.Transactions[1:], block.Header)
	if err != nil {
		return nil, nil, errors.NewStackedErrorf(err, "failed to apply reward and regular txs")
	}

	return statedb, receipts, nil
}

func (bc *Blockchain) applyRewardAndRegularTxs(statedb *state.Statedb, rewardTx *types.Transaction, regularTxs []*types.Transaction, blockHeader *types.BlockHeader) ([]*types.Receipt, error) {

	receipts := make([]*types.Receipt, len(regularTxs)+1)
	// validate and apply reward txs
	if err := txs.ValidateRewardTx(rewardTx, blockHeader); err != nil {
		return nil, errors.NewStackedError(err, "failed to validate reward tx")
	}

	rewardReceipt, err := txs.ApplyRewardTx(rewardTx, statedb)
	if err != nil {
		return nil, errors.NewStackedError(err, "failed to apply reward tx")
	}
	receipts[0] = rewardReceipt

	// batch validate signature to improve perf
	if err := types.BatchValidateTxs(regularTxs); err != nil {
		return nil, errors.NewStackedErrorf(err, "failed to batch validate %v txs", len(regularTxs))
	}

	// process regular txs
	for i, tx := range regularTxs {
		txIdx := i + 1

		if err := tx.ValidateState(statedb, blockHeader.Height); err != nil {
			return nil, errors.NewStackedErrorf(err, "failed to validate tx[%v] against statedb", txIdx)
		}

		receipt, err := bc.ApplyTransaction(tx, txIdx, blockHeader.Creator, statedb, blockHeader)
		if err != nil {
			return nil, errors.NewStackedErrorf(err, "failed to apply tx[%v]", txIdx)
		}

		receipts[txIdx] = receipt
	}

	return receipts, nil
}

// ApplyTransaction applies a transaction, changes corresponding statedb and generates its receipt
func (bc *Blockchain) ApplyTransaction(tx *types.Transaction, txIndex int, coinbase common.Address, statedb *state.Statedb,
	blockHeader *types.BlockHeader) (*types.Receipt, error) {
	ctx := &svm.Context{
		Tx:          tx,
		TxIndex:     txIndex,
		Statedb:     statedb,
		BlockHeader: blockHeader,
		BcStore:     bc.bcStore,
	}
	receipt, err := svm.Process(ctx, blockHeader.Height)
	if err != nil {
		return nil, errors.NewStackedError(err, "failed to process tx via svm")
	}

	return receipt, nil
}

// deleteCanonicalBlock deletes the canonical block info for the specified height.
func deleteCanonicalBlock(bcStore store.BlockchainStore, height uint64) (bool, error) {
	hash, err := bcStore.GetBlockHash(height)
	if err == leveldbErrors.ErrNotFound {
		return false, nil
	}

	if err != nil {
		return false, errors.NewStackedErrorf(err, "failed to get block hash by height %v", height)
	}

	// delete the tx/debt indices
	block, err := bcStore.GetBlock(hash)
	if err != nil {
		return false, errors.NewStackedErrorf(err, "failed to get block by hash %v", hash)
	}

	if err = bcStore.DeleteIndices(block); err != nil {
		return false, errors.NewStackedErrorf(err, "failed to delete tx/debt indices of block %v", block.HeaderHash)
	}

	// delete the block hash in canonical chain.
	deleted, err := bcStore.DeleteBlockHash(height)
	if err != nil {
		return false, errors.NewStackedErrorf(err, "failed to delete block hash by height %v", height)
	}

	return deleted, nil
}

// OverwriteStaleBlocks overwrites the stale canonical height-to-hash mappings.
func OverwriteStaleBlocks(bcStore store.BlockchainStore, staleHash common.Hash, rp *recoveryPoint) error {
	var overwritten bool
	var err error

	// When recover the blockchain, the stale block hash my be already overwritten before program crash.
	if _, staleHash, err = overwriteSingleStaleBlock(bcStore, staleHash); err != nil {
		return errors.NewStackedErrorf(err, "failed to overwrite single stale block, hash = %v", staleHash)
	}

	for !staleHash.Equal(common.EmptyHash) {
		if rp != nil {
			rp.onOverwriteStaleBlocks(staleHash)
		}

		if overwritten, staleHash, err = overwriteSingleStaleBlock(bcStore, staleHash); err != nil {
			return errors.NewStackedErrorf(err, "failed to overwrite single stale block, hash = %v", staleHash)
		}

		if !overwritten {
			break
		}
	}

	if rp != nil {
		rp.onOverwriteStaleBlocks(common.EmptyHash)
	}

	return nil
}

func overwriteSingleStaleBlock(bcStore store.BlockchainStore, hash common.Hash) (overwritten bool, preBlockHash common.Hash, err error) {
	header, err := bcStore.GetBlockHeader(hash)
	if err != nil {
		return false, common.EmptyHash, errors.NewStackedErrorf(err, "failed to get block header by hash %v", hash)
	}

	canonicalHash, err := bcStore.GetBlockHash(header.Height)
	if err == nil {
		if hash.Equal(canonicalHash) {
			return false, header.PreviousBlockHash, nil
		}

		// delete the tx/debt indices in previous canonical chain.
		canonicalBlock, err := bcStore.GetBlock(canonicalHash)
		if err != nil {
			return false, common.EmptyHash, errors.NewStackedErrorf(err, "failed to get block by hash %v", canonicalHash)
		}

		if err = bcStore.DeleteIndices(canonicalBlock); err != nil {
			return false, common.EmptyHash, errors.NewStackedErrorf(err, "failed to delete tx/debt indices of block %v", canonicalBlock.HeaderHash)
		}
	}

	// add the tx/debt indices in new canonical chain.
	block, err := bcStore.GetBlock(hash)
	if err != nil {
		return false, common.EmptyHash, errors.NewStackedErrorf(err, "failed to get block by hash %v", hash)
	}

	if err = bcStore.AddIndices(block); err != nil {
		return false, common.EmptyHash, errors.NewStackedErrorf(err, "failed to add tx/debt indices of block %v", block.HeaderHash)
	}

	// update the block hash in canonical chain.
	if err = bcStore.PutBlockHash(header.Height, hash); err != nil {
		return false, common.EmptyHash, errors.NewStackedErrorf(err, "failed to put block height to hash map in canonical chain, height = %v, hash = %v", header.Height, hash)
	}

	return true, header.PreviousBlockHash, nil
}

// GetStore returns the blockchain store instance.
func (bc *Blockchain) GetStore() store.BlockchainStore {
	return bc.bcStore
}
