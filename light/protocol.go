package light

import (
	"WuyaChain/common"
	"WuyaChain/core/state"
	"WuyaChain/core/store"
	"WuyaChain/core/types"
	"WuyaChain/event"
	"WuyaChain/log"
	"math/big"
	"sync"
	"time"
)

const (
	statusDataMsgCode           uint16 = 0
	announceRequestCode         uint16 = 1
	announceCode                uint16 = 2
	syncHashRequestCode         uint16 = 3
	syncHashResponseCode        uint16 = 4
	downloadHeadersRequestCode  uint16 = 5
	downloadHeadersResponseCode uint16 = 6

	msgWaitTimeout = time.Second * 60
)

//Protocol base class for high level transfer protocol.
type Protocol struct {
	// Name should contain the official protocol name,
	// often a three-letter word.
	Name string

	// Version should contain the version number of the protocol.
	Version uint

	// Length should contain the number of message codes used by the protocol.
	Length uint16

	//// AddPeer find a new peer will call this method
	//AddPeer func(peer *Peer, rw MsgReadWriter) bool
	//
	//// DeletePeer this method will be called when a peer is disconnected
	//DeletePeer func(peer *Peer)

	// GetPeer this method will be called for get peer information
	GetPeer func(address common.Address) interface{}
}


// BlockChain define some interfaces related to underlying blockchain
type BlockChain interface {
	GetCurrentState() (*state.Statedb, error)
	GetState(root common.Hash) (*state.Statedb, error)
	GetStateByRootAndBlockHash(root, blockHash common.Hash) (*state.Statedb, error)
	GetStore() store.BlockchainStore
	GetHeadRollbackEventManager() *event.EventManager
	CurrentHeader() *types.BlockHeader
	WriteHeader(*types.BlockHeader) error
	PutCurrentHeader(*types.BlockHeader)
	PutTd(*big.Int)
}


// TransactionPool define some interfaces related to add and get txs
type TransactionPool interface {
	AddTransaction(tx *types.Transaction) error
	GetTransaction(txHash common.Hash) *types.Transaction
}


// LightProtocol service implementation of seele
type LightProtocol struct {
	//p2p.Protocol

	bServerMode         bool
	networkID           string
	txPool              TransactionPool
	chain               BlockChain
	//peerSet             *peerSet
	odrBackend          *odrBackend
	//downloader          *Downloader
	wg                  sync.WaitGroup
	quitCh              chan struct{}
	syncCh              chan struct{}
	chainHeaderChangeCh chan common.Hash
	log                 *log.WuyaLog

	shard uint
}


func codeToStr(code uint16) string {
	switch code {
	case statusDataMsgCode:
		return "statusDataMsgCode"
	case announceRequestCode:
		return "announceRequestCode"
	case announceCode:
		return "announceCode"
	case syncHashRequestCode:
		return "syncHashRequestCode"
	case syncHashResponseCode:
		return "syncHashResponseCode"
	case downloadHeadersRequestCode:
		return "downloadHeadersRequestCode"
	case downloadHeadersResponseCode:
		return "downloadHeadersResponseCode"
	case blockRequestCode:
		return "blockRequestCode"
	case blockResponseCode:
		return "blockResponseCode"
	case addTxRequestCode:
		return "addTxRequestCode"
	case addTxResponseCode:
		return "addTxResponseCode"
	case trieRequestCode:
		return "trieRequestCode"
	case trieResponseCode:
		return "trieResponseCode"
	case receiptRequestCode:
		return "receiptRequestCode"
	case receiptResponseCode:
		return "receiptResponseCode"
	case txByHashRequestCode:
		return "txByHashRequestCode"
	case txByHashResponseCode:
		return "txByHashResponseCode"
	case protocolMsgCodeLength:
		return "protocolMsgCodeLength"
	}

	return "unknown"
}
