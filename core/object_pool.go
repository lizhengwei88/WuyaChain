package core

import (
	"WuyaChain/common"
	"WuyaChain/core/state"
	"WuyaChain/core/store"
	"WuyaChain/core/types"
	"WuyaChain/log"
	"math/big"
	"sync"
	"time"
)

var CachedCapacity = CachedBlocks * 500

type blockchain interface {
	GetCurrentState() (*state.Statedb, error)
	GetStore() store.BlockchainStore
}

// poolObject object for pool, like transaction and debt
type poolObject interface {
	FromAccount() common.Address
	Price() *big.Int
	Nonce() uint64
	GetHash() common.Hash
	Size() int
	ToAccount() common.Address
}

// poolItem the item for pool collection
type poolItem struct {
	poolObject
	common.BaseHeapItem
	timestamp time.Time
}

type getObjectFromBlockFunc func(block *types.Block) []poolObject
type canRemoveFunc func(chain blockchain, state *state.Statedb, item *poolItem) (bool, bool)
type objectValidationFunc func(state *state.Statedb, obj poolObject) error
type afterAddFunc func(obj poolObject)

// Pool is a thread-safe container for block object received from the network or submitted locally.
// An object will be removed from the pool once included in a blockchain or pending time too long (> timeoutDuration).
type Pool struct {
	mutex              sync.RWMutex
	capacity           int
	chain              blockchain
	hashToTxMap        map[common.Hash]*poolItem
	pendingQueue       *pendingQueue
	processingObjects  map[common.Hash]struct{}
	log                *log.WuyaLog
	getObjectFromBlock getObjectFromBlockFunc
	canRemove          canRemoveFunc
	objectValidation   objectValidationFunc
	afterAdd           afterAddFunc
	cachedTxs          *CachedTxs
}

// NewPool creates and returns a transaction pool.
func NewPool(capacity int, chain blockchain, getObjectFromBlock getObjectFromBlockFunc,
	canRemove canRemoveFunc, log *log.WuyaLog, objectValidation objectValidationFunc, afterAdd afterAddFunc, cachedTxs *CachedTxs) *Pool {
	pool := &Pool{
		capacity:           capacity,
		chain:              chain,
		hashToTxMap:        make(map[common.Hash]*poolItem),
		 pendingQueue:       newPendingQueue(),
		processingObjects:  make(map[common.Hash]struct{}),
		log:                log,
		getObjectFromBlock: getObjectFromBlock,
		canRemove:          canRemove,
		objectValidation:   objectValidation,
		afterAdd:           afterAdd,
		// cachedTxs:          NewCachedTxs(CachedCapacity),
		cachedTxs: cachedTxs,
	}
	// pool.cachedTxs.init(chain)

	go pool.loopCheckingPool()

	return pool
}

// check the pool frequently, remove finalized and old txs, reinject the txs not on the chain yet
func (pool *Pool) loopCheckingPool() {
	for {
		pool.mutex.RLock()
		pendingQueueCount := pool.pendingQueue.count()
		pool.mutex.RUnlock()
		if pendingQueueCount > 0 {
			time.Sleep(10 * time.Second)
		} else {
			pool.removeObjects()
			pool.mutex.Lock()
			if len(pool.hashToTxMap) > 0 {
				for _, poolTx := range pool.hashToTxMap {
					if _,ok := pool.processingObjects[poolTx.GetHash()]; ok{
						continue
					}
					pool.pendingQueue.add(poolTx)
					pool.afterAdd(poolTx.poolObject)
				}
			}
			pool.mutex.Unlock()
			time.Sleep(5 * time.Second)
		}
	}
}


// removeObjects removes finalized and old transactions in hashToTxMap
func (pool *Pool) removeObjects() {
	state, err := pool.chain.GetCurrentState()
	if err != nil {
		pool.log.Warn("failed to get current state, err: %s", err)
		return
	}

	objMap := pool.getObjectMap()
	for objHash, poolTx := range objMap {
		objectRemove, cachedTxsRemove := pool.canRemove(pool.chain, state, poolTx)
		if objectRemove {
			if cachedTxsRemove {
				pool.cachedTxs.remove(objHash)
			}
			pool.removeOject(objHash)
		}
	}
}

func (pool *Pool) getObjectMap() map[common.Hash]*poolItem {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	txMap := make(map[common.Hash]*poolItem)
	for hash, tx := range pool.hashToTxMap {
		txMap[hash] = tx
	}

	return txMap
}

// removeOject removes tx of specified tx hash from pool
func (pool *Pool) removeOject(objHash common.Hash) {
	defer pool.mutex.Unlock()
	pool.mutex.Lock()
	pool.doRemoveObject(objHash)
}

// doRemoveObject removes a transaction from pool.
func (pool *Pool) doRemoveObject(objHash common.Hash) {
	if tx := pool.hashToTxMap[objHash]; tx != nil {
		pool.pendingQueue.remove(tx.FromAccount(), tx.Nonce())
		delete(pool.processingObjects, objHash)
		delete(pool.hashToTxMap, objHash)
	}
}