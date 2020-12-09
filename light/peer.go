package light

import (
	"WuyaChain/common"
	"WuyaChain/log"
	"WuyaChain/p2p"
	"math/big"
	"sync"
)

type peer struct {
	*p2p.Peer
	quitCh          chan struct{}
	peerStrID       string
	peerID          common.Address
	version         uint // Seele protocol version negotiated
	head            common.Hash
	headBlockNum    uint64
	td              *big.Int // total difficulty
	lock            sync.RWMutex
	protocolManager *LightProtocol
	rw              p2p.MsgReadWriter // the read write method for this peer

	curSyncMagic    uint32
	blockNumBegin   uint64        // first block number of blockHashArr
	blockHashArr    []common.Hash // block hashes that should be identical with remote server peer, and is only useful in client mode.
	updatedAncestor uint64

	lastAnnounceCodeTime int64
	log             *log.WuyaLog
}

// findIdxByHash finds index of hash in p.blockHashArr, and returns -1 if not found
func (p *peer) findIdxByHash(hash common.Hash) int {
	for idx := 0; idx < len(p.blockHashArr); idx++ {
		if p.blockHashArr[idx] == hash {
			return idx
		}
	}

	return -1
}

