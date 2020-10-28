package miner

import (
	"WuyaChain/common"
	"WuyaChain/core/types"
)

type Task struct {
	header *types.BlockHeader
	txs []*types.Transaction
	coinbase common.Address
}

func NewTask(header *types.BlockHeader,coinbase common.Address) *Task {
	return &Task{
		header: header,
		coinbase: coinbase,
	}

}
