package miner

import (
	"WuyaChain/common"
	"WuyaChain/core"
	"WuyaChain/core/types"
	"WuyaChain/log"
	"fmt"
	"math/big"
	"time"
)

type Miner struct {
	log      *log.WuyaLog
	wuya     *core.Blockchain
	current  *Task
	recv     chan *types.Block
	coinbase common.Address
}

func NewMiner(addr common.Address, ) *Miner {
	miner := &Miner{
		coinbase: addr,
		log:      log.GetLogger("miner"),
		recv:                 make(chan *types.Block, 1),
	}
	return miner
}

func (miner *Miner) Start() error {
	fmt.Println("miner start")
	err := miner.prepareNewBlock()
	if err != nil {
		miner.log.Warn(err.Error())
		return err
	}
	return nil
}

func (miner *Miner) prepareNewBlock() error {
	fmt.Println("miner start prepare")
	miner.log.Debug("starting mining the new block")
	fmt.Println("log.debug")
	timestamp := time.Now().UnixNano()
	fmt.Println("timestamp:", timestamp)
	block := miner.wuya.CurrentBlock()
	fmt.Println("block:", block)
	header := newHeaderByParent(block, miner.coinbase, timestamp)
	miner.current = NewTask(header, miner.coinbase)
	return nil
}

func newHeaderByParent(problock *types.Block, coinbase common.Address, timestamp int64) *types.BlockHeader {
	return &types.BlockHeader{
		PreviousBlockHash: problock.HeaderHash,
		Creator:           coinbase,
		Height:            problock.Header.Height + 1,
		CreateTimestamp:   big.NewInt(timestamp),
	}
}
