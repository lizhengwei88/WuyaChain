package node

type BasicConfig struct {
    Name string `json:"name"`
    Version string `json:"version"`
    DataDir string `json:"dataDir"`
    RpcAddr string `json:"address"`
    Coinbase string `json:"coinbase"`
    MinerAlgorithm string `json:"algorithm"`
}
