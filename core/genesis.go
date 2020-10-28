package core

import (
	"WuyaChain/common"
	"WuyaChain/core/store"
	"WuyaChain/core/types"
	"WuyaChain/crypto"
	"WuyaChain/database"
	"encoding/json"
	"fmt"
	leveldbErrors "github.com/syndtr/goleveldb/leveldb/errors"
	"math/big"
)

type Genesis struct {
	header *types.BlockHeader
	info   *GenesisInfo
}

const genesisBlockHeight = uint64(0)

// GenesisInfo genesis info for generating genesis block, it could be used for initializing account balance
type GenesisInfo struct {

	// Difficult initial difficult for mining. Use bigger difficult as you can. Because block is chosen by total difficult
	Difficult int64 `json:"difficult"`

	// ShardNumber is the shard number of genesis block.
	ShardNumber uint `json:"shard"`

	// CreateTimestamp is the initial time of genesis
	CreateTimestamp *big.Int `json:"timestamp"`

	//master Account
	MasterAccount common.Address `json:"master"`

	// balance of the master account
	Balance *big.Int `json:"balance"`
}

func (info *GenesisInfo) Hash() common.Hash {
	data, err := json.Marshal(info)
	if err != nil {
		panic(fmt.Sprintf("Failed to Marshal err %s", err))
	}
	return crypto.HashBytes(data)
}

func GetGenesis(info *GenesisInfo) *Genesis {
	return &Genesis{
		header: &types.BlockHeader{
			PreviousBlockHash: common.EmptyHash,
			Creator:           common.EmptyAddress,
			Height:            genesisBlockHeight,
			Difficulty: big.NewInt(info.Difficult),
			CreateTimestamp:   info.CreateTimestamp,
		},
		info: info,
	}
}

func (genesis *Genesis) InitializeAndValidate(bcStore store.BlockchainStore, accountStateDB database.Database) error {
	storedGenesisHash, err := bcStore.GetBlockHash(genesisBlockHeight)
	fmt.Println("来看一看，瞧一瞧：", storedGenesisHash,err)
	if err == leveldbErrors.ErrNotFound {
		return genesis.store(bcStore, accountStateDB)
	}
	//TODO 正伟在这里做逻辑库里有数据取出
	return err
}

// store atomically stores the genesis block in the blockchain store.
func (genesis *Genesis) store(bcStore store.BlockchainStore, accountStateDB database.Database) error {
	 //statedb := getStateDB(genesis.info)
	//
	//batch := accountStateDB.NewBatch()
	//fmt.Println("state:",statedb,batch)
	//if _, err := statedb.Commit(batch); err != nil {
	//	return errors.NewStackedError(err, "failed to commit batch into statedb")
	//}
	//
	//if err := batch.Commit(); err != nil {
	//	return errors.NewStackedError(err, "failed to commit batch into database")
	//}
	fmt.Println("============================================")
	fmt.Println("genesis.header.hash:", genesis.header.Hash())
	fmt.Println("genesis.header:", genesis.header)
	fmt.Println("genesis.header.Difficulty:", genesis.header.Difficulty)
	fmt.Println("=============================================")

	err:=bcStore.PutBlockHeader(genesis.header.Hash(),genesis.header,genesis.header.Difficulty,true)
    if err!=nil{
    	fmt.Println("PutBlockHead is err:",err)
		return err
	}
	return nil
}
