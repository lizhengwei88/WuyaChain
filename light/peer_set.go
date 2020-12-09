package light

import (
	"WuyaChain/common"
	"math/rand"
	"sync"
)

type peerFilter struct {
	blockHash common.Hash
}

type peerSet struct {
	peerMap                 map[common.Address]*peer
	peerLastAnnounceTimeMap map[*peer]int64
	lock                    sync.RWMutex
}


func (p *peerSet) choosePeers(filter peerFilter) (choosePeers []*peer) {
	p.lock.Lock()
	defer p.lock.Unlock()

	mapLen := len(p.peerMap)
	peerL := make([]*peer, mapLen)
	var filteredPeers []*peer

	idx := 0
	for _, v := range p.peerMap {
		peerL[idx] = v
		idx++

		if !filter.blockHash.IsEmpty() && v.findIdxByHash(filter.blockHash) != -1 {
			filteredPeers = append(filteredPeers, v)
		}
	}

	const maxPeers = 3

	// choose filtered peers
	if len := len(filteredPeers); len > 0 {
		if len <= maxPeers {
			return filteredPeers
		}

		perm := rand.Perm(len)
		for i := 0; i < maxPeers; i++ {
			choosePeers = append(choosePeers, filteredPeers[perm[i]])
		}

		return
	}

	common.Shuffle(peerL)
	cnt := 0
	for _, p := range peerL {
		cnt++
		choosePeers = append(choosePeers, p)
		if cnt >= maxPeers {
			return
		}
	}

	return
}
