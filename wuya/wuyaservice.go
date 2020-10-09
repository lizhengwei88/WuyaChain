package wuya

import (
	"WuyaChain/database"
	"WuyaChain/database/leveldb"
	"WuyaChain/node"
	"WuyaChain/log"
	"context"
	"fmt"
	"path/filepath"
)

type WuyaService struct {
	log *log.WuyaLog
	networkID string
	netVersion string
    chainDBPath string
	chainDB database.Database
}

// ServiceContext is a collection of service configuration inherited from node
type ServiceContext struct {
	DataDir string
}

func NewWuyaService(ctx context.Context,conf *node.Config, log *log.WuyaLog)(*WuyaService,error)  {
   w:=&WuyaService{
    log: log,
   	networkID: conf.P2PConfig.NetworkID,
   	netVersion: conf.BasicConfig.Version,
   }
   serviceContext:=ctx.Value("ServiceContext").(ServiceContext)
   //Init blockchain DB
   w.initBlockChainDB(&serviceContext)
   return w,nil
}

func(w *WuyaService) initBlockChainDB(serviceContext *ServiceContext) (err error) {
	w.chainDBPath=filepath.Join(serviceContext.DataDir,BlockChainDir)
    w.log.Info("NewWuyaService BlockChain datadir is %s",w.chainDBPath)
	fmt.Println("log is 002")
    w.chainDB,err=leveldb.NewLevelDB(w.chainDBPath)
	if err!=nil{
		fmt.Println("log is 003")
     w.log.Error("NewWuyaService Create BlockChain datadir is %s",w.chainDBPath)
     return err
	}
    return nil
}