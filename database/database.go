package database

type Database interface {
	Close()
	Get(key []byte) ([]byte, error)
	Put(key []byte, value []byte) error
	Delete(key []byte) error
	NewBatch() Batch
}

// Batch is the interface of batch for database
type Batch interface {
	Put(key []byte, value []byte)
	Delete(key []byte)
	Commit() error
	Rollback()
}
