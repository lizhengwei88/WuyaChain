package core

import (
	"encoding/json"
	"fmt"
	"WuyaChain/common"
	"WuyaChain/crypto"
	"math/big"
)

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
     data,err:=json.Marshal(info)
     if err!=nil{
     	panic(fmt.Sprintf("Failed to Marshal err %s",err))
	 }
	 return crypto.HashBytes(data)
}