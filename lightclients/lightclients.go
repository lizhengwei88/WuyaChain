package lightclients

import (
	"WuyaChain/common"
	"WuyaChain/consensus"

	"WuyaChain/light"

	"WuyaChain/node"
	"context"

	lru "github.com/hashicorp/golang-lru"
	"math/big"

)


// LightClientsManager manages light clients of other shards and provides services for debt validation.
type LightClientsManager struct {
	lightClients        []*light.ServiceClient
	lightClientsBackend []*light.LightBackend
	confirmedTxs        []*lru.Cache
	packedDebts         []*lru.Cache

	localShard uint
}


// NewLightClientManager create a new LightClientManager instance.
func NewLightClientManager(targetShard uint, context context.Context, config *node.Config, engine consensus.Engine) (*LightClientsManager, error) {
	clients := make([]*light.ServiceClient, common.ShardCount+1)
	backends := make([]*light.LightBackend, common.ShardCount+1)
	confirmedTxs := make([]*lru.Cache, common.ShardCount+1)


	copyConf := config.Clone()
	//var err error
	for i := 1; i <= common.ShardCount; i++ {
		if i == int(targetShard) {
			continue
		}

		shard := uint(i)
		copyConf.WuyaConfig.GenesisConfig.ShardNumber = shard

		if shard == uint(1) {
			copyConf.WuyaConfig.GenesisConfig.Masteraccount, _ = common.HexToAddress("0xd9dd0a837a3eb6f6a605a5929555b36ced68fdd1")
			copyConf.WuyaConfig.GenesisConfig.Balance = big.NewInt(175000000000000000)
		} else if shard == uint(2) {
			copyConf.WuyaConfig.GenesisConfig.Masteraccount, _ = common.HexToAddress("0xc71265f11acdacffe270c4f45dceff31747b6ac1")
			copyConf.WuyaConfig.GenesisConfig.Balance = big.NewInt(175000000000000000)
		} else if shard == uint(3) {
			copyConf.WuyaConfig.GenesisConfig.Masteraccount, _ = common.HexToAddress("0x509bb3c2285a542e96d3500e1d04f478be12faa1")
			copyConf.WuyaConfig.GenesisConfig.Balance = big.NewInt(175000000000000000)
		} else if shard == uint(4) {
			copyConf.WuyaConfig.GenesisConfig.Masteraccount, _ = common.HexToAddress("0xc6c5c85c585ee33aae502b874afe6cbc3727ebf1")
			copyConf.WuyaConfig.GenesisConfig.Balance = big.NewInt(175000000000000000)
		} else {
			copyConf.WuyaConfig.GenesisConfig.Masteraccount, _ = common.HexToAddress("0x0000000000000000000000000000000000000000")
			copyConf.WuyaConfig.GenesisConfig.Balance = big.NewInt(0)
		}

	//	dbFolder := filepath.Join("db", fmt.Sprintf("lightchainforshard_%d", i))
		//clients[i], err = light.NewServiceClient(context, copyConf, log.GetLogger(fmt.Sprintf("lightclient_%d", i)), dbFolder, shard, engine)
		//if err != nil {
		//	return nil, err
		//}

	//	backends[i] = light.NewLightBackend(clients[i])

		// At most, shardCount * 8K (txs+dets) hash values cached.
		// In case of 8 shards, 64K hash values cached, consuming about 2M memory.
		confirmedTxs[i] = common.MustNewCache(4096)
		}

	return &LightClientsManager{
		lightClients:        clients,
		lightClientsBackend: backends,
		confirmedTxs:        confirmedTxs,
		localShard:          targetShard,
	}, nil
}
