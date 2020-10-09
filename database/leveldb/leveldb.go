package leveldb

import (
	"WuyaChain/database"
	"github.com/syndtr/goleveldb/leveldb"
)

type LevelDB struct {
	db *leveldb.DB
	quitChan chan struct{}
}

//NewLevelDB constructs and returns a LevelDB instance
func NewLevelDB(path string) (database.Database,error)  {
    db,err:=leveldb.OpenFile(path,nil)
    if err!=nil{
    	return nil, err
	}
	result:=&LevelDB{
		db:db,
		quitChan: make(chan struct{}),
	}
	return result,nil
}