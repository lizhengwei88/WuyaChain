package wuya

import (
	"WuyaChain/core"
	"WuyaChain/core/store"
	"WuyaChain/database"
	"WuyaChain/database/leveldb"
	"WuyaChain/log"
	"WuyaChain/miner"
	"WuyaChain/node"
	"context"
	"path/filepath"
)

type WuyaService struct {
	log            *log.WuyaLog
	networkID      string
	netVersion     string
	chainDBPath    string
	chainDB        database.Database
	accountStateDB database.Database // database used to store account state info.
	chain          *core.Blockchain
	miner          *miner.Miner
}

//func (w *WuyaService) GetLevelDb(conf *node.Config) {
//	fmt.Println("<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
//	bcStore := store.NewCachedStore(store.NewBlockchainDatabase(w.chainDB))
//	//bcStore :=  store.NewBlockchainDatabase(w.chainDB)
//	fmt.Println("bcStore:", bcStore)
//	genesis := core.GetGenesis(&conf.WuyaConfig.GenesisConfig)
//	fmt.Println("======================initGenesisAndChain-genesis==========")
//	fmt.Printf("gensis:%#v", genesis)
//	fmt.Println("======================initGenesisAndChain-genesis==========")
//	var hegiht = []byte{72, 0, 0, 0, 0, 0, 0, 0, 0}
//	hash, found := w.chainDB.Get(hegiht)
//	fmt.Println("看这里。。。。。。。。。。。。。。。", found, hash)
//
//	fmt.Println("看这里。。。。。。。。。。。。。。。")
//	genesis.InitializeAndValidate(bcStore,w.accountStateDB)
//
//}

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

func NewWuyaService(ctx context.Context, conf *node.Config, log *log.WuyaLog, startHeight int) (*WuyaService, error) {
	w := &WuyaService{
		log:        log,
		networkID:  conf.P2PConfig.NetworkID,
		netVersion: conf.BasicConfig.Version,
	}
	serviceContext := ctx.Value("ServiceContext").(ServiceContext)

	//Init blockchain DB
	w.initBlockChainDB(&serviceContext)

	//w.miner = miner.NewMiner(conf.WuyaConfig.Coinbase)

	//init and validate genesis
	if err := w.initGenesisAndChain(&serviceContext, conf, startHeight); err != nil {
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
	//bcStore := store.NewCachedStore(store.NewBlockchainDatabase(w.chainDB))
	bcStore := store.NewBlockchainDatabase(w.chainDB)
	genesis := core.GetGenesis(&conf.WuyaConfig.GenesisConfig)

	if err = genesis.InitializeAndValidate(bcStore,w.accountStateDB); err != nil {
		//w.Stop()
		w.log.Error("NewSeeleService genesis.Initialize err. %s", err)
		return err
	}

	w.chain, err = core.NewBlockchain(bcStore, startHeight)
	if err != nil {
		w.log.Error("failed to init chain in NewWuyaService.%s", err)
		return err
	}

	return nil
}
