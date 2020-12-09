package wuya

import (
	"WuyaChain/common"
	"WuyaChain/consensus"
	"WuyaChain/core"
	"WuyaChain/core/store"
	"WuyaChain/core/types"
	"WuyaChain/database"
	"WuyaChain/database/leveldb"
	"WuyaChain/event"
	"WuyaChain/log"
	"WuyaChain/miner"
	"WuyaChain/node"
	"context"
	"fmt"
	"path/filepath"
)

const chainHeaderChangeBuffSize = 100
type WuyaService struct {
	log            *log.WuyaLog
	txPool             *core.TransactionPool
	networkID      string
	netVersion     string
	chainDBPath    string
	chainDB        database.Database
	accountStateDB database.Database // database used to store account state info.
	accountStateDBPath string
	chain          *core.Blockchain
	miner          *miner.Miner
	lastHeader               common.Hash
	chainHeaderChangeChannel chan common.Hash
}

//miner get miner
func (w *WuyaService) Miner() *miner.Miner {
	return w.miner
}

//BlockChain get blockchain
func (w *WuyaService) BlcokChain() *core.Blockchain {
	return w.chain
}

// ServiceContext is a collection of service configuration inherited from node
type ServiceContext struct {
	DataDir string
}

func NewWuyaService(ctx context.Context, conf *node.Config, log *log.WuyaLog, engine consensus.Engine, startHeight int) (*WuyaService, error) {
	w := &WuyaService{
		log:        log,
		networkID:  conf.P2PConfig.NetworkID,
		netVersion: conf.BasicConfig.Version,
	}
	serviceContext := ctx.Value("ServiceContext").(ServiceContext)

	//Init blockchain DB
	w.initBlockChainDB(&serviceContext)

	//leveldb.StartMetrics(s.chainDB, "chaindb", log)

	// Initialize account state info DB.
	if err := w.initAccountStateDB(&serviceContext); err != nil {
		return nil, err
	}

	 //w.miner = miner.NewMiner(conf.WuyaConfig.Coinbase)
	w.miner = miner.NewMiner(conf.WuyaConfig.Coinbase, w, engine)

	//init and validate genesis
	if err := w.initGenesisAndChain(&serviceContext, conf, startHeight); err != nil {
		return nil, err
	}

	if err := w.initPool(conf); err != nil {
		return nil, err
	}

	return w, nil
}

func (w *WuyaService) initBlockChainDB(serviceContext *ServiceContext) (err error) {
	w.chainDBPath = filepath.Join(serviceContext.DataDir, BlockChainDir)
	w.log.Info("NewWuyaService BlockChain datadir is %s", w.chainDBPath)
	w.chainDB, err = leveldb.NewLevelDB(w.chainDBPath)

	if err != nil {
		w.log.Error("NewWuyaService Create BlockChain datadir is %s", w.chainDBPath)
		return err
	}
	return nil
}

func (w *WuyaService) initGenesisAndChain(serviceContext *ServiceContext, conf *node.Config, startHeight int) (err error) {
	 bcStore := store.NewCachedStore(store.NewBlockchainDatabase(w.chainDB))
	//bcStore := store.NewBlockchainDatabase(w.chainDB)
	genesis := core.GetGenesis(&conf.WuyaConfig.GenesisConfig)

	if err = genesis.InitializeAndValidate(bcStore,w.accountStateDB); err != nil {
		//w.Stop()
		w.log.Error("NewSeeleService genesis.Initialize err. %s", err)
		return err
	}

	recoveryPointFile := filepath.Join(serviceContext.DataDir, BlockChainRecoveryPointFile)
	if w.chain, err = core.NewBlockchain(bcStore, w.accountStateDB, recoveryPointFile, w.miner.GetEngine() , startHeight); err != nil {
		w.Stop()
		w.log.Error("failed to init chain in NewSeeleService. %s", err)
		return err
	}

	//w.chain, err = core.NewBlockchain(bcStore, startHeight)
	//if err != nil {
	//	fmt.Println("1111111111")
	//	w.log.Error("failed to init chain in NewWuyaService.%s", err)
	//	return err
	//}

	return nil
}


func (w *WuyaService) initAccountStateDB(serviceContext *ServiceContext) (err error) {
	w.accountStateDBPath = filepath.Join(serviceContext.DataDir, AccountStateDir)
	w.log.Info("NewSeeleService account state datadir is %s", w.accountStateDBPath)

	if w.accountStateDB, err = leveldb.NewLevelDB(w.accountStateDBPath); err != nil {
		w.Stop()
		w.log.Error("NewSeeleService Create BlockChain err: failed to create account state DB, %s", err)
		return err
	}

	return nil
}


// Stop implements node.Service, terminating all internal goroutines.
func (w *WuyaService) Stop() error {
	//TODO
	// s.txPool.Stop() s.chain.Stop()
	// retries? leave it to future
 	if w.chainDB != nil {
		w.chainDB.Close()
		w.chainDB = nil
	}

	if w.accountStateDB != nil {
		w.accountStateDB.Close()
		w.accountStateDB = nil
	}

	return nil
}


// AccountStateDB return account state db
func (s *WuyaService) AccountStateDB() database.Database { return s.accountStateDB }


// BlockChain get blockchain
func (w *WuyaService) BlockChain() *core.Blockchain { return w.chain }

// TxPool tx pool
func (w *WuyaService) TxPool() *core.TransactionPool { return w.txPool }
 // NetVersion net version
func (w *WuyaService) NetVersion() string { return w.netVersion }

// NetWorkID net id
func (w *WuyaService) NetWorkID() string { return w.networkID }


func (w *WuyaService) initPool(conf *node.Config) (err error) {
	if w.lastHeader, err = w.chain.GetStore().GetHeadBlockHash(); err != nil {
		w.Stop()
		return fmt.Errorf("failed to get chain header, %s", err)
	}

	w.chainHeaderChangeChannel = make(chan common.Hash, chainHeaderChangeBuffSize)
	w.txPool = core.NewTransactionPool(conf.WuyaConfig.TxConf, w.chain)

	event.ChainHeaderChangedEventMananger.AddAsyncListener(w.chainHeaderChanged)
	//go w.MonitorChainHeaderChange()

	return nil
}

// chainHeaderChanged handle chain header changed event.
// add forked transaction back
// deleted invalid transaction
func (w *WuyaService) chainHeaderChanged(e event.Event) {
	newBlock := e.(*types.Block)
	if newBlock == nil || newBlock.HeaderHash.IsEmpty() {
		return
	}

	w.chainHeaderChangeChannel <- newBlock.HeaderHash
}