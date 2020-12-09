package light

import (
	"WuyaChain/consensus"
	"WuyaChain/core"
	"WuyaChain/core/store"
	"WuyaChain/database"
	"WuyaChain/database/leveldb"
	"WuyaChain/log"
	"WuyaChain/node"
	"WuyaChain/p2p"
	"WuyaChain/wuya"
	"context"
	"path/filepath"
)

// ServiceClient implements service for light mode.
type ServiceClient struct {
	networkID     string
	netVersion    string
	p2pServer     *p2p.Server
	seeleProtocol *LightProtocol
	log           *log.WuyaLog
	odrBackend    *odrBackend

	txPool  *txPool
	//chain   *LightChain
	lightDB database.Database // database used to store blocks and account state.
	chain   *LightChain
	shard uint
}


// NewServiceClient create ServiceClient
func NewServiceClient(ctx context.Context, conf *node.Config, log *log.WuyaLog, dbFolder string, shard uint, engine consensus.Engine) (s *ServiceClient, err error) {
	s = &ServiceClient{
		log:        log,
		networkID:  conf.P2PConfig.NetworkID,
		netVersion: conf.BasicConfig.Version,
		shard:      shard,
	}

	serviceContext := ctx.Value("ServiceContext").(wuya.ServiceContext)
	// Initialize blockchain DB.
	chainDBPath := filepath.Join(serviceContext.DataDir, dbFolder)
	log.Info("NewServiceClient BlockChain datadir is %s", chainDBPath)
	s.lightDB, err = leveldb.NewLevelDB(chainDBPath)
	if err != nil {
		log.Error("NewServiceClient Create lightDB err. %s", err)
		return nil, err
	}

	bcStore := store.NewCachedStore(store.NewBlockchainDatabase(s.lightDB))
	 s.odrBackend = newOdrBackend(bcStore, shard)
	// initialize and validate genesis
	genesis := core.GetGenesis(&conf.WuyaConfig.GenesisConfig)

	err = genesis.InitializeAndValidate(bcStore, s.lightDB)
	if err != nil {
		s.lightDB.Close()
		s.odrBackend.close()
		log.Error("NewServiceClient genesis.Initialize err. %s", err)
		return nil, err
	}

	s.chain, err = newLightChain(bcStore, s.lightDB, s.odrBackend, engine)
	if err != nil {
		s.lightDB.Close()
		s.odrBackend.close()
		log.Error("failed to init chain in NewServiceClient. %s", err)
		return nil, err
	}

	// s.txPool = newTxPool(s.chain, s.odrBackend, s.chain.headerChangedEventManager, s.chain.headRollbackEventManager)

	//s.seeleProtocol, err = NewLightProtocol(conf.P2PConfig.NetworkID, s.txPool, nil, s.chain, false, s.odrBackend, log, shard)
	//if err != nil {
	//	s.lightDB.Close()
	//	s.odrBackend.close()
	//	log.Error("failed to create seeleProtocol in NewServiceClient, %s", err)
	//	return nil, err
	//}
	//
	//s.odrBackend.start(s.seeleProtocol.peerSet)
	log.Info("Light mode started.")
	return s, nil
}