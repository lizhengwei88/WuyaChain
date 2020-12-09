package core

import (
	"WuyaChain/common"
	"container/heap"
)

// txCollection represents the nonce sorted transactions of an account.
type txCollection struct {
	txs       map[uint64]*poolItem
	nonceHeap *common.Heap
}

func newTxCollection() *txCollection {
	return &txCollection{
		txs: make(map[uint64]*poolItem),
		nonceHeap: common.NewHeap(func(i, j common.HeapItem) bool {
			iNonce := i.(*poolItem).Nonce()
			jNonce := j.(*poolItem).Nonce()
			return iNonce < jNonce
		}),
	}
}

func (collection *txCollection) len() int {
	return collection.nonceHeap.Len()
}


func (collection *txCollection) remove(nonce uint64) bool {
	if tx := collection.txs[nonce]; tx != nil {
		heap.Remove(collection.nonceHeap, tx.GetHeapIndex())
		delete(collection.txs, nonce)
		return true
	}

	return false
}

func (collection *txCollection) add(tx *poolItem) bool {
	if existTx := collection.txs[tx.Nonce()]; existTx != nil {
		existTx.poolObject = tx.poolObject
		existTx.timestamp = tx.timestamp
		return false
	}

	heap.Push(collection.nonceHeap, tx)
	collection.txs[tx.Nonce()] = tx

	return true
}

func (collection *txCollection) peek() *poolItem {
	if item := collection.nonceHeap.Peek(); item != nil {
		return item.(*poolItem)
	}

	return nil
}

// cmp compares to the specified tx collection based on price and timestamp.
//   For higher price, return 1.
//   For lower price, return -1.
//   Otherwise:
//     For earier timestamp, return 1.
//     For later timestamp, return -1.
//     Otherwise, return 0.
func (collection *txCollection) cmp(other *txCollection) int {
	if other == nil {
		return 1
	}

	iTx, jTx := collection.peek(), other.peek()
	if iTx == nil && jTx == nil {
		return 0
	}

	if jTx == nil {
		return 1
	}

	if iTx == nil {
		return -1
	}

	if r := iTx.Price().Cmp(jTx.Price()); r != 0 {
		return r
	}

	if iTx.timestamp.Before(jTx.timestamp) {
		return 1
	}

	if iTx.timestamp.After(jTx.timestamp) {
		return -1
	}

	return 0
}
