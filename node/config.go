package node

import (
    "WuyaChain/common"
    "WuyaChain/core"
    "WuyaChain/p2p"
)

type Config struct {

   BasicConfig BasicConfig
    // The configuration of p2p network
    P2PConfig p2p.Config
    // The WuyaConfig is the configuration to create the wuya service.
    WuyaConfig WuyaConfig
    // metrics config info
    //MetricsConfig *metrics.Config
}

type BasicConfig struct {
    Name string `json:"name"`
    Version string `json:"version"`
    DataDir string `json:"dataDir"`
    RpcAddr string `json:"address"`
    Coinbase string `json:"coinbase"`
    MinerAlgorithm string `json:"algorithm"`
}

// Config is the wuya's configuration to create wuya service
type WuyaConfig struct {
    Coinbase common.Address
    GenesisConfig core.GenesisInfo
}
