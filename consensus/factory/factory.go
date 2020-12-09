package factory

import (
	"WuyaChain/common"

	"WuyaChain/consensus"
	"WuyaChain/consensus/ethash"
	"WuyaChain/consensus/pow"
	"WuyaChain/consensus/spow"


	"fmt"

)


// GetConsensusEngine get consensus engine according to miner algorithm name
// WARNING: engine may be a heavy instance. we should have as less as possible in our process.
func GetConsensusEngine(minerAlgorithm string, folder string, percentage int) (consensus.Engine, error) {
	var minerEngine consensus.Engine
	if minerAlgorithm == common.EthashAlgorithm {
		minerEngine = ethash.New(ethash.GetDefaultConfig(), nil, false)
	} else if minerAlgorithm == common.Sha256Algorithm {
		minerEngine = pow.NewEngine(1)
	} else if minerAlgorithm == common.SpowAlgorithm {
		minerEngine = spow.NewSpowEngine(1, folder, percentage)
	} else {
		return nil, fmt.Errorf("unknown miner algorithm")
	}

	return minerEngine, nil
}


//func GetBFTEngine(privateKey *ecdsa.PrivateKey, folder string) (consensus.Engine, error) {
//	path := filepath.Join(folder, common.BFTDataFolder)
//	db, err := leveldb.NewLevelDB(path)
//	if err != nil {
//		return nil, errors.NewStackedError(err, "create bft folder failed")
//	}
//
//	return backend.New(istanbul.DefaultConfig, privateKey, db), nil
//}
