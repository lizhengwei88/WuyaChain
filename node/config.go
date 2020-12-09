package node

import (
    "WuyaChain/common"
    "WuyaChain/core"
    "WuyaChain/metrics"
    "WuyaChain/p2p"
    "crypto/ecdsa"
)

type Config struct {

   BasicConfig BasicConfig
    // The configuration of p2p network
    P2PConfig p2p.Config
    // The WuyaConfig is the configuration to create the wuya service.
    WuyaConfig WuyaConfig
    // metrics config info
    //MetricsConfig *metrics.Config

    // metrics config info
    MetricsConfig *metrics.Config
}

type BasicConfig struct {
    Name string `json:"name"`
    Version string `json:"version"`
    DataDir string `json:"dataDir"`
    // The file system path of the temporary dataset, used for spow
    DataSetDir string `json:"dataSetDir"`

    RpcAddr string `json:"address"`
    Coinbase string `json:"coinbase"`
    MinerAlgorithm string `json:"algorithm"`
}

// Config is the wuya's configuration to create wuya service
type WuyaConfig struct {
    TxConf core.TransactionPoolConfig
    Coinbase common.Address
    CoinbasePrivateKey *ecdsa.PrivateKey
    GenesisConfig core.GenesisInfo
}

func (conf *Config) Clone() *Config {
    cloned := *conf
    if conf.MetricsConfig != nil {
        temp := *conf.MetricsConfig
        cloned.MetricsConfig = &temp
    }

    return &cloned
}
