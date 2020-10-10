package discovery

import (
	"WuyaChain/common"
	"net"
)

type Node struct {
	ID common.Address //public key
	IP net.IP
	UDPPort,TCPPort int
	Shard uint
	sha common.Hash
}
func NewNode(id common.Address,ip net.IP,port int,shard uint)  *Node{
    return &Node{
    	ID: id,
    	IP: ip,
    	UDPPort: port,
    	Shard: shard,
	}
}

