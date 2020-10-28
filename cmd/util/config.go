package util

import (
	"WuyaChain/core"
	"WuyaChain/node"
	"WuyaChain/p2p"
)

// Config is the Configuration of node
type Config struct {
	// basic config for Node
	BasicConfig node.BasicConfig `json:"basic"`
    P2PConfig p2p.Config `json:"p2p"`
	GenesisConfig core.GenesisInfo `json:"genesis"`
}