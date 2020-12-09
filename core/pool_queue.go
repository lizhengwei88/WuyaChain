package core

import (
	"WuyaChain/common"
	"container/heap"
)

type heapedTxList struct {
	common.BaseHeapItem
	*txCollection
}

type heapedTxListPair struct {
	best  *heapedTxList
	worst *heapedTxList
}

// pendingQueue represents the heaped transactions that grouped by account.
type pendingQueue struct {
	txs       map[common.Address]*heapedTxListPair
	bestHeap  *common.Heap
	worstHeap *common.Heap
}

func newPendingQueue() *pendingQueue {
	return &pendingQueue{
		txs: make(map[common.Address]*heapedTxListPair),
		bestHeap: common.NewHeap(func(i, j common.HeapItem) bool {
			iCollection := i.(*heapedTxList).txCollection
			jCollection := j.(*heapedTxList).txCollection
			return iCollection.cmp(jCollection) > 0
		}),
		worstHeap: common.NewHeap(func(i, j common.HeapItem) bool {
			iCollection := i.(*heapedTxList).txCollection
			jCollection := j.(*heapedTxList).txCollection
			return iCollection.cmp(jCollection) <= 0
		}),
	}
}

func (q *pendingQueue) count() int {
	sum := 0

	for _, pair := range q.txs {
		sum += pair.best.len()
	}

	return sum
}

func (q *pendingQueue) remove(addr common.Address, nonce uint64) {
	pair := q.txs[addr]
	if pair == nil {
		return
	}

	if !pair.best.remove(nonce) {
		return
	}

	if pair.best.len() == 0 {
		delete(q.txs, addr)
		heap.Remove(q.bestHeap, pair.best.GetHeapIndex())
		heap.Remove(q.worstHeap, pair.worst.GetHeapIndex())
	} else {
		heap.Fix(q.bestHeap, pair.best.GetHeapIndex())
		heap.Fix(q.worstHeap, pair.worst.GetHeapIndex())
	}
}

func (q *pendingQueue) add(tx *poolItem) {
	if pair := q.txs[tx.FromAccount()]; pair != nil {
		pair.best.add(tx)

		heap.Fix(q.bestHeap, pair.best.GetHeapIndex())
		heap.Fix(q.worstHeap, pair.worst.GetHeapIndex())
	} else {
		collection := newTxCollection()
		collection.add(tx)

		pair := &heapedTxListPair{
			best:  &heapedTxList{txCollection: collection},
			worst: &heapedTxList{txCollection: collection},
		}

		q.txs[tx.FromAccount()] = pair
		heap.Push(q.bestHeap, pair.best)
		heap.Push(q.worstHeap, pair.worst)
	}
}