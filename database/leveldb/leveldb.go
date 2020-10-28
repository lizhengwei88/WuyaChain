package leveldb

import (
	"WuyaChain/database"
	"github.com/syndtr/goleveldb/leveldb"
)


var (
	// ErrEmptyKey key is empty
	//ErrEmptyKey = errors.New("key could not be empty")
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

// Close is used to close the db when not used
func (db *LevelDB) Close() {
	close(db.quitChan)
	db.db.Close()
}

// Get gets the value for the given key
func (db *LevelDB) Get(key []byte) ([]byte, error) {
	return db.db.Get(key, nil)
}

func (db *LevelDB) Put(key []byte,value []byte)  error {
        if len(key)<1{
        	return nil
		}
	return db.db.Put(key,value,nil)
}

func (db *LevelDB) Delete(key []byte)  error {
	return db.db.Delete(key,nil)
}

// NewBatch constructs and returns a batch object
func (db *LevelDB) NewBatch() database.Batch {
	batch := &Batch{
		leveldb: db.db,
		batch:   new(leveldb.Batch),
	}
	return batch
}