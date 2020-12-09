package p2p

import (
	"WuyaChain/log"
	"sync"
)

// Peer represents a connected remote node.
type Peer struct {
	protocolErr   chan error
	closed        chan struct{}
	//Node          *discovery.Node // remote peer that this peer connects
	disconnection chan string
	protocolMap   map[string]protocolRW // protocol cap => protocol read write wrapper
	//rw            *connection

	wg   sync.WaitGroup
	log  *log.WuyaLog
	lock sync.Mutex
}


type protocolRW struct {
	Protocol
	bQuited bool
	offset  uint16
	in      chan Message // read message channel, message will be transferred here when it is a protocol message
	//rw      MsgReadWriter
	close   chan struct{}
}
