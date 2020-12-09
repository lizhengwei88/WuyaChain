package p2p

import (
	"WuyaChain/common"
	"WuyaChain/core"
	"WuyaChain/crypto"
	"WuyaChain/log"
	"WuyaChain/p2p/discovery"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"net"
	"sync"
)

const (
	// Maximum number of peers that can be connected for each shard
	maxConnsPerShard = 40
	// Maximum number of peers that node actively connects to.
	maxActiveConnsPerShard = 15
)

type Config struct {
	// p2p.server will listen for incoming tcp connections. And it is for udp address used for Kad protocol
	ListenAddr string `json:"address"`

	// NetworkID used to define net type, for example main net and test net.
	NetworkID string `json:"networkID"`

	// SubPrivateKey which will be make PrivateKey
	SubPrivateKey string `json:"privateKey"`

	// PrivateKey private key for p2p module, do not use it as any accounts
	PrivateKey *ecdsa.PrivateKey `json:"-"`
}

type Server struct {
	Config
	running bool  //Start state
	quit chan struct{}
	peerLock sync.Mutex // lock for peer set
	log      *log.WuyaLog


	// MaxPendingPeers is the maximum number of peers that can be pending in the
	// handshake phase, counted separately for inbound and outbound connections.
	// Zero defaults to preset values.
	MaxPendingPeers int
	genesis core.GenesisInfo

	// genesisHash is used for handshake
	genesisHash common.Hash
	SelfNode *discovery.Node
	// maxConnections represents max connections that node can connect to.
	// Reject connections if srv.PeerCount > maxConnections.
	maxConnections int

	// maxActiveConnections represents max connections that node can actively connect to.
	// Need not connect to a new node if srv.PeerCount > maxActiveConnections.
	maxActiveConnections int


}

func NewServer(genesis core.GenesisInfo,config Config,protocol []Protocol )  *Server{
	shard := genesis.ShardNumber
	genesis.ShardNumber = 0

	// set the master account and balance to empty to calculate hash
	masteraccount := genesis.Masteraccount
	balance := genesis.Balance
	genesis.Masteraccount, _ = common.HexToAddress("0x0000000000000000000000000000000000000000")
	genesis.Balance = big.NewInt(0)

	hash := genesis.Hash()
	genesis.ShardNumber = shard
	genesis.Masteraccount = masteraccount
	genesis.Balance = balance

	return &Server{
		Config:               config,
		running:              false,
		log:                  log.GetLogger("p2p"),
		quit:                 make(chan struct{}),
		//peerSet:              NewPeerSet(),
		//nodeSet:              NewNodeSet(),
		MaxPendingPeers:      0,
		//Protocols:            protocols,
		genesis:              genesis,
		genesisHash:          hash,
		maxConnections:       maxConnsPerShard * common.ShardCount,
		maxActiveConnections: maxActiveConnsPerShard * common.ShardCount,
	}
}

func (srv *Server) Start(nodeDir string,shard uint) (err error)  {
    if srv.running{
		return errors.New("server already running")
	}
	address:=crypto.GetAddress(&srv.PrivateKey.PublicKey)
	addr,err:=net.ResolveUDPAddr("udp",srv.ListenAddr)
	if err!=nil{
		return err
	}
	srv.SelfNode=discovery.NewNode(*address,addr.IP,addr.Port,shard)
	return nil
}

