package miner

import (
	"WuyaChain/common"
	"WuyaChain/common/memory"
	"WuyaChain/consensus"
	"WuyaChain/core"
	"WuyaChain/core/types"
	"WuyaChain/event"
	"WuyaChain/log"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"time"
)

// WuyaBackend wraps all methods required for minier.
type WuyaBackend interface {
	TxPool() *core.TransactionPool
	BlockChain() *core.Blockchain
}

type Miner struct {
	mining   int32
	canStart int32
	stopped  int32
	stopper  int32 // manually stop miner
	wg       sync.WaitGroup
	stopChan chan struct{}
	log      *log.WuyaLog

	isFirstDownloader    int32
	isFirstBlockPrepared int32

	current  *Task
	recv     chan *types.Block
	coinbase common.Address
	wuya     WuyaBackend
	engine   consensus.Engine
	msgChan  chan bool // use msgChan to receive msg setting miner to start or stop, and miner will deal with these msgs sequentially
}

// NewMiner constructs and returns a miner instance
func NewMiner(addr common.Address, wuya WuyaBackend, engine consensus.Engine) *Miner {
	miner := &Miner{
		coinbase:             addr,
		canStart:             1,
		stopped:              0,
		stopper:              0,
		wuya:                 wuya,
		wg:                   sync.WaitGroup{},
		recv:                 make(chan *types.Block, 1),
		log:                  log.GetLogger("miner"),
		isFirstDownloader:    1,
		isFirstBlockPrepared: 0,
		//debtVerifier:         verifier,
		engine:  engine,
		msgChan: make(chan bool, 100),
	}

	event.BlockDownloaderEventManager.AddListener(miner.downloaderEventCallback)
	event.TransactionInsertedEventManager.AddAsyncListener(miner.newTxOrDebtCallback)
	//event.DebtsInsertedEventManager.AddAsyncListener(miner.newTxOrDebtCallback)
	go miner.handleMsg()
	return miner
}

func (miner *Miner) handleMsg() {
	for {
		select {
		case msg := <-miner.msgChan:
			if msg == true {
				if miner.CanStart() {
					err := miner.Start()
					if err != nil {
						miner.log.Error("error start miner,%s", err.Error())
					}
				} else {
					miner.log.Warn("cannot start miner,stopper:%d, stopped:%d,mining:%d,canStart:%d",
						atomic.LoadInt32(&miner.stopper),
						atomic.LoadInt32(&miner.stopped),
						atomic.LoadInt32(&miner.mining),
						atomic.LoadInt32(&miner.canStart))
				}
			} else {
				if atomic.LoadInt32(&miner.stopped) == 0 && atomic.LoadInt32(&miner.mining) == 1 {
					miner.Stop()

				} else {
					miner.log.Warn("miner is not working,stopper:%d, stopped:%d,mining:%d,canStart:%d",
						atomic.LoadInt32(&miner.stopper),
						atomic.LoadInt32(&miner.stopped),
						atomic.LoadInt32(&miner.mining),
						atomic.LoadInt32(&miner.canStart))
				}
			}
		}
	}
}

func (miner *Miner) CanStart() bool {
	if atomic.LoadInt32(&miner.stopper) == 0 &&
		atomic.LoadInt32(&miner.stopped) == 1 &&
		atomic.LoadInt32(&miner.mining) == 0 &&
		atomic.LoadInt32(&miner.canStart) == 1 {
		return true
	} else {
		return false
	}
}

// Stop is used to stop the miner
func (miner *Miner) Stop() {
	// set stopped to 1 to prevent restart
	atomic.StoreInt32(&miner.stopped, 1)
	miner.stopMining()
	atomic.StoreInt32(&miner.mining, 0)
	if istanbul, ok := miner.engine.(consensus.Istanbul); ok {
		if err := istanbul.Stop(); err != nil {
			panic(fmt.Sprintf("failed to stop istanbul engine: %v", err))
		}

	}

}

func (miner *Miner) stopMining() {
	// notify all threads to terminate
	if miner.stopChan != nil {
		close(miner.stopChan)
	}
	atomic.StoreInt32(&miner.mining, 0)

	// wait for all threads to terminate
	miner.wg.Wait()
	miner.log.Info("Miner stopped.")
}

// downloaderEventCallback handles events which indicate the downloader state
func (miner *Miner) downloaderEventCallback(e event.Event) {

	switch e.(int) {
	case event.DownloaderStartEvent:
		miner.log.Info("got download start event, stop miner")
		atomic.StoreInt32(&miner.canStart, 0)
		miner.msgChan <- false

	case event.DownloaderDoneEvent, event.DownloaderFailedEvent:
		atomic.StoreInt32(&miner.canStart, 1)
		atomic.StoreInt32(&miner.isFirstDownloader, 0)
		miner.msgChan <- true
	}
}

func (miner *Miner) Start() error {
	miner.stopChan = make(chan struct{})

	// try to prepare the first block
	err := miner.prepareNewBlock(miner.recv)
	if err != nil {
		miner.log.Warn(err.Error())
		return err
	}

	go miner.waitBlock()
	atomic.StoreInt32(&miner.mining, 1)
	atomic.StoreInt32(&miner.stopped, 0)
	miner.log.Info("Miner started")
	return nil
}

func (miner *Miner) prepareNewBlock(recv chan *types.Block) error {
	miner.log.Debug("starting mining the new block")
	timestamp := time.Now().Unix()
	//获取block及状态
	parent, stateDB, err := miner.wuya.BlockChain().GetCurrentInfo()
	if err != nil {
		return fmt.Errorf("failed to get current info, %s", err)
	}

	header := newHeaderByParent(parent, miner.coinbase, timestamp)

	err = miner.engine.Prepare(miner.wuya.BlockChain(), header)
	if err != nil {
		return fmt.Errorf("failed to prepare header, %s", err)
	}

	miner.current = NewTask(header, miner.coinbase)

	err = miner.current.applyTransactionsAndDebts(miner.wuya, stateDB, miner.wuya.BlockChain().AccountDB(), miner.log)
	fmt.Println("miner.current", miner.current.header.Height)

	if err != nil {
		return fmt.Errorf("failed to apply transaction %s", err)
	}

	miner.log.Info("committing a new task to engine, height:%d, difficult:%d", header.Height, header.Difficulty)
	miner.commitTask(miner.current, recv)
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

// newTxOrDebtCallback handles the new tx event
func (miner *Miner) newTxOrDebtCallback(e event.Event) {
	miner.msgChan <- true
}

// waitBlock waits for blocks to be mined continuously
func (miner *Miner) waitBlock() {
 out:
	for {
		select {
		case result := <-miner.recv:
			for {
				if result == nil {
					break
				}
			 	miner.log.Info("found a new mined block, block height:%d, hash:%s, time: %d", result.Header.Height, result.HeaderHash.Hex(), time.Now().UnixNano())
				ret := miner.saveBlock(result)
				if ret != nil {
					miner.log.Error("failed to save the block, for %s", ret.Error())
					break
				}
				//mining lock

				if h, ok := miner.engine.(consensus.Handler); ok {
					h.NewChainHead()
				}

				miner.log.Info("saved mined block successfully")
				event.BlockMinedEventManager.Fire(result) // notify p2p to broadcast the block
				break
			}
			atomic.StoreInt32(&miner.stopped, 1)
			atomic.StoreInt32(&miner.mining, 0)
			// loop mining after mining completed
			miner.newTxOrDebtCallback(event.EmptyEvent)
		case <-miner.stopChan:
			break out
		}
	}
}

// saveBlock saves the block in the given result to the blockchain
func (miner *Miner) saveBlock(result *types.Block) error {
	now := time.Now()
	// entrance
 	memory.Print(miner.log, "miner saveBlock entrance", now, false)
	txPool := miner.wuya.TxPool().Pool

	ret := miner.wuya.BlockChain().WriteBlock(result, txPool)

	// entrance
	memory.Print(miner.log, "miner saveBlock exit", now, true)

	return ret
}

func (miner *Miner) GetEngine() consensus.Engine {
	return miner.engine
}

// commitTask commits the given task to the miner
func (miner *Miner) commitTask(task *Task, recv chan *types.Block) {
	block := task.generateBlock()
	miner.engine.Seal(miner.wuya.BlockChain(), block, miner.stopChan, recv)
}
