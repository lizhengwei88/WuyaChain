package miner

import (
	"WuyaChain/common"
	"WuyaChain/common/memory"
	"WuyaChain/consensus"
	"WuyaChain/core/state"
	"WuyaChain/core/txs"
	"WuyaChain/core/types"
	"WuyaChain/database"
	"WuyaChain/log"
	"fmt"
	"math/big"
	"time"
)

type Task struct {
	header *types.BlockHeader
	txs []*types.Transaction
	receipts []*types.Receipt
	coinbase common.Address
}

func NewTask(header *types.BlockHeader,coinbase common.Address) *Task {
	return &Task{
		header: header,
		coinbase: coinbase,
	}

}

// generateBlock builds a block from task
func (task *Task) generateBlock() *types.Block {
	return types.NewBlock(task.header, task.txs, task.receipts)
	//return types.NewBlock(task.header, task.txs, task.receipts, task.debts)
}


// applyTransactionsAndDebts TODO need to check more about the transactions, such as gas limit
func (task *Task) applyTransactionsAndDebts(wuya WuyaBackend, statedb *state.Statedb, accountStateDB database.Database, log *log.WuyaLog) error {
 	now := time.Now()
	// entrance
	memory.Print(log, "task applyTransactionsAndDebts entrance", now, false)

	// the reward tx will always be at the first of the block's transactions
	reward, err := task.handleMinerRewardTx(statedb)
	if err != nil {
		return err
	}

	//task.chooseTransactions(seele, statedb, log, size)
	fmt.Printf("mining block height:%d, reward:%s, transaction number:%d",
		task.header.Height, reward, len(task.txs))
	log.Info("mining block height:%d, reward:%s, transaction number:%d",
		task.header.Height, reward, len(task.txs))
	batch := accountStateDB.NewBatch()

	root, err := statedb.Commit(batch)
	if err != nil {
		return err
	}

	task.header.StateHash = root

	// exit
	memory.Print(log, "task applyTransactionsAndDebts exit", now, true)

	return nil
}

// handleMinerRewardTx handles the miner reward transaction.
func (task *Task) handleMinerRewardTx(statedb *state.Statedb) (*big.Int, error) {

	reward := consensus.GetReward(task.header.Height)

	rewardTx, err := txs.NewRewardTx(task.coinbase, reward, task.header.CreateTimestamp.Uint64())
	if err != nil {
		return nil, err
	}

	rewardTxReceipt, err := txs.ApplyRewardTx(rewardTx, statedb)
	if err != nil {
		return nil, err
	}

	task.txs = append(task.txs, rewardTx)

	// add the receipt of the reward tx
	task.receipts = append(task.receipts, rewardTxReceipt)

	return reward, nil
}