package light

import (
	"WuyaChain/consensus"
	"WuyaChain/core/store"
	"WuyaChain/core/types"
	"WuyaChain/event"
	"WuyaChain/log"
	"math/big"
	"sync"
)

// LightChain represents a canonical chain that by default only handles block headers.
type LightChain struct {
	mutex                     sync.RWMutex
	bcStore                   store.BlockchainStore
	odrBackend                *odrBackend
	engine                    consensus.Engine
	currentHeader             *types.BlockHeader
	canonicalTD               *big.Int
	headerChangedEventManager *event.EventManager
	headRollbackEventManager  *event.EventManager
	log                       *log.WuyaLog
}
