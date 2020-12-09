package core

import (
	"WuyaChain/common"
	"WuyaChain/common/errors"
	"WuyaChain/core/state"
	"WuyaChain/core/types"
	"WuyaChain/event"
	"WuyaChain/log"
	"time"
)

const transactionTimeoutDuration = 3 * time.Hour

// TransactionPool is a thread-safe container for transactions received from the network or submitted locally.
// A transaction will be removed from the pool once included in a blockchain or pending time too long (> transactionTimeoutDuration).
type TransactionPool struct {
	*Pool
}


// NewTransactionPool creates and returns a transaction pool.
func NewTransactionPool(config TransactionPoolConfig, chain blockchain) *TransactionPool {
	log := log.GetLogger("txpool")
	getObjectFromBlock := func(block *types.Block) []poolObject {
		return txsToObjects(block.GetExcludeRewardTransactions())
	}
	// 1st bool: can remove from object pool
	// 2nd bool: can remove from cachedTxs
	canRemove := func(chain blockchain, state *state.Statedb, item *poolItem) (bool, bool) {
		nowTimestamp := time.Now()
		txIndex, _ := chain.GetStore().GetTxIndex(item.GetHash())
		nonce := state.GetNonce(item.FromAccount())
		duration := nowTimestamp.Sub(item.timestamp)

		// Transactions have been processed or are too old need to delete
		if txIndex != nil || item.Nonce() < nonce || duration > transactionTimeoutDuration {
			if txIndex == nil {
				if item.Nonce() < nonce {
					log.Debug("remove tx %s because nonce too low, account %s, tx nonce %d, target nonce %d", item.GetHash().Hex(),
						item.FromAccount().Hex(), item.Nonce(), nonce)
					return true, false // the true stand for "not timeout"
				} else if duration > transactionTimeoutDuration {
					log.Debug("remove tx %s because not packed for more than three hours", item.GetHash().Hex())
					return true, true
				}
			}
			return true, false
		}

		return false, false
	}

	objectValidation := func(state *state.Statedb, obj poolObject) error {
		tx := obj.(*types.Transaction)
		if err := tx.Validate(state, common.ThirdForkHeight); err != nil {
			return errors.NewStackedError(err, "failed to validate tx")
		}

		return nil
	}

	afterAdd := func(obj poolObject) {
		log.Debug("receive transaction and add it. transaction hash: %v, time: %d", obj.GetHash(), time.Now().UnixNano())

		// fire event
		event.TransactionInsertedEventManager.Fire(obj.(*types.Transaction))
	}

	cachedTxs := NewCachedTxs(CachedCapacity)
	cachedTxs.init(chain)

	pool := NewPool(config.Capacity, chain, getObjectFromBlock, canRemove, log, objectValidation, afterAdd, cachedTxs)

	return &TransactionPool{pool}
}

func txsToObjects(txs []*types.Transaction) []poolObject {
	objects := make([]poolObject, len(txs))
	for index, tx := range txs {
		objects[index] = tx
	}

	return objects
}
