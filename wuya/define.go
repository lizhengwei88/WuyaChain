package wuya

const (
	// BlockChainDir blockchain data directory based on config.DataRoot
	BlockChainDir="/db/blockchain"
	// AccountStateDir account state info directory based on config.DataRoot
	AccountStateDir = "/db/accountState"
	// BlockChainRecoveryPointFile is used to store the recovery point info of blockchain.
	BlockChainRecoveryPointFile = "recoveryPoint.json"
)