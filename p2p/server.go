package p2p

import (
	"WuyaChain/common"
	"WuyaChain/core"
	"WuyaChain/crypto"
	"WuyaChain/p2p/discovery"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"net"
)

type Config struct {
	// p2p.server will listen for incoming tcp connections. And it is for udp address used for Kad protocol
	ListenAddr string `json:"address"`

	// NetworkID used to define net type, for example main net and test net.
	NetworkID string `json:"networkID"`

	PrivateKey *ecdsa.PrivateKey `json:"-"`
}

type Server struct {
	Config
	running bool  //Start state
	genesis core.GenesisInfo

	// genesisHash is used for handshake
	genesisHash common.Hash
	SelfNode *discovery.Node

}

func NewServer(genesis core.GenesisInfo,config Config,protocol []Protocol )  *Server{
   shard:=genesis.ShardNumber
   genesis.ShardNumber=0
   //balance:=genesis.Balance
   hash:=genesis.Hash()
	genesis.MasterAccount, _ = common.HexToAddress("0x0000000000000000000000000000000000000000")
	//masterAccount:=genesis.MasterAccount
   genesis.Balance=big.NewInt(0)
   genesis.ShardNumber=shard
   return &Server{
	   Config:               config,
	   genesis:              genesis,
	   genesisHash:          hash,
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