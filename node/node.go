package node

import (
	"WuyaChain/log"
	"WuyaChain/p2p"
	"errors"
	"fmt"
)

// error infos
var (
	ErrConfigIsNull       = errors.New("config info is null")
	ErrLogIsNull          = errors.New("SeeleLog is null")
	ErrNodeRunning        = errors.New("node is already running")
	ErrNodeStopped        = errors.New("node is not started")
	ErrServiceStartFailed = errors.New("failed to start node service")
	ErrServiceStopFailed  = errors.New("failed to stop node service")
)

type Service interface {
	ProtoCol() []p2p.Protocol
	Stop() error
}

type Node struct {
	config   *Config
	log      *log.WuyaLog
	server   *p2p.Server
	services []Service
}

// New creates a new P2P node.
func NewPToP(nodeCofig *Config) (*Node, error) {
	nlog := log.GetLogger("node")
	node := &Node{
		config: nodeCofig,
		log:    nlog,
	}
	//err:=node.checkConfig()
	//if err!=nil{
	//	return nil,err
	//}
	return node, nil
}

func (n *Node) Start() error {
	protocols := make([]p2p.Protocol, 0)
	if n.server != nil {
		return ErrNodeRunning
	}
	//start p2p server
	for _, service := range n.services {
		protocols = append(protocols, service.ProtoCol()...)
	}

	p2pServer := p2p.NewServer(n.config.WuyaConfig.GenesisConfig, n.config.P2PConfig, protocols)
    err:=p2pServer.Start(n.config.BasicConfig.DataDir,n.config.WuyaConfig.GenesisConfig.ShardNumber)
	if err!=nil{
		return ErrServiceStartFailed
	}
	n.server=p2pServer
	fmt.Println("p2p net启动成功！！！")
    return nil
}

func (n *Node) checkConfig() error {
	//specificShard:=n.config.WuyaConfig.GenesisConfig.ShardNumber

	return nil
}
